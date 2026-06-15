package storage

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/ting-boundless/boundless/pkg/config"
	"github.com/ting-boundless/boundless/pkg/httpx"
)

// Client uploads objects to an S3-compatible bucket (MinIO, OSS, etc.).
type Client struct {
	host      string
	secure    bool
	accessKey string
	secretKey string
	bucket    string
	region    string
	http      *http.Client
}

// Config holds S3 connection parameters (12-Factor via env).
type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	Region    string
	UseSSL    bool
}

// ConfigFromEnv loads S3 settings from the environment.
func ConfigFromEnv() Config {
	useSSL := httpx.Env("S3_USE_SSL", "") == "true"
	endpoint := EndpointFromEnv()
	if strings.HasPrefix(strings.ToLower(endpoint), "https://") {
		useSSL = true
	}
	region := httpx.Env("S3_REGION", "")
	if region == "" {
		region = "us-east-1"
	}
	return Config{
		Endpoint:  endpoint,
		AccessKey: httpx.Env("S3_ACCESS_KEY", ""),
		SecretKey: httpx.Env("S3_SECRET_KEY", ""),
		Bucket:    httpx.Env("S3_BUCKET", ""),
		Region:    region,
		UseSSL:    useSSL,
	}
}

// NewClient opens an S3-compatible client. Returns nil,nil when storage is not configured.
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	if config.IsPlaceholder(cfg.Endpoint) || cfg.Bucket == "" ||
		config.IsPlaceholder(cfg.AccessKey) || config.IsPlaceholder(cfg.SecretKey) ||
		cfg.AccessKey == "" || cfg.SecretKey == "" {
		return nil, nil
	}
	if err := Ping(ctx, cfg.Endpoint); err != nil {
		return nil, err
	}

	host, secure, err := parseEndpoint(cfg.Endpoint, cfg.UseSSL)
	if err != nil {
		return nil, err
	}

	return &Client{
		host:      host,
		secure:    secure,
		accessKey: cfg.AccessKey,
		secretKey: cfg.SecretKey,
		bucket:    cfg.Bucket,
		region:    cfg.Region,
		http:      &http.Client{Timeout: 60 * time.Second},
	}, nil
}

// Enabled reports whether uploads can be performed.
func (c *Client) Enabled() bool {
	return c != nil && c.bucket != "" && c.accessKey != ""
}

// Bucket returns the configured bucket name.
func (c *Client) Bucket() string {
	if c == nil {
		return ""
	}
	return c.bucket
}

// PutObject uploads content using SigV4 (path-style; MinIO-compatible).
func (c *Client) PutObject(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	if !c.Enabled() {
		return fmt.Errorf("s3 client not configured")
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	u := c.objectURL(key)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, r)
	if err != nil {
		return err
	}
	if size >= 0 {
		req.ContentLength = size
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-amz-content-sha256", "UNSIGNED-PAYLOAD")

	if err := signS3Request(req, c.accessKey, c.secretKey, c.region, "UNSIGNED-PAYLOAD"); err != nil {
		return err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("s3 put: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("s3 put: status %d", resp.StatusCode)
	}
	return nil
}

const emptyPayloadHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

// GetObject fetches an object body (caller must close resp.Body).
func (c *Client) GetObject(ctx context.Context, key string) (*http.Response, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("s3 client not configured")
	}

	u := c.objectURL(key)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	if err := signS3Request(req, c.accessKey, c.secretKey, c.region, emptyPayloadHash); err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("s3 get: %w", err)
	}
	if resp.StatusCode >= 300 {
		defer resp.Body.Close()
		return nil, fmt.Errorf("s3 get: status %d", resp.StatusCode)
	}
	return resp, nil
}

// DeleteObject removes an object from the bucket.
func (c *Client) DeleteObject(ctx context.Context, key string) error {
	if !c.Enabled() {
		return fmt.Errorf("s3 client not configured")
	}

	u := c.objectURL(key)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return err
	}

	if err := signS3Request(req, c.accessKey, c.secretKey, c.region, emptyPayloadHash); err != nil {
		return err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("s3 delete: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("s3 delete: status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) objectURL(key string) string {
	scheme := "http"
	if c.secure {
		scheme = "https"
	}
	path := "/" + c.bucket + "/" + strings.TrimPrefix(key, "/")
	return scheme + "://" + c.host + path
}

func parseEndpoint(raw string, useSSL bool) (host string, secure bool, err error) {
	if !strings.Contains(raw, "://") {
		raw = "http://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", false, fmt.Errorf("s3 endpoint: %w", err)
	}
	host = u.Host
	if host == "" {
		return "", false, fmt.Errorf("s3 endpoint: missing host")
	}
	secure = useSSL || u.Scheme == "https"
	return host, secure, nil
}

func signS3Request(req *http.Request, accessKey, secretKey, region, payloadHash string) error {
	now := time.Now().UTC()
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")
	req.Header.Set("X-Amz-Date", amzDate)
	if req.Header.Get("Host") == "" {
		req.Header.Set("Host", req.URL.Host)
	}

	canonicalURI := req.URL.EscapedPath()
	if canonicalURI == "" {
		canonicalURI = "/"
	}

	signedHeaders := signedHeaderNames(req)
	sort.Strings(signedHeaders)

	var canonicalHeaders strings.Builder
	for _, h := range signedHeaders {
		canonicalHeaders.WriteString(h)
		canonicalHeaders.WriteByte(':')
		canonicalHeaders.WriteString(strings.TrimSpace(req.Header.Get(http.CanonicalHeaderKey(h))))
		canonicalHeaders.WriteByte('\n')
	}
	signedHeadersStr := strings.Join(signedHeaders, ";")

	canonicalRequest := strings.Join([]string{
		req.Method,
		canonicalURI,
		req.URL.RawQuery,
		canonicalHeaders.String(),
		signedHeadersStr,
		payloadHash,
	}, "\n")

	crHash := sha256Hex([]byte(canonicalRequest))
	credentialScope := dateStamp + "/" + region + "/s3/aws4_request"
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		credentialScope,
		crHash,
	}, "\n")

	signingKey := deriveSigningKey(secretKey, dateStamp, region, "s3")
	signature := hex.EncodeToString(hmacSHA256(signingKey, []byte(stringToSign)))

	auth := fmt.Sprintf(
		"AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		accessKey, credentialScope, signedHeadersStr, signature,
	)
	req.Header.Set("Authorization", auth)
	return nil
}

func signedHeaderNames(req *http.Request) []string {
	var names []string
	for k := range req.Header {
		lower := strings.ToLower(k)
		switch lower {
		case "host", "content-type":
			names = append(names, lower)
		default:
			if strings.HasPrefix(lower, "x-amz-") {
				names = append(names, lower)
			}
		}
	}
	if len(names) == 0 {
		names = []string{"host"}
	}
	hasHost := false
	for _, n := range names {
		if n == "host" {
			hasHost = true
			break
		}
	}
	if !hasHost {
		names = append(names, "host")
	}
	return names
}

func deriveSigningKey(secret, dateStamp, region, service string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secret), []byte(dateStamp))
	kRegion := hmacSHA256(kDate, []byte(region))
	kService := hmacSHA256(kRegion, []byte(service))
	return hmacSHA256(kService, []byte("aws4_request"))
}

func hmacSHA256(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write(data)
	return mac.Sum(nil)
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
