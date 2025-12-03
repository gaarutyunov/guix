# WGSL Code Generation in Guix

## Overview

Guix now supports WGSL (WebGPU Shader Language) code generation alongside Go/WASM output. This allows developers to write GPU shaders using the same Guix DSL used for application logic, with automatic type synchronization between CPU and GPU code.

## Implementation Status

### ✅ Completed Features

1. **GPU AST Extensions** (`pkg/ast/gpu_ast.go`)
   - `GPUDecorator` - Represents GPU decorators (@gpu, @vertex, etc.)
   - `GPUStructDecl` - GPU struct declarations
   - `GPUBindingDecl` - Resource binding declarations
   - `GPUFuncDecl` - Shader function declarations
   - `GPUParameter` - Shader function parameters
   - `GPUReturnType` - Shader function return types

2. **Type Mapping System** (`pkg/codegen/gpu_types.go`)
   - Guix/Go types → WGSL types mapping
   - Scalar types: `float32` → `f32`, `int32` → `i32`, etc.
   - Vector types: `vec2`, `vec3`, `vec4` with proper alignment
   - Matrix types: `mat2`, `mat3`, `mat4` with column-major layout
   - Array types: `[]T` → `array<T>`, `[N]T` → `array<T, N>`
   - Automatic memory layout calculation
   - Padding field generation for proper alignment

3. **WGSL Code Generator** (`pkg/codegen/wgsl_generator.go`)
   - Generates WGSL shader code from GPU AST nodes
   - Supports struct declarations
   - Supports binding declarations with address spaces
   - Supports shader functions (@vertex, @fragment, @compute)
   - Generates proper WGSL syntax for:
     - Variable declarations
     - Assignments
     - Control flow (if/else, for loops)
     - Function calls
     - Binary operations
     - Index expressions

4. **GPU Go Code Generator** (`pkg/codegen/gpu_go_generator.go`)
   - Generates matching Go structs for GPU types
   - Provides `ToBytes()` method for GPU upload
   - Generates size validation in `init()` functions
   - Embeds WGSL shader source using `//go:embed`

5. **Build Pipeline Integration** (`cmd/guix/main.go`)
   - Detects GPU declarations in `.gx` files
   - Generates `.wgsl` shader files
   - Generates `_gpu_gen.go` Go struct files
   - Works with existing `guix generate` command

6. **Parser Extensions** (`pkg/parser/parser.go`)
   - Extended lexer to recognize GPU decorators
   - Added `@gpu`, `@vertex`, `@fragment`, `@compute`, etc.
   - Added `@binding`, `@location`, `@builtin`, `@workgroup`
   - Added `@uniform`, `@storage`, `@align`

7. **Test Coverage** (`pkg/codegen/wgsl_test.go`)
   - WGSL struct generation tests
   - Type mapping tests
   - Go struct generation tests

## Current Limitations

⚠️ **Parser Integration Not Complete**

The parser **lexer** recognizes GPU decorator tokens, but the **grammar** does not yet parse GPU declarations. This means:

- ❌ Cannot parse `.gx` files with GPU syntax yet
- ❌ Must construct GPU AST nodes programmatically for testing
- ✅ AST structure is complete and ready
- ✅ Code generators work correctly with AST nodes
- ✅ Build pipeline integration is ready

### What Works Now

```go
// This works - programmatic AST construction
file := &guixast.File{
    Package: "shaders",
    GPUStructs: []*guixast.GPUStructDecl{
        {
            Name: "Uniforms",
            Struct: &guixast.GPUStructType{
                Fields: []*guixast.GPUField{
                    {Name: "viewportSize", Type: &guixast.GPUType{Name: "vec2"}},
                },
            },
        },
    },
}

gen := codegen.NewWGSLGenerator()
wgslCode, _ := gen.Generate(file) // Generates valid WGSL!
```

### What Doesn't Work Yet

```go
// This doesn't work yet - parser doesn't support GPU syntax
input := `
package shaders

@gpu type Uniforms struct {
    viewportSize vec2
}
`

parser.ParseString(input) // Error: unexpected token "@gpu"
```

## Next Steps

