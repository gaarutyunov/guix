//go:build js && wasm

package runtime

import (
	"math"
)

// Vec2 represents a 2D vector
type Vec2 struct {
	X, Y float32
}

// Vec3 represents a 3D vector
type Vec3 struct {
	X, Y, Z float32
}

// Vec4 represents a 4D vector
type Vec4 struct {
	X, Y, Z, W float32
}

// Mat4 represents a 4x4 matrix in column-major order
type Mat4 [16]float32

// NewVec2 creates a new 2D vector
func NewVec2(x, y float32) Vec2 {
	return Vec2{X: x, Y: y}
}

// NewVec3 creates a new 3D vector
func NewVec3(x, y, z float32) Vec3 {
	return Vec3{X: x, Y: y, Z: z}
}

// NewVec4 creates a new 4D vector
func NewVec4(x, y, z, w float32) Vec4 {
	return Vec4{X: x, Y: y, Z: z, W: w}
}

// Add adds two Vec3 vectors
func (v Vec3) Add(other Vec3) Vec3 {
	return Vec3{v.X + other.X, v.Y + other.Y, v.Z + other.Z}
}

// Sub subtracts two Vec3 vectors
func (v Vec3) Sub(other Vec3) Vec3 {
	return Vec3{v.X - other.X, v.Y - other.Y, v.Z - other.Z}
}

// Mul multiplies a Vec3 by a scalar
func (v Vec3) Mul(scalar float32) Vec3 {
	return Vec3{v.X * scalar, v.Y * scalar, v.Z * scalar}
}

// Dot computes the dot product of two Vec3 vectors
func (v Vec3) Dot(other Vec3) float32 {
	return v.X*other.X + v.Y*other.Y + v.Z*other.Z
}

// Cross computes the cross product of two Vec3 vectors
func (v Vec3) Cross(other Vec3) Vec3 {
	return Vec3{
		v.Y*other.Z - v.Z*other.Y,
		v.Z*other.X - v.X*other.Z,
		v.X*other.Y - v.Y*other.X,
	}
}

// Length returns the length of the Vec3
func (v Vec3) Length() float32 {
	return float32(math.Sqrt(float64(v.X*v.X + v.Y*v.Y + v.Z*v.Z)))
}

// Normalize returns a normalized version of the Vec3
func (v Vec3) Normalize() Vec3 {
	length := v.Length()
	if length == 0 {
		return v
	}
	return Vec3{v.X / length, v.Y / length, v.Z / length}
}

// Identity returns a 4x4 identity matrix
func Identity() Mat4 {
	return Mat4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

// Perspective creates a perspective projection matrix
func Perspective(fovY, aspect, near, far float32) Mat4 {
	f := float32(1.0 / math.Tan(float64(fovY)/2.0))
	nf := 1.0 / (near - far)

	return Mat4{
		f / aspect, 0, 0, 0,
		0, f, 0, 0,
		0, 0, (far + near) * nf, -1,
		0, 0, 2 * far * near * nf, 0,
	}
}

// Orthographic creates an orthographic projection matrix
func Orthographic(left, right, bottom, top, near, far float32) Mat4 {
	rl := 1.0 / (right - left)
	tb := 1.0 / (top - bottom)
	fn := 1.0 / (far - near)

	return Mat4{
		2 * rl, 0, 0, 0,
		0, 2 * tb, 0, 0,
		0, 0, -2 * fn, 0,
		-(right + left) * rl, -(top + bottom) * tb, -(far + near) * fn, 1,
	}
}

// LookAt creates a view matrix looking at a target
func LookAt(eye, target, up Vec3) Mat4 {
	zAxis := eye.Sub(target).Normalize()
	xAxis := up.Cross(zAxis).Normalize()
	yAxis := zAxis.Cross(xAxis)

	return Mat4{
		xAxis.X, yAxis.X, zAxis.X, 0,
		xAxis.Y, yAxis.Y, zAxis.Y, 0,
		xAxis.Z, yAxis.Z, zAxis.Z, 0,
		-xAxis.Dot(eye), -yAxis.Dot(eye), -zAxis.Dot(eye), 1,
	}
}

// Translation creates a translation matrix
func Translation(x, y, z float32) Mat4 {
	return Mat4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		x, y, z, 1,
	}
}

// Scale creates a scaling matrix
func Scale(x, y, z float32) Mat4 {
	return Mat4{
		x, 0, 0, 0,
		0, y, 0, 0,
		0, 0, z, 0,
		0, 0, 0, 1,
	}
}

