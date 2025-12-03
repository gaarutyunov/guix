//go:build js && wasm

package chart

import "math"

// ViewportTransform handles transformation between data and screen coordinates
type ViewportTransform struct {
	// Canvas dimensions
	CanvasWidth  float32
	CanvasHeight float32

	// Chart area (after padding)
	ChartX      float32
	ChartY      float32
	ChartWidth  float32
	ChartHeight float32

	// Data ranges
	DataXMin float64
	DataXMax float64
	DataYMin float64
	DataYMax float64

	// Zoom and pan state
	ZoomX float64
	ZoomY float64
	PanX  float64
	PanY  float64
}

// NewViewportTransform creates a new viewport transform
func NewViewportTransform(canvasWidth, canvasHeight int, padding Padding) *ViewportTransform {
	return &ViewportTransform{
		CanvasWidth:  float32(canvasWidth),
		CanvasHeight: float32(canvasHeight),
		ChartX:       padding.Left,
		ChartY:       padding.Top,
		ChartWidth:   float32(canvasWidth) - padding.Left - padding.Right,
		ChartHeight:  float32(canvasHeight) - padding.Top - padding.Bottom,
		ZoomX:        1.0,
		ZoomY:        1.0,
		PanX:         0.0,
		PanY:         0.0,
	}
}

// SetDataRange sets the data coordinate range
func (vt *ViewportTransform) SetDataRange(xMin, xMax, yMin, yMax float64) {
	vt.DataXMin = xMin
	vt.DataXMax = xMax
	vt.DataYMin = yMin
	vt.DataYMax = yMax
}

// DataToScreen converts data coordinates to screen pixel coordinates
func (vt *ViewportTransform) DataToScreen(dataX, dataY float64) (float32, float32) {
	// Apply zoom and pan
	effectiveXMin := vt.DataXMin + vt.PanX
	effectiveXMax := vt.DataXMax + vt.PanX
	effectiveYMin := vt.DataYMin + vt.PanY
	effectiveYMax := vt.DataYMax + vt.PanY

	dataXRange := (effectiveXMax - effectiveXMin) / vt.ZoomX
	dataYRange := (effectiveYMax - effectiveYMin) / vt.ZoomY

	// Normalize to 0-1
	normalizedX := (dataX - effectiveXMin) / dataXRange
	normalizedY := (dataY - effectiveYMin) / dataYRange

	// Convert to screen coordinates
	screenX := vt.ChartX + float32(normalizedX)*vt.ChartWidth
	screenY := vt.ChartY + vt.ChartHeight - float32(normalizedY)*vt.ChartHeight // Flip Y

	return screenX, screenY
}

// ScreenToData converts screen pixel coordinates to data coordinates
func (vt *ViewportTransform) ScreenToData(screenX, screenY float32) (float64, float64) {
	// Normalize screen coordinates to 0-1
	normalizedX := float64(screenX-vt.ChartX) / float64(vt.ChartWidth)
	normalizedY := 1.0 - float64(screenY-vt.ChartY)/float64(vt.ChartHeight) // Flip Y

	// Apply zoom and pan
	effectiveXMin := vt.DataXMin + vt.PanX
	effectiveXMax := vt.DataXMax + vt.PanX
	effectiveYMin := vt.DataYMin + vt.PanY
	effectiveYMax := vt.DataYMax + vt.PanY

	dataXRange := (effectiveXMax - effectiveXMin) / vt.ZoomX
	dataYRange := (effectiveYMax - effectiveYMin) / vt.ZoomY

	// Convert to data coordinates
	dataX := effectiveXMin + normalizedX*dataXRange
	dataY := effectiveYMin + normalizedY*dataYRange

	return dataX, dataY
}

// IsVisible checks if a data point is visible in the current viewport
func (vt *ViewportTransform) IsVisible(dataX, dataY float64) bool {
	effectiveXMin := vt.DataXMin + vt.PanX
	effectiveXMax := vt.DataXMax + vt.PanX
	effectiveYMin := vt.DataYMin + vt.PanY
	effectiveYMax := vt.DataYMax + vt.PanY

	return dataX >= effectiveXMin && dataX <= effectiveXMax &&
		dataY >= effectiveYMin && dataY <= effectiveYMax
}

// Zoom applies zoom at a specific point
func (vt *ViewportTransform) Zoom(zoomFactor float64, centerX, centerY float32) {
	// Convert center point to data coordinates
	dataX, dataY := vt.ScreenToData(centerX, centerY)

	// Apply zoom
	vt.ZoomX *= zoomFactor
	vt.ZoomY *= zoomFactor

	// Clamp zoom levels
	vt.ZoomX = math.Max(0.1, math.Min(100.0, vt.ZoomX))
	vt.ZoomY = math.Max(0.1, math.Min(100.0, vt.ZoomY))

	// Adjust pan to keep the center point fixed
	newScreenX, newScreenY := vt.DataToScreen(dataX, dataY)
	vt.PanX += float64(centerX - newScreenX)
	vt.PanY += float64(centerY - newScreenY)
}

// Pan moves the viewport by screen pixel delta
func (vt *ViewportTransform) Pan(deltaX, deltaY float32) {
	dataXRange := (vt.DataXMax - vt.DataXMin) / vt.ZoomX
	dataYRange := (vt.DataYMax - vt.DataYMin) / vt.ZoomY

	// Convert pixel delta to data delta
	dataDeltaX := float64(deltaX) * dataXRange / float64(vt.ChartWidth)
	dataDeltaY := float64(deltaY) * dataYRange / float64(vt.ChartHeight)

	vt.PanX -= dataDeltaX
	vt.PanY += dataDeltaY // Flip Y
}

// Reset resets zoom and pan to default
func (vt *ViewportTransform) Reset() {
	vt.ZoomX = 1.0
	vt.ZoomY = 1.0
	vt.PanX = 0.0
	vt.PanY = 0.0
}

// CalculateCandleWidth calculates the optimal width for candles based on data density
func (vt *ViewportTransform) CalculateCandleWidth(dataCount int, widthRatio float32) float32 {
	if dataCount <= 0 {
		return 10.0
	}

	// Calculate pixels per data point
	pixelsPerPoint := vt.ChartWidth / float32(dataCount)

	// Apply width ratio (0.0-1.0)
	candleWidth := pixelsPerPoint * widthRatio

	// Clamp to reasonable values
	if candleWidth < 1.0 {
		candleWidth = 1.0
	} else if candleWidth > vt.ChartWidth/4.0 {
		candleWidth = vt.ChartWidth / 4.0
	}

	return candleWidth
}
