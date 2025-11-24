// +build js,wasm

package main

import (
	"github.com/gaarutyunov/guix/pkg/runtime"
)

func main() {
	// Wait for DOM to be ready
	runtime.WaitForDOMReady(func() {
		// Create and mount the app
		app := NewApp()
		if _, err := runtime.Render("#root", app); err != nil {
			panic(err)
		}
	})

	// Block forever - WASM must not exit
	select {}
}
