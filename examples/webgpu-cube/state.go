//go:build js && wasm

package main

import "github.com/gaarutyunov/guix/pkg/runtime"

// Rotation state - accessed by render loop and updated by controls
var (
	rotationX    float32 = 0
	rotationY    float32 = 0
	autoRotate   bool    = true
	speed        float32 = 1.0
	renderer     *runtime.SceneRenderer
	loadingChan  chan bool
	firstRender  = true
)

// makeLoadingChannel creates and initializes the loading state channel
func makeLoadingChannel() chan bool {
	if loadingChan == nil {
		loadingChan = make(chan bool, 1)
		loadingChan <- true // Start with loading = true
	}
	return loadingChan
}

// makeCommandChannel creates a new command channel and starts the processor goroutine
func makeCommandChannel() chan ControlCommand {
	commands := make(chan ControlCommand, 10)
	go func() {
		for cmd := range commands {
			processControlCommand(cmd)
		}
	}()
	return commands
}

// makeKeyboardHandler creates a keyboard event handler for the given command channel
func makeKeyboardHandler(commands chan ControlCommand) func(runtime.Event) {
	return func(e runtime.Event) {
		switch e.Key {
		case "ArrowUp":
			commands <- ControlCommand{Type: "rotX", Value: -0.2}
		case "ArrowDown":
			commands <- ControlCommand{Type: "rotX", Value: 0.2}
		case "ArrowLeft":
			commands <- ControlCommand{Type: "rotY", Value: -0.2}
		case "ArrowRight":
			commands <- ControlCommand{Type: "rotY", Value: 0.2}
		case " ":
			commands <- ControlCommand{Type: "autoRotate"}
		}
	}
}

// makeRenderUpdateCallback creates the render update callback for WebGPU
func makeRenderUpdateCallback() func(float64, interface{}) {
	return func(delta float64, r interface{}) {
		// Initialize renderer reference on first call
		if renderer == nil {
			if sceneRenderer, ok := r.(*runtime.SceneRenderer); ok {
				renderer = sceneRenderer
			}
		}

		// Mark loading complete on first frame
		if firstRender && renderer != nil {
			firstRender = false
			if loadingChan != nil {
				select {
				case loadingChan <- false:
				default:
				}
			}
		}

		// Update rotation state and mesh transform
		updateRotation(delta)
	}
}

// updateRotation is called from the render loop to update rotation based on delta time
func updateRotation(delta float64) {
	if autoRotate {
		rotationY += float32(delta) * 0.001 * speed
		rotationX += float32(delta) * 0.0005 * speed
	}

	// Update mesh transform if renderer is available
	if renderer != nil && len(renderer.Meshes) > 0 {
		transform := runtime.NewTransform()
		transform.Position = runtime.Vec3{X: 0, Y: 0, Z: 0}
		transform.Rotation = runtime.Vec3{X: rotationX, Y: rotationY, Z: 0}
		transform.Scale = runtime.Vec3{X: 1, Y: 1, Z: 1}
		renderer.UpdateMeshTransform(0, transform)
	}
}

// processControlCommand handles control commands from the UI
func processControlCommand(cmd ControlCommand) {
	switch cmd.Type {
	case "rotX":
		rotationX += cmd.Value
	case "rotY":
		rotationY += cmd.Value
	case "autoRotate":
		autoRotate = !autoRotate
	case "speed":
		speed = cmd.Value
	}
}
