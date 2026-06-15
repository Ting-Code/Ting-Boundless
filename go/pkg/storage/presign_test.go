package storage

import (
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestPresignGetURL_ContainsSignature(t *testing.T) {
	c := &Client{
		host:      "127.0.0.1:9000",
		secure:    false,
		accessKey: "minioadmin",
		secretKey: "minioadmin",
		bucket:    "ting",
		region:    "us-east-1",
	}
	fixed := time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC)

	raw, expiresAt, err := c.presignGetURLAt("tenant/u1/abc/readme.md", time.Hour, fixed)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(raw, "X-Amz-Signature=") {
		t.Fatalf("url=%q", raw)
	}
	if !strings.Contains(raw, "/ting/tenant/u1/abc/readme.md") {
		t.Fatalf("path missing: %q", raw)
	}
	if expiresAt != fixed.Add(time.Hour) {
		t.Fatalf("expiresAt=%v", expiresAt)
	}

	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if u.Query().Get("X-Amz-Algorithm") != "AWS4-HMAC-SHA256" {
		t.Fatal("missing algorithm")
	}
}

func TestPresignGetURL_Deterministic(t *testing.T) {
	c := &Client{
		host: "localhost:9000", accessKey: "ak", secretKey: "sk",
		bucket: "b", region: "us-east-1",
	}
	fixed := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	u1, _, _ := c.presignGetURLAt("k", time.Minute, fixed)
	u2, _, _ := c.presignGetURLAt("k", time.Minute, fixed)
	if u1 != u2 {
		t.Fatalf("non-deterministic: %q vs %q", u1, u2)
	}
}

func TestAwsURIEncode(t *testing.T) {
	if awsURIEncode("a/b") != "a%2Fb" {
		t.Fatal(awsURIEncode("a/b"))
	}
	if awsURIEncode("plain") != "plain" {
		t.Fatal()
	}
}
