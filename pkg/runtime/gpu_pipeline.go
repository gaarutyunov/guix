//go:build js && wasm

package runtime

import (
	"fmt"
	"syscall/js"
)

// mapToJSObject converts a Go map to a JavaScript object,
// properly handling js.Value objects and nested structures
func mapToJSObject(m map[string]interface{}) js.Value {
	obj := js.Global().Get("Object").New()
	for key, value := range m {
		switch v := value.(type) {
		case []interface{}:
			// Convert slice to JavaScript array
			arr := js.Global().Get("Array").New(len(v))
			for i, elem := range v {
				if elemMap, ok := elem.(map[string]interface{}); ok {
					arr.SetIndex(i, mapToJSObject(elemMap))
				} else {
					arr.SetIndex(i, elem)
				}
			}
			obj.Set(key, arr)
		case map[string]interface{}:
			// Recursively convert nested maps
			obj.Set(key, mapToJSObject(v))
		default:
			// Direct assignment (handles js.Value, primitives, etc.)
			obj.Set(key, v)
		}
	}
	return obj
}

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
		// Convert bind group layouts to JavaScript array
		jsLayouts := js.Global().Get("Array").New(len(config.BindGroupLayouts))
		for i, layout := range config.BindGroupLayouts {
			jsLayouts.SetIndex(i, layout)
		}

		// Create layout descriptor as JavaScript object
		layoutDescriptor := js.Global().Get("Object").New()
		layoutDescriptor.Set("bindGroupLayouts", jsLayouts)
		if config.Label != "" {
			layoutDescriptor.Set("label", config.Label+" Layout")
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
	// Create target as js.Value to avoid nested map conversion issues
	target := js.Global().Get("Object").New()
	target.Set("format", config.ColorFormat)

	fragmentState := map[string]interface{}{
		"module":     config.FragmentShader,
		"entryPoint": config.FragmentEntryPoint,
		"targets":    []interface{}{target},
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
		// Create depth-stencil as js.Value to avoid nested map conversion issues
		depthStencil := js.Global().Get("Object").New()
		depthStencil.Set("format", config.DepthFormat)
		depthStencil.Set("depthWriteEnabled", true)
		depthStencil.Set("depthCompare", "less")
		pipelineDescriptor["depthStencil"] = depthStencil
	}

	// Convert pipeline descriptor to JavaScript object
	jsPipelineDescriptor := mapToJSObject(pipelineDescriptor)

	// Create the pipeline
	pipeline := ctx.Device.Call("createRenderPipeline", jsPipelineDescriptor)
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
		// Convert bind group layouts to JavaScript array
		jsLayouts := js.Global().Get("Array").New(len(config.BindGroupLayouts))
		for i, layout := range config.BindGroupLayouts {
			jsLayouts.SetIndex(i, layout)
		}

		// Create layout descriptor as JavaScript object
		layoutDescriptor := js.Global().Get("Object").New()
		layoutDescriptor.Set("bindGroupLayouts", jsLayouts)
		if config.Label != "" {
			layoutDescriptor.Set("label", config.Label+" Layout")
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

	// Convert pipeline descriptor to JavaScript object
	jsPipelineDescriptor := mapToJSObject(pipelineDescriptor)

	// Create the pipeline
	pipeline := ctx.Device.Call("createComputePipeline", jsPipelineDescriptor)
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
	// Create blend components as js.Value to avoid nested map conversion issues
	colorBlend := js.Global().Get("Object").New()
	colorBlend.Set("srcFactor", BlendFactorSrcAlpha)
	colorBlend.Set("dstFactor", BlendFactorOneMinusSrcAlpha)
	colorBlend.Set("operation", BlendOperationAdd)

	alphaBlend := js.Global().Get("Object").New()
	alphaBlend.Set("srcFactor", BlendFactorOne)
	alphaBlend.Set("dstFactor", BlendFactorOneMinusSrcAlpha)
	alphaBlend.Set("operation", BlendOperationAdd)

	blend := js.Global().Get("Object").New()
	blend.Set("color", colorBlend)
	blend.Set("alpha", alphaBlend)

	target := js.Global().Get("Object").New()
	target.Set("format", config.ColorFormat)
	target.Set("blend", blend)

	fragmentState := map[string]interface{}{
		"module":     config.FragmentShader,
		"entryPoint": config.FragmentEntryPoint,
		"targets":    []interface{}{target},
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
		// Create depth-stencil as js.Value to avoid nested map conversion issues
		depthStencil := js.Global().Get("Object").New()
		depthStencil.Set("format", config.DepthFormat)
		depthStencil.Set("depthWriteEnabled", true)
		depthStencil.Set("depthCompare", "less")
		pipelineDescriptor["depthStencil"] = depthStencil
	}

	// Convert pipeline descriptor to JavaScript object
	jsPipelineDescriptor := mapToJSObject(pipelineDescriptor)

	// Create the pipeline
	pipeline := ctx.Device.Call("createRenderPipeline", jsPipelineDescriptor)
	if !pipeline.Truthy() {
		return nil, fmt.Errorf("failed to create render pipeline with blending")
	}

	return &RenderPipeline{
		Pipeline: pipeline,
		Label:    config.Label,
	}, nil
}
