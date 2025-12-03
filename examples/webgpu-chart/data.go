//go:build js && wasm

package main

import "github.com/gaarutyunov/guix/pkg/runtime/chart"

// ChartData holds sample data for charts
type ChartData struct {
	Bitcoin []chart.OHLCV
}

// OHLCV is an alias for chart.OHLCV to make it accessible in .gx files
type OHLCV = chart.OHLCV

// ChartConfig holds chart configuration
type ChartConfig struct {
	Data       []chart.OHLCV
	Title      string
	Background ChartColor
}

// ChartColor represents an RGBA color
type ChartColor struct {
	R, G, B, A float32
}

// NewChartConfig creates a new chart configuration with defaults
func NewChartConfig() *ChartConfig {
	return &ChartConfig{
		Background: ChartColor{R: 0.08, G: 0.09, B: 0.12, A: 1.0},
	}
}

// GetFallbackData generates realistic fallback Bitcoin OHLCV data using Markov chains
// This generates initial data when Binance API is unavailable
func GetFallbackData() []chart.OHLCV {
	// Generate 1000 candles using Markov chain algorithm for realistic price action
	return GenerateFallbackData(1000)
}

// GetChartData fetches Bitcoin data from Binance API with fallback to static data
// This function will attempt to fetch 1000 candles from Binance (hourly data = ~41 days)
// If the fetch fails (e.g., CORS in browser), it falls back to static sample data
func GetChartData() []chart.OHLCV {
	// Try to fetch from Binance (may fail due to CORS in browser)
	data, err := FetchBinanceData("BTCUSDT", "1h", 1000)
	if err != nil {
		// Silently fall back to static data
		return GetFallbackData()
	}

	return data
}
