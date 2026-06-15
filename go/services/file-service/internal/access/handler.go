package access

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/ting-boundless/boundless/pkg/errs"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/services/file-service/internal/store"
)

type fileStore interface {
	GetByID(ctx context.Context, id string) (store.File, error)
	ListByOwner(ctx context.Context, ownerID, tenantID string, limit int) ([]store.File, error)
	DeleteByID(ctx context.Context, id, ownerID, tenantID string) error
}

type blobStore interface {
	Enabled() bool
	GetObject(ctx context.Context, key string) (*http.Response, error)
	PresignGetURL(key string, expires time.Duration) (string, time.Time, error)
	DeleteObject(ctx context.Context, key string) error
}

// MetaHandler serves GET /v1/files/{id}.
type MetaHandler struct {
	files fileStore
	log   *slog.Logger
}

// DownloadHandler serves GET /v1/files/{id}/download.
type DownloadHandler struct {
	files fileStore
	s3    blobStore
	log   *slog.Logger
}

// URLHandler serves GET /v1/files/{id}/url (presigned GET).
type URLHandler struct {
	files         fileStore
	s3            blobStore
	defaultExpiry time.Duration
	maxExpiry     time.Duration
	log           *slog.Logger
}

// Config wires file access handlers.
type Config struct {
	Files         fileStore
	S3            blobStore
	DefaultExpiry time.Duration
	Log           *slog.Logger
}

// NewMeta returns metadata handler.
func NewMeta(cfg Config) http.Handler {
	return &MetaHandler{files: cfg.Files, log: cfg.Log}
}

// NewDownload returns download handler.
func NewDownload(cfg Config) http.Handler {
	return &DownloadHandler{files: cfg.Files, s3: cfg.S3, log: cfg.Log}
}

// NewURL returns presigned URL handler.
func NewURL(cfg Config) http.Handler {
	exp := cfg.DefaultExpiry
	if exp <= 0 {
		exp = time.Hour
	}
	return &URLHandler{
		files:         cfg.Files,
		s3:            cfg.S3,
		defaultExpiry: exp,
		maxExpiry:     7 * 24 * time.Hour,
		log:           cfg.Log,
	}
}

func (h *MetaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id, ok := identity.FromContext(r.Context())
	if !ok || id.UserID == "" {
		httpx.WriteError(w, id.RequestID, errs.Unauthorized("auth.unauthenticated", "authentication required"))
		return
	}
	if h.files == nil {
		httpx.WriteError(w, id.RequestID, errs.Internal("database_unavailable", "database not connected"))
		return
	}

	fileID := r.PathValue("id")
	if fileID == "" {
		httpx.WriteError(w, id.RequestID, errs.BadRequest("file_id_required", "file id is required"))
		return
	}

	row, err := h.files.GetByID(r.Context(), fileID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.WriteError(w, id.RequestID, errs.NotFound("file_not_found", "file not found"))
			return
		}
		httpx.WriteError(w, id.RequestID, errs.Internal("lookup_failed", "could not load file metadata"))
		return
	}
	if !canAccess(id, row) {
		httpx.WriteError(w, id.RequestID, errs.Forbidden("file_forbidden", "not allowed to access this file"))
		return
	}

	httpx.JSON(w, http.StatusOK, fileJSON(row))
}

func (h *DownloadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id, ok := identity.FromContext(r.Context())
	if !ok || id.UserID == "" {
		httpx.WriteError(w, id.RequestID, errs.Unauthorized("auth.unauthenticated", "authentication required"))
		return
	}
	if h.files == nil {
		httpx.WriteError(w, id.RequestID, errs.Internal("database_unavailable", "database not connected"))
		return
	}
	if !h.s3.Enabled() {
		httpx.WriteError(w, id.RequestID, errs.Internal("storage_unavailable", "object storage not configured"))
		return
	}

	fileID := r.PathValue("id")
	if fileID == "" {
		httpx.WriteError(w, id.RequestID, errs.BadRequest("file_id_required", "file id is required"))
		return
	}

	row, err := h.files.GetByID(r.Context(), fileID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.WriteError(w, id.RequestID, errs.NotFound("file_not_found", "file not found"))
			return
		}
		httpx.WriteError(w, id.RequestID, errs.Internal("lookup_failed", "could not load file metadata"))
		return
	}
	if !canAccess(id, row) {
		httpx.WriteError(w, id.RequestID, errs.Forbidden("file_forbidden", "not allowed to access this file"))
		return
	}

	resp, err := h.s3.GetObject(r.Context(), row.ObjectKey)
	if err != nil {
		if h.log != nil {
			h.log.Error("s3 download failed", slog.Any("error", err), slog.String("key", row.ObjectKey))
		}
		httpx.WriteError(w, id.RequestID, errs.Internal("download_failed", "failed to fetch object"))
		return
	}
	defer resp.Body.Close()

	if ct := row.ContentType; ct != "" {
		w.Header().Set("Content-Type", ct)
	} else if ct := resp.Header.Get("Content-Type"); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	if row.SizeBytes > 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(row.SizeBytes, 10))
	}
	name := filepath.Base(row.ObjectKey)
	if name != "" && name != "." {
		w.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)
	}
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, resp.Body)
}

func (h *URLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id, ok := identity.FromContext(r.Context())
	if !ok || id.UserID == "" {
		httpx.WriteError(w, id.RequestID, errs.Unauthorized("auth.unauthenticated", "authentication required"))
		return
	}
	if h.files == nil {
		httpx.WriteError(w, id.RequestID, errs.Internal("database_unavailable", "database not connected"))
		return
	}
	if !h.s3.Enabled() {
		httpx.WriteError(w, id.RequestID, errs.Internal("storage_unavailable", "object storage not configured"))
		return
	}

	fileID := r.PathValue("id")
	if fileID == "" {
		httpx.WriteError(w, id.RequestID, errs.BadRequest("file_id_required", "file id is required"))
		return
	}

	row, err := h.files.GetByID(r.Context(), fileID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.WriteError(w, id.RequestID, errs.NotFound("file_not_found", "file not found"))
			return
		}
		httpx.WriteError(w, id.RequestID, errs.Internal("lookup_failed", "could not load file metadata"))
		return
	}
	if !canAccess(id, row) {
		httpx.WriteError(w, id.RequestID, errs.Forbidden("file_forbidden", "not allowed to access this file"))
		return
	}

	expiry := h.defaultExpiry
	if raw := r.URL.Query().Get("expires"); raw != "" {
		sec, err := strconv.Atoi(raw)
		if err != nil || sec <= 0 {
			httpx.WriteError(w, id.RequestID, errs.BadRequest("invalid_expires", "expires must be a positive number of seconds"))
			return
		}
		expiry = time.Duration(sec) * time.Second
	}
	if expiry > h.maxExpiry {
		expiry = h.maxExpiry
	}

	signed, expiresAt, err := h.s3.PresignGetURL(row.ObjectKey, expiry)
	if err != nil {
		if h.log != nil {
			h.log.Error("presign failed", slog.Any("error", err), slog.String("key", row.ObjectKey))
		}
		httpx.WriteError(w, id.RequestID, errs.Internal("presign_failed", "could not create download URL"))
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{
		"file_id":    row.ID,
		"url":        signed,
		"expires_at": expiresAt,
		"method":     "GET",
	})
}

func canAccess(id identity.Identity, f store.File) bool {
	if id.UserID != f.OwnerID {
		return false
	}
	if id.TenantID != "" && f.TenantID != "" && id.TenantID != f.TenantID {
		return false
	}
	return true
}
