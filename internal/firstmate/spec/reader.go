package spec

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// KindToFilename maps spec kinds to their filenames.
var KindToFilename = map[string]string{
	"SPECS":      "SPECS.md",
	"NOTES":      "NOTES.md",
	"TESTS":      "TESTS.md",
	"BENCHMARKS": "BENCHMARKS.md",
}

// SpecFile represents a found spec file.
type SpecFile struct {
	Path    string
	Kind    string
	Content string
}

// FindAll walks root and returns all spec files.
func FindAll(root string) ([]*SpecFile, error) {
	var specs []*SpecFile
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == "vendor" || name == "testdata" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		base := filepath.Base(path)
		for kind, filename := range KindToFilename {
			if base == filename {
				content, err := os.ReadFile(path)
				if err != nil {
					return nil
				}
				rel, _ := filepath.Rel(root, path)
				specs = append(specs, &SpecFile{
					Path:    rel,
					Kind:    kind,
					Content: string(content),
				})
				break
			}
		}
		return nil
	})
	return specs, err
}

// ReadByKind returns content of all spec files of a given kind.
// If pattern is set, only returns files whose content contains the pattern.
func ReadByKind(root, kind, pattern string) (string, error) {
	specs, err := FindAll(root)
	if err != nil {
		return "", fmt.Errorf("find specs: %w", err)
	}

	var sb strings.Builder
	count := 0
	for _, s := range specs {
		if s.Kind != kind {
			continue
		}
		if pattern != "" && !strings.Contains(strings.ToLower(s.Content), strings.ToLower(pattern)) {
			continue
		}
		fmt.Fprintf(&sb, "=== %s ===\n", s.Path)
		sb.WriteString(s.Content)
		sb.WriteString("\n\n")
		count++
	}
	if count == 0 {
		return fmt.Sprintf("No %s files found", kind), nil
	}
	return sb.String(), nil
}

// SearchResult holds a spec file that matched a query, with the matching line numbers.
type SearchResult struct {
	Path         string `json:"path"`
	Kind         string `json:"kind"`
	Content      string `json:"content"`
	MatchedLines []int  `json:"matched_lines"`
}

// CodeRef is a reference to a query match inside a Go source file.
type CodeRef struct {
	Path    string `json:"path"`
	Line    int    `json:"line"`
	Col     int    `json:"col"`
	Excerpt string `json:"excerpt"`
}

// FindResult is the combined result of a spec + code search.
type FindResult struct {
	Specs    []*SearchResult `json:"specs"`
	CodeRefs []*CodeRef      `json:"code_refs"`
}

// SearchAll searches all spec files and Go source files for query, returning a FindResult.
func SearchAll(root, query string) (*FindResult, error) {
	specs, err := FindAll(root)
	if err != nil {
		return nil, fmt.Errorf("find specs: %w", err)
	}

	queryLower := strings.ToLower(query)
	result := &FindResult{}

	for _, s := range specs {
		var matchedLines []int
		lineNum := 0
		for line := range strings.SplitSeq(s.Content, "\n") {
			lineNum++
			if strings.Contains(strings.ToLower(line), queryLower) {
				matchedLines = append(matchedLines, lineNum)
			}
		}
		if len(matchedLines) > 0 {
			result.Specs = append(result.Specs, &SearchResult{
				Path:         s.Path,
				Kind:         s.Kind,
				Content:      s.Content,
				MatchedLines: matchedLines,
			})
		}
	}

	// Search Go source files for code references.
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == "vendor" || name == "testdata" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		lineNum := 0
		for line := range strings.SplitSeq(string(content), "\n") {
			lineNum++
			lineLower := strings.ToLower(line)
			col := strings.Index(lineLower, queryLower)
			if col >= 0 {
				result.CodeRefs = append(result.CodeRefs, &CodeRef{
					Path:    rel,
					Line:    lineNum,
					Col:     col + 1, // 1-based
					Excerpt: strings.TrimSpace(line),
				})
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk source: %w", err)
	}

	return result, nil
}

// GetByID searches all spec files for a spec ID like "SPEC-001".
func GetByID(root, id string) (string, error) {
	specs, err := FindAll(root)
	if err != nil {
		return "", fmt.Errorf("find specs: %w", err)
	}

	var sb strings.Builder
	idLower := strings.ToLower(id)
	found := 0

	for _, s := range specs {
		lines := strings.Split(s.Content, "\n")
		for lineNum, line := range lines {
			if strings.Contains(strings.ToLower(line), idLower) {
				// Return surrounding context (3 lines)
				start := lineNum - 1
				if start < 0 {
					start = 0
				}
				end := lineNum + 3
				if end > len(lines) {
					end = len(lines)
				}
				fmt.Fprintf(&sb, "=== %s:%d ===\n", s.Path, lineNum+1)
				for _, ctxLine := range lines[start:end] {
					sb.WriteString(ctxLine + "\n")
				}
				sb.WriteString("\n")
				found++
			}
		}
	}

	if found == 0 {
		return fmt.Sprintf("No spec with ID %q found", id), nil
	}
	return sb.String(), nil
}

// ListAll returns all spec file paths.
func ListAll(root string) (string, error) {
	specs, err := FindAll(root)
	if err != nil {
		return "", fmt.Errorf("find specs: %w", err)
	}
	if len(specs) == 0 {
		return "No spec files found.", nil
	}
	var sb strings.Builder
	for _, s := range specs {
		fmt.Fprintf(&sb, "[%s] %s\n", s.Kind, s.Path)
	}
	return sb.String(), nil
}
