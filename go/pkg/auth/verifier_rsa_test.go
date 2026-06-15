package auth

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestVerifier_RS256ViaJWKS(t *testing.T) {
	issuer, err := NewIssuer(IssuerConfig{
		Issuer:          "http://auth.test/oidc",
		Audience:        "ting-api",
		GenerateIfEmpty: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := issuer.JWKSJSON()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	cfg := Config{
		Issuers:  []string{"http://auth.test/oidc"},
		JWKSURLs: []string{srv.URL},
		Audience: "ting-api",
	}
	v, err := NewVerifier(t.Context(), cfg, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	if err != nil {
		t.Fatal(err)
	}

	tok, err := issuer.AccessToken("wx-user", "", []string{"user"}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	id, err := v.Verify(tok)
	if err != nil {
		t.Fatal(err)
	}
	if id.UserID != "wx-user" {
		t.Fatalf("user_id=%q", id.UserID)
	}
}

func TestVerifier_MultiIssuer(t *testing.T) {
	cfg := Config{
		Issuers:   []string{"http://logto/oidc", "http://auth/oidc"},
		DevSecret: "multi-secret",
	}
	v, err := NewVerifier(t.Context(), cfg, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	if err != nil {
		t.Fatal(err)
	}

	for _, iss := range cfg.Issuers {
		c := Config{Issuers: []string{iss}, DevSecret: cfg.DevSecret}
		tok, err := DevToken(c, "u1", "t1", nil, time.Hour)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := v.Verify(tok); err != nil {
			t.Fatalf("issuer %q: %v", iss, err)
		}
	}

	bad := Config{Issuers: []string{"http://evil/oidc"}, DevSecret: cfg.DevSecret}
	tok, err := DevToken(bad, "u1", "t1", nil, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := v.Verify(tok); err == nil {
		t.Fatal("expected issuer mismatch")
	}
}
