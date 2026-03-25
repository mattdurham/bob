package analysis

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/mattdurham/bob/internal/firstmate/graph"
)

// RunTests runs go test with coverage and annotates nodes.
func RunTests(cwd string, store *graph.Store) (string, error) {
	coverProfile := "/tmp/first-mate-cover.out"
	output, err := runCommand(cwd, "go", "test", "-coverprofile="+coverProfile, "-covermode=atomic", "./...")

	var sb strings.Builder
	sb.WriteString("=== go test ===\n")
	sb.WriteString(output)
	sb.WriteByte('\n')

	if err != nil {
		fmt.Fprintf(&sb, "Tests failed: %v\n", err)
	} else {
		sb.WriteString("All tests passed.\n")
	}

	// Parse test results into maps for O(1) annotation.
	passed := make(map[string]bool)
	failed := make(map[string]bool)
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "--- PASS:"):
			passed[extractTestName(line)] = true
		case strings.HasPrefix(line, "--- FAIL:"):
			failed[extractTestName(line)] = true
		}
	}

	// Annotate test function nodes
	nodes, _ := store.AllNodes()
	for _, n := range nodes {
		if n.Kind != "function" {
			continue
		}
		if passed[n.Name] {
			n.TestStatus = "pass"
			store.UpdateNodeAnalysis(n)
		} else if failed[n.Name] {
			n.TestStatus = "fail"
			store.UpdateNodeAnalysis(n)
		}
	}

	return sb.String(), nil
}

func extractTestName(line string) string {
	// "--- PASS: TestFoo (0.00s)"
	parts := strings.Fields(line)
	if len(parts) >= 3 {
		return parts[2]
	}
	return ""
}
