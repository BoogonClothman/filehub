package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FileEntry represents a file or directory in listings.
type FileEntry struct {
	Name      string `json:"name"`
	IsDir     bool   `json:"isDir"`
	Size      int64  `json:"size"`
	ModTime   string `json:"modTime"`
	Preview   bool   `json:"preview"`
}

// DirList is the response for directory listing.
type DirList struct {
	Path    string      `json:"path"`
	Entries []FileEntry `json:"entries"`
}

// Files handles GET (list / download), POST (upload), DELETE.
func (h *Handler) Files(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listOrDownload(w, r)
	case http.MethodPost:
		h.upload(w, r)
	case http.MethodDelete:
		h.deleteFile(w, r)
	default:
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) listOrDownload(w http.ResponseWriter, r *http.Request) {
	rel := r.URL.Query().Get("path")
	abs, ok := h.safePath(rel)
	if !ok {
		jsonError(w, "invalid path", http.StatusBadRequest)
		return
	}

	info, err := os.Stat(abs)
	if err != nil {
		if os.IsNotExist(err) {
			jsonError(w, "not found", http.StatusNotFound)
			return
		}
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Download if it's a file or ?download=1
	if !info.IsDir() || r.URL.Query().Get("download") == "1" {
		h.download(w, r, abs, info)
		return
	}

	// List directory
	entries, err := os.ReadDir(abs)
	if err != nil {
		jsonError(w, "cannot read directory", http.StatusInternalServerError)
		return
	}

	result := DirList{
		Path:    rel,
		Entries: make([]FileEntry, 0, len(entries)),
	}

	for _, e := range entries {
		ei, err := e.Info()
		if err != nil {
			continue
		}
		name := e.Name()
		// Hide hidden files
		if strings.HasPrefix(name, ".") {
			continue
		}
		entry := FileEntry{
			Name:    name,
			IsDir:   e.IsDir(),
			Size:    ei.Size(),
			ModTime: ei.ModTime().Format("2006-01-02 15:04"),
			Preview: false,
		}
		if !e.IsDir() && isPreviewable(name) {
			entry.Preview = true
		}
		result.Entries = append(result.Entries, entry)
	}

	// Sort: directories first, then by name
	sort.Slice(result.Entries, func(i, j int) bool {
		if result.Entries[i].IsDir != result.Entries[j].IsDir {
			return result.Entries[i].IsDir
		}
		return strings.ToLower(result.Entries[i].Name) < strings.ToLower(result.Entries[j].Name)
	})

	jsonOK(w, result)
}

func (h *Handler) download(w http.ResponseWriter, r *http.Request, abs string, info os.FileInfo) {
	name := filepath.Base(abs)
	ct := detectMIME(name)
	if ct == "application/octet-stream" {
		// Force download for non-browser-friendly types
		w.Header().Set("Content-Disposition", "attachment; filename=\""+name+"\"")
	}
	w.Header().Set("Content-Type", ct)
	http.ServeFile(w, r, abs)
	logRequest(r, "download "+name)
}

func (h *Handler) upload(w http.ResponseWriter, r *http.Request) {
	// Limit upload size
	maxBytes := h.maxUploadMB * 1024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes+1)

	if err := r.ParseMultipartForm(maxBytes); err != nil {
		jsonError(w, "file too large or malformed request", http.StatusBadRequest)
		return
	}
	defer r.MultipartForm.RemoveAll()

	dirPath := r.FormValue("path")
	absDir, ok := h.safePath(dirPath)
	if !ok {
		jsonError(w, "invalid path", http.StatusBadRequest)
		return
	}

	// Ensure target directory exists
	if err := os.MkdirAll(absDir, 0755); err != nil {
		jsonError(w, "cannot create directory", http.StatusInternalServerError)
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		jsonError(w, "no files provided", http.StatusBadRequest)
		return
	}

	uploaded := make([]string, 0, len(files))
	for _, fh := range files {
		name := safeFilename(fh.Filename)
		if name == "" || name == "." || name == ".." {
			continue
		}

		dest := filepath.Join(absDir, name)

		src, err := fh.Open()
		if err != nil {
			continue
		}

		dst, err := os.Create(dest)
		if err != nil {
			src.Close()
			continue
		}

		written, err := ioCopy(dst, src, maxBytes)
		src.Close()
		dst.Close()

		if err != nil || written > maxBytes {
			os.Remove(dest)
			continue
		}

		uploaded = append(uploaded, name)
		logRequest(r, "uploaded "+name+" ("+formatSize(written)+")")
	}

	if len(uploaded) == 0 {
		jsonError(w, "upload failed", http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]interface{}{
		"uploaded": uploaded,
		"count":    len(uploaded),
	})
}

func (h *Handler) deleteFile(w http.ResponseWriter, r *http.Request) {
	rel := r.URL.Query().Get("path")
	abs, ok := h.safePath(rel)
	if !ok {
		jsonError(w, "invalid path", http.StatusBadRequest)
		return
	}

	info, err := os.Stat(abs)
	if err != nil {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}

	if err := os.RemoveAll(abs); err != nil {
		jsonError(w, "cannot delete", http.StatusInternalServerError)
		return
	}

	if info.IsDir() {
		logRequest(r, "deleted dir "+filepath.Base(abs))
	} else {
		logRequest(r, "deleted file "+filepath.Base(abs))
	}

	jsonOK(w, map[string]string{"status": "deleted"})
}

// ioCopy copies up to maxBytes, returning the number of bytes written.
func ioCopy(dst *os.File, src interface{ Read([]byte) (int, error) }, maxBytes int64) (int64, error) {
	if maxBytes <= 0 {
		return ioCopyN(dst, src, -1)
	}
	return ioCopyN(dst, src, maxBytes+1) // +1 to detect overflow
}

func ioCopyN(dst *os.File, src interface{ Read([]byte) (int, error) }, limit int64) (int64, error) {
	if limit < 0 {
		buf := make([]byte, 32*1024)
		var total int64
		for {
			nr, err := src.Read(buf)
			if nr > 0 {
				nw, ew := dst.Write(buf[:nr])
				if ew != nil {
					return total, ew
				}
				total += int64(nw)
			}
			if err != nil {
				if err.Error() == "EOF" {
					return total, nil
				}
				return total, err
			}
		}
	}

	buf := make([]byte, 32*1024)
	var total int64
	for total < limit {
		nr, err := src.Read(buf)
		if nr > 0 {
			if total+int64(nr) > limit {
				nr = int(limit - total)
			}
			nw, ew := dst.Write(buf[:nr])
			if ew != nil {
				return total, ew
			}
			total += int64(nw)
		}
		if err != nil {
			if err.Error() == "EOF" {
				return total, nil
			}
			return total, err
		}
	}
	return total, nil
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return formatInt(bytes) + " B"
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return formatFloat(float64(bytes)/float64(div)) + " " + string("KMGTPE"[exp]) + "B"
}

func formatInt(n int64) string {
	if n < 0 {
		return "-" + formatUint(uint64(-n))
	}
	return formatUint(uint64(n))
}

func formatUint(n uint64) string {
	var buf [20]byte
	i := len(buf)
	for {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
		if n == 0 {
			return string(buf[i:])
		}
	}
}

func formatFloat(f float64) string {
	// Simple 1-decimal formatting without fmt
	n := int(f * 10)
	return formatInt(int64(n/10)) + "." + formatInt(int64(n%10))
}
