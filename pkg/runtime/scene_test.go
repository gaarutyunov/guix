//go:build js && wasm
// +build js,wasm

package runtime

import (
	"testing"
)

// MockScene is a test implementation of the Scene interface
type MockScene struct {
	rendered bool
}

func (m *MockScene) RenderScene() *GPUNode {
	m.rendered = true
	mesh := Mesh()
	sceneNode := SceneNode(mesh)
	return sceneNode
}

func TestSceneInterface(t *testing.T) {
	scene := &MockScene{}

	// Verify it implements Scene interface
	var _ Scene = scene

	// Call RenderScene
	node := scene.RenderScene()

	if !scene.rendered {
		t.Error("Expected RenderScene to set rendered flag")
	}

	if node == nil {
		t.Fatal("Expected non-nil GPUNode")
	}

	if node.Type != SceneNodeType {
		t.Errorf("Expected SceneNodeType, got %d", node.Type)
	}
}

func TestSceneBuilderNoArgs(t *testing.T) {
	node := SceneNode()

	if node == nil {
		t.Fatal("Expected non-nil GPUNode")
	}

	if node.Type != SceneNodeType {
		t.Errorf("Expected SceneNodeType, got %d", node.Type)
	}

	if node.Tag != "scene" {
		t.Errorf("Expected tag 'scene', got '%s'", node.Tag)
	}

	if len(node.Children) != 0 {
		t.Errorf("Expected 0 children, got %d", len(node.Children))
	}

	if node.Properties == nil {
		t.Error("Expected non-nil Properties map")
	}
}

func TestSceneBuilderWithChildren(t *testing.T) {
	mesh := Mesh()
	camera := PerspectiveCamera()

	node := SceneNode(mesh, camera)

	if len(node.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(node.Children))
	}

	if node.Children[0] != mesh {
		t.Error("Expected first child to be mesh")
	}

	if node.Children[1] != camera {
		t.Error("Expected second child to be camera")
	}
}

func TestSceneBuilderWithProps(t *testing.T) {
	bg := Background(0.1, 0.2, 0.3, 1.0)

	node := SceneNode(bg)

	if len(node.Properties) == 0 {
		t.Error("Expected properties to be set")
	}

	bgVal, ok := node.Properties["background"]
	if !ok {
		t.Error("Expected background property to be set")
	}

	vec, ok := bgVal.(*Vec4)
	if !ok {
		t.Error("Expected background to be Vec4")
	}

	if vec.X != 0.1 || vec.Y != 0.2 || vec.Z != 0.3 || vec.W != 1.0 {
		t.Errorf("Expected background (0.1, 0.2, 0.3, 1.0), got (%f, %f, %f, %f)",
			vec.X, vec.Y, vec.Z, vec.W)
	}
}

func TestSceneBuilderWithPropsAndChildren(t *testing.T) {
	bg := Background(0.5, 0.5, 0.5, 1.0)
	mesh := Mesh()
	light := AmbientLight()

	node := SceneNode(bg, mesh, light)

	// Check property
	if _, ok := node.Properties["background"]; !ok {
		t.Error("Expected background property")
	}

	// Check children
	if len(node.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(node.Children))
	}

	if node.Children[0] != mesh {
		t.Error("Expected first child to be mesh")
	}

	if node.Children[1] != light {
		t.Error("Expected second child to be light")
	}
}

func TestBackgroundProp(t *testing.T) {
	prop := Background(0.1, 0.2, 0.3, 0.8)

	if prop.Key != "background" {
		t.Errorf("Expected key 'background', got '%s'", prop.Key)
	}

	vec, ok := prop.Value.(*Vec4)
	if !ok {
		t.Fatal("Expected Vec4 value")
	}

	if vec.X != 0.1 {
		t.Errorf("Expected X=0.1, got %f", vec.X)
	}
	if vec.Y != 0.2 {
		t.Errorf("Expected Y=0.2, got %f", vec.Y)
	}
	if vec.Z != 0.3 {
		t.Errorf("Expected Z=0.3, got %f", vec.Z)
	}
	if vec.W != 0.8 {
		t.Errorf("Expected W=0.8, got %f", vec.W)
	}
}

