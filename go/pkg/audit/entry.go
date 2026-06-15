package audit

import (
	"net/http"

	"github.com/ting-boundless/boundless/pkg/identity"
)

// EntryEvent builds a Gateway entry audit event for denied or invalid access.
func EntryEvent(eventType string, r *http.Request, requestID string, extra map[string]any) Event {
	ev := NewEvent(SourceGateway, eventType)
	ev.Subject = r.Method + " " + r.URL.Path
	ev.Data = map[string]any{
		"path":       r.URL.Path,
		"method":     r.Method,
		"request_id": requestID,
	}
	if ip := clientIP(r); ip != "" {
		ev.Data["client_ip"] = ip
	}
	for k, v := range extra {
		ev.Data[k] = v
	}
	return ev
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}

// EmitEntry schedules a gateway entry audit event when emitter is configured.
func EmitEntry(emitter Emitter, r *http.Request, requestID, eventType string, extra map[string]any) {
	if emitter == nil {
		return
	}
	_ = emitter.Emit(r.Context(), EntryEvent(eventType, r, requestID, extra))
}

// RequestIDFrom extracts the gateway-assigned request id from headers.
func RequestIDFrom(r *http.Request) string {
	return r.Header.Get(identity.HeaderRequestID)
}
