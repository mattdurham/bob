package extension

// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

import (
	"context"
	"fmt"

	"github.com/tetratelabs/wazero/api"
)

// requiredExports lists the WASM exports that every extension must provide.
var requiredExports = []string{"_init", "_on_event", "_alloc", "_free"}

// validateExports returns an error if m is missing any required export.
func validateExports(m api.Module) error {
	for _, name := range requiredExports {
		if fn := m.ExportedFunction(name); fn == nil {
			return fmt.Errorf("extension missing required export: %s", name)
		}
	}
	return nil
}

// callInit calls the extension's _init() export.
// For native Go WASM modules, _initialize is called first to set up the Go runtime.
// Returns an error if _init is absent or returns a non-zero status code.
func callInit(ctx context.Context, m api.Module) error {
	// Native Go WASM (GOOS=wasip1) exports _initialize to bootstrap the runtime.
	// Call it before _init so Go globals and the scheduler are ready.
	if fn := m.ExportedFunction("_initialize"); fn != nil {
		if _, err := fn.Call(ctx); err != nil {
			return fmt.Errorf("_initialize trap: %w", err)
		}
	}
	fn := m.ExportedFunction("_init")
	if fn == nil {
		return fmt.Errorf("extension missing _init export")
	}
	results, err := fn.Call(ctx)
	if err != nil {
		return fmt.Errorf("_init trap: %w", err)
	}
	if len(results) > 0 && results[0] != 0 {
		return fmt.Errorf("_init returned error code %d", results[0])
	}
	return nil
}
