package identity_test

import (
	"testing"

	"github.com/ting-boundless/boundless/pkg/identity"
)

func TestHasRole(t *testing.T) {
	id := identity.Identity{Roles: []string{"admin", "user"}}
	if !identity.HasRole(id, "admin") {
		t.Fatal("expected admin")
	}
	if identity.HasRole(id, "owner") {
		t.Fatal("unexpected owner")
	}
}
