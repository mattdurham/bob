package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

//go:embed templates/*.html
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

// workflowIDPattern validates workflow IDs (format: repo/name or repo-name)
// Allows alphanumeric, underscore, hyphen, forward slash, and dot (for directory names)
var workflowIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_/.-]+$`)

// StartUIServer starts the web UI server
func StartUIServer(host, port string) error {
	addr := fmt.Sprintf("%s:%s", host, port)

	// Parse templates
	tmpl, err := template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return fmt.Errorf("failed to parse templates: %w", err)
	}

	// Routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleDashboard(w, r, tmpl)
	})
	http.HandleFunc("/workflow/", func(w http.ResponseWriter, r *http.Request) {
		handleWorkflowDetail(w, r, tmpl)
	})
	http.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		handleTasks(w, r, tmpl)
	})
	http.Handle("/static/", http.FileServer(http.FS(staticFS)))

	// Start server
	fmt.Printf("🏴‍☠️ Bob Web UI starting on http://%s\n", addr)
	fmt.Printf("📊 View your workflows and tasks in your browser\n")
	fmt.Printf("Press Ctrl+C to stop\n\n")

	return http.ListenAndServe(addr, nil)
}

// DashboardData contains data for the dashboard page
type DashboardData struct {
	Workflows []WorkflowSummary
	TaskCount int
	Error     string
}

// WorkflowSummary contains summary info for a workflow (simplified)
type WorkflowSummary struct {
	WorkflowID   string
	Workflow     string
	CurrentStep  string
	WorktreePath string
}

// WorkflowDetailData contains data for the workflow detail page (simplified)
type WorkflowDetailData struct {
	Summary  WorkflowSummary
	Metadata map[string]interface{}
	Error    string
}

func handleDashboard(w http.ResponseWriter, r *http.Request, tmpl *template.Template) {
	data := DashboardData{}

	// Load workflows
	workflows, err := loadWorkflows()
	if err != nil {
		data.Error = fmt.Sprintf("Error loading workflows: %v", err)
	} else {
		data.Workflows = workflows
	}

	// Load task count
	taskCount, err := countTasks()
	if err != nil {
		log.Printf("Error counting tasks: %v", err)
	}
	data.TaskCount = taskCount

	if err := tmpl.ExecuteTemplate(w, "dashboard.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleWorkflowDetail(w http.ResponseWriter, r *http.Request, tmpl *template.Template) {
	// Extract workflow ID from URL
	workflowID := strings.TrimPrefix(r.URL.Path, "/workflow/")
	if workflowID == "" {
		http.Error(w, "Workflow ID required", http.StatusBadRequest)
		return
	}

	// Validate workflow ID format (prevent path traversal)
	// Allow "/" for workflow IDs like "bob/codex-integration", but block path traversal
	if strings.Contains(workflowID, "..") || strings.Contains(workflowID, "\\") {
		http.Error(w, "Invalid workflow ID", http.StatusBadRequest)
		return
	}

	// Additional validation: must match expected format (alphanumeric, underscore, hyphen, forward slash)
	if !workflowIDPattern.MatchString(workflowID) {
		http.Error(w, "Invalid workflow ID format", http.StatusBadRequest)
		return
	}

	data := WorkflowDetailData{}

	// Load workflow details - loadWorkflowDetail handles safe filename encoding internally
	workflow, err := loadWorkflowDetail(workflowID)
	if err != nil {
		data.Error = fmt.Sprintf("Error loading workflow: %v", err)
	} else {
		data.Summary = workflow.Summary
		data.Metadata = workflow.Metadata
	}

	if err := tmpl.ExecuteTemplate(w, "workflow.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleTasks(w http.ResponseWriter, r *http.Request, tmpl *template.Template) {
	// TODO: Implement task listing
	if _, err := fmt.Fprintf(w, "Tasks page - Coming soon!"); err != nil {
		log.Printf("Warning: failed to write response: %v", err)
	}
}

// loadWorkflows loads all workflow summaries from ~/.bob/state/
func loadWorkflows() ([]WorkflowSummary, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	workflowsDir := filepath.Join(homeDir, ".bob", "state")
	entries, err := os.ReadDir(workflowsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []WorkflowSummary{}, nil
		}
		return nil, err
	}

	var workflows []WorkflowSummary
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(workflowsDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Error reading %s: %v", entry.Name(), err)
			continue
		}

		var state WorkflowState
		if err := json.Unmarshal(data, &state); err != nil {
			log.Printf("Error parsing %s: %v", entry.Name(), err)
			continue
		}

		workflows = append(workflows, WorkflowSummary{
			WorkflowID:   state.WorkflowID,
			Workflow:     state.Workflow,
			CurrentStep:  state.CurrentStep,
			WorktreePath: state.WorktreePath,
		})
	}

	// Sort by workflow ID (alphabetical)
	sort.Slice(workflows, func(i, j int) bool {
		return workflows[i].WorkflowID < workflows[j].WorkflowID
	})

	return workflows, nil
}

// loadWorkflowDetail loads full details for a specific workflow
// Takes the real workflow ID and handles safe filename encoding internally
func loadWorkflowDetail(workflowID string) (*WorkflowDetailData, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// Load state file directly
	stateDir := filepath.Join(homeDir, ".bob", "state")
	filename := workflowIDToFilename(workflowID)
	filePath := filepath.Join(stateDir, filename)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var state WorkflowState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &WorkflowDetailData{
		Summary: WorkflowSummary{
			WorkflowID:   state.WorkflowID,
			Workflow:     state.Workflow,
			CurrentStep:  state.CurrentStep,
			WorktreePath: state.WorktreePath,
			// Removed fields: TaskDescription, LoopCount, IssueCount, LastUpdate, StartedAt
		},
		Metadata: make(map[string]interface{}), // Empty metadata
	}, nil
}

// countTasks counts tasks in .bob/issues/ across all repositories
func countTasks() (int, error) {
	// TODO: Implement task counting from .bob/issues/
	// For now, return 0
	return 0, nil
}
