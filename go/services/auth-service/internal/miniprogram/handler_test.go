package miniprogram_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ting-boundless/boundless/pkg/audit"
	"github.com/ting-boundless/boundless/pkg/auth"
	"github.com/ting-boundless/boundless/services/auth-service/internal/miniprogram"
	"github.com/ting-boundless/boundless/services/auth-service/internal/wechat"
)

type stubIdentity struct {
	userID string
}

func (s stubIdentity) ResolveWeChatUser(context.Context, string, string) (string, error) {
	return s.userID, nil
}

type syncAudit struct {
	mu     sync.Mutex
	events []audit.Event
}

func (a *syncAudit) Emit(_ context.Context, e audit.Event) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.events = append(a.events, e)
	return nil
}

func TestLogin_MockWeChat_MintsJWT(t *testing.T) {
	issuer, err := auth.NewIssuer(auth.IssuerConfig{
		Issuer:          "http://test/oidc",
		Audience:        "ting-test",
		GenerateIfEmpty: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	aud := &syncAudit{}
	h := miniprogram.NewHandler(miniprogram.Config{
		WeChat:   wechat.NewClient(wechat.Config{MockMode: true}),
		Identity: stubIdentity{userID: "user-wx-1"},
		Issuer:   issuer,
		Audit:    aud,
		TTL:      time.Hour,
		Log:      slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	body := strings.NewReader(`{"code":"abc"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/miniprogram/login", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-Id", "rid-1")

	rr := httptest.NewRecorder()
	h.Login(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}

	var resp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int64  `json:"expires_in"`
		UserID      string `json:"user_id"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.UserID != "user-wx-1" || resp.AccessToken == "" || resp.TokenType != "Bearer" {
		t.Fatalf("resp=%+v", resp)
	}

	aud.mu.Lock()
	defer aud.mu.Unlock()
	if len(aud.events) != 1 || aud.events[0].Type != "user.login.success" {
		t.Fatalf("audit=%+v", aud.events)
	}
}

func TestLogin_RequiresCode(t *testing.T) {
	issuer, err := auth.NewIssuer(auth.IssuerConfig{GenerateIfEmpty: true})
	if err != nil {
		t.Fatal(err)
	}
	h := miniprogram.NewHandler(miniprogram.Config{
		WeChat:   wechat.NewClient(wechat.Config{MockMode: true}),
		Identity: stubIdentity{userID: "u1"},
		Issuer:   issuer,
	})

	rr := httptest.NewRecorder()
	h.Login(rr, httptest.NewRequest(http.MethodPost, "/v1/auth/miniprogram/login", strings.NewReader(`{}`)))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rr.Code)
	}
}
