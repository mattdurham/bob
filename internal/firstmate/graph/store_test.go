package graph_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mattdurham/bob/internal/firstmate/graph"
)

func tempStore(t *testing.T) *graph.Store {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	store, err := graph.Open(path)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func TestStoreSaveAndLoadNode(t *testing.T) {
	store := tempStore(t)

	n := &graph.Node{
		ID:          "pkg.Foo",
		Kind:        "function",
		Name:        "Foo",
		File:        "foo.go",
		Line:        10,
		Cyclomatic:  3,
		Cognitive:   2,
		CalleeIDs:   []string{"pkg.Bar"},
		CallerIDs:   []string{},
		ChildIDs:    []string{},
		LintCount:   1,
		LintIssues:  "some issue",
	}
	if err := store.SaveNode(n); err != nil {
		t.Fatalf("save node: %v", err)
	}

	got, err := store.GetNodeByID("pkg.Foo")
	if err != nil {
		t.Fatalf("get node: %v", err)
	}
	if got == nil {
		t.Fatal("expected node to be found")
	}
	if got.Name != "Foo" {
		t.Errorf("got name %q, want Foo", got.Name)
	}
	if got.Cyclomatic != 3 {
		t.Errorf("got cyclomatic %d, want 3", got.Cyclomatic)
	}
	if len(got.CalleeIDs) != 1 || got.CalleeIDs[0] != "pkg.Bar" {
		t.Errorf("got calleeIDs %v, want [pkg.Bar]", got.CalleeIDs)
	}
}

func TestStoreSaveAndLoadEdge(t *testing.T) {
	store := tempStore(t)

	e := &graph.Edge{From: "a", To: "b", Kind: "call"}
	if err := store.SaveEdge(e); err != nil {
		t.Fatalf("save edge: %v", err)
	}
	// Duplicate should be ignored (INSERT OR IGNORE)
	if err := store.SaveEdge(e); err != nil {
		t.Fatalf("save duplicate edge: %v", err)
	}

	edges, err := store.AllEdges()
	if err != nil {
		t.Fatalf("all edges: %v", err)
	}
	if len(edges) != 1 {
		t.Errorf("got %d edges, want 1", len(edges))
	}
}

func TestStoreReset(t *testing.T) {
	store := tempStore(t)

	store.SaveNode(&graph.Node{ID: "n1", Kind: "function", Name: "f", CalleeIDs: []string{}, CallerIDs: []string{}, ChildIDs: []string{}})
	store.SaveEdge(&graph.Edge{From: "n1", To: "n2", Kind: "call"})

	if err := store.Reset(); err != nil {
		t.Fatalf("reset: %v", err)
	}

	nodes, _ := store.AllNodes()
	if len(nodes) != 0 {
		t.Errorf("expected 0 nodes after reset, got %d", len(nodes))
	}
}

func TestStoreGetNodeByIDMissing(t *testing.T) {
	store := tempStore(t)
	n, err := store.GetNodeByID("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != nil {
		t.Error("expected nil for missing node")
	}
}

func TestDBPath(t *testing.T) {
	path, err := graph.DBPath()
	if err != nil {
		t.Fatalf("DBPath: %v", err)
	}
	if path == "" {
		t.Error("expected non-empty path")
	}
	// Verify the directory exists
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("directory %s does not exist: %v", dir, err)
	}
}

func TestStoreSnapshot(t *testing.T) {
	store := tempStore(t)

	nodes := []*graph.Node{
		{ID: "a", Kind: "function", Name: "a", CalleeIDs: []string{}, CallerIDs: []string{}, ChildIDs: []string{}},
	}
	edges := []*graph.Edge{
		{From: "a", To: "b", Kind: "call"},
	}

	if err := store.SaveSnapshot("test-snap", nodes, edges); err != nil {
		t.Fatalf("save snapshot: %v", err)
	}

	snap, err := store.LoadLatestSnapshot()
	if err != nil {
		t.Fatalf("load snapshot: %v", err)
	}
	if snap.Name != "test-snap" {
		t.Errorf("got snap name %q, want test-snap", snap.Name)
	}
	if len(snap.Nodes) != 1 {
		t.Errorf("got %d nodes in snap, want 1", len(snap.Nodes))
	}
}
