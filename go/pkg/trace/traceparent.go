// Package trace implements W3C Trace Context helpers for log correlation and propagation.
package trace

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

// HeaderTraceparent is the W3C Trace Context header name.
const HeaderTraceparent = "traceparent"

// TraceIDFromParent extracts the 32-hex trace id from a traceparent value.
// Returns empty string when the value is missing or invalid.
func TraceIDFromParent(traceparent string) string {
	parts := strings.Split(strings.TrimSpace(traceparent), "-")
	if len(parts) < 2 {
		return ""
	}
	traceID := parts[1]
	if len(traceID) != 32 {
		return ""
	}
	for _, c := range traceID {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return ""
		}
	}
	return strings.ToLower(traceID)
}

// NewTraceparent returns a valid W3C traceparent for a new root span.
func NewTraceparent() string {
	var traceID [16]byte
	var spanID [8]byte
	_, _ = rand.Read(traceID[:])
	_, _ = rand.Read(spanID[:])
	return fmt.Sprintf("00-%s-%s-01", hex.EncodeToString(traceID[:]), hex.EncodeToString(spanID[:]))
}

// EnsureRequest sets traceparent on the request when absent and mirrors it on the response.
func EnsureRequest(h interface {
	Get(key string) string
	Set(key, value string)
}, setResponseHeader func(key, value string)) string {
	tp := h.Get(HeaderTraceparent)
	if tp == "" {
		tp = NewTraceparent()
		h.Set(HeaderTraceparent, tp)
	}
	if setResponseHeader != nil {
		setResponseHeader(HeaderTraceparent, tp)
	}
	return tp
}
