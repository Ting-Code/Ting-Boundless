package logger

import (
	"context"
	"log/slog"
	"testing"
)

type countHandler struct {
	n *int
}

func (h countHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h countHandler) Handle(context.Context, slog.Record) error {
	*h.n++
	return nil
}

func (h countHandler) WithAttrs([]slog.Attr) slog.Handler { return h }

func (h countHandler) WithGroup(string) slog.Handler { return h }

func TestFanoutHandler_ForwardsToAll(t *testing.T) {
	var a, b int
	log := slog.New(NewFanout(countHandler{&a}, countHandler{&b}))
	log.Info("x")

	if a != 1 || b != 1 {
		t.Fatalf("a=%d b=%d", a, b)
	}
}

func TestFanoutHandler_SkipsNilChildren(t *testing.T) {
	var n int
	log := slog.New(NewFanout(nil, countHandler{&n}))
	log.Info("x")
	if n != 1 {
		t.Fatalf("n=%d", n)
	}
}
