package httpx

import (
	"context"
	"net/http"
	"sync"
	"time"
)

// Check is a named readiness probe for a dependency (DB, Redis, MQ, ...).
type Check struct {
	Name string
	// Probe returns nil when the dependency is reachable.
	Probe func(ctx context.Context) error
}

// Health holds liveness and readiness state for a service.
//
// /healthz = process liveness only (no dependency checks).
// /readyz  = readiness; runs all registered dependency checks.
//
// Note (see docs/ARCHITECTURE.md): the Gateway must NOT add Logto JWKS as a hard
// readiness check, because JWKS is cached and a Logto outage should only degrade,
// not take the Gateway offline.
type Health struct {
	mu     sync.RWMutex
	checks []Check
}

// NewHealth creates a Health registry.
func NewHealth() *Health { return &Health{} }

// Register adds a readiness check.
func (h *Health) Register(c Check) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checks = append(h.checks, c)
}

// Handler wires /healthz, /readyz, and /metrics onto the given mux.
func (h *Health) Handler(mux *http.ServeMux) {
	RegisterMetrics(mux)

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		h.mu.RLock()
		checks := h.checks
		h.mu.RUnlock()

		results := make(map[string]string, len(checks))
		ready := true
		for _, c := range checks {
			if err := c.Probe(ctx); err != nil {
				ready = false
				results[c.Name] = "error: " + err.Error()
			} else {
				results[c.Name] = "ok"
			}
		}

		status := http.StatusOK
		if !ready {
			status = http.StatusServiceUnavailable
		}
		JSON(w, status, map[string]any{"ready": ready, "checks": results})
	})
}
