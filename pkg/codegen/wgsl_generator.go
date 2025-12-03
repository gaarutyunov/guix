// Package codegen provides WGSL code generation from Guix GPU AST
package codegen

import (
	"bytes"
	"fmt"
	"strings"

	guixast "github.com/gaarutyunov/guix/pkg/ast"
)

// WGSLGenerator generates WGSL shader code from Guix GPU AST
type WGSLGenerator struct {
	guixast.BaseVisitor
	output      bytes.Buffer
	indentLevel int
	structs     map[string]*guixast.GPUStructDecl
	bindings    []*guixast.GPUBindingDecl
	functions   []*guixast.GPUFuncDecl
}

// NewWGSLGenerator creates a new WGSL code generator
func NewWGSLGenerator() *WGSLGenerator {
	return &WGSLGenerator{
		structs: make(map[string]*guixast.GPUStructDecl),
	}
}

// Generate generates WGSL code from a Guix file
func (g *WGSLGenerator) Generate(file *guixast.File) ([]byte, error) {
	// Collect GPU declarations
	for _, gpuStruct := range file.GPUStructs {
		g.structs[gpuStruct.Name] = gpuStruct
	}
	g.bindings = file.GPUBindings
	g.functions = file.GPUFunctions

	// Generate structs
	for _, gpuStruct := range file.GPUStructs {
		g.generateStruct(gpuStruct)
		g.writeln("")
	}

	// Generate bindings
	for _, binding := range file.GPUBindings {
		g.generateBinding(binding)
		g.writeln("")
	}

	// Generate functions
	for _, function := range file.GPUFunctions {
		g.generateFunction(function)
		g.writeln("")
	}

	return g.output.Bytes(), nil
}

// Helper methods for writing output

func (g *WGSLGenerator) write(s string) {
	g.output.WriteString(s)
}

func (g *WGSLGenerator) writeln(s string) {
	if s != "" {
		g.write(g.indent() + s)
	}
	g.output.WriteString("\n")
}

func (g *WGSLGenerator) indent() string {
	return strings.Repeat("    ", g.indentLevel)
}

func (g *WGSLGenerator) increaseIndent() {
	g.indentLevel++
}

func (g *WGSLGenerator) decreaseIndent() {
	if g.indentLevel > 0 {
		g.indentLevel--
	}
}

// Generate struct declaration
func (g *WGSLGenerator) generateStruct(s *guixast.GPUStructDecl) {
	// Write struct header
	g.writeln(fmt.Sprintf("struct %s {", s.Name))
	g.increaseIndent()

	// Write fields
	for _, field := range s.Struct.Fields {
		g.generateField(field)
	}

	g.decreaseIndent()
	g.writeln("}")
}

// Generate struct field
func (g *WGSLGenerator) generateField(field *guixast.GPUField) {
	var decorators []string

	// Collect decorators
	for _, decorator := range field.Decorators {
		decorators = append(decorators, g.formatDecorator(decorator))
	}

	// Get WGSL type
	wgslType, err := MapGPUTypeToWGSL(field.Type)
	if err != nil {
		// Fall back to field type name if mapping fails
		wgslType = field.Type.Name
	}

	// Write decorators and field
	if len(decorators) > 0 {
		decoratorStr := strings.Join(decorators, " ")
		g.writeln(fmt.Sprintf("%s %s: %s,", decoratorStr, field.Name, wgslType))
	} else {
		g.writeln(fmt.Sprintf("%s: %s,", field.Name, wgslType))
	}
}

// Generate binding declaration
func (g *WGSLGenerator) generateBinding(binding *guixast.GPUBindingDecl) {
	var parts []string

	// Collect decorators (@group, @binding, etc.)
	for _, decorator := range binding.Decorators {
		parts = append(parts, g.formatDecorator(decorator))
	}

	// Get address space and access mode from decorators
	addressSpace := ""
	for _, decorator := range binding.Decorators {
		// Strip @ prefix for comparison
		name := strings.TrimPrefix(decorator.Name, "@")
		switch name {
		case "uniform":
			addressSpace = "<uniform>"
		case "storage":
			// Check if there's a read/write/read_write argument
			if len(decorator.Args) > 0 {
				// For now, default to read
				addressSpace = "<storage, read>"
			} else {
				addressSpace = "<storage>"
			}
		}
	}

	// Get WGSL type
	wgslType, err := MapGPUTypeToWGSL(binding.Type)
	if err != nil {
		wgslType = binding.Type.Name
	}

	// Build the binding declaration
	decoratorStr := strings.Join(parts, " ")
	if addressSpace != "" {
		g.writeln(fmt.Sprintf("%s var%s %s: %s;", decoratorStr, addressSpace, binding.Name, wgslType))
	} else {
		g.writeln(fmt.Sprintf("%s var %s: %s;", decoratorStr, binding.Name, wgslType))
	}
}

