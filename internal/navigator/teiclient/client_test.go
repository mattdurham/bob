package teiclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew_DefaultTimeout(t *testing.T) {
	c := New("http://localhost:7462")
	if c.httpClient.Timeout == 0 {
		t.Error("expected non-zero http client timeout")
	}
}

func TestEmbed_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([][]float32{{0.1, 0.2, 0.3}})
	}))
	defer srv.Close()

	c := New(srv.URL)
	result, err := c.Embed(context.Background(), []string{"hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if len(result[0]) != 3 {
		t.Fatalf("expected 3 floats, got %d", len(result[0]))
	}
}

func TestEmbed_BatchMultiple(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([][]float32{
			{0.1, 0.2},
			{0.3, 0.4},
			{0.5, 0.6},
		})
	}))
	defer srv.Close()

	c := New(srv.URL)
	result, err := c.Embed(context.Background(), []string{"a", "b", "c"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 results, got %d", len(result))
	}
}

func TestEmbed_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := New(srv.URL)
	_, err := c.Embed(context.Background(), []string{"hello"})
	if err == nil {
		t.Error("expected error for 500 response")
	}
}

func TestEmbed_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not-json"))
	}))
	defer srv.Close()

	c := New(srv.URL)
	_, err := c.Embed(context.Background(), []string{"hello"})
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestHealth_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(srv.URL)
	if err := c.Health(context.Background()); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestHealth_Unavailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	c := New(srv.URL)
	if err := c.Health(context.Background()); err == nil {
		t.Error("expected non-nil error for 503 response")
	}
}
