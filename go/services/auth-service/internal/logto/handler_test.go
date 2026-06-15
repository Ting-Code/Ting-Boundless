package logto

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/ting-boundless/boundless/pkg/audit"
)

type fakeAudit struct {
	mu     sync.Mutex
	events []audit.Event
}

func (f *fakeAudit) Emit(_ context.Context, e audit.Event) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events = append(f.events, e)
	return nil
}

func TestHandler_AcceptsSignedWebhook(t *testing.T) {
	key := "hook-secret"
	body := []byte(`{"hookId":"h1","event":"PostSignIn","createdAt":"2024-01-01T00:00:00.000Z","userId":"u1"}`)
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write(body)
	sig := hex.EncodeToString(mac.Sum(nil))

	aud := &fakeAudit{}
	h := NewHandler(Config{
		SigningKey: key,
		Audit:      aud,
		Log:        slog.New(slog.NewTextHandler(os.Stderr, nil)),
	})

	req := httptest.NewRequest(http.MethodPost, "/internal/webhooks/logto", bytes.NewReader(body))
	req.Header.Set(signatureHeader, sig)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
	if len(aud.events) != 1 || aud.events[0].Type != "user.login.success" {
		t.Fatalf("events=%+v", aud.events)
	}
}

func TestHandler_InvalidSignature(t *testing.T) {
	h := NewHandler(Config{
		SigningKey: "key",
		Log:        slog.New(slog.NewTextHandler(os.Stderr, nil)),
	})
	body := []byte(`{"hookId":"h","event":"PostSignIn","createdAt":"2024-01-01T00:00:00.000Z"}`)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set(signatureHeader, "00")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rr.Code)
	}
}

func TestHandler_SkipVerifyDev(t *testing.T) {
	aud := &fakeAudit{}
	h := NewHandler(Config{
		SkipVerify: true,
		Audit:      aud,
		Log:        slog.New(slog.NewTextHandler(os.Stderr, nil)),
	})
	body := []byte(`{"hookId":"h","event":"Identifier.Lockout","createdAt":"2024-01-01T00:00:00.000Z"}`)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("status %d", rr.Code)
	}
	if len(aud.events) != 1 || aud.events[0].Type != "user.login.failed" {
		t.Fatalf("events=%+v", aud.events)
	}
}

func TestHandler_DuplicateWithoutDeliveriesStore(t *testing.T) {
	aud := &fakeAudit{}
	h := NewHandler(Config{SkipVerify: true, Audit: aud, Log: slog.New(slog.NewTextHandler(os.Stderr, nil))})
	body := []byte(`{"hookId":"h","event":"PostSignIn","createdAt":"2024-01-01T00:00:00.000Z"}`)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusAccepted {
			t.Fatalf("iter %d status %d", i, rr.Code)
		}
	}
	if len(aud.events) != 2 {
		t.Fatalf("without deliveries store both requests emit audit: got %d", len(aud.events))
	}
}
