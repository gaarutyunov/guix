// Package visitors provides AST visitor implementations for various compiler passes
package visitors

import (
	"fmt"
	"strings"

	"github.com/gaarutyunov/guix/pkg/ast"
)

// DebugPrinter prints a formatted representation of the AST for debugging
type DebugPrinter struct {
	ast.BaseVisitor

	// Output buffer
	output strings.Builder

	// Current indentation level
	indent int
}

// NewDebugPrinter creates a new debug printer
func NewDebugPrinter() *DebugPrinter {
	return &DebugPrinter{
		indent: 0,
	}
}

// String returns the formatted output
func (d *DebugPrinter) String() string {
	return d.output.String()
}

// print writes indented output
func (d *DebugPrinter) print(format string, args ...interface{}) {
	d.output.WriteString(strings.Repeat("  ", d.indent))
	d.output.WriteString(fmt.Sprintf(format, args...))
	d.output.WriteString("\n")
}

// VisitFile prints a file node
func (d *DebugPrinter) VisitFile(node *ast.File) interface{} {
	d.print("File: package %s", node.Package)
	d.indent++

	if len(node.Imports) > 0 {
		d.print("Imports:")
		d.indent++
		for _, imp := range node.Imports {
			imp.Accept(d)
		}
		d.indent--
	}

	if len(node.Types) > 0 {
		d.print("Types:")
		d.indent++
		for _, typeDef := range node.Types {
			typeDef.Accept(d)
		}
		d.indent--
	}

	if len(node.Components) > 0 {
		d.print("Components:")
		d.indent++
		for _, comp := range node.Components {
			comp.Accept(d)
		}
		d.indent--
	}

	d.indent--
	return nil
}

// VisitImport prints an import node
func (d *DebugPrinter) VisitImport(node *ast.Import) interface{} {
	if len(node.Paths) == 1 {
		d.print("Import: %s", node.Paths[0])
	} else {
		d.print("Import: %v", node.Paths)
	}
	return nil
}

// VisitTypeDef prints a type definition
func (d *DebugPrinter) VisitTypeDef(node *ast.TypeDef) interface{} {
	d.print("TypeDef: %s", node.Name)
	d.indent++
	if node.Struct != nil {
		node.Struct.Accept(d)
	}
	d.indent--
	return nil
}

// VisitStructType prints a struct type
func (d *DebugPrinter) VisitStructType(node *ast.StructType) interface{} {
	d.print("Struct:")
	d.indent++
	for _, field := range node.Fields {
		field.Accept(d)
	}
	d.indent--
	return nil
}

// VisitStructField prints a struct field
func (d *DebugPrinter) VisitStructField(node *ast.StructField) interface{} {
	d.print("Field: %s %s", node.Name, d.typeString(node.Type))
	return nil
}

// VisitComponent prints a component
func (d *DebugPrinter) VisitComponent(node *ast.Component) interface{} {
	params := make([]string, len(node.Params))
	for i, p := range node.Params {
		params[i] = fmt.Sprintf("%s %s", p.Name, d.typeString(p.Type))
	}

	results := make([]string, len(node.Results))
	for i, r := range node.Results {
		results[i] = d.typeString(r)
	}

	if len(results) > 0 {
		d.print("Component: %s(%s) (%s)", node.Name, strings.Join(params, ", "), strings.Join(results, ", "))
	} else {
		d.print("Function: %s(%s)", node.Name, strings.Join(params, ", "))
	}

	d.indent++
	if node.Body != nil {
		node.Body.Accept(d)
	}
	d.indent--

	return nil
}

// VisitBody prints a component body
func (d *DebugPrinter) VisitBody(node *ast.Body) interface{} {
	if len(node.VarDecls) > 0 {
		d.print("Variables:")
		d.indent++
		for _, varDecl := range node.VarDecls {
			varDecl.Accept(d)
		}
		d.indent--
	}

	if len(node.Statements) > 0 {
		d.print("Statements:")
		d.indent++
		for _, stmt := range node.Statements {
			stmt.Accept(d)
		}
		d.indent--
	}

	if len(node.Children) > 0 {
		d.print("Children:")
		d.indent++
		for _, child := range node.Children {
			child.Accept(d)
		}
		d.indent--
	}

	return nil
}

