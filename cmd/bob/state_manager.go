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

// WorkflowState represents the current state of a workflow instance
type WorkflowState struct {
	WorkflowID      string                 `json:"workflowId"`
	SessionID       string                 `json:"sessionId,omitempty"` // Optional: for grouping related agents
	AgentID         string                 `json:"agentId,omitempty"`   // Optional: for multi-agent workflows
	Workflow        string                 `json:"workflow"`
	WorktreePath    string                 `json:"worktreePath"`
	TaskDescription string                 `json:"taskDescription"`
	CurrentStep     string                 `json:"currentStep"`
	ProgressHistory []ProgressEntry        `json:"progressHistory"`
	Issues          []Issue                `json:"issues"`
	LoopCount       int                    `json:"loopCount"`
	Metadata        map[string]interface{} `json:"metadata"`
	RejoinHistory   []RejoinEvent          `json:"rejoinHistory,omitempty"`
	ResetHistory    []ResetEvent           `json:"resetHistory,omitempty"`
	StartedAt       time.Time              `json:"startedAt"`
	UpdatedAt       time.Time              `json:"updatedAt"`
}

// ProgressEntry tracks a single progress report
type ProgressEntry struct {
	Step      string                 `json:"step"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Issue represents a problem found during review/testing
type Issue struct {
	Phase       string `json:"phase"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	File        string `json:"file,omitempty"`
	Line        int    `json:"line,omitempty"`
}

// RejoinEvent tracks a workflow rejoin action
type RejoinEvent struct {
	Timestamp       time.Time `json:"timestamp"`
	FromStep        string    `json:"fromStep"`
	ToStep          string    `json:"toStep"`
	ResetSubsequent bool      `json:"resetSubsequent"`
	Reason          string    `json:"reason,omitempty"`
}

// ResetEvent tracks a workflow reset action
type ResetEvent struct {
	Timestamp    time.Time `json:"timestamp"`
	PreviousStep string    `json:"previousStep"`
	Reason       string    `json:"reason,omitempty"`
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

	// Create initial state
	state := &WorkflowState{
		WorkflowID:      workflowID,
		SessionID:       sessionID,
		AgentID:         agentID,
		Workflow:        workflow,
		WorktreePath:    actualWorktreePath,
		TaskDescription: taskDescription,
		CurrentStep:     def.Steps[0].Name, // Start at first step
		ProgressHistory: []ProgressEntry{
			{
				Step:      def.Steps[0].Name,
				Timestamp: time.Now(),
			},
		},
		Issues:    []Issue{},
		LoopCount: 0,
		Metadata:  make(map[string]interface{}),
		StartedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := sm.saveState(state); err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"workflowId":      workflowID,
		"workflow":        workflow,
		"currentStep":     state.CurrentStep,
		"steps":           def.Steps,
		"registeredAt":    state.StartedAt,
		"worktreePath":    actualWorktreePath,
		"createdWorktree": createdWorktree,
	}

	if createdWorktree {
		result["branch"] = branchName
		result["message"] = fmt.Sprintf("Created worktree at: %s\nBranch: %s\nRun: cd %s", actualWorktreePath, branchName, actualWorktreePath)
	}

	if sessionID != "" {
		result["sessionId"] = sessionID
	}
	if agentID != "" {
		result["agentId"] = agentID
	}

	return result, nil
}

