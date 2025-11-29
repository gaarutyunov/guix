//go:build js && wasm

package runtime

import (
	"fmt"
	"syscall/js"
)

// GPUContext holds the WebGPU device, adapter, and queue
type GPUContext struct {
	GPU     js.Value // navigator.gpu
	Adapter js.Value // GPUAdapter
	Device  js.Value // GPUDevice
	Queue   js.Value // GPUQueue
}

// WebGPUFeatures represents optional WebGPU features
type WebGPUFeatures struct {
	DepthClipControl       bool
	Depth32FloatStencil8   bool
	TextureCompressionBC   bool
	TextureCompressionETC2 bool
	TextureCompressionASTC bool
	TimestampQuery         bool
	IndirectFirstInstance  bool
}

// GPULimits represents device limits
type GPULimits struct {
	MaxTextureDimension1D                     uint32
	MaxTextureDimension2D                     uint32
	MaxTextureDimension3D                     uint32
	MaxTextureArrayLayers                     uint32
	MaxBindGroups                             uint32
	MaxDynamicUniformBuffersPerPipelineLayout uint32
	MaxDynamicStorageBuffersPerPipelineLayout uint32
	MaxSampledTexturesPerShaderStage          uint32
	MaxSamplersPerShaderStage                 uint32
	MaxStorageBuffersPerShaderStage           uint32
	MaxStorageTexturesPerShaderStage          uint32
	MaxUniformBuffersPerShaderStage           uint32
	MaxUniformBufferBindingSize               uint32
	MaxStorageBufferBindingSize               uint32
	MinUniformBufferOffsetAlignment           uint32
	MinStorageBufferOffsetAlignment           uint32
	MaxVertexBuffers                          uint32
	MaxVertexAttributes                       uint32
	MaxVertexBufferArrayStride                uint32
	MaxInterStageShaderComponents             uint32
	MaxComputeWorkgroupStorageSize            uint32
	MaxComputeInvocationsPerWorkgroup         uint32
	MaxComputeWorkgroupSizeX                  uint32
	MaxComputeWorkgroupSizeY                  uint32
	MaxComputeWorkgroupSizeZ                  uint32
	MaxComputeWorkgroupsPerDimension          uint32
}

var (
	// GlobalGPUContext is the shared GPU context for the application
	GlobalGPUContext *GPUContext
)

// IsWebGPUSupported checks if WebGPU is available in the browser
func IsWebGPUSupported() bool {
	navigator := js.Global().Get("navigator")
	if !navigator.Truthy() {
		return false
	}
	gpu := navigator.Get("gpu")
	return gpu.Truthy() && !gpu.IsUndefined() && !gpu.IsNull()
}

// InitWebGPU initializes the WebGPU context
// This must be called before any GPU operations
func InitWebGPU() (*GPUContext, error) {
	log("[WebGPU] Starting initialization")

	if !IsWebGPUSupported() {
		logError("[WebGPU] WebGPU is not supported in this browser")
		return nil, fmt.Errorf("WebGPU is not supported in this browser")
	}
	log("[WebGPU] WebGPU support detected")

	navigator := js.Global().Get("navigator")
	gpu := navigator.Get("gpu")

	ctx := &GPUContext{
		GPU: gpu,
	}

	// Request adapter (async operation)
	log("[WebGPU] Requesting GPU adapter...")
	adapterPromise := gpu.Call("requestAdapter")

	// Wait for adapter promise to resolve
	adapter, err := awaitPromise(adapterPromise)
	if err != nil {
		logError(fmt.Sprintf("[WebGPU] Failed to request GPU adapter: %v", err))
		return nil, fmt.Errorf("failed to request GPU adapter: %w", err)
	}
	if !adapter.Truthy() {
		logError("[WebGPU] GPU adapter is null")
		return nil, fmt.Errorf("GPU adapter is null - WebGPU may not be available")
	}
	log("[WebGPU] GPU adapter acquired")
	ctx.Adapter = adapter

	// Request device (async operation)
	log("[WebGPU] Requesting GPU device...")
	deviceDescriptor := map[string]interface{}{
		"label": "Guix WebGPU Device",
	}
	devicePromise := adapter.Call("requestDevice", deviceDescriptor)

	device, err := awaitPromise(devicePromise)
	if err != nil {
		logError(fmt.Sprintf("[WebGPU] Failed to request GPU device: %v", err))
		return nil, fmt.Errorf("failed to request GPU device: %w", err)
	}
	if !device.Truthy() {
		logError("[WebGPU] GPU device is null")
		return nil, fmt.Errorf("GPU device is null")
	}
	log("[WebGPU] GPU device acquired")
	ctx.Device = device

	// Get the queue
	log("[WebGPU] Getting GPU queue...")
	ctx.Queue = device.Get("queue")
	if !ctx.Queue.Truthy() {
		logError("[WebGPU] GPU queue is null")
		return nil, fmt.Errorf("GPU queue is null")
	}
	log("[WebGPU] GPU queue acquired")

	// Set up device error handling
	device.Call("addEventListener", "uncapturederror", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			event := args[0]
			error := event.Get("error")
			message := error.Get("message").String()
			logError(fmt.Sprintf("WebGPU uncaptured error: %s", message))
		}
		return nil
	}))

	// Store global context
	GlobalGPUContext = ctx
	log("[WebGPU] Initialization complete")

	return ctx, nil
}

