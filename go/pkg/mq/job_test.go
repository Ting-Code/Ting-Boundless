package mq

import (
	"testing"
	"time"
)

func TestParseJob_Valid(t *testing.T) {
	raw := []byte(`{"id":"j1","type":"ping","time":"2026-06-15T12:00:00Z","payload":{"x":1}}`)
	j, err := ParseJob(raw)
	if err != nil {
		t.Fatal(err)
	}
	if j.ID != "j1" || j.Type != "ping" {
		t.Fatalf("job=%+v", j)
	}
	if j.Payload["x"].(float64) != 1 {
		t.Fatalf("payload=%v", j.Payload)
	}
}

func TestParseJob_MissingType(t *testing.T) {
	_, err := ParseJob([]byte(`{"id":"j1"}`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestJob_MarshalRoundTrip(t *testing.T) {
	want := Job{
		ID:   "abc",
		Type: "ping",
		Time: time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),
	}
	raw, err := want.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseJob(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != want.ID || got.Type != want.Type {
		t.Fatalf("got=%+v", got)
	}
}
