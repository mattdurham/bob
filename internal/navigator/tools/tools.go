package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mattdurham/bob/internal/navigator/agent"
	"github.com/mattdurham/bob/internal/navigator/embedder"
	"github.com/mattdurham/bob/internal/navigator/store"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server holds shared state for all navigator MCP tools.
type Server struct {
	store    *store.Store
	agent    *agent.Agent
	embedder *embedder.Embedder
}

// New creates a tool server. agent and embedder may be nil.
func New(s *store.Store, a *agent.Agent, e *embedder.Embedder) *Server {
	return &Server{store: s, agent: a, embedder: e}
}

// Register registers all navigator tools with the MCP server.
func (s *Server) Register(srv *mcp.Server) {
	type RememberArgs struct {
		Content    string            `json:"content"    jsonschema:"The finding, insight, or decision to store"`
		Repo       string            `json:"repo"       jsonschema:"Optional: repository name, e.g. grafana/tempo. Omit for general knowledge."`
		Scope      string            `json:"scope"      jsonschema:"Package or file path this applies to, e.g. pkg/store"`
		Tags       []string          `json:"tags"       jsonschema:"Optional labels, e.g. [concurrency, nil-map, fix]"`
		Confidence string            `json:"confidence" jsonschema:"One of: verified, observed, tentative"`
		Source     string            `json:"source"     jsonschema:"Origin of this finding, e.g. pr-review, debugging, code-review"`
		Meta       map[string]string `json:"meta"       jsonschema:"Optional key-value metadata, e.g. {pr: '#234', commit: 'abc123'}"`
	}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "remember",
		Description: "Store a finding, fix, or insight into the knowledge base for future agents to use",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args RememberArgs) (*mcp.CallToolResult, any, error) {
		return s.remember(ctx, args.Content, args.Repo, args.Scope, args.Tags, args.Confidence, args.Source, args.Meta)
	})

	type RecallArgs struct {
		Query string `json:"query" jsonschema:"What to search for"`
		Scope string `json:"scope" jsonschema:"Optional: narrow to a package or file path"`
		Limit int    `json:"limit" jsonschema:"Max results (default 10)"`
	}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "recall",
		Description: "Search the knowledge base and return raw matching findings",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args RecallArgs) (*mcp.CallToolResult, any, error) {
		return s.recall(ctx, args.Query, args.Scope, args.Limit)
	})

	type ConsultArgs struct {
		Question string `json:"question" jsonschema:"The question to ask the senior developer"`
		Scope    string `json:"scope"    jsonschema:"Optional: narrow context to a package or file path"`
	}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "consult",
		Description: "Ask the navigator a question. It will search accumulated knowledge and return an opinionated answer — like asking a senior developer who has reviewed this codebase before",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ConsultArgs) (*mcp.CallToolResult, any, error) {
		return s.consult(ctx, args.Question, args.Scope)
	})

	type ReindexArgs struct{}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "reindex",
		Description: "Backfill vector embeddings for all thoughts that don't have them yet",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ReindexArgs) (*mcp.CallToolResult, any, error) {
		return s.reindex(ctx)
	})

	type ListArgs struct {
		Repo string `json:"repo" jsonschema:"Optional: filter by repository name"`
	}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list",
		Description: "List all thoughts, optionally filtered by repo",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ListArgs) (*mcp.CallToolResult, any, error) {
		return s.list(ctx, args.Repo)
	})

	type StatusArgs struct{}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "status",
		Description: "Check navigator status: total thoughts, embedded count, queue depth",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args StatusArgs) (*mcp.CallToolResult, any, error) {
		return s.status(ctx)
	})

	type UpdateArgs struct {
		ID         int64             `json:"id"         jsonschema:"Thought ID to update"`
		Content    string            `json:"content"    jsonschema:"New content (replaces existing)"`
		Repo       string            `json:"repo"       jsonschema:"Optional: repository name"`
		Scope      string            `json:"scope"      jsonschema:"Package or file path"`
		Tags       []string          `json:"tags"       jsonschema:"Labels"`
		Confidence string            `json:"confidence" jsonschema:"One of: verified, observed, tentative"`
		Source     string            `json:"source"     jsonschema:"Origin of this finding"`
		Meta       map[string]string `json:"meta"       jsonschema:"Optional key-value metadata (replaces existing)"`
	}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "update",
		Description: "Update an existing thought's content and metadata. Clears embeddings (run reindex to regenerate).",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args UpdateArgs) (*mcp.CallToolResult, any, error) {
		return s.update(ctx, args.ID, args.Content, args.Repo, args.Scope, args.Tags, args.Confidence, args.Source, args.Meta)
	})

	type ForgetArgs struct {
		ID int64 `json:"id" jsonschema:"Thought ID to delete"`
	}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "forget",
		Description: "Delete a thought and its embeddings permanently",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ForgetArgs) (*mcp.CallToolResult, any, error) {
		return s.forget(ctx, args.ID)
	})
}

