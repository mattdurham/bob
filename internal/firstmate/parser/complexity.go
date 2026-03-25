package parser

import "go/ast"

// CyclomaticComplexity computes cyclomatic complexity for a function body.
// Complexity = 1 + count of decision points (if, for, range, case, select-case, &&, ||).
func CyclomaticComplexity(body *ast.BlockStmt) int {
	if body == nil {
		return 1
	}
	counter := &cyclomaticVisitor{}
	ast.Inspect(body, counter.visit)
	return 1 + counter.count
}

type cyclomaticVisitor struct {
	count int
}

func (v *cyclomaticVisitor) visit(n ast.Node) bool {
	switch x := n.(type) {
	case *ast.IfStmt:
		v.count++
	case *ast.ForStmt:
		v.count++
	case *ast.RangeStmt:
		v.count++
	case *ast.CaseClause:
		if x.List != nil { // non-default case
			v.count++
		}
	case *ast.CommClause:
		if x.Comm != nil { // non-default select-case
			v.count++
		}
	case *ast.BinaryExpr:
		if x.Op.String() == "&&" || x.Op.String() == "||" {
			v.count++
		}
	}
	return true
}

// CognitiveComplexity computes cognitive complexity for a function body.
// Weighted by nesting depth.
func CognitiveComplexity(body *ast.BlockStmt) int {
	if body == nil {
		return 0
	}
	visitor := &cognitiveVisitor{}
	visitor.walkBlock(body, 0)
	return visitor.score
}

type cognitiveVisitor struct {
	score int
}

func (v *cognitiveVisitor) walkBlock(block *ast.BlockStmt, depth int) {
	if block == nil {
		return
	}
	for _, stmt := range block.List {
		v.walkStmt(stmt, depth)
	}
}

func (v *cognitiveVisitor) walkStmt(stmt ast.Stmt, depth int) {
	switch s := stmt.(type) {
	case *ast.IfStmt:
		v.score += 1 + depth
		v.walkBlock(s.Body, depth+1)
		if s.Else != nil {
			v.score++
			switch e := s.Else.(type) {
			case *ast.BlockStmt:
				v.walkBlock(e, depth+1)
			case *ast.IfStmt:
				v.walkStmt(e, depth+1)
			}
		}
	case *ast.ForStmt:
		v.score += 1 + depth
		v.walkBlock(s.Body, depth+1)
	case *ast.RangeStmt:
		v.score += 1 + depth
		v.walkBlock(s.Body, depth+1)
	case *ast.SwitchStmt:
		v.score += 1 + depth
		v.walkBlock(s.Body, depth+1)
	case *ast.TypeSwitchStmt:
		v.score += 1 + depth
		v.walkBlock(s.Body, depth+1)
	case *ast.SelectStmt:
		v.score += 1 + depth
		v.walkBlock(s.Body, depth+1)
	case *ast.CaseClause:
		v.walkBlock(&ast.BlockStmt{List: s.Body}, depth)
	case *ast.CommClause:
		v.walkBlock(&ast.BlockStmt{List: s.Body}, depth)
	}
}
