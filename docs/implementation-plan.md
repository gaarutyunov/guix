# Implementation Plan: Grammar & Visitor Refactoring

## Overview
This document outlines the step-by-step implementation plan for refactoring the Guix compiler to use a unified grammar and visitor pattern.

## Goals
1. ✅ Eliminate grammar ambiguity for expression statements
2. ✅ Enable support for arbitrary function calls in helper functions
3. ✅ Improve code maintainability and extensibility
4. ✅ Maintain backward compatibility during migration

## Timeline Estimate
- **Phase 1**: 2-3 days (Infrastructure)
- **Phase 2**: 2-3 days (Visitors)
- **Phase 3**: 3-4 days (Generator Migration)
- **Phase 4**: 2-3 days (Grammar Update)
- **Phase 5**: 1-2 days (Cleanup & Documentation)

**Total**: 10-15 days

---

## Phase 1: Visitor Infrastructure (Non-Breaking)

### Step 1.1: Create Visitor Package
**Files**: `pkg/visitor/visitor.go`, `pkg/visitor/base.go`

```go
// pkg/visitor/visitor.go
package visitor

import ast "github.com/gaarutyunov/guix/pkg/ast"

type Visitor interface {
    // Core nodes
    VisitFile(*ast.File) interface{}
    VisitComponent(*ast.Component) interface{}
    VisitHelperFunc(*ast.HelperFunc) interface{}

    // Structure nodes
    VisitBody(*ast.Body) interface{}
    VisitFuncBody(*ast.FuncBody) interface{}

    // Statement nodes
    VisitStatement(*ast.Statement) interface{}
    VisitBodyStatement(*ast.BodyStatement) interface{}
    VisitVarDecl(*ast.VarDecl) interface{}
    VisitAssignment(*ast.Assignment) interface{}
    VisitReturn(*ast.Return) interface{}
    VisitIfStmt(*ast.IfStmt) interface{}
    VisitElse(*ast.Else) interface{}
    VisitForLoop(*ast.ForLoop) interface{}

    // Expression nodes
    VisitExpr(*ast.Expr) interface{}
    VisitPrimary(*ast.Primary) interface{}
    VisitBinaryOp(*ast.BinaryOp) interface{}
    VisitUnaryExpr(*ast.UnaryExpr) interface{}
    VisitCallOrSelect(*ast.CallOrSelect) interface{}
    VisitFuncLit(*ast.FuncLit) interface{}
    VisitCompositeLit(*ast.CompositeLit) interface{}
    VisitLiteral(*ast.Literal) interface{}
    VisitMakeCall(*ast.MakeCall) interface{}
    VisitChannelOp(*ast.ChannelOp) interface{}

    // UI nodes
    VisitNode(*ast.Node) interface{}
    VisitElement(*ast.Element) interface{}
    VisitTextNode(*ast.TextNode) interface{}
    VisitTemplate(*ast.Template) interface{}
    VisitTemplateFragment(*ast.TemplateFragment) interface{}
    VisitProp(*ast.Prop) interface{}
}
```

**Tests**: `pkg/visitor/visitor_test.go`
```go
func TestVisitorInterface(t *testing.T) {
    // Ensure BaseVisitor implements Visitor
    var _ Visitor = (*BaseVisitor)(nil)
}
```

### Step 1.2: Add Accept Methods to AST
**Files**: `pkg/ast/visitor.go`

```go
// pkg/ast/visitor.go
package ast

// Visitor interface forward declaration
type Visitor interface {
    VisitFile(*File) interface{}
    VisitComponent(*Component) interface{}
    // ... (same as pkg/visitor/Visitor)
}

// Accept methods for all node types
func (n *File) Accept(v Visitor) interface{} {
    return v.VisitFile(n)
}

func (n *Component) Accept(v Visitor) interface{} {
    return v.VisitComponent(n)
}

func (n *Body) Accept(v Visitor) interface{} {
    return v.VisitBody(n)
}

func (n *Statement) Accept(v Visitor) interface{} {
    return v.VisitStatement(n)
}

func (n *Expr) Accept(v Visitor) interface{} {
    return v.VisitExpr(n)
}

// ... etc for all 30+ node types
```

