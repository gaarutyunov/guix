// Package main provides a calculator example using Guix.
//
// To generate the component code, run:
//
//	go generate
//
// Or use the guix CLI directly:
//
//	guix generate -p .
//
//go:generate guix generate -p .
package main

// NewCalculatorStateChannel creates and initializes a calculator state channel
func NewCalculatorStateChannel() chan CalculatorState {
	ch := make(chan CalculatorState, 10)
	ch <- CalculatorState{
		Display:          "0",
		PreviousValue:    0,
		Operator:         "",
		WaitingForOperand: false,
	}
	return ch
}
