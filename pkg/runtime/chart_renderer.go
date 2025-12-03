//go:build js && wasm

package runtime

import (
	_ "embed"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"syscall/js"
)

//go:embed chart/shaders/candlestick.wgsl
var candlestickShader string

//go:embed chart/shaders/line.wgsl
var lineShader string

// ohlcvData represents extracted OHLCV data
type ohlcvData struct {
	Timestamp int64
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
}

// ChartRenderer manages rendering of charts
type ChartRenderer struct {
	Canvas              *GPUCanvas
	Chart               *GPUNode
	CandlestickPipeline *RenderPipeline
	LinePipeline        *RenderPipeline
	LineFillPipeline    *RenderPipeline
	UniformBuffer       *GPUBuffer
	CandleDataBuffer    *GPUBuffer
	LineDataBuffer      *GPUBuffer
	BindGroup           js.Value
	LineBindGroup       js.Value
	CandlestickModule   js.Value
	LineModule          js.Value
	AxisSeries          []*GPUNode
	CandlestickSeries   []*GPUNode
	LineSeries          []*GPUNode
	BackgroundColor     Vec4
	Padding             interface{}
	DataXRange          [2]float64
	DataYRange          [2]float64
	CandleWidth         float32
	initialized         bool
}

// NewChartRenderer creates a new chart renderer
func NewChartRenderer(canvas *GPUCanvas, chart *GPUNode) (*ChartRenderer, error) {
	log("[ChartRenderer] Creating chart renderer")

	if canvas == nil {
		logError("[ChartRenderer] Canvas is nil")
		return nil, fmt.Errorf("canvas is nil")
	}
	if chart == nil {
		logError("[ChartRenderer] Chart is nil")
		return nil, fmt.Errorf("chart is nil")
	}

	renderer := &ChartRenderer{
		Canvas:            canvas,
		Chart:             chart,
		AxisSeries:        make([]*GPUNode, 0),
		CandlestickSeries: make([]*GPUNode, 0),
		LineSeries:        make([]*GPUNode, 0),
		BackgroundColor:   NewVec4(0.08, 0.09, 0.12, 1.0),
		Padding:           map[string]float32{"top": 60, "right": 20, "bottom": 40, "left": 80},
		DataXRange:        [2]float64{0, 1},
		DataYRange:        [2]float64{0, 1},
		CandleWidth:       10.0,
	}

	// Extract chart properties
	if bg, ok := chart.Properties["background"].(Vec4); ok {
		renderer.BackgroundColor = bg
	}
	if pad, ok := chart.Properties["padding"]; ok {
		renderer.Padding = pad
	}

	// Build chart scene graph
	if err := renderer.buildChart(chart); err != nil {
		logError(fmt.Sprintf("[ChartRenderer] Failed to build chart: %v", err))
		return nil, err
	}

	log("[ChartRenderer] Chart renderer created successfully")
	return renderer, nil
}

// buildChart traverses chart node tree and collects series
func (cr *ChartRenderer) buildChart(node *GPUNode) error {
	if node == nil {
		log("[ChartRenderer] buildChart called with nil node")
		return nil
	}

	log(fmt.Sprintf("[ChartRenderer] buildChart processing node type: %v", node.Type))

	// Collect series by type
	switch node.Type {
	case ChartAxisNodeType:
		log("[ChartRenderer] Found axis node")
		cr.AxisSeries = append(cr.AxisSeries, node)
	case ChartSeriesNodeType:
		seriesType, _ := node.Properties["seriesType"].(string)
		log(fmt.Sprintf("[ChartRenderer] Found series node, seriesType: %s", seriesType))
		switch seriesType {
		case "candlestick":
			log("[ChartRenderer] Adding candlestick series")
			cr.CandlestickSeries = append(cr.CandlestickSeries, node)
		case "line", "area":
			log(fmt.Sprintf("[ChartRenderer] Adding %s series", seriesType))
			cr.LineSeries = append(cr.LineSeries, node)
		default:
			log(fmt.Sprintf("[ChartRenderer] Unknown series type: %s", seriesType))
		}
	default:
		log(fmt.Sprintf("[ChartRenderer] Skipping node type: %v", node.Type))
	}

	// Recursively process children
	log(fmt.Sprintf("[ChartRenderer] Processing %d children", len(node.Children)))
	for i, child := range node.Children {
		log(fmt.Sprintf("[ChartRenderer] Processing child %d/%d", i+1, len(node.Children)))
		if err := cr.buildChart(child); err != nil {
			return err
		}
	}

	log(fmt.Sprintf("[ChartRenderer] buildChart complete - Axis: %d, Candlestick: %d, Line: %d",
		len(cr.AxisSeries), len(cr.CandlestickSeries), len(cr.LineSeries)))

	return nil
}

