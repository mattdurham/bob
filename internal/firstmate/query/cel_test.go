package query_test

import (
	"testing"

	"github.com/mattdurham/bob/internal/firstmate/graph"
	"github.com/mattdurham/bob/internal/firstmate/query"
)

func makeNode(id, kind string, cyclo int) *graph.Node {
	return &graph.Node{
		ID:         id,
		Kind:       kind,
		Name:       id,
		Cyclomatic: cyclo,
		CalleeIDs:  []string{},
		CallerIDs:  []string{},
		ChildIDs:   []string{},
	}
}

func TestQueryNodes_FilterByKind(t *testing.T) {
	eng, err := query.NewEngine()
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	nodes := []*graph.Node{
		makeNode("pkg.Foo", "function", 5),
		makeNode("pkg.Bar", "type", 0),
		makeNode("pkg.Baz", "function", 15),
	}

	got, err := eng.QueryNodes(nodes, `kind == "function"`, "", 0)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("got %d nodes, want 2", len(got))
	}
}

func TestQueryNodes_FilterByCyclomatic(t *testing.T) {
	eng, err := query.NewEngine()
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	nodes := []*graph.Node{
		makeNode("pkg.Simple", "function", 3),
		makeNode("pkg.Complex", "function", 15),
		makeNode("pkg.VeryComplex", "function", 25),
	}

	got, err := eng.QueryNodes(nodes, `cyclomatic > 10`, "", 0)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("got %d nodes, want 2", len(got))
	}
}

func TestQueryNodes_TopN(t *testing.T) {
	eng, err := query.NewEngine()
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	nodes := []*graph.Node{
		makeNode("a", "function", 1),
		makeNode("b", "function", 2),
		makeNode("c", "function", 3),
	}

	got, err := eng.QueryNodes(nodes, `kind == "function"`, "cyclomatic", 2)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("got %d nodes, want 2 (top_n=2)", len(got))
	}
}

func TestQueryNodes_InvalidExpr(t *testing.T) {
	eng, err := query.NewEngine()
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	_, err = eng.QueryNodes(nil, `this is not valid cel`, "", 0)
	if err == nil {
		t.Error("expected error for invalid CEL expression")
	}
}

func TestQueryEdges_FilterByKind(t *testing.T) {
	eng, err := query.NewEdgeEngine()
	if err != nil {
		t.Fatalf("new edge engine: %v", err)
	}

	edges := []*graph.Edge{
		{From: "a", To: "b", Kind: "call"},
		{From: "c", To: "d", Kind: "implements"},
		{From: "e", To: "f", Kind: "call"},
	}

	got, err := eng.QueryEdges(edges, `kind == "call"`)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("got %d edges, want 2", len(got))
	}
}

func TestHelp(t *testing.T) {
	h := query.Help()
	if h == "" {
		t.Error("expected non-empty help text")
	}
	if len(h) < 100 {
		t.Errorf("help text too short: %d chars", len(h))
	}
}

func TestExamples(t *testing.T) {
	e := query.Examples()
	if e == "" {
		t.Error("expected non-empty examples text")
	}
}
