//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"syscall/js"
	"time"

	"github.com/gaarutyunov/guix/pkg/runtime/chart"
)

// BinanceKline represents a Binance kline (candlestick) response item
type BinanceKline []interface{}

// FetchBinanceData fetches OHLCV data from Binance public API
// symbol: trading pair (e.g., "BTCUSDT")
// interval: time interval (e.g., "1h", "4h", "1d")
// limit: number of candles to fetch (max 1000)
// endTime: optional end time in milliseconds (0 for latest data)
// Note: This may fail due to CORS when called from browser, in which case fallback data is used
func FetchBinanceDataWithEndTime(symbol, interval string, limit int, endTime int64) ([]chart.OHLCV, error) {
	// Construct Binance API URL
	url := fmt.Sprintf("https://api.binance.com/api/v3/klines?symbol=%s&interval=%s&limit=%d",
		symbol, interval, limit)

	// Add endTime parameter if specified
	if endTime > 0 {
		url += fmt.Sprintf("&endTime=%d", endTime)
	}

	// Create a promise to fetch data
	promise := js.Global().Call("fetch", url)

	// Wait for the promise to resolve
	resultChan := make(chan js.Value, 1)
	errorChan := make(chan error, 1)

	onSuccess := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		response := args[0]

		// Check if response is OK
		if !response.Get("ok").Bool() {
			errorChan <- fmt.Errorf("HTTP error: %d", response.Get("status").Int())
			return nil
		}

		// Parse JSON
		jsonPromise := response.Call("json")
		onJSONSuccess := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			resultChan <- args[0]
			return nil
		})
		defer onJSONSuccess.Release()

		onJSONError := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			errorChan <- fmt.Errorf("JSON parse error")
			return nil
		})
		defer onJSONError.Release()

		jsonPromise.Call("then", onJSONSuccess).Call("catch", onJSONError)
		return nil
	})
	defer onSuccess.Release()

	onError := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// CORS or network error - silently fail and use fallback
		errorChan <- fmt.Errorf("network error")
		return nil
	})
	defer onError.Release()

	promise.Call("then", onSuccess).Call("catch", onError)

	// Wait for result or error with timeout (reduced to 2s since CORS fails immediately)
	select {
	case result := <-resultChan:
		return parseBinanceKlines(result)
	case err := <-errorChan:
		return nil, err
	case <-time.After(2 * time.Second):
		return nil, fmt.Errorf("timeout")
	}
}

// parseBinanceKlines converts Binance kline data to OHLCV format
func parseBinanceKlines(data js.Value) ([]chart.OHLCV, error) {
	// Convert js.Value to JSON string
	jsonStr := js.Global().Get("JSON").Call("stringify", data).String()

	// Parse JSON
	var rawKlines [][]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &rawKlines); err != nil {
		return nil, fmt.Errorf("failed to parse kline data: %w", err)
	}

	// Convert to OHLCV format
	ohlcvData := make([]chart.OHLCV, 0, len(rawKlines))
	for _, kline := range rawKlines {
		if len(kline) < 11 {
			continue
		}

		// Binance kline format:
		// [
		//   0: Open time (int64)
		//   1: Open (string)
		//   2: High (string)
		//   3: Low (string)
		//   4: Close (string)
		//   5: Volume (string)
		//   6: Close time
		//   7: Quote asset volume
		//   8: Number of trades
		//   9: Taker buy base asset volume
		//   10: Taker buy quote asset volume
		//   11: Ignore
		// ]

		timestamp := int64(kline[0].(float64))

		// Parse price strings to float64
		open := parseFloat(kline[1])
		high := parseFloat(kline[2])
		low := parseFloat(kline[3])
		close := parseFloat(kline[4])
		volume := parseFloat(kline[5])

		ohlcvData = append(ohlcvData, chart.OHLCV{
			Timestamp: timestamp,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
		})
	}

	return ohlcvData, nil
}

// parseFloat converts interface{} to float64
func parseFloat(val interface{}) float64 {
	switch v := val.(type) {
	case float64:
		// Handle NaN and Inf
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return 0
		}
		return v
	case string:
		var f float64
		_, _ = fmt.Sscanf(v, "%f", &f)
		// Handle NaN and Inf from parsed strings
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return 0
		}
		return f
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return 0
	}
}

// FetchBinanceData fetches latest data from Binance (wrapper for backward compatibility)
func FetchBinanceData(symbol, interval string, limit int) ([]chart.OHLCV, error) {
	return FetchBinanceDataWithEndTime(symbol, interval, limit, 0)
}

// FetchBinanceDataAsync fetches data asynchronously and returns a channel
func FetchBinanceDataAsync(symbol, interval string, limit int) <-chan []chart.OHLCV {
	resultChan := make(chan []chart.OHLCV, 1)

	go func() {
		data, err := FetchBinanceData(symbol, interval, limit)
		if err != nil {
			// Send fallback data on error (CORS, timeout, etc.)
			resultChan <- GetFallbackData()
		} else {
			resultChan <- data
		}
		close(resultChan)
	}()

	return resultChan
}
