package storage

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/ting-boundless/boundless/pkg/config"
	"github.com/ting-boundless/boundless/pkg/httpx"
)

// EndpointFromEnv reads S3_ENDPOINT.
func EndpointFromEnv() string {
	return httpx.Env("S3_ENDPOINT", "http://localhost:9000")
}

// Ping checks that the S3-compatible endpoint accepts TCP connections.
func Ping(ctx context.Context, endpoint string) error {
	if config.IsPlaceholder(endpoint) {
		return fmt.Errorf("s3 endpoint not configured (%q)", endpoint)
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("s3 endpoint parse: %w", err)
	}
	host := u.Host
	if host == "" {
		host = u.Path
	}
	d := net.Dialer{Timeout: 2 * time.Second}
	conn, err := d.DialContext(ctx, "tcp", host)
	if err != nil {
		return fmt.Errorf("s3 endpoint unreachable: %w", err)
	}
	_ = conn.Close()
	return nil
}
