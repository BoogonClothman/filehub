package handlers

import (
	"net/http"
	"os"
)

// Preview serves image files for inline preview.
func (h *Handler) Preview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rel := r.URL.Query().Get("path")
	abs, ok := h.safePath(rel)
	if !ok {
		jsonError(w, "invalid path", http.StatusBadRequest)
		return
	}

	info, err := os.Stat(abs)
	if err != nil || info.IsDir() || !isPreviewable(abs) {
		jsonError(w, "not found or not previewable", http.StatusNotFound)
		return
	}

	ct := detectMIME(info.Name())
	w.Header().Set("Content-Type", ct)
	http.ServeFile(w, r, abs)
}
