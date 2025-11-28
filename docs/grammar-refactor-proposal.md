# Grammar Refactoring Proposal

## Core Principle: Parse Liberally, Validate Strictly

Instead of forcing the parser to distinguish between statement types, parse a unified structure and disambiguate during AST transformation.

## Proposed AST Changes

### Current (Problematic)
```go
type Statement struct {
    VarDecl        *VarDecl
    Assignment     *Assignment      // x = y
    AssignmentExpr *AssignmentExpr  // x.y = z
    Expr           *Expr            // f() or x.y()
    Return         *Return
    If             *IfStmt
}

type Assignment struct {
    Left  string
    Op    string
    Right *Expr
}

type AssignmentExpr struct {
    Left  *CallOrSelect
    Op    string
    Right *Expr
}
```

**Problem**: Parser tries CallOrSelect for both AssignmentExpr and Expr, commits to wrong one.

### Proposed (Unified)
```go
type Statement struct {
    VarDecl  *VarDecl
    ExprStmt *ExpressionStatement  // Unified: handles both expr and assignment
    Return   *Return
    If       *IfStmt
    For      *ForLoop
}

// Unified statement that can be expr or assignment
type ExpressionStatement struct {
    Pos   lexer.Position
    Left  *CallOrSelect         `@@`
    Op    string                `(@("<-" | ":=" | "=" | "+=" | "-=" | "*=" | "/="))?`
    Right *Expr                 `(@@)?`
}
```

**Key Change**: `Op` and `Right` are optional. Parser always succeeds matching `CallOrSelect`, then optionally matches operator and right-hand side.

**Disambiguation**:
- If `Op` is empty: Expression statement (function call)
- If `Op` is present: Assignment statement

## BodyStatement Similar Pattern

```go
type BodyStatement struct {
    VarDecl  *VarDecl
    ExprStmt *ExpressionStatement
    Return   *Return
    If       *IfStmt
}
```

Same unified approach for component bodies.

## Benefits

1. **No Ambiguity**: Parser always knows what to match
2. **Simpler Grammar**: Fewer alternations, less backtracking
3. **Deferred Disambiguation**: Handled in well-defined transformation phase
4. **Extensible**: Easy to add new operators or statement types

## Migration Path

1. Update AST types (breaking change)
2. Update parser grammar
3. Add transformation visitor to convert ExpressionStatement to appropriate Go AST node
4. Update codegen to handle unified type
5. Test with existing examples

## Example Parsing

### Input: `console.Call("log", msg)`
```go
ExpressionStatement{
    Left: CallOrSelect{
        Base: "console",
        Fields: ["Call"],
        Args: [Expr{...}, Expr{...}],
    },
    Op: "",      // Empty!
    Right: nil,
}
```
→ Transformed to: Expression statement

### Input: `state.Display = "5"`
```go
ExpressionStatement{
    Left: CallOrSelect{
        Base: "state",
        Fields: ["Display"],
        Args: nil,
    },
    Op: "=",
    Right: Expr{...},
}
```
→ Transformed to: Assignment statement

### Input: `log("message")`
```go
ExpressionStatement{
    Left: CallOrSelect{
        Base: "log",
        Fields: [],
        Args: [Expr{...}],
    },
    Op: "",
    Right: nil,
}
```
→ Transformed to: Expression statement
