package logto

import (
	"testing"
	"time"
)

func TestToAuditEvent_PostSignIn(t *testing.T) {
	raw := []byte(`{
		"hookId":"hook-1",
		"event":"PostSignIn",
		"createdAt":"2024-06-01T12:00:00.000Z",
		"userId":"logto-user-9",
		"userIp":"203.0.113.1"
	}`)
	p, err := ParsePayload(raw)
	if err != nil {
		t.Fatal(err)
	}
	ev, ok := ToAuditEvent(p, "platform-42")
	if !ok {
		t.Fatal("expected mapped event")
	}
	if ev.Type != "user.login.success" {
		t.Fatalf("type=%q", ev.Type)
	}
	if ev.ActorUserID != "platform-42" {
		t.Fatalf("actor=%q", ev.ActorUserID)
	}
	if ev.Data["client_ip"] != "203.0.113.1" {
		t.Fatalf("ip=%v", ev.Data["client_ip"])
	}
	if ev.Time != time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC) {
		t.Fatalf("time=%v", ev.Time)
	}
}

func TestToAuditEvent_UnknownSkipped(t *testing.T) {
	raw := []byte(`{"hookId":"h","event":"Role.Created","createdAt":"2024-01-01T00:00:00.000Z"}`)
	p, err := ParsePayload(raw)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := ToAuditEvent(p, ""); ok {
		t.Fatal("expected skip")
	}
}

func TestDeliveryKeyStable(t *testing.T) {
	p := Payload{HookID: "a", CreatedAt: "t", Event: "PostSignIn"}
	if DeliveryKey(p) != "a:t:PostSignIn" {
		t.Fatal(DeliveryKey(p))
	}
}
