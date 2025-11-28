# Visitor Pattern Refactoring Proposal

## Current Architecture (Direct AST Walking)

**Problem**: Code generation logic is tightly coupled with AST traversal.

```go
func (g *Generator) generateStatement(stmt *Statement) ast.Stmt {
    if stmt.Assignment != nil {
        // Assignment logic here...
    }
    if stmt.Expr != nil {
        // Expression logic here...
    }
    if stmt.If != nil {
        // If logic here...
    }
    // ... 15 more type checks
}
```

**Issues**:
1. ❌ Hard to add preprocessing (validation, transformation)
2. ❌ Can't compose multiple passes easily
3. ❌ Generator class becomes monolithic
4. ❌ Difficult to test individual transformations
5. ❌ No clear separation between traversal and logic

## Proposed Architecture (Visitor Pattern)

### Core Visitor Interface

```go
// pkg/visitor/visitor.go
package visitor

import (
    "github.com/gaarutyunov/guix/pkg/ast"
    "go/ast" as goast
)

// Visitor interface - one method per AST node type
type Visitor interface {
    VisitFile(*ast.File) interface{}
    VisitComponent(*ast.Component) interface{}
    VisitBody(*ast.Body) interface{}
    VisitStatement(*ast.Statement) interface{}
    VisitBodyStatement(*ast.BodyStatement) interface{}
    VisitExpr(*ast.Expr) interface{}
    VisitPrimary(*ast.Primary) interface{}
    VisitNode(*ast.Node) interface{}
    VisitElement(*ast.Element) interface{}
    VisitTemplate(*ast.Template) interface{}
    VisitVarDecl(*ast.VarDecl) interface{}
    VisitAssignment(*ast.Assignment) interface{}
    VisitReturn(*ast.Return) interface{}
    VisitIfStmt(*ast.IfStmt) interface{}
    VisitForLoop(*ast.ForLoop) interface{}
}
```

### Add Accept Methods to AST Nodes

```go
// pkg/ast/visitor.go
package ast

import "github.com/gaarutyunov/guix/pkg/visitor"

// Accept methods for each node type
func (f *File) Accept(v visitor.Visitor) interface{} {
    return v.VisitFile(f)
}

func (c *Component) Accept(v visitor.Visitor) interface{} {
    return v.VisitComponent(c)
}

func (s *Statement) Accept(v visitor.Visitor) interface{} {
    return v.VisitStatement(s)
}

// ... etc for all node types
```

### Base Visitor (Default Traversal)

```go
// pkg/visitor/base.go
package visitor

import "github.com/gaarutyunov/guix/pkg/ast"

// BaseVisitor provides default traversal behavior
// Other visitors can embed this and override specific methods
type BaseVisitor struct{}

func (v *BaseVisitor) VisitFile(node *ast.File) interface{} {
    for _, comp := range node.Components {
        comp.Accept(v)
    }
    for _, helper := range node.HelperFuncs {
        helper.Accept(v)
    }
    return nil
}

func (v *BaseVisitor) VisitComponent(node *ast.Component) interface{} {
    if node.Body != nil {
        node.Body.Accept(v)
    }
    return nil
}

func (v *BaseVisitor) VisitBody(node *ast.Body) interface{} {
    for _, decl := range node.VarDecls {
        decl.Accept(v)
    }
    for _, stmt := range node.Statements {
        stmt.Accept(v)
    }
    for _, child := range node.Children {
        child.Accept(v)
    }
    return nil
}

func (v *BaseVisitor) VisitStatement(node *ast.Statement) interface{} {
    if node.VarDecl != nil {
        return node.VarDecl.Accept(v)
    }
    if node.ExprStmt != nil {
        return node.ExprStmt.Accept(v)
    }
    if node.Return != nil {
        return node.Return.Accept(v)
    }
    if node.If != nil {
        return node.If.Accept(v)
    }
    return nil
}

// ... default implementations for all node types
```

## Specialized Visitors

### 1. Statement Disambiguator (Transformation Pass)

```go
// pkg/visitor/disambiguate.go
package visitor

import (
    "github.com/gaarutyunov/guix/pkg/ast"
)

// StatementDisambiguator transforms ambiguous ExpressionStatement nodes
// into properly typed nodes (Assignment or Expression)
type StatementDisambiguator struct {
    BaseVisitor
}

func NewStatementDisambiguator() *StatementDisambiguator {
    return &StatementDisambiguator{}
}

func (v *StatementDisambiguator) VisitExpressionStatement(node *ast.ExpressionStatement) interface{} {
    if node.Op != "" {
        // Has operator → it's an assignment
        return &ast.Assignment{
            Left:  node.Left,
            Op:    node.Op,
            Right: node.Right,
        }
    }

    // No operator → it's an expression
    return &ast.Expression{
        Value: node.Left,
    }
}
```

### 2. Semantic Analyzer (Validation Pass)

