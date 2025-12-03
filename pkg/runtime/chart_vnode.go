//go:build js && wasm

package runtime

// Chart node types
const (
	ChartNodeType GPUNodeType = iota + 100
	ChartAxisNodeType
	ChartSeriesNodeType
	ChartOverlayNodeType
)

// Padding represents padding around a chart
type Padding struct {
	Top    float32
	Right  float32
	Bottom float32
	Left   float32
}

// Chart interface for chart components
type Chart interface {
	RenderChart() *GPUNode
}

// GPUChart wraps a Chart interface in a VNode for rendering
func GPUChart(chart Chart) *VNode {
	return &VNode{
		Type:       ElementNode,
		Tag:        "webgpu-chart",
		Properties: map[string]interface{}{"chart": chart},
	}
}

// ChartNode creates a chart container node
func ChartNode(options ...interface{}) *GPUNode {
	node := &GPUNode{
		Type:       ChartNodeType,
		Tag:        "chart",
		Properties: make(map[string]interface{}),
		Transform:  NewTransform(),
		Children:   make([]*GPUNode, 0),
	}

	// Set defaults
	node.Properties["background"] = NewVec4(0.08, 0.09, 0.12, 1.0)
	node.Properties["padding"] = Padding{Top: 60, Right: 20, Bottom: 40, Left: 80}
	node.Properties["interactive"] = true

	for _, opt := range options {
		switch o := opt.(type) {
		case *GPUNode:
			node.Children = append(node.Children, o)
		case GPUProp:
			node.Properties[o.Key] = o.Value
		}
	}

	return node
}

// XAxis creates an X-axis node
func XAxis(options ...interface{}) *GPUNode {
	node := &GPUNode{
		Type:       ChartAxisNodeType,
		Tag:        "xaxis",
		Properties: make(map[string]interface{}),
	}

	// Set defaults
	node.Properties["axisType"] = "x"
	node.Properties["position"] = "bottom"
	node.Properties["gridLines"] = true
	node.Properties["gridColor"] = NewVec4(0.2, 0.2, 0.25, 0.5)
	node.Properties["timeScale"] = false

	for _, opt := range options {
		switch o := opt.(type) {
		case GPUProp:
			node.Properties[o.Key] = o.Value
		}
	}

	return node
}

// YAxis creates a Y-axis node
func YAxis(options ...interface{}) *GPUNode {
	node := &GPUNode{
		Type:       ChartAxisNodeType,
		Tag:        "yaxis",
		Properties: make(map[string]interface{}),
	}

	// Set defaults
	node.Properties["axisType"] = "y"
	node.Properties["position"] = "right"
	node.Properties["gridLines"] = true
	node.Properties["gridColor"] = NewVec4(0.2, 0.2, 0.25, 0.5)

	for _, opt := range options {
		switch o := opt.(type) {
		case GPUProp:
			node.Properties[o.Key] = o.Value
		}
	}

	return node
}

// CandlestickSeries creates a candlestick series node
func CandlestickSeries(options ...interface{}) *GPUNode {
	node := &GPUNode{
		Type:       ChartSeriesNodeType,
		Tag:        "candlestick",
		Properties: make(map[string]interface{}),
	}

	// Set defaults
	node.Properties["seriesType"] = "candlestick"
	node.Properties["upColor"] = NewVec4(0.18, 0.80, 0.44, 1.0)   // Green
	node.Properties["downColor"] = NewVec4(0.91, 0.27, 0.38, 1.0) // Red
	node.Properties["wickColor"] = NewVec4(0.6, 0.6, 0.65, 1.0)   // Gray
	node.Properties["barWidth"] = float32(0.8)

	for _, opt := range options {
		switch o := opt.(type) {
		case GPUProp:
			node.Properties[o.Key] = o.Value
		}
	}

	return node
}

// LineSeries creates a line series node
func LineSeries(options ...interface{}) *GPUNode {
	node := &GPUNode{
		Type:       ChartSeriesNodeType,
		Tag:        "line",
		Properties: make(map[string]interface{}),
	}

	// Set defaults
	node.Properties["seriesType"] = "line"
	node.Properties["strokeColor"] = NewVec4(0.18, 0.80, 0.44, 1.0)
	node.Properties["strokeWidth"] = float32(2.0)
	node.Properties["fill"] = false
	node.Properties["fillColor"] = NewVec4(0.18, 0.80, 0.44, 0.3)

	for _, opt := range options {
		switch o := opt.(type) {
		case GPUProp:
			node.Properties[o.Key] = o.Value
		}
	}

	return node
}

// Chart property functions

// ChartBackground sets chart background color
func ChartBackground(r, g, b, a float32) GPUProp {
	return GPUProp{Key: "background", Value: NewVec4(r, g, b, a)}
}

// ChartPadding sets chart padding
func ChartPadding(top, right, bottom, left float32) GPUProp {
	return GPUProp{
		Key: "padding",
		Value: Padding{
			Top:    top,
			Right:  right,
			Bottom: bottom,
			Left:   left,
		},
	}
}

// ChartInteractive enables/disables interactivity
func ChartInteractive(enabled bool) GPUProp {
	return GPUProp{Key: "interactive", Value: enabled}
}

// AxisPosition sets axis position
func AxisPosition(pos string) GPUProp {
	return GPUProp{Key: "position", Value: pos}
}

// TimeScale enables time-based scaling
func TimeScale(enabled bool) GPUProp {
	return GPUProp{Key: "timeScale", Value: enabled}
}

// GridLines enables/disables grid lines
func GridLines(enabled bool) GPUProp {
	return GPUProp{Key: "gridLines", Value: enabled}
}

// GridColor sets grid line color
func GridColor(r, g, b, a float32) GPUProp {
	return GPUProp{Key: "gridColor", Value: NewVec4(r, g, b, a)}
}

// Label sets axis label
func Label(text string) GPUProp {
	return GPUProp{Key: "label", Value: text}
}

// ChartData sets series data
func ChartData(data interface{}) GPUProp {
	return GPUProp{Key: "data", Value: data}
}

// UpColor sets up candle color
func UpColor(r, g, b, a float32) GPUProp {
	return GPUProp{Key: "upColor", Value: NewVec4(r, g, b, a)}
}

// DownColor sets down candle color
func DownColor(r, g, b, a float32) GPUProp {
	return GPUProp{Key: "downColor", Value: NewVec4(r, g, b, a)}
}

// WickColor sets wick color
func WickColor(r, g, b, a float32) GPUProp {
	return GPUProp{Key: "wickColor", Value: NewVec4(r, g, b, a)}
}

// BarWidth sets bar width ratio
func BarWidth(ratio float32) GPUProp {
	return GPUProp{Key: "barWidth", Value: ratio}
}

// StrokeColor sets line stroke color
func StrokeColor(r, g, b, a float32) GPUProp {
	return GPUProp{Key: "strokeColor", Value: NewVec4(r, g, b, a)}
}

// StrokeWidth sets line width
func StrokeWidth(width float32) GPUProp {
	return GPUProp{Key: "strokeWidth", Value: width}
}

// FillColor sets fill color
func FillColor(r, g, b, a float32) GPUProp {
	return GPUProp{Key: "fillColor", Value: NewVec4(r, g, b, a)}
}

// FillEnabled enables/disables fill
func FillEnabled(enabled bool) GPUProp {
	return GPUProp{Key: "fill", Value: enabled}
}

// Axis position constants
const (
	AxisTop    = "top"
	AxisBottom = "bottom"
	AxisLeft   = "left"
	AxisRight  = "right"
)
