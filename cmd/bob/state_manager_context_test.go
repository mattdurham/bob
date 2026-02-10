package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateDynamicContext_NoFiles(t *testing.T) {
	sm := NewStateManager()
	tempDir := t.TempDir()

	// No bots/ directory exists
	context := sm.generateDynamicContext(tempDir, "work", "PLAN")

	if context != "" {
		t.Errorf("Expected empty context when no files exist, got: %s", context)
	}
}

func TestGenerateDynamicContext_EmptyFile(t *testing.T) {
	sm := NewStateManager()
	tempDir := t.TempDir()
	botsDir := filepath.Join(tempDir, "bots")
	_ = os.MkdirAll(botsDir, 0755)

	// Create empty file
	_ = os.WriteFile(filepath.Join(botsDir, "plan.md"), []byte(""), 0644)

	context := sm.generateDynamicContext(tempDir, "work", "PLAN")

	if context != "" {
		t.Errorf("Expected empty context for empty file, got: %s", context)
	}
}

func TestGenerateDynamicContext_SmallFile(t *testing.T) {
	sm := NewStateManager()
	tempDir := t.TempDir()
	botsDir := filepath.Join(tempDir, "bots")
	_ = os.MkdirAll(botsDir, 0755)

	// Create file with less than minFindingsLength bytes
	_ = os.WriteFile(filepath.Join(botsDir, "plan.md"), []byte("ok"), 0644)

	context := sm.generateDynamicContext(tempDir, "work", "PLAN")

	if context != "" {
		t.Errorf("Expected empty context for small file, got: %s", context)
	}
}

func TestGenerateDynamicContext_ReviewWithIssues(t *testing.T) {
	sm := NewStateManager()
	tempDir := t.TempDir()
	botsDir := filepath.Join(tempDir, "bots")
	_ = os.MkdirAll(botsDir, 0755)

	// Create review.md with issues
	reviewContent := `# Review Findings

## Issues Found
1. Missing error handling in state_manager.go:245
2. Unclosed file descriptor in state_manager.go:167
3. Unused variable in guidance.go:52

These issues need to be addressed before proceeding.
`
	_ = os.WriteFile(filepath.Join(botsDir, "review.md"), []byte(reviewContent), 0644)

	// PLAN step after REVIEW would indicate a loop back
	context := sm.generateDynamicContext(tempDir, "work", "PLAN")

	if context == "" {
		t.Fatal("Expected non-empty context with issues, got empty string")
	}

	// Context should mention the issues
	if !strings.Contains(context, "Missing error handling") {
		t.Errorf("Expected context to mention first issue, got: %s", context)
	}

	if !strings.Contains(context, "Unclosed file descriptor") {
		t.Errorf("Expected context to mention second issue, got: %s", context)
	}
}

func TestGenerateDynamicContext_CleanStep(t *testing.T) {
	sm := NewStateManager()
	tempDir := t.TempDir()
	botsDir := filepath.Join(tempDir, "bots")
	_ = os.MkdirAll(botsDir, 0755)

	// Create brainstorm.md (not a checkpoint phase)
	brainstormContent := `# Brainstorm

## Approaches Considered
- Approach 1
- Approach 2

All looks good, ready to proceed.
`
	_ = os.WriteFile(filepath.Join(botsDir, "brainstorm.md"), []byte(brainstormContent), 0644)

	context := sm.generateDynamicContext(tempDir, "work", "BRAINSTORM")

	// BRAINSTORM is not a checkpoint phase, so context might be empty or minimal
	// This test verifies the function doesn't crash
	_ = context
}

func TestGetGuidance_WithDynamicContext(t *testing.T) {
	sm := NewStateManager()
	tempDir := t.TempDir()
	worktreeDir := filepath.Join(tempDir, "worktree")
	_ = os.MkdirAll(worktreeDir, 0755)
	botsDir := filepath.Join(worktreeDir, "bots")
	_ = os.MkdirAll(botsDir, 0755)

	// Initialize git repo for main
	_ = os.Chdir(tempDir)
	_ = exec.Command("git", "init").Run()
	_ = exec.Command("git", "config", "user.email", "test@test.com").Run()
	_ = exec.Command("git", "config", "user.name", "Test").Run()
	_ = os.WriteFile("README.md", []byte("test"), 0644)
	_ = exec.Command("git", "add", "README.md").Run()
	_ = exec.Command("git", "commit", "-m", "init").Run()

	// Create worktree
	_ = exec.Command("git", "worktree", "add", "-b", "test-branch", worktreeDir).Run()

	// Register workflow on worktree
	_, err := sm.Register("work", worktreeDir, "Test task", "", "", "")
	if err != nil {
		t.Fatalf("Failed to register workflow: %v", err)
	}

	// Create plan.md with content
	planContent := `# Issues to Address
1. Fix error handling
2. Close file descriptors
`
	_ = os.WriteFile(filepath.Join(botsDir, "plan.md"), []byte(planContent), 0644)

	// Move to PLAN step
	state, _ := sm.loadState(sm.worktreeToID(worktreeDir, "", ""))
	state.CurrentStep = "PLAN"
	_ = sm.saveState(state)

	// Get guidance
	result, err := sm.GetGuidance(worktreeDir, "", "")
	if err != nil {
		t.Fatalf("GetGuidance failed: %v", err)
	}

	prompt, ok := result["prompt"].(string)
	if !ok {
		t.Fatal("Expected prompt to be a string")
	}

	// Verify base prompt is present
	if !strings.Contains(prompt, "PLAN Phase") {
		t.Error("Expected base prompt to contain 'PLAN Phase'")
	}

	// Verify dynamic context is appended
	if !strings.Contains(prompt, "## Current Context") {
		t.Error("Expected prompt to contain '## Current Context' section")
	}

	// Verify issues are mentioned
	if !strings.Contains(prompt, "Fix error handling") {
		t.Error("Expected context to mention issues from plan.md")
	}
}

// Helper to run git commands for tests
