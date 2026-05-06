package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"github.com/mattdurham/bob/bob/extension"
	"github.com/mattdurham/bob/bob/harness"
	anthropicprovider "github.com/mattdurham/bob/bob/provider/anthropic"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "bob: "+err.Error())
		os.Exit(1)
	}

	// Build provider.
	p := anthropicprovider.New(cfg.APIKey)

	// Build extension host with a simple stderr logger.
	h := extension.NewHost(func(level int, msg string) {
		levelStr := [...]string{"DEBUG", "INFO", "WARN", "ERROR"}
		if level >= 0 && level < len(levelStr) {
			log.Printf("[%s] %s", levelStr[level], msg)
		} else {
			log.Printf("[?] %s", msg)
		}
	})

	ctx := context.Background()

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

	// Ensure the active model is set.
	m := harness.New(p, h)
	m.SetExtensionPaths(extPaths)
	if cfg.Model != "" {
		m.SetActiveModel(cfg.Model)
	}

	prog := tea.NewProgram(&m)
	m.SetProgram(prog)

	if _, err := prog.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "bob: "+err.Error())
		h.Close(ctx) //nolint:errcheck
		os.Exit(1)
	}

	if err := h.Close(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "bob: close extension host: %v\n", err)
	}
}
