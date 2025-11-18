// Copyright 2025 Patrick Deglon
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package frontend provides utilities for frontend assets management,
// template rendering, and HTMX integration.
package frontend

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/patdeg/common"
)

// AssetManager manages static assets with versioning and caching
type AssetManager struct {
	basePath    string
	urlPrefix   string
	cache       map[string]*Asset
	hashCache   map[string]string
	mu          sync.RWMutex
	development bool
}

// Asset represents a static asset
type Asset struct {
	Path        string
	Content     []byte
	ContentType string
	Hash        string
	ModTime     time.Time
}

// NewAssetManager creates a new asset manager
func NewAssetManager(basePath, urlPrefix string, development bool) *AssetManager {
	return &AssetManager{
		basePath:    basePath,
		urlPrefix:   strings.TrimRight(urlPrefix, "/"),
		cache:       make(map[string]*Asset),
		hashCache:   make(map[string]string),
		development: development,
	}
}

// GetAssetURL returns a versioned URL for an asset
func (am *AssetManager) GetAssetURL(path string) string {
	hash := am.getAssetHash(path)
	if hash != "" {
		return fmt.Sprintf("%s/%s?v=%s", am.urlPrefix, path, hash)
	}
	return fmt.Sprintf("%s/%s", am.urlPrefix, path)
}

// ServeHTTP serves static assets with caching headers
func (am *AssetManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Remove URL prefix
	path := strings.TrimPrefix(r.URL.Path, am.urlPrefix)
	path = strings.TrimPrefix(path, "/")

	// Secure path validation to prevent directory traversal attacks
	validPath, err := common.ValidatePath(am.basePath, path)
	if err != nil {
		common.Error("Path traversal attempt blocked: %s (requested: %s)", err, path)
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	// Get or load asset using validated path
	asset, err := am.getAssetFromPath(validPath)
	if err != nil {
		http.Error(w, "Asset not found", http.StatusNotFound)
		return
	}

	// Set headers
	w.Header().Set("Content-Type", asset.ContentType)

	if !am.development {
		// Production caching
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		w.Header().Set("ETag", asset.Hash)

		// Check if-none-match
		if r.Header.Get("If-None-Match") == asset.Hash {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	} else {
		// Development - no caching
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	}

	// Serve content
	if _, err := w.Write(asset.Content); err != nil {
		common.Error("Failed to write asset response: %v", err)
	}
}

// getAssetFromPath loads an asset from a validated full path.
// The fullPath argument must have been produced by common.ValidatePath to
// prevent directory traversal.
func (am *AssetManager) getAssetFromPath(fullPath string) (*Asset, error) {
	// Extract relative path for caching
	relPath, err := filepath.Rel(am.basePath, fullPath)
	if err != nil {
		return nil, err
	}

	// Check cache in production
	if !am.development {
		am.mu.RLock()
		if asset, ok := am.cache[relPath]; ok {
			am.mu.RUnlock()
			return asset, nil
		}
		am.mu.RUnlock()
	}

	// Load from filesystem using validated path
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}

	// #nosec G304 -- fullPath must be validated via common.ValidatePath before calling.
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}

	// Calculate hash using SHA-256 (more secure than MD5)
	hash := fmt.Sprintf("%x", sha256.Sum256(content))

	// Determine content type
	contentType := getContentType(relPath)

	asset := &Asset{
		Path:        relPath,
		Content:     content,
		ContentType: contentType,
		Hash:        hash[:8], // Use first 8 chars of hash
		ModTime:     info.ModTime(),
	}

	// Cache in production
	if !am.development {
		am.mu.Lock()
		am.cache[relPath] = asset
		am.hashCache[relPath] = asset.Hash
		am.mu.Unlock()
	}

	return asset, nil
}

