package auth

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/ting-boundless/boundless/pkg/audit"
	"github.com/ting-boundless/boundless/pkg/auth"
)

type syncAudit struct {
	mu     sync.Mutex
	events []audit.Event
}

func (s *syncAudit) Emit(_ context.Context, e audit.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, e)
	return nil
}

func TestAuthenticate_EmitsAccessDeniedAudit(t *testing.T) {
	cfg := auth.Config{DevSecret: "s"}
	v, err := auth.NewVerifier(t.Context(), cfg, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	if err != nil {
		t.Fatal(err)
	}
	aud := &syncAudit{}

	req := httptest.NewRequest(http.MethodGet, "/v1/users/me", nil)
	rr := httptest.NewRecorder()
	Authenticate(v, nil, DefaultAnonPrefixes(), nil, aud, nil, SensitivePrefixes{})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rr.Code)
	}
	aud.mu.Lock()
	defer aud.mu.Unlock()
	if len(aud.events) != 1 || aud.events[0].Type != "api.access.denied" {
		t.Fatalf("events=%+v", aud.events)
	}
}

func TestAuthenticate_EmitsInvalidTokenAudit(t *testing.T) {
	cfg := auth.Config{DevSecret: "s"}
	v, err := auth.NewVerifier(t.Context(), cfg, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	if err != nil {
		t.Fatal(err)
	}
	aud := &syncAudit{}

	req := httptest.NewRequest(http.MethodGet, "/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer bad")
	rr := httptest.NewRecorder()
	Authenticate(v, nil, DefaultAnonPrefixes(), nil, aud, nil, SensitivePrefixes{})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rr.Code)
	}
	aud.mu.Lock()
	defer aud.mu.Unlock()
	if len(aud.events) != 1 || aud.events[0].Type != "api.token.invalid" {
		t.Fatalf("events=%+v", aud.events)
	}
}
