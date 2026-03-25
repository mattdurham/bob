package spec_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mattdurham/bob/internal/firstmate/spec"
)

func writeSpecFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestFindAll(t *testing.T) {
	dir := t.TempDir()
	writeSpecFile(t, dir, "SPECS.md", "# Specs\n1. Do something")
	writeSpecFile(t, filepath.Join(dir, "sub"), "NOTES.md", "# Notes\n## Decision")

	specs, err := spec.FindAll(dir)
	if err != nil {
		t.Fatalf("find all: %v", err)
	}
	if len(specs) != 2 {
		t.Errorf("got %d spec files, want 2", len(specs))
	}
}

func TestReadByKind(t *testing.T) {
	dir := t.TempDir()
	writeSpecFile(t, dir, "SPECS.md", "# My Specs\n1. Invariant one")
	writeSpecFile(t, dir, "NOTES.md", "# Notes here")

	result, err := spec.ReadByKind(dir, "SPECS", "")
	if err != nil {
		t.Fatalf("read by kind: %v", err)
	}
	if !strings.Contains(result, "Invariant one") {
		t.Errorf("expected SPECS content, got: %s", result)
	}
}

func TestReadByKindWithPattern(t *testing.T) {
	dir := t.TempDir()
	writeSpecFile(t, dir, "SPECS.md", "# Specs\nThis mentions ratelimit")
	writeSpecFile(t, filepath.Join(dir, "sub"), "SPECS.md", "# Other\nNothing relevant")

	result, err := spec.ReadByKind(dir, "SPECS", "ratelimit")
	if err != nil {
		t.Fatalf("read by kind: %v", err)
	}
	if !strings.Contains(result, "ratelimit") {
		t.Errorf("expected filtered content, got: %s", result)
	}
	if strings.Contains(result, "Nothing relevant") {
		t.Error("should not include non-matching spec file")
	}
}

func TestSearchAll(t *testing.T) {
	dir := t.TempDir()
	writeSpecFile(t, dir, "SPECS.md", "# Specs\nThe TokenBucket interface is the entry point")

	result, err := spec.SearchAll(dir, "TokenBucket")
	if err != nil {
		t.Fatalf("search all: %v", err)
	}
	if len(result.Specs) == 0 {
		t.Error("expected at least one spec match, got none")
	}
	if len(result.Specs[0].MatchedLines) == 0 {
		t.Error("expected matched lines")
	}
}

func TestListAll(t *testing.T) {
	dir := t.TempDir()
	writeSpecFile(t, dir, "SPECS.md", "content")
	writeSpecFile(t, dir, "TESTS.md", "content")

	result, err := spec.ListAll(dir)
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if !strings.Contains(result, "SPECS") {
		t.Errorf("expected SPECS in list, got: %s", result)
	}
	if !strings.Contains(result, "TESTS") {
		t.Errorf("expected TESTS in list, got: %s", result)
	}
}

func TestListAllEmpty(t *testing.T) {
	dir := t.TempDir()
	result, err := spec.ListAll(dir)
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if !strings.Contains(result, "No spec files found") {
		t.Errorf("expected empty message, got: %s", result)
	}
}

func TestGetByID(t *testing.T) {
	dir := t.TempDir()
	writeSpecFile(t, dir, "SPECS.md", "# Specs\n## SPEC-001\nMust do X\n## SPEC-002\nMust do Y")

	result, err := spec.GetByID(dir, "SPEC-001")
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if !strings.Contains(result, "SPEC-001") {
		t.Errorf("expected SPEC-001 match, got: %s", result)
	}
}

func TestFindAllSkipsVendor(t *testing.T) {
	dir := t.TempDir()
	writeSpecFile(t, filepath.Join(dir, "vendor", "lib"), "SPECS.md", "vendor specs")
	writeSpecFile(t, dir, "SPECS.md", "real specs")

	specs, err := spec.FindAll(dir)
	if err != nil {
		t.Fatalf("find all: %v", err)
	}
	for _, s := range specs {
		if strings.Contains(s.Path, "vendor") {
			t.Errorf("should not include vendor spec: %s", s.Path)
		}
	}
}
