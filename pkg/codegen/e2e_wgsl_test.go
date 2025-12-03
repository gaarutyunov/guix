package codegen

import (
	"strings"
	"testing"

	"github.com/gaarutyunov/guix/pkg/parser"
)

func TestE2EWGSLGeneration(t *testing.T) {
	input := `package shaders

@gpu type Uniforms struct {
	viewportSize vec2
	color vec4
}

@binding(0, 0) @uniform var uniforms Uniforms

@vertex
func vsMain(@builtin(vertex_index) idx uint32) vec4 {
	return vec4(0.0, 0.0, 0.0, 1.0)
}

@fragment
func fsMain(@location(0) color vec4) vec4 {
	return color
}`

	// Parse
	p, err := parser.New()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	file, err := p.ParseString(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Verify parsing
	if len(file.GPUStructs) != 1 {
		t.Errorf("Expected 1 GPU struct, got %d", len(file.GPUStructs))
	}

	if len(file.GPUBindings) != 1 {
		t.Errorf("Expected 1 GPU binding, got %d", len(file.GPUBindings))
	}

	if len(file.GPUFunctions) != 2 {
		t.Errorf("Expected 2 GPU functions, got %d", len(file.GPUFunctions))
	}

	// Generate WGSL
	wgslGen := NewWGSLGenerator()
	wgslOutput, err := wgslGen.Generate(file)
	if err != nil {
		t.Fatalf("Failed to generate WGSL: %v", err)
	}

	wgslCode := string(wgslOutput)
	t.Logf("Generated WGSL:\n%s", wgslCode)

	// Verify WGSL output contains expected elements
	expectedStrings := []string{
		"struct Uniforms {",
		"viewportSize: vec2<f32>,",
		"color: vec4<f32>,",
		"@binding(0, 0)",
		"@uniform",
		"var<uniform> uniforms: Uniforms;",
		"@vertex",
		"fn vsMain(",
		"@builtin(vertex_index) idx: u32",
		"-> vec4<f32>",
		"@fragment",
		"fn fsMain(",
		"@location(0) color: vec4<f32>",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(wgslCode, expected) {
			t.Errorf("Expected WGSL to contain %q, but it didn't", expected)
		}
	}

	// Generate GPU Go code
	gpuGoGen := NewGPUGoGenerator(file.Package, "shaders.wgsl")
	gpuGoOutput, err := gpuGoGen.Generate(file)
	if err != nil {
		t.Fatalf("Failed to generate GPU Go code: %v", err)
	}

	goCode := string(gpuGoOutput)
	t.Logf("Generated Go code:\n%s", goCode)

	// Verify Go output contains expected elements
	expectedGoStrings := []string{
		"type Uniforms struct {",
		"ViewportSize",
		"[2]float32",
		"Color",
		"[4]float32",
		"func (s *Uniforms) ToBytes() []byte",
		"//go:embed shaders.wgsl",
		"var ShaderSource string",
	}

	for _, expected := range expectedGoStrings {
		if !strings.Contains(goCode, expected) {
			t.Errorf("Expected Go code to contain %q, but it didn't", expected)
		}
	}
}

func TestE2ECompleteShader(t *testing.T) {
	input := `package candle

@gpu type ChartUniforms struct {
	viewportSize vec2
	dataRange vec4
	candleWidth float32
	upColor vec4
	downColor vec4
}

@gpu type Candle struct {
	timestamp float32
	open float32
	high float32
	low float32
	close float32
}

@binding(0, 0) @uniform var uniforms ChartUniforms
@binding(0, 1) @storage var candles []Candle

@vertex
func vsCandle(@builtin(vertex_index) vid uint32, @builtin(instance_index) iid uint32) vec4 {
	c := candles[iid]
	bullish := c.close >= c.open
	x := (c.timestamp - uniforms.dataRange.x) / (uniforms.dataRange.z - uniforms.dataRange.x)
	return vec4(x, 0.0, 0.0, 1.0)
}

@fragment
func fsCandle(@location(0) color vec4) vec4 {
	return color
}`

	// Parse
	p, err := parser.New()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	file, err := p.ParseString(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Generate WGSL
	wgslGen := NewWGSLGenerator()
	wgslOutput, err := wgslGen.Generate(file)
	if err != nil {
		t.Fatalf("Failed to generate WGSL: %v", err)
	}

	wgslCode := string(wgslOutput)
	t.Logf("Generated complete WGSL:\n%s", wgslCode)

	// Verify complex features
	if !strings.Contains(wgslCode, "array<Candle>") {
		t.Error("Expected array type for candles slice")
	}

	if !strings.Contains(wgslCode, "let c = candles[iid];") {
		t.Error("Expected array indexing in generated code")
	}

	if !strings.Contains(wgslCode, "c.close >= c.open") {
		t.Error("Expected comparison expression")
	}

	if !strings.Contains(wgslCode, "uniforms.dataRange.x") {
		t.Error("Expected field access")
	}

	// Generate Go code
	gpuGoGen := NewGPUGoGenerator(file.Package, "candle.wgsl")
	gpuGoOutput, err := gpuGoGen.Generate(file)
	if err != nil {
		t.Fatalf("Failed to generate GPU Go code: %v", err)
	}

	goCode := string(gpuGoOutput)

	// Verify Go struct generation
	if !strings.Contains(goCode, "type ChartUniforms struct") {
		t.Error("Expected ChartUniforms struct")
	}

	if !strings.Contains(goCode, "type Candle struct") {
		t.Error("Expected Candle struct")
	}
}
