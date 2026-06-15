package bff

import (
	"strings"

	"github.com/ting-boundless/boundless/pkg/httpx"
)

// sessionCookieSecure controls the Set-Cookie Secure attribute (HTTPS-only).
// Explicit SESSION_COOKIE_SECURE wins; otherwise true when APP_ENV=production.
func sessionCookieSecure() bool {
	switch strings.ToLower(strings.TrimSpace(httpx.Env("SESSION_COOKIE_SECURE", ""))) {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	default:
		return strings.EqualFold(httpx.Env("APP_ENV", ""), "production")
	}
}
