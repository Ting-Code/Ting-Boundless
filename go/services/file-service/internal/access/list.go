package access

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/ting-boundless/boundless/pkg/errs"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/services/file-service/internal/store"
)

// ListHandler serves GET /v1/files/ (owner's files).
type ListHandler struct {
	files fileStore
	log   *slog.Logger
}

// NewList returns a list handler.
func NewList(cfg Config) http.Handler {
	return &ListHandler{files: cfg.Files, log: cfg.Log}
}

func (h *ListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id, _ := identity.FromContext(r.Context())
	if h.files == nil {
		httpx.WriteError(w, id.RequestID, errs.Internal("database_unavailable", "database not connected"))
		return
	}

	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limit = n
		}
	}

	rows, err := h.files.ListByOwner(r.Context(), id.UserID, id.TenantID, limit)
	if err != nil {
		if h.log != nil {
			h.log.Error("list files failed", slog.Any("error", err))
		}
		httpx.WriteError(w, id.RequestID, errs.Internal("list_failed", "could not list files"))
		return
	}

	files := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		files = append(files, fileJSON(row))
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"files": files})
}

func fileJSON(row store.File) map[string]any {
	return map[string]any{
		"file_id":      row.ID,
		"tenant_id":    row.TenantID,
		"owner_id":     row.OwnerID,
		"bucket":       row.Bucket,
		"object_key":   row.ObjectKey,
		"content_type": row.ContentType,
		"size_bytes":   row.SizeBytes,
		"created_at":   row.CreatedAt,
	}
}
