package http

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Manifest represents the structure of docs/assets/manifest.json
type Manifest struct {
	CoreDocs map[string]struct {
		DocPath  string `json:"doc_path"`
		HTMLPath string `json:"html_path"`
	} `json:"core_docs"`
	Guides map[string]struct {
		DocPath  string `json:"doc_path"`
		HTMLPath string `json:"html_path"`
	} `json:"guides"`
	Modules map[string]struct {
		DocPath  string `json:"doc_path"`
		HTMLPath string `json:"html_path"`
	} `json:"modules"`
	Templates map[string]struct {
		DocPath  string `json:"doc_path"`
		HTMLPath string `json:"html_path"`
	} `json:"templates"`
}

type DocsHandler struct {
	pageMap    map[string]string
	projectRoot string
	assetsDir  string
}

func findProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "."
}

func NewDocsHandler(manifestRelPath string, assetsRelDir string) *DocsHandler {
	root := findProjectRoot()

	handler := &DocsHandler{
		projectRoot: root,
		pageMap: map[string]string{
			"index":                        "docs/html/index.html",
			"onboarding":                   "docs/html/core/onboarding-guide.html",
			"architecture":                 "docs/html/core/architecture-reference.html",
			"external-client-integration": "docs/html/guides/external-client-integration-guide.html",
			"error-builder":                "docs/html/guides/error-builder-guide.html",
			"user-module":                  "docs/html/modules/user-module.html",
			"feature-template":             "docs/html/templates/feature-template.html",
			"guide-template":               "docs/html/templates/guide-template.html",
			"swagger-description-template": "docs/html/templates/swagger-description-template.html",
		},
		assetsDir: filepath.Join(root, assetsRelDir),
	}

	manifestPath := filepath.Join(root, manifestRelPath)
	if manifestData, err := os.ReadFile(manifestPath); err == nil {
		var m Manifest
		if err := json.Unmarshal(manifestData, &m); err == nil {
			for k, v := range m.CoreDocs {
				if v.HTMLPath != "" {
					handler.pageMap[k] = v.HTMLPath
				}
			}
			for k, v := range m.Guides {
				if v.HTMLPath != "" {
					handler.pageMap[k] = v.HTMLPath
				}
			}
			for k, v := range m.Modules {
				if v.HTMLPath != "" {
					handler.pageMap[k] = v.HTMLPath
				}
			}
			for k, v := range m.Templates {
				if v.HTMLPath != "" {
					handler.pageMap[k] = v.HTMLPath
				}
			}
		}
	}

	return handler
}

// ServeDocs handles GET /docs and GET /docs?page=<page_key>
func (h *DocsHandler) ServeDocs(w http.ResponseWriter, r *http.Request) {
	page := r.URL.Query().Get("page")

	var targetRelFile string
	if page == "" {
		targetRelFile = h.pageMap["index"]
	} else {
		mapped, exists := h.pageMap[page]
		if !exists {
			h.serve404(w, "Documentation page not found in PAGE_MAP")
			return
		}
		targetRelFile = mapped
	}

	targetAbsFile := filepath.Join(h.projectRoot, targetRelFile)

	// Read and serve HTML content
	content, err := os.ReadFile(targetAbsFile)
	if err != nil {
		h.serve404(w, "Documentation HTML file missing on disk")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

// ServeAssets handles GET /docs/assets/{filename}
func (h *DocsHandler) ServeAssets(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")
	if filename == "" {
		filename = strings.TrimPrefix(r.URL.Path, "/docs/assets/")
	}

	// Prevent directory traversal
	filename = filepath.Clean(filename)
	if strings.Contains(filename, "..") || strings.HasPrefix(filename, "/") {
		http.Error(w, "Invalid asset path", http.StatusBadRequest)
		return
	}

	assetPath := filepath.Join(h.assetsDir, filename)
	content, err := os.ReadFile(assetPath)
	if err != nil {
		http.Error(w, "Asset not found", http.StatusNotFound)
		return
	}

	// Set content type by extension
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".css":
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".json":
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	// Cache header
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

func (h *DocsHandler) serve404(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	notFoundHTML := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>404 Page Not Found - Go DDD Docs</title>
  <link rel="stylesheet" href="/docs/assets/docs.css">
</head>
<body>
  <div class="docs-container" style="justify-content: center; align-items: center; min-height: 80vh;">
    <div style="text-align: center; max-width: 500px; padding: 2rem;">
      <h1 style="font-size: 3rem; color: #f87171;">404</h1>
      <h2>Documentation Page Not Found</h2>
      <p style="margin-top: 1rem; margin-bottom: 2rem; color: #94a3b8;">The page key you requested does not exist or the HTML artifact is missing.</p>
      <a class="back-link" href="/docs">← Return to Documentation Home</a>
    </div>
  </div>
</body>
</html>`
	_, _ = w.Write([]byte(notFoundHTML))
}

// GetPageMap returns a copy of the active PAGE_MAP for inspection/testing
func (h *DocsHandler) GetPageMap() map[string]string {
	res := make(map[string]string)
	for k, v := range h.pageMap {
		res[k] = v
	}
	return res
}
