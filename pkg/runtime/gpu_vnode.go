//go:build js && wasm

package runtime

import (
	"syscall/js"
)

// GPUNodeType represents the type of a GPU node
type GPUNodeType uint8

const (
	// GPUCanvasNodeType represents a WebGPU canvas
	GPUCanvasNodeType GPUNodeType = iota
	// SceneNodeType represents a 3D scene
	SceneNodeType
	// MeshNodeType represents a 3D mesh
	MeshNodeType
	// CameraNodeType represents a camera
	CameraNodeType
	// LightNodeType represents a light
	LightNodeType
	// GroupNodeType represents a group/container
	GroupNodeType
)

// GPUNode represents a GPU/3D scene graph node
type GPUNode struct {
	Type         GPUNodeType
	Tag          string                 // Node type identifier
	Properties   map[string]interface{} // Node properties
	Transform    Transform              // 3D transformation
	Children     []*GPUNode             // Child nodes
	RenderFunc   func(*GPUCanvas, float64) // Custom render function
	InitFunc     func(*GPUCanvas)       // Initialization function
	Canvas       *GPUCanvas             // Canvas reference (for canvas nodes)
	Material     *Material              // Material (for mesh nodes)
	Geometry     Geometry               // Geometry (for mesh nodes)
	Camera       *Camera                // Camera (for camera nodes)
	Light        *Light                 // Light (for light nodes)
}

// Geometry interface for different geometry types
type Geometry interface {
	GetVertices() []float32
	GetIndices() []uint16
}

// Material represents a material with color and properties
type Material struct {
	Color      Vec4    // RGBA color
	Metalness  float32 // 0-1
	Roughness  float32 // 0-1
	Emissive   Vec3    // Emissive color
	Shader     *CustomShader // Custom shader override
}

// CustomShader represents custom vertex/fragment shaders
type CustomShader struct {
	VertexShader   string
	FragmentShader string
	VertexEntry    string
	FragmentEntry  string
}

// Light represents a light source
type Light struct {
	Type      string  // "ambient", "directional", "point", "spot"
	Color     Vec3    // RGB color
	Intensity float32 // Light intensity
	Position  Vec3    // Position (for point/spot)
	Direction Vec3    // Direction (for directional/spot)
}

// GPU property types

// GPUProp represents a GPU node property
type GPUProp struct {
	Key   string
	Value interface{}
}

// Width sets the width property for GPU canvas
func Width(value int) GPUProp {
	return GPUProp{Key: "width", Value: value}
}

// Height sets the height property for GPU canvas
func Height(value int) GPUProp {
	return GPUProp{Key: "height", Value: value}
}

// Position sets the position for a 3D node
func Position(x, y, z float32) GPUProp {
	return GPUProp{Key: "position", Value: NewVec3(x, y, z)}
}

// Rotation sets the rotation for a 3D node (in radians)
func Rotation(x, y, z float32) GPUProp {
	return GPUProp{Key: "rotation", Value: NewVec3(x, y, z)}
}

// ScaleValue sets the scale for a 3D node
func ScaleValue(x, y, z float32) GPUProp {
	return GPUProp{Key: "scale", Value: NewVec3(x, y, z)}
}

// Color sets the color property
func Color(r, g, b, a float32) GPUProp {
	return GPUProp{Key: "color", Value: NewVec4(r, g, b, a)}
}

// Metalness sets the metalness property
func Metalness(value float32) GPUProp {
	return GPUProp{Key: "metalness", Value: value}
}

// Roughness sets the roughness property
func Roughness(value float32) GPUProp {
	return GPUProp{Key: "roughness", Value: value}
}

// Intensity sets light intensity
func Intensity(value float32) GPUProp {
	return GPUProp{Key: "intensity", Value: value}
}

// FOV sets camera field of view (in radians)
func FOV(value float32) GPUProp {
	return GPUProp{Key: "fov", Value: value}
}

// Near sets camera near plane
func Near(value float32) GPUProp {
	return GPUProp{Key: "near", Value: value}
}

// Far sets camera far plane
func Far(value float32) GPUProp {
	return GPUProp{Key: "far", Value: value}
}

// LookAtPos sets camera look-at target
func LookAtPos(x, y, z float32) GPUProp {
	return GPUProp{Key: "lookAt", Value: NewVec3(x, y, z)}
}

