package pcg

import (
	"math/rand"
	"testing"
)

func TestNewSeedManager(t *testing.T) {
	tests := []struct {
		name     string
		baseSeed int64
		wantSeed bool // true if we expect the provided seed, false if we expect auto-generated
	}{
		{
			name:     "with provided seed",
			baseSeed: 12345,
			wantSeed: true,
		},
		{
			name:     "with zero seed should auto-generate",
			baseSeed: 0,
			wantSeed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewSeedManager(tt.baseSeed)

			if sm == nil {
				t.Error("NewSeedManager returned nil")
				return
			}

			if tt.wantSeed {
				if sm.GetBaseSeed() != tt.baseSeed {
					t.Errorf("Expected base seed %d, got %d", tt.baseSeed, sm.GetBaseSeed())
				}
			} else {
				if sm.GetBaseSeed() == 0 {
					t.Error("Expected auto-generated seed, but got 0")
				}
			}

			if sm.contextSeeds == nil {
				t.Error("contextSeeds map not initialized")
			}
		})
	}
}

func TestSeedManager_GetBaseSeed(t *testing.T) {
	expectedSeed := int64(987654321)
	sm := NewSeedManager(expectedSeed)

	if got := sm.GetBaseSeed(); got != expectedSeed {
		t.Errorf("GetBaseSeed() = %d, want %d", got, expectedSeed)
	}
}

func TestSeedManager_DeriveContextSeed(t *testing.T) {
	sm := NewSeedManager(12345)

	tests := []struct {
		name        string
		contentType ContentType
		contextName string
	}{
		{
			name:        "terrain context",
			contentType: ContentTypeTerrain,
			contextName: "forest_level_1",
		},
		{
			name:        "items context",
			contentType: ContentTypeItems,
			contextName: "magic_sword",
		},
		{
			name:        "quests context",
			contentType: ContentTypeQuests,
			contextName: "fetch_quest_1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First call should create and cache the seed
			seed1 := sm.DeriveContextSeed(tt.contentType, tt.contextName)
			if seed1 == 0 {
				t.Error("DeriveContextSeed returned 0")
			}

			// Second call should return the same cached seed
			seed2 := sm.DeriveContextSeed(tt.contentType, tt.contextName)
			if seed1 != seed2 {
				t.Errorf("DeriveContextSeed not deterministic: first=%d, second=%d", seed1, seed2)
			}

			// Different context should produce different seed
			differentSeed := sm.DeriveContextSeed(tt.contentType, tt.contextName+"_different")
			if seed1 == differentSeed {
				t.Error("Different contexts produced the same seed")
			}
		})
	}
}

func TestSeedManager_DeriveParameterSeed(t *testing.T) {
	sm := NewSeedManager(12345)
	baseSeed := int64(54321)

	// Test with simple parameters (no constraints map that can vary)
	params1 := GenerationParams{
		Difficulty:  5,
		PlayerLevel: 10,
	}

	params2 := GenerationParams{
		Difficulty:  5,
		PlayerLevel: 10,
	}

	params3 := GenerationParams{
		Difficulty:  6, // Different difficulty
		PlayerLevel: 10,
	}

	// Same parameters should produce same seed
	seed1 := sm.DeriveParameterSeed(baseSeed, params1)
	seed2 := sm.DeriveParameterSeed(baseSeed, params2)
	if seed1 != seed2 {
		t.Errorf("Same parameters produced different seeds: %d vs %d", seed1, seed2)
	}

	// Different parameters should produce different seed
	seed3 := sm.DeriveParameterSeed(baseSeed, params3)
	if seed1 == seed3 {
		t.Error("Different parameters produced the same seed")
	}

	// Test that different base seeds produce different results
	seed4 := sm.DeriveParameterSeed(baseSeed+1, params1)
	if seed1 == seed4 {
		t.Error("Different base seeds produced the same seed")
	}
}

