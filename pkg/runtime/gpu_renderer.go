//go:build js && wasm

package runtime

import (
	"fmt"
	"syscall/js"
)

// SceneRenderer manages rendering of a 3D scene
type SceneRenderer struct {
	Canvas        *GPUCanvas
	Scene         *GPUNode
	ActiveCamera  *Camera
	Pipeline      *RenderPipeline
	UniformBuffer *GPUBuffer
	DepthTexture  js.Value
	Meshes        []*MeshInstance
	Lights        []*Light
	AmbientLight  *Light
}

// MeshInstance represents an instantiated mesh with buffers
type MeshInstance struct {
	Transform    Transform
	Geometry     Geometry
	Material     *Material
	VertexBuffer *GPUBuffer
	IndexBuffer  *GPUBuffer
	IndexCount   int
}

// NewSceneRenderer creates a new scene renderer
func NewSceneRenderer(canvas *GPUCanvas, scene *GPUNode) (*SceneRenderer, error) {
	log("[Renderer] Creating scene renderer")

	if canvas == nil {
		logError("[Renderer] Canvas is nil")
		return nil, fmt.Errorf("canvas is nil")
	}
	if scene == nil {
		logError("[Renderer] Scene is nil")
		return nil, fmt.Errorf("scene is nil")
	}

	renderer := &SceneRenderer{
		Canvas: canvas,
		Scene:  scene,
		Meshes: make([]*MeshInstance, 0),
		Lights: make([]*Light, 0),
	}

	// Extract scene data
	log("[Renderer] Building scene graph")
	if err := renderer.buildScene(scene); err != nil {
		logError(fmt.Sprintf("[Renderer] Failed to build scene: %v", err))
		return nil, err
	}
	log(fmt.Sprintf("[Renderer] Scene built: %d meshes, %d lights", len(renderer.Meshes), len(renderer.Lights)))

	// Create depth texture
	log("[Renderer] Creating depth texture")
	depthTexture, err := canvas.CreateDepthTexture()
	if err != nil {
		logError(fmt.Sprintf("[Renderer] Failed to create depth texture: %v", err))
		return nil, fmt.Errorf("failed to create depth texture: %w", err)
	}
	renderer.DepthTexture = depthTexture

	// Create uniform buffer for MVP matrix
	log("[Renderer] Creating uniform buffer")
	uniformBuffer, err := CreateUniformBuffer(canvas.GPUContext, 256, "mvp-uniforms")
	if err != nil {
		logError(fmt.Sprintf("[Renderer] Failed to create uniform buffer: %v", err))
		return nil, fmt.Errorf("failed to create uniform buffer: %w", err)
	}
	renderer.UniformBuffer = uniformBuffer

	// Create render pipeline
	log("[Renderer] Creating render pipeline")
	if err := renderer.createPipeline(); err != nil {
		logError(fmt.Sprintf("[Renderer] Failed to create pipeline: %v", err))
		return nil, err
	}

	log("[Renderer] Scene renderer created successfully")
	return renderer, nil
}

// buildScene traverses the scene graph and extracts renderable objects
func (sr *SceneRenderer) buildScene(node *GPUNode) error {
	if node == nil {
		return nil
	}

	switch node.Type {
	case MeshNodeType:
		// Create mesh instance
		if node.Geometry != nil {
			mesh, err := sr.createMeshInstance(node)
			if err != nil {
				logError(fmt.Sprintf("Failed to create mesh: %v", err))
			} else {
				sr.Meshes = append(sr.Meshes, mesh)
			}
		}

	case CameraNodeType:
		// Set active camera
		if node.Camera != nil {
			sr.ActiveCamera = node.Camera
			sr.ActiveCamera.Aspect = sr.Canvas.GetAspectRatio()
		}

	case LightNodeType:
		// Add light
		if node.Light != nil {
			if node.Light.Type == "ambient" {
				sr.AmbientLight = node.Light
			} else {
				sr.Lights = append(sr.Lights, node.Light)
			}
		}
	}

	// Recursively process children
	for _, child := range node.Children {
		if err := sr.buildScene(child); err != nil {
			return err
		}
	}

	return nil
}

