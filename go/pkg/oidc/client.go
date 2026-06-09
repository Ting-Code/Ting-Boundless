// Package oidc implements OIDC authorization-code helpers for the Gateway BFF.
package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ting-boundless/boundless/pkg/httpx"
)

// ClientConfig holds OIDC client registration (12-Factor via env).
type ClientConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Scopes       string
	Resource     string // Logto API resource indicator (defaults to OIDC_AUDIENCE)
	AuthURL      string
	TokenURL     string
}

// ConfigFromEnv loads OIDC client settings.
func ConfigFromEnv() ClientConfig {
	issuer := strings.TrimRight(httpx.Env("OIDC_ISSUER", ""), "/")
	resource := httpx.Env("OIDC_RESOURCE", "")
	if resource == "" {
		resource = httpx.Env("OIDC_AUDIENCE", "")
	}
	cfg := ClientConfig{
		Issuer:       issuer,
		ClientID:     httpx.Env("OIDC_CLIENT_ID", ""),
		ClientSecret: httpx.Env("OIDC_CLIENT_SECRET", ""),
		RedirectURI:  httpx.Env("OIDC_REDIRECT_URI", ""),
		Scopes:       httpx.Env("OIDC_SCOPES", "openid profile email"),
		Resource:     resource,
		AuthURL:      httpx.Env("OIDC_AUTHORIZATION_URL", ""),
		TokenURL:     httpx.Env("OIDC_TOKEN_URL", ""),
	}
	if cfg.AuthURL == "" && issuer != "" {
		cfg.AuthURL = issuer + "/auth"
	}
	if cfg.TokenURL == "" && issuer != "" {
		cfg.TokenURL = issuer + "/token"
	}
	if cfg.RedirectURI == "" {
		base := strings.TrimRight(httpx.Env("GATEWAY_PUBLIC_URL", "http://127.0.0.1:8080"), "/")
		cfg.RedirectURI = base + "/callback"
	}
	return cfg
}

// Ready reports whether the OIDC BFF flow can run (Logto or compatible IdP).
func (c ClientConfig) Ready() bool {
	return c.ClientID != "" && c.ClientSecret != "" && c.AuthURL != "" && c.TokenURL != "" && c.RedirectURI != ""
}

// Client performs authorization-code exchange with the IdP.
type Client struct {
	cfg    ClientConfig
	client *http.Client
}

// NewClient builds an OIDC client.
func NewClient(cfg ClientConfig) *Client {
	return &Client{
		cfg: cfg,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// AuthorizeURL builds the browser redirect URL for the authorization code flow.
func (c *Client) AuthorizeURL(state, nonce string) (string, error) {
	if !c.cfg.Ready() {
		return "", fmt.Errorf("oidc client not configured")
	}
	u, err := url.Parse(c.cfg.AuthURL)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("client_id", c.cfg.ClientID)
	q.Set("redirect_uri", c.cfg.RedirectURI)
	q.Set("response_type", "code")
	q.Set("scope", c.cfg.Scopes)
	q.Set("state", state)
	q.Set("nonce", nonce)
	if c.cfg.Resource != "" {
		q.Set("resource", c.cfg.Resource)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// TokenResponse is the OIDC token endpoint response.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// ExchangeCode trades an authorization code for tokens.
func (c *Client) ExchangeCode(ctx context.Context, code string) (TokenResponse, error) {
	if !c.cfg.Ready() {
		return TokenResponse{}, fmt.Errorf("oidc client not configured")
	}
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", c.cfg.RedirectURI)
	form.Set("client_id", c.cfg.ClientID)
	form.Set("client_secret", c.cfg.ClientSecret)
	if c.cfg.Resource != "" {
		form.Set("resource", c.cfg.Resource)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return TokenResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return TokenResponse{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return TokenResponse{}, fmt.Errorf("token endpoint %d: %s", resp.StatusCode, string(body))
	}

	var tr TokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return TokenResponse{}, fmt.Errorf("decode token response: %w", err)
	}
	if tr.AccessToken == "" {
		return TokenResponse{}, fmt.Errorf("token response missing access_token")
	}
	return tr, nil
}
