# Guix Compiler Refactoring Proposal

## Executive Summary

This proposal outlines a comprehensive refactoring of the Guix compiler to:
1. **Eliminate grammar ambiguity** preventing arbitrary function calls in .gx files
2. **Introduce visitor pattern** for better code organization and extensibility
3. **Enable new features** while maintaining backward compatibility

## Problem Statement

### Current Issue
The parser cannot distinguish between:
- Function calls: `log(...)`
- Method calls: `console.Call(...)`
- Field assignments: `state.Display = value`
- UI elements: `Div(...) { ... }`

All start with the same pattern (`Ident ("." Ident)* ("(" ... ")")?`), causing parse errors.

### Impact
- ❌ Cannot add logging to helper functions
- ❌ Grammar is brittle and hard to extend
- ❌ Code generation is tightly coupled to AST structure
- ❌ Difficult to add preprocessing/validation passes

## Proposed Solution

### Two-Pronged Approach

#### 1. Grammar Restructuring
**Principle**: Parse liberally, disambiguate later

**Key Change**: Introduce `ExpressionStatement` that unifies expressions and assignments:

```go
type ExpressionStatement struct {
    Left  *CallOrSelect         `@@`
    Op    string                `(@("=" | ":=" | ...))?`  // Optional!
    Right *Expr                 `(@@)?`                   // Optional!
}
```

**Disambiguation**:
- If `Op` is empty → Expression statement
- If `Op` is present → Assignment statement

**Benefits**:
- ✅ No parser ambiguity
- ✅ Simpler grammar
- ✅ Supports all statement types

#### 2. Visitor Pattern
**Principle**: Separate concerns into composable passes

**Architecture**:
```
Parser → AST
   ↓
Semantic Analyzer (Visitor) → Validated AST
   ↓
Statement Validator (Visitor) → Transformed AST
   ↓
Code Generator (Visitor) → Go AST → Code
```

**Benefits**:
- ✅ Modular, testable components
- ✅ Easy to add new passes (optimization, etc.)
- ✅ Clean separation of concerns
- ✅ Maintainable codebase

## Detailed Proposals

### 1. Grammar Refactoring
See: [`grammar-refactor-proposal.md`](./grammar-refactor-proposal.md)

**Key Points**:
- Unified `ExpressionStatement` type
- Self-disambiguating via optional operator
- Backward compatible during transition
- Examples of parsing different statement types

### 2. Visitor Pattern Architecture
See: [`visitor-pattern-proposal.md`](./visitor-pattern-proposal.md)

**Key Points**:
- Visitor interface definition
- BaseVisitor for default traversal
- Specialized visitors (SemanticAnalyzer, CodeGenerator)
- Migration strategy
- Examples of extending with new visitors

### 3. Implementation Plan
See: [`implementation-plan.md`](./implementation-plan.md)

**Key Points**:
- 5 phases over 10-15 days
- Non-breaking initial phases
- Comprehensive testing strategy
- Risk mitigation
- Success criteria

## Benefits

### Immediate
- ✅ **Function calls work**: `log(fmt.Sprintf(...))`
- ✅ **Method calls work**: `console.Call("log", ...)`
- ✅ **Field assignments work**: `state.Display = value`

### Long-term
- ✅ **Extensibility**: Easy to add new language features
- ✅ **Maintainability**: Cleaner, modular codebase
- ✅ **Testability**: Each visitor tested independently
- ✅ **Debugging**: Debug printer visitor for AST inspection
- ✅ **Optimization**: Easy to add optimization passes
- ✅ **Validation**: Separate semantic analysis pass
- ✅ **Documentation**: Self-documenting visitor architecture

## Examples

### Before (Doesn't Work)
```go
func handleNumber(...) {
    log(fmt.Sprintf("Debug: %v", value))  // ❌ Parse error
    state.Display = digit                  // ❌ Ambiguity error
}
```

### After (Works!)
```go
func handleNumber(...) {
    log(fmt.Sprintf("Debug: %v", value))  // ✅ Expression statement
    state.Display = digit                  // ✅ Assignment
    state.Counter = state.Counter + 1      // ✅ Complex assignment
}
```

## Migration Path

### Phase 1: Infrastructure (Non-Breaking)
- Add visitor interfaces
- Add Accept() methods to AST
- Create BaseVisitor
- **No changes to existing code**

### Phase 2: Visitors (Non-Breaking)
- Implement SemanticAnalyzer
- Implement DebugPrinter
- Test independently
- **No changes to codegen yet**

