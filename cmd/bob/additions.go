package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// AdditionsCache caches prompt additions from .bob/additions/ on bob branch.
// All methods are safe for concurrent use by multiple goroutines.
type AdditionsCache struct {
	additions map[string]map[string]string // [workflow][step]content
	mu        sync.RWMutex
	repoPath  string
	loaded    bool
}

// NewAdditionsCache creates a new additions cache for a repository.
// Returns error if the repository path is invalid or contains path traversal attempts.
func NewAdditionsCache(repoPath string) (*AdditionsCache, error) {
	// Validate and sanitize the repository path
	validPath, err := validateRepoPath(repoPath)
	if err != nil {
		return nil, err
	}

	return &AdditionsCache{
		additions: make(map[string]map[string]string),
		repoPath:  validPath,
		loaded:    false,
	}, nil
}

// validateRepoPath validates and sanitizes a repository path.
// Returns error if path is relative, contains traversal, or is otherwise invalid.
func validateRepoPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("repository path cannot be empty")
	}

	// Check if path contains obvious traversal patterns before cleaning
	if strings.Contains(path, "..") {
		return "", fmt.Errorf("repository path contains traversal: %s", path)
	}

	// Reject relative paths - must be absolute
	if !filepath.IsAbs(path) {
		return "", fmt.Errorf("repository path must be absolute: %s", path)
	}

	// Clean the path to normalize it
	cleanPath := filepath.Clean(path)

	// After cleaning, ".." should not appear
	if strings.Contains(cleanPath, "..") {
		return "", fmt.Errorf("repository path contains traversal after cleaning: %s", path)
	}

	return cleanPath, nil
}

// LoadAdditions loads all additions from .bob/additions/ on bob branch.
// Returns nil if bob branch or .bob/additions/ doesn't exist (not an error).
func (ac *AdditionsCache) LoadAdditions() error {
	// Quick check if already loaded (read lock)
	ac.mu.RLock()
	if ac.loaded {
		ac.mu.RUnlock()
		return nil
	}
	ac.mu.RUnlock()

	// Acquire write lock and double-check
	ac.mu.Lock()
	defer ac.mu.Unlock()

	// Double-check after acquiring write lock
	if ac.loaded {
		return nil
	}

	// Check if bob branch exists
	cmd := exec.Command("git", "-C", ac.repoPath, "rev-parse", "--verify", "bob")
	if err := cmd.Run(); err != nil {
		// Bob branch doesn't exist - this is OK, not an error
		return nil
	}

	// List all files in .bob/additions/
	cmd = exec.Command("git", "-C", ac.repoPath, "ls-tree", "-r", "--name-only", "bob:.bob/additions/")
	output, err := cmd.Output()
	if err != nil {
		// .bob/additions/ doesn't exist - this is OK, not an error
		ac.loaded = true
		return nil
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		ac.loaded = true
		return nil
	}

	files := strings.Split(outputStr, "\n")
	for _, file := range files {
		if file == "" || !strings.HasSuffix(file, ".md") {
			continue
		}

		// Parse workflow and step from path: workflow/STEP.md
		// (ls-tree output is relative to .bob/additions/)
		parts := strings.Split(file, "/")
		if len(parts) < 2 {
			continue
		}

		workflow := parts[0]
		step := strings.TrimSuffix(parts[1], ".md")

		// Security: Validate no path traversal in workflow or step names
		if strings.Contains(workflow, "..") || strings.Contains(step, "..") {
			continue
		}
		// Security: Reject if workflow or step contains path separators
		if strings.ContainsAny(workflow, "/\\") || strings.ContainsAny(step, "/\\") {
			continue
		}

		// Security: Construct and validate the full path stays within .bob/additions/
		fullPath := filepath.Join(".bob/additions", workflow, step+".md")
		cleanPath := filepath.Clean(fullPath)
		expectedPrefix := filepath.Clean(".bob/additions") + string(filepath.Separator)
		if !strings.HasPrefix(cleanPath+string(filepath.Separator), expectedPrefix) {
			continue
		}

		// Read file content from bob branch
		content, err := readFileFromBobBranch(ac.repoPath, fullPath)
		if err != nil {
			// Log error but continue processing other files
			continue
		}
		if content == "" {
			continue
		}

		// Initialize workflow map if needed
		if ac.additions[workflow] == nil {
			ac.additions[workflow] = make(map[string]string)
		}

		ac.additions[workflow][step] = content
	}

	ac.loaded = true
	return nil
}

// GetAddition retrieves an addition for a workflow and step.
// Returns empty string if not found.
func (ac *AdditionsCache) GetAddition(workflow, step string) string {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	if workflowMap, ok := ac.additions[workflow]; ok {
		if content, ok := workflowMap[step]; ok {
			return content
		}
	}

	return ""
}

// readFileFromBobBranch reads a file from the bob branch.
// Returns empty string (not error) if file doesn't exist.
// Returns error for actual git failures.
func readFileFromBobBranch(repoPath, filePath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "show", "bob:"+filePath)
	output, err := cmd.Output()
	if err != nil {
		// Check if this is just a "file not found" error
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Exit code 128 typically means file doesn't exist
			if exitErr.ExitCode() == 128 {
				return "", nil
			}
		}
		// Real error - return it with context
		return "", fmt.Errorf("failed to read %s from bob branch: %w", filePath, err)
	}

	// Optimize: avoid conversion if output is empty
	if len(output) == 0 {
		return "", nil
	}

	return string(output), nil
}
