package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
	_ "modernc.org/sqlite/vec"
)

const schema = `
CREATE TABLE IF NOT EXISTS thoughts (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	content    TEXT    NOT NULL,
	repo       TEXT    NOT NULL DEFAULT '',
	scope      TEXT    NOT NULL DEFAULT '',
	tags       TEXT    NOT NULL DEFAULT '[]',
	confidence TEXT    NOT NULL DEFAULT 'observed',
	source     TEXT    NOT NULL DEFAULT '',
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE VIRTUAL TABLE IF NOT EXISTS thoughts_fts USING fts5(
	content,
	repo,
	scope,
	tags,
	content='thoughts',
	content_rowid='id'
);

CREATE TRIGGER IF NOT EXISTS thoughts_ai AFTER INSERT ON thoughts BEGIN
	INSERT INTO thoughts_fts(rowid, content, repo, scope, tags)
	VALUES (new.id, new.content, new.repo, new.scope, new.tags);
END;

CREATE TRIGGER IF NOT EXISTS thoughts_ad AFTER DELETE ON thoughts BEGIN
	INSERT INTO thoughts_fts(thoughts_fts, rowid, content, repo, scope, tags)
	VALUES ('delete', old.id, old.content, old.repo, old.scope, old.tags);
END;

CREATE TRIGGER IF NOT EXISTS thoughts_au AFTER UPDATE ON thoughts BEGIN
	INSERT INTO thoughts_fts(thoughts_fts, rowid, content, repo, scope, tags)
	VALUES ('delete', old.id, old.content, old.repo, old.scope, old.tags);
	INSERT INTO thoughts_fts(rowid, content, repo, scope, tags)
	VALUES (new.id, new.content, new.repo, new.scope, new.tags);
END;

-- chunk-to-thought mapping (multiple chunks per thought for long content)
CREATE TABLE IF NOT EXISTS thought_chunks (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	thought_id INTEGER NOT NULL,
	FOREIGN KEY (thought_id) REFERENCES thoughts(id) ON DELETE CASCADE
);

-- vec0 virtual table for vector search, rowid maps to thought_chunks.id
CREATE VIRTUAL TABLE IF NOT EXISTS thoughts_vec USING vec0(
	embedding float[768]
);

-- generic key-value metadata for thoughts
CREATE TABLE IF NOT EXISTS thought_meta (
	thought_id INTEGER NOT NULL,
	key        TEXT    NOT NULL,
	value      TEXT    NOT NULL,
	PRIMARY KEY (thought_id, key),
	FOREIGN KEY (thought_id) REFERENCES thoughts(id) ON DELETE CASCADE
);
`

// Thought is a stored finding or insight from an agent.
type Thought struct {
	ID         int64
	Content    string
	Repo       string // optional — e.g. "grafana/tempo", empty for general knowledge
	Scope      string
	Tags       []string
	Confidence string // "verified", "observed", "tentative"
	Source     string // "pr-review", "debugging", "code-review", etc.
	CreatedAt  time.Time
}

// Store manages the thoughts SQLite database.
type Store struct {
	db *sql.DB
}

// DBPath returns the default database path at ~/.bob/navigator/thoughts.db.
func DBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dir: %w", err)
	}
	dir := filepath.Join(home, ".bob", "navigator")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create data dir: %w", err)
	}
	return filepath.Join(dir, "thoughts.db"), nil
}

// Open opens (or creates) the database and applies the schema.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)

	// Migrate: add repo column if missing (existing DBs won't have it).
	db.Exec(`ALTER TABLE thoughts ADD COLUMN repo TEXT NOT NULL DEFAULT ''`)

	// Rebuild FTS and triggers to include repo column.
	db.Exec(`DROP TABLE IF EXISTS thoughts_fts`)
	db.Exec(`DROP TRIGGER IF EXISTS thoughts_ai`)
	db.Exec(`DROP TRIGGER IF EXISTS thoughts_ad`)
	db.Exec(`DROP TRIGGER IF EXISTS thoughts_au`)

	// Migrate vec table: if old vec table exists without chunk mapping, rebuild.
	// Drop both and let schema recreate — embeddings recomputed via reindex.
	var hasChunks int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE name='thought_chunks'`).Scan(&hasChunks); err == nil && hasChunks == 0 {
		db.Exec(`DROP TABLE IF EXISTS thoughts_vec`)
	}

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}

	// Rebuild FTS index from existing data.
	db.Exec(`INSERT INTO thoughts_fts(thoughts_fts) VALUES('rebuild')`)

	return &Store{db: db}, nil
}

// ThoughtsWithoutEmbeddings returns all thoughts that don't have any chunks
// in the vec table. Used for backfilling embeddings.
func (s *Store) ThoughtsWithoutEmbeddings() ([]*Thought, error) {
	rows, err := s.db.Query(`
		SELECT t.id, t.content, t.repo, t.scope, t.tags, t.confidence, t.source, t.created_at
		FROM thoughts t
		WHERE t.id NOT IN (SELECT DISTINCT thought_id FROM thought_chunks)
		ORDER BY t.id`)
	if err != nil {
		return nil, fmt.Errorf("query unembedded: %w", err)
	}
	defer rows.Close()
	return scanThoughts(rows)
}

// Stats returns the total number of thoughts and the number with embeddings.
func (s *Store) Stats() (total, embedded int, err error) {
	err = s.db.QueryRow(`SELECT COUNT(*) FROM thoughts`).Scan(&total)
	if err != nil {
		return 0, 0, err
	}
	err = s.db.QueryRow(`SELECT COUNT(DISTINCT thought_id) FROM thought_chunks`).Scan(&embedded)
	if err != nil {
		return total, 0, err
	}
	return total, embedded, nil
}

// All returns every thought in the database.
func (s *Store) All() ([]*Thought, error) {
	rows, err := s.db.Query(`
		SELECT id, content, repo, scope, tags, confidence, source, created_at
		FROM thoughts ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list all: %w", err)
	}
	defer rows.Close()
	return scanThoughts(rows)
}

