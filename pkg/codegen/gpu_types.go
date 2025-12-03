// Package codegen provides GPU type mapping utilities for WGSL code generation
package codegen

import (
	"fmt"
	"strings"

	guixast "github.com/gaarutyunov/guix/pkg/ast"
)

// GPUTypeInfo contains information about a GPU type
type GPUTypeInfo struct {
	WGSLType  string // WGSL type name (e.g., "vec2<f32>", "mat4x4<f32>")
	GoType    string // Go type name (e.g., "[2]float32", "[16]float32")
	Size      int    // Size in bytes
	Alignment int    // Alignment in bytes
	IsVector  bool   // Is this a vector type?
	IsMatrix  bool   // Is this a matrix type?
	IsScalar  bool   // Is this a scalar type?
}

// guixToWGSL maps Guix/Go type names to WGSL types
var guixToWGSL = map[string]GPUTypeInfo{
	// Scalar types
	"float32": {WGSLType: "f32", GoType: "float32", Size: 4, Alignment: 4, IsScalar: true},
	"int32":   {WGSLType: "i32", GoType: "int32", Size: 4, Alignment: 4, IsScalar: true},
	"uint32":  {WGSLType: "u32", GoType: "uint32", Size: 4, Alignment: 4, IsScalar: true},
	"bool":    {WGSLType: "bool", GoType: "bool", Size: 4, Alignment: 4, IsScalar: true},

	// 2D vectors
	"vec2":  {WGSLType: "vec2<f32>", GoType: "[2]float32", Size: 8, Alignment: 8, IsVector: true},
	"ivec2": {WGSLType: "vec2<i32>", GoType: "[2]int32", Size: 8, Alignment: 8, IsVector: true},
	"uvec2": {WGSLType: "vec2<u32>", GoType: "[2]uint32", Size: 8, Alignment: 8, IsVector: true},

	// 3D vectors (aligned to 16 bytes)
	"vec3":  {WGSLType: "vec3<f32>", GoType: "[3]float32", Size: 12, Alignment: 16, IsVector: true},
	"ivec3": {WGSLType: "vec3<i32>", GoType: "[3]int32", Size: 12, Alignment: 16, IsVector: true},
	"uvec3": {WGSLType: "vec3<u32>", GoType: "[3]uint32", Size: 12, Alignment: 16, IsVector: true},

	// 4D vectors (aligned to 16 bytes)
	"vec4":  {WGSLType: "vec4<f32>", GoType: "[4]float32", Size: 16, Alignment: 16, IsVector: true},
	"ivec4": {WGSLType: "vec4<i32>", GoType: "[4]int32", Size: 16, Alignment: 16, IsVector: true},
	"uvec4": {WGSLType: "vec4<u32>", GoType: "[4]uint32", Size: 16, Alignment: 16, IsVector: true},

	// Matrices (column-major, aligned to 16 bytes per column)
	"mat2":   {WGSLType: "mat2x2<f32>", GoType: "[4]float32", Size: 16, Alignment: 8, IsMatrix: true},
	"mat3":   {WGSLType: "mat3x3<f32>", GoType: "[9]float32", Size: 48, Alignment: 16, IsMatrix: true},
	"mat4":   {WGSLType: "mat4x4<f32>", GoType: "[16]float32", Size: 64, Alignment: 16, IsMatrix: true},
	"mat2x3": {WGSLType: "mat2x3<f32>", GoType: "[6]float32", Size: 32, Alignment: 16, IsMatrix: true},
	"mat2x4": {WGSLType: "mat2x4<f32>", GoType: "[8]float32", Size: 32, Alignment: 16, IsMatrix: true},
	"mat3x2": {WGSLType: "mat3x2<f32>", GoType: "[6]float32", Size: 24, Alignment: 8, IsMatrix: true},
	"mat3x4": {WGSLType: "mat3x4<f32>", GoType: "[12]float32", Size: 48, Alignment: 16, IsMatrix: true},
	"mat4x2": {WGSLType: "mat4x2<f32>", GoType: "[8]float32", Size: 32, Alignment: 8, IsMatrix: true},
	"mat4x3": {WGSLType: "mat4x3<f32>", GoType: "[12]float32", Size: 64, Alignment: 16, IsMatrix: true},
}

// GPUTypeMappingError represents an error in GPU type mapping
type GPUTypeMappingError struct {
	TypeName string
	Reason   string
}

func (e *GPUTypeMappingError) Error() string {
	return fmt.Sprintf("GPU type mapping error for '%s': %s", e.TypeName, e.Reason)
}

// MapGPUTypeToWGSL converts a Guix GPU type to WGSL type string
func MapGPUTypeToWGSL(gpuType *guixast.GPUType) (string, error) {
	if gpuType == nil {
		return "", &GPUTypeMappingError{TypeName: "<nil>", Reason: "nil type"}
	}

	// Handle slice types
	if gpuType.IsSlice {
		elemType, err := MapGPUTypeToWGSL(&guixast.GPUType{
			Name:    gpuType.Name,
			Generic: gpuType.Generic,
		})
		if err != nil {
			return "", err
		}

		// WGSL uses array<T> for runtime-sized arrays
		return fmt.Sprintf("array<%s>", elemType), nil
	}

	// Handle array types with size
	if gpuType.Generic != nil && gpuType.Generic.Name != "" {
		elemType, err := MapGPUTypeToWGSL(&guixast.GPUType{
			Name: gpuType.Name,
		})
		if err != nil {
			return "", err
		}

		// WGSL uses array<T, N> for fixed-size arrays
		return fmt.Sprintf("array<%s, %s>", elemType, gpuType.Generic.Name), nil
	}

	// Look up the type in our mapping table
	typeInfo, ok := guixToWGSL[gpuType.Name]
	if !ok {
		// If not a built-in type, assume it's a user-defined struct
		return gpuType.Name, nil
	}

	return typeInfo.WGSLType, nil
}

