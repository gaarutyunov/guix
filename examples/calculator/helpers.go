//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"
	"strconv"
	"strings"
)

// appendToken appends a token to a slice of tokens
func appendToken(tokens []string, token string) []string {
	return append(tokens, token)
}

// clearTokens returns an empty slice
func clearTokens() []string {
	return []string{}
}

// tokensToString converts a slice of tokens to a display string
func tokensToString(tokens []string) string {
	return strings.Join(tokens, " ")
}

// calculateFromTokens evaluates a slice of tokens with proper operator precedence
func calculateFromTokens(tokens []string) float64 {
	// Handle empty case
	if len(tokens) == 0 {
		return 0
	}

	// Handle single number
	if len(tokens) == 1 {
		result, _ := strconv.ParseFloat(tokens[0], 64)
		return result
	}

	// Make a copy of tokens to work with
	workingTokens := make([]string, len(tokens))
	copy(workingTokens, tokens)

	// First pass: handle * and / (higher precedence)
	i := 1
	for i < len(workingTokens) {
		if i+1 < len(workingTokens) {
			operator := workingTokens[i]
			if operator == "*" || operator == "/" {
				// Get left and right operands
				left, _ := strconv.ParseFloat(workingTokens[i-1], 64)
				right, _ := strconv.ParseFloat(workingTokens[i+1], 64)

				// Calculate result
				result := calculate(left, right, operator)

				// Replace the three tokens (num op num) with the result
				resultStr := formatNumber(result)
				workingTokens = append(workingTokens[:i-1], append([]string{resultStr}, workingTokens[i+2:]...)...)

				// Don't increment i, check same position again
				continue
			}
		}
		i = i + 2
	}

	// Second pass: handle + and - (lower precedence)
	i = 1
	for i < len(workingTokens) {
		if i+1 < len(workingTokens) {
			operator := workingTokens[i]
			if operator == "+" || operator == "-" {
				// Get left and right operands
				left, _ := strconv.ParseFloat(workingTokens[i-1], 64)
				right, _ := strconv.ParseFloat(workingTokens[i+1], 64)

				// Calculate result
				result := calculate(left, right, operator)

				// Replace the three tokens with the result
				resultStr := formatNumber(result)
				workingTokens = append(workingTokens[:i-1], append([]string{resultStr}, workingTokens[i+2:]...)...)

				// Don't increment i, check same position again
				continue
			}
		}
		i = i + 2
	}

	// Parse final result
	result, _ := strconv.ParseFloat(workingTokens[0], 64)
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