func TestSeedManager_CreateRNG(t *testing.T) {
	sm := NewSeedManager(12345)
	contentType := ContentTypeTerrain
	name := "test_terrain"
	params := GenerationParams{
		Difficulty:  5,
		PlayerLevel: 10,
	}

	// Create two RNGs with same parameters
	rng1 := sm.CreateRNG(contentType, name, params)
	rng2 := sm.CreateRNG(contentType, name, params)

	if rng1 == nil || rng2 == nil {
		t.Error("CreateRNG returned nil")
		return
	}

	// They should produce the same sequence
	val1a := rng1.Int63()
	val2a := rng2.Int63()
	if val1a != val2a {
		t.Errorf("RNGs with same parameters produced different values: %d vs %d", val1a, val2a)
	}

	// Next values should also match
	val1b := rng1.Int63()
	val2b := rng2.Int63()
	if val1b != val2b {
		t.Errorf("RNGs sequence diverged: %d vs %d", val1b, val2b)
	}
}

func TestSeedManager_CreateSubRNG(t *testing.T) {
	sm := NewSeedManager(12345)

	// Create a parent RNG
	parentRNG := rand.New(rand.NewSource(54321))
	phase := "room_generation"

	// Create sub-RNGs
	subRNG1 := sm.CreateSubRNG(parentRNG, phase)

	// Reset parent RNG and create another sub-RNG with same phase
	parentRNG = rand.New(rand.NewSource(54321))
	subRNG2 := sm.CreateSubRNG(parentRNG, phase)

	if subRNG1 == nil || subRNG2 == nil {
		t.Error("CreateSubRNG returned nil")
		return
	}

	// They should produce the same sequence
	val1 := subRNG1.Int63()
	val2 := subRNG2.Int63()
	if val1 != val2 {
		t.Errorf("Sub-RNGs with same phase produced different values: %d vs %d", val1, val2)
	}
}

func TestSeedManager_SaveableState(t *testing.T) {
	sm := NewSeedManager(12345)

	// Generate some context seeds to populate the state
	sm.DeriveContextSeed(ContentTypeTerrain, "forest")
	sm.DeriveContextSeed(ContentTypeItems, "sword")

	// Get saveable state
	state := sm.GetSaveableState()

	if state.BaseSeed != 12345 {
		t.Errorf("Expected base seed 12345, got %d", state.BaseSeed)
	}

	if len(state.ContextSeeds) != 2 {
		t.Errorf("Expected 2 context seeds, got %d", len(state.ContextSeeds))
	}

	// Create new seed manager and load state
	sm2 := NewSeedManager(0) // Different initial seed
	sm2.LoadState(state)

	if sm2.GetBaseSeed() != 12345 {
		t.Errorf("LoadState didn't restore base seed correctly")
	}

	// Should produce same context seeds
	forestSeed1 := sm.DeriveContextSeed(ContentTypeTerrain, "forest")
	forestSeed2 := sm2.DeriveContextSeed(ContentTypeTerrain, "forest")
	if forestSeed1 != forestSeed2 {
		t.Errorf("LoadState didn't preserve context seeds: %d vs %d", forestSeed1, forestSeed2)
	}
}

func TestNewGenerationContext(t *testing.T) {
	sm := NewSeedManager(12345)
	contentType := ContentTypeTerrain
	name := "test_context"
	params := GenerationParams{
		Difficulty:  5,
		PlayerLevel: 10,
	}

	gc := NewGenerationContext(sm, contentType, name, params)

	if gc == nil {
		t.Error("NewGenerationContext returned nil")
		return
	}

	if gc.RNG == nil {
		t.Error("GenerationContext RNG is nil")
	}

	if gc.Phase != "main" {
		t.Errorf("Expected phase 'main', got '%s'", gc.Phase)
	}

	if gc.SeedMgr != sm {
		t.Error("GenerationContext doesn't reference correct SeedManager")
	}

	if gc.SubRNGs == nil {
		t.Error("SubRNGs map not initialized")
	}
}

