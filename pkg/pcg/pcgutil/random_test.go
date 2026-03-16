package pcgutil

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWeightedRandomSelect(t *testing.T) {
	tests := []struct {
		name     string
		weights  map[string]int
		fallback string
		want     string
	}{
		{
			name:     "EmptyWeights",
			weights:  map[string]int{},
			fallback: "default",
			want:     "default",
		},
		{
			name: "AllZeroWeights",
			weights: map[string]int{
				"a": 0,
				"b": 0,
			},
			fallback: "fallback",
			want:     "fallback",
		},
		{
			name: "SingleNonZeroWeight",
			weights: map[string]int{
				"only": 100,
			},
			fallback: "fallback",
			want:     "only",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rng := rand.New(rand.NewSource(42))
			got := WeightedRandomSelect(rng, tt.weights, tt.fallback)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWeightedRandomSelect_Distribution(t *testing.T) {
	weights := map[string]int{
		"high":   80,
		"medium": 15,
		"low":    5,
	}

	counts := make(map[string]int)
	iterations := 10000
	rng := rand.New(rand.NewSource(12345))

	for i := 0; i < iterations; i++ {
		result := WeightedRandomSelect(rng, weights, "fallback")
		counts[result]++
	}

	assert.Greater(t, counts["high"], counts["medium"], "high weight should be selected more than medium")
	assert.Greater(t, counts["medium"], counts["low"], "medium weight should be selected more than low")
	assert.Equal(t, 0, counts["fallback"], "fallback should never be selected with valid weights")
}

func TestWeightedRandomSelect_MixedWeights(t *testing.T) {
	weights := map[string]int{
		"a": 50,
		"b": 0,
		"c": 50,
	}

	counts := make(map[string]int)
	iterations := 1000
	rng := rand.New(rand.NewSource(99))

	for i := 0; i < iterations; i++ {
		result := WeightedRandomSelect(rng, weights, "fallback")
		counts[result]++
	}

	assert.Equal(t, 0, counts["b"], "zero weight items should never be selected")
	assert.Greater(t, counts["a"], 0, "non-zero weight items should be selected")
	assert.Greater(t, counts["c"], 0, "non-zero weight items should be selected")
}
