//go:build js && wasm

package chart

// OHLCV represents a single candlestick data point
type OHLCV struct {
	Timestamp int64   // Unix timestamp in milliseconds
	Open      float64 // Opening price
	High      float64 // Highest price
	Low       float64 // Lowest price
	Close     float64 // Closing price
	Volume    float64 // Trading volume
}

// Point represents a generic 2D data point
type Point struct {
	X float64
	Y float64
}

// TimeValue represents a time-series data point
type TimeValue struct {
	Timestamp int64   // Unix timestamp in milliseconds
	Value     float64 // Data value
}

// Range represents a min-max range for axes
type Range struct {
	Min float64
	Max float64
}

// Padding represents padding around chart area
type Padding struct {
	Top    float32
	Right  float32
	Bottom float32
	Left   float32
}

// Viewport represents the visible data range
type Viewport struct {
	XRange Range // Visible X-axis range
	YRange Range // Visible Y-axis range
}

// CalculateOHLCVRange calculates the Y-axis range for OHLCV data
func CalculateOHLCVRange(data []OHLCV) Range {
	if len(data) == 0 {
		return Range{Min: 0, Max: 1}
	}

	min := data[0].Low
	max := data[0].High

	for _, candle := range data {
		if candle.Low < min {
			min = candle.Low
		}
		if candle.High > max {
			max = candle.High
		}
	}

	// Add 5% padding to the range
	padding := (max - min) * 0.05
	return Range{
		Min: min - padding,
		Max: max + padding,
	}
}

// CalculateTimeRange calculates the X-axis time range for OHLCV data
func CalculateTimeRange(data []OHLCV) Range {
	if len(data) == 0 {
		return Range{Min: 0, Max: 1}
	}

	min := float64(data[0].Timestamp)
	max := float64(data[len(data)-1].Timestamp)

	return Range{Min: min, Max: max}
}

// CalculatePointRange calculates the range for Point data
func CalculatePointRange(data []Point, axis string) Range {
	if len(data) == 0 {
		return Range{Min: 0, Max: 1}
	}

	var min, max float64
	if axis == "x" {
		min, max = data[0].X, data[0].X
		for _, p := range data {
			if p.X < min {
				min = p.X
			}
			if p.X > max {
				max = p.X
			}
		}
	} else {
		min, max = data[0].Y, data[0].Y
		for _, p := range data {
			if p.Y < min {
				min = p.Y
			}
			if p.Y > max {
				max = p.Y
			}
		}
	}

	return Range{Min: min, Max: max}
}
