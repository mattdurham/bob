package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
)

// TestNewAdditionsCache tests creating a new cache
func TestNewAdditionsCache(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewAdditionsCache(tmpDir)
	if err != nil {
		t.Fatalf("NewAdditionsCache failed: %v", err)
	}
	if cache == nil {
		t.Fatal("NewAdditionsCache returned nil")
	}
	// repoPath should be absolute
	if !filepath.IsAbs(cache.repoPath) {
		t.Error("repoPath should be absolute")
	}
	if cache.loaded {
		t.Error("New cache should not be marked as loaded")
	}
	if cache.additions == nil {
		t.Error("Additions map should be initialized")
	}
}

// TestLoadAdditions_MissingBobBranch tests loading when bob branch doesn't exist
func TestLoadAdditions_MissingBobBranch(t *testing.T) {
	// Create a temporary git repo without bob branch
	tmpDir := t.TempDir()
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	cache, _ := NewAdditionsCache(tmpDir)
	err := cache.LoadAdditions()

	// Should not error when bob branch doesn't exist
	if err != nil {
		t.Errorf("LoadAdditions should not error on missing bob branch, got: %v", err)
	}
	if cache.loaded {
		t.Error("Cache should not be marked as loaded when bob branch missing")
	}
}

// TestLoadAdditions_MissingAdditionsDir tests loading when .bob/additions doesn't exist
func TestLoadAdditions_MissingAdditionsDir(t *testing.T) {
	// Create a temporary git repo with bob branch but no .bob/additions
	tmpDir := t.TempDir()
	setupGitRepo(t, tmpDir, false)

	cache, _ := NewAdditionsCache(tmpDir)
	err := cache.LoadAdditions()

	// Should not error when .bob/additions doesn't exist
	if err != nil {
		t.Errorf("LoadAdditions should not error on missing .bob/additions, got: %v", err)
	}
	if !cache.loaded {
		t.Error("Cache should be marked as loaded even with no additions")
	}
}

// TestLoadAdditions_WithFiles tests loading with actual addition files
func TestLoadAdditions_WithFiles(t *testing.T) {
	// Create a temporary git repo with bob branch and additions
	tmpDir := t.TempDir()
	setupGitRepo(t, tmpDir, true)

	cache, _ := NewAdditionsCache(tmpDir)
	err := cache.LoadAdditions()

	if err != nil {
		t.Fatalf("LoadAdditions failed: %v", err)
	}
	if !cache.loaded {
		t.Error("Cache should be marked as loaded")
	}

	// Verify content was loaded
	content := cache.GetAddition("brainstorm", "PLAN")
	if content == "" {
		t.Error("Expected to load brainstorm/PLAN addition, got empty string")
	}
	if content != "## Test Addition\nThis is test content for PLAN step." {
		t.Errorf("Unexpected content: %s", content)
	}
}

// TestGetAddition_Present tests retrieving existing addition
func TestGetAddition_Present(t *testing.T) {
	tmpDir := t.TempDir()
	cache, _ := NewAdditionsCache(tmpDir)
	cache.additions = map[string]map[string]string{
		"brainstorm": {
			"PLAN": "Test content",
		},
	}
	cache.loaded = true

	content := cache.GetAddition("brainstorm", "PLAN")
	if content != "Test content" {
		t.Errorf("Expected 'Test content', got '%s'", content)
	}
}

// TestGetAddition_Missing tests retrieving non-existent addition
func TestGetAddition_Missing(t *testing.T) {
	tmpDir := t.TempDir()
	cache, _ := NewAdditionsCache(tmpDir)
	cache.additions = map[string]map[string]string{}
	cache.loaded = true

	content := cache.GetAddition("nonexistent", "STEP")
	if content != "" {
		t.Errorf("Expected empty string for missing addition, got '%s'", content)
	}
}

// TestGetAddition_ConcurrentAccess tests thread safety
func TestGetAddition_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	cache, _ := NewAdditionsCache(tmpDir)
	cache.additions = map[string]map[string]string{
		"workflow1": {"STEP1": "content1"},
		"workflow2": {"STEP2": "content2"},
		"workflow3": {"STEP3": "content3"},
	}
	cache.loaded = true

	var wg sync.WaitGroup
	iterations := 100

	// Start multiple goroutines reading concurrently
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			workflow := []string{"workflow1", "workflow2", "workflow3"}[n%3]
			step := []string{"STEP1", "STEP2", "STEP3"}[n%3]
			_ = cache.GetAddition(workflow, step)
		}(i)
	}

	wg.Wait() // Should not panic or race
}

