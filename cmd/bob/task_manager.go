package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

// Task represents a workflow task
type Task struct {
	ID            string                 `json:"id"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description"`
	Type          string                 `json:"type"`     // feature, bug, chore, refactor, docs, test
	Priority      string                 `json:"priority"` // high, medium, low
	State         string                 `json:"state"`    // pending, in_progress, blocked, completed, cancelled
	Assignee      string                 `json:"assignee,omitempty"`
	Tags          []string               `json:"tags,omitempty"`
	Blocks        []string               `json:"blocks,omitempty"`    // Task IDs this task blocks
	BlockedBy     []string               `json:"blockedBy,omitempty"` // Task IDs blocking this task
	Comments      []Comment              `json:"comments,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	WorkflowState map[string]string      `json:"workflowState,omitempty"` // Key-value workflow state (workflow, currentStep, worktreePath, etc.)
	CreatedAt     time.Time              `json:"createdAt"`
	UpdatedAt     time.Time              `json:"updatedAt"`
	CompletedAt   *time.Time             `json:"completedAt,omitempty"`
}

// Comment represents a task comment
type Comment struct {
	ID        string    `json:"id"`
	Author    string    `json:"author"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
}

// GitHubFile represents a file in GitHub API
type GitHubFile struct {
	Content string `json:"content"`
	SHA     string `json:"sha"`
}

// GitHubCommit represents a commit request
type GitHubCommit struct {
	Message string `json:"message"`
	Content string `json:"content"`
	Branch  string `json:"branch"`
	SHA     string `json:"sha,omitempty"`
}

// GitHubRepo represents repository information
type GitHubRepo struct {
	Owner string
	Name  string
	Token string
}

// TaskManager manages tasks for git repositories
type TaskManager struct {
	branchName string
	issuesDir  string
	client     *http.Client
}

// NewTaskManager creates a new task manager
func NewTaskManager() *TaskManager {
	return &TaskManager{
		branchName: "bob",
		issuesDir:  ".bob/issues",
		client:     &http.Client{Timeout: 30 * time.Second},
	}
}

// getRepoPath returns the git repository root path
func (tm *TaskManager) getRepoPath(repoPath string) (string, error) {
	if repoPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		repoPath = cwd
	}

	// Get the top-level git directory
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository: %s", repoPath)
	}

	return strings.TrimSpace(string(output)), nil
}

// getGitHubRepo extracts GitHub repository information from git remote
func (tm *TaskManager) getGitHubRepo(repoPath string) (*GitHubRepo, error) {
	cmd := exec.Command("git", "-C", repoPath, "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git remote: %w", err)
	}

	remoteURL := strings.TrimSpace(string(output))

	// Parse GitHub URL (supports both HTTPS and SSH formats)
	// HTTPS: https://github.com/owner/repo.git
	// SSH: git@github.com:owner/repo.git
	var owner, name string

	if strings.Contains(remoteURL, "github.com") {
		// Remove .git suffix
		remoteURL = strings.TrimSuffix(remoteURL, ".git")

		if strings.HasPrefix(remoteURL, "git@github.com:") {
			// SSH format
			parts := strings.TrimPrefix(remoteURL, "git@github.com:")
			ownerRepo := strings.Split(parts, "/")
			if len(ownerRepo) == 2 {
				owner, name = ownerRepo[0], ownerRepo[1]
			}
		} else if strings.Contains(remoteURL, "github.com/") {
			// HTTPS format
			parts := strings.Split(remoteURL, "github.com/")
			if len(parts) == 2 {
				ownerRepo := strings.Split(parts[1], "/")
				if len(ownerRepo) == 2 {
					owner, name = ownerRepo[0], ownerRepo[1]
				}
			}
		}
	}

	if owner == "" || name == "" {
		return nil, fmt.Errorf("not a GitHub repository or invalid remote URL: %s", remoteURL)
	}

	// Get GitHub token from environment
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable not set")
	}

	return &GitHubRepo{
		Owner: owner,
		Name:  name,
		Token: token,
	}, nil
}