// GetOrInitGPUContext returns the global GPU context or initializes it if needed
func GetOrInitGPUContext() (*GPUContext, error) {
	if GlobalGPUContext != nil {
		return GlobalGPUContext, nil
	}
	return InitWebGPU()
}

// awaitPromise waits for a JavaScript Promise to resolve and returns the result
func awaitPromise(promise js.Value) (js.Value, error) {
	if !promise.Truthy() || promise.Type() != js.TypeObject {
		return js.Undefined(), fmt.Errorf("invalid promise")
	}

	resultChan := make(chan js.Value, 1)
	errorChan := make(chan error, 1)

	// Create resolve callback
	var resolveFn js.Func
	resolveFn = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			resultChan <- args[0]
		} else {
			resultChan <- js.Undefined()
		}
		resolveFn.Release()
		return nil
	})

	// Create reject callback
	var rejectFn js.Func
	rejectFn = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		var err error
		if len(args) > 0 {
			errMsg := args[0].String()
			err = fmt.Errorf("%s", errMsg)
		} else {
			err = fmt.Errorf("promise rejected with no reason")
		}
		errorChan <- err
		rejectFn.Release()
		return nil
	})

	// Attach callbacks to promise
	promise.Call("then", resolveFn).Call("catch", rejectFn)

	// Wait for result or error
	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errorChan:
		return js.Undefined(), err
	}
}

// GetPreferredCanvasFormat returns the preferred canvas format for the current system
func GetPreferredCanvasFormat() string {
	if !IsWebGPUSupported() {
		return "rgba8unorm" // fallback
	}
	navigator := js.Global().Get("navigator")
	gpu := navigator.Get("gpu")
	format := gpu.Call("getPreferredCanvasFormat")
	if format.Truthy() {
		return format.String()
	}
	return "bgra8unorm" // common fallback
}

// GPUBufferUsage constants matching WebGPU specification
const (
	GPUBufferUsageMapRead      = 0x0001
	GPUBufferUsageMapWrite     = 0x0002
	GPUBufferUsageCopySrc      = 0x0004
	GPUBufferUsageCopyDst      = 0x0008
	GPUBufferUsageIndex        = 0x0010
	GPUBufferUsageVertex       = 0x0020
	GPUBufferUsageUniform      = 0x0040
	GPUBufferUsageStorage      = 0x0080
	GPUBufferUsageIndirect     = 0x0100
	GPUBufferUsageQueryResolve = 0x0200
)

// GPUTextureUsage constants matching WebGPU specification
const (
	GPUTextureUsageCopySrc          = 0x01
	GPUTextureUsageCopyDst          = 0x02
	GPUTextureUsageTextureBinding   = 0x04
	GPUTextureUsageStorageBinding   = 0x08
	GPUTextureUsageRenderAttachment = 0x10
)

// GPUShaderStage constants
const (
	GPUShaderStageVertex   = 0x1
	GPUShaderStageFragment = 0x2
	GPUShaderStageCompute  = 0x4
)

