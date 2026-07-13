package main

import (
	"encoding/json"
	"os"
)

// Config holds the application configuration.
type Config struct {
	Port        int    `json:"port"`
	DataRoot    string `json:"dataRoot"`
	MaxUploadMB int64  `json:"maxUploadMB"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		Port:        5000,
		DataRoot:    "./data",
		MaxUploadMB: 100,
	}
}

// LoadConfig reads config from file, or creates one with defaults if not found.
func LoadConfig(path string) (Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Write default config
			if writeErr := writeConfig(path, cfg); writeErr != nil {
				return cfg, writeErr
			}
			return cfg, nil
		}
		return cfg, err
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	// Apply minimums
	if cfg.Port <= 0 {
		cfg.Port = 5000
	}
	if cfg.DataRoot == "" {
		cfg.DataRoot = "./data"
	}
	if cfg.MaxUploadMB <= 0 {
		cfg.MaxUploadMB = 100
	}

	return cfg, nil
}

func writeConfig(path string, cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
