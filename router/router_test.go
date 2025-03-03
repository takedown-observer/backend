package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/takedown-observer/backend/api"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestHandler(t *testing.T) *api.Handler {
	// Create a temporary database
	tmpDir, err := os.MkdirTemp("", "takedown-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	db, err := gorm.Open(sqlite.Open(filepath.Join(tmpDir, "test.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	return api.NewHandler(db)
}

func TestRouter(t *testing.T) {
	handler := setupTestHandler(t)
	router := New(handler)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "GET dashboard",
			method:         "GET",
			path:           "/dashboard",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET invalid path",
			method:         "GET",
			path:           "/invalid",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "GET invalid API path",
			method:         "GET",
			path:           "/api/invalid",
			expectedStatus: http.StatusNotFound,
		},
	}

	// Create a static directory and index.html for testing
	tmpDir, err := os.MkdirTemp("", "static-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create static directory
	staticDir := filepath.Join(tmpDir, "static")
	if err := os.MkdirAll(staticDir, 0755); err != nil {
		t.Fatalf("Failed to create static directory: %v", err)
	}

	// Create index.html
	indexPath := filepath.Join(staticDir, "index.html")
	if err := os.WriteFile(indexPath, []byte("<html><body>Test</body></html>"), 0644); err != nil {
		t.Fatalf("Failed to create index.html: %v", err)
	}

	// Set working directory to temp directory for static file serving
	originalWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.method == "OPTIONS" {
				req.Header.Set("Origin", "https://twitter.com")
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.method == "OPTIONS" {
				cors := w.Header().Get("Access-Control-Allow-Origin")
				if cors != "https://twitter.com" {
					t.Errorf("Expected CORS header to be https://twitter.com, got %s", cors)
				}
			}
		})
	}
}
