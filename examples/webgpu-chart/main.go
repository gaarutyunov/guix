//go:build js && wasm

package main

import (
	"github.com/gaarutyunov/guix/pkg/runtime"
)

func main() {
	log("[Go] WASM module started")
	log("[Go] WebGPU Bitcoin Chart Example")
	log("[Go] Waiting for DOM to be ready...")

	// Wait for DOM to be ready before initializing
	runtime.WaitForDOMReady(func() {
		log("[Go] DOM is ready, initializing app")

		// Fetch initial chart data
		log("[Go] Loading chart data...")
		initialData := GetChartData()
		log("[Go] Loaded", len(initialData), "candles")

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

		// Initialize scroll manager after mount
		log("[Go] Initializing scroll manager...")
		scrollManager := NewScrollManager("chart-canvas", initialData)
		log("[Go] Scroll manager initialized")

		// Store scroll manager for cleanup (not implemented yet)
		_ = scrollManager
	})

	// Keep the program running
	select {}
}