// readFileFromGitHub reads a file from GitHub API
func (tm *TaskManager) readFileFromGitHub(repo *GitHubRepo, path string, branch string) ([]byte, string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s",
		repo.Owner, repo.Name, path, branch)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}

	req.Header.Set("Authorization", "Bearer "+repo.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := tm.client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, "", nil // File doesn't exist
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("GitHub API error %d: %s", resp.StatusCode, string(body))
	}

	var ghFile GitHubFile
	if err := json.NewDecoder(resp.Body).Decode(&ghFile); err != nil {
		return nil, "", err
	}

	// Decode base64 content
	content, err := base64.StdEncoding.DecodeString(ghFile.Content)
	if err != nil {
		return nil, "", err
	}

	return content, ghFile.SHA, nil
}

// writeFileToGitHub writes a file to GitHub API
func (tm *TaskManager) writeFileToGitHub(repo *GitHubRepo, path string, content []byte, message string, sha string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s",
		repo.Owner, repo.Name, path)

	// Encode content to base64
	encodedContent := base64.StdEncoding.EncodeToString(content)

	payload := map[string]interface{}{
		"message": message,
		"content": encodedContent,
		"branch":  tm.branchName,
	}

	if sha != "" {
		payload["sha"] = sha
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewReader(payloadBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+repo.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := tm.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ensureBranchExists ensures the branch exists on GitHub
func (tm *TaskManager) ensureBranchExists(repo *GitHubRepo, repoPath string) error {
	// Check if branch exists
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/branches/%s",
		repo.Owner, repo.Name, tm.branchName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+repo.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := tm.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusOK {
		// Branch exists
		return nil
	}

	// Branch doesn't exist, create it
	// Get the default branch's SHA
	defaultURL := fmt.Sprintf("https://api.github.com/repos/%s/%s", repo.Owner, repo.Name)
	req, err = http.NewRequest("GET", defaultURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+repo.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err = tm.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	var repoInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&repoInfo); err != nil {
		return err
	}

	defaultBranch := repoInfo["default_branch"].(string)

	// Get the default branch's SHA
	refURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/ref/heads/%s",
		repo.Owner, repo.Name, defaultBranch)
	req, err = http.NewRequest("GET", refURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+repo.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err = tm.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	var refInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&refInfo); err != nil {
		return err
	}

	objectInfo := refInfo["object"].(map[string]interface{})
	sha := objectInfo["sha"].(string)

	// Create new branch
	createURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/refs", repo.Owner, repo.Name)
	payload := map[string]interface{}{
		"ref": "refs/heads/" + tm.branchName,
		"sha": sha,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err = http.NewRequest("POST", createURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+repo.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err = tm.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create branch: %d: %s", resp.StatusCode, string(body))
	}

	// Branch created successfully - individual task files will be created as needed
	return nil
}

// listTaskFiles lists all task files in the issues directory
func (tm *TaskManager) listTaskFiles(repo *GitHubRepo) ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s",
		repo.Owner, repo.Name, tm.issuesDir, tm.branchName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+repo.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := tm.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		// Directory doesn't exist yet
		return []string{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error %d: %s", resp.StatusCode, string(body))
	}

	var files []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, err
	}

	var taskFiles []string
	for _, file := range files {
		if name, ok := file["name"].(string); ok && strings.HasSuffix(name, ".json") {
			taskFiles = append(taskFiles, name)
		}
	}

	return taskFiles, nil
}

// readTaskFile reads a single task file from GitHub
func (tm *TaskManager) readTaskFile(repo *GitHubRepo, filename string) (*Task, string, error) {
	path := fmt.Sprintf("%s/%s", tm.issuesDir, filename)
	data, sha, err := tm.readFileFromGitHub(repo, path, tm.branchName)
	if err != nil {
		return nil, "", err
	}

	if data == nil {
		return nil, "", nil
	}

	var task Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, "", err
	}

	return &task, sha, nil
}

// writeTaskFile writes a single task file to GitHub
func (tm *TaskManager) writeTaskFile(repo *GitHubRepo, task *Task, sha string) error {
	filename := fmt.Sprintf("%s.json", task.ID)
	path := fmt.Sprintf("%s/%s", tm.issuesDir, filename)

	data, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return err
	}

	message := fmt.Sprintf("Update task %s", task.ID)
	if sha == "" {
		message = fmt.Sprintf("Create task %s", task.ID)
	}

	// Write to GitHub
	return tm.writeFileToGitHub(repo, path, data, message, sha)
}

