package auth

import (
	"strings"

	"github.com/ting-boundless/boundless/pkg/httpx"
)

// AnonPrefixes lists path prefixes that may pass through the Gateway without
// credentials. All other paths require a valid Bearer token or BFF session cookie.
type AnonPrefixes []string

// DefaultAnonPrefixes are used when GATEWAY_ANON_PREFIXES is unset.
func DefaultAnonPrefixes() AnonPrefixes {
	return AnonPrefixes{
		"/healthz",
		"/readyz",
		"/metrics",
		"/sign-in",
		"/callback",
		"/sign-out",
		"/admin",
		"/v1/business/ping",
		"/v1/auth/",
	}
}

// AnonPrefixesFromEnv loads comma-separated prefixes from GATEWAY_ANON_PREFIXES.
func AnonPrefixesFromEnv() AnonPrefixes {
	return ParseAnonPrefixes(httpx.Env("GATEWAY_ANON_PREFIXES", ""))
}

// ParseAnonPrefixes parses a comma-separated prefix list. Empty input returns defaults.
func ParseAnonPrefixes(raw string) AnonPrefixes {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return DefaultAnonPrefixes()
	}
	parts := strings.Split(raw, ",")
	out := make(AnonPrefixes, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		out = append(out, p)
	}
	if len(out) == 0 {
		return DefaultAnonPrefixes()
	}
	return out
}

// Allows reports whether path may proceed without authentication.
func (p AnonPrefixes) Allows(path string) bool {
	for _, prefix := range p {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}
