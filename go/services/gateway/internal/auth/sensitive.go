package auth

import (
	"strings"

	"github.com/ting-boundless/boundless/pkg/httpx"
)

// SensitivePrefixes defines paths that require a fresh (non-revoked) session or subject.
type SensitivePrefixes struct {
	prefix []string
}

const defaultSensitiveEnv = "/v1/files/"

// DefaultSensitivePrefixes returns built-in sensitive path prefixes.
func DefaultSensitivePrefixes() SensitivePrefixes {
	return ParseSensitivePrefixes(defaultSensitiveEnv)
}

// ResolveSensitivePrefixes loads GATEWAY_SENSITIVE_PREFIXES or defaults.
// Set GATEWAY_SENSITIVE_PREFIXES= to disable revocation checks on all paths.
func ResolveSensitivePrefixes() SensitivePrefixes {
	raw := httpx.Env("GATEWAY_SENSITIVE_PREFIXES", "")
	if raw == "-" || raw == "none" {
		return SensitivePrefixes{}
	}
	return ParseSensitivePrefixes(raw)
}

// ParseSensitivePrefixes parses comma-separated path prefixes.
func ParseSensitivePrefixes(raw string) SensitivePrefixes {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ParseSensitivePrefixes(defaultSensitiveEnv)
	}
	out := SensitivePrefixes{}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if !strings.HasPrefix(part, "/") {
			part = "/" + part
		}
		out.prefix = append(out.prefix, part)
	}
	return out
}

// RequiresRevocationCheck reports whether path needs revocation lookup.
func (p SensitivePrefixes) RequiresRevocationCheck(path string) bool {
	for _, prefix := range p.prefix {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

// PrefixPaths returns configured prefixes (for logging/tests).
func (p SensitivePrefixes) PrefixPaths() []string {
	return append([]string(nil), p.prefix...)
}
