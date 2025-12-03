//go:build js && wasm

package runtime

import (
	"testing"
)

func TestChartNode(t *testing.T) {
	// Test Chart node creation
	node := ChartNode()

	if node == nil {
		t.Fatal("ChartNode returned nil")
	}

	if node.Type != ChartNodeType {
		t.Errorf("Expected ChartNodeType, got %v", node.Type)
	}

	if node.Tag != "chart" {
		t.Errorf("Expected tag 'chart', got %s", node.Tag)
	}

	// Check default properties
	if _, ok := node.Properties["background"]; !ok {
		t.Error("Chart should have default background property")
	}

	if _, ok := node.Properties["padding"]; !ok {
		t.Error("Chart should have default padding property")
	}

	if interactive, ok := node.Properties["interactive"]; !ok || interactive != true {
		t.Error("Chart should have interactive property set to true by default")
	}
}

func TestChartNodeWithOptions(t *testing.T) {
	// Test Chart with custom background
	bg := ChartBackground(0.5, 0.5, 0.5, 1.0)
	node := ChartNode(bg)

	if bgVal, ok := node.Properties["background"]; !ok {
		t.Error("Chart should have background property")
	} else if vec, ok := bgVal.(*Vec4); !ok {
		t.Error("Background should be Vec4")
	} else {
		if vec.X != 0.5 || vec.Y != 0.5 || vec.Z != 0.5 || vec.W != 1.0 {
			t.Errorf("Background color incorrect: got %v", vec)
		}
	}
}

func TestXAxisNode(t *testing.T) {
	node := XAxis()

	if node == nil {
		t.Fatal("XAxis returned nil")
	}

	if node.Type != ChartAxisNodeType {
		t.Errorf("Expected ChartAxisNodeType, got %v", node.Type)
	}

	if node.Tag != "xaxis" {
		t.Errorf("Expected tag 'xaxis', got %s", node.Tag)
	}

	// Check default properties
	if axisType, ok := node.Properties["axisType"]; !ok || axisType != "x" {
		t.Error("XAxis should have axisType property set to 'x'")
	}

	if position, ok := node.Properties["position"]; !ok || position != "bottom" {
		t.Error("XAxis should have default position 'bottom'")
	}
}

func TestYAxisNode(t *testing.T) {
	node := YAxis()

	if node == nil {
		t.Fatal("YAxis returned nil")
	}

	if node.Type != ChartAxisNodeType {
		t.Errorf("Expected ChartAxisNodeType, got %v", node.Type)
	}

	if node.Tag != "yaxis" {
		t.Errorf("Expected tag 'yaxis', got %s", node.Tag)
	}

	// Check default properties
	if axisType, ok := node.Properties["axisType"]; !ok || axisType != "y" {
		t.Error("YAxis should have axisType property set to 'y'")
	}

	if position, ok := node.Properties["position"]; !ok || position != "right" {
		t.Error("YAxis should have default position 'right'")
	}
}

func TestCandlestickSeriesNode(t *testing.T) {
	node := CandlestickSeries()

	if node == nil {
		t.Fatal("CandlestickSeries returned nil")
	}

	if node.Type != ChartSeriesNodeType {
		t.Errorf("Expected ChartSeriesNodeType, got %v", node.Type)
	}

	if node.Tag != "candlestick" {
		t.Errorf("Expected tag 'candlestick', got %s", node.Tag)
	}

	// Check default properties
	if seriesType, ok := node.Properties["seriesType"]; !ok || seriesType != "candlestick" {
		t.Error("CandlestickSeries should have seriesType property set to 'candlestick'")
	}

	if _, ok := node.Properties["upColor"]; !ok {
		t.Error("CandlestickSeries should have default upColor")
	}

	if _, ok := node.Properties["downColor"]; !ok {
		t.Error("CandlestickSeries should have default downColor")
	}

	if _, ok := node.Properties["wickColor"]; !ok {
		t.Error("CandlestickSeries should have default wickColor")
	}
}

