package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mattdurham/bob/internal/firstmate/analysis"
	"github.com/mattdurham/bob/internal/firstmate/graph"
	"github.com/mattdurham/bob/internal/firstmate/parser"
	"github.com/mattdurham/bob/internal/firstmate/pr"
	"github.com/mattdurham/bob/internal/firstmate/query"
	"github.com/mattdurham/bob/internal/firstmate/spec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server holds shared state for all tools.
type Server struct {
	store     *graph.Store
	celEngine *query.Engine
	edgeEngine *query.EdgeEngine
}

// NewServer creates a new tool server with the given store.
func NewServer(store *graph.Store) (*Server, error) {
	celEng, err := query.NewEngine()
	if err != nil {
		return nil, fmt.Errorf("create CEL engine: %w", err)
	}
	edgeEng, err := query.NewEdgeEngine()
	if err != nil {
		return nil, fmt.Errorf("create edge CEL engine: %w", err)
	}
	return &Server{
		store:      store,
		celEngine:  celEng,
		edgeEngine: edgeEng,
	}, nil
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}

func errResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: "ERROR: " + msg}},
	}
}

// Register registers all tools with the MCP server.
func (s *Server) Register(mcpServer *mcp.Server) {
	// Graph management
	type ParseTreeArgs struct{}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "parse_tree",
		Description: "Parse all Go files under the current working directory into the graph",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ParseTreeArgs) (*mcp.CallToolResult, any, error) {
		return s.parseTree(ctx)
	})

	type ParseFilesArgs struct {
		Paths string `json:"paths" jsonschema:"comma-separated list of file paths to parse"`
	}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "parse_files",
		Description: "Parse specific Go files (comma-separated paths)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ParseFilesArgs) (*mcp.CallToolResult, any, error) {
		return s.parseFiles(ctx, args.Paths)
	})

	type ParsePackagesArgs struct {
		Packages string `json:"packages" jsonschema:"comma-separated package paths to parse"`
	}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "parse_packages",
		Description: "Parse specific Go package paths (comma-separated)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ParsePackagesArgs) (*mcp.CallToolResult, any, error) {
		return s.parsePackages(ctx, args.Packages)
	})

	type ResetGraphArgs struct{}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "reset_graph",
		Description: "Drop all nodes and edges from the graph",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ResetGraphArgs) (*mcp.CallToolResult, any, error) {
		return s.resetGraph(ctx)
	})

	type GetGraphArgs struct{}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "get_graph",
		Description: "Return JSON of all nodes and edges in the graph",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetGraphArgs) (*mcp.CallToolResult, any, error) {
		return s.getGraph(ctx)
	})

	type ListNodesArgs struct{}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "list_nodes",
		Description: "Return all node IDs and kinds",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ListNodesArgs) (*mcp.CallToolResult, any, error) {
		return s.listNodes(ctx)
	})

	type ListEdgesArgs struct{}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "list_edges",
		Description: "Return all edges in the graph",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ListEdgesArgs) (*mcp.CallToolResult, any, error) {
		return s.listEdges(ctx)
	})

	type GetNodesArgs struct {
		IDs string `json:"ids" jsonschema:"comma-separated node IDs to retrieve"`
	}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "get_nodes",
		Description: "Get specific nodes by ID (comma-separated)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetNodesArgs) (*mcp.CallToolResult, any, error) {
		return s.getNodes(ctx, args.IDs)
	})

	// Queries
	type QueryNodesArgs struct {
		Expr   string `json:"expr" jsonschema:"CEL expression to filter nodes"`
		SortBy string `json:"sort_by,omitempty" jsonschema:"field name to sort results by"`
		TopN   int    `json:"top_n,omitempty" jsonschema:"maximum number of results to return"`
	}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "query_nodes",
		Description: "CEL expression query over graph nodes",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args QueryNodesArgs) (*mcp.CallToolResult, any, error) {
		return s.queryNodes(ctx, args.Expr, args.SortBy, args.TopN)
	})

	type QueryEdgesArgs struct {
		Expr string `json:"expr" jsonschema:"CEL expression to filter edges"`
	}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "query_edges",
		Description: "CEL expression query over graph edges",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args QueryEdgesArgs) (*mcp.CallToolResult, any, error) {
		return s.queryEdges(ctx, args.Expr)
	})

	// Call graph
	type CallGraphArgs struct {
		FunctionID string `json:"function_id" jsonschema:"ID of the function to start traversal from"`
		Direction  string `json:"direction,omitempty" jsonschema:"callees, callers, or both (default: callees)"`
		Depth      int    `json:"depth,omitempty" jsonschema:"maximum traversal depth (default: 3)"`
	}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "call_graph",
		Description: "BFS/DFS traversal of the call graph from a function",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args CallGraphArgs) (*mcp.CallToolResult, any, error) {
		return s.callGraph(ctx, args.FunctionID, args.Direction, args.Depth)
	})

	type CallPathArgs struct {
		From string `json:"from" jsonschema:"source function ID"`
		To   string `json:"to" jsonschema:"target function ID"`
	}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "call_path",
		Description: "Find shortest call path between two functions using BFS",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args CallPathArgs) (*mcp.CallToolResult, any, error) {
		return s.callPath(ctx, args.From, args.To)
	})

	type FindImplementationsArgs struct {
		InterfaceID string `json:"interface_id" jsonschema:"interface node ID to find implementations of"`
	}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "find_implementations",
		Description: "Find all types implementing a given interface",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args FindImplementationsArgs) (*mcp.CallToolResult, any, error) {
		return s.findImplementations(ctx, args.InterfaceID)
	})

	type FindDeadcodeArgs struct{}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "find_deadcode",
		Description: "Find exported functions and types with no callers (potential dead code)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args FindDeadcodeArgs) (*mcp.CallToolResult, any, error) {
		return s.findDeadcode(ctx)
	})

	type FindTodosArgs struct{}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "find_todos",
		Description: "Walk AST for TODO, FIXME, and HACK comments",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args FindTodosArgs) (*mcp.CallToolResult, any, error) {
		return s.findTodos(ctx)
	})

	type FindHotspotsArgs struct{}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "find_hotspots",
		Description: "Find functions with cyclomatic complexity > 10, sorted by complexity descending",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args FindHotspotsArgs) (*mcp.CallToolResult, any, error) {
		return s.findHotspots(ctx)
	})

	type FindRacesArgs struct{}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "find_races",
		Description: "Heuristic AST analysis for potential race conditions",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args FindRacesArgs) (*mcp.CallToolResult, any, error) {
		return s.findRaces(ctx)
	})

	// Snapshots
	type GraphSnapshotArgs struct{}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "graph_snapshot",
		Description: "Save current nodes and edges as a named snapshot",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GraphSnapshotArgs) (*mcp.CallToolResult, any, error) {
		return s.graphSnapshot(ctx)
	})

	type GraphDiffArgs struct{}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "graph_diff",
		Description: "Compare current graph to most recent snapshot",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GraphDiffArgs) (*mcp.CallToolResult, any, error) {
		return s.graphDiff(ctx)
	})

	// Analysis
	type RunVetArgs struct{}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "run_vet",
		Description: "Run go vet ./... in current directory",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args RunVetArgs) (*mcp.CallToolResult, any, error) {
		return s.runVet(ctx)
	})

	type RunLintArgs struct{}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "run_lint",
		Description: "Run golangci-lint run ./... in current directory",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args RunLintArgs) (*mcp.CallToolResult, any, error) {
		return s.runLint(ctx)
	})

	type RunEscapeArgs struct{}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "run_escape",
		Description: "Run escape analysis via go build -gcflags='-m -m' ./...",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args RunEscapeArgs) (*mcp.CallToolResult, any, error) {
		return s.runEscape(ctx)
	})

	type RunTestsArgs struct{}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "run_tests",
		Description: "Run go test with coverage profile",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args RunTestsArgs) (*mcp.CallToolResult, any, error) {
		return s.runTests(ctx)
	})

	type RunBenchArgs struct{}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "run_bench",
		Description: "Run go test -bench=. -benchmem ./...",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args RunBenchArgs) (*mcp.CallToolResult, any, error) {
		return s.runBench(ctx)
	})

	type RunProfileArgs struct {
		ProfileFile string `json:"profile_file" jsonschema:"path to the .prof pprof file to analyze"`
	}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "run_profile",
		Description: "Parse a pprof profile file and annotate nodes with flat/cum percentages",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args RunProfileArgs) (*mcp.CallToolResult, any, error) {
		return s.runProfile(ctx, args.ProfileFile)
	})

	type RunAnalysisArgs struct{}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "run_analysis",
		Description: "Run vet + lint + escape + tests sequentially and return combined summary",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args RunAnalysisArgs) (*mcp.CallToolResult, any, error) {
		return s.runAnalysis(ctx)
	})

	type RunChecksArgs struct{}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "run_checks",
		Description: "Run vet + lint and return combined summary",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args RunChecksArgs) (*mcp.CallToolResult, any, error) {
		return s.runChecks(ctx)
	})

	// Spec lookup
	type ReadDocsArgs struct {
		Kind    string `json:"kind" jsonschema:"spec kind: SPECS, NOTES, TESTS, or BENCHMARKS"`
		Pattern string `json:"pattern,omitempty" jsonschema:"optional search string to filter content"`
	}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "read_docs",
		Description: "Read spec files by kind (SPECS/NOTES/TESTS/BENCHMARKS), optionally filtered by pattern",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ReadDocsArgs) (*mcp.CallToolResult, any, error) {
		return s.readDocs(ctx, args.Kind, args.Pattern)
	})

	type FindSpecArgs struct {
		Query string `json:"query" jsonschema:"symbol or keyword to search for in spec files"`
	}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "find_spec",
		Description: "Search all spec files for a symbol or keyword",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args FindSpecArgs) (*mcp.CallToolResult, any, error) {
		return s.findSpec(ctx, args.Query)
	})

	type ListSpecsArgs struct{}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "list_specs",
		Description: "List all spec files (SPECS.md, NOTES.md, TESTS.md, BENCHMARKS.md) in the project",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ListSpecsArgs) (*mcp.CallToolResult, any, error) {
		return s.listSpecs(ctx)
	})

	type GetSpecArgs struct {
		ID string `json:"id" jsonschema:"spec ID to look up (e.g. SPEC-001)"`
	}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "get_spec",
		Description: "Search for a specific spec ID across all spec files",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetSpecArgs) (*mcp.CallToolResult, any, error) {
		return s.getSpec(ctx, args.ID)
	})

	// PR monitoring
	type MonitorPRArgs struct {
		PR string `json:"pr" jsonschema:"PR URL or number to monitor"`
	}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "monitor_pr",
		Description: "Check PR status, CI checks, and reviews via the gh CLI",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args MonitorPRArgs) (*mcp.CallToolResult, any, error) {
		return s.monitorPR(ctx, args.PR)
	})

	// Help
	type QueryHelpArgs struct{}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "query_help",
		Description: "Return CEL query syntax documentation and available node/edge fields",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args QueryHelpArgs) (*mcp.CallToolResult, any, error) {
		return textResult(query.Help()), nil, nil
	})

	type QueryExamplesArgs struct{}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "query_examples",
		Description: "Return grouped example CEL queries",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args QueryExamplesArgs) (*mcp.CallToolResult, any, error) {
		return textResult(query.Examples()), nil, nil
	})

	type ToolsHelpArgs struct{}
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "tools_help",
		Description: "Return a summary of all available tools",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ToolsHelpArgs) (*mcp.CallToolResult, any, error) {
		return textResult(ToolsHelp()), nil, nil
	})
}

