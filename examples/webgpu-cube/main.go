//go:build js && wasm

package main

import (
	"fmt"
	"github.com/gaarutyunov/guix/pkg/runtime"
	"syscall/js"
)

func main() {
	fmt.Println("[Go] WASM module started")
	fmt.Println("[Go] WebGPU Rotating Cube Example (Declarative DSL + Channel Reactivity)")
	fmt.Println("[Go] Waiting for DOM to be ready...")

	// Wait for DOM to be ready before initializing
	runtime.WaitForDOMReady(func() {
		fmt.Println("[Go] DOM is ready, initializing app")

		// Create command channel for controls
		commands := make(chan ControlCommand, 10)

		// Set up render update callback
		renderCallback = createRenderUpdateCallback()

		// Create and mount app component
		appComponent := NewApp()
		runtimeApp := runtime.NewApp(appComponent)
		appComponent.BindApp(runtimeApp)

		if err := runtimeApp.Mount("#app"); err != nil {
			panic(err)
		}

		fmt.Println("[Go] App mounted successfully")

		// Mount controls component
		mountControls(commands)

		// Set up keyboard event handler on document
		setupKeyboardHandler(commands)

		// Set up command processor
		setupRenderCallback(commands)
	})

	// Keep the program running
	select {}
}

// mountControls mounts the Controls component to the controls container
func mountControls(commands chan ControlCommand) {
	document := js.Global().Get("document")
	container := document.Call("querySelector", "#controls-container")
	if !container.Truthy() {
		fmt.Println("[Go] Warning: controls container not found")
		return
	}

	// Create and mount controls component
	controlsComponent := NewControls(WithCommands(commands))
	controlsVNode := controlsComponent.Render()
	runtime.Mount(controlsVNode, container)

	fmt.Println("[Go] Controls mounted")
}

// setupKeyboardHandler attaches keyboard event listener to document
func setupKeyboardHandler(commands chan ControlCommand) {
	document := js.Global().Get("document")
	document.Call("addEventListener", "keydown", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			event := args[0]
			key := event.Get("key").String()
			switch key {
			case "ArrowUp":
				commands <- ControlCommand{Type: "rotX", Value: -0.2}
			case "ArrowDown":
				commands <- ControlCommand{Type: "rotX", Value: 0.2}
			case "ArrowLeft":
				commands <- ControlCommand{Type: "rotY", Value: -0.2}
			case "ArrowRight":
				commands <- ControlCommand{Type: "rotY", Value: 0.2}
			case " ": // Space
				commands <- ControlCommand{Type: "autoRotate"}
				event.Call("preventDefault")
			}
		}
		return nil
	}))
	fmt.Println("[Go] Keyboard handler set up")
}

// setupRenderCallback starts the command processor goroutine
func setupRenderCallback(commands chan ControlCommand) {
	go func() {
		for cmd := range commands {
			processControlCommand(cmd)
		}
	}()

	fmt.Println("[Go] Command processor started")
}

// createRenderUpdateCallback creates the render update callback for the GPU canvas
func createRenderUpdateCallback() func(float64, interface{}) {
	return func(delta float64, r interface{}) {
		// Initialize renderer reference on first call
		if renderer == nil {
			if sceneRenderer, ok := r.(*runtime.SceneRenderer); ok {
				renderer = sceneRenderer
				fmt.Println("[Go] Renderer initialized")
			}
		}

		// Update rotation state and mesh transform
		updateRotation(delta)
	}
}