// VisitVarDecl prints a variable declaration
func (d *DebugPrinter) VisitVarDecl(node *ast.VarDecl) interface{} {
	d.print("VarDecl: %s := ...", strings.Join(node.Names, ", "))
	d.indent++
	for _, val := range node.Values {
		val.Accept(d)
	}
	d.indent--
	return nil
}

// VisitAssignment prints an assignment
func (d *DebugPrinter) VisitAssignment(node *ast.Assignment) interface{} {
	selector := node.Left
	if len(node.LeftSelector) > 0 {
		selector = selector + "." + strings.Join(node.LeftSelector, ".")
	}
	d.print("Assignment: %s %s ...", selector, node.Op)
	d.indent++
	if node.Right != nil {
		node.Right.Accept(d)
	}
	d.indent--
	return nil
}

// VisitReturn prints a return statement
func (d *DebugPrinter) VisitReturn(node *ast.Return) interface{} {
	d.print("Return:")
	d.indent++
	for _, val := range node.Values {
		val.Accept(d)
	}
	d.indent--
	return nil
}

// VisitIfStmt prints an if statement
func (d *DebugPrinter) VisitIfStmt(node *ast.IfStmt) interface{} {
	d.print("If:")
	d.indent++
	d.print("Condition:")
	d.indent++
	if node.Cond != nil {
		node.Cond.Accept(d)
	}
	d.indent--
	d.print("Then:")
	d.indent++
	if node.Body != nil {
		node.Body.Accept(d)
	}
	d.indent--
	if node.Else != nil {
		node.Else.Accept(d)
	}
	d.indent--
	return nil
}

// VisitElse prints an else clause
func (d *DebugPrinter) VisitElse(node *ast.Else) interface{} {
	if node.IfStmt != nil {
		d.print("Else If:")
		d.indent++
		node.IfStmt.Accept(d)
		d.indent--
	} else if node.Body != nil {
		d.print("Else:")
		d.indent++
		node.Body.Accept(d)
		d.indent--
	}
	return nil
}

// VisitForLoop prints a for loop
func (d *DebugPrinter) VisitForLoop(node *ast.ForLoop) interface{} {
	if node.Key != "" {
		d.print("For: %s, %s in ...", node.Key, node.Val)
	} else {
		d.print("For: %s in ...", node.Val)
	}
	d.indent++
	if node.Range != nil {
		d.print("Range:")
		d.indent++
		node.Range.Accept(d)
		d.indent--
	}
	if node.Body != nil {
		d.print("Body:")
		d.indent++
		node.Body.Accept(d)
		d.indent--
	}
	d.indent--
	return nil
}

// VisitElement prints an element
func (d *DebugPrinter) VisitElement(node *ast.Element) interface{} {
	d.print("Element: %s", node.Tag)
	d.indent++
	if len(node.Props) > 0 {
		d.print("Props:")
		d.indent++
		for _, prop := range node.Props {
			prop.Accept(d)
		}
		d.indent--
	}
	if len(node.Children) > 0 {
		d.print("Children:")
		d.indent++
		for _, child := range node.Children {
			child.Accept(d)
		}
		d.indent--
	}
	d.indent--
	return nil
}

// VisitProp prints a property
func (d *DebugPrinter) VisitProp(node *ast.Prop) interface{} {
	d.print("Prop: %s(...)", node.Name)
	d.indent++
	for _, arg := range node.Args {
		if arg != nil {
			arg.Accept(d)
		}
	}
	d.indent--
	return nil
}

// VisitNode prints a node
func (d *DebugPrinter) VisitNode(node *ast.Node) interface{} {
	if node.Text != nil {
		node.Text.Accept(d)
	}
	if node.Template != nil {
		node.Template.Accept(d)
	}
	if node.IfExpr != nil {
		node.IfExpr.Accept(d)
	}
	if node.ForLoop != nil {
		node.ForLoop.Accept(d)
	}
	if node.ChannelRecv != nil {
		node.ChannelRecv.Accept(d)
	}
	if node.Element != nil {
		node.Element.Accept(d)
	}
	if node.ExprStmt != nil {
		node.ExprStmt.Accept(d)
	}
	return nil
}

// VisitTextNode prints a text node
func (d *DebugPrinter) VisitTextNode(node *ast.TextNode) interface{} {
	d.print("Text: %q", node.Text)
	return nil
}

// VisitTemplate prints a template
func (d *DebugPrinter) VisitTemplate(node *ast.Template) interface{} {
	d.print("Template:")
	d.indent++
	for _, frag := range node.Fragments {
		frag.Accept(d)
	}
	d.indent--
	return nil
}

