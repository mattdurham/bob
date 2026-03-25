package query

import (
	"fmt"
	"sort"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/mattdurham/bob/internal/firstmate/graph"
)

// Engine evaluates CEL expressions over graph nodes and edges.
type Engine struct {
	env *cel.Env
}

// NewEngine creates a new CEL query engine.
func NewEngine() (*Engine, error) {
	env, err := cel.NewEnv(
		cel.Variable("id", cel.StringType),
		cel.Variable("kind", cel.StringType),
		cel.Variable("name", cel.StringType),
		cel.Variable("file", cel.StringType),
		cel.Variable("line", cel.IntType),
		cel.Variable("text", cel.StringType),
		cel.Variable("external", cel.BoolType),
		cel.Variable("cyclomatic", cel.IntType),
		cel.Variable("cognitive", cel.IntType),
		cel.Variable("receiver", cel.StringType),
		cel.Variable("params", cel.StringType),
		cel.Variable("returns", cel.StringType),
		cel.Variable("parent_id", cel.StringType),
		cel.Variable("callee_ids", cel.StringType),
		cel.Variable("caller_ids", cel.StringType),
		cel.Variable("child_ids", cel.StringType),
		cel.Variable("lint_count", cel.IntType),
		cel.Variable("lint_issues", cel.StringType),
		cel.Variable("race_count", cel.IntType),
		cel.Variable("race_issues", cel.StringType),
		cel.Variable("heap_allocs", cel.IntType),
		cel.Variable("stack_allocs", cel.IntType),
		cel.Variable("coverage", cel.DoubleType),
		cel.Variable("test_status", cel.StringType),
		cel.Variable("bench_ns_op", cel.DoubleType),
		cel.Variable("bench_b_op", cel.DoubleType),
		cel.Variable("bench_allocs_op", cel.DoubleType),
		cel.Variable("pprof_flat_pct", cel.DoubleType),
		cel.Variable("pprof_cum_pct", cel.DoubleType),
		cel.Variable("vet_issues", cel.StringType),
	)
	if err != nil {
		return nil, fmt.Errorf("create CEL env: %w", err)
	}
	return &Engine{env: env}, nil
}

// QueryNodes evaluates a CEL expression against all nodes, returning matching nodes
// optionally sorted and limited.
func (e *Engine) QueryNodes(nodes []*graph.Node, expr, sortBy string, topN int) ([]*graph.Node, error) {
	ast, iss := e.env.Compile(expr)
	if iss != nil && iss.Err() != nil {
		return nil, fmt.Errorf("compile CEL expression: %w", iss.Err())
	}
	prg, err := e.env.Program(ast)
	if err != nil {
		return nil, fmt.Errorf("create CEL program: %w", err)
	}

	var matched []*graph.Node
	for _, n := range nodes {
		activation := n.ToMap()
		out, _, err := prg.Eval(activation)
		if err != nil {
			continue
		}
		if b, ok := out.Value().(bool); ok && b {
			matched = append(matched, n)
		}
	}

	if sortBy != "" {
		sortNodes(matched, sortBy)
	}

	if topN > 0 && len(matched) > topN {
		matched = matched[:topN]
	}

	return matched, nil
}

func sortNodes(nodes []*graph.Node, field string) {
	// Pre-compute field values once to avoid O(N log N) ToMap calls.
	vals := make([]any, len(nodes))
	for i, n := range nodes {
		vals[i] = n.ToMap()[field]
	}
	sort.Slice(nodes, func(i, j int) bool {
		vi, vj := vals[i], vals[j]
		switch a := vi.(type) {
		case int64:
			if b, ok := vj.(int64); ok {
				return a > b // descending
			}
		case float64:
			if b, ok := vj.(float64); ok {
				return a > b
			}
		case string:
			if b, ok := vj.(string); ok {
				return a < b // ascending for strings
			}
		}
		return false
	})
}

// EdgeEngine evaluates CEL expressions over edges.
type EdgeEngine struct {
	env *cel.Env
}

// NewEdgeEngine creates a CEL engine for edge queries.
func NewEdgeEngine() (*EdgeEngine, error) {
	env, err := cel.NewEnv(
		cel.Variable("from", cel.StringType),
		cel.Variable("to", cel.StringType),
		cel.Variable("kind", cel.StringType),
	)
	if err != nil {
		return nil, fmt.Errorf("create CEL env for edges: %w", err)
	}
	return &EdgeEngine{env: env}, nil
}

