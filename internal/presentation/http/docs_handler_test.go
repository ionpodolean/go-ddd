package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDocsHandler_ServeDocs(t *testing.T) {
	handler := NewDocsHandler("docs/assets/manifest.json", "docs/assets")

	tests := []struct {
		name           string
		url            string
		expectedStatus int
		containsText   string
	}{
		{
			name:           "Home page when page param missing",
			url:            "/docs",
			expectedStatus: http.StatusOK,
			containsText:   "Go DDD Technical Documentation",
		},
		{
			name:           "Onboarding guide page",
			url:            "/docs?page=onboarding",
			expectedStatus: http.StatusOK,
			containsText:   "Developer Onboarding Guide",
		},
		{
			name:           "Architecture reference page",
			url:            "/docs?page=architecture",
			expectedStatus: http.StatusOK,
			containsText:   "Architecture Reference",
		},
		{
			name:           "External client integration page",
			url:            "/docs?page=external-client-integration",
			expectedStatus: http.StatusOK,
			containsText:   "External Client Integration Guide",
		},
		{
			name:           "Error builder guide page",
			url:            "/docs?page=error-builder",
			expectedStatus: http.StatusOK,
			containsText:   "Error Builder &amp; Handling Guide",
		},
		{
			name:           "User module page",
			url:            "/docs?page=user-module",
			expectedStatus: http.StatusOK,
			containsText:   "User Management Module",
		},
		{
			name:           "Feature template page",
			url:            "/docs?page=feature-template",
			expectedStatus: http.StatusOK,
			containsText:   "Feature Module Title",
		},
		{
			name:           "Unknown page key returns 404",
			url:            "/docs?page=unknown-key-12345",
			expectedStatus: http.StatusNotFound,
			containsText:   "404",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			rr := httptest.NewRecorder()

			handler.ServeDocs(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			body := rr.Body.String()
			if !strings.Contains(body, tt.containsText) {
				t.Errorf("expected body to contain %q, body snippet: %q", tt.containsText, body[:min(200, len(body))])
			}

			if tt.expectedStatus == http.StatusOK {
				contentType := rr.Header().Get("Content-Type")
				if !strings.Contains(contentType, "text/html") {
					t.Errorf("expected Content-Type text/html, got %s", contentType)
				}
			}
		})
	}
}

func TestDocsHandler_ServeAssets(t *testing.T) {
	handler := NewDocsHandler("docs/assets/manifest.json", "docs/assets")
	router := http.NewServeMux()
	router.HandleFunc("GET /docs/assets/{filename}", handler.ServeAssets)

	t.Run("Serve docs.css asset", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/docs/assets/docs.css", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}

		contentType := rr.Header().Get("Content-Type")
		if !strings.Contains(contentType, "text/css") {
			t.Errorf("expected text/css, got %s", contentType)
		}

		cacheControl := rr.Header().Get("Cache-Control")
		if cacheControl == "" {
			t.Errorf("expected Cache-Control header to be set")
		}
	})

	t.Run("Serve manifest.json asset", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/docs/assets/manifest.json", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}

		contentType := rr.Header().Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			t.Errorf("expected application/json, got %s", contentType)
		}
	})

	t.Run("Missing asset returns 404", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/docs/assets/nonexistent.png", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", rr.Code)
		}
	})
}
