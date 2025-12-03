//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
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
func FetchBinanceData(symbol, interval string, limit int) ([]chart.OHLCV, error) {
	log("[Binance] Fetching data for", symbol, "interval:", interval, "limit:", limit)

	// Construct Binance API URL
	url := fmt.Sprintf("https://api.binance.com/api/v3/klines?symbol=%s&interval=%s&limit=%d",
		symbol, interval, limit)

	log("[Binance] URL:", url)

	// Create a promise to fetch data
	promise := js.Global().Call("fetch", url)

	// Wait for the promise to resolve
	resultChan := make(chan js.Value, 1)
	errorChan := make(chan error, 1)

	onSuccess := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		response := args[0]
		log("[Binance] Fetch successful, status:", response.Get("status").Int())

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
			errorChan <- fmt.Errorf("JSON parse error: %v", args[0].String())
			return nil
		})
		defer onJSONError.Release()

		jsonPromise.Call("then", onJSONSuccess).Call("catch", onJSONError)
		return nil
	})
	defer onSuccess.Release()

	onError := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		err := fmt.Errorf("Fetch error: %v", args[0].String())
		errorChan <- err
		return nil
	})
	defer onError.Release()

	promise.Call("then", onSuccess).Call("catch", onError)

	// Wait for result or error with timeout
	select {
	case result := <-resultChan:
		log("[Binance] Data received, parsing...")
		return parseBinanceKlines(result)
	case err := <-errorChan:
		log("[Binance] Error:", err.Error())
		return nil, err
	case <-time.After(10 * time.Second):
		log("[Binance] Timeout after 10 seconds")
		return nil, fmt.Errorf("timeout fetching data")
	}
}

// parseBinanceKlines converts Binance kline data to OHLCV format
func parseBinanceKlines(data js.Value) ([]chart.OHLCV, error) {
	log("[Binance] Parsing kline data...")

	// Convert js.Value to JSON string
	jsonStr := js.Global().Get("JSON").Call("stringify", data).String()

	// Parse JSON
	var rawKlines [][]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &rawKlines); err != nil {
		log("[Binance] JSON unmarshal error:", err.Error())
		return nil, fmt.Errorf("failed to parse kline data: %w", err)
	}

	log("[Binance] Parsed", len(rawKlines), "candles")

	// Convert to OHLCV format
	ohlcvData := make([]chart.OHLCV, 0, len(rawKlines))
	for i, kline := range rawKlines {
		if len(kline) < 11 {
			log("[Binance] Skipping invalid kline at index", i)
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

	log("[Binance] Successfully parsed", len(ohlcvData), "OHLCV candles")
	if len(ohlcvData) > 0 {
		log("[Binance] First candle timestamp:", ohlcvData[0].Timestamp, "price:", ohlcvData[0].Close)
		log("[Binance] Last candle timestamp:", ohlcvData[len(ohlcvData)-1].Timestamp, "price:", ohlcvData[len(ohlcvData)-1].Close)
	}

	return ohlcvData, nil
}

// parseFloat converts interface{} to float64
func parseFloat(val interface{}) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case string:
		var f float64
		fmt.Sscanf(v, "%f", &f)
		return f
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return 0
	}
}

// FetchBinanceDataAsync fetches data asynchronously and returns a channel
func FetchBinanceDataAsync(symbol, interval string, limit int) <-chan []chart.OHLCV {
	resultChan := make(chan []chart.OHLCV, 1)

	go func() {
		data, err := FetchBinanceData(symbol, interval, limit)
		if err != nil {
			log("[Binance] Fetch error:", err.Error())
			// Send fallback data on error
			resultChan <- GetFallbackData()
		} else {
			resultChan <- data
		}
		close(resultChan)
	}()

	return resultChan
}
