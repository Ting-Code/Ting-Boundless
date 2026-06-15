package audit

import (
	"context"
	"log/slog"
	"time"
)

// Async wraps an Emitter and delivers events in a background goroutine so the
// request path is not blocked (Gateway entry events).
type Async struct {
	inner Emitter
	log   *slog.Logger
}

// NewAsync builds a non-blocking emitter. Returns nil when inner is nil or disabled.
func NewAsync(inner Emitter, log *slog.Logger) *Async {
	if inner == nil {
		return nil
	}
	if http, ok := inner.(*HTTPEmitter); ok && !http.Enabled() {
		return nil
	}
	return &Async{inner: inner, log: log}
}

// Emit schedules delivery without blocking the caller.
func (a *Async) Emit(ctx context.Context, e Event) error {
	if a == nil || a.inner == nil {
		return nil
	}
	ev := e
	go func() {
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := a.inner.Emit(c, ev); err != nil && a.log != nil {
			a.log.Warn("async audit emit failed",
				slog.String("type", ev.Type),
				slog.Any("error", err),
			)
		}
	}()
	return nil
}

// Enabled reports whether async delivery is configured.
func (a *Async) Enabled() bool {
	return a != nil && a.inner != nil
}
