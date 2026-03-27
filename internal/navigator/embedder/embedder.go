// Package embedder generates vector embeddings from text using a local
// nomic-embed-text GGUF model via yzma (llama.cpp Go bindings, no CGO).
package embedder

import (
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hybridgroup/yzma/pkg/llama"
)

const (
	// Dims is the embedding dimension for nomic-embed-text-v1.5.
	Dims = 768
)

// EmbedJob represents a pending embedding request.
type EmbedJob struct {
	ThoughtID int64
	Content   string
}

// StoreFunc is called by the background worker to persist an embedding.
type StoreFunc func(thoughtID int64, embedding []float32) error

// Embedder generates float32 embeddings from text. It processes embedding
// jobs asynchronously via a background worker to avoid blocking MCP calls.
type Embedder struct {
	mu    sync.Mutex
	model llama.Model
	vocab llama.Vocab

	queue chan EmbedJob
	done  chan struct{}
}

// DefaultLibPath returns the default directory containing llama.cpp shared libraries.
func DefaultLibPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".bob", "navigator", "lib")
}

// DefaultModelPath returns the default path for the GGUF model file.
func DefaultModelPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".bob", "navigator", "models", "nomic-embed-text-v1.5.Q8_0.gguf")
}

// New loads the llama.cpp library and embedding model. Returns nil, nil if
// the library or model files are not present (embeddings not installed).
// The storeFn callback is invoked by the background worker to persist embeddings.
func New(libPath, modelPath string, storeFn StoreFunc) (*Embedder, error) {
	// Check that lib directory and model file exist; skip silently if not installed.
	if info, err := os.Stat(libPath); err != nil || !info.IsDir() {
		return nil, nil
	}
	if _, err := os.Stat(modelPath); err != nil {
		return nil, nil
	}

	if err := llama.Load(libPath); err != nil {
		return nil, fmt.Errorf("load llama library: %w", err)
	}

	llama.LogSet(llama.LogSilent())
	llama.Init()

	params := llama.ModelDefaultParams()
	model, err := llama.ModelLoadFromFile(modelPath, params)
	if err != nil {
		return nil, fmt.Errorf("load model: %w", err)
	}

	vocab := llama.ModelGetVocab(model)

	e := &Embedder{
		model: model,
		vocab: vocab,
		queue: make(chan EmbedJob, 256),
		done:  make(chan struct{}),
	}

	go e.worker(storeFn)

	return e, nil
}

// Enqueue adds a thought to the embedding queue. Returns immediately.
func (e *Embedder) Enqueue(id int64, content string) {
	select {
	case e.queue <- EmbedJob{ThoughtID: id, Content: content}:
	default:
		log.Printf("navigator: embedding queue full, dropping thought-%d", id)
	}
}

// Embed returns a normalized 768-dim float32 vector for the given text.
// Used synchronously for query-time embedding (recall).
func (e *Embedder) Embed(text string) ([]float32, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.embedLocked(text)
}

// Pending returns the number of jobs waiting in the queue.
func (e *Embedder) Pending() int {
	return len(e.queue)
}

// EmbedChunks splits text into chunks and returns an embedding for each.
// Short texts produce a single chunk. Used for synchronous reindex.
func (e *Embedder) EmbedChunks(text string) ([][]float32, error) {
	chunks := chunkText(text)
	vecs := make([][]float32, 0, len(chunks))
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, chunk := range chunks {
		vec, err := e.embedLocked(chunk)
		if err != nil {
			return nil, err
		}
		vecs = append(vecs, vec)
	}
	return vecs, nil
}

func (e *Embedder) worker(storeFn StoreFunc) {
	for job := range e.queue {
		chunks := chunkText(job.Content)
		for _, chunk := range chunks {
			e.mu.Lock()
			vec, err := e.embedLocked(chunk)
			e.mu.Unlock()

			if err != nil {
				log.Printf("navigator: embed thought-%d chunk failed: %v", job.ThoughtID, err)
				break
			}
			if err := storeFn(job.ThoughtID, vec); err != nil {
				log.Printf("navigator: store embedding thought-%d failed: %v", job.ThoughtID, err)
				break
			}
		}
	}
	close(e.done)
}

// chunkText splits text at paragraph boundaries into pieces that fit the
// embedding model's context window (~6000 chars ≈ ~2000 tokens, safe margin
// under nomic-embed-text's 8192 token limit). Short texts return as-is.
func chunkText(text string) []string {
	const maxChunkChars = 6000
	if len(text) <= maxChunkChars {
		return []string{text}
	}

	paragraphs := strings.Split(text, "\n\n")
	var chunks []string
	var current strings.Builder

	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		// If adding this paragraph exceeds the limit, flush current chunk.
		if current.Len() > 0 && current.Len()+len(p)+2 > maxChunkChars {
			chunks = append(chunks, current.String())
			current.Reset()
		}
		// If a single paragraph exceeds the limit, split it by lines.
		if len(p) > maxChunkChars {
			if current.Len() > 0 {
				chunks = append(chunks, current.String())
				current.Reset()
			}
			lines := strings.Split(p, "\n")
			for _, line := range lines {
				if current.Len() > 0 && current.Len()+len(line)+1 > maxChunkChars {
					chunks = append(chunks, current.String())
					current.Reset()
				}
				if current.Len() > 0 {
					current.WriteByte('\n')
				}
				current.WriteString(line)
			}
			continue
		}
		if current.Len() > 0 {
			current.WriteString("\n\n")
		}
		current.WriteString(p)
	}
	if current.Len() > 0 {
		chunks = append(chunks, current.String())
	}
	return chunks
}

func (e *Embedder) embedLocked(text string) ([]float32, error) {
	tokens := llama.Tokenize(e.vocab, text, true, true)

	n := uint32(len(tokens) + 16)
	ctxParams := llama.ContextDefaultParams()
	ctxParams.NCtx = n
	ctxParams.NBatch = n
	ctxParams.NUbatch = n
	ctxParams.Embeddings = 1

	ctx, err := llama.InitFromModel(e.model, ctxParams)
	if err != nil {
		return nil, fmt.Errorf("init context: %w", err)
	}
	defer llama.Free(ctx)

	batch := llama.BatchGetOne(tokens)
	ret, err := llama.Decode(ctx, batch)
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	if ret != 0 {
		return nil, fmt.Errorf("decode returned %d", ret)
	}

	nEmbd := llama.ModelNEmbd(e.model)
	vec, err := llama.GetEmbeddingsSeq(ctx, 0, nEmbd)
	if err != nil {
		return nil, fmt.Errorf("get embeddings: %w", err)
	}

	// Normalize.
	var sum float64
	for _, v := range vec {
		sum += float64(v * v)
	}
	norm := float32(1.0 / math.Sqrt(sum))
	for i := range vec {
		vec[i] *= norm
	}

	return vec, nil
}

// Close drains the queue and frees the model.
func (e *Embedder) Close() {
	close(e.queue)
	<-e.done
	if e.model != 0 {
		llama.ModelFree(e.model)
	}
	llama.Close()
}
