// Package ast defines the Abstract Syntax Tree for Guix components
package ast

// BaseVisitor provides default traversal implementations for all AST nodes.
// Visitors can embed this struct and override only the methods they need.
// By default, BaseVisitor traverses the entire AST without modification.
type BaseVisitor struct{}

// Ensure BaseVisitor implements Visitor interface
var _ Visitor = (*BaseVisitor)(nil)

// Top-level declarations

func (v *BaseVisitor) VisitFile(node *File) interface{} {
	for _, imp := range node.Imports {
		imp.Accept(v)
	}
	for _, gpuStruct := range node.GPUStructs {
		gpuStruct.Accept(v)
	}
	for _, gpuBinding := range node.GPUBindings {
		gpuBinding.Accept(v)
	}
	for _, gpuFunc := range node.GPUFunctions {
		gpuFunc.Accept(v)
	}
	for _, typeDef := range node.Types {
		typeDef.Accept(v)
	}
	for _, comp := range node.Components {
		comp.Accept(v)
	}
	for _, method := range node.Methods {
		method.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitImport(node *Import) interface{} {
	return nil
}

func (v *BaseVisitor) VisitTypeDef(node *TypeDef) interface{} {
	if node.Struct != nil {
		node.Struct.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitStructType(node *StructType) interface{} {
	for _, field := range node.Fields {
		field.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitStructField(node *StructField) interface{} {
	if node.Type != nil {
		node.Type.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitComponent(node *Component) interface{} {
	for _, param := range node.Params {
		param.Accept(v)
	}
	for _, result := range node.Results {
		result.Accept(v)
	}
	if node.Body != nil {
		node.Body.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitMethod(node *Method) interface{} {
	if node.Receiver != nil {
		node.Receiver.Accept(v)
	}
	for _, param := range node.Params {
		param.Accept(v)
	}
	for _, result := range node.Results {
		result.Accept(v)
	}
	if node.Body != nil {
		node.Body.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitReceiver(node *Receiver) interface{} {
	return nil
}

func (v *BaseVisitor) VisitParameter(node *Parameter) interface{} {
	if node.Type != nil {
		node.Type.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitType(node *Type) interface{} {
	if node.Generic != nil {
		node.Generic.Accept(v)
	}
	for _, param := range node.FuncParams {
		param.Accept(v)
	}
	for _, result := range node.FuncResults {
		result.Accept(v)
	}
	return nil
}

// Body and statements

func (v *BaseVisitor) VisitBody(node *Body) interface{} {
	for _, varDecl := range node.VarDecls {
		varDecl.Accept(v)
	}
	for _, stmt := range node.Statements {
		stmt.Accept(v)
	}
	for _, child := range node.Children {
		child.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitBodyStatement(node *BodyStatement) interface{} {
	if node.VarDecl != nil {
		node.VarDecl.Accept(v)
	}
	if node.AssignStmt != nil {
		node.AssignStmt.Accept(v)
	}
	if node.Return != nil {
		node.Return.Accept(v)
	}
	if node.If != nil {
		node.If.Accept(v)
	}
	if node.For != nil {
		node.For.Accept(v)
	}
	if node.Switch != nil {
		node.Switch.Accept(v)
	}
	if node.Select != nil {
		node.Select.Accept(v)
	}
	if node.GoStmt != nil {
		node.GoStmt.Accept(v)
	}
	// Deprecated - kept for backward compatibility
	if node.Assignment != nil {
		node.Assignment.Accept(v)
	}
	if node.CallStmt != nil {
		node.CallStmt.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitStatement(node *Statement) interface{} {
	if node.CallStmt != nil {
		node.CallStmt.Accept(v)
	}
	if node.VarDecl != nil {
		node.VarDecl.Accept(v)
	}
	if node.AssignStmt != nil {
		node.AssignStmt.Accept(v)
	}
	if node.Return != nil {
		node.Return.Accept(v)
	}
	if node.If != nil {
		node.If.Accept(v)
	}
	if node.For != nil {
		node.For.Accept(v)
	}
	if node.Switch != nil {
		node.Switch.Accept(v)
	}
	if node.Select != nil {
		node.Select.Accept(v)
	}
	if node.GoStmt != nil {
		node.GoStmt.Accept(v)
	}
	// Deprecated - kept for backward compatibility
	if node.Assignment != nil {
		node.Assignment.Accept(v)
	}
	if node.Expr != nil {
		node.Expr.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitCallStmt(node *CallStmt) interface{} {
	for _, arg := range node.Args {
		arg.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitAssignmentStmt(node *AssignmentStmt) interface{} {
	if node.Index != nil {
		node.Index.Accept(v)
	}
	if node.Right != nil {
		node.Right.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitVarDecl(node *VarDecl) interface{} {
	for _, val := range node.Values {
		val.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitAssignment(node *Assignment) interface{} {
	if node.Right != nil {
		node.Right.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitReturn(node *Return) interface{} {
	for _, val := range node.Values {
		val.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitGoStmt(node *GoStmt) interface{} {
	if node.Func != nil {
		node.Func.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitSwitchStmt(node *SwitchStmt) interface{} {
	if node.Expr != nil {
		node.Expr.Accept(v)
	}
	for _, caseClause := range node.Cases {
		caseClause.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitCaseClause(node *CaseClause) interface{} {
	for _, val := range node.Values {
		val.Accept(v)
	}
	for _, stmt := range node.Statements {
		stmt.Accept(v)
	}
	for _, stmt := range node.DefStmts {
		stmt.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitSelectStmt(node *SelectStmt) interface{} {
	for _, commClause := range node.Cases {
		commClause.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitCommClause(node *CommClause) interface{} {
	if node.Comm != nil {
		node.Comm.Accept(v)
	}
	for _, stmt := range node.Statements {
		stmt.Accept(v)
	}
	for _, stmt := range node.DefStmts {
		stmt.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitCommCase(node *CommCase) interface{} {
	if node.Send != nil {
		node.Send.Accept(v)
	}
	if node.Recv != nil {
		node.Recv.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitSendStmt(node *SendStmt) interface{} {
	if node.Value != nil {
		node.Value.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitRecvStmt(node *RecvStmt) interface{} {
	return nil
}

func (v *BaseVisitor) VisitIfStmt(node *IfStmt) interface{} {
	if node.Cond != nil {
		node.Cond.Accept(v)
	}
	if node.Body != nil {
		node.Body.Accept(v)
	}
	if node.Else != nil {
		node.Else.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitElse(node *Else) interface{} {
	if node.IfStmt != nil {
		node.IfStmt.Accept(v)
	}
	if node.Body != nil {
		node.Body.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitForLoop(node *ForLoop) interface{} {
	// Range-based for loop
	if node.Range != nil {
		node.Range.Accept(v)
	}
	if node.Body != nil {
		node.Body.Accept(v)
	}
	// C-style for loop
	if node.Init != nil {
		node.Init.Accept(v)
	}
	if node.Cond != nil {
		node.Cond.Accept(v)
	}
	if node.Post != nil {
		node.Post.Accept(v)
	}
	if node.CBody != nil {
		node.CBody.Accept(v)
	}
	return nil
}

// Nodes and expressions

func (v *BaseVisitor) VisitNode(node *Node) interface{} {
	if node.Text != nil {
		node.Text.Accept(v)
	}
	if node.Template != nil {
		node.Template.Accept(v)
	}
	if node.IfExpr != nil {
		node.IfExpr.Accept(v)
	}
	if node.ForLoop != nil {
		node.ForLoop.Accept(v)
	}
	if node.ChannelRecv != nil {
		node.ChannelRecv.Accept(v)
	}
	if node.Element != nil {
		node.Element.Accept(v)
	}
	if node.ExprStmt != nil {
		node.ExprStmt.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitElement(node *Element) interface{} {
	for _, prop := range node.Props {
		prop.Accept(v)
	}
	for _, child := range node.Children {
		child.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitProp(node *Prop) interface{} {
	for _, arg := range node.Args {
		if arg != nil {
			arg.Accept(v)
		}
	}
	return nil
}

func (v *BaseVisitor) VisitExprStmt(node *ExprStmt) interface{} {
	if node.Expr != nil {
		node.Expr.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitExpr(node *Expr) interface{} {
	if node.Left != nil {
		node.Left.Accept(v)
	}
	for _, binOp := range node.BinOps {
		binOp.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitBinaryOp(node *BinaryOp) interface{} {
	if node.Right != nil {
		node.Right.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitPrimary(node *Primary) interface{} {
	if node.Unary != nil {
		node.Unary.Accept(v)
	}
	if node.Literal != nil {
		node.Literal.Accept(v)
	}
	if node.CompositeLit != nil {
		node.CompositeLit.Accept(v)
	}
	if node.MakeCall != nil {
		node.MakeCall.Accept(v)
	}
	if node.IndexExpr != nil {
		node.IndexExpr.Accept(v)
	}
	if node.CallOrSel != nil {
		node.CallOrSel.Accept(v)
	}
	if node.FuncLit != nil {
		node.FuncLit.Accept(v)
	}
	if node.ChannelOp != nil {
		node.ChannelOp.Accept(v)
	}
	if node.Paren != nil {
		node.Paren.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitUnaryExpr(node *UnaryExpr) interface{} {
	if node.Right != nil {
		node.Right.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitLiteral(node *Literal) interface{} {
	return nil
}

func (v *BaseVisitor) VisitIndexExpr(node *IndexExpr) interface{} {
	if node.Index != nil {
		node.Index.Accept(v)
	}
	if node.Slice != nil {
		node.Slice.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitSliceExpr(node *SliceExpr) interface{} {
	if node.Low != nil {
		node.Low.Accept(v)
	}
	if node.High != nil {
		node.High.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitCallOrSelect(node *CallOrSelect) interface{} {
	for _, arg := range node.Args {
		arg.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitSelector(node *Selector) interface{} {
	return nil
}

func (v *BaseVisitor) VisitCall(node *Call) interface{} {
	for _, arg := range node.Args {
		arg.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitMakeCall(node *MakeCall) interface{} {
	if node.ChanType != nil {
		node.ChanType.Accept(v)
	}
	if node.ChanSize != nil {
		node.ChanSize.Accept(v)
	}
	if node.SliceType != nil {
		node.SliceType.Accept(v)
	}
	if node.SliceLen != nil {
		node.SliceLen.Accept(v)
	}
	if node.SliceCap != nil {
		node.SliceCap.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitFuncLit(node *FuncLit) interface{} {
	for _, param := range node.Params {
		param.Accept(v)
	}
	for _, result := range node.Results {
		result.Accept(v)
	}
	if node.Body != nil {
		node.Body.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitFuncBody(node *FuncBody) interface{} {
	for _, stmt := range node.Statements {
		stmt.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitCompositeLit(node *CompositeLit) interface{} {
	for _, elem := range node.Elements {
		elem.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitKeyValue(node *KeyValue) interface{} {
	if node.Value != nil {
		node.Value.Accept(v)
	}
	return nil
}

// Templates and special nodes

func (v *BaseVisitor) VisitTextNode(node *TextNode) interface{} {
	return nil
}

func (v *BaseVisitor) VisitTemplate(node *Template) interface{} {
	for _, frag := range node.Fragments {
		frag.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitFragment(node *Fragment) interface{} {
	if node.Expr != nil {
		node.Expr.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitIfExpr(node *IfExpr) interface{} {
	if node.Cond != nil {
		node.Cond.Accept(v)
	}
	if node.TrueBody != nil {
		node.TrueBody.Accept(v)
	}
	if node.FalseBody != nil {
		node.FalseBody.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitChannelRecv(node *ChannelRecv) interface{} {
	return nil
}

func (v *BaseVisitor) VisitChannelOp(node *ChannelOp) interface{} {
	return nil
}

// GPU/WGSL nodes

func (v *BaseVisitor) VisitGPUDecorator(node *GPUDecorator) interface{} {
	for _, arg := range node.Args {
		arg.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitGPUStructDecl(node *GPUStructDecl) interface{} {
	for _, decorator := range node.Decorators {
		decorator.Accept(v)
	}
	if node.Struct != nil {
		node.Struct.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitGPUStructType(node *GPUStructType) interface{} {
	for _, field := range node.Fields {
		field.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitGPUField(node *GPUField) interface{} {
	for _, decorator := range node.Decorators {
		decorator.Accept(v)
	}
	if node.Type != nil {
		node.Type.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitGPUType(node *GPUType) interface{} {
	if node.Generic != nil {
		node.Generic.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitGPUBindingDecl(node *GPUBindingDecl) interface{} {
	for _, decorator := range node.Decorators {
		decorator.Accept(v)
	}
	if node.Type != nil {
		node.Type.Accept(v)
	}
	if node.InitialExpr != nil {
		node.InitialExpr.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitGPUFuncDecl(node *GPUFuncDecl) interface{} {
	for _, decorator := range node.Decorators {
		decorator.Accept(v)
	}
	for _, param := range node.Params {
		param.Accept(v)
	}
	if node.Results != nil {
		node.Results.Accept(v)
	}
	if node.Body != nil {
		node.Body.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitGPUParameter(node *GPUParameter) interface{} {
	for _, decorator := range node.Decorators {
		decorator.Accept(v)
	}
	if node.Type != nil {
		node.Type.Accept(v)
	}
	return nil
}

func (v *BaseVisitor) VisitGPUReturnType(node *GPUReturnType) interface{} {
	for _, decorator := range node.Decorators {
		decorator.Accept(v)
	}
	if node.Type != nil {
		node.Type.Accept(v)
	}
	return nil
}
