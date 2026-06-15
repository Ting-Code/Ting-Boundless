package access

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/ting-boundless/boundless/pkg/errs"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/services/file-service/internal/store"
)

type fileDeleter interface {
	GetByID(ctx context.Context, id string) (store.File, error)
	DeleteByID(ctx context.Context, id, ownerID, tenantID string) error
}

type blobDeleter interface {
	Enabled() bool
	DeleteObject(ctx context.Context, key string) error
}

// DeleteHandler serves DELETE /v1/files/{id}.
type DeleteHandler struct {
	files fileDeleter
	s3    blobDeleter
	log   *slog.Logger
}

// NewDelete returns a delete handler.
func NewDelete(cfg Config) http.Handler {
	return &DeleteHandler{files: cfg.Files, s3: cfg.S3, log: cfg.Log}
}

func (h *DeleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id, ok := identity.FromContext(r.Context())
	if !ok || id.UserID == "" {
		httpx.WriteError(w, id.RequestID, errs.Unauthorized("auth.unauthenticated", "authentication required"))
		return
	}
	if h.files == nil {
		httpx.WriteError(w, id.RequestID, errs.Internal("database_unavailable", "database not connected"))
		return
	}
	if h.s3 == nil || !h.s3.Enabled() {
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

	if err := h.s3.DeleteObject(r.Context(), row.ObjectKey); err != nil {
		if h.log != nil {
			h.log.Error("s3 delete failed", slog.Any("error", err), slog.String("key", row.ObjectKey))
		}
		httpx.WriteError(w, id.RequestID, errs.Internal("delete_failed", "failed to remove object"))
		return
	}

	if err := h.files.DeleteByID(r.Context(), fileID, id.UserID, id.TenantID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.WriteError(w, id.RequestID, errs.NotFound("file_not_found", "file not found"))
			return
		}
		httpx.WriteError(w, id.RequestID, errs.Internal("delete_failed", "failed to remove file metadata"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
