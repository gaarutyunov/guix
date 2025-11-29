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
	commands := make(chan ControlCommand, 10)

	// Start command processor goroutine
	go func() {
		for cmd := range commands {
			processControlCommand(cmd)
		}
	}()

	return &AppWithControls{
		App:              NewApp(),
		commands:         commands,
		controlsInstance: NewControls(WithCommands(commands)),
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
	return runtime.Div(
		runtime.Class("webgpu-container"),
		runtime.TabIndex(0),
		runtime.OnKeyDown(makeKeyboardHandler(a.commands)),
		runtime.Canvas(
			runtime.ID("webgpu-canvas"),
			runtime.Width(600),
			runtime.Height(400),
			runtime.GPURenderUpdate(makeRenderUpdateCallback()),
			runtime.GPUScene(NewCubeScene(0, 0)),
		),
		a.controlsInstance.Render(),
		runtime.Div(runtime.Class("loading"), runtime.Text("Loading WebGPU...")),
	)
}
