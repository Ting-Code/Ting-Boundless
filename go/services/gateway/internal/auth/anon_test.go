package auth

import "testing"

func TestParseAnonPrefixes_DefaultWhenEmpty(t *testing.T) {
	got := ParseAnonPrefixes("")
	def := DefaultAnonPrefixes()
	if len(got.ExactPaths()) != len(def.ExactPaths()) {
		t.Fatalf("exact len=%d want %d", len(got.ExactPaths()), len(def.ExactPaths()))
	}
	if len(got.PrefixPaths()) != len(def.PrefixPaths()) {
		t.Fatalf("prefix len=%d want %d", len(got.PrefixPaths()), len(def.PrefixPaths()))
	}
}

func TestParseAnonPrefixes_ExactAndPrefixSyntax(t *testing.T) {
	p := ParseAnonPrefixes("=/sign-in,/admin,/v1/auth/")
	if !p.Allows("/sign-in") {
		t.Fatal("/sign-in should match exact")
	}
	if p.Allows("/sign-in/dev") {
		t.Fatal("/sign-in/dev must not match exact /sign-in")
	}
	if !p.Allows("/admin/items") {
		t.Fatal("/admin/items should match prefix /admin")
	}
	if !p.Allows("/v1/auth/miniprogram/login") {
		t.Fatal("/v1/auth/* should match")
	}
}

func TestAnonPrefixes_DefaultTightening(t *testing.T) {
	p := DefaultAnonPrefixes()
	cases := []struct {
		path string
		want bool
	}{
		{"/healthz", true},
		{"/sign-in", true},
		{"/callback", true},
		{"/sign-out", true},
		{"/sign-in/dev", false},
		{"/sign-in/extra", false},
		{"/v1/business/ping", true},
		{"/admin/items", true},
		{"/admin", true},
		{"/v1/auth/miniprogram/login", true},
		{"/internal/webhooks/logto", true},
		{"/internal/webhooks/other", false},
		{"/v1/business/me", false},
		{"/v1/business/items", false},
		{"/v1/users/me", false},
		{"/", true},
	}
	for _, tc := range cases {
		if got := p.Allows(tc.path); got != tc.want {
			t.Fatalf("Allows(%q)=%v want %v", tc.path, got, tc.want)
		}
	}
}

func TestResolveAnonPrefixes_DevLoginAddsSignInDev(t *testing.T) {
	t.Setenv("GATEWAY_ANON_PREFIXES", "")
	t.Setenv("GATEWAY_BFF_DEV_LOGIN", "false")
	if ResolveAnonPrefixes().Allows("/sign-in/dev") {
		t.Fatal("dev path should be blocked when GATEWAY_BFF_DEV_LOGIN=false")
	}

	t.Setenv("GATEWAY_BFF_DEV_LOGIN", "true")
	if !ResolveAnonPrefixes().Allows("/sign-in/dev") {
		t.Fatal("dev path should be allowed when GATEWAY_BFF_DEV_LOGIN=true")
	}
}

func TestResolveAnonPrefixes_RespectsCustomEnv(t *testing.T) {
	t.Setenv("GATEWAY_ANON_PREFIXES", "=/only-this")
	t.Setenv("GATEWAY_BFF_DEV_LOGIN", "false")
	p := ResolveAnonPrefixes()
	if !p.Allows("/only-this") {
		t.Fatal("custom exact rule should match")
	}
	if p.Allows("/healthz") {
		t.Fatal("default rules should not apply when env is set")
	}
}

func TestAnonPrefixes_CustomEnvOverridesDefaults(t *testing.T) {
	p := ParseAnonPrefixes("=/custom")
	if !p.Allows("/custom") {
		t.Fatal("expected custom exact path")
	}
	if p.Allows("/sign-in") {
		t.Fatal("defaults should not apply")
	}
}
