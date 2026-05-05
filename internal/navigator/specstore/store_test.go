package specstore

import (
	"testing"
)

func openMem(t *testing.T) *Store {
	t.Helper()
	s, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func sampleSection(path, text, hash string) *Section {
	return &Section{Path: path, Text: text, FileHash: hash}
}

func sampleVec(val float32) []float32 {
	v := make([]float32, 768)
	for i := range v {
		v[i] = val
	}
	return v
}

func TestOpen_CreatesSchema(t *testing.T) {
	s := openMem(t)
	// Count should be 0 for a new store, proving schema is applied
	n, err := s.Count()
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 rows, got %d", n)
	}
}

func TestUpsert_Insert(t *testing.T) {
	s := openMem(t)
	err := s.Upsert(sampleSection("pkg/foo/SPECS.md#section-1", "text content here", "hash1"), sampleVec(0.1))
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	n, _ := s.Count()
	if n != 1 {
		t.Errorf("expected count 1, got %d", n)
	}
}

func TestUpsert_Update(t *testing.T) {
	s := openMem(t)
	sec := sampleSection("pkg/foo/SPECS.md#section-1", "original text", "hash1")
	if err := s.Upsert(sec, sampleVec(0.1)); err != nil {
		t.Fatalf("first Upsert: %v", err)
	}
	sec.Text = "updated text content"
	sec.FileHash = "hash2"
	if err := s.Upsert(sec, sampleVec(0.2)); err != nil {
		t.Fatalf("second Upsert: %v", err)
	}
	n, _ := s.Count()
	if n != 1 {
		t.Errorf("expected count still 1 after update, got %d", n)
	}
}

func TestUpsert_EmbeddingStored(t *testing.T) {
	s := openMem(t)
	if err := s.Upsert(sampleSection("pkg/foo/SPECS.md#section-1", "some content", "hash1"), sampleVec(0.5)); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	results, err := s.VectorSearch(sampleVec(0.5), 5)
	if err != nil {
		t.Fatalf("VectorSearch: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Path != "pkg/foo/SPECS.md#section-1" {
		t.Errorf("unexpected path: %s", results[0].Path)
	}
}

func TestDeleteByFilePrefix(t *testing.T) {
	s := openMem(t)
	for i, hash := range []string{"h1", "h2", "h3"} {
		path := "pkg/foo/SPECS.md#section-" + string(rune('1'+i))
		if err := s.Upsert(sampleSection(path, "content here", hash), sampleVec(float32(i)/10)); err != nil {
			t.Fatalf("Upsert: %v", err)
		}
	}
	if err := s.DeleteByFilePrefix("pkg/foo/SPECS.md"); err != nil {
		t.Fatalf("DeleteByFilePrefix: %v", err)
	}
	n, _ := s.Count()
	if n != 0 {
		t.Errorf("expected 0 rows after delete, got %d", n)
	}
}

func TestDeleteByFilePrefix_OtherFiles(t *testing.T) {
	s := openMem(t)
	s.Upsert(sampleSection("pkg/foo/SPECS.md#section-1", "content here", "h1"), sampleVec(0.1))
	s.Upsert(sampleSection("pkg/bar/SPECS.md#section-1", "other content", "h2"), sampleVec(0.2))
	if err := s.DeleteByFilePrefix("pkg/foo/SPECS.md"); err != nil {
		t.Fatalf("DeleteByFilePrefix: %v", err)
	}
	n, _ := s.Count()
	if n != 1 {
		t.Errorf("expected 1 row (other file kept), got %d", n)
	}
}

func TestHashForFile_Missing(t *testing.T) {
	s := openMem(t)
	_, exists, err := s.HashForFile("pkg/foo/SPECS.md")
	if err != nil {
		t.Fatalf("HashForFile: %v", err)
	}
	if exists {
		t.Error("expected exists=false for unknown file")
	}
}

func TestHashForFile_Found(t *testing.T) {
	s := openMem(t)
	s.Upsert(sampleSection("pkg/foo/SPECS.md#section-1", "content", "abc123"), sampleVec(0.1))
	hash, exists, err := s.HashForFile("pkg/foo/SPECS.md")
	if err != nil {
		t.Fatalf("HashForFile: %v", err)
	}
	if !exists {
		t.Error("expected exists=true")
	}
	if hash != "abc123" {
		t.Errorf("expected hash 'abc123', got %q", hash)
	}
}

func TestVectorSearch_TopK(t *testing.T) {
	s := openMem(t)
	for i := 0; i < 5; i++ {
		path := "pkg/foo/SPECS.md#section-" + string(rune('1'+i))
		s.Upsert(sampleSection(path, "content", "hash"), sampleVec(float32(i+1)/10))
	}
	results, err := s.VectorSearch(sampleVec(0.5), 3)
	if err != nil {
		t.Fatalf("VectorSearch: %v", err)
	}
	if len(results) > 3 {
		t.Errorf("expected at most 3 results, got %d", len(results))
	}
}

func TestVectorSearch_Empty(t *testing.T) {
	s := openMem(t)
	results, err := s.VectorSearch(sampleVec(0.5), 5)
	if err != nil {
		t.Fatalf("VectorSearch on empty store: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty results, got %d", len(results))
	}
}

func TestCount(t *testing.T) {
	s := openMem(t)
	for i := 0; i < 3; i++ {
		path := "pkg/foo/SPECS.md#section-" + string(rune('1'+i))
		s.Upsert(sampleSection(path, "content here long enough", "hash"), sampleVec(float32(i)/10))
	}
	n, err := s.Count()
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if n != 3 {
		t.Errorf("expected 3, got %d", n)
	}
}

func TestOpen_FileDB(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.sqlite"
	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open file: %v", err)
	}
	defer s.Close()
	n, err := s.Count()
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0, got %d", n)
	}
}

func TestVectorSearch_OrderedByDistance(t *testing.T) {
	s := openMem(t)
	// Insert sections with distinct embedding values
	for i, val := range []float32{0.1, 0.5, 0.9} {
		path := "pkg/SPECS.md#s" + string(rune('1'+i))
		s.Upsert(sampleSection(path, "content here", "hash"), sampleVec(val))
	}
	// Query near 0.5 — should come back ordered by distance
	results, err := s.VectorSearch(sampleVec(0.5), 3)
	if err != nil {
		t.Fatalf("VectorSearch: %v", err)
	}
	for i := 1; i < len(results); i++ {
		if results[i].Distance < results[i-1].Distance {
			t.Errorf("results not in ascending distance order at index %d", i)
		}
	}
}