// getAssetHash returns the hash for an asset
func (am *AssetManager) getAssetHash(path string) string {
	if am.development {
		// Always return current timestamp in development
		return fmt.Sprintf("%d", time.Now().Unix())
	}

	am.mu.RLock()
	if hash, ok := am.hashCache[path]; ok {
		am.mu.RUnlock()
		return hash
	}
	am.mu.RUnlock()

	// Try to load asset to get hash using validated path
	validPath, err := common.ValidatePath(am.basePath, path)
	if err != nil {
		common.Error("Path traversal attempt blocked when computing hash: %s (requested: %s)", err, path)
		return ""
	}

	if asset, err := am.getAssetFromPath(validPath); err == nil {
		return asset.Hash
	}

	return ""
}

// getContentType determines content type from file extension
func getContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".html":
		return "text/html"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	case ".eot":
		return "application/vnd.ms-fontobject"
	default:
		return "application/octet-stream"
	}
}

// TemplateManager manages HTML templates with caching
type TemplateManager struct {
	basePath    string
	funcMap     template.FuncMap
	cache       map[string]*template.Template
	mu          sync.RWMutex
	development bool
}

// NewTemplateManager creates a new template manager
func NewTemplateManager(basePath string, development bool) *TemplateManager {
	return &TemplateManager{
		basePath:    basePath,
		cache:       make(map[string]*template.Template),
		development: development,
		funcMap:     DefaultFuncMap(),
	}
}

// Render renders a template with data
func (tm *TemplateManager) Render(w io.Writer, name string, data interface{}) error {
	tmpl, err := tm.getTemplate(name)
	if err != nil {
		return fmt.Errorf("failed to get template: %v", err)
	}

	return tmpl.Execute(w, data)
}

