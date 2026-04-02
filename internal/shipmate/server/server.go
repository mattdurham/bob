// Package server implements the shipmate MCP tool server.
// It exposes the shipmate_record tool that lets agents emit synthetic OTEL spans.
package server

import (
	"context"
	"strings"

	"github.com/mattdurham/bob/internal/shipmate/recorder"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Recorderface is the interface the MCP server uses to create synthetic spans.
// recorder.Recorder implements this interface.
type Recorderface interface {
	Record(ctx context.Context, args recorder.RecordArgs) error
}

// SessionIDProvider is the interface the MCP server uses to read the current session.id.
// proxy.Proxy implements this.
type SessionIDProvider interface {
	SessionID() string
}

// Server holds MCP tool state.
type Server struct {
	rec      Recorderface
	sessions SessionIDProvider
}

// New creates a tool server.
func New(rec Recorderface, sessions SessionIDProvider) *Server {
	return &Server{rec: rec, sessions: sessions}
}

// Register registers all shipmate tools with the MCP server.
func (s *Server) Register(srv *mcp.Server) {
	type RecordArgs struct {
		Name       string            `json:"name"       jsonschema:"Span name"`
		Agent      string            `json:"agent"      jsonschema:"Agent identity, e.g. coder-1"`
		Text       string            `json:"text"       jsonschema:"Free-form description"`
		Attributes map[string]string `json:"attributes" jsonschema:"Optional key-value span attributes"`
	}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "shipmate_record",
		Description: "Emit a synthetic OTEL span annotated with your agent identity and description. The span is correlated to the active Claude Code session via session.id.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args RecordArgs) (*mcp.CallToolResult, any, error) {
		return s.record(ctx, args.Name, args.Agent, args.Text, args.Attributes)
	})
}

func (s *Server) record(ctx context.Context, name, agent, text string, attributes map[string]string) (*mcp.CallToolResult, any, error) {
	if strings.TrimSpace(name) == "" {
		return errResult("name is required"), nil, nil
	}
	if strings.TrimSpace(agent) == "" {
		return errResult("agent is required"), nil, nil
	}

	args := recorder.RecordArgs{
		Name:       name,
		Agent:      agent,
		Text:       text,
		SessionID:  s.sessions.SessionID(),
		Attributes: attributes,
	}
	if err := s.rec.Record(ctx, args); err != nil {
		return errResult("record failed: " + err.Error()), nil, nil
	}
	return textResult("span recorded: " + name), nil, nil
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}

func errResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: "ERROR: " + msg}},
	}
}
