package bff

import "testing"

func TestDevRolesFromQuery(t *testing.T) {
	got := devRolesFromQuery("")
	want := []string{"user", "admin"}
	if len(got) != len(want) {
		t.Fatalf("default: got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("default: got %v want %v", got, want)
		}
	}

	got = devRolesFromQuery("admin,operator")
	if len(got) != 2 || got[0] != "admin" || got[1] != "operator" {
		t.Fatalf("custom: got %v", got)
	}

	got = devRolesFromQuery("  ,  ")
	if len(got) != 2 {
		t.Fatalf("whitespace-only should fall back to default, got %v", got)
	}
}
