// Package ast defines the Abstract Syntax Tree for Guix components
package ast

import (
	"github.com/alecthomas/participle/v2/lexer"
)

// File represents a complete .gx file
type File struct {
	Pos        lexer.Position
	Package    string       `"package" @Ident`
	Imports    []*Import    `@@*`
	Types      []*TypeDef   `@@*`
	Components []*Component `@@*`
}

// TypeDef represents a type definition
type TypeDef struct {
	Pos    lexer.Position
	Name   string        `"type" @Ident`
	Struct *StructType   `@@`
}

// StructType represents a struct type definition
type StructType struct {
	Pos    lexer.Position
	Fields []*StructField `"struct" "{" @@* "}"`
}

// StructField represents a field in a struct
type StructField struct {
	Pos  lexer.Position
	Name string `@Ident`
	Type *Type  `@@`
}

// Import represents an import statement
type Import struct {
	Pos  lexer.Position
	Path string `"import" @String`
}

// Component represents a component or function definition
// If Results is non-empty, it's a component function that returns Component interface
// If Results is empty, it's a regular helper function
type Component struct {
	Pos     lexer.Position
	Name    string       `"func" @Ident`
	Params  []*Parameter `"(" (@@ ("," @@)*)? ")"`
	Results []*Type      `("(" (@@ ("," @@)*)? ")")?`
	Body    *Body        `@@`
}

// Parameter represents a component parameter with name and type
type Parameter struct {
	Pos  lexer.Position
	Name string `@Ident`
	Type *Type  `@@`
}

// Type represents a type specification
type Type struct {
	Pos         lexer.Position
	IsChannel   bool   `@("<-")?`
	IsChan      bool   `@("chan")?`
	Name        string `@Ident`
	Generic     *Type  `("[" @@ "]")?`
	IsPointer   bool   `@("*")?`
	IsFunc      bool   `@("func")?`
	FuncParams  []*Type
	FuncResults []*Type
}

// Body represents a component body with optional variable declarations, statements, and UI tree
type Body struct {
	Pos        lexer.Position
	VarDecls   []*VarDecl       `"{" @@*`
	Statements []*BodyStatement `@@*`
	Children   []*Node          `@@* "}"`
}

// BodyStatement represents a statement in a component body
type BodyStatement struct {
	Pos        lexer.Position
	VarDecl    *VarDecl    `@@`
	Assignment *Assignment `| @@`
	Return     *Return     `| @@`
	If         *IfStmt     `| @@`
	For        *ForLoop    `| @@`
}

// Node represents any node in the component tree
type Node struct {
	Pos         lexer.Position
	Text        *TextNode    `@@`
	Template    *Template    `| @@`
	IfExpr      *IfExpr      `| @@`
	ForLoop     *ForLoop     `| @@`
	ChannelRecv *ChannelRecv `| @@`
	Element     *Element     `| @@`
	ExprStmt    *ExprStmt    `| @@`
}

// ExprStmt represents an expression statement (function call without braces)
// This is now handled as part of VarDecls since they both use assignments
type ExprStmt struct {
	Pos  lexer.Position
	Expr *CallOrSelect `@@`
}

// Element represents an element like Div, Input, Button
// Example: Div(Class("container"), OnClick(handler)) { ... }
type Element struct {
	Pos      lexer.Position
	Tag      string  `@Ident`
	Props    []*Prop `("(" (@@ ("," @@)*)? ")")?`
	Children []*Node `("{" @@* "}")?`
}

// Prop represents a property or event handler
type Prop struct {
	Pos   lexer.Position
	Name  string `@Ident`
	Value *Expr  `"(" @@ ")"`
}

// Expr represents an expression with optional binary operations
type Expr struct {
	Pos     lexer.Position
	Left    *Primary      `@@`
	BinOps  []*BinaryOp   `@@*`
}

// BinaryOp represents a binary operation (operator and right operand)
type BinaryOp struct {
	Pos   lexer.Position
	Op    string   `@("==" | "!=" | "<=" | ">=" | "<" | ">" | "&&" | "||" | "+" | "-" | "*" | "/")`
	Right *Primary `@@`
}

// Primary represents a primary expression (operand in binary expressions)
type Primary struct {
	Pos          lexer.Position
	Unary        *UnaryExpr     `  @@`
	Literal      *Literal       `| @@`
	CompositeLit *CompositeLit  `| @@`
	MakeCall     *MakeCall      `| @@`
	CallOrSel    *CallOrSelect  `| @@`
	FuncLit      *FuncLit       `| @@`
	ChannelOp    *ChannelOp     `| @@`
	Paren        *Expr          `| "(" @@ ")"`
	Ident        string         `| @Ident`
}

// CompositeLit represents a composite literal (struct initialization)
// Example: CalculatorState{Display: "0", PreviousValue: 0}
type CompositeLit struct {
	Pos      lexer.Position
	Type     string         `@Ident`
	Elements []*KeyValue    `"{" (@@ ("," @@)*)? ","? "}"`
}

// KeyValue represents a key-value pair in a composite literal
type KeyValue struct {
	Pos   lexer.Position
	Key   string `@Ident ":"`
	Value *Expr  `@@`
}

