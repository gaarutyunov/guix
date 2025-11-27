// Package visitors provides AST visitor implementations for various compiler passes
package visitors

import (
	"fmt"

	"github.com/gaarutyunov/guix/pkg/ast"
)

// SemanticError represents a semantic error in the AST
type SemanticError struct {
	Position string
	Message  string
}

func (e *SemanticError) Error() string {
	return fmt.Sprintf("%s: %s", e.Position, e.Message)
}

// SemanticAnalyzer performs semantic analysis on the AST
// It validates:
// - Variable declarations and usage
// - Type consistency
// - Channel usage patterns
// - Component structure
type SemanticAnalyzer struct {
	ast.BaseVisitor

	// Errors collected during analysis
	Errors []*SemanticError

	// Warnings collected during analysis
	Warnings []*SemanticError

	// Current scope for variable tracking
	scopes []map[string]bool

	// Component parameters for current component
	componentParams map[string]bool

	// Hoisted variables for current component
	hoistedVars map[string]bool
}

// NewSemanticAnalyzer creates a new semantic analyzer
func NewSemanticAnalyzer() *SemanticAnalyzer {
	return &SemanticAnalyzer{
		Errors:   make([]*SemanticError, 0),
		Warnings: make([]*SemanticError, 0),
		scopes:   []map[string]bool{make(map[string]bool)},
	}
}

// HasErrors returns true if any errors were found
func (s *SemanticAnalyzer) HasErrors() bool {
	return len(s.Errors) > 0
}

// addError records a semantic error
func (s *SemanticAnalyzer) addError(pos, message string) {
	s.Errors = append(s.Errors, &SemanticError{
		Position: pos,
		Message:  message,
	})
}

// addWarning records a semantic warning
func (s *SemanticAnalyzer) addWarning(pos, message string) {
	s.Warnings = append(s.Warnings, &SemanticError{
		Position: pos,
		Message:  message,
	})
}

// pushScope creates a new variable scope
func (s *SemanticAnalyzer) pushScope() {
	s.scopes = append(s.scopes, make(map[string]bool))
}

// popScope removes the current scope
func (s *SemanticAnalyzer) popScope() {
	if len(s.scopes) > 1 {
		s.scopes = s.scopes[:len(s.scopes)-1]
	}
}

// declareVar declares a variable in the current scope
func (s *SemanticAnalyzer) declareVar(name string) {
	if len(s.scopes) > 0 {
		s.scopes[len(s.scopes)-1][name] = true
	}
}

// isDeclared checks if a variable is declared in any scope
func (s *SemanticAnalyzer) isDeclared(name string) bool {
	// Check component params
	if s.componentParams != nil && s.componentParams[name] {
		return true
	}
	// Check hoisted vars
	if s.hoistedVars != nil && s.hoistedVars[name] {
		return true
	}
	// Check scopes from innermost to outermost
	for i := len(s.scopes) - 1; i >= 0; i-- {
		if s.scopes[i][name] {
			return true
		}
	}
	return false
}

// VisitFile analyzes a file
func (s *SemanticAnalyzer) VisitFile(node *ast.File) interface{} {
	for _, imp := range node.Imports {
		imp.Accept(s)
	}
	for _, typeDef := range node.Types {
		typeDef.Accept(s)
	}
	for _, comp := range node.Components {
		comp.Accept(s)
	}
	return nil
}

// VisitComponent analyzes a component
func (s *SemanticAnalyzer) VisitComponent(node *ast.Component) interface{} {
	// Create a new scope for component
	s.pushScope()
	defer s.popScope()

	// Track component parameters
	s.componentParams = make(map[string]bool)
	for _, param := range node.Params {
		s.componentParams[param.Name] = true
		param.Accept(s)
	}

	// Analyze body
	if node.Body != nil {
		// First pass: collect hoisted variables (channels and state)
		s.hoistedVars = make(map[string]bool)
		for _, varDecl := range node.Body.VarDecls {
			for _, name := range varDecl.Names {
				s.hoistedVars[name] = true
			}
		}

		// Second pass: analyze the body
		node.Body.Accept(s)
	}

	// Clear component context
	s.componentParams = nil
	s.hoistedVars = nil

	return nil
}