// Background sets scene background color
func Background(r, g, b, a float32) GPUProp {
	return GPUProp{Key: "background", Value: NewVec4(r, g, b, a)}
}

// OnGPUInit sets an initialization function for GPU canvas
func OnGPUInit(fn func(*GPUCanvas)) GPUProp {
	return GPUProp{Key: "onInit", Value: fn}
}

// OnGPURender sets a render function for GPU canvas
func OnGPURender(fn func(*GPUCanvas, float64)) GPUProp {
	return GPUProp{Key: "onRender", Value: fn}
}

// GeometryProp wraps geometry for property passing
func GeometryProp(geom Geometry) GPUProp {
	return GPUProp{Key: "geometry", Value: geom}
}

// MaterialProp wraps material for property passing
func MaterialProp(mat *Material) GPUProp {
	return GPUProp{Key: "material", Value: mat}
}

// GPU Node Builders

// GPUCanvas creates a WebGPU canvas element
func GPUCanvas(options ...interface{}) *VNode {
	// Create a regular VNode that will contain a canvas element
	node := &VNode{
		Type:       ElementNode,
		Tag:        "canvas",
		Attributes: make(map[string]string),
		Properties: make(map[string]interface{}),
		Events:     make(map[string]EventHandler),
	}

	// Default properties
	width := 800
	height := 600
	var initFunc func(*GPUCanvas)
	var renderFunc func(*GPUCanvas, float64)

	// Process options
	for _, opt := range options {
		switch o := opt.(type) {
		case GPUProp:
			switch o.Key {
			case "width":
				if w, ok := o.Value.(int); ok {
					width = w
				}
			case "height":
				if h, ok := o.Value.(int); ok {
					height = h
				}
			case "onInit":
				if fn, ok := o.Value.(func(*GPUCanvas)); ok {
					initFunc = fn
				}
			case "onRender":
				if fn, ok := o.Value.(func(*GPUCanvas, float64)); ok {
					renderFunc = fn
				}
			}
		case *VNode:
			// GPU children (Scene, Mesh, etc.) - we'll handle these differently
			node.Children = append(node.Children, o)
		case *GPUNode:
			// Convert GPUNode to VNode representation if needed
			// For now, store as property
			if node.Properties["gpuChildren"] == nil {
				node.Properties["gpuChildren"] = []*GPUNode{}
			}
			children := node.Properties["gpuChildren"].([]*GPUNode)
			node.Properties["gpuChildren"] = append(children, o)
		}
	}

	// Store GPU-specific properties
	node.Properties["gpuWidth"] = width
	node.Properties["gpuHeight"] = height
	node.Properties["gpuInitFunc"] = initFunc
	node.Properties["gpuRenderFunc"] = renderFunc
	node.Attributes["width"] = string(rune(width))
	node.Attributes["height"] = string(rune(height))
	node.Attributes["class"] = "guix-gpu-canvas"

	return node
}

// Scene creates a 3D scene node
func Scene(options ...interface{}) *GPUNode {
	node := &GPUNode{
		Type:       SceneNodeType,
		Tag:        "scene",
		Properties: make(map[string]interface{}),
		Transform:  NewTransform(),
		Children:   make([]*GPUNode, 0),
	}

	for _, opt := range options {
		switch o := opt.(type) {
		case *GPUNode:
			node.Children = append(node.Children, o)
		case GPUProp:
			node.Properties[o.Key] = o.Value
		}
	}

	return node
}

// Mesh creates a 3D mesh node
func Mesh(options ...interface{}) *GPUNode {
	node := &GPUNode{
		Type:       MeshNodeType,
		Tag:        "mesh",
		Properties: make(map[string]interface{}),
		Transform:  NewTransform(),
	}

	for _, opt := range options {
		switch o := opt.(type) {
		case GPUProp:
			switch o.Key {
			case "position":
				if v, ok := o.Value.(Vec3); ok {
					node.Transform.Position = v
				}
			case "rotation":
				if v, ok := o.Value.(Vec3); ok {
					node.Transform.Rotation = v
				}
			case "scale":
				if v, ok := o.Value.(Vec3); ok {
					node.Transform.Scale = v
				}
			case "geometry":
				if g, ok := o.Value.(Geometry); ok {
					node.Geometry = g
				}
			case "material":
				if m, ok := o.Value.(*Material); ok {
					node.Material = m
				}
			default:
				node.Properties[o.Key] = o.Value
			}
		}
	}

	return node
}

