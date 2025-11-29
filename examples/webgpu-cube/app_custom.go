//go:build js && wasm

package main

import "github.com/gaarutyunov/guix/pkg/runtime"

// AppWithControls extends the generated App with Controls integration
type AppWithControls struct {
	*App
	commands         chan ControlCommand
	controlsInstance *Controls
}

// NewAppWithControls creates an App with integrated Controls
func NewAppWithControls() *AppWithControls {
	log("[Go] NewAppWithControls: Creating command channel...")
	commands := make(chan ControlCommand, 10)

	// Start command processor goroutine
	log("[Go] NewAppWithControls: Starting command processor...")
	go func() {
		for cmd := range commands {
			processControlCommand(cmd)
		}
	}()

	log("[Go] NewAppWithControls: Creating base App...")
	app := NewApp()
	log("[Go] NewAppWithControls: Creating Controls...")
	controls := NewControls(WithCommands(commands))
	log("[Go] NewAppWithControls: Done")

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
	log("[Go] AppWithControls.Render: Called")
	log("[Go] AppWithControls.Render: Creating keyboard handler...")
	keyHandler := makeKeyboardHandler(a.commands)
	log("[Go] AppWithControls.Render: Creating render callback...")
	renderCallback := makeRenderUpdateCallback()
	log("[Go] AppWithControls.Render: Creating cube scene...")
	cubeScene := NewCubeScene(0, 0)
	log("[Go] AppWithControls.Render: Rendering controls...")
	controlsVNode := a.controlsInstance.Render()
	log("[Go] AppWithControls.Render: Building VNode tree...")

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

	log("[Go] AppWithControls.Render: Done, returning VNode")
	return result
}