// initialize sets up GPU resources
func (cr *ChartRenderer) initialize() error {
	if cr.initialized {
		return nil
	}

	log("[ChartRenderer] Initializing GPU resources")

	ctx := cr.Canvas.GPUContext

	// Create shader modules
	candleModule, err := CreateShaderModule(ctx, candlestickShader, "candlestick-shader")
	if err != nil {
		return fmt.Errorf("failed to create candlestick shader: %w", err)
	}
	cr.CandlestickModule = candleModule.Module

	lineModule, err := CreateShaderModule(ctx, lineShader, "line-shader")
	if err != nil {
		return fmt.Errorf("failed to create line shader: %w", err)
	}
	cr.LineModule = lineModule.Module

	// Create pipelines
	if err := cr.createPipelines(); err != nil {
		return fmt.Errorf("failed to create pipelines: %w", err)
	}

	// Create uniform buffer
	uniformBuffer, err := CreateUniformBuffer(ctx, 256, "chart-uniforms")
	if err != nil {
		return fmt.Errorf("failed to create uniform buffer: %w", err)
	}
	cr.UniformBuffer = uniformBuffer

	cr.initialized = true
	log("[ChartRenderer] GPU resources initialized")
	return nil
}

// createPipelines creates render pipelines
func (cr *ChartRenderer) createPipelines() error {
	ctx := cr.Canvas.GPUContext

	// Candlestick pipeline
	candlePipeline, err := CreateRenderPipeline(ctx, PipelineConfig{
		Label:              "candlestick-pipeline",
		VertexShader:       cr.CandlestickModule,
		FragmentShader:     cr.CandlestickModule,
		VertexEntryPoint:   "vs_main",
		FragmentEntryPoint: "fs_main",
		ColorFormat:        cr.Canvas.Format,
		PrimitiveTopology:  PrimitiveTopologyTriangleList,
		CullMode:           CullModeNone,
	})
	if err != nil {
		return fmt.Errorf("failed to create candlestick pipeline: %w", err)
	}
	cr.CandlestickPipeline = candlePipeline

	// Line pipeline
	linePipeline, err := CreateRenderPipeline(ctx, PipelineConfig{
		Label:              "line-pipeline",
		VertexShader:       cr.LineModule,
		FragmentShader:     cr.LineModule,
		VertexEntryPoint:   "vs_line",
		FragmentEntryPoint: "fs_main",
		ColorFormat:        cr.Canvas.Format,
		PrimitiveTopology:  PrimitiveTopologyTriangleList,
		CullMode:           CullModeNone,
	})
	if err != nil {
		return fmt.Errorf("failed to create line pipeline: %w", err)
	}
	cr.LinePipeline = linePipeline

	// Line fill pipeline
	lineFillPipeline, err := CreatePipelineWithBlending(ctx, PipelineConfig{
		Label:              "line-fill-pipeline",
		VertexShader:       cr.LineModule,
		FragmentShader:     cr.LineModule,
		VertexEntryPoint:   "vs_fill",
		FragmentEntryPoint: "fs_main",
		ColorFormat:        cr.Canvas.Format,
		PrimitiveTopology:  PrimitiveTopologyTriangleList,
		CullMode:           CullModeNone,
	})
	if err != nil {
		return fmt.Errorf("failed to create line fill pipeline: %w", err)
	}
	cr.LineFillPipeline = lineFillPipeline

	return nil
}

