package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// StateManager manages workflow states
type StateManager struct {
	stateDir string
}

// NewStateManager creates a new state manager
func NewStateManager() *StateManager {
	homeDir, _ := os.UserHomeDir()
	stateDir := filepath.Join(homeDir, ".bob", "state")
	_ = os.MkdirAll(stateDir, 0755)

	return &StateManager{
		stateDir: stateDir,
	}
}


// Register registers a new workflow instance
func (sm *StateManager) Register(workflow, worktreePath, taskDescription string, sessionID, agentID string) (map[string]interface{}, error) {
	workflowID := sm.worktreeToID(worktreePath, sessionID, agentID)

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
		WorktreePath:    worktreePath,
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
		"workflowId":   workflowID,
		"workflow":     workflow,
		"currentStep":  state.CurrentStep,
		"steps":        def.Steps,
		"registeredAt": state.StartedAt,
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

	// Get next step
	nextStep, _ := GetNextStep(state.Workflow, state.CurrentStep)

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
		"nextStep":        nextStep,
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

// loadState loads workflow state from disk
func (sm *StateManager) loadState(workflowID string) (*WorkflowState, error) {
	filename := strings.ReplaceAll(workflowID, "/", "-") + ".json"
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
	filename := strings.ReplaceAll(state.WorkflowID, "/", "-") + ".json"
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
