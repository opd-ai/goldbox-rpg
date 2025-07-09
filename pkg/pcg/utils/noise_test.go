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

// Helper functions to test for NaN and Inf without importing math
func isNaN(f float64) bool {
	return f != f
}

func isInf(f float64) bool {
	return f > 1e308 || f < -1e308
}
