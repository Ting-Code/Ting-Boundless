package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// HTTPEmitter posts audit events to audit-service.
type HTTPEmitter struct {
	baseURL string
	token   string
	client  *http.Client
}

// HTTPEmitterConfig configures the audit HTTP client.
type HTTPEmitterConfig struct {
	BaseURL string
	Token   string
	Timeout time.Duration
}

// NewHTTPEmitter builds an emitter targeting POST /internal/audit/events.
func NewHTTPEmitter(cfg HTTPEmitterConfig) *HTTPEmitter {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &HTTPEmitter{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		token:   cfg.Token,
		client:  &http.Client{Timeout: timeout},
	}
}

// Emit delivers an audit event. Disabled when BaseURL is empty.
func (e *HTTPEmitter) Emit(ctx context.Context, ev Event) error {
	if e == nil || e.baseURL == "" {
		return nil
	}
	body, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("marshal audit event: %w", err)
	}
	if ev.ID == "" || ev.Source == "" || ev.Type == "" {
		return fmt.Errorf("audit event: id, source, and type are required")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.baseURL+"/internal/audit/events", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if e.token != "" {
		req.Header.Set("X-Internal-Token", e.token)
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("audit post: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("audit post: status %d", resp.StatusCode)
	}
	return nil
}

// Enabled reports whether the emitter will deliver events.
func (e *HTTPEmitter) Enabled() bool {
	return e != nil && e.baseURL != ""
}