// RotationX creates a rotation matrix around the X axis
func RotationX(angle float32) Mat4 {
	c := float32(math.Cos(float64(angle)))
	s := float32(math.Sin(float64(angle)))

	return Mat4{
		1, 0, 0, 0,
		0, c, s, 0,
		0, -s, c, 0,
		0, 0, 0, 1,
	}
}

// RotationY creates a rotation matrix around the Y axis
func RotationY(angle float32) Mat4 {
	c := float32(math.Cos(float64(angle)))
	s := float32(math.Sin(float64(angle)))

	return Mat4{
		c, 0, -s, 0,
		0, 1, 0, 0,
		s, 0, c, 0,
		0, 0, 0, 1,
	}
}

// RotationZ creates a rotation matrix around the Z axis
func RotationZ(angle float32) Mat4 {
	c := float32(math.Cos(float64(angle)))
	s := float32(math.Sin(float64(angle)))

	return Mat4{
		c, s, 0, 0,
		-s, c, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

// Multiply multiplies two Mat4 matrices
func (m Mat4) Multiply(other Mat4) Mat4 {
	var result Mat4

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			result[i*4+j] = 0
			for k := 0; k < 4; k++ {
				result[i*4+j] += m[k*4+j] * other[i*4+k]
			}
		}
	}

	return result
}

// ToBytes converts the matrix to a byte slice for GPU upload
func (m Mat4) ToBytes() []byte {
	bytes := make([]byte, 64) // 16 floats * 4 bytes
	for i, v := range m {
		u := math.Float32bits(v)
		bytes[i*4] = byte(u)
		bytes[i*4+1] = byte(u >> 8)
		bytes[i*4+2] = byte(u >> 16)
		bytes[i*4+3] = byte(u >> 24)
	}
	return bytes
}

// Transform represents a 3D transformation with position, rotation, and scale
type Transform struct {
	Position Vec3
	Rotation Vec3 // Euler angles in radians
	Scale    Vec3
}

// NewTransform creates a new transform with default values
func NewTransform() Transform {
	return Transform{
		Position: Vec3{0, 0, 0},
		Rotation: Vec3{0, 0, 0},
		Scale:    Vec3{1, 1, 1},
	}
}

// Matrix returns the transformation matrix
func (t Transform) Matrix() Mat4 {
	// Create individual transformation matrices
	translationMat := Translation(t.Position.X, t.Position.Y, t.Position.Z)
	rotationMatX := RotationX(t.Rotation.X)
	rotationMatY := RotationY(t.Rotation.Y)
	rotationMatZ := RotationZ(t.Rotation.Z)
	scaleMat := Scale(t.Scale.X, t.Scale.Y, t.Scale.Z)

	// Combine: Translation * RotationZ * RotationY * RotationX * Scale
	result := scaleMat
	result = rotationMatX.Multiply(result)
	result = rotationMatY.Multiply(result)
	result = rotationMatZ.Multiply(result)
	result = translationMat.Multiply(result)

	return result
}

// Camera represents a 3D camera
type Camera struct {
	Position Vec3
	Target   Vec3
	Up       Vec3
	FOV      float32 // Field of view in radians
	Aspect   float32 // Aspect ratio (width/height)
	Near     float32 // Near clipping plane
	Far      float32 // Far clipping plane
}

// NewPerspectiveCamera creates a new perspective camera
func NewPerspectiveCamera(fov, aspect, near, far float32) Camera {
	return Camera{
		Position: Vec3{0, 0, 5},
		Target:   Vec3{0, 0, 0},
		Up:       Vec3{0, 1, 0},
		FOV:      fov,
		Aspect:   aspect,
		Near:     near,
		Far:      far,
	}
}

// ViewMatrix returns the camera's view matrix
func (c Camera) ViewMatrix() Mat4 {
	return LookAt(c.Position, c.Target, c.Up)
}

// ProjectionMatrix returns the camera's projection matrix
func (c Camera) ProjectionMatrix() Mat4 {
	return Perspective(c.FOV, c.Aspect, c.Near, c.Far)
}

// ViewProjectionMatrix returns the combined view-projection matrix
func (c Camera) ViewProjectionMatrix() Mat4 {
	view := c.ViewMatrix()
	proj := c.ProjectionMatrix()
	return proj.Multiply(view)
}

// LookAt sets the camera to look at a target
func (c *Camera) LookAtTarget(target Vec3) {
	c.Target = target
}

// DegreesToRadians converts degrees to radians
func DegreesToRadians(degrees float32) float32 {
	return degrees * float32(math.Pi) / 180.0
}

// RadiansToDegrees converts radians to degrees
func RadiansToDegrees(radians float32) float32 {
	return radians * 180.0 / float32(math.Pi)
}
