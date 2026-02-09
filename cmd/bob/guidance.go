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
	filename := fmt.Sprintf("prompts/%s/%s.md", workflow, stepToFilename(workflow, step))

	content, err := promptFiles.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("prompt not found for workflow=%s step=%s: %w", workflow, step, err)
	}

	return string(content), nil
}

// stepToFilename converts step name to filename based on workflow
func stepToFilename(workflow, step string) string {
	// Workflow-specific step mappings
	workflowSteps := map[string]map[string]string{
		"brainstorm": {
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
		},
		"code-review": {
			"INIT":     "01-init",
			"REVIEW":   "02-review",
			"FIX":      "03-fix",
			"TEST":     "04-test",
			"COMMIT":   "05-commit",
			"MONITOR":  "06-monitor",
			"COMPLETE": "07-complete",
		},
		"explore": {
			"INIT":     "01-init",
			"WORKTREE": "02-worktree",
			"DISCOVER": "03-discover",
			"ANALYZE":  "04-analyze",
			"DOCUMENT": "05-document",
			"COMPLETE": "06-complete",
		},
		"performance": {
			"INIT":      "01-init",
			"BENCHMARK": "02-benchmark",
			"ANALYZE":   "03-analyze",
			"OPTIMIZE":  "04-optimize",
			"VERIFY":    "05-verify",
			"COMMIT":    "06-commit",
			"MONITOR":   "07-monitor",
			"COMPLETE":  "08-complete",
		},
		"test-bob": {
			"INIT":     "01-init",
			"PROMPT":   "02-prompt",
			"COMPLETE": "03-complete",
		},
	}

	// Check workflow-specific mapping first
	if stepMap, ok := workflowSteps[workflow]; ok {
		if filename, ok := stepMap[step]; ok {
			return filename
		}
	}

	// Fallback: lowercase step name with dash
	// e.g., "MY_STEP" -> "my-step"
	return strings.ToLower(strings.ReplaceAll(step, "_", "-"))
}
