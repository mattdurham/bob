package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// setupTestGitRepo creates a temporary git repo with a worktree for testing
func setupTestGitRepo(t *testing.T) (tmpDir string, worktreePath string, sm *StateManager) {
	tmpDir, err := os.MkdirTemp("", "bob-test-*")
	if err != nil {
		t.Fatal(err)
	}

	// Create a real git repo
	repoDir := filepath.Join(tmpDir, "test-repo")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Initialize git repo
	cmd := exec.Command("git", "init", "-b", "main")
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	// Create initial commit
	testFile := filepath.Join(repoDir, "README.md")
	if err := os.WriteFile(testFile, []byte("# Test"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = repoDir
	// Ignore error - git add may fail if no changes, which is OK
	if err := cmd.Run(); err != nil {
		t.Logf("git add returned error (may be OK if no changes): %v", err)
	}

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to set git user.email: %v", err)
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to set git user.name: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	// Create a worktree
	worktreePath = filepath.Join(tmpDir, "test-worktree")
	cmd = exec.Command("git", "worktree", "add", "-b", "test-branch", worktreePath)
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	// Create state manager
	stateDir := filepath.Join(tmpDir, "state")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		t.Fatal(err)
	}

	sm = &StateManager{
		stateDir:       stateDir,
		additionsCache: make(map[string]*AdditionsCache),
	}

	return tmpDir, worktreePath, sm
}

func TestRejoin(t *testing.T) {
	// Create temporary directory and git repo
	tmpDir, worktreePath, sm := setupTestGitRepo(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Register initial workflow
	_, err := sm.Register("work", worktreePath, "Initial task description", "", "", "")
	if err != nil {
		t.Fatalf("Failed to register workflow: %v", err)
	}

	// Test 1: Rejoin at same step with new task description
	t.Run("RejoinSameStepNewDescription", func(t *testing.T) {
		result, err := sm.Rejoin(worktreePath, "INIT", "Updated task description", false, "", "")
		if err != nil {
			t.Fatalf("Rejoin failed: %v", err)
		}

		if !result["rejoined"].(bool) {
			t.Error("Expected rejoined=true")
		}

		// Note: taskDescription no longer tracked in simplified state
		// Just verify rejoin succeeded - description update now ignored
	})

	// Test 2: Rejoin at different step with reset
	t.Run("RejoinDifferentStepWithReset", func(t *testing.T) {
		// First advance to next step (workflow auto-advances from INIT to WORKTREE)
		_, _ = sm.ReportProgress(worktreePath, "INIT", nil, "", "")

		// Get current step to verify advancement
		status, _ := sm.GetStatus(worktreePath, "", "")
		currentStep := status["currentStep"].(string)

		// Now rejoin at INIT with reset
		result, err := sm.Rejoin(worktreePath, "INIT", "", true, "", "")
		if err != nil {
			t.Fatalf("Rejoin failed: %v", err)
		}

		if result["fromStep"].(string) != currentStep {
			t.Errorf("Expected fromStep=%s, got %s", currentStep, result["fromStep"].(string))
		}

		if result["currentStep"].(string) != "INIT" {
			t.Errorf("Expected currentStep=INIT, got %s", result["currentStep"].(string))
		}

		// Note: Progress history no longer tracked in simplified state
		// Just verify the currentStep is correct
		status, _ = sm.GetStatus(worktreePath, "", "")
		if status["currentStep"].(string) != "INIT" {
			t.Errorf("Expected currentStep=INIT after rejoin, got %s", status["currentStep"].(string))
		}
	})

	// Test 3: Rejoin with invalid step
	t.Run("RejoinInvalidStep", func(t *testing.T) {
		_, err := sm.Rejoin(worktreePath, "INVALID_STEP", "", false, "", "")
		if err == nil {
			t.Error("Expected error for invalid step")
		}
	})

	// Test 4: Rejoin non-existent workflow
	t.Run("RejoinNonExistentWorkflow", func(t *testing.T) {
		_, err := sm.Rejoin("/nonexistent/path", "INIT", "", false, "", "")
		if err == nil {
			t.Error("Expected error for non-existent workflow")
		}
	})

	// Test 5: Verify rejoin history tracking (removed - history no longer tracked)
	t.Run("RejoinBasicFunctionality", func(t *testing.T) {
		// Just verify rejoin works without errors
		result, err := sm.Rejoin(worktreePath, "INIT", "", false, "", "")
		if err != nil {
			t.Errorf("Rejoin failed: %v", err)
		}
		if !result["rejoined"].(bool) {
			t.Error("Expected rejoined=true")
		}
	})
}

func TestReset(t *testing.T) {
	// Create temporary directory and git repo
	tmpDir, worktreePath, sm := setupTestGitRepo(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Test 1: Reset with archive
	t.Run("ResetWithArchive", func(t *testing.T) {
		// Register workflow
		_, err := sm.Register("work", worktreePath, "Test task", "", "", "")
		if err != nil {
			t.Fatalf("Failed to register workflow: %v", err)
		}

		// Reset with archive
		result, err := sm.Reset(worktreePath, true, "", "")
		if err != nil {
			t.Fatalf("Reset failed: %v", err)
		}

		if !result["reset"].(bool) {
			t.Error("Expected reset=true")
		}

		// Note: Archiving removed in simplified state - flag is ignored
		if result["archived"].(bool) {
			t.Error("Expected archived=false (archiving removed)")
		}

		// Verify state file was deleted
		workflowID := sm.worktreeToID(worktreePath, "", "")
		filename := workflowIDToFilename(workflowID)
		statePath := filepath.Join(sm.stateDir, filename)
		if _, err := os.Stat(statePath); !os.IsNotExist(err) {
			t.Error("State file was not deleted")
		}
	})

	// Test 2: Reset without archive
	t.Run("ResetWithoutArchive", func(t *testing.T) {
		// Register new workflow
		_, err := sm.Register("work", worktreePath, "Test task", "", "", "")
		if err != nil {
			t.Fatalf("Failed to register workflow: %v", err)
		}

		// Reset without archive
		result, err := sm.Reset(worktreePath, false, "", "")
		if err != nil {
			t.Fatalf("Reset failed: %v", err)
		}

		if !result["reset"].(bool) {
			t.Error("Expected reset=true")
		}

		if result["archived"].(bool) {
			t.Error("Expected archived=false")
		}

		// Verify no archive was created
		if _, ok := result["archivePath"]; ok {
			t.Error("Archive path should not be present when archived=false")
		}
	})

	// Test 3: Reset non-existent workflow
	t.Run("ResetNonExistentWorkflow", func(t *testing.T) {
		_, err := sm.Reset("/nonexistent/path", true, "", "")
		if err == nil {
			t.Error("Expected error for non-existent workflow")
		}
	})
}

func TestRejoinWithSessionAndAgent(t *testing.T) {
	// Create temporary directory and git repo
	tmpDir, worktreePath, sm := setupTestGitRepo(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()
	sessionID := "test-session"
	agentID := "test-agent"

	// Register workflow with session and agent IDs
	_, err := sm.Register("work", worktreePath, "Test task", "", sessionID, agentID)
	if err != nil {
		t.Fatalf("Failed to register workflow: %v", err)
	}

	// Rejoin with matching session and agent IDs
	result, err := sm.Rejoin(worktreePath, "INIT", "Updated description", false, sessionID, agentID)
	if err != nil {
		t.Fatalf("Rejoin failed: %v", err)
	}

	if !result["rejoined"].(bool) {
		t.Error("Expected rejoined=true")
	}

	// Note: taskDescription no longer stored in simplified state
	// Just verify status can be retrieved
	status, err := sm.GetStatus(worktreePath, sessionID, agentID)
	if err != nil {
		t.Errorf("GetStatus failed: %v", err)
	}
	if status["workflowId"] == nil {
		t.Error("Expected workflowId in status")
	}
}

func TestRejoinTimestamps(t *testing.T) {
	// Create temporary directory and git repo
	tmpDir, worktreePath, sm := setupTestGitRepo(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Register workflow
	_, err := sm.Register("work", worktreePath, "Test task", "", "", "")
	if err != nil {
		t.Fatalf("Failed to register workflow: %v", err)
	}

	// Wait a bit to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	// Rejoin
	result, err := sm.Rejoin(worktreePath, "INIT", "", false, "", "")
	if err != nil {
		t.Fatalf("Rejoin failed: %v", err)
	}

	// Verify timestamp was updated
	timestamp := result["timestamp"].(time.Time)
	// Note: StartedAt no longer tracked in simplified state
	// Just verify timestamp exists
	if timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}
}

func TestFileBasedRouting(t *testing.T) {
	// Create temporary directory and git repo
	tmpDir, worktreePath, sm := setupTestGitRepo(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Register workflow (using "work" workflow which has REVIEW as a checkpoint phase)
	_, err := sm.Register("work", worktreePath, "Test file-based routing", "", "", "")
	if err != nil {
		t.Fatalf("Failed to register workflow: %v", err)
	}

	// Create bots directory
	botsDir := filepath.Join(worktreePath, "bots")
	if err := os.MkdirAll(botsDir, 0755); err != nil {
		t.Fatalf("Failed to create bots directory: %v", err)
	}

	// Advance to REVIEW step (a checkpoint phase)
	// First advance through INIT -> WORKTREE -> BRAINSTORM -> PLAN -> EXECUTE -> TEST -> REVIEW
	// Create dummy findings files for checkpoint phases
	checkpointPhases := []string{"TEST"}
	for _, phase := range checkpointPhases {
		findingsFile := filepath.Join(botsDir, strings.ToLower(phase)+".md")
		if err := os.WriteFile(findingsFile, []byte("# Review\n\nNo issues found.\n"), 0644); err != nil {
			t.Fatalf("Failed to create findings file for %s: %v", phase, err)
		}
	}

	steps := []string{"INIT", "WORKTREE", "BRAINSTORM", "PLAN", "EXECUTE", "TEST"}
	for _, step := range steps {
		_, err := sm.ReportProgress(worktreePath, step, nil, "", "")
		if err != nil {
			t.Fatalf("Failed to advance through %s: %v", step, err)
		}
	}

	// Test 1: Missing file for checkpoint phase (should error with new enforcement)
	t.Run("MissingFileForCheckpoint", func(t *testing.T) {
		// Ensure no review.md file exists
		reviewFile := filepath.Join(botsDir, "review.md")
		_ = os.Remove(reviewFile)

		// Report progress on REVIEW step without findings file
		_, err := sm.ReportProgress(worktreePath, "REVIEW", nil, "", "")
		if err == nil {
			t.Error("Expected error for missing findings file in checkpoint phase")
		}
		if err != nil && !strings.Contains(err.Error(), "findings file not found") {
			t.Errorf("Expected 'findings file not found' error, got: %v", err)
		}
	})

	// Test 2: Empty/short file (should advance forward)
	t.Run("EmptyFile", func(t *testing.T) {
		reviewFile := filepath.Join(botsDir, "review.md")

		// Write empty file
		if err := os.WriteFile(reviewFile, []byte(""), 0644); err != nil {
			t.Fatalf("Failed to write empty review file: %v", err)
		}

		result, err := sm.ReportProgress(worktreePath, "REVIEW", nil, "", "")
		if err != nil {
			t.Errorf("Empty file should advance forward, got error: %v", err)
		}

		// Should have advanced to next step
		if result["currentStep"].(string) == "REVIEW" {
			t.Error("Expected to advance past REVIEW step with empty file")
		}
	})

	// Test 3: Short file (< minFindingsLength, should advance)
	t.Run("ShortFile", func(t *testing.T) {
		// Reset to REVIEW step
		state, _ := sm.loadState(sm.worktreeToID(worktreePath, "", ""))
		state.CurrentStep = "REVIEW"
		_ = sm.saveState(state)

		reviewFile := filepath.Join(botsDir, "review.md")

		// Write file with less than minFindingsLength (10 bytes)
		if err := os.WriteFile(reviewFile, []byte("OK"), 0644); err != nil {
			t.Fatalf("Failed to write short review file: %v", err)
		}

		result, err := sm.ReportProgress(worktreePath, "REVIEW", nil, "", "")
		if err != nil {
			t.Errorf("Short file should advance forward, got error: %v", err)
		}

		// Should have advanced
		if result["currentStep"].(string) == "REVIEW" {
			t.Error("Expected to advance past REVIEW step with short file")
		}
	})

	// Test 4: File with "no issues" content (should advance)
	t.Run("NoIssuesFound", func(t *testing.T) {
		// Reset to REVIEW step
		state, _ := sm.loadState(sm.worktreeToID(worktreePath, "", ""))
		state.CurrentStep = "REVIEW"
		_ = sm.saveState(state)

		reviewFile := filepath.Join(botsDir, "review.md")

		// Write file indicating no issues
		noIssuesContent := "# Code Review\n\nTotal Issues: 0\n\nNo issues found."
		if err := os.WriteFile(reviewFile, []byte(noIssuesContent), 0644); err != nil {
			t.Fatalf("Failed to write no-issues review file: %v", err)
		}

		result, err := sm.ReportProgress(worktreePath, "REVIEW", nil, "", "")
		if err != nil {
			t.Errorf("No-issues file should advance forward, got error: %v", err)
		}

		// Should have advanced (assuming fallback classification works)
		currentStep := result["currentStep"].(string)
		if currentStep == "REVIEW" {
			// This might stay at REVIEW if Claude API classifies it, which is OK
			// The important thing is no error occurred
			t.Logf("Stayed at REVIEW (classification may have detected false positive)")
		}
	})
}

func TestNonCheckpointPhaseRouting(t *testing.T) {
	// Create temporary directory and git repo
	tmpDir, worktreePath, sm := setupTestGitRepo(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Register workflow
	_, err := sm.Register("work", worktreePath, "Test non-checkpoint routing", "", "", "")
	if err != nil {
		t.Fatalf("Failed to register workflow: %v", err)
	}

	// Test: Non-checkpoint phases should auto-advance without checking files
	t.Run("NonCheckpointAutoAdvance", func(t *testing.T) {
		// Report progress on INIT (non-checkpoint phase) without any files
		result, err := sm.ReportProgress(worktreePath, "INIT", nil, "", "")
		if err != nil {
			t.Errorf("Non-checkpoint phase should advance without error: %v", err)
		}

		// Should have advanced
		if result["currentStep"].(string) == "INIT" {
			t.Error("Expected to advance past INIT step automatically")
		}
	})
}
