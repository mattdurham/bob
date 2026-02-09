package main

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed prompts/**/*.md
var promptFiles embed.FS

// LoadPrompt loads the prompt markdown file for a given workflow and step
func LoadPrompt(workflow, step string) (string, error) {
	// Map workflow and step to file path
	filename := fmt.Sprintf("prompts/%s/%s.md", workflow, stepToFilename(step))

	content, err := promptFiles.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("prompt not found for workflow=%s step=%s: %w", workflow, step, err)
	}

	return string(content), nil
}

// stepToFilename converts step name to filename
func stepToFilename(step string) string {
	// Map step names to numbered filenames
	stepFiles := map[string]string{
		// brainstorm workflow
		"INIT":       "01-init",
		"WORKTREE":   "02-worktree",
		"BRAINSTORM": "03-brainstorm",
		"PLAN":       "04-plan",
		"EXECUTE":    "05-execute",
		"TEST":       "06-test",
		"REVIEW":     "07-review",
		"COMMIT":     "08-commit",
		"MONITOR":    "09-monitor",
		"COMPLETE":   "10-complete",

		// code-review workflow (shares some steps with brainstorm)
		"FIX": "03-fix",

		// performance workflow
		"BENCHMARK": "02-benchmark",
		"ANALYZE":   "03-analyze",
		"OPTIMIZE":  "04-optimize",
		"VERIFY":    "05-verify",

		// explore workflow
		"DISCOVER": "01-discover",
		"DOCUMENT": "03-document",
	}

	if filename, ok := stepFiles[step]; ok {
		return filename
	}

	// Fallback: lowercase step name with dash
	// e.g., "MY_STEP" -> "my-step"
	return strings.ToLower(strings.ReplaceAll(step, "_", "-"))
}