// QueryEdges evaluates a CEL expression against all edges.
func (e *EdgeEngine) QueryEdges(edges []*graph.Edge, expr string) ([]*graph.Edge, error) {
	ast, iss := e.env.Compile(expr)
	if iss != nil && iss.Err() != nil {
		return nil, fmt.Errorf("compile CEL expression: %w", iss.Err())
	}
	prg, err := e.env.Program(ast)
	if err != nil {
		return nil, fmt.Errorf("create CEL program: %w", err)
	}

	var matched []*graph.Edge
	for _, ed := range edges {
		activation := map[string]any{
			"from": ed.From,
			"to":   ed.To,
			"kind": ed.Kind,
		}
		out, _, err := prg.Eval(activation)
		if err != nil {
			continue
		}
		if b, ok := out.Value().(bool); ok && b {
			matched = append(matched, ed)
		}
	}
	return matched, nil
}

// Help returns static documentation on CEL query syntax.
func Help() string {
	return strings.TrimSpace(`
CEL Query Syntax for first-mate
================================

Nodes fields (use directly in expressions):
  id              string  - unique node identifier (e.g. "pkg.FuncName")
  kind            string  - "function", "type", "interface", "var", "const", "file", "package"
  name            string  - short name
  file            string  - relative file path
  line            int     - line number
  text            string  - source text snippet
  external        bool    - true if from external package
  cyclomatic      int     - cyclomatic complexity
  cognitive       int     - cognitive complexity
  receiver        string  - receiver type for methods
  params          string  - parameter types string
  returns         string  - return types string
  parent_id       string  - containing node ID
  callee_ids      string  - JSON array of called function IDs
  caller_ids      string  - JSON array of calling function IDs
  child_ids       string  - JSON array of child node IDs
  lint_count      int     - number of lint issues
  lint_issues     string  - semicolon-separated lint messages
  race_count      int     - number of detected races
  race_issues     string  - semicolon-separated race messages
  heap_allocs     int     - heap allocations from escape analysis
  stack_allocs    int     - stack allocations
  coverage        double  - test coverage percentage
  test_status     string  - "pass", "fail", "covered", "untested"
  bench_ns_op     double  - nanoseconds per operation from benchmark
  bench_b_op      double  - bytes per operation from benchmark
  bench_allocs_op double  - allocations per operation from benchmark
  pprof_flat_pct  double  - pprof flat percentage
  pprof_cum_pct   double  - pprof cumulative percentage
  vet_issues      string  - semicolon-separated vet messages

Edge fields:
  from  string  - source node ID
  to    string  - target node ID
  kind  string  - "call", "implements", "contains", "imports"

CEL operators:
  ==, !=, <, >, <=, >=       comparison
  &&, ||, !                   logical
  +, -, *, /                  arithmetic
  contains(), startsWith(), endsWith()  string methods
  matches()                   regex match

Examples:
  kind == "function" && cyclomatic > 10
  kind == "function" && lint_count > 0
  file.contains("_test.go")
  name.startsWith("Test")
`)
}

// Examples returns example CEL queries grouped by category.
func Examples() string {
	return strings.TrimSpace(`
CEL Query Examples
==================

Complexity:
  kind == "function" && cyclomatic > 10
  kind == "function" && cyclomatic > 20
  kind == "function" && cognitive > 15

Code Quality:
  lint_count > 0
  lint_count > 3
  vet_issues != ""

Testing:
  kind == "function" && name.startsWith("Test")
  kind == "function" && test_status == "fail"
  kind == "function" && coverage < 0.5

Analysis:
  kind == "function" && heap_allocs > 10
  kind == "function" && pprof_flat_pct > 5.0
  kind == "function" && bench_ns_op > 1000

Structure:
  kind == "interface"
  kind == "type" && external == false
  kind == "function" && receiver != ""
  file.contains("_test.go") && kind == "function"

Race Detection:
  race_count > 0

Benchmarks:
  kind == "function" && name.startsWith("Bench")
  bench_allocs_op > 10

Edge queries:
  kind == "call"
  kind == "implements"
  from.contains("handler")
`)
}