// Render renders the chart
func (cr *ChartRenderer) Render() {
	log("[ChartRenderer] Render() called")

	if !cr.initialized {
		log("[ChartRenderer] Not initialized, initializing now...")
		if err := cr.initialize(); err != nil {
			logError(fmt.Sprintf("[ChartRenderer] Failed to initialize: %v", err))
			return
		}
		log("[ChartRenderer] Initialization complete")
	}

	log(fmt.Sprintf("[ChartRenderer] Rendering - Candlestick series: %d, Line series: %d", len(cr.CandlestickSeries), len(cr.LineSeries)))

	// Get canvas texture
	log("[ChartRenderer] Getting current texture")
	textureView := cr.Canvas.Context.Call("getCurrentTexture").Call("createView")
	if !textureView.Truthy() {
		logError("[ChartRenderer] Failed to get texture view")
		return
	}
	log("[ChartRenderer] Texture view obtained")

	// Create command encoder
	log("[ChartRenderer] Creating command encoder")
	encoder := cr.Canvas.GPUContext.Device.Call("createCommandEncoder", map[string]interface{}{
		"label": "chart-command-encoder",
	})

	// Clear background
	log(fmt.Sprintf("[ChartRenderer] Clearing background with color (%.2f, %.2f, %.2f, %.2f)", cr.BackgroundColor.X, cr.BackgroundColor.Y, cr.BackgroundColor.Z, cr.BackgroundColor.W))
	clearValue := js.Global().Get("Object").New()
	clearValue.Set("r", cr.BackgroundColor.X)
	clearValue.Set("g", cr.BackgroundColor.Y)
	clearValue.Set("b", cr.BackgroundColor.Z)
	clearValue.Set("a", cr.BackgroundColor.W)

	colorAttachment := js.Global().Get("Object").New()
	colorAttachment.Set("view", textureView)
	colorAttachment.Set("clearValue", clearValue)
	colorAttachment.Set("loadOp", "clear")
	colorAttachment.Set("storeOp", "store")

	colorAttachments := js.Global().Get("Array").New(1)
	colorAttachments.SetIndex(0, colorAttachment)

	renderPassDesc := js.Global().Get("Object").New()
	renderPassDesc.Set("label", "chart-render-pass")
	renderPassDesc.Set("colorAttachments", colorAttachments)

	log("[ChartRenderer] Beginning render pass")
	pass := encoder.Call("beginRenderPass", renderPassDesc)

	// Render candlestick series
	log(fmt.Sprintf("[ChartRenderer] Rendering %d candlestick series", len(cr.CandlestickSeries)))
	for i, series := range cr.CandlestickSeries {
		log(fmt.Sprintf("[ChartRenderer] Rendering candlestick series %d/%d", i+1, len(cr.CandlestickSeries)))
		cr.renderCandlestickSeries(pass, series)
	}

	// Render line series
	log(fmt.Sprintf("[ChartRenderer] Rendering %d line series", len(cr.LineSeries)))
	for i, series := range cr.LineSeries {
		log(fmt.Sprintf("[ChartRenderer] Rendering line series %d/%d", i+1, len(cr.LineSeries)))
		cr.renderLineSeries(pass, series)
	}

	log("[ChartRenderer] Ending render pass")
	pass.Call("end")

	// Submit command buffer
	log("[ChartRenderer] Submitting command buffer")
	commandBuffer := encoder.Call("finish")
	commandBuffers := js.Global().Get("Array").New(1)
	commandBuffers.SetIndex(0, commandBuffer)
	cr.Canvas.GPUContext.Device.Get("queue").Call("submit", commandBuffers)
	log("[ChartRenderer] Command buffer submitted successfully")
}

