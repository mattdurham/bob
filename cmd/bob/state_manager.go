package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const minFindingsLength = 10 // Minimum content length to consider findings meaningful

// WorkflowState represents the current state of a workflow instance
// Simplified to only track essential information: workflow type, location, and current step
// All other data (findings, issues, progress) is stored in markdown files under bots/
type WorkflowState struct {
	WorkflowID   string `json:"workflowId"`
	Workflow     string `json:"workflow"`
	WorktreePath string `json:"worktreePath"`
	CurrentStep  string `json:"currentStep"`
}

// StateManager manages workflow states
type StateManager struct {
	stateDir       string
	additionsCache map[string]*AdditionsCache // Cache per repo path
	cacheMutex     sync.RWMutex               // Protects additionsCache map
}

// NewStateManager creates a new state manager
func NewStateManager() *StateManager {
	homeDir, _ := os.UserHomeDir()
	stateDir := filepath.Join(homeDir, ".bob", "state")
	_ = os.MkdirAll(stateDir, 0755)

	return &StateManager{
		stateDir:       stateDir,
		additionsCache: make(map[string]*AdditionsCache),
	}
}

// Register registers a new workflow instance
// If worktreePath points to main repo and featureName is provided, automatically creates a worktree
func (sm *StateManager) Register(workflow, worktreePath, taskDescription, featureName string, sessionID, agentID string) (map[string]interface{}, error) {
	var createdWorktree bool
	var branchName string
	var actualWorktreePath = worktreePath

	// Check if this is the main repository
	isMain, repoRoot, err := sm.isMainRepo(worktreePath)
	if err != nil {
		return nil, err
	}

	// If on main repo and featureName provided, create worktree
	if isMain && featureName != "" {
		newWorktreePath, branch, err := sm.createWorktree(repoRoot, featureName)
		if err != nil {
			return nil, fmt.Errorf("failed to create worktree: %w", err)
		}
		actualWorktreePath = newWorktreePath
		branchName = branch
		createdWorktree = true
	} else if isMain && featureName == "" {
		return nil, fmt.Errorf("cannot register workflow on main branch without featureName parameter. Provide featureName to auto-create worktree, or use an existing worktree path")
	}

	workflowID := sm.worktreeToID(actualWorktreePath, sessionID, agentID)

	// Check if already exists
	if _, err := sm.loadState(workflowID); err == nil {
		return nil, fmt.Errorf("workflow already registered for this worktree")
	}

	// Get workflow definition
	def, err := GetWorkflowDefinition(workflow)
	if err != nil {
		return nil, err
	}

	// Validate workflow has at least one step
	if len(def.Steps) == 0 {
		return nil, fmt.Errorf("workflow '%s' has no steps defined", workflow)
	}

	// Create initial state (simplified - only essential fields)
	state := &WorkflowState{
		WorkflowID:   workflowID,
		Workflow:     workflow,
		WorktreePath: actualWorktreePath,
		CurrentStep:  def.Steps[0].Name, // Start at first step
	}

	if err := sm.saveState(state); err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"workflowId":      workflowID,
		"workflow":        workflow,
		"currentStep":     state.CurrentStep,
		"steps":           def.Steps,
		"registeredAt":    time.Now(),
		"worktreePath":    actualWorktreePath,
		"createdWorktree": createdWorktree,
	}

	if createdWorktree {
		result["branch"] = branchName
		result["message"] = fmt.Sprintf("Created worktree at: %s\nBranch: %s\nRun: cd %s", actualWorktreePath, branchName, actualWorktreePath)
	}

	// Keep sessionID/agentID in result for backward compatibility, even though not stored in state
	if sessionID != "" {
		result["sessionId"] = sessionID
	}
	if agentID != "" {
		result["agentId"] = agentID
	}

	return result, nil
}