**Tests**: `pkg/ast/visitor_test.go`
```go
type mockVisitor struct{}

func (m *mockVisitor) VisitFile(n *File) interface{} { return "file" }
// ... implement all methods

func TestAcceptMethods(t *testing.T) {
    visitor := &mockVisitor{}

    file := &File{Components: []*Component{}}
    result := file.Accept(visitor)

    assert.Equal(t, "file", result)
}
```

### Step 1.3: Implement BaseVisitor
**Files**: `pkg/visitor/base.go`

```go
// pkg/visitor/base.go
package visitor

import ast "github.com/gaarutyunov/guix/pkg/ast"

type BaseVisitor struct{}

func (v *BaseVisitor) VisitFile(n *ast.File) interface{} {
    for _, comp := range n.Components {
        comp.Accept(v)
    }
    for _, helper := range n.HelperFuncs {
        helper.Accept(v)
    }
    return nil
}

func (v *BaseVisitor) VisitComponent(n *ast.Component) interface{} {
    if n.Body != nil {
        n.Body.Accept(v)
    }
    return nil
}

func (v *BaseVisitor) VisitBody(n *ast.Body) interface{} {
    for _, decl := range n.VarDecls {
        decl.Accept(v)
    }
    for _, stmt := range n.Statements {
        stmt.Accept(v)
    }
    for _, child := range n.Children {
        child.Accept(v)
    }
    return nil
}

// ... implement all visitor methods with default traversal
```

**Tests**: `pkg/visitor/base_test.go`
```go
func TestBaseVisitorTraversal(t *testing.T) {
    visited := []string{}

    type trackingVisitor struct {
        BaseVisitor
    }

    tv := &trackingVisitor{}
    tv.VisitComponent = func(n *ast.Component) interface{} {
        visited = append(visited, "component:"+n.Name)
        return tv.BaseVisitor.VisitComponent(n)
    }

    comp := &ast.Component{Name: "TestComp", Body: &ast.Body{}}
    comp.Accept(tv)

    assert.Contains(t, visited, "component:TestComp")
}
```

### Step 1.4: Deliverables
- ✅ `pkg/visitor` package created
- ✅ All AST nodes have `Accept()` methods
- ✅ `BaseVisitor` provides default traversal
- ✅ Unit tests pass
- ✅ No changes to existing codegen (backward compatible)

---

## Phase 2: Specialized Visitors

### Step 2.1: Semantic Analyzer
**Files**: `pkg/visitor/semantic.go`

```go
package visitor

import (
    "fmt"
    ast "github.com/gaarutyunov/guix/pkg/ast"
)

type SemanticAnalyzer struct {
    BaseVisitor
    errors []error
    scopes []*Scope
}

type Scope struct {
    parent *Scope
    vars   map[string]*VarInfo
}

type VarInfo struct {
    name     string
    declType string
    used     bool
}

func NewSemanticAnalyzer() *SemanticAnalyzer {
    return &SemanticAnalyzer{
        errors: make([]error, 0),
        scopes: []*Scope{NewGlobalScope()},
    }
}

func (v *SemanticAnalyzer) pushScope() {
    v.scopes = append(v.scopes, NewScope(v.currentScope()))
}

func (v *SemanticAnalyzer) popScope() {
    v.scopes = v.scopes[:len(v.scopes)-1]
}

func (v *SemanticAnalyzer) currentScope() *Scope {
    return v.scopes[len(v.scopes)-1]
}

func (v *SemanticAnalyzer) VisitComponent(n *ast.Component) interface{} {
    v.pushScope()

    // Add parameters to scope
    for _, param := range n.Params {
        if !v.currentScope().define(param.Name, param.Type) {
            v.addError(n.Pos, "duplicate parameter: %s", param.Name)
        }
    }

    // Visit body
    if n.Body != nil {
        n.Body.Accept(v)
    }

    v.popScope()
    return nil
}

func (v *SemanticAnalyzer) VisitVarDecl(n *ast.VarDecl) interface{} {
    // Check RHS uses defined variables
    for _, val := range n.Values {
        v.checkExprUsesDefinedVars(val)
    }

    // Define variables
    for _, name := range n.Names {
        if !v.currentScope().define(name, "") {
            v.addError(n.Pos, "duplicate variable: %s", name)
        }
    }

    return nil
}

func (v *SemanticAnalyzer) Errors() []error {
    return v.errors
}

func (v *SemanticAnalyzer) addError(pos lexer.Position, format string, args ...interface{}) {
    v.errors = append(v.errors, fmt.Errorf(
        "%s:%d:%d: "+format,
        append([]interface{}{pos.Filename, pos.Line, pos.Column}, args...)...,
    ))
}
```

