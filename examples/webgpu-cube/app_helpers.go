//go:build js && wasm

package main

// StartCommandProcessor starts the goroutine that processes control commands
func (a *App) StartCommandProcessor() {
	go func() {
		for cmd := range a.commands {
			processControlCommand(cmd)
		}
	}()
}
