//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"
	"strconv"
)

func handleNumber(stateChannel chan CalculatorState, state CalculatorState, digit string) {
	if state.WaitingForOperand {
		state.Display = digit
		state.WaitingForOperand = false
	} else {
		if state.Display == "0" {
			state.Display = digit
		} else {
			state.Display = state.Display + digit
		}
	}

	stateChannel <- state
}

func handleOperator(stateChannel chan CalculatorState, state CalculatorState, operator string) {
	displayValue, _ := strconv.ParseFloat(state.Display, 64)

	if state.Operator != "" && !state.WaitingForOperand {
		// Perform the pending calculation
		result := calculate(state.PreviousValue, displayValue, state.Operator)
		state.Display = formatNumber(result)
		state.PreviousValue = result
	} else {
		state.PreviousValue = displayValue
	}

	state.WaitingForOperand = true
	state.Operator = operator

	stateChannel <- state
}

func handleEquals(stateChannel chan CalculatorState, state CalculatorState) {
	if state.Operator == "" {
		stateChannel <- state
		return
	}

	displayValue, _ := strconv.ParseFloat(state.Display, 64)
	result := calculate(state.PreviousValue, displayValue, state.Operator)

	state.Display = formatNumber(result)
	state.Operator = ""
	state.WaitingForOperand = true
	state.PreviousValue = 0

	stateChannel <- state
}

func handleClear(stateChannel chan CalculatorState) {
	stateChannel <- CalculatorState{
		Display:          "0",
		PreviousValue:    0,
		Operator:         "",
		WaitingForOperand: false,
	}
}

func calculate(a, b float64, operator string) float64 {
	switch operator {
	case "+":
		return a + b
	case "-":
		return a - b
	case "*":
		return a * b
	case "/":
		if b != 0 {
			return a / b
		}
		return 0
	default:
		return b
	}
}

func formatNumber(num float64) string {
	// Format as string, removing trailing zeros
	return fmt.Sprintf("%g", num)
}
