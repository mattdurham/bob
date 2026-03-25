package analysis

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/mattdurham/bob/internal/firstmate/graph"
)

// RunProfile parses a pprof profile file and annotates nodes.
func RunProfile(profileFile string, store *graph.Store) (string, error) {
	output, err := runCommand("", "go", "tool", "pprof", "-text", profileFile)

	var sb strings.Builder
	sb.WriteString("=== pprof analysis ===\n")

	if err != nil {
		fmt.Fprintf(&sb, "pprof failed: %v\n%s\n", err, output)
		return sb.String(), nil
	}

	type pprofEntry struct {
		flat     float64
		cum      float64
		funcName string
	}
	var entries []pprofEntry

	scanner := bufio.NewScanner(strings.NewReader(output))
	lineCount := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineCount++
		if lineCount <= 5 { // skip header lines
			sb.WriteString(line + "\n")
			continue
		}
		// Format: flat  flat%  sum%  cum  cum%  funcname
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}
		entries = append(entries, pprofEntry{
			flat:     parsePct(fields[1]),
			cum:      parsePct(fields[4]),
			funcName: fields[5],
		})
	}

	// Return top 20
	top := entries
	if len(top) > 20 {
		top = top[:20]
	}
	sb.WriteString("\nTop functions by CPU:\n")
	for i, e := range top {
		fmt.Fprintf(&sb, "%3d. flat=%.2f%% cum=%.2f%% %s\n", i+1, e.flat, e.cum, e.funcName)
	}

	// Build lookup map for O(1) annotation.
	entryMap := make(map[string]pprofEntry, len(entries))
	for _, e := range entries {
		entryMap[e.funcName] = e
	}

	nodes, _ := store.AllNodes()
	for _, n := range nodes {
		// Try "pkg.FuncName" then bare name.
		key := n.ID
		e, ok := entryMap[key]
		if !ok {
			e, ok = entryMap[n.Name]
		}
		if ok {
			n.PprofFlatPct = e.flat
			n.PprofCumPct = e.cum
			store.UpdateNodeAnalysis(n)
		}
	}

	return sb.String(), nil
}

func parsePct(s string) float64 {
	s = strings.TrimSuffix(s, "%")
	v, _ := strconv.ParseFloat(s, 64)
	return v
}
