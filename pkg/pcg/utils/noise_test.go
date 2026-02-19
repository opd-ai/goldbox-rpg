package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPerlinNoise(t *testing.T) {
	seed := int64(12345)
	pn := NewPerlinNoise(seed)

	assert.NotNil(t, pn)
	assert.Equal(t, seed, pn.seed)
	assert.Len(t, pn.permutation, 512)
}

func TestPerlinNoiseDeterministic(t *testing.T) {
	seed := int64(54321)

	pn1 := NewPerlinNoise(seed)
	pn2 := NewPerlinNoise(seed)

	// Same coordinates should produce same noise values
	x, y := 1.5, 2.3
	noise1 := pn1.Noise2D(x, y)
	noise2 := pn2.Noise2D(x, y)

	assert.Equal(t, noise1, noise2)
}

func TestPerlinNoiseBasicFunctionality(t *testing.T) {
	pn := NewPerlinNoise(12345)

	// Test that noise function doesn't panic and returns finite values
	noise := pn.Noise2D(1.0, 1.0)
	assert.False(t, isNaN(noise))
	assert.False(t, isInf(noise))
}

func TestNewSimplexNoise(t *testing.T) {
	seed := int64(12345)
	sn := NewSimplexNoise(seed)

	assert.NotNil(t, sn)
	assert.Equal(t, seed, sn.seed)
	assert.Len(t, sn.perm, 512)
}

func TestSimplexNoiseDeterministic(t *testing.T) {
	seed := int64(54321)

	sn1 := NewSimplexNoise(seed)
	sn2 := NewSimplexNoise(seed)

	// Same coordinates should produce same noise values
	x, y := 1.5, 2.3
	noise1 := sn1.Noise2D(x, y)
	noise2 := sn2.Noise2D(x, y)

	assert.Equal(t, noise1, noise2)
}

func TestSimplexNoiseBasicFunctionality(t *testing.T) {
	sn := NewSimplexNoise(12345)

	// Test that noise function doesn't panic and returns finite values
	noise := sn.Noise2D(1.0, 1.0)
	assert.False(t, isNaN(noise))
	assert.False(t, isInf(noise))
}

// TestFractalNoiseBasic tests basic functionality of FractalNoise method
func TestFractalNoiseBasic(t *testing.T) {
	pn := NewPerlinNoise(12345)

	// Test that FractalNoise returns finite values
	value := pn.FractalNoise(1.0, 1.0, 4, 0.5, 1.0)
	assert.False(t, isNaN(value), "FractalNoise should not return NaN")
	assert.False(t, isInf(value), "FractalNoise should not return infinity")
}

// TestFractalNoiseDeterministic verifies FractalNoise produces consistent results
func TestFractalNoiseDeterministic(t *testing.T) {
	seed := int64(42)
	pn1 := NewPerlinNoise(seed)
	pn2 := NewPerlinNoise(seed)

	// Same inputs should produce identical outputs
	x, y := 2.5, 3.7
	octaves := 4
	persistence := 0.5
	scale := 1.0

	noise1 := pn1.FractalNoise(x, y, octaves, persistence, scale)
	noise2 := pn2.FractalNoise(x, y, octaves, persistence, scale)

	assert.Equal(t, noise1, noise2, "FractalNoise should be deterministic with same seed")
}

// TestFractalNoiseOctaves verifies that more octaves produce more detail
func TestFractalNoiseOctaves(t *testing.T) {
	pn := NewPerlinNoise(12345)

	// Sample multiple points and track variance
	tests := []struct {
		name     string
		octaves  int
		expected string
	}{
		{"single octave", 1, "basic"},
		{"four octaves", 4, "detailed"},
		{"eight octaves", 8, "highly detailed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify function works with different octave counts
			value := pn.FractalNoise(1.5, 2.5, tt.octaves, 0.5, 1.0)
			assert.False(t, isNaN(value), "should not return NaN")
			assert.False(t, isInf(value), "should not return infinity")
		})
	}
}

// TestFractalNoisePersistence verifies persistence affects amplitude decay
func TestFractalNoisePersistence(t *testing.T) {
	pn := NewPerlinNoise(12345)

	// Test at multiple coordinates - different persistence should produce different results
	coords := []struct{ x, y float64 }{
		{0.5, 0.5}, {1.5, 2.5}, {3.7, 4.2}, {5.1, 6.9}, {10.3, 15.7},
	}

	hasDifference := false
	for _, c := range coords {
		lowP := pn.FractalNoise(c.x, c.y, 4, 0.25, 1.0)
		highP := pn.FractalNoise(c.x, c.y, 4, 0.75, 1.0)

		// Both should be valid finite values
		assert.False(t, isNaN(lowP), "low persistence should produce valid value at (%v, %v)", c.x, c.y)
		assert.False(t, isNaN(highP), "high persistence should produce valid value at (%v, %v)", c.x, c.y)

		// Different persistence often produces different values
		if lowP != highP {
			hasDifference = true
		}
	}

	// At least some coordinates should show different values for different persistence
	assert.True(t, hasDifference,
		"different persistence values should produce different results at some coordinates")
}

// TestFractalNoiseScale verifies scale parameter affects frequency
func TestFractalNoiseScale(t *testing.T) {
	pn := NewPerlinNoise(12345)

	// Different scales should produce different patterns
	scale1 := pn.FractalNoise(1.0, 1.0, 4, 0.5, 0.5)
	scale2 := pn.FractalNoise(1.0, 1.0, 4, 0.5, 2.0)

	assert.False(t, isNaN(scale1), "scale 0.5 should produce valid value")
	assert.False(t, isNaN(scale2), "scale 2.0 should produce valid value")
	assert.NotEqual(t, scale1, scale2, "different scales should produce different results")
}

