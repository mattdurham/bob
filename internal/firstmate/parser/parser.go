package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/mattdurham/bob/internal/firstmate/graph"
)

// Parser parses Go source files into a graph.
type Parser struct {
	fset *token.FileSet
}

// New creates a new Parser.
func New() *Parser {
	return &Parser{fset: token.NewFileSet()}
}

// ParseDir parses all Go files under root and returns a graph.
func (p *Parser) ParseDir(root string) (*graph.Graph, error) {
	g := graph.New()
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable paths
		}
		if d.IsDir() {
			// Skip vendor, testdata, hidden directories
			name := d.Name()
			if name == "vendor" || name == "testdata" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		if err := p.parseFile(path, root, g); err != nil {
			// Non-fatal: log and continue
			_ = err
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk dir: %w", err)
	}
	p.resolveCallEdges(g)
	p.resolveImplements(g)
	return g, nil
}

// ParseFiles parses specific files.
func (p *Parser) ParseFiles(files []string, root string) (*graph.Graph, error) {
	g := graph.New()
	for _, path := range files {
		if err := p.parseFile(path, root, g); err != nil {
			_ = err
		}
	}
	p.resolveCallEdges(g)
	return g, nil
}

func (p *Parser) parseFile(path, root string, g *graph.Graph) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file %s: %w", path, err)
	}

	f, err := parser.ParseFile(p.fset, path, src, parser.ParseComments)
	if err != nil {
		// Try without strict mode — partial parse is better than nothing
		f, err = parser.ParseFile(p.fset, path, src, parser.ParseComments|parser.AllErrors)
		if err != nil {
			return fmt.Errorf("parse file %s: %w", path, err)
		}
	}

	rel, err := filepath.Rel(root, path)
	if err != nil {
		rel = path
	}

	pkgName := f.Name.Name
	pkgID := "pkg:" + pkgName

	// Ensure package node exists
	if _, ok := g.GetNode(pkgID); !ok {
		g.AddNode(&graph.Node{
			ID:       pkgID,
			Kind:     "package",
			Name:     pkgName,
			File:     rel,
			ChildIDs: []string{},
			CalleeIDs: []string{},
			CallerIDs: []string{},
		})
	}

	// File node
	fileID := "file:" + rel
	fileNode := &graph.Node{
		ID:        fileID,
		Kind:      "file",
		Name:      filepath.Base(path),
		File:      rel,
		ParentID:  pkgID,
		ChildIDs:  []string{},
		CalleeIDs: []string{},
		CallerIDs: []string{},
	}
	g.AddNode(fileNode)

	// Package -> file edge
	g.AddEdge(&graph.Edge{From: pkgID, To: fileID, Kind: "contains"})
	if pkgNode, ok := g.GetNode(pkgID); ok {
		pkgNode.AddChild(fileID)
	}

	// Import edges
	for _, imp := range f.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		importID := "pkg:" + importPath
		// Ensure external package node exists
		if _, ok := g.GetNode(importID); !ok {
			g.AddNode(&graph.Node{
				ID:        importID,
				Kind:      "package",
				Name:      importPath,
				External:  true,
				ChildIDs:  []string{},
				CalleeIDs: []string{},
				CallerIDs: []string{},
			})
		}
		g.AddEdge(&graph.Edge{From: fileID, To: importID, Kind: "imports"})
	}

	srcStr := string(src)

	// Walk declarations
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			p.parseFuncDecl(d, pkgName, fileID, rel, srcStr, g)
		case *ast.GenDecl:
			p.parseGenDecl(d, pkgName, fileID, rel, srcStr, g)
		}
	}

	return nil
}