func TestLineSeriesNode(t *testing.T) {
	node := LineSeries()

	if node == nil {
		t.Fatal("LineSeries returned nil")
	}

	if node.Type != ChartSeriesNodeType {
		t.Errorf("Expected ChartSeriesNodeType, got %v", node.Type)
	}

	if node.Tag != "line" {
		t.Errorf("Expected tag 'line', got %s", node.Tag)
	}

	// Check default properties
	if seriesType, ok := node.Properties["seriesType"]; !ok || seriesType != "line" {
		t.Error("LineSeries should have seriesType property set to 'line'")
	}
}

func TestChartPropertyFunctions(t *testing.T) {
	// Test ChartBackground
	bg := ChartBackground(0.1, 0.2, 0.3, 1.0)
	if bg.Key != "background" {
		t.Errorf("ChartBackground key should be 'background', got %s", bg.Key)
	}

	// Test ChartPadding
	padding := ChartPadding(10, 20, 30, 40)
	if padding.Key != "padding" {
		t.Errorf("ChartPadding key should be 'padding', got %s", padding.Key)
	}
	if p, ok := padding.Value.(Padding); !ok {
		t.Error("ChartPadding value should be Padding type")
	} else {
		if p.Top != 10 || p.Right != 20 || p.Bottom != 30 || p.Left != 40 {
			t.Errorf("ChartPadding values incorrect: got %+v", p)
		}
	}

	// Test ChartInteractive
	interactive := ChartInteractive(true)
	if interactive.Key != "interactive" {
		t.Errorf("ChartInteractive key should be 'interactive', got %s", interactive.Key)
	}
	if interactive.Value != true {
		t.Errorf("ChartInteractive value should be true, got %v", interactive.Value)
	}

	// Test AxisPosition
	pos := AxisPosition("top")
	if pos.Key != "position" {
		t.Errorf("AxisPosition key should be 'position', got %s", pos.Key)
	}
	if pos.Value != "top" {
		t.Errorf("AxisPosition value should be 'top', got %v", pos.Value)
	}

	// Test TimeScale
	ts := TimeScale(true)
	if ts.Key != "timeScale" {
		t.Errorf("TimeScale key should be 'timeScale', got %s", ts.Key)
	}

	// Test GridLines
	gl := GridLines(false)
	if gl.Key != "gridLines" {
		t.Errorf("GridLines key should be 'gridLines', got %s", gl.Key)
	}

	// Test ChartData
	data := []interface{}{1, 2, 3}
	cd := ChartData(data)
	if cd.Key != "data" {
		t.Errorf("ChartData key should be 'data', got %s", cd.Key)
	}

	// Test UpColor
	uc := UpColor(0.2, 0.8, 0.4, 1.0)
	if uc.Key != "upColor" {
		t.Errorf("UpColor key should be 'upColor', got %s", uc.Key)
	}

	// Test DownColor
	dc := DownColor(0.9, 0.3, 0.4, 1.0)
	if dc.Key != "downColor" {
		t.Errorf("DownColor key should be 'downColor', got %s", dc.Key)
	}

	// Test WickColor
	wc := WickColor(0.6, 0.6, 0.6, 1.0)
	if wc.Key != "wickColor" {
		t.Errorf("WickColor key should be 'wickColor', got %s", wc.Key)
	}

	// Test BarWidth
	bw := BarWidth(0.8)
	if bw.Key != "barWidth" {
		t.Errorf("BarWidth key should be 'barWidth', got %s", bw.Key)
	}
}

func TestGPUChart(t *testing.T) {
	// Create a mock chart
	mockChart := &mockChartImpl{}

	vnode := GPUChart(mockChart)

	if vnode == nil {
		t.Fatal("GPUChart returned nil")
	}

	if vnode.Type != ElementNode {
		t.Errorf("Expected ElementNode type, got %v", vnode.Type)
	}

	if vnode.Tag != "webgpu-chart" {
		t.Errorf("Expected tag 'webgpu-chart', got %s", vnode.Tag)
	}

	if chart, ok := vnode.Properties["chart"]; !ok {
		t.Error("GPUChart VNode should have 'chart' property")
	} else if chart != mockChart {
		t.Error("Chart property should contain the original chart instance")
	}
}

// Mock Chart implementation for testing
type mockChartImpl struct{}

func (m *mockChartImpl) RenderChart() *GPUNode {
	return ChartNode()
}
