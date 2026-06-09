package auth

import "testing"

func TestParseAnonPrefixes_DefaultWhenEmpty(t *testing.T) {
	got := ParseAnonPrefixes("")
	if len(got) != len(DefaultAnonPrefixes()) {
		t.Fatalf("len=%d want %d", len(got), len(DefaultAnonPrefixes()))
	}
}

func TestAnonPrefixes_Allows(t *testing.T) {
	p := DefaultAnonPrefixes()
	cases := []struct {
		path string
		want bool
	}{
		{"/sign-in", true},
		{"/sign-in/dev", true},
		{"/v1/business/ping", true},
		{"/admin/items", true},
		{"/v1/business/me", false},
		{"/v1/business/items", false},
		{"/v1/auth/miniprogram/login", true},
	}
	for _, tc := range cases {
		if got := p.Allows(tc.path); got != tc.want {
			t.Fatalf("Allows(%q)=%v want %v", tc.path, got, tc.want)
		}
	}
}
