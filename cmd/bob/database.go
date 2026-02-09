package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Database manages the shared SQLite database
type Database struct {
	db *sql.DB
}

// NewDatabase creates or opens the shared database
func NewDatabase() (*Database, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("getting home dir: %w", err)
	}

	dbDir := filepath.Join(homeDir, ".bob", "state")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("creating db directory: %w", err)
	}

	dbPath := filepath.Join(dbDir, "db.sql")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	database := &Database{db: db}
	if err := database.initSchema(); err != nil {
		db.Close()
		return nil, err
	}

	return database, nil
}

// initSchema creates tables if they don't exist
func (d *Database) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS workflows (
		id TEXT PRIMARY KEY,
		workflow TEXT NOT NULL,
		current_step TEXT NOT NULL,
		task_description TEXT,
		loop_count INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS workflow_progress (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		workflow_id TEXT NOT NULL,
		step TEXT NOT NULL,
		metadata TEXT,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (workflow_id) REFERENCES workflows(id)
	);

	CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		repo_path TEXT NOT NULL,
		title TEXT NOT NULL,
		description TEXT,
		type TEXT,
		priority TEXT,
		state TEXT,
		assignee TEXT,
		tags TEXT,
		blocks TEXT,
		blocked_by TEXT,
		metadata TEXT,
		workflow_state TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		completed_at TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS task_comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id TEXT NOT NULL,
		author TEXT,
		text TEXT NOT NULL,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (task_id) REFERENCES tasks(id)
	);

	CREATE INDEX IF NOT EXISTS idx_workflows_workflow ON workflows(workflow);
	CREATE INDEX IF NOT EXISTS idx_workflows_current_step ON workflows(current_step);
	CREATE INDEX IF NOT EXISTS idx_workflow_progress_workflow_id ON workflow_progress(workflow_id);
	CREATE INDEX IF NOT EXISTS idx_tasks_repo_path ON tasks(repo_path);
	CREATE INDEX IF NOT EXISTS idx_tasks_state ON tasks(state);
	CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority);
	`

	_, err := d.db.Exec(schema)
	return err
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// Workflow operations

func (d *Database) SaveWorkflow(id, workflow, currentStep, taskDescription string, loopCount int) error {
	_, err := d.db.Exec(`
		INSERT INTO workflows (id, workflow, current_step, task_description, loop_count, updated_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(id) DO UPDATE SET
			current_step = excluded.current_step,
			task_description = excluded.task_description,
			loop_count = excluded.loop_count,
			updated_at = CURRENT_TIMESTAMP
	`, id, workflow, currentStep, taskDescription, loopCount)
	return err
}

func (d *Database) AddWorkflowProgress(workflowID, step string, metadata map[string]interface{}) error {
	metadataJSON, _ := json.Marshal(metadata)
	_, err := d.db.Exec(`
		INSERT INTO workflow_progress (workflow_id, step, metadata)
		VALUES (?, ?, ?)
	`, workflowID, step, string(metadataJSON))
	return err
}

func (d *Database) GetActiveWorkflows() ([]map[string]interface{}, error) {
	rows, err := d.db.Query(`
		SELECT id, workflow, current_step, task_description, loop_count,
		       (SELECT COUNT(*) FROM workflow_progress WHERE workflow_id = workflows.id) as progress_count
		FROM workflows
		WHERE updated_at > datetime('now', '-1 day')
		ORDER BY updated_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workflows []map[string]interface{}
	for rows.Next() {
		var id, workflow, currentStep, taskDescription string
		var loopCount, progressCount int
		if err := rows.Scan(&id, &workflow, &currentStep, &taskDescription, &loopCount, &progressCount); err != nil {
			continue
		}
		workflows = append(workflows, map[string]interface{}{
			"id":              id,
			"workflow":        workflow,
			"currentStep":     currentStep,
			"taskDescription": taskDescription,
			"loopCount":       loopCount,
			"progressCount":   progressCount,
		})
	}
	return workflows, nil
}

func (d *Database) GetWorkflowsByType(workflow string) ([]map[string]interface{}, error) {
	rows, err := d.db.Query(`
		SELECT id, workflow, current_step, task_description, loop_count,
		       (SELECT COUNT(*) FROM workflow_progress WHERE workflow_id = workflows.id) as progress_count
		FROM workflows
		WHERE workflow = ?
		ORDER BY updated_at DESC
	`, workflow)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workflows []map[string]interface{}
	for rows.Next() {
		var id, wf, currentStep, taskDescription string
		var loopCount, progressCount int
		if err := rows.Scan(&id, &wf, &currentStep, &taskDescription, &loopCount, &progressCount); err != nil {
			continue
		}
		workflows = append(workflows, map[string]interface{}{
			"id":              id,
			"workflow":        wf,
			"currentStep":     currentStep,
			"taskDescription": taskDescription,
			"loopCount":       loopCount,
			"progressCount":   progressCount,
		})
	}
	return workflows, nil
}

// Task operations

