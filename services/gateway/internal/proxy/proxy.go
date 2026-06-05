// Package proxy provides prefix-based reverse proxying for the Gateway.
package proxy

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sort"
	"strings"
)

// Routes maps a path prefix to an upstream base URL.
type Routes map[string]string

// Router dispatches requests to upstreams by longest-prefix match.
type Router struct {
	prefixes []string
	proxies  map[string]*httputil.ReverseProxy
	log      *slog.Logger
}

// New builds a Router from the route table.
func New(routes Routes, log *slog.Logger) (*Router, error) {
	r := &Router{proxies: make(map[string]*httputil.ReverseProxy), log: log}
	for prefix, target := range routes {
		u, err := url.Parse(target)
		if err != nil {
			return nil, err
		}
		r.prefixes = append(r.prefixes, prefix)
		r.proxies[prefix] = httputil.NewSingleHostReverseProxy(u)
	}
	// Longest prefix first so "/v1/users/" wins over "/".
	sort.Slice(r.prefixes, func(i, j int) bool {
		return len(r.prefixes[i]) > len(r.prefixes[j])
	})
	return r, nil
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for _, prefix := range r.prefixes {
		if strings.HasPrefix(req.URL.Path, prefix) {
			r.proxies[prefix].ServeHTTP(w, req)
			return
		}
	}
	http.Error(w, `{"error":{"code":"not_found","message":"no route"}}`, http.StatusNotFound)
}