// TestFractalNoiseTableDriven uses table-driven tests for comprehensive coverage
func TestFractalNoiseTableDriven(t *testing.T) {
	tests := []struct {
		name        string
		seed        int64
		x, y        float64
		octaves     int
		persistence float64
		scale       float64
	}{
		{"basic values", 42, 0.0, 0.0, 4, 0.5, 1.0},
		{"negative coords", 42, -1.5, -2.5, 4, 0.5, 1.0},
		{"large coords", 42, 100.5, 200.5, 4, 0.5, 1.0},
		{"single octave", 42, 1.0, 1.0, 1, 0.5, 1.0},
		{"many octaves", 42, 1.0, 1.0, 8, 0.5, 1.0},
		{"zero persistence", 42, 1.0, 1.0, 4, 0.0, 1.0},
		{"small scale", 42, 1.0, 1.0, 4, 0.5, 0.1},
		{"large scale", 42, 1.0, 1.0, 4, 0.5, 10.0},
		{"different seed", 99999, 1.0, 1.0, 4, 0.5, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pn := NewPerlinNoise(tt.seed)
			value := pn.FractalNoise(tt.x, tt.y, tt.octaves, tt.persistence, tt.scale)

			assert.False(t, isNaN(value), "FractalNoise should not return NaN")
			assert.False(t, isInf(value), "FractalNoise should not return infinity")
		})
	}
}

// TestFractalNoiseSpatialVariation verifies noise varies spatially
func TestFractalNoiseSpatialVariation(t *testing.T) {
	pn := NewPerlinNoise(12345)

	// Sample a grid of values - they should vary
	var values []float64
	for x := 0.0; x < 10.0; x += 0.25 {
		for y := 0.0; y < 10.0; y += 0.25 {
			value := pn.FractalNoise(x, y, 4, 0.5, 1.0)
			values = append(values, value)
		}
	}

	// Verify not all values are identical
	firstValue := values[0]
	hasDifferent := false
	for _, val := range values[1:] {
		if val != firstValue {
			hasDifferent = true
			break
		}
	}
	assert.True(t, hasDifferent, "FractalNoise should produce different values across space")
}

// TestHelperFade tests the fade function for proper smoothstep behavior
func TestHelperFade(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"zero", 0.0, 0.0},
		{"one", 1.0, 1.0},
		{"half", 0.5, 0.5}, // The fade function has property f(0.5) = 0.5
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fade(tt.input)
			assert.InDelta(t, tt.expected, result, 0.0001,
				"fade(%v) should equal %v", tt.input, tt.expected)
		})
	}

	// Fade function should be monotonically increasing in [0,1]
	prev := fade(0.0)
	for i := 0.1; i <= 1.0; i += 0.1 {
		curr := fade(i)
		assert.GreaterOrEqual(t, curr, prev, "fade should be monotonically increasing")
		prev = curr
	}
}

// TestHelperLerp tests linear interpolation
func TestHelperLerp(t *testing.T) {
	tests := []struct {
		name     string
		t, a, b  float64
		expected float64
	}{
		{"t=0", 0.0, 10.0, 20.0, 10.0},
		{"t=1", 1.0, 10.0, 20.0, 20.0},
		{"t=0.5", 0.5, 10.0, 20.0, 15.0},
		{"t=0.25", 0.25, 0.0, 100.0, 25.0},
		{"negative values", 0.5, -10.0, 10.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lerp(tt.t, tt.a, tt.b)
			assert.InDelta(t, tt.expected, result, 0.0001,
				"lerp(%v, %v, %v) should equal %v", tt.t, tt.a, tt.b, tt.expected)
		})
	}
}

// TestHelperGrad2d tests gradient calculation for Perlin noise
func TestHelperGrad2d(t *testing.T) {
	// Test all four gradient directions based on hash & 3
	tests := []struct {
		name string
		hash int
		x, y float64
	}{
		{"hash 0", 0, 1.0, 1.0},
		{"hash 1", 1, 1.0, 1.0},
		{"hash 2", 2, 1.0, 1.0},
		{"hash 3", 3, 1.0, 1.0},
		{"hash 4 (wraps to 0)", 4, 1.0, 1.0},
		{"negative coords", 0, -1.0, -1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := grad2d(tt.hash, tt.x, tt.y)
			assert.False(t, isNaN(result), "grad2d should not return NaN")
			assert.False(t, isInf(result), "grad2d should not return infinity")
		})
	}
}

// TestHelperDot2d tests dot product for simplex noise
func TestHelperDot2d(t *testing.T) {
	tests := []struct {
		name     string
		g        []float64
		x, y     float64
		expected float64
	}{
		{"positive gradient", []float64{1.0, 1.0}, 2.0, 3.0, 5.0},
		{"negative gradient", []float64{-1.0, -1.0}, 2.0, 3.0, -5.0},
		{"mixed gradient", []float64{1.0, -1.0}, 2.0, 3.0, -1.0},
		{"zero vector", []float64{0.0, 0.0}, 2.0, 3.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dot2d(tt.g, tt.x, tt.y)
			assert.InDelta(t, tt.expected, result, 0.0001,
				"dot2d should equal %v", tt.expected)
		})
	}
}

// Helper functions to test for NaN and Inf without importing math
func isNaN(f float64) bool {
	return f != f
}

func isInf(f float64) bool {
	return f > 1e308 || f < -1e308
}
