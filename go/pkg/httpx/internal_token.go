package httpx

import (
	"log/slog"
	"strings"
)

// RequireInternalToken reports whether INTERNAL_API_TOKEN must be non-empty.
// True when APP_ENV=production or REQUIRE_INTERNAL_TOKEN=true.
func RequireInternalToken() bool {
	if Env("REQUIRE_INTERNAL_TOKEN", "") == "true" {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(Env("APP_ENV", "")), "production")
}

// LoadInternalToken reads INTERNAL_API_TOKEN. When RequireInternalToken() is true
// and the token is empty, logs an error and returns false so callers can exit.
// When optional in dev, logs a warning once when unset.
func LoadInternalToken(log *slog.Logger) (token string, ok bool) {
	token = strings.TrimSpace(Env("INTERNAL_API_TOKEN", ""))
	if RequireInternalToken() && token == "" {
		if log != nil {
			log.Error("INTERNAL_API_TOKEN is required",
				slog.String("hint", "set APP_ENV=production only with a strong token, or set REQUIRE_INTERNAL_TOKEN=true"),
			)
		}
		return "", false
	}
	if token == "" && log != nil {
		log.Warn("INTERNAL_API_TOKEN unset; gateway trust checks disabled (dev only)")
	}
	return token, true
}
