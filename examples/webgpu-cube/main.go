//go:build js && wasm

package main

import (
	"fmt"
	"github.com/gaarutyunov/guix/pkg/runtime"
)

func main() {
	fmt.Println("[Go] WASM module started")
	fmt.Println("[Go] WebGPU Rotating Cube Example (Declarative DSL)")
	fmt.Println("[Go] Waiting for DOM to be ready...")

	// Wait for DOM to be ready before initializing
	runtime.WaitForDOMReady(func() {
		fmt.Println("[Go] DOM is ready, initializing app")

		// Create and mount app component
		appComponent := NewApp()
		runtimeApp := runtime.NewApp(appComponent)
		appComponent.BindApp(runtimeApp)

		if err := runtimeApp.Mount("#app"); err != nil {
			panic(err)
		}

		fmt.Println("[Go] App mounted successfully")
	})

	// Keep the program running
	select {}
}
