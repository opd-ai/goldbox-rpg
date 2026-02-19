package quests

import (
	"fmt"
	"testing"

	"goldbox-rpg/pkg/pcg"
)

func TestNewObjectiveGenerator(t *testing.T) {
	og := NewObjectiveGenerator()

	if og == nil {
		t.Fatal("NewObjectiveGenerator returned nil")
	}
}

func TestGenerateKillObjective(t *testing.T) {
	og := NewObjectiveGenerator()

	testCases := []struct {
		name       string
		difficulty int
		genCtx     *pcg.GenerationContext
		wantErr    bool
	}{
		{
			name:       "valid_difficulty_5",
			difficulty: 5,
			genCtx:     createTestGenerationContext(),
			wantErr:    false,
		},
		{
			name:       "low_difficulty_1",
			difficulty: 1,
			genCtx:     createTestGenerationContext(),
			wantErr:    false,
		},
		{
			name:       "high_difficulty_10",
			difficulty: 10,
			genCtx:     createTestGenerationContext(),
			wantErr:    false,
		},
		{
			name:       "invalid_difficulty_too_low",
			difficulty: 0,
			genCtx:     createTestGenerationContext(),
			wantErr:    true,
		},
		{
			name:       "invalid_difficulty_too_high",
			difficulty: 11,
			genCtx:     createTestGenerationContext(),
			wantErr:    true,
		},
		{
			name:       "nil_generation_context",
			difficulty: 5,
			genCtx:     nil,
			wantErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			objective, err := og.GenerateKillObjective(tc.difficulty, tc.genCtx)

			if tc.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if objective == nil {
				t.Fatal("Generated objective is nil")
			}

			// Validate objective structure
			if objective.ID == "" {
				t.Error("Objective ID is empty")
			}

			if objective.Type != "kill" {
				t.Errorf("Expected type 'kill', got '%s'", objective.Type)
			}

			if objective.Description == "" {
				t.Error("Objective description is empty")
			}

			if objective.Target == "" {
				t.Error("Objective target is empty")
			}

			if objective.Quantity <= 0 {
				t.Errorf("Invalid quantity: %d", objective.Quantity)
			}

			if objective.Progress != 0 {
				t.Errorf("Expected progress 0, got %d", objective.Progress)
			}

			if objective.Complete {
				t.Error("New objective should not be complete")
			}

			if objective.Conditions == nil {
				t.Error("Objective conditions should not be nil")
			}
		})
	}
}

func TestGenerateFetchObjective(t *testing.T) {
	og := NewObjectiveGenerator()

	testCases := []struct {
		name        string
		playerLevel int
		genCtx      *pcg.GenerationContext
		wantErr     bool
	}{
		{
			name:        "low_level_player",
			playerLevel: 3,
			genCtx:      createTestGenerationContext(),
			wantErr:     false,
		},
		{
			name:        "mid_level_player",
			playerLevel: 8,
			genCtx:      createTestGenerationContext(),
			wantErr:     false,
		},
		{
			name:        "high_level_player",
			playerLevel: 17,
			genCtx:      createTestGenerationContext(),
			wantErr:     false,
		},
		{
			name:        "invalid_level_too_low",
			playerLevel: 0,
			genCtx:      createTestGenerationContext(),
			wantErr:     true,
		},
		{
			name:        "invalid_level_too_high",
			playerLevel: 21,
			genCtx:      createTestGenerationContext(),
			wantErr:     true,
		},
		{
			name:        "nil_generation_context",
			playerLevel: 5,
			genCtx:      nil,
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			objective, err := og.GenerateFetchObjective(tc.playerLevel, tc.genCtx)

			if tc.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if objective == nil {
				t.Fatal("Generated objective is nil")
			}

			// Validate objective structure
			if objective.ID == "" {
				t.Error("Objective ID is empty")
			}

			if objective.Type != "fetch" {
				t.Errorf("Expected type 'fetch', got '%s'", objective.Type)
			}

			if objective.Description == "" {
				t.Error("Objective description is empty")
			}

			if objective.Target == "" {
				t.Error("Objective target is empty")
			}

			if objective.Quantity <= 0 {
				t.Errorf("Invalid quantity: %d", objective.Quantity)
			}

			if objective.Progress != 0 {
				t.Errorf("Expected progress 0, got %d", objective.Progress)
			}

			if objective.Complete {
				t.Error("New objective should not be complete")
			}

			if objective.Conditions == nil {
				t.Error("Objective conditions should not be nil")
			}

			// Check that pickup and delivery conditions are set
			if _, hasPickup := objective.Conditions["pickup"]; !hasPickup {
				t.Error("Fetch objective should have pickup condition")
			}

			if _, hasDelivery := objective.Conditions["delivery"]; !hasDelivery {
				t.Error("Fetch objective should have delivery condition")
			}
		})
	}
}

