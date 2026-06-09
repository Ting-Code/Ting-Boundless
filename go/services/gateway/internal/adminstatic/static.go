// Package adminstatic serves the built @ting/admin SPA at /admin (local dev without nginx).
package adminstatic

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Handler serves files from distDir. Unknown paths fall back to index.html (SPA routing).
func Handler(distDir string) http.Handler {
	distDir = filepath.Clean(distDir)
	index := filepath.Join(distDir, "index.html")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.NotFound(w, r)
			return
		}

		rel := strings.TrimPrefix(r.URL.Path, "/admin")
		rel = strings.TrimPrefix(rel, "/")
		if rel == "" {
			http.ServeFile(w, r, index)
			return
		}

		target := filepath.Join(distDir, filepath.Clean(rel))
		if !strings.HasPrefix(target, distDir+string(filepath.Separator)) && target != distDir {
			http.NotFound(w, r)
			return
		}

		info, err := os.Stat(target)
		if err != nil || info.IsDir() {
			http.ServeFile(w, r, index)
			return
		}

		http.ServeFile(w, r, target)
	})
}

// ResolveDir returns the first existing directory from candidates.
func ResolveDir(candidates ...string) string {
	for _, dir := range candidates {
		if dir == "" {
			continue
		}
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			if _, err := os.Stat(filepath.Join(dir, "index.html")); err == nil {
				return dir
			}
		}
	}
	return ""
}
