//go:build js && wasm

package chart

import "github.com/gaarutyunov/guix/pkg/runtime"

// LineSeriesConfig holds configuration for line series
type LineSeriesConfig struct {
	Data        []Point
	DataChannel chan []Point
	StrokeColor runtime.Vec4
	StrokeWidth float32
	Fill        bool
	FillColor   runtime.Vec4
}

// LineSeries creates a line series node
func LineSeries(opts ...SeriesOption) *runtime.GPUNode {
	node := &runtime.GPUNode{
		Type:       runtime.GroupNodeType,
		Tag:        SeriesNodeType,
		Properties: make(map[string]interface{}),
		Transform:  runtime.NewTransform(),
	}

	// Set defaults
	config := &LineSeriesConfig{
		Data:        []Point{},
		StrokeColor: runtime.NewVec4(0.18, 0.80, 0.44, 1.0), // Green
		StrokeWidth: 2.0,
		Fill:        false,
		FillColor:   runtime.NewVec4(0.18, 0.80, 0.44, 0.3), // Transparent green
	}

	node.Properties["seriesType"] = "line"
	node.Properties["config"] = config

	for _, opt := range opts {
		opt(node)
	}

	return node
}

// DataPropLine sets static data for the line series
func DataPropLine(data []Point) SeriesOption {
	return func(node *runtime.GPUNode) {
		if config, ok := node.Properties["config"].(*LineSeriesConfig); ok {
			config.Data = data
		}
	}
}

// DataChannelLine sets a reactive data channel for the line series
func DataChannelLine(ch chan []Point) SeriesOption {
	return func(node *runtime.GPUNode) {
		if config, ok := node.Properties["config"].(*LineSeriesConfig); ok {
			config.DataChannel = ch
		}
	}
}

// StrokeColor sets the line color
func StrokeColor(r, g, b, a float32) SeriesOption {
	return func(node *runtime.GPUNode) {
		if config, ok := node.Properties["config"].(*LineSeriesConfig); ok {
			config.StrokeColor = runtime.NewVec4(r, g, b, a)
		}
	}
}

// StrokeWidth sets the line width in pixels
func StrokeWidth(width float32) SeriesOption {
	return func(node *runtime.GPUNode) {
		if config, ok := node.Properties["config"].(*LineSeriesConfig); ok {
			config.StrokeWidth = width
		}
	}
}

// Fill enables/disables area fill under the line
func Fill(enabled bool) SeriesOption {
	return func(node *runtime.GPUNode) {
		if config, ok := node.Properties["config"].(*LineSeriesConfig); ok {
			config.Fill = enabled
		}
	}
}

// FillColor sets the fill color for area under the line
func FillColor(r, g, b, a float32) SeriesOption {
	return func(node *runtime.GPUNode) {
		if config, ok := node.Properties["config"].(*LineSeriesConfig); ok {
			config.FillColor = runtime.NewVec4(r, g, b, a)
		}
	}
}

// AreaSeries creates an area series (line with fill enabled)
func AreaSeries(opts ...SeriesOption) *runtime.GPUNode {
	// Create a line series with fill enabled by default
	node := LineSeries(opts...)
	if config, ok := node.Properties["config"].(*LineSeriesConfig); ok {
		config.Fill = true
	}
	node.Properties["seriesType"] = "area"
	return node
}
