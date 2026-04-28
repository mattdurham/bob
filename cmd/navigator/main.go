// navigator is a stdio MCP server that acts as a persistent knowledge base
// for Claude agents. Agents call remember() to record findings and consult()
// to ask questions backed by accumulated knowledge.
//
// Subcommands:
//
//	navigator                           — run as MCP stdio server (default)
//	navigator export <file.json.gz>     — dump all thoughts to gzipped JSON
//	navigator import <file.json.gz>     — load thoughts from gzipped JSON
//	navigator serve <repo-path>         — run spec indexing daemon
//	navigator index <repo-path>         — index spec files and exit
//	navigator query <text>              — query indexed spec files
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
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/mattdurham/bob/internal/navigator/agent"
	"github.com/mattdurham/bob/internal/navigator/daemon"
	"github.com/mattdurham/bob/internal/navigator/docker"
	"github.com/mattdurham/bob/internal/navigator/embedder"
	"github.com/mattdurham/bob/internal/navigator/indexer"
	"github.com/mattdurham/bob/internal/navigator/specstore"
	"github.com/mattdurham/bob/internal/navigator/store"
	"github.com/mattdurham/bob/internal/navigator/teiclient"
	"github.com/mattdurham/bob/internal/navigator/tools"
	"github.com/mattdurham/bob/internal/navigator/watcher"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type serveConfig struct {
	teiPort    int
	teiImage   string
	specDB     string
	dockerArgs string
}

type indexConfig struct {
	teiPort  int
	teiImage string
	specDB   string
}

type queryConfig struct {
	specDB string
	limit  int
}

func defaultSpecDB() string {
	exe, err := os.Executable()
	if err != nil {
		return filepath.Join(os.TempDir(), "specs.sqlite")
	}
	return filepath.Join(filepath.Dir(exe), "specs.sqlite")
}

func defaultSocketDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return os.TempDir()
	}
	dir := filepath.Join(home, ".bob", "navigator")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Printf("navigator: failed to create socket dir %s: %v", dir, err)
	}
	return dir
}

func main() {
	dbPath := flag.String("db", "", "path to thoughts database (default: ~/.bob/navigator/thoughts.db)")
	libPath := flag.String("lib", "", "path to llama.cpp lib dir (default: ~/.bob/navigator/lib)")
	modelPath := flag.String("model", "", "path to embedding GGUF model")
	teiPort := flag.Int("tei-port", 7462, "port for text-embeddings-inference container")
	teiImage := flag.String("tei-image", "ghcr.io/huggingface/text-embeddings-inference:latest", "TEI Docker image")
	specDB := flag.String("spec-db", defaultSpecDB(), "path to spec vector database")
	dockerArgs := flag.String("docker-args", "", "extra args to pass to docker run")
	limit := flag.Int("limit", 10, "max results for query subcommand")
	flag.Parse()

	// Resolve database path for the thoughts store.
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
	case "serve":
		repoPath := flag.Arg(1)
		if repoPath == "" {
			log.Fatal("usage: navigator serve <repo-path>")
		}
		cfg := serveConfig{
			teiPort:    *teiPort,
			teiImage:   *teiImage,
			specDB:     *specDB,
			dockerArgs: *dockerArgs,
		}
		if err := runServe(repoPath, cfg); err != nil {
			log.Fatalf("navigator serve: %v", err)
		}
	case "index":
		repoPath := flag.Arg(1)
		if repoPath == "" {
			log.Fatal("usage: navigator index <repo-path>")
		}
		cfg := indexConfig{
			teiPort:  *teiPort,
			teiImage: *teiImage,
			specDB:   *specDB,
		}
		if err := runIndex(repoPath, cfg); err != nil {
			log.Fatalf("navigator index: %v", err)
		}
	case "query":
		queryText := flag.Arg(1)
		if queryText == "" {
			log.Fatal("usage: navigator query <text>")
		}
		cfg := queryConfig{
			specDB: *specDB,
			limit:  *limit,
		}
		if err := runQuery(queryText, cfg); err != nil {
			log.Fatalf("navigator query: %v", err)
		}
	default:
		runServer(path, *libPath, *modelPath)
	}
}

