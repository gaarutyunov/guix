//go:build js && wasm

package main

import (
	"github.com/gaarutyunov/guix/pkg/runtime"
	"syscall/js"
)

var console = js.Global().Get("console")

func log(args ...interface{}) {
	console.Call("log", args...)
}

func main() {
	log("[Go] WASM module started")
	log("[Go] WebGPU Rotating Cube Example (Declarative DSL + Channel Reactivity)")
	log("[Go] Waiting for DOM to be ready...")

	// Wait for DOM to be ready before initializing
	runtime.WaitForDOMReady(func() {
		log("[Go] DOM is ready, initializing app")

		// Create app using generated component
		log("[Go] Creating App...")
		appComponent := NewApp()
		log("[Go] App created")

		// Create runtime app
		log("[Go] Creating runtime app...")
		runtimeApp := runtime.NewApp(appComponent)
		log("[Go] Runtime app created")

		// Bind app
		log("[Go] Binding app...")
		appComponent.BindApp(runtimeApp)
		log("[Go] App bound")

		// Mount
		log("[Go] Mounting to #app...")
		if err := runtimeApp.Mount("#app"); err != nil {
			log("[Go] Mount error:", err)
			panic(err)
		}

		log("[Go] App mounted successfully")
	})

	// Keep the program running
	select {}
}