**Tests**: `pkg/visitor/semantic_test.go`

### Step 2.2: Debug Printer
**Files**: `pkg/visitor/debug.go`

```go
package visitor

import (
    "fmt"
    "strings"
    ast "github.com/gaarutyunov/guix/pkg/ast"
)

type DebugPrinter struct {
    BaseVisitor
    indent int
    output strings.Builder
}

func NewDebugPrinter() *DebugPrinter {
    return &DebugPrinter{indent: 0}
}

func (v *DebugPrinter) VisitComponent(n *ast.Component) interface{} {
    v.printf("Component: %s\n", n.Name)
    v.indent++
    v.BaseVisitor.VisitComponent(n)
    v.indent--
    return nil
}

func (v *DebugPrinter) VisitStatement(n *ast.Statement) interface{} {
    v.printf("Statement\n")
    v.indent++
    v.BaseVisitor.VisitStatement(n)
    v.indent--
    return nil
}

func (v *DebugPrinter) printf(format string, args ...interface{}) {
    v.output.WriteString(strings.Repeat("  ", v.indent))
    v.output.WriteString(fmt.Sprintf(format, args...))
}

func (v *DebugPrinter) String() string {
    return v.output.String()
}
```

**Usage**:
```go
printer := visitor.NewDebugPrinter()
file.Accept(printer)
fmt.Println(printer.String())
```

### Step 2.3: Deliverables
- ✅ `SemanticAnalyzer` validates AST
- ✅ `DebugPrinter` helps debugging
- ✅ Unit tests for each visitor
- ✅ Integration tests with example .gx files

---

## Phase 3: Migrate Generator to Visitor Pattern

### Step 3.1: Make Generator Implement Visitor
**Files**: `pkg/codegen/generator_visitor.go`

