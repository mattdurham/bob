package specstore

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"
	_ "modernc.org/sqlite/vec"
)

const schema = `
CREATE TABLE IF NOT EXISTS specs (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    path      TEXT NOT NULL UNIQUE,
    text      TEXT NOT NULL,
    file_hash TEXT NOT NULL
);

CREATE VIRTUAL TABLE IF NOT EXISTS specs_vec USING vec0(
    embedding float[768]
);
`

// Store holds spec sections and their vector embeddings.
type Store struct {
	db *sql.DB
}

// Section is a stored spec section.
type Section struct {
	ID       int64
	Path     string
	Text     string
	FileHash string
}

// Result is a vector search result.
type Result struct {
	Path     string
	Text     string
	Distance float64
}

// Open opens (or creates) the spec store database.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("specstore: open: %w", err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("specstore: schema: %w", err)
	}
	return &Store{db: db}, nil
}

// Close closes the database.
func (s *Store) Close() error {
	return s.db.Close()
}

// Upsert inserts or replaces a spec section and its embedding.
func (s *Store) Upsert(sec *Section, embedding []float32) error {
	// Fetch the existing id for this path before replacing, so we can clean up
	// the old specs_vec row. INSERT OR REPLACE with AUTOINCREMENT deletes the
	// old row and creates a new one with a new id, leaving the old specs_vec
	// row as a ghost embedding if we don't remove it explicitly.
	var oldID int64
	var hasOld bool
	err := s.db.QueryRow(`SELECT id FROM specs WHERE path = ?`, sec.Path).Scan(&oldID)
	if err == nil {
		hasOld = true
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("specstore: lookup existing id: %w", err)
	}

	res, err := s.db.Exec(
		`INSERT OR REPLACE INTO specs (path, text, file_hash) VALUES (?, ?, ?)`,
		sec.Path, sec.Text, sec.FileHash,
	)
	if err != nil {
		return fmt.Errorf("specstore: upsert specs: %w", err)
	}
	newID, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("specstore: last insert id: %w", err)
	}

	// Delete the old vec row when the id changed (i.e., the row was replaced).
	if hasOld && oldID != newID {
		if _, err := s.db.Exec(`DELETE FROM specs_vec WHERE rowid = ?`, oldID); err != nil {
			return fmt.Errorf("specstore: delete old vec %d: %w", oldID, err)
		}
	}

	vec, err := json.Marshal(embedding)
	if err != nil {
		return fmt.Errorf("specstore: marshal embedding: %w", err)
	}
	if _, err := s.db.Exec(
		`INSERT OR REPLACE INTO specs_vec(rowid, embedding) VALUES (?, ?)`,
		newID, string(vec),
	); err != nil {
		return fmt.Errorf("specstore: upsert vec: %w", err)
	}
	return nil
}

// escapeLike escapes LIKE special characters (%, _, \) so that a literal
// string can be used as the prefix operand with ESCAPE '\'.
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

// DeleteByFilePrefix removes all specs rows where path starts with filePrefix,
// and their corresponding vec rows.
func (s *Store) DeleteByFilePrefix(filePrefix string) error {
	escaped := escapeLike(filePrefix)
	rows, err := s.db.Query(`SELECT id FROM specs WHERE path LIKE ? || '%' ESCAPE '\'`, escaped)
	if err != nil {
		return fmt.Errorf("specstore: query ids for delete: %w", err)
	}
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return fmt.Errorf("specstore: scan id: %w", err)
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return fmt.Errorf("specstore: rows error: %w", err)
	}

	for _, id := range ids {
		if _, err := s.db.Exec(`DELETE FROM specs_vec WHERE rowid=?`, id); err != nil {
			return fmt.Errorf("specstore: delete vec %d: %w", id, err)
		}
	}
	if _, err := s.db.Exec(`DELETE FROM specs WHERE path LIKE ? || '%' ESCAPE '\'`, escaped); err != nil {
		return fmt.Errorf("specstore: delete specs: %w", err)
	}
	return nil
}

// HashForFile returns the stored file_hash for any spec entry under filePath.
// Returns exists=false when no entries are found.
func (s *Store) HashForFile(filePath string) (string, bool, error) {
	var hash string
	err := s.db.QueryRow(
		`SELECT file_hash FROM specs WHERE path LIKE ? || '%' ESCAPE '\' LIMIT 1`,
		escapeLike(filePath),
	).Scan(&hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("specstore: hash for file: %w", err)
	}
	return hash, true, nil
}

// VectorSearch finds the nearest spec sections by embedding vector.
func (s *Store) VectorSearch(embedding []float32, limit int) ([]Result, error) {
	if limit <= 0 {
		limit = 10
	}

	// Pre-check: avoid fragile sqlite-vec error message matching by returning
	// early when the table is empty.
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM specs`).Scan(&count); err != nil {
		return nil, fmt.Errorf("specstore: pre-check count: %w", err)
	}
	if count == 0 {
		return nil, nil
	}

	vec, err := json.Marshal(embedding)
	if err != nil {
		return nil, fmt.Errorf("specstore: marshal query vec: %w", err)
	}

	rows, err := s.db.Query(`
		SELECT s.path, s.text, v.distance
		FROM specs s
		JOIN (
			SELECT rowid, distance
			FROM specs_vec
			WHERE embedding MATCH ?
			ORDER BY distance
			LIMIT ?
		) v ON s.id = v.rowid
		ORDER BY v.distance`,
		string(vec), limit,
	)
	if err != nil {
		return nil, fmt.Errorf("specstore: vector search: %w", err)
	}
	defer rows.Close()

	var results []Result
	for rows.Next() {
		var r Result
		if err := rows.Scan(&r.Path, &r.Text, &r.Distance); err != nil {
			return nil, fmt.Errorf("specstore: scan result: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// Count returns the total number of stored spec sections.
func (s *Store) Count() (int, error) {
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM specs`).Scan(&n); err != nil {
		return 0, fmt.Errorf("specstore: count: %w", err)
	}
	return n, nil
}
