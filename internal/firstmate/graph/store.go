package graph

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS nodes (
    id TEXT PRIMARY KEY,
    kind TEXT, name TEXT, file TEXT, line INTEGER, text TEXT,
    external INTEGER DEFAULT 0,
    cyclomatic INTEGER DEFAULT 0, cognitive INTEGER DEFAULT 0,
    receiver TEXT DEFAULT '', params TEXT DEFAULT '', returns TEXT DEFAULT '',
    parent_id TEXT DEFAULT '',
    callee_ids TEXT DEFAULT '[]', caller_ids TEXT DEFAULT '[]', child_ids TEXT DEFAULT '[]',
    lint_count INTEGER DEFAULT 0, lint_issues TEXT DEFAULT '',
    race_count INTEGER DEFAULT 0, race_issues TEXT DEFAULT '',
    heap_allocs INTEGER DEFAULT 0, stack_allocs INTEGER DEFAULT 0,
    coverage REAL DEFAULT 0, test_status TEXT DEFAULT '',
    bench_ns_op REAL DEFAULT 0, bench_b_op REAL DEFAULT 0, bench_allocs_op REAL DEFAULT 0,
    pprof_flat_pct REAL DEFAULT 0, pprof_cum_pct REAL DEFAULT 0,
    vet_issues TEXT DEFAULT ''
);
CREATE TABLE IF NOT EXISTS edges (
    from_id TEXT, to_id TEXT, kind TEXT,
    PRIMARY KEY (from_id, to_id, kind)
);
CREATE TABLE IF NOT EXISTS snapshots (
    name TEXT PRIMARY KEY,
    created_at INTEGER,
    nodes_json TEXT,
    edges_json TEXT
);
`

// Store handles SQLite persistence for the graph.
type Store struct {
	db *sql.DB
}

// DBPath returns the default SQLite database path.
func DBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	dir := filepath.Join(home, ".local", "share", "first-mate")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("could not create data directory: %w", err)
	}
	return filepath.Join(dir, "graph.db"), nil
}

// Open opens (or creates) the SQLite database and runs migrations.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	return &Store{db: db}, nil
}

// Close closes the database.
func (s *Store) Close() error {
	return s.db.Close()
}

// SaveNode upserts a node.
func (s *Store) SaveNode(n *Node) error {
	calleeJSON, _ := json.Marshal(n.CalleeIDs)
	callerJSON, _ := json.Marshal(n.CallerIDs)
	childJSON, _ := json.Marshal(n.ChildIDs)
	external := 0
	if n.External {
		external = 1
	}
	_, err := s.db.Exec(`
		INSERT INTO nodes (id, kind, name, file, line, text, external,
			cyclomatic, cognitive, receiver, params, returns, parent_id,
			callee_ids, caller_ids, child_ids,
			lint_count, lint_issues, race_count, race_issues,
			heap_allocs, stack_allocs, coverage, test_status,
			bench_ns_op, bench_b_op, bench_allocs_op,
			pprof_flat_pct, pprof_cum_pct, vet_issues)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(id) DO UPDATE SET
			kind=excluded.kind, name=excluded.name, file=excluded.file,
			line=excluded.line, text=excluded.text, external=excluded.external,
			cyclomatic=excluded.cyclomatic, cognitive=excluded.cognitive,
			receiver=excluded.receiver, params=excluded.params, returns=excluded.returns,
			parent_id=excluded.parent_id, callee_ids=excluded.callee_ids,
			caller_ids=excluded.caller_ids, child_ids=excluded.child_ids,
			lint_count=excluded.lint_count, lint_issues=excluded.lint_issues,
			race_count=excluded.race_count, race_issues=excluded.race_issues,
			heap_allocs=excluded.heap_allocs, stack_allocs=excluded.stack_allocs,
			coverage=excluded.coverage, test_status=excluded.test_status,
			bench_ns_op=excluded.bench_ns_op, bench_b_op=excluded.bench_b_op,
			bench_allocs_op=excluded.bench_allocs_op,
			pprof_flat_pct=excluded.pprof_flat_pct, pprof_cum_pct=excluded.pprof_cum_pct,
			vet_issues=excluded.vet_issues`,
		n.ID, n.Kind, n.Name, n.File, n.Line, n.Text, external,
		n.Cyclomatic, n.Cognitive, n.Receiver, n.Params, n.Returns, n.ParentID,
		string(calleeJSON), string(callerJSON), string(childJSON),
		n.LintCount, n.LintIssues, n.RaceCount, n.RaceIssues,
		n.HeapAllocs, n.StackAllocs, n.Coverage, n.TestStatus,
		n.BenchNsOp, n.BenchBOp, n.BenchAllocsOp,
		n.PprofFlatPct, n.PprofCumPct, n.VetIssues,
	)
	return err
}

// SaveEdge upserts an edge.
func (s *Store) SaveEdge(e *Edge) error {
	_, err := s.db.Exec(`
		INSERT OR IGNORE INTO edges (from_id, to_id, kind) VALUES (?,?,?)`,
		e.From, e.To, e.Kind,
	)
	return err
}

// LoadAll loads all nodes and edges from the database into the graph.
func (s *Store) LoadAll(g *Graph) error {
	if err := s.loadNodes(g); err != nil {
		return err
	}
	return s.loadEdges(g)
}

func (s *Store) loadNodes(g *Graph) error {
	rows, err := s.db.Query(`SELECT id, kind, name, file, line, text, external,
		cyclomatic, cognitive, receiver, params, returns, parent_id,
		callee_ids, caller_ids, child_ids,
		lint_count, lint_issues, race_count, race_issues,
		heap_allocs, stack_allocs, coverage, test_status,
		bench_ns_op, bench_b_op, bench_allocs_op,
		pprof_flat_pct, pprof_cum_pct, vet_issues
		FROM nodes`)
	if err != nil {
		return fmt.Errorf("query nodes: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		n := &Node{}
		var external int
		var calleeJSON, callerJSON, childJSON string
		err := rows.Scan(
			&n.ID, &n.Kind, &n.Name, &n.File, &n.Line, &n.Text, &external,
			&n.Cyclomatic, &n.Cognitive, &n.Receiver, &n.Params, &n.Returns, &n.ParentID,
			&calleeJSON, &callerJSON, &childJSON,
			&n.LintCount, &n.LintIssues, &n.RaceCount, &n.RaceIssues,
			&n.HeapAllocs, &n.StackAllocs, &n.Coverage, &n.TestStatus,
			&n.BenchNsOp, &n.BenchBOp, &n.BenchAllocsOp,
			&n.PprofFlatPct, &n.PprofCumPct, &n.VetIssues,
		)
		if err != nil {
			return fmt.Errorf("scan node: %w", err)
		}
		n.External = external != 0
		_ = json.Unmarshal([]byte(calleeJSON), &n.CalleeIDs)
		_ = json.Unmarshal([]byte(callerJSON), &n.CallerIDs)
		_ = json.Unmarshal([]byte(childJSON), &n.ChildIDs)
		g.AddNode(n)
	}
	return rows.Err()
}

func (s *Store) loadEdges(g *Graph) error {
	rows, err := s.db.Query(`SELECT from_id, to_id, kind FROM edges`)
	if err != nil {
		return fmt.Errorf("query edges: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		e := &Edge{}
		if err := rows.Scan(&e.From, &e.To, &e.Kind); err != nil {
			return fmt.Errorf("scan edge: %w", err)
		}
		g.AddEdge(e)
	}
	return rows.Err()
}

// Reset deletes all nodes and edges.
func (s *Store) Reset() error {
	if _, err := s.db.Exec(`DELETE FROM nodes`); err != nil {
		return err
	}
	_, err := s.db.Exec(`DELETE FROM edges`)
	return err
}

// AllNodes returns all nodes from the database directly.
func (s *Store) AllNodes() ([]*Node, error) {
	g := New()
	if err := s.loadNodes(g); err != nil {
		return nil, err
	}
	return g.Nodes(), nil
}

// AllEdges returns all edges from the database directly.
func (s *Store) AllEdges() ([]*Edge, error) {
	rows, err := s.db.Query(`SELECT from_id, to_id, kind FROM edges`)
	if err != nil {
		return nil, fmt.Errorf("query edges: %w", err)
	}
	defer rows.Close()
	var edges []*Edge
	for rows.Next() {
		e := &Edge{}
		if err := rows.Scan(&e.From, &e.To, &e.Kind); err != nil {
			return nil, fmt.Errorf("scan edge: %w", err)
		}
		edges = append(edges, e)
	}
	return edges, rows.Err()
}

// UpdateNodeAnalysis updates only the analysis fields of a node.
func (s *Store) UpdateNodeAnalysis(n *Node) error {
	_, err := s.db.Exec(`
		UPDATE nodes SET
			lint_count=?, lint_issues=?, race_count=?, race_issues=?,
			heap_allocs=?, stack_allocs=?, coverage=?, test_status=?,
			bench_ns_op=?, bench_b_op=?, bench_allocs_op=?,
			pprof_flat_pct=?, pprof_cum_pct=?, vet_issues=?
		WHERE id=?`,
		n.LintCount, n.LintIssues, n.RaceCount, n.RaceIssues,
		n.HeapAllocs, n.StackAllocs, n.Coverage, n.TestStatus,
		n.BenchNsOp, n.BenchBOp, n.BenchAllocsOp,
		n.PprofFlatPct, n.PprofCumPct, n.VetIssues,
		n.ID,
	)
	return err
}

// GetNodeByID fetches a single node by ID.
func (s *Store) GetNodeByID(id string) (*Node, error) {
	row := s.db.QueryRow(`SELECT id, kind, name, file, line, text, external,
		cyclomatic, cognitive, receiver, params, returns, parent_id,
		callee_ids, caller_ids, child_ids,
		lint_count, lint_issues, race_count, race_issues,
		heap_allocs, stack_allocs, coverage, test_status,
		bench_ns_op, bench_b_op, bench_allocs_op,
		pprof_flat_pct, pprof_cum_pct, vet_issues
		FROM nodes WHERE id=?`, id)

	n := &Node{}
	var external int
	var calleeJSON, callerJSON, childJSON string
	err := row.Scan(
		&n.ID, &n.Kind, &n.Name, &n.File, &n.Line, &n.Text, &external,
		&n.Cyclomatic, &n.Cognitive, &n.Receiver, &n.Params, &n.Returns, &n.ParentID,
		&calleeJSON, &callerJSON, &childJSON,
		&n.LintCount, &n.LintIssues, &n.RaceCount, &n.RaceIssues,
		&n.HeapAllocs, &n.StackAllocs, &n.Coverage, &n.TestStatus,
		&n.BenchNsOp, &n.BenchBOp, &n.BenchAllocsOp,
		&n.PprofFlatPct, &n.PprofCumPct, &n.VetIssues,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	n.External = external != 0
	_ = json.Unmarshal([]byte(calleeJSON), &n.CalleeIDs)
	_ = json.Unmarshal([]byte(callerJSON), &n.CallerIDs)
	_ = json.Unmarshal([]byte(childJSON), &n.ChildIDs)
	return n, nil
}
