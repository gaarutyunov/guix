// Candlestick Chart Shader
// Renders OHLCV candlestick data with GPU instancing

struct ChartUniforms {
    viewportSize: vec2<f32>,     // Canvas width and height in pixels
    dataRange: vec4<f32>,        // minX, maxX, minY, maxY (data coordinates)
    padding: vec4<f32>,          // top, right, bottom, left padding in pixels
    candleWidth: f32,            // Width of each candle in pixels
    upColor: vec4<f32>,          // Color for up candles (close >= open)
    downColor: vec4<f32>,        // Color for down candles (close < open)
    wickColor: vec4<f32>,        // Color for candle wicks
}

struct Candle {
    timestamp: f32,              // Unix timestamp in milliseconds
    open: f32,
    high: f32,
    low: f32,
    close: f32,
    volume: f32,
}

struct VertexOutput {
    @builtin(position) position: vec4<f32>,
    @location(0) color: vec4<f32>,
    @location(1) @interpolate(flat) isWick: u32,  // 1 for wick, 0 for body
}

@group(0) @binding(0) var<uniform> uniforms: ChartUniforms;
@group(0) @binding(1) var<storage, read> candles: array<Candle>;

// Transform data coordinates to clip space
fn dataToClip(dataX: f32, dataY: f32) -> vec2<f32> {
    // Calculate chart area (viewport minus padding)
    let chartWidth = uniforms.viewportSize.x - uniforms.padding.w - uniforms.padding.y;
    let chartHeight = uniforms.viewportSize.y - uniforms.padding.x - uniforms.padding.z;

    // Normalize data to 0-1 range
    let normalizedX = (dataX - uniforms.dataRange.x) / (uniforms.dataRange.y - uniforms.dataRange.x);
    let normalizedY = (dataY - uniforms.dataRange.z) / (uniforms.dataRange.w - uniforms.dataRange.z);

    // Convert to pixel coordinates (adding padding)
    let pixelX = uniforms.padding.w + normalizedX * chartWidth;
    let pixelY = uniforms.padding.x + (1.0 - normalizedY) * chartHeight;  // Flip Y axis

    // Convert to clip space (-1 to 1)
    let clipX = (pixelX / uniforms.viewportSize.x) * 2.0 - 1.0;
    let clipY = (pixelY / uniforms.viewportSize.y) * 2.0 - 1.0;

    return vec2<f32>(clipX, clipY);
}

// Vertex shader - generates geometry for candles
// Each candle is rendered as 2 instances: body (quad) and wick (line)
@vertex
fn vs_main(
    @builtin(vertex_index) vertexIndex: u32,
    @builtin(instance_index) instanceIndex: u32
) -> VertexOutput {
    var output: VertexOutput;

    // Determine if this is a wick (odd instance) or body (even instance)
    let candleIndex = instanceIndex / 2u;
    let isWick = instanceIndex % 2u;

    let candle = candles[candleIndex];
    let centerX = candle.timestamp;

    // Determine color based on candle direction
    let isUp = candle.close >= candle.open;
    let bodyColor = select(uniforms.downColor, uniforms.upColor, isUp);

    var position: vec2<f32>;

    if (isWick == 1u) {
        // Render wick as a thin line from low to high
        let wickWidth = max(1.0, uniforms.candleWidth * 0.1);

        // Generate line vertices (6 vertices for a thin rect)
        switch (vertexIndex) {
            case 0u: { position = vec2<f32>(centerX - wickWidth / 2.0, candle.low); }
            case 1u: { position = vec2<f32>(centerX + wickWidth / 2.0, candle.low); }
            case 2u: { position = vec2<f32>(centerX - wickWidth / 2.0, candle.high); }
            case 3u: { position = vec2<f32>(centerX + wickWidth / 2.0, candle.low); }
            case 4u: { position = vec2<f32>(centerX + wickWidth / 2.0, candle.high); }
            case 5u: { position = vec2<f32>(centerX - wickWidth / 2.0, candle.high); }
            default: { position = vec2<f32>(0.0, 0.0); }
        }

        output.color = uniforms.wickColor;
        output.isWick = 1u;
    } else {
        // Render body as a rect from open to close
        let bodyTop = max(candle.open, candle.close);
        let bodyBottom = min(candle.open, candle.close);
        let halfWidth = uniforms.candleWidth / 2.0;

        // Generate quad vertices
        switch (vertexIndex) {
            case 0u: { position = vec2<f32>(centerX - halfWidth, bodyBottom); }
            case 1u: { position = vec2<f32>(centerX + halfWidth, bodyBottom); }
            case 2u: { position = vec2<f32>(centerX - halfWidth, bodyTop); }
            case 3u: { position = vec2<f32>(centerX + halfWidth, bodyBottom); }
            case 4u: { position = vec2<f32>(centerX + halfWidth, bodyTop); }
            case 5u: { position = vec2<f32>(centerX - halfWidth, bodyTop); }
            default: { position = vec2<f32>(0.0, 0.0); }
        }

        output.color = bodyColor;
        output.isWick = 0u;
    }

    // Transform to clip space
    let clipPos = dataToClip(position.x, position.y);
    output.position = vec4<f32>(clipPos.x, clipPos.y, 0.0, 1.0);

    return output;
}

// Fragment shader
@fragment
fn fs_main(input: VertexOutput) -> @location(0) vec4<f32> {
    return input.color;
}
