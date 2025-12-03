// Package ast defines GPU-specific AST nodes for WGSL code generation
package ast

import "github.com/alecthomas/participle/v2/lexer"

// GPUDecorator represents a GPU-specific decorator like @gpu, @vertex, @binding, etc.
type GPUDecorator struct {
	Pos  lexer.Position
	Name string  `"@" @("gpu" | "vertex" | "fragment" | "compute" | "uniform" | "storage" | "binding" | "location" | "builtin" | "workgroup")`
	Args []*Expr `("(" (@@ ("," @@)*)? ")")?`
}

// GPUStructDecl represents a struct marked with @gpu decorator
type GPUStructDecl struct {
	Pos        lexer.Position
	Decorators []*GPUDecorator `@@*`
	Name       string          `"type" @Ident`
	Struct     *GPUStructType  `@@`
}

// GPUStructType represents a GPU struct type
type GPUStructType struct {
	Pos    lexer.Position
	Fields []*GPUField `"struct" "{" @@* "}"`
}

// GPUField represents a field in a GPU struct
type GPUField struct {
	Pos        lexer.Position
	Decorators []*GPUDecorator `@@*`
	Name       string          `@Ident`
	Type       *GPUType        `@@`
}

// GPUType represents a GPU type (vec2, vec3, vec4, mat4, etc.)
type GPUType struct {
	Pos       lexer.Position
	IsSlice   bool   `@("[" "]")?`
	IsPointer bool   `@("*")?`
	Name      string `@Ident`
	Generic   *Type  `("[" @@ "]")?` // For array sizes or generic types
}

// GPUBindingDecl represents a binding declaration
// Example: @binding(0, 0) @uniform var uniforms ChartUniforms
type GPUBindingDecl struct {
	Pos         lexer.Position
	Decorators  []*GPUDecorator `@@+`
	Name        string          `"var" @Ident`
	Type        *GPUType        `@@`
	InitialExpr *Expr           `("=" @@)?`
}

// GPUFuncDecl represents a GPU shader function
type GPUFuncDecl struct {
	Pos        lexer.Position
	Decorators []*GPUDecorator `@@*`
	Name       string          `"func" @Ident`
	Params     []*GPUParameter `"(" (@@ ("," @@)*)? ")"`
	Results    *GPUReturnType  `@@?`
	Body       *Body           `@@`
}

// GPUParameter represents a parameter in a GPU function
type GPUParameter struct {
	Pos        lexer.Position
	Decorators []*GPUDecorator `@@*`
	Name       string          `@Ident`
	Type       *GPUType        `@@`
}

// GPUReturnType represents a return type for GPU functions
type GPUReturnType struct {
	Pos        lexer.Position
	Decorators []*GPUDecorator `@@*`
	Type       *GPUType        `@@`
}

// ShaderStage represents the shader stage (vertex, fragment, compute)
type ShaderStage int

const (
	ShaderStageVertex ShaderStage = iota
	ShaderStageFragment
	ShaderStageCompute
)

// AddressSpace represents the WGSL address space
type AddressSpace int

const (
	AddressSpaceUniform AddressSpace = iota
	AddressSpaceStorage
	AddressSpacePrivate
	AddressSpaceFunction
)

// AccessMode represents the WGSL access mode
type AccessMode int

const (
	AccessModeRead AccessMode = iota
	AccessModeWrite
	AccessModeReadWrite
)

// Accept methods for visitor pattern

func (n *GPUDecorator) Accept(v Visitor) interface{}   { return v.VisitGPUDecorator(n) }
func (n *GPUStructDecl) Accept(v Visitor) interface{}  { return v.VisitGPUStructDecl(n) }
func (n *GPUStructType) Accept(v Visitor) interface{}  { return v.VisitGPUStructType(n) }
func (n *GPUField) Accept(v Visitor) interface{}       { return v.VisitGPUField(n) }
func (n *GPUType) Accept(v Visitor) interface{}        { return v.VisitGPUType(n) }
func (n *GPUBindingDecl) Accept(v Visitor) interface{} { return v.VisitGPUBindingDecl(n) }
func (n *GPUFuncDecl) Accept(v Visitor) interface{}    { return v.VisitGPUFuncDecl(n) }
func (n *GPUParameter) Accept(v Visitor) interface{}   { return v.VisitGPUParameter(n) }
func (n *GPUReturnType) Accept(v Visitor) interface{}  { return v.VisitGPUReturnType(n) }