// --- Graph management ---

func (s *Server) parseTree(_ context.Context) (*mcp.CallToolResult, any, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return errResult(fmt.Sprintf("get working dir: %v", err)), nil, nil
	}

	p := parser.New()
	g, err := p.ParseDir(cwd)
	if err != nil {
		return errResult(fmt.Sprintf("parse dir: %v", err)), nil, nil
	}

	nodes := g.Nodes()
	edges := g.Edges()

	// Persist to store
	for _, n := range nodes {
		if err := s.store.SaveNode(n); err != nil {
			return errResult(fmt.Sprintf("save node %s: %v", n.ID, err)), nil, nil
		}
	}
	for _, e := range edges {
		if err := s.store.SaveEdge(e); err != nil {
			_ = err // ignore edge constraint violations
		}
	}

	return textResult(fmt.Sprintf("Parsed %d nodes, %d edges from %s", len(nodes), len(edges), cwd)), nil, nil
}

func (s *Server) parseFiles(_ context.Context, paths string) (*mcp.CallToolResult, any, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return errResult(fmt.Sprintf("get working dir: %v", err)), nil, nil
	}

	files := splitTrim(paths)
	if len(files) == 0 {
		return errResult("no files specified"), nil, nil
	}

	// Resolve to absolute paths
	absFiles := make([]string, 0, len(files))
	for _, f := range files {
		if !filepath.IsAbs(f) {
			f = filepath.Join(cwd, f)
		}
		absFiles = append(absFiles, f)
	}

	p := parser.New()
	g, err := p.ParseFiles(absFiles, cwd)
	if err != nil {
		return errResult(fmt.Sprintf("parse files: %v", err)), nil, nil
	}

	nodes := g.Nodes()
	edges := g.Edges()
	for _, n := range nodes {
		s.store.SaveNode(n)
	}
	for _, e := range edges {
		s.store.SaveEdge(e)
	}

	return textResult(fmt.Sprintf("Parsed %d nodes, %d edges from %d files", len(nodes), len(edges), len(absFiles))), nil, nil
}