func TestGenerationContext_GetSubRNG(t *testing.T) {
	sm := NewSeedManager(12345)
	gc := NewGenerationContext(sm, ContentTypeTerrain, "test", GenerationParams{})

	phase := "test_phase"

	// First call should create the sub-RNG
	subRNG1 := gc.GetSubRNG(phase)
	if subRNG1 == nil {
		t.Error("GetSubRNG returned nil")
		return
	}

	// Second call should return the same cached sub-RNG
	subRNG2 := gc.GetSubRNG(phase)
	if subRNG1 != subRNG2 {
		t.Error("GetSubRNG didn't return cached sub-RNG")
	}

	// Different phase should return different sub-RNG
	subRNG3 := gc.GetSubRNG("different_phase")
	if subRNG1 == subRNG3 {
		t.Error("Different phases returned the same sub-RNG")
	}
}

func TestGenerationContext_RollDice(t *testing.T) {
	sm := NewSeedManager(12345)
	gc := NewGenerationContext(sm, ContentTypeTerrain, "test", GenerationParams{})

	tests := []struct {
		name  string
		sides int
		valid bool
	}{
		{"valid 6-sided die", 6, true},
		{"valid 20-sided die", 20, true},
		{"invalid 0 sides", 0, false},
		{"invalid negative sides", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gc.RollDice(tt.sides)

			if !tt.valid {
				if result != 0 {
					t.Errorf("Expected 0 for invalid sides, got %d", result)
				}
				return
			}

			if result < 1 || result > tt.sides {
				t.Errorf("RollDice(%d) = %d, want result between 1 and %d", tt.sides, result, tt.sides)
			}
		})
	}
}

func TestGenerationContext_RollMultipleDice(t *testing.T) {
	sm := NewSeedManager(12345)
	gc := NewGenerationContext(sm, ContentTypeTerrain, "test", GenerationParams{})

	count := 3
	sides := 6
	results := gc.RollMultipleDice(count, sides)

	if len(results) != count {
		t.Errorf("Expected %d results, got %d", count, len(results))
	}

	for i, result := range results {
		if result < 1 || result > sides {
			t.Errorf("Result[%d] = %d, want result between 1 and %d", i, result, sides)
		}
	}
}

func TestGenerationContext_RollDiceSum(t *testing.T) {
	sm := NewSeedManager(12345)
	gc := NewGenerationContext(sm, ContentTypeTerrain, "test", GenerationParams{})

	count := 3
	sides := 6
	sum := gc.RollDiceSum(count, sides)

	minSum := count * 1
	maxSum := count * sides

	if sum < minSum || sum > maxSum {
		t.Errorf("RollDiceSum(%d, %d) = %d, want sum between %d and %d", count, sides, sum, minSum, maxSum)
	}
}

func TestGenerationContext_RandomChoice(t *testing.T) {
	sm := NewSeedManager(12345)
	gc := NewGenerationContext(sm, ContentTypeTerrain, "test", GenerationParams{})

	// Test with valid choices
	choices := []string{"option1", "option2", "option3"}
	choice := gc.RandomChoice(choices)

	found := false
	for _, validChoice := range choices {
		if choice == validChoice {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("RandomChoice returned invalid choice: %s", choice)
	}

	// Test with empty choices
	emptyChoice := gc.RandomChoice([]string{})
	if emptyChoice != "" {
		t.Errorf("RandomChoice with empty slice should return empty string, got: %s", emptyChoice)
	}
}

func TestGenerationContext_RandomFloat(t *testing.T) {
	sm := NewSeedManager(12345)
	gc := NewGenerationContext(sm, ContentTypeTerrain, "test", GenerationParams{})

	result := gc.RandomFloat()

	if result < 0.0 || result >= 1.0 {
		t.Errorf("RandomFloat() = %f, want result between 0.0 and 1.0", result)
	}
}

func TestGenerationContext_RandomFloatRange(t *testing.T) {
	sm := NewSeedManager(12345)
	gc := NewGenerationContext(sm, ContentTypeTerrain, "test", GenerationParams{})

	min := 2.5
	max := 7.8
	result := gc.RandomFloatRange(min, max)

	if result < min || result > max {
		t.Errorf("RandomFloatRange(%f, %f) = %f, want result between %f and %f", min, max, result, min, max)
	}
}

func TestGenerationContext_RandomIntRange(t *testing.T) {
	sm := NewSeedManager(12345)
	gc := NewGenerationContext(sm, ContentTypeTerrain, "test", GenerationParams{})

	tests := []struct {
		name string
		min  int
		max  int
	}{
		{"normal range", 5, 15},
		{"single value", 10, 10},
		{"reversed range", 15, 5}, // min >= max case
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gc.RandomIntRange(tt.min, tt.max)

			if tt.min >= tt.max {
				if result != tt.min {
					t.Errorf("RandomIntRange(%d, %d) = %d, expected %d when min >= max", tt.min, tt.max, result, tt.min)
				}
			} else {
				if result < tt.min || result > tt.max {
					t.Errorf("RandomIntRange(%d, %d) = %d, want result between %d and %d", tt.min, tt.max, result, tt.min, tt.max)
				}
			}
		})
	}
}