// ReportProgress updates the workflow state
func (sm *StateManager) ReportProgress(worktreePath, currentStep string, metadata map[string]interface{}, sessionID, agentID string) (map[string]interface{}, error) {
	workflowID := sm.worktreeToID(worktreePath, sessionID, agentID)
	state, err := sm.loadState(workflowID)
	if err != nil {
		return nil, fmt.Errorf("workflow not found (did you register it first?): %w", err)
	}

	previousStep := state.CurrentStep
	nextStep := currentStep

	// AUTO-ROUTING: If agent is reporting on current step (not transitioning),
	// check if this is a checkpoint phase and classify findings
	if currentStep == previousStep && sm.isCheckpointPhase(state.Workflow, currentStep) {
		// Check for findings text in metadata
		if findingsText, ok := metadata["findings"].(string); ok {
			// Use Claude API to classify if findings contain issues
			claudeClient := NewClaudeClient()
			hasIssues, err := claudeClient.ClassifyFindings(findingsText)

			if err == nil {
				if hasIssues {
					// Issues found - loop back to fix them
					def, err := GetWorkflowDefinition(state.Workflow)
					if err == nil {
						// Find current step's canLoopTo
						for _, step := range def.Steps {
							if step.Name == currentStep && len(step.CanLoopTo) > 0 {
								// Loop to first available target
								nextStep = step.CanLoopTo[0]
								break
							}
						}
					}
				} else {
					// No issues - advance forward
					nextStepName, err := GetNextStep(state.Workflow, currentStep)
					if err == nil {
						nextStep = nextStepName
					}
				}
			} else {
				// If classification fails, log error but don't fail the workflow
				fmt.Fprintf(os.Stderr, "Warning: Claude classification failed: %v\n", err)
				// Fall back to checking findings length
				if len(strings.TrimSpace(findingsText)) < 10 {
					nextStepName, err := GetNextStep(state.Workflow, currentStep)
					if err == nil {
						nextStep = nextStepName
					}
				}
			}
		}
	}

	// AUTO-ADVANCE: For non-checkpoint phases, if agent reports current step, auto-advance
	if currentStep == previousStep && !sm.isCheckpointPhase(state.Workflow, currentStep) {
		nextStepName, err := GetNextStep(state.Workflow, currentStep)
		if err == nil {
			nextStep = nextStepName
		}
	}

	// Check if this is a loop back
	if sm.isLoopBack(state.Workflow, previousStep, nextStep) {
		state.LoopCount++
	}

	// Update state
	state.CurrentStep = nextStep
	state.UpdatedAt = time.Now()
	state.ProgressHistory = append(state.ProgressHistory, ProgressEntry{
		Step:      nextStep,
		Timestamp: time.Now(),
		Metadata:  metadata,
	})

	for k, v := range metadata {
		state.Metadata[k] = v
	}

	if err := sm.saveState(state); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"recorded":     true,
		"currentStep":  state.CurrentStep,
		"previousStep": previousStep,
		"loopCount":    state.LoopCount,
		"timestamp":    state.UpdatedAt,
		"autoRouted":   currentStep == previousStep, // Indicate if auto-routing was used
	}, nil
}

// GetGuidance returns guidance for the current step
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

	// Prepend task description if this is the initial step and description exists
	if state.TaskDescription != "" && len(state.ProgressHistory) == 0 {
		prompt = fmt.Sprintf("## Task Context\n\n%s\n\n---\n\n%s", state.TaskDescription, prompt)
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
		"currentStep":     state.CurrentStep,
		"prompt":          prompt,
		"taskDescription": state.TaskDescription,
		"canLoopBack":     canLoopBack,
		"loopCount":       state.LoopCount,
	}, nil
}

// RecordIssues records issues and determines if loop is needed
func (sm *StateManager) RecordIssues(worktreePath, step string, issues []Issue, sessionID, agentID string) (map[string]interface{}, error) {
	workflowID := sm.worktreeToID(worktreePath, sessionID, agentID)
	state, err := sm.loadState(workflowID)
	if err != nil {
		return nil, fmt.Errorf("workflow not found: %w", err)
	}

	// Add issues
	for _, issue := range issues {
		issue.Phase = step
		state.Issues = append(state.Issues, issue)
	}
	state.UpdatedAt = time.Now()

	if err := sm.saveState(state); err != nil {
		return nil, err
	}

	// Determine if loop is needed
	shouldLoop := len(issues) > 0
	var loopBackTo string

	if shouldLoop {
		def, _ := GetWorkflowDefinition(state.Workflow)
		for _, rule := range def.LoopRules {
			if rule.FromStep == step && rule.Condition == "issues_found" {
				loopBackTo = rule.ToStep
				break
			}
		}
	}

	return map[string]interface{}{
		"recorded":    true,
		"issueCount":  len(issues),
		"shouldLoop":  shouldLoop,
		"loopBackTo":  loopBackTo,
		"totalIssues": len(state.Issues),
	}, nil
}

