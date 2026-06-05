// Package logger provides a JSON structured logger with ECS-style field names.
//
// It is a thin wrapper over log/slog. Output is JSON to stdout (12-Factor:
// logs are an event stream; collection is the platform's job, not the service's).
// Field names follow the Elastic Common Schema (ECS) so that all languages emit
// the same shape; see platform-contracts/schemas/logging.schema.json.
package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

// ECS-style keys.
const (
	keyTime    = "@timestamp"
	keyLevel   = "log.level"
	keyMessage = "message"
	keyService = "service.name"
)

// New returns a JSON slog.Logger tagged with the service name. level is one of
// debug, info, warn, error (defaults to info).
func New(service, level string) *slog.Logger {
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:       parseLevel(level),
		ReplaceAttr: ecsReplace,
	})
	return slog.New(h).With(slog.String(keyService, service))
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// ecsReplace renames the default slog keys to ECS-style names.
func ecsReplace(_ []string, a slog.Attr) slog.Attr {
	switch a.Key {
	case slog.TimeKey:
		a.Key = keyTime
	case slog.LevelKey:
		a.Key = keyLevel
	case slog.MessageKey:
		a.Key = keyMessage
	}
	return a
}

type ctxKey struct{}

// Into stores a logger in the context.
func Into(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

// From retrieves a logger from the context, falling back to slog.Default.
func From(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}
