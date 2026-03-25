package analysis

import (
	"fmt"
	"strings"

	"github.com/mattdurham/bob/internal/firstmate/graph"
)

// RunEscape runs escape analysis and annotates nodes.
func RunEscape(cwd string, store *graph.Store) (string, error) {
	output, err := runCommand(cwd, "go", "build", "-gcflags=-m -m", "./...")

	var sb strings.Builder
	sb.WriteString("=== Escape Analysis ===\n")

	if err != nil && len(output) == 0 {
		fmt.Fprintf(&sb, "Build failed: %v\n", err)
		return sb.String(), nil
	}

	// Parse lines like:
	// ./file.go:line:col: func does not escape
	// ./file.go:line:col: moved to heap: var
	heapByFile := make(map[string]int)
	stackByFile := make(map[string]int)

	for line := range strings.SplitSeq(output, "\n") {
		line = strings.TrimSpace(line)
		line, _ = strings.CutPrefix(line, "./")
		parts := strings.SplitN(line, ":", 4)
		if len(parts) < 4 {
			continue
		}
		file := parts[0]
		msg := strings.TrimSpace(strings.Join(parts[3:], ":"))
		if strings.Contains(msg, "moved to heap") || strings.Contains(msg, "escapes to heap") {
			heapByFile[file]++
		} else {
			stackByFile[file]++
		}
	}

	// Annotate nodes
	nodes, _ := store.AllNodes()
	for _, n := range nodes {
		h, s := heapByFile[n.File], stackByFile[n.File]
		if h > 0 || s > 0 {
			n.HeapAllocs = h
			n.StackAllocs = s
			store.UpdateNodeAnalysis(n)
		}
	}

	heapTotal := 0
	for _, v := range heapByFile {
		heapTotal += v
	}
	stackTotal := 0
	for _, v := range stackByFile {
		stackTotal += v
	}
	fmt.Fprintf(&sb, "Total heap escapes: %d, stack: %d\n", heapTotal, stackTotal)

	return sb.String(), nil
}