// renderCandlestickSeries renders a candlestick series
func (cr *ChartRenderer) renderCandlestickSeries(pass js.Value, series *GPUNode) {
	log("[ChartRenderer] renderCandlestickSeries() called")

	// Extract data from series properties
	data, ok := series.Properties["data"]
	if !ok {
		logError("[ChartRenderer] No 'data' property found in series")
		return
	}
	log(fmt.Sprintf("[ChartRenderer] Data property found, type: %T", data))

	// Extract OHLCV data using reflection
	log("[ChartRenderer] Extracting OHLCV data...")
	candles := cr.extractOHLCVData(data)
	if len(candles) == 0 {
		logError("[ChartRenderer] No candles extracted from data")
		return
	}
	log(fmt.Sprintf("[ChartRenderer] Extracted %d candles", len(candles)))

	// Log first candle for debugging
	if len(candles) > 0 {
		c := candles[0]
		log(fmt.Sprintf("[ChartRenderer] First candle - Timestamp: %d, O: %.2f, H: %.2f, L: %.2f, C: %.2f, V: %.2f",
			c.Timestamp, c.Open, c.High, c.Low, c.Close, c.Volume))
	}

	// Calculate data ranges
	log("[ChartRenderer] Calculating data ranges...")
	cr.calculateDataRanges(candles)
	log(fmt.Sprintf("[ChartRenderer] Data ranges - X: [%.2f, %.2f], Y: [%.2f, %.2f]",
		cr.DataXRange[0], cr.DataXRange[1], cr.DataYRange[0], cr.DataYRange[1]))

	// Create data buffer
	log("[ChartRenderer] Creating candle data buffer...")
	dataBuffer := cr.createCandleDataBuffer(candles)
	if dataBuffer == nil {
		logError("[ChartRenderer] Failed to create candle data buffer")
		return
	}
	log("[ChartRenderer] Candle data buffer created successfully")

	// Extract colors
	upColor := NewVec4(0.18, 0.80, 0.44, 1.0)
	downColor := NewVec4(0.91, 0.27, 0.38, 1.0)
	wickColor := NewVec4(0.6, 0.6, 0.65, 1.0)

	if c, ok := series.Properties["upColor"].(Vec4); ok {
		upColor = c
	}
	if c, ok := series.Properties["downColor"].(Vec4); ok {
		downColor = c
	}
	if c, ok := series.Properties["wickColor"].(Vec4); ok {
		wickColor = c
	}

	// Calculate candle width in DATA COORDINATES (not pixels!)
	// The shader expects candleWidth in the same units as the timestamp
	dataXRange := cr.DataXRange[1] - cr.DataXRange[0]
	candleWidth := float32(dataXRange/float64(len(candles))) * 0.8

	padding := cr.getPadding()
	chartWidth := float32(cr.Canvas.Width) - padding["left"] - padding["right"]
	log(fmt.Sprintf("[ChartRenderer] Canvas: %dx%d, Chart width: %.2f, Candle width in data coords: %.2f",
		cr.Canvas.Width, cr.Canvas.Height, chartWidth, candleWidth))

	// Create uniforms
	log("[ChartRenderer] Creating uniforms...")
	uniformData := cr.createCandleUniforms(upColor, downColor, wickColor, candleWidth)
	if err := cr.Canvas.GPUContext.WriteBuffer(cr.UniformBuffer.Buffer, 0, uniformData); err != nil {
		logError(fmt.Sprintf("[ChartRenderer] Failed to write uniform data: %v", err))
		return
	}
	log("[ChartRenderer] Uniforms written to buffer")

	// Create bind group
	log("[ChartRenderer] Creating bind group...")
	bindGroup := cr.createCandleBindGroup(dataBuffer)
	log("[ChartRenderer] Bind group created")

	// Draw
	log(fmt.Sprintf("[ChartRenderer] Issuing draw call - 6 vertices, %d instances", len(candles)*2))
	pass.Call("setPipeline", cr.CandlestickPipeline.Pipeline)
	pass.Call("setBindGroup", 0, bindGroup)
	pass.Call("draw", 6, len(candles)*2, 0, 0) // 6 vertices per quad, 2 instances per candle (body + wick)
	log("[ChartRenderer] Draw call completed")
}

