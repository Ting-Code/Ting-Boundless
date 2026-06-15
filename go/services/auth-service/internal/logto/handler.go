package logto

import (
	"context"
	"io"
	"log/slog"
	"net/http"

	"github.com/ting-boundless/boundless/pkg/audit"
	"github.com/ting-boundless/boundless/pkg/errs"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/pkg/revocation"
	"github.com/ting-boundless/boundless/services/auth-service/internal/store"
)

// Handler processes Logto webhook deliveries.
type Handler struct {
	signingKey string
	skipVerify bool
	identities *store.IdentityStore
	deliveries *store.Deliveries
	audit      audit.Emitter
	revocations *revocation.Store
	log        *slog.Logger
}

// Config wires Logto webhook handling.
type Config struct {
	SigningKey string
	SkipVerify bool
	Identities *store.IdentityStore
	Deliveries *store.Deliveries
	Audit      audit.Emitter
	Revocations *revocation.Store
	Log        *slog.Logger
}

// NewHandler builds a Logto webhook handler.
func NewHandler(cfg Config) *Handler {
	return &Handler{
		signingKey: cfg.SigningKey,
		skipVerify: cfg.SkipVerify,
		identities: cfg.Identities,
		deliveries: cfg.Deliveries,
		audit:      cfg.Audit,
		revocations: cfg.Revocations,
		log:        cfg.Log,
	}
}

// ServeHTTP handles POST /internal/webhooks/logto.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rid := r.Header.Get(identity.HeaderRequestID)

	raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		httpx.WriteError(w, rid, errs.BadRequest("invalid_body", "could not read request body"))
		return
	}

	if h.signingKey != "" {
		if err := VerifySignature(h.signingKey, raw, r.Header.Get(signatureHeader)); err != nil {
			h.log.Warn("logto webhook signature failed", slog.Any("error", err))
			httpx.WriteError(w, rid, errs.Unauthorized("webhook.invalid_signature", "invalid webhook signature"))
			return
		}
	} else if !h.skipVerify {
		httpx.WriteError(w, rid, errs.Internal("webhook.not_configured", "LOGTO_WEBHOOK_SIGNING_KEY not configured"))
		return
	}

	payload, err := ParsePayload(raw)
	if err != nil {
		httpx.WriteError(w, rid, errs.BadRequest("invalid_payload", "malformed webhook payload"))
		return
	}

	if h.deliveries != nil {
		first, err := h.deliveries.TryRecord(r.Context(), DeliveryKey(payload))
		if err != nil {
			h.log.Error("webhook idempotency failed", slog.Any("error", err))
			httpx.WriteError(w, rid, errs.Internal("webhook.idempotency_failed", "could not record delivery"))
			return
		}
		if !first {
			httpx.JSON(w, http.StatusAccepted, map[string]string{"status": "duplicate"})
			return
		}
	}

	platformUserID := ""
	if logtoUID := logtoUserID(payload); logtoUID != "" && h.identities != nil {
		uid, err := h.identities.ResolveLogtoUser(r.Context(), logtoUID)
		if err != nil {
			h.log.Error("resolve logto user failed", slog.Any("error", err), slog.String("logto_user_id", logtoUID))
		} else {
			platformUserID = uid
		}
	}

	if ev, ok := ToAuditEvent(payload, platformUserID); ok && h.audit != nil {
		if err := h.audit.Emit(r.Context(), ev); err != nil {
			h.log.Error("audit emit failed", slog.Any("error", err), slog.String("type", ev.Type))
			httpx.WriteError(w, rid, errs.Internal("audit.emit_failed", "could not emit audit event"))
			return
		}
	}

	if err := h.applyRevocation(r.Context(), payload, platformUserID); err != nil {
		h.log.Error("revocation apply failed", slog.Any("error", err), slog.String("event", payload.Event))
		httpx.WriteError(w, rid, errs.Internal("revocation.failed", "could not apply revocation"))
		return
	}

	httpx.JSON(w, http.StatusAccepted, map[string]string{"status": "accepted"})
}

func (h *Handler) applyRevocation(ctx context.Context, payload Payload, platformUserID string) error {
	if h.revocations == nil || !h.revocations.Enabled() {
		return nil
	}
	switch payload.Event {
	case "User.Deleted":
		if logtoUID := logtoUserID(payload); logtoUID != "" {
			if err := h.revocations.RevokeSubject(ctx, logtoUID); err != nil {
				return err
			}
		}
		if platformUserID != "" {
			if err := h.revocations.RevokeSubject(ctx, platformUserID); err != nil {
				return err
			}
		}
	}
	return nil
}
