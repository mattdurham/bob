package main

import (
	"fmt"
	"os"
	"strings"
)

// Config holds runtime configuration for the bob binary.
type Config struct {
	// APIKey is the Anthropic API key (ANTHROPIC_API_KEY).
	APIKey string

	// ExtensionsDir is the directory scanned for .wasm extension files (BOB_EXTENSIONS_DIR).
	ExtensionsDir string

	// Model is the default LLM model to use (BOB_MODEL, default: claude-sonnet-4-5).
	Model string

	// Provider is the LLM provider name (BOB_PROVIDER, default: anthropic).
	Provider string
}

// LoadConfig reads configuration from environment variables.
//
// Variable precedence (highest to lowest):
//  1. Environment variables
//
// Defaults:
//   - BOB_MODEL: claude-sonnet-4-5
//   - BOB_PROVIDER: anthropic
//
// Returns an error if the provider is "anthropic" and ANTHROPIC_API_KEY is empty.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		APIKey:        os.Getenv("ANTHROPIC_API_KEY"),
		ExtensionsDir: expandTilde(os.Getenv("BOB_EXTENSIONS_DIR")),
		Model:         os.Getenv("BOB_MODEL"),
		Provider:      os.Getenv("BOB_PROVIDER"),
	}

	if cfg.Model == "" {
		cfg.Model = "claude-sonnet-4-5"
	}
	if cfg.Provider == "" {
		cfg.Provider = "anthropic"
	}

	if cfg.Provider == "anthropic" && cfg.APIKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY is required when BOB_PROVIDER=anthropic")
	}

	return cfg, nil
}

// expandTilde replaces a leading "~" with the user's home directory.
// Returns the path unchanged if it doesn't start with "~" or if the home
// directory cannot be determined.
func expandTilde(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return home + path[1:]
}
