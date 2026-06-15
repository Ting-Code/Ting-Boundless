package outbox

import (
	"testing"
	"time"
)

func TestRowToEvent(t *testing.T) {
	payload := []byte(`{"item_id":"i1","tenant_id":"t1","actor_user_id":"u1","title":"hello"}`)
	ev, err := rowToEvent("550e8400-e29b-41d4-a716-446655440000", "business.item.created", payload, time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if ev.ID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("id=%q", ev.ID)
	}
	if ev.Source != "business-service" {
		t.Fatalf("source=%q", ev.Source)
	}
	if ev.Type != "business.item.created" {
		t.Fatalf("type=%q", ev.Type)
	}
	if ev.ActorUserID != "u1" || ev.TenantID != "t1" || ev.Subject != "item:i1" {
		t.Fatalf("ev=%+v", ev)
	}
	if ev.Time != time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC) {
		t.Fatalf("time=%v", ev.Time)
	}
}
