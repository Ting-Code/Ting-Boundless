package identityapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ting-boundless/boundless/services/auth-service/internal/identityapi"
)

type stubResolver struct {
	userID string
	err    error
}

func (s stubResolver) ResolveProviderUser(context.Context, string, string) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return s.userID, nil
}

func TestResolveHandler_OK(t *testing.T) {
	h := identityapi.NewResolveHandler(stubResolver{userID: "platform-u1"})
	body := strings.NewReader(`{"provider":"logto","provider_uid":"sub-abc"}`)
	req := httptest.NewRequest(http.MethodPost, "/internal/identity/resolve", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var resp struct {
		UserID string `json:"user_id"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.UserID != "platform-u1" {
		t.Fatalf("user_id=%q", resp.UserID)
	}
}

func TestResolveHandler_BadRequest(t *testing.T) {
	h := identityapi.NewResolveHandler(stubResolver{})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/internal/identity/resolve", strings.NewReader(`{}`)))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestResolveHandler_Unconfigured(t *testing.T) {
	var h *identityapi.ResolveHandler
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/internal/identity/resolve", nil))
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d", rr.Code)
	}
}
