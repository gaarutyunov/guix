//go:build js && wasm

package main

import (
	"syscall/js"

	"github.com/gaarutyunov/guix/pkg/runtime/chart"
)

// ScrollManager manages viewport scrolling and data pagination for the chart
type ScrollManager struct {
	canvas        js.Value
	app           js.Value
	chartData     *ChartDataManager
	panX          float64
	panY          float64
	isDragging    bool
	lastMouseX    float64
	lastMouseY    float64
	touchStartX   float64
	touchStartY   float64
	wheelCallback js.Func
	keyCallback   js.Func
}

// ChartDataManager manages the chart data with pagination and generation
type ChartDataManager struct {
	allData      []chart.OHLCV
	visibleStart int
	visibleEnd   int
	prefetchSize int
	totalFetched int
	symbol       string
	interval     string
	isFetching   bool
	generator    *MarkovChainGenerator
}

// NewChartDataManager creates a new chart data manager
func NewChartDataManager(initialData []chart.OHLCV, symbol, interval string) *ChartDataManager {
	visibleSize := 100 // Show 100 candles at a time (as requested)
	if len(initialData) < visibleSize {
		visibleSize = len(initialData)
	}

	// Initialize Markov chain generator from last candle
	var generator *MarkovChainGenerator
	if len(initialData) > 0 {
		lastCandle := initialData[len(initialData)-1]
		generator = NewMarkovChainGenerator(lastCandle.Close, 1.0)
	} else {
		generator = NewMarkovChainGenerator(45000.0, 1.0)
	}

	return &ChartDataManager{
		allData:      initialData,
		visibleStart: 0,
		visibleEnd:   visibleSize,
		prefetchSize: 50, // Prefetch 50 candles when nearing edge
		totalFetched: len(initialData),
		symbol:       symbol,
		interval:     interval,
		isFetching:   false,
		generator:    generator,
	}
}

// GetVisibleData returns the currently visible data
func (cdm *ChartDataManager) GetVisibleData() []chart.OHLCV {
	if cdm.visibleEnd > len(cdm.allData) {
		cdm.visibleEnd = len(cdm.allData)
	}
	if cdm.visibleStart < 0 {
		cdm.visibleStart = 0
	}
	return cdm.allData[cdm.visibleStart:cdm.visibleEnd]
}

// ShiftViewport shifts the visible data window
func (cdm *ChartDataManager) ShiftViewport(delta int) bool {
	oldStart := cdm.visibleStart
	oldEnd := cdm.visibleEnd

	cdm.visibleStart += delta
	cdm.visibleEnd += delta

	// Clamp to data bounds
	if cdm.visibleStart < 0 {
		cdm.visibleStart = 0
		cdm.visibleEnd = oldEnd - oldStart
	}
	if cdm.visibleEnd > len(cdm.allData) {
		cdm.visibleEnd = len(cdm.allData)
		cdm.visibleStart = cdm.visibleEnd - (oldEnd - oldStart)
		if cdm.visibleStart < 0 {
			cdm.visibleStart = 0
		}
	}

	// Check if we need to prefetch more data (near the end, fetch older data)
	needsFetch := cdm.visibleEnd > len(cdm.allData)-cdm.prefetchSize

	// Return whether viewport actually changed
	return cdm.visibleStart != oldStart || cdm.visibleEnd != oldEnd || needsFetch
}

// FetchOrGenerateMoreData tries to fetch more data from Binance, falls back to generation
func (cdm *ChartDataManager) FetchOrGenerateMoreData(count int) {
	if cdm.isFetching {
		return
	}

	cdm.isFetching = true
	go func() {
		defer func() { cdm.isFetching = false }()

		// Get the oldest timestamp (last element in array = oldest data)
		if len(cdm.allData) == 0 {
			return
		}

		oldestTimestamp := cdm.allData[len(cdm.allData)-1].Timestamp

		log("[Data] Fetching more data before timestamp:", oldestTimestamp)

		// Try to fetch from Binance with endTime parameter
		data, err := FetchBinanceDataWithEndTime(cdm.symbol, cdm.interval, count, oldestTimestamp)

		if err != nil || len(data) == 0 {
			// Failed to fetch, generate using Markov chains instead
			log("[Data] Fetch failed, generating", count, "candles")
			data = cdm.generator.GenerateCandles(count, cdm.interval)
		} else {
			log("[Data] Fetched", len(data), "candles from Binance")
		}

		// Append to existing data
		cdm.allData = append(cdm.allData, data...)
		cdm.totalFetched = len(cdm.allData)

		log("[Data] Total candles:", cdm.totalFetched)
	}()
}

// NewScrollManager creates a new scroll manager
func NewScrollManager(canvasID string, initialData []chart.OHLCV) *ScrollManager {
	canvas := js.Global().Get("document").Call("getElementById", canvasID)
	app := js.Global().Get("document").Call("getElementById", "app")

	sm := &ScrollManager{
		canvas:     canvas,
		app:        app,
		chartData:  NewChartDataManager(initialData, "BTCUSDT", "1h"),
		panX:       0,
		panY:       0,
		isDragging: false,
	}

	sm.attachEventListeners()
	return sm
}

