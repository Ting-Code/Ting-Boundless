// Package wechat calls WeChat mini-program jscode2session.
package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const code2SessionURL = "https://api.weixin.qq.com/sns/jscode2session"

// Session holds open identifiers returned by WeChat.
type Session struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid,omitempty"`
}

// Client exchanges wx.login() codes for session data.
type Client struct {
	appID      string
	appSecret  string
	mockMode   bool
	httpClient *http.Client
}

// Config configures the WeChat API client.
type Config struct {
	AppID     string
	AppSecret string
	MockMode  bool
}

// NewClient builds a WeChat client. MockMode returns deterministic openids from codes.
func NewClient(cfg Config) *Client {
	return &Client{
		appID:      cfg.AppID,
		appSecret:  cfg.AppSecret,
		mockMode:   cfg.MockMode,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Code2Session exchanges a login code for openid (and optional unionid).
func (c *Client) Code2Session(ctx context.Context, code string) (Session, error) {
	if code == "" {
		return Session{}, fmt.Errorf("code required")
	}
	if c.mockMode {
		openid := "mock_" + code
		unionid := ""
		if parts := strings.SplitN(code, "|", 2); len(parts) == 2 && parts[0] != "" && parts[1] != "" {
			openid = "mock_" + parts[0]
			unionid = "union_" + parts[1]
		}
		return Session{OpenID: openid, UnionID: unionid, SessionKey: "mock-session"}, nil
	}
	if c.appID == "" || c.appSecret == "" {
		return Session{}, fmt.Errorf("wechat credentials not configured")
	}

	q := url.Values{
		"appid":      {c.appID},
		"secret":     {c.appSecret},
		"js_code":    {code},
		"grant_type": {"authorization_code"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, code2SessionURL+"?"+q.Encode(), nil)
	if err != nil {
		return Session{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Session{}, err
	}
	defer resp.Body.Close()

	var body struct {
		OpenID     string `json:"openid"`
		SessionKey string `json:"session_key"`
		UnionID    string `json:"unionid"`
		ErrCode    int    `json:"errcode"`
		ErrMsg     string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return Session{}, fmt.Errorf("wechat response: %w", err)
	}
	if body.ErrCode != 0 {
		return Session{}, fmt.Errorf("wechat errcode %s: %s", strconv.Itoa(body.ErrCode), body.ErrMsg)
	}
	if body.OpenID == "" {
		return Session{}, fmt.Errorf("wechat: empty openid")
	}
	return Session{
		OpenID:     body.OpenID,
		SessionKey: body.SessionKey,
		UnionID:    body.UnionID,
	}, nil
}