// renderLineSeries renders a line series
func (cr *ChartRenderer) renderLineSeries(pass js.Value, series *GPUNode) {
	log("[ChartRenderer] renderLineSeries() called")

	// Extract data from series properties
	data, ok := series.Properties["data"]
	if !ok {
		logError("[ChartRenderer] No 'data' property found in line series")
		return
	}
	log(fmt.Sprintf("[ChartRenderer] Line data property found, type: %T", data))

	// Try to extract point data
	var points []interface{}
	switch d := data.(type) {
	case []interface{}:
		points = d
	default:
		logError(fmt.Sprintf("[ChartRenderer] Invalid line data type: %T", data))
		return
	}

	if len(points) < 2 {
		logError(fmt.Sprintf("[ChartRenderer] Not enough points for line series: %d (need at least 2)", len(points)))
		return
	}
	log(fmt.Sprintf("[ChartRenderer] Line series has %d points", len(points)))

	// Log first point for debugging
	if len(points) > 0 {
		if pointMap, ok := points[0].(map[string]interface{}); ok {
			x, _ := pointMap["X"].(float64)
			y, _ := pointMap["Y"].(float64)
			log(fmt.Sprintf("[ChartRenderer] First point - X: %.2f, Y: %.2f", x, y))
		}
	}

	// Calculate data ranges
	log("[ChartRenderer] Calculating line data ranges...")
	cr.calculateLineDataRanges(points)
	log(fmt.Sprintf("[ChartRenderer] Line data ranges - X: [%.2f, %.2f], Y: [%.2f, %.2f]",
		cr.DataXRange[0], cr.DataXRange[1], cr.DataYRange[0], cr.DataYRange[1]))

	// Create data buffer
	log("[ChartRenderer] Creating line data buffer...")
	dataBuffer := cr.createLineDataBuffer(points)
	if dataBuffer == nil {
		logError("[ChartRenderer] Failed to create line data buffer")
		return
	}
	log("[ChartRenderer] Line data buffer created successfully")

	// Extract line properties
	strokeColor := NewVec4(0.18, 0.80, 0.44, 1.0)
	strokeWidth := float32(2.0)
	fill := false
	fillColor := NewVec4(0.18, 0.80, 0.44, 0.3)

	if c, ok := series.Properties["strokeColor"].(Vec4); ok {
		strokeColor = c
	}
	if w, ok := series.Properties["strokeWidth"].(float32); ok {
		strokeWidth = w
	}
	if f, ok := series.Properties["fill"].(bool); ok {
		fill = f
	}
	if c, ok := series.Properties["fillColor"].(Vec4); ok {
		fillColor = c
	}

	log(fmt.Sprintf("[ChartRenderer] Line properties - Stroke width: %.2f, Fill: %v", strokeWidth, fill))

	// Create uniforms
	log("[ChartRenderer] Creating line uniforms...")
	uniformData := cr.createLineUniforms(strokeColor, strokeWidth, fill, fillColor)
	if err := cr.Canvas.GPUContext.WriteBuffer(cr.UniformBuffer.Buffer, 0, uniformData); err != nil {
		logError(fmt.Sprintf("[ChartRenderer] Failed to write line uniform data: %v", err))
		return
	}
	log("[ChartRenderer] Line uniforms written to buffer")

	// Create bind group
	log("[ChartRenderer] Creating line bind group...")
	bindGroup := cr.createLineBindGroup(dataBuffer)
	log("[ChartRenderer] Line bind group created")

	// Draw fill first if enabled
	if fill {
		log(fmt.Sprintf("[ChartRenderer] Drawing fill - 6 vertices, %d instances", len(points)-1))
		pass.Call("setPipeline", cr.LineFillPipeline.Pipeline)
		pass.Call("setBindGroup", 0, bindGroup)
		pass.Call("draw", 6, len(points)-1, 0, 0)
		log("[ChartRenderer] Fill draw completed")
	}

	// Draw line
	log(fmt.Sprintf("[ChartRenderer] Drawing line - 6 vertices, %d instances", len(points)-1))
	pass.Call("setPipeline", cr.LinePipeline.Pipeline)
	pass.Call("setBindGroup", 0, bindGroup)
	pass.Call("draw", 6, len(points)-1, 0, 0)
	log("[ChartRenderer] Line draw completed")
}

// Helper functions

// extractOHLCVData uses reflection to extract OHLCV data from any slice type
func (cr *ChartRenderer) extractOHLCVData(data interface{}) []ohlcvData {
	log(fmt.Sprintf("[ChartRenderer] extractOHLCVData called with type: %T", data))

	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Slice {
		logError(fmt.Sprintf("[ChartRenderer] Data is not a slice: %T (kind: %v)", data, v.Kind()))
		return nil
	}

	log(fmt.Sprintf("[ChartRenderer] Data slice length: %d", v.Len()))

	if v.Len() == 0 {
		logError("[ChartRenderer] Data slice is empty")
		return nil
	}

	// Check first element type
	firstItem := v.Index(0)
	log(fmt.Sprintf("[ChartRenderer] First item type: %v, kind: %v", firstItem.Type(), firstItem.Kind()))

	result := make([]ohlcvData, v.Len())
	successCount := 0

	for i := 0; i < v.Len(); i++ {
		item := v.Index(i)
		if item.Kind() == reflect.Struct {
			// Extract fields by name
			timestamp := getInt64Field(item, "Timestamp")
			open := getFloat64Field(item, "Open")
			high := getFloat64Field(item, "High")
			low := getFloat64Field(item, "Low")
			close := getFloat64Field(item, "Close")
			volume := getFloat64Field(item, "Volume")

			result[i] = ohlcvData{
				Timestamp: timestamp,
				Open:      open,
				High:      high,
				Low:       low,
				Close:     close,
				Volume:    volume,
			}

			if i == 0 {
				log(fmt.Sprintf("[ChartRenderer] Successfully extracted fields from first item - T:%d O:%.2f H:%.2f L:%.2f C:%.2f V:%.2f",
					timestamp, open, high, low, close, volume))
			}
			successCount++
		} else {
			if i == 0 {
				logError(fmt.Sprintf("[ChartRenderer] Item %d is not a struct, it's a %v", i, item.Kind()))
			}
		}
	}

	log(fmt.Sprintf("[ChartRenderer] Successfully extracted %d/%d OHLCV items", successCount, v.Len()))

	return result
}