To complete the WGSL code generation feature, the following work is needed:

### 1. Parser Grammar Extension

Update the participle grammar in `pkg/ast/` to parse GPU declarations:

```go
// In pkg/ast/ast.go, update File struct tags
type File struct {
    Pos          lexer.Position
    Package      string            `"package" @Ident`
    Imports      []*Import         `@@*`
    Types        []*TypeDef        `@@*`
    GPUStructs   []*GPUStructDecl  `@@*`  // ← Parser rules needed
    GPUBindings  []*GPUBindingDecl `@@*`  // ← Parser rules needed
    GPUFunctions []*GPUFuncDecl    `@@*`  // ← Parser rules needed
    Components   []*Component      `@@*`
    Methods      []*Method         `@@*`
}
```

### 2. Grammar Implementation

The grammar needs to support:

```ebnf
GPUStructDecl   = Decorator* "type" Ident GPUStructType
GPUStructType   = "struct" "{" GPUField* "}"
GPUField        = Decorator* Ident GPUType
GPUBindingDecl  = Decorator+ "var" Ident GPUType
GPUFuncDecl     = Decorator* "func" Ident "(" Parameters? ")" GPUReturnType? Body
```

### 3. Decorator Parsing

Decorators need proper parsing:

```ebnf
Decorator = "@" ("gpu" | "vertex" | "fragment" | "compute" |
                 "uniform" | "storage" | "binding" | "location" |
                 "builtin" | "workgroup") ("(" Args ")")?
```

### 4. Integration Testing

Once parsing works, add end-to-end tests:

```go
func TestE2EWGSLGeneration(t *testing.T) {
    input := `
    package shaders

    @gpu type Uniforms struct {
        viewportSize vec2
        dataRange    vec4
    }

    @binding(0, 0) @uniform var uniforms Uniforms

    @vertex
    func vsMain(@builtin(vertex_index) idx uint32) vec4 {
        return vec4(0.0, 0.0, 0.0, 1.0)
    }
    `

    // Parse → Generate WGSL → Verify output
}
```

### 5. Example Application

Create a working example in `examples/wgsl-shader/`:

```
examples/wgsl-shader/
  ├── shader.gx        # GPU declarations in Guix
  ├── shader.wgsl      # Generated WGSL (auto)
  ├── shader_gpu_gen.go # Generated Go structs (auto)
  ├── main.go          # Example usage
  └── index.html       # WebGPU demo
```

## Usage (Once Parser is Complete)

### 1. Define GPU Structs

```go
// candle_shader.gx
package shaders

@gpu type ChartUniforms struct {
    viewportSize  vec2
    dataRange     vec4
    candleWidth   float32
    upColor       vec4
    downColor     vec4
}

@gpu type Candle struct {
    timestamp float32
    open      float32
    high      float32
    low       float32
    close     float32
    volume    float32
}
```

### 2. Define Bindings

```go
@binding(0, 0) @uniform var uniforms ChartUniforms
@binding(0, 1) @storage(read) var candles []Candle
```

### 3. Write Shader Functions

```go
@vertex
func vsMain(
    @builtin(vertex_index) vertexIndex uint32,
    @builtin(instance_index) candleIndex uint32
) VertexOutput {
    var output VertexOutput

    candle := candles[candleIndex]
    isBullish := candle.close >= candle.open

    output.color = select(uniforms.downColor, uniforms.upColor, isBullish)

    return output
}
```

### 4. Generate Code

```bash
guix generate -p .
```

This generates:
- `candle_shader.wgsl` - WGSL shader code
- `candle_shader_gpu_gen.go` - Go structs with `ToBytes()` methods

### 5. Use in Go

```go
// main.go
package main

import "myproject/shaders"

func main() {
    // Create uniforms
    uniforms := shaders.ChartUniforms{
        ViewportSize: [2]float32{800, 600},
        DataRange:    [4]float32{0, 0, 100, 100},
        CandleWidth:  2.0,
        UpColor:      [4]float32{0, 1, 0, 1},
        DownColor:    [4]float32{1, 0, 0, 1},
    }

    // Upload to GPU
    uniformBuffer.WriteData(uniforms.ToBytes())

    // Use shader source
    shader := device.CreateShaderModule(shaders.ShaderSource)
}
```

