package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/mattdurham/bob/internal/navigator/indexer"
	"github.com/mattdurham/bob/internal/navigator/specstore"
)

// Embedder is the subset of teiclient.Client needed by the daemon.
type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
}

// QueryableStore is the subset of specstore.Store needed for queries.
type QueryableStore interface {
	VectorSearch(embedding []float32, limit int) ([]specstore.Result, error)
	Count() (int, error)
}

// Server runs an HTTP server on a Unix domain socket.
type Server struct {
	socketPath string
	pidPath    string
	indexer    *indexer.Indexer
	store      QueryableStore
	embedder   Embedder
	srv        *http.Server
	listener   net.Listener

	// ctx/cancel and wg manage the lifecycle of background goroutines started
	// by handleIndex. Stop() cancels the context and waits for all goroutines
	// to finish before closing the listener, preventing use-after-close on the
	// store.
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// Paths returns the socket and PID file paths given a base directory.
func Paths(dir string) (socketPath, pidPath string) {
	return dir + "/navigator.sock", dir + "/navigator.pid"
}

// NewServer creates a new daemon server.
func NewServer(socketPath, pidPath string, idx *indexer.Indexer, store QueryableStore, embedder Embedder) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Server{
		socketPath: socketPath,
		pidPath:    pidPath,
		indexer:    idx,
		store:      store,
		embedder:   embedder,
		ctx:        ctx,
		cancel:     cancel,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/query", s.handleQuery)
	mux.HandleFunc("/index", s.handleIndex)
	mux.HandleFunc("/status", s.handleStatus)
	s.srv = &http.Server{Handler: mux}

	return s
}

// Start writes the PID file and begins listening on the Unix socket.
func (s *Server) Start(ctx context.Context) error {
	// Remove stale socket if it exists
	os.Remove(s.socketPath)

	ln, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return fmt.Errorf("daemon: listen: %w", err)
	}
	s.listener = ln

	pid := os.Getpid()
	if err := os.WriteFile(s.pidPath, []byte(strconv.Itoa(pid)), 0o644); err != nil {
		ln.Close()
		return fmt.Errorf("daemon: write pid: %w", err)
	}

	go func() {
		<-ctx.Done()
		s.Stop()
	}()

	if err := s.srv.Serve(ln); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Stop cancels background goroutines, waits for them to finish, then
// gracefully shuts down the HTTP server and removes the PID and socket files.
func (s *Server) Stop() error {
	// Signal all handleIndex goroutines to stop.
	s.cancel()
	// Wait for goroutines to finish before closing the store / listener.
	s.wg.Wait()

	os.Remove(s.pidPath)
	os.Remove(s.socketPath)

	// Graceful shutdown with a 5-second drain for in-flight connections.
	shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutCancel()
	return s.srv.Shutdown(shutCtx)
}

func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}

	embeddings, err := s.embedder.Embed(r.Context(), []string{req.Query})
	if err != nil {
		http.Error(w, fmt.Sprintf("embed: %v", err), http.StatusInternalServerError)
		return
	}
	if len(embeddings) == 0 {
		http.Error(w, "empty embedding result", http.StatusInternalServerError)
		return
	}

	results, err := s.store.VectorSearch(embeddings[0], req.Limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("search: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"results": results})
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.indexer.IndexAll(s.ctx); err != nil {
			log.Printf("navigator: background index failed: %v", err)
		}
	}()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	count, err := s.store.Count()
	if err != nil {
		http.Error(w, fmt.Sprintf("count: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"state": "ready",
		"count": count,
	})
}
