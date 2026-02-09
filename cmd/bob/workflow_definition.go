package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed workflows/*.json
var workflowsFS embed.FS

// WorkflowDefinition defines a complete workflow with steps and loop rules
type WorkflowDefinition struct {
	Keyword     string     `json:"keyword"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Steps       []Step     `json:"steps"`
	LoopRules   []LoopRule `json:"loopRules"`
}

// Step represents a single workflow step
type Step struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	CanLoopTo    []string `json:"canLoopTo,omitempty"`
	Requirements []string `json:"requirements,omitempty"`
}

// LoopRule defines when and how to loop back to a previous step
type LoopRule struct {
	FromStep    string `json:"fromStep"`
	ToStep      string `json:"toStep"`
	Condition   string `json:"condition"`
	Description string `json:"description"`
}

// GetWorkflowDefinition returns the workflow definition for a given keyword
// If basePath is provided, checks for .bob/workflows/*.json files first
func GetWorkflowDefinition(keyword string, basePath ...string) (*WorkflowDefinition, error) {
	// Try loading from external path first if provided
	if len(basePath) > 0 && basePath[0] != "" {
		externalPath := filepath.Join(basePath[0], ".bob", "workflows", keyword+".json")
		if data, err := os.ReadFile(externalPath); err == nil {
			var workflow WorkflowDefinition
			if err := json.Unmarshal(data, &workflow); err != nil {
				return nil, fmt.Errorf("failed to parse external workflow %s: %w", keyword, err)
			}
			return &workflow, nil
		}
	}

	// Fall back to embedded workflows
	filename := fmt.Sprintf("workflows/%s.json", keyword)
	data, err := workflowsFS.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("workflow not found: %s", keyword)
	}

	var workflow WorkflowDefinition
	if err := json.Unmarshal(data, &workflow); err != nil {
		return nil, fmt.Errorf("failed to parse workflow %s: %w", keyword, err)
	}

	return &workflow, nil
}

// ListWorkflows returns all available workflow keywords
// If basePath is provided, includes workflows from .bob/workflows/*.json
func ListWorkflows(basePath ...string) ([]string, error) {
	workflowMap := make(map[string]bool)

	// Load external workflows first if basePath provided
	if len(basePath) > 0 && basePath[0] != "" {
		externalDir := filepath.Join(basePath[0], ".bob", "workflows")
		if entries, err := os.ReadDir(externalDir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
					keyword := strings.TrimSuffix(entry.Name(), ".json")
					workflowMap[keyword] = true
				}
			}
		}
	}

	// Load embedded workflows
	entries, err := workflowsFS.ReadDir("workflows")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			keyword := strings.TrimSuffix(entry.Name(), ".json")
			workflowMap[keyword] = true
		}
	}

	// Convert map to slice
	var workflows []string
	for keyword := range workflowMap {
		workflows = append(workflows, keyword)
	}

	return workflows, nil
}

// GetNextStep returns the next step in the workflow
func GetNextStep(workflow string, currentStep string, basePath ...string) (string, error) {
	def, err := GetWorkflowDefinition(workflow, basePath...)
	if err != nil {
		return "", err
	}

	for i, step := range def.Steps {
		if step.Name == currentStep {
			if i+1 < len(def.Steps) {
				return def.Steps[i+1].Name, nil
			}
			return "", fmt.Errorf("already at final step")
		}
	}

	return "", fmt.Errorf("unknown step: %s", currentStep)
}