func (s *Server) parsePackages(_ context.Context, packages string) (*mcp.CallToolResult, any, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return errResult(fmt.Sprintf("get working dir: %v", err)), nil, nil
	}

	pkgs := splitTrim(packages)
	if len(pkgs) == 0 {
		return errResult("no packages specified"), nil, nil
	}

	p := parser.New()
	totalNodes := 0
	totalEdges := 0

	for _, pkg := range pkgs {
		pkgPath := pkg
		if !filepath.IsAbs(pkgPath) {
			pkgPath = filepath.Join(cwd, pkg)
		}
		g, err := p.ParseDir(pkgPath)
		if err != nil {
			continue
		}
		nodes := g.Nodes()
		edges := g.Edges()
		for _, n := range nodes {
			s.store.SaveNode(n)
		}
		for _, e := range edges {
			s.store.SaveEdge(e)
		}
		totalNodes += len(nodes)
		totalEdges += len(edges)
	}

	return textResult(fmt.Sprintf("Parsed %d nodes, %d edges from %d packages", totalNodes, totalEdges, len(pkgs))), nil, nil
}

func (s *Server) resetGraph(_ context.Context) (*mcp.CallToolResult, any, error) {
	if err := s.store.Reset(); err != nil {
		return errResult(fmt.Sprintf("reset graph: %v", err)), nil, nil
	}
	return textResult("Graph reset: all nodes and edges removed."), nil, nil
}

