package graph

import "sync"

// Graph holds nodes and edges in memory.
type Graph struct {
	mu      sync.RWMutex
	nodes   map[string]*Node
	edges   []*Edge
	edgeSet map[string]struct{} // "from:to:kind" → dedup key
}

// New creates an empty Graph.
func New() *Graph {
	return &Graph{
		nodes:   make(map[string]*Node),
		edgeSet: make(map[string]struct{}),
	}
}

// AddNode adds or updates a node.
func (g *Graph) AddNode(n *Node) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.nodes[n.ID] = n
}

// GetNode returns a node by ID.
func (g *Graph) GetNode(id string) (*Node, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	n, ok := g.nodes[id]
	return n, ok
}

// Nodes returns all nodes as a slice.
func (g *Graph) Nodes() []*Node {
	g.mu.RLock()
	defer g.mu.RUnlock()
	out := make([]*Node, 0, len(g.nodes))
	for _, n := range g.nodes {
		out = append(out, n)
	}
	return out
}

// AddEdge adds an edge if not already present.
func (g *Graph) AddEdge(e *Edge) {
	g.mu.Lock()
	defer g.mu.Unlock()
	key := e.From + ":" + e.To + ":" + e.Kind
	if _, exists := g.edgeSet[key]; exists {
		return
	}
	g.edgeSet[key] = struct{}{}
	g.edges = append(g.edges, e)
}

// Edges returns all edges.
func (g *Graph) Edges() []*Edge {
	g.mu.RLock()
	defer g.mu.RUnlock()
	out := make([]*Edge, len(g.edges))
	copy(out, g.edges)
	return out
}

// Reset clears all nodes and edges.
func (g *Graph) Reset() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.nodes = make(map[string]*Node)
	g.edges = nil
	g.edgeSet = make(map[string]struct{})
}

// Merge merges another graph into this one.
func (g *Graph) Merge(other *Graph) {
	other.mu.RLock()
	defer other.mu.RUnlock()
	g.mu.Lock()
	defer g.mu.Unlock()
	for id, n := range other.nodes {
		if existing, ok := g.nodes[id]; ok {
			// Merge relationship slices
			for _, c := range n.CalleeIDs {
				existing.AddCallee(c)
			}
			for _, c := range n.CallerIDs {
				existing.AddCaller(c)
			}
			for _, c := range n.ChildIDs {
				existing.AddChild(c)
			}
		} else {
			g.nodes[id] = n
		}
	}
	for _, e := range other.edges {
		key := e.From + ":" + e.To + ":" + e.Kind
		if _, exists := g.edgeSet[key]; !exists {
			g.edgeSet[key] = struct{}{}
			g.edges = append(g.edges, e)
		}
	}
}
