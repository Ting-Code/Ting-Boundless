// Package proxy provides prefix-based reverse proxying for the Gateway.
package proxy

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sort"
	"strings"

	"github.com/ting-boundless/boundless/pkg/errs"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/identity"
)

// Routes maps a path prefix to an upstream base URL.
type Routes map[string]string

// Router dispatches requests to upstreams by longest-prefix match.
type Router struct {
	prefixes []string
	proxies  map[string]*httputil.ReverseProxy
	log      *slog.Logger
}

// New builds a Router from the route table. When internalToken is non-empty, each
// proxied request receives X-Internal-Token so upstreams can trust Gateway traffic.
// Client traceparent and other hop headers are forwarded unchanged by the reverse proxy.
func New(routes Routes, log *slog.Logger, internalToken string) (*Router, error) {
	r := &Router{proxies: make(map[string]*httputil.ReverseProxy), log: log}
	for prefix, target := range routes {
		u, err := url.Parse(target)
		if err != nil {
			return nil, err
		}
		r.prefixes = append(r.prefixes, prefix)
		p := httputil.NewSingleHostReverseProxy(u)
		if internalToken != "" {
			token := internalToken
			director := p.Director
			p.Director = func(req *http.Request) {
				director(req)
				req.Header.Set("X-Internal-Token", token)
			}
		}
		r.proxies[prefix] = p
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
	httpx.WriteError(w, req.Header.Get(identity.HeaderRequestID), errs.NotFound("route.not_found", "no route"))
}
