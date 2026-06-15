package storage

import (
	"net/http"
	"net/url"
	"testing"
)

func TestSignS3Request_ProducesAuthorization(t *testing.T) {
	req, err := http.NewRequest(http.MethodPut, "http://127.0.0.1:9000/my-bucket/tenant/u1/id/file.txt", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("x-amz-content-sha256", "UNSIGNED-PAYLOAD")

	if err := signS3Request(req, "minioadmin", "minioadmin", "us-east-1", "UNSIGNED-PAYLOAD"); err != nil {
		t.Fatal(err)
	}
	auth := req.Header.Get("Authorization")
	if auth == "" || !stringsHasPrefix(auth, "AWS4-HMAC-SHA256") {
		t.Fatalf("auth=%q", auth)
	}
	if req.Header.Get("X-Amz-Date") == "" {
		t.Fatal("missing X-Amz-Date")
	}
}

func TestParseEndpoint(t *testing.T) {
	host, secure, err := parseEndpoint("http://localhost:9000", false)
	if err != nil || host != "localhost:9000" || secure {
		t.Fatalf("host=%q secure=%v err=%v", host, secure, err)
	}
	_, secure, err = parseEndpoint("https://oss.example.com", false)
	if err != nil || !secure {
		t.Fatalf("secure=%v err=%v", secure, err)
	}
}

func TestObjectURL(t *testing.T) {
	c := &Client{host: "localhost:9000", bucket: "ting", secure: false}
	u, err := url.Parse(c.objectURL("t1/u1/abc/readme.md"))
	if err != nil {
		t.Fatal(err)
	}
	if u.Path != "/ting/t1/u1/abc/readme.md" {
		t.Fatalf("path=%q", u.Path)
	}
}

func stringsHasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
