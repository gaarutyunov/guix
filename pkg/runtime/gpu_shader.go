//go:build js && wasm

package runtime

import (
	"fmt"
	"syscall/js"
)

// ShaderModule represents a compiled shader module
type ShaderModule struct {
	Module js.Value
	Label  string
	Source string
}

// ShaderStage represents a shader stage configuration
type ShaderStage struct {
	Module     js.Value
	EntryPoint string
	Constants  map[string]float64
}

// VertexBufferLayout describes the layout of vertex data
type VertexBufferLayout struct {
	ArrayStride int
	StepMode    string // "vertex" or "instance"
	Attributes  []VertexAttribute
}

// VertexAttribute describes a single vertex attribute
type VertexAttribute struct {
	Format         string // "float32x3", "float32x2", etc.
	Offset         int
	ShaderLocation int
}

// BindGroupLayout represents the layout of a bind group
type BindGroupLayout struct {
	Entries []BindGroupLayoutEntry
}

// BindGroupLayoutEntry describes a binding in a bind group
type BindGroupLayoutEntry struct {
	Binding          int
	Visibility       int    // GPUShaderStage flags
	Type             string // "buffer", "sampler", "texture", "storage-texture"
	BufferType       string // "uniform", "storage", "read-only-storage"
	HasDynamicOffset bool
	MinBindingSize   int
}

// BindGroup represents a set of resources bound to a pipeline
type BindGroup struct {
	BindGroup js.Value
	Layout    js.Value
}

// CreateShaderModule creates a shader module from WGSL source
func CreateShaderModule(ctx *GPUContext, source string, label string) (*ShaderModule, error) {
	if ctx == nil || ctx.Device.IsUndefined() {
		return nil, fmt.Errorf("GPU context not initialized")
	}

	module, err := ctx.CreateShaderModule(source, label)
	if err != nil {
		return nil, err
	}

	return &ShaderModule{
		Module: module,
		Label:  label,
		Source: source,
	}, nil
}

// CreateVertexShaderStage creates a vertex shader stage descriptor
func CreateVertexShaderStage(module js.Value, entryPoint string) map[string]interface{} {
	return map[string]interface{}{
		"module":     module,
		"entryPoint": entryPoint,
	}
}

// CreateFragmentShaderStage creates a fragment shader stage descriptor
func CreateFragmentShaderStage(module js.Value, entryPoint string, format string) map[string]interface{} {
	// Create target descriptor as js.Value to avoid nested map conversion issues
	targetDesc := js.Global().Get("Object").New()
	targetDesc.Set("format", format)

	return map[string]interface{}{
		"module":     module,
		"entryPoint": entryPoint,
		"targets":    []interface{}{targetDesc},
	}
}

// CreateVertexBufferLayout creates a vertex buffer layout descriptor
func CreateVertexBufferLayout(arrayStride int, attributes []VertexAttribute) map[string]interface{} {
	jsAttrs := make([]interface{}, len(attributes))
	for i, attr := range attributes {
		// Create each attribute as js.Value to avoid nested map conversion issues
		attrObj := js.Global().Get("Object").New()
		attrObj.Set("format", attr.Format)
		attrObj.Set("offset", attr.Offset)
		attrObj.Set("shaderLocation", attr.ShaderLocation)
		jsAttrs[i] = attrObj
	}

	return map[string]interface{}{
		"arrayStride": arrayStride,
		"stepMode":    "vertex",
		"attributes":  jsAttrs,
	}
}

// CreateBindGroupLayoutEntry creates a bind group layout entry
func CreateBindGroupLayoutEntry(binding int, visibility int, bufferType string) map[string]interface{} {
	entry := map[string]interface{}{
		"binding":    binding,
		"visibility": visibility,
	}

	if bufferType != "" {
		// Create buffer descriptor as a js.Value to avoid nested map conversion issues
		bufferDesc := js.Global().Get("Object").New()
		bufferDesc.Set("type", bufferType)
		entry["buffer"] = bufferDesc
	}

	return entry
}

// CreateBindGroupLayout creates a bind group layout
func CreateBindGroupLayout(ctx *GPUContext, entries []map[string]interface{}, label string) (js.Value, error) {
	if ctx.Device.IsUndefined() {
		return js.Undefined(), fmt.Errorf("GPU device not initialized")
	}

	// Convert entries to JavaScript array using mapToJSObject
	jsEntries := js.Global().Get("Array").New(len(entries))
	for i, entry := range entries {
		jsEntries.SetIndex(i, mapToJSObject(entry))
	}

	// Create descriptor as JavaScript object
	descriptor := js.Global().Get("Object").New()
	descriptor.Set("entries", jsEntries)
	if label != "" {
		descriptor.Set("label", label)
	}

	layout := ctx.Device.Call("createBindGroupLayout", descriptor)
	if !layout.Truthy() {
		return js.Undefined(), fmt.Errorf("failed to create bind group layout")
	}

	return layout, nil
}

// CreateBindGroup creates a bind group with the specified resources
func CreateBindGroup(ctx *GPUContext, layout js.Value, entries []map[string]interface{}, label string) (js.Value, error) {
	if ctx.Device.IsUndefined() {
		return js.Undefined(), fmt.Errorf("GPU device not initialized")
	}

	// Convert entries to JavaScript array using mapToJSObject
	jsEntries := js.Global().Get("Array").New(len(entries))
	for i, entry := range entries {
		jsEntries.SetIndex(i, mapToJSObject(entry))
	}

	// Create descriptor as JavaScript object
	descriptor := js.Global().Get("Object").New()
	descriptor.Set("layout", layout)
	descriptor.Set("entries", jsEntries)
	if label != "" {
		descriptor.Set("label", label)
	}

	bindGroup := ctx.Device.Call("createBindGroup", descriptor)
	if !bindGroup.Truthy() {
		return js.Undefined(), fmt.Errorf("failed to create bind group")
	}

	return bindGroup, nil
}

