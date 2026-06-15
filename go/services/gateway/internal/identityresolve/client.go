// Package identityresolve calls auth-service to map IdP subjects to platform user IDs.
package identityresolve

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/pkg/otel"
)

const providerLogto = "logto"

// Client resolves external identity provider subjects via auth-service.
type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

// NewClient builds a resolver. Returns nil when authServiceURL is empty.
func NewClient(authServiceURL, internalToken string) *Client {
	authServiceURL = strings.TrimRight(strings.TrimSpace(authServiceURL), "/")
	if authServiceURL == "" {
		return nil
	}
	return &Client{
		baseURL: authServiceURL,
		token:   internalToken,
		http:    otel.NewHTTPClient(3 * time.Second),
	}
}

// FromEnv loads resolver settings from AUTH_SERVICE_URL and INTERNAL_API_TOKEN.
func FromEnv() *Client {
	return NewClient(httpx.Env("AUTH_SERVICE_URL", "http://127.0.0.1:8084"), httpx.Env("INTERNAL_API_TOKEN", ""))
}

type resolveRequest struct {
	Provider    string `json:"provider"`
	ProviderUID string `json:"provider_uid"`
}

type resolveResponse struct {
	UserID string `json:"user_id"`
}

// ResolveLogto maps a Logto OIDC subject to a platform user_id.
func (c *Client) ResolveLogto(ctx context.Context, logtoSub, requestID string) (string, error) {
	if c == nil {
		return "", fmt.Errorf("identity resolver not configured")
	}
	return c.resolve(ctx, providerLogto, logtoSub, requestID)
}

func (c *Client) resolve(ctx context.Context, provider, providerUID, requestID string) (string, error) {
	body, err := json.Marshal(resolveRequest{Provider: provider, ProviderUID: providerUID})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/identity/resolve", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("X-Internal-Token", c.token)
	}
	if requestID != "" {
		req.Header.Set(identity.HeaderRequestID, requestID)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("identity resolve: status %d", resp.StatusCode)
	}

	var out resolveResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.UserID == "" {
		return "", fmt.Errorf("identity resolve: empty user_id")
	}
	return out.UserID, nil
}
