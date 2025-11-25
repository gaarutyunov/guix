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
	Components []*Component `@@*`
}

// Import represents an import statement
type Import struct {
	Pos  lexer.Position
	Path string `"import" @String`
}

// Component represents a component definition
// Example: func Counter(counterChannel: <-chan int) { ... }
type Component struct {
	Pos    lexer.Position
	Name   string       `"func" @Ident`
	Params []*Parameter `"(" (@@ ("," @@)*)? ")"`
	Body   *Body        `@@`
}

// Parameter represents a component parameter with name and type
type Parameter struct {
	Pos  lexer.Position
	Name string `@Ident ":"`
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

// Body represents a component body with optional variable declarations and UI tree
type Body struct {
	Pos      lexer.Position
	VarDecls []*VarDecl `"{" @@*`
	Children []*Node    `@@* "}"`
}

// Node represents any node in the component tree
type Node struct {
	Pos         lexer.Position
	Element     *Element     `@@`
	Text        *TextNode    `| @@`
	Template    *Template    `| @@`
	IfExpr      *IfExpr      `| @@`
	ForLoop     *ForLoop     `| @@`
	ChannelRecv *ChannelRecv `| @@`
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

// Expr represents an expression (simplified to avoid recursion)
type Expr struct {
	Pos       lexer.Position
	Literal   *Literal   `@@`
	Ident     string     `| @Ident`
	MakeCall  *MakeCall  `| @@`
	Call      *Call      `| @@`
	FuncLit   *FuncLit   `| @@`
	ChannelOp *ChannelOp `| @@`
}

// Literal represents a literal value
type Literal struct {
	Pos    lexer.Position
	String *string `@String`
	Number *string `| @Number`
	Bool   *string `| @("true" | "false")`
}

// Call represents a function call
// Example: OnClick(handler)
type Call struct {
	Pos  lexer.Position
	Func string  `@Ident`
	Args []*Expr `"(" (@@ ("," @@)*)? ")"`
}

// MakeCall represents a make() function call with type argument
// Example: make(chan int, 10)
type MakeCall struct {
	Pos      lexer.Position
	Func     string  `@"make"`
	ChanType *Type   `"(" "chan" @@`
	Size     *Expr   `("," @@)? ")"`
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
	Assignment *Assignment `@@`
	Expr       *Expr       `| @@`
	Return     *Return     `| @@`
	If         *IfStmt     `| @@`
	For        *ForLoop    `| @@`
	VarDecl    *VarDecl    `| @@`
}

// Assignment represents an assignment statement
type Assignment struct {
	Pos   lexer.Position
	Left  string `@Ident`
	Op    string `@("<-" | ":=" | "=" | "+=" | "-=" | "*=" | "/=")`
	Right *Expr  `@@`
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
// Example: counter := make(chan int)
type VarDecl struct {
	Pos   lexer.Position
	Name  string `@Ident`
	Op    string `@":="`
	Value *Expr  `@@`
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
