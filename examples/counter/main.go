//go:build js && wasm
// +build js,wasm

package main

import (
	"github.com/gaarutyunov/guix/pkg/runtime"
)

func main() {
	// Create the app component
	appComponent := NewApp()

	// Create the runtime app
	runtimeApp := runtime.NewApp(appComponent)

	// Bind the component to enable channel reactivity
	appComponent.BindApp(runtimeApp)

	// Mount the app (Mount internally waits for DOM)
	if err := runtimeApp.Mount("#root"); err != nil {
		panic(err)
	}

	// Block forever - WASM must not exit
	select {}
}