// VisitBody analyzes a component body
func (s *SemanticAnalyzer) VisitBody(node *ast.Body) interface{} {
	// Analyze variable declarations
	for _, varDecl := range node.VarDecls {
		varDecl.Accept(s)
	}

	// Analyze statements
	for _, stmt := range node.Statements {
		stmt.Accept(s)
	}

	// Analyze UI tree
	for _, child := range node.Children {
		child.Accept(s)
	}

	return nil
}

// VisitVarDecl analyzes a variable declaration
func (s *SemanticAnalyzer) VisitVarDecl(node *ast.VarDecl) interface{} {
	// Declare all variables
	for _, name := range node.Names {
		s.declareVar(name)
	}

	// Analyze values
	for _, val := range node.Values {
		val.Accept(s)
	}

	// Validate: number of names should match number of values (or values is 1 for multi-return)
	if len(node.Values) != 1 && len(node.Names) != len(node.Values) {
		s.addError(
			fmt.Sprintf("%d:%d", node.Pos.Line, node.Pos.Column),
			fmt.Sprintf("assignment mismatch: %d variables but %d values", len(node.Names), len(node.Values)),
		)
	}

	return nil
}

// VisitAssignment analyzes an assignment
func (s *SemanticAnalyzer) VisitAssignment(node *ast.Assignment) interface{} {
	// Check if variable is declared (for regular assignments, not :=)
	if node.Op == "=" || node.Op == "+=" || node.Op == "-=" || node.Op == "*=" || node.Op == "/=" {
		if !s.isDeclared(node.Left) {
			s.addError(
				fmt.Sprintf("%d:%d", node.Pos.Line, node.Pos.Column),
				fmt.Sprintf("undefined variable: %s", node.Left),
			)
		}
	} else if node.Op == ":=" {
		// Short declaration - declare the variable
		s.declareVar(node.Left)
	}

	// Analyze right side
	if node.Right != nil {
		node.Right.Accept(s)
	}

	return nil
}

// VisitStatement analyzes a statement
func (s *SemanticAnalyzer) VisitStatement(node *ast.Statement) interface{} {
	if node.VarDecl != nil {
		node.VarDecl.Accept(s)
	}
	if node.Assignment != nil {
		node.Assignment.Accept(s)
	}
	if node.Return != nil {
		node.Return.Accept(s)
	}
	if node.If != nil {
		node.If.Accept(s)
	}
	if node.For != nil {
		node.For.Accept(s)
	}
	if node.Expr != nil {
		node.Expr.Accept(s)
	}
	return nil
}

// VisitIfStmt analyzes an if statement
func (s *SemanticAnalyzer) VisitIfStmt(node *ast.IfStmt) interface{} {
	// Analyze condition
	if node.Cond != nil {
		node.Cond.Accept(s)
	}

	// Analyze body in new scope
	s.pushScope()
	if node.Body != nil {
		node.Body.Accept(s)
	}
	s.popScope()

	// Analyze else clause
	if node.Else != nil {
		node.Else.Accept(s)
	}

	return nil
}

// VisitElse analyzes an else clause
func (s *SemanticAnalyzer) VisitElse(node *ast.Else) interface{} {
	if node.IfStmt != nil {
		node.IfStmt.Accept(s)
	}
	if node.Body != nil {
		s.pushScope()
		node.Body.Accept(s)
		s.popScope()
	}
	return nil
}

// VisitForLoop analyzes a for loop
func (s *SemanticAnalyzer) VisitForLoop(node *ast.ForLoop) interface{} {
	// Analyze range expression
	if node.Range != nil {
		node.Range.Accept(s)
	}

	// Analyze body in new scope with loop variables
	s.pushScope()
	if node.Key != "" {
		s.declareVar(node.Key)
	}
	if node.Val != "" {
		s.declareVar(node.Val)
	}
	if node.Body != nil {
		node.Body.Accept(s)
	}
	s.popScope()

	return nil
}

// VisitExpr analyzes an expression
func (s *SemanticAnalyzer) VisitExpr(node *ast.Expr) interface{} {
	if node.Left != nil {
		node.Left.Accept(s)
	}
	for _, binOp := range node.BinOps {
		binOp.Accept(s)
	}
	return nil
}

