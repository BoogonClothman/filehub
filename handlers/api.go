package handlers

import (
	"encoding/json"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Handler holds shared dependencies for HTTP handlers.
type Handler struct {
	dataRoot    string
	maxUploadMB int64
}

// New creates a new Handler.
func New(dataRoot string, maxUploadMB int64) *Handler {
	return &Handler{
		dataRoot:    dataRoot,
		maxUploadMB: maxUploadMB,
	}
}

// safePath prevents directory traversal by resolving and validating a relative path
// stays within dataRoot. Returns the absolute path and true if valid.
func (h *Handler) safePath(rel string) (string, bool) {
	// Clean the path and remove any leading slashes
	rel = filepath.Clean(rel)
	rel = strings.TrimPrefix(rel, "/")
	rel = strings.TrimPrefix(rel, "\\")

	abs := filepath.Join(h.dataRoot, rel)

	// Verify the resolved path is within dataRoot
	abs, err := filepath.Abs(abs)
	if err != nil {
		return "", false
	}

	root, err := filepath.Abs(h.dataRoot)
	if err != nil {
		return "", false
	}

	// Ensure abs is inside root
	relCheck, err := filepath.Rel(root, abs)
	if err != nil || strings.HasPrefix(relCheck, "..") {
		return "", false
	}

	return abs, true
}

// safeFilename removes characters that are illegal in Windows filenames.
func safeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	return replacer.Replace(name)
}

// jsonError writes a JSON error response.
func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// jsonOK writes a JSON success response.
func jsonOK(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(data)
}

// detectMIME returns the MIME type based on file extension.
func detectMIME(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	// Common types not in the default mime package
	custom := map[string]string{
		".svg":  "image/svg+xml",
		".webp": "image/webp",
		".ico":  "image/x-icon",
		".md":   "text/markdown; charset=utf-8",
		".json": "application/json",
		".js":   "application/javascript; charset=utf-8",
		".css":  "text/css; charset=utf-8",
		".html": "text/html; charset=utf-8",
		".txt":  "text/plain; charset=utf-8",
		".xml":  "application/xml; charset=utf-8",
	}
	if ct, ok := custom[ext]; ok {
		return ct
	}
	ct := mime.TypeByExtension(ext)
	if ct != "" {
		return ct
	}
	return "application/octet-stream"
}

// isPreviewable returns true for image types that can be previewed.
func isPreviewable(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".bmp", ".ico":
		return true
	}
	return false
}

// copyReader copies from reader to file, limited to maxBytes.
func copyLimited(dst *os.File, src io.Reader, maxBytes int64) (int64, error) {
	if maxBytes <= 0 {
		return io.Copy(dst, src)
	}
	return io.CopyN(dst, src, maxBytes+1) // +1 to detect overflow
}

// logRequest logs incoming API calls.
func logRequest(r *http.Request, msg string) {
	log.Printf("[%s] %s %s — %s", r.Method, r.URL.Path, r.URL.RawQuery, msg)
}
