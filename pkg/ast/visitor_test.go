package ast

import (
	"testing"
)

// CountingVisitor counts the number of times each node type is visited
type CountingVisitor struct {
	BaseVisitor
	Counts map[string]int
}

func NewCountingVisitor() *CountingVisitor {
	return &CountingVisitor{
		Counts: make(map[string]int),
	}
}

func (v *CountingVisitor) VisitFile(node *File) interface{} {
	v.Counts["File"]++
	// Manually traverse children passing the outer visitor (v)
	for _, imp := range node.Imports {
		imp.Accept(v)
	}
	for _, comp := range node.Components {
		comp.Accept(v)
	}
	return nil
}

func (v *CountingVisitor) VisitComponent(node *Component) interface{} {
	v.Counts["Component"]++
	// Manually traverse children
	if node.Body != nil {
		node.Body.Accept(v)
	}
	return nil
}

func (v *CountingVisitor) VisitBody(node *Body) interface{} {
	// Traverse children (don't count Body itself for simplicity)
	for _, stmt := range node.Statements {
		stmt.Accept(v)
	}
	for _, child := range node.Children {
		child.Accept(v)
	}
	return nil
}

func (v *CountingVisitor) VisitBodyStatement(node *BodyStatement) interface{} {
	// Traverse children
	if node.Assignment != nil {
		node.Assignment.Accept(v)
	}
	return nil
}

func (v *CountingVisitor) VisitNode(node *Node) interface{} {
	// Traverse children
	if node.Element != nil {
		node.Element.Accept(v)
	}
	if node.Text != nil {
		node.Text.Accept(v)
	}
	return nil
}

func (v *CountingVisitor) VisitElement(node *Element) interface{} {
	v.Counts["Element"]++
	// Manually traverse children
	for _, child := range node.Children {
		child.Accept(v)
	}
	return nil
}

func (v *CountingVisitor) VisitExpr(node *Expr) interface{} {
	v.Counts["Expr"]++
	// Manually traverse children
	if node.Left != nil {
		node.Left.Accept(v)
	}
	for _, binOp := range node.BinOps {
		binOp.Accept(v)
	}
	return nil
}

func (v *CountingVisitor) VisitPrimary(node *Primary) interface{} {
	// Traverse children
	if node.Literal != nil {
		node.Literal.Accept(v)
	}
	return nil
}

func (v *CountingVisitor) VisitBinaryOp(node *BinaryOp) interface{} {
	// Traverse children
	if node.Right != nil {
		node.Right.Accept(v)
	}
	return nil
}

func (v *CountingVisitor) VisitLiteral(node *Literal) interface{} {
	v.Counts["Literal"]++
	return nil
}

func (v *CountingVisitor) VisitTextNode(node *TextNode) interface{} {
	v.Counts["TextNode"]++
	return nil
}

func (v *CountingVisitor) VisitIfStmt(node *IfStmt) interface{} {
	v.Counts["IfStmt"]++
	// Manually traverse children
	if node.Cond != nil {
		node.Cond.Accept(v)
	}
	if node.Body != nil {
		node.Body.Accept(v)
	}
	return nil
}

func (v *CountingVisitor) VisitFuncBody(node *FuncBody) interface{} {
	// Traverse children
	for _, stmt := range node.Statements {
		stmt.Accept(v)
	}
	return nil
}

func (v *CountingVisitor) VisitStatement(node *Statement) interface{} {
	// Traverse children
	if node.Assignment != nil {
		node.Assignment.Accept(v)
	}
	return nil
}

func (v *CountingVisitor) VisitAssignment(node *Assignment) interface{} {
	v.Counts["Assignment"]++
	// Manually traverse children
	if node.Right != nil {
		node.Right.Accept(v)
	}
	return nil
}

func TestBaseVisitor_TraversesFile(t *testing.T) {
	// Create a simple AST
	strVal := "test"
	file := &File{
		Package: "main",
		Imports: []*Import{
			{Path: "fmt"},
		},
		Components: []*Component{
			{
				Name: "TestComp",
				Params: []*Parameter{
					{Name: "x", Type: &Type{Name: "int"}},
				},
				Body: &Body{
					Children: []*Node{
						{
							Element: &Element{
								Tag: "Div",
								Children: []*Node{
									{Text: &TextNode{Text: "Hello"}},
								},
							},
						},
					},
				},
			},
		},
	}

	// Create counting visitor
	visitor := NewCountingVisitor()

	// Visit the file
	file.Accept(visitor)

	// Verify counts
	if visitor.Counts["File"] != 1 {
		t.Errorf("Expected 1 File, got %d", visitor.Counts["File"])
	}
	if visitor.Counts["Component"] != 1 {
		t.Errorf("Expected 1 Component, got %d", visitor.Counts["Component"])
	}
	if visitor.Counts["Element"] != 1 {
		t.Errorf("Expected 1 Element, got %d", visitor.Counts["Element"])
	}
	if visitor.Counts["TextNode"] != 1 {
		t.Errorf("Expected 1 TextNode, got %d", visitor.Counts["TextNode"])
	}

	_ = strVal // Use the variable to avoid compiler warning
}