// createMeshInstance creates GPU buffers for a mesh
func (sr *SceneRenderer) createMeshInstance(node *GPUNode) (*MeshInstance, error) {
	vertices := node.Geometry.GetVertices()
	indices := node.Geometry.GetIndices()

	// Create vertex buffer
	vertexBuffer, err := CreateVertexBuffer(sr.Canvas.GPUContext, vertices, "mesh-vertices")
	if err != nil {
		return nil, fmt.Errorf("failed to create vertex buffer: %w", err)
	}

	// Create index buffer
	indexBuffer, err := CreateIndexBuffer(sr.Canvas.GPUContext, indices, "mesh-indices")
	if err != nil {
		return nil, fmt.Errorf("failed to create index buffer: %w", err)
	}

	material := node.Material
	if material == nil {
		// Default material
		material = &Material{
			Color:     Vec4{0.8, 0.8, 0.8, 1.0},
			Metalness: 0.0,
			Roughness: 0.5,
		}
	}

	return &MeshInstance{
		Transform:    node.Transform,
		Geometry:     node.Geometry,
		Material:     material,
		VertexBuffer: vertexBuffer,
		IndexBuffer:  indexBuffer,
		IndexCount:   len(indices),
	}, nil
}

// createPipeline creates the render pipeline with shaders
func (sr *SceneRenderer) createPipeline() error {
	ctx := sr.Canvas.GPUContext

	// Create vertex shader
	vertexShaderModule, err := CreateShaderModule(ctx, VertexShaderWithMVP, "vertex-shader")
	if err != nil {
		return fmt.Errorf("failed to create vertex shader: %w", err)
	}

	// Create fragment shader
	fragmentShaderModule, err := CreateShaderModule(ctx, FragmentShaderWithLighting, "fragment-shader")
	if err != nil {
		return fmt.Errorf("failed to create fragment shader: %w", err)
	}

	// Create vertex buffer layout
	vertexBufferLayout := CreateVertexBufferLayout(24, []VertexAttribute{
		{Format: VertexFormatFloat32x3, Offset: 0, ShaderLocation: 0},  // position
		{Format: VertexFormatFloat32x3, Offset: 12, ShaderLocation: 1}, // normal
	})

	// Create bind group layout for uniforms
	bindGroupLayoutEntries := []map[string]interface{}{
		CreateBindGroupLayoutEntry(0, GPUShaderStageVertex, "uniform"),
	}

	bindGroupLayout, err := CreateBindGroupLayout(ctx, bindGroupLayoutEntries, "uniform-bind-group-layout")
	if err != nil {
		return fmt.Errorf("failed to create bind group layout: %w", err)
	}

	// Create pipeline
	config := PipelineConfig{
		Label:              "scene-pipeline",
		VertexShader:       vertexShaderModule.Module,
		FragmentShader:     fragmentShaderModule.Module,
		VertexEntryPoint:   "vs_main",
		FragmentEntryPoint: "fs_main",
		VertexBuffers:      []map[string]interface{}{vertexBufferLayout},
		ColorFormat:        sr.Canvas.Format,
		DepthFormat:        "depth24plus",
		PrimitiveTopology:  PrimitiveTopologyTriangleList,
		CullMode:           CullModeBack,
		BindGroupLayouts:   []js.Value{bindGroupLayout},
	}

	pipeline, err := CreateRenderPipeline(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to create pipeline: %w", err)
	}

	sr.Pipeline = pipeline

	return nil
}

