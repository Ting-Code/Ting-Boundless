package httpx

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"

	"github.com/ting-boundless/boundless/pkg/errs"
	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/pkg/logger"
	"github.com/ting-boundless/boundless/pkg/trace"
)

// RequestID ensures every request has a request id in context and response.
// A client-supplied X-Request-Id is NOT trusted at the edge; the Gateway should
// strip it first. Internally, an existing id is preserved as the correlation id.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get(identity.HeaderRequestID)
		if rid == "" {
			rid = NewRequestID()
			r.Header.Set(identity.HeaderRequestID, rid)
		}
		w.Header().Set(identity.HeaderRequestID, rid)
		next.ServeHTTP(w, r)
	})
}

// TraceContext ensures every request has a W3C traceparent for log correlation
// and downstream propagation. Incoming traceparent is preserved; absent values
// get a new root span at this hop.
func TraceContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		trace.EnsureRequest(r.Header, w.Header().Set)
		next.ServeHTTP(w, r)
	})
}

// AccessLog logs one structured line per request with ECS correlation fields.
func AccessLog(base *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			l := logger.WithRequest(base, r)
			ctx := logger.Into(r.Context(), l)

			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r.WithContext(ctx))

			l.Info("http_request",
				slog.String("http.request.method", r.Method),
				slog.String("url.path", NormalizePath(r.URL.Path)),
				slog.Int("http.response.status_code", sw.status),
				slog.Duration("event.duration", time.Since(start)),
			)
		})
	}
}

// Recover converts panics into 500s instead of crashing the server.
func Recover(base *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					base.Error("panic recovered", slog.Any("panic", rec))
					rid := r.Header.Get(identity.HeaderRequestID)
					WriteError(w, rid, errs.Internal("internal", "internal server error"))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// Chain applies middlewares in order (outermost first).
func Chain(h http.Handler, mw ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mw) - 1; i >= 0; i-- {
		h = mw[i](h)
	}
	return h
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (s *statusWriter) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// NewRequestID returns a fresh correlation id (Gateway edge must always generate).
func NewRequestID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
