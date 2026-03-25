package parser_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mattdurham/bob/internal/firstmate/parser"
)

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

func TestParseDir_BasicFunction(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "main.go", `package main

func Add(a, b int) int {
	return a + b
}

func main() {
	Add(1, 2)
}
`)

	p := parser.New()
	g, err := p.ParseDir(dir)
	if err != nil {
		t.Fatalf("parse dir: %v", err)
	}

	nodes := g.Nodes()
	if len(nodes) == 0 {
		t.Fatal("expected nodes, got none")
	}

	// Verify Add function node exists
	addNode, ok := g.GetNode("main.Add")
	if !ok {
		t.Error("expected to find node main.Add")
	} else {
		if addNode.Kind != "function" {
			t.Errorf("got kind %q, want function", addNode.Kind)
		}
		if addNode.Cyclomatic < 1 {
			t.Errorf("cyclomatic should be >= 1, got %d", addNode.Cyclomatic)
		}
	}
}

func TestParseDir_Interface(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "types.go", `package mypack

type Reader interface {
	Read(p []byte) (n int, err error)
}
`)

	p := parser.New()
	g, err := p.ParseDir(dir)
	if err != nil {
		t.Fatalf("parse dir: %v", err)
	}

	n, ok := g.GetNode("mypack.Reader")
	if !ok {
		t.Error("expected to find node mypack.Reader")
	} else if n.Kind != "interface" {
		t.Errorf("got kind %q, want interface", n.Kind)
	}
}

func TestParseDir_PackageAndFileNodes(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "foo.go", `package foo

func Hello() string {
	return "hello"
}
`)

	p := parser.New()
	g, err := p.ParseDir(dir)
	if err != nil {
		t.Fatalf("parse dir: %v", err)
	}

	// Check package node
	_, ok := g.GetNode("pkg:foo")
	if !ok {
		t.Error("expected pkg:foo node")
	}

	// Check file node
	_, ok = g.GetNode("file:foo.go")
	if !ok {
		t.Error("expected file:foo.go node")
	}
}

func TestParseDir_MethodReceiver(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "service.go", `package svc

type Service struct{}

func (s *Service) Run() error {
	return nil
}
`)

	p := parser.New()
	g, err := p.ParseDir(dir)
	if err != nil {
		t.Fatalf("parse dir: %v", err)
	}

	// Method node should have receiver
	var found bool
	for _, n := range g.Nodes() {
		if n.Name == "Run" && n.Kind == "function" {
			if n.Receiver == "" {
				t.Error("expected non-empty receiver for Run method")
			}
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find Run method node")
	}
}

func TestParseDir_SkipsVendor(t *testing.T) {
	dir := t.TempDir()
	vendorDir := filepath.Join(dir, "vendor", "somelib")
	os.MkdirAll(vendorDir, 0o755)
	writeFile(t, vendorDir, "lib.go", `package somelib
func Secret() {}
`)
	writeFile(t, dir, "main.go", `package main
func main() {}
`)

	p := parser.New()
	g, err := p.ParseDir(dir)
	if err != nil {
		t.Fatalf("parse dir: %v", err)
	}

	for _, n := range g.Nodes() {
		if n.Name == "Secret" {
			t.Error("should not have parsed vendor directory")
		}
	}
}