// tryAdvanceStep attempts to get the next step in workflow
// Returns (nextStep, isCompleted). If final step reached, returns (currentStep, true)
// Logs error if advancement fails for reasons other than reaching final step
func (sm *StateManager) tryAdvanceStep(workflow, currentStep string) (string, bool) {
	nextStep, err := GetNextStep(workflow, currentStep)
	if err != nil {
		// Check if this is the final step using typed error
		if errors.Is(err, ErrFinalStep) {
			return currentStep, true // Workflow completed
		}
		// Other error - log and stay at current step
		fmt.Fprintf(os.Stderr, "Warning: failed to advance from step %s: %v\n", currentStep, err)
		return currentStep, false
	}
	return nextStep, false
}

// ReportProgress updates the workflow state
// KEY CHANGE: Reads markdown files directly instead of using metadata for validation
func (sm *StateManager) ReportProgress(worktreePath, currentStep string, metadata map[string]interface{}, sessionID, agentID string) (map[string]interface{}, error) {
	workflowID := sm.worktreeToID(worktreePath, sessionID, agentID)
	state, err := sm.loadState(workflowID)
	if err != nil {
		return nil, fmt.Errorf("workflow not found (did you register it first?): %w", err)
	}

	previousStep := state.CurrentStep
	nextStep := currentStep
	var workflowCompleted bool

	// AUTO-ROUTING: If agent is reporting on current step (not transitioning),
	// check if this is a checkpoint phase and classify findings from markdown file
	if currentStep == previousStep && sm.isCheckpointPhase(state.Workflow, currentStep) {
		// Read markdown file directly instead of using metadata
		findingsFile := filepath.Join(worktreePath, "bots", sm.stepToMarkdownFilename(currentStep))
		findingsContent, err := os.ReadFile(findingsFile)

		if err != nil {
			// Distinguish file-not-found from other errors
			if os.IsNotExist(err) {
				// Enforce contract: findings file must exist for checkpoint phases
				// Agents must ALWAYS write findings before reporting progress
				return nil, fmt.Errorf("findings file not found for checkpoint step %q; agents must ALWAYS write findings before reporting progress. Expected file: %s", currentStep, findingsFile)
			} else {
				// Unexpected I/O error (permissions, disk full, etc.) - fail loudly
				return nil, fmt.Errorf("failed to read findings file (permissions/I/O error): %w", err)
			}
		} else if len(findingsContent) < minFindingsLength {
			// File exists but empty = no issues found, advance forward
			nextStep, workflowCompleted = sm.tryAdvanceStep(state.Workflow, currentStep)
		} else {
			// File exists with content - classify it with Claude API
			claudeClient := NewClaudeClient()
			hasIssues, err := claudeClient.ClassifyFindings(string(findingsContent))

			if err == nil {
				if hasIssues {
					// Issues found - loop back to fix them
					workflowDef, err := GetWorkflowDefinition(state.Workflow)
					if err == nil {
						// Find current step's canLoopTo
						for _, step := range workflowDef.Steps {
							if step.Name == currentStep && len(step.CanLoopTo) > 0 {
								// Loop to first available target
								nextStep = step.CanLoopTo[0]
								break
							}
						}
					}
				} else {
					// No issues - advance forward
					nextStep, workflowCompleted = sm.tryAdvanceStep(state.Workflow, currentStep)
				}
			} else {
				// If classification fails, log error and advance forward (safe default)
				fmt.Fprintf(os.Stderr, "Warning: Claude classification failed: %v\n", err)
				nextStep, workflowCompleted = sm.tryAdvanceStep(state.Workflow, currentStep)
			}
		}
	}

	// AUTO-ADVANCE: For non-checkpoint phases, if agent reports current step, auto-advance
	if currentStep == previousStep && !sm.isCheckpointPhase(state.Workflow, currentStep) {
		nextStep, workflowCompleted = sm.tryAdvanceStep(state.Workflow, currentStep)
	}

	// Update state (simplified - just update current step)
	state.CurrentStep = nextStep

	if err := sm.saveState(state); err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"recorded":     true,
		"currentStep":  state.CurrentStep,
		"previousStep": previousStep,
		"loopCount":    0, // Removed loop count tracking
		"timestamp":    time.Now(),
		"autoRouted":   currentStep == previousStep, // Indicate if auto-routing was used
	}

	// Add completion flag if workflow finished
	if workflowCompleted {
		result["completed"] = true
		result["message"] = "Workflow completed"
	}

	return result, nil
}

