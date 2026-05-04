package indexer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mattdurham/bob/internal/navigator/specstore"
)

// mockEmbedder records calls and returns fixed embeddings.
type mockEmbedder struct {
	calls  int
	result [][]float32
	err    error
}

func (m *mockEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	m.calls++
	if m.err != nil {
		return nil, m.err
	}
	if m.result != nil {
		return m.result, nil
	}
	// Return one embedding per text
	result := make([][]float32, len(texts))
	for i := range result {
		v := make([]float32, 768)
		for j := range v {
			v[j] = float32(i+1) / 100
		}
		result[i] = v
	}
	return result, nil
}

// mockStore records calls.
type mockStore struct {
	upserted  int
	deleted   []string
	hashes    map[string]string // path prefix → hash
	upsertErr error
}

func newMockStore() *mockStore {
	return &mockStore{hashes: make(map[string]string)}
}

func (m *mockStore) Upsert(sec *specstore.Section, embedding []float32) error {
	if m.upsertErr != nil {
		return m.upsertErr
	}
	m.upserted++
	return nil
}

func (m *mockStore) DeleteByFilePrefix(filePrefix string) error {
	m.deleted = append(m.deleted, filePrefix)
	return nil
}

func (m *mockStore) HashForFile(filePath string) (string, bool, error) {
	hash, ok := m.hashes[filePath]
	return hash, ok, nil
}

func (m *mockStore) Count() (int, error) {
	return m.upserted, nil
}

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeFile %s: %v", name, err)
	}
	return path
}

const specsContent = `## 1. Rate Limiting
This invariant ensures that requests are throttled at the configured rate.
The system must reject all requests above the threshold.

## 2. Idempotency
All write operations must be idempotent when given the same input.
`

func TestIndexAll_NewFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "SPECS.md", specsContent)

	store := newMockStore()
	embedder := &mockEmbedder{}
	idx := New(store, embedder, dir)

	if err := idx.IndexAll(context.Background()); err != nil {
		t.Fatalf("IndexAll: %v", err)
	}
	if embedder.calls == 0 {
		t.Error("expected Embed to be called for new file")
	}
	if store.upserted == 0 {
		t.Error("expected Upsert to be called")
	}
}

func TestIndexAll_UnchangedFile(t *testing.T) {
	dir := t.TempDir()
	absPath := writeFile(t, dir, "SPECS.md", specsContent)

	// Pre-compute the hash of this file
	hash, err := hashFile(absPath)
	if err != nil {
		t.Fatalf("hashFile: %v", err)
	}

	store := newMockStore()
	// Store already has the same hash for this relative path
	relPath := "SPECS.md"
	store.hashes[relPath] = hash

	embedder := &mockEmbedder{}
	idx := New(store, embedder, dir)

	if err := idx.IndexAll(context.Background()); err != nil {
		t.Fatalf("IndexAll: %v", err)
	}
	if embedder.calls > 0 {
		t.Error("expected Embed NOT called when hash unchanged")
	}
}

func TestIndexAll_ChangedFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "SPECS.md", specsContent)

	store := newMockStore()
	store.hashes["SPECS.md"] = "old-different-hash"

	embedder := &mockEmbedder{}
	idx := New(store, embedder, dir)

	if err := idx.IndexAll(context.Background()); err != nil {
		t.Fatalf("IndexAll: %v", err)
	}
	if embedder.calls == 0 {
		t.Error("expected Embed called when hash changed")
	}
}

func TestIndexAll_MultipleFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "SPECS.md", specsContent)
	writeFile(t, dir, "NOTES.md", specsContent)
	writeFile(t, dir, "CLAUDE.md", "1. First invariant text that is long enough to pass filter\n2. Second invariant text long enough to pass the filter too\n")

	store := newMockStore()
	embedder := &mockEmbedder{}
	idx := New(store, embedder, dir)

	if err := idx.IndexAll(context.Background()); err != nil {
		t.Fatalf("IndexAll: %v", err)
	}
	if embedder.calls < 2 {
		t.Errorf("expected at least 2 Embed calls for 3 files, got %d", embedder.calls)
	}
}

func TestIndexFile_IndexesNewFile(t *testing.T) {
	dir := t.TempDir()
	absPath := writeFile(t, dir, "SPECS.md", specsContent)

	store := newMockStore()
	embedder := &mockEmbedder{}
	idx := New(store, embedder, dir)

	if err := idx.IndexFile(context.Background(), absPath, false); err != nil {
		t.Fatalf("IndexFile: %v", err)
	}
	if embedder.calls != 1 {
		t.Errorf("expected 1 Embed call, got %d", embedder.calls)
	}
	if store.upserted == 0 {
		t.Error("expected Upsert called for each section")
	}
}

func TestIndexFile_Removed(t *testing.T) {
	dir := t.TempDir()
	absPath := filepath.Join(dir, "SPECS.md")

	store := newMockStore()
	embedder := &mockEmbedder{}
	idx := New(store, embedder, dir)

	if err := idx.IndexFile(context.Background(), absPath, true); err != nil {
		t.Fatalf("IndexFile removed: %v", err)
	}
	if embedder.calls > 0 {
		t.Error("expected Embed NOT called for removed file")
	}
	if len(store.deleted) == 0 {
		t.Error("expected DeleteByFilePrefix called for removed file")
	}
}

func TestIndexFile_ShortSections_Skipped(t *testing.T) {
	dir := t.TempDir()
	// Content that produces no valid sections (too short)
	absPath := writeFile(t, dir, "SPECS.md", "## Hi\nok")

	store := newMockStore()
	embedder := &mockEmbedder{}
	idx := New(store, embedder, dir)

	if err := idx.IndexFile(context.Background(), absPath, false); err != nil {
		t.Fatalf("IndexFile: %v", err)
	}
	if embedder.calls > 0 {
		t.Error("expected Embed NOT called when no valid sections")
	}
}

func TestHashFile_Consistent(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "SPECS.md", specsContent)

	h1, err := hashFile(path)
	if err != nil {
		t.Fatalf("hashFile: %v", err)
	}
	h2, err := hashFile(path)
	if err != nil {
		t.Fatalf("hashFile second: %v", err)
	}
	if h1 != h2 {
		t.Errorf("hashFile not deterministic: %q vs %q", h1, h2)
	}
}