func (p *Parser) parseFuncDecl(d *ast.FuncDecl, pkgName, fileID, rel, srcStr string, g *graph.Graph) {
	if d.Name == nil {
		return
	}

	name := d.Name.Name
	receiver := ""

	nodeID := pkgName + "." + name
	if d.Recv != nil && len(d.Recv.List) > 0 {
		recv := d.Recv.List[0].Type
		receiver = typeString(recv)
		nodeID = pkgName + ".(*" + strings.TrimPrefix(receiver, "*") + ")." + name
	}

	pos := p.fset.Position(d.Pos())
	params := fieldListString(d.Type.Params)
	returns := fieldListString(d.Type.Results)
	text := extractText(srcStr, p.fset, d.Pos(), d.End())

	cyclo := CyclomaticComplexity(d.Body)
	cogn := CognitiveComplexity(d.Body)

	node := &graph.Node{
		ID:         nodeID,
		Kind:       "function",
		Name:       name,
		File:       rel,
		Line:       pos.Line,
		Text:       text,
		Receiver:   receiver,
		Params:     params,
		Returns:    returns,
		ParentID:   fileID,
		Cyclomatic: cyclo,
		Cognitive:  cogn,
		CalleeIDs:  []string{},
		CallerIDs:  []string{},
		ChildIDs:   []string{},
	}
	g.AddNode(node)
	g.AddEdge(&graph.Edge{From: fileID, To: nodeID, Kind: "contains"})
	if fileNode, ok := g.GetNode(fileID); ok {
		fileNode.AddChild(nodeID)
	}

	// Collect call expressions
	if d.Body != nil {
		ast.Inspect(d.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			calleeName := callExprName(call)
			if calleeName != "" {
				// Store as unresolved call; resolved later
				node.AddCallee("unresolved:" + calleeName)
			}
			return true
		})
	}
}

func (p *Parser) parseGenDecl(d *ast.GenDecl, pkgName, fileID, rel, srcStr string, g *graph.Graph) {
	for _, spec := range d.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			kind := "type"
			if _, ok := s.Type.(*ast.InterfaceType); ok {
				kind = "interface"
			}
			pos := p.fset.Position(s.Pos())
			nodeID := pkgName + "." + s.Name.Name
			text := extractText(srcStr, p.fset, d.Pos(), d.End())
			node := &graph.Node{
				ID:        nodeID,
				Kind:      kind,
				Name:      s.Name.Name,
				File:      rel,
				Line:      pos.Line,
				Text:      text,
				ParentID:  fileID,
				ChildIDs:  []string{},
				CalleeIDs: []string{},
				CallerIDs: []string{},
			}
			g.AddNode(node)
			g.AddEdge(&graph.Edge{From: fileID, To: nodeID, Kind: "contains"})
			if fileNode, ok := g.GetNode(fileID); ok {
				fileNode.AddChild(nodeID)
			}
		case *ast.ValueSpec:
			kind := "var"
			if d.Tok == token.CONST {
				kind = "const"
			}
			for _, name := range s.Names {
				pos := p.fset.Position(name.Pos())
				nodeID := pkgName + "." + name.Name
				node := &graph.Node{
					ID:        nodeID,
					Kind:      kind,
					Name:      name.Name,
					File:      rel,
					Line:      pos.Line,
					ParentID:  fileID,
					ChildIDs:  []string{},
					CalleeIDs: []string{},
					CallerIDs: []string{},
				}
				g.AddNode(node)
				g.AddEdge(&graph.Edge{From: fileID, To: nodeID, Kind: "contains"})
				if fileNode, ok := g.GetNode(fileID); ok {
					fileNode.AddChild(nodeID)
				}
			}
		}
	}
}

// resolveCallEdges converts "unresolved:funcName" callees to real node IDs.
func (p *Parser) resolveCallEdges(g *graph.Graph) {
	// Build name->ID lookup for functions
	nameToID := make(map[string]string)
	for _, n := range g.Nodes() {
		if n.Kind == "function" {
			// Last segment of dotted name
			parts := strings.Split(n.Name, ".")
			shortName := parts[len(parts)-1]
			nameToID[shortName] = n.ID
			nameToID[n.ID] = n.ID // exact match
			nameToID[n.Name] = n.ID
		}
	}

	for _, n := range g.Nodes() {
		if n.Kind != "function" {
			continue
		}
		resolved := make([]string, 0, len(n.CalleeIDs))
		for _, callee := range n.CalleeIDs {
			if name, ok := strings.CutPrefix(callee, "unresolved:"); ok {
				if id, ok := nameToID[name]; ok {
					resolved = append(resolved, id)
					g.AddEdge(&graph.Edge{From: n.ID, To: id, Kind: "call"})
					// Update callerIDs on target
					if target, ok := g.GetNode(id); ok {
						target.AddCaller(n.ID)
					}
				}
				// Drop unresolved callees
			} else {
				resolved = append(resolved, callee)
			}
		}
		n.CalleeIDs = resolved
	}
}

