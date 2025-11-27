//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"
	"syscall/js"

	"github.com/gaarutyunov/guix/pkg/runtime"
)

var console = js.Global().Get("console")

func log(args ...interface{}) {
	// Convert all args to strings to avoid js.ValueOf errors
	jsArgs := make([]interface{}, len(args))
	for i, arg := range args {
		jsArgs[i] = fmt.Sprint(arg)
	}
	console.Call("log", jsArgs...)
}

func main() {
	log("MAIN: Starting calculator example")

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
