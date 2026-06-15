package logto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestVerifySignature(t *testing.T) {
	key := "test-signing-key"
	body := []byte(`{"hookId":"h1","event":"PostSignIn","createdAt":"2024-01-01T00:00:00.000Z"}`)
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write(body)
	sig := hex.EncodeToString(mac.Sum(nil))

	if err := VerifySignature(key, body, sig); err != nil {
		t.Fatal(err)
	}
	if err := VerifySignature(key, body, "bad"); err == nil {
		t.Fatal("expected mismatch")
	}
}