func (s *Server) remember(_ context.Context, content, repo, scope string, tags []string, confidence, source string, meta map[string]string) (*mcp.CallToolResult, any, error) {
	if strings.TrimSpace(content) == "" {
		return errResult("content is required"), nil, nil
	}
	if confidence == "" {
		confidence = "observed"
	}

	t := &store.Thought{
		Content:    content,
		Repo:       repo,
		Scope:      scope,
		Tags:       tags,
		Confidence: confidence,
		Source:     source,
	}
	if err := s.store.Save(t); err != nil {
		return errResult(fmt.Sprintf("save failed: %v", err)), nil, nil
	}

	for k, v := range meta {
		if err := s.store.SetMeta(t.ID, k, v); err != nil {
			return errResult(fmt.Sprintf("meta failed: %v", err)), nil, nil
		}
	}

	if s.embedder != nil {
		s.embedder.Enqueue(t.ID, content)
	}
	return textResult(fmt.Sprintf("stored as thought-%d", t.ID)), nil, nil
}

func (s *Server) recall(_ context.Context, query, scope string, limit int) (*mcp.CallToolResult, any, error) {
	if strings.TrimSpace(query) == "" {
		return errResult("query is required"), nil, nil
	}

	var (
		thoughts []*store.Thought
		err      error
	)

	// Vector search is synchronous — we need the embedding to query.
	if s.embedder != nil {
		vec, embErr := s.embedder.Embed(query)
		if embErr == nil {
			thoughts, err = s.store.VectorSearch(vec, scope, limit)
		}
	}
	if thoughts == nil {
		thoughts, err = s.store.Search(query, scope, limit)
	}
	if err != nil {
		return errResult(fmt.Sprintf("search failed: %v", err)), nil, nil
	}
	if len(thoughts) == 0 {
		return textResult("no findings matched that query"), nil, nil
	}

	var out []map[string]any
	for _, t := range thoughts {
		entry := map[string]any{
			"id":         t.ID,
			"content":    t.Content,
			"scope":      t.Scope,
			"tags":       t.Tags,
			"confidence": t.Confidence,
			"source":     t.Source,
			"created_at": t.CreatedAt.Format("2006-01-02"),
		}
		if t.Repo != "" {
			entry["repo"] = t.Repo
		}
		if meta, err := s.store.GetMeta(t.ID); err == nil && len(meta) > 0 {
			entry["meta"] = meta
		}
		out = append(out, entry)
	}
	b, _ := json.MarshalIndent(out, "", "  ")
	return textResult(string(b)), nil, nil
}

func (s *Server) consult(ctx context.Context, question, scope string) (*mcp.CallToolResult, any, error) {
	if s.agent == nil {
		return errResult("consult requires NAVIGATOR_API_KEY — set it and restart navigator"), nil, nil
	}
	if strings.TrimSpace(question) == "" {
		return errResult("question is required"), nil, nil
	}
	answer, err := s.agent.Consult(ctx, question, scope)
	if err != nil {
		return errResult(fmt.Sprintf("agent error: %v", err)), nil, nil
	}
	return textResult(answer), nil, nil
}