```go
// pkg/visitor/semantic.go
package visitor

import (
    "fmt"
    "github.com/gaarutyunov/guix/pkg/ast"
)

// SemanticAnalyzer validates the AST for semantic errors
type SemanticAnalyzer struct {
    BaseVisitor
    errors []error
    scope  *Scope  // Track variable scopes
}

func NewSemanticAnalyzer() *SemanticAnalyzer {
    return &SemanticAnalyzer{
        errors: make([]error, 0),
        scope:  NewScope(nil),
    }
}

func (v *SemanticAnalyzer) VisitComponent(node *ast.Component) interface{} {
    // Enter component scope
    v.scope = NewScope(v.scope)

    // Add component parameters to scope
    for _, param := range node.Params {
        if !v.scope.Define(param.Name, param.Type) {
            v.errors = append(v.errors, fmt.Errorf(
                "duplicate parameter: %s", param.Name))
        }
    }

    // Visit body
    if node.Body != nil {
        node.Body.Accept(v)
    }

    // Exit component scope
    v.scope = v.scope.Parent()

    return nil
}

func (v *SemanticAnalyzer) VisitVarDecl(node *ast.VarDecl) interface{} {
    // Check for undefined variables in RHS
    for _, val := range node.Values {
        v.checkExprVariables(val)
    }

    // Add to scope
    for _, name := range node.Names {
        if !v.scope.Define(name, nil) {
            v.errors = append(v.errors, fmt.Errorf(
                "duplicate variable: %s", name))
        }
    }

    return nil
}

func (v *SemanticAnalyzer) Errors() []error {
    return v.errors
}
```

### 3. Code Generator (Final Pass)

```go
// pkg/codegen/generator_visitor.go
package codegen

import (
    "go/ast"
    "go/token"

    guixast "github.com/gaarutyunov/guix/pkg/ast"
    "github.com/gaarutyunov/guix/pkg/visitor"
)

// Generator now implements Visitor interface
type Generator struct {
    visitor.BaseVisitor

    fset                *token.FileSet
    pkg                 string
    components          map[string]bool
    hoistedVars         map[string]bool
    channelReceiveVars  map[string]string
    componentParams     map[string]bool
    verbose             bool

    // Output
    currentFile     *ast.File
    currentDecls    []ast.Decl
}

func (g *Generator) VisitFile(node *guixast.File) interface{} {
    g.components = make(map[string]bool)

    // Collect component names
    for _, comp := range node.Components {
        g.components[comp.Name] = true
    }

    // Generate imports
    imports := g.generateImports(node)

    // Generate components
    var decls []ast.Decl
    decls = append(decls, imports)

    for _, comp := range node.Components {
        compDecls := comp.Accept(g).([]ast.Decl)
        decls = append(decls, compDecls...)
    }

    // Generate helper functions
    for _, helper := range node.HelperFuncs {
        helperDecl := helper.Accept(g).(*ast.FuncDecl)
        decls = append(decls, helperDecl)
    }

    return &ast.File{
        Name:  ast.NewIdent(g.pkg),
        Decls: decls,
    }
}

func (g *Generator) VisitComponent(node *guixast.Component) interface{} {
    // Setup component context
    g.hoistedVars = make(map[string]bool)
    g.channelReceiveVars = make(map[string]string)
    g.componentParams = make(map[string]bool)

    for _, param := range node.Params {
        g.componentParams[param.Name] = true
    }

    // Collect hoisted vars and channel receives
    g.analyzeComponentVars(node)

    var decls []ast.Decl

    // Generate struct
    decls = append(decls, g.generateStruct(node))

    // Generate option type and functions
    if len(node.Params) > 0 {
        decls = append(decls, g.generateOptionType(node))
        decls = append(decls, g.generateOptionFunctions(node)...)
    }

    // Generate constructor
    decls = append(decls, g.generateConstructor(node))

    // Generate methods
    decls = append(decls, g.generateBindApp(node))
    decls = append(decls, g.generateRenderMethod(node))
    decls = append(decls, g.generateMountMethod(node))
    decls = append(decls, g.generateUnmountMethod(node))
    decls = append(decls, g.generateUpdateMethod(node))

    return decls
}

func (g *Generator) VisitStatement(node *guixast.Statement) interface{} {
    if node.VarDecl != nil {
        return node.VarDecl.Accept(g)
    }

    if node.ExprStmt != nil {
        exprStmt := node.ExprStmt.Accept(g)
        return exprStmt
    }

    if node.Return != nil {
        return node.Return.Accept(g)
    }

    if node.If != nil {
        return node.If.Accept(g)
    }

    return &ast.EmptyStmt{}
}

func (g *Generator) VisitExpressionStatement(node *guixast.ExpressionStatement) interface{} {
    if node.Op != "" {
        // Assignment
        return &ast.AssignStmt{
            Lhs: []ast.Expr{g.generateCallOrSelect(node.Left)},
            Tok: g.assignOpToToken(node.Op),
            Rhs: []ast.Expr{node.Right.Accept(g).(ast.Expr)},
        }
    }

    // Expression statement
    return &ast.ExprStmt{
        X: g.generateCallOrSelect(node.Left),
    }
}

// Helper methods remain similar but call Accept() instead of direct recursion
```

