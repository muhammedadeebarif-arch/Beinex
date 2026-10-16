package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ems-backend/internal/response"
)

func TestSPAHandler(t *testing.T) {
	// Create a temporary mock directory with compiled frontend files
	tempDir, err := os.MkdirTemp("", "spa-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	indexContent := "<!doctype html><html><head><title>Test SPA</title></head><body><app-root></app-root></body></html>"
	if err := os.WriteFile(filepath.Join(tempDir, "index.html"), []byte(indexContent), 0644); err != nil {
		t.Fatalf("failed to write index.html: %v", err)
	}

	jsContent := "console.log('main bundle');"
	if err := os.WriteFile(filepath.Join(tempDir, "main.js"), []byte(jsContent), 0644); err != nil {
		t.Fatalf("failed to write main.js: %v", err)
	}

	spaHandler := NewSPAHandler(tempDir)

	t.Run("Serves index.html on root /", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		spaHandler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200 OK, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "<app-root>") {
			t.Errorf("expected index.html content, got %s", rec.Body.String())
		}
	})

	t.Run("Serves existing static asset /main.js with correct content", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/main.js", nil)
		rec := httptest.NewRecorder()

		spaHandler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200 OK, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "main bundle") {
			t.Errorf("expected main.js content, got %s", rec.Body.String())
		}
	})

	t.Run("Returns 404 for missing static file with extension", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/missing.js", nil)
		rec := httptest.NewRecorder()

		spaHandler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404 for missing .js file, got %d", rec.Code)
		}
		if strings.Contains(rec.Body.String(), "<app-root>") {
			t.Error("must NOT return index.html for missing static asset")
		}
	})

	t.Run("Serves index.html for Angular client-side route", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/employees", nil)
		rec := httptest.NewRecorder()

		spaHandler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200 OK, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "<app-root>") {
			t.Errorf("expected index.html content, got %s", rec.Body.String())
		}
	})

	t.Run("Never serves index.html for unknown /api/ route", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/unknown", nil)
		rec := httptest.NewRecorder()

		spaHandler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404 for unknown api endpoint, got %d", rec.Code)
		}

		var errResp response.ErrorResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
			t.Fatalf("expected JSON error response, got %s", rec.Body.String())
		}

		if errResp.Error.Code != "NOT_FOUND" {
			t.Errorf("expected code NOT_FOUND, got %s", errResp.Error.Code)
		}
	})
}
