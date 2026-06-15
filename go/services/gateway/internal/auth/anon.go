package auth

import (
	"strings"

	"github.com/ting-boundless/boundless/pkg/httpx"
)

// AnonPrefixes defines paths that may pass through the Gateway without end-user
// credentials. Use exact rules for BFF/OIDC endpoints; use prefix rules only
// where a subtree is intentionally public (admin static, auth integration).
type AnonPrefixes struct {
	exact  []string
	prefix []string
}

// defaultAnonEnv is the built-in GATEWAY_ANON_PREFIXES when env is unset.
// Syntax: =/path for exact match; /path for prefix match (see ParseAnonPrefixes).
const defaultAnonEnv = "=/healthz,=/readyz,=/metrics,=/sign-in,=/callback,=/sign-out,=/internal/webhooks/logto,=/,/v1/business/ping,/admin,/v1/auth/"

// DefaultAnonPrefixes returns the production-safe anonymous path rules.
// Does not include /sign-in/dev; use ResolveAnonPrefixes for env + dev toggles.
func DefaultAnonPrefixes() AnonPrefixes {
	return ParseAnonPrefixes(defaultAnonEnv)
}

// ResolveAnonPrefixes loads rules from GATEWAY_ANON_PREFIXES (or defaults) and
// appends /sign-in/dev when GATEWAY_BFF_DEV_LOGIN=true.
func ResolveAnonPrefixes() AnonPrefixes {
	rules := ParseAnonPrefixes(httpx.Env("GATEWAY_ANON_PREFIXES", ""))
	if httpx.Env("GATEWAY_BFF_DEV_LOGIN", "") == "true" {
		rules = rules.withExact("/sign-in/dev")
	}
	return rules
}

// AnonPrefixesFromEnv loads comma-separated rules from GATEWAY_ANON_PREFIXES only.
// Prefer ResolveAnonPrefixes in gateway main.
func AnonPrefixesFromEnv() AnonPrefixes {
	return ParseAnonPrefixes(httpx.Env("GATEWAY_ANON_PREFIXES", ""))
}

// ParseAnonPrefixes parses GATEWAY_ANON_PREFIXES.
//
//   - =/path — exact path only (e.g. =/sign-in matches /sign-in but not /sign-in/dev)
//   - /path  — prefix match (e.g. /admin matches /admin/items)
//
// Empty input returns DefaultAnonPrefixes().
func ParseAnonPrefixes(raw string) AnonPrefixes {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ParseAnonPrefixes(defaultAnonEnv)
	}
	out := AnonPrefixes{}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.HasPrefix(part, "=") {
			path := strings.TrimSpace(part[1:])
			if path != "" && !strings.HasPrefix(path, "/") {
				path = "/" + path
			}
			if path != "" {
				out.exact = append(out.exact, path)
			}
			continue
		}
		if !strings.HasPrefix(part, "/") {
			part = "/" + part
		}
		out.prefix = append(out.prefix, part)
	}
	if len(out.exact) == 0 && len(out.prefix) == 0 {
		return ParseAnonPrefixes(defaultAnonEnv)
	}
	return out
}

func (p AnonPrefixes) withExact(path string) AnonPrefixes {
	for _, e := range p.exact {
		if e == path {
			return p
		}
	}
	p.exact = append(p.exact, path)
	return p
}

// IsEmpty reports whether no exact or prefix rules are configured.
func (p AnonPrefixes) IsEmpty() bool {
	return len(p.exact) == 0 && len(p.prefix) == 0
}

// Allows reports whether path may proceed without authentication.
func (p AnonPrefixes) Allows(path string) bool {
	for _, e := range p.exact {
		if path == e {
			return true
		}
	}
	for _, prefix := range p.prefix {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

// ExactPaths returns a copy of exact-match paths (for logging/tests).
func (p AnonPrefixes) ExactPaths() []string {
	return append([]string(nil), p.exact...)
}

// PrefixPaths returns a copy of prefix-match paths (for logging/tests).
func (p AnonPrefixes) PrefixPaths() []string {
	return append([]string(nil), p.prefix...)
}