func runServe(repoPath string, cfg serveConfig) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var extraArgs []string
	if cfg.dockerArgs != "" {
		extraArgs = strings.Fields(cfg.dockerArgs)
	}

	mgr := docker.New("navigator-tei", cfg.teiImage, cfg.teiPort, extraArgs)
	baseURL, err := mgr.EnsureRunning(ctx)
	if err != nil {
		return fmt.Errorf("docker: %w", err)
	}

	teiClient := teiclient.New(baseURL)
	if err := teiClient.Health(ctx); err != nil {
		return fmt.Errorf("TEI health check: %w", err)
	}

	ss, err := specstore.Open(cfg.specDB)
	if err != nil {
		return fmt.Errorf("open spec store: %w", err)
	}
	defer ss.Close()

	idx := indexer.New(ss, teiClient, repoPath)
	if err := idx.IndexAll(ctx); err != nil {
		return fmt.Errorf("initial index: %w", err)
	}

	socketDir := defaultSocketDir()
	socketPath, pidPath := daemon.Paths(socketDir)
	srv := daemon.NewServer(socketPath, pidPath, idx, ss, teiClient)

	w, err := watcher.New(func(path string, removed bool) {
		if err := idx.IndexFile(context.Background(), path, removed); err != nil {
			log.Printf("navigator: index error for %s: %v", path, err)
		}
	}, 200*time.Millisecond)
	if err != nil {
		return fmt.Errorf("watcher: %w", err)
	}
	defer w.Close()

	if err := w.Watch(ctx, repoPath); err != nil {
		return fmt.Errorf("watcher watch: %w", err)
	}

	return srv.Start(ctx)
}

func runIndex(repoPath string, cfg indexConfig) error {
	ctx := context.Background()

	mgr := docker.New("navigator-tei", cfg.teiImage, cfg.teiPort, nil)
	baseURL, err := mgr.EnsureRunning(ctx)
	if err != nil {
		return fmt.Errorf("docker: %w", err)
	}

	teiClient := teiclient.New(baseURL)
	if err := teiClient.Health(ctx); err != nil {
		return fmt.Errorf("TEI health check: %w", err)
	}

	ss, err := specstore.Open(cfg.specDB)
	if err != nil {
		return fmt.Errorf("open spec store: %w", err)
	}
	defer ss.Close()

	idx := indexer.New(ss, teiClient, repoPath)
	if err := idx.IndexAll(ctx); err != nil {
		return fmt.Errorf("index: %w", err)
	}

	count, err := ss.Count()
	if err != nil {
		return err
	}
	fmt.Printf("%d sections indexed\n", count)
	return nil
}

func runQuery(queryText string, cfg queryConfig) error {
	ctx := context.Background()

	socketDir := defaultSocketDir()
	socketPath, _ := daemon.Paths(socketDir)

	c := daemon.NewClient(socketPath)
	results, err := c.Query(ctx, queryText, cfg.limit)
	if err != nil {
		// Fallback: direct query without daemon
		ss, openErr := specstore.Open(cfg.specDB)
		if openErr != nil {
			return fmt.Errorf("daemon unavailable (%v) and could not open spec store: %w", err, openErr)
		}
		defer ss.Close()
		count, _ := ss.Count()
		if count == 0 {
			fmt.Println("No spec sections indexed. Run: navigator index <repo-path>")
			return nil
		}
		fmt.Fprintf(os.Stderr, "daemon not running, falling back to direct query (no embeddings available without TEI)\n")
		return nil
	}

	if len(results) == 0 {
		fmt.Println("No results found.")
		return nil
	}
	for _, r := range results {
		fmt.Printf("[distance=%.4f] %s\n  %s\n\n", r.Distance, r.Path, truncate(r.Text, 120))
	}
	return nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
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
