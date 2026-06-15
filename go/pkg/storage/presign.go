package storage

import (
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

const presignPayloadHash = "UNSIGNED-PAYLOAD"
const maxPresignExpiry = 7 * 24 * time.Hour

// PresignGetURL returns a time-limited signed GET URL for an object key.
func (c *Client) PresignGetURL(key string, expires time.Duration) (string, time.Time, error) {
	if !c.Enabled() {
		return "", time.Time{}, fmt.Errorf("s3 client not configured")
	}
	if expires <= 0 {
		expires = time.Hour
	}
	if expires > maxPresignExpiry {
		expires = maxPresignExpiry
	}
	return c.presignGetURLAt(key, expires, time.Now().UTC())
}

func (c *Client) presignGetURLAt(key string, expires time.Duration, now time.Time) (string, time.Time, error) {
	expiresSec := int(expires.Seconds())
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")
	credentialScope := dateStamp + "/" + c.region + "/s3/aws4_request"
	credential := c.accessKey + "/" + credentialScope

	objectPath := "/" + c.bucket + "/" + strings.TrimPrefix(key, "/")
	canonicalURI := awsCanonicalURI(objectPath)

	queryPairs := [][2]string{
		{"X-Amz-Algorithm", "AWS4-HMAC-SHA256"},
		{"X-Amz-Credential", credential},
		{"X-Amz-Date", amzDate},
		{"X-Amz-Expires", strconv.Itoa(expiresSec)},
		{"X-Amz-SignedHeaders", "host"},
	}
	sort.Slice(queryPairs, func(i, j int) bool {
		return queryPairs[i][0] < queryPairs[j][0]
	})

	canonicalQuery := buildCanonicalQuery(queryPairs)
	canonicalHeaders := "host:" + c.host + "\n"
	signedHeaders := "host"

	canonicalRequest := strings.Join([]string{
		"GET",
		canonicalURI,
		canonicalQuery,
		canonicalHeaders,
		signedHeaders,
		presignPayloadHash,
	}, "\n")

	crHash := sha256Hex([]byte(canonicalRequest))
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		credentialScope,
		crHash,
	}, "\n")

	signingKey := deriveSigningKey(c.secretKey, dateStamp, c.region, "s3")
	signature := hex.EncodeToString(hmacSHA256(signingKey, []byte(stringToSign)))

	rawQuery := canonicalQuery + "&X-Amz-Signature=" + awsURIEncode(signature)
	scheme := "http"
	if c.secure {
		scheme = "https"
	}
	signed := scheme + "://" + c.host + canonicalURI + "?" + rawQuery
	return signed, now.Add(expires), nil
}

func buildCanonicalQuery(pairs [][2]string) string {
	parts := make([]string, len(pairs))
	for i, p := range pairs {
		parts[i] = awsURIEncode(p[0]) + "=" + awsURIEncode(p[1])
	}
	return strings.Join(parts, "&")
}

func awsCanonicalURI(path string) string {
	if path == "" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	segments := strings.Split(path, "/")
	for i, seg := range segments {
		segments[i] = awsURIEncode(seg)
	}
	out := strings.Join(segments, "/")
	if !strings.HasPrefix(out, "/") {
		out = "/" + out
	}
	return out
}

func awsURIEncode(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') ||
			c == '_' || c == '-' || c == '.' || c == '~' {
			b.WriteByte(c)
		} else {
			b.WriteString(fmt.Sprintf("%%%02X", c))
		}
	}
	return b.String()
}
