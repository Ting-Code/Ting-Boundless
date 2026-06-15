package logto

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ting-boundless/boundless/pkg/audit"
)

// Payload is the common Logto webhook envelope (plus event-specific fields).
type Payload struct {
	HookID    string          `json:"hookId"`
	Event     string          `json:"event"`
	CreatedAt string          `json:"createdAt"`
	UserID    string          `json:"userId"`
	UserIP    string          `json:"userIp"`
	IP        string          `json:"ip"`
	User      json.RawMessage `json:"user"`
	Data      json.RawMessage `json:"data"`
}

// ParsePayload decodes a verified webhook body.
func ParsePayload(raw []byte) (Payload, error) {
	var p Payload
	if err := json.Unmarshal(raw, &p); err != nil {
		return Payload{}, fmt.Errorf("invalid json: %w", err)
	}
	if p.HookID == "" || p.Event == "" || p.CreatedAt == "" {
		return Payload{}, fmt.Errorf("hookId, event, and createdAt are required")
	}
	return p, nil
}

// DeliveryKey uniquely identifies a webhook delivery for idempotency.
func DeliveryKey(p Payload) string {
	return p.HookID + ":" + p.CreatedAt + ":" + p.Event
}

// ToAuditEvent maps a Logto webhook to a CloudEvents-style audit event.
func ToAuditEvent(p Payload, platformUserID string) (audit.Event, bool) {
	eventType, ok := auditTypeForEvent(p.Event)
	if !ok {
		return audit.Event{}, false
	}

	ev := audit.NewEvent(audit.SourceIdP, eventType)
	ev.Subject = logtoSubject(p)
	ev.ActorUserID = platformUserID
	ev.Data = map[string]any{
		"logto_hook_id":    p.HookID,
		"logto_event":      p.Event,
		"logto_created_at": p.CreatedAt,
		"logto_user_id":    logtoUserID(p),
	}
	if ip := clientIP(p); ip != "" {
		ev.Data["client_ip"] = ip
	}
	if len(p.User) > 0 {
		ev.Data["logto_user"] = json.RawMessage(p.User)
	}
	ev.ID = "logto-" + sha256Hex(DeliveryKey(p))
	if t, err := time.Parse(time.RFC3339, p.CreatedAt); err == nil {
		ev.Time = t.UTC()
	}
	return ev, true
}

func auditTypeForEvent(event string) (string, bool) {
	switch event {
	case "PostSignIn":
		return "user.login.success", true
	case "PostRegister":
		return "user.register.success", true
	case "PostResetPassword":
		return "user.password.reset", true
	case "Identifier.Lockout":
		return "user.login.failed", true
	case "User.Created":
		return "user.created", true
	case "User.Deleted":
		return "user.deleted", true
	default:
		return "", false
	}
}

func logtoUserID(p Payload) string {
	if p.UserID != "" {
		return p.UserID
	}
	var user struct {
		ID string `json:"id"`
	}
	if len(p.User) > 0 && json.Unmarshal(p.User, &user) == nil && user.ID != "" {
		return user.ID
	}
	if len(p.Data) > 0 && json.Unmarshal(p.Data, &user) == nil {
		return user.ID
	}
	return ""
}

func logtoSubject(p Payload) string {
	if uid := logtoUserID(p); uid != "" {
		return "user:" + uid
	}
	return "logto:" + p.Event
}

func clientIP(p Payload) string {
	if p.UserIP != "" {
		return p.UserIP
	}
	return p.IP
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}
