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
	"sort"
	"strings"
	"time"
)

//go:embed templates/*.html
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

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

// WorkflowSummary contains summary info for a workflow
type WorkflowSummary struct {
	WorkflowID      string
	Workflow        string
	TaskDescription string
	CurrentStep     string
	LoopCount       int
	IssueCount      int
	WorktreePath    string
	LastUpdate      time.Time
	StartedAt       time.Time
}

// WorkflowDetailData contains data for the workflow detail page
type WorkflowDetailData struct {
	Summary         WorkflowSummary
	ProgressHistory []ProgressEntry
	Issues          []Issue
	Metadata        map[string]interface{}
	Error           string
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

	data := WorkflowDetailData{}

	// Load workflow details
	workflow, err := loadWorkflowDetail(workflowID)
	if err != nil {
		data.Error = fmt.Sprintf("Error loading workflow: %v", err)
	} else {
		data.Summary = workflow.Summary
		data.ProgressHistory = workflow.ProgressHistory
		data.Issues = workflow.Issues
		data.Metadata = workflow.Metadata
	}

	if err := tmpl.ExecuteTemplate(w, "workflow.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleTasks(w http.ResponseWriter, r *http.Request, tmpl *template.Template) {
	// TODO: Implement task listing
	fmt.Fprintf(w, "Tasks page - Coming soon!")
}

// loadWorkflows loads all workflow summaries from ~/.claude/workflows/
func loadWorkflows() ([]WorkflowSummary, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	workflowsDir := filepath.Join(homeDir, ".claude", "workflows")
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
			WorkflowID:      state.WorkflowID,
			Workflow:        state.Workflow,
			TaskDescription: state.TaskDescription,
			CurrentStep:     state.CurrentStep,
			LoopCount:       state.LoopCount,
			IssueCount:      len(state.Issues),
			WorktreePath:    state.WorktreePath,
			LastUpdate:      state.UpdatedAt,
			StartedAt:       state.StartedAt,
		})
	}

	// Sort by last update (most recent first)
	sort.Slice(workflows, func(i, j int) bool {
		return workflows[i].LastUpdate.After(workflows[j].LastUpdate)
	})

	return workflows, nil
}

// loadWorkflowDetail loads full details for a specific workflow
func loadWorkflowDetail(workflowID string) (*WorkflowDetailData, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// Find the workflow file
	workflowsDir := filepath.Join(homeDir, ".claude", "workflows")
	filePath := filepath.Join(workflowsDir, workflowID+".json")

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
			WorkflowID:      state.WorkflowID,
			Workflow:        state.Workflow,
			TaskDescription: state.TaskDescription,
			CurrentStep:     state.CurrentStep,
			LoopCount:       state.LoopCount,
			IssueCount:      len(state.Issues),
			WorktreePath:    state.WorktreePath,
			LastUpdate:      state.UpdatedAt,
			StartedAt:       state.StartedAt,
		},
		ProgressHistory: state.ProgressHistory,
		Issues:          state.Issues,
		Metadata:        state.Metadata,
	}, nil
}

// countTasks counts tasks in .bob/issues/ across all repositories
func countTasks() (int, error) {
	// TODO: Implement task counting from .bob/issues/
	// For now, return 0
	return 0, nil
}