// Update modifies an existing thought's content and metadata.
// Clears existing embeddings so the thought gets re-embedded on next reindex.
func (s *Store) Update(id int64, t *Thought) error {
	tags, err := json.Marshal(t.Tags)
	if err != nil {
		return fmt.Errorf("marshal tags: %w", err)
	}
	res, err := s.db.Exec(
		`UPDATE thoughts SET content=?, repo=?, scope=?, tags=?, confidence=?, source=? WHERE id=?`,
		t.Content, t.Repo, t.Scope, string(tags), t.Confidence, t.Source, id,
	)
	if err != nil {
		return fmt.Errorf("update: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("thought-%d not found", id)
	}
	// Clear old embeddings — content changed, need re-embed.
	s.deleteEmbeddings(id)
	return nil
}

// Delete removes a thought and its embeddings and metadata.
func (s *Store) Delete(id int64) error {
	res, err := s.db.Exec(`DELETE FROM thoughts WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("thought-%d not found", id)
	}
	s.deleteEmbeddings(id)
	s.db.Exec(`DELETE FROM thought_meta WHERE thought_id=?`, id)
	return nil
}

func (s *Store) deleteEmbeddings(thoughtID int64) {
	rows, _ := s.db.Query(`SELECT id FROM thought_chunks WHERE thought_id=?`, thoughtID)
	if rows != nil {
		var ids []int64
		for rows.Next() {
			var id int64
			rows.Scan(&id)
			ids = append(ids, id)
		}
		rows.Close()
		for _, id := range ids {
			s.db.Exec(`DELETE FROM thoughts_vec WHERE rowid=?`, id)
		}
	}
	s.db.Exec(`DELETE FROM thought_chunks WHERE thought_id=?`, thoughtID)
}

// Close closes the database.
func (s *Store) Close() error {
	return s.db.Close()
}

// Save stores a new thought and populates its ID.
func (s *Store) Save(t *Thought) error {
	tags, err := json.Marshal(t.Tags)
	if err != nil {
		return fmt.Errorf("marshal tags: %w", err)
	}
	res, err := s.db.Exec(
		`INSERT INTO thoughts (content, repo, scope, tags, confidence, source) VALUES (?, ?, ?, ?, ?, ?)`,
		t.Content, t.Repo, t.Scope, string(tags), t.Confidence, t.Source,
	)
	if err != nil {
		return fmt.Errorf("insert: %w", err)
	}
	t.ID, _ = res.LastInsertId()
	return nil
}

// SaveWithEmbedding stores a thought and its vector embedding together.
func (s *Store) SaveWithEmbedding(t *Thought, embedding []float32) error {
	if err := s.Save(t); err != nil {
		return err
	}
	return s.AddEmbedding(t.ID, embedding)
}

// AddEmbedding inserts a single embedding chunk for a thought.
func (s *Store) AddEmbedding(thoughtID int64, embedding []float32) error {
	res, err := s.db.Exec(`INSERT INTO thought_chunks (thought_id) VALUES (?)`, thoughtID)
	if err != nil {
		return fmt.Errorf("insert chunk: %w", err)
	}
	chunkID, _ := res.LastInsertId()

	vec, err := json.Marshal(embedding)
	if err != nil {
		return fmt.Errorf("marshal embedding: %w", err)
	}
	if _, err := s.db.Exec(
		`INSERT INTO thoughts_vec(rowid, embedding) VALUES (?, ?)`,
		chunkID, string(vec),
	); err != nil {
		return fmt.Errorf("insert embedding: %w", err)
	}
	return nil
}

// Search runs a full-text search against thoughts, optionally filtered by scope.
// Returns up to limit results ordered by FTS rank.
func (s *Store) Search(query, scope string, limit int) ([]*Thought, error) {
	if limit <= 0 {
		limit = 10
	}
	ftsQuery := sanitizeFTS(query)

	var (
		rows *sql.Rows
		err  error
	)
	if scope != "" {
		rows, err = s.db.Query(`
			SELECT t.id, t.content, t.repo, t.scope, t.tags, t.confidence, t.source, t.created_at
			FROM thoughts t
			JOIN thoughts_fts f ON t.id = f.rowid
			WHERE thoughts_fts MATCH ? AND t.scope LIKE ?
			ORDER BY rank
			LIMIT ?`,
			ftsQuery, "%"+scope+"%", limit,
		)
	} else {
		rows, err = s.db.Query(`
			SELECT t.id, t.content, t.repo, t.scope, t.tags, t.confidence, t.source, t.created_at
			FROM thoughts t
			JOIN thoughts_fts f ON t.id = f.rowid
			WHERE thoughts_fts MATCH ?
			ORDER BY rank
			LIMIT ?`,
			ftsQuery, limit,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("fts search: %w", err)
	}
	defer rows.Close()
	return scanThoughts(rows)
}

// VectorSearch finds the nearest thoughts by embedding vector.
func (s *Store) VectorSearch(embedding []float32, scope string, limit int) ([]*Thought, error) {
	if limit <= 0 {
		limit = 10
	}
	vec, err := json.Marshal(embedding)
	if err != nil {
		return nil, fmt.Errorf("marshal query vec: %w", err)
	}

	// Join vec → chunks → thoughts, dedup by thought_id keeping best distance.
	var rows *sql.Rows
	if scope != "" {
		rows, err = s.db.Query(`
			SELECT t.id, t.content, t.repo, t.scope, t.tags, t.confidence, t.source, t.created_at
			FROM thoughts t
			JOIN thought_chunks c ON t.id = c.thought_id
			JOIN (
				SELECT rowid, distance
				FROM thoughts_vec
				WHERE embedding MATCH ?
				ORDER BY distance
				LIMIT ?
			) v ON c.id = v.rowid
			WHERE t.scope LIKE ?
			GROUP BY t.id
			ORDER BY MIN(v.distance)`,
			string(vec), limit*4, "%"+scope+"%",
		)
	} else {
		rows, err = s.db.Query(`
			SELECT t.id, t.content, t.repo, t.scope, t.tags, t.confidence, t.source, t.created_at
			FROM thoughts t
			JOIN thought_chunks c ON t.id = c.thought_id
			JOIN (
				SELECT rowid, distance
				FROM thoughts_vec
				WHERE embedding MATCH ?
				ORDER BY distance
				LIMIT ?
			) v ON c.id = v.rowid
			GROUP BY t.id
			ORDER BY MIN(v.distance)`,
			string(vec), limit*4,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("vector search: %w", err)
	}
	defer rows.Close()
	return scanThoughts(rows)
}

func scanThoughts(rows *sql.Rows) ([]*Thought, error) {
	var thoughts []*Thought
	for rows.Next() {
		t := &Thought{}
		var tags, createdAt string
		if err := rows.Scan(&t.ID, &t.Content, &t.Repo, &t.Scope, &tags, &t.Confidence, &t.Source, &createdAt); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		_ = json.Unmarshal([]byte(tags), &t.Tags)
		t.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		thoughts = append(thoughts, t)
	}
	return thoughts, rows.Err()
}

// DeleteMeta removes all key-value pairs for a thought.
func (s *Store) DeleteMeta(thoughtID int64) {
	s.db.Exec(`DELETE FROM thought_meta WHERE thought_id=?`, thoughtID)
}

// SetMeta sets a key-value pair on a thought.
func (s *Store) SetMeta(thoughtID int64, key, value string) error {
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO thought_meta (thought_id, key, value) VALUES (?, ?, ?)`,
		thoughtID, key, value,
	)
	return err
}

// GetMeta returns all key-value pairs for a thought.
func (s *Store) GetMeta(thoughtID int64) (map[string]string, error) {
	rows, err := s.db.Query(`SELECT key, value FROM thought_meta WHERE thought_id = ?`, thoughtID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	meta := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		meta[k] = v
	}
	return meta, rows.Err()
}

// sanitizeFTS converts a free-text query into FTS5 OR-joined terms
// so any matching word scores a hit, ranked by overlap.
func sanitizeFTS(q string) string {
	words := strings.Fields(q)
	if len(words) == 0 {
		return q
	}
	for i, w := range words {
		w = strings.ReplaceAll(w, `"`, `""`)
		words[i] = `"` + w + `"`
	}
	return strings.Join(words, " OR ")
}