### Phase 3: Migrate Generator (Non-Breaking)
- Make Generator implement Visitor
- Refactor using Accept() calls
- **Output remains identical**

### Phase 4: Grammar Update (Breaking)
- Introduce ExpressionStatement
- Update parser
- Update generator
- **New functionality enabled**

### Phase 5: Cleanup
- Remove old code
- Update documentation
- Ship!

## Testing Strategy

### Unit Tests
- Each visitor tested independently
- AST Accept() methods verified
- Grammar parsing edge cases

### Integration Tests
- Regenerate all examples
- Compare with previous output
- Test new function call support

### Regression Tests
- All existing tests must pass
- Backward compatibility verified

## Timeline

- **Phase 1**: 2-3 days
- **Phase 2**: 2-3 days
- **Phase 3**: 3-4 days
- **Phase 4**: 2-3 days
- **Phase 5**: 1-2 days

**Total**: 10-15 days

## Success Criteria

- ✅ Function calls work in helper functions
- ✅ Method calls work (console.Call, etc.)
- ✅ Field assignments work (state.field = value)
- ✅ All existing examples regenerate correctly
- ✅ All tests pass
- ✅ Documentation complete
- ✅ Code is more maintainable than before

## Risks & Mitigation

### Risk: Breaking Changes
**Mitigation**: Gradual migration, keep old code until proven

### Risk: Performance
**Mitigation**: Benchmark before/after, optimize if needed

### Risk: Complexity
**Mitigation**: Comprehensive docs, clear examples, code reviews

## Comparison with Current Approach

### Current (Type Switches)
```go
func generateStatement(stmt *Statement) ast.Stmt {
    if stmt.Assignment != nil {
        // 50 lines of assignment logic
    }
    if stmt.Expr != nil {
        // 30 lines of expr logic
    }
    if stmt.If != nil {
        // 40 lines of if logic
    }
    // ... 10 more type checks
}
```

**Issues**:
- Monolithic function
- Hard to test
- Tightly coupled
- Difficult to extend

### Proposed (Visitor Pattern)
```go
func (g *Generator) VisitExpressionStatement(n *ExprStmt) interface{} {
    if n.Op != "" {
        return g.generateAssignment(n)
    }
    return g.generateExpression(n)
}

func (g *Generator) VisitIfStmt(n *IfStmt) interface{} {
    return g.generateIf(n)
}
```

**Benefits**:
- ✅ Each method focused on one thing
- ✅ Easy to test
- ✅ Loosely coupled
- ✅ Easy to extend

## Future Possibilities

With visitor pattern in place, we can easily add:

### Optimization Pass
```go
type Optimizer struct {
    BaseVisitor
}

func (o *Optimizer) VisitExpr(n *Expr) interface{} {
    // Constant folding
    // Dead code elimination
    // etc.
}
```

### Type Inference
```go
type TypeInferencer struct {
    BaseVisitor
    types map[*ast.Node]Type
}
```

### Documentation Generator
```go
type DocGenerator struct {
    BaseVisitor
}

func (d *DocGenerator) VisitComponent(n *Component) interface{} {
    // Generate markdown docs
}
```

### Linter
```go
type Linter struct {
    BaseVisitor
    warnings []Warning
}
```

## Conclusion

This refactoring provides:
1. **Immediate value**: Solves function call ambiguity
2. **Long-term value**: Better architecture for future growth
3. **Low risk**: Gradual migration with testing at each phase
4. **High reward**: More maintainable, extensible codebase

## Recommendation

**Proceed with implementation** following the phased approach outlined in [`implementation-plan.md`](./implementation-plan.md).

## Questions?

- Grammar details → [`grammar-refactor-proposal.md`](./grammar-refactor-proposal.md)
- Visitor architecture → [`visitor-pattern-proposal.md`](./visitor-pattern-proposal.md)
- Implementation steps → [`implementation-plan.md`](./implementation-plan.md)

## Appendix: Quick Reference

### Current State
- ❌ Function calls don't parse
- ❌ Grammar is ambiguous
- ❌ Codegen is monolithic
- ❌ Hard to add features

### After Refactoring
- ✅ Function calls work
- ✅ Grammar is clean
- ✅ Codegen is modular
- ✅ Easy to extend

### What Users Get
```go
// Calculator example with logging
func handleOperator(state State, op string) {
    log(fmt.Sprintf("Operator: %s, State: %+v", op, state))

    if state.Operator != "" {
        result := calculate(state)
        state.Display = formatNumber(result)
        log(fmt.Sprintf("Result: %f", result))
    }

    state.Operator = op
}
```

**All of this just works!** 🎉
