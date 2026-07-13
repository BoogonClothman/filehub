package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"filehub/handlers"
)

//go:embed public/*
var publicFS embed.FS

// exeDir returns the directory containing the running executable.
// Relative paths (config, data) are resolved against this directory,
// so data follows the binary regardless of working directory.
func exeDir() string {
	exe, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}
	return filepath.Dir(exe)
}

// resolvePath returns path unchanged if absolute, otherwise joins it with base.
func resolvePath(base, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(base, path)
}

// Run parses CLI flags, loads config, and builds the HTTP server.
// It returns the configured server and the application config.
// The caller is responsible for starting the server and managing its lifecycle.
func Run() (*http.Server, Config) {
	exe := exeDir()

	port := flag.Int("port", 0, "HTTP port (default from config.json or 5000)")
	dataRoot := flag.String("data", "", "data root directory (default from config.json or ./data)")
	flag.Parse()

	cfg, err := LoadConfig(filepath.Join(exe, "config.json"))
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if *port > 0 {
		cfg.Port = *port
	}
	if *dataRoot != "" {
		cfg.DataRoot = *dataRoot
	}

	// Resolve data root relative to executable dir (not CWD).
	// An absolute path in config.json is used as-is.
	cfg.DataRoot = resolvePath(exe, cfg.DataRoot)

	if err := os.MkdirAll(cfg.DataRoot, 0755); err != nil {
		log.Fatalf("Failed to create data root: %v", err)
	}

	h := handlers.New(cfg.DataRoot, cfg.MaxUploadMB)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/files", h.Files)
	mux.HandleFunc("/api/dirs", h.Dirs)
	mux.HandleFunc("/api/rename", h.Rename)
	mux.HandleFunc("/api/preview", h.Preview)

	subFS, err := fs.Sub(publicFS, "public")
	if err != nil {
		log.Fatalf("Failed to create sub filesystem: %v", err)
	}
	mux.Handle("/", http.FileServer(http.FS(subFS)))

	addr := fmt.Sprintf("0.0.0.0:%d", cfg.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return srv, cfg
}

// Serve starts the HTTP server in a goroutine and blocks until stop is closed,
// then performs a graceful shutdown with a 10-second deadline.
func Serve(srv *http.Server, ready chan<- struct{}, stop <-chan struct{}) {
	go func() {
		if ready != nil {
			close(ready)
		}
		log.Printf("📁 FileHub running at http://0.0.0.0:%d/", mustPort(srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-stop

	log.Println("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
	log.Println("Server stopped.")
}

func mustPort(addr string) int {
	var p int
	fmt.Sscanf(addr, "0.0.0.0:%d", &p)
	return p
}
