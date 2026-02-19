package utils

import (
	"math"
)

// PerlinNoise generates Perlin noise for terrain generation
type PerlinNoise struct {
	seed        int64
	permutation []int
}

// NewPerlinNoise creates a new Perlin noise generator with seed
func NewPerlinNoise(seed int64) *PerlinNoise {
	pn := &PerlinNoise{
		seed:        seed,
		permutation: make([]int, 512),
	}

	// Initialize permutation table with seed
	pn.initPermutation()
	return pn
}

// initPermutation initializes the permutation table deterministically from seed
func (pn *PerlinNoise) initPermutation() {
	// Standard Perlin permutation table
	p := []int{
		151, 160, 137, 91, 90, 15, 131, 13, 201, 95, 96, 53, 194, 233, 7, 225,
		140, 36, 103, 30, 69, 142, 8, 99, 37, 240, 21, 10, 23, 190, 6, 148,
		247, 120, 234, 75, 0, 26, 197, 62, 94, 252, 219, 203, 117, 35, 11, 32,
		57, 177, 33, 88, 237, 149, 56, 87, 174, 20, 125, 136, 171, 168, 68, 175,
		74, 165, 71, 134, 139, 48, 27, 166, 77, 146, 158, 231, 83, 111, 229, 122,
		60, 211, 133, 230, 220, 105, 92, 41, 55, 46, 245, 40, 244, 102, 143, 54,
		65, 25, 63, 161, 1, 216, 80, 73, 209, 76, 132, 187, 208, 89, 18, 169,
		200, 196, 135, 130, 116, 188, 159, 86, 164, 100, 109, 198, 173, 186, 3, 64,
		52, 217, 226, 250, 124, 123, 5, 202, 38, 147, 118, 126, 255, 82, 85, 212,
		207, 206, 59, 227, 47, 16, 58, 17, 182, 189, 28, 42, 223, 183, 170, 213,
		119, 248, 152, 2, 44, 154, 163, 70, 221, 153, 101, 155, 167, 43, 172, 9,
		129, 22, 39, 253, 19, 98, 108, 110, 79, 113, 224, 232, 178, 185, 112, 104,
		218, 246, 97, 228, 251, 34, 242, 193, 238, 210, 144, 12, 191, 179, 162, 241,
		81, 51, 145, 235, 249, 14, 239, 107, 49, 192, 214, 31, 181, 199, 106, 157,
		184, 84, 204, 176, 115, 121, 50, 45, 127, 4, 150, 254, 138, 236, 205, 93,
		222, 114, 67, 29, 24, 72, 243, 141, 128, 195, 78, 66, 215, 61, 156, 180,
	}

	// Shuffle based on seed (simple linear congruential generator)
	rng := pn.seed
	for i := len(p) - 1; i > 0; i-- {
		rng = (rng*1103515245 + 12345) & 0x7fffffff
		j := int(rng) % (i + 1)
		p[i], p[j] = p[j], p[i]
	}

	// Double the permutation table
	for i := 0; i < 256; i++ {
		pn.permutation[i] = p[i]
		pn.permutation[i+256] = p[i]
	}
}

// Noise2D generates 2D Perlin noise value at coordinates
func (pn *PerlinNoise) Noise2D(x, y float64) float64 {
	// Find unit square that contains point
	xi := int(math.Floor(x)) & 255
	yi := int(math.Floor(y)) & 255

	// Find relative x,y of point in square
	xf := x - math.Floor(x)
	yf := y - math.Floor(y)

	// Compute fade curves for each of x,y
	u := fade(xf)
	v := fade(yf)

	// Hash coordinates of square corners
	aa := pn.permutation[pn.permutation[xi]+yi]
	ab := pn.permutation[pn.permutation[xi]+yi+1]
	ba := pn.permutation[pn.permutation[xi+1]+yi]
	bb := pn.permutation[pn.permutation[xi+1]+yi+1]

	// Add blended results from 4 corners of square
	x1 := lerp(u, grad2d(aa, xf, yf), grad2d(ba, xf-1, yf))
	x2 := lerp(u, grad2d(ab, xf, yf-1), grad2d(bb, xf-1, yf-1))

	return lerp(v, x1, x2)
}

// FractalNoise generates fractal noise by combining multiple octaves
func (pn *PerlinNoise) FractalNoise(x, y float64, octaves int, persistence, scale float64) float64 {
	var value float64
	var amplitude float64 = 1.0
	var frequency float64 = scale

	for i := 0; i < octaves; i++ {
		value += pn.Noise2D(x*frequency, y*frequency) * amplitude
		amplitude *= persistence
		frequency *= 2.0
	}

	return value
}

// SimplexNoise provides a faster alternative to Perlin noise for procedural
// terrain and texture generation. Simplex noise produces smoother gradients
// and has better computational performance than classic Perlin noise,
// especially in higher dimensions.
//
// The generator is deterministic: given the same seed, it always produces
// the same noise values. This allows for reproducible procedural generation.
//
// Example usage:
//
//	sn := NewSimplexNoise(42)
//	value := sn.Noise2D(x, y) // Returns value in range [-1, 1]
type SimplexNoise struct {
	seed int64
	grad [][]float64
	perm []int
}

// NewSimplexNoise creates a new Simplex noise generator
func NewSimplexNoise(seed int64) *SimplexNoise {
	sn := &SimplexNoise{
		seed: seed,
		grad: [][]float64{
			{1, 1},
			{-1, 1},
			{1, -1},
			{-1, -1},
			{1, 0},
			{-1, 0},
			{1, 0},
			{-1, 0},
			{0, 1},
			{0, -1},
			{0, 1},
			{0, -1},
		},
		perm: make([]int, 512),
	}

	sn.initPermutation()
	return sn
}

