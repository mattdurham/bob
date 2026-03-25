package parser_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	fmparser "github.com/mattdurham/bob/internal/firstmate/parser"
)

func parseFunc(t *testing.T, src string) *ast.FuncDecl {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", "package p\n"+src, 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	for _, decl := range f.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			return fn
		}
	}
	t.Fatal("no func decl found")
	return nil
}

func TestCyclomaticComplexity_Simple(t *testing.T) {
	fn := parseFunc(t, `func f() int { return 1 }`)
	got := fmparser.CyclomaticComplexity(fn.Body)
	if got != 1 {
		t.Errorf("got %d, want 1", got)
	}
}

func TestCyclomaticComplexity_WithIf(t *testing.T) {
	fn := parseFunc(t, `func f(x int) int {
		if x > 0 {
			return x
		}
		return -x
	}`)
	got := fmparser.CyclomaticComplexity(fn.Body)
	if got != 2 {
		t.Errorf("got %d, want 2 (1 base + 1 if)", got)
	}
}

func TestCyclomaticComplexity_ForLoop(t *testing.T) {
	fn := parseFunc(t, `func f(n int) int {
		sum := 0
		for i := 0; i < n; i++ {
			sum += i
		}
		return sum
	}`)
	got := fmparser.CyclomaticComplexity(fn.Body)
	if got < 2 {
		t.Errorf("got %d, want >= 2 (for loop adds decision point)", got)
	}
}

func TestCyclomaticComplexity_NilBody(t *testing.T) {
	got := fmparser.CyclomaticComplexity(nil)
	if got != 1 {
		t.Errorf("got %d, want 1 for nil body", got)
	}
}

func TestCognitiveComplexity_Simple(t *testing.T) {
	fn := parseFunc(t, `func f() int { return 1 }`)
	got := fmparser.CognitiveComplexity(fn.Body)
	if got != 0 {
		t.Errorf("got %d, want 0 for trivial function", got)
	}
}

func TestCognitiveComplexity_Nested(t *testing.T) {
	fn := parseFunc(t, `func f(x, y int) int {
		if x > 0 {
			if y > 0 {
				return x + y
			}
		}
		return 0
	}`)
	got := fmparser.CognitiveComplexity(fn.Body)
	// Outer if: 1+0=1, inner if: 1+1=2, total=3
	if got < 2 {
		t.Errorf("got %d, want >= 2 for nested ifs", got)
	}
}
