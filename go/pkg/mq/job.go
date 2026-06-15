package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Job is the canonical async job envelope (JSON on the wire).
type Job struct {
	ID      string         `json:"id"`
	Type    string         `json:"type"`
	Time    time.Time      `json:"time"`
	Actor   string         `json:"actor_user_id,omitempty"`
	Tenant  string         `json:"tenant_id,omitempty"`
	Payload map[string]any `json:"payload,omitempty"`
}

// ParseJob decodes a RabbitMQ message body.
func ParseJob(raw []byte) (Job, error) {
	var j Job
	if err := json.Unmarshal(raw, &j); err != nil {
		return Job{}, fmt.Errorf("job json: %w", err)
	}
	if j.Type == "" {
		return Job{}, fmt.Errorf("job type is required")
	}
	if j.ID == "" {
		return Job{}, fmt.Errorf("job id is required")
	}
	if j.Time.IsZero() {
		j.Time = time.Now().UTC()
	}
	return j, nil
}

// Marshal serializes a job for publishing.
func (j Job) Marshal() ([]byte, error) {
	if j.Type == "" || j.ID == "" {
		return nil, fmt.Errorf("job id and type are required")
	}
	if j.Time.IsZero() {
		j.Time = time.Now().UTC()
	}
	return json.Marshal(j)
}

// Handler processes a decoded job. Return an error to trigger retry/DLQ.
type Handler func(ctx context.Context, job Job) error