// deleteTaskFile deletes a single task file from GitHub
func (tm *TaskManager) deleteTaskFile(repo *GitHubRepo, taskID string, sha string) error {
	filename := fmt.Sprintf("%s.json", taskID)
	path := fmt.Sprintf("%s/%s", tm.issuesDir, filename)

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s",
		repo.Owner, repo.Name, path)

	payload := map[string]interface{}{
		"message": fmt.Sprintf("Delete task %s", taskID),
		"sha":     sha,
		"branch":  tm.branchName,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("DELETE", url, bytes.NewReader(payloadBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+repo.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := tm.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// loadTasks loads all tasks for a repository using GitHub API
// Note: This reads all task files - for individual operations use readTaskFile
func (tm *TaskManager) loadTasks(repoPath string) ([]Task, error) {
	repoRoot, err := tm.getRepoPath(repoPath)
	if err != nil {
		return nil, err
	}

	repo, err := tm.getGitHubRepo(repoRoot)
	if err != nil {
		return nil, err
	}

	// Ensure branch exists
	if err := tm.ensureBranchExists(repo, repoRoot); err != nil {
		return nil, err
	}

	// List all task files
	taskFiles, err := tm.listTaskFiles(repo)
	if err != nil {
		return nil, err
	}

	var tasks []Task
	for _, filename := range taskFiles {
		task, _, err := tm.readTaskFile(repo, filename)
		if err != nil {
			// Log error but continue with other tasks
			continue
		}
		if task != nil {
			tasks = append(tasks, *task)
		}
	}

	return tasks, nil
}

// generateTaskID generates a unique task ID
func (tm *TaskManager) generateTaskID(tasks []Task) string {
	maxID := 0
	for _, task := range tasks {
		var id int
		if _, err := fmt.Sscanf(task.ID, "task-%d", &id); err == nil {
			if id > maxID {
				maxID = id
			}
		}
	}
	return fmt.Sprintf("task-%d", maxID+1)
}

// CreateTask creates a new task
func (tm *TaskManager) CreateTask(repoPath, title, description, taskType, priority string, tags []string, metadata map[string]interface{}) (map[string]interface{}, error) {
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}

	repoRoot, err := tm.getRepoPath(repoPath)
	if err != nil {
		return nil, err
	}

	repo, err := tm.getGitHubRepo(repoRoot)
	if err != nil {
		return nil, err
	}

	// Ensure branch exists
	if err := tm.ensureBranchExists(repo, repoRoot); err != nil {
		return nil, err
	}

	// Load existing tasks to generate unique ID
	tasks, err := tm.loadTasks(repoPath)
	if err != nil {
		return nil, err
	}

	// Validate type
	validTypes := map[string]bool{"feature": true, "bug": true, "chore": true, "refactor": true, "docs": true, "test": true}
	if taskType != "" && !validTypes[taskType] {
		taskType = "feature" // default
	} else if taskType == "" {
		taskType = "feature"
	}

	// Validate priority
	validPriorities := map[string]bool{"high": true, "medium": true, "low": true}
	if priority != "" && !validPriorities[priority] {
		priority = "medium" // default
	} else if priority == "" {
		priority = "medium"
	}

	task := Task{
		ID:          tm.generateTaskID(tasks),
		Title:       title,
		Description: description,
		Type:        taskType,
		Priority:    priority,
		State:       "pending",
		Tags:        tags,
		Blocks:      []string{},
		BlockedBy:   []string{},
		Comments:    []Comment{},
		Metadata:    metadata,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Write task to its own file (no SHA since it's a new file)
	if err := tm.writeTaskFile(repo, &task, ""); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"task":    task,
		"message": fmt.Sprintf("Created task %s: %s", task.ID, task.Title),
	}, nil
}

// GetTask retrieves a specific task
func (tm *TaskManager) GetTask(repoPath, taskID string) (map[string]interface{}, error) {
	repoRoot, err := tm.getRepoPath(repoPath)
	if err != nil {
		return nil, err
	}

	repo, err := tm.getGitHubRepo(repoRoot)
	if err != nil {
		return nil, err
	}

	filename := fmt.Sprintf("%s.json", taskID)
	task, _, err := tm.readTaskFile(repo, filename)
	if err != nil {
		return nil, err
	}

	if task == nil {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	return map[string]interface{}{
		"task": *task,
	}, nil
}

// ListTasks lists tasks with optional filters
func (tm *TaskManager) ListTasks(repoPath, state, priority, taskType, assignee string, tags []string) (map[string]interface{}, error) {
	tasks, err := tm.loadTasks(repoPath)
	if err != nil {
		return nil, err
	}

	// Apply filters
	var filtered []Task
	for _, task := range tasks {
		if state != "" && task.State != state {
			continue
		}
		if priority != "" && task.Priority != priority {
			continue
		}
		if taskType != "" && task.Type != taskType {
			continue
		}
		if assignee != "" && task.Assignee != assignee {
			continue
		}
		if len(tags) > 0 {
			hasTag := false
			for _, filterTag := range tags {
				for _, taskTag := range task.Tags {
					if taskTag == filterTag {
						hasTag = true
						break
					}
				}
				if hasTag {
					break
				}
			}
			if !hasTag {
				continue
			}
		}

		filtered = append(filtered, task)
	}

	// Sort by priority (high > medium > low) then by created date
	sort.Slice(filtered, func(i, j int) bool {
		priorityOrder := map[string]int{"high": 3, "medium": 2, "low": 1}
		if priorityOrder[filtered[i].Priority] != priorityOrder[filtered[j].Priority] {
			return priorityOrder[filtered[i].Priority] > priorityOrder[filtered[j].Priority]
		}
		return filtered[i].CreatedAt.Before(filtered[j].CreatedAt)
	})

	return map[string]interface{}{
		"tasks": filtered,
		"count": len(filtered),
	}, nil
}

// UpdateTask updates a task
func (tm *TaskManager) UpdateTask(repoPath, taskID string, updates map[string]interface{}) (map[string]interface{}, error) {
	repoRoot, err := tm.getRepoPath(repoPath)
	if err != nil {
		return nil, err
	}

	repo, err := tm.getGitHubRepo(repoRoot)
	if err != nil {
		return nil, err
	}

	// Read the task file
	filename := fmt.Sprintf("%s.json", taskID)
	task, sha, err := tm.readTaskFile(repo, filename)
	if err != nil {
		return nil, err
	}

	if task == nil {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	// Apply updates
	if title, ok := updates["title"].(string); ok && title != "" {
		task.Title = title
	}
	if desc, ok := updates["description"].(string); ok {
		task.Description = desc
	}
	if taskType, ok := updates["type"].(string); ok && taskType != "" {
		task.Type = taskType
	}
	if priority, ok := updates["priority"].(string); ok && priority != "" {
		task.Priority = priority
	}
	if state, ok := updates["state"].(string); ok && state != "" {
		task.State = state
		if state == "completed" {
			now := time.Now()
			task.CompletedAt = &now
		}
	}
	if assignee, ok := updates["assignee"].(string); ok {
		task.Assignee = assignee
	}

	task.UpdatedAt = time.Now()

	// Write back the updated task
	if err := tm.writeTaskFile(repo, task, sha); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"task":    *task,
		"message": fmt.Sprintf("Updated task %s", taskID),
	}, nil
}

// DeleteTask deletes a task and cleans up dependencies in related tasks
func (tm *TaskManager) DeleteTask(repoPath, taskID string) (map[string]interface{}, error) {
	repoRoot, err := tm.getRepoPath(repoPath)
	if err != nil {
		return nil, err
	}

	repo, err := tm.getGitHubRepo(repoRoot)
	if err != nil {
		return nil, err
	}

	// Read the task to get its SHA and validate it exists
	filename := fmt.Sprintf("%s.json", taskID)
	task, sha, err := tm.readTaskFile(repo, filename)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	// Clean up dependencies: load all tasks and remove references
	allTasks, err := tm.loadTasks(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks for dependency cleanup: %w", err)
	}

	// Track tasks that need updating
	var tasksToUpdate []struct {
		task *Task
		sha  string
	}

	// Find tasks that reference the deleted task
	for i := range allTasks {
		t := &allTasks[i]
		if t.ID == taskID {
			continue // Skip the task being deleted
		}

		modified := false

		// Remove from Blocks list
		if contains(t.Blocks, taskID) {
			t.Blocks = removeFromSlice(t.Blocks, taskID)
			modified = true
		}

		// Remove from BlockedBy list
		if contains(t.BlockedBy, taskID) {
			t.BlockedBy = removeFromSlice(t.BlockedBy, taskID)
			modified = true
		}

		if modified {
			// Read the task file to get current SHA
			taskFilename := fmt.Sprintf("%s.json", t.ID)
			_, taskSHA, err := tm.readTaskFile(repo, taskFilename)
			if err != nil {
				return nil, fmt.Errorf("failed to read task %s for update: %w", t.ID, err)
			}
			t.UpdatedAt = time.Now()
			tasksToUpdate = append(tasksToUpdate, struct {
				task *Task
				sha  string
			}{t, taskSHA})
		}
	}

	// Update all affected tasks first (before deleting)
	for _, tu := range tasksToUpdate {
		if err := tm.writeTaskFile(repo, tu.task, tu.sha); err != nil {
			return nil, fmt.Errorf("failed to update task %s during dependency cleanup: %w", tu.task.ID, err)
		}
	}

	// Now delete the task file
	if err := tm.deleteTaskFile(repo, taskID, sha); err != nil {
		return nil, fmt.Errorf("failed to delete task file: %w", err)
	}

	return map[string]interface{}{
		"taskId":                taskID,
		"title":                 task.Title,
		"message":               fmt.Sprintf("Deleted task %s and cleaned up %d dependent task(s)", taskID, len(tasksToUpdate)),
		"dependenciesCleanedUp": len(tasksToUpdate),
	}, nil
}

// removeFromSlice removes an item from a string slice and returns the new slice
func removeFromSlice(slice []string, item string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

// AddDependency adds a dependency between tasks
func (tm *TaskManager) AddDependency(repoPath, taskID, blocksTaskID string) (map[string]interface{}, error) {
	repoRoot, err := tm.getRepoPath(repoPath)
	if err != nil {
		return nil, err
	}

	repo, err := tm.getGitHubRepo(repoRoot)
	if err != nil {
		return nil, err
	}

	// Read first task
	filename1 := fmt.Sprintf("%s.json", taskID)
	task1, sha1, err := tm.readTaskFile(repo, filename1)
	if err != nil {
		return nil, err
	}
	if task1 == nil {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	// Read second task
	filename2 := fmt.Sprintf("%s.json", blocksTaskID)
	task2, sha2, err := tm.readTaskFile(repo, filename2)
	if err != nil {
		return nil, err
	}
	if task2 == nil {
		return nil, fmt.Errorf("task not found: %s", blocksTaskID)
	}

	// Add to blocks list
	if !contains(task1.Blocks, blocksTaskID) {
		task1.Blocks = append(task1.Blocks, blocksTaskID)
	}

	// Add to blockedBy list
	if !contains(task2.BlockedBy, taskID) {
		task2.BlockedBy = append(task2.BlockedBy, taskID)
	}

	task1.UpdatedAt = time.Now()
	task2.UpdatedAt = time.Now()

	// Write both tasks back
	if err := tm.writeTaskFile(repo, task1, sha1); err != nil {
		return nil, err
	}

	if err := tm.writeTaskFile(repo, task2, sha2); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message": fmt.Sprintf("Task %s now blocks %s", taskID, blocksTaskID),
	}, nil
}

// AddComment adds a comment to a task
func (tm *TaskManager) AddComment(repoPath, taskID, author, text string) (map[string]interface{}, error) {
	repoRoot, err := tm.getRepoPath(repoPath)
	if err != nil {
		return nil, err
	}

	repo, err := tm.getGitHubRepo(repoRoot)
	if err != nil {
		return nil, err
	}

	// Read the task file
	filename := fmt.Sprintf("%s.json", taskID)
	task, sha, err := tm.readTaskFile(repo, filename)
	if err != nil {
		return nil, err
	}

	if task == nil {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	comment := Comment{
		ID:        fmt.Sprintf("comment-%d", len(task.Comments)+1),
		Author:    author,
		Text:      text,
		CreatedAt: time.Now(),
	}

	task.Comments = append(task.Comments, comment)
	task.UpdatedAt = time.Now()

	// Write back the updated task
	if err := tm.writeTaskFile(repo, task, sha); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"comment": comment,
		"message": fmt.Sprintf("Added comment to task %s", taskID),
	}, nil
}

// GetReadyTasks returns tasks that are ready to work on (no blockers)
func (tm *TaskManager) GetReadyTasks(repoPath string) (map[string]interface{}, error) {
	tasks, err := tm.loadTasks(repoPath)
	if err != nil {
		return nil, err
	}

	var ready []Task
	for _, task := range tasks {
		if task.State == "pending" && len(task.BlockedBy) == 0 {
			ready = append(ready, task)
		}
	}

	// Sort by priority
	sort.Slice(ready, func(i, j int) bool {
		priorityOrder := map[string]int{"high": 3, "medium": 2, "low": 1}
		if priorityOrder[ready[i].Priority] != priorityOrder[ready[j].Priority] {
			return priorityOrder[ready[i].Priority] > priorityOrder[ready[j].Priority]
		}
		return ready[i].CreatedAt.Before(ready[j].CreatedAt)
	})

	return map[string]interface{}{
		"tasks": ready,
		"count": len(ready),
	}, nil
}

// SetWorkflowState sets or updates key-value pairs in task's workflow state
func (tm *TaskManager) SetWorkflowState(repoPath, taskID string, state map[string]interface{}) (map[string]interface{}, error) {
	repoRoot, err := tm.getRepoPath(repoPath)
	if err != nil {
		return nil, err
	}

	repo, err := tm.getGitHubRepo(repoRoot)
	if err != nil {
		return nil, err
	}

	// Read the task file
	filename := fmt.Sprintf("%s.json", taskID)
	task, sha, err := tm.readTaskFile(repo, filename)
	if err != nil {
		return nil, err
	}

	if task == nil {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	// Initialize WorkflowState if nil
	if task.WorkflowState == nil {
		task.WorkflowState = make(map[string]string)
	}

	// Merge new state (convert interface{} to string)
	for key, value := range state {
		task.WorkflowState[key] = fmt.Sprintf("%v", value)
	}

	task.UpdatedAt = time.Now()

	// Write back the updated task
	if err := tm.writeTaskFile(repo, task, sha); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"taskId":        taskID,
		"workflowState": task.WorkflowState,
		"message":       fmt.Sprintf("Updated workflow state for task %s", taskID),
	}, nil
}

// GetWorkflowState retrieves workflow state from a task
func (tm *TaskManager) GetWorkflowState(repoPath, taskID string) (map[string]interface{}, error) {
	repoRoot, err := tm.getRepoPath(repoPath)
	if err != nil {
		return nil, err
	}

	repo, err := tm.getGitHubRepo(repoRoot)
	if err != nil {
		return nil, err
	}

	// Read the task file
	filename := fmt.Sprintf("%s.json", taskID)
	task, _, err := tm.readTaskFile(repo, filename)
	if err != nil {
		return nil, err
	}

	if task == nil {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	return map[string]interface{}{
		"taskId":        taskID,
		"workflowState": task.WorkflowState,
	}, nil
}

// DeleteWorkflowStateKey deletes a key from task's workflow state
func (tm *TaskManager) DeleteWorkflowStateKey(repoPath, taskID, key string) (map[string]interface{}, error) {
	repoRoot, err := tm.getRepoPath(repoPath)
	if err != nil {
		return nil, err
	}

	repo, err := tm.getGitHubRepo(repoRoot)
	if err != nil {
		return nil, err
	}

	// Read the task file
	filename := fmt.Sprintf("%s.json", taskID)
	task, sha, err := tm.readTaskFile(repo, filename)
	if err != nil {
		return nil, err
	}

	if task == nil {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	// Delete key if it exists
	if task.WorkflowState != nil {
		delete(task.WorkflowState, key)
	}

	task.UpdatedAt = time.Now()

	// Write back the updated task
	if err := tm.writeTaskFile(repo, task, sha); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"taskId":        taskID,
		"workflowState": task.WorkflowState,
		"message":       fmt.Sprintf("Deleted key '%s' from task %s workflow state", key, taskID),
	}, nil
}

// helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