```go
package codegen

import (
    "go/ast"
    "go/token"
    guixast "github.com/gaarutyunov/guix/pkg/ast"
    "github.com/gaarutyunov/guix/pkg/visitor"
)

// Ensure Generator implements Visitor
var _ visitor.Visitor = (*Generator)(nil)

func (g *Generator) VisitFile(n *guixast.File) interface{} {
    // Collect component names
    g.components = make(map[string]bool)
    for _, comp := range n.Components {
        g.components[comp.Name] = true
    }

    // Generate imports
    imports := g.generateImports(n)

    // Generate declarations
    var decls []ast.Decl
    decls = append(decls, imports)

    // Visit components
    for _, comp := range n.Components {
        compDecls := comp.Accept(g).([]ast.Decl)
        decls = append(decls, compDecls...)
    }

    // Visit helper functions
    for _, helper := range n.HelperFuncs {
        helperDecl := helper.Accept(g).(*ast.FuncDecl)
        decls = append(decls, helperDecl)
    }

    return &ast.File{
        Name:  ast.NewIdent(g.pkg),
        Decls: decls,
    }
}

func (g *Generator) VisitComponent(n *guixast.Component) interface{} {
    // Setup context
    g.setupComponentContext(n)

    var decls []ast.Decl

    // Generate struct
    decls = append(decls, g.generateStruct(n))

    // Generate option functions
    if len(n.Params) > 0 {
        decls = append(decls, g.generateOptionType(n))
        decls = append(decls, g.generateOptionFunctions(n)...)
    }

    // Generate methods
    decls = append(decls, g.generateConstructor(n))
    decls = append(decls, g.generateBindApp(n))
    decls = append(decls, g.generateRenderMethod(n))
    decls = append(decls, g.generateMountMethod(n))
    decls = append(decls, g.generateUnmountMethod(n))
    decls = append(decls, g.generateUpdateMethod(n))

    return decls
}

func (g *Generator) VisitStatement(n *guixast.Statement) interface{} {
    if n.VarDecl != nil {
        return n.VarDecl.Accept(g)
    }
    if n.Assignment != nil {
        return n.Assignment.Accept(g)
    }
    if n.Return != nil {
        return n.Return.Accept(g)
    }
    if n.If != nil {
        return n.If.Accept(g)
    }
    if n.Expr != nil {
        return &ast.ExprStmt{
            X: n.Expr.Accept(g).(ast.Expr),
        }
    }
    return &ast.EmptyStmt{}
}

// Implement all other Visit* methods...
```

### Step 3.2: Update Main Generation Flow
**Files**: `pkg/codegen/codegen.go`

```go
func (g *Generator) Generate(file *guixast.File) ([]byte, error) {
    // Use visitor pattern
    goFile := file.Accept(g).(*ast.File)

    // Format
    var buf bytes.Buffer
    buf.WriteString("//go:build js && wasm\n")
    buf.WriteString("// +build js,wasm\n\n")
    buf.WriteString("// Code generated by guix. DO NOT EDIT.\n\n")

    if err := format.Node(&buf, g.fset, goFile); err != nil {
        return nil, fmt.Errorf("format error: %w", err)
    }

    return buf.Bytes(), nil
}
```

### Step 3.3: Deliverables
- ✅ Generator implements Visitor interface
- ✅ All generation uses Accept() calls
- ✅ Existing tests still pass
- ✅ Code generation output identical to before

---

## Phase 4: Grammar Update (Breaking Change)

### Step 4.1: Add ExpressionStatement Type
**Files**: `pkg/ast/ast.go`

```go
// Unified statement type
type ExpressionStatement struct {
    Pos   lexer.Position
    Left  *CallOrSelect         `@@`
    Op    string                `(@("<-" | ":=" | "=" | "+=" | "-=" | "*=" | "/="))?`
    Right *Expr                 `(@@)?`
}

// Update Statement
type Statement struct {
    Pos      lexer.Position
    VarDecl  *VarDecl              `@@`
    ExprStmt *ExpressionStatement  `| @@`
    Return   *Return               `| @@`
    If       *IfStmt               `| @@`
    For      *ForLoop              `| @@`
}

// Remove old Assignment type (breaking change)
```

### Step 4.2: Create Statement Disambiguator
**Files**: `pkg/visitor/disambiguate.go`

```go
package visitor

import (
    ast "github.com/gaarutyunov/guix/pkg/ast"
)

// StatementDisambiguator is no longer needed with new grammar!
// ExpressionStatement is self-disambiguating:
// - If Op is present → assignment
// - If Op is empty → expression

// But we keep this for validation
type StatementValidator struct {
    BaseVisitor
    errors []error
}

func (v *StatementValidator) VisitExpressionStatement(n *ast.ExpressionStatement) interface{} {
    if n.Op != "" {
        // Assignment - must have right side
        if n.Right == nil {
            v.errors = append(v.errors, fmt.Errorf(
                "%s:%d:%d: assignment missing right-hand side",
                n.Pos.Filename, n.Pos.Line, n.Pos.Column))
        }
    } else {
        // Expression - must be callable
        if n.Left.Args == nil && len(n.Left.Fields) == 0 {
            v.errors = append(v.errors, fmt.Errorf(
                "%s:%d:%d: statement has no effect",
                n.Pos.Filename, n.Pos.Line, n.Pos.Column))
        }
    }
    return nil
}
```

