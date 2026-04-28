package mdparser

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// Section is a single extractable chunk from a spec file.
type Section struct {
	ID   string
	Text string
}

var (
	h2Re        = regexp.MustCompile(`^## `)
	numberedH2  = regexp.MustCompile(`^(\d+)\.\s+(.+)`)
	invariantRe = regexp.MustCompile(`^(\d+)\.\s+(.+)`)
	nonAlnum    = regexp.MustCompile(`[^a-z0-9]+`)
)

var specFilenameSet = map[string]bool{
	"SPECS.md":      true,
	"NOTES.md":      true,
	"BENCHMARKS.md": true,
	"TESTS.md":      true,
	"CLAUDE.md":     true,
}

// Parse extracts sections from content. filename is used to pick the parsing mode.
func Parse(filename, content string) []Section {
	base := filepath.Base(filename)
	if !isSpecFilename(base) {
		return nil
	}
	if base == "CLAUDE.md" {
		return parseInvariants(content)
	}
	return parseH2Sections(content)
}

func isSpecFilename(name string) bool {
	return specFilenameSet[name]
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = nonAlnum.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

func parseH2Sections(content string) []Section {
	lines := strings.Split(content, "\n")
	var sections []Section
	var currentHeading string
	var currentBody []string
	inSection := false

	flush := func() {
		if !inSection {
			return
		}
		text := currentHeading + "\n" + strings.Join(currentBody, "\n")
		text = strings.TrimSpace(text)
		if len(text) < 20 {
			return
		}
		id := buildH2ID(currentHeading)
		sections = append(sections, Section{ID: id, Text: text})
	}

	for _, line := range lines {
		if h2Re.MatchString(line) {
			flush()
			currentHeading = strings.TrimPrefix(line, "## ")
			currentBody = nil
			inSection = true
		} else if inSection {
			currentBody = append(currentBody, line)
		}
		// Lines before first ## (including H1) are preamble — ignored
	}
	flush()
	return sections
}

func buildH2ID(heading string) string {
	if m := numberedH2.FindStringSubmatch(heading); m != nil {
		num := m[1]
		rest := slugify(m[2])
		return fmt.Sprintf("section-%s-%s", num, rest)
	}
	return "section-" + slugify(heading)
}

func parseInvariants(content string) []Section {
	lines := strings.Split(content, "\n")
	var sections []Section
	var currentNum string
	var currentLines []string

	flush := func() {
		if currentNum == "" {
			return
		}
		text := strings.TrimSpace(strings.Join(currentLines, "\n"))
		if len(text) < 20 {
			return
		}
		sections = append(sections, Section{
			ID:   "invariant-" + currentNum,
			Text: text,
		})
	}

	for _, line := range lines {
		if m := invariantRe.FindStringSubmatch(line); m != nil {
			flush()
			currentNum = m[1]
			currentLines = []string{m[2]}
		} else if currentNum != "" {
			currentLines = append(currentLines, line)
		}
	}
	flush()
	return sections
}
