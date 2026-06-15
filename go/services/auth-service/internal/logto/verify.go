package logto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const signatureHeader = "Logto-Signature-Sha-256"

// VerifySignature checks the Logto webhook HMAC-SHA256 hex digest.
func VerifySignature(signingKey string, rawBody []byte, headerValue string) error {
	if signingKey == "" {
		return fmt.Errorf("signing key not configured")
	}
	if headerValue == "" {
		return fmt.Errorf("missing %s header", signatureHeader)
	}
	mac := hmac.New(sha256.New, []byte(signingKey))
	mac.Write(rawBody)
	expected := mac.Sum(nil)
	got, err := hex.DecodeString(headerValue)
	if err != nil {
		return fmt.Errorf("invalid signature encoding")
	}
	if !hmac.Equal(expected, got) {
		return fmt.Errorf("signature mismatch")
	}
	return nil
}
