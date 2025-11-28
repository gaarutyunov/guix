//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"
	"strconv"
	"strings"
)

// calculateFromExpression parses and evaluates an expression string
func calculateFromExpression(expr string) float64 {
	// Split into tokens
	tokens := strings.Fields(expr)

	// Handle empty case
	if len(tokens) == 0 {
		return 0
	}

	// Parse first number
	result, _ := strconv.ParseFloat(tokens[0], 64)

	// Loop through operator-number pairs
	for i := 1; i < len(tokens); i = i + 2 {
		if i+1 < len(tokens) {
			operator := tokens[i]
			num, _ := strconv.ParseFloat(tokens[i+1], 64)
			result = calculate(result, num, operator)
		}
	}

	return result
}

// calculate performs a single operation
func calculate(a float64, b float64, operator string) float64 {
	result := b
	if operator == "+" {
		result = a + b
	} else if operator == "-" {
		result = a - b
	} else if operator == "*" {
		result = a * b
	} else if operator == "/" {
		if b != 0 {
			result = a / b
		} else {
			result = 0
		}
	}
	return result
}

// formatNumber formats a float64 for display
func formatNumber(num float64) string {
	return fmt.Sprintf("%g", num)
}
