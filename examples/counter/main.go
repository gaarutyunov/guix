//go:build js && wasm
// +build js,wasm

package main

import (
	"github.com/gaarutyunov/guix/pkg/runtime"
)

func main() {
	log("MAIN: Starting counter example")

	// Create the app component
	log("MAIN: Creating App component")
	appComponent := NewApp()
	log("MAIN: App component created")

	// Create the runtime app
	log("MAIN: Creating runtime app")
	runtimeApp := runtime.NewApp(appComponent)
	log("MAIN: Runtime app created")

	// Bind the component to enable channel reactivity
	log("MAIN: Binding app component")
	appComponent.BindApp(runtimeApp)
	log("MAIN: App component bound")

	// Mount the app (Mount internally waits for DOM)
	log("MAIN: Mounting app to #root")
	if err := runtimeApp.Mount("#root"); err != nil {
		console.Call("error", "MAIN: Mount failed:", err.Error())
		panic(err)
	}
	log("MAIN: App mounted successfully")

	// Block forever - WASM must not exit
	log("MAIN: Entering event loop")
	select {}
}