// MapGPUTypeToGo converts a Guix GPU type to Go type string
func MapGPUTypeToGo(gpuType *guixast.GPUType) (string, error) {
	if gpuType == nil {
		return "", &GPUTypeMappingError{TypeName: "<nil>", Reason: "nil type"}
	}

	// Handle pointer types
	if gpuType.IsPointer {
		elemType, err := MapGPUTypeToGo(&guixast.GPUType{
			IsSlice: gpuType.IsSlice,
			Name:    gpuType.Name,
			Generic: gpuType.Generic,
		})
		if err != nil {
			return "", err
		}
		return "*" + elemType, nil
	}

	// Handle slice types
	if gpuType.IsSlice {
		// For slices, just use the type name directly
		// e.g., []vec4 not [][4]float32
		return "[]" + gpuType.Name, nil
	}

	// Handle array types with size
	if gpuType.Generic != nil && gpuType.Generic.Name != "" {
		// For arrays, use the type name
		return fmt.Sprintf("[%s]%s", gpuType.Generic.Name, gpuType.Name), nil
	}

	// Look up the type in our mapping table
	typeInfo, ok := guixToWGSL[gpuType.Name]
	if !ok {
		// If not a built-in type, assume it's a user-defined struct
		return gpuType.Name, nil
	}

	return typeInfo.GoType, nil
}

// GetGPUTypeInfo returns detailed information about a GPU type
func GetGPUTypeInfo(typeName string) (GPUTypeInfo, bool) {
	info, ok := guixToWGSL[typeName]
	return info, ok
}

// CalculateStructSize calculates the size and alignment of a GPU struct
// following WGSL alignment rules
func CalculateStructSize(fields []*guixast.GPUField) (size int, alignment int, err error) {
	offset := 0
	maxAlign := 0

	for _, field := range fields {
		typeInfo, ok := guixToWGSL[field.Type.Name]
		if !ok {
			// Unknown type - can't calculate size
			return 0, 0, &GPUTypeMappingError{
				TypeName: field.Type.Name,
				Reason:   "unknown type, cannot calculate struct size",
			}
		}

		// Update maximum alignment
		if typeInfo.Alignment > maxAlign {
			maxAlign = typeInfo.Alignment
		}

		// Align current offset to field alignment
		if offset%typeInfo.Alignment != 0 {
			offset += typeInfo.Alignment - (offset % typeInfo.Alignment)
		}

		// Add field size
		offset += typeInfo.Size
	}

	// Final struct alignment is the maximum field alignment
	// Final struct size is rounded up to alignment
	if offset%maxAlign != 0 {
		offset += maxAlign - (offset % maxAlign)
	}

	return offset, maxAlign, nil
}

// GeneratePaddingFields generates padding fields to ensure proper alignment
func GeneratePaddingFields(fields []*guixast.GPUField) ([]*guixast.GPUField, error) {
	result := make([]*guixast.GPUField, 0)
	offset := 0
	paddingCounter := 0

	for _, field := range fields {
		typeInfo, ok := guixToWGSL[field.Type.Name]
		if !ok {
			// Unknown type - just add the field without padding
			result = append(result, field)
			continue
		}

		// Calculate padding needed before this field
		if offset%typeInfo.Alignment != 0 {
			padding := typeInfo.Alignment - (offset % typeInfo.Alignment)

			// Add padding field
			padField := &guixast.GPUField{
				Name: fmt.Sprintf("_pad%d", paddingCounter),
				Type: &guixast.GPUType{
					IsSlice: true,
					Name:    "byte",
				},
			}
			result = append(result, padField)
			paddingCounter++
			offset += padding
		}

		// Add the actual field
		result = append(result, field)
		offset += typeInfo.Size
	}

	return result, nil
}

// IsGPUBuiltinType checks if a type is a GPU built-in type
func IsGPUBuiltinType(typeName string) bool {
	_, ok := guixToWGSL[typeName]
	return ok
}

// WGSLBuiltinFunctions maps Guix function names to WGSL built-in functions
var WGSLBuiltinFunctions = map[string]string{
	// Math functions
	"abs":   "abs",
	"floor": "floor",
	"ceil":  "ceil",
	"round": "round",
	"sqrt":  "sqrt",
	"sin":   "sin",
	"cos":   "cos",
	"tan":   "tan",
	"asin":  "asin",
	"acos":  "acos",
	"atan":  "atan",
	"pow":   "pow",
	"exp":   "exp",
	"log":   "log",

	// Vector/matrix functions
	"dot":       "dot",
	"cross":     "cross",
	"length":    "length",
	"normalize": "normalize",
	"distance":  "distance",

	// Interpolation
	"mix":        "mix",
	"clamp":      "clamp",
	"smoothstep": "smoothstep",

	// Min/max
	"min": "min",
	"max": "max",

	// Select (ternary)
	"select": "select",
}

// IsWGSLBuiltinFunction checks if a function name is a WGSL built-in
func IsWGSLBuiltinFunction(name string) bool {
	_, ok := WGSLBuiltinFunctions[name]
	return ok
}

// MapFunctionToWGSL maps a Guix function call to WGSL
func MapFunctionToWGSL(name string) string {
	if wgslName, ok := WGSLBuiltinFunctions[name]; ok {
		return wgslName
	}
	// Return as-is for user-defined functions
	return name
}

// FormatWGSLType formats a type name for WGSL output
func FormatWGSLType(typeName string) string {
	return strings.TrimSpace(typeName)
}
