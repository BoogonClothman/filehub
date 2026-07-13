package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
)

// RenameRequest is the JSON body for renaming.
type RenameRequest struct {
	Path    string `json:"path"`
	NewName string `json:"newName"`
}

// Rename handles file/directory rename.
func (h *Handler) Rename(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RenameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Path == "" || req.NewName == "" {
		jsonError(w, "path and newName are required", http.StatusBadRequest)
		return
	}

	// Sanitize new name
	req.NewName = safeFilename(req.NewName)
	if req.NewName == "" || req.NewName == "." || req.NewName == ".." {
		jsonError(w, "invalid name", http.StatusBadRequest)
		return
	}

	src, ok := h.safePath(req.Path)
	if !ok {
		jsonError(w, "invalid path", http.StatusBadRequest)
		return
	}

	// Check source exists
	if _, err := os.Stat(src); err != nil {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}

	// Destination is in the same directory
	dst := filepath.Join(filepath.Dir(src), req.NewName)

	// Additional safety: ensure dst is inside dataRoot
	dstAbs, err := filepath.Abs(dst)
	if err != nil {
		jsonError(w, "invalid destination", http.StatusBadRequest)
		return
	}
	root, _ := filepath.Abs(h.dataRoot)
	if rel, err := filepath.Rel(root, dstAbs); err != nil || filepath.IsAbs(rel) && rel[0:3] == ".." {
		jsonError(w, "invalid destination", http.StatusBadRequest)
		return
	}

	if _, ok := h.safePath(filepath.Join(filepath.Dir(req.Path), req.NewName)); !ok {
		jsonError(w, "invalid destination", http.StatusBadRequest)
		return
	}

	// Check destination doesn't exist
	if _, err := os.Stat(dst); err == nil {
		jsonError(w, "name already exists", http.StatusConflict)
		return
	}

	if err := os.Rename(src, dst); err != nil {
		jsonError(w, "cannot rename", http.StatusInternalServerError)
		return
	}

	logRequest(r, "renamed "+filepath.Base(req.Path)+" → "+req.NewName)
	jsonOK(w, map[string]string{
		"status":  "renamed",
		"newName": req.NewName,
	})
}
