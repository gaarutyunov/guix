//go:build js && wasm

package main

// StartCommandProcessor starts the goroutine that processes control commands
// and sends state updates after each command
func (a *App) StartCommandProcessor() {
	go func() {
		for cmd := range a.commands {
			processControlCommand(cmd)
			// Send updated state after processing command
			sendControlState(a.controlState)
		}
	}()
}
