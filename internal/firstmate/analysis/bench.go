package analysis

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/mattdurham/bob/internal/firstmate/graph"
)

// RunBench runs benchmarks and annotates nodes.
func RunBench(cwd string, store *graph.Store) (string, error) {
	output, err := runCommand(cwd, "go", "test", "-bench=.", "-benchmem", "./...")

	var sb strings.Builder
	sb.WriteString("=== go test -bench ===\n")
	sb.WriteString(output)
	sb.WriteByte('\n')

	if err != nil {
		fmt.Fprintf(&sb, "Benchmark run failed: %v\n", err)
	}

	// Parse benchmark results
	// BenchmarkFoo-8   1000000   1234 ns/op   256 B/op   3 allocs/op
	type benchResult struct {
		nsOp     float64
		bOp      float64
		allocsOp float64
	}
	resultMap := make(map[string]benchResult)

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "Benchmark") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		name := fields[0]
		// Strip -N suffix
		if idx := strings.LastIndex(name, "-"); idx > 0 {
			name = name[:idx]
		}
		r := benchResult{}
		for i := 0; i < len(fields)-1; i++ {
			switch fields[i+1] {
			case "ns/op":
				r.nsOp, _ = strconv.ParseFloat(fields[i], 64)
			case "B/op":
				r.bOp, _ = strconv.ParseFloat(fields[i], 64)
			case "allocs/op":
				r.allocsOp, _ = strconv.ParseFloat(fields[i], 64)
			}
		}
		resultMap[name] = r
	}

	// Annotate matching nodes using map lookup.
	nodes, _ := store.AllNodes()
	for _, n := range nodes {
		if n.Kind != "function" {
			continue
		}
		if r, ok := resultMap[n.Name]; ok {
			n.BenchNsOp = r.nsOp
			n.BenchBOp = r.bOp
			n.BenchAllocsOp = r.allocsOp
			store.UpdateNodeAnalysis(n)
		}
	}

	return sb.String(), nil
}