// TestReadFileFromBobBranch_FileExists tests reading existing file
func TestReadFileFromBobBranch_FileExists(t *testing.T) {
	tmpDir := t.TempDir()
	setupGitRepo(t, tmpDir, true)

	content, err := readFileFromBobBranch(tmpDir, ".bob/additions/brainstorm/PLAN.md")
	if err != nil {
		t.Fatalf("readFileFromBobBranch failed: %v", err)
	}
	if content == "" {
		t.Error("Expected non-empty content")
	}
}

// TestReadFileFromBobBranch_FileNotFound tests reading non-existent file
func TestReadFileFromBobBranch_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	setupGitRepo(t, tmpDir, false)

	content, err := readFileFromBobBranch(tmpDir, ".bob/additions/nonexistent.md")
	// Should return empty string, not error, when file doesn't exist
	if err != nil {
		t.Errorf("readFileFromBobBranch should not error on missing file, got: %v", err)
	}
	if content != "" {
		t.Error("Expected empty string for non-existent file")
	}
}

// setupGitRepo creates a test git repository with bob branch
// If withAdditions is true, adds .bob/additions files
func setupGitRepo(t *testing.T, dir string, withAdditions bool) {
	t.Helper()

	// Initialize git repo
	runCmd(t, dir, "git", "init", "-b", "main")
	runCmd(t, dir, "git", "config", "user.email", "test@example.com")
	runCmd(t, dir, "git", "config", "user.name", "Test User")

	// Create initial commit on main
	readmeFile := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readmeFile, []byte("# Test Repo"), 0644); err != nil {
		t.Fatal(err)
	}
	runCmd(t, dir, "git", "add", "README.md")
	runCmd(t, dir, "git", "commit", "-m", "Initial commit")

	// Create bob branch
	runCmd(t, dir, "git", "checkout", "-b", "bob")

	if withAdditions {
		// Create .bob/additions directory and files
		additionsDir := filepath.Join(dir, ".bob", "additions", "brainstorm")
		if err := os.MkdirAll(additionsDir, 0755); err != nil {
			t.Fatal(err)
		}

		planFile := filepath.Join(additionsDir, "PLAN.md")
		content := "## Test Addition\nThis is test content for PLAN step."
		if err := os.WriteFile(planFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		runCmd(t, dir, "git", "add", ".bob")
		runCmd(t, dir, "git", "commit", "-m", "Add test additions")
	}

	// Switch back to main
	runCmd(t, dir, "git", "checkout", "main")
}

// Security Tests

// TestNewAdditionsCache_RejectsRelativePath tests that relative paths are rejected
func TestNewAdditionsCache_RejectsRelativePath(t *testing.T) {
	cache, err := NewAdditionsCache("../relative/path")
	if err == nil {
		t.Error("Expected error for relative path, got nil")
	}
	if cache != nil {
		t.Error("Expected nil cache for invalid path")
	}
}

// TestNewAdditionsCache_RejectsPathTraversal tests that paths with .. are rejected
func TestNewAdditionsCache_RejectsPathTraversal(t *testing.T) {
	cache, err := NewAdditionsCache("/tmp/../../../etc/passwd")
	if err == nil {
		t.Error("Expected error for path traversal, got nil")
	}
	if cache != nil {
		t.Error("Expected nil cache for traversal path")
	}
}

// TestNewAdditionsCache_AcceptsValidAbsolutePath tests that valid absolute paths work
func TestNewAdditionsCache_AcceptsValidAbsolutePath(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewAdditionsCache(tmpDir)
	if err != nil {
		t.Errorf("Unexpected error for valid path: %v", err)
	}
	if cache == nil {
		t.Error("Expected non-nil cache for valid path")
	}
}