func (s *Server) getGraph(_ context.Context) (*mcp.CallToolResult, any, error) {
	nodes, err := s.store.AllNodes()
	if err != nil {
		return errResult(fmt.Sprintf("load nodes: %v", err)), nil, nil
	}
	edges, err := s.store.AllEdges()
	if err != nil {
		return errResult(fmt.Sprintf("load edges: %v", err)), nil, nil
	}
	data := map[string]interface{}{
		"nodes": nodes,
		"edges": edges,
	}
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return errResult(fmt.Sprintf("marshal graph: %v", err)), nil, nil
	}
	return textResult(string(b)), nil, nil
}

func (s *Server) listNodes(_ context.Context) (*mcp.CallToolResult, any, error) {
	nodes, err := s.store.AllNodes()
	if err != nil {
		return errResult(fmt.Sprintf("load nodes: %v", err)), nil, nil
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "Total: %d nodes\n\n", len(nodes))
	for _, n := range nodes {
		fmt.Fprintf(&sb, "[%s] %s\n", n.Kind, n.ID)
	}
	return textResult(sb.String()), nil, nil
}

func (s *Server) listEdges(_ context.Context) (*mcp.CallToolResult, any, error) {
	edges, err := s.store.AllEdges()
	if err != nil {
		return errResult(fmt.Sprintf("load edges: %v", err)), nil, nil
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "Total: %d edges\n\n", len(edges))
	for _, e := range edges {
		fmt.Fprintf(&sb, "%s -[%s]-> %s\n", e.From, e.Kind, e.To)
	}
	return textResult(sb.String()), nil, nil
}

func (s *Server) getNodes(_ context.Context, ids string) (*mcp.CallToolResult, any, error) {
	nodeIDs := splitTrim(ids)
	if len(nodeIDs) == 0 {
		return errResult("no IDs specified"), nil, nil
	}
	var result []*graph.Node
	for _, id := range nodeIDs {
		n, err := s.store.GetNodeByID(id)
		if err != nil {
			return errResult(fmt.Sprintf("get node %s: %v", id, err)), nil, nil
		}
		if n != nil {
			result = append(result, n)
		}
	}
	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return errResult(fmt.Sprintf("marshal nodes: %v", err)), nil, nil
	}
	return textResult(string(b)), nil, nil
}

// --- Queries ---

func (s *Server) queryNodes(_ context.Context, expr, sortBy string, topN int) (*mcp.CallToolResult, any, error) {
	if expr == "" {
		return errResult("expr is required"), nil, nil
	}
	nodes, err := s.store.AllNodes()
	if err != nil {
		return errResult(fmt.Sprintf("load nodes: %v", err)), nil, nil
	}
	matched, err := s.celEngine.QueryNodes(nodes, expr, sortBy, topN)
	if err != nil {
		return errResult(fmt.Sprintf("CEL query: %v", err)), nil, nil
	}
	b, err := json.MarshalIndent(matched, "", "  ")
	if err != nil {
		return errResult(fmt.Sprintf("marshal result: %v", err)), nil, nil
	}
	return textResult(fmt.Sprintf("Found %d nodes:\n%s", len(matched), string(b))), nil, nil
}

