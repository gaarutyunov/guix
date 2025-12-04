package codegen

import (
	"strings"
	"testing"

	guixast "github.com/gaarutyunov/guix/pkg/ast"
)

func TestWGSLGenerator(t *testing.T) {
	tests := []struct {
		name     string
		file     *guixast.File
		contains []string
	}{
		{
			name: "simple GPU struct",
			file: &guixast.File{
				Package: "shaders",
				GPUStructs: []*guixast.GPUStructDecl{
					{
						Name: "Uniforms",
						Struct: &guixast.GPUStructType{
							Fields: []*guixast.GPUField{
								{
									Name: "viewportSize",
									Type: &guixast.GPUType{Name: "vec2"},
								},
								{
									Name: "dataRange",
									Type: &guixast.GPUType{Name: "vec4"},
								},
							},
						},
					},
				},
			},
			contains: []string{
				"struct Uniforms {",
				"viewportSize: vec2<f32>,",
				"dataRange: vec4<f32>,",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate WGSL
			gen := NewWGSLGenerator()
			output, err := gen.Generate(tt.file)
			if err != nil {
				t.Fatalf("Failed to generate WGSL: %v", err)
			}

			outputStr := string(output)

			// Check that output contains expected strings
			for _, expected := range tt.contains {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected output to contain %q, but it didn't.\nOutput:\n%s", expected, outputStr)
				}
			}
		})
	}
}

func TestGPUTypeMapping(t *testing.T) {
	tests := []struct {
		name      string
		gpuType   *guixast.GPUType
		wgslType  string
		goType    string
		shouldErr bool
	}{
		{
			name:     "vec2",
			gpuType:  &guixast.GPUType{Name: "vec2"},
			wgslType: "vec2<f32>",
			goType:   "[2]float32",
		},
		{
			name:     "vec4",
			gpuType:  &guixast.GPUType{Name: "vec4"},
			wgslType: "vec4<f32>",
			goType:   "[4]float32",
		},
		{
			name:     "mat4",
			gpuType:  &guixast.GPUType{Name: "mat4"},
			wgslType: "mat4x4<f32>",
			goType:   "[16]float32",
		},
		{
			name:     "float32",
			gpuType:  &guixast.GPUType{Name: "float32"},
			wgslType: "f32",
			goType:   "float32",
		},
		{
			name:     "slice of vec4",
			gpuType:  &guixast.GPUType{IsSlice: true, Name: "vec4"},
			wgslType: "array<vec4<f32>>",
			goType:   "[]vec4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test WGSL mapping
			wgslResult, err := MapGPUTypeToWGSL(tt.gpuType)
			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if wgslResult != tt.wgslType {
				t.Errorf("WGSL mapping: expected %q, got %q", tt.wgslType, wgslResult)
			}

			// Test Go mapping
			goResult, err := MapGPUTypeToGo(tt.gpuType)
			if err != nil {
				t.Fatalf("Unexpected error in Go mapping: %v", err)
			}

			if goResult != tt.goType {
				t.Errorf("Go mapping: expected %q, got %q", tt.goType, goResult)
			}
		})
	}
}
