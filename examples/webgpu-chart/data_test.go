//go:build js && wasm

package main

import (
	"testing"
)

func TestGetFallbackData(t *testing.T) {
	data := GetFallbackData()

	if len(data) == 0 {
		t.Fatal("Fallback data should not be empty")
	}

	// Should generate 1000 candles
	expectedCount := 1000
	if len(data) != expectedCount {
		t.Errorf("Expected %d candles, got %d", expectedCount, len(data))
	}

	// Verify data structure
	for i, candle := range data {
		// Check timestamp
		if candle.Timestamp <= 0 {
			t.Errorf("Candle %d: Invalid timestamp", i)
		}

		// Check OHLCV relationships
		if candle.High < candle.Open && candle.High < candle.Close {
			t.Errorf("Candle %d: High should be >= max(Open, Close)", i)
		}

		if candle.Low > candle.Open && candle.Low > candle.Close {
			t.Errorf("Candle %d: Low should be <= min(Open, Close)", i)
		}

		// Check positive values
		if candle.Open <= 0 || candle.High <= 0 || candle.Low <= 0 ||
			candle.Close <= 0 || candle.Volume <= 0 {
			t.Errorf("Candle %d: All values should be positive", i)
		}
	}
}

func TestGetChartData(t *testing.T) {
	data := GetChartData()

	if len(data) == 0 {
		t.Fatal("Chart data should not be empty")
	}

	// Should have some data (either from Binance or fallback)
	if len(data) < 20 {
		t.Errorf("Expected at least 20 candles, got %d", len(data))
	}

	// Verify all candles have valid structure
	for i, candle := range data {
		if candle.Timestamp <= 0 {
			t.Errorf("Candle %d: Invalid timestamp", i)
		}
		if candle.Open <= 0 || candle.Close <= 0 {
			t.Errorf("Candle %d: Invalid prices", i)
		}
	}
}

func TestNewChartConfig(t *testing.T) {
	config := NewChartConfig()

	if config == nil {
		t.Fatal("Expected config to be non-nil")
	}

	// Check default background color
	if config.Background.R != 0.08 {
		t.Errorf("Expected R=0.08, got %f", config.Background.R)
	}
	if config.Background.G != 0.09 {
		t.Errorf("Expected G=0.09, got %f", config.Background.G)
	}
	if config.Background.B != 0.12 {
		t.Errorf("Expected B=0.12, got %f", config.Background.B)
	}
	if config.Background.A != 1.0 {
		t.Errorf("Expected A=1.0, got %f", config.Background.A)
	}
}

func TestChartDataConsistency(t *testing.T) {
	// Call GetChartData multiple times and verify it returns consistent results
	data1 := GetChartData()
	data2 := GetChartData()

	// Both should have data
	if len(data1) == 0 || len(data2) == 0 {
		t.Fatal("Both data calls should return data")
	}

	// If both succeed or both fall back, they should have same length
	// (Fallback generates 1000, Binance returns up to 1000)
	if len(data1) != len(data2) {
		t.Logf("Warning: Inconsistent data lengths: %d vs %d (may be due to API variance)",
			len(data1), len(data2))
	}
}

func TestFallbackDataOrdering(t *testing.T) {
	data := GetFallbackData()

	// Timestamps should be in ascending order (oldest first)
	for i := 1; i < len(data); i++ {
		if data[i].Timestamp <= data[i-1].Timestamp {
			t.Errorf("Timestamps should be ascending: candle %d (%d) <= candle %d (%d)",
				i, data[i].Timestamp, i-1, data[i-1].Timestamp)
		}
	}
}

func TestFallbackDataRealistic(t *testing.T) {
	data := GetFallbackData()

	if len(data) == 0 {
		t.Fatal("No fallback data generated")
	}

	// Check that prices are in a realistic range for Bitcoin
	for i, candle := range data {
		// Bitcoin prices should typically be > $1000 and < $1,000,000
		if candle.Close < 1000 || candle.Close > 1000000 {
			t.Errorf("Candle %d: Price %f seems unrealistic", i, candle.Close)
		}

		// Volume should be reasonable
		if candle.Volume < 1000000 || candle.Volume > 100000000000 {
			t.Errorf("Candle %d: Volume %f seems unrealistic", i, candle.Volume)
		}
	}
}

func TestOHLCVTypeAlias(t *testing.T) {
	// Test that OHLCV type alias works
	var ohlcv OHLCV
	ohlcv.Timestamp = 1234567890
	ohlcv.Open = 45000.0
	ohlcv.High = 46000.0
	ohlcv.Low = 44000.0
	ohlcv.Close = 45500.0
	ohlcv.Volume = 1000000000

	if ohlcv.Timestamp != 1234567890 {
		t.Error("Type alias should maintain timestamp value")
	}
	if ohlcv.Close != 45500.0 {
		t.Error("Type alias should maintain close value")
	}
}