// GetGuidance returns guidance for the current step (simplified)
func (sm *StateManager) GetGuidance(worktreePath string, sessionID, agentID string) (map[string]interface{}, error) {
	workflowID := sm.worktreeToID(worktreePath, sessionID, agentID)
	state, err := sm.loadState(workflowID)
	if err != nil {
		return nil, fmt.Errorf("workflow not found: %w", err)
	}

	// Load prompt for current step
	prompt, err := LoadPrompt(state.Workflow, state.CurrentStep)
	if err != nil {
		return nil, err
	}

	// Append project-specific additions if they exist
	repoPath, err := sm.getRepoPath(worktreePath)
	if err == nil {
		cache := sm.getOrLoadAdditionsCache(repoPath)
		if cache != nil {
			if addition := cache.GetAddition(state.Workflow, state.CurrentStep); addition != "" {
				prompt = fmt.Sprintf("%s\n\n---\n\n### Project-Specific Guidance\n\n%s", prompt, addition)
			}
		}
	}

	// Get loop targets
	def, _ := GetWorkflowDefinition(state.Workflow)
	var canLoopBack []string
	for _, step := range def.Steps {
		if step.Name == state.CurrentStep {
			canLoopBack = step.CanLoopTo
			break
		}
	}

	return map[string]interface{}{
		"currentStep": state.CurrentStep,
		"prompt":      prompt,
		"canLoopBack": canLoopBack,
		"loopCount":   0, // Removed loop count tracking
	}, nil
}

// RecordIssues is deprecated - issues are now stored in markdown files under bots/
// Kept for backward compatibility but does nothing
func (sm *StateManager) RecordIssues(worktreePath, step string, issues []interface{}, sessionID, agentID string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"recorded":    false,
		"issueCount":  0,
		"shouldLoop":  false,
		"loopBackTo":  "",
		"totalIssues": 0,
		"deprecated":  true,
		"warning":     "This function is deprecated. Issues are now stored in markdown files under bots/<step>.md. Write findings to those files instead.",
	}, nil
}

// GetStatus returns complete workflow status
// GetStatus returns complete workflow status (simplified)
func (sm *StateManager) GetStatus(worktreePath string, sessionID, agentID string) (map[string]interface{}, error) {
	workflowID := sm.worktreeToID(worktreePath, sessionID, agentID)
	state, err := sm.loadState(workflowID)
	if err != nil {
		return nil, fmt.Errorf("workflow not found: %w", err)
	}

	return map[string]interface{}{
		"workflowId":   state.WorkflowID,
		"workflow":     state.Workflow,
		"worktreePath": state.WorktreePath,
		"currentStep":  state.CurrentStep,
		// Deprecated fields kept for backward compatibility (always empty)
		"taskDescription": "",
		"loopCount":       0,
		"progressHistory": []interface{}{},
		"updatedAt":       "",
	}, nil
}

// worktreeToID converts a worktree path (+ optional session/agent) to a unique ID
func (sm *StateManager) worktreeToID(worktreePath, sessionID, agentID string) string {
	// Pattern: ~/source/<repo>-worktrees/<name> -> <repo>/<name>
	var baseID string
	parts := strings.Split(worktreePath, "-worktrees/")
	if len(parts) == 2 {
		repo := filepath.Base(strings.TrimSuffix(parts[0], "/"))
		name := strings.TrimSuffix(parts[1], "/")
		baseID = fmt.Sprintf("%s/%s", repo, name)
	} else {
		// Fallback: hash the path
		hash := sha256.Sum256([]byte(worktreePath))
		baseID = fmt.Sprintf("%x", hash[:8])
	}

	// Add session ID if provided
	if sessionID != "" {
		baseID = fmt.Sprintf("%s-session-%s", baseID, sessionID)
	}

	// Add agent ID if provided
	if agentID != "" {
		baseID = fmt.Sprintf("%s-agent-%s", baseID, agentID)
	}

	return baseID
}

