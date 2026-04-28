package watcher

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"context"
)

func TestIsSpecFile_True(t *testing.T) {
	for _, name := range []string{"SPECS.md", "NOTES.md", "BENCHMARKS.md", "TESTS.md", "CLAUDE.md"} {
		if !isSpecFile("/some/path/" + name) {
			t.Errorf("expected %s to be a spec file", name)
		}
	}
}

func TestIsSpecFile_False(t *testing.T) {
	for _, name := range []string{"README.md", "main.go", "go.mod"} {
		if isSpecFile("/some/path/" + name) {
			t.Errorf("expected %s NOT to be a spec file", name)
		}
	}
}

func TestWatcher_DetectsCreate(t *testing.T) {
	dir := t.TempDir()

	var mu sync.Mutex
	var calls []string
	handler := func(path string, removed bool) {
		mu.Lock()
		calls = append(calls, path)
		mu.Unlock()
	}

	w, err := New(handler, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer w.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := w.Watch(ctx, dir); err != nil {
		t.Fatalf("Watch: %v", err)
	}

	specPath := filepath.Join(dir, "SPECS.md")
	if err := os.WriteFile(specPath, []byte("content"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		n := len(calls)
		mu.Unlock()
		if n > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(calls) == 0 {
		t.Error("expected handler to be called for SPECS.md create")
	}
}

func TestWatcher_DetectsWrite(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "SPECS.md")
	if err := os.WriteFile(specPath, []byte("initial"), 0o644); err != nil {
		t.Fatalf("setup WriteFile: %v", err)
	}

	var mu sync.Mutex
	var calls []string
	handler := func(path string, removed bool) {
		mu.Lock()
		calls = append(calls, path)
		mu.Unlock()
	}

	w, err := New(handler, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer w.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := w.Watch(ctx, dir); err != nil {
		t.Fatalf("Watch: %v", err)
	}

	if err := os.WriteFile(specPath, []byte("updated"), 0o644); err != nil {
		t.Fatalf("update WriteFile: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		n := len(calls)
		mu.Unlock()
		if n > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(calls) == 0 {
		t.Error("expected handler called for SPECS.md write")
	}
}

func TestWatcher_DetectsRemove(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "SPECS.md")
	if err := os.WriteFile(specPath, []byte("content"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	var mu sync.Mutex
	type callInfo struct {
		path    string
		removed bool
	}
	var calls []callInfo
	handler := func(path string, removed bool) {
		mu.Lock()
		calls = append(calls, callInfo{path, removed})
		mu.Unlock()
	}

	w, err := New(handler, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer w.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := w.Watch(ctx, dir); err != nil {
		t.Fatalf("Watch: %v", err)
	}

	if err := os.Remove(specPath); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		n := len(calls)
		mu.Unlock()
		if n > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	mu.Lock()
	defer mu.Unlock()
	found := false
	for _, c := range calls {
		if c.removed {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected handler called with removed=true for SPECS.md remove")
	}
}

func TestWatcher_Debounce(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "SPECS.md")
	if err := os.WriteFile(specPath, []byte("initial"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	var mu sync.Mutex
	var calls int
	handler := func(path string, removed bool) {
		mu.Lock()
		calls++
		mu.Unlock()
	}

	debounce := 200 * time.Millisecond
	w, err := New(handler, debounce)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer w.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := w.Watch(ctx, dir); err != nil {
		t.Fatalf("Watch: %v", err)
	}

	// Write multiple times quickly
	for i := 0; i < 5; i++ {
		os.WriteFile(specPath, []byte("update"), 0o644)
		time.Sleep(10 * time.Millisecond)
	}

	// Wait for debounce to settle
	time.Sleep(debounce + 100*time.Millisecond)

	mu.Lock()
	n := calls
	mu.Unlock()

	if n > 2 {
		t.Errorf("expected debounce to coalesce events, got %d calls", n)
	}
}

func TestWatcher_IgnoresNonSpec(t *testing.T) {
	dir := t.TempDir()

	var mu sync.Mutex
	var calls []string
	handler := func(path string, removed bool) {
		mu.Lock()
		calls = append(calls, path)
		mu.Unlock()
	}

	w, err := New(handler, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer w.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := w.Watch(ctx, dir); err != nil {
		t.Fatalf("Watch: %v", err)
	}

	readmePath := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readmePath, []byte("readme"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Wait a bit to ensure no events come through
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(calls) != 0 {
		t.Errorf("expected handler NOT called for README.md, got %d calls: %v", len(calls), calls)
	}
}
