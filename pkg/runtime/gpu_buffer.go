//go:build js && wasm

package runtime

import (
	"encoding/binary"
	"fmt"
	"math"
	"syscall/js"
)

// GPUBuffer wraps a WebGPU buffer with helper methods
type GPUBuffer struct {
	Buffer js.Value
	Size   int
	Usage  int
	Label  string
}

// CreateVertexBuffer creates a vertex buffer and uploads data
func CreateVertexBuffer(ctx *GPUContext, data []float32, label string) (*GPUBuffer, error) {
	size := len(data) * 4 // 4 bytes per float32
	usage := GPUBufferUsageVertex | GPUBufferUsageCopyDst

	buffer, err := ctx.CreateBuffer(size, usage, label)
	if err != nil {
		return nil, err
	}

	// Convert float32 slice to bytes
	bytes := float32SliceToBytes(data)
	if err := ctx.WriteBuffer(buffer, 0, bytes); err != nil {
		return nil, err
	}

	return &GPUBuffer{
		Buffer: buffer,
		Size:   size,
		Usage:  usage,
		Label:  label,
	}, nil
}

// CreateIndexBuffer creates an index buffer and uploads data
func CreateIndexBuffer(ctx *GPUContext, data []uint16, label string) (*GPUBuffer, error) {
	size := len(data) * 2 // 2 bytes per uint16
	usage := GPUBufferUsageIndex | GPUBufferUsageCopyDst

	buffer, err := ctx.CreateBuffer(size, usage, label)
	if err != nil {
		return nil, err
	}

	// Convert uint16 slice to bytes
	bytes := uint16SliceToBytes(data)
	if err := ctx.WriteBuffer(buffer, 0, bytes); err != nil {
		return nil, err
	}

	return &GPUBuffer{
		Buffer: buffer,
		Size:   size,
		Usage:  usage,
		Label:  label,
	}, nil
}

// CreateUniformBuffer creates a uniform buffer with specified size
func CreateUniformBuffer(ctx *GPUContext, size int, label string) (*GPUBuffer, error) {
	// Align to 256 bytes (WebGPU uniform buffer alignment requirement)
	alignedSize := (size + 255) &^ 255

	usage := GPUBufferUsageUniform | GPUBufferUsageCopyDst

	buffer, err := ctx.CreateBuffer(alignedSize, usage, label)
	if err != nil {
		return nil, err
	}

	return &GPUBuffer{
		Buffer: buffer,
		Size:   alignedSize,
		Usage:  usage,
		Label:  label,
	}, nil
}

// CreateStorageBuffer creates a storage buffer with specified size
func CreateStorageBuffer(ctx *GPUContext, size int, label string) (*GPUBuffer, error) {
	usage := GPUBufferUsageStorage | GPUBufferUsageCopyDst

	buffer, err := ctx.CreateBuffer(size, usage, label)
	if err != nil {
		return nil, err
	}

	return &GPUBuffer{
		Buffer: buffer,
		Size:   size,
		Usage:  usage,
		Label:  label,
	}, nil
}

// Write writes data to the buffer
func (b *GPUBuffer) Write(ctx *GPUContext, offset int, data []byte) error {
	if offset+len(data) > b.Size {
		return fmt.Errorf("write would exceed buffer size")
	}
	return ctx.WriteBuffer(b.Buffer, offset, data)
}

// WriteFloat32 writes float32 data to the buffer
func (b *GPUBuffer) WriteFloat32(ctx *GPUContext, offset int, data []float32) error {
	bytes := float32SliceToBytes(data)
	return b.Write(ctx, offset, bytes)
}

// WriteUint16 writes uint16 data to the buffer
func (b *GPUBuffer) WriteUint16(ctx *GPUContext, offset int, data []uint16) error {
	bytes := uint16SliceToBytes(data)
	return b.Write(ctx, offset, bytes)
}

// WriteUint32 writes uint32 data to the buffer
func (b *GPUBuffer) WriteUint32(ctx *GPUContext, offset int, data []uint32) error {
	bytes := uint32SliceToBytes(data)
	return b.Write(ctx, offset, bytes)
}

// Destroy destroys the buffer
func (b *GPUBuffer) Destroy() {
	if b.Buffer.Truthy() {
		b.Buffer.Call("destroy")
	}
}

// Utility functions for converting Go slices to bytes

func float32SliceToBytes(data []float32) []byte {
	bytes := make([]byte, len(data)*4)
	for i, v := range data {
		binary.LittleEndian.PutUint32(bytes[i*4:], math.Float32bits(v))
	}
	return bytes
}

func uint16SliceToBytes(data []uint16) []byte {
	bytes := make([]byte, len(data)*2)
	for i, v := range data {
		binary.LittleEndian.PutUint16(bytes[i*2:], v)
	}
	return bytes
}

func uint32SliceToBytes(data []uint32) []byte {
	bytes := make([]byte, len(data)*4)
	for i, v := range data {
		binary.LittleEndian.PutUint32(bytes[i*4:], v)
	}
	return bytes
}

// Geometry helper structures

// BoxGeometry creates vertex data for a box
type BoxGeometry struct {
	Width  float32
	Height float32
	Depth  float32
}

// NewBoxGeometry creates a new box geometry
func NewBoxGeometry(width, height, depth float32) *BoxGeometry {
	return &BoxGeometry{
		Width:  width,
		Height: height,
		Depth:  depth,
	}
}