func TestMeshBuilder(t *testing.T) {
	node := Mesh()

	if node == nil {
		t.Fatal("Expected non-nil GPUNode")
	}

	if node.Type != MeshNodeType {
		t.Errorf("Expected MeshNodeType, got %d", node.Type)
	}

	if node.Tag != "mesh" {
		t.Errorf("Expected tag 'mesh', got '%s'", node.Tag)
	}
}

func TestMeshWithProps(t *testing.T) {
	pos := Position(1, 2, 3)
	rot := Rotation(0.5, 0.5, 0)

	mesh := Mesh(pos, rot)

	if len(mesh.Properties) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(mesh.Properties))
	}

	posVec, ok := mesh.Properties["position"].(*Vec3)
	if !ok {
		t.Fatal("Expected position to be Vec3")
	}

	if posVec.X != 1 || posVec.Y != 2 || posVec.Z != 3 {
		t.Errorf("Expected position (1, 2, 3), got (%f, %f, %f)",
			posVec.X, posVec.Y, posVec.Z)
	}
}

func TestPerspectiveCameraBuilder(t *testing.T) {
	node := PerspectiveCamera()

	if node == nil {
		t.Fatal("Expected non-nil GPUNode")
	}

	if node.Type != CameraNodeType {
		t.Errorf("Expected CameraNodeType, got %d", node.Type)
	}

	if node.Tag != "perspective_camera" {
		t.Errorf("Expected tag 'perspective_camera', got '%s'", node.Tag)
	}
}

func TestAmbientLightBuilder(t *testing.T) {
	node := AmbientLight()

	if node == nil {
		t.Fatal("Expected non-nil GPUNode")
	}

	if node.Type != LightNodeType {
		t.Errorf("Expected LightNodeType, got %d", node.Type)
	}

	if node.Tag != "ambient_light" {
		t.Errorf("Expected tag 'ambient_light', got '%s'", node.Tag)
	}
}

func TestDirectionalLightBuilder(t *testing.T) {
	node := DirectionalLight()

	if node == nil {
		t.Fatal("Expected non-nil GPUNode")
	}

	if node.Type != LightNodeType {
		t.Errorf("Expected LightNodeType, got %d", node.Type)
	}

	if node.Tag != "directional_light" {
		t.Errorf("Expected tag 'directional_light', got '%s'", node.Tag)
	}
}

func TestComplexSceneGraph(t *testing.T) {
	// Build a complex scene graph like the cube example
	scene := SceneNode(
		Background(0.1, 0.1, 0.15, 1.0),
		Mesh(
			Position(0, 0, 0),
			Rotation(0.5, 0.5, 0),
			ScaleValue(1, 1, 1),
		),
		PerspectiveCamera(
			FOV(1.047), // ~60 degrees
			Near(0.1),
			Far(100),
			Position(0, 2, 6),
		),
		AmbientLight(
			Color(1, 1, 1, 1),
			Intensity(0.4),
		),
		DirectionalLight(
			Position(5, 10, 7),
			Color(1, 1, 1, 1),
			Intensity(0.8),
		),
	)

	// Verify structure
	if scene.Type != SceneNodeType {
		t.Error("Expected SceneNodeType")
	}

	// Should have background prop + 4 children
	if len(scene.Properties) == 0 {
		t.Error("Expected background property")
	}

	if len(scene.Children) != 4 {
		t.Fatalf("Expected 4 children (mesh, camera, 2 lights), got %d", len(scene.Children))
	}

	// Verify children types
	if scene.Children[0].Type != MeshNodeType {
		t.Error("Expected first child to be Mesh")
	}

	if scene.Children[1].Type != CameraNodeType {
		t.Error("Expected second child to be Camera")
	}

	if scene.Children[2].Type != LightNodeType {
		t.Error("Expected third child to be Light")
	}

	if scene.Children[3].Type != LightNodeType {
		t.Error("Expected fourth child to be Light")
	}
}
