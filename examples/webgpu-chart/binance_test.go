//go:build js && wasm

package main

import (
	"math"
	"testing"
)

func TestParseFloat(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected float64
	}{
		{"float64", float64(123.456), 123.456},
		{"string valid", "789.012", 789.012},
		{"string integer", "100", 100.0},
		{"int", int(42), 42.0},
		{"int64", int64(999), 999.0},
		{"zero float", float64(0), 0.0},
		{"zero string", "0", 0.0},
		{"negative float", float64(-50.5), -50.5},
		{"negative string", "-123.45", -123.45},
		{"invalid string", "invalid", 0.0},
		{"empty string", "", 0.0},
		{"nil", nil, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseFloat(tt.input)
			if math.Abs(result-tt.expected) > 0.0001 {
				t.Errorf("Expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestParseFloatPrecision(t *testing.T) {
	// Test that string parsing maintains precision
	tests := []struct {
		input    string
		expected float64
	}{
		{"12345.6789", 12345.6789},
		{"0.00000001", 0.00000001},
		{"999999999.999999", 999999999.999999},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseFloat(tt.input)
			// Allow small floating point error
			if math.Abs(result-tt.expected) > 0.000001 {
				t.Errorf("Expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestParseFloatTypes(t *testing.T) {
	// Test all supported types
	f64 := 123.456
	i := 42
	var i64 int64 = 999

	if parseFloat(f64) != 123.456 {
		t.Error("Failed to parse float64")
	}
	if parseFloat(i) != 42.0 {
		t.Error("Failed to parse int")
	}
	if parseFloat(i64) != 999.0 {
		t.Error("Failed to parse int64")
	}
}

func TestParseFloatStringFormats(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{"scientific notation", "1.23e2", 123.0},
		{"scientific notation negative", "1.5e-3", 0.0015},
		{"leading zeros", "00123.45", 123.45},
		{"trailing zeros", "123.450000", 123.45},
		{"no decimal", "12345", 12345.0},
		{"only decimal", ".5", 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseFloat(tt.input)
			if math.Abs(result-tt.expected) > 0.0001 {
				t.Errorf("Expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestParseFloatEdgeCases(t *testing.T) {
	// Test edge cases that should return 0
	edgeCases := []interface{}{
		nil,
		"not-a-number",
		"NaN",
		"infinity",
		struct{}{},
		[]int{1, 2, 3},
		map[string]int{"a": 1},
	}

	for _, input := range edgeCases {
		result := parseFloat(input)
		if result != 0.0 {
			t.Errorf("Expected 0.0 for edge case, got %f", result)
		}
	}
}

func TestParseFloatNegativeNumbers(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected float64
	}{
		{float64(-123.456), -123.456},
		{"-789.012", -789.012},
		{int(-42), -42.0},
		{int64(-999), -999.0},
	}

	for _, tt := range tests {
		result := parseFloat(tt.input)
		if result != tt.expected {
			t.Errorf("Expected %f, got %f for input %v", tt.expected, result, tt.input)
		}
	}
}

func TestParseFloatLargeNumbers(t *testing.T) {
	// Test parsing of large numbers (like volumes)
	largeNum := "50000000000.123456" // 50 billion
	result := parseFloat(largeNum)

	expected := 50000000000.123456
	if math.Abs(result-expected) > 1.0 { // Allow small error for large numbers
		t.Errorf("Expected approximately %f, got %f", expected, result)
	}
}

func TestParseFloatSmallNumbers(t *testing.T) {
	// Test parsing of very small numbers
	smallNum := "0.00000123"
	result := parseFloat(smallNum)

	expected := 0.00000123
	if math.Abs(result-expected) > 0.000000001 {
		t.Errorf("Expected %f, got %f", expected, result)
	}
}

func TestBinanceKlineType(t *testing.T) {
	// Test that BinanceKline type can hold interface{} values
	var kline BinanceKline = []interface{}{
		float64(1234567890), // timestamp
		"45000.00",          // open
		"46000.00",          // high
		"44000.00",          // low
		"45500.00",          // close
		"1000000000",        // volume
	}

	if len(kline) != 6 {
		t.Errorf("Expected 6 elements, got %d", len(kline))
	}

	// Test parsing each element
	timestamp := parseFloat(kline[0])
	if timestamp != 1234567890 {
		t.Errorf("Failed to parse timestamp: %f", timestamp)
	}

	open := parseFloat(kline[1])
	if open != 45000.00 {
		t.Errorf("Failed to parse open: %f", open)
	}

	volume := parseFloat(kline[5])
	if volume != 1000000000 {
		t.Errorf("Failed to parse volume: %f", volume)
	}
}

func TestParseFloatRealisticBinanceValues(t *testing.T) {
	// Test with realistic values from Binance API
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{"BTC price", "45123.56", 45123.56},
		{"Volume", "28500000000.123", 28500000000.123},
		{"Low precision price", "0.00001234", 0.00001234},
		{"High volume", "999999999999.99", 999999999999.99},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseFloat(tt.input)
			// Allow small relative error
			relError := math.Abs((result - tt.expected) / tt.expected)
			if relError > 0.000001 && math.Abs(result-tt.expected) > 0.01 {
				t.Errorf("Expected %f, got %f (relative error: %f)",
					tt.expected, result, relError)
			}
		})
	}
}

func TestParseFloatZeroValues(t *testing.T) {
	zeros := []interface{}{
		float64(0),
		"0",
		"0.0",
		"0.00000",
		int(0),
		int64(0),
	}

	for i, zero := range zeros {
		result := parseFloat(zero)
		if result != 0.0 {
			t.Errorf("Zero test %d: Expected 0.0, got %f", i, result)
		}
	}
}
