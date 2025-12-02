//go:build js && wasm

package main

import (
	"github.com/gaarutyunov/guix/pkg/runtime"
)

// updateMeshTransform updates the mesh transform with the given rotation values
func updateMeshTransform(rendererInterface interface{}, rotationX, rotationY float64) {
	if renderer, ok := rendererInterface.(*runtime.SceneRenderer); ok && len(renderer.Meshes) > 0 {
		transform := runtime.NewTransform()
		transform.Position = runtime.Vec3{X: 0, Y: 0, Z: 0}
		transform.Rotation = runtime.Vec3{X: float32(rotationX), Y: float32(rotationY), Z: 0}
		transform.Scale = runtime.Vec3{X: 1, Y: 1, Z: 1}
		renderer.UpdateMeshTransform(0, transform)
	}
}
