package pcg

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
	"time"
)

// SeedManager provides deterministic seeding for reproducible content generation
// Follows the established deterministic patterns in the existing dice system
type SeedManager struct {
	baseSeed     int64
	contextSeeds map[string]int64
}

// NewSeedManager creates a new seed manager with a base seed
func NewSeedManager(baseSeed int64) *SeedManager {
	if baseSeed == 0 {
		baseSeed = time.Now().UnixNano()
	}

	return &SeedManager{
		baseSeed:     baseSeed,
		contextSeeds: make(map[string]int64),
	}
}

// GetBaseSeed returns the base seed used for all generation
func (sm *SeedManager) GetBaseSeed() int64 {
	return sm.baseSeed
}

// DeriveContextSeed creates a deterministic seed for a specific context
// This ensures that the same content type/name combination always produces
// the same seed, enabling reproducible generation across sessions
func (sm *SeedManager) DeriveContextSeed(contentType ContentType, name string) int64 {
	context := fmt.Sprintf("%s:%s", contentType, name)

	if seed, exists := sm.contextSeeds[context]; exists {
		return seed
	}

	// Create deterministic seed by hashing base seed + context
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("%d:%s", sm.baseSeed, context)))
	hash := hasher.Sum(nil)

	// Convert first 8 bytes of hash to int64
	seed := int64(binary.BigEndian.Uint64(hash[:8]))

	sm.contextSeeds[context] = seed
	return seed
}

// DeriveParameterSeed creates a seed based on generation parameters
// This allows for controlled variation within the same generator context
func (sm *SeedManager) DeriveParameterSeed(baseSeed int64, params GenerationParams) int64 {
	hasher := sha256.New()

	// Include critical parameters that should affect generation
	paramString := fmt.Sprintf("%d:%d:%d",
		baseSeed,
		params.Difficulty,
		params.PlayerLevel)

	// Include any additional constraints that should affect seeding
	for key, value := range params.Constraints {
		paramString += fmt.Sprintf(":%s=%v", key, value)
	}

	hasher.Write([]byte(paramString))
	hash := hasher.Sum(nil)

	return int64(binary.BigEndian.Uint64(hash[:8]))
}

// CreateRNG creates a new random number generator with the derived seed
// This provides the same pattern as the existing DiceRoller system
func (sm *SeedManager) CreateRNG(contentType ContentType, name string, params GenerationParams) *rand.Rand {
	contextSeed := sm.DeriveContextSeed(contentType, name)
	finalSeed := sm.DeriveParameterSeed(contextSeed, params)

	return rand.New(rand.NewSource(finalSeed))
}

// CreateSubRNG creates a child RNG for a specific generation phase
// This allows deterministic sub-generation within a larger generation process
func (sm *SeedManager) CreateSubRNG(parentRNG *rand.Rand, phase string) *rand.Rand {
	// Use the parent RNG to get a deterministic seed for the sub-phase
	subSeed := parentRNG.Int63()

	// Hash the phase name with the sub-seed for determinism
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("%d:%s", subSeed, phase)))
	hash := hasher.Sum(nil)

	finalSeed := int64(binary.BigEndian.Uint64(hash[:8]))
	return rand.New(rand.NewSource(finalSeed))
}

// SaveableState represents the state that can be saved/loaded for reproducibility
type SaveableState struct {
	BaseSeed     int64            `yaml:"base_seed"`
	ContextSeeds map[string]int64 `yaml:"context_seeds"`
}

// GetSaveableState returns the current state for persistence
func (sm *SeedManager) GetSaveableState() SaveableState {
	return SaveableState{
		BaseSeed:     sm.baseSeed,
		ContextSeeds: sm.contextSeeds,
	}
}

// LoadState restores the seed manager from saved state
func (sm *SeedManager) LoadState(state SaveableState) {
	sm.baseSeed = state.BaseSeed
	sm.contextSeeds = make(map[string]int64)

	for context, seed := range state.ContextSeeds {
		sm.contextSeeds[context] = seed
	}
}

// GenerationContext provides context and seeded RNG for generators
type GenerationContext struct {
	RNG     *rand.Rand
	Seed    int64
	Phase   string
	SeedMgr *SeedManager
	SubRNGs map[string]*rand.Rand
}

// NewGenerationContext creates a new generation context
func NewGenerationContext(seedMgr *SeedManager, contentType ContentType, name string, params GenerationParams) *GenerationContext {
	rng := seedMgr.CreateRNG(contentType, name, params)

	return &GenerationContext{
		RNG:     rng,
		Seed:    seedMgr.DeriveContextSeed(contentType, name),
		Phase:   "main",
		SeedMgr: seedMgr,
		SubRNGs: make(map[string]*rand.Rand),
	}
}

// GetSubRNG returns a deterministic sub-RNG for the specified phase
func (gc *GenerationContext) GetSubRNG(phase string) *rand.Rand {
	if subRNG, exists := gc.SubRNGs[phase]; exists {
		return subRNG
	}

	subRNG := gc.SeedMgr.CreateSubRNG(gc.RNG, phase)
	gc.SubRNGs[phase] = subRNG
	return subRNG
}

// RollDice provides dice rolling functionality using the context's RNG
// This integrates with the existing dice system patterns
func (gc *GenerationContext) RollDice(sides int) int {
	if sides <= 0 {
		return 0
	}
	return gc.RNG.Intn(sides) + 1
}

// RollMultipleDice rolls multiple dice and returns individual results
func (gc *GenerationContext) RollMultipleDice(count, sides int) []int {
	results := make([]int, count)
	for i := 0; i < count; i++ {
		results[i] = gc.RollDice(sides)
	}
	return results
}

// RollDiceSum rolls multiple dice and returns the sum
func (gc *GenerationContext) RollDiceSum(count, sides int) int {
	total := 0
	for i := 0; i < count; i++ {
		total += gc.RollDice(sides)
	}
	return total
}

// RandomChoice selects a random element from a slice
func (gc *GenerationContext) RandomChoice(choices []string) string {
	if len(choices) == 0 {
		return ""
	}
	return choices[gc.RNG.Intn(len(choices))]
}

// RandomFloat returns a random float64 between 0.0 and 1.0
func (gc *GenerationContext) RandomFloat() float64 {
	return gc.RNG.Float64()
}

// RandomFloatRange returns a random float64 between min and max
func (gc *GenerationContext) RandomFloatRange(min, max float64) float64 {
	return min + gc.RNG.Float64()*(max-min)
}

// RandomIntRange returns a random int between min and max (inclusive)
func (gc *GenerationContext) RandomIntRange(min, max int) int {
	if min >= max {
		return min
	}
	return min + gc.RNG.Intn(max-min+1)
}

// WeightedChoice selects from choices based on weights
func (gc *GenerationContext) WeightedChoice(choices []string, weights []float64) string {
	if len(choices) == 0 || len(choices) != len(weights) {
		return ""
	}

	// Calculate total weight
	totalWeight := 0.0
	for _, weight := range weights {
		totalWeight += weight
	}

	if totalWeight <= 0 {
		return gc.RandomChoice(choices)
	}

	// Random value between 0 and total weight
	randomValue := gc.RNG.Float64() * totalWeight

	// Find the selected choice
	currentWeight := 0.0
	for i, weight := range weights {
		currentWeight += weight
		if randomValue <= currentWeight {
			return choices[i]
		}
	}

	// Fallback to last choice
	return choices[len(choices)-1]
}
