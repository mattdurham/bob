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
		"work": {
			"INIT":       "01-init",
			"PROMPT":     "02-prompt",
			"WORKTREE":   "03-worktree",
			"BRAINSTORM": "04-brainstorm",
			"PLAN":       "05-plan",
			"EXECUTE":    "06-execute",
			"TEST":       "07-test",
			"REVIEW":     "08-review",
			"COMMIT":     "09-commit",
			"MONITOR":    "10-monitor",
			"COMPLETE":   "11-complete",
		},
		"code-review": {
			"INIT":     "01-init",
			"PROMPT":   "02-prompt",
			"WORKTREE": "03-worktree",
			"REVIEW":   "04-review",
			"FIX":      "05-fix",
			"TEST":     "06-test",
			"COMMIT":   "07-commit",
			"MONITOR":  "08-monitor",
			"COMPLETE": "09-complete",
		},
		"explore": {
			"INIT":     "01-init",
			"PROMPT":   "02-prompt",
			"WORKTREE": "03-worktree",
			"DISCOVER": "04-discover",
			"ANALYZE":  "05-analyze",
			"DOCUMENT": "06-document",
			"COMPLETE": "07-complete",
		},
		"performance": {
			"INIT":      "01-init",
			"PROMPT":    "02-prompt",
			"WORKTREE":  "03-worktree",
			"BENCHMARK": "04-benchmark",
			"ANALYZE":   "05-analyze",
			"OPTIMIZE":  "06-optimize",
			"VERIFY":    "07-verify",
			"COMMIT":    "08-commit",
			"MONITOR":   "09-monitor",
			"COMPLETE":  "10-complete",
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
