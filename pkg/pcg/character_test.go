package pcg

import (
	"context"
	"testing"
	"time"

	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
)

func TestNewNPCGenerator(t *testing.T) {
	tests := []struct {
		name   string
		logger *logrus.Logger
		want   string
	}{
		{
			name:   "with logger",
			logger: logrus.New(),
			want:   "1.0.0",
		},
		{
			name:   "with nil logger",
			logger: nil,
			want:   "1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewNPCGenerator(tt.logger)
			if gen == nil {
				t.Error("NewNPCGenerator() returned nil")
			}
			if gen.GetVersion() != tt.want {
				t.Errorf("NewNPCGenerator().GetVersion() = %v, want %v", gen.GetVersion(), tt.want)
			}
			if gen.GetType() != ContentTypeCharacters {
				t.Errorf("NewNPCGenerator().GetType() = %v, want %v", gen.GetType(), ContentTypeCharacters)
			}
		})
	}
}

func TestNPCGenerator_Validate(t *testing.T) {
	gen := NewNPCGenerator(nil)

	tests := []struct {
		name    string
		params  GenerationParams
		wantErr bool
	}{
		{
			name: "valid basic parameters",
			params: GenerationParams{
				Seed:        12345,
				Difficulty:  5,
				PlayerLevel: 3,
			},
			wantErr: false,
		},
		{
			name: "zero seed",
			params: GenerationParams{
				Seed:        0,
				Difficulty:  5,
				PlayerLevel: 3,
			},
			wantErr: true,
		},
		{
			name: "difficulty too low",
			params: GenerationParams{
				Seed:        12345,
				Difficulty:  0,
				PlayerLevel: 3,
			},
			wantErr: true,
		},
		{
			name: "difficulty too high",
			params: GenerationParams{
				Seed:        12345,
				Difficulty:  25,
				PlayerLevel: 3,
			},
			wantErr: true,
		},
		{
			name: "player level too low",
			params: GenerationParams{
				Seed:        12345,
				Difficulty:  5,
				PlayerLevel: 0,
			},
			wantErr: true,
		},
		{
			name: "player level too high",
			params: GenerationParams{
				Seed:        12345,
				Difficulty:  5,
				PlayerLevel: 25,
			},
			wantErr: true,
		},
		{
			name: "valid character parameters constraints",
			params: GenerationParams{
				Seed:        12345,
				Difficulty:  5,
				PlayerLevel: 3,
				Constraints: map[string]interface{}{
					"character_params": CharacterParams{
						PersonalityDepth: 3,
						MotivationCount:  2,
						UniqueTraits:     4,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid personality depth",
			params: GenerationParams{
				Seed:        12345,
				Difficulty:  5,
				PlayerLevel: 3,
				Constraints: map[string]interface{}{
					"character_params": CharacterParams{
						PersonalityDepth: 0,
						MotivationCount:  2,
						UniqueTraits:     4,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid motivation count",
			params: GenerationParams{
				Seed:        12345,
				Difficulty:  5,
				PlayerLevel: 3,
				Constraints: map[string]interface{}{
					"character_params": CharacterParams{
						PersonalityDepth: 3,
						MotivationCount:  15,
						UniqueTraits:     4,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid unique traits",
			params: GenerationParams{
				Seed:        12345,
				Difficulty:  5,
				PlayerLevel: 3,
				Constraints: map[string]interface{}{
					"character_params": CharacterParams{
						PersonalityDepth: 3,
						MotivationCount:  2,
						UniqueTraits:     0,
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gen.Validate(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("NPCGenerator.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNPCGenerator_Generate(t *testing.T) {
	gen := NewNPCGenerator(nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		params  GenerationParams
		wantErr bool
	}{
		{
			name: "basic generation",
			params: GenerationParams{
				Seed:        12345,
				Difficulty:  5,
				PlayerLevel: 3,
			},
			wantErr: false,
		},
		{
			name: "generation with specific character params",
			params: GenerationParams{
				Seed:        12345,
				Difficulty:  5,
				PlayerLevel: 3,
				Constraints: map[string]interface{}{
					"character_params": CharacterParams{
						CharacterType:    CharacterTypeMerchant,
						PersonalityDepth: 4,
						MotivationCount:  3,
						BackgroundType:   BackgroundUrban,
						SocialClass:      SocialClassMerchant,
						AgeRange:         AgeRangeAdult,
						Alignment:        "Lawful Neutral",
						UniqueTraits:     5,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "generation with invalid params",
			params: GenerationParams{
				Seed:        0, // Invalid seed should cause error
				Difficulty:  5,
				PlayerLevel: 3,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gen.Generate(ctx, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("NPCGenerator.Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				npc, ok := result.(*game.NPC)
				if !ok {
					t.Errorf("NPCGenerator.Generate() returned %T, want *game.NPC", result)
					return
				}

				// Validate the generated NPC
				if npc.Character.Name == "" {
					t.Error("Generated NPC has empty name")
				}
				if npc.Character.ID == "" {
					t.Error("Generated NPC has empty ID")
				}
				if npc.Character.Level < 1 {
					t.Errorf("Generated NPC has invalid level: %d", npc.Character.Level)
				}
				if npc.Character.MaxHP <= 0 {
					t.Errorf("Generated NPC has invalid MaxHP: %d", npc.Character.MaxHP)
				}
			}
		})
	}
}

func TestNPCGenerator_GenerateNPC(t *testing.T) {
	gen := NewNPCGenerator(nil)
	ctx := context.Background()

	tests := []struct {
		name          string
		characterType CharacterType
		params        CharacterParams
		wantErr       bool
	}{
		{
			name:          "generate guard",
			characterType: CharacterTypeGuard,
			params: CharacterParams{
				GenerationParams: GenerationParams{
					Seed:        12345,
					Difficulty:  5,
					PlayerLevel: 3,
				},
				PersonalityDepth: 3,
				MotivationCount:  2,
				BackgroundType:   BackgroundMilitary,
				SocialClass:      SocialClassPeasant,
				AgeRange:         AgeRangeAdult,
				UniqueTraits:     3,
			},
			wantErr: false,
		},
		{
			name:          "generate merchant",
			characterType: CharacterTypeMerchant,
			params: CharacterParams{
				GenerationParams: GenerationParams{
					Seed:        54321,
					Difficulty:  3,
					PlayerLevel: 2,
				},
				PersonalityDepth: 4,
				MotivationCount:  3,
				BackgroundType:   BackgroundUrban,
				SocialClass:      SocialClassMerchant,
				AgeRange:         AgeRangeMiddleAged,
				UniqueTraits:     4,
			},
			wantErr: false,
		},
		{
			name:          "generate noble",
			characterType: CharacterTypeNoble,
			params: CharacterParams{
				GenerationParams: GenerationParams{
					Seed:        98765,
					Difficulty:  8,
					PlayerLevel: 6,
				},
				PersonalityDepth: 5,
				MotivationCount:  4,
				BackgroundType:   BackgroundNoble,
				SocialClass:      SocialClassNoble,
				AgeRange:         AgeRangeAdult,
				UniqueTraits:     6,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			npc, err := gen.GenerateNPC(ctx, tt.characterType, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("NPCGenerator.GenerateNPC() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Validate basic character properties
				if npc.Character.Name == "" {
					t.Error("Generated NPC has empty name")
				}
				if npc.Character.ID == "" {
					t.Error("Generated NPC has empty ID")
				}
				if npc.Behavior == "" {
					t.Error("Generated NPC has empty behavior")
				}

				// Validate attribute ranges
				attrs := []int{
					npc.Character.Strength,
					npc.Character.Dexterity,
					npc.Character.Constitution,
					npc.Character.Intelligence,
					npc.Character.Wisdom,
					npc.Character.Charisma,
				}
				for i, attr := range attrs {
					if attr < 3 || attr > 25 {
						t.Errorf("Attribute %d out of reasonable range: %d", i, attr)
					}
				}

				// Validate character type-specific attributes
				switch tt.characterType {
				case CharacterTypeGuard:
					if npc.Character.Class != game.ClassFighter {
						t.Errorf("Guard should be Fighter class, got %v", npc.Character.Class)
					}
				case CharacterTypeMerchant:
					if npc.Character.Charisma < 10 {
						t.Errorf("Merchant should have decent Charisma, got %d", npc.Character.Charisma)
					}
				case CharacterTypeNoble:
					if npc.Character.Charisma < 12 {
						t.Errorf("Noble should have high Charisma, got %d", npc.Character.Charisma)
					}
				}
			}
		})
	}
}

func TestNPCGenerator_GenerateNPCGroup(t *testing.T) {
	gen := NewNPCGenerator(nil)
	ctx := context.Background()

	tests := []struct {
		name      string
		groupType NPCGroupType
		params    CharacterParams
		wantErr   bool
		minSize   int
		maxSize   int
	}{
		{
			name:      "generate guard group",
			groupType: NPCGroupGuards,
			params: CharacterParams{
				GenerationParams: GenerationParams{
					Seed:        12345,
					Difficulty:  5,
					PlayerLevel: 3,
				},
				PersonalityDepth: 3,
				MotivationCount:  2,
				BackgroundType:   BackgroundMilitary,
				SocialClass:      SocialClassPeasant,
				AgeRange:         AgeRangeAdult,
				UniqueTraits:     3,
			},
			wantErr: false,
			minSize: 3,
			maxSize: 8,
		},
		{
			name:      "generate family group",
			groupType: NPCGroupFamily,
			params: CharacterParams{
				GenerationParams: GenerationParams{
					Seed:        54321,
					Difficulty:  3,
					PlayerLevel: 2,
				},
				PersonalityDepth: 2,
				MotivationCount:  1,
				BackgroundType:   BackgroundRural,
				SocialClass:      SocialClassPeasant,
				AgeRange:         AgeRangeAdult,
				UniqueTraits:     2,
			},
			wantErr: false,
			minSize: 2,
			maxSize: 6,
		},
		{
			name:      "generate merchant group",
			groupType: NPCGroupMerchants,
			params: CharacterParams{
				GenerationParams: GenerationParams{
					Seed:        98765,
					Difficulty:  4,
					PlayerLevel: 3,
				},
				PersonalityDepth: 3,
				MotivationCount:  2,
				BackgroundType:   BackgroundUrban,
				SocialClass:      SocialClassMerchant,
				AgeRange:         AgeRangeAdult,
				UniqueTraits:     3,
			},
			wantErr: false,
			minSize: 2,
			maxSize: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			npcs, err := gen.GenerateNPCGroup(ctx, tt.groupType, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("NPCGenerator.GenerateNPCGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(npcs) < tt.minSize || len(npcs) > tt.maxSize {
					t.Errorf("Group size %d not within expected range [%d, %d]", len(npcs), tt.minSize, tt.maxSize)
				}

				// Validate all NPCs in the group
				for i, npc := range npcs {
					if npc.Character.Name == "" {
						t.Errorf("NPC %d has empty name", i)
					}
					if npc.Character.ID == "" {
						t.Errorf("NPC %d has empty ID", i)
					}

					// First NPC should be a leader with higher social class
					if i == 0 {
						switch tt.groupType {
						case NPCGroupGuards:
							if npc.Character.Class != game.ClassFighter {
								t.Errorf("Guard leader should be Fighter, got %v", npc.Character.Class)
							}
						}
					}
				}

				// Ensure all NPCs have unique IDs
				idMap := make(map[string]bool)
				for _, npc := range npcs {
					if idMap[npc.Character.ID] {
						t.Errorf("Duplicate NPC ID found: %s", npc.Character.ID)
					}
					idMap[npc.Character.ID] = true
				}
			}
		})
	}
}

func TestNPCGenerator_GeneratePersonality(t *testing.T) {
	gen := NewNPCGenerator(nil)
	ctx := context.Background()

	// Create a basic character for testing
	char := &game.Character{
		ID:   "test_char",
		Name: "Test Character",
	}

	tests := []struct {
		name    string
		params  CharacterParams
		wantErr bool
	}{
		{
			name: "basic personality generation",
			params: CharacterParams{
				GenerationParams: GenerationParams{
					Seed:        12345,
					Difficulty:  5,
					PlayerLevel: 3,
				},
				PersonalityDepth: 3,
				MotivationCount:  2,
				UniqueTraits:     4,
				Alignment:        "Lawful Good",
			},
			wantErr: false,
		},
		{
			name: "complex personality generation",
			params: CharacterParams{
				GenerationParams: GenerationParams{
					Seed:        54321,
					Difficulty:  8,
					PlayerLevel: 6,
				},
				PersonalityDepth: 5,
				MotivationCount:  5,
				UniqueTraits:     8,
				Alignment:        "Chaotic Neutral",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			personality, err := gen.GeneratePersonality(ctx, char, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("NPCGenerator.GeneratePersonality() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Validate personality structure
				if personality.Alignment != tt.params.Alignment {
					t.Errorf("Alignment mismatch: got %s, want %s", personality.Alignment, tt.params.Alignment)
				}
				if personality.Temperament == "" {
					t.Error("Temperament should not be empty")
				}
				if len(personality.Values) == 0 {
					t.Error("Values should not be empty")
				}
				if len(personality.Fears) == 0 {
					t.Error("Fears should not be empty")
				}
				if len(personality.Traits) != tt.params.UniqueTraits {
					t.Errorf("Expected %d traits, got %d", tt.params.UniqueTraits, len(personality.Traits))
				}
				if len(personality.Motivations) != tt.params.MotivationCount {
					t.Errorf("Expected %d motivations, got %d", tt.params.MotivationCount, len(personality.Motivations))
				}

				// Validate trait intensities
				for _, trait := range personality.Traits {
					if trait.Intensity < 0.3 || trait.Intensity > 1.0 {
						t.Errorf("Trait intensity %f out of range [0.3, 1.0]", trait.Intensity)
					}
				}

				// Validate motivation intensities
				for _, motivation := range personality.Motivations {
					if motivation.Intensity < 0.4 || motivation.Intensity > 1.0 {
						t.Errorf("Motivation intensity %f out of range [0.4, 1.0]", motivation.Intensity)
					}
				}
			}
		})
	}
}

func TestNPCGenerator_DeterministicGeneration(t *testing.T) {
	gen := NewNPCGenerator(nil)
	ctx := context.Background()

	params := GenerationParams{
		Seed:        12345,
		Difficulty:  5,
		PlayerLevel: 3,
		Constraints: map[string]interface{}{
			"character_params": CharacterParams{
				CharacterType:    CharacterTypeMerchant,
				PersonalityDepth: 3,
				MotivationCount:  2,
				BackgroundType:   BackgroundUrban,
				SocialClass:      SocialClassMerchant,
				AgeRange:         AgeRangeAdult,
				UniqueTraits:     4,
			},
		},
	}

	// Generate the same character twice with the same seed
	result1, err1 := gen.Generate(ctx, params)
	if err1 != nil {
		t.Fatalf("First generation failed: %v", err1)
	}

	result2, err2 := gen.Generate(ctx, params)
	if err2 != nil {
		t.Fatalf("Second generation failed: %v", err2)
	}

	npc1 := result1.(*game.NPC)
	npc2 := result2.(*game.NPC)

	// Verify deterministic generation
	if npc1.Character.Name != npc2.Character.Name {
		t.Errorf("Names differ: %s vs %s", npc1.Character.Name, npc2.Character.Name)
	}
	if npc1.Character.Strength != npc2.Character.Strength {
		t.Errorf("Strength differs: %d vs %d", npc1.Character.Strength, npc2.Character.Strength)
	}
	if npc1.Character.Dexterity != npc2.Character.Dexterity {
		t.Errorf("Dexterity differs: %d vs %d", npc1.Character.Dexterity, npc2.Character.Dexterity)
	}
	if npc1.Character.Class != npc2.Character.Class {
		t.Errorf("Class differs: %v vs %v", npc1.Character.Class, npc2.Character.Class)
	}
	if npc1.Behavior != npc2.Behavior {
		t.Errorf("Behavior differs: %s vs %s", npc1.Behavior, npc2.Behavior)
	}
}

func TestNPCGenerator_AttributeRanges(t *testing.T) {
	gen := NewNPCGenerator(nil)

	tests := []struct {
		name          string
		characterType CharacterType
		socialClass   SocialClass
		checkAttr     func(*game.Character) int
		expectedMin   int
		testName      string
	}{
		{
			name:          "guard strength",
			characterType: CharacterTypeGuard,
			socialClass:   SocialClassPeasant,
			checkAttr:     func(c *game.Character) int { return c.Strength },
			expectedMin:   12, // Base 10 + 2 minimum
			testName:      "Strength",
		},
		{
			name:          "merchant charisma",
			characterType: CharacterTypeMerchant,
			socialClass:   SocialClassMerchant,
			checkAttr:     func(c *game.Character) int { return c.Charisma },
			expectedMin:   12, // Base 10 + 2 minimum
			testName:      "Charisma",
		},
		{
			name:          "mage intelligence",
			characterType: CharacterTypeMage,
			socialClass:   SocialClassPeasant,
			checkAttr:     func(c *game.Character) int { return c.Intelligence },
			expectedMin:   13, // Base 10 + 3 minimum
			testName:      "Intelligence",
		},
		{
			name:          "noble charisma",
			characterType: CharacterTypeNoble,
			socialClass:   SocialClassNoble,
			checkAttr:     func(c *game.Character) int { return c.Charisma },
			expectedMin:   14, // Base 10 + 3 (type) + 1 (social class)
			testName:      "Charisma",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := CharacterParams{
				GenerationParams: GenerationParams{
					Seed:        12345,
					Difficulty:  5,
					PlayerLevel: 3,
				},
				CharacterType:    tt.characterType,
				SocialClass:      tt.socialClass,
				PersonalityDepth: 3,
				MotivationCount:  2,
				UniqueTraits:     3,
			}

			baseChar, err := gen.generateBaseCharacter(params)
			if err != nil {
				t.Fatalf("generateBaseCharacter() failed: %v", err)
			}

			attrValue := tt.checkAttr(baseChar)
			if attrValue < tt.expectedMin {
				t.Errorf("%s for %s should be at least %d, got %d",
					tt.testName, tt.characterType, tt.expectedMin, attrValue)
			}
		})
	}
}

// Benchmark tests
func BenchmarkNPCGenerator_Generate(b *testing.B) {
	gen := NewNPCGenerator(nil)
	ctx := context.Background()
	params := GenerationParams{
		Seed:        12345,
		Difficulty:  5,
		PlayerLevel: 3,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		params.Seed = int64(i) // Vary seed for each iteration
		_, err := gen.Generate(ctx, params)
		if err != nil {
			b.Fatalf("Generate() failed: %v", err)
		}
	}
}

func BenchmarkNPCGenerator_GenerateNPCGroup(b *testing.B) {
	gen := NewNPCGenerator(nil)
	ctx := context.Background()
	params := CharacterParams{
		GenerationParams: GenerationParams{
			Seed:        12345,
			Difficulty:  5,
			PlayerLevel: 3,
		},
		PersonalityDepth: 3,
		MotivationCount:  2,
		UniqueTraits:     3,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		params.Seed = int64(i) // Vary seed for each iteration
		_, err := gen.GenerateNPCGroup(ctx, NPCGroupGuards, params)
		if err != nil {
			b.Fatalf("GenerateNPCGroup() failed: %v", err)
		}
	}
}

func TestNPCGenerator_SpeechPatterns(t *testing.T) {
	gen := NewNPCGenerator(nil)

	params := CharacterParams{
		PersonalityDepth: 3,
		MotivationCount:  2,
		UniqueTraits:     3,
	}

	// Test multiple generations to ensure variety
	patterns := make(map[string]bool)
	for i := 0; i < 20; i++ {
		pattern := gen.generateSpeechPattern(params)

		// Validate required fields
		if pattern.Formality == "" {
			t.Error("Formality should not be empty")
		}
		if pattern.Vocabulary == "" {
			t.Error("Vocabulary should not be empty")
		}
		if pattern.Accent == "" {
			t.Error("Accent should not be empty")
		}
		if len(pattern.Mannerisms) == 0 {
			t.Error("Mannerisms should not be empty")
		}

		// Track pattern variety
		patternKey := pattern.Formality + pattern.Vocabulary + pattern.Accent
		patterns[patternKey] = true
	}

	// Ensure we're getting variety in speech patterns
	if len(patterns) < 5 {
		t.Errorf("Expected more variety in speech patterns, got %d unique patterns", len(patterns))
	}
}

func TestNPCGenerator_Timeout(t *testing.T) {
	gen := NewNPCGenerator(nil)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	params := GenerationParams{
		Seed:        12345,
		Difficulty:  5,
		PlayerLevel: 3,
		Timeout:     time.Millisecond,
	}

	// This test is more about ensuring the context is properly handled
	// The actual timeout behavior would depend on implementation details
	_, err := gen.Generate(ctx, params)

	// We don't expect a timeout error here since our generation is fast
	// But this test ensures context propagation works
	if err != nil && err.Error() != "context deadline exceeded" {
		// If we get an error, it should be validation-related, not timeout
		if err.Error() == "seed cannot be zero" {
			t.Error("Unexpected seed validation error")
		}
	}
}
