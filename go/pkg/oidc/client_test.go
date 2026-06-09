package oidc

import (
	"net/url"
	"testing"
)

func TestAuthorizeURL(t *testing.T) {
	c := NewClient(ClientConfig{
		ClientID:     "app",
		ClientSecret: "secret",
		RedirectURI:  "http://localhost:8080/callback",
		Scopes:       "openid profile",
		AuthURL:      "http://idp/oidc/auth",
		TokenURL:     "http://idp/oidc/token",
	})

	u, err := c.AuthorizeURL("state-1", "nonce-1")
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(u)
	if err != nil {
		t.Fatal(err)
	}
	q := parsed.Query()
	if q.Get("client_id") != "app" {
		t.Fatalf("client_id=%q", q.Get("client_id"))
	}
	if q.Get("response_type") != "code" {
		t.Fatalf("response_type=%q", q.Get("response_type"))
	}
	if q.Get("state") != "state-1" {
		t.Fatalf("state=%q", q.Get("state"))
	}
}

func TestConfigFromEnv_RedirectDefault(t *testing.T) {
	t.Setenv("GATEWAY_PUBLIC_URL", "http://127.0.0.1:8080")
	t.Setenv("OIDC_ISSUER", "http://idp/oidc")
	cfg := ConfigFromEnv()
	if cfg.RedirectURI != "http://127.0.0.1:8080/callback" {
		t.Fatalf("redirect=%q", cfg.RedirectURI)
	}
	if cfg.AuthURL != "http://idp/oidc/auth" {
		t.Fatalf("auth=%q", cfg.AuthURL)
	}
}
