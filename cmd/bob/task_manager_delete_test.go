package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestDeleteTask_Success tests successful task deletion
func TestDeleteTask_Success(t *testing.T) {
	// Setup mock GitHub API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/repos/owner/repo/contents/.bob/issues":
			// List tasks
			files := []map[string]interface{}{
				{"name": "task-1.json"},
			}
			_ = json.NewEncoder(w).Encode(files)

		case r.Method == "GET" && r.URL.Path == "/repos/owner/repo/contents/.bob/issues/task-1.json":
			// Read task-1
			task := Task{
				ID:          "task-1",
				Title:       "Test Task",
				Description: "To be deleted",
				State:       "pending",
				Blocks:      []string{},
				BlockedBy:   []string{},
			}
			data, _ := json.Marshal(task)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"content": data,
				"sha":     "test-sha-1",
			})

		case r.Method == "DELETE" && r.URL.Path == "/repos/owner/repo/contents/.bob/issues/task-1.json":
			// Delete task
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"message": "deleted"})

		default:
			t.Logf("Unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// TODO: Test implementation once DeleteTask method is added
	// This test is currently a placeholder
	t.Skip("DeleteTask method not yet implemented")
}

// TestDeleteTask_WithDependencies tests deletion with dependency cleanup
func TestDeleteTask_WithDependencies(t *testing.T) {
	// Setup mock GitHub API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/repos/owner/repo/contents/.bob/issues":
			// List tasks
			files := []map[string]interface{}{
				{"name": "task-1.json"},
				{"name": "task-2.json"},
			}
			_ = json.NewEncoder(w).Encode(files)

		case r.Method == "GET" && r.URL.Path == "/repos/owner/repo/contents/.bob/issues/task-1.json":
			// Read task-1 (to be deleted, blocks task-2)
			task := Task{
				ID:        "task-1",
				Title:     "Blocking Task",
				State:     "pending",
				Blocks:    []string{"task-2"},
				BlockedBy: []string{},
				UpdatedAt: time.Now(),
			}
			data, _ := json.Marshal(task)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"content": data,
				"sha":     "test-sha-1",
			})

		case r.Method == "GET" && r.URL.Path == "/repos/owner/repo/contents/.bob/issues/task-2.json":
			// Read task-2 (blocked by task-1)
			task := Task{
				ID:        "task-2",
				Title:     "Blocked Task",
				State:     "pending",
				Blocks:    []string{},
				BlockedBy: []string{"task-1"},
				UpdatedAt: time.Now(),
			}
			data, _ := json.Marshal(task)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"content": data,
				"sha":     "test-sha-2",
			})

		case r.Method == "PUT" && r.URL.Path == "/repos/owner/repo/contents/.bob/issues/task-2.json":
			// Update task-2 (remove dependency)
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"message": "updated"})

		case r.Method == "DELETE" && r.URL.Path == "/repos/owner/repo/contents/.bob/issues/task-1.json":
			// Delete task-1
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"message": "deleted"})

		default:
			t.Logf("Unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// TODO: Test implementation once DeleteTask method is added
	t.Skip("DeleteTask method not yet implemented")
}

// TestDeleteTask_NotFound tests deletion of non-existent task
func TestDeleteTask_NotFound(t *testing.T) {
	// Setup mock GitHub API server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/repos/owner/repo/contents/.bob/issues/task-999.json" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// TODO: Test implementation once DeleteTask method is added
	t.Skip("DeleteTask method not yet implemented")
}

// TestDeleteTask_MultipleDependencies tests task with multiple dependents
func TestDeleteTask_MultipleDependencies(t *testing.T) {
	// Setup test for task that blocks multiple other tasks
	// TODO: Implement once DeleteTask method is added
	t.Skip("DeleteTask method not yet implemented")
}

// TestRemoveFromSlice tests the helper function
func TestRemoveFromSlice(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		remove   string
		expected []string
	}{
		{
			name:     "remove from middle",
			slice:    []string{"a", "b", "c"},
			remove:   "b",
			expected: []string{"a", "c"},
		},
		{
			name:     "remove from start",
			slice:    []string{"a", "b", "c"},
			remove:   "a",
			expected: []string{"b", "c"},
		},
		{
			name:     "remove from end",
			slice:    []string{"a", "b", "c"},
			remove:   "c",
			expected: []string{"a", "b"},
		},
		{
			name:     "remove non-existent",
			slice:    []string{"a", "b", "c"},
			remove:   "d",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "remove from empty",
			slice:    []string{},
			remove:   "a",
			expected: []string{},
		},
		{
			name:     "remove only element",
			slice:    []string{"a"},
			remove:   "a",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeFromSlice(tt.slice, tt.remove)

			// Check length
			if len(result) != len(tt.expected) {
				t.Errorf("expected length %d, got %d", len(tt.expected), len(result))
				return
			}

			// Check contents
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("at index %d: expected %q, got %q", i, tt.expected[i], v)
				}
			}
		})
	}
}