// getInt64Field extracts an int64 field from a struct
func getInt64Field(v reflect.Value, fieldName string) int64 {
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return 0
	}
	switch field.Kind() {
	case reflect.Int64:
		return field.Int()
	case reflect.Int, reflect.Int32:
		return field.Int()
	default:
		return 0
	}
}

// getFloat64Field extracts a float64 field from a struct
func getFloat64Field(v reflect.Value, fieldName string) float64 {
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return 0
	}
	switch field.Kind() {
	case reflect.Float64, reflect.Float32:
		return field.Float()
	case reflect.Int, reflect.Int32, reflect.Int64:
		return float64(field.Int())
	default:
		return 0
	}
}

func (cr *ChartRenderer) getPadding() map[string]float32 {
	padding := map[string]float32{"top": 60, "right": 20, "bottom": 40, "left": 80}

	// Try different padding types
	switch p := cr.Padding.(type) {
	case map[string]float32:
		for k, v := range p {
			padding[k] = v
		}
	}

	return padding
}

func (cr *ChartRenderer) calculateDataRanges(candles []ohlcvData) {
	if len(candles) == 0 {
		return
	}

	minX, maxX := math.MaxFloat64, -math.MaxFloat64
	minY, maxY := math.MaxFloat64, -math.MaxFloat64

	for _, c := range candles {
		timestamp := float64(c.Timestamp)
		high := c.High
		low := c.Low

		if timestamp < minX {
			minX = timestamp
		}
		if timestamp > maxX {
			maxX = timestamp
		}
		if low < minY {
			minY = low
		}
		if high > maxY {
			maxY = high
		}
	}

	// Add padding to Y range
	yPadding := (maxY - minY) * 0.05
	cr.DataXRange = [2]float64{minX, maxX}
	cr.DataYRange = [2]float64{minY - yPadding, maxY + yPadding}
}

func (cr *ChartRenderer) calculateLineDataRanges(points []interface{}) {
	if len(points) == 0 {
		return
	}

	minX, maxX := math.MaxFloat64, -math.MaxFloat64
	minY, maxY := math.MaxFloat64, -math.MaxFloat64

	for _, p := range points {
		pointMap, ok := p.(map[string]interface{})
		if !ok {
			continue
		}

		x, _ := pointMap["X"].(float64)
		y, _ := pointMap["Y"].(float64)

		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
	}

	cr.DataXRange = [2]float64{minX, maxX}
	cr.DataYRange = [2]float64{minY, maxY}
}

func (cr *ChartRenderer) createCandleDataBuffer(candles []ohlcvData) *GPUBuffer {
	// Each candle: timestamp(f32), open(f32), high(f32), low(f32), close(f32), volume(f32) = 24 bytes
	bufferSize := len(candles) * 24
	data := make([]byte, bufferSize)

	for i, c := range candles {
		offset := i * 24
		timestamp := float64(c.Timestamp)
		open := c.Open
		high := c.High
		low := c.Low
		close := c.Close
		volume := c.Volume

		binary.LittleEndian.PutUint32(data[offset:], math.Float32bits(float32(timestamp)))
		binary.LittleEndian.PutUint32(data[offset+4:], math.Float32bits(float32(open)))
		binary.LittleEndian.PutUint32(data[offset+8:], math.Float32bits(float32(high)))
		binary.LittleEndian.PutUint32(data[offset+12:], math.Float32bits(float32(low)))
		binary.LittleEndian.PutUint32(data[offset+16:], math.Float32bits(float32(close)))
		binary.LittleEndian.PutUint32(data[offset+20:], math.Float32bits(float32(volume)))
	}

	// Create buffer and write data
	buffer, err := CreateStorageBuffer(cr.Canvas.GPUContext, bufferSize, "candle-data")
	if err != nil {
		logError(fmt.Sprintf("[ChartRenderer] Failed to create candle buffer: %v", err))
		return nil
	}

	if err := buffer.Write(cr.Canvas.GPUContext, 0, data); err != nil {
		logError(fmt.Sprintf("[ChartRenderer] Failed to write candle data: %v", err))
		return nil
	}

	return buffer
}

