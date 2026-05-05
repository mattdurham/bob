package docker

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func makeManager(rf runFunc) *Manager {
	m := New("test-container", "test-image:latest", 7462, nil)
	m.execRun = rf
	// Inject a fake lookPath so tests don't require Docker to be installed.
	m.lookPath = func(name string) (string, error) {
		return "/usr/bin/" + name, nil
	}
	return m
}

func TestIsRunning_True(t *testing.T) {
	m := makeManager(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return []byte(`{"Running":true}`), nil
	})
	if !m.isRunning(context.Background()) {
		t.Error("expected isRunning to return true")
	}
}

func TestIsRunning_False(t *testing.T) {
	m := makeManager(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return []byte(`{"Running":false}`), nil
	})
	if m.isRunning(context.Background()) {
		t.Error("expected isRunning to return false")
	}
}

func TestIsRunning_NotFound(t *testing.T) {
	m := makeManager(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("exit status 1")
	})
	if m.isRunning(context.Background()) {
		t.Error("expected isRunning to return false on error")
	}
}

func TestEnsureRunning_AlreadyRunning(t *testing.T) {
	calls := 0
	m := makeManager(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		calls++
		if name == "docker" && len(args) > 0 && args[0] == "inspect" {
			return []byte(`{"Running":true}`), nil
		}
		return nil, fmt.Errorf("unexpected call to %s %v", name, args)
	})
	url, err := m.EnsureRunning(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "http://localhost:7462" {
		t.Errorf("expected http://localhost:7462, got %s", url)
	}
	// Only the inspect call should have been made
	if calls != 1 {
		t.Errorf("expected 1 call (inspect), got %d", calls)
	}
}

func TestEnsureRunning_StartsWithGPU(t *testing.T) {
	var runArgs []string
	m := makeManager(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if name == "docker" && len(args) > 0 && args[0] == "inspect" {
			return []byte(`{"Running":false}`), nil
		}
		if name == "docker" && len(args) > 0 && args[0] == "run" {
			runArgs = args
			return []byte("container-id"), nil
		}
		return nil, fmt.Errorf("unexpected: %s %v", name, args)
	})
	// Override health poll to succeed immediately
	m.healthCheck = func(ctx context.Context, baseURL string) error { return nil }
	_, err := m.EnsureRunning(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	hasGPU := false
	for _, a := range runArgs {
		if a == "--gpus" {
			hasGPU = true
			break
		}
	}
	if !hasGPU {
		t.Errorf("expected GPU run args, got %v", runArgs)
	}
}

func TestEnsureRunning_FallbackCPU(t *testing.T) {
	gpuTried := false
	cpuTried := false
	m := makeManager(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if name == "docker" && len(args) > 0 && args[0] == "inspect" {
			return []byte(`{"Running":false}`), nil
		}
		if name == "docker" && len(args) > 0 && args[0] == "run" {
			for _, a := range args {
				if a == "--gpus" {
					gpuTried = true
					return nil, fmt.Errorf("GPU not available")
				}
			}
			cpuTried = true
			return []byte("container-id"), nil
		}
		return nil, fmt.Errorf("unexpected: %s %v", name, args)
	})
	m.healthCheck = func(ctx context.Context, baseURL string) error { return nil }
	_, err := m.EnsureRunning(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !gpuTried {
		t.Error("GPU run was not attempted")
	}
	if !cpuTried {
		t.Error("CPU fallback was not attempted")
	}
}

func TestStop(t *testing.T) {
	var called []string
	m := makeManager(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		called = append(called, strings.Join(append([]string{name}, args...), " "))
		return nil, nil
	})
	if err := m.Stop(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(called) < 2 {
		t.Fatalf("expected at least 2 calls (stop + rm), got %d: %v", len(called), called)
	}
	hasStop := false
	hasRm := false
	for _, c := range called {
		if strings.Contains(c, "stop") && strings.Contains(c, "test-container") {
			hasStop = true
		}
		if strings.Contains(c, "rm") && strings.Contains(c, "test-container") {
			hasRm = true
		}
	}
	if !hasStop {
		t.Errorf("docker stop not called, got: %v", called)
	}
	if !hasRm {
		t.Errorf("docker rm not called, got: %v", called)
	}
}

func TestDefaultHealthCheck_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	if err := defaultHealthCheck(context.Background(), srv.URL); err != nil {
		t.Errorf("expected nil error for healthy server, got %v", err)
	}
}

func TestDefaultHealthCheck_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately cancelled

	err := defaultHealthCheck(ctx, srv.URL)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}