func (s *Server) queryEdges(_ context.Context, expr string) (*mcp.CallToolResult, any, error) {
	if expr == "" {
		return errResult("expr is required"), nil, nil
	}
	edges, err := s.store.AllEdges()
	if err != nil {
		return errResult(fmt.Sprintf("load edges: %v", err)), nil, nil
	}
	matched, err := s.edgeEngine.QueryEdges(edges, expr)
	if err != nil {
		return errResult(fmt.Sprintf("CEL query: %v", err)), nil, nil
	}
	b, err := json.MarshalIndent(matched, "", "  ")
	if err != nil {
		return errResult(fmt.Sprintf("marshal result: %v", err)), nil, nil
	}
	return textResult(fmt.Sprintf("Found %d edges:\n%s", len(matched), string(b))), nil, nil
}

// --- Call graph ---

func (s *Server) callGraph(_ context.Context, functionID, direction string, depth int) (*mcp.CallToolResult, any, error) {
	if functionID == "" {
		return errResult("function_id is required"), nil, nil
	}
	if direction == "" {
		direction = "callees"
	}
	if depth <= 0 {
		depth = 3
	}

	nodes, err := s.store.AllNodes()
	if err != nil {
		return errResult(fmt.Sprintf("load nodes: %v", err)), nil, nil
	}

	nodeMap := make(map[string]*graph.Node, len(nodes))
	for _, n := range nodes {
		nodeMap[n.ID] = n
	}

	startNode, ok := nodeMap[functionID]
	if !ok {
		return errResult(fmt.Sprintf("function not found: %s", functionID)), nil, nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Call graph for: %s (direction: %s, depth: %d)\n\n", functionID, direction, depth)

	visited := make(map[string]bool)
	var traverse func(id string, level int)
	traverse = func(id string, level int) {
		if level > depth || visited[id] {
			return
		}
		visited[id] = true
		n, ok := nodeMap[id]
		if !ok {
			return
		}
		indent := strings.Repeat("  ", level)
		fmt.Fprintf(&sb, "%s%s (%s)\n", indent, n.ID, n.Kind)

		var nextIDs []string
		switch direction {
		case "callers":
			nextIDs = n.CallerIDs
		case "both":
			nextIDs = append(n.CalleeIDs, n.CallerIDs...)
		default:
			nextIDs = n.CalleeIDs
		}
		for _, next := range nextIDs {
			traverse(next, level+1)
		}
	}

	_ = startNode
	traverse(functionID, 0)
	return textResult(sb.String()), nil, nil
}

func (s *Server) callPath(_ context.Context, from, to string) (*mcp.CallToolResult, any, error) {
	if from == "" || to == "" {
		return errResult("from and to are required"), nil, nil
	}

	nodes, err := s.store.AllNodes()
	if err != nil {
		return errResult(fmt.Sprintf("load nodes: %v", err)), nil, nil
	}

	nodeMap := make(map[string]*graph.Node, len(nodes))
	for _, n := range nodes {
		nodeMap[n.ID] = n
	}

	// BFS to find shortest path
	type step struct {
		id   string
		path []string
	}
	queue := []step{{id: from, path: []string{from}}}
	visited := map[string]bool{from: true}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		if cur.id == to {
			return textResult(fmt.Sprintf("Shortest path (%d steps):\n%s", len(cur.path)-1, strings.Join(cur.path, " -> "))), nil, nil
		}

		n, ok := nodeMap[cur.id]
		if !ok {
			continue
		}
		for _, callee := range n.CalleeIDs {
			if !visited[callee] {
				visited[callee] = true
				newPath := make([]string, len(cur.path)+1)
				copy(newPath, cur.path)
				newPath[len(cur.path)] = callee
				queue = append(queue, step{id: callee, path: newPath})
			}
		}
	}

	return textResult(fmt.Sprintf("No path found from %s to %s", from, to)), nil, nil
}

