//go:build js && wasm

package main

import (
	"testing"

	"github.com/gaarutyunov/guix/pkg/runtime/chart"
)

func TestNewChartDataManager(t *testing.T) {
	initialData := GenerateFallbackData(500)
	symbol := "BTCUSDT"
	interval := "1h"

	cdm := NewChartDataManager(initialData, symbol, interval)

	if cdm == nil {
		t.Fatal("Expected ChartDataManager to be non-nil")
	}

	if len(cdm.allData) != 500 {
		t.Errorf("Expected 500 data points, got %d", len(cdm.allData))
	}

	if cdm.symbol != symbol {
		t.Errorf("Expected symbol %s, got %s", symbol, cdm.symbol)
	}

	if cdm.interval != interval {
		t.Errorf("Expected interval %s, got %s", interval, cdm.interval)
	}

	if cdm.generator == nil {
		t.Error("Expected generator to be initialized")
	}

	// Should show 200 candles at a time
	if cdm.visibleEnd != 200 {
		t.Errorf("Expected visibleEnd=200, got %d", cdm.visibleEnd)
	}

	if cdm.visibleStart != 0 {
		t.Errorf("Expected visibleStart=0, got %d", cdm.visibleStart)
	}
}

func TestNewChartDataManagerSmallDataset(t *testing.T) {
	// Test with fewer than 200 candles
	initialData := GenerateFallbackData(50)
	cdm := NewChartDataManager(initialData, "BTCUSDT", "1h")

	// Should show all 50 candles
	if cdm.visibleEnd != 50 {
		t.Errorf("Expected visibleEnd=50, got %d", cdm.visibleEnd)
	}
}

func TestNewChartDataManagerEmptyData(t *testing.T) {
	// Test with empty data
	var initialData []chart.OHLCV
	cdm := NewChartDataManager(initialData, "BTCUSDT", "1h")

	if cdm == nil {
		t.Fatal("Expected ChartDataManager to be non-nil even with empty data")
	}

	if cdm.visibleEnd != 0 {
		t.Errorf("Expected visibleEnd=0, got %d", cdm.visibleEnd)
	}

	// Generator should still be initialized with default price
	if cdm.generator == nil {
		t.Error("Expected generator to be initialized")
	}
}

func TestGetVisibleData(t *testing.T) {
	initialData := GenerateFallbackData(500)
	cdm := NewChartDataManager(initialData, "BTCUSDT", "1h")

	visibleData := cdm.GetVisibleData()

	if len(visibleData) != 200 {
		t.Errorf("Expected 200 visible candles, got %d", len(visibleData))
	}

	// Visible data should be a slice of the original data
	for i, candle := range visibleData {
		if candle.Timestamp != initialData[i].Timestamp {
			t.Errorf("Visible data mismatch at index %d", i)
		}
	}
}

func TestGetVisibleDataBounds(t *testing.T) {
	initialData := GenerateFallbackData(100)
	cdm := NewChartDataManager(initialData, "BTCUSDT", "1h")

	// Set invalid bounds
	cdm.visibleEnd = 200 // More than available data
	visibleData := cdm.GetVisibleData()

	// Should clamp to available data
	if len(visibleData) != 100 {
		t.Errorf("Expected clamped length of 100, got %d", len(visibleData))
	}

	// Test negative start
	cdm.visibleStart = -10
	cdm.visibleEnd = 50
	_ = cdm.GetVisibleData()

	if cdm.visibleStart != 0 {
		t.Errorf("Expected clamped start of 0, got %d", cdm.visibleStart)
	}
}

func TestShiftViewport(t *testing.T) {
	initialData := GenerateFallbackData(500)
	cdm := NewChartDataManager(initialData, "BTCUSDT", "1h")

	// Initial state: 0-200
	if cdm.visibleStart != 0 || cdm.visibleEnd != 200 {
		t.Errorf("Initial state incorrect: start=%d, end=%d",
			cdm.visibleStart, cdm.visibleEnd)
	}

	// Shift right (scroll to newer data)
	changed := cdm.ShiftViewport(50)
	if !changed {
		t.Error("Expected viewport to change")
	}

	if cdm.visibleStart != 50 || cdm.visibleEnd != 250 {
		t.Errorf("After shift: expected start=50, end=250, got start=%d, end=%d",
			cdm.visibleStart, cdm.visibleEnd)
	}
}

func TestShiftViewportLeftBound(t *testing.T) {
	initialData := GenerateFallbackData(500)
	cdm := NewChartDataManager(initialData, "BTCUSDT", "1h")

	// Try to shift left beyond boundary (already at start=0)
	changed := cdm.ShiftViewport(-100)

	// Should clamp to 0
	if cdm.visibleStart != 0 {
		t.Errorf("Expected start=0, got %d", cdm.visibleStart)
	}

	if cdm.visibleEnd != 200 {
		t.Errorf("Expected end=200, got %d", cdm.visibleEnd)
	}

	// Since we're already at the boundary, nothing changed
	if changed {
		t.Error("Expected changed to be false when already at boundary")
	}
}

