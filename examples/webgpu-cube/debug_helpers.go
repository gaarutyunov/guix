//go:build js && wasm

package main

import "fmt"

func logCommand(cmd ControlCommand) {
	log(fmt.Sprintf("[Command] Type: %s, Value: %f", cmd.Type, cmd.Value))
}

func logStateChange(field string, oldVal, newVal interface{}) {
	log(fmt.Sprintf("[State Change] %s: %v -> %v", field, oldVal, newVal))
}

func logStateSend(state ControlState) {
	log(fmt.Sprintf("[State Send] AutoRotate: %t, Speed: %f", state.AutoRotate, state.Speed))
}

func logStateReceive(state ControlState) {
	log(fmt.Sprintf("[State Receive] AutoRotate: %t, Speed: %f", state.AutoRotate, state.Speed))
}