// RenderString renders a template to a string
func (tm *TemplateManager) RenderString(name string, data interface{}) (string, error) {
	var buf bytes.Buffer
	if err := tm.Render(&buf, name, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// getTemplate loads or retrieves a template from cache
func (tm *TemplateManager) getTemplate(name string) (*template.Template, error) {
	// Check cache in production
	if !tm.development {
		tm.mu.RLock()
		if tmpl, ok := tm.cache[name]; ok {
			tm.mu.RUnlock()
			return tmpl, nil
		}
		tm.mu.RUnlock()
	}

	// Load template
	tmplPath := filepath.Join(tm.basePath, name)

	tmpl, err := template.New(filepath.Base(name)).
		Funcs(tm.funcMap).
		ParseFiles(tmplPath)
	if err != nil {
		return nil, err
	}

	// Look for layout
	layoutPath := filepath.Join(tm.basePath, "layout.html")
	if _, err := os.Stat(layoutPath); err == nil {
		tmpl, err = tmpl.ParseFiles(layoutPath)
		if err != nil {
			return nil, err
		}
	}

	// Cache in production
	if !tm.development {
		tm.mu.Lock()
		tm.cache[name] = tmpl
		tm.mu.Unlock()
	}

	return tmpl, nil
}

// AddFunc adds a template function
func (tm *TemplateManager) AddFunc(name string, fn interface{}) {
	tm.funcMap[name] = fn
}

// DefaultFuncMap returns default template functions
func DefaultFuncMap() template.FuncMap {
	return template.FuncMap{
		// String functions
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"title": strings.Title,
		"trim":  strings.TrimSpace,

		// Date functions
		"date": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
		"datetime": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},
		"timeago": func(t time.Time) string {
			d := time.Since(t)
			switch {
			case d < time.Minute:
				return "just now"
			case d < time.Hour:
				return fmt.Sprintf("%d minutes ago", int(d.Minutes()))
			case d < 24*time.Hour:
				return fmt.Sprintf("%d hours ago", int(d.Hours()))
			default:
				return fmt.Sprintf("%d days ago", int(d.Hours()/24))
			}
		},

		// Formatting functions
		"json": func(v interface{}) (string, error) {
			b, err := json.Marshal(v)
			return string(b), err
		},
		"truncate": func(s string, n int) string {
			if len(s) <= n {
				return s
			}
			return s[:n] + "..."
		},

		// URL functions
		"url": func(path string, params ...string) string {
			if len(params) == 0 {
				return path
			}

			query := url.Values{}
			for i := 0; i < len(params)-1; i += 2 {
				query.Add(params[i], params[i+1])
			}

			if strings.Contains(path, "?") {
				return path + "&" + query.Encode()
			}
			return path + "?" + query.Encode()
		},

		// Logic functions
		"default": func(def, val interface{}) interface{} {
			if val == nil || val == "" || val == 0 || val == false {
				return def
			}
			return val
		},
		"contains":  strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,
	}
}

// HTMXResponse helps build HTMX responses
type HTMXResponse struct {
	w       http.ResponseWriter
	headers map[string]string
}

// NewHTMXResponse creates a new HTMX response builder
func NewHTMXResponse(w http.ResponseWriter) *HTMXResponse {
	return &HTMXResponse{
		w:       w,
		headers: make(map[string]string),
	}
}

// Trigger sets HX-Trigger header
func (h *HTMXResponse) Trigger(events ...string) *HTMXResponse {
	h.headers["HX-Trigger"] = strings.Join(events, ",")
	return h
}

// TriggerAfterSwap sets HX-Trigger-After-Swap header
func (h *HTMXResponse) TriggerAfterSwap(events ...string) *HTMXResponse {
	h.headers["HX-Trigger-After-Swap"] = strings.Join(events, ",")
	return h
}

// TriggerAfterSettle sets HX-Trigger-After-Settle header
func (h *HTMXResponse) TriggerAfterSettle(events ...string) *HTMXResponse {
	h.headers["HX-Trigger-After-Settle"] = strings.Join(events, ",")
	return h
}

// Redirect sets HX-Redirect header
func (h *HTMXResponse) Redirect(url string) *HTMXResponse {
	h.headers["HX-Redirect"] = url
	return h
}

// Refresh sets HX-Refresh header
func (h *HTMXResponse) Refresh() *HTMXResponse {
	h.headers["HX-Refresh"] = "true"
	return h
}

// PushURL sets HX-Push-Url header
func (h *HTMXResponse) PushURL(url string) *HTMXResponse {
	h.headers["HX-Push-Url"] = url
	return h
}

// ReplaceURL sets HX-Replace-Url header
func (h *HTMXResponse) ReplaceURL(url string) *HTMXResponse {
	h.headers["HX-Replace-Url"] = url
	return h
}

// Retarget sets HX-Retarget header
func (h *HTMXResponse) Retarget(selector string) *HTMXResponse {
	h.headers["HX-Retarget"] = selector
	return h
}

// Reswap sets HX-Reswap header
func (h *HTMXResponse) Reswap(method string) *HTMXResponse {
	h.headers["HX-Reswap"] = method
	return h
}

// Write writes the response with HTMX headers
func (h *HTMXResponse) Write(content []byte) (int, error) {
	// Set HTMX headers
	for k, v := range h.headers {
		h.w.Header().Set(k, v)
	}

	// Write content
	return h.w.Write(content)
}

// WriteHTML writes HTML content with HTMX headers
func (h *HTMXResponse) WriteHTML(html string) (int, error) {
	h.w.Header().Set("Content-Type", "text/html")
	return h.Write([]byte(html))
}

// IsHTMXRequest checks if a request is from HTMX
func IsHTMXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// GetHTMXTarget returns the HTMX target element
func GetHTMXTarget(r *http.Request) string {
	return r.Header.Get("HX-Target")
}

// GetHTMXTrigger returns the HTMX trigger element
func GetHTMXTrigger(r *http.Request) string {
	return r.Header.Get("HX-Trigger")
}

// GetHTMXTriggerName returns the HTMX trigger name
func GetHTMXTriggerName(r *http.Request) string {
	return r.Header.Get("HX-Trigger-Name")
}

// GetHTMXPrompt returns the HTMX prompt response
func GetHTMXPrompt(r *http.Request) string {
	return r.Header.Get("HX-Prompt")
}
