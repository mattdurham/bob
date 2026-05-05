package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"time"
)

type runFunc func(ctx context.Context, name string, args ...string) ([]byte, error)

// Manager manages the text-embeddings-inference Docker container.
type Manager struct {
	containerName string
	image         string
	port          int
	extraArgs     []string
	execRun       runFunc
	healthCheck   func(ctx context.Context, baseURL string) error
	// lookPath is used to locate the docker binary. It defaults to
	// exec.LookPath and can be replaced in tests to avoid requiring Docker.
	lookPath func(string) (string, error)
}

// New creates a Manager with the real exec runner.
func New(containerName, image string, port int, extraArgs []string) *Manager {
	m := &Manager{
		containerName: containerName,
		image:         image,
		port:          port,
		extraArgs:     extraArgs,
		lookPath:      exec.LookPath,
	}
	m.execRun = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return exec.CommandContext(ctx, name, args...).CombinedOutput()
	}
	m.healthCheck = defaultHealthCheck
	return m
}

func defaultHealthCheck(ctx context.Context, baseURL string) error {
	client := &http.Client{Timeout: 5 * time.Second}
	deadline := time.Now().Add(60 * time.Second)
	wait := time.Second
	for time.Now().Before(deadline) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/health", nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
		if wait < 16*time.Second {
			wait *= 2
		}
	}
	return fmt.Errorf("TEI container health check timed out after 60s; try: docker pull %s", "nomic-ai/nomic-embed-text-v1.5")
}

type containerState struct {
	Running bool `json:"Running"`
}

func (m *Manager) isRunning(ctx context.Context) bool {
	out, err := m.execRun(ctx, "docker", "inspect", "--format", "{{json .State}}", m.containerName)
	if err != nil {
		return false
	}
	var state containerState
	if err := json.Unmarshal(out, &state); err != nil {
		return false
	}
	return state.Running
}

// EnsureRunning checks if the container is running; starts it if not.
// Returns the base URL on success.
func (m *Manager) EnsureRunning(ctx context.Context) (string, error) {
	if _, err := m.lookPath("docker"); err != nil {
		return "", fmt.Errorf("docker is required: install Docker Desktop or Docker Engine")
	}

	baseURL := fmt.Sprintf("http://localhost:%d", m.port)

	if m.isRunning(ctx) {
		return baseURL, nil
	}

	// Try with GPU first
	gpuArgs := m.buildRunArgs(true)
	if _, err := m.execRun(ctx, "docker", gpuArgs...); err != nil {
		// Fallback to CPU
		cpuArgs := m.buildRunArgs(false)
		if _, err2 := m.execRun(ctx, "docker", cpuArgs...); err2 != nil {
			return "", fmt.Errorf("docker run failed (GPU: %v, CPU: %v)", err, err2)
		}
	}

	if err := m.healthCheck(ctx, baseURL); err != nil {
		return "", fmt.Errorf("container started but health check failed: %w", err)
	}

	return baseURL, nil
}

func (m *Manager) buildRunArgs(useGPU bool) []string {
	args := []string{
		"run", "-d",
		"--name", m.containerName,
		"-p", fmt.Sprintf("%d:80", m.port),
	}
	if useGPU {
		args = append(args, "--gpus", "all")
	}
	args = append(args, m.extraArgs...)
	args = append(args, m.image, "--model-id", "nomic-ai/nomic-embed-text-v1.5")
	return args
}

// Stop stops and removes the container.
func (m *Manager) Stop(ctx context.Context) error {
	if _, err := m.execRun(ctx, "docker", "stop", m.containerName); err != nil {
		return fmt.Errorf("docker stop: %w", err)
	}
	if _, err := m.execRun(ctx, "docker", "rm", m.containerName); err != nil {
		return fmt.Errorf("docker rm: %w", err)
	}
	return nil
}
