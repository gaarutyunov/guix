// Line Chart Shader
// Renders line series data with anti-aliasing

struct ChartUniforms {
    viewportSize: vec2<f32>,     // Canvas width and height in pixels
    dataRange: vec4<f32>,        // minX, maxX, minY, maxY (data coordinates)
    padding: vec4<f32>,          // top, right, bottom, left padding in pixels
    strokeWidth: f32,            // Line width in pixels
    strokeColor: vec4<f32>,      // Line color
    fillEnabled: u32,            // 1 if fill is enabled, 0 otherwise
    fillColor: vec4<f32>,        // Fill color (under the line)
}

struct Point {
    x: f32,
    y: f32,
}

struct VertexOutput {
    @builtin(position) position: vec4<f32>,
    @location(0) color: vec4<f32>,
    @location(1) @interpolate(flat) isFill: u32,
}

@group(0) @binding(0) var<uniform> uniforms: ChartUniforms;
@group(0) @binding(1) var<storage, read> points: array<Point>;

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

// Vertex shader for line segments
// Uses instanced rendering where each instance is a line segment
@vertex
fn vs_line(
    @builtin(vertex_index) vertexIndex: u32,
    @builtin(instance_index) instanceIndex: u32
) -> VertexOutput {
    var output: VertexOutput;

    // Each instance represents a line segment between two points
    let p1 = points[instanceIndex];
    let p2 = points[instanceIndex + 1u];

    // Convert to clip space
    let clip1 = dataToClip(p1.x, p1.y);
    let clip2 = dataToClip(p2.x, p2.y);

    // Calculate perpendicular direction for line width
    let dx = clip2.x - clip1.x;
    let dy = clip2.y - clip1.y;
    let len = sqrt(dx * dx + dy * dy);

    var perpX: f32;
    var perpY: f32;

    if (len > 0.0001) {
        perpX = -dy / len;
        perpY = dx / len;
    } else {
        perpX = 0.0;
        perpY = 1.0;
    }

    // Scale perpendicular by half stroke width (in clip space)
    let halfWidth = (uniforms.strokeWidth / uniforms.viewportSize.x);
    perpX *= halfWidth;
    perpY *= halfWidth * (uniforms.viewportSize.x / uniforms.viewportSize.y);

    // Generate quad vertices for the line segment
    var position: vec2<f32>;
    switch (vertexIndex) {
        case 0u: { position = vec2<f32>(clip1.x - perpX, clip1.y - perpY); }
        case 1u: { position = vec2<f32>(clip1.x + perpX, clip1.y + perpY); }
        case 2u: { position = vec2<f32>(clip2.x - perpX, clip2.y - perpY); }
        case 3u: { position = vec2<f32>(clip1.x + perpX, clip1.y + perpY); }
        case 4u: { position = vec2<f32>(clip2.x + perpX, clip2.y + perpY); }
        case 5u: { position = vec2<f32>(clip2.x - perpX, clip2.y - perpY); }
        default: { position = vec2<f32>(0.0, 0.0); }
    }

    output.position = vec4<f32>(position.x, position.y, 0.0, 1.0);
    output.color = uniforms.strokeColor;
    output.isFill = 0u;

    return output;
}

// Vertex shader for filled area under the line
@vertex
fn vs_fill(
    @builtin(vertex_index) vertexIndex: u32,
    @builtin(instance_index) instanceIndex: u32
) -> VertexOutput {
    var output: VertexOutput;

    // Each instance represents a triangle strip segment
    let p1 = points[instanceIndex];
    let p2 = points[instanceIndex + 1u];

    // Convert to clip space
    let clip1 = dataToClip(p1.x, p1.y);
    let clip2 = dataToClip(p2.x, p2.y);

    // Get the bottom of the chart (y = dataRange.z)
    let clipBottom = dataToClip(p1.x, uniforms.dataRange.z).y;

    // Generate triangle strip vertices
    var position: vec2<f32>;
    switch (vertexIndex) {
        case 0u: { position = vec2<f32>(clip1.x, clipBottom); }
        case 1u: { position = vec2<f32>(clip1.x, clip1.y); }
        case 2u: { position = vec2<f32>(clip2.x, clipBottom); }
        case 3u: { position = vec2<f32>(clip1.x, clip1.y); }
        case 4u: { position = vec2<f32>(clip2.x, clip2.y); }
        case 5u: { position = vec2<f32>(clip2.x, clipBottom); }
        default: { position = vec2<f32>(0.0, 0.0); }
    }

    output.position = vec4<f32>(position.x, position.y, 0.0, 1.0);
    output.color = uniforms.fillColor;
    output.isFill = 1u;

    return output;
}

// Fragment shader
@fragment
fn fs_main(input: VertexOutput) -> @location(0) vec4<f32> {
    return input.color;
}
