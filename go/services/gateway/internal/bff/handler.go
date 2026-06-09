// Package bff implements Gateway OIDC BFF routes for Web/Admin cookie sessions.
package bff

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ting-boundless/boundless/pkg/auth"
	"github.com/ting-boundless/boundless/pkg/errs"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/pkg/oidc"
	"github.com/ting-boundless/boundless/services/gateway/internal/session"
)

// Handler serves /sign-in, /callback, /sign-out (and optional dev login).
type Handler struct {
	oidc      *oidc.Client
	oidcCfg   oidc.ClientConfig
	sessions  *session.Store
	verifier  *auth.Verifier
	authCfg   auth.Config
	devLogin  bool
	log       *slog.Logger
}

// New builds a BFF handler.
func New(
	oidcCfg oidc.ClientConfig,
	sessions *session.Store,
	verifier *auth.Verifier,
	authCfg auth.Config,
	log *slog.Logger,
) *Handler {
	return &Handler{
		oidc:     oidc.NewClient(oidcCfg),
		oidcCfg:  oidcCfg,
		sessions: sessions,
		verifier: verifier,
		authCfg:  authCfg,
		devLogin: httpx.Env("GATEWAY_BFF_DEV_LOGIN", "") == "true",
		log:      log,
	}
}

// SignIn starts the OIDC authorization code flow.
func (h *Handler) SignIn(w http.ResponseWriter, r *http.Request) {
	if !h.oidcCfg.Ready() {
		errs.Write(w, "", errs.Internal("oidc_not_configured", "OIDC BFF not configured; set OIDC_CLIENT_ID/SECRET or use GATEWAY_BFF_DEV_LOGIN"))
		return
	}
	if !h.sessions.Enabled() {
		errs.Write(w, "", errs.Internal("session_unavailable", "redis required for BFF sessions"))
		return
	}

	returnTo := safeReturnTo(r.URL.Query().Get("return_to"), "/")
	state, err := randomHex(16)
	if err != nil {
		errs.Write(w, "", errs.Internal("state_error", "could not start login"))
		return
	}
	nonce, err := randomHex(16)
	if err != nil {
		errs.Write(w, "", errs.Internal("state_error", "could not start login"))
		return
	}

	if err := h.sessions.SavePending(r.Context(), state, session.PendingLogin{
		ReturnTo: returnTo,
		Nonce:    nonce,
	}); err != nil {
		h.log.Error("save oidc state", slog.Any("error", err))
		errs.Write(w, "", errs.Internal("state_error", "could not start login"))
		return
	}

	authorizeURL, err := h.oidc.AuthorizeURL(state, nonce)
	if err != nil {
		errs.Write(w, "", errs.Internal("oidc_error", "could not build authorize URL"))
		return
	}
	http.Redirect(w, r, authorizeURL, http.StatusFound)
}

// Callback completes OIDC and sets the HttpOnly session cookie.
func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	if errMsg := r.URL.Query().Get("error"); errMsg != "" {
		desc := r.URL.Query().Get("error_description")
		errs.Write(w, "", errs.Unauthorized("oidc_denied", errMsg+": "+desc))
		return
	}

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" || state == "" {
		errs.Write(w, "", errs.BadRequest("invalid_callback", "missing code or state"))
		return
	}

	pending, err := h.sessions.ConsumePending(r.Context(), state)
	if err != nil {
		errs.Write(w, "", errs.Unauthorized("invalid_state", "login session expired or invalid"))
		return
	}

	tokens, err := h.oidc.ExchangeCode(r.Context(), code)
	if err != nil {
		h.log.Error("token exchange", slog.Any("error", err))
		errs.Write(w, "", errs.Unauthorized("token_exchange_failed", "could not complete login"))
		return
	}

	token := tokens.AccessToken
	if token == "" {
		token = tokens.IDToken
	}
	id, err := h.verifier.Verify(token)
	if err != nil {
		h.log.Error("verify access token", slog.Any("error", err))
		errs.Write(w, "", errs.Unauthorized("invalid_token", "invalid token from IdP"))
		return
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	if tokens.ExpiresIn > 0 {
		expiresAt = time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)
	}

	sid, err := h.sessions.Create(r.Context(), session.Data{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		Identity:     id,
		ExpiresAt:    expiresAt,
	})
	if err != nil {
		h.log.Error("create session", slog.Any("error", err))
		errs.Write(w, "", errs.Internal("session_error", "could not create session"))
		return
	}

	h.setSessionCookie(w, sid, expiresAt)
	http.Redirect(w, r, pending.ReturnTo, http.StatusFound)
}

