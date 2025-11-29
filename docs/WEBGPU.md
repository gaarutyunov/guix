# WebGPU Support in Guix

Guix now features first-class WebGPU support, enabling high-performance 3D graphics and compute operations in your web applications. This document provides a comprehensive guide to using WebGPU in Guix.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Getting Started](#getting-started)
- [Core Concepts](#core-concepts)
- [API Reference](#api-reference)
- [Examples](#examples)
- [Performance Tips](#performance-tips)
- [Troubleshooting](#troubleshooting)

## Overview

### What is WebGPU?

WebGPU is a modern graphics API for the web that provides:

- **Low-level GPU access** similar to Vulkan, Metal, and DirectX 12
- **High performance** 3D rendering and compute operations
- **Cross-platform** support across browsers and operating systems
- **Safety** through built-in validation and error handling

### Guix WebGPU Features

- ✅ **Declarative 3D API**: Scene graph with meshes, cameras, and lights
- ✅ **PBR Materials**: Physically-based rendering with metalness/roughness
- ✅ **Built-in Geometries**: Box, sphere, plane primitives
- ✅ **Lighting System**: Ambient, directional, and point lights
- ✅ **Camera System**: Perspective projection with look-at
- ✅ **3D Math**: Vectors, matrices, transformations
- ✅ **Shader Support**: WGSL shader compilation
- ✅ **Buffer Management**: Vertex, index, and uniform buffers
- ✅ **Pipeline Management**: Render and compute pipelines
- 🚧 **Custom Shaders**: User-defined WGSL shaders (in progress)
- 🚧 **Textures**: Image and procedural textures (in progress)
- 🚧 **Post-processing**: Effects and filters (planned)

## Architecture

### Runtime Structure

```
pkg/runtime/
├── webgpu.go          # Core WebGPU context and device management
├── gpu_canvas.go      # Canvas element with WebGPU context
├── gpu_shader.go      # Shader compilation and pipelines
├── gpu_buffer.go      # Buffer management and geometries
├── gpu_pipeline.go    # Render/compute pipeline creation
├── gpu_math.go        # 3D math (vectors, matrices, transforms)
├── gpu_vnode.go       # GPU VNode builders (Scene, Mesh, Camera)
└── gpu_renderer.go    # Scene graph renderer
```

### Data Flow

```
Scene Graph (GPUNode tree)
    ↓
SceneRenderer.buildScene()
    ↓
Create GPU Resources (buffers, pipelines)
    ↓
Render Loop (requestAnimationFrame)
    ↓
Update Transforms
    ↓
Render Pass (draw commands)
    ↓
Submit to GPU Queue
    ↓
Present to Canvas
```

## Getting Started

### Browser Requirements

WebGPU requires a modern browser:

- **Chrome 113+** or **Edge 113+** (recommended)
- **Safari Technology Preview** (experimental)
- **Firefox Nightly** with `dom.webgpu.enabled` flag

### Basic Example

```go
package main

import (
    "github.com/gaarutyunov/guix/pkg/runtime"
)

func main() {
    // Initialize WebGPU
    gpuCtx, _ := runtime.InitWebGPU()

    // Create canvas
    canvas, _ := runtime.CreateGPUCanvas(runtime.GPUCanvasConfig{
        Width:  800,
        Height: 600,
    })
    canvas.Mount("#app")

    // Create scene
    scene := runtime.Scene(
        runtime.Background(0.1, 0.1, 0.15, 1.0),
    )

    // Add a cube
    cube := runtime.Mesh(
        runtime.GeometryProp(runtime.NewBoxGeometry(2, 2, 2)),
        runtime.MaterialProp(runtime.StandardMaterial(
            runtime.Color(1, 0.5, 0.2, 1),
        )),
    )

    // Add camera
    camera := runtime.PerspectiveCamera(
        runtime.FOV(runtime.DegreesToRadians(60)),
        runtime.Position(0, 2, 6),
    )

    scene.Children = append(scene.Children, cube, camera)

    // Create renderer
    renderer, _ := runtime.NewSceneRenderer(canvas, scene)

    // Render loop
    canvas.SetRenderFunc(func(c *runtime.GPUCanvas, delta float64) {
        renderer.Render()
    })
    canvas.Start()

    select {} // Keep running
}
```

## Core Concepts

### Scene Graph

The scene graph is a hierarchical tree structure of 3D objects:

```
Scene (root)
├── Mesh (cube)
│   ├── Geometry (vertices, indices)
│   └── Material (color, properties)
├── Camera (perspective)
│   └── Transform (position, rotation)
└── Lights
    ├── AmbientLight
    └── DirectionalLight
```

### Transforms

Every 3D object has a transform with:

- **Position**: `Vec3{X, Y, Z}` - location in 3D space
- **Rotation**: `Vec3{X, Y, Z}` - Euler angles in radians
- **Scale**: `Vec3{X, Y, Z}` - scale factors per axis

```go
transform := runtime.NewTransform()
transform.Position = runtime.Vec3{X: 0, Y: 1, Z: 0}
transform.Rotation = runtime.Vec3{X: 0, Y: 3.14, Z: 0} // 180° on Y
transform.Scale = runtime.Vec3{X: 2, Y: 1, Z: 1}       // Stretch on X
```

### Materials

Materials define how surfaces appear:

```go
material := runtime.StandardMaterial(
    runtime.Color(1.0, 0.5, 0.2, 1.0), // RGBA (orange)
    runtime.Metalness(0.8),            // 0=dielectric, 1=metal
    runtime.Roughness(0.2),            // 0=smooth, 1=rough
)
```

### Geometries

Built-in primitive geometries:

```go
// Box: width, height, depth
box := runtime.NewBoxGeometry(2.0, 2.0, 2.0)

// Sphere: radius, width segments, height segments
sphere := runtime.NewSphereGeometry(1.0, 32, 16)

// Plane: width, height
plane := runtime.NewPlaneGeometry(10.0, 10.0)
```

### Cameras

Perspective camera for 3D scenes:

```go
camera := runtime.PerspectiveCamera(
    runtime.FOV(runtime.DegreesToRadians(60)), // Field of view
    runtime.Near(0.1),                         // Near clipping plane
    runtime.Far(100.0),                        // Far clipping plane
    runtime.Position(0, 2, 6),                 // Camera position
    runtime.LookAtPos(0, 0, 0),               // Look at target
)
```

### Lighting

Three types of lights:

```go
// Ambient: uniform lighting from all directions
ambient := runtime.AmbientLight(
    runtime.Color(1, 1, 1, 1),
    runtime.Intensity(0.3),
)

// Directional: parallel rays from a direction (like sun)
directional := runtime.DirectionalLight(
    runtime.Position(5, 10, 7),   // Light direction
    runtime.Intensity(0.8),
)

// Point: radiates from a point (like a bulb)
point := runtime.PointLight(
    runtime.Position(0, 5, 0),
    runtime.Intensity(1.0),
)
```

## API Reference

### GPU Context

```go
// Check WebGPU support
supported := runtime.IsWebGPUSupported()

// Initialize WebGPU (gets adapter and device)
ctx, err := runtime.InitWebGPU()

// Get or initialize global context
ctx, err := runtime.GetOrInitGPUContext()

// Get preferred canvas format
format := runtime.GetPreferredCanvasFormat() // "bgra8unorm" or "rgba8unorm"
```

### Canvas

```go
// Create canvas
config := runtime.GPUCanvasConfig{
    Width:            800,
    Height:           600,
    DevicePixelRatio: 1.0,
    AlphaMode:        "premultiplied",
    FrameLoop:        "always",
}
canvas, err := runtime.CreateGPUCanvas(config)

// Mount to DOM
canvas.Mount("#app")

// Set render function
canvas.SetRenderFunc(func(c *runtime.GPUCanvas, delta float64) {
    // Render code
})

// Control render loop
canvas.Start()  // Begin rendering
canvas.Stop()   // Stop rendering
canvas.RenderOnce() // Render single frame

// Resize
canvas.Resize(1024, 768)

// Cleanup
canvas.Unmount()
```

### Scene Nodes

```go
// Scene (root node)
scene := runtime.Scene(
    runtime.Background(r, g, b, a),
)

// Mesh
mesh := runtime.Mesh(
    runtime.GeometryProp(geometry),
    runtime.MaterialProp(material),
    runtime.Position(x, y, z),
    runtime.Rotation(rx, ry, rz),
    runtime.ScaleValue(sx, sy, sz),
)

// Camera
camera := runtime.PerspectiveCamera(
    runtime.FOV(fov),
    runtime.Near(near),
    runtime.Far(far),
    runtime.Position(x, y, z),
    runtime.LookAtPos(tx, ty, tz),
)

// Lights
ambient := runtime.AmbientLight(
    runtime.Color(r, g, b, a),
    runtime.Intensity(intensity),
)

directional := runtime.DirectionalLight(
    runtime.Position(x, y, z),
    runtime.Color(r, g, b, a),
    runtime.Intensity(intensity),
)

point := runtime.PointLight(
    runtime.Position(x, y, z),
    runtime.Color(r, g, b, a),
    runtime.Intensity(intensity),
)

// Group (container)
group := runtime.Group(
    runtime.Position(x, y, z),
    runtime.Rotation(rx, ry, rz),
)
```

### Math

```go
// Vectors
v2 := runtime.NewVec2(x, y)
v3 := runtime.NewVec3(x, y, z)
v4 := runtime.NewVec4(x, y, z, w)

// Vector operations
result := v1.Add(v2)
result := v1.Sub(v2)
result := v1.Mul(scalar)
dot := v1.Dot(v2)
cross := v1.Cross(v2)
length := v.Length()
normalized := v.Normalize()

// Matrices
identity := runtime.Identity()
perspective := runtime.Perspective(fov, aspect, near, far)
orthographic := runtime.Orthographic(left, right, bottom, top, near, far)
lookAt := runtime.LookAt(eye, target, up)
translation := runtime.Translation(x, y, z)
scale := runtime.Scale(x, y, z)
rotationX := runtime.RotationX(angle)
rotationY := runtime.RotationY(angle)
rotationZ := runtime.RotationZ(angle)
result := mat1.Multiply(mat2)

// Transforms
transform := runtime.NewTransform()
transform.Position = runtime.Vec3{X: 1, Y: 2, Z: 3}
transform.Rotation = runtime.Vec3{X: 0, Y: 0, Z: 0}
transform.Scale = runtime.Vec3{X: 1, Y: 1, Z: 1}
matrix := transform.Matrix() // Get 4x4 matrix

// Angle conversion
radians := runtime.DegreesToRadians(90)
degrees := runtime.RadiansToDegrees(3.14159)
```

### Buffers

```go
// Create vertex buffer
vertices := []float32{...}
buffer, err := runtime.CreateVertexBuffer(ctx, vertices, "my-vertices")

// Create index buffer
indices := []uint16{...}
buffer, err := runtime.CreateIndexBuffer(ctx, indices, "my-indices")

// Create uniform buffer
buffer, err := runtime.CreateUniformBuffer(ctx, 256, "my-uniforms")

// Write to buffer
buffer.Write(ctx, offset, bytes)
buffer.WriteFloat32(ctx, offset, floats)
buffer.WriteUint16(ctx, offset, uints)

// Cleanup
buffer.Destroy()
```

### Shaders

```go
// Create shader module
code := `
@vertex
fn vs_main(@location(0) position: vec3f) -> @builtin(position) vec4f {
    return vec4f(position, 1.0);
}
`
shader, err := runtime.CreateShaderModule(ctx, code, "my-shader")

// Built-in shaders
runtime.BasicVertexShader
runtime.BasicFragmentShader
runtime.VertexShaderWithPosition
runtime.VertexShaderWithMVP
runtime.FragmentShaderWithLighting
```

### Renderer

```go
// Create scene renderer
renderer, err := runtime.NewSceneRenderer(canvas, scene)

// Render frame
renderer.Render()

// Update mesh transform
transform := runtime.NewTransform()
transform.Rotation.Y += 0.01
renderer.UpdateMeshTransform(0, transform) // Update first mesh

// Cleanup
renderer.Cleanup()
```

## Examples

### Rotating Cube

See [`examples/webgpu-cube/`](../examples/webgpu-cube/) for a complete rotating cube example with controls.

### Custom Animation

```go
rotationY := float32(0.0)

canvas.SetRenderFunc(func(c *runtime.GPUCanvas, delta float64) {
    // Update rotation
    rotationY += float32(delta) * 0.001

    // Update mesh transform
    transform := runtime.NewTransform()
    transform.Rotation.Y = rotationY
    renderer.UpdateMeshTransform(0, transform)

    // Render
    renderer.Render()
})
```

### Multiple Meshes

```go
scene := runtime.Scene()

// Create multiple cubes in a grid
for x := -2; x <= 2; x++ {
    for z := -2; z <= 2; z++ {
        mesh := runtime.Mesh(
            runtime.GeometryProp(runtime.NewBoxGeometry(0.8, 0.8, 0.8)),
            runtime.MaterialProp(material),
            runtime.Position(float32(x)*1.5, 0, float32(z)*1.5),
        )
        scene.Children = append(scene.Children, mesh)
    }
}
```

## Performance Tips

### 1. Minimize Buffer Updates

```go
// BAD: Creating new buffers every frame
canvas.SetRenderFunc(func(c *runtime.GPUCanvas, delta float64) {
    buffer, _ := runtime.CreateVertexBuffer(ctx, vertices, "temp")
    // ...
})

// GOOD: Create buffers once, update uniforms only
buffer, _ := runtime.CreateVertexBuffer(ctx, vertices, "static")
canvas.SetRenderFunc(func(c *runtime.GPUCanvas, delta float64) {
    uniformBuffer.Write(ctx, 0, mvp.ToBytes())
    // ...
})
```

### 2. Use Index Buffers

```go
// Reduces vertex data by ~50% for typical meshes
indices := []uint16{0, 1, 2, 0, 2, 3} // Two triangles, 6 indices
vertices := []float32{...}             // Only 4 vertices needed
```

### 3. Batch Similar Objects

```go
// Group objects by material to minimize pipeline changes
scene.Children = append(scene.Children,
    metalMeshes...,     // All metal objects
    plasticMeshes...,   // All plastic objects
    glassMeshes...,     // All glass objects
)
```

### 4. Level of Detail (LOD)

```go
// Use simpler geometry for distant objects
distance := camera.Position.Sub(mesh.Transform.Position).Length()
if distance > 10 {
    mesh.Geometry = lowPolyGeometry
} else {
    mesh.Geometry = highPolyGeometry
}
```

### 5. Frustum Culling

```go
// Don't render objects outside camera view
if !isInFrustum(mesh, camera) {
    continue // Skip rendering
}
```

## Troubleshooting

### WebGPU Not Supported

**Problem**: Browser doesn't support WebGPU

**Solution**:
- Use Chrome 113+, Edge 113+, or Safari Technology Preview
- Enable WebGPU in browser flags:
  - Chrome: `chrome://flags/#enable-unsafe-webgpu`
  - Firefox: `about:config` → `dom.webgpu.enabled`

### Black Screen / No Rendering

**Problem**: Canvas shows but nothing renders

**Checks**:
1. Verify WebGPU initialized: `IsWebGPUSupported()`
2. Check browser console for GPU errors
3. Ensure scene has camera and mesh
4. Verify render loop started: `canvas.Start()`

```go
if err := renderer.Render(); err != nil {
    runtime.LogError(fmt.Sprintf("Render error: %v", err))
}
```

### Performance Issues

**Problem**: Low frame rate

**Solutions**:
- Reduce canvas size
- Use simpler geometries (fewer polygons)
- Minimize draw calls (batch objects)
- Profile with browser DevTools Performance tab

### Shader Compilation Errors

**Problem**: Shader fails to compile

**Solution**:
- Check WGSL syntax (use [WebGPU Shader Validator](https://shader-playground.timjones.io/))
- Ensure entry points are named correctly (`vs_main`, `fs_main`)
- Verify attribute locations match vertex buffer layout

### Memory Leaks

**Problem**: Memory usage increases over time

**Solution**:
- Call `Cleanup()` on renderer when done
- Destroy buffers: `buffer.Destroy()`
- Release unused resources

```go
defer renderer.Cleanup()
defer buffer.Destroy()
```

## Browser Compatibility

| Browser | Version | Status |
|---------|---------|--------|
| Chrome | 113+ | ✅ Full support |
| Edge | 113+ | ✅ Full support |
| Safari | Technology Preview | 🚧 Experimental |
| Firefox | Nightly (with flag) | 🚧 Experimental |
| Opera | 99+ | ✅ Full support |

## Future Roadmap

- [ ] **Textures**: Image loading and texture mapping
- [ ] **Normal Maps**: Detailed surface geometry
- [ ] **Shadow Maps**: Real-time shadows
- [ ] **Post-Processing**: Bloom, SSAO, tone mapping
- [ ] **Compute Shaders**: GPU compute for physics, particles
- [ ] **Instancing**: Efficient rendering of many objects
- [ ] **glTF Loader**: Load 3D models
- [ ] **Animation System**: Skeletal animation
- [ ] **Physics Integration**: Collision detection

## Resources

- [WebGPU Specification](https://www.w3.org/TR/webgpu/)
- [WGSL Specification](https://www.w3.org/TR/WGSL/)
- [WebGPU Samples](https://webgpu.github.io/webgpu-samples/)
- [Learn WebGPU](https://eliemichel.github.io/LearnWebGPU/)
- [WebGPU Fundamentals](https://webgpufundamentals.org/)

## License

WebGPU support is part of the Guix project and follows the same license.
