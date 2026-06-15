// Package siteproxy reverse-proxies public SSR paths to the Next.js site service.
package siteproxy

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/ting-boundless/boundless/pkg/config"
	"github.com/ting-boundless/boundless/pkg/httpx"
)

// IsAPIPath reports whether the path is handled by the API reverse proxy table.
func IsAPIPath(path string) bool {
	return strings.HasPrefix(path, "/v1/") || strings.HasPrefix(path, "/internal/")
}

// Handler proxies browser requests to the Next.js app (SITE_SERVICE_URL).
type Handler struct {
	upstream http.Handler
	log      *slog.Logger
}

// FromEnv builds a site proxy. Returns nil when SITE_SERVICE_URL is unset or a placeholder.
func FromEnv(log *slog.Logger) (*Handler, error) {
	raw := strings.TrimSpace(httpx.Env("SITE_SERVICE_URL", "http://127.0.0.1:3006"))
	if raw == "" || config.IsPlaceholder(raw) {
		if log != nil {
			log.Warn("site proxy disabled (set SITE_SERVICE_URL for Next SSR)")
		}
		return nil, nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("site url: %w", err)
	}
	p := httputil.NewSingleHostReverseProxy(u)
	if log != nil {
		log.Info("site proxy enabled", slog.String("url", raw))
	}
	return &Handler{upstream: p, log: log}, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.upstream == nil {
		http.Error(w, `{"error":{"code":"not_found","message":"site not configured"}}`, http.StatusNotFound)
		return
	}
	h.upstream.ServeHTTP(w, r)
}

// ComposeAPIAndSite routes API prefixes to api; other paths to site when configured.
func ComposeAPIAndSite(api http.Handler, site *Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if IsAPIPath(r.URL.Path) {
			api.ServeHTTP(w, r)
			return
		}
		if site != nil {
			site.ServeHTTP(w, r)
			return
		}
		api.ServeHTTP(w, r)
	})
}
