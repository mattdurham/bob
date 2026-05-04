package watcher

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Handler is called when a spec file changes.
// path is the absolute file path. removed is true if the file was deleted.
type Handler func(path string, removed bool)

var specFilenames = map[string]bool{
	"SPECS.md":      true,
	"NOTES.md":      true,
	"BENCHMARKS.md": true,
	"TESTS.md":      true,
	"CLAUDE.md":     true,
}

func isSpecFile(path string) bool {
	return specFilenames[filepath.Base(path)]
}

// Watcher watches a directory tree for spec file changes.
type Watcher struct {
	fw       *fsnotify.Watcher
	handler  Handler
	debounce time.Duration

	// mu protects timerMap. It is also held by the run goroutine's fire
	// closure; Close acquires it to stop pending timers before returning.
	mu       sync.Mutex
	timerMap map[string]*time.Timer
}

// New creates a new Watcher.
func New(handler Handler, debounce time.Duration) (*Watcher, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &Watcher{
		fw:       fw,
		handler:  handler,
		debounce: debounce,
		timerMap: make(map[string]*time.Timer),
	}, nil
}

// Watch adds root and all subdirectories to the watcher and starts processing events.
func (w *Watcher) Watch(ctx context.Context, root string) error {
	if err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return w.fw.Add(path)
		}
		return nil
	}); err != nil {
		return err
	}

	go w.run(ctx)
	return nil
}

func (w *Watcher) run(ctx context.Context) {
	fire := func(path string, removed bool) {
		w.mu.Lock()
		defer w.mu.Unlock()
		if t, ok := w.timerMap[path]; ok {
			// If Stop returns false the AfterFunc goroutine is already running
			// (or has run). That goroutine will call delete(w.timerMap, path)
			// after the handler, which would delete the NEW timer we are about
			// to set. Remove the stale entry now so the old goroutine's delete
			// is a no-op on the new timer.
			if !t.Stop() {
				delete(w.timerMap, path)
			}
		}
		w.timerMap[path] = time.AfterFunc(w.debounce, func() {
			w.handler(path, removed)
			w.mu.Lock()
			delete(w.timerMap, path)
			w.mu.Unlock()
		})
	}

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-w.fw.Events:
			if !ok {
				return
			}
			if !isSpecFile(event.Name) {
				// Still watch newly created directories
				if event.Has(fsnotify.Create) {
					if fi, err := os.Stat(event.Name); err == nil && fi.IsDir() {
						w.fw.Add(event.Name)
					}
				}
				continue
			}
			removed := event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename)
			fire(event.Name, removed)
		case err, ok := <-w.fw.Errors:
			if !ok {
				return
			}
			log.Printf("navigator: watcher error: %v", err)
		}
	}
}

// Close stops the watcher and cancels any pending debounce timers so that the
// handler is not invoked after Close returns.
func (w *Watcher) Close() error {
	err := w.fw.Close()
	w.mu.Lock()
	for path, t := range w.timerMap {
		t.Stop()
		delete(w.timerMap, path)
	}
	w.mu.Unlock()
	return err
}
