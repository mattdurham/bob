package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRejoin(t *testing.T) {
	// Create temporary state directory
	tmpDir, err := os.MkdirTemp("", "bob-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	sm := &StateManager{
		stateDir:       tmpDir,
		additionsCache: make(map[string]*AdditionsCache),
	}

	worktreePath := "/test/worktree"

	// Register initial workflow
	_, err = sm.Register("brainstorm", worktreePath, "Initial task description", "", "", "")
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
	// Create temporary state directory
	tmpDir, err := os.MkdirTemp("", "bob-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	sm := &StateManager{
		stateDir:       tmpDir,
		additionsCache: make(map[string]*AdditionsCache),
	}

	worktreePath := "/test/worktree"

	// Test 1: Reset with archive
	t.Run("ResetWithArchive", func(t *testing.T) {
		// Register workflow
		_, err := sm.Register("brainstorm", worktreePath, "Test task", "", "", "")
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
		_, err := sm.Register("brainstorm", worktreePath, "Test task", "", "", "")
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
	// Create temporary state directory
	tmpDir, err := os.MkdirTemp("", "bob-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	sm := &StateManager{
		stateDir:       tmpDir,
		additionsCache: make(map[string]*AdditionsCache),
	}

	worktreePath := "/test/worktree"
	sessionID := "test-session"
	agentID := "test-agent"

	// Register workflow with session and agent IDs
	_, err = sm.Register("brainstorm", worktreePath, "Test task", "", sessionID, agentID)
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
	// Create temporary state directory
	tmpDir, err := os.MkdirTemp("", "bob-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	sm := &StateManager{
		stateDir:       tmpDir,
		additionsCache: make(map[string]*AdditionsCache),
	}

	worktreePath := "/test/worktree"

	// Register workflow
	_, err = sm.Register("brainstorm", worktreePath, "Test task", "", "", "")
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
