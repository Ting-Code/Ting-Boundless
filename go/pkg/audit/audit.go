// Package audit defines the CloudEvents-style audit event and the emitter
// interface used across services.
//
// Delivery rules (see docs/ARCHITECTURE.md "Audit Sources And Delivery"):
//   - Domain events tied to a business write MUST use the Transactional Outbox
//     pattern: persist the event in the same DB transaction, then dispatch async.
//   - Gateway entry events (no business write) may be emitted async without outbox.
//   - Identity events come from auth-service (Logto webhook / code2session).
package audit

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"
)

// newID returns a random 128-bit hex id (stdlib only; swap for UUID later).
func newID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// SpecVersion is the CloudEvents spec version used for the envelope.
const SpecVersion = "1.0"

// Source identifies the origin of an event (CloudEvents "source").
type Source string

const (
	SourceIdP     Source = "auth-service"
	SourceGateway Source = "gateway"
	// Business services use their own name as the source.
)

// Event is a CloudEvents-style audit event. Keep fields in sync with
// platform-contracts/schemas/audit-event.schema.json.
type Event struct {
	ID          string         `json:"id"`
	Source      string         `json:"source"`
	Type        string         `json:"type"` // e.g. user.login.success, resource.delete
	Subject     string         `json:"subject,omitempty"`
	Time        time.Time      `json:"time"`
	TenantID    string         `json:"tenant_id,omitempty"`
	ActorUserID string         `json:"actor_user_id,omitempty"`
	Data        map[string]any `json:"data,omitempty"`
}

// NewEvent builds an event with an id and timestamp filled in.
func NewEvent(source Source, eventType string) Event {
	return Event{
		ID:     newID(),
		Source: string(source),
		Type:   eventType,
		Time:   time.Now().UTC(),
	}
}

// Emitter delivers audit events.
//
// Business-service implementations must write to the outbox inside the business
// transaction. The Gateway may use a fire-and-forget async implementation.
type Emitter interface {
	Emit(ctx context.Context, e Event) error
}
