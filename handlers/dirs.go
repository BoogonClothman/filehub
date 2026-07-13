package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
)

// CreateDirRequest is the JSON body for creating a directory.
type CreateDirRequest struct {
	Path string `json:"path"`
}

// Dirs handles directory creation.
func (h *Handler) Dirs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateDirRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Path == "" {
		jsonError(w, "path is required", http.StatusBadRequest)
		return
	}

	abs, ok := h.safePath(req.Path)
	if !ok {
		jsonError(w, "invalid path", http.StatusBadRequest)
		return
	}

	if err := os.MkdirAll(abs, 0755); err != nil {
		jsonError(w, "cannot create directory", http.StatusInternalServerError)
		return
	}

	name := filepath.Base(abs)
	logRequest(r, "created dir "+name)
	jsonOK(w, map[string]string{
		"status": "created",
		"name":   name,
	})
}