// isCheckpointPhase checks if a step is a checkpoint that requires Claude classification
func (sm *StateManager) isCheckpointPhase(workflow, stepName string) bool {
	// Checkpoint phases that require findings classification
	checkpointPhases := []string{"REVIEW", "TEST", "MONITOR", "PROMPT"}

	for _, phase := range checkpointPhases {
		if stepName == phase {
			return true
		}
	}

	return false
}

// stepToMarkdownFilename maps a workflow step name to its markdown filename
// e.g., "REVIEW" -> "review.md", "TEST" -> "test.md"
func (sm *StateManager) stepToMarkdownFilename(stepName string) string {
	// Convert step name to lowercase for filename
	// e.g., "REVIEW" -> "review.md", "BRAINSTORM" -> "brainstorm.md"
	return strings.ToLower(stepName) + ".md"
}

// workflowIDToFilename converts a workflow ID to a safe, collision-free filename
// Uses URL encoding to ensure reversibility and avoid collisions like:
// - "foo/bar-baz" and "foo-bar/baz" both mapping to "foo-bar-baz.json"
func workflowIDToFilename(workflowID string) string {
	// URL encode the workflow ID to make it filesystem-safe
	// This handles /, \, and other special characters
	return url.PathEscape(workflowID) + ".json"
}

// loadState loads workflow state from disk
func (sm *StateManager) loadState(workflowID string) (*WorkflowState, error) {
	filename := workflowIDToFilename(workflowID)
	path := filepath.Join(sm.stateDir, filename)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var state WorkflowState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

// saveState saves workflow state to disk
func (sm *StateManager) saveState(state *WorkflowState) error {
	filename := workflowIDToFilename(state.WorkflowID)
	path := filepath.Join(sm.stateDir, filename)

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(path, data, 0644)
}

// ListAgents lists all workflows, optionally filtered by worktree (simplified)
func (sm *StateManager) ListAgents(sessionID, worktreePath string) (map[string]interface{}, error) {
	files, err := os.ReadDir(sm.stateDir)
	if err != nil {
		return nil, err
	}

	var agents []map[string]interface{}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		path := filepath.Join(sm.stateDir, file.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var state WorkflowState
		if err := json.Unmarshal(data, &state); err != nil {
			continue
		}

		// Apply filter (only worktreePath filter now, sessionID ignored)
		if worktreePath != "" && state.WorktreePath != worktreePath {
			continue
		}

		agents = append(agents, map[string]interface{}{
			"workflowId":   state.WorkflowID,
			"workflow":     state.Workflow,
			"worktreePath": state.WorktreePath,
			"currentStep":  state.CurrentStep,
		})
	}

	return map[string]interface{}{
		"agents": agents,
		"count":  len(agents),
	}, nil
}

// GetSessionStatus is deprecated - sessions are no longer tracked
// Kept for backward compatibility but returns minimal info
func (sm *StateManager) GetSessionStatus(sessionID string) (map[string]interface{}, error) {
	// Log deprecation warning
	fmt.Fprintf(os.Stderr, "Warning: GetSessionStatus is deprecated and will be removed in a future version. Use GetStatus instead.\n")

	return map[string]interface{}{
		"sessionId":  sessionID,
		"agentCount": 0,
		"agents":     []interface{}{},
		"deprecated": "Session tracking removed - use ListAgents instead",
	}, nil
}

// getRepoPath returns the git repository root path from a worktree path
func (sm *StateManager) getRepoPath(worktreePath string) (string, error) {
	cmd := exec.Command("git", "-C", worktreePath, "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository: %s", worktreePath)
	}

	return strings.TrimSpace(string(output)), nil
}