// VisitFragment prints a template fragment
func (d *DebugPrinter) VisitFragment(node *ast.Fragment) interface{} {
	if node.Text != "" {
		d.print("Text: %q", node.Text)
	}
	if node.Expr != nil {
		d.print("Expr:")
		d.indent++
		node.Expr.Accept(d)
		d.indent--
	}
	return nil
}

// VisitExpr prints an expression
func (d *DebugPrinter) VisitExpr(node *ast.Expr) interface{} {
	d.print("Expr:")
	d.indent++
	if node.Left != nil {
		node.Left.Accept(d)
	}
	for _, binOp := range node.BinOps {
		binOp.Accept(d)
	}
	d.indent--
	return nil
}

// VisitBinaryOp prints a binary operation
func (d *DebugPrinter) VisitBinaryOp(node *ast.BinaryOp) interface{} {
	d.print("BinOp: %s", node.Op)
	d.indent++
	if node.Right != nil {
		node.Right.Accept(d)
	}
	d.indent--
	return nil
}

// VisitPrimary prints a primary expression
func (d *DebugPrinter) VisitPrimary(node *ast.Primary) interface{} {
	if node.Ident != "" {
		d.print("Ident: %s", node.Ident)
	}
	if node.Unary != nil {
		node.Unary.Accept(d)
	}
	if node.Literal != nil {
		node.Literal.Accept(d)
	}
	if node.CompositeLit != nil {
		node.CompositeLit.Accept(d)
	}
	if node.MakeCall != nil {
		node.MakeCall.Accept(d)
	}
	if node.CallOrSel != nil {
		node.CallOrSel.Accept(d)
	}
	if node.FuncLit != nil {
		node.FuncLit.Accept(d)
	}
	if node.ChannelOp != nil {
		node.ChannelOp.Accept(d)
	}
	if node.Paren != nil {
		d.print("Paren:")
		d.indent++
		node.Paren.Accept(d)
		d.indent--
	}
	return nil
}

// VisitUnaryExpr prints a unary expression
func (d *DebugPrinter) VisitUnaryExpr(node *ast.UnaryExpr) interface{} {
	d.print("Unary: %s", node.Op)
	d.indent++
	if node.Right != nil {
		node.Right.Accept(d)
	}
	d.indent--
	return nil
}

// VisitLiteral prints a literal
func (d *DebugPrinter) VisitLiteral(node *ast.Literal) interface{} {
	if node.String != nil {
		d.print("String: %s", *node.String)
	}
	if node.Number != nil {
		d.print("Number: %s", *node.Number)
	}
	if node.Bool != nil {
		d.print("Bool: %s", *node.Bool)
	}
	return nil
}

// VisitCallOrSelect prints a call or selector
func (d *DebugPrinter) VisitCallOrSelect(node *ast.CallOrSelect) interface{} {
	selector := node.Base
	if len(node.Fields) > 0 {
		selector = selector + "." + strings.Join(node.Fields, ".")
	}
	if len(node.Args) > 0 {
		d.print("Call: %s(...)", selector)
		d.indent++
		for i, arg := range node.Args {
			d.print("Arg %d:", i)
			d.indent++
			arg.Accept(d)
			d.indent--
		}
		d.indent--
	} else {
		d.print("Selector: %s", selector)
	}
	return nil
}

// VisitMakeCall prints a make call
func (d *DebugPrinter) VisitMakeCall(node *ast.MakeCall) interface{} {
	if node.ChanType != nil {
		d.print("Make: chan %s", d.typeString(node.ChanType))
		if node.ChanSize != nil {
			d.indent++
			d.print("Size:")
			d.indent++
			node.ChanSize.Accept(d)
			d.indent--
			d.indent--
		}
	} else if node.SliceType != nil {
		d.print("Make: []%s", d.typeString(node.SliceType))
		if node.SliceLen != nil {
			d.indent++
			d.print("Len:")
			d.indent++
			node.SliceLen.Accept(d)
			d.indent--
			d.indent--
		}
		if node.SliceCap != nil {
			d.indent++
			d.print("Cap:")
			d.indent++
			node.SliceCap.Accept(d)
			d.indent--
			d.indent--
		}
	}
	return nil
}

