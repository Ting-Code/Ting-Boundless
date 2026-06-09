package logger

import (
	"log/slog"
	"net/http"

	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/pkg/trace"
)

// RequestAttrs returns ECS fields from trusted internal request headers.
func RequestAttrs(r *http.Request) []slog.Attr {
	attrs := []slog.Attr{
		slog.String("request_id", r.Header.Get(identity.HeaderRequestID)),
	}
	if tid := trace.TraceIDFromParent(r.Header.Get(trace.HeaderTraceparent)); tid != "" {
		attrs = append(attrs, slog.String("trace_id", tid))
	}
	id := identity.FromHeaders(r.Header)
	if id.UserID != "" {
		attrs = append(attrs, slog.String("user_id", id.UserID))
	}
	if id.TenantID != "" {
		attrs = append(attrs, slog.String("tenant_id", id.TenantID))
	}
	return attrs
}

// WithRequest returns a child logger enriched with request correlation fields.
func WithRequest(base *slog.Logger, r *http.Request) *slog.Logger {
	attrs := RequestAttrs(r)
	args := make([]any, 0, len(attrs)*2)
	for _, a := range attrs {
		args = append(args, a.Key, a.Value.Any())
	}
	return base.With(args...)
}
