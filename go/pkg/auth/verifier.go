// Package auth verifies standard JWTs (JWKS or dev HMAC) and maps claims to identity.
package auth

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"

	"github.com/ting-boundless/boundless/pkg/config"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/identity"
)

// Config holds OIDC verification parameters (12-Factor via env).
type Config struct {
	Issuers   []string // trusted iss (Logto + auth-service internal issuer)
	JWKSURLs  []string
	Audience  string
	DevSecret string // local dev only: HS256 when JWKS/Logto unavailable
}

// ConfigFromEnv loads verification settings from the environment.
func ConfigFromEnv() Config {
	cfg := Config{
		Audience:  httpx.Env("OIDC_AUDIENCE", ""),
		DevSecret: httpx.Env("GATEWAY_DEV_JWT_SECRET", ""),
	}
	cfg.Issuers = appendNonEmpty(cfg.Issuers, httpx.Env("OIDC_ISSUER", ""))
	cfg.Issuers = appendNonEmpty(cfg.Issuers, httpx.Env("AUTH_OIDC_ISSUER", ""))
	cfg.JWKSURLs = appendNonEmpty(cfg.JWKSURLs, httpx.Env("OIDC_JWKS_URL", ""))
	cfg.JWKSURLs = appendNonEmpty(cfg.JWKSURLs, httpx.Env("AUTH_JWKS_URL", ""))
	return cfg
}

func appendNonEmpty(list []string, v string) []string {
	if v == "" || config.IsPlaceholder(v) {
		return list
	}
	return append(list, v)
}

// Verifier validates Bearer JWTs and produces trusted identity fields.
type Verifier struct {
	cfg  Config
	jwks keyfunc.Keyfunc
	log  *slog.Logger
}

// NewVerifier builds a verifier. JWKS is optional when DevSecret is set for local dev.
func NewVerifier(ctx context.Context, cfg Config, log *slog.Logger) (*Verifier, error) {
	v := &Verifier{cfg: cfg, log: log}

	if len(cfg.JWKSURLs) > 0 {
		k, err := keyfunc.NewDefaultCtx(ctx, cfg.JWKSURLs)
		if err != nil {
			if cfg.DevSecret == "" {
				return nil, fmt.Errorf("jwks: %w", err)
			}
			log.Warn("jwks unavailable; dev HMAC fallback only", slog.Any("error", err))
		} else {
			v.jwks = k
		}
	} else if cfg.DevSecret == "" {
		log.Warn("jwt verification disabled (no JWKS URL and no GATEWAY_DEV_JWT_SECRET)")
	}

	return v, nil
}

// Enabled reports whether any verification method is configured.
func (v *Verifier) Enabled() bool {
	return v.jwks != nil || v.cfg.DevSecret != ""
}

// Verify parses and validates a raw JWT string (no "Bearer " prefix).
func (v *Verifier) Verify(tokenString string) (identity.Identity, error) {
	if !v.Enabled() {
		return identity.Identity{}, fmt.Errorf("jwt verifier not configured")
	}

	var claims jwt.MapClaims
	var err error

	if v.jwks != nil {
		claims, err = v.parseWithJWKS(tokenString)
		if err == nil {
			return mapClaims(claims), nil
		}
		if v.cfg.DevSecret == "" {
			return identity.Identity{}, err
		}
	}

	if v.cfg.DevSecret != "" {
		claims, err = v.parseWithHMAC(tokenString)
		if err != nil {
			return identity.Identity{}, err
		}
		return mapClaims(claims), nil
	}

	return identity.Identity{}, err
}

func (v *Verifier) parseWithJWKS(tokenString string) (jwt.MapClaims, error) {
	tok, err := jwt.Parse(tokenString, v.jwks.Keyfunc,
		jwt.WithValidMethods([]string{"RS256", "RS384", "RS512", "ES256", "ES384", "ES512"}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, fmt.Errorf("jwt: %w", err)
	}
	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok || !tok.Valid {
		return nil, fmt.Errorf("jwt: invalid claims")
	}
	if err := v.validateStandardClaims(claims); err != nil {
		return nil, err
	}
	return claims, nil
}

func (v *Verifier) parseWithHMAC(tokenString string) (jwt.MapClaims, error) {
	tok, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method %q", t.Method.Alg())
		}
		return []byte(v.cfg.DevSecret), nil
	}, jwt.WithExpirationRequired())
	if err != nil {
		return nil, fmt.Errorf("jwt: %w", err)
	}
	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok || !tok.Valid {
		return nil, fmt.Errorf("jwt: invalid claims")
	}
	if err := v.validateStandardClaims(claims); err != nil {
		return nil, err
	}
	return claims, nil
}

func (v *Verifier) validateStandardClaims(claims jwt.MapClaims) error {
	if len(v.cfg.Issuers) > 0 {
		iss, _ := claims["iss"].(string)
		if !issuerAllowed(iss, v.cfg.Issuers) {
			return fmt.Errorf("jwt: issuer mismatch")
		}
	}
	if v.cfg.Audience != "" {
		if !audienceContains(claims["aud"], v.cfg.Audience) {
			return fmt.Errorf("jwt: audience mismatch")
		}
	}
	if exp, err := claims.GetExpirationTime(); err == nil && exp.Before(time.Now()) {
		return fmt.Errorf("jwt: expired")
	}
	return nil
}

func audienceContains(audClaim any, want string) bool {
	switch aud := audClaim.(type) {
	case string:
		return aud == want
	case []any:
		for _, item := range aud {
			if s, ok := item.(string); ok && s == want {
				return true
			}
		}
	}
	return false
}

func issuerAllowed(iss string, allowed []string) bool {
	for _, want := range allowed {
		if iss == want {
			return true
		}
	}
	return false
}

func mapClaims(claims jwt.MapClaims) identity.Identity {
	sub, _ := claims["sub"].(string)
	userID := stringClaim(claims, "user_id")
	if userID == "" {
		userID = sub
	}
	return identity.Identity{
		UserID:   userID,
		TenantID: stringClaim(claims, "tenant_id"),
		Subject:  sub,
		Roles:    stringSliceClaim(claims, "roles"),
		Scopes:   scopesClaim(claims),
	}
}

func stringClaim(claims jwt.MapClaims, key string) string {
	v, _ := claims[key].(string)
	return v
}

func stringSliceClaim(claims jwt.MapClaims, key string) []string {
	raw, ok := claims[key]
	if !ok {
		return nil
	}
	switch v := raw.(type) {
	case []string:
		return v
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	case string:
		if v == "" {
			return nil
		}
		return strings.Split(v, ",")
	}
	return nil
}

func scopesClaim(claims jwt.MapClaims) []string {
	if s, ok := claims["scope"].(string); ok && s != "" {
		return strings.Fields(s)
	}
	return stringSliceClaim(claims, "scopes")
}
