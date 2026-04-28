package daemon

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/mattdurham/bob/internal/navigator/specstore"
)

// Client connects to a running daemon over its Unix socket.
type Client struct {
	socketPath string
	httpClient *http.Client
}

// NewClient creates a client that communicates over the Unix socket at socketPath.
func NewClient(socketPath string) *Client {
	dialer := &net.Dialer{}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return dialer.DialContext(ctx, "unix", socketPath)
		},
	}
	return &Client{
		socketPath: socketPath,
		httpClient: &http.Client{Transport: transport},
	}
}

// Query sends a query to the daemon.
func (c *Client) Query(ctx context.Context, query string, limit int) ([]specstore.Result, error) {
	body, err := json.Marshal(map[string]any{"query": query, "limit": limit})
	if err != nil {
		return nil, fmt.Errorf("daemon client: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://unix/query", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("daemon client: request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("daemon client: query: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("daemon client: query returned status %d", resp.StatusCode)
	}

	var result struct {
		Results []specstore.Result `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("daemon client: decode: %w", err)
	}
	return result.Results, nil
}

// Status returns daemon status.
func (c *Client) Status(ctx context.Context) (map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://unix/status", nil)
	if err != nil {
		return nil, fmt.Errorf("daemon client: status request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("daemon client: status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("daemon client: status returned %d", resp.StatusCode)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("daemon client: decode status: %w", err)
	}
	return result, nil
}