// Generate function
func (g *WGSLGenerator) generateFunction(fn *guixast.GPUFuncDecl) {
	// Collect entry point decorators (@vertex, @fragment, @compute)
	var entryDecorators []string
	var otherDecorators []string

	for _, decorator := range fn.Decorators {
		// Strip @ prefix for comparison
		name := strings.TrimPrefix(decorator.Name, "@")
		if name == "vertex" || name == "fragment" || name == "compute" {
			entryDecorators = append(entryDecorators, g.formatDecorator(decorator))
		} else {
			otherDecorators = append(otherDecorators, g.formatDecorator(decorator))
		}
	}

	// Write entry point decorators
	for _, decorator := range entryDecorators {
		g.writeln(decorator)
	}

	// Write function signature
	paramsStr := g.generateParameters(fn.Params)
	returnStr := ""
	if fn.Results != nil {
		returnTypeStr, err := MapGPUTypeToWGSL(fn.Results.Type)
		if err != nil {
			returnTypeStr = fn.Results.Type.Name
		}

		// Add return type decorators
		if len(fn.Results.Decorators) > 0 {
			decorators := make([]string, len(fn.Results.Decorators))
			for i, decorator := range fn.Results.Decorators {
				decorators[i] = g.formatDecorator(decorator)
			}
			returnStr = fmt.Sprintf(" -> %s %s", strings.Join(decorators, " "), returnTypeStr)
		} else {
			returnStr = fmt.Sprintf(" -> %s", returnTypeStr)
		}
	}

	g.writeln(fmt.Sprintf("fn %s(%s)%s {", fn.Name, paramsStr, returnStr))
	g.increaseIndent()

	// Generate function body
	g.generateBody(fn.Body)

	g.decreaseIndent()
	g.writeln("}")
}

// Generate function parameters
func (g *WGSLGenerator) generateParameters(params []*guixast.GPUParameter) string {
	if len(params) == 0 {
		return ""
	}

	parts := make([]string, len(params))
	for i, param := range params {
		parts[i] = g.generateParameter(param)
	}

	return strings.Join(parts, ", ")
}

// Generate a single parameter
func (g *WGSLGenerator) generateParameter(param *guixast.GPUParameter) string {
	// Collect decorators
	var decorators []string
	for _, decorator := range param.Decorators {
		decorators = append(decorators, g.formatDecorator(decorator))
	}

	// Get WGSL type
	wgslType, err := MapGPUTypeToWGSL(param.Type)
	if err != nil {
		wgslType = param.Type.Name
	}

	// Format parameter
	if len(decorators) > 0 {
		return fmt.Sprintf("%s %s: %s", strings.Join(decorators, " "), param.Name, wgslType)
	}
	return fmt.Sprintf("%s: %s", param.Name, wgslType)
}

// Format a decorator for WGSL output
func (g *WGSLGenerator) formatDecorator(decorator *guixast.GPUDecorator) string {
	// Strip @ prefix if present (decorator.Name comes from Directive token which includes @)
	name := strings.TrimPrefix(decorator.Name, "@")

	if len(decorator.Args) == 0 {
		return fmt.Sprintf("@%s", name)
	}

	// Format arguments
	args := make([]string, len(decorator.Args))
	for i, arg := range decorator.Args {
		args[i] = g.generateExpression(arg)
	}

	return fmt.Sprintf("@%s(%s)", name, strings.Join(args, ", "))
}

// Generate function body
func (g *WGSLGenerator) generateBody(body *guixast.Body) {
	if body == nil {
		return
	}

	// Generate variable declarations
	for _, varDecl := range body.VarDecls {
		g.generateVarDecl(varDecl)
	}

	// Generate statements
	for _, stmt := range body.Statements {
		g.generateStatement(stmt)
	}

	// Note: Children nodes are not applicable in shader functions
}

// Generate variable declaration
func (g *WGSLGenerator) generateVarDecl(varDecl *guixast.VarDecl) {
	// WGSL uses 'var' or 'let' for variables
	// For := assignments, use 'let' (immutable)
	// For var declarations, use 'var' (mutable)

	if len(varDecl.Names) == 1 && len(varDecl.Values) == 1 {
		// Simple case: single variable
		g.writeln(fmt.Sprintf("let %s = %s;", varDecl.Names[0], g.generateExpression(varDecl.Values[0])))
	} else {
		// Multiple assignments - generate separately
		for i := range varDecl.Names {
			if i < len(varDecl.Values) {
				g.writeln(fmt.Sprintf("let %s = %s;", varDecl.Names[i], g.generateExpression(varDecl.Values[i])))
			}
		}
	}
}

