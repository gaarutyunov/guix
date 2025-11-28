//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"
	"github.com/gaarutyunov/guix/pkg/runtime"
)

var (
	rotationX float32 = 0
	rotationY float32 = 0
	autoRotate bool = true
	speed float32 = 1.0
)

func main() {
	fmt.Println("WebGPU Rotating Cube Example")

	// Wait for DOM to be ready
	runtime.WaitForDOMReady()

	// Check WebGPU support
	if !runtime.IsWebGPUSupported() {
		runtime.LogError("WebGPU is not supported in this browser")
		showError("WebGPU is not supported in this browser. Please use a browser with WebGPU support (Chrome 113+, Edge 113+)")
		return
	}

	// Initialize WebGPU
	fmt.Println("Initializing WebGPU...")
	gpuCtx, err := runtime.InitWebGPU()
	if err != nil {
		runtime.LogError(fmt.Sprintf("Failed to initialize WebGPU: %v", err))
		showError(fmt.Sprintf("Failed to initialize WebGPU: %v", err))
		return
	}

	fmt.Println("WebGPU initialized successfully")
	runtime.LogInfo("GPU Adapter: " + gpuCtx.Adapter.String())

	// Create GPU canvas
	config := runtime.GPUCanvasConfig{
		Width:            600,
		Height:           400,
		DevicePixelRatio: 1.0,
		AlphaMode:        "premultiplied",
		FrameLoop:        "always",
	}

	canvas, err := runtime.CreateGPUCanvas(config)
	if err != nil {
		runtime.LogError(fmt.Sprintf("Failed to create GPU canvas: %v", err))
		showError(fmt.Sprintf("Failed to create GPU canvas: %v", err))
		return
	}

	fmt.Println("GPU canvas created successfully")

	// Mount canvas to DOM
	if err := canvas.Mount("#app"); err != nil {
		runtime.LogError(fmt.Sprintf("Failed to mount canvas: %v", err))
		showError(fmt.Sprintf("Failed to mount canvas: %v", err))
		return
	}

	fmt.Println("Canvas mounted")

	// Create scene
	scene := createScene()

	// Create renderer
	renderer, err := runtime.NewSceneRenderer(canvas, scene)
	if err != nil {
		runtime.LogError(fmt.Sprintf("Failed to create renderer: %v", err))
		showError(fmt.Sprintf("Failed to create renderer: %v", err))
		return
	}

	fmt.Println("Scene renderer created")

	// Set up controls
	setupControls(renderer)

	// Set render function
	canvas.SetRenderFunc(func(c *runtime.GPUCanvas, delta float64) {
		if autoRotate {
			rotationY += float32(delta) * 0.001 * speed
			rotationX += float32(delta) * 0.0005 * speed
		}

		// Update mesh transform
		if len(renderer.Meshes) > 0 {
			transform := runtime.NewTransform()
			transform.Position = runtime.Vec3{X: 0, Y: 0, Z: 0}
			transform.Rotation = runtime.Vec3{X: rotationX, Y: rotationY, Z: 0}
			transform.Scale = runtime.Vec3{X: 1, Y: 1, Z: 1}
			renderer.UpdateMeshTransform(0, transform)
		}

		// Render
		renderer.Render()
	})

	// Start render loop
	canvas.Start()

	fmt.Println("Render loop started")

	// Keep the program running
	select {}
}

func createScene() *runtime.GPUNode {
	// Create scene
	scene := runtime.Scene(
		runtime.Background(0.1, 0.1, 0.15, 1.0),
	)

	// Create cube geometry
	cubeGeometry := runtime.NewBoxGeometry(2.0, 2.0, 2.0)

	// Create material
	material := runtime.StandardMaterial(
		runtime.Color(0.91, 0.27, 0.38, 1.0), // #e94560
		runtime.Metalness(0.3),
		runtime.Roughness(0.4),
	)

	// Create mesh
	mesh := runtime.Mesh(
		runtime.GeometryProp(cubeGeometry),
		runtime.MaterialProp(material),
		runtime.Position(0, 0, 0),
		runtime.Rotation(0, 0, 0),
		runtime.ScaleValue(1, 1, 1),
	)

	// Create camera
	camera := runtime.PerspectiveCamera(
		runtime.FOV(runtime.DegreesToRadians(60)),
		runtime.Near(0.1),
		runtime.Far(100),
		runtime.Position(0, 2, 6),
		runtime.LookAtPos(0, 0, 0),
	)

	// Create lights
	ambientLight := runtime.AmbientLight(
		runtime.Color(1, 1, 1, 1),
		runtime.Intensity(0.4),
	)

	directionalLight := runtime.DirectionalLight(
		runtime.Position(5, 10, 7),
		runtime.Color(1, 1, 1, 1),
		runtime.Intensity(0.8),
	)

	// Add nodes to scene
	scene.Children = append(scene.Children, mesh, camera, ambientLight, directionalLight)

	return scene
}