func TestGenerationContext_WeightedChoice(t *testing.T) {
	sm := NewSeedManager(12345)
	gc := NewGenerationContext(sm, ContentTypeTerrain, "test", GenerationParams{})

	tests := []struct {
		name    string
		choices []string
		weights []float64
		wantErr bool
	}{
		{
			name:    "valid weighted choice",
			choices: []string{"A", "B", "C"},
			weights: []float64{1.0, 2.0, 1.0},
			wantErr: false,
		},
		{
			name:    "empty choices",
			choices: []string{},
			weights: []float64{},
			wantErr: true,
		},
		{
			name:    "mismatched lengths",
			choices: []string{"A", "B"},
			weights: []float64{1.0, 2.0, 3.0},
			wantErr: true,
		},
		{
			name:    "zero weights",
			choices: []string{"A", "B"},
			weights: []float64{0.0, 0.0},
			wantErr: false, // Should fallback to random choice
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gc.WeightedChoice(tt.choices, tt.weights)

			if tt.wantErr {
				if result != "" && len(tt.choices) == 0 {
					t.Errorf("WeightedChoice with empty choices should return empty string, got: %s", result)
				}
				return
			}

			if len(tt.choices) == 0 {
				return // Already handled above
			}

			// Check if result is one of the valid choices
			found := false
			for _, choice := range tt.choices {
				if result == choice {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("WeightedChoice returned invalid choice: %s", result)
			}
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkSeedManager_DeriveContextSeed(b *testing.B) {
	sm := NewSeedManager(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.DeriveContextSeed(ContentTypeTerrain, "benchmark_test")
	}
}

func BenchmarkGenerationContext_RollDice(b *testing.B) {
	sm := NewSeedManager(12345)
	gc := NewGenerationContext(sm, ContentTypeTerrain, "benchmark", GenerationParams{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gc.RollDice(20)
	}
}

// Test deterministic behavior across multiple runs
func TestDeterministicBehavior(t *testing.T) {
	baseSeed := int64(987654321)
	contentType := ContentTypeTerrain
	name := "deterministic_test"
	params := GenerationParams{
		Difficulty:  10,
		PlayerLevel: 5,
	}

	// Run the same generation twice
	results1 := runGenerationSequence(baseSeed, contentType, name, params)
	results2 := runGenerationSequence(baseSeed, contentType, name, params)

	// Results should be identical
	if len(results1) != len(results2) {
		t.Errorf("Different number of results: %d vs %d", len(results1), len(results2))
		return
	}

	for i, val1 := range results1 {
		if val1 != results2[i] {
			t.Errorf("Results differ at position %d: %d vs %d", i, val1, results2[i])
		}
	}
}

// Helper function for deterministic testing
func runGenerationSequence(baseSeed int64, contentType ContentType, name string, params GenerationParams) []int {
	sm := NewSeedManager(baseSeed)
	gc := NewGenerationContext(sm, contentType, name, params)

	results := make([]int, 10)
	for i := 0; i < 10; i++ {
		results[i] = gc.RollDice(20)
	}

	return results
}
