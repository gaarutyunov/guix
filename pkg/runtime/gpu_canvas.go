//go:build js && wasm

package runtime

import (
	"fmt"
	"syscall/js"
)

// GPUCanvas represents a canvas element with WebGPU context
type GPUCanvas struct {
	Canvas        js.Value    // HTML canvas element
	Context       js.Value    // GPUCanvasContext
	GPUContext    *GPUContext // Shared GPU context
	Width         int
	Height        int
	Format        string                    // Texture format (e.g., "bgra8unorm")
	FrameCallback js.Func                   // Animation frame callback
	RenderFunc    func(*GPUCanvas, float64) // User render function
	AnimationID   js.Value                  // requestAnimationFrame ID
	Running       bool
	FrameCount    int
	LastTime      float64
}

// GPUCanvasConfig holds configuration for creating a GPU canvas
type GPUCanvasConfig struct {
	Width            int
	Height           int
	DevicePixelRatio float64
	AlphaMode        string // "opaque", "premultiplied"
	FrameLoop        string // "always", "demand", "never"
}

// DefaultGPUCanvasConfig returns default canvas configuration
func DefaultGPUCanvasConfig() GPUCanvasConfig {
	return GPUCanvasConfig{
		Width:            800,
		Height:           600,
		DevicePixelRatio: 1.0,
		AlphaMode:        "premultiplied",
		FrameLoop:        "always",
	}
}

// CreateGPUCanvas creates a new GPU-enabled canvas element
func CreateGPUCanvas(config GPUCanvasConfig) (*GPUCanvas, error) {
	log("[Canvas] Creating GPU canvas")

	// Get or initialize GPU context
	gpuCtx, err := GetOrInitGPUContext()
	if err != nil {
		logError(fmt.Sprintf("[Canvas] Failed to get GPU context: %v", err))
		return nil, fmt.Errorf("failed to initialize WebGPU: %w", err)
	}

	// Create canvas element
	log("[Canvas] Creating canvas element")
	document := js.Global().Get("document")
	canvas := document.Call("createElement", "canvas")
	if !canvas.Truthy() {
		logError("[Canvas] Failed to create canvas element")
		return nil, fmt.Errorf("failed to create canvas element")
	}

	// Set canvas size
	canvas.Set("width", config.Width)
	canvas.Set("height", config.Height)
	canvas.Get("style").Set("width", fmt.Sprintf("%dpx", config.Width))
	canvas.Get("style").Set("height", fmt.Sprintf("%dpx", config.Height))
	log(fmt.Sprintf("[Canvas] Canvas size set to %dx%d", config.Width, config.Height))

	// Get WebGPU context
	log("[Canvas] Getting WebGPU context from canvas")
	gpuCanvasCtx := canvas.Call("getContext", "webgpu")
	if !gpuCanvasCtx.Truthy() {
		logError("[Canvas] Failed to get webgpu context")
		return nil, fmt.Errorf("failed to get webgpu context from canvas")
	}

	// Get preferred format
	format := GetPreferredCanvasFormat()
	log(fmt.Sprintf("[Canvas] Using format: %s", format))

	// Configure the canvas context
	log("[Canvas] Configuring canvas context")
	configObj := map[string]interface{}{
		"device":    gpuCtx.Device,
		"format":    format,
		"alphaMode": config.AlphaMode,
	}
	gpuCanvasCtx.Call("configure", configObj)

	gpuCanvas := &GPUCanvas{
		Canvas:     canvas,
		Context:    gpuCanvasCtx,
		GPUContext: gpuCtx,
		Width:      config.Width,
		Height:     config.Height,
		Format:     format,
		Running:    false,
		FrameCount: 0,
		LastTime:   0,
	}

	log("[Canvas] GPU canvas created successfully")
	return gpuCanvas, nil
}

// SetRenderFunc sets the function to be called on each frame
func (gc *GPUCanvas) SetRenderFunc(renderFunc func(*GPUCanvas, float64)) {
	gc.RenderFunc = renderFunc
}

// Start begins the render loop
func (gc *GPUCanvas) Start() {
	if gc.Running {
		return
	}

	log("[Canvas] Starting render loop")
	gc.Running = true
	gc.startRenderLoop()
}

// Stop halts the render loop
func (gc *GPUCanvas) Stop() {
	gc.Running = false
	if gc.AnimationID.Truthy() {
		js.Global().Call("cancelAnimationFrame", gc.AnimationID)
		gc.AnimationID = js.Undefined()
	}
}

// startRenderLoop initiates the requestAnimationFrame loop
func (gc *GPUCanvas) startRenderLoop() {
	var renderFrame js.Func
	renderFrame = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if !gc.Running {
			renderFrame.Release()
			return nil
		}

		// Get current time from args
		var currentTime float64
		if len(args) > 0 {
			currentTime = args[0].Float()
		} else {
			currentTime = js.Global().Get("performance").Call("now").Float()
		}

		// Calculate delta time
		delta := currentTime - gc.LastTime
		if gc.LastTime == 0 {
			delta = 0
		}
		gc.LastTime = currentTime

		// Call user render function
		if gc.RenderFunc != nil {
			gc.RenderFunc(gc, delta)
		}

		gc.FrameCount++

		// Request next frame
		gc.AnimationID = js.Global().Call("requestAnimationFrame", renderFrame)
		return nil
	})

	gc.FrameCallback = renderFrame
	gc.AnimationID = js.Global().Call("requestAnimationFrame", renderFrame)
}

// GetCurrentTexture returns the current texture to render to
func (gc *GPUCanvas) GetCurrentTexture() js.Value {
	if !gc.Context.Truthy() {
		logError("GPU canvas context not initialized")
		return js.Undefined()
	}
	return gc.Context.Call("getCurrentTexture")
}

