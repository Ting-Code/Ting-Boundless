package miniprogram

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/ting-boundless/boundless/pkg/audit"
	"github.com/ting-boundless/boundless/pkg/errs"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/services/auth-service/internal/wechat"
)

type identityResolver interface {
	ResolveWeChatUser(ctx context.Context, openid, unionid string) (string, error)
}

type tokenIssuer interface {
	AccessToken(userID, tenantID string, roles []string, ttl time.Duration) (string, error)
}

// Handler serves mini-program login.
type Handler struct {
	wechat   *wechat.Client
	ident    identityResolver
	issuer   tokenIssuer
	audit    audit.Emitter
	ttl      time.Duration
	log      *slog.Logger
}

// Config wires dependencies for mini-program login.
type Config struct {
	WeChat   *wechat.Client
	Identity identityResolver
	Issuer   tokenIssuer
	Audit    audit.Emitter
	TTL      time.Duration
	Log      *slog.Logger
}

// NewHandler builds a login handler.
func NewHandler(cfg Config) *Handler {
	ttl := cfg.TTL
	if ttl <= 0 {
		ttl = time.Hour
	}
	return &Handler{
		wechat:   cfg.WeChat,
		ident:    cfg.Identity,
		issuer:   cfg.Issuer,
		audit:    cfg.Audit,
		ttl:      ttl,
		log:      cfg.Log,
	}
}

type loginRequest struct {
	Code string `json:"code"`
}

type loginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	UserID      string `json:"user_id"`
}

// Login handles POST /v1/auth/miniprogram/login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	rid := r.Header.Get("X-Request-Id")

	if h == nil || h.wechat == nil || h.issuer == nil {
		httpx.WriteError(w, rid, errs.Internal("auth.unavailable", "auth service not configured"))
		return
	}
	if h.ident == nil {
		httpx.WriteError(w, rid, errs.Internal("auth.unavailable", "user database not configured"))
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		httpx.WriteError(w, rid, errs.BadRequest("auth.invalid_request", "code is required"))
		return
	}

	sess, err := h.wechat.Code2Session(r.Context(), req.Code)
	if err != nil {
		h.log.Warn("wechat code2session failed", slog.Any("error", err))
		httpx.WriteError(w, rid, errs.Unauthorized("auth.wechat_failed", "wechat login failed"))
		return
	}

	userID, err := h.ident.ResolveWeChatUser(r.Context(), sess.OpenID, sess.UnionID)
	if err != nil {
		h.log.Error("resolve wechat user failed", slog.Any("error", err))
		httpx.WriteError(w, rid, errs.Internal("auth.identity_failed", "could not resolve user"))
		return
	}

	tok, err := h.issuer.AccessToken(userID, "", []string{"user"}, h.ttl)
	if err != nil {
		h.log.Error("mint access token failed", slog.Any("error", err))
		httpx.WriteError(w, rid, errs.Internal("auth.token_failed", "could not issue token"))
		return
	}

	if h.audit != nil {
		ev := audit.NewEvent(audit.SourceIdP, "user.login.success")
		ev.ActorUserID = userID
		ev.Subject = "user:" + userID
		ev.Data = map[string]any{
			"channel": "wechat_miniprogram",
			"openid":  sess.OpenID,
		}
		if sess.UnionID != "" {
			ev.Data["unionid"] = sess.UnionID
		}
		if err := h.audit.Emit(r.Context(), ev); err != nil {
			h.log.Warn("audit emit failed", slog.Any("error", err))
		}
	}

	httpx.JSON(w, http.StatusOK, loginResponse{
		AccessToken: tok,
		TokenType:   "Bearer",
		ExpiresIn:   int64(h.ttl.Seconds()),
		UserID:      userID,
	})
}
