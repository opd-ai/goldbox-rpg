package pcg

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestNewFactionGenerator(t *testing.T) {
	tests := []struct {
		name   string
		logger *logrus.Logger
	}{
		{
			name:   "with nil logger",
			logger: nil,
		},
		{
			name:   "with provided logger",
			logger: logrus.New(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fg := NewFactionGenerator(tt.logger)

			if fg == nil {
				t.Fatal("NewFactionGenerator returned nil")
			}

			if fg.version != "1.0.0" {
				t.Errorf("expected version 1.0.0, got %s", fg.version)
			}

			if fg.logger == nil {
				t.Error("logger should not be nil")
			}

			if fg.rng == nil {
				t.Error("rng should not be nil")
			}
		})
	}
}

func TestFactionGenerator_GetType(t *testing.T) {
	fg := NewFactionGenerator(nil)
	expected := ContentTypeFactions

	if fg.GetType() != expected {
		t.Errorf("expected %s, got %s", expected, fg.GetType())
	}
}

func TestFactionGenerator_GetVersion(t *testing.T) {
	fg := NewFactionGenerator(nil)
	expected := "1.0.0"

	if fg.GetVersion() != expected {
		t.Errorf("expected %s, got %s", expected, fg.GetVersion())
	}
}

func TestFactionGenerator_Validate(t *testing.T) {
	fg := NewFactionGenerator(nil)

	tests := []struct {
		name    string
		params  GenerationParams
		wantErr bool
	}{
		{
			name: "valid parameters",
			params: GenerationParams{
				Seed:       12345,
				Difficulty: 5,
			},
			wantErr: false,
		},
		{
			name: "zero seed",
			params: GenerationParams{
				Seed:       0,
				Difficulty: 5,
			},
			wantErr: true,
		},
		{
			name: "difficulty too low",
			params: GenerationParams{
				Seed:       12345,
				Difficulty: 0,
			},
			wantErr: true,
		},
		{
			name: "difficulty too high",
			params: GenerationParams{
				Seed:       12345,
				Difficulty: 25,
			},
			wantErr: true,
		},
		{
			name: "valid faction constraints",
			params: GenerationParams{
				Seed:       12345,
				Difficulty: 5,
				Constraints: map[string]interface{}{
					"faction_params": FactionParams{
						FactionCount:  5,
						ConflictLevel: 0.5,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid faction count",
			params: GenerationParams{
				Seed:       12345,
				Difficulty: 5,
				Constraints: map[string]interface{}{
					"faction_params": FactionParams{
						FactionCount:  25,
						ConflictLevel: 0.5,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid conflict level",
			params: GenerationParams{
				Seed:       12345,
				Difficulty: 5,
				Constraints: map[string]interface{}{
					"faction_params": FactionParams{
						FactionCount:  5,
						ConflictLevel: 1.5,
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fg.Validate(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFactionGenerator_Generate(t *testing.T) {
	fg := NewFactionGenerator(nil)
	ctx := context.Background()

	tests := []struct {
		name                string
		params              GenerationParams
		wantErr             bool
		expectedMinFactions int
		expectedMaxFactions int
	}{
		{
			name: "default generation",
			params: GenerationParams{
				Seed:       12345,
				Difficulty: 5,
			},
			wantErr:             false,
			expectedMinFactions: 3,
			expectedMaxFactions: 10,
		},
		{
			name: "specific faction count",
			params: GenerationParams{
				Seed:       12345,
				Difficulty: 5,
				Constraints: map[string]interface{}{
					"faction_params": FactionParams{
						FactionCount:  7,
						ConflictLevel: 0.4,
						EconomicFocus: 0.6,
					},
				},
			},
			wantErr:             false,
			expectedMinFactions: 7,
			expectedMaxFactions: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fg.Generate(ctx, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				system, ok := result.(*GeneratedFactionSystem)
				if !ok {
					t.Fatal("Generate() did not return *GeneratedFactionSystem")
				}

				// Verify basic structure
				if system.ID == "" {
					t.Error("system should have an ID")
				}

				if system.Name == "" {
					t.Error("system should have a name")
				}

				// Check faction count
				factionCount := len(system.Factions)
				if factionCount < tt.expectedMinFactions || factionCount > tt.expectedMaxFactions {
					t.Errorf("expected %d-%d factions, got %d", tt.expectedMinFactions, tt.expectedMaxFactions, factionCount)
				}

				// Verify all factions have required fields
				for i, faction := range system.Factions {
					if faction.ID == "" {
						t.Errorf("faction %d should have an ID", i)
					}
					if faction.Name == "" {
						t.Errorf("faction %d should have a name", i)
					}
					if faction.Power < 1 {
						t.Errorf("faction %d should have power >= 1, got %d", i, faction.Power)
					}
				}

				// Verify relationships exist
				expectedRelationships := (factionCount * (factionCount - 1)) / 2
				if len(system.Relationships) != expectedRelationships {
					t.Errorf("expected %d relationships, got %d", expectedRelationships, len(system.Relationships))
				}

				// Verify territories exist
				if len(system.Territories) == 0 {
					t.Error("system should have at least one territory")
				}

				// Check timestamp
				if system.Generated.IsZero() {
					t.Error("system should have a generation timestamp")
				}
			}
		})
	}
}

func TestFactionGenerator_DeterministicGeneration(t *testing.T) {
	fg1 := NewFactionGenerator(nil)
	fg2 := NewFactionGenerator(nil)
	ctx := context.Background()

	params := GenerationParams{
		Seed:       42,
		Difficulty: 5,
		Constraints: map[string]interface{}{
			"faction_params": FactionParams{
				FactionCount:  5,
				ConflictLevel: 0.3,
			},
		},
	}

	result1, err1 := fg1.Generate(ctx, params)
	if err1 != nil {
		t.Fatalf("first generation failed: %v", err1)
	}

	result2, err2 := fg2.Generate(ctx, params)
	if err2 != nil {
		t.Fatalf("second generation failed: %v", err2)
	}

	system1 := result1.(*GeneratedFactionSystem)
	system2 := result2.(*GeneratedFactionSystem)

	// Verify deterministic generation produces same results
	if len(system1.Factions) != len(system2.Factions) {
		t.Errorf("faction counts differ: %d vs %d", len(system1.Factions), len(system2.Factions))
	}

	// Check that faction names and types are consistent
	for i := 0; i < len(system1.Factions) && i < len(system2.Factions); i++ {
		f1, f2 := system1.Factions[i], system2.Factions[i]
		if f1.Name != f2.Name {
			t.Errorf("faction %d names differ: %s vs %s", i, f1.Name, f2.Name)
		}
		if f1.Type != f2.Type {
			t.Errorf("faction %d types differ: %s vs %s", i, f1.Type, f2.Type)
		}
		if f1.Power != f2.Power {
			t.Errorf("faction %d power differs: %d vs %d", i, f1.Power, f2.Power)
		}
	}
}

func TestFactionGenerator_HelperMethods(t *testing.T) {
	fg := NewFactionGenerator(nil)

	t.Run("generateID", func(t *testing.T) {
		id1 := fg.generateID("test")
		id2 := fg.generateID("test")

		if id1 == id2 {
			t.Error("generateID should produce unique IDs")
		}

		if id1[:5] != "test_" {
			t.Errorf("ID should start with prefix, got %s", id1)
		}
	})

	t.Run("selectFactionType", func(t *testing.T) {
		factionType := fg.selectFactionType()
		validTypes := []FactionType{
			FactionTypeMilitary, FactionTypeEconomic, FactionTypeReligious,
			FactionTypeCriminal, FactionTypeScholarly, FactionTypePolitical,
			FactionTypeMercenary, FactionTypeMagical,
		}

		found := false
		for _, validType := range validTypes {
			if factionType == validType {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("invalid faction type: %s", factionType)
		}
	})

	t.Run("determineRelationshipStatus", func(t *testing.T) {
		tests := []struct {
			opinion  float64
			expected RelationshipStatus
		}{
			{0.8, RelationStatusAllied},
			{0.5, RelationStatusFriendly},
			{0.0, RelationStatusNeutral},
			{-0.5, RelationStatusTense},
			{-0.8, RelationStatusHostile},
			{-0.95, RelationStatusWar},
		}

		for _, tt := range tests {
			result := fg.determineRelationshipStatus(tt.opinion)
			if result != tt.expected {
				t.Errorf("opinion %f should give %s, got %s", tt.opinion, tt.expected, result)
			}
		}
	})
}

func TestFactionGenerator_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	fg := NewFactionGenerator(nil)
	ctx := context.Background()

	params := GenerationParams{
		Seed:       time.Now().UnixNano(),
		Difficulty: 10,
		Constraints: map[string]interface{}{
			"faction_params": FactionParams{
				FactionCount:  10,
				ConflictLevel: 0.5,
			},
		},
	}

	start := time.Now()
	result, err := fg.Generate(ctx, params)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	system := result.(*GeneratedFactionSystem)

	// Performance benchmarks
	if duration > time.Second {
		t.Errorf("generation took too long: %v", duration)
	}

	t.Logf("Generated faction system with %d factions in %v", len(system.Factions), duration)
	t.Logf("System contains %d relationships, %d territories, %d trade deals, %d conflicts",
		len(system.Relationships), len(system.Territories), len(system.TradeDeals), len(system.Conflicts))
}

// Benchmark tests
func BenchmarkFactionGenerator_Generate(b *testing.B) {
	fg := NewFactionGenerator(nil)
	ctx := context.Background()

	params := GenerationParams{
		Seed:       42,
		Difficulty: 5,
		Constraints: map[string]interface{}{
			"faction_params": FactionParams{
				FactionCount:  6,
				ConflictLevel: 0.4,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		params.Seed = int64(i + 1) // Vary seed for each iteration
		_, err := fg.Generate(ctx, params)
		if err != nil {
			b.Fatalf("generation failed: %v", err)
		}
	}
}

func BenchmarkFactionGenerator_SmallSystem(b *testing.B) {
	fg := NewFactionGenerator(nil)
	ctx := context.Background()

	params := GenerationParams{
		Seed:       42,
		Difficulty: 3,
		Constraints: map[string]interface{}{
			"faction_params": FactionParams{
				FactionCount:  3,
				ConflictLevel: 0.2,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		params.Seed = int64(i + 1)
		_, err := fg.Generate(ctx, params)
		if err != nil {
			b.Fatalf("generation failed: %v", err)
		}
	}
}

func BenchmarkFactionGenerator_LargeSystem(b *testing.B) {
	fg := NewFactionGenerator(nil)
	ctx := context.Background()

	params := GenerationParams{
		Seed:       42,
		Difficulty: 10,
		Constraints: map[string]interface{}{
			"faction_params": FactionParams{
				FactionCount:  15,
				ConflictLevel: 0.7,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		params.Seed = int64(i + 1)
		_, err := fg.Generate(ctx, params)
		if err != nil {
			b.Fatalf("generation failed: %v", err)
		}
	}
}
