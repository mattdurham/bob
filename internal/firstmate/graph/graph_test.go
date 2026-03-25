package graph_test

import (
	"testing"

	"github.com/mattdurham/bob/internal/firstmate/graph"
)

func TestGraphAddAndGetNode(t *testing.T) {
	g := graph.New()
	n := &graph.Node{
		ID:        "pkg.MyFunc",
		Kind:      "function",
		Name:      "MyFunc",
		CalleeIDs: []string{},
		CallerIDs: []string{},
		ChildIDs:  []string{},
	}
	g.AddNode(n)

	got, ok := g.GetNode("pkg.MyFunc")
	if !ok {
		t.Fatal("expected node to be found")
	}
	if got.Name != "MyFunc" {
		t.Errorf("got name %q, want %q", got.Name, "MyFunc")
	}
}

func TestGraphMissingNode(t *testing.T) {
	g := graph.New()
	_, ok := g.GetNode("nonexistent")
	if ok {
		t.Error("expected node not to be found")
	}
}

func TestGraphAddEdge(t *testing.T) {
	g := graph.New()
	e := &graph.Edge{From: "a", To: "b", Kind: "call"}
	g.AddEdge(e)
	g.AddEdge(e) // duplicate should be ignored

	edges := g.Edges()
	if len(edges) != 1 {
		t.Errorf("got %d edges, want 1", len(edges))
	}
}

func TestGraphReset(t *testing.T) {
	g := graph.New()
	g.AddNode(&graph.Node{ID: "n1", Kind: "function", CalleeIDs: []string{}, CallerIDs: []string{}, ChildIDs: []string{}})
	g.AddEdge(&graph.Edge{From: "n1", To: "n2", Kind: "call"})
	g.Reset()

	if len(g.Nodes()) != 0 {
		t.Error("expected no nodes after reset")
	}
	if len(g.Edges()) != 0 {
		t.Error("expected no edges after reset")
	}
}

func TestNodeAddCallee(t *testing.T) {
	n := &graph.Node{ID: "a", CalleeIDs: []string{}}
	n.AddCallee("b")
	n.AddCallee("b") // duplicate
	n.AddCallee("c")

	if len(n.CalleeIDs) != 2 {
		t.Errorf("got %d callees, want 2", len(n.CalleeIDs))
	}
}

func TestNodeToMap(t *testing.T) {
	n := &graph.Node{
		ID:         "pkg.Foo",
		Kind:       "function",
		Name:       "Foo",
		Cyclomatic: 5,
		Coverage:   0.75,
		CalleeIDs:  []string{"pkg.Bar"},
		CallerIDs:  []string{},
		ChildIDs:   []string{},
	}
	m := n.ToMap()

	if m["id"] != "pkg.Foo" {
		t.Errorf("id mismatch: %v", m["id"])
	}
	if m["cyclomatic"].(int64) != 5 {
		t.Errorf("cyclomatic mismatch: %v", m["cyclomatic"])
	}
	if m["coverage"].(float64) != 0.75 {
		t.Errorf("coverage mismatch: %v", m["coverage"])
	}
}
