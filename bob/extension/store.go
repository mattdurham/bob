// Package extension implements a wazero-based WASM extension host.
package extension

import "sync"

// Store is a thread-safe, in-process key-value store for extension state.
type Store struct {
	mu   sync.RWMutex
	data map[string]string
}

// NewStore returns an initialised Store.
func NewStore() *Store {
	return &Store{data: make(map[string]string)}
}

// Set stores value under key.
func (s *Store) Set(k, v string) {
	s.mu.Lock()
	s.data[k] = v
	s.mu.Unlock()
}

// Get retrieves the value stored under key. Returns ("", false) if absent.
func (s *Store) Get(k string) (string, bool) {
	s.mu.RLock()
	v, ok := s.data[k]
	s.mu.RUnlock()
	return v, ok
}