// VisitFuncLit prints a function literal
func (d *DebugPrinter) VisitFuncLit(node *ast.FuncLit) interface{} {
	params := make([]string, len(node.Params))
	for i, p := range node.Params {
		params[i] = fmt.Sprintf("%s %s", p.Name, d.typeString(p.Type))
	}
	d.print("FuncLit: func(%s)", strings.Join(params, ", "))
	d.indent++
	if node.Body != nil {
		node.Body.Accept(d)
	}
	d.indent--
	return nil
}

// VisitFuncBody prints a function body
func (d *DebugPrinter) VisitFuncBody(node *ast.FuncBody) interface{} {
	for _, stmt := range node.Statements {
		stmt.Accept(d)
	}
	return nil
}

// VisitStatement prints a statement
func (d *DebugPrinter) VisitStatement(node *ast.Statement) interface{} {
	if node.VarDecl != nil {
		node.VarDecl.Accept(d)
	}
	if node.Assignment != nil {
		node.Assignment.Accept(d)
	}
	if node.Return != nil {
		node.Return.Accept(d)
	}
	if node.If != nil {
		node.If.Accept(d)
	}
	if node.For != nil {
		node.For.Accept(d)
	}
	if node.Expr != nil {
		node.Expr.Accept(d)
	}
	return nil
}

// VisitCompositeLit prints a composite literal
func (d *DebugPrinter) VisitCompositeLit(node *ast.CompositeLit) interface{} {
	d.print("CompositeLit: %s{...}", node.Type)
	d.indent++
	for _, elem := range node.Elements {
		elem.Accept(d)
	}
	d.indent--
	return nil
}

// VisitKeyValue prints a key-value pair
func (d *DebugPrinter) VisitKeyValue(node *ast.KeyValue) interface{} {
	d.print("Field: %s", node.Key)
	d.indent++
	if node.Value != nil {
		node.Value.Accept(d)
	}
	d.indent--
	return nil
}

// VisitChannelOp prints a channel operation
func (d *DebugPrinter) VisitChannelOp(node *ast.ChannelOp) interface{} {
	d.print("ChannelOp: %s%s", node.Op, node.Channel)
	return nil
}

// VisitChannelRecv prints a channel receive
func (d *DebugPrinter) VisitChannelRecv(node *ast.ChannelRecv) interface{} {
	d.print("ChannelRecv: <-%s", node.Channel)
	return nil
}

// VisitIfExpr prints an if expression
func (d *DebugPrinter) VisitIfExpr(node *ast.IfExpr) interface{} {
	d.print("IfExpr:")
	d.indent++
	d.print("Cond:")
	d.indent++
	if node.Cond != nil {
		node.Cond.Accept(d)
	}
	d.indent--
	d.print("True:")
	d.indent++
	if node.TrueBody != nil {
		node.TrueBody.Accept(d)
	}
	d.indent--
	if node.FalseBody != nil {
		d.print("False:")
		d.indent++
		node.FalseBody.Accept(d)
		d.indent--
	}
	d.indent--
	return nil
}

// typeString formats a type as a string
func (d *DebugPrinter) typeString(t *ast.Type) string {
	if t == nil {
		return "nil"
	}
	prefix := ""
	if t.IsChannel {
		prefix = "<-"
	}
	if t.IsChan {
		prefix = "chan "
	}
	if t.IsPointer {
		prefix += "*"
	}
	name := t.Name
	if t.Generic != nil {
		name = fmt.Sprintf("%s[%s]", name, d.typeString(t.Generic))
	}
	return prefix + name
}

// Implement remaining visitor interface methods
func (d *DebugPrinter) VisitBodyStatement(node *ast.BodyStatement) interface{} {
	return d.BaseVisitor.VisitBodyStatement(node)
}

func (d *DebugPrinter) VisitExprStmt(node *ast.ExprStmt) interface{} {
	return d.BaseVisitor.VisitExprStmt(node)
}

func (d *DebugPrinter) VisitParameter(node *ast.Parameter) interface{} {
	return d.BaseVisitor.VisitParameter(node)
}

func (d *DebugPrinter) VisitType(node *ast.Type) interface{} {
	return d.BaseVisitor.VisitType(node)
}

func (d *DebugPrinter) VisitSelector(node *ast.Selector) interface{} {
	return d.BaseVisitor.VisitSelector(node)
}

func (d *DebugPrinter) VisitCall(node *ast.Call) interface{} {
	return d.BaseVisitor.VisitCall(node)
}
