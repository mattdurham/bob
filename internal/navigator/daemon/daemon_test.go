package daemon

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mattdurham/bob/internal/navigator/indexer"
	"github.com/mattdurham/bob/internal/navigator/specstore"
)

// mockQueryableStore implements QueryableStore.
type mockQueryableStore struct {
	results []specstore.Result
	count   int
}

func (m *mockQueryableStore) VectorSearch(embedding []float32, limit int) ([]specstore.Result, error) {
	return m.results, nil
}

func (m *mockQueryableStore) Count() (int, error) {
	return m.count, nil
}

// mockEmbedder implements teiclient.Embedder-like interface.
type mockEmbedder struct{}

func (m *mockEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	result := make([][]float32, len(texts))
	for i := range result {
		result[i] = make([]float32, 768)
	}
	return result, nil
}

// mockIndexer wraps indexer.Indexer (nil-safe for tests).
type mockIndexerStore struct {
	upserted int
	hashes   map[string]string
}

func (m *mockIndexerStore) Upsert(sec *specstore.Section, embedding []float32) error {
	m.upserted++
	return nil
}
func (m *mockIndexerStore) DeleteByFilePrefix(p string) error { return nil }
func (m *mockIndexerStore) HashForFile(p string) (string, bool, error) {
	if m.hashes == nil {
		return "", false, nil
	}
	h, ok := m.hashes[p]
	return h, ok, nil
}
func (m *mockIndexerStore) Count() (int, error) { return m.upserted, nil }

func buildServer(t *testing.T, socketPath string) *Server {
	t.Helper()
	dir := t.TempDir()
	pidPath := filepath.Join(dir, "test.pid")
	store := &mockQueryableStore{count: 42}
	embedder := &mockEmbedder{}
	// Use a real but empty indexer (root won't be walked in tests)
	idxStore := &mockIndexerStore{hashes: make(map[string]string)}
	idx := indexer.New(idxStore, embedder, t.TempDir())
	return NewServer(socketPath, pidPath, idx, store, embedder)
}

func TestPaths_Deterministic(t *testing.T) {
	dir := "/tmp/test-daemon"
	sock1, pid1 := Paths(dir)
	sock2, pid2 := Paths(dir)
	if sock1 != sock2 || pid1 != pid2 {
		t.Errorf("Paths not deterministic: %q/%q vs %q/%q", sock1, pid1, sock2, pid2)
	}
}

func TestServer_PIDFile(t *testing.T) {
	dir := t.TempDir()
	socketPath := filepath.Join(dir, "test.sock")
	pidPath := filepath.Join(dir, "test.pid")
	store := &mockQueryableStore{count: 1}
	embedder := &mockEmbedder{}
	idxStore := &mockIndexerStore{}
	idx := indexer.New(idxStore, embedder, t.TempDir())
	srv := NewServer(socketPath, pidPath, idx, store, embedder)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start(ctx)
	}()

	// Wait for server to start
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(pidPath); err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if _, err := os.Stat(pidPath); err != nil {
		t.Error("expected PID file to exist after Start")
	}

	if err := srv.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	if _, err := os.Stat(pidPath); err == nil {
		t.Error("expected PID file to be removed after Stop")
	}
}

func TestServer_StatusEndpoint(t *testing.T) {
	dir := t.TempDir()
	socketPath := filepath.Join(dir, "test.sock")
	srv := buildServer(t, socketPath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go srv.Start(ctx)

	// Wait for socket
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(socketPath); err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	client := NewClient(socketPath)
	status, err := client.Status(context.Background())
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if _, ok := status["state"]; !ok {
		t.Errorf("expected 'state' key in status, got %v", status)
	}

	srv.Stop()
}

func TestServer_QueryEndpoint(t *testing.T) {
	dir := t.TempDir()
	socketPath := filepath.Join(dir, "test.sock")
	pidPath := filepath.Join(dir, "test.pid")
	store := &mockQueryableStore{
		results: []specstore.Result{
			{Path: "pkg/foo/SPECS.md#section-1", Text: "some content", Distance: 0.12},
		},
		count: 1,
	}
	embedder := &mockEmbedder{}
	idxStore := &mockIndexerStore{}
	idx := indexer.New(idxStore, embedder, t.TempDir())
	srv := NewServer(socketPath, pidPath, idx, store, embedder)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go srv.Start(ctx)

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(socketPath); err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	client := NewClient(socketPath)
	results, err := client.Query(context.Background(), "test query", 5)
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}

	srv.Stop()
}

func TestServer_IndexEndpoint(t *testing.T) {
	dir := t.TempDir()
	socketPath := filepath.Join(dir, "test.sock")
	srv := buildServer(t, socketPath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go srv.Start(ctx)

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(socketPath); err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Call /index via HTTP directly over the unix socket
	httpClient := NewClient(socketPath).httpClient
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "http://unix/index", nil)
	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("POST /index: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)
	if ok, _ := body["ok"].(bool); !ok {
		t.Errorf("expected {ok:true}, got %v", body)
	}

	srv.Stop()
}

func TestClient_Query(t *testing.T) {
	dir := t.TempDir()
	socketPath := filepath.Join(dir, "test.sock")
	pidPath := filepath.Join(dir, "test.pid")
	store := &mockQueryableStore{
		results: []specstore.Result{
			{Path: "x/SPECS.md#section-1", Text: "hello", Distance: 0.5},
		},
	}
	embedder := &mockEmbedder{}
	idxStore := &mockIndexerStore{}
	idx := indexer.New(idxStore, embedder, t.TempDir())
	srv := NewServer(socketPath, pidPath, idx, store, embedder)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go srv.Start(ctx)

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(socketPath); err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	c := NewClient(socketPath)
	results, err := c.Query(ctx, "hello world", 10)
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if results[0].Path != "x/SPECS.md#section-1" {
		t.Errorf("unexpected path: %s", results[0].Path)
	}

	srv.Stop()
}
