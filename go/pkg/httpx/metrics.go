package httpx

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// RegisterMetrics wires GET /metrics (Prometheus) on the mux.
func RegisterMetrics(mux *http.ServeMux) {
	mux.Handle("GET /metrics", promhttp.Handler())
}