// Generate a statement
func (g *WGSLGenerator) generateStatement(stmt *guixast.BodyStatement) {
	if stmt.VarDecl != nil {
		g.generateVarDecl(stmt.VarDecl)
	} else if stmt.AssignStmt != nil {
		g.generateAssignment(stmt.AssignStmt)
	} else if stmt.Return != nil {
		g.generateReturn(stmt.Return)
	} else if stmt.If != nil {
		g.generateIf(stmt.If)
	} else if stmt.For != nil {
		g.generateFor(stmt.For)
	} else if stmt.CallStmt != nil {
		g.generateCallStmt(stmt.CallStmt)
	}
}

// Generate assignment statement
func (g *WGSLGenerator) generateAssignment(assign *guixast.AssignmentStmt) {
	// Build left side
	left := assign.Base
	for _, field := range assign.Fields {
		left += "." + field
	}

	if assign.Index != nil {
		left += fmt.Sprintf("[%s]", g.generateExpression(assign.Index))
	}

	// Map operator
	op := assign.Op
	switch op {
	case ":=":
		// WGSL doesn't have :=, use = instead
		op = "="
	case "<-":
		// Channel operations not supported in WGSL
		// This shouldn't appear in shader code
		return
	}

	// Generate right side
	right := g.generateExpression(assign.Right)

	g.writeln(fmt.Sprintf("%s %s %s;", left, op, right))
}

// Generate return statement
func (g *WGSLGenerator) generateReturn(ret *guixast.Return) {
	if len(ret.Values) == 0 {
		g.writeln("return;")
	} else if len(ret.Values) == 1 {
		g.writeln(fmt.Sprintf("return %s;", g.generateExpression(ret.Values[0])))
	} else {
		// Multiple return values - create a struct or tuple (not fully supported yet)
		values := make([]string, len(ret.Values))
		for i, val := range ret.Values {
			values[i] = g.generateExpression(val)
		}
		g.writeln(fmt.Sprintf("return %s;", strings.Join(values, ", ")))
	}
}

// Generate if statement
func (g *WGSLGenerator) generateIf(ifStmt *guixast.IfStmt) {
	cond := g.generateExpression(ifStmt.Cond)
	g.writeln(fmt.Sprintf("if (%s) {", cond))
	g.increaseIndent()

	// Generate if body
	if ifStmt.Body != nil {
		for _, stmt := range ifStmt.Body.Statements {
			g.generateFuncStatement(stmt)
		}
	}

	g.decreaseIndent()

	// Generate else clause
	if ifStmt.Else != nil {
		if ifStmt.Else.IfStmt != nil {
			g.write(g.indent() + "} else ")
			g.output.WriteString("\n")
			g.generateIf(ifStmt.Else.IfStmt)
		} else if ifStmt.Else.Body != nil {
			g.writeln("} else {")
			g.increaseIndent()
			for _, stmt := range ifStmt.Else.Body.Statements {
				g.generateFuncStatement(stmt)
			}
			g.decreaseIndent()
			g.writeln("}")
		}
	} else {
		g.writeln("}")
	}
}

// Generate for loop
func (g *WGSLGenerator) generateFor(forLoop *guixast.ForLoop) {
	if forLoop.Range != nil {
		// Range-based for loop not directly supported in WGSL
		// Convert to C-style for loop
		g.writeln(fmt.Sprintf("// Range-based for loop not directly supported"))
		g.writeln(fmt.Sprintf("// for %s in %s", forLoop.Val, g.generateExpression(forLoop.Range)))
	} else {
		// C-style for loop
		init := ""
		if forLoop.Init != nil {
			if len(forLoop.Init.Names) > 0 && len(forLoop.Init.Values) > 0 {
				init = fmt.Sprintf("var %s: u32 = %s", forLoop.Init.Names[0], g.generateExpression(forLoop.Init.Values[0]))
			}
		}

		cond := g.generateExpression(forLoop.Cond)

		post := ""
		if forLoop.Post != nil {
			postLeft := forLoop.Post.Base
			for _, field := range forLoop.Post.Fields {
				postLeft += "." + field
			}
			post = fmt.Sprintf("%s %s %s", postLeft, forLoop.Post.Op, g.generateExpression(forLoop.Post.Right))
		}

		g.writeln(fmt.Sprintf("for (%s; %s; %s) {", init, cond, post))
		g.increaseIndent()

		if forLoop.CBody != nil {
			for _, stmt := range forLoop.CBody.Statements {
				g.generateFuncStatement(stmt)
			}
		}

		g.decreaseIndent()
		g.writeln("}")
	}
}

