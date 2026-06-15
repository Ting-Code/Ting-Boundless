package jwks_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ting-boundless/boundless/pkg/auth"
	"github.com/ting-boundless/boundless/services/auth-service/internal/jwks"
)

func TestHandler_ReturnsJWKS(t *testing.T) {
	issuer, err := auth.NewIssuer(auth.IssuerConfig{
		Issuer:          "http://test/oidc",
		GenerateIfEmpty: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	h := jwks.NewHandler(issuer)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/auth/jwks", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if ct := rr.Header().Get("Cache-Control"); ct == "" {
		t.Fatal("missing cache-control")
	}
	var doc struct {
		Keys []map[string]any `json:"keys"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Keys) == 0 {
		t.Fatal("expected at least one jwk")
	}
}

func TestHandler_Unconfigured(t *testing.T) {
	var h *jwks.Handler
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/auth/jwks", nil))
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d", rr.Code)
	}
}
