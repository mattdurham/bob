package graph

import (
	"encoding/json"
	"fmt"
	"time"
)

// Snapshot stores a named graph snapshot.
type Snapshot struct {
	Name      string
	CreatedAt int64
	Nodes     []*Node
	Edges     []*Edge
}

// SaveSnapshot saves the current graph state as a named snapshot.
func (s *Store) SaveSnapshot(name string, nodes []*Node, edges []*Edge) error {
	nodesJSON, err := json.Marshal(nodes)
	if err != nil {
		return fmt.Errorf("marshal nodes: %w", err)
	}
	edgesJSON, err := json.Marshal(edges)
	if err != nil {
		return fmt.Errorf("marshal edges: %w", err)
	}
	_, err = s.db.Exec(`
		INSERT OR REPLACE INTO snapshots (name, created_at, nodes_json, edges_json)
		VALUES (?, ?, ?, ?)`,
		name, time.Now().Unix(), string(nodesJSON), string(edgesJSON),
	)
	return err
}

// LoadLatestSnapshot loads the most recently created snapshot.
func (s *Store) LoadLatestSnapshot() (*Snapshot, error) {
	row := s.db.QueryRow(`SELECT name, created_at, nodes_json, edges_json FROM snapshots ORDER BY created_at DESC LIMIT 1`)
	snap := &Snapshot{}
	var nodesJSON, edgesJSON string
	err := row.Scan(&snap.Name, &snap.CreatedAt, &nodesJSON, &edgesJSON)
	if err != nil {
		return nil, fmt.Errorf("no snapshots found: %w", err)
	}
	if err := json.Unmarshal([]byte(nodesJSON), &snap.Nodes); err != nil {
		return nil, fmt.Errorf("unmarshal nodes: %w", err)
	}
	if err := json.Unmarshal([]byte(edgesJSON), &snap.Edges); err != nil {
		return nil, fmt.Errorf("unmarshal edges: %w", err)
	}
	return snap, nil
}

// DiffResult holds the result of comparing the current graph with a snapshot.
type DiffResult struct {
	Added   []string
	Removed []string
	Changed []string
}

// DiffWithCurrent compares a snapshot against the current store and returns differences.
func (s *Store) DiffWithCurrent(snap *Snapshot) (*DiffResult, error) {
	currentNodes, err := s.AllNodes()
	if err != nil {
		return nil, err
	}

	currentMap := make(map[string]*Node, len(currentNodes))
	for _, n := range currentNodes {
		currentMap[n.ID] = n
	}

	snapMap := make(map[string]*Node, len(snap.Nodes))
	for _, n := range snap.Nodes {
		snapMap[n.ID] = n
	}

	result := &DiffResult{}

	// Find added and changed
	for id, cur := range currentMap {
		if snap, ok := snapMap[id]; !ok {
			result.Added = append(result.Added, id)
		} else if cur.Kind != snap.Kind || cur.File != snap.File || cur.Line != snap.Line {
			result.Changed = append(result.Changed, id)
		}
	}

	// Find removed
	for id := range snapMap {
		if _, ok := currentMap[id]; !ok {
			result.Removed = append(result.Removed, id)
		}
	}

	return result, nil
}
