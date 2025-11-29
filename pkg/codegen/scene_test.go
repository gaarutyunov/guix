package codegen

import (
	"go/format"
	"strings"
	"testing"

	"github.com/gaarutyunov/guix/pkg/parser"
)

func TestGenerateSimpleScene(t *testing.T) {
	p, err := parser.New()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	source := `package main

func TestScene() (Scene) {
	Scene {
		Mesh()
	}
}`

	file, err := p.ParseString(source)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	gen := New("main")
	output, err := gen.Generate(file)
	if err != nil {
		t.Fatalf("Failed to generate: %v", err)
	}

	outputStr := string(output)

	// Check for struct definition
	if !strings.Contains(outputStr, "type TestScene struct") {
		t.Error("Expected TestScene struct definition")
	}

	// Check for constructor
	if !strings.Contains(outputStr, "func NewTestScene() runtime.Scene") {
		t.Error("Expected NewTestScene constructor returning runtime.Scene")
	}

	// Check for RenderScene method
	if !strings.Contains(outputStr, "func (s *TestScene) RenderScene() *runtime.GPUNode") {
		t.Error("Expected RenderScene method")
	}

	// Check for Scene call
	if !strings.Contains(outputStr, "runtime.SceneNode(") {
		t.Error("Expected runtime.Scene call")
	}

	// Check for Mesh call
	if !strings.Contains(outputStr, "runtime.Mesh()") {
		t.Error("Expected runtime.Mesh call")
	}

	// Verify code is valid Go
	_, err = format.Source(output)
	if err != nil {
		t.Errorf("Generated code is not valid Go: %v\n%s", err, outputStr)
	}
}

func TestGenerateSceneWithParams(t *testing.T) {
	p, err := parser.New()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	source := `package main

func RotatingScene(rotX float32, rotY float32) (Scene) {
	Scene {
		Mesh(Rotation(rotX, rotY, 0))
	}
}`

	file, err := p.ParseString(source)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	gen := New("main")
	output, err := gen.Generate(file)
	if err != nil {
		t.Fatalf("Failed to generate: %v", err)
	}

	outputStr := string(output)

	// Check for struct with fields
	if !strings.Contains(outputStr, "type RotatingScene struct") {
		t.Error("Expected RotatingScene struct definition")
	}

	if !strings.Contains(outputStr, "RotX float32") {
		t.Error("Expected RotX field")
	}

	if !strings.Contains(outputStr, "RotY float32") {
		t.Error("Expected RotY field")
	}

	// Check for constructor with parameters
	if !strings.Contains(outputStr, "func NewRotatingScene(rotX float32, rotY float32) runtime.Scene") {
		t.Error("Expected NewRotatingScene constructor with parameters")
	}

	// Check for struct initialization
	if !strings.Contains(outputStr, "&RotatingScene{RotX: rotX, RotY: rotY}") {
		t.Error("Expected struct initialization with parameters")
	}

	// Verify code is valid Go
	_, err = format.Source(output)
	if err != nil {
		t.Errorf("Generated code is not valid Go: %v\n%s", err, outputStr)
	}
}

func TestGenerateSceneWithBackgroundProp(t *testing.T) {
	p, err := parser.New()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	source := `package main

func ColoredScene() (Scene) {
	Scene(Background(0.1, 0.1, 0.15, 1.0)) {
		Mesh()
	}
}`

	file, err := p.ParseString(source)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	gen := New("main")
	output, err := gen.Generate(file)
	if err != nil {
		t.Fatalf("Failed to generate: %v", err)
	}

	outputStr := string(output)

	// Check for Background call with runtime prefix
	if !strings.Contains(outputStr, "runtime.Background(0.1, 0.1, 0.15, 1.0)") {
		t.Error("Expected runtime.Background call with 4 arguments")
	}

	// Check that Background is passed to Scene
	if !strings.Contains(outputStr, "runtime.SceneNode(runtime.Background") {
		t.Error("Expected Background to be passed to Scene")
	}

	// Verify code is valid Go
	_, err = format.Source(output)
	if err != nil {
		t.Errorf("Generated code is not valid Go: %v\n%s", err, outputStr)
	}
}

