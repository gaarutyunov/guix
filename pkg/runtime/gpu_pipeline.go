//go:build js && wasm

package runtime

import (
	"fmt"
	"syscall/js"
)

// RenderPipeline wraps a WebGPU render pipeline
type RenderPipeline struct {
	Pipeline js.Value
	Layout   js.Value
	Label    string
}

// PipelineConfig holds configuration for creating a render pipeline
type PipelineConfig struct {
	Label              string
	VertexShader       js.Value
	FragmentShader     js.Value
	VertexEntryPoint   string
	FragmentEntryPoint string
	VertexBuffers      []map[string]interface{}
	ColorFormat        string
	DepthFormat        string
	PrimitiveTopology  string
	CullMode           string
	BindGroupLayouts   []js.Value
}

// DefaultPipelineConfig returns a default pipeline configuration
func DefaultPipelineConfig() PipelineConfig {
	return PipelineConfig{
		VertexEntryPoint:   "vs_main",
		FragmentEntryPoint: "fs_main",
		ColorFormat:        GetPreferredCanvasFormat(),
		PrimitiveTopology:  "triangle-list",
		CullMode:           "back",
	}
}

// CreateRenderPipeline creates a render pipeline with the specified configuration
func CreateRenderPipeline(ctx *GPUContext, config PipelineConfig) (*RenderPipeline, error) {
	if ctx.Device.IsUndefined() {
		return nil, fmt.Errorf("GPU device not initialized")
	}

	if config.VertexShader.IsUndefined() {
		return nil, fmt.Errorf("vertex shader is required")
	}
	if config.FragmentShader.IsUndefined() {
		return nil, fmt.Errorf("fragment shader is required")
	}

	// Create pipeline layout
	var pipelineLayout js.Value
	if len(config.BindGroupLayouts) > 0 {
		layoutDescriptor := map[string]interface{}{
			"bindGroupLayouts": config.BindGroupLayouts,
		}
		if config.Label != "" {
			layoutDescriptor["label"] = config.Label + " Layout"
		}
		pipelineLayout = ctx.Device.Call("createPipelineLayout", layoutDescriptor)
	} else {
		pipelineLayout = js.Undefined()
	}

	// Create vertex state
	vertexState := map[string]interface{}{
		"module":     config.VertexShader,
		"entryPoint": config.VertexEntryPoint,
	}
	if len(config.VertexBuffers) > 0 {
		vertexState["buffers"] = config.VertexBuffers
	}

	// Create fragment state
	fragmentState := map[string]interface{}{
		"module":     config.FragmentShader,
		"entryPoint": config.FragmentEntryPoint,
		"targets": []interface{}{
			map[string]interface{}{
				"format": config.ColorFormat,
			},
		},
	}

	// Create primitive state
	primitiveState := map[string]interface{}{
		"topology": config.PrimitiveTopology,
	}
	if config.CullMode != "" {
		primitiveState["cullMode"] = config.CullMode
	}

	// Create pipeline descriptor
	pipelineDescriptor := map[string]interface{}{
		"vertex":    vertexState,
		"fragment":  fragmentState,
		"primitive": primitiveState,
	}

	if config.Label != "" {
		pipelineDescriptor["label"] = config.Label
	}

	if pipelineLayout.Truthy() {
		pipelineDescriptor["layout"] = pipelineLayout
	} else {
		pipelineDescriptor["layout"] = "auto"
	}

	// Add depth-stencil state if depth format is specified
	if config.DepthFormat != "" {
		pipelineDescriptor["depthStencil"] = map[string]interface{}{
			"format":            config.DepthFormat,
			"depthWriteEnabled": true,
			"depthCompare":      "less",
		}
	}

	// Create the pipeline
	pipeline := ctx.Device.Call("createRenderPipeline", pipelineDescriptor)
	if !pipeline.Truthy() {
		return nil, fmt.Errorf("failed to create render pipeline")
	}

	return &RenderPipeline{
		Pipeline: pipeline,
		Layout:   pipelineLayout,
		Label:    config.Label,
	}, nil
}

// ComputePipeline wraps a WebGPU compute pipeline
type ComputePipeline struct {
	Pipeline js.Value
	Layout   js.Value
	Label    string
}

// ComputePipelineConfig holds configuration for creating a compute pipeline
type ComputePipelineConfig struct {
	Label            string
	ComputeShader    js.Value
	EntryPoint       string
	BindGroupLayouts []js.Value
}

