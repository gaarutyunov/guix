//go:build js && wasm

package chart

import (
	"fmt"
	"math"
	"time"

	"github.com/gaarutyunov/guix/pkg/runtime"
)

// AxisPosition defines where an axis is positioned
type AxisPosition string

const (
	AxisTop    AxisPosition = "top"
	AxisBottom AxisPosition = "bottom"
	AxisLeft   AxisPosition = "left"
	AxisRight  AxisPosition = "right"
)

// AxisConfig holds axis configuration
type AxisConfig struct {
	Position   AxisPosition
	Label      string
	Range      Range
	TickCount  int
	TickFormat func(float64) string
	GridLines  bool
	GridColor  runtime.Vec4
	TimeScale  bool
}

// XAxis creates an X-axis node
func XAxis(opts ...AxisOption) *runtime.GPUNode {
	node := &runtime.GPUNode{
		Type:       runtime.GroupNodeType,
		Tag:        AxisNodeType,
		Properties: make(map[string]interface{}),
		Transform:  runtime.NewTransform(),
	}

	// Set defaults for X-axis
	config := &AxisConfig{
		Position:   AxisBottom,
		TickCount:  10,
		TickFormat: func(v float64) string { return fmt.Sprintf("%.2f", v) },
		GridLines:  true,
		GridColor:  runtime.NewVec4(0.2, 0.2, 0.25, 0.5),
		TimeScale:  false,
	}
	node.Properties["axisType"] = "x"
	node.Properties["config"] = config

	for _, opt := range opts {
		opt(node)
	}

	return node
}

// YAxis creates a Y-axis node
func YAxis(opts ...AxisOption) *runtime.GPUNode {
	node := &runtime.GPUNode{
		Type:       runtime.GroupNodeType,
		Tag:        AxisNodeType,
		Properties: make(map[string]interface{}),
		Transform:  runtime.NewTransform(),
	}

	// Set defaults for Y-axis
	config := &AxisConfig{
		Position:   AxisLeft,
		TickCount:  10,
		TickFormat: func(v float64) string { return fmt.Sprintf("%.2f", v) },
		GridLines:  true,
		GridColor:  runtime.NewVec4(0.2, 0.2, 0.25, 0.5),
		TimeScale:  false,
	}
	node.Properties["axisType"] = "y"
	node.Properties["config"] = config

	for _, opt := range opts {
		opt(node)
	}

	return node
}

// Position sets the axis position
func AxisPos(pos AxisPosition) AxisOption {
	return func(node *runtime.GPUNode) {
		if config, ok := node.Properties["config"].(*AxisConfig); ok {
			config.Position = pos
		}
	}
}

// Label sets the axis label
func Label(text string) AxisOption {
	return func(node *runtime.GPUNode) {
		if config, ok := node.Properties["config"].(*AxisConfig); ok {
			config.Label = text
		}
	}
}

// AxisRange sets the axis range
func AxisRange(min, max float64) AxisOption {
	return func(node *runtime.GPUNode) {
		if config, ok := node.Properties["config"].(*AxisConfig); ok {
			config.Range = Range{Min: min, Max: max}
		}
	}
}

// TickCount sets the number of ticks on the axis
func TickCount(n int) AxisOption {
	return func(node *runtime.GPUNode) {
		if config, ok := node.Properties["config"].(*AxisConfig); ok {
			config.TickCount = n
		}
	}
}

// TickFormat sets the tick formatting function
func TickFormat(fn func(float64) string) AxisOption {
	return func(node *runtime.GPUNode) {
		if config, ok := node.Properties["config"].(*AxisConfig); ok {
			config.TickFormat = fn
		}
	}
}

// GridLines enables/disables grid lines
func GridLines(enabled bool) AxisOption {
	return func(node *runtime.GPUNode) {
		if config, ok := node.Properties["config"].(*AxisConfig); ok {
			config.GridLines = enabled
		}
	}
}

// GridColor sets the grid line color
func GridColor(r, g, b, a float32) AxisOption {
	return func(node *runtime.GPUNode) {
		if config, ok := node.Properties["config"].(*AxisConfig); ok {
			config.GridColor = runtime.NewVec4(r, g, b, a)
		}
	}
}

// TimeScale enables time-based scaling for the axis
func TimeScale(enabled bool) AxisOption {
	return func(node *runtime.GPUNode) {
		if config, ok := node.Properties["config"].(*AxisConfig); ok {
			config.TimeScale = enabled
		}
	}
}

// CalculateTicks generates tick values for an axis
func CalculateTicks(r Range, count int) []float64 {
	if count <= 0 {
		return []float64{}
	}

	ticks := make([]float64, count)
	step := (r.Max - r.Min) / float64(count-1)

	for i := 0; i < count; i++ {
		ticks[i] = r.Min + float64(i)*step
	}

	return ticks
}

// NiceNumber finds a "nice" number approximately equal to x
func NiceNumber(x float64, round bool) float64 {
	exp := math.Floor(math.Log10(x))
	f := x / math.Pow(10, exp)
	var nf float64

	if round {
		if f < 1.5 {
			nf = 1
		} else if f < 3 {
			nf = 2
		} else if f < 7 {
			nf = 5
		} else {
			nf = 10
		}
	} else {
		if f <= 1 {
			nf = 1
		} else if f <= 2 {
			nf = 2
		} else if f <= 5 {
			nf = 5
		} else {
			nf = 10
		}
	}

	return nf * math.Pow(10, exp)
}

// CalculateNiceTicks generates "nice" tick values
func CalculateNiceTicks(r Range, maxTicks int) []float64 {
	tickRange := NiceNumber(r.Max-r.Min, false)
	tickSpacing := NiceNumber(tickRange/float64(maxTicks-1), true)
	niceMin := math.Floor(r.Min/tickSpacing) * tickSpacing
	niceMax := math.Ceil(r.Max/tickSpacing) * tickSpacing

	ticks := []float64{}
	for tick := niceMin; tick <= niceMax; tick += tickSpacing {
		ticks = append(ticks, tick)
	}

	return ticks
}

// FormatTime formats a timestamp for display
func FormatTime(timestamp int64, format string) string {
	t := time.Unix(timestamp/1000, (timestamp%1000)*1000000)
	switch format {
	case "date":
		return t.Format("2006-01-02")
	case "time":
		return t.Format("15:04:05")
	case "datetime":
		return t.Format("2006-01-02 15:04")
	case "short":
		return t.Format("Jan 02")
	default:
		return t.Format(format)
	}
}

// FormatCurrency formats a value as currency
func FormatCurrency(value float64, currency string) string {
	switch currency {
	case "USD", "usd", "$":
		return fmt.Sprintf("$%.2f", value)
	case "EUR", "eur", "€":
		return fmt.Sprintf("€%.2f", value)
	case "GBP", "gbp", "£":
		return fmt.Sprintf("£%.2f", value)
	default:
		return fmt.Sprintf("%.2f %s", value, currency)
	}
}

// FormatNumber formats a number with thousands separators
func FormatNumber(value float64, decimals int) string {
	format := fmt.Sprintf("%%.%df", decimals)
	return fmt.Sprintf(format, value)
}