// SignOut clears the BFF session cookie.
func (h *Handler) SignOut(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(h.sessions.CookieName()); err == nil && c.Value != "" {
		_ = h.sessions.Delete(r.Context(), c.Value)
	}
	h.clearSessionCookie(w)
	returnTo := safeReturnTo(r.URL.Query().Get("return_to"), "/")
	http.Redirect(w, r, returnTo, http.StatusFound)
}

// DevSignIn creates a local session without Logto (GATEWAY_BFF_DEV_LOGIN=true only).
func (h *Handler) DevSignIn(w http.ResponseWriter, r *http.Request) {
	if !h.devLogin {
		http.NotFound(w, r)
		return
	}
	if !h.sessions.Enabled() {
		errs.Write(w, "", errs.Internal("session_unavailable", "redis required for BFF sessions"))
		return
	}
	if h.authCfg.DevSecret == "" {
		errs.Write(w, "", errs.Internal("dev_auth_unconfigured", "set GATEWAY_DEV_JWT_SECRET for dev login"))
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = "dev-user"
	}
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		tenantID = "dev-tenant"
	}
	returnTo := safeReturnTo(r.URL.Query().Get("return_to"), "/")

	tok, err := auth.DevToken(h.authCfg, userID, tenantID, []string{"user"}, 24*time.Hour)
	if err != nil {
		errs.Write(w, "", errs.Internal("dev_token_error", "could not mint dev token"))
		return
	}
	id, err := h.verifier.Verify(tok)
	if err != nil {
		errs.Write(w, "", errs.Internal("dev_token_error", "dev token verification failed"))
		return
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	sid, err := h.sessions.Create(r.Context(), session.Data{
		AccessToken: tok,
		Identity:    id,
		ExpiresAt:   expiresAt,
	})
	if err != nil {
		errs.Write(w, "", errs.Internal("session_error", "could not create session"))
		return
	}

	h.setSessionCookie(w, sid, expiresAt)
	http.Redirect(w, r, returnTo, http.StatusFound)
}

func (h *Handler) setSessionCookie(w http.ResponseWriter, sessionID string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.sessions.CookieName(),
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
		// Secure: true in production behind HTTPS
	})
}

func (h *Handler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.sessions.CookieName(),
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
	})
}

func safeReturnTo(raw, fallback string) string {
	if raw == "" {
		return fallback
	}
	if !strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, "//") {
		return fallback
	}
	if u, err := url.Parse(raw); err != nil || u.Host != "" {
		return fallback
	}
	return raw
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// IdentityFromRequest resolves identity from session cookie (used by auth middleware).
func IdentityFromRequest(r *http.Request, sessions *session.Store, verifier *auth.Verifier) (identity.Identity, error) {
	if sessions == nil || !sessions.Enabled() || verifier == nil {
		return identity.Identity{}, fmt.Errorf("session auth unavailable")
	}
	c, err := r.Cookie(sessions.CookieName())
	if err != nil || c.Value == "" {
		return identity.Identity{}, fmt.Errorf("no session cookie")
	}
	data, err := sessions.Get(r.Context(), c.Value)
	if err != nil {
		return identity.Identity{}, err
	}
	id, err := verifier.Verify(data.AccessToken)
	if err != nil {
		_ = sessions.Delete(r.Context(), c.Value)
		return identity.Identity{}, err
	}
	return id, nil
}