func setupControls(renderer *runtime.SceneRenderer) {
	document := js.Global().Get("document")

	// Create control panel HTML
	controlsHTML := `
		<div id="controls" style="margin-top: 16px; text-align: center;">
			<div style="margin-bottom: 12px;">
				<button id="btn-up" style="width: 48px; height: 48px; background: #16213e; color: white; border: none; border-radius: 8px; font-size: 20px; cursor: pointer; margin: 0 4px;">↑</button>
			</div>
			<div style="margin-bottom: 12px;">
				<button id="btn-left" style="width: 48px; height: 48px; background: #16213e; color: white; border: none; border-radius: 8px; font-size: 20px; cursor: pointer; margin: 0 4px;">←</button>
				<button id="btn-toggle" style="width: 48px; height: 48px; background: #16213e; color: white; border: none; border-radius: 8px; font-size: 20px; cursor: pointer; margin: 0 4px;">⏸</button>
				<button id="btn-right" style="width: 48px; height: 48px; background: #16213e; color: white; border: none; border-radius: 8px; font-size: 20px; cursor: pointer; margin: 0 4px;">→</button>
			</div>
			<div style="margin-bottom: 12px;">
				<button id="btn-down" style="width: 48px; height: 48px; background: #16213e; color: white; border: none; border-radius: 8px; font-size: 20px; cursor: pointer; margin: 0 4px;">↓</button>
			</div>
			<div id="speed-control" style="margin-top: 12px; display: flex; align-items: center; justify-content: center; gap: 8px;">
				<span style="color: white;">Speed:</span>
				<input type="range" id="speed-slider" min="0.1" max="3.0" step="0.1" value="1.0" style="width: 200px;">
				<span id="speed-value" style="color: white;">1.0</span>
			</div>
			<p style="color: #888888; font-size: 12px; margin-top: 16px;">
				Use arrow keys or buttons to rotate. Space to toggle auto-rotation.
			</p>
		</div>
	`

	// Insert controls into DOM
	app := document.Call("querySelector", "#app")
	if app.Truthy() {
		app.Set("innerHTML", app.Get("innerHTML").String()+controlsHTML)
	}

	// Button event handlers
	btnUp := document.Call("getElementById", "btn-up")
	if btnUp.Truthy() {
		btnUp.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			rotationX -= 0.2
			return nil
		}))
	}

	btnDown := document.Call("getElementById", "btn-down")
	if btnDown.Truthy() {
		btnDown.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			rotationX += 0.2
			return nil
		}))
	}

	btnLeft := document.Call("getElementById", "btn-left")
	if btnLeft.Truthy() {
		btnLeft.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			rotationY -= 0.2
			return nil
		}))
	}

	btnRight := document.Call("getElementById", "btn-right")
	if btnRight.Truthy() {
		btnRight.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			rotationY += 0.2
			return nil
		}))
	}

	btnToggle := document.Call("getElementById", "btn-toggle")
	if btnToggle.Truthy() {
		btnToggle.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			autoRotate = !autoRotate
			if autoRotate {
				btnToggle.Set("textContent", "⏸")
				// Show speed control
				speedControl := document.Call("getElementById", "speed-control")
				if speedControl.Truthy() {
					speedControl.Get("style").Set("display", "flex")
				}
			} else {
				btnToggle.Set("textContent", "▶")
				// Hide speed control
				speedControl := document.Call("getElementById", "speed-control")
				if speedControl.Truthy() {
					speedControl.Get("style").Set("display", "none")
				}
			}
			return nil
		}))
	}

	// Speed slider
	speedSlider := document.Call("getElementById", "speed-slider")
	if speedSlider.Truthy() {
		speedSlider.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			speed = float32(speedSlider.Get("valueAsNumber").Float())
			speedValue := document.Call("getElementById", "speed-value")
			if speedValue.Truthy() {
				speedValue.Set("textContent", fmt.Sprintf("%.1f", speed))
			}
			return nil
		}))
	}

	// Keyboard event handlers
	document.Call("addEventListener", "keydown", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			event := args[0]
			key := event.Get("key").String()
			switch key {
			case "ArrowUp":
				rotationX -= 0.2
			case "ArrowDown":
				rotationX += 0.2
			case "ArrowLeft":
				rotationY -= 0.2
			case "ArrowRight":
				rotationY += 0.2
			case " ": // Space
				autoRotate = !autoRotate
				if btnToggle.Truthy() {
					if autoRotate {
						btnToggle.Set("textContent", "⏸")
					} else {
						btnToggle.Set("textContent", "▶")
					}
				}
				event.Call("preventDefault")
			}
		}
		return nil
	}))
}

func showError(message string) {
	document := js.Global().Get("document")
	app := document.Call("querySelector", "#app")
	if app.Truthy() {
		errorHTML := fmt.Sprintf(`
			<div style="padding: 20px; background: #ff4444; color: white; border-radius: 8px; margin: 20px;">
				<h2>Error</h2>
				<p>%s</p>
			</div>
		`, message)
		app.Set("innerHTML", errorHTML)
	}
}
