package tools

import "strings"

// ToolsHelp returns a summary of all available tools.
func ToolsHelp() string {
	return strings.TrimSpace(`
first-mate — Go Code Graph Analysis MCP Server
===============================================

Graph Management:
  parse_tree          Parse all Go files under current working directory into the graph
  parse_files         Parse specific files (comma-separated paths)
  parse_packages      Parse specific package paths (comma-separated)
  reset_graph         Drop all nodes and edges from the graph
  get_graph           Return JSON of all nodes and edges
  list_nodes          Return all node IDs and kinds
  list_edges          Return all edges
  get_nodes           Get specific nodes by ID (comma-separated)

Queries:
  query_nodes         CEL expression query over nodes (args: expr, sort_by, top_n)
  query_edges         CEL expression query over edges (args: expr)

Call Graph:
  call_graph          BFS traversal from a function (args: function_id, direction, depth)
  call_path           Find shortest path between two functions (args: from, to)
  find_implementations  Find all types implementing an interface (args: interface_id)
  find_deadcode       Find exported functions with no callers
  find_todos          Walk AST for TODO/FIXME/HACK comments
  find_hotspots       Query functions with cyclomatic > 10
  find_races          Heuristic AST analysis for potential race conditions

Snapshots:
  graph_snapshot      Save current graph as a named snapshot
  graph_diff          Compare current graph to most recent snapshot

Analysis (shell-out):
  run_vet             Run go vet ./...
  run_lint            Run golangci-lint run ./...
  run_escape          Run go build -gcflags='-m -m' escape analysis
  run_tests           Run go test with coverage
  run_bench           Run go test -bench=. -benchmem
  run_profile         Parse a pprof profile file (args: profile_file)
  run_analysis        Run vet + lint + escape + tests
  run_checks          Run vet + lint

Spec Lookup:
  read_docs           Read spec files by kind: SPECS/NOTES/TESTS/BENCHMARKS (args: kind, pattern)
  find_spec           Search spec files for a symbol/keyword (args: query)
  list_specs          List all spec files in the project
  get_spec            Search for a specific spec ID like "SPEC-001" (args: id)

PR Monitoring:
  monitor_pr          Check PR status, CI, and reviews via gh CLI (args: pr)

Help:
  query_help          CEL query syntax and available fields
  query_examples      Example CEL queries by category
  tools_help          This summary
`)
}
