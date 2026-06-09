package httpx

import (
	"regexp"
	"strings"
)

var (
	uuidSegment = regexp.MustCompile(`(?i)[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
	idSegment   = regexp.MustCompile(`(?i)^[0-9a-z]{12,}$`)
	numSegment  = regexp.MustCompile(`^[0-9]+$`)
)

// NormalizePath replaces high-cardinality path segments with :id for access logs.
func NormalizePath(path string) string {
	if path == "" || path == "/" {
		return path
	}
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "" {
			continue
		}
		if uuidSegment.MatchString(part) || idSegment.MatchString(part) || numSegment.MatchString(part) {
			parts[i] = ":id"
		}
	}
	out := strings.Join(parts, "/")
	if !strings.HasPrefix(out, "/") {
		out = "/" + out
	}
	return out
}