// Render renders the scene
func (sr *SceneRenderer) Render() {
	if sr.ActiveCamera == nil {
		logError("No active camera in scene")
		return
	}

	ctx := sr.Canvas.GPUContext

	// Create command encoder
	encoder, err := ctx.CreateCommandEncoder("scene-encoder")
	if err != nil {
		logError(fmt.Sprintf("Failed to create encoder: %v", err))
		return
	}

	// Get background color from scene properties
	bgColor := [4]float32{0.1, 0.1, 0.15, 1.0}
	if sr.Scene != nil && sr.Scene.Properties != nil {
		if bg, ok := sr.Scene.Properties["background"].(Vec4); ok {
			bgColor = [4]float32{bg.X, bg.Y, bg.Z, bg.W}
		}
	}

	// Begin render pass
	renderPass := sr.Canvas.BeginRenderPassWithDepth(
		encoder,
		bgColor,
		sr.DepthTexture,
		"clear",
		1.0,
	)

	if !renderPass.Truthy() {
		logError("Failed to begin render pass")
		return
	}

	// Set pipeline
	renderPass.Call("setPipeline", sr.Pipeline.Pipeline)

	// Update camera aspect ratio
	sr.ActiveCamera.Aspect = sr.Canvas.GetAspectRatio()

	// Get view-projection matrix
	viewProjection := sr.ActiveCamera.ViewProjectionMatrix()

	// Render each mesh
	for _, mesh := range sr.Meshes {
		// Calculate model-view-projection matrix
		model := mesh.Transform.Matrix()
		mvp := viewProjection.Multiply(model)

		// Update uniform buffer
		if err := sr.UniformBuffer.Write(ctx, 0, mvp.ToBytes()); err != nil {
			logError(fmt.Sprintf("Failed to write uniforms: %v", err))
			continue
		}

		// Create bind group for this mesh
		// Note: WebGPU requires a GPUBufferBinding object (with buffer, offset, size) not just the buffer
		bufferBinding := CreateBufferBinding(sr.UniformBuffer.Buffer, 0, sr.UniformBuffer.Size)
		bindGroupEntries := []map[string]interface{}{
			CreateBindGroupEntry(0, bufferBinding),
		}

		bindGroup, err := CreateBindGroup(
			ctx,
			sr.Pipeline.Pipeline.Call("getBindGroupLayout", 0),
			bindGroupEntries,
			"mesh-bind-group",
		)
		if err != nil {
			logError(fmt.Sprintf("Failed to create bind group: %v", err))
			continue
		}

		// Set bind group
		renderPass.Call("setBindGroup", 0, bindGroup)

		// Set vertex buffer
		renderPass.Call("setVertexBuffer", 0, mesh.VertexBuffer.Buffer)

		// Set index buffer
		renderPass.Call("setIndexBuffer", mesh.IndexBuffer.Buffer, "uint16")

		// Draw indexed
		renderPass.Call("drawIndexed", mesh.IndexCount, 1, 0, 0, 0)
	}

	// End render pass
	renderPass.Call("end")

	// Finish and submit
	commandBuffer := encoder.Call("finish")
	ctx.Submit(commandBuffer)
}

// UpdateMeshTransform updates the transform of a mesh by index
func (sr *SceneRenderer) UpdateMeshTransform(index int, transform Transform) {
	if index >= 0 && index < len(sr.Meshes) {
		sr.Meshes[index].Transform = transform
	}
}

// Cleanup releases GPU resources
func (sr *SceneRenderer) Cleanup() {
	// Destroy mesh buffers
	for _, mesh := range sr.Meshes {
		if mesh.VertexBuffer != nil {
			mesh.VertexBuffer.Destroy()
		}
		if mesh.IndexBuffer != nil {
			mesh.IndexBuffer.Destroy()
		}
	}

	// Destroy uniform buffer
	if sr.UniformBuffer != nil {
		sr.UniformBuffer.Destroy()
	}

	// Destroy depth texture
	if sr.DepthTexture.Truthy() {
		sr.DepthTexture.Call("destroy")
	}
}
