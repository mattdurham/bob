package main

import (
	"os"
	"os/exec"
	"path/filepath"
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