### Step 4.3: Update Generator
**Files**: `pkg/codegen/generator_visitor.go`

```go
func (g *Generator) VisitExpressionStatement(n *guixast.ExpressionStatement) interface{} {
    if n.Op != "" {
        // Generate assignment
        lhs := g.generateCallOrSelect(n.Left)
        rhs := n.Right.Accept(g).(ast.Expr)

        if n.Op == "<-" {
            return &ast.SendStmt{
                Chan:  lhs,
                Value: rhs,
            }
        }

        return &ast.AssignStmt{
            Lhs: []ast.Expr{lhs},
            Tok: g.assignOpToToken(n.Op),
            Rhs: []ast.Expr{rhs},
        }
    }

    // Generate expression statement
    return &ast.ExprStmt{
        X: g.generateCallOrSelect(n.Left),
    }
}
```

### Step 4.4: Update Parser
Update Participle grammar to use new `ExpressionStatement` type.

### Step 4.5: Deliverables
- ✅ New grammar parses successfully
- ✅ All examples regenerate correctly
- ✅ Function calls in helpers work: `log(fmt.Sprintf(...))`
- ✅ Field assignments work: `state.Display = value`
- ✅ All tests pass

---

## Phase 5: Cleanup & Documentation

### Step 5.1: Remove Old Code
- Remove old `generateStatement()` methods
- Remove old `Assignment` type references
- Update all comments

### Step 5.2: Documentation
**Files**:
- `docs/architecture.md` - Visitor pattern overview
- `docs/adding-features.md` - How to add new features
- `docs/visitors.md` - Available visitors and their purposes

### Step 5.3: Examples
**Files**: `examples/visitors/`

```go
// examples/visitors/count_components.go
// Example: Count components in a file
type ComponentCounter struct {
    visitor.BaseVisitor
    count int
}

func (v *ComponentCounter) VisitComponent(n *ast.Component) interface{} {
    v.count++
    return v.BaseVisitor.VisitComponent(n)
}

// Usage:
counter := &ComponentCounter{}
file.Accept(counter)
fmt.Printf("Found %d components\n", counter.count)
```

### Step 5.4: Deliverables
- ✅ Clean, maintainable codebase
- ✅ Comprehensive documentation
- ✅ Examples of extending with visitors
- ✅ All tests passing

---

## Testing Strategy

### Unit Tests
- Test each visitor independently
- Test AST node Accept() methods
- Test grammar parsing edge cases

### Integration Tests
- Regenerate all examples
- Compare output with previous version
- Test new functionality (function calls)

### Regression Tests
- Keep existing test suite
- Ensure backward compatibility
- Test error messages

---

## Success Criteria

- ✅ Function calls work in helper functions: `log(...)`
- ✅ Method calls work: `console.Call(...)`
- ✅ Field assignments work: `state.field = value`
- ✅ No grammar ambiguity errors
- ✅ All existing tests pass
- ✅ New visitor-based architecture in place
- ✅ Documentation complete

---

## Risk Mitigation

### Risk: Breaking Changes
**Mitigation**:
- Implement visitors first (non-breaking)
- Migrate gradually
- Keep old code until new code proven

### Risk: Performance Regression
**Mitigation**:
- Benchmark before/after
- Profile code generation
- Optimize hot paths if needed

### Risk: Complexity Increase
**Mitigation**:
- Comprehensive documentation
- Clear examples
- Code reviews at each phase

---

## Next Steps

1. Review this plan with team
2. Get approval for breaking changes
3. Create feature branch
4. Begin Phase 1 implementation
5. Review after each phase