// GetVertices returns vertex data for the box (position + normal)
func (bg *BoxGeometry) GetVertices() []float32 {
	w := bg.Width / 2
	h := bg.Height / 2
	d := bg.Depth / 2

	// Each face has 4 vertices, each vertex has 6 floats (3 position + 3 normal)
	// Total: 6 faces * 4 vertices * 6 floats = 144 floats
	vertices := []float32{
		// Front face (z+)
		-w, -h, d, 0, 0, 1,
		w, -h, d, 0, 0, 1,
		w, h, d, 0, 0, 1,
		-w, h, d, 0, 0, 1,

		// Back face (z-)
		w, -h, -d, 0, 0, -1,
		-w, -h, -d, 0, 0, -1,
		-w, h, -d, 0, 0, -1,
		w, h, -d, 0, 0, -1,

		// Top face (y+)
		-w, h, d, 0, 1, 0,
		w, h, d, 0, 1, 0,
		w, h, -d, 0, 1, 0,
		-w, h, -d, 0, 1, 0,

		// Bottom face (y-)
		-w, -h, -d, 0, -1, 0,
		w, -h, -d, 0, -1, 0,
		w, -h, d, 0, -1, 0,
		-w, -h, d, 0, -1, 0,

		// Right face (x+)
		w, -h, d, 1, 0, 0,
		w, -h, -d, 1, 0, 0,
		w, h, -d, 1, 0, 0,
		w, h, d, 1, 0, 0,

		// Left face (x-)
		-w, -h, -d, -1, 0, 0,
		-w, -h, d, -1, 0, 0,
		-w, h, d, -1, 0, 0,
		-w, h, -d, -1, 0, 0,
	}

	return vertices
}

// GetIndices returns index data for the box
func (bg *BoxGeometry) GetIndices() []uint16 {
	// Each face has 2 triangles (6 indices)
	// Total: 6 faces * 6 indices = 36 indices
	indices := []uint16{
		// Front
		0, 1, 2, 0, 2, 3,
		// Back
		4, 5, 6, 4, 6, 7,
		// Top
		8, 9, 10, 8, 10, 11,
		// Bottom
		12, 13, 14, 12, 14, 15,
		// Right
		16, 17, 18, 16, 18, 19,
		// Left
		20, 21, 22, 20, 22, 23,
	}

	return indices
}

// PlaneGeometry creates vertex data for a plane
type PlaneGeometry struct {
	Width  float32
	Height float32
}

// NewPlaneGeometry creates a new plane geometry
func NewPlaneGeometry(width, height float32) *PlaneGeometry {
	return &PlaneGeometry{
		Width:  width,
		Height: height,
	}
}

// GetVertices returns vertex data for the plane (position + normal + uv)
func (pg *PlaneGeometry) GetVertices() []float32 {
	w := pg.Width / 2
	h := pg.Height / 2

	// 4 vertices, each with 8 floats (3 position + 3 normal + 2 uv)
	vertices := []float32{
		-w, -h, 0, 0, 0, 1, 0, 0, // bottom-left
		w, -h, 0, 0, 0, 1, 1, 0, // bottom-right
		w, h, 0, 0, 0, 1, 1, 1, // top-right
		-w, h, 0, 0, 0, 1, 0, 1, // top-left
	}

	return vertices
}

// GetIndices returns index data for the plane
func (pg *PlaneGeometry) GetIndices() []uint16 {
	return []uint16{
		0, 1, 2,
		0, 2, 3,
	}
}

// SphereGeometry creates vertex data for a sphere
type SphereGeometry struct {
	Radius         float32
	WidthSegments  int
	HeightSegments int
}

// NewSphereGeometry creates a new sphere geometry
func NewSphereGeometry(radius float32, widthSegments, heightSegments int) *SphereGeometry {
	return &SphereGeometry{
		Radius:         radius,
		WidthSegments:  widthSegments,
		HeightSegments: heightSegments,
	}
}

// GetVertices returns vertex data for the sphere (position + normal)
func (sg *SphereGeometry) GetVertices() []float32 {
	vertices := make([]float32, 0)

	for y := 0; y <= sg.HeightSegments; y++ {
		v := float32(y) / float32(sg.HeightSegments)
		phi := v * float32(math.Pi)

		for x := 0; x <= sg.WidthSegments; x++ {
			u := float32(x) / float32(sg.WidthSegments)
			theta := u * 2 * float32(math.Pi)

			// Position
			px := -sg.Radius * float32(math.Cos(float64(theta))) * float32(math.Sin(float64(phi)))
			py := sg.Radius * float32(math.Cos(float64(phi)))
			pz := sg.Radius * float32(math.Sin(float64(theta))) * float32(math.Sin(float64(phi)))

			// Normal (normalized position for a sphere centered at origin)
			nx := px / sg.Radius
			ny := py / sg.Radius
			nz := pz / sg.Radius

			vertices = append(vertices, px, py, pz, nx, ny, nz)
		}
	}

	return vertices
}

// GetIndices returns index data for the sphere
func (sg *SphereGeometry) GetIndices() []uint16 {
	indices := make([]uint16, 0)

	for y := 0; y < sg.HeightSegments; y++ {
		for x := 0; x < sg.WidthSegments; x++ {
			a := uint16(y*(sg.WidthSegments+1) + x)
			b := uint16(y*(sg.WidthSegments+1) + x + 1)
			c := uint16((y+1)*(sg.WidthSegments+1) + x + 1)
			d := uint16((y+1)*(sg.WidthSegments+1) + x)

			indices = append(indices, a, b, d)
			indices = append(indices, b, c, d)
		}
	}

	return indices
}
