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
	_ = cmd.Run()

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = repoDir
	_ = cmd.Run()

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = repoDir
	_ = cmd.Run()

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

		// Verify task description was updated
		status, _ := sm.GetStatus(worktreePath, "", "")
		if status["taskDescription"].(string) != "Updated task description" {
			t.Error("Task description was not updated")
		}
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

		// Verify progress history was reset
		status, _ = sm.GetStatus(worktreePath, "", "")
		history := status["progressHistory"].([]ProgressEntry)

		// Should only have entries up to and including INIT (plus the rejoin entry)
		foundNonInitStep := false
		for _, entry := range history {
			if entry.Step != "INIT" {
				// Check if this is the rejoin entry by looking at metadata
				if metadata, ok := entry.Metadata["rejoin"].(bool); !ok || !metadata {
					foundNonInitStep = true
					break
				}
			}
		}
		if foundNonInitStep {
			t.Error("Found non-INIT steps in history after reset (excluding rejoin entry)")
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

	// Test 5: Verify rejoin history tracking
	t.Run("RejoinHistoryTracking", func(t *testing.T) {
		workflowID := sm.worktreeToID(worktreePath, "", "")
		state, _ := sm.loadState(workflowID)

		if len(state.RejoinHistory) == 0 {
			t.Error("Expected rejoin history to be populated")
		}

		// Verify last rejoin event
		lastRejoin := state.RejoinHistory[len(state.RejoinHistory)-1]
		if lastRejoin.ToStep != "INIT" {
			t.Errorf("Expected last rejoin to INIT, got %s", lastRejoin.ToStep)
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

		if !result["archived"].(bool) {
			t.Error("Expected archived=true")
		}

		// Verify archive file exists
		archivePath := result["archivePath"].(string)
		if _, err := os.Stat(archivePath); os.IsNotExist(err) {
			t.Error("Archive file was not created")
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

	// Verify task description was updated
	status, _ := sm.GetStatus(worktreePath, sessionID, agentID)
	if status["taskDescription"].(string) != "Updated description" {
		t.Error("Task description was not updated")
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
	workflowID := sm.worktreeToID(worktreePath, "", "")
	state, _ := sm.loadState(workflowID)

	if !timestamp.After(state.StartedAt) {
		t.Error("Expected rejoin timestamp to be after start time")
	}
}