func TestShiftViewportRightBound(t *testing.T) {
	initialData := GenerateFallbackData(500)
	cdm := NewChartDataManager(initialData, "BTCUSDT", "1h")

	// Shift to near the end
	cdm.ShiftViewport(400)

	// Should clamp to available data
	if cdm.visibleEnd > 500 {
		t.Errorf("visibleEnd should not exceed data length: got %d", cdm.visibleEnd)
	}

	if cdm.visibleEnd != 500 {
		t.Errorf("Expected end=500, got %d", cdm.visibleEnd)
	}
}

func TestShiftViewportPrefetchTrigger(t *testing.T) {
	initialData := GenerateFallbackData(500)
	cdm := NewChartDataManager(initialData, "BTCUSDT", "1h")
	cdm.prefetchSize = 100

	// Shift to near the end (within prefetch window)
	cdm.ShiftViewport(350) // visibleEnd will be 450, which is > 500-100

	// Check if we're in prefetch range
	inPrefetchRange := cdm.visibleEnd > len(cdm.allData)-cdm.prefetchSize
	if !inPrefetchRange {
		t.Error("Expected to be in prefetch range")
	}
}

func TestShiftViewportZeroDelta(t *testing.T) {
	initialData := GenerateFallbackData(500)
	cdm := NewChartDataManager(initialData, "BTCUSDT", "1h")

	oldStart := cdm.visibleStart
	oldEnd := cdm.visibleEnd

	// Shift by zero
	changed := cdm.ShiftViewport(0)

	if cdm.visibleStart != oldStart || cdm.visibleEnd != oldEnd {
		t.Error("Viewport should not change with zero delta")
	}

	if changed {
		t.Error("Expected changed to be false with zero delta")
	}
}

func TestShiftViewportMaintainsSize(t *testing.T) {
	initialData := GenerateFallbackData(500)
	cdm := NewChartDataManager(initialData, "BTCUSDT", "1h")

	initialSize := cdm.visibleEnd - cdm.visibleStart

	// Shift multiple times
	for i := 0; i < 10; i++ {
		cdm.ShiftViewport(10)

		currentSize := cdm.visibleEnd - cdm.visibleStart

		// Size should remain constant (unless at boundaries)
		if currentSize != initialSize && cdm.visibleEnd < 500 {
			t.Errorf("Viewport size changed: initial=%d, current=%d",
				initialSize, currentSize)
		}
	}
}

func TestChartDataManagerPrefetchSize(t *testing.T) {
	initialData := GenerateFallbackData(500)
	cdm := NewChartDataManager(initialData, "BTCUSDT", "1h")

	if cdm.prefetchSize != 100 {
		t.Errorf("Expected prefetchSize=100, got %d", cdm.prefetchSize)
	}
}

func TestChartDataManagerTotalFetched(t *testing.T) {
	initialData := GenerateFallbackData(500)
	cdm := NewChartDataManager(initialData, "BTCUSDT", "1h")

	if cdm.totalFetched != 500 {
		t.Errorf("Expected totalFetched=500, got %d", cdm.totalFetched)
	}
}

func TestChartDataManagerIsFetching(t *testing.T) {
	initialData := GenerateFallbackData(500)
	cdm := NewChartDataManager(initialData, "BTCUSDT", "1h")

	if cdm.isFetching {
		t.Error("Expected isFetching to be false initially")
	}
}

func TestShiftViewportMultipleDirections(t *testing.T) {
	initialData := GenerateFallbackData(500)
	cdm := NewChartDataManager(initialData, "BTCUSDT", "1h")

	// Shift right
	cdm.ShiftViewport(100)
	rightPos := cdm.visibleStart

	// Shift left
	cdm.ShiftViewport(-50)
	leftPos := cdm.visibleStart

	if rightPos-50 != leftPos {
		t.Errorf("Position inconsistent: right=%d, left=%d", rightPos, leftPos)
	}

	// Should be at position 50
	if leftPos != 50 {
		t.Errorf("Expected position 50, got %d", leftPos)
	}
}

func TestGetVisibleDataAfterShift(t *testing.T) {
	initialData := GenerateFallbackData(500)
	cdm := NewChartDataManager(initialData, "BTCUSDT", "1h")

	// Shift viewport
	cdm.ShiftViewport(100)

	visibleData := cdm.GetVisibleData()

	// Should show candles 100-300
	if len(visibleData) != 200 {
		t.Errorf("Expected 200 visible candles, got %d", len(visibleData))
	}

	// First visible candle should be the 100th candle from original data
	if visibleData[0].Timestamp != initialData[100].Timestamp {
		t.Error("Visible data doesn't match expected range after shift")
	}
}

func TestChartDataManagerGeneratorInitialization(t *testing.T) {
	// Test with data
	initialData := GenerateFallbackData(100)
	cdm := NewChartDataManager(initialData, "BTCUSDT", "1h")

	if cdm.generator == nil {
		t.Fatal("Expected generator to be initialized")
	}

	// Generator should start from last candle's close price
	lastPrice := initialData[len(initialData)-1].Close
	if cdm.generator.currentPrice != lastPrice {
		t.Errorf("Expected generator price to be %f, got %f",
			lastPrice, cdm.generator.currentPrice)
	}
}
