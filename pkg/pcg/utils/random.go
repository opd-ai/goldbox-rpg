package utils

import "math/rand"

// WeightedRandomSelect selects a key from a weighted map using weighted random selection.
// Returns the selected key and true if successful, or the zero value and false if weights is empty.
func WeightedRandomSelect[K comparable](rng *rand.Rand, weights map[K]int, fallback K) K {
	if len(weights) == 0 {
		return fallback
	}

	totalWeight := 0
	for _, weight := range weights {
		totalWeight += weight
	}

	if totalWeight == 0 {
		return fallback
	}

	randomValue := rng.Intn(totalWeight)
	currentWeight := 0

	for key, weight := range weights {
		currentWeight += weight
		if randomValue < currentWeight {
			return key
		}
	}

	return fallback
}