func TestBaseVisitor_TraversesExpressions(t *testing.T) {
	// Create an expression AST
	numVal := "42"
	expr := &Expr{
		Left: &Primary{
			Literal: &Literal{Number: &numVal},
		},
		BinOps: []*BinaryOp{
			{
				Op: "+",
				Right: &Primary{
					Literal: &Literal{Number: &numVal},
				},
			},
		},
	}

	// Create counting visitor
	visitor := NewCountingVisitor()

	// Visit the expression
	expr.Accept(visitor)

	// Verify counts - should see 1 Expr and 2 Literals (left and right of binop)
	if visitor.Counts["Expr"] != 1 {
		t.Errorf("Expected 1 Expr, got %d", visitor.Counts["Expr"])
	}
	if visitor.Counts["Literal"] != 2 {
		t.Errorf("Expected 2 Literals, got %d", visitor.Counts["Literal"])
	}
}

func TestBaseVisitor_TraversesStatements(t *testing.T) {
	// Create a component with statements
	numVal := "0"
	comp := &Component{
		Name: "Test",
		Body: &Body{
			Statements: []*BodyStatement{
				{
					Assignment: &Assignment{
						Left: "x",
						Op:   "=",
						Right: &Expr{
							Left: &Primary{
								Literal: &Literal{Number: &numVal},
							},
						},
					},
				},
			},
		},
	}

	// Create counting visitor
	visitor := NewCountingVisitor()

	// Visit the component
	comp.Accept(visitor)

	// Verify counts
	if visitor.Counts["Component"] != 1 {
		t.Errorf("Expected 1 Component, got %d", visitor.Counts["Component"])
	}
	if visitor.Counts["Assignment"] != 1 {
		t.Errorf("Expected 1 Assignment, got %d", visitor.Counts["Assignment"])
	}
	if visitor.Counts["Expr"] != 1 {
		t.Errorf("Expected 1 Expr, got %d", visitor.Counts["Expr"])
	}
	if visitor.Counts["Literal"] != 1 {
		t.Errorf("Expected 1 Literal, got %d", visitor.Counts["Literal"])
	}
}

func TestBaseVisitor_TraversesIfStatements(t *testing.T) {
	// Create an if statement
	trueVal := "true"
	numVal := "1"
	ifStmt := &IfStmt{
		Cond: &Expr{
			Left: &Primary{
				Literal: &Literal{Bool: &trueVal},
			},
		},
		Body: &FuncBody{
			Statements: []*Statement{
				{
					Assignment: &Assignment{
						Left: "x",
						Op:   "=",
						Right: &Expr{
							Left: &Primary{
								Literal: &Literal{Number: &numVal},
							},
						},
					},
				},
			},
		},
	}

	// Create counting visitor
	visitor := NewCountingVisitor()

	// Visit the if statement
	ifStmt.Accept(visitor)

	// Verify counts
	if visitor.Counts["IfStmt"] != 1 {
		t.Errorf("Expected 1 IfStmt, got %d", visitor.Counts["IfStmt"])
	}
	if visitor.Counts["Assignment"] != 1 {
		t.Errorf("Expected 1 Assignment, got %d", visitor.Counts["Assignment"])
	}
	// Should have 2 Exprs: one for condition, one for assignment right side
	if visitor.Counts["Expr"] != 2 {
		t.Errorf("Expected 2 Exprs, got %d", visitor.Counts["Expr"])
	}
	// Should have 2 Literals: one in condition, one in assignment
	if visitor.Counts["Literal"] != 2 {
		t.Errorf("Expected 2 Literals, got %d", visitor.Counts["Literal"])
	}
}

// TransformingVisitor demonstrates a visitor that transforms the AST
type TransformingVisitor struct {
	BaseVisitor
}

func (v *TransformingVisitor) VisitTextNode(node *TextNode) interface{} {
	// Transform: reverse the text
	runes := []rune(node.Text)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	node.Text = string(runes)
	return nil
}

func TestTransformingVisitor(t *testing.T) {
	// Create a text node
	textNode := &TextNode{Text: "Hello"}

	// Create transforming visitor
	visitor := &TransformingVisitor{}

	// Visit the node
	textNode.Accept(visitor)

	// Verify transformation
	if textNode.Text != "olleH" {
		t.Errorf("Expected 'olleH', got '%s'", textNode.Text)
	}
}