// GetCurrentTextureView returns a view of the current texture
func (gc *GPUCanvas) GetCurrentTextureView() js.Value {
	texture := gc.GetCurrentTexture()
	if !texture.Truthy() {
		return js.Undefined()
	}
	return texture.Call("createView")
}

// BeginRenderPass begins a render pass with the current texture as the color attachment
func (gc *GPUCanvas) BeginRenderPass(encoder js.Value, clearColor [4]float32, loadOp string) js.Value {
	if !encoder.Truthy() {
		logError("Command encoder is undefined")
		return js.Undefined()
	}

	textureView := gc.GetCurrentTextureView()
	if !textureView.Truthy() {
		logError("Failed to get current texture view")
		return js.Undefined()
	}

	// If loadOp is empty, default to "clear"
	if loadOp == "" {
		loadOp = "clear"
	}

	// Create color attachment descriptor
	colorAttachment := map[string]interface{}{
		"view":    textureView,
		"loadOp":  loadOp,
		"storeOp": "store",
	}

	// Only add clearValue if using "clear" loadOp
	if loadOp == "clear" {
		colorAttachment["clearValue"] = map[string]interface{}{
			"r": clearColor[0],
			"g": clearColor[1],
			"b": clearColor[2],
			"a": clearColor[3],
		}
	}

	// Create render pass descriptor
	renderPassDescriptor := map[string]interface{}{
		"colorAttachments": []interface{}{colorAttachment},
	}

	return encoder.Call("beginRenderPass", renderPassDescriptor)
}

// BeginRenderPassWithDepth begins a render pass with depth testing
func (gc *GPUCanvas) BeginRenderPassWithDepth(
	encoder js.Value,
	clearColor [4]float32,
	depthTexture js.Value,
	depthLoadOp string,
	depthClearValue float32,
) js.Value {
	if !encoder.Truthy() {
		logError("Command encoder is undefined")
		return js.Undefined()
	}

	textureView := gc.GetCurrentTextureView()
	if !textureView.Truthy() {
		logError("Failed to get current texture view")
		return js.Undefined()
	}

	if depthLoadOp == "" {
		depthLoadOp = "clear"
	}

	// Create depth attachment
	depthAttachment := map[string]interface{}{
		"view":            depthTexture.Call("createView"),
		"depthLoadOp":     depthLoadOp,
		"depthStoreOp":    "store",
		"depthClearValue": depthClearValue,
	}

	// Create color attachment
	colorAttachment := map[string]interface{}{
		"view":    textureView,
		"loadOp":  "clear",
		"storeOp": "store",
		"clearValue": map[string]interface{}{
			"r": clearColor[0],
			"g": clearColor[1],
			"b": clearColor[2],
			"a": clearColor[3],
		},
	}

	// Create render pass descriptor
	renderPassDescriptor := map[string]interface{}{
		"colorAttachments":       []interface{}{colorAttachment},
		"depthStencilAttachment": depthAttachment,
	}

	return encoder.Call("beginRenderPass", renderPassDescriptor)
}

// Resize resizes the canvas and reconfigures the GPU context
func (gc *GPUCanvas) Resize(width, height int) error {
	gc.Width = width
	gc.Height = height

	// Update canvas element size
	gc.Canvas.Set("width", width)
	gc.Canvas.Set("height", height)
	gc.Canvas.Get("style").Set("width", fmt.Sprintf("%dpx", width))
	gc.Canvas.Get("style").Set("height", fmt.Sprintf("%dpx", height))

	// Reconfigure GPU context
	configObj := map[string]interface{}{
		"device":    gc.GPUContext.Device,
		"format":    gc.Format,
		"alphaMode": "premultiplied",
	}
	gc.Context.Call("configure", configObj)

	return nil
}

// Mount attaches the canvas to a DOM element
func (gc *GPUCanvas) Mount(selector string) error {
	log(fmt.Sprintf("[Canvas] Mounting canvas to selector: %s", selector))
	document := js.Global().Get("document")
	container := document.Call("querySelector", selector)
	if !container.Truthy() {
		logError(fmt.Sprintf("[Canvas] Container element not found: %s", selector))
		return fmt.Errorf("container element not found: %s", selector)
	}

	container.Call("appendChild", gc.Canvas)
	log("[Canvas] Canvas mounted successfully")
	return nil
}

// Unmount removes the canvas from the DOM
func (gc *GPUCanvas) Unmount() {
	gc.Stop()
	if gc.FrameCallback.Value.Truthy() {
		gc.FrameCallback.Release()
	}
	if gc.Canvas.Truthy() {
		parent := gc.Canvas.Get("parentNode")
		if parent.Truthy() {
			parent.Call("removeChild", gc.Canvas)
		}
	}
}

// GetAspectRatio returns the aspect ratio of the canvas
func (gc *GPUCanvas) GetAspectRatio() float32 {
	if gc.Height == 0 {
		return 1.0
	}
	return float32(gc.Width) / float32(gc.Height)
}

// CreateDepthTexture creates a depth texture for the canvas
func (gc *GPUCanvas) CreateDepthTexture() (js.Value, error) {
	return gc.GPUContext.CreateTexture(
		gc.Width,
		gc.Height,
		"depth24plus",
		GPUTextureUsageRenderAttachment,
		"depth-texture",
	)
}

// RenderOnce renders a single frame without starting the animation loop
func (gc *GPUCanvas) RenderOnce() {
	if gc.RenderFunc != nil {
		currentTime := js.Global().Get("performance").Call("now").Float()
		delta := currentTime - gc.LastTime
		if gc.LastTime == 0 {
			delta = 0
		}
		gc.LastTime = currentTime
		gc.RenderFunc(gc, delta)
		gc.FrameCount++
	}
}