func (s *Server) findImplementations(_ context.Context, interfaceID string) (*mcp.CallToolResult, any, error) {
	if interfaceID == "" {
		return errResult("interface_id is required"), nil, nil
	}

	edges, err := s.store.AllEdges()
	if err != nil {
		return errResult(fmt.Sprintf("load edges: %v", err)), nil, nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Implementations of %s:\n\n", interfaceID)
	count := 0
	for _, e := range edges {
		if e.Kind == "implements" && e.To == interfaceID {
			fmt.Fprintf(&sb, "  %s\n", e.From)
			count++
		}
	}
	if count == 0 {
		sb.WriteString("  (none found)\n")
	}
	return textResult(sb.String()), nil, nil
}

func (s *Server) findDeadcode(_ context.Context) (*mcp.CallToolResult, any, error) {
	nodes, err := s.store.AllNodes()
	if err != nil {
		return errResult(fmt.Sprintf("load nodes: %v", err)), nil, nil
	}

	var sb strings.Builder
	sb.WriteString("Potential dead code (exported with no callers):\n\n")
	count := 0

	for _, n := range nodes {
		if n.Kind != "function" && n.Kind != "type" {
			continue
		}
		if n.External {
			continue
		}
		if len(n.CallerIDs) > 0 {
			continue
		}
		// Check if exported (starts with uppercase)
		if len(n.Name) == 0 || n.Name[0] < 'A' || n.Name[0] > 'Z' {
			continue
		}
		// Skip main, init, Test*, Bench*
		if n.Name == "main" || n.Name == "init" ||
			strings.HasPrefix(n.Name, "Test") ||
			strings.HasPrefix(n.Name, "Bench") ||
			strings.HasPrefix(n.Name, "Example") {
			continue
		}
		fmt.Fprintf(&sb, "  [%s] %s (%s:%d)\n", n.Kind, n.ID, n.File, n.Line)
		count++
	}

	if count == 0 {
		sb.WriteString("  (none found)\n")
	} else {
		fmt.Fprintf(&sb, "\nTotal: %d potential dead code items\n", count)
	}
	return textResult(sb.String()), nil, nil
}

func (s *Server) findTodos(_ context.Context) (*mcp.CallToolResult, any, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return errResult(fmt.Sprintf("get working dir: %v", err)), nil, nil
	}

	findings, err := findTodosInDir(cwd)
	if err != nil {
		return errResult(fmt.Sprintf("find todos: %v", err)), nil, nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "TODO/FIXME/HACK comments (%d found):\n\n", len(findings))
	for _, f := range findings {
		sb.WriteString(f + "\n")
	}
	return textResult(sb.String()), nil, nil
}

func findTodosInDir(root string) ([]string, error) {
	var findings []string
	fset := newFset()
	err := walkGoFiles(root, func(path, rel string) error {
		f, err := parseFileForComments(fset, path)
		if err != nil {
			return nil
		}
		for _, cg := range f.Comments {
			for _, c := range cg.List {
				text := c.Text
				upper := strings.ToUpper(text)
				if strings.Contains(upper, "TODO") || strings.Contains(upper, "FIXME") || strings.Contains(upper, "HACK") {
					pos := fset.Position(c.Pos())
					findings = append(findings, fmt.Sprintf("%s:%d: %s", rel, pos.Line, strings.TrimSpace(text)))
				}
			}
		}
		return nil
	})
	return findings, err
}

func (s *Server) findHotspots(_ context.Context) (*mcp.CallToolResult, any, error) {
	nodes, err := s.store.AllNodes()
	if err != nil {
		return errResult(fmt.Sprintf("load nodes: %v", err)), nil, nil
	}

	var hotspots []*graph.Node
	for _, n := range nodes {
		if n.Kind == "function" && n.Cyclomatic > 10 {
			hotspots = append(hotspots, n)
		}
	}

	sort.Slice(hotspots, func(i, j int) bool {
		return hotspots[i].Cyclomatic > hotspots[j].Cyclomatic
	})

	var sb strings.Builder
	fmt.Fprintf(&sb, "High-complexity functions (%d found):\n\n", len(hotspots))
	for _, n := range hotspots {
		sb.WriteString(fmt.Sprintf("  cyclomatic=%d cognitive=%d  %s  (%s:%d)\n",
			n.Cyclomatic, n.Cognitive, n.ID, n.File, n.Line))
	}
	if len(hotspots) == 0 {
		sb.WriteString("  (none found with cyclomatic > 10)\n")
	}
	return textResult(sb.String()), nil, nil
}

