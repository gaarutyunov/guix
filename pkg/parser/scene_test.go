package parser

import (
	"testing"
)

func TestParseSceneWithNoProps(t *testing.T) {
	p, err := New()
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

	if len(file.Components) != 1 {
		t.Fatalf("Expected 1 component, got %d", len(file.Components))
	}

	comp := file.Components[0]
	if comp.Name != "TestScene" {
		t.Errorf("Expected component name 'TestScene', got '%s'", comp.Name)
	}

	if len(comp.Results) != 1 {
		t.Fatalf("Expected 1 result type, got %d", len(comp.Results))
	}

	if comp.Results[0].Name != "Scene" {
		t.Errorf("Expected result type 'Scene', got '%s'", comp.Results[0].Name)
	}

	if comp.Body == nil {
		t.Fatal("Expected body to be non-nil")
	}

	if len(comp.Body.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(comp.Body.Children))
	}

	child := comp.Body.Children[0]
	if child.Element == nil {
		t.Fatal("Expected element child")
	}

	if child.Element.Tag != "Scene" {
		t.Errorf("Expected Scene element, got '%s'", child.Element.Tag)
	}

	if len(child.Element.Props) != 0 {
		t.Errorf("Expected 0 props, got %d", len(child.Element.Props))
	}

	if len(child.Element.Children) != 1 {
		t.Fatalf("Expected 1 child element, got %d", len(child.Element.Children))
	}
}

func TestParseSceneWithSingleArgProp(t *testing.T) {
	p, err := New()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	source := `package main

func TestScene() (Scene) {
	Scene(Width(800)) {
		Mesh()
	}
}`

	file, err := p.ParseString(source)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(file.Components) != 1 {
		t.Fatalf("Expected 1 component, got %d", len(file.Components))
	}

	comp := file.Components[0]
	if comp.Body == nil {
		t.Fatal("Expected body to be non-nil")
	}

	if len(comp.Body.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(comp.Body.Children))
	}

	sceneElem := comp.Body.Children[0].Element
	if sceneElem == nil {
		t.Fatal("Expected Scene element")
	}

	if len(sceneElem.Props) != 1 {
		t.Fatalf("Expected 1 prop, got %d", len(sceneElem.Props))
	}

	prop := sceneElem.Props[0]
	if prop.Name != "Width" {
		t.Errorf("Expected prop name 'Width', got '%s'", prop.Name)
	}

	if len(prop.Args) != 1 {
		t.Fatalf("Expected 1 arg, got %d", len(prop.Args))
	}
}

func TestParseSceneWithMultiArgProps(t *testing.T) {
	p, err := New()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	source := `package main

func TestScene() (Scene) {
	Scene(Background(0.1, 0.1, 0.15, 1.0)) {
		Mesh()
	}
}`

	file, err := p.ParseString(source)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(file.Components) != 1 {
		t.Fatalf("Expected 1 component, got %d", len(file.Components))
	}

	comp := file.Components[0]
	if comp.Body == nil {
		t.Fatal("Expected body to be non-nil")
	}

	if len(comp.Body.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(comp.Body.Children))
	}

	sceneElem := comp.Body.Children[0].Element
	if sceneElem == nil {
		t.Fatal("Expected Scene element")
	}

	if len(sceneElem.Props) != 1 {
		t.Fatalf("Expected 1 prop, got %d", len(sceneElem.Props))
	}

	prop := sceneElem.Props[0]
	if prop.Name != "Background" {
		t.Errorf("Expected prop name 'Background', got '%s'", prop.Name)
	}

	if len(prop.Args) != 4 {
		t.Fatalf("Expected 4 args, got %d", len(prop.Args))
	}
}

func TestParseSceneWithNestedElements(t *testing.T) {
	p, err := New()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	source := `package main

func CubeScene(rotX float32, rotY float32) (Scene) {
	Scene(Background(0.1, 0.1, 0.15, 1.0)) {
		Mesh(
			Position(0, 0, 0),
			Rotation(rotX, rotY, 0)
		)

		PerspectiveCamera(
			FOV(60),
			Position(0, 2, 6)
		)
	}
}`

	file, err := p.ParseString(source)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(file.Components) != 1 {
		t.Fatalf("Expected 1 component, got %d", len(file.Components))
	}

	comp := file.Components[0]
	if comp.Name != "CubeScene" {
		t.Errorf("Expected component name 'CubeScene', got '%s'", comp.Name)
	}

	if len(comp.Params) != 2 {
		t.Fatalf("Expected 2 params, got %d", len(comp.Params))
	}

	sceneElem := comp.Body.Children[0].Element
	if sceneElem == nil {
		t.Fatal("Expected Scene element")
	}

	// Scene should have 2 children: Mesh and PerspectiveCamera
	if len(sceneElem.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(sceneElem.Children))
	}

	// Check Mesh element
	meshNode := sceneElem.Children[0]
	if meshNode.Element == nil {
		t.Fatal("Expected Mesh element")
	}
	if meshNode.Element.Tag != "Mesh" {
		t.Errorf("Expected Mesh element, got '%s'", meshNode.Element.Tag)
	}
	if len(meshNode.Element.Props) != 2 {
		t.Errorf("Expected 2 props on Mesh, got %d", len(meshNode.Element.Props))
	}

	// Check PerspectiveCamera element
	cameraNode := sceneElem.Children[1]
	if cameraNode.Element == nil {
		t.Fatal("Expected PerspectiveCamera element")
	}
	if cameraNode.Element.Tag != "PerspectiveCamera" {
		t.Errorf("Expected PerspectiveCamera element, got '%s'", cameraNode.Element.Tag)
	}
	if len(cameraNode.Element.Props) != 2 {
		t.Errorf("Expected 2 props on PerspectiveCamera, got %d", len(cameraNode.Element.Props))
	}
}

func TestParseSceneWithComponentParams(t *testing.T) {
	p, err := New()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	source := `package main

func ParameterizedScene(width float32, height float32, bgColor string) (Scene) {
	Scene(Width(width), Height(height)) {
		Mesh()
	}
}`

	file, err := p.ParseString(source)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	comp := file.Components[0]
	if len(comp.Params) != 3 {
		t.Fatalf("Expected 3 params, got %d", len(comp.Params))
	}

	sceneElem := comp.Body.Children[0].Element
	if len(sceneElem.Props) != 2 {
		t.Fatalf("Expected 2 props, got %d", len(sceneElem.Props))
	}

	// Verify prop names
	if sceneElem.Props[0].Name != "Width" {
		t.Errorf("Expected first prop 'Width', got '%s'", sceneElem.Props[0].Name)
	}
	if sceneElem.Props[1].Name != "Height" {
		t.Errorf("Expected second prop 'Height', got '%s'", sceneElem.Props[1].Name)
	}
}

func TestParseEmptyScene(t *testing.T) {
	p, err := New()
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

	comp := file.Components[0]
	sceneElem := comp.Body.Children[0].Element

	if len(sceneElem.Children) != 0 {
		t.Errorf("Expected 0 children, got %d", len(sceneElem.Children))
	}
}
