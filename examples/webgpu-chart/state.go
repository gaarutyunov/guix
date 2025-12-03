//go:build js && wasm

package main

import "github.com/gaarutyunov/guix/pkg/runtime/chart"

// Global state for chart data management
var (
	globalScrollManager *ScrollManager
	globalApp           *App
)

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

// TriggerChartUpdate triggers the app to rerender with updated data
func TriggerChartUpdate() {
	if globalApp != nil {
		globalApp.Update()
	}
}