func (s *Server) findRaces(_ context.Context) (*mcp.CallToolResult, any, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return errResult(fmt.Sprintf("get working dir: %v", err)), nil, nil
	}

	findings, err := analysis.FindRaces(cwd)
	if err != nil {
		return errResult(fmt.Sprintf("find races: %v", err)), nil, nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Potential race conditions (%d found):\n\n", len(findings))
	for _, f := range findings {
		fmt.Fprintf(&sb, "  %s:%d: %s\n", f.File, f.Line, f.Message)
	}
	if len(findings) == 0 {
		sb.WriteString("  (none found by heuristic analysis)\n")
		sb.WriteString("  Note: run 'go test -race ./...' for authoritative race detection.\n")
	}
	return textResult(sb.String()), nil, nil
}

// --- Snapshots ---

func (s *Server) graphSnapshot(_ context.Context) (*mcp.CallToolResult, any, error) {
	nodes, err := s.store.AllNodes()
	if err != nil {
		return errResult(fmt.Sprintf("load nodes: %v", err)), nil, nil
	}
	edges, err := s.store.AllEdges()
	if err != nil {
		return errResult(fmt.Sprintf("load edges: %v", err)), nil, nil
	}

	name := fmt.Sprintf("snap-%d", time.Now().Unix())
	if err := s.store.SaveSnapshot(name, nodes, edges); err != nil {
		return errResult(fmt.Sprintf("save snapshot: %v", err)), nil, nil
	}

	return textResult(fmt.Sprintf("Snapshot saved: %s (%d nodes, %d edges)", name, len(nodes), len(edges))), nil, nil
}

func (s *Server) graphDiff(_ context.Context) (*mcp.CallToolResult, any, error) {
	snap, err := s.store.LoadLatestSnapshot()
	if err != nil {
		return errResult(fmt.Sprintf("load snapshot: %v", err)), nil, nil
	}

	diff, err := s.store.DiffWithCurrent(snap)
	if err != nil {
		return errResult(fmt.Sprintf("diff: %v", err)), nil, nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Diff vs snapshot %q:\n\n", snap.Name)
	fmt.Fprintf(&sb, "Added (%d):\n", len(diff.Added))
	for _, id := range diff.Added {
		fmt.Fprintf(&sb, "  + %s\n", id)
	}
	fmt.Fprintf(&sb, "\nRemoved (%d):\n", len(diff.Removed))
	for _, id := range diff.Removed {
		fmt.Fprintf(&sb, "  - %s\n", id)
	}
	fmt.Fprintf(&sb, "\nChanged (%d):\n", len(diff.Changed))
	for _, id := range diff.Changed {
		fmt.Fprintf(&sb, "  ~ %s\n", id)
	}
	return textResult(sb.String()), nil, nil
}

// --- Analysis ---

func (s *Server) runVet(_ context.Context) (*mcp.CallToolResult, any, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return errResult(fmt.Sprintf("get working dir: %v", err)), nil, nil
	}
	result, err := analysis.RunVet(cwd, s.store)
	if err != nil {
		return errResult(fmt.Sprintf("run vet: %v", err)), nil, nil
	}
	return textResult(result), nil, nil
}

func (s *Server) runLint(_ context.Context) (*mcp.CallToolResult, any, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return errResult(fmt.Sprintf("get working dir: %v", err)), nil, nil
	}
	result, err := analysis.RunLint(cwd, s.store)
	if err != nil {
		return errResult(fmt.Sprintf("run lint: %v", err)), nil, nil
	}
	return textResult(result), nil, nil
}

func (s *Server) runEscape(_ context.Context) (*mcp.CallToolResult, any, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return errResult(fmt.Sprintf("get working dir: %v", err)), nil, nil
	}
	result, err := analysis.RunEscape(cwd, s.store)
	if err != nil {
		return errResult(fmt.Sprintf("run escape: %v", err)), nil, nil
	}
	return textResult(result), nil, nil
}

func (s *Server) runTests(_ context.Context) (*mcp.CallToolResult, any, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return errResult(fmt.Sprintf("get working dir: %v", err)), nil, nil
	}
	result, err := analysis.RunTests(cwd, s.store)
	if err != nil {
		return errResult(fmt.Sprintf("run tests: %v", err)), nil, nil
	}
	return textResult(result), nil, nil
}

func (s *Server) runBench(_ context.Context) (*mcp.CallToolResult, any, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return errResult(fmt.Sprintf("get working dir: %v", err)), nil, nil
	}
	result, err := analysis.RunBench(cwd, s.store)
	if err != nil {
		return errResult(fmt.Sprintf("run bench: %v", err)), nil, nil
	}
	return textResult(result), nil, nil
}

