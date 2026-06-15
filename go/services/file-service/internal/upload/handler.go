package upload

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ting-boundless/boundless/pkg/errs"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/services/file-service/internal/store"
)

const formField = "file"

type fileMetadata interface {
	Insert(ctx context.Context, f store.File) (store.File, error)
}

type blobStore interface {
	Enabled() bool
	Bucket() string
	PutObject(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
}

// Handler serves POST /v1/files/ (multipart upload).
type Handler struct {
	files    fileMetadata
	s3       blobStore
	maxBytes int64
	log      *slog.Logger
}

// Config wires upload dependencies.
type Config struct {
	Files    fileMetadata
	S3       blobStore
	MaxBytes int64
	Log      *slog.Logger
}

// New builds an upload handler.
func New(cfg Config) http.Handler {
	max := cfg.MaxBytes
	if max <= 0 {
		max = 20 << 20 // align with nginx client_max_body_size 20m
	}
	return &Handler{
		files:    cfg.Files,
		s3:       cfg.S3,
		maxBytes: max,
		log:      cfg.Log,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	if err := r.ParseMultipartForm(h.maxBytes); err != nil {
		httpx.WriteError(w, id.RequestID, errs.BadRequest("invalid_multipart", "could not parse multipart form"))
		return
	}
	part, header, err := r.FormFile(formField)
	if err != nil {
		httpx.WriteError(w, id.RequestID, errs.BadRequest("file_required", "multipart field 'file' is required"))
		return
	}
	defer part.Close()

	if header.Size > h.maxBytes {
		httpx.WriteError(w, id.RequestID, errs.BadRequest("file_too_large", "file exceeds size limit"))
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	fileID, err := newFileID()
	if err != nil {
		httpx.WriteError(w, id.RequestID, errs.Internal("id_error", "could not allocate file id"))
		return
	}

	tenantID := id.TenantID
	objectKey := objectKey(tenantID, id.UserID, fileID, header.Filename)

	if err := h.s3.PutObject(r.Context(), objectKey, io.LimitReader(part, h.maxBytes+1), header.Size, contentType); err != nil {
		if h.log != nil {
			h.log.Error("s3 upload failed", slog.Any("error", err), slog.String("key", objectKey))
		}
		httpx.WriteError(w, id.RequestID, errs.Internal("upload_failed", "failed to store object"))
		return
	}

	row, err := h.files.Insert(r.Context(), store.File{
		ID:          fileID,
		TenantID:    tenantID,
		OwnerID:     id.UserID,
		Bucket:      h.s3.Bucket(),
		ObjectKey:   objectKey,
		ContentType: contentType,
		SizeBytes:   header.Size,
	})
	if err != nil {
		if h.log != nil {
			h.log.Error("file metadata insert failed", slog.Any("error", err), slog.String("id", fileID))
		}
		httpx.WriteError(w, id.RequestID, errs.Internal("metadata_failed", "upload stored but metadata failed"))
		return
	}

	httpx.JSON(w, http.StatusCreated, map[string]any{
		"file_id":      row.ID,
		"tenant_id":    row.TenantID,
		"owner_id":     row.OwnerID,
		"bucket":       row.Bucket,
		"object_key":   row.ObjectKey,
		"content_type": row.ContentType,
		"size_bytes":   row.SizeBytes,
		"created_at":   row.CreatedAt,
	})
}

func objectKey(tenantID, ownerID, fileID, filename string) string {
	safe := sanitizeFilename(filename)
	if tenantID == "" {
		tenantID = "_"
	}
	return path.Join(tenantID, ownerID, fileID, safe)
}

func sanitizeFilename(name string) string {
	base := filepath.Base(strings.TrimSpace(name))
	if base == "" || base == "." || base == ".." {
		return "upload"
	}
	return base
}

func newFileID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// MaxBytesFromEnv reads FILE_MAX_BYTES (default 20 MiB).
func MaxBytesFromEnv() int64 {
	if s := httpx.Env("FILE_MAX_BYTES", ""); s != "" {
		if n, err := strconv.ParseInt(s, 10, 64); err == nil && n > 0 {
			return n
		}
	}
	return 20 << 20
}