func (d *Database) SaveTask(task Task) error {
	tagsJSON, _ := json.Marshal(task.Tags)
	blocksJSON, _ := json.Marshal(task.Blocks)
	blockedByJSON, _ := json.Marshal(task.BlockedBy)
	metadataJSON, _ := json.Marshal(task.Metadata)
	workflowStateJSON, _ := json.Marshal(task.WorkflowState)

	var completedAt *time.Time
	if task.CompletedAt != nil {
		completedAt = task.CompletedAt
	}

	_, err := d.db.Exec(`
		INSERT INTO tasks (id, repo_path, title, description, type, priority, state, assignee,
		                   tags, blocks, blocked_by, metadata, workflow_state, created_at, updated_at, completed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			title = excluded.title,
			description = excluded.description,
			type = excluded.type,
			priority = excluded.priority,
			state = excluded.state,
			assignee = excluded.assignee,
			tags = excluded.tags,
			blocks = excluded.blocks,
			blocked_by = excluded.blocked_by,
			metadata = excluded.metadata,
			workflow_state = excluded.workflow_state,
			updated_at = excluded.updated_at,
			completed_at = excluded.completed_at
	`, task.ID, "", task.Title, task.Description, task.Type, task.Priority, task.State, task.Assignee,
		string(tagsJSON), string(blocksJSON), string(blockedByJSON), string(metadataJSON),
		string(workflowStateJSON), task.CreatedAt, task.UpdatedAt, completedAt)
	return err
}

func (d *Database) GetAllTasks() ([]Task, error) {
	rows, err := d.db.Query(`
		SELECT id, title, description, type, priority, state, assignee,
		       tags, blocks, blocked_by, metadata, workflow_state, created_at, updated_at, completed_at
		FROM tasks
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		var tagsJSON, blocksJSON, blockedByJSON, metadataJSON, workflowStateJSON string
		var completedAt sql.NullTime

		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Type, &task.Priority,
			&task.State, &task.Assignee, &tagsJSON, &blocksJSON, &blockedByJSON,
			&metadataJSON, &workflowStateJSON, &task.CreatedAt, &task.UpdatedAt, &completedAt); err != nil {
			continue
		}

		json.Unmarshal([]byte(tagsJSON), &task.Tags)
		json.Unmarshal([]byte(blocksJSON), &task.Blocks)
		json.Unmarshal([]byte(blockedByJSON), &task.BlockedBy)
		json.Unmarshal([]byte(metadataJSON), &task.Metadata)
		json.Unmarshal([]byte(workflowStateJSON), &task.WorkflowState)

		if completedAt.Valid {
			task.CompletedAt = &completedAt.Time
		}

		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (d *Database) GetTaskStats() map[string]int {
	stats := map[string]int{
		"total":       0,
		"pending":     0,
		"in_progress": 0,
		"blocked":     0,
		"completed":   0,
	}

	rows, err := d.db.Query(`SELECT state, COUNT(*) FROM tasks GROUP BY state`)
	if err != nil {
		return stats
	}
	defer rows.Close()

	for rows.Next() {
		var state string
		var count int
		if err := rows.Scan(&state, &count); err != nil {
			continue
		}
		stats[state] = count
		stats["total"] += count
	}

	return stats
}

func (d *Database) GetTask(taskID string) (Task, error) {
	var task Task
	var tagsJSON, blocksJSON, blockedByJSON, metadataJSON, workflowStateJSON string
	var completedAt sql.NullTime

	err := d.db.QueryRow(`
		SELECT id, title, description, type, priority, state, assignee,
		       tags, blocks, blocked_by, metadata, workflow_state, created_at, updated_at, completed_at
		FROM tasks
		WHERE id = ?
	`, taskID).Scan(&task.ID, &task.Title, &task.Description, &task.Type, &task.Priority,
		&task.State, &task.Assignee, &tagsJSON, &blocksJSON, &blockedByJSON,
		&metadataJSON, &workflowStateJSON, &task.CreatedAt, &task.UpdatedAt, &completedAt)

	if err != nil {
		return task, err
	}

	json.Unmarshal([]byte(tagsJSON), &task.Tags)
	json.Unmarshal([]byte(blocksJSON), &task.Blocks)
	json.Unmarshal([]byte(blockedByJSON), &task.BlockedBy)
	json.Unmarshal([]byte(metadataJSON), &task.Metadata)
	json.Unmarshal([]byte(workflowStateJSON), &task.WorkflowState)

	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}

	return task, nil
}

func (d *Database) UpdateTask(taskID string, updates map[string]interface{}) error {
	// Build dynamic UPDATE query
	query := "UPDATE tasks SET updated_at = CURRENT_TIMESTAMP"
	args := []interface{}{}

	if state, ok := updates["state"].(string); ok {
		query += ", state = ?"
		args = append(args, state)
	}

	if assignee, ok := updates["assignedTo"].(string); ok {
		query += ", assignee = ?"
		args = append(args, assignee)
	}

	if priority, ok := updates["priority"].(string); ok {
		query += ", priority = ?"
		args = append(args, priority)
	}

	if title, ok := updates["title"].(string); ok {
		query += ", title = ?"
		args = append(args, title)
	}

	if description, ok := updates["description"].(string); ok {
		query += ", description = ?"
		args = append(args, description)
	}

	query += " WHERE id = ?"
	args = append(args, taskID)

	_, err := d.db.Exec(query, args...)
	return err
}