func (cr *ChartRenderer) createLineDataBuffer(points []interface{}) *GPUBuffer {
	// Each point: x(f32), y(f32) = 8 bytes
	bufferSize := len(points) * 8
	data := make([]byte, bufferSize)

	for i, p := range points {
		pointMap, ok := p.(map[string]interface{})
		if !ok {
			continue
		}

		offset := i * 8
		x, _ := pointMap["X"].(float64)
		y, _ := pointMap["Y"].(float64)

		binary.LittleEndian.PutUint32(data[offset:], math.Float32bits(float32(x)))
		binary.LittleEndian.PutUint32(data[offset+4:], math.Float32bits(float32(y)))
	}

	// Create buffer and write data
	buffer, err := CreateStorageBuffer(cr.Canvas.GPUContext, bufferSize, "line-data")
	if err != nil {
		logError(fmt.Sprintf("[ChartRenderer] Failed to create line buffer: %v", err))
		return nil
	}

	if err := buffer.Write(cr.Canvas.GPUContext, 0, data); err != nil {
		logError(fmt.Sprintf("[ChartRenderer] Failed to write line data: %v", err))
		return nil
	}

	return buffer
}

func (cr *ChartRenderer) createCandleUniforms(upColor, downColor, wickColor Vec4, candleWidth float32) []byte {
	padding := cr.getPadding()

	// Uniform layout matches WGSL struct
	data := make([]byte, 256)
	offset := 0

	// viewportSize: vec2<f32>
	binary.LittleEndian.PutUint32(data[offset:], math.Float32bits(float32(cr.Canvas.Width)))
	binary.LittleEndian.PutUint32(data[offset+4:], math.Float32bits(float32(cr.Canvas.Height)))
	offset += 16 // vec2 aligned to 16 bytes

	// dataRange: vec4<f32>
	binary.LittleEndian.PutUint32(data[offset:], math.Float32bits(float32(cr.DataXRange[0])))
	binary.LittleEndian.PutUint32(data[offset+4:], math.Float32bits(float32(cr.DataXRange[1])))
	binary.LittleEndian.PutUint32(data[offset+8:], math.Float32bits(float32(cr.DataYRange[0])))
	binary.LittleEndian.PutUint32(data[offset+12:], math.Float32bits(float32(cr.DataYRange[1])))
	offset += 16

	// padding: vec4<f32>
	binary.LittleEndian.PutUint32(data[offset:], math.Float32bits(padding["top"]))
	binary.LittleEndian.PutUint32(data[offset+4:], math.Float32bits(padding["right"]))
	binary.LittleEndian.PutUint32(data[offset+8:], math.Float32bits(padding["bottom"]))
	binary.LittleEndian.PutUint32(data[offset+12:], math.Float32bits(padding["left"]))
	offset += 16

	// candleWidth: f32
	binary.LittleEndian.PutUint32(data[offset:], math.Float32bits(candleWidth))
	offset += 16 // aligned to 16 bytes

	// upColor: vec4<f32>
	binary.LittleEndian.PutUint32(data[offset:], math.Float32bits(upColor.X))
	binary.LittleEndian.PutUint32(data[offset+4:], math.Float32bits(upColor.Y))
	binary.LittleEndian.PutUint32(data[offset+8:], math.Float32bits(upColor.Z))
	binary.LittleEndian.PutUint32(data[offset+12:], math.Float32bits(upColor.W))
	offset += 16

	// downColor: vec4<f32>
	binary.LittleEndian.PutUint32(data[offset:], math.Float32bits(downColor.X))
	binary.LittleEndian.PutUint32(data[offset+4:], math.Float32bits(downColor.Y))
	binary.LittleEndian.PutUint32(data[offset+8:], math.Float32bits(downColor.Z))
	binary.LittleEndian.PutUint32(data[offset+12:], math.Float32bits(downColor.W))
	offset += 16

	// wickColor: vec4<f32>
	binary.LittleEndian.PutUint32(data[offset:], math.Float32bits(wickColor.X))
	binary.LittleEndian.PutUint32(data[offset+4:], math.Float32bits(wickColor.Y))
	binary.LittleEndian.PutUint32(data[offset+8:], math.Float32bits(wickColor.Z))
	binary.LittleEndian.PutUint32(data[offset+12:], math.Float32bits(wickColor.W))

	return data
}

