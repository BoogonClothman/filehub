package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"filehub/handlers"
)

//go:embed public/*
var publicFS embed.FS

func main() {
	port := flag.Int("port", 0, "HTTP port (default from config.json or 5000)")
	dataRoot := flag.String("data", "", "data root directory (default from config.json or ./data)")
	flag.Parse()

	// Load config
	cfg, err := LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// CLI flags override config
	if *port > 0 {
		cfg.Port = *port
	}
	if *dataRoot != "" {
		cfg.DataRoot = *dataRoot
	}

	// Resolve absolute data root
	absRoot, err := filepath.Abs(cfg.DataRoot)
	if err != nil {
		log.Fatalf("Failed to resolve data root: %v", err)
	}
	cfg.DataRoot = absRoot

	// Ensure data directory exists
	if err := os.MkdirAll(cfg.DataRoot, 0755); err != nil {
		log.Fatalf("Failed to create data root: %v", err)
	}

	// Setup API handlers
	h := handlers.New(cfg.DataRoot, cfg.MaxUploadMB)

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/files", h.Files)
	mux.HandleFunc("/api/dirs", h.Dirs)
	mux.HandleFunc("/api/rename", h.Rename)
	mux.HandleFunc("/api/preview", h.Preview)

	// Static files (embedded) — serve from public/ subdirectory
	subFS, err := fs.Sub(publicFS, "public")
	if err != nil {
		log.Fatalf("Failed to create sub filesystem: %v", err)
	}
	mux.Handle("/", http.FileServer(http.FS(subFS)))

	addr := fmt.Sprintf("0.0.0.0:%d", cfg.Port)
	log.Printf("📁 FileHub running at http://0.0.0.0:%d/", cfg.Port)
	log.Printf("   Data root: %s", cfg.DataRoot)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
