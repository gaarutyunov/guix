//go:build js && wasm

package main

import (
	"fmt"
	"github.com/gaarutyunov/guix/pkg/runtime"
)

func main() {
	fmt.Println("[Go] WASM module started")
	fmt.Println("[Go] WebGPU Rotating Cube Example (Declarative DSL + Channel Reactivity)")
	fmt.Println("[Go] Waiting for DOM to be ready...")

	// Wait for DOM to be ready before initializing
	runtime.WaitForDOMReady(func() {
		fmt.Println("[Go] DOM is ready, initializing app")

		// Create and mount app component with controls
		fmt.Println("[Go] Creating AppWithControls...")
		appComponent := NewAppWithControls()
		fmt.Println("[Go] AppWithControls created")

		fmt.Println("[Go] Creating runtime app...")
		runtimeApp := runtime.NewApp(appComponent)
		fmt.Println("[Go] Runtime app created")

		fmt.Println("[Go] Binding app...")
		appComponent.BindApp(runtimeApp)
		fmt.Println("[Go] App bound")

		fmt.Println("[Go] Mounting to #app...")
		if err := runtimeApp.Mount("#app"); err != nil {
			fmt.Println("[Go] Mount error:", err)
			panic(err)
		}

		fmt.Println("[Go] App mounted successfully with integrated controls")
	})

	// Keep the program running
	select {}
}
