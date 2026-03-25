package graph

import (
	"slices"
	"strings"
)

// Node represents a Go code element in the graph.
type Node struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"` // "function", "type", "interface", "var", "const", "file", "package"
	Name     string `json:"name"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Text     string `json:"text"`
	External bool   `json:"external"`

	// Function-specific
	Cyclomatic int    `json:"cyclomatic"`
	Cognitive  int    `json:"cognitive"`
	Receiver   string `json:"receiver"`
	Params     string `json:"params"`
	Returns    string `json:"returns"`

	// Relationships
	ParentID  string   `json:"parent_id"`
	CalleeIDs []string `json:"callee_ids"`
	CallerIDs []string `json:"caller_ids"`
	ChildIDs  []string `json:"child_ids"`

	// Analysis annotations
	LintCount    int     `json:"lint_count"`
	LintIssues   string  `json:"lint_issues"`
	RaceCount    int     `json:"race_count"`
	RaceIssues   string  `json:"race_issues"`
	HeapAllocs   int     `json:"heap_allocs"`
	StackAllocs  int     `json:"stack_allocs"`
	Coverage     float64 `json:"coverage"`
	TestStatus   string  `json:"test_status"`
	BenchNsOp    float64 `json:"bench_ns_op"`
	BenchBOp     float64 `json:"bench_b_op"`
	BenchAllocsOp float64 `json:"bench_allocs_op"`
	PprofFlatPct float64 `json:"pprof_flat_pct"`
	PprofCumPct  float64 `json:"pprof_cum_pct"`
	VetIssues    string  `json:"vet_issues"`
}

// Edge represents a directed relationship between two nodes.
type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Kind string `json:"kind"` // "call", "implements", "contains", "imports"
}

// ToMap converts a Node to a map for CEL evaluation.
func (n *Node) ToMap() map[string]any {
	return map[string]any{
		"id":              n.ID,
		"kind":            n.Kind,
		"name":            n.Name,
		"file":            n.File,
		"line":            int64(n.Line),
		"text":            n.Text,
		"external":        n.External,
		"cyclomatic":      int64(n.Cyclomatic),
		"cognitive":       int64(n.Cognitive),
		"receiver":        n.Receiver,
		"params":          n.Params,
		"returns":         n.Returns,
		"parent_id":       n.ParentID,
		"callee_ids":      joinIDs(n.CalleeIDs),
		"caller_ids":      joinIDs(n.CallerIDs),
		"child_ids":       joinIDs(n.ChildIDs),
		"lint_count":      int64(n.LintCount),
		"lint_issues":     n.LintIssues,
		"race_count":      int64(n.RaceCount),
		"race_issues":     n.RaceIssues,
		"heap_allocs":     int64(n.HeapAllocs),
		"stack_allocs":    int64(n.StackAllocs),
		"coverage":        n.Coverage,
		"test_status":     n.TestStatus,
		"bench_ns_op":     n.BenchNsOp,
		"bench_b_op":      n.BenchBOp,
		"bench_allocs_op": n.BenchAllocsOp,
		"pprof_flat_pct":  n.PprofFlatPct,
		"pprof_cum_pct":   n.PprofCumPct,
		"vet_issues":      n.VetIssues,
	}
}

// joinIDs encodes a string slice as a JSON array without reflection or allocs.
func joinIDs(ids []string) string {
	if len(ids) == 0 {
		return "[]"
	}
	var sb strings.Builder
	sb.WriteByte('[')
	for i, id := range ids {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteByte('"')
		sb.WriteString(id)
		sb.WriteByte('"')
	}
	sb.WriteByte(']')
	return sb.String()
}

// AddCallee adds a callee ID if not already present.
func (n *Node) AddCallee(id string) {
	if !slices.Contains(n.CalleeIDs, id) {
		n.CalleeIDs = append(n.CalleeIDs, id)
	}
}

// AddCaller adds a caller ID if not already present.
func (n *Node) AddCaller(id string) {
	if !slices.Contains(n.CallerIDs, id) {
		n.CallerIDs = append(n.CallerIDs, id)
	}
}

// AddChild adds a child ID if not already present.
func (n *Node) AddChild(id string) {
	if !slices.Contains(n.ChildIDs, id) {
		n.ChildIDs = append(n.ChildIDs, id)
	}
}