// CreateComputePipeline creates a compute pipeline
func CreateComputePipeline(ctx *GPUContext, config ComputePipelineConfig) (*ComputePipeline, error) {
	if ctx.Device.IsUndefined() {
		return nil, fmt.Errorf("GPU device not initialized")
	}

	if config.ComputeShader.IsUndefined() {
		return nil, fmt.Errorf("compute shader is required")
	}

	if config.EntryPoint == "" {
		config.EntryPoint = "cs_main"
	}

	// Create pipeline layout
	var pipelineLayout js.Value
	if len(config.BindGroupLayouts) > 0 {
		layoutDescriptor := map[string]interface{}{
			"bindGroupLayouts": config.BindGroupLayouts,
		}
		if config.Label != "" {
			layoutDescriptor["label"] = config.Label + " Layout"
		}
		pipelineLayout = ctx.Device.Call("createPipelineLayout", layoutDescriptor)
	} else {
		pipelineLayout = js.Undefined()
	}

	// Create compute state
	computeState := map[string]interface{}{
		"module":     config.ComputeShader,
		"entryPoint": config.EntryPoint,
	}

	// Create pipeline descriptor
	pipelineDescriptor := map[string]interface{}{
		"compute": computeState,
	}

	if config.Label != "" {
		pipelineDescriptor["label"] = config.Label
	}

	if pipelineLayout.Truthy() {
		pipelineDescriptor["layout"] = pipelineLayout
	} else {
		pipelineDescriptor["layout"] = "auto"
	}

	// Create the pipeline
	pipeline := ctx.Device.Call("createComputePipeline", pipelineDescriptor)
	if !pipeline.Truthy() {
		return nil, fmt.Errorf("failed to create compute pipeline")
	}

	return &ComputePipeline{
		Pipeline: pipeline,
		Layout:   pipelineLayout,
		Label:    config.Label,
	}, nil
}

// PrimitiveTopology constants
const (
	PrimitiveTopologyPointList     = "point-list"
	PrimitiveTopologyLineList      = "line-list"
	PrimitiveTopologyLineStrip     = "line-strip"
	PrimitiveTopologyTriangleList  = "triangle-list"
	PrimitiveTopologyTriangleStrip = "triangle-strip"
)

// CullMode constants
const (
	CullModeNone  = "none"
	CullModeFront = "front"
	CullModeBack  = "back"
)

// CompareFunction constants
const (
	CompareFunctionNever        = "never"
	CompareFunctionLess         = "less"
	CompareFunctionEqual        = "equal"
	CompareFunctionLessEqual    = "less-equal"
	CompareFunctionGreater      = "greater"
	CompareFunctionNotEqual     = "not-equal"
	CompareFunctionGreaterEqual = "greater-equal"
	CompareFunctionAlways       = "always"
)

// BlendFactor constants
const (
	BlendFactorZero             = "zero"
	BlendFactorOne              = "one"
	BlendFactorSrc              = "src"
	BlendFactorOneMinusSrc      = "one-minus-src"
	BlendFactorSrcAlpha         = "src-alpha"
	BlendFactorOneMinusSrcAlpha = "one-minus-src-alpha"
	BlendFactorDst              = "dst"
	BlendFactorOneMinusDst      = "one-minus-dst"
	BlendFactorDstAlpha         = "dst-alpha"
	BlendFactorOneMinusDstAlpha = "one-minus-dst-alpha"
)

// BlendOperation constants
const (
	BlendOperationAdd             = "add"
	BlendOperationSubtract        = "subtract"
	BlendOperationReverseSubtract = "reverse-subtract"
	BlendOperationMin             = "min"
	BlendOperationMax             = "max"
)

// CreatePipelineWithBlending creates a render pipeline with alpha blending
func CreatePipelineWithBlending(ctx *GPUContext, config PipelineConfig) (*RenderPipeline, error) {
	if ctx.Device.IsUndefined() {
		return nil, fmt.Errorf("GPU device not initialized")
	}

	// Create vertex state
	vertexState := map[string]interface{}{
		"module":     config.VertexShader,
		"entryPoint": config.VertexEntryPoint,
	}
	if len(config.VertexBuffers) > 0 {
		vertexState["buffers"] = config.VertexBuffers
	}

	// Create fragment state with blending
	fragmentState := map[string]interface{}{
		"module":     config.FragmentShader,
		"entryPoint": config.FragmentEntryPoint,
		"targets": []interface{}{
			map[string]interface{}{
				"format": config.ColorFormat,
				"blend": map[string]interface{}{
					"color": map[string]interface{}{
						"srcFactor": BlendFactorSrcAlpha,
						"dstFactor": BlendFactorOneMinusSrcAlpha,
						"operation": BlendOperationAdd,
					},
					"alpha": map[string]interface{}{
						"srcFactor": BlendFactorOne,
						"dstFactor": BlendFactorOneMinusSrcAlpha,
						"operation": BlendOperationAdd,
					},
				},
			},
		},
	}

	// Create primitive state
	primitiveState := map[string]interface{}{
		"topology": config.PrimitiveTopology,
	}
	if config.CullMode != "" {
		primitiveState["cullMode"] = config.CullMode
	}

	// Create pipeline descriptor
	pipelineDescriptor := map[string]interface{}{
		"vertex":    vertexState,
		"fragment":  fragmentState,
		"primitive": primitiveState,
		"layout":    "auto",
	}

	if config.Label != "" {
		pipelineDescriptor["label"] = config.Label
	}

	// Add depth-stencil state if depth format is specified
	if config.DepthFormat != "" {
		pipelineDescriptor["depthStencil"] = map[string]interface{}{
			"format":            config.DepthFormat,
			"depthWriteEnabled": true,
			"depthCompare":      "less",
		}
	}

	// Create the pipeline
	pipeline := ctx.Device.Call("createRenderPipeline", pipelineDescriptor)
	if !pipeline.Truthy() {
		return nil, fmt.Errorf("failed to create render pipeline with blending")
	}

	return &RenderPipeline{
		Pipeline: pipeline,
		Label:    config.Label,
	}, nil
}
