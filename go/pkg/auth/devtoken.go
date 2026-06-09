package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// DevToken builds a signed HS256 JWT for local testing (GATEWAY_DEV_JWT_SECRET).
func DevToken(cfg Config, userID, tenantID string, roles []string, ttl time.Duration) (string, error) {
	if cfg.DevSecret == "" {
		return "", jwt.ErrTokenMalformed
	}
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":       userID,
		"user_id":   userID,
		"tenant_id": tenantID,
		"roles":     roles,
		"iat":       now.Unix(),
		"exp":       now.Add(ttl).Unix(),
	}
	if cfg.Issuer != "" {
		claims["iss"] = cfg.Issuer
	}
	if cfg.Audience != "" {
		claims["aud"] = cfg.Audience
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(cfg.DevSecret))
}
