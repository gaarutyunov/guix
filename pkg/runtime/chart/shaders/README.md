# Chart Shader Types - Guix GPU Code Generation

This directory demonstrates Guix's WGSL code generation capabilities by providing type-safe GPU struct definitions for chart rendering.

## Generated Files

The following files are generated from `.gx` source files using `guix generate`:

### Candlestick Chart Types

**Source**: `candlestick_types.gx`

**Generated Files**:
- `candlestick_types.wgsl` - WGSL struct definitions
- `candlestick_types_gpu_gen.go` - Go structs with `ToBytes()` methods and size validation
- `candlestick_types_gen.go` - Regular Guix component code

**Structs**:
- `ChartUniforms` - Uniform buffer data for candlestick rendering
- `Candle` - OHLCV candlestick data point

### Line Chart Types

**Source**: `line_types.gx`

**Generated Files**:
- `line_types.wgsl` - WGSL struct definitions
- `line_types_gpu_gen.go` - Go structs with `ToBytes()` methods and size validation
- `line_types_gen.go` - Regular Guix component code

**Structs**:
- `LineUniforms` - Uniform buffer data for line rendering
- `Point` - 2D point data

## Benefits of Guix GPU Code Generation

1. **Type Safety**: Guix ensures CPU (Go) and GPU (WGSL) types stay synchronized
2. **Memory Layout Validation**: Generated `init()` functions verify struct sizes match between Go and WGSL
3. **Automatic Serialization**: `ToBytes()` methods provide zero-copy serialization to GPU buffers
4. **Single Source of Truth**: Define types once in `.gx` files, generate both Go and WGSL

## Usage Example

### 1. Define GPU Types in Guix

```go
// candlestick_types.gx
package shaders

@gpu type ChartUniforms struct {
	viewportSize vec2    // Canvas width and height
	dataRange    vec4    // minX, maxX, minY, maxY
	candleWidth  float32 // Candle width in pixels
	upColor      vec4    // Color for up candles
	downColor    vec4    // Color for down candles
}

@gpu type Candle struct {
	timestamp float32
	open      float32
	high      float32
	low       float32
	close     float32
}
```

### 2. Generate Code

```bash
guix generate -p pkg/runtime/chart/shaders
```

### 3. Use in Go

```go
import "github.com/gaarutyunov/guix/pkg/runtime/chart/shaders"

// Create uniform data
uniforms := shaders.ChartUniforms{
	ViewportSize: [2]float32{800, 600},
	DataRange:    [4]float32{0, 100, 0, 1000},
	CandleWidth:  10.0,
	UpColor:      [4]float32{0, 1, 0, 1},  // Green
	DownColor:    [4]float32{1, 0, 0, 1},  // Red
}

// Upload to GPU uniform buffer
uniformBytes := uniforms.ToBytes()
device.Queue().WriteBuffer(uniformBuffer, 0, uniformBytes)
```

### 4. Use in WGSL Shaders

The generated WGSL struct definitions can be copied into your shader code, or you can reference the generated `.wgsl` file:

```wgsl
// Include generated type definitions
struct ChartUniforms {
    viewportSize: vec2<f32>,
    dataRange: vec4<f32>,
    candleWidth: f32,
    upColor: vec4<f32>,
    downColor: vec4<f32>,
}

struct Candle {
    timestamp: f32,
    open: f32,
    high: f32,
    low: f32,
    close: f32,
}

// Bind the uniform buffer
@group(0) @binding(0) var<uniform> uniforms: ChartUniforms;
@group(0) @binding(1) var<storage, read> candles: array<Candle>;

// Use in shader functions
@vertex
fn vs_main(@builtin(instance_index) idx: u32) -> @builtin(position) vec4<f32> {
    let candle = candles[idx];
    let isUp = candle.close >= candle.open;
    let color = select(uniforms.downColor, uniforms.upColor, isUp);
    // ... render logic
}
```

## Type Safety Guarantees

The generated Go code includes size validation that runs at initialization:

```go
func init() {
	expectedSize := 112
	if unsafe.Sizeof(ChartUniforms{}) != expectedSize {
		panic("ChartUniforms size mismatch with WGSL")
	}
}
```

This ensures that if you modify the Guix types, both Go and WGSL code will be regenerated consistently, preventing subtle memory layout bugs.

## Comparison with Manual WGSL

### Before (Manual WGSL + Manual Go Structs)

**Problems:**
- Need to maintain identical struct definitions in two languages
- Manual padding calculations required
- No compile-time validation that Go/WGSL match
- Easy to introduce subtle bugs when updating types

**WGSL (candlestick.wgsl)**:
```wgsl
struct ChartUniforms {
    viewportSize: vec2<f32>,
    dataRange: vec4<f32>,
    padding: vec4<f32>,
    candleWidth: f32,
    upColor: vec4<f32>,
    downColor: vec4<f32>,
    wickColor: vec4<f32>,
}
```

**Go (manual)**:
```go
type ChartUniforms struct {
	ViewportSize [2]float32
	DataRange    [4]float32
	Padding      [4]float32
	CandleWidth  float32
	_padding1    [3]float32  // Manual padding - error prone!
	UpColor      [4]float32
	DownColor    [4]float32
	WickColor    [4]float32
}
```

### After (Guix GPU Code Generation)

**Benefits:**
- Single source of truth in `.gx` files
- Automatic padding and alignment
- Runtime size validation
- Type-safe serialization with `ToBytes()`

**Guix (candlestick_types.gx)**:
```go
@gpu type ChartUniforms struct {
	viewportSize vec2
	dataRange    vec4
	padding      vec4
	candleWidth  float32
	upColor      vec4
	downColor    vec4
	wickColor    vec4
}
```

**Generated automatically:**
- ✅ WGSL struct definition
- ✅ Go struct with correct padding
- ✅ ToBytes() serialization method
- ✅ Size validation in init()

## Regenerating Code

When you modify the `.gx` files, regenerate all code with:

```bash
guix generate -p pkg/runtime/chart/shaders
```

This ensures Go and WGSL stay perfectly synchronized.

## Future Enhancements

The current implementation generates struct definitions and bindings. Future enhancements could include:

- Full shader function generation (vertex/fragment shaders)
- Support for more complex WGSL features (control flow, built-in functions)
- Automatic binding generation with `@binding()` decorators
- Integration with Guix component system for declarative GPU pipelines

## Manual Shader Functions

For now, complex shader logic (vertex/fragment functions) should be written manually in WGSL:

- `candlestick.wgsl` - Complete candlestick shader with rendering logic
- `line.wgsl` - Complete line chart shader with rendering logic

These shaders use the struct definitions from the generated `.wgsl` files or include them directly.
