# WebGPU Chart Example

This example demonstrates Guix's WebGPU-powered charting capabilities with a Bitcoin price candlestick chart.

## Features

- **GPU-accelerated rendering**: All chart elements rendered via WebGPU for high performance
- **Candlestick charts**: OHLCV (Open-High-Low-Close-Volume) visualization
- **Declarative syntax**: Define charts using Guix's component-based approach
- **Automatic axis scaling**: Smart tick generation and formatting
- **Responsive design**: Adapts to different screen sizes

## Prerequisites

- Go 1.21 or later
- A WebGPU-enabled browser:
  - Chrome/Edge 113+
  - Firefox Nightly (with `dom.webgpu.enabled` flag)
  - Safari Technology Preview

## Building

### 1. Generate Guix Components

```bash
go generate
```

This will generate `app_gen.go` and `chart_gen.go` from the `.gx` source files.

### 2. Build WASM Binary

```bash
GOOS=js GOARCH=wasm go build -o main.wasm
```

### 3. Copy wasm_exec.js

```bash
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .
```

## Running

Start a local web server:

```bash
# Using Python
python3 -m http.server 8080

# Or using Go
go run ../../cmd/serve/main.go -port 8080
```

Then open http://localhost:8080 in your WebGPU-enabled browser.

## Chart Components

The chart is defined declaratively in `chart.gx`:

```go
chart.ChartNode(
    chart.ChartBackground(0.08, 0.09, 0.12, 1.0),
    chart.ChartPaddingProp(60, 20, 40, 80),

    // X-Axis with time scale
    chart.XAxis(
        chart.TimeScale(true),
        chart.GridLines(true),
    ),

    // Y-Axis with price formatting
    chart.YAxis(
        chart.TickFormat(func(v float64) string {
            return chart.FormatCurrency(v, "USD")
        }),
    ),

    // Candlestick series
    chart.CandlestickSeries(
        chart.DataProp(data),
        chart.UpColor(0.18, 0.80, 0.44, 1.0),   // Green
        chart.DownColor(0.91, 0.27, 0.38, 1.0), // Red
    ),
)
```

## Data Format

OHLCV data is defined as:

```go
type OHLCV struct {
    Timestamp int64   // Unix timestamp in milliseconds
    Open      float64 // Opening price
    High      float64 // Highest price
    Low       float64 // Lowest price
    Close     float64 // Closing price
    Volume    float64 // Trading volume
}
```

## Browser Compatibility

| Browser | Version | Status |
|---------|---------|--------|
| Chrome  | 113+    | ✅ Supported |
| Edge    | 113+    | ✅ Supported |
| Firefox | Nightly | ⚠️ Experimental |
| Safari  | TP      | ⚠️ Experimental |

## Troubleshooting

### WebGPU not available

If you see "WebGPU Not Supported", make sure you're using a compatible browser:

- **Chrome/Edge**: WebGPU is enabled by default in version 113+
- **Firefox**: Enable `dom.webgpu.enabled` in `about:config`
- **Safari**: Use Safari Technology Preview

### Chart not rendering

1. Check the browser console for errors
2. Verify WebGPU is available: `navigator.gpu`
3. Make sure the WASM binary loaded correctly
4. Check that `wasm_exec.js` matches your Go version

## Architecture

The chart system follows Guix's declarative component model:

- **Chart Container** (`ChartNode`): Root container with background and padding
- **Axes** (`XAxis`, `YAxis`): Configurable axes with scales and ticks
- **Series** (`CandlestickSeries`): Data visualization components
- **Shaders** (WGSL): GPU shaders for rendering chart elements

All rendering happens on the GPU via WebGPU for maximum performance.

## License

MIT License - see LICENSE file for details
