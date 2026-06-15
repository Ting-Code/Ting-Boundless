package wechat

import (
	"context"
	"testing"
)

func TestClient_MockMode(t *testing.T) {
	c := NewClient(Config{MockMode: true})
	sess, err := c.Code2Session(context.Background(), "abc123")
	if err != nil {
		t.Fatal(err)
	}
	if sess.OpenID != "mock_abc123" {
		t.Fatalf("openid=%q", sess.OpenID)
	}
	if sess.UnionID != "" {
		t.Fatalf("unionid=%q", sess.UnionID)
	}
}

func TestClient_MockModeUnionid(t *testing.T) {
	c := NewClient(Config{MockMode: true})
	sess, err := c.Code2Session(context.Background(), "app_a|shared_union")
	if err != nil {
		t.Fatal(err)
	}
	if sess.OpenID != "mock_app_a" {
		t.Fatalf("openid=%q", sess.OpenID)
	}
	if sess.UnionID != "union_shared_union" {
		t.Fatalf("unionid=%q", sess.UnionID)
	}
}

func TestClient_CodeRequired(t *testing.T) {
	c := NewClient(Config{MockMode: true})
	if _, err := c.Code2Session(context.Background(), ""); err == nil {
		t.Fatal("expected error")
	}
}
