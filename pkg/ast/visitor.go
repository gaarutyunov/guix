// Package ast defines the Abstract Syntax Tree for Guix components
package ast

// Visitor interface defines methods for visiting each AST node type.
// Implementations can traverse and transform the AST by implementing
// these methods. The return type is interface{} to allow flexibility
// in what visitors return (e.g., transformed nodes, analysis results, etc.).
type Visitor interface {
	// Top-level declarations
	VisitFile(*File) interface{}
	VisitImport(*Import) interface{}
	VisitTypeDef(*TypeDef) interface{}
	VisitStructType(*StructType) interface{}
	VisitStructField(*StructField) interface{}
	VisitComponent(*Component) interface{}
	VisitMethod(*Method) interface{}
	VisitReceiver(*Receiver) interface{}
	VisitParameter(*Parameter) interface{}
	VisitType(*Type) interface{}

	// Body and statements
	VisitBody(*Body) interface{}
	VisitBodyStatement(*BodyStatement) interface{}
	VisitStatement(*Statement) interface{}
	VisitCallStmt(*CallStmt) interface{}
	VisitAssignmentStmt(*AssignmentStmt) interface{}
	VisitVarDecl(*VarDecl) interface{}
	VisitAssignment(*Assignment) interface{}
	VisitGoStmt(*GoStmt) interface{}
	VisitSwitchStmt(*SwitchStmt) interface{}
	VisitCaseClause(*CaseClause) interface{}
	VisitSelectStmt(*SelectStmt) interface{}
	VisitCommClause(*CommClause) interface{}
	VisitCommCase(*CommCase) interface{}
	VisitSendStmt(*SendStmt) interface{}
	VisitRecvStmt(*RecvStmt) interface{}
	VisitReturn(*Return) interface{}
	VisitIfStmt(*IfStmt) interface{}
	VisitElse(*Else) interface{}
	VisitForLoop(*ForLoop) interface{}

	// Nodes and expressions
	VisitNode(*Node) interface{}
	VisitElement(*Element) interface{}
	VisitProp(*Prop) interface{}
	VisitExprStmt(*ExprStmt) interface{}
	VisitExpr(*Expr) interface{}
	VisitBinaryOp(*BinaryOp) interface{}
	VisitPrimary(*Primary) interface{}
	VisitUnaryExpr(*UnaryExpr) interface{}
	VisitLiteral(*Literal) interface{}
	VisitIndexExpr(*IndexExpr) interface{}
	VisitSliceExpr(*SliceExpr) interface{}
	VisitCallOrSelect(*CallOrSelect) interface{}
	VisitSelector(*Selector) interface{} // Deprecated but still in AST
	VisitCall(*Call) interface{}         // Deprecated but still in AST
	VisitMakeCall(*MakeCall) interface{}
	VisitFuncLit(*FuncLit) interface{}
	VisitFuncBody(*FuncBody) interface{}
	VisitCompositeLit(*CompositeLit) interface{}
	VisitKeyValue(*KeyValue) interface{}

	// Templates and special nodes
	VisitTextNode(*TextNode) interface{}
	VisitTemplate(*Template) interface{}
	VisitFragment(*Fragment) interface{}
	VisitIfExpr(*IfExpr) interface{}
	VisitChannelRecv(*ChannelRecv) interface{}
	VisitChannelOp(*ChannelOp) interface{}
}

// Node interface that all AST nodes must implement to support visitor pattern
type ASTNode interface {
	Accept(v Visitor) interface{}
}