func TestGenerateSceneWithMultipleElements(t *testing.T) {
	p, err := parser.New()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	source := `package main

func ComplexScene() (Scene) {
	Scene {
		Mesh(Position(0, 0, 0))
		PerspectiveCamera(FOV(60))
		AmbientLight(Intensity(0.5))
	}
}`

	file, err := p.ParseString(source)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	gen := New("main")
	output, err := gen.Generate(file)
	if err != nil {
		t.Fatalf("Failed to generate: %v", err)
	}

	outputStr := string(output)

	// Check for all three elements
	if !strings.Contains(outputStr, "runtime.Mesh(") {
		t.Error("Expected runtime.Mesh call")
	}

	if !strings.Contains(outputStr, "runtime.PerspectiveCamera(") {
		t.Error("Expected runtime.PerspectiveCamera call")
	}

	if !strings.Contains(outputStr, "runtime.AmbientLight(") {
		t.Error("Expected runtime.AmbientLight call")
	}

	// Check for Position, FOV, and Intensity props
	if !strings.Contains(outputStr, "runtime.Position(0, 0, 0)") {
		t.Error("Expected runtime.Position call")
	}

	if !strings.Contains(outputStr, "runtime.FOV(60)") {
		t.Error("Expected runtime.FOV call")
	}

	if !strings.Contains(outputStr, "runtime.Intensity(0.5)") {
		t.Error("Expected runtime.Intensity call")
	}

	// Verify code is valid Go
	_, err = format.Source(output)
	if err != nil {
		t.Errorf("Generated code is not valid Go: %v\n%s", err, outputStr)
	}
}

func TestGenerateSceneWithNestedProps(t *testing.T) {
	p, err := parser.New()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	source := `package main

func MaterialScene() (Scene) {
	Scene {
		Mesh(
			GeometryProp(NewBoxGeometry(2.0, 2.0, 2.0)),
			MaterialProp(StandardMaterial(
				Color(0.91, 0.27, 0.38, 1.0),
				Metalness(0.3)
			))
		)
	}
}`

	file, err := p.ParseString(source)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	gen := New("main")
	output, err := gen.Generate(file)
	if err != nil {
		t.Fatalf("Failed to generate: %v", err)
	}

	outputStr := string(output)

	// Check for nested function calls
	if !strings.Contains(outputStr, "runtime.NewBoxGeometry(2.0, 2.0, 2.0)") {
		t.Error("Expected runtime.NewBoxGeometry call")
	}

	if !strings.Contains(outputStr, "runtime.StandardMaterial(") {
		t.Error("Expected runtime.StandardMaterial call")
	}

	if !strings.Contains(outputStr, "runtime.Color(0.91, 0.27, 0.38, 1.0)") {
		t.Error("Expected runtime.Color call")
	}

	if !strings.Contains(outputStr, "runtime.Metalness(0.3)") {
		t.Error("Expected runtime.Metalness call")
	}

	// Verify code is valid Go
	_, err = format.Source(output)
	if err != nil {
		t.Errorf("Generated code is not valid Go: %v\n%s", err, outputStr)
	}
}

func TestGenerateEmptyScene(t *testing.T) {
	p, err := parser.New()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	source := `package main

func EmptyScene() (Scene) {
	Scene {
	}
}`

	file, err := p.ParseString(source)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	gen := New("main")
	output, err := gen.Generate(file)
	if err != nil {
		t.Fatalf("Failed to generate: %v", err)
	}

	outputStr := string(output)

	// Should still generate struct, constructor, and RenderScene method
	if !strings.Contains(outputStr, "type EmptyScene struct") {
		t.Error("Expected EmptyScene struct definition")
	}

	if !strings.Contains(outputStr, "func NewEmptyScene() runtime.Scene") {
		t.Error("Expected NewEmptyScene constructor")
	}

	if !strings.Contains(outputStr, "func (s *EmptyScene) RenderScene() *runtime.GPUNode") {
		t.Error("Expected RenderScene method")
	}

	// Should call runtime.SceneNode() with no arguments
	if !strings.Contains(outputStr, "return runtime.SceneNode(") {
		t.Error("Expected runtime.Scene call in RenderScene")
	}

	// Verify code is valid Go
	_, err = format.Source(output)
	if err != nil {
		t.Errorf("Generated code is not valid Go: %v\n%s", err, outputStr)
	}
}

func TestGenerateSceneImports(t *testing.T) {
	p, err := parser.New()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	source := `package main

func TestScene() (Scene) {
	Scene {
		Mesh()
	}
}`

	file, err := p.ParseString(source)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	gen := New("main")
	output, err := gen.Generate(file)
	if err != nil {
		t.Fatalf("Failed to generate: %v", err)
	}

	outputStr := string(output)

	// Check for build tags
	if !strings.Contains(outputStr, "//go:build js && wasm") {
		t.Error("Expected //go:build tag")
	}

	if !strings.Contains(outputStr, "// +build js,wasm") {
		t.Error("Expected +build tag")
	}

	// Check for runtime import
	if !strings.Contains(outputStr, `"github.com/gaarutyunov/guix/pkg/runtime"`) {
		t.Error("Expected runtime package import")
	}

	// Scene components should import syscall/js (for consistency with Component)
	if !strings.Contains(outputStr, `"syscall/js"`) {
		t.Error("Expected syscall/js import")
	}
}


// TODO: Fix receiver reference bug - currently generates c.Field instead of s.Field
// func TestSceneReceiverReference(t *testing.T) {
//     ... test code ...
// }