// Maximum number of repository additions caches to maintain in memory
const maxCacheSize = 100

// getOrLoadAdditionsCache gets or creates an additions cache for a repository.
// Returns nil if the repository path is invalid.
func (sm *StateManager) getOrLoadAdditionsCache(repoPath string) *AdditionsCache {
	// Check if cache exists (read lock)
	sm.cacheMutex.RLock()
	cache, exists := sm.additionsCache[repoPath]
	sm.cacheMutex.RUnlock()

	if exists {
		return cache
	}

	// Create and load cache (write lock)
	sm.cacheMutex.Lock()
	defer sm.cacheMutex.Unlock()

	// Double-check after acquiring write lock (another goroutine might have created it)
	if cache, exists := sm.additionsCache[repoPath]; exists {
		return cache
	}

	// Evict a random entry if at capacity (Go maps have randomized iteration)
	// For true LRU/FIFO, use a separate timestamp or linked list structure
	if len(sm.additionsCache) >= maxCacheSize {
		for k := range sm.additionsCache {
			delete(sm.additionsCache, k)
			log.Printf("Evicted additions cache for %s (cache full)", k)
			break
		}
	}

	// Create new cache and validate repo path
	cache, err := NewAdditionsCache(repoPath)
	if err != nil {
		log.Printf("Failed to create additions cache for %s: %v", repoPath, err)
		return nil
	}

	// Load additions from bob branch
	if err := cache.LoadAdditions(); err != nil {
		// Log error but still use cache - missing bob branch is OK, other errors are logged
		log.Printf("Warning: Failed to load additions for %s: %v", repoPath, err)
	}

	sm.additionsCache[repoPath] = cache
	return cache
}

// Rejoin allows resuming a workflow at any step (simplified)
func (sm *StateManager) Rejoin(worktreePath, step, taskDescription string, resetSubsequent bool, sessionID, agentID string) (map[string]interface{}, error) {
	workflowID := sm.worktreeToID(worktreePath, sessionID, agentID)
	state, err := sm.loadState(workflowID)
	if err != nil {
		return nil, fmt.Errorf("workflow not found: %w", err)
	}

	// Validate step if provided
	if step != "" {
		def, err := GetWorkflowDefinition(state.Workflow)
		if err != nil {
			return nil, err
		}

		// Check if step exists in workflow
		validStep := false
		for _, s := range def.Steps {
			if s.Name == step {
				validStep = true
				break
			}
		}

		if !validStep {
			return nil, fmt.Errorf("invalid step '%s' for workflow '%s'", step, state.Workflow)
		}
	} else {
		// If no step specified, continue from current step
		step = state.CurrentStep
	}

	fromStep := state.CurrentStep

	// Note: taskDescription parameter ignored (no longer stored in state)
	// Note: resetSubsequent parameter ignored (no history to reset)

	// Update current step
	state.CurrentStep = step

	if err := sm.saveState(state); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"rejoined":        true,
		"workflowId":      workflowID,
		"fromStep":        fromStep,
		"currentStep":     state.CurrentStep,
		"resetSubsequent": resetSubsequent, // Returned for compatibility but has no effect
		"timestamp":       time.Now(),
	}, nil
}

// Reset clears workflow state for worktree (simplified - no archiving)
func (sm *StateManager) Reset(worktreePath string, archive bool, sessionID, agentID string) (map[string]interface{}, error) {
	workflowID := sm.worktreeToID(worktreePath, sessionID, agentID)

	// Note: archive parameter ignored (no complex state to archive)
	// Just delete the state file
	filename := workflowIDToFilename(workflowID)
	statePath := filepath.Join(sm.stateDir, filename)
	if err := os.Remove(statePath); err != nil {
		return nil, fmt.Errorf("failed to remove state file: %w", err)
	}

	return map[string]interface{}{
		"reset":      true,
		"workflowId": workflowID,
		"archived":   false, // Always false now
		"timestamp":  time.Now(),
	}, nil
}