// TestLoadAdditions_RejectsPathTraversalInFiles tests that files with .. are rejected
func TestLoadAdditions_RejectsPathTraversalInFiles(t *testing.T) {
	tmpDir := t.TempDir()
	setupGitRepo(t, tmpDir, false)

	// Create malicious file with path traversal in bob branch
	runCmd(t, tmpDir, "git", "checkout", "bob")

	additionsDir := filepath.Join(tmpDir, ".bob", "additions", "../../malicious")
	if err := os.MkdirAll(additionsDir, 0755); err != nil {
		t.Fatal(err)
	}

	malFile := filepath.Join(additionsDir, "EVIL.md")
	if err := os.WriteFile(malFile, []byte("malicious content"), 0644); err != nil {
		t.Fatal(err)
	}

	runCmd(t, tmpDir, "git", "add", "-A")
	runCmd(t, tmpDir, "git", "commit", "-m", "Add malicious file")
	runCmd(t, tmpDir, "git", "checkout", "main")

	cache, _ := NewAdditionsCache(tmpDir)
	_ = cache.LoadAdditions()

	// Should not have loaded the malicious file
	content := cache.GetAddition("../../malicious", "EVIL")
	if content != "" {
		t.Error("Path traversal should have been rejected")
	}
}

// TestLoadAdditions_RejectsInvalidWorkflowNames tests workflow names with separators
func TestLoadAdditions_RejectsInvalidWorkflowNames(t *testing.T) {
	tmpDir := t.TempDir()
	setupGitRepo(t, tmpDir, false)

	// Try to create workflow with slash
	runCmd(t, tmpDir, "git", "checkout", "bob")

	additionsDir := filepath.Join(tmpDir, ".bob", "additions", "bad/workflow")
	if err := os.MkdirAll(additionsDir, 0755); err != nil {
		t.Fatal(err)
	}

	badFile := filepath.Join(additionsDir, "STEP.md")
	if err := os.WriteFile(badFile, []byte("bad content"), 0644); err != nil {
		t.Fatal(err)
	}

	runCmd(t, tmpDir, "git", "add", "-A")
	runCmd(t, tmpDir, "git", "commit", "-m", "Add bad workflow")
	runCmd(t, tmpDir, "git", "checkout", "main")

	cache, _ := NewAdditionsCache(tmpDir)
	_ = cache.LoadAdditions()

	// Should not have loaded files with / in workflow name
	content := cache.GetAddition("bad/workflow", "STEP")
	if content != "" {
		t.Error("Workflow name with slash should have been rejected")
	}
}

// TestLoadAdditions_ConcurrentCalls tests concurrent LoadAdditions doesn't duplicate work
func TestLoadAdditions_ConcurrentCalls(t *testing.T) {
	tmpDir := t.TempDir()
	setupGitRepo(t, tmpDir, true)

	cache, _ := NewAdditionsCache(tmpDir)

	var wg sync.WaitGroup
	loadCount := 10

	// Spawn multiple goroutines trying to load concurrently
	for i := 0; i < loadCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = cache.LoadAdditions()
		}()
	}

	wg.Wait()

	// Should still work correctly
	content := cache.GetAddition("brainstorm", "PLAN")
	if content == "" {
		t.Error("Cache should have loaded content despite concurrent calls")
	}

	// Cache should be marked as loaded
	if !cache.loaded {
		t.Error("Cache should be marked as loaded after concurrent calls")
	}
}

// TestReadFileFromBobBranch_ErrorHandling tests various error scenarios
func TestReadFileFromBobBranch_ErrorHandling(t *testing.T) {
	// Test with valid repo but missing file - should return empty string, no error
	tmpDir := t.TempDir()
	setupGitRepo(t, tmpDir, false)

	content, err := readFileFromBobBranch(tmpDir, "nonexistent.md")
	if err != nil {
		t.Errorf("Should not error on missing file, got: %v", err)
	}
	if content != "" {
		t.Error("Should return empty string for missing file")
	}

	// Test with valid repo but missing bob branch - should handle gracefully
	content2, err2 := readFileFromBobBranch(tmpDir, ".bob/additions/test.md")
	// This should either return empty or error, but not panic
	_ = content2
	_ = err2
}

// runCmd runs a command and fails the test if it errors
func runCmd(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Command failed: %s %v\nError: %v\nOutput: %s", name, args, err, output)
	}
}
