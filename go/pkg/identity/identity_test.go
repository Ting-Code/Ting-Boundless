package identity_test

import (
	"net/http"
	"testing"

	"github.com/ting-boundless/boundless/pkg/identity"
)

func TestStripUntrustedAndInjectRoundTrip(t *testing.T) {
	h := make(http.Header)
	h.Set(identity.HeaderUserID, "forged")
	h.Set(identity.HeaderTenantID, "t-forged")
	h.Set(identity.HeaderRoles, "admin")
	h.Set(identity.HeaderRequestID, "client-rid")

	identity.StripUntrusted(h)

	id := identity.Identity{
		RequestID: "trusted-rid",
		UserID:    "u1",
		TenantID:  "t1",
		Roles:     []string{"admin", "user"},
		Scopes:    []string{"read"},
		Subject:   "sub-1",
	}
	id.Inject(h)

	got := identity.FromHeaders(h)
	if got.UserID != "u1" || got.TenantID != "t1" || got.RequestID != "trusted-rid" {
		t.Fatalf("got=%+v", got)
	}
	if len(got.Roles) != 2 || got.Roles[0] != "admin" {
		t.Fatalf("roles=%v", got.Roles)
	}
	if got.Subject != "sub-1" {
		t.Fatalf("subject=%q", got.Subject)
	}
}

func TestAuthenticated(t *testing.T) {
	if !identity.Identity{UserID: "u1"}.Authenticated() {
		t.Fatal("expected authenticated")
	}
	if identity.Identity{}.Authenticated() {
		t.Fatal("expected anonymous")
	}
}

func TestHasRole(t *testing.T) {
	id := identity.Identity{Roles: []string{"user", "admin"}}
	if !identity.HasRole(id, "admin") {
		t.Fatal("expected admin role")
	}
}
