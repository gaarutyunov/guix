//go:build js && wasm

package main

import "github.com/gaarutyunov/guix/pkg/runtime"

// Rotation state - accessed by render loop and updated by controls
var (
	rotationX      float32 = 0
	rotationY      float32 = 0
	autoRotate     bool    = true
	speed          float32 = 1.0
	renderer       *runtime.SceneRenderer
	renderCallback func(float64, interface{})
)

// getRenderUpdateCallback returns the render update callback
func getRenderUpdateCallback() func(float64, interface{}) {
	return renderCallback
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
