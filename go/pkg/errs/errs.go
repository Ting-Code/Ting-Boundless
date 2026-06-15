// Package errs is the unified error model shared by all services.
//
// The JSON shape matches platform-contracts/schemas/error-response.schema.json
// so that every client sees the same error envelope regardless of language.
package errs

import (
	"encoding/json"
	"net/http"
)

// Error is a unified application error.
type Error struct {
	// Code is a stable, machine-readable string, e.g. "user.not_found".
	Code string `json:"code"`
	// Message is a human-readable description (safe to show to clients).
	Message string `json:"message"`
	// Status is the HTTP status code (not serialized in the body).
	Status int `json:"-"`
}

func (e *Error) Error() string { return e.Code + ": " + e.Message }

// New constructs an Error.
func New(status int, code, message string) *Error {
	return &Error{Code: code, Message: message, Status: status}
}

// Common constructors.
func BadRequest(code, msg string) *Error   { return New(http.StatusBadRequest, code, msg) }
func Unauthorized(code, msg string) *Error { return New(http.StatusUnauthorized, code, msg) }
func Forbidden(code, msg string) *Error    { return New(http.StatusForbidden, code, msg) }
func NotFound(code, msg string) *Error     { return New(http.StatusNotFound, code, msg) }
func Internal(code, msg string) *Error     { return New(http.StatusInternalServerError, code, msg) }

// envelope is the wire format: {"error": {code, message, request_id}}.
type envelope struct {
	Error struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		RequestID string `json:"request_id,omitempty"`
	} `json:"error"`
}

// Write serializes err as the unified error response. Non-*Error values are
// treated as internal errors and the message is not leaked.
//
// Prefer httpx.WriteError in HTTP handlers so responses align with pkg/contracts.
func Write(w http.ResponseWriter, requestID string, err error) {
	e, ok := err.(*Error)
	if !ok {
		e = Internal("internal", "internal server error")
	}
	var env envelope
	env.Error.Code = e.Code
	env.Error.Message = e.Message
	env.Error.RequestID = requestID

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(e.Status)
	_ = json.NewEncoder(w).Encode(env)
}