// isMainRepo checks if the given path is the main repository (not a worktree)
func (sm *StateManager) isMainRepo(path string) (bool, string, error) {
	// Get the git directory for this path
	cmd := exec.Command("git", "-C", path, "rev-parse", "--git-dir")
	output, err := cmd.Output()
	if err != nil {
		return false, "", fmt.Errorf("not a git repository: %s", path)
	}

	gitDir := strings.TrimSpace(string(output))

	// Get the repository root
	cmd = exec.Command("git", "-C", path, "rev-parse", "--show-toplevel")
	output, err = cmd.Output()
	if err != nil {
		return false, "", fmt.Errorf("failed to get repository root: %w", err)
	}
	repoRoot := strings.TrimSpace(string(output))

	// Check if this is a worktree by checking if .git is a file (worktree) or directory (main)
	gitPath := filepath.Join(path, ".git")
	info, err := os.Stat(gitPath)
	if err != nil {
		// If .git doesn't exist in this path, it's likely the repo root with .git directory elsewhere
		isMain := gitDir == ".git" || gitDir == filepath.Join(path, ".git")
		return isMain, repoRoot, nil
	}

	// If .git is a directory, it's the main repo
	// If .git is a file, it's a worktree
	isMain := info.IsDir()
	return isMain, repoRoot, nil
}

// createWorktree creates a new git worktree for the given feature
func (sm *StateManager) createWorktree(repoPath, featureName string) (string, string, error) {
	// Get repository name
	repoName := filepath.Base(repoPath)

	// Create worktree directory path
	worktreesDir := filepath.Join(filepath.Dir(repoPath), repoName+"-worktrees")
	worktreePath := filepath.Join(worktreesDir, featureName)

	// Create branch name
	branchName := "feature/" + featureName

	// Detect the default branch (main or master) without changing current branch
	var baseBranch string
	cmd := exec.Command("git", "-C", repoPath, "symbolic-ref", "refs/remotes/origin/HEAD")
	output, err := cmd.Output()
	if err == nil {
		// Parse "refs/remotes/origin/main" -> "main"
		ref := strings.TrimSpace(string(output))
		baseBranch = strings.TrimPrefix(ref, "refs/remotes/origin/")
	} else {
		// Fallback: check if main or master exists locally
		cmd = exec.Command("git", "-C", repoPath, "rev-parse", "--verify", "main")
		if cmd.Run() == nil {
			baseBranch = "main"
		} else {
			cmd = exec.Command("git", "-C", repoPath, "rev-parse", "--verify", "master")
			if cmd.Run() == nil {
				baseBranch = "master"
			} else {
				// Last resort: use current HEAD
				cmd = exec.Command("git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD")
				output, err = cmd.Output()
				if err != nil {
					return "", "", fmt.Errorf("failed to determine base branch: %w", err)
				}
				baseBranch = strings.TrimSpace(string(output))
			}
		}
	}

	// Create worktrees directory if it doesn't exist
	if err := os.MkdirAll(worktreesDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create worktrees directory: %w", err)
	}

	// Create the worktree from base branch (without modifying current working tree)
	cmd = exec.Command("git", "-C", repoPath, "worktree", "add", "-b", branchName, worktreePath, baseBranch)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("failed to create worktree: %w\nOutput: %s", err, string(output))
	}

	// Create bots/ directory in the new worktree
	botsDir := filepath.Join(worktreePath, "bots")
	if err := os.MkdirAll(botsDir, 0755); err != nil {
		// Cleanup: remove the partially-created worktree
		_ = exec.Command("git", "-C", repoPath, "worktree", "remove", worktreePath, "--force").Run()
		_ = exec.Command("git", "-C", repoPath, "branch", "-D", branchName).Run()
		return "", "", fmt.Errorf("failed to create bots directory: %w", err)
	}

	return worktreePath, branchName, nil
}