// GetStatus returns complete workflow status
func (sm *StateManager) GetStatus(worktreePath string, sessionID, agentID string) (map[string]interface{}, error) {
	workflowID := sm.worktreeToID(worktreePath, sessionID, agentID)
	state, err := sm.loadState(workflowID)
	if err != nil {
		return nil, fmt.Errorf("workflow not found: %w", err)
	}

	return map[string]interface{}{
		"workflowId":      state.WorkflowID,
		"workflow":        state.Workflow,
		"taskDescription": state.TaskDescription,
		"currentStep":     state.CurrentStep,
		"loopCount":       state.LoopCount,
		"issueCount":      len(state.Issues),
		"progressHistory": state.ProgressHistory,
		"issues":          state.Issues,
		"startedAt":       state.StartedAt,
		"updatedAt":       state.UpdatedAt,
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

// isLoopBack checks if moving from prevStep to currentStep is a loop back
func (sm *StateManager) isLoopBack(workflow, prevStep, currentStep string) bool {
	def, err := GetWorkflowDefinition(workflow)
	if err != nil {
		return false
	}

	// Find positions
	prevPos := -1
	currentPos := -1
	for i, step := range def.Steps {
		if step.Name == prevStep {
			prevPos = i
		}
		if step.Name == currentStep {
			currentPos = i
		}
	}

	// Loop back if moving to earlier step
	return currentPos < prevPos
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

// ListAgents lists all agents, optionally filtered by session or worktree
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

		// Apply filters
		if sessionID != "" && state.SessionID != sessionID {
			continue
		}
		if worktreePath != "" && state.WorktreePath != worktreePath {
			continue
		}

		agents = append(agents, map[string]interface{}{
			"workflowId":      state.WorkflowID,
			"sessionId":       state.SessionID,
			"agentId":         state.AgentID,
			"workflow":        state.Workflow,
			"worktreePath":    state.WorktreePath,
			"currentStep":     state.CurrentStep,
			"taskDescription": state.TaskDescription,
			"loopCount":       state.LoopCount,
			"issueCount":      len(state.Issues),
			"startedAt":       state.StartedAt,
			"updatedAt":       state.UpdatedAt,
		})
	}

	return map[string]interface{}{
		"agents": agents,
		"count":  len(agents),
	}, nil
}

// GetSessionStatus returns aggregated status of all agents in a session
func (sm *StateManager) GetSessionStatus(sessionID string) (map[string]interface{}, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("sessionId is required")
	}

	files, err := os.ReadDir(sm.stateDir)
	if err != nil {
		return nil, err
	}

	var agents []map[string]interface{}
	stepCounts := make(map[string]int)
	var totalIssues int
	var totalLoops int
	var earliestStart time.Time
	var latestUpdate time.Time

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

		// Filter by session ID
		if state.SessionID != sessionID {
			continue
		}

		// Aggregate data
		stepCounts[state.CurrentStep]++
		totalIssues += len(state.Issues)
		totalLoops += state.LoopCount

		if earliestStart.IsZero() || state.StartedAt.Before(earliestStart) {
			earliestStart = state.StartedAt
		}
		if state.UpdatedAt.After(latestUpdate) {
			latestUpdate = state.UpdatedAt
		}

		agents = append(agents, map[string]interface{}{
			"agentId":     state.AgentID,
			"workflow":    state.Workflow,
			"currentStep": state.CurrentStep,
			"loopCount":   state.LoopCount,
			"issueCount":  len(state.Issues),
			"updatedAt":   state.UpdatedAt,
		})
	}

	if len(agents) == 0 {
		return nil, fmt.Errorf("no agents found for session: %s", sessionID)
	}

	return map[string]interface{}{
		"sessionId":    sessionID,
		"agentCount":   len(agents),
		"agents":       agents,
		"stepCounts":   stepCounts,
		"totalIssues":  totalIssues,
		"totalLoops":   totalLoops,
		"startedAt":    earliestStart,
		"lastActivity": latestUpdate,
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

	// Evict oldest entry if at capacity (simple FIFO)
	if len(sm.additionsCache) >= maxCacheSize {
		// Remove first entry from map (Go maps don't guarantee order, but this is best effort)
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

// Rejoin allows resuming a workflow at any step with optional state reset
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

	// Update task description if provided
	if taskDescription != "" {
		state.TaskDescription = taskDescription
	}

	// Reset subsequent steps if requested
	if resetSubsequent && step != "" {
		// Reuse the already-loaded definition from validation above
		def, err := GetWorkflowDefinition(state.Workflow)
		if err != nil {
			return nil, fmt.Errorf("failed to load workflow definition for reset: %w", err)
		}

		// Find step position
		stepPos := -1
		for i, s := range def.Steps {
			if s.Name == step {
				stepPos = i
				break
			}
		}

		// Remove progress history entries after the rejoin step
		if stepPos >= 0 {
			var newHistory []ProgressEntry
			for _, entry := range state.ProgressHistory {
				// Find entry position
				entryPos := -1
				for i, s := range def.Steps {
					if s.Name == entry.Step {
						entryPos = i
						break
					}
				}
				// Keep entries at or before rejoin step
				if entryPos >= 0 && entryPos <= stepPos {
					newHistory = append(newHistory, entry)
				}
			}
			state.ProgressHistory = newHistory
		}

		// Clear issues from subsequent steps
		var remainingIssues []Issue
		for _, issue := range state.Issues {
			issuePos := -1
			for i, s := range def.Steps {
				if s.Name == issue.Phase {
					issuePos = i
					break
				}
			}
			if issuePos >= 0 && issuePos <= stepPos {
				remainingIssues = append(remainingIssues, issue)
			}
		}
		state.Issues = remainingIssues
	}

	// Update current step
	state.CurrentStep = step
	state.UpdatedAt = time.Now()

	// Add rejoin event to history
	rejoinEvent := RejoinEvent{
		Timestamp:       time.Now(),
		FromStep:        fromStep,
		ToStep:          step,
		ResetSubsequent: resetSubsequent,
		Reason:          fmt.Sprintf("Rejoined workflow at step %s", step),
	}
	state.RejoinHistory = append(state.RejoinHistory, rejoinEvent)

	// Add to progress history
	state.ProgressHistory = append(state.ProgressHistory, ProgressEntry{
		Step:      step,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"rejoin":          true,
			"fromStep":        fromStep,
			"resetSubsequent": resetSubsequent,
		},
	})

	if err := sm.saveState(state); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"rejoined":        true,
		"workflowId":      workflowID,
		"fromStep":        fromStep,
		"currentStep":     state.CurrentStep,
		"resetSubsequent": resetSubsequent,
		"timestamp":       state.UpdatedAt,
	}, nil
}

