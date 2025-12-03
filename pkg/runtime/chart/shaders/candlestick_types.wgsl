struct ChartUniforms {
    viewportSize: vec2<f32>,
    dataRange: vec4<f32>,
    padding: vec4<f32>,
    candleWidth: f32,
    upColor: vec4<f32>,
    downColor: vec4<f32>,
    wickColor: vec4<f32>,
}

struct Candle {
    timestamp: f32,
    open: f32,
    high: f32,
    low: f32,
    close: f32,
    volume: f32,
}