// Generate call statement
func (g *WGSLGenerator) generateCallStmt(call *guixast.CallStmt) {
	funcName := string(call.Base)
	for _, field := range call.Fields {
		funcName += "." + field
	}

	args := make([]string, len(call.Args))
	for i, arg := range call.Args {
		args[i] = g.generateExpression(arg)
	}

	g.writeln(fmt.Sprintf("%s(%s);", MapFunctionToWGSL(funcName), strings.Join(args, ", ")))
}

// Generate function body statement (from FuncBody)
func (g *WGSLGenerator) generateFuncStatement(stmt *guixast.Statement) {
	if stmt.VarDecl != nil {
		g.generateVarDecl(stmt.VarDecl)
	} else if stmt.AssignStmt != nil {
		g.generateAssignment(stmt.AssignStmt)
	} else if stmt.Return != nil {
		g.generateReturn(stmt.Return)
	} else if stmt.If != nil {
		g.generateIf(stmt.If)
	} else if stmt.For != nil {
		g.generateFor(stmt.For)
	} else if stmt.CallStmt != nil {
		g.generateCallStmt(stmt.CallStmt)
	}
}

// Generate expression
func (g *WGSLGenerator) generateExpression(expr *guixast.Expr) string {
	if expr == nil {
		return ""
	}

	// Start with primary expression
	result := g.generatePrimary(expr.Left)

	// Add binary operations
	for _, binOp := range expr.BinOps {
		result += fmt.Sprintf(" %s %s", binOp.Op, g.generatePrimary(binOp.Right))
	}

	return result
}

// Generate primary expression
func (g *WGSLGenerator) generatePrimary(primary *guixast.Primary) string {
	if primary == nil {
		return ""
	}

	if primary.Literal != nil {
		return g.generateLiteral(primary.Literal)
	} else if primary.Ident != "" {
		return primary.Ident
	} else if primary.CallOrSel != nil {
		return g.generateCallOrSelect(primary.CallOrSel)
	} else if primary.Unary != nil {
		return fmt.Sprintf("%s%s", primary.Unary.Op, g.generatePrimary(primary.Unary.Right))
	} else if primary.IndexExpr != nil {
		return g.generateIndexExpr(primary.IndexExpr)
	} else if primary.Paren != nil {
		return fmt.Sprintf("(%s)", g.generateExpression(primary.Paren))
	}

	return ""
}

// Generate literal
func (g *WGSLGenerator) generateLiteral(lit *guixast.Literal) string {
	if lit.String != nil {
		return *lit.String
	} else if lit.Number != nil {
		// Add type suffix for floats if needed
		num := *lit.Number
		if strings.Contains(num, ".") && !strings.HasSuffix(num, "f") {
			return num // WGSL infers float type
		}
		return num
	} else if lit.Bool != nil {
		return *lit.Bool
	}
	return ""
}

// Generate call or selector
func (g *WGSLGenerator) generateCallOrSelect(cos *guixast.CallOrSelect) string {
	// Build base with fields
	base := cos.Base
	for _, field := range cos.Fields {
		base += "." + field
	}

	// If it has parentheses, it's a call
	if cos.HasParens {
		args := make([]string, len(cos.Args))
		for i, arg := range cos.Args {
			args[i] = g.generateExpression(arg)
		}
		return fmt.Sprintf("%s(%s)", MapFunctionToWGSL(base), strings.Join(args, ", "))
	}

	// Otherwise it's a selector
	return base
}

// Generate index expression
func (g *WGSLGenerator) generateIndexExpr(idx *guixast.IndexExpr) string {
	if idx.Slice != nil {
		// Slice expression - not fully supported in WGSL
		low := ""
		high := ""
		if idx.Slice.Low != nil {
			low = g.generateExpression(idx.Slice.Low)
		}
		if idx.Slice.High != nil {
			high = g.generateExpression(idx.Slice.High)
		}
		return fmt.Sprintf("%s[%s:%s]", idx.Base, low, high)
	} else if idx.Index != nil {
		return fmt.Sprintf("%s[%s]", idx.Base, g.generateExpression(idx.Index))
	}
	return idx.Base
}