## Architecture

```
.gx file (with GPU declarations)
           ↓
    [Guix Parser] ← NOT YET IMPLEMENTED
           ↓
    [GPU AST] ✅ COMPLETE
           ↓
    ┌──────┴──────┐
    ↓             ↓
[WGSL Gen] ✅  [Go Gen] ✅
    ↓             ↓
.wgsl file   _gpu_gen.go
```

## Files Added/Modified

### New Files
- `pkg/ast/gpu_ast.go` - GPU AST node definitions
- `pkg/codegen/gpu_types.go` - Type mapping system
- `pkg/codegen/wgsl_generator.go` - WGSL code generator
- `pkg/codegen/gpu_go_generator.go` - GPU Go code generator
- `pkg/codegen/wgsl_test.go` - Test coverage

### Modified Files
- `pkg/ast/ast.go` - Added GPU declaration fields to File
- `pkg/ast/visitor.go` - Added GPU visitor methods
- `pkg/ast/base_visitor.go` - Added GPU visitor implementations
- `pkg/parser/parser.go` - Extended lexer for GPU decorators
- `cmd/guix/main.go` - Integrated GPU code generation

## Type Mapping Reference

| Guix Type | Go Type        | WGSL Type      | Size | Align |
|-----------|----------------|----------------|------|-------|
| float32   | float32        | f32            | 4    | 4     |
| int32     | int32          | i32            | 4    | 4     |
| uint32    | uint32         | u32            | 4    | 4     |
| bool      | bool           | bool           | 4    | 4     |
| vec2      | [2]float32     | vec2<f32>      | 8    | 8     |
| vec3      | [3]float32     | vec3<f32>      | 12   | 16    |
| vec4      | [4]float32     | vec4<f32>      | 16   | 16    |
| mat4      | [16]float32    | mat4x4<f32>    | 64   | 16    |
| []T       | []T            | array<T>       | -    | -     |
| [N]T      | [N]T           | array<T, N>    | -    | -     |

## Summary

This implementation provides the **foundation** for WGSL code generation in Guix:

✅ **Complete:**
- AST structure for GPU declarations
- Type mapping system
- WGSL code generator
- GPU Go code generator
- Build pipeline integration
- Test coverage

⚠️ **Incomplete:**
- Parser grammar for GPU syntax (requires participle grammar extension)

The implementation is production-ready **except** for parser support. Once the parser grammar is extended to recognize GPU declarations, the entire system will work end-to-end.

## Testing

Run tests:
```bash
go test ./pkg/codegen/wgsl_test.go -v
```

Expected output:
```
=== RUN   TestWGSLGenerator
=== RUN   TestWGSLGenerator/simple_GPU_struct
--- PASS: TestWGSLGenerator (0.00s)
    --- PASS: TestWGSLGenerator/simple_GPU_struct (0.00s)
=== RUN   TestGPUTypeMapping
=== RUN   TestGPUTypeMapping/vec2
=== RUN   TestGPUTypeMapping/vec4
=== RUN   TestGPUTypeMapping/mat4
=== RUN   TestGPUTypeMapping/float32
=== RUN   TestGPUTypeMapping/slice_of_vec4
--- PASS: TestGPUTypeMapping (0.00s)
PASS
```

## Future Enhancements

1. **Shader Composition** - Import and reuse shader functions
2. **Generic Shader Functions** - Type-parameterized shaders
3. **Conditional Compilation** - Debug-only shader code
4. **Compute Shader Support** - Enhanced @compute decorator handling
5. **Texture Bindings** - @binding support for textures and samplers
6. **Error Messages** - Map WGSL line numbers back to Guix source
7. **Hot Reload** - Watch mode for shader development
8. **Validation** - Check WGSL validity before generation

## References

- [WGSL Specification](https://www.w3.org/TR/WGSL/)
- [WebGPU API](https://www.w3.org/TR/webgpu/)
- [Guix Documentation](../README.md)
- [Participle Parser](https://github.com/alecthomas/participle)