// initPermutation initializes permutation table for simplex noise
func (sn *SimplexNoise) initPermutation() {
	p := make([]int, 256)
	for i := 0; i < 256; i++ {
		p[i] = i
	}

	// Shuffle based on seed
	rng := sn.seed
	for i := len(p) - 1; i > 0; i-- {
		rng = (rng*1103515245 + 12345) & 0x7fffffff
		j := int(rng) % (i + 1)
		p[i], p[j] = p[j], p[i]
	}

	// Double the permutation table
	for i := 0; i < 256; i++ {
		sn.perm[i] = p[i]
		sn.perm[i+256] = p[i]
	}
}

// Noise2D generates 2D Simplex noise
func (sn *SimplexNoise) Noise2D(x, y float64) float64 {
	F2 := 0.5 * (math.Sqrt(3.0) - 1.0)
	G2 := (3.0 - math.Sqrt(3.0)) / 6.0

	// Skew the input space to determine which simplex cell we're in
	s := (x + y) * F2
	i := int(math.Floor(x + s))
	j := int(math.Floor(y + s))

	t := float64(i+j) * G2
	X0 := float64(i) - t
	Y0 := float64(j) - t
	x0 := x - X0
	y0 := y - Y0

	// Determine which simplex we are in
	var i1, j1 int
	if x0 > y0 {
		i1, j1 = 1, 0
	} else {
		i1, j1 = 0, 1
	}

	// Calculate the contribution from the three corners
	x1 := x0 - float64(i1) + G2
	y1 := y0 - float64(j1) + G2
	x2 := x0 - 1.0 + 2.0*G2
	y2 := y0 - 1.0 + 2.0*G2

	// Work out the hashed gradient indices of the three simplex corners
	ii := i & 255
	jj := j & 255
	gi0 := sn.perm[ii+sn.perm[jj]] % 12
	gi1 := sn.perm[ii+i1+sn.perm[jj+j1]] % 12
	gi2 := sn.perm[ii+1+sn.perm[jj+1]] % 12

	// Calculate the contribution from the three corners
	var n0, n1, n2 float64

	t0 := 0.5 - x0*x0 - y0*y0
	if t0 < 0 {
		n0 = 0.0
	} else {
		t0 *= t0
		n0 = t0 * t0 * dot2d(sn.grad[gi0], x0, y0)
	}

	t1 := 0.5 - x1*x1 - y1*y1
	if t1 < 0 {
		n1 = 0.0
	} else {
		t1 *= t1
		n1 = t1 * t1 * dot2d(sn.grad[gi1], x1, y1)
	}

	t2 := 0.5 - x2*x2 - y2*y2
	if t2 < 0 {
		n2 = 0.0
	} else {
		t2 *= t2
		n2 = t2 * t2 * dot2d(sn.grad[gi2], x2, y2)
	}

	// Add contributions from each corner to get the final noise value
	return 70.0 * (n0 + n1 + n2)
}

// Exported helper functions for extending noise algorithms

// Fade implements the quintic smoothstep function 6t^5 - 15t^4 + 10t^3.
// This function produces smoother transitions than linear interpolation
// and has zero first and second derivatives at t=0 and t=1.
// Used in Perlin noise to create smooth gradients between lattice points.
//
// The input t should be in range [0, 1] for standard use.
// Returns a smoothed value also in range [0, 1].
func Fade(t float64) float64 {
	return t * t * t * (t*(t*6-15) + 10)
}

// Lerp performs linear interpolation between two values.
// Given interpolation factor t and values a, b, returns: a + t*(b-a)
//
// Parameters:
//   - t: interpolation factor, typically in range [0, 1]
//   - a: starting value (returned when t=0)
//   - b: ending value (returned when t=1)
//
// This is a fundamental building block for noise algorithms
// and can be chained for multi-dimensional interpolation.
func Lerp(t, a, b float64) float64 {
	return a + t*(b-a)
}

// Grad2D calculates the gradient contribution for 2D Perlin noise.
// The hash value selects one of four gradient directions based on its
// lowest 2 bits, creating the characteristic Perlin noise pattern.
//
// Parameters:
//   - hash: integer used to deterministically select gradient direction
//   - x, y: relative position within the unit cell
//
// Returns the dot product of the gradient vector and the position offset.
func Grad2D(hash int, x, y float64) float64 {
	h := hash & 3
	u := x
	if h >= 2 {
		u = y
	}
	v := y
	if h >= 2 {
		v = x
	}

	uSign := 1.0
	if (h & 1) != 0 {
		uSign = -1.0
	}

	vSign := 1.0
	if (h & 2) != 0 {
		vSign = -1.0
	}

	return uSign*u + vSign*v
}

// Dot2D calculates the 2D dot product between a gradient vector and position.
// This is used in Simplex noise to compute the contribution of each vertex.
//
// Parameters:
//   - g: 2-element slice representing the gradient vector [gx, gy]
//   - x, y: position coordinates
//
// Returns g[0]*x + g[1]*y
func Dot2D(g []float64, x, y float64) float64 {
	return g[0]*x + g[1]*y
}

// Internal aliases for backward compatibility within the package
func fade(t float64) float64 {
	return Fade(t)
}

func lerp(t, a, b float64) float64 {
	return Lerp(t, a, b)
}

func grad2d(hash int, x, y float64) float64 {
	return Grad2D(hash, x, y)
}

func dot2d(g []float64, x, y float64) float64 {
	return Dot2D(g, x, y)
}
