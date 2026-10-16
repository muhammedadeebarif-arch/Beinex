package handler

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"ems-backend/internal/response"
)

// SPAHandler serves the compiled Angular Single Page Application and its static assets.
type SPAHandler struct {
	staticDir string
	indexPath string
}

// NewSPAHandler constructs a new SPAHandler.
func NewSPAHandler(staticDir string) *SPAHandler {
	indexPath := filepath.Join(staticDir, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		log.Printf("[WARN] Frontend index.html not found at %s: %v", indexPath, err)
	} else {
		log.Printf("[INFO] Serving Angular frontend from %s", staticDir)
	}

	return &SPAHandler{
		staticDir: staticDir,
		indexPath: indexPath,
	}
}

func (h *SPAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. Never intercept API routes or return HTML for missing API endpoints
	if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/api" {
		response.NotFound(w, "API endpoint not found.")
		return
	}

	cleanPath := filepath.Clean(r.URL.Path)
	targetFile := filepath.Join(h.staticDir, cleanPath)

	// 2. Check if the static asset exists directly (e.g., JS, CSS, images, favicon)
	if info, err := os.Stat(targetFile); err == nil && !info.IsDir() {
		http.ServeFile(w, r, targetFile)
		return
	}

	// 3. If a specific asset file with an extension was requested but not found, return 404 (not index.html)
	if ext := filepath.Ext(cleanPath); ext != "" && ext != ".html" {
		http.NotFound(w, r)
		return
	}

	// 4. For root path and client-side Angular routes, serve index.html
	if _, err := os.Stat(h.indexPath); err == nil {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		http.ServeFile(w, r, h.indexPath)
		return
	}

	// If index.html is missing (e.g., frontend hasn't been built yet)
	response.NotFound(w, "Frontend application is not compiled. Please run 'npm run build' in frontend/.")
}