## Pipeline Architecture

### Main Generation Flow

```go
// cmd/guix/generate.go
func generateCode(inputFile string) ([]byte, error) {
    // Parse
    file, err := parser.Parse(inputFile)
    if err != nil {
        return nil, err
    }

    // Pass 1: Semantic analysis
    analyzer := visitor.NewSemanticAnalyzer()
    file.Accept(analyzer)
    if len(analyzer.Errors()) > 0 {
        return nil, analyzer.Errors()[0]
    }

    // Pass 2: Code generation
    generator := codegen.NewGenerator("main")
    goFile := file.Accept(generator).(*ast.File)

    // Format and return
    return format.Source(goFile)
}
```

## Benefits of This Architecture

### 1. **Separation of Concerns**
- ✅ Parsing: Grammar handles syntax
- ✅ Validation: Semantic analyzer checks correctness
- ✅ Transformation: Disambiguator resolves ambiguities
- ✅ Generation: Generator produces Go code

### 2. **Extensibility**
```go
// Easy to add new passes
type TypeInferenceVisitor struct { ... }
type OptimizationVisitor struct { ... }
type DocumentationGenerator struct { ... }

// Chain them
file.Accept(semanticAnalyzer)
file.Accept(typeInference)
file.Accept(optimizer)
file.Accept(codeGenerator)
```

### 3. **Testability**
```go
func TestSemanticAnalyzer(t *testing.T) {
    ast := &File{...}
    analyzer := NewSemanticAnalyzer()
    ast.Accept(analyzer)
    assert.Empty(t, analyzer.Errors())
}

func TestCodeGenerator(t *testing.T) {
    ast := &Component{...}
    gen := NewGenerator("main")
    result := ast.Accept(gen)
    assert.NotNil(t, result)
}
```

### 4. **Debugging**
```go
// Debug visitor to print AST
type DebugPrinter struct {
    BaseVisitor
    indent int
}

func (v *DebugPrinter) VisitComponent(node *Component) interface{} {
    fmt.Printf("%sComponent: %s\n", strings.Repeat("  ", v.indent), node.Name)
    v.indent++
    v.BaseVisitor.VisitComponent(node)
    v.indent--
    return nil
}

// Use: file.Accept(NewDebugPrinter())
```

### 5. **Composition**
```go
// Combine multiple transformations
type CompositeVisitor struct {
    visitors []Visitor
}

func (v *CompositeVisitor) VisitFile(node *File) interface{} {
    for _, visitor := range v.visitors {
        node.Accept(visitor)
    }
    return nil
}

// Use:
composite := &CompositeVisitor{
    visitors: []Visitor{
        NewSemanticAnalyzer(),
        NewOptimizer(),
        NewCodeGenerator(),
    },
}
file.Accept(composite)
```

## Migration Strategy

### Phase 1: Add Visitor Infrastructure (Non-Breaking)
1. Create `pkg/visitor` package with interfaces
2. Add `Accept()` methods to AST nodes
3. Create `BaseVisitor` with default traversal
4. No changes to existing code

### Phase 2: Create Specialized Visitors
1. Implement `SemanticAnalyzer`
2. Implement `StatementDisambiguator`
3. Test separately

### Phase 3: Migrate Generator
1. Make `Generator` implement `Visitor`
2. Refactor methods to use `Accept()` calls
3. Keep old methods as helpers temporarily
4. Test thoroughly

### Phase 4: Update Grammar (Breaking)
1. Introduce `ExpressionStatement`
2. Remove separate `Assignment`/`AssignmentExpr`
3. Update parser
4. Add disambiguation pass
5. Integration testing

### Phase 5: Cleanup
1. Remove old generator methods
2. Update documentation
3. Add visitor examples

## Example: Adding a New Feature

**Feature**: Add support for switch statements

### Without Visitor (Current)
- ❌ Modify parser grammar
- ❌ Add to multiple AST types
- ❌ Add to `generateStatement()`
- ❌ Add to `generateBodyStatement()`
- ❌ Update all switch statements checking node types
- ❌ Hard to test in isolation

### With Visitor (Proposed)
```go
// 1. Add to AST
type SwitchStmt struct {
    Expr  *Expr
    Cases []*CaseClause
}

// 2. Add to visitor interface
type Visitor interface {
    // ... existing methods
    VisitSwitchStmt(*ast.SwitchStmt) interface{}
}

// 3. Implement in generator
func (g *Generator) VisitSwitchStmt(node *ast.SwitchStmt) interface{} {
    cases := make([]ast.Stmt, len(node.Cases))
    for i, c := range node.Cases {
        cases[i] = c.Accept(g).(ast.Stmt)
    }

    return &ast.SwitchStmt{
        Tag:  node.Expr.Accept(g).(ast.Expr),
        Body: &ast.BlockStmt{List: cases},
    }
}

// Done! BaseVisitor handles traversal automatically
```