// GPUMapMode constants
const (
	GPUMapModeRead  = 0x0001
	GPUMapModeWrite = 0x0002
)

// CreateBuffer creates a GPU buffer with the specified size and usage
func (ctx *GPUContext) CreateBuffer(size int, usage int, label string) (js.Value, error) {
	if ctx.Device.IsUndefined() {
		return js.Undefined(), fmt.Errorf("GPU device not initialized")
	}

	descriptor := map[string]interface{}{
		"size":  size,
		"usage": usage,
	}
	if label != "" {
		descriptor["label"] = label
	}

	buffer := ctx.Device.Call("createBuffer", descriptor)
	if !buffer.Truthy() {
		return js.Undefined(), fmt.Errorf("failed to create buffer")
	}

	return buffer, nil
}

// CreateTexture creates a GPU texture with the specified parameters
func (ctx *GPUContext) CreateTexture(width, height int, format string, usage int, label string) (js.Value, error) {
	if ctx.Device.IsUndefined() {
		return js.Undefined(), fmt.Errorf("GPU device not initialized")
	}

	descriptor := map[string]interface{}{
		"size": map[string]interface{}{
			"width":              width,
			"height":             height,
			"depthOrArrayLayers": 1,
		},
		"format": format,
		"usage":  usage,
	}
	if label != "" {
		descriptor["label"] = label
	}

	texture := ctx.Device.Call("createTexture", descriptor)
	if !texture.Truthy() {
		return js.Undefined(), fmt.Errorf("failed to create texture")
	}

	return texture, nil
}

// CreateShaderModule creates a shader module from WGSL source code
func (ctx *GPUContext) CreateShaderModule(code string, label string) (js.Value, error) {
	if ctx.Device.IsUndefined() {
		return js.Undefined(), fmt.Errorf("GPU device not initialized")
	}

	descriptor := map[string]interface{}{
		"code": code,
	}
	if label != "" {
		descriptor["label"] = label
	}

	shaderModule := ctx.Device.Call("createShaderModule", descriptor)
	if !shaderModule.Truthy() {
		return js.Undefined(), fmt.Errorf("failed to create shader module")
	}

	return shaderModule, nil
}

// WriteBuffer writes data to a GPU buffer at the specified offset
func (ctx *GPUContext) WriteBuffer(buffer js.Value, offset int, data []byte) error {
	if ctx.Queue.IsUndefined() {
		return fmt.Errorf("GPU queue not initialized")
	}

	// Create a Uint8Array from the byte data
	jsArray := js.Global().Get("Uint8Array").New(len(data))
	js.CopyBytesToJS(jsArray, data)

	// Write to buffer
	ctx.Queue.Call("writeBuffer", buffer, offset, jsArray, 0, len(data))

	return nil
}

// Submit submits command buffers to the GPU queue
func (ctx *GPUContext) Submit(commandBuffers ...js.Value) {
	if ctx.Queue.IsUndefined() {
		logError("Cannot submit commands: GPU queue not initialized")
		return
	}

	// Create JavaScript array of command buffers
	jsArray := js.Global().Get("Array").New()
	for _, cb := range commandBuffers {
		if cb.Truthy() {
			jsArray.Call("push", cb)
		}
	}

	ctx.Queue.Call("submit", jsArray)
}

// CreateCommandEncoder creates a new command encoder
func (ctx *GPUContext) CreateCommandEncoder(label string) (js.Value, error) {
	if ctx.Device.IsUndefined() {
		return js.Undefined(), fmt.Errorf("GPU device not initialized")
	}

	descriptor := map[string]interface{}{}
	if label != "" {
		descriptor["label"] = label
	}

	encoder := ctx.Device.Call("createCommandEncoder", descriptor)
	if !encoder.Truthy() {
		return js.Undefined(), fmt.Errorf("failed to create command encoder")
	}

	return encoder, nil
}

// Destroy releases GPU resources (call on cleanup)
func (ctx *GPUContext) Destroy() {
	if ctx.Device.Truthy() {
		ctx.Device.Call("destroy")
	}
	GlobalGPUContext = nil
}