// UnaryExpr represents a unary expression (e.g., !x, -x)
type UnaryExpr struct {
	Pos   lexer.Position
	Op    string   `@("!" | "-" | "+")`
	Right *Primary `@@`
}

// Literal represents a literal value
type Literal struct {
	Pos    lexer.Position
	String *string `@String`
	Number *string `| @Number`
	Bool   *string `| @("true" | "false")`
}

// CallOrSelect represents either a selector (obj.field.field) or a call (func() or obj.method())
// This unifies both to avoid grammar ambiguity
type CallOrSelect struct {
	Pos    lexer.Position
	Base   string   `@Ident`
	Fields []string `("." @Ident)*`
	Args   []*Expr  `("(" (@@ ("," @@)*)? ")")?`
}

// Selector represents a selector expression - deprecated, use CallOrSelect
// Example: e.Target.Value
type Selector struct {
	Pos    lexer.Position
	Base   string   `@Ident`
	Fields []string `("." @Ident)+`
}

// Call represents a function call or method call - deprecated, use CallOrSelect
// Example: OnClick(handler) or strconv.Atoi(value)
type Call struct {
	Pos    lexer.Position
	Base   string   `@Ident`
	Fields []string `("." @Ident)*`
	Args   []*Expr  `"(" (@@ ("," @@)*)? ")"`
}

// MakeCall represents a make() function call with type argument
// Example: make(chan int, 10)
type MakeCall struct {
	Pos      lexer.Position
	Func     string `@"make"`
	ChanType *Type  `"(" "chan" @@`
	Size     *Expr  `("," @@)? ")"`
}

// FuncLit represents a function literal
// Example: func(ev Event) { ... }
type FuncLit struct {
	Pos     lexer.Position
	Params  []*Parameter `"func" "(" (@@ ("," @@)*)? ")"`
	Results []*Type      `("(" (@@ ("," @@)*)? ")")?`
	Body    *FuncBody    `@@`
}

// FuncBody represents a function body
type FuncBody struct {
	Pos        lexer.Position
	Statements []*Statement `"{" @@* "}"`
}

// Statement represents a statement in a function body
type Statement struct {
	Pos        lexer.Position
	ExprStmt   *ExpressionStmt   `@@`  // Try this first to support function calls
	VarDecl    *VarDecl          `| @@`
	Return     *Return           `| @@`
	If         *IfStmt           `| @@`
	For        *ForLoop          `| @@`
	Assignment *Assignment       `| @@` // Deprecated: kept for backward compatibility
	Expr       *Expr             `| @@` // Deprecated: kept for backward compatibility
}

// ExpressionStmt represents a unified expression statement that handles both
// function calls and assignments. This eliminates parser ambiguity.
// If Op is empty, it's an expression (function call).
// If Op is present, it's an assignment.
type ExpressionStmt struct {
	Pos    lexer.Position
	Base   string   `@Ident`
	Fields []string `("." @Ident)*`
	Args   []*Expr  `( "(" (@@ ("," @@)*)? ")" )?`  // Optional args for function call
	Op     string   `( @("<-" | ":=" | "=" | "+=" | "-=" | "*=" | "/=") )?`  // Optional operator for assignment
	Right  *Expr    `( @@ )?`  // Right side for assignment
}

// Assignment represents an assignment statement (DEPRECATED - use ExpressionStmt)
// Kept for backward compatibility during migration
type Assignment struct {
	Pos          lexer.Position
	Left         string   `@Ident`
	LeftSelector []string `("." @Ident)*`
	Op           string   `@("<-" | ":=" | "=" | "+=" | "-=" | "*=" | "/=")`
	Right        *Expr    `@@`
}

// Return represents a return statement
type Return struct {
	Pos    lexer.Position
	Values []*Expr `"return" (@@)?`
}

// IfStmt represents an if statement
type IfStmt struct {
	Pos  lexer.Position
	Cond *Expr     `"if" @@`
	Body *FuncBody `@@`
	Else *Else     `("else" @@)?`
}

// Else represents an else clause
type Else struct {
	Pos    lexer.Position
	IfStmt *IfStmt   `@@`
	Body   *FuncBody `| @@`
}

// VarDecl represents a variable declaration
// Example: counter := make(chan int) or n, err := strconv.Atoi(value)
type VarDecl struct {
	Pos    lexer.Position
	Names  []string `@Ident ("," @Ident)*`
	Op     string   `@":="`
	Values []*Expr  `@@ ("," @@)*`
}

// TextNode represents plain text
type TextNode struct {
	Pos  lexer.Position
	Text string `@String`
}

// Template represents a template string with interpolation
// Example: `Counter: {<-counterChannel}`
type Template struct {
	Pos       lexer.Position
	Fragments []*Fragment `Backtick @@* BacktickEnd`
}

// Fragment represents part of a template (text or expression)
type Fragment struct {
	Pos  lexer.Position
	Text string `@TemplateText`
	Expr *Expr  `| ("{" @@ "}")`
}