// PerspectiveCamera creates a perspective camera node
func PerspectiveCamera(options ...interface{}) *GPUNode {
	node := &GPUNode{
		Type:       CameraNodeType,
		Tag:        "camera",
		Properties: make(map[string]interface{}),
		Transform:  NewTransform(),
		Camera:     &Camera{
			Position: Vec3{0, 0, 5},
			Target:   Vec3{0, 0, 0},
			Up:       Vec3{0, 1, 0},
			FOV:      DegreesToRadians(60),
			Aspect:   1.0,
			Near:     0.1,
			Far:      100,
		},
	}

	for _, opt := range options {
		switch o := opt.(type) {
		case GPUProp:
			switch o.Key {
			case "position":
				if v, ok := o.Value.(Vec3); ok {
					node.Camera.Position = v
					node.Transform.Position = v
				}
			case "lookAt":
				if v, ok := o.Value.(Vec3); ok {
					node.Camera.Target = v
				}
			case "fov":
				if v, ok := o.Value.(float32); ok {
					node.Camera.FOV = v
				}
			case "near":
				if v, ok := o.Value.(float32); ok {
					node.Camera.Near = v
				}
			case "far":
				if v, ok := o.Value.(float32); ok {
					node.Camera.Far = v
				}
			default:
				node.Properties[o.Key] = o.Value
			}
		}
	}

	return node
}

// AmbientLight creates an ambient light node
func AmbientLight(options ...interface{}) *GPUNode {
	node := &GPUNode{
		Type:       LightNodeType,
		Tag:        "ambient-light",
		Properties: make(map[string]interface{}),
		Light: &Light{
			Type:      "ambient",
			Color:     Vec3{1, 1, 1},
			Intensity: 0.5,
		},
	}

	for _, opt := range options {
		switch o := opt.(type) {
		case GPUProp:
			switch o.Key {
			case "color":
				if v, ok := o.Value.(Vec4); ok {
					node.Light.Color = Vec3{v.X, v.Y, v.Z}
				} else if v, ok := o.Value.(Vec3); ok {
					node.Light.Color = v
				}
			case "intensity":
				if v, ok := o.Value.(float32); ok {
					node.Light.Intensity = v
				}
			default:
				node.Properties[o.Key] = o.Value
			}
		}
	}

	return node
}

// DirectionalLight creates a directional light node
func DirectionalLight(options ...interface{}) *GPUNode {
	node := &GPUNode{
		Type:       LightNodeType,
		Tag:        "directional-light",
		Properties: make(map[string]interface{}),
		Transform:  NewTransform(),
		Light: &Light{
			Type:      "directional",
			Color:     Vec3{1, 1, 1},
			Intensity: 1.0,
			Position:  Vec3{5, 10, 7},
			Direction: Vec3{-1, -1, -1},
		},
	}

	for _, opt := range options {
		switch o := opt.(type) {
		case GPUProp:
			switch o.Key {
			case "position":
				if v, ok := o.Value.(Vec3); ok {
					node.Light.Position = v
					node.Transform.Position = v
				}
			case "color":
				if v, ok := o.Value.(Vec4); ok {
					node.Light.Color = Vec3{v.X, v.Y, v.Z}
				} else if v, ok := o.Value.(Vec3); ok {
					node.Light.Color = v
				}
			case "intensity":
				if v, ok := o.Value.(float32); ok {
					node.Light.Intensity = v
				}
			default:
				node.Properties[o.Key] = o.Value
			}
		}
	}

	return node
}

// PointLight creates a point light node
func PointLight(options ...interface{}) *GPUNode {
	node := &GPUNode{
		Type:       LightNodeType,
		Tag:        "point-light",
		Properties: make(map[string]interface{}),
		Transform:  NewTransform(),
		Light: &Light{
			Type:      "point",
			Color:     Vec3{1, 1, 1},
			Intensity: 1.0,
			Position:  Vec3{0, 5, 0},
		},
	}

	for _, opt := range options {
		switch o := opt.(type) {
		case GPUProp:
			switch o.Key {
			case "position":
				if v, ok := o.Value.(Vec3); ok {
					node.Light.Position = v
					node.Transform.Position = v
				}
			case "color":
				if v, ok := o.Value.(Vec4); ok {
					node.Light.Color = Vec3{v.X, v.Y, v.Z}
				} else if v, ok := o.Value.(Vec3); ok {
					node.Light.Color = v
				}
			case "intensity":
				if v, ok := o.Value.(float32); ok {
					node.Light.Intensity = v
				}
			default:
				node.Properties[o.Key] = o.Value
			}
		}
	}

	return node
}