// attachEventListeners attaches all scroll event listeners
func (sm *ScrollManager) attachEventListeners() {
	// Mouse wheel for horizontal scrolling
	sm.wheelCallback = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		event.Call("preventDefault")

		deltaY := event.Get("deltaY").Float()
		deltaX := event.Get("deltaX").Float()

		// Use vertical wheel for horizontal scroll if no horizontal movement
		scrollDelta := deltaX
		if deltaX == 0 {
			scrollDelta = deltaY
		}

		// Shift viewport based on scroll
		// Negative delta = scroll left (show older data), positive = scroll right (show newer data)
		shift := int(scrollDelta / 20) // Adjust sensitivity
		if sm.chartData.ShiftViewport(shift) {
			sm.updateChart()
		}

		return nil
	})
	sm.canvas.Call("addEventListener", "wheel", sm.wheelCallback)

	// Mouse drag for panning
	mouseDownCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		sm.isDragging = true
		sm.lastMouseX = event.Get("clientX").Float()
		sm.lastMouseY = event.Get("clientY").Float()
		return nil
	})
	sm.canvas.Call("addEventListener", "mousedown", mouseDownCallback)

	mouseMoveCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if !sm.isDragging {
			return nil
		}

		event := args[0]
		currentX := event.Get("clientX").Float()
		deltaX := currentX - sm.lastMouseX

		// Convert mouse delta to data shift
		shift := int(-deltaX / 5) // Negative because moving mouse right scrolls left
		if sm.chartData.ShiftViewport(shift) {
			sm.updateChart()
		}

		sm.lastMouseX = currentX
		return nil
	})
	js.Global().Get("document").Call("addEventListener", "mousemove", mouseMoveCallback)

	mouseUpCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		sm.isDragging = false
		return nil
	})
	js.Global().Get("document").Call("addEventListener", "mouseup", mouseUpCallback)

	// Touch events for mobile
	touchStartCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		if event.Get("touches").Length() > 0 {
			touch := event.Get("touches").Index(0)
			sm.touchStartX = touch.Get("clientX").Float()
			sm.touchStartY = touch.Get("clientY").Float()
		}
		return nil
	})
	sm.canvas.Call("addEventListener", "touchstart", touchStartCallback)

	touchMoveCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		event.Call("preventDefault")

		if event.Get("touches").Length() > 0 {
			touch := event.Get("touches").Index(0)
			currentX := touch.Get("clientX").Float()
			deltaX := currentX - sm.touchStartX

			shift := int(-deltaX / 5)
			if sm.chartData.ShiftViewport(shift) {
				sm.updateChart()
			}

			sm.touchStartX = currentX
		}
		return nil
	})
	sm.canvas.Call("addEventListener", "touchmove", touchMoveCallback)

	// Keyboard navigation
	sm.keyCallback = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		key := event.Get("key").String()

		shift := 0
		switch key {
		case "ArrowLeft":
			shift = -10 // Scroll left (older data)
		case "ArrowRight":
			shift = 10 // Scroll right (newer data)
		case "Home":
			sm.chartData.visibleStart = 0
			sm.chartData.visibleEnd = 100
			sm.updateChart()
			return nil
		case "End":
			size := sm.chartData.visibleEnd - sm.chartData.visibleStart
			sm.chartData.visibleEnd = len(sm.chartData.allData)
			sm.chartData.visibleStart = sm.chartData.visibleEnd - size
			if sm.chartData.visibleStart < 0 {
				sm.chartData.visibleStart = 0
			}
			sm.updateChart()
			return nil
		}

		if shift != 0 {
			event.Call("preventDefault")
			if sm.chartData.ShiftViewport(shift) {
				sm.updateChart()
			}
		}

		return nil
	})
	sm.app.Call("addEventListener", "keydown", sm.keyCallback)
}

// updateChart triggers a chart rerender with updated data
func (sm *ScrollManager) updateChart() {
	log("[Scroll] Viewport updated:", sm.chartData.visibleStart, "-", sm.chartData.visibleEnd, "of", len(sm.chartData.allData))

	// Trigger the app to rerender with new visible data
	TriggerChartUpdate()

	// Check if we need to fetch/generate more data
	if sm.chartData.visibleEnd > len(sm.chartData.allData)-sm.chartData.prefetchSize {
		log("[Scroll] Near end, fetching/generating more data...")
		sm.chartData.FetchOrGenerateMoreData(200) // Fetch or generate 200 more candles
	}
}

// Cleanup releases event listeners
func (sm *ScrollManager) Cleanup() {
	if sm.wheelCallback.Truthy() {
		sm.canvas.Call("removeEventListener", "wheel", sm.wheelCallback)
		sm.wheelCallback.Release()
	}
	if sm.keyCallback.Truthy() {
		sm.app.Call("removeEventListener", "keydown", sm.keyCallback)
		sm.keyCallback.Release()
	}
}
