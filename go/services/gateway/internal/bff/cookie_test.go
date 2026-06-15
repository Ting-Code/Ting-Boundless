package bff

import (
	"os"
	"testing"
)

func TestSessionCookieSecure(t *testing.T) {
	t.Setenv("SESSION_COOKIE_SECURE", "")
	t.Setenv("APP_ENV", "")

	if sessionCookieSecure() {
		t.Fatal("expected false by default")
	}

	t.Setenv("APP_ENV", "production")
	if !sessionCookieSecure() {
		t.Fatal("expected true when APP_ENV=production")
	}

	t.Setenv("SESSION_COOKIE_SECURE", "false")
	if sessionCookieSecure() {
		t.Fatal("explicit false should override production")
	}

	t.Setenv("APP_ENV", "development")
	t.Setenv("SESSION_COOKIE_SECURE", "true")
	if !sessionCookieSecure() {
		t.Fatal("explicit true should enable Secure")
	}
}

func TestSessionCookieSecureUnset(t *testing.T) {
	os.Unsetenv("SESSION_COOKIE_SECURE")
	os.Unsetenv("APP_ENV")
	if sessionCookieSecure() {
		t.Fatal("unset env should default to false")
	}
}