// Group creates a container for grouping nodes
func Group(options ...interface{}) *GPUNode {
	node := &GPUNode{
		Type:       GroupNodeType,
		Tag:        "group",
		Properties: make(map[string]interface{}),
		Transform:  NewTransform(),
		Children:   make([]*GPUNode, 0),
	}

	for _, opt := range options {
		switch o := opt.(type) {
		case *GPUNode:
			node.Children = append(node.Children, o)
		case GPUProp:
			switch o.Key {
			case "position":
				if v, ok := o.Value.(Vec3); ok {
					node.Transform.Position = v
				}
			case "rotation":
				if v, ok := o.Value.(Vec3); ok {
					node.Transform.Rotation = v
				}
			case "scale":
				if v, ok := o.Value.(Vec3); ok {
					node.Transform.Scale = v
				}
			default:
				node.Properties[o.Key] = o.Value
			}
		}
	}

	return node
}

// Helper functions for materials and geometries

// StandardMaterial creates a standard PBR material
func StandardMaterial(options ...interface{}) *Material {
	mat := &Material{
		Color:     Vec4{1, 1, 1, 1},
		Metalness: 0.0,
		Roughness: 0.5,
		Emissive:  Vec3{0, 0, 0},
	}

	for _, opt := range options {
		if prop, ok := opt.(GPUProp); ok {
			switch prop.Key {
			case "color":
				if v, ok := prop.Value.(Vec4); ok {
					mat.Color = v
				}
			case "metalness":
				if v, ok := prop.Value.(float32); ok {
					mat.Metalness = v
				}
			case "roughness":
				if v, ok := prop.Value.(float32); ok {
					mat.Roughness = v
				}
			}
		}
	}

	return mat
}

// BoxGeometryNode creates box geometry
func BoxGeometryNode(width, height, depth float32) Geometry {
	return NewBoxGeometry(width, height, depth)
}

// PlaneGeometryNode creates plane geometry
func PlaneGeometryNode(width, height float32) Geometry {
	return NewPlaneGeometry(width, height)
}

// SphereGeometryNode creates sphere geometry
func SphereGeometryNode(radius float32, widthSegments, heightSegments int) Geometry {
	return NewSphereGeometry(radius, widthSegments, heightSegments)
}

// InitGPUCanvas initializes a GPU canvas from a VNode
func InitGPUCanvas(node *VNode) (*GPUCanvas, error) {
	width := 800
	height := 600

	if w, ok := node.Properties["gpuWidth"].(int); ok {
		width = w
	}
	if h, ok := node.Properties["gpuHeight"].(int); ok {
		height = h
	}

	config := GPUCanvasConfig{
		Width:  width,
		Height: height,
		DevicePixelRatio: 1.0,
		AlphaMode: "premultiplied",
		FrameLoop: "always",
	}

	canvas, err := CreateGPUCanvas(config)
	if err != nil {
		return nil, err
	}

	// Run init function if provided
	if initFunc, ok := node.Properties["gpuInitFunc"].(func(*GPUCanvas)); ok && initFunc != nil {
		initFunc(canvas)
	}

	// Set render function if provided
	if renderFunc, ok := node.Properties["gpuRenderFunc"].(func(*GPUCanvas, float64)); ok && renderFunc != nil {
		canvas.SetRenderFunc(renderFunc)
	}

	return canvas, nil
}

// AttachGPUCanvas attaches a GPU canvas to a DOM node
func AttachGPUCanvas(domNode js.Value, gpuCanvas *GPUCanvas) {
	// Replace the placeholder canvas with the GPU canvas
	if domNode.Truthy() && gpuCanvas.Canvas.Truthy() {
		parent := domNode.Get("parentNode")
		if parent.Truthy() {
			parent.Call("replaceChild", gpuCanvas.Canvas, domNode)
		}
	}
}
