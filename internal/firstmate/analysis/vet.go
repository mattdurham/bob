package analysis

import (
	"fmt"
	"strings"

	"github.com/mattdurham/bob/internal/firstmate/graph"
)

// RunVet runs go vet in the given directory and annotates nodes in the store.
func RunVet(cwd string, store *graph.Store) (string, error) {
	output, err := runCommand(cwd, "go", "vet", "./...")

	var sb strings.Builder
	sb.WriteString("=== go vet ===\n")

	if len(output) == 0 && err == nil {
		sb.WriteString("No issues found.\n")
		return sb.String(), nil
	}

	// Parse lines like: file.go:line:col: message
	issuesByFile := make(map[string][]string)
	for line := range strings.SplitSeq(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 4)
		if len(parts) >= 3 {
			issuesByFile[parts[0]] = append(issuesByFile[parts[0]], line)
		}
	}

	// Annotate nodes
	nodes, _ := store.AllNodes()
	for _, n := range nodes {
		if issues, ok := issuesByFile[n.File]; ok {
			n.VetIssues = strings.Join(issues, ";")
			store.UpdateNodeAnalysis(n)
		}
	}

	if err != nil {
		fmt.Fprintf(&sb, "go vet reported issues:\n%s\n", output)
	} else {
		sb.WriteString("No issues found.\n")
	}

	return sb.String(), nil
}
