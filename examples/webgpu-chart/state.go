//go:build js && wasm

package main

import "github.com/gaarutyunov/guix/pkg/runtime/chart"

// Global state for chart data management
var (
	globalScrollManager *ScrollManager
	globalApp           *App
	chartDataChannel    chan []chart.OHLCV
)

// InitChartDataChannel initializes the global chart data channel
func InitChartDataChannel() chan []chart.OHLCV {
	if chartDataChannel == nil {
		chartDataChannel = make(chan []chart.OHLCV, 1) // Buffered channel
	}
	return chartDataChannel
}

// GetChartDataChannel returns the global chart data channel
func GetChartDataChannel() chan []chart.OHLCV {
	return chartDataChannel
}

// GetVisibleChartData returns the currently visible chart data
// This is called by the app during rendering to get only the visible window of data
func GetVisibleChartData() []chart.OHLCV {
	if globalScrollManager != nil && globalScrollManager.chartData != nil {
		return globalScrollManager.chartData.GetVisibleData()
	}
	// Fallback: return initial data if scroll manager not initialized
	return GetChartData()
}

// SetGlobalScrollManager sets the global scroll manager instance
func SetGlobalScrollManager(sm *ScrollManager) {
	globalScrollManager = sm
}

// SetGlobalApp sets the global app instance
func SetGlobalApp(app *App) {
	globalApp = app
}

// SendChartDataUpdate sends new chart data through the channel
func SendChartDataUpdate(data []chart.OHLCV) {
	if chartDataChannel != nil {
		// Non-blocking send - if channel is full, drain it first
		select {
		case chartDataChannel <- data:
			// Successfully sent
		default:
			// Channel full, drain and send new data
			select {
			case <-chartDataChannel:
			default:
			}
			chartDataChannel <- data
		}
	}
}