func (cr *ChartRenderer) createLineUniforms(strokeColor Vec4, strokeWidth float32, fill bool, fillColor Vec4) []byte {
	padding := cr.getPadding()

	// Uniform layout matches WGSL struct
	data := make([]byte, 256)
	offset := 0

	// viewportSize: vec2<f32>
	binary.LittleEndian.PutUint32(data[offset:], math.Float32bits(float32(cr.Canvas.Width)))
	binary.LittleEndian.PutUint32(data[offset+4:], math.Float32bits(float32(cr.Canvas.Height)))
	offset += 16

	// dataRange: vec4<f32>
	binary.LittleEndian.PutUint32(data[offset:], math.Float32bits(float32(cr.DataXRange[0])))
	binary.LittleEndian.PutUint32(data[offset+4:], math.Float32bits(float32(cr.DataXRange[1])))
	binary.LittleEndian.PutUint32(data[offset+8:], math.Float32bits(float32(cr.DataYRange[0])))
	binary.LittleEndian.PutUint32(data[offset+12:], math.Float32bits(float32(cr.DataYRange[1])))
	offset += 16

	// padding: vec4<f32>
	binary.LittleEndian.PutUint32(data[offset:], math.Float32bits(padding["top"]))
	binary.LittleEndian.PutUint32(data[offset+4:], math.Float32bits(padding["right"]))
	binary.LittleEndian.PutUint32(data[offset+8:], math.Float32bits(padding["bottom"]))
	binary.LittleEndian.PutUint32(data[offset+12:], math.Float32bits(padding["left"]))
	offset += 16

	// strokeWidth: f32
	binary.LittleEndian.PutUint32(data[offset:], math.Float32bits(strokeWidth))
	offset += 16

	// strokeColor: vec4<f32>
	binary.LittleEndian.PutUint32(data[offset:], math.Float32bits(strokeColor.X))
	binary.LittleEndian.PutUint32(data[offset+4:], math.Float32bits(strokeColor.Y))
	binary.LittleEndian.PutUint32(data[offset+8:], math.Float32bits(strokeColor.Z))
	binary.LittleEndian.PutUint32(data[offset+12:], math.Float32bits(strokeColor.W))
	offset += 16

	// fillEnabled: u32
	fillValue := uint32(0)
	if fill {
		fillValue = 1
	}
	binary.LittleEndian.PutUint32(data[offset:], fillValue)
	offset += 16

	// fillColor: vec4<f32>
	binary.LittleEndian.PutUint32(data[offset:], math.Float32bits(fillColor.X))
	binary.LittleEndian.PutUint32(data[offset+4:], math.Float32bits(fillColor.Y))
	binary.LittleEndian.PutUint32(data[offset+8:], math.Float32bits(fillColor.Z))
	binary.LittleEndian.PutUint32(data[offset+12:], math.Float32bits(fillColor.W))

	return data
}

func (cr *ChartRenderer) createCandleBindGroup(dataBuffer *GPUBuffer) js.Value {
	entries := js.Global().Get("Array").New(2)

	// Binding 0: Uniform buffer
	entry0 := js.Global().Get("Object").New()
	entry0.Set("binding", 0)
	bufferBinding0 := js.Global().Get("Object").New()
	bufferBinding0.Set("buffer", cr.UniformBuffer.Buffer)
	entry0.Set("resource", bufferBinding0)
	entries.SetIndex(0, entry0)

	// Binding 1: Storage buffer
	entry1 := js.Global().Get("Object").New()
	entry1.Set("binding", 1)
	bufferBinding1 := js.Global().Get("Object").New()
	bufferBinding1.Set("buffer", dataBuffer.Buffer)
	entry1.Set("resource", bufferBinding1)
	entries.SetIndex(1, entry1)

	bindGroupDesc := js.Global().Get("Object").New()
	bindGroupDesc.Set("layout", cr.CandlestickPipeline.Pipeline.Call("getBindGroupLayout", 0))
	bindGroupDesc.Set("entries", entries)

	return cr.Canvas.GPUContext.Device.Call("createBindGroup", bindGroupDesc)
}

func (cr *ChartRenderer) createLineBindGroup(dataBuffer *GPUBuffer) js.Value {
	entries := js.Global().Get("Array").New(2)

	// Binding 0: Uniform buffer
	entry0 := js.Global().Get("Object").New()
	entry0.Set("binding", 0)
	bufferBinding0 := js.Global().Get("Object").New()
	bufferBinding0.Set("buffer", cr.UniformBuffer.Buffer)
	entry0.Set("resource", bufferBinding0)
	entries.SetIndex(0, entry0)

	// Binding 1: Storage buffer
	entry1 := js.Global().Get("Object").New()
	entry1.Set("binding", 1)
	bufferBinding1 := js.Global().Get("Object").New()
	bufferBinding1.Set("buffer", dataBuffer.Buffer)
	entry1.Set("resource", bufferBinding1)
	entries.SetIndex(1, entry1)

	bindGroupDesc := js.Global().Get("Object").New()
	bindGroupDesc.Set("layout", cr.LinePipeline.Pipeline.Call("getBindGroupLayout", 0))
	bindGroupDesc.Set("entries", entries)

	return cr.Canvas.GPUContext.Device.Call("createBindGroup", bindGroupDesc)
}
