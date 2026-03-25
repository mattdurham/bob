package analysis

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// RaceFinding represents a potential race condition detected by heuristic analysis.
type RaceFinding struct {
	File    string
	Line    int
	Message string
}

// FindRaces performs heuristic AST analysis to detect potential race conditions.
func FindRaces(root string) ([]RaceFinding, error) {
	var findings []RaceFinding
	fset := token.NewFileSet()

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == "vendor" || name == "testdata" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		src, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		f, err := parser.ParseFile(fset, path, src, 0)
		if err != nil {
			return nil
		}

		rel, _ := filepath.Rel(root, path)
		fileFindings := analyzeFile(f, fset, rel)
		findings = append(findings, fileFindings...)
		return nil
	})
	return findings, err
}

func analyzeFile(f *ast.File, fset *token.FileSet, relPath string) []RaceFinding {
	var findings []RaceFinding

	ast.Inspect(f, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.GoStmt:
			// Check for goroutine closure capturing loop variables
			if fn, ok := node.Call.Fun.(*ast.FuncLit); ok {
				if capturesLoopVars(fn) {
					pos := fset.Position(node.Pos())
					findings = append(findings, RaceFinding{
						File:    relPath,
						Line:    pos.Line,
						Message: "goroutine closure may capture loop variable",
					})
				}
			}
		}
		return true
	})

	// Check for WaitGroup misuse: Done called before Add
	findings = append(findings, checkWaitGroup(f, fset, relPath)...)

	return findings
}

// capturesLoopVars is a simplified heuristic: if the goroutine closure
// has no parameters but references identifiers that look like loop vars.
func capturesLoopVars(fn *ast.FuncLit) bool {
	// If function literal takes no params and calls something with typical loop var names
	if fn.Type.Params != nil && len(fn.Type.Params.List) > 0 {
		return false // has params, likely safe
	}
	found := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		ident, ok := n.(*ast.Ident)
		if ok && (ident.Name == "i" || ident.Name == "j" || ident.Name == "v" || ident.Name == "k") {
			found = true
		}
		return !found
	})
	return found
}

func checkWaitGroup(f *ast.File, fset *token.FileSet, relPath string) []RaceFinding {
	var findings []RaceFinding
	// Look for wg.Done() inside goroutines without corresponding wg.Add()
	// This is a simplified heuristic
	ast.Inspect(f, func(n ast.Node) bool {
		goStmt, ok := n.(*ast.GoStmt)
		if !ok {
			return true
		}
		fn, ok := goStmt.Call.Fun.(*ast.FuncLit)
		if !ok {
			return true
		}
		hasDone := false
		ast.Inspect(fn.Body, func(inner ast.Node) bool {
			call, ok := inner.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if ok && sel.Sel.Name == "Done" {
				hasDone = true
			}
			return true
		})
		if hasDone {
			pos := fset.Position(goStmt.Pos())
			findings = append(findings, RaceFinding{
				File:    relPath,
				Line:    pos.Line,
				Message: "goroutine calls Done() - ensure Add() was called before goroutine launch",
			})
		}
		return true
	})
	return findings
}
