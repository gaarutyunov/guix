//go:build js && wasm

package chart

import "github.com/gaarutyunov/guix/pkg/runtime"

// CandleSeriesConfig holds configuration for candlestick series
type CandleSeriesConfig struct {
	Data        []OHLCV
	DataChannel chan []OHLCV
	UpColor     runtime.Vec4
	DownColor   runtime.Vec4
	WickColor   runtime.Vec4
	BarWidth    float32 // 0.0-1.0 ratio of available space
}

// CandlestickSeries creates a candlestick series node
func CandlestickSeries(opts ...SeriesOption) *runtime.GPUNode {
	node := &runtime.GPUNode{
		Type:       runtime.GroupNodeType,
		Tag:        SeriesNodeType,
		Properties: make(map[string]interface{}),
		Transform:  runtime.NewTransform(),
	}

	// Set defaults
	config := &CandleSeriesConfig{
		Data:      []OHLCV{},
		UpColor:   runtime.NewVec4(0.18, 0.80, 0.44, 1.0), // Green
		DownColor: runtime.NewVec4(0.91, 0.27, 0.38, 1.0), // Red
		WickColor: runtime.NewVec4(0.6, 0.6, 0.65, 1.0),   // Gray
		BarWidth:  0.8,
	}

	node.Properties["seriesType"] = "candlestick"
	node.Properties["config"] = config

	for _, opt := range opts {
		opt(node)
	}

	return node
}

// OHLCSeries creates an OHLC bar series node (alternative to candlesticks)
func OHLCSeries(opts ...SeriesOption) *runtime.GPUNode {
	node := CandlestickSeries(opts...)
	node.Properties["seriesType"] = "ohlc"
	return node
}

// DataProp sets static data for the series
func DataProp(data []OHLCV) SeriesOption {
	return func(node *runtime.GPUNode) {
		if config, ok := node.Properties["config"].(*CandleSeriesConfig); ok {
			config.Data = data
		}
	}
}

// DataChannel sets a reactive data channel for the series
func DataChannel(ch chan []OHLCV) SeriesOption {
	return func(node *runtime.GPUNode) {
		if config, ok := node.Properties["config"].(*CandleSeriesConfig); ok {
			config.DataChannel = ch
		}
	}
}

// UpColor sets the color for up candles (close >= open)
func UpColor(r, g, b, a float32) SeriesOption {
	return func(node *runtime.GPUNode) {
		if config, ok := node.Properties["config"].(*CandleSeriesConfig); ok {
			config.UpColor = runtime.NewVec4(r, g, b, a)
		}
	}
}

// DownColor sets the color for down candles (close < open)
func DownColor(r, g, b, a float32) SeriesOption {
	return func(node *runtime.GPUNode) {
		if config, ok := node.Properties["config"].(*CandleSeriesConfig); ok {
			config.DownColor = runtime.NewVec4(r, g, b, a)
		}
	}
}

// WickColor sets the color for candle wicks
func WickColor(r, g, b, a float32) SeriesOption {
	return func(node *runtime.GPUNode) {
		if config, ok := node.Properties["config"].(*CandleSeriesConfig); ok {
			config.WickColor = runtime.NewVec4(r, g, b, a)
		}
	}
}

// BarWidth sets the width ratio for candles (0.0-1.0)
func BarWidth(ratio float32) SeriesOption {
	return func(node *runtime.GPUNode) {
		if config, ok := node.Properties["config"].(*CandleSeriesConfig); ok {
			if ratio < 0 {
				ratio = 0
			} else if ratio > 1 {
				ratio = 1
			}
			config.BarWidth = ratio
		}
	}
}
