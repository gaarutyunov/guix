//go:build js && wasm

package main

// StartCommandProcessor starts the goroutine that processes control commands
// and sends state updates after each command
func (a *App) StartCommandProcessor() {
	go func() {
		for cmd := range a.commands {
			// Process command using component's hoisted variables
			switch cmd.Type {
			case "rotX":
				a.rotationX += float64(cmd.Value)
			case "rotY":
				a.rotationY += float64(cmd.Value)
			case "autoRotate":
				a.autoRotate = !a.autoRotate
			case "speed":
				a.speed = float64(cmd.Value)
			}

			// Send updated state after processing command
			select {
			case a.controlState <- ControlState{AutoRotate: a.autoRotate, Speed: float32(a.speed)}:
			default:
				// Channel full, skip this update
			}
		}
	}()
}