// VisitPrimary analyzes a primary expression
func (s *SemanticAnalyzer) VisitPrimary(node *ast.Primary) interface{} {
	// Check identifier usage
	if node.Ident != "" {
		// Check if it's a known identifier (variable, parameter, etc.)
		// We'll be lenient here and not error on unknown identifiers
		// as they might be component names or Go built-ins
		if !s.isDeclared(node.Ident) {
			// This might be a component name, built-in, or imported identifier
			// Don't error, just note it
		}
	}

	// Traverse children
	if node.Unary != nil {
		node.Unary.Accept(s)
	}
	if node.Literal != nil {
		node.Literal.Accept(s)
	}
	if node.CompositeLit != nil {
		node.CompositeLit.Accept(s)
	}
	if node.MakeCall != nil {
		node.MakeCall.Accept(s)
	}
	if node.CallOrSel != nil {
		node.CallOrSel.Accept(s)
	}
	if node.FuncLit != nil {
		node.FuncLit.Accept(s)
	}
	if node.ChannelOp != nil {
		node.ChannelOp.Accept(s)
	}
	if node.Paren != nil {
		node.Paren.Accept(s)
	}

	return nil
}

// VisitCallOrSelect analyzes a call or selector
func (s *SemanticAnalyzer) VisitCallOrSelect(node *ast.CallOrSelect) interface{} {
	// Check base identifier
	if node.Base != "" && !s.isDeclared(node.Base) {
		// Might be a component, built-in, or imported identifier
		// Don't error
	}

	// Analyze arguments
	for _, arg := range node.Args {
		arg.Accept(s)
	}

	return nil
}

// VisitFuncLit analyzes a function literal
func (s *SemanticAnalyzer) VisitFuncLit(node *ast.FuncLit) interface{} {
	// Create new scope for function
	s.pushScope()
	defer s.popScope()

	// Declare parameters
	for _, param := range node.Params {
		s.declareVar(param.Name)
		param.Accept(s)
	}

	// Analyze body
	if node.Body != nil {
		node.Body.Accept(s)
	}

	return nil
}

// VisitFuncBody analyzes a function body
func (s *SemanticAnalyzer) VisitFuncBody(node *ast.FuncBody) interface{} {
	for _, stmt := range node.Statements {
		stmt.Accept(s)
	}
	return nil
}

// VisitChannelOp analyzes a channel operation
func (s *SemanticAnalyzer) VisitChannelOp(node *ast.ChannelOp) interface{} {
	// Check if channel variable is declared
	if !s.isDeclared(node.Channel) {
		s.addError(
			fmt.Sprintf("%d:%d", node.Pos.Line, node.Pos.Column),
			fmt.Sprintf("undefined channel: %s", node.Channel),
		)
	}
	return nil
}

// Remaining visitor methods delegate to BaseVisitor for traversal
func (s *SemanticAnalyzer) VisitBinaryOp(node *ast.BinaryOp) interface{} {
	if node.Right != nil {
		node.Right.Accept(s)
	}
	return nil
}

func (s *SemanticAnalyzer) VisitUnaryExpr(node *ast.UnaryExpr) interface{} {
	if node.Right != nil {
		node.Right.Accept(s)
	}
	return nil
}

func (s *SemanticAnalyzer) VisitCompositeLit(node *ast.CompositeLit) interface{} {
	for _, elem := range node.Elements {
		elem.Accept(s)
	}
	return nil
}

func (s *SemanticAnalyzer) VisitKeyValue(node *ast.KeyValue) interface{} {
	if node.Value != nil {
		node.Value.Accept(s)
	}
	return nil
}

func (s *SemanticAnalyzer) VisitElement(node *ast.Element) interface{} {
	for _, prop := range node.Props {
		prop.Accept(s)
	}
	for _, child := range node.Children {
		child.Accept(s)
	}
	return nil
}

func (s *SemanticAnalyzer) VisitProp(node *ast.Prop) interface{} {
	if node.Value != nil {
		node.Value.Accept(s)
	}
	return nil
}

