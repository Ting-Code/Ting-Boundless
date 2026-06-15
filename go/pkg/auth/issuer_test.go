package auth

import (
	"encoding/json"
	"testing"
	"time"
)

func TestIssuer_AccessTokenAndJWKS(t *testing.T) {
	iss, err := NewIssuer(IssuerConfig{
		Issuer:          "http://auth.test/oidc",
		Audience:        "ting-test",
		GenerateIfEmpty: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	tok, err := iss.AccessToken("user-1", "tenant-a", []string{"user"}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if tok == "" {
		t.Fatal("empty token")
	}

	raw, err := iss.JWKSJSON()
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Keys []struct {
			Kty string `json:"kty"`
			Kid string `json:"kid"`
		} `json:"keys"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Keys) != 1 || doc.Keys[0].Kty != "RSA" || doc.Keys[0].Kid == "" {
		t.Fatalf("unexpected jwks: %+v", doc)
	}
}
