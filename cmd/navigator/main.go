// navigator is a stdio MCP server that acts as a persistent knowledge base
// for Claude agents. Agents call remember() to record findings and consult()
// to ask questions backed by accumulated knowledge.
//
// Subcommands:
//
//	navigator                      — run as MCP stdio server (default)
//	navigator export <file.json.gz> — dump all thoughts to gzipped JSON
//	navigator import <file.json.gz> — load thoughts from gzipped JSON
package main

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/mattdurham/bob/internal/navigator/agent"
	"github.com/mattdurham/bob/internal/navigator/embedder"
	"github.com/mattdurham/bob/internal/navigator/store"
	"github.com/mattdurham/bob/internal/navigator/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	dbPath := flag.String("db", "", "path to thoughts database (default: ~/.bob/navigator/thoughts.db)")
	libPath := flag.String("lib", "", "path to llama.cpp lib dir (default: ~/.bob/navigator/lib)")
	modelPath := flag.String("model", "", "path to embedding GGUF model")
	flag.Parse()

	// Resolve database path.
	path := *dbPath
	if path == "" {
		var err error
		path, err = store.DBPath()
		if err != nil {
			log.Fatalf("navigator: resolve db path: %v", err)
		}
	}

	cmd := flag.Arg(0)
	switch cmd {
	case "export":
		file := flag.Arg(1)
		if file == "" {
			log.Fatal("usage: navigator export <file.json.gz>")
		}
		runExport(path, file)
	case "import":
		file := flag.Arg(1)
		if file == "" {
			log.Fatal("usage: navigator import <file.json.gz>")
		}
		runImport(path, file)
	default:
		runServer(path, *libPath, *modelPath)
	}
}

// exportThought is the JSON shape for import/export.
type exportThought struct {
	Content    string            `json:"content"`
	Repo       string            `json:"repo,omitempty"`
	Scope      string            `json:"scope"`
	Tags       []string          `json:"tags"`
	Confidence string            `json:"confidence"`
	Source     string            `json:"source"`
	Meta       map[string]string `json:"meta,omitempty"`
}

func runExport(dbPath, filePath string) {
	s, err := store.Open(dbPath)
	if err != nil {
		log.Fatalf("export: open store: %v", err)
	}
	defer s.Close()

	thoughts, err := s.All()
	if err != nil {
		log.Fatalf("export: %v", err)
	}

	var out []exportThought
	for _, t := range thoughts {
		et := exportThought{
			Content:    t.Content,
			Repo:       t.Repo,
			Scope:      t.Scope,
			Tags:       t.Tags,
			Confidence: t.Confidence,
			Source:     t.Source,
		}
		if meta, err := s.GetMeta(t.ID); err == nil && len(meta) > 0 {
			et.Meta = meta
		}
		out = append(out, et)
	}

	f, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("export: create file: %v", err)
	}
	defer f.Close()

	gz := gzip.NewWriter(f)
	enc := json.NewEncoder(gz)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		log.Fatalf("export: encode: %v", err)
	}
	gz.Close()

	fmt.Printf("exported %d thoughts to %s\n", len(out), filePath)
}

func runImport(dbPath, filePath string) {
	s, err := store.Open(dbPath)
	if err != nil {
		log.Fatalf("import: open store: %v", err)
	}
	defer s.Close()

	f, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("import: open file: %v", err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		log.Fatalf("import: gzip: %v", err)
	}
	defer gz.Close()

	data, err := io.ReadAll(gz)
	if err != nil {
		log.Fatalf("import: read: %v", err)
	}

	var thoughts []exportThought
	if err := json.Unmarshal(data, &thoughts); err != nil {
		log.Fatalf("import: decode: %v", err)
	}

	var imported int
	for _, t := range thoughts {
		thought := &store.Thought{
			Content:    t.Content,
			Repo:       t.Repo,
			Scope:      t.Scope,
			Tags:       t.Tags,
			Confidence: t.Confidence,
			Source:     t.Source,
		}
		if err := s.Save(thought); err != nil {
			log.Printf("import: skip thought: %v", err)
			continue
		}
		for k, v := range t.Meta {
			s.SetMeta(thought.ID, k, v)
		}
		imported++
	}
	fmt.Printf("imported %d/%d thoughts from %s\n", imported, len(thoughts), filePath)
	fmt.Println("run reindex via MCP to generate embeddings for imported thoughts")
}

func runServer(dbPath, libPath, modelPath string) {
	s, err := store.Open(dbPath)
	if err != nil {
		log.Fatalf("navigator: open store: %v", err)
	}
	defer s.Close()

	lp := libPath
	if lp == "" {
		lp = embedder.DefaultLibPath()
	}
	mp := modelPath
	if mp == "" {
		mp = embedder.DefaultModelPath()
	}
	emb, err := embedder.New(lp, mp, s.AddEmbedding)
	if err != nil {
		log.Fatalf("navigator: embedder: %v", err)
	}
	if emb != nil {
		defer emb.Close()
	}

	var a *agent.Agent
	if key := os.Getenv("NAVIGATOR_API_KEY"); key != "" {
		a = agent.New(key, s)
	}

	toolServer := tools.New(s, a, emb)
	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "navigator",
		Version: "v0.1.0",
	}, nil)
	toolServer.Register(srv)

	if err := srv.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("navigator: %v", err)
	}
}