// Reset clears workflow state for worktree (optionally archives before reset)
func (sm *StateManager) Reset(worktreePath string, archive bool, sessionID, agentID string) (map[string]interface{}, error) {
	workflowID := sm.worktreeToID(worktreePath, sessionID, agentID)
	state, err := sm.loadState(workflowID)
	if err != nil {
		return nil, fmt.Errorf("workflow not found: %w", err)
	}

	// Archive if requested
	var archivePath string
	if archive {
		timestamp := time.Now().Format("20060102-150405")
		// Safely remove .json extension if present
		baseFilename := strings.TrimSuffix(workflowIDToFilename(workflowID), ".json")
		archiveFilename := fmt.Sprintf("%s-archived-%s.json", baseFilename, timestamp)
		archivePath = filepath.Join(sm.stateDir, "archive")
		if err := os.MkdirAll(archivePath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create archive directory: %w", err)
		}

		archiveFullPath := filepath.Join(archivePath, archiveFilename)

		// Add reset event before archiving
		resetEvent := ResetEvent{
			Timestamp:    time.Now(),
			PreviousStep: state.CurrentStep,
			Reason:       "Workflow reset",
		}
		state.ResetHistory = append(state.ResetHistory, resetEvent)
		state.UpdatedAt = time.Now()

		data, err := json.MarshalIndent(state, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal state for archive: %w", err)
		}

		if err := os.WriteFile(archiveFullPath, data, 0644); err != nil {
			return nil, fmt.Errorf("failed to archive state: %w", err)
		}
		archivePath = archiveFullPath
	}

	// Delete the state file
	filename := workflowIDToFilename(workflowID)
	statePath := filepath.Join(sm.stateDir, filename)
	if err := os.Remove(statePath); err != nil {
		return nil, fmt.Errorf("failed to remove state file: %w", err)
	}

	result := map[string]interface{}{
		"reset":      true,
		"workflowId": workflowID,
		"archived":   archive,
		"timestamp":  time.Now(),
	}

	if archive {
		result["archivePath"] = archivePath
	}

	return result, nil
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