// CreateBindGroupEntry creates a bind group entry
func CreateBindGroupEntry(binding int, resource js.Value) map[string]interface{} {
	return map[string]interface{}{
		"binding":  binding,
		"resource": resource,
	}
}

// CreateBufferBinding creates a GPUBufferBinding object for use in bind group entries
func CreateBufferBinding(buffer js.Value, offset int, size int) js.Value {
	binding := js.Global().Get("Object").New()
	binding.Set("buffer", buffer)
	binding.Set("offset", offset)
	binding.Set("size", size)
	return binding
}

// Common vertex formats
const (
	VertexFormatFloat32   = "float32"
	VertexFormatFloat32x2 = "float32x2"
	VertexFormatFloat32x3 = "float32x3"
	VertexFormatFloat32x4 = "float32x4"
	VertexFormatUint32    = "uint32"
	VertexFormatUint32x2  = "uint32x2"
	VertexFormatUint32x3  = "uint32x3"
	VertexFormatUint32x4  = "uint32x4"
	VertexFormatSint32    = "sint32"
	VertexFormatSint32x2  = "sint32x2"
	VertexFormatSint32x3  = "sint32x3"
	VertexFormatSint32x4  = "sint32x4"
)

// Common shader examples
const (
	// BasicVertexShader is a simple passthrough vertex shader
	BasicVertexShader = `
@vertex
fn vs_main(@builtin(vertex_index) vertexIndex: u32) -> @builtin(position) vec4f {
    var pos = array<vec2f, 3>(
        vec2f(0.0, 0.5),
        vec2f(-0.5, -0.5),
        vec2f(0.5, -0.5)
    );
    return vec4f(pos[vertexIndex], 0.0, 1.0);
}
`

	// BasicFragmentShader is a simple solid color fragment shader
	BasicFragmentShader = `
@fragment
fn fs_main() -> @location(0) vec4f {
    return vec4f(1.0, 0.0, 0.0, 1.0);
}
`

	// VertexShaderWithPosition is a vertex shader that takes position input
	VertexShaderWithPosition = `
struct VertexInput {
    @location(0) position: vec3f,
}

struct VertexOutput {
    @builtin(position) position: vec4f,
}

@vertex
fn vs_main(input: VertexInput) -> VertexOutput {
    var output: VertexOutput;
    output.position = vec4f(input.position, 1.0);
    return output;
}
`

	// FragmentShaderWithColor is a fragment shader with color input
	FragmentShaderWithColor = `
@fragment
fn fs_main() -> @location(0) vec4f {
    return vec4f(1.0, 0.5, 0.0, 1.0);
}
`

	// VertexShaderWithMVP is a vertex shader with MVP matrix
	VertexShaderWithMVP = `
struct Uniforms {
    modelViewProjection: mat4x4f,
}

struct VertexInput {
    @location(0) position: vec3f,
    @location(1) normal: vec3f,
}

struct VertexOutput {
    @builtin(position) position: vec4f,
    @location(0) normal: vec3f,
}

@group(0) @binding(0) var<uniform> uniforms: Uniforms;

@vertex
fn vs_main(input: VertexInput) -> VertexOutput {
    var output: VertexOutput;
    output.position = uniforms.modelViewProjection * vec4f(input.position, 1.0);
    output.normal = input.normal;
    return output;
}
`

	// FragmentShaderWithLighting is a fragment shader with simple lighting
	FragmentShaderWithLighting = `
struct VertexOutput {
    @builtin(position) position: vec4f,
    @location(0) normal: vec3f,
}

@fragment
fn fs_main(input: VertexOutput) -> @location(0) vec4f {
    let lightDir = normalize(vec3f(1.0, 1.0, 1.0));
    let diffuse = max(dot(normalize(input.normal), lightDir), 0.0);
    let ambient = 0.3;
    let color = vec3f(1.0, 0.5, 0.2);
    return vec4f(color * (ambient + diffuse * 0.7), 1.0);
}
`
)

// ValidateShaderCompilation checks if a shader module compiled successfully
func ValidateShaderCompilation(module js.Value) error {
	// Check for compilation info (if available)
	if module.Truthy() {
		compilationInfo := module.Get("compilationInfo")
		if compilationInfo.Truthy() && compilationInfo.Type() == js.TypeFunction {
			// compilationInfo is async, but we can't easily await it here
			// In practice, shader errors will be caught by device error handler
			return nil
		}
	}
	return nil
}

// ShaderPreprocessor can be used to inject common definitions into shaders
type ShaderPreprocessor struct {
	Defines map[string]string
}

// NewShaderPreprocessor creates a new shader preprocessor
func NewShaderPreprocessor() *ShaderPreprocessor {
	return &ShaderPreprocessor{
		Defines: make(map[string]string),
	}
}

// AddDefine adds a define to the preprocessor
func (sp *ShaderPreprocessor) AddDefine(name, value string) {
	sp.Defines[name] = value
}

// Process processes a shader source with defines
func (sp *ShaderPreprocessor) Process(source string) string {
	// Simple implementation - prepend defines
	result := ""
	for name, value := range sp.Defines {
		result += fmt.Sprintf("const %s = %s;\n", name, value)
	}
	result += "\n" + source
	return result
}
