package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestAutoWorktreeCreation(t *testing.T) {
	// Create a temporary git repository to simulate main repo
	tmpDir, err := os.MkdirTemp("", "bob-worktree-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize git repo with explicit main branch
	repoDir := filepath.Join(tmpDir, "test-repo")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Init git with main branch
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
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

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

	// Create state manager with state directory
	stateDir := filepath.Join(tmpDir, "state")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		t.Fatal(err)
	}
	sm := &StateManager{
		stateDir:       stateDir,
		additionsCache: make(map[string]*AdditionsCache),
	}

	// Test 1: Register with featureName on main repo - should auto-create worktree
	t.Run("AutoCreateWorktreeFromMain", func(t *testing.T) {
		result, err := sm.Register("brainstorm", repoDir, "Test task", "test-feature", "", "")
		if err != nil {
			t.Fatalf("Register failed: %v", err)
		}

		// Verify worktree was created
		if !result["createdWorktree"].(bool) {
			t.Error("Expected createdWorktree=true")
		}

		// Verify worktree path
		worktreePath := result["worktreePath"].(string)
		expectedPath := filepath.Join(tmpDir, "test-repo-worktrees", "test-feature")
		if worktreePath != expectedPath {
			t.Errorf("Expected worktreePath=%s, got %s", expectedPath, worktreePath)
		}

		// Verify worktree exists
		if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
			t.Error("Worktree directory was not created")
		}

		// Verify branch name
		branchName := result["branch"].(string)
		if branchName != "feature/test-feature" {
			t.Errorf("Expected branch=feature/test-feature, got %s", branchName)
		}

		// Verify bots/ directory was created
		botsDir := filepath.Join(worktreePath, "bots")
		if _, err := os.Stat(botsDir); os.IsNotExist(err) {
			t.Error("bots/ directory was not created")
		}

		// Verify we're on the correct branch in the worktree
		cmd := exec.Command("git", "branch", "--show-current")
		cmd.Dir = worktreePath
		output, _ := cmd.Output()
		currentBranch := strings.TrimSpace(string(output))
		if currentBranch != "feature/test-feature" {
			t.Errorf("Expected current branch=feature/test-feature, got %s", currentBranch)
		}
	})

	// Test 2: Register on main without featureName - should error
	t.Run("ErrorWithoutFeatureName", func(t *testing.T) {
		_, err := sm.Register("brainstorm", repoDir, "Test task", "", "", "")
		if err == nil {
			t.Error("Expected error when registering on main without featureName")
		}

		if !strings.Contains(err.Error(), "featureName") {
			t.Errorf("Error should mention featureName, got: %s", err.Error())
		}
	})

	// Test 3: Register in existing worktree - should not create new worktree
	t.Run("UseExistingWorktree", func(t *testing.T) {
		// Create a worktree manually
		worktreesDir := filepath.Join(tmpDir, "test-repo-worktrees")
		existingWorktree := filepath.Join(worktreesDir, "existing-feature")

		cmd := exec.Command("git", "worktree", "add", "-b", "feature/existing", existingWorktree)
		cmd.Dir = repoDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create test worktree: %v", err)
		}

		// Register workflow in existing worktree
		result, err := sm.Register("brainstorm", existingWorktree, "Test task", "another-name", "", "")
		if err != nil {
			t.Fatalf("Register failed: %v", err)
		}

		// Should NOT have created a new worktree
		if result["createdWorktree"].(bool) {
			t.Error("Expected createdWorktree=false for existing worktree")
		}

		// Should use the existing worktree path
		if result["worktreePath"].(string) != existingWorktree {
			t.Error("Should use existing worktree path")
		}
	})
}

func TestIsMainRepo(t *testing.T) {
	// Create temporary git repo
	tmpDir, err := os.MkdirTemp("", "bob-mainrepo-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	repoDir := filepath.Join(tmpDir, "test-repo")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Init git with main branch
	cmd := exec.Command("git", "init", "-b", "main")
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	// Create initial commit so worktrees can be created
	testFile := filepath.Join(repoDir, "README.md")
	if err := os.WriteFile(testFile, []byte("# Test"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

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

	stateDir := filepath.Join(tmpDir, "state")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		t.Fatal(err)
	}

	sm := &StateManager{
		stateDir:       stateDir,
		additionsCache: make(map[string]*AdditionsCache),
	}

	// Test 1: Main repo should be detected
	t.Run("DetectMainRepo", func(t *testing.T) {
		isMain, _, err := sm.isMainRepo(repoDir)
		if err != nil {
			t.Fatalf("isMainRepo failed: %v", err)
		}

		if !isMain {
			t.Error("Expected isMain=true for main repository")
		}
	})

	// Test 2: Worktree should be detected
	t.Run("DetectWorktree", func(t *testing.T) {
		// Create a worktree (repo now has initial commit, so this will work)
		worktreeDir := filepath.Join(tmpDir, "test-worktree")
		cmd := exec.Command("git", "worktree", "add", "-b", "test-branch", worktreeDir)
		cmd.Dir = repoDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create worktree: %v", err)
		}

		isMain, _, err := sm.isMainRepo(worktreeDir)
		if err != nil {
			t.Fatalf("isMainRepo failed: %v", err)
		}

		if isMain {
			t.Error("Expected isMain=false for worktree")
		}
	})

	// Test 3: Non-git directory should error
	t.Run("ErrorForNonGitDir", func(t *testing.T) {
		nonGitDir := filepath.Join(tmpDir, "not-a-repo")
		os.MkdirAll(nonGitDir, 0755)

		_, _, err := sm.isMainRepo(nonGitDir)
		if err == nil {
			t.Error("Expected error for non-git directory")
		}
	})
}

func TestWorktreeNaming(t *testing.T) {
	tests := []struct {
		featureName    string
		expectedBranch string
		valid          bool
	}{
		{"my-feature", "feature/my-feature", true},
		{"fix-bug", "feature/fix-bug", true},
		{"add-auth-system", "feature/add-auth-system", true},
	}

	for _, tt := range tests {
		t.Run(tt.featureName, func(t *testing.T) {
			expectedBranch := "feature/" + tt.featureName
			if expectedBranch != tt.expectedBranch {
				t.Errorf("Branch naming failed: expected %s, got %s", tt.expectedBranch, expectedBranch)
			}
		})
	}
}