// resolveImplements detects which types implement which interfaces.
func (p *Parser) resolveImplements(g *graph.Graph) {
	// Collect interfaces and their method sets
	type ifaceInfo struct {
		id      string
		methods map[string]bool
	}
	var ifaces []ifaceInfo

	for _, n := range g.Nodes() {
		if n.Kind != "interface" {
			continue
		}
		// Parse methods from text (simplified: just track by name)
		methods := parseInterfaceMethods(n.Text)
		if len(methods) > 0 {
			ifaces = append(ifaces, ifaceInfo{id: n.ID, methods: methods})
		}
	}

	// For each type, check if it has all methods of an interface
	// Collect type->methods mapping
	typeMethods := make(map[string]map[string]bool)
	for _, n := range g.Nodes() {
		if n.Kind != "function" || n.Receiver == "" {
			continue
		}
		recv := strings.TrimPrefix(n.Receiver, "*")
		// Extract package from node ID
		parts := strings.SplitN(n.ID, ".", 2)
		if len(parts) < 2 {
			continue
		}
		typeID := parts[0] + "." + recv
		if typeMethods[typeID] == nil {
			typeMethods[typeID] = make(map[string]bool)
		}
		typeMethods[typeID][n.Name] = true
	}

	for typeID, methods := range typeMethods {
		for _, iface := range ifaces {
			if len(iface.methods) == 0 {
				continue
			}
			implements := true
			for method := range iface.methods {
				if !methods[method] {
					implements = false
					break
				}
			}
			if implements {
				g.AddEdge(&graph.Edge{From: typeID, To: iface.id, Kind: "implements"})
			}
		}
	}
}

// parseInterfaceMethods extracts method names from interface source text.
func parseInterfaceMethods(text string) map[string]bool {
	methods := make(map[string]bool)
	// Simple heuristic: look for lines like "MethodName(" inside interface {}
	inInterface := false
	for line := range strings.SplitSeq(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "interface {") || strings.Contains(trimmed, "interface{") {
			inInterface = true
			continue
		}
		if inInterface {
			if trimmed == "}" {
				break
			}
			// Method signature: starts with uppercase letter and contains "("
			idx := strings.Index(trimmed, "(")
			if idx > 0 {
				methodName := trimmed[:idx]
				methodName = strings.TrimSpace(methodName)
				if methodName != "" && methodName[0] >= 'A' && methodName[0] <= 'Z' {
					methods[methodName] = true
				}
			}
		}
	}
	return methods
}

// callExprName extracts the callee name from a call expression.
func callExprName(call *ast.CallExpr) string {
	switch f := call.Fun.(type) {
	case *ast.Ident:
		return f.Name
	case *ast.SelectorExpr:
		return f.Sel.Name
	}
	return ""
}

// typeString returns a string representation of a type expression.
func typeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + typeString(t.X)
	case *ast.SelectorExpr:
		return typeString(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + typeString(t.Elt)
	case *ast.MapType:
		return "map[" + typeString(t.Key) + "]" + typeString(t.Value)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.Ellipsis:
		return "..." + typeString(t.Elt)
	}
	return ""
}

// fieldListString converts a field list to a compact string.
func fieldListString(fl *ast.FieldList) string {
	if fl == nil {
		return ""
	}
	var parts []string
	for _, field := range fl.List {
		typStr := typeString(field.Type)
		if len(field.Names) == 0 {
			parts = append(parts, typStr)
		} else {
			for _, name := range field.Names {
				parts = append(parts, name.Name+" "+typStr)
			}
		}
	}
	return strings.Join(parts, ", ")
}

// extractText extracts source text for a given position range.
func extractText(src string, fset *token.FileSet, start, end token.Pos) string {
	f := fset.File(start)
	if f == nil {
		return ""
	}
	startOff := f.Offset(start)
	endOff := f.Offset(end)
	if startOff < 0 || endOff > len(src) || startOff >= endOff {
		return ""
	}
	text := src[startOff:endOff]
	// Truncate very long texts
	const maxLen = 4096
	if len(text) > maxLen {
		text = text[:maxLen] + "..."
	}
	return text
}
