package mdparser

import (
	"strings"
	"testing"
)

func TestParse_H2_BasicSection(t *testing.T) {
	content := "## 1. Foo\nsome text here"
	sections := Parse("SPECS.md", content)
	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}
	if sections[0].ID != "section-1-foo" {
		t.Errorf("expected ID section-1-foo, got %s", sections[0].ID)
	}
	if !strings.Contains(sections[0].Text, "Foo") {
		t.Errorf("text should contain heading: %s", sections[0].Text)
	}
	if !strings.Contains(sections[0].Text, "some text here") {
		t.Errorf("text should contain body: %s", sections[0].Text)
	}
}

func TestParse_H2_MultipleSection(t *testing.T) {
	content := "## 1. Alpha\nbody one here and more\n## 2. Beta\nbody two here and more"
	sections := Parse("SPECS.md", content)
	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(sections))
	}
	if sections[0].ID != "section-1-alpha" {
		t.Errorf("expected section-1-alpha, got %s", sections[0].ID)
	}
	if sections[1].ID != "section-2-beta" {
		t.Errorf("expected section-2-beta, got %s", sections[1].ID)
	}
}

func TestParse_H2_EmptySection(t *testing.T) {
	// Heading text itself is >= 20 chars, no body — still returned
	content := "## This Is A Long Enough Heading Title\n"
	sections := Parse("SPECS.md", content)
	if len(sections) != 1 {
		t.Fatalf("expected 1 section for long heading, got %d", len(sections))
	}
}

func TestParse_H2_ShortSection(t *testing.T) {
	// Total text < 20 chars → skipped
	content := "## Foo\nbar"
	sections := Parse("SPECS.md", content)
	if len(sections) != 0 {
		t.Fatalf("expected 0 sections (too short), got %d", len(sections))
	}
}

func TestParse_H2_Numbering(t *testing.T) {
	content := "## 2. Rate Limiting\nsome description text here"
	sections := Parse("SPECS.md", content)
	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}
	if sections[0].ID != "section-2-rate-limiting" {
		t.Errorf("expected section-2-rate-limiting, got %s", sections[0].ID)
	}
}

func TestParse_H2_NoNumber(t *testing.T) {
	content := "## Architecture\nsome description text here about the architecture"
	sections := Parse("SPECS.md", content)
	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}
	if sections[0].ID != "section-architecture" {
		t.Errorf("expected section-architecture, got %s", sections[0].ID)
	}
}

func TestParse_H2_H1Skipped(t *testing.T) {
	content := "# Title\npreamble content\n## 1. Real Section\nbody text here and more content"
	sections := Parse("SPECS.md", content)
	if len(sections) != 1 {
		t.Fatalf("expected 1 section (H1 is preamble), got %d", len(sections))
	}
	if sections[0].ID != "section-1-real-section" {
		t.Errorf("expected section-1-real-section, got %s", sections[0].ID)
	}
}

func TestParse_CLAUDE_BasicInvariant(t *testing.T) {
	content := "1. Foo bar baz this is a longer invariant"
	sections := Parse("CLAUDE.md", content)
	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}
	if sections[0].ID != "invariant-1" {
		t.Errorf("expected invariant-1, got %s", sections[0].ID)
	}
	if !strings.Contains(sections[0].Text, "Foo bar baz") {
		t.Errorf("text should contain invariant text: %s", sections[0].Text)
	}
}

func TestParse_CLAUDE_MultipleInvariants(t *testing.T) {
	content := "1. First invariant text that is long enough\n2. Second invariant text long enough\n3. Third invariant text long enough"
	sections := Parse("CLAUDE.md", content)
	if len(sections) != 3 {
		t.Fatalf("expected 3 sections, got %d", len(sections))
	}
	if sections[0].ID != "invariant-1" {
		t.Errorf("expected invariant-1, got %s", sections[0].ID)
	}
	if sections[2].ID != "invariant-3" {
		t.Errorf("expected invariant-3, got %s", sections[2].ID)
	}
}

func TestParse_CLAUDE_ShortInvariant(t *testing.T) {
	content := "1. Too short"
	sections := Parse("CLAUDE.md", content)
	if len(sections) != 0 {
		t.Fatalf("expected 0 sections (too short), got %d", len(sections))
	}
}

func TestParse_UnknownFilename(t *testing.T) {
	content := "## Some Section\nbody text here and more content"
	sections := Parse("README.md", content)
	if len(sections) != 0 {
		t.Fatalf("expected 0 sections for unknown filename, got %d", len(sections))
	}
}

func TestSlugify_SpecialChars(t *testing.T) {
	result := slugify("Rate-Limiting / v2")
	if result != "rate-limiting-v2" {
		t.Errorf("expected rate-limiting-v2, got %s", result)
	}
}

func TestSlugify_LeadingNumber(t *testing.T) {
	result := slugify("2. Foo")
	if result != "2-foo" {
		t.Errorf("expected 2-foo, got %s", result)
	}
}
