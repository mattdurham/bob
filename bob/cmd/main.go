package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"charm.land/fantasy"
	fantasyanthropicprovider "charm.land/fantasy/providers/anthropic"
	fantasygoogleprovider "charm.land/fantasy/providers/google"
	fantasyopenapiprovider "charm.land/fantasy/providers/openai"
	"github.com/mattdurham/bob/bob/extension"
	"github.com/mattdurham/bob/bob/harness"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "bob: "+err.Error())
		os.Exit(1)
	}

	ctx := context.Background()

	// Build fantasy provider based on configured provider name.
	var fantasyProv fantasy.Provider
	var provErr error

	switch cfg.Provider {
	case "anthropic":
		fantasyProv, provErr = fantasyanthropicprovider.New(
			fantasyanthropicprovider.WithAPIKey(cfg.AnthropicAPIKey),
		)
	case "openai":
		fantasyProv, provErr = fantasyopenapiprovider.New(
			fantasyopenapiprovider.WithAPIKey(cfg.OpenAIAPIKey),
		)
	case "gemini":
		fantasyProv, provErr = fantasygoogleprovider.New(
			fantasygoogleprovider.WithGeminiAPIKey(cfg.GeminiAPIKey),
		)
	default:
		fmt.Fprintf(os.Stderr, "bob: unknown provider %q\n", cfg.Provider)
		os.Exit(1)
	}

	if provErr != nil {
		fmt.Fprintf(os.Stderr, "bob: create provider: %v\n", provErr)
		os.Exit(1)
	}

	langModel, provErr := fantasyProv.LanguageModel(ctx, cfg.Model)
	if provErr != nil {
		fmt.Fprintf(os.Stderr, "bob: get language model %q from provider %q: %v\n", cfg.Model, cfg.Provider, provErr)
		os.Exit(1)
	}

	// Build extension host with a simple stderr logger.
	h := extension.NewHost(func(level int, msg string) {
		levelStr := [...]string{"DEBUG", "INFO", "WARN", "ERROR"}
		if level >= 0 && level < len(levelStr) {
			log.Printf("[%s] %s", levelStr[level], msg)
		} else {
			log.Printf("[?] %s", msg)
		}
	})

	// Load .wasm extensions from the configured directory.
	var extPaths []string
	if cfg.ExtensionsDir != "" {
		entries, readErr := os.ReadDir(cfg.ExtensionsDir)
		if readErr != nil {
			fmt.Fprintf(os.Stderr, "bob: extensions dir %q not found, skipping\n", cfg.ExtensionsDir)
		} else {
			for _, e := range entries {
				if e.IsDir() || filepath.Ext(e.Name()) != ".wasm" {
					continue
				}
				path := filepath.Join(cfg.ExtensionsDir, e.Name())
				if loadErr := h.Load(ctx, path); loadErr != nil {
					fmt.Fprintf(os.Stderr, "bob: load extension %q: %v\n", e.Name(), loadErr)
					continue
				}
				extPaths = append(extPaths, path)
			}
		}
	}

	defer func() {
		if err := h.Close(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "bob: close extension host: %v\n", err)
		}
	}()

	m := harness.New(langModel, cfg.Provider, h)
	m.SetExtensionPaths(extPaths)

	prog := tea.NewProgram(&m)
	m.SetProgram(prog)

	if _, err := prog.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "bob: "+err.Error())
		os.Exit(1)
	}
}
