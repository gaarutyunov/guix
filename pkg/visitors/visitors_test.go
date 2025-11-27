package visitors

import (
	"strings"
	"testing"

	"github.com/gaarutyunov/guix/pkg/ast"
)

func TestSemanticAnalyzer_UndefinedVariable(t *testing.T) {
	// Create assignment to undefined variable
	numVal := "0"
	comp := &ast.Component{
		Name: "Test",
		Body: &ast.Body{
			Statements: []*ast.BodyStatement{
				{
					Assignment: &ast.Assignment{
						Left:  "x",
						Op:    "=",
						Right: &ast.Expr{
							Left: &ast.Primary{
								Literal: &ast.Literal{Number: &numVal},
							},
						},
					},
				},
			},
		},
	}

	analyzer := NewSemanticAnalyzer()
	comp.Accept(analyzer)

	if !analyzer.HasErrors() {
		t.Error("Expected error for undefined variable, got none")
	}

	if len(analyzer.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(analyzer.Errors))
	}

	if !strings.Contains(analyzer.Errors[0].Message, "undefined variable: x") {
		t.Errorf("Expected 'undefined variable: x', got '%s'", analyzer.Errors[0].Message)
	}
}

func TestSemanticAnalyzer_ValidAssignment(t *testing.T) {
	// Create valid short declaration
	numVal := "0"
	comp := &ast.Component{
		Name: "Test",
		Body: &ast.Body{
			Statements: []*ast.BodyStatement{
				{
					Assignment: &ast.Assignment{
						Left:  "x",
						Op:    ":=",
						Right: &ast.Expr{
							Left: &ast.Primary{
								Literal: &ast.Literal{Number: &numVal},
							},
						},
					},
				},
				{
					Assignment: &ast.Assignment{
						Left:  "x",
						Op:    "=",
						Right: &ast.Expr{
							Left: &ast.Primary{
								Literal: &ast.Literal{Number: &numVal},
							},
						},
					},
				},
			},
		},
	}

	analyzer := NewSemanticAnalyzer()
	comp.Accept(analyzer)

	if analyzer.HasErrors() {
		t.Errorf("Expected no errors, got %d: %v", len(analyzer.Errors), analyzer.Errors)
	}
}

func TestSemanticAnalyzer_ComponentParams(t *testing.T) {
	// Component parameter should be accessible
	comp := &ast.Component{
		Name: "Test",
		Params: []*ast.Parameter{
			{Name: "x", Type: &ast.Type{Name: "int"}},
		},
		Body: &ast.Body{
			Statements: []*ast.BodyStatement{
				{
					Assignment: &ast.Assignment{
						Left:  "x",
						Op:    "=",
						Right: &ast.Expr{
							Left: &ast.Primary{
								CallOrSel: &ast.CallOrSelect{
									Base: "x",
								},
							},
						},
					},
				},
			},
		},
	}

	analyzer := NewSemanticAnalyzer()
	comp.Accept(analyzer)

	if analyzer.HasErrors() {
		t.Errorf("Expected no errors, got %d: %v", len(analyzer.Errors), analyzer.Errors)
	}
}

func TestSemanticAnalyzer_Scoping(t *testing.T) {
	// Variable declared in if should not leak
	trueVal := "true"
	numVal := "1"
	comp := &ast.Component{
		Name: "Test",
		Body: &ast.Body{
			Statements: []*ast.BodyStatement{
				{
					If: &ast.IfStmt{
						Cond: &ast.Expr{
							Left: &ast.Primary{
								Literal: &ast.Literal{Bool: &trueVal},
							},
						},
						Body: &ast.FuncBody{
							Statements: []*ast.Statement{
								{
									Assignment: &ast.Assignment{
										Left:  "x",
										Op:    ":=",
										Right: &ast.Expr{
											Left: &ast.Primary{
												Literal: &ast.Literal{Number: &numVal},
											},
										},
									},
								},
							},
						},
					},
				},
				{
					// This should error - x not in scope
					Assignment: &ast.Assignment{
						Left:  "x",
						Op:    "=",
						Right: &ast.Expr{
							Left: &ast.Primary{
								Literal: &ast.Literal{Number: &numVal},
							},
						},
					},
				},
			},
		},
	}

	analyzer := NewSemanticAnalyzer()
	comp.Accept(analyzer)

	if !analyzer.HasErrors() {
		t.Error("Expected error for variable out of scope, got none")
	}

	if !strings.Contains(analyzer.Errors[0].Message, "undefined variable: x") {
		t.Errorf("Expected 'undefined variable: x', got '%s'", analyzer.Errors[0].Message)
	}
}

func TestDebugPrinter_SimpleComponent(t *testing.T) {
	// Create a simple component with Component return type
	comp := &ast.Component{
		Name: "Test",
		Params: []*ast.Parameter{
			{Name: "x", Type: &ast.Type{Name: "int"}},
		},
		Results: []*ast.Type{
			{Name: "Component"},
		},
		Body: &ast.Body{
			Children: []*ast.Node{
				{
					Element: &ast.Element{
						Tag: "Div",
						Children: []*ast.Node{
							{Text: &ast.TextNode{Text: "Hello"}},
						},
					},
				},
			},
		},
	}

	printer := NewDebugPrinter()
	comp.Accept(printer)

	output := printer.String()

	// Verify output contains expected elements
	if !strings.Contains(output, "Component: Test") {
		t.Errorf("Expected 'Component: Test' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "Element: Div") {
		t.Errorf("Expected 'Element: Div' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "Text: \"Hello\"") {
		t.Errorf("Expected 'Text: \"Hello\"' in output, got:\n%s", output)
	}
}

func TestDebugPrinter_Assignment(t *testing.T) {
	// Create an assignment
	numVal := "42"
	assignment := &ast.Assignment{
		Left:  "x",
		Op:    ":=",
		Right: &ast.Expr{
			Left: &ast.Primary{
				Literal: &ast.Literal{Number: &numVal},
			},
		},
	}

	printer := NewDebugPrinter()
	assignment.Accept(printer)

	output := printer.String()

	// Verify output
	if !strings.Contains(output, "Assignment: x :=") {
		t.Errorf("Expected 'Assignment: x :=' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "Number: 42") {
		t.Errorf("Expected 'Number: 42' in output, got:\n%s", output)
	}
}

func TestDebugPrinter_Expression(t *testing.T) {
	// Create a binary expression: 2 + 3
	num2 := "2"
	num3 := "3"
	expr := &ast.Expr{
		Left: &ast.Primary{
			Literal: &ast.Literal{Number: &num2},
		},
		BinOps: []*ast.BinaryOp{
			{
				Op: "+",
				Right: &ast.Primary{
					Literal: &ast.Literal{Number: &num3},
				},
			},
		},
	}

	printer := NewDebugPrinter()
	expr.Accept(printer)

	output := printer.String()

	// Verify output
	if !strings.Contains(output, "BinOp: +") {
		t.Errorf("Expected 'BinOp: +' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "Number: 2") {
		t.Errorf("Expected 'Number: 2' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "Number: 3") {
		t.Errorf("Expected 'Number: 3' in output, got:\n%s", output)
	}
}
