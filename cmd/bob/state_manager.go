package main

import (
	"crypto/sha256"
	"encoding/json"
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

	// Get workflow definition (with custom workflow support)
	repoPath, err := sm.getRepoPath(actualWorktreePath)
	if err != nil {
		log.Printf("Warning: failed to get repo path, using embedded workflows: %v", err)
		repoPath = ""
	}
	def, err := GetWorkflowDefinition(workflow, repoPath)
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
func (sm *StateManager) tryAdvanceStep(workflow, worktreePath, currentStep string) (string, bool) {
	repoPath, err := sm.getRepoPath(worktreePath)
	if err != nil {
		log.Printf("Warning: failed to get repo path, using embedded workflows: %v", err)
		repoPath = ""
	}
	nextStep, err := GetNextStep(workflow, currentStep, repoPath)
	if err != nil {
		// Check if this is the final step
		if strings.Contains(err.Error(), "final step") {
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
				// File missing = no issues found, advance forward
				nextStep, workflowCompleted = sm.tryAdvanceStep(state.Workflow, state.WorktreePath, currentStep)
			} else {
				// Unexpected I/O error (permissions, disk full, etc.) - fail loudly
				return nil, fmt.Errorf("failed to read findings file (permissions/I/O error): %w", err)
			}
		} else if len(findingsContent) < minFindingsLength {
			// File exists but empty = no issues found, advance forward
			nextStep, workflowCompleted = sm.tryAdvanceStep(state.Workflow, state.WorktreePath, currentStep)
		} else {
			// File exists with content - classify it with Claude API
			claudeClient := NewClaudeClient()
			hasIssues, err := claudeClient.ClassifyFindings(string(findingsContent))

			if err == nil {
				if hasIssues {
					// Issues found - loop back to fix them
					repoPath, err := sm.getRepoPath(state.WorktreePath)
					if err != nil {
						log.Printf("Warning: failed to get repo path, using embedded workflows: %v", err)
						repoPath = ""
					}
					workflowDef, err := GetWorkflowDefinition(state.Workflow, repoPath)
					if err != nil {
						log.Printf("Warning: failed to get workflow definition: %v", err)
					} else {
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
					nextStep, workflowCompleted = sm.tryAdvanceStep(state.Workflow, state.WorktreePath, currentStep)
				}
			} else {
				// If classification fails, log error and advance forward (safe default)
				fmt.Fprintf(os.Stderr, "Warning: Claude classification failed: %v\n", err)
				nextStep, workflowCompleted = sm.tryAdvanceStep(state.Workflow, state.WorktreePath, currentStep)
			}
		}
	// AUTO-ADVANCE: For non-checkpoint phases, if agent reports current step, auto-advance
}
	if currentStep == previousStep && !sm.isCheckpointPhase(state.Workflow, currentStep) {
		nextStep, workflowCompleted = sm.tryAdvanceStep(state.Workflow, state.WorktreePath, currentStep)
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
	if err != nil {
		log.Printf("Warning: failed to get repo path for additions: %v", err)
		repoPath = ""
	}
	if repoPath != "" {
		cache := sm.getOrLoadAdditionsCache(repoPath)
		if cache != nil {
			if addition := cache.GetAddition(state.Workflow, state.CurrentStep); addition != "" {
				prompt = fmt.Sprintf("%s\n\n---\n\n### Project-Specific Guidance\n\n%s", prompt, addition)
			}
		}
	}

	// Append dynamic context based on bots/*.md files
	dynamicContext := sm.generateDynamicContext(worktreePath, state.Workflow, state.CurrentStep)
	if dynamicContext != "" {
		prompt = fmt.Sprintf("%s\n\n---\n\n## Current Context\n\n%s", prompt, dynamicContext)
	}

	// Get loop targets (get repoPath again for workflow definition)
	repoPath, err = sm.getRepoPath(worktreePath)
	if err != nil {
		log.Printf("Warning: failed to get repo path for workflow definition: %v", err)
		repoPath = ""
	}
	def, err := GetWorkflowDefinition(state.Workflow, repoPath)
	if err != nil {
		log.Printf("Warning: failed to get workflow definition: %v", err)
		return nil, fmt.Errorf("failed to load workflow definition: %w", err)
	}
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
		"recorded":    true,
		"issueCount":  0,
		"shouldLoop":  false,
		"loopBackTo":  "",
		"totalIssues": 0,
		"deprecated":  "Issues are now stored in markdown files under bots/",
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
		repoPath, err := sm.getRepoPath(worktreePath)
		if err != nil {
			log.Printf("Warning: failed to get repo path, using embedded workflows: %v", err)
			repoPath = ""
		}
		def, err := GetWorkflowDefinition(state.Workflow, repoPath)
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

// generateDynamicContext reads bots/*.md files and generates contextual guidance
// based on current workflow state and step
func (sm *StateManager) generateDynamicContext(worktreePath, workflow, currentStep string) string {
	// Check if we can loop back from this step (meaning we might have looped TO this step)
	// Read previous step's file if we just looped back
	var content string
	var err error

	// Try to detect loop back by checking if previous step files exist with issues
	// For PLAN step, check if review.md exists (common loop: REVIEW -> PLAN)
	// For EXECUTE step, check if test.md exists (common loop: TEST -> EXECUTE)
	previousStepFiles := map[string][]string{
		"PLAN":    {"review", "test"}, // Could have looped from REVIEW or TEST
		"EXECUTE": {"test", "review"}, // Could have looped from TEST or REVIEW
		"REVIEW":  {"monitor"},        // Could have looped from MONITOR
	}

	// First try previous step files
	if prevSteps, ok := previousStepFiles[currentStep]; ok {
		for _, prevStep := range prevSteps {
			content, err = sm.readBotsFile(worktreePath, prevStep)
			if err == nil && len(content) >= minFindingsLength {
				// Found previous step's file with content - use it
				break
			}
		}
	}

	// If no previous step content, try current step
	if content == "" || err != nil {
		content, err = sm.readBotsFile(worktreePath, currentStep)
		if err != nil || len(content) < minFindingsLength {
			return "" // No meaningful content
		}
	}

	// Parse findings from markdown
	findings := sm.parseMarkdownFindings(content)
	if len(findings) == 0 {
		return ""
	}

	// Format context based on step
	return sm.formatContextForStep(currentStep, findings)
}

// readBotsFile reads a markdown file from the bots/ directory
func (sm *StateManager) readBotsFile(worktreePath, step string) (string, error) {
	filename := sm.stepToMarkdownFilename(step)
	filePath := filepath.Join(worktreePath, "bots", filename)

	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// parseMarkdownFindings extracts key findings from markdown content
// Looks for bullet points, numbered lists, and extracts them as findings
func (sm *StateManager) parseMarkdownFindings(content string) []string {
	var findings []string
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines
		if trimmed == "" {
			continue
		}

		// Extract bullet points (-, *, +)
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "+ ") {
			finding := strings.TrimSpace(trimmed[2:])
			if finding != "" {
				findings = append(findings, finding)
			}
			continue
		}

		// Extract numbered lists (1., 2., etc.)
		for i := 1; i <= 99; i++ {
			prefix := fmt.Sprintf("%d. ", i)
			if strings.HasPrefix(trimmed, prefix) {
				finding := strings.TrimSpace(trimmed[len(prefix):])
				if finding != "" {
					findings = append(findings, finding)
				}
				break
			}
		}

		// Limit to first 10 findings
		if len(findings) >= 10 {
			break
		}
	}

	return findings
}

// formatContextForStep formats findings into a context message based on the step
func (sm *StateManager) formatContextForStep(step string, findings []string) string {
	if len(findings) == 0 {
		return ""
	}

	var sb strings.Builder

	// Check if this is a checkpoint step that might have looped back
	checkpoint := sm.isCheckpointPhase("work", step)

	if checkpoint {
		sb.WriteString("⚠️ Issues found that need attention\n\n")
	}

	// Format findings based on step type
	switch step {
	case "PLAN":
		sb.WriteString("Issues to address in your plan:\n")
	case "EXECUTE":
		sb.WriteString("Issues to fix in your implementation:\n")
	case "REVIEW":
		sb.WriteString("Issues found during review:\n")
	case "TEST":
		sb.WriteString("Test failures to address:\n")
	default:
		sb.WriteString("Issues found:\n")
	}

	// List findings
	for i, finding := range findings {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, finding))
	}

	// Add task instruction based on step
	sb.WriteString("\nYour task:\n")
	switch step {
	case "PLAN":
		sb.WriteString("- Update your plan to address these issues\n")
		sb.WriteString("- Consider impacts and dependencies\n")
		sb.WriteString("- Write updated plan to bots/plan.md\n")
	case "EXECUTE":
		sb.WriteString("- Fix these issues in your implementation\n")
		sb.WriteString("- Ensure tests pass after fixes\n")
		sb.WriteString("- Follow TDD principles\n")
	case "REVIEW":
		sb.WriteString("- Review code for these issues\n")
		sb.WriteString("- Document findings in bots/review.md\n")
		sb.WriteString("- Suggest specific fixes\n")
	case "TEST":
		sb.WriteString("- Fix failing tests\n")
		sb.WriteString("- Verify all tests pass\n")
		sb.WriteString("- Check test coverage\n")
	default:
		sb.WriteString("- Address these issues before proceeding\n")
	}

	return sb.String()
}