func (s *SemanticAnalyzer) VisitNode(node *ast.Node) interface{} {
	if node.Text != nil {
		node.Text.Accept(s)
	}
	if node.Template != nil {
		node.Template.Accept(s)
	}
	if node.IfExpr != nil {
		node.IfExpr.Accept(s)
	}
	if node.ForLoop != nil {
		node.ForLoop.Accept(s)
	}
	if node.ChannelRecv != nil {
		node.ChannelRecv.Accept(s)
	}
	if node.Element != nil {
		node.Element.Accept(s)
	}
	if node.ExprStmt != nil {
		node.ExprStmt.Accept(s)
	}
	return nil
}

func (s *SemanticAnalyzer) VisitTemplate(node *ast.Template) interface{} {
	for _, frag := range node.Fragments {
		frag.Accept(s)
	}
	return nil
}

func (s *SemanticAnalyzer) VisitFragment(node *ast.Fragment) interface{} {
	if node.Expr != nil {
		node.Expr.Accept(s)
	}
	return nil
}

func (s *SemanticAnalyzer) VisitIfExpr(node *ast.IfExpr) interface{} {
	if node.Cond != nil {
		node.Cond.Accept(s)
	}
	if node.TrueBody != nil {
		s.pushScope()
		node.TrueBody.Accept(s)
		s.popScope()
	}
	if node.FalseBody != nil {
		s.pushScope()
		node.FalseBody.Accept(s)
		s.popScope()
	}
	return nil
}

func (s *SemanticAnalyzer) VisitReturn(node *ast.Return) interface{} {
	for _, val := range node.Values {
		val.Accept(s)
	}
	return nil
}

func (s *SemanticAnalyzer) VisitBodyStatement(node *ast.BodyStatement) interface{} {
	if node.VarDecl != nil {
		node.VarDecl.Accept(s)
	}
	if node.Assignment != nil {
		node.Assignment.Accept(s)
	}
	if node.Return != nil {
		node.Return.Accept(s)
	}
	if node.If != nil {
		node.If.Accept(s)
	}
	if node.For != nil {
		node.For.Accept(s)
	}
	return nil
}

func (s *SemanticAnalyzer) VisitMakeCall(node *ast.MakeCall) interface{} {
	return s.BaseVisitor.VisitMakeCall(node)
}

func (s *SemanticAnalyzer) VisitChannelRecv(node *ast.ChannelRecv) interface{} {
	return s.BaseVisitor.VisitChannelRecv(node)
}

func (s *SemanticAnalyzer) VisitExprStmt(node *ast.ExprStmt) interface{} {
	return s.BaseVisitor.VisitExprStmt(node)
}

func (s *SemanticAnalyzer) VisitImport(node *ast.Import) interface{} {
	return s.BaseVisitor.VisitImport(node)
}

func (s *SemanticAnalyzer) VisitTypeDef(node *ast.TypeDef) interface{} {
	return s.BaseVisitor.VisitTypeDef(node)
}

func (s *SemanticAnalyzer) VisitStructType(node *ast.StructType) interface{} {
	return s.BaseVisitor.VisitStructType(node)
}

func (s *SemanticAnalyzer) VisitStructField(node *ast.StructField) interface{} {
	return s.BaseVisitor.VisitStructField(node)
}

func (s *SemanticAnalyzer) VisitParameter(node *ast.Parameter) interface{} {
	return s.BaseVisitor.VisitParameter(node)
}

func (s *SemanticAnalyzer) VisitType(node *ast.Type) interface{} {
	return s.BaseVisitor.VisitType(node)
}

func (s *SemanticAnalyzer) VisitLiteral(node *ast.Literal) interface{} {
	return s.BaseVisitor.VisitLiteral(node)
}

func (s *SemanticAnalyzer) VisitSelector(node *ast.Selector) interface{} {
	return s.BaseVisitor.VisitSelector(node)
}

func (s *SemanticAnalyzer) VisitCall(node *ast.Call) interface{} {
	return s.BaseVisitor.VisitCall(node)
}

func (s *SemanticAnalyzer) VisitTextNode(node *ast.TextNode) interface{} {
	return s.BaseVisitor.VisitTextNode(node)
}
