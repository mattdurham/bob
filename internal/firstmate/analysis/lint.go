package analysis

import (
	"fmt"
	"strings"

	"github.com/mattdurham/bob/internal/firstmate/graph"
)

// RunLint runs golangci-lint in the given directory and annotates nodes.
func RunLint(cwd string, store *graph.Store) (string, error) {
	output, err := runCommand(cwd, "golangci-lint", "run", "./...")

	var sb strings.Builder
	sb.WriteString("=== golangci-lint ===\n")

	if len(output) == 0 && err == nil {
		sb.WriteString("No issues found.\n")
		return sb.String(), nil
	}

	// Parse lines like: file.go:line:col: message (linter)
	type issue struct {
		file string
		msg  string
	}
	var issues []issue
	for line := range strings.SplitSeq(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "level=") {
			continue
		}
		parts := strings.SplitN(line, ":", 4)
		if len(parts) >= 3 {
			issues = append(issues, issue{
				file: parts[0],
				msg:  parts[1] + ":" + strings.TrimSpace(strings.Join(parts[2:], ":")),
			})
		}
	}

	// Group by file
	byFile := make(map[string][]string)
	for _, iss := range issues {
		byFile[iss.file] = append(byFile[iss.file], iss.msg)
	}

	// Annotate nodes
	nodes, _ := store.AllNodes()
	for _, n := range nodes {
		if fileIssues, ok := byFile[n.File]; ok {
			n.LintCount = len(fileIssues)
			n.LintIssues = strings.Join(fileIssues, ";")
			store.UpdateNodeAnalysis(n)
		}
	}

	fmt.Fprintf(&sb, "Found %d issues.\n", len(issues))
	if len(output) > 0 {
		sb.WriteString(output)
		sb.WriteByte('\n')
	}

	return sb.String(), nil
}