// IfExpr represents an if expression (not statement)
// Example: if cond { a } else { b }
type IfExpr struct {
	Pos       lexer.Position
	Cond      *Expr `"if" @@`
	TrueBody  *Body `@@`
	FalseBody *Body `("else" @@)?`
}

// ForLoop represents a for loop
type ForLoop struct {
	Pos   lexer.Position
	Key   string `"for" (@Ident ",")? `
	Val   string `@Ident`
	Range *Expr  `"in" @@`
	Body  *Body  `@@`
}

// ChannelRecv represents a channel receive operation
// Example: <-counterChannel
type ChannelRecv struct {
	Pos     lexer.Position
	Channel string `"<-" @Ident`
}

// ChannelOp represents channel operations
type ChannelOp struct {
	Pos     lexer.Position
	Op      string `@("<-")`
	Channel string `@Ident`
}

// Note: More complex expressions like binary, unary, selector, and index
// expressions are not yet supported in the parser to avoid infinite recursion.
// These will be added in a future version with proper precedence handling.

// Accept methods for visitor pattern

// Top-level declarations
func (n *File) Accept(v Visitor) interface{}        { return v.VisitFile(n) }
func (n *Import) Accept(v Visitor) interface{}      { return v.VisitImport(n) }
func (n *TypeDef) Accept(v Visitor) interface{}     { return v.VisitTypeDef(n) }
func (n *StructType) Accept(v Visitor) interface{}  { return v.VisitStructType(n) }
func (n *StructField) Accept(v Visitor) interface{} { return v.VisitStructField(n) }
func (n *Component) Accept(v Visitor) interface{}   { return v.VisitComponent(n) }
func (n *Parameter) Accept(v Visitor) interface{}   { return v.VisitParameter(n) }
func (n *Type) Accept(v Visitor) interface{}        { return v.VisitType(n) }

// Body and statements
func (n *Body) Accept(v Visitor) interface{}           { return v.VisitBody(n) }
func (n *BodyStatement) Accept(v Visitor) interface{}  { return v.VisitBodyStatement(n) }
func (n *Statement) Accept(v Visitor) interface{}      { return v.VisitStatement(n) }
func (n *ExpressionStmt) Accept(v Visitor) interface{} { return v.VisitExpressionStmt(n) }
func (n *VarDecl) Accept(v Visitor) interface{}        { return v.VisitVarDecl(n) }
func (n *Assignment) Accept(v Visitor) interface{}     { return v.VisitAssignment(n) }
func (n *Return) Accept(v Visitor) interface{}        { return v.VisitReturn(n) }
func (n *IfStmt) Accept(v Visitor) interface{}        { return v.VisitIfStmt(n) }
func (n *Else) Accept(v Visitor) interface{}          { return v.VisitElse(n) }
func (n *ForLoop) Accept(v Visitor) interface{}       { return v.VisitForLoop(n) }

// Nodes and expressions
func (n *Node) Accept(v Visitor) interface{}           { return v.VisitNode(n) }
func (n *Element) Accept(v Visitor) interface{}        { return v.VisitElement(n) }
func (n *Prop) Accept(v Visitor) interface{}           { return v.VisitProp(n) }
func (n *ExprStmt) Accept(v Visitor) interface{}       { return v.VisitExprStmt(n) }
func (n *Expr) Accept(v Visitor) interface{}           { return v.VisitExpr(n) }
func (n *BinaryOp) Accept(v Visitor) interface{}       { return v.VisitBinaryOp(n) }
func (n *Primary) Accept(v Visitor) interface{}        { return v.VisitPrimary(n) }
func (n *UnaryExpr) Accept(v Visitor) interface{}      { return v.VisitUnaryExpr(n) }
func (n *Literal) Accept(v Visitor) interface{}        { return v.VisitLiteral(n) }
func (n *CallOrSelect) Accept(v Visitor) interface{}   { return v.VisitCallOrSelect(n) }
func (n *Selector) Accept(v Visitor) interface{}       { return v.VisitSelector(n) }
func (n *Call) Accept(v Visitor) interface{}           { return v.VisitCall(n) }
func (n *MakeCall) Accept(v Visitor) interface{}       { return v.VisitMakeCall(n) }
func (n *FuncLit) Accept(v Visitor) interface{}        { return v.VisitFuncLit(n) }
func (n *FuncBody) Accept(v Visitor) interface{}       { return v.VisitFuncBody(n) }
func (n *CompositeLit) Accept(v Visitor) interface{}   { return v.VisitCompositeLit(n) }
func (n *KeyValue) Accept(v Visitor) interface{}       { return v.VisitKeyValue(n) }

// Templates and special nodes
func (n *TextNode) Accept(v Visitor) interface{}    { return v.VisitTextNode(n) }
func (n *Template) Accept(v Visitor) interface{}    { return v.VisitTemplate(n) }
func (n *Fragment) Accept(v Visitor) interface{}    { return v.VisitFragment(n) }
func (n *IfExpr) Accept(v Visitor) interface{}      { return v.VisitIfExpr(n) }
func (n *ChannelRecv) Accept(v Visitor) interface{} { return v.VisitChannelRecv(n) }
func (n *ChannelOp) Accept(v Visitor) interface{}   { return v.VisitChannelOp(n) }
