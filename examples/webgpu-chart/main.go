//go:build js && wasm

package main

import (
	"github.com/gaarutyunov/guix/pkg/runtime"
)

func main() {
	// Initialize the app
	app := NewApp()

	// Mount to DOM
	if err := runtime.Mount(app, "#app"); err != nil {
		runtime.LogError("Failed to mount app: " + err.Error())
		return
	}

	// Keep the app running
	select {}
}
