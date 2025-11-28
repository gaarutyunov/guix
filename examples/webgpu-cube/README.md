# Guix WebGPU - Rotating Cube Example

This example demonstrates Guix's WebGPU support with a rotating 3D cube that can be controlled with keyboard or mouse.

## Features

- **3D Rendering**: WebGPU-powered 3D graphics
- **Interactive Controls**: Rotate the cube with arrow keys or buttons
- **Auto-Rotation**: Toggle automatic rotation with spacebar
- **Speed Control**: Adjust rotation speed with slider
- **PBR Material**: Physically-based rendering with metalness and roughness
- **Directional Lighting**: Realistic lighting with ambient and directional lights

## Prerequisites

- **Go 1.21+** installed
- **Browser with WebGPU support**:
  - Chrome 113+ or Edge 113+
  - Safari Technology Preview (experimental)
  - Firefox Nightly (with `dom.webgpu.enabled` flag)

## Building

1. **Navigate to the example directory**:
   ```bash
   cd examples/webgpu-cube
   ```

2. **Copy the WASM exec helper**:
   ```bash
   cp $(go env GOROOT)/misc/wasm/wasm_exec.js .
   ```

3. **Build the WASM binary**:
   ```bash
   GOOS=js GOARCH=wasm go build -o main.wasm
   ```

4. **Serve the application**:
   ```bash
   # Using Python 3
   python3 -m http.server 8080

   # Or using Go
   go run ../../cmd/serve/main.go

   # Or using any other HTTP server
   ```

5. **Open in browser**:
   ```
   http://localhost:8080
   ```

## Controls

- **Arrow Keys** or **Buttons**: Rotate the cube manually
- **Spacebar** or **Play/Pause Button**: Toggle auto-rotation
- **Speed Slider**: Adjust auto-rotation speed (when enabled)

## How It Works

### Scene Graph

The example creates a scene graph with:

- **Scene**: Root node with background color
- **Mesh**: A cube with box geometry and PBR material
- **Camera**: Perspective camera positioned at (0, 2, 6)
- **Lights**:
  - Ambient light for base illumination
  - Directional light for realistic shading

### Render Loop

The application uses `requestAnimationFrame` for smooth 60 FPS rendering:

1. Update rotation based on user input or auto-rotation
2. Calculate model-view-projection (MVP) matrix
3. Update uniform buffer with new transform
4. Submit draw commands to GPU
5. Present frame to canvas

### WebGPU Pipeline

The renderer uses:

- **Vertex Shader**: Transforms vertices with MVP matrix
- **Fragment Shader**: Applies lighting calculations
- **Depth Testing**: Ensures correct triangle ordering
- **Backface Culling**: Improves performance

## Code Structure

```
main.go           - Application entry point and setup
index.html        - HTML page with styles and WASM loader
main.wasm         - Compiled WebAssembly binary (generated)
wasm_exec.js      - Go WASM runtime (copied from Go installation)
```

## Customization

### Change Cube Color

In `createScene()`:

```go
material := runtime.StandardMaterial(
    runtime.Color(0.91, 0.27, 0.38, 1.0), // Change RGBA values
    runtime.Metalness(0.3),
    runtime.Roughness(0.4),
)
```

### Adjust Camera

```go
camera := runtime.PerspectiveCamera(
    runtime.FOV(runtime.DegreesToRadians(60)), // Field of view
    runtime.Position(0, 2, 6),                 // Camera position
    runtime.LookAtPos(0, 0, 0),               // Look at target
)
```

### Modify Lighting

```go
ambientLight := runtime.AmbientLight(
    runtime.Intensity(0.4), // Ambient light strength
)

directionalLight := runtime.DirectionalLight(
    runtime.Position(5, 10, 7),  // Light direction
    runtime.Intensity(0.8),      // Light strength
)
```

## Troubleshooting

### "WebGPU is not supported"

- Ensure you're using a compatible browser (Chrome 113+, Edge 113+)
- Check that WebGPU is enabled in browser flags:
  - Chrome/Edge: `chrome://flags/#enable-unsafe-webgpu`
  - Firefox: `about:config` → `dom.webgpu.enabled` → `true`

### Black screen or no rendering

- Check browser console for errors
- Verify WASM file loaded successfully
- Ensure GPU drivers are up to date

### Performance issues

- Close other GPU-intensive applications
- Reduce canvas size in `main.go`:
  ```go
  config := runtime.GPUCanvasConfig{
      Width:  400,  // Smaller width
      Height: 300,  // Smaller height
  }
  ```

## Learn More

- [WebGPU Specification](https://www.w3.org/TR/webgpu/)
- [WebGPU Samples](https://webgpu.github.io/webgpu-samples/)
- [Guix Documentation](../../README.md)

## License

This example is part of the Guix project and follows the same license.
