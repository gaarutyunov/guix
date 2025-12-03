struct LineUniforms {
    viewportSize: vec2<f32>,
    dataRange: vec4<f32>,
    padding: vec4<f32>,
    strokeWidth: f32,
    strokeColor: vec4<f32>,
    fillEnabled: u32,
    fillColor: vec4<f32>,
}

struct Point {
    x: f32,
    y: f32,
}

