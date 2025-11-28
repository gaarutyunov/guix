//go:build js && wasm
// +build js,wasm

package main

import (
	"testing"
)

func TestCalculateFromTokens_OperatorPrecedence(t *testing.T) {
	tests := []struct {
		name     string
		tokens   []string
		expected float64
	}{
		{
			name:     "multiplication before addition",
			tokens:   []string{"1", "+", "2", "*", "3"},
			expected: 7, // 1 + (2 * 3) = 1 + 6 = 7
		},
		{
			name:     "division before subtraction",
			tokens:   []string{"10", "-", "8", "/", "2"},
			expected: 6, // 10 - (8 / 2) = 10 - 4 = 6
		},
		{
			name:     "multiple multiplications and additions",
			tokens:   []string{"2", "+", "3", "*", "4", "+", "5"},
			expected: 19, // 2 + (3 * 4) + 5 = 2 + 12 + 5 = 19
		},
		{
			name:     "left to right for same precedence",
			tokens:   []string{"10", "+", "5", "-", "3"},
			expected: 12, // (10 + 5) - 3 = 15 - 3 = 12
		},
		{
			name:     "simple addition",
			tokens:   []string{"2", "+", "3"},
			expected: 5,
		},
		{
			name:     "simple multiplication",
			tokens:   []string{"6", "*", "7"},
			expected: 42,
		},
		{
			name:     "complex expression",
			tokens:   []string{"5", "*", "2", "+", "3", "*", "4", "-", "1"},
			expected: 21, // (5 * 2) + (3 * 4) - 1 = 10 + 12 - 1 = 21
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateFromTokens(tt.tokens)
			if result != tt.expected {
				t.Errorf("calculateFromTokens(%v) = %v, want %v", tt.tokens, result, tt.expected)
			}
		})
	}
}
