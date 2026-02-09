package main

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// CreateMCPServer creates and configures the MCP server
func CreateMCPServer() *server.MCPServer {
	s := server.NewMCPServer("bob", version)

	stateManager := NewStateManager()
	taskManager := NewTaskManager()

	// Workflow tools
	registerWorkflowTools(s, stateManager, taskManager)

	// Task tools
	registerTaskTools(s, taskManager)

	return s
}

func registerWorkflowTools(s *server.MCPServer, stateManager *StateManager, taskManager *TaskManager) {
	// workflow_list_workflows
	s.AddTool(
		mcp.NewTool("workflow_list_workflows",
			mcp.WithDescription("List all available workflow types (keywords). Includes embedded workflows and custom workflows from .bob/workflows/ if repoPath provided."),
			mcp.WithString("repoPath",
				mcp.Description("Optional: Git repository path to check for custom workflows in .bob/workflows/"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			repoPath := request.GetString("repoPath", "")
			workflows, err := ListWorkflows(repoPath)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			data, _ := json.Marshal(map[string]interface{}{"workflows": workflows})
			return mcp.NewToolResultText(string(data)), nil
		},
	)

	// workflow_get_definition
	s.AddTool(
		mcp.NewTool("workflow_get_definition",
			mcp.WithDescription("Get the full definition of a workflow by keyword (e.g., 'brainstorm'). Checks for custom workflows in .bob/workflows/ if repoPath provided."),
			mcp.WithString("workflow",
				mcp.Required(),
				mcp.Description("Workflow keyword (e.g., 'brainstorm', 'code-review', 'performance')"),
			),
			mcp.WithString("repoPath",
				mcp.Description("Optional: Git repository path to check for custom workflows"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			workflow, err := request.RequireString("workflow")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			repoPath := request.GetString("repoPath", "")

			var def *WorkflowDefinition
			if repoPath != "" {
				def, err = GetWorkflowDefinition(workflow, repoPath)
			} else {
				def, err = GetWorkflowDefinition(workflow)
			}

			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			data, _ := json.Marshal(def)
			return mcp.NewToolResultText(string(data)), nil
		},
	)

	// workflow_register
	s.AddTool(
		mcp.NewTool("workflow_register",
			mcp.WithDescription("Register a new workflow session in a git worktree. Creates workflow state and initializes tracking."),
			mcp.WithString("workflow",
				mcp.Required(),
				mcp.Description("Workflow keyword (e.g., 'brainstorm', 'code-review', 'performance')"),
			),
			mcp.WithString("worktreePath",
				mcp.Required(),
				mcp.Description("Absolute path to the git worktree where workflow will run"),
			),
			mcp.WithString("taskDescription",
				mcp.Description("Optional: Description of the specific task to be prefixed to initial guidance"),
			),
			mcp.WithString("sessionID",
				mcp.Description("Optional: Session identifier for tracking"),
			),
			mcp.WithString("agentID",
				mcp.Description("Optional: Agent identifier"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			workflow, _ := request.RequireString("workflow")
			worktreePath, _ := request.RequireString("worktreePath")
			taskDescription := request.GetString("taskDescription", "")
			sessionID := request.GetString("sessionID", "")
			agentID := request.GetString("agentID", "")

			result, err := stateManager.Register(workflow, worktreePath, taskDescription, sessionID, agentID)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			data, _ := json.Marshal(result)
			return mcp.NewToolResultText(string(data)), nil
		},
	)

	// workflow_report_progress
	s.AddTool(
		mcp.NewTool("workflow_report_progress",
			mcp.WithDescription("Report progress completion for current workflow step and transition to next step. Stores metadata for state management."),
			mcp.WithString("worktreePath",
				mcp.Required(),
				mcp.Description("Absolute path to the git worktree"),
			),
			mcp.WithString("currentStep",
				mcp.Required(),
				mcp.Description("Name of the step being completed"),
			),
			mcp.WithObject("metadata",
				mcp.Description("Optional: Key-value metadata to store (e.g., metrics, findings, decisions)"),
			),
			mcp.WithString("sessionID",
				mcp.Description("Optional: Session identifier"),
			),
			mcp.WithString("agentID",
				mcp.Description("Optional: Agent identifier"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			worktreePath, _ := request.RequireString("worktreePath")
			currentStep, _ := request.RequireString("currentStep")

			args := request.GetArguments()
			metadata, _ := args["metadata"].(map[string]interface{})
			sessionID := request.GetString("sessionID", "")
			agentID := request.GetString("agentID", "")

			result, err := stateManager.ReportProgress(worktreePath, currentStep, metadata, sessionID, agentID)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			data, _ := json.Marshal(result)
			return mcp.NewToolResultText(string(data)), nil
		},
	)

	// workflow_get_guidance
	s.AddTool(
		mcp.NewTool("workflow_get_guidance",
			mcp.WithDescription("Get guidance prompt for current workflow step. Call at start of each step to retrieve context, instructions, and stored state."),
			mcp.WithString("worktreePath",
				mcp.Required(),
				mcp.Description("Absolute path to the git worktree"),
			),
			mcp.WithString("sessionID",
				mcp.Description("Optional: Session identifier"),
			),
			mcp.WithString("agentID",
				mcp.Description("Optional: Agent identifier"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			worktreePath, _ := request.RequireString("worktreePath")
			sessionID := request.GetString("sessionID", "")
			agentID := request.GetString("agentID", "")

			result, err := stateManager.GetGuidance(worktreePath, sessionID, agentID)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			data, _ := json.Marshal(result)
			return mcp.NewToolResultText(string(data)), nil
		},
	)

	// workflow_record_issues
	s.AddTool(
		mcp.NewTool("workflow_record_issues",
			mcp.WithDescription("Record issues found during workflow execution for tracking and resolution."),
			mcp.WithString("worktreePath",
				mcp.Required(),
				mcp.Description("Absolute path to the git worktree"),
			),
			mcp.WithString("currentStep",
				mcp.Required(),
				mcp.Description("Current workflow step where issues were found"),
			),
			mcp.WithArray("issues",
				mcp.Required(),
				mcp.Description("Array of issue objects with fields: phase, description, severity, file, line"),
			),
			mcp.WithString("sessionID",
				mcp.Description("Optional: Session identifier"),
			),
			mcp.WithString("agentID",
				mcp.Description("Optional: Agent identifier"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			worktreePath, _ := request.RequireString("worktreePath")

			args := request.GetArguments()
			issuesRaw, _ := args["issues"].([]interface{})
			sessionID := request.GetString("sessionID", "")
			agentID := request.GetString("agentID", "")

			// Convert to []Issue
			var issues []Issue
			for _, issueRaw := range issuesRaw {
				issueMap, ok := issueRaw.(map[string]interface{})
				if !ok {
					continue
				}
				issue := Issue{
					Phase:       getString(issueMap, "phase"),
					Description: getString(issueMap, "description"),
					Severity:    getString(issueMap, "severity"),
					File:        getString(issueMap, "file"),
					Line:        getInt(issueMap, "line"),
				}
				issues = append(issues, issue)
			}

			// Get current step for RecordIssues
			currentStep := request.GetString("currentStep", "")
			result, err := stateManager.RecordIssues(worktreePath, currentStep, issues, sessionID, agentID)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			data, _ := json.Marshal(result)
			return mcp.NewToolResultText(string(data)), nil
		},
	)

	// workflow_get_status
	s.AddTool(
		mcp.NewTool("workflow_get_status",
			mcp.WithDescription("Get current workflow status including step, progress history, and metadata."),
			mcp.WithString("worktreePath",
				mcp.Required(),
				mcp.Description("Absolute path to the git worktree"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			worktreePath, _ := request.RequireString("worktreePath")
			sessionID := request.GetString("sessionID", "")
			agentID := request.GetString("agentID", "")

			result, err := stateManager.GetStatus(worktreePath, sessionID, agentID)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			data, _ := json.Marshal(result)
			return mcp.NewToolResultText(string(data)), nil
		},
	)

	// workflow_list_agents
	s.AddTool(
		mcp.NewTool("workflow_list_agents",
			mcp.WithDescription("List all active agents in the workflow session."),
			mcp.WithString("worktreePath",
				mcp.Required(),
				mcp.Description("Absolute path to the git worktree"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			worktreePath, _ := request.RequireString("worktreePath")
			sessionID := request.GetString("sessionID", "")

			result, err := stateManager.ListAgents(sessionID, worktreePath)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			data, _ := json.Marshal(result)
			return mcp.NewToolResultText(string(data)), nil
		},
	)

	// workflow_get_session_status
	s.AddTool(
		mcp.NewTool("workflow_get_session_status",
			mcp.WithDescription("Get session-specific status and progress for a workflow."),
			mcp.WithString("worktreePath",
				mcp.Required(),
				mcp.Description("Absolute path to the git worktree"),
			),
			mcp.WithString("sessionID",
				mcp.Required(),
				mcp.Description("Session identifier"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sessionID, _ := request.RequireString("sessionID")

			result, err := stateManager.GetSessionStatus(sessionID)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			data, _ := json.Marshal(result)
			return mcp.NewToolResultText(string(data)), nil
		},
	)
}

func registerTaskTools(s *server.MCPServer, taskManager *TaskManager) {
	// task_create
	s.AddTool(
		mcp.NewTool("task_create",
			mcp.WithDescription("Create a new task in the .bob/issues/ directory on the bob branch."),
			mcp.WithString("repoPath",
				mcp.Required(),
				mcp.Description("Git repository path"),
			),
			mcp.WithString("title",
				mcp.Required(),
				mcp.Description("Task title"),
			),
			mcp.WithString("description",
				mcp.Required(),
				mcp.Description("Task description"),
			),
			mcp.WithString("priority",
				mcp.Description("Priority: low, medium, high, critical"),
			),
			mcp.WithArray("labels",
				mcp.Description("Array of label strings"),
			),
			mcp.WithArray("dependencies",
				mcp.Description("Array of task IDs this task depends on"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			repoPath, _ := request.RequireString("repoPath")
			title, _ := request.RequireString("title")
			description, _ := request.RequireString("description")
			priority := request.GetString("priority", "medium")
			taskType := request.GetString("taskType", "task")

			args := request.GetArguments()
			labelsRaw, _ := args["labels"].([]interface{})
			labels := toStringSlice(labelsRaw)

			metadata := make(map[string]interface{})

			// Create the task first
			result, err := taskManager.CreateTask(repoPath, title, description, taskType, priority, labels, metadata)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			// Extract task ID from result
			taskMap, _ := result["task"].(Task)
			taskID := taskMap.ID

			// Add dependencies if provided
			if depsRaw, ok := args["dependencies"].([]interface{}); ok {
				deps := toStringSlice(depsRaw)
				for _, depID := range deps {
					// Add dependency: this task depends on depID
					_, err := taskManager.AddDependency(repoPath, taskID, depID)
					if err != nil {
						// Log error but don't fail task creation
						result["dependencyErrors"] = append(
							result["dependencyErrors"].([]string),
							err.Error(),
						)
					}
				}
			}

			data, _ := json.Marshal(result)
			return mcp.NewToolResultText(string(data)), nil
		},
	)

	// task_get
	s.AddTool(
		mcp.NewTool("task_get",
			mcp.WithDescription("Get a task by ID from the .bob/issues/ directory."),
			mcp.WithString("repoPath",
				mcp.Required(),
				mcp.Description("Git repository path"),
			),
			mcp.WithString("taskId",
				mcp.Required(),
				mcp.Description("Task ID (e.g., 'task-001')"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			repoPath, _ := request.RequireString("repoPath")
			taskID, _ := request.RequireString("taskId")

			task, err := taskManager.GetTask(repoPath, taskID)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			data, _ := json.Marshal(task)
			return mcp.NewToolResultText(string(data)), nil
		},
	)

	// task_list
	s.AddTool(
		mcp.NewTool("task_list",
			mcp.WithDescription("List all tasks from .bob/issues/ directory with optional filters."),
			mcp.WithString("repoPath",
				mcp.Required(),
				mcp.Description("Git repository path"),
			),
			mcp.WithString("status",
				mcp.Description("Filter by status: open, in_progress, completed, blocked"),
			),
			mcp.WithString("priority",
				mcp.Description("Filter by priority: low, medium, high, critical"),
			),
			mcp.WithString("label",
				mcp.Description("Filter by label"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			repoPath, _ := request.RequireString("repoPath")
			state := request.GetString("status", "")
			priority := request.GetString("priority", "")
			taskType := request.GetString("taskType", "")
			assignee := request.GetString("assignee", "")

			args := request.GetArguments()
			tagsRaw, _ := args["labels"].([]interface{})
			tags := toStringSlice(tagsRaw)

			tasks, err := taskManager.ListTasks(repoPath, state, priority, taskType, assignee, tags)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			data, _ := json.Marshal(tasks)
			return mcp.NewToolResultText(string(data)), nil
		},
	)

	// task_update
	s.AddTool(
		mcp.NewTool("task_update",
			mcp.WithDescription("Update a task's fields (status, priority, labels, etc.)."),
			mcp.WithString("repoPath",
				mcp.Required(),
				mcp.Description("Git repository path"),
			),
			mcp.WithString("taskId",
				mcp.Required(),
				mcp.Description("Task ID to update"),
			),
			mcp.WithString("status",
				mcp.Description("New status: open, in_progress, completed, blocked"),
			),
			mcp.WithString("priority",
				mcp.Description("New priority: low, medium, high, critical"),
			),
			mcp.WithArray("labels",
				mcp.Description("New labels array"),
			),
			mcp.WithString("assignee",
				mcp.Description("Assignee name or ID"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			repoPath, _ := request.RequireString("repoPath")
			taskID, _ := request.RequireString("taskId")

			args := request.GetArguments()
			updates := make(map[string]interface{})

			if status, ok := args["status"].(string); ok {
				updates["state"] = status
			}
			if priority, ok := args["priority"].(string); ok {
				updates["priority"] = priority
			}
			if labelsRaw, ok := args["labels"].([]interface{}); ok {
				updates["tags"] = toStringSlice(labelsRaw)
			}
			if assignee, ok := args["assignee"].(string); ok {
				updates["assignee"] = assignee
			}

			task, err := taskManager.UpdateTask(repoPath, taskID, updates)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			data, _ := json.Marshal(task)
			return mcp.NewToolResultText(string(data)), nil
		},
	)

	// task_add_dependency
	s.AddTool(
		mcp.NewTool("task_add_dependency",
			mcp.WithDescription("Add a dependency relationship between two tasks."),
			mcp.WithString("repoPath",
				mcp.Required(),
				mcp.Description("Git repository path"),
			),
			mcp.WithString("taskId",
				mcp.Required(),
				mcp.Description("Task ID that depends on another"),
			),
			mcp.WithString("dependsOn",
				mcp.Required(),
				mcp.Description("Task ID that must be completed first"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			repoPath, _ := request.RequireString("repoPath")
			taskID, _ := request.RequireString("taskId")
			dependsOn, _ := request.RequireString("dependsOn")

			result, err := taskManager.AddDependency(repoPath, taskID, dependsOn)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			data, _ := json.Marshal(result)
			return mcp.NewToolResultText(string(data)), nil
		},
	)

	// task_add_comment
	s.AddTool(
		mcp.NewTool("task_add_comment",
			mcp.WithDescription("Add a comment to a task for notes, updates, or discussion."),
			mcp.WithString("repoPath",
				mcp.Required(),
				mcp.Description("Git repository path"),
			),
			mcp.WithString("taskId",
				mcp.Required(),
				mcp.Description("Task ID to comment on"),
			),
			mcp.WithString("comment",
				mcp.Required(),
				mcp.Description("Comment text"),
			),
			mcp.WithString("author",
				mcp.Description("Comment author name"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			repoPath, _ := request.RequireString("repoPath")
			taskID, _ := request.RequireString("taskId")
			comment, _ := request.RequireString("comment")
			author := request.GetString("author", "")

			task, err := taskManager.AddComment(repoPath, taskID, author, comment)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			data, _ := json.Marshal(task)
			return mcp.NewToolResultText(string(data)), nil
		},
	)

	// task_get_ready
	s.AddTool(
		mcp.NewTool("task_get_ready",
			mcp.WithDescription("Get tasks that are ready to work on (open status, no blocking dependencies)."),
			mcp.WithString("repoPath",
				mcp.Required(),
				mcp.Description("Git repository path"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			repoPath, _ := request.RequireString("repoPath")

			tasks, err := taskManager.GetReadyTasks(repoPath)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			data, _ := json.Marshal(map[string]interface{}{"tasks": tasks})
			return mcp.NewToolResultText(string(data)), nil
		},
	)

	// task_set_workflow_state
	s.AddTool(
		mcp.NewTool("task_set_workflow_state",
			mcp.WithDescription("Set or update workflow state key-value pairs for a task."),
			mcp.WithString("repoPath",
				mcp.Required(),
				mcp.Description("Git repository path"),
			),
			mcp.WithString("taskId",
				mcp.Required(),
				mcp.Description("Task ID"),
			),
			mcp.WithObject("state",
				mcp.Required(),
				mcp.Description("Key-value pairs to set in workflow state"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			repoPath, _ := request.RequireString("repoPath")
			taskID, _ := request.RequireString("taskId")

			args := request.GetArguments()
			state, _ := args["state"].(map[string]interface{})

			result, err := taskManager.SetWorkflowState(repoPath, taskID, state)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			data, _ := json.Marshal(result)
			return mcp.NewToolResultText(string(data)), nil
		},
	)

	// task_get_workflow_state
	s.AddTool(
		mcp.NewTool("task_get_workflow_state",
			mcp.WithDescription("Get all workflow state for a task."),
			mcp.WithString("repoPath",
				mcp.Required(),
				mcp.Description("Git repository path"),
			),
			mcp.WithString("taskId",
				mcp.Required(),
				mcp.Description("Task ID"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			repoPath, _ := request.RequireString("repoPath")
			taskID, _ := request.RequireString("taskId")

			result, err := taskManager.GetWorkflowState(repoPath, taskID)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			data, _ := json.Marshal(result)
			return mcp.NewToolResultText(string(data)), nil
		},
	)

	// task_delete_workflow_state_key
	s.AddTool(
		mcp.NewTool("task_delete_workflow_state_key",
			mcp.WithDescription("Delete a specific key from task's workflow state."),
			mcp.WithString("repoPath",
				mcp.Required(),
				mcp.Description("Git repository path"),
			),
			mcp.WithString("taskId",
				mcp.Required(),
				mcp.Description("Task ID"),
			),
			mcp.WithString("key",
				mcp.Required(),
				mcp.Description("State key to delete"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			repoPath, _ := request.RequireString("repoPath")
			taskID, _ := request.RequireString("taskId")
			key, _ := request.RequireString("key")

			result, err := taskManager.DeleteWorkflowStateKey(repoPath, taskID, key)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			data, _ := json.Marshal(result)
			return mcp.NewToolResultText(string(data)), nil
		},
	)
}

// Helper functions
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return 0
}

func toStringSlice(arr []interface{}) []string {
	result := make([]string, 0, len(arr))
	for _, v := range arr {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}
	return result
}
