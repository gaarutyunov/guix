//go:build js && wasm

package main

import (
	"fmt"

	"github.com/gaarutyunov/guix/pkg/runtime"
)

// AppWithControls extends the generated App with Controls integration
type AppWithControls struct {
	*App
	commands         chan ControlCommand
	controlsInstance *Controls
}

// NewAppWithControls creates an App with integrated Controls
func NewAppWithControls() *AppWithControls {
	fmt.Println("[Go] NewAppWithControls: Creating command channel...")
	commands := make(chan ControlCommand, 10)

	// Start command processor goroutine
	fmt.Println("[Go] NewAppWithControls: Starting command processor...")
	go func() {
		for cmd := range commands {
			processControlCommand(cmd)
		}
	}()

	fmt.Println("[Go] NewAppWithControls: Creating base App...")
	app := NewApp()
	fmt.Println("[Go] NewAppWithControls: Creating Controls...")
	controls := NewControls(WithCommands(commands))
	fmt.Println("[Go] NewAppWithControls: Done")

	return &AppWithControls{
		App:              app,
		commands:         commands,
		controlsInstance: controls,
	}
}

// BindApp binds the runtime app to both App and Controls
func (a *AppWithControls) BindApp(app *runtime.App) {
	a.App.BindApp(app)
	if a.controlsInstance != nil {
		a.controlsInstance.BindApp(app)
	}
}

// Render renders the full app with controls and event handlers
func (a *AppWithControls) Render() *runtime.VNode {
	fmt.Println("[Go] AppWithControls.Render: Called")
	fmt.Println("[Go] AppWithControls.Render: Creating keyboard handler...")
	keyHandler := makeKeyboardHandler(a.commands)
	fmt.Println("[Go] AppWithControls.Render: Creating render callback...")
	renderCallback := makeRenderUpdateCallback()
	fmt.Println("[Go] AppWithControls.Render: Creating cube scene...")
	cubeScene := NewCubeScene(0, 0)
	fmt.Println("[Go] AppWithControls.Render: Rendering controls...")
	controlsVNode := a.controlsInstance.Render()
	fmt.Println("[Go] AppWithControls.Render: Building VNode tree...")

	result := runtime.Div(
		runtime.Class("webgpu-container"),
		runtime.TabIndex(0),
		runtime.OnKeyDown(keyHandler),
		runtime.Canvas(
			runtime.ID("webgpu-canvas"),
			runtime.Width(600),
			runtime.Height(400),
			runtime.GPURenderUpdate(renderCallback),
			runtime.GPUScene(cubeScene),
		),
		controlsVNode,
		runtime.Div(runtime.Class("loading"), runtime.Text("Loading WebGPU...")),
	)

	fmt.Println("[Go] AppWithControls.Render: Done, returning VNode")
	return result
}