func TestGenerateExploreObjective(t *testing.T) {
	og := NewObjectiveGenerator()

	testCases := []struct {
		name    string
		genCtx  *pcg.GenerationContext
		wantErr bool
	}{
		{
			name:    "valid_context",
			genCtx:  createTestGenerationContext(),
			wantErr: false,
		},
		{
			name:    "nil_generation_context",
			genCtx:  nil,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			objective, err := og.GenerateExploreObjective(tc.genCtx)

			if tc.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if objective == nil {
				t.Fatal("Generated objective is nil")
			}

			// Validate objective structure
			if objective.ID == "" {
				t.Error("Objective ID is empty")
			}

			if objective.Type != "explore" {
				t.Errorf("Expected type 'explore', got '%s'", objective.Type)
			}

			if objective.Description == "" {
				t.Error("Objective description is empty")
			}

			if objective.Target == "" {
				t.Error("Objective target is empty")
			}

			if objective.Quantity < 70 || objective.Quantity > 100 {
				t.Errorf("Invalid percentage: %d (should be 70-100)", objective.Quantity)
			}

			if objective.Progress != 0 {
				t.Errorf("Expected progress 0, got %d", objective.Progress)
			}

			if objective.Complete {
				t.Error("New objective should not be complete")
			}

			if objective.Conditions == nil {
				t.Error("Objective conditions should not be nil")
			}

			// Check that area and percentage conditions are set
			if _, hasArea := objective.Conditions["area"]; !hasArea {
				t.Error("Explore objective should have area condition")
			}

			if _, hasPercentage := objective.Conditions["percentage"]; !hasPercentage {
				t.Error("Explore objective should have percentage condition")
			}
		})
	}
}

func TestSelectEnemyTypesForDifficulty(t *testing.T) {
	og := NewObjectiveGenerator()

	testCases := []struct {
		difficulty int
		minEnemies int
	}{
		{difficulty: 1, minEnemies: 3},
		{difficulty: 5, minEnemies: 3},
		{difficulty: 10, minEnemies: 3},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("difficulty_%d", tc.difficulty), func(t *testing.T) {
			enemies := og.selectEnemyTypesForDifficulty(tc.difficulty)

			if len(enemies) < tc.minEnemies {
				t.Errorf("Expected at least %d enemies for difficulty %d, got %d",
					tc.minEnemies, tc.difficulty, len(enemies))
			}

			// Check that all enemies are non-empty strings
			for _, enemy := range enemies {
				if enemy == "" {
					t.Error("Found empty enemy name")
				}
			}
		})
	}
}

func TestSelectItemTypesForLevel(t *testing.T) {
	og := NewObjectiveGenerator()

	testCases := []struct {
		level    int
		minItems int
	}{
		{level: 1, minItems: 1},
		{level: 5, minItems: 1},
		{level: 10, minItems: 1},
		{level: 15, minItems: 1},
		{level: 20, minItems: 1},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("level_%d", tc.level), func(t *testing.T) {
			items := og.selectItemTypesForLevel(tc.level)

			if len(items) < tc.minItems {
				t.Errorf("Expected at least %d items for level %d, got %d",
					tc.minItems, tc.level, len(items))
			}

			// Check that all items are non-empty strings
			for _, item := range items {
				if item == "" {
					t.Error("Found empty item name")
				}
			}
		})
	}
}

func TestIsCommonItem(t *testing.T) {
	testCases := []struct {
		item     string
		expected bool
	}{
		{"Health Potion", true},
		{"Rope", true},
		{"Torch", true},
		{"Magic Sword", false},
		{"Dragon Scale", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.item, func(t *testing.T) {
			result := isCommonItem(tc.item)
			if result != tc.expected {
				t.Errorf("isCommonItem(%q) = %v, want %v", tc.item, result, tc.expected)
			}
		})
	}
}

func TestObjectivesGeneratorDeterministicGeneration(t *testing.T) {
	og := NewObjectiveGenerator()

	// Create two identical generation contexts with the same seed
	genCtx1 := createTestGenerationContextWithSeed(12345)
	genCtx2 := createTestGenerationContextWithSeed(12345)

	// Generate objectives with same parameters
	obj1, err1 := og.GenerateKillObjective(5, genCtx1)
	if err1 != nil {
		t.Fatalf("First generation failed: %v", err1)
	}

	obj2, err2 := og.GenerateKillObjective(5, genCtx2)
	if err2 != nil {
		t.Fatalf("Second generation failed: %v", err2)
	}

	// They should be identical (except for ID which might include timestamps)
	if obj1.Type != obj2.Type {
		t.Errorf("Types differ: %s vs %s", obj1.Type, obj2.Type)
	}

	if obj1.Target != obj2.Target {
		t.Errorf("Targets differ: %s vs %s", obj1.Target, obj2.Target)
	}

	if obj1.Quantity != obj2.Quantity {
		t.Errorf("Quantities differ: %d vs %d", obj1.Quantity, obj2.Quantity)
	}
}

// Helper functions

func createTestGenerationContext() *pcg.GenerationContext {
	return createTestGenerationContextWithSeed(42)
}

func createTestGenerationContextWithSeed(seed int64) *pcg.GenerationContext {
	seedMgr := pcg.NewSeedManager(seed)
	params := pcg.GenerationParams{
		Seed:        seed,
		Difficulty:  5,
		PlayerLevel: 5,
	}
	return pcg.NewGenerationContext(seedMgr, pcg.ContentTypeQuests, "test", params)
}
