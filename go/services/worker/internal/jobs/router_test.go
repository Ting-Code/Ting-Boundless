package jobs_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/ting-boundless/boundless/pkg/mq"
	"github.com/ting-boundless/boundless/services/worker/internal/jobs"
)

type stubEffects struct {
	calls int
}

func (s *stubEffects) Record(context.Context, mq.Job, string) (bool, error) {
	s.calls++
	return true, nil
}

func TestRouter_Ping(t *testing.T) {
	h := jobs.NewRouter(slog.New(slog.NewTextHandler(io.Discard, nil)), nil)
	if err := h(context.Background(), mq.Job{ID: "1", Type: "ping"}); err != nil {
		t.Fatal(err)
	}
}

func TestRouter_BusinessItemCreated(t *testing.T) {
	fx := &stubEffects{}
	h := jobs.NewRouter(nil, fx)
	if err := h(context.Background(), mq.Job{
		ID:     "3",
		Type:   "business.item.created",
		Tenant: "t1",
		Actor:  "u1",
		Payload: map[string]any{
			"item_id": "abc",
			"title":   "hello",
		},
	}); err != nil {
		t.Fatal(err)
	}
	if fx.calls != 1 {
		t.Fatalf("expected 1 record call, got %d", fx.calls)
	}
}

func TestRouter_BusinessItemMissingItemID(t *testing.T) {
	h := jobs.NewRouter(nil, &stubEffects{})
	err := h(context.Background(), mq.Job{ID: "5", Type: "business.item.deleted"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRouter_BusinessItemUpdated(t *testing.T) {
	h := jobs.NewRouter(nil, &stubEffects{})
	if err := h(context.Background(), mq.Job{
		ID: "4", Type: "business.item.updated",
		Payload: map[string]any{"item_id": "abc"},
	}); err != nil {
		t.Fatal(err)
	}
}

func TestRouter_UnknownTypeAcked(t *testing.T) {
	h := jobs.NewRouter(nil, nil)
	if err := h(context.Background(), mq.Job{ID: "2", Type: "future.job"}); err != nil {
		t.Fatal(err)
	}
}
