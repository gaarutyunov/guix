package parser

import (
	"testing"
)

func TestParseGPUStruct(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "simple GPU struct",
			input: `package shaders

@gpu type Uniforms struct {
	viewportSize vec2
	dataRange vec4
}`,
			wantErr: false,
		},
		{
			name: "GPU struct with decorators on fields",
			input: `package shaders

@gpu type VertexOutput struct {
	@builtin(position) position vec4
	@location(0) color vec4
}`,
			wantErr: false,
		},
		{
			name: "multiple GPU structs",
			input: `package shaders

@gpu type Uniforms struct {
	viewportSize vec2
}

@gpu type Candle struct {
	timestamp float32
	open float32
	close float32
}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := New()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			file, err := p.ParseString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				// Verify we parsed GPU structs
				if len(file.GPUStructs) == 0 {
					t.Errorf("Expected GPU structs to be parsed, got none")
				}

				t.Logf("Successfully parsed %d GPU struct(s)", len(file.GPUStructs))
				for i, s := range file.GPUStructs {
					t.Logf("  GPU Struct %d: %s with %d fields", i+1, s.Name, len(s.Struct.Fields))
				}
			}
		})
	}
}

func TestParseGPUBinding(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "uniform binding",
			input: `package shaders

@binding(0, 0) @uniform var uniforms Uniforms`,
			wantErr: false,
		},
		{
			name: "storage binding",
			input: `package shaders

@binding(0, 1) @storage var candles []Candle`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := New()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			file, err := p.ParseString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if len(file.GPUBindings) == 0 {
					t.Errorf("Expected GPU bindings to be parsed, got none")
				}

				t.Logf("Successfully parsed %d GPU binding(s)", len(file.GPUBindings))
			}
		})
	}
}

func TestParseGPUFunction(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "vertex shader function",
			input: `package shaders

@vertex
func vsMain(@builtin(vertex_index) idx uint32) vec4 {
	return vec4(0.0, 0.0, 0.0, 1.0)
}`,
			wantErr: false,
		},
		{
			name: "fragment shader function",
			input: `package shaders

@fragment
func fsMain(@location(0) color vec4) vec4 {
	return color
}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := New()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			file, err := p.ParseString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseString() error = %v, wantErr %v", err, tt.wantErr)
				if err != nil {
					t.Logf("Parse error details: %v", err)
				}
				return
			}

			if err == nil {
				if len(file.GPUFunctions) == 0 {
					t.Errorf("Expected GPU functions to be parsed, got none")
				}

				t.Logf("Successfully parsed %d GPU function(s)", len(file.GPUFunctions))
			}
		})
	}
}

func TestParseCompleteGPUShader(t *testing.T) {
	input := `package shaders

@gpu type Uniforms struct {
	viewportSize vec2
	color vec4
}

@gpu type VertexOutput struct {
	@builtin(position) position vec4
	@location(0) color vec4
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

	p, err := New()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	file, err := p.ParseString(input)
	if err != nil {
		t.Fatalf("Failed to parse complete shader: %v", err)
	}

	// Verify all components were parsed
	if len(file.GPUStructs) < 2 {
		t.Errorf("Expected at least 2 GPU structs, got %d", len(file.GPUStructs))
	}

	if len(file.GPUBindings) < 1 {
		t.Errorf("Expected at least 1 GPU binding, got %d", len(file.GPUBindings))
	}

	if len(file.GPUFunctions) < 2 {
		t.Errorf("Expected at least 2 GPU functions, got %d", len(file.GPUFunctions))
	}

	t.Logf("Successfully parsed complete shader:")
	t.Logf("  - %d GPU structs", len(file.GPUStructs))
	t.Logf("  - %d GPU bindings", len(file.GPUBindings))
	t.Logf("  - %d GPU functions", len(file.GPUFunctions))
}

func TestParseGPUWithRegularCode(t *testing.T) {
	input := `package mixed

import "fmt"

@gpu type Uniforms struct {
	color vec4
}

@binding(0, 0) @uniform var uniforms Uniforms

@vertex
func vsMain(@builtin(vertex_index) idx uint32) vec4 {
	return vec4(0.0, 0.0, 0.0, 1.0)
}

type RegularStruct struct {
	Name string
}

@props func MyComponent(data chan int) (Component) {
	Div() {
		"Hello"
	}
}`

	p, err := New()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	file, err := p.ParseString(input)
	if err != nil {
		t.Fatalf("Failed to parse mixed file: %v", err)
	}

	// Verify all components
	if len(file.Imports) == 0 {
		t.Error("Expected imports to be parsed")
	}

	if len(file.GPUStructs) == 0 {
		t.Error("Expected GPU structs to be parsed")
	}

	if len(file.Types) == 0 {
		t.Error("Expected regular types to be parsed")
	}

	if len(file.Components) == 0 {
		t.Error("Expected components to be parsed")
	}

	if len(file.GPUBindings) == 0 {
		t.Error("Expected GPU bindings to be parsed")
	}

	if len(file.GPUFunctions) == 0 {
		t.Error("Expected GPU functions to be parsed")
	}

	t.Logf("Successfully parsed mixed file:")
	t.Logf("  - %d imports", len(file.Imports))
	t.Logf("  - %d GPU structs", len(file.GPUStructs))
	t.Logf("  - %d regular types", len(file.Types))
	t.Logf("  - %d components", len(file.Components))
	t.Logf("  - %d GPU bindings", len(file.GPUBindings))
	t.Logf("  - %d GPU functions", len(file.GPUFunctions))
}
