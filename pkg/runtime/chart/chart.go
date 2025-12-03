//go:build js && wasm

package chart

import "github.com/gaarutyunov/guix/pkg/runtime"

// ChartNodeType represents chart-specific node types
const (
	ChartContainerType = "chart"
	AxisNodeType       = "axis"
	SeriesNodeType     = "series"
	OverlayNodeType    = "overlay"
)

// ChartProp represents a chart property
type ChartProp struct {
	Key   string
	Value interface{}
}

// ChartOption is a function that configures a chart
type ChartOption func(*runtime.GPUNode)

// AxisOption is a function that configures an axis
type AxisOption func(*runtime.GPUNode)

// SeriesOption is a function that configures a series
type SeriesOption func(*runtime.GPUNode)

// Chart creates a chart container node
func Chart(opts ...ChartOption) *runtime.GPUNode {
	node := &runtime.GPUNode{
		Type:       runtime.GroupNodeType,
		Tag:        ChartContainerType,
		Properties: make(map[string]interface{}),
		Transform:  runtime.NewTransform(),
		Children:   make([]*runtime.GPUNode, 0),
	}

	// Set default background
	node.Properties["background"] = runtime.NewVec4(0.08, 0.09, 0.12, 1.0)
	node.Properties["padding"] = Padding{Top: 60, Right: 20, Bottom: 40, Left: 80}
	node.Properties["interactive"] = true

	for _, opt := range opts {
		opt(node)
	}

	return node
}

// Background sets the chart background color
func Background(r, g, b, a float32) ChartOption {
	return func(node *runtime.GPUNode) {
		node.Properties["background"] = runtime.NewVec4(r, g, b, a)
	}
}

// Padding sets the chart padding
func ChartPadding(top, right, bottom, left float32) ChartOption {
	return func(node *runtime.GPUNode) {
		node.Properties["padding"] = Padding{
			Top:    top,
			Right:  right,
			Bottom: bottom,
			Left:   left,
		}
	}
}

// Interactive enables/disables chart interactivity
func Interactive(enabled bool) ChartOption {
	return func(node *runtime.GPUNode) {
		node.Properties["interactive"] = enabled
	}
}

// WithChild adds a child node to the chart
func WithChild(child *runtime.GPUNode) ChartOption {
	return func(node *runtime.GPUNode) {
		if child != nil {
			node.Children = append(node.Children, child)
		}
	}
}

// ChartNode creates a chart node with children
func ChartNode(options ...interface{}) *runtime.GPUNode {
	node := &runtime.GPUNode{
		Type:       runtime.GroupNodeType,
		Tag:        ChartContainerType,
		Properties: make(map[string]interface{}),
		Transform:  runtime.NewTransform(),
		Children:   make([]*runtime.GPUNode, 0),
	}

	// Set defaults
	node.Properties["background"] = runtime.NewVec4(0.08, 0.09, 0.12, 1.0)
	node.Properties["padding"] = Padding{Top: 60, Right: 20, Bottom: 40, Left: 80}
	node.Properties["interactive"] = true

	for _, opt := range options {
		switch o := opt.(type) {
		case *runtime.GPUNode:
			node.Children = append(node.Children, o)
		case ChartProp:
			node.Properties[o.Key] = o.Value
		case runtime.GPUProp:
			node.Properties[o.Key] = o.Value
		}
	}

	return node
}

// ChartBackground creates a background property
func ChartBackground(r, g, b, a float32) ChartProp {
	return ChartProp{Key: "background", Value: runtime.NewVec4(r, g, b, a)}
}

// ChartPaddingProp creates a padding property
func ChartPaddingProp(top, right, bottom, left float32) ChartProp {
	return ChartProp{
		Key: "padding",
		Value: Padding{
			Top:    top,
			Right:  right,
			Bottom: bottom,
			Left:   left,
		},
	}
}

// ChartInteractive creates an interactive property
func ChartInteractive(enabled bool) ChartProp {
	return ChartProp{Key: "interactive", Value: enabled}
}
