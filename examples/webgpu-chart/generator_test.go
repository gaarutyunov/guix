//go:build js && wasm

package main

import (
	"testing"
	"time"
)

func TestNewMarkovChainGenerator(t *testing.T) {
	startPrice := 45000.0
	volatility := 1.0

	gen := NewMarkovChainGenerator(startPrice, volatility)

	if gen == nil {
		t.Fatal("Expected generator to be non-nil")
	}

	if gen.currentPrice != startPrice {
		t.Errorf("Expected currentPrice %f, got %f", startPrice, gen.currentPrice)
	}

	if gen.volatility != volatility {
		t.Errorf("Expected volatility %f, got %f", volatility, gen.volatility)
	}

	if gen.rand == nil {
		t.Error("Expected rand to be initialized")
	}
}

func TestGenerateCandles(t *testing.T) {
	gen := NewMarkovChainGenerator(45000.0, 1.0)
	count := 100
	interval := "1h"

	candles := gen.GenerateCandles(count, interval)

	if len(candles) != count {
		t.Errorf("Expected %d candles, got %d", count, len(candles))
	}

	// Verify OHLCV relationships
	for i, candle := range candles {
		// High should be >= max(open, close)
		maxPrice := candle.Open
		if candle.Close > maxPrice {
			maxPrice = candle.Close
		}
		if candle.High < maxPrice {
			t.Errorf("Candle %d: High (%f) should be >= max(Open, Close) (%f)",
				i, candle.High, maxPrice)
		}

		// Low should be <= min(open, close)
		minPrice := candle.Open
		if candle.Close < minPrice {
			minPrice = candle.Close
		}
		if candle.Low > minPrice {
			t.Errorf("Candle %d: Low (%f) should be <= min(Open, Close) (%f)",
				i, candle.Low, minPrice)
		}

		// Volume should be within reasonable range
		if candle.Volume < 100000000 || candle.Volume > 10000000000 {
			t.Errorf("Candle %d: Volume (%f) out of expected range", i, candle.Volume)
		}

		// Prices should be positive
		if candle.Open <= 0 || candle.High <= 0 || candle.Low <= 0 || candle.Close <= 0 {
			t.Errorf("Candle %d: All prices should be positive", i)
		}
	}
}

func TestGenerateCandlesTimestamps(t *testing.T) {
	gen := NewMarkovChainGenerator(45000.0, 1.0)
	count := 50
	interval := "1h"
	intervalMs := int64(60 * 60 * 1000)

	now := time.Now().UnixMilli()
	candles := gen.GenerateCandles(count, interval)

	// First candle should be in the past
	if candles[0].Timestamp > now {
		t.Error("First candle timestamp should be in the past")
	}

	// Timestamps should be sequential with correct intervals
	for i := 1; i < len(candles); i++ {
		expectedDelta := intervalMs
		actualDelta := candles[i].Timestamp - candles[i-1].Timestamp

		if actualDelta != expectedDelta {
			t.Errorf("Candle %d: Expected timestamp delta %d, got %d",
				i, expectedDelta, actualDelta)
		}
	}
}

func TestGetIntervalMilliseconds(t *testing.T) {
	tests := []struct {
		interval string
		expected int64
	}{
		{"1m", 60 * 1000},
		{"5m", 5 * 60 * 1000},
		{"15m", 15 * 60 * 1000},
		{"30m", 30 * 60 * 1000},
		{"1h", 60 * 60 * 1000},
		{"4h", 4 * 60 * 60 * 1000},
		{"1d", 24 * 60 * 60 * 1000},
		{"unknown", 60 * 60 * 1000}, // Default to 1h
	}

	for _, tt := range tests {
		t.Run(tt.interval, func(t *testing.T) {
			result := getIntervalMilliseconds(tt.interval)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestMarkovStateTransitions(t *testing.T) {
	gen := NewMarkovChainGenerator(45000.0, 1.0)

	// Test that getMarkovState returns valid states
	validStates := map[string]bool{
		"uptrend":   true,
		"downtrend": true,
		"ranging":   true,
	}

	// Run multiple times to test different states
	for i := 0; i < 100; i++ {
		state := gen.getMarkovState()
		if !validStates[state] {
			t.Errorf("Invalid state returned: %s", state)
		}
	}
}

func TestGenerateCandlesPriceMovement(t *testing.T) {
	startPrice := 45000.0
	gen := NewMarkovChainGenerator(startPrice, 1.0)

	candles := gen.GenerateCandles(100, "1h")

	// Check that prices don't jump too drastically
	for i := 1; i < len(candles); i++ {
		prevClose := candles[i-1].Close
		currentOpen := candles[i].Open

		// Price shouldn't change more than 10% between candles
		maxChange := prevClose * 0.1
		actualChange := currentOpen - prevClose

		if actualChange > maxChange || actualChange < -maxChange {
			t.Errorf("Candle %d: Price change too large: %f (%.2f%%)",
				i, actualChange, (actualChange/prevClose)*100)
		}
	}
}

func TestGenerateFallbackData(t *testing.T) {
	count := 1000
	data := GenerateFallbackData(count)

	if len(data) != count {
		t.Errorf("Expected %d candles, got %d", count, len(data))
	}

	// Verify first candle
	if data[0].Close <= 0 {
		t.Error("First candle price should be positive")
	}

	// Verify all candles have valid OHLCV relationships
	for i, candle := range data {
		if candle.High < candle.Open || candle.High < candle.Close {
			t.Errorf("Candle %d: High invalid", i)
		}
		if candle.Low > candle.Open || candle.Low > candle.Close {
			t.Errorf("Candle %d: Low invalid", i)
		}
		if candle.Volume <= 0 {
			t.Errorf("Candle %d: Volume should be positive", i)
		}
	}
}

func TestGenerateCandlesWithDifferentVolatility(t *testing.T) {
	tests := []struct {
		name       string
		volatility float64
	}{
		{"low volatility", 0.5},
		{"normal volatility", 1.0},
		{"high volatility", 2.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewMarkovChainGenerator(45000.0, tt.volatility)
			candles := gen.GenerateCandles(50, "1h")

			if len(candles) != 50 {
				t.Errorf("Expected 50 candles, got %d", len(candles))
			}

			// All candles should be valid regardless of volatility
			for i, candle := range candles {
				if candle.High < candle.Low {
					t.Errorf("Candle %d: High (%f) < Low (%f)",
						i, candle.High, candle.Low)
				}
			}
		})
	}
}

func TestMarkovChainContinuity(t *testing.T) {
	// Test that generating more candles continues from the last price
	gen := NewMarkovChainGenerator(45000.0, 1.0)

	firstBatch := gen.GenerateCandles(10, "1h")
	if len(firstBatch) == 0 {
		t.Fatal("First batch should have candles")
	}

	secondBatch := gen.GenerateCandles(10, "1h")
	if len(secondBatch) == 0 {
		t.Fatal("Second batch should have candles")
	}

	// The generator continues from currentPrice, so there should be
	// some relationship between batches (though not necessarily equal
	// due to random price movement)
	if secondBatch[0].Open <= 0 {
		t.Error("Second batch should have valid opening price")
	}

	// Generator's current price should be updated
	if gen.currentPrice <= 0 {
		t.Error("Generator's current price should be positive")
	}
}
