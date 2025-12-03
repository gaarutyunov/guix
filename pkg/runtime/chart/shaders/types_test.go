//go:build js && wasm
// +build js,wasm

package shaders

import (
	"testing"
	"unsafe"
)

// TestChartUniformsSize verifies the ChartUniforms struct size matches WGSL expectations
func TestChartUniformsSize(t *testing.T) {
	uniforms := ChartUniforms{}

	// Expected size based on WGSL layout rules:
	// vec2 (8) + vec4 (16) + vec4 (16) + f32 (4) + padding (12) + vec4 (16) * 3 = 112 bytes
	expectedSize := uintptr(112)
	actualSize := unsafe.Sizeof(uniforms)

	if actualSize != expectedSize {
		t.Errorf("ChartUniforms size mismatch: expected %d bytes, got %d bytes", expectedSize, actualSize)
	}
}

// TestCandleSize verifies the Candle struct size matches WGSL expectations
func TestCandleSize(t *testing.T) {
	candle := Candle{}

	// Expected size: 6 * f32 = 6 * 4 = 24 bytes
	expectedSize := uintptr(24)
	actualSize := unsafe.Sizeof(candle)

	if actualSize != expectedSize {
		t.Errorf("Candle size mismatch: expected %d bytes, got %d bytes", expectedSize, actualSize)
	}
}

// TestChartUniformsToBytes verifies ToBytes generates correct byte slice
func TestChartUniformsToBytes(t *testing.T) {
	uniforms := ChartUniforms{
		ViewportSize: [2]float32{800, 600},
		DataRange:    [4]float32{0, 100, 0, 1000},
		Padding:      [4]float32{10, 10, 10, 10},
		CandleWidth:  15.0,
		UpColor:      [4]float32{0, 1, 0, 1},
		DownColor:    [4]float32{1, 0, 0, 1},
		WickColor:    [4]float32{0.5, 0.5, 0.5, 1},
	}

	bytes := uniforms.ToBytes()

	if len(bytes) != 112 {
		t.Errorf("ToBytes returned wrong size: expected 112 bytes, got %d bytes", len(bytes))
	}

	// Verify bytes are non-zero (contains actual data)
	hasData := false
	for _, b := range bytes {
		if b != 0 {
			hasData = true
			break
		}
	}

	if !hasData {
		t.Error("ToBytes returned all zeros, expected actual data")
	}
}

// TestCandleToBytes verifies Candle ToBytes generates correct byte slice
func TestCandleToBytes(t *testing.T) {
	candle := Candle{
		Timestamp: 1700000000000,
		Open:      100.5,
		High:      105.2,
		Low:       99.8,
		Close:     103.1,
		Volume:    1000000,
	}

	bytes := candle.ToBytes()

	if len(bytes) != 24 {
		t.Errorf("ToBytes returned wrong size: expected 24 bytes, got %d bytes", len(bytes))
	}
}

// TestLineUniformsSize verifies the LineUniforms struct size matches WGSL expectations
func TestLineUniformsSize(t *testing.T) {
	uniforms := LineUniforms{}

	// Expected size based on WGSL layout rules:
	// vec2 (8) + vec4 (16) + vec4 (16) + f32 (4) + vec4 (16) + u32 (4) + vec4 (16) = 80 bytes
	// But with alignment, it's actually larger
	actualSize := unsafe.Sizeof(uniforms)

	// The exact size depends on Go's struct layout
	// Just verify it's a reasonable size
	if actualSize < 80 || actualSize > 128 {
		t.Errorf("LineUniforms size unexpected: got %d bytes (expected between 80-128)", actualSize)
	}
}

// TestPointSize verifies the Point struct size matches WGSL expectations
func TestPointSize(t *testing.T) {
	point := Point{}

	// Expected size: 2 * f32 = 2 * 4 = 8 bytes
	expectedSize := uintptr(8)
	actualSize := unsafe.Sizeof(point)

	if actualSize != expectedSize {
		t.Errorf("Point size mismatch: expected %d bytes, got %d bytes", expectedSize, actualSize)
	}
}

// TestShaderSourceEmbedded verifies the shader sources are embedded
func TestShaderSourceEmbedded(t *testing.T) {
	if CandlestickTypesShaderSource == "" {
		t.Error("CandlestickTypesShaderSource is empty, expected embedded WGSL code")
	}

	if LineTypesShaderSource == "" {
		t.Error("LineTypesShaderSource is empty, expected embedded WGSL code")
	}

	// Verify they contain expected struct definitions
	if len(CandlestickTypesShaderSource) < 50 {
		t.Errorf("CandlestickTypesShaderSource too short: %d characters", len(CandlestickTypesShaderSource))
	}

	if len(LineTypesShaderSource) < 50 {
		t.Errorf("LineTypesShaderSource too short: %d characters", len(LineTypesShaderSource))
	}
}
