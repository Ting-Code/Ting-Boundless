package logger

import (
	"context"
	"log/slog"
)

// FanoutHandler forwards each record to every underlying handler.
type FanoutHandler struct {
	handlers []slog.Handler
}

// NewFanout returns a handler that duplicates records to all children.
func NewFanout(handlers ...slog.Handler) *FanoutHandler {
	out := make([]slog.Handler, 0, len(handlers))
	for _, h := range handlers {
		if h != nil {
			out = append(out, h)
		}
	}
	return &FanoutHandler{handlers: out}
}

func (f *FanoutHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range f.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (f *FanoutHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range f.handlers {
		if err := h.Handle(ctx, r.Clone()); err != nil {
			return err
		}
	}
	return nil
}

func (f *FanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := make([]slog.Handler, len(f.handlers))
	for i, h := range f.handlers {
		next[i] = h.WithAttrs(attrs)
	}
	return &FanoutHandler{handlers: next}
}

func (f *FanoutHandler) WithGroup(name string) slog.Handler {
	next := make([]slog.Handler, len(f.handlers))
	for i, h := range f.handlers {
		next[i] = h.WithGroup(name)
	}
	return &FanoutHandler{handlers: next}
}

// WithHandler rebuilds a logger using the same level/options as New but a custom handler.
func WithHandler(service string, level string, handler slog.Handler) *slog.Logger {
	if handler == nil {
		return New(service, level)
	}
	return slog.New(handler).With(slog.String(keyService, service))
}