func (s *Server) reindex(_ context.Context) (*mcp.CallToolResult, any, error) {
	if s.embedder == nil {
		return errResult("reindex requires the embedding model — install with make install-navigator"), nil, nil
	}
	thoughts, err := s.store.ThoughtsWithoutEmbeddings()
	if err != nil {
		return errResult(fmt.Sprintf("list failed: %v", err)), nil, nil
	}
	if len(thoughts) == 0 {
		return textResult("all thoughts already have embeddings"), nil, nil
	}

	var embedded, totalChunks int
	for _, t := range thoughts {
		vecs, err := s.embedder.EmbedChunks(t.Content)
		if err != nil {
			return textResult(fmt.Sprintf("embedded %d/%d (failed on thought-%d: %v)", embedded, len(thoughts), t.ID, err)), nil, nil
		}
		for _, vec := range vecs {
			if err := s.store.AddEmbedding(t.ID, vec); err != nil {
				return textResult(fmt.Sprintf("embedded %d/%d (store failed on thought-%d: %v)", embedded, len(thoughts), t.ID, err)), nil, nil
			}
			totalChunks++
		}
		embedded++
	}
	return textResult(fmt.Sprintf("embedded %d/%d thoughts (%d chunks)", embedded, len(thoughts), totalChunks)), nil, nil
}

func (s *Server) status(_ context.Context) (*mcp.CallToolResult, any, error) {
	total, embedded, err := s.store.Stats()
	if err != nil {
		return errResult(fmt.Sprintf("stats failed: %v", err)), nil, nil
	}
	pending := 0
	if s.embedder != nil {
		pending = s.embedder.Pending()
	}
	return textResult(fmt.Sprintf("thoughts: %d total, %d embedded, %d in queue", total, embedded, pending)), nil, nil
}

func (s *Server) list(_ context.Context, repo string) (*mcp.CallToolResult, any, error) {
	thoughts, err := s.store.All()
	if err != nil {
		return errResult(fmt.Sprintf("list failed: %v", err)), nil, nil
	}

	var out []map[string]any
	for _, t := range thoughts {
		if repo != "" && t.Repo != repo {
			continue
		}
		entry := map[string]any{
			"id":         t.ID,
			"scope":      t.Scope,
			"tags":       t.Tags,
			"confidence": t.Confidence,
			"source":     t.Source,
			"created_at": t.CreatedAt.Format("2006-01-02"),
			"excerpt":    truncate(t.Content, 120),
		}
		if t.Repo != "" {
			entry["repo"] = t.Repo
		}
		out = append(out, entry)
	}
	if len(out) == 0 {
		return textResult("no thoughts stored"), nil, nil
	}
	b, _ := json.MarshalIndent(out, "", "  ")
	return textResult(string(b)), nil, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func (s *Server) update(_ context.Context, id int64, content, repo, scope string, tags []string, confidence, source string, meta map[string]string) (*mcp.CallToolResult, any, error) {
	if id == 0 {
		return errResult("id is required"), nil, nil
	}
	if strings.TrimSpace(content) == "" {
		return errResult("content is required"), nil, nil
	}
	if confidence == "" {
		confidence = "observed"
	}
	t := &store.Thought{
		Content:    content,
		Repo:       repo,
		Scope:      scope,
		Tags:       tags,
		Confidence: confidence,
		Source:     source,
	}
	if err := s.store.Update(id, t); err != nil {
		return errResult(fmt.Sprintf("update failed: %v", err)), nil, nil
	}
	// Replace meta.
	// Delete old meta first — Update doesn't clear it.
	s.store.DeleteMeta(id)
	for k, v := range meta {
		s.store.SetMeta(id, k, v)
	}
	// Re-embed in background.
	if s.embedder != nil {
		s.embedder.Enqueue(id, content)
	}
	return textResult(fmt.Sprintf("updated thought-%d", id)), nil, nil
}

func (s *Server) forget(_ context.Context, id int64) (*mcp.CallToolResult, any, error) {
	if id == 0 {
		return errResult("id is required"), nil, nil
	}
	if err := s.store.Delete(id); err != nil {
		return errResult(fmt.Sprintf("delete failed: %v", err)), nil, nil
	}
	return textResult(fmt.Sprintf("deleted thought-%d", id)), nil, nil
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}

func errResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: "ERROR: " + msg}},
	}
}
