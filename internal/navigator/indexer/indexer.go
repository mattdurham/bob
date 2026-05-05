package indexer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/mattdurham/bob/internal/navigator/mdparser"
	"github.com/mattdurham/bob/internal/navigator/specstore"
)

// Embedder is satisfied by teiclient.Client.
type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
}

// Store is satisfied by specstore.Store.
type Store interface {
	Upsert(sec *specstore.Section, embedding []float32) error
	DeleteByFilePrefix(filePrefix string) error
	HashForFile(filePath string) (string, bool, error)
	Count() (int, error)
}

// Indexer orchestrates the full scan→parse→embed→store pipeline.
type Indexer struct {
	store    Store
	embedder Embedder
	root     string
}

// New creates an Indexer.
func New(store Store, embedder Embedder, root string) *Indexer {
	return &Indexer{store: store, embedder: embedder, root: root}
}

var specFileSet = map[string]bool{
	"SPECS.md":      true,
	"NOTES.md":      true,
	"BENCHMARKS.md": true,
	"TESTS.md":      true,
	"CLAUDE.md":     true,
}

func isSpecFilename(name string) bool {
	return specFileSet[filepath.Base(name)]
}

// IndexAll walks root and indexes all spec files not already up to date.
// It pre-computes each file's hash to skip unchanged files and passes the
// pre-computed hash to the internal indexing logic to avoid a second disk read.
func (idx *Indexer) IndexAll(ctx context.Context) error {
	return filepath.WalkDir(idx.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !isSpecFilename(path) {
			return nil
		}
		rel, err := relPath(idx.root, path)
		if err != nil {
			return err
		}
		fileHash, err := hashFile(path)
		if err != nil {
			return fmt.Errorf("indexer: hash %s: %w", path, err)
		}
		storedHash, exists, err := idx.store.HashForFile(rel)
		if err != nil {
			return fmt.Errorf("indexer: hash lookup %s: %w", path, err)
		}
		if exists && storedHash == fileHash {
			return nil
		}
		// Pass the already-computed hash to avoid a second disk read inside
		// the indexing pipeline.
		return idx.indexFileKnownHash(ctx, path, rel, fileHash)
	})
}

// IndexFile indexes (or re-indexes) a single file. If removed=true, deletes its entries.
func (idx *Indexer) IndexFile(ctx context.Context, absPath string, removed bool) error {
	rel, err := relPath(idx.root, absPath)
	if err != nil {
		rel = filepath.Base(absPath)
	}

	if removed {
		return idx.store.DeleteByFilePrefix(rel)
	}

	fileHash, err := hashFile(absPath)
	if err != nil {
		return fmt.Errorf("indexer: hash %s: %w", absPath, err)
	}

	storedHash, exists, err := idx.store.HashForFile(rel)
	if err != nil {
		return fmt.Errorf("indexer: hash lookup: %w", err)
	}
	if exists && storedHash == fileHash {
		return nil
	}

	return idx.indexFileKnownHash(ctx, absPath, rel, fileHash)
}

// indexFileKnownHash performs the embed+store pipeline for a file whose hash
// has already been computed, avoiding a second disk read.
func (idx *Indexer) indexFileKnownHash(ctx context.Context, absPath, rel, fileHash string) error {
	content, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("indexer: read %s: %w", absPath, err)
	}

	sections := mdparser.Parse(filepath.Base(absPath), string(content))
	if len(sections) == 0 {
		return nil
	}

	texts := make([]string, len(sections))
	for i, s := range sections {
		texts[i] = s.Text
	}

	// Batch all sections for this file in a single Embed call
	embeddings, err := idx.embedder.Embed(ctx, texts)
	if err != nil {
		return fmt.Errorf("indexer: embed %s: %w", absPath, err)
	}
	if len(embeddings) != len(sections) {
		return fmt.Errorf("indexer: embed returned %d vectors for %d sections", len(embeddings), len(sections))
	}

	if err := idx.store.DeleteByFilePrefix(rel); err != nil {
		return fmt.Errorf("indexer: delete stale %s: %w", rel, err)
	}

	for i, sec := range sections {
		storeSec := &specstore.Section{
			Path:     rel + "#" + sec.ID,
			Text:     sec.Text,
			FileHash: fileHash,
		}
		if err := idx.store.Upsert(storeSec, embeddings[i]); err != nil {
			return fmt.Errorf("indexer: upsert %s: %w", storeSec.Path, err)
		}
	}
	return nil
}

func hashFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:]), nil
}

func relPath(root, abs string) (string, error) {
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return "", fmt.Errorf("indexer: rel path: %w", err)
	}
	return rel, nil
}