func (s *Server) runProfile(_ context.Context, profileFile string) (*mcp.CallToolResult, any, error) {
	if profileFile == "" {
		return errResult("profile_file is required"), nil, nil
	}
	result, err := analysis.RunProfile(profileFile, s.store)
	if err != nil {
		return errResult(fmt.Sprintf("run profile: %v", err)), nil, nil
	}
	return textResult(result), nil, nil
}

func (s *Server) runAnalysis(ctx context.Context) (*mcp.CallToolResult, any, error) {
	var sb strings.Builder
	sb.WriteString("=== Full Analysis ===\n\n")

	cwd, err := os.Getwd()
	if err != nil {
		return errResult(fmt.Sprintf("get working dir: %v", err)), nil, nil
	}

	if r, err := analysis.RunVet(cwd, s.store); err == nil {
		sb.WriteString(r)
	}
	if r, err := analysis.RunLint(cwd, s.store); err == nil {
		sb.WriteString(r)
	}
	if r, err := analysis.RunEscape(cwd, s.store); err == nil {
		sb.WriteString(r)
	}
	if r, err := analysis.RunTests(cwd, s.store); err == nil {
		sb.WriteString(r)
	}

	return textResult(sb.String()), nil, nil
}

func (s *Server) runChecks(_ context.Context) (*mcp.CallToolResult, any, error) {
	var sb strings.Builder
	sb.WriteString("=== Checks ===\n\n")

	cwd, err := os.Getwd()
	if err != nil {
		return errResult(fmt.Sprintf("get working dir: %v", err)), nil, nil
	}

	if r, err := analysis.RunVet(cwd, s.store); err == nil {
		sb.WriteString(r)
	}
	if r, err := analysis.RunLint(cwd, s.store); err == nil {
		sb.WriteString(r)
	}

	return textResult(sb.String()), nil, nil
}

// --- Spec lookup ---

func (s *Server) readDocs(_ context.Context, kind, pattern string) (*mcp.CallToolResult, any, error) {
	if kind == "" {
		return errResult("kind is required (SPECS, NOTES, TESTS, BENCHMARKS)"), nil, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return errResult(fmt.Sprintf("get working dir: %v", err)), nil, nil
	}
	result, err := spec.ReadByKind(cwd, kind, pattern)
	if err != nil {
		return errResult(fmt.Sprintf("read docs: %v", err)), nil, nil
	}
	return textResult(result), nil, nil
}

func (s *Server) findSpec(_ context.Context, query string) (*mcp.CallToolResult, any, error) {
	if query == "" {
		return errResult("query is required"), nil, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return errResult(fmt.Sprintf("get working dir: %v", err)), nil, nil
	}
	result, err := spec.SearchAll(cwd, query)
	if err != nil {
		return errResult(fmt.Sprintf("find spec: %v", err)), nil, nil
	}
	if len(result.Specs) == 0 && len(result.CodeRefs) == 0 {
		return textResult(fmt.Sprintf("No matches for %q in spec files or source", query)), nil, nil
	}
	out, err := json.Marshal(result)
	if err != nil {
		return errResult(fmt.Sprintf("marshal results: %v", err)), nil, nil
	}
	return textResult(string(out)), nil, nil
}

func (s *Server) listSpecs(_ context.Context) (*mcp.CallToolResult, any, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return errResult(fmt.Sprintf("get working dir: %v", err)), nil, nil
	}
	result, err := spec.ListAll(cwd)
	if err != nil {
		return errResult(fmt.Sprintf("list specs: %v", err)), nil, nil
	}
	return textResult(result), nil, nil
}

func (s *Server) getSpec(_ context.Context, id string) (*mcp.CallToolResult, any, error) {
	if id == "" {
		return errResult("id is required"), nil, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return errResult(fmt.Sprintf("get working dir: %v", err)), nil, nil
	}
	result, err := spec.GetByID(cwd, id)
	if err != nil {
		return errResult(fmt.Sprintf("get spec: %v", err)), nil, nil
	}
	return textResult(result), nil, nil
}

// --- PR monitoring ---

func (s *Server) monitorPR(_ context.Context, prRef string) (*mcp.CallToolResult, any, error) {
	if prRef == "" {
		return errResult("pr is required"), nil, nil
	}
	result, err := pr.MonitorPR(prRef)
	if err != nil {
		return errResult(fmt.Sprintf("monitor pr: %v", err)), nil, nil
	}
	return textResult(result), nil, nil
}

// --- helpers ---

func splitTrim(s string) []string {
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
