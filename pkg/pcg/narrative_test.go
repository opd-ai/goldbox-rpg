package pcg

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNarrativeGenerator(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel) // Suppress logs during testing

	ng := NewNarrativeGenerator(logger)

	assert.NotNil(t, ng)
	assert.Equal(t, "1.0.0", ng.version)
	assert.NotNil(t, ng.logger)
	assert.NotNil(t, ng.rng)
	assert.NotEmpty(t, ng.storyArchetypes)
	assert.NotEmpty(t, ng.narrativeThemes)
	assert.NotEmpty(t, ng.characterArchetypes)
}

func TestNarrativeGenerator_Generate(t *testing.T) {
	tests := []struct {
		name           string
		params         GenerationParams
		expectError    bool
		expectedFields []string
	}{
		{
			name: "successful classic narrative generation",
			params: GenerationParams{
				Seed:        12345,
				Difficulty:  5,
				PlayerLevel: 3,
				Constraints: map[string]interface{}{
					"narrative_params": NarrativeParams{
						NarrativeType:   NarrativeLinear,
						Theme:           "classic",
						CampaignLength:  "medium",
						ComplexityLevel: 3,
						CharacterFocus:  true,
						MainAntagonist:  "dark_lord",
						ConflictType:    "good_vs_evil",
						TonePreference:  "heroic",
					},
				},
			},
			expectError:    false,
			expectedFields: []string{"ID", "Title", "Theme", "MainPlotline", "NPCs", "KeyLocations"},
		},
		{
			name: "successful grimdark narrative generation",
			params: GenerationParams{
				Seed:        54321,
				Difficulty:  8,
				PlayerLevel: 5,
				Constraints: map[string]interface{}{
					"narrative_params": NarrativeParams{
						NarrativeType:   NarrativeBranching,
						Theme:           "grimdark",
						CampaignLength:  "long",
						ComplexityLevel: 5,
						CharacterFocus:  false,
						MainAntagonist:  "corrupt_noble",
						ConflictType:    "political",
						TonePreference:  "dark",
					},
				},
			},
			expectError:    false,
			expectedFields: []string{"ID", "Title", "Theme", "MainPlotline", "NPCs", "KeyLocations"},
		},
		{
			name: "invalid parameters - missing narrative_params",
			params: GenerationParams{
				Seed:        12345,
				Difficulty:  5,
				PlayerLevel: 3,
				Constraints: map[string]interface{}{},
			},
			expectError: true,
		},
		{
			name: "invalid parameters - empty theme",
			params: GenerationParams{
				Seed:        12345,
				Difficulty:  5,
				PlayerLevel: 3,
				Constraints: map[string]interface{}{
					"narrative_params": NarrativeParams{
						NarrativeType:  NarrativeLinear,
						Theme:          "", // Invalid empty theme
						CampaignLength: "medium",
					},
				},
			},
			expectError: true,
		},
	}

	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	ng := NewNarrativeGenerator(logger)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := ng.Generate(ctx, tt.params)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				narrative, ok := result.(*CampaignNarrative)
				require.True(t, ok, "Result should be a CampaignNarrative")

				// Check that all expected fields are present and non-empty
				for _, field := range tt.expectedFields {
					switch field {
					case "ID":
						assert.NotEmpty(t, narrative.ID)
					case "Title":
						assert.NotEmpty(t, narrative.Title)
					case "Theme":
						assert.NotEmpty(t, narrative.Theme)
					case "MainPlotline":
						assert.NotNil(t, narrative.MainPlotline)
						assert.NotEmpty(t, narrative.MainPlotline.ID)
					case "NPCs":
						assert.NotEmpty(t, narrative.NPCs)
					case "KeyLocations":
						assert.NotEmpty(t, narrative.KeyLocations)
					}
				}

				// Verify metadata
				assert.NotEmpty(t, narrative.Metadata)
				assert.Contains(t, narrative.Metadata, "character_count")
				assert.Contains(t, narrative.Metadata, "location_count")
				assert.Contains(t, narrative.Metadata, "subplot_count")
			}
		})
	}
}

func TestNarrativeGenerator_Validate(t *testing.T) {
	tests := []struct {
		name        string
		params      GenerationParams
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid parameters",
			params: GenerationParams{
				Constraints: map[string]interface{}{
					"narrative_params": NarrativeParams{
						Theme:          "classic",
						CampaignLength: "medium",
					},
				},
			},
			expectError: false,
		},
		{
			name: "missing narrative_params",
			params: GenerationParams{
				Constraints: map[string]interface{}{},
			},
			expectError: true,
			errorMsg:    "expected narrative_params",
		},
		{
			name: "invalid campaign length",
			params: GenerationParams{
				Constraints: map[string]interface{}{
					"narrative_params": NarrativeParams{
						Theme:          "classic",
						CampaignLength: "invalid", // Invalid length
					},
				},
			},
			expectError: true,
			errorMsg:    "campaign length must be one of",
		},
		{
			name: "empty theme",
			params: GenerationParams{
				Constraints: map[string]interface{}{
					"narrative_params": NarrativeParams{
						Theme:          "", // Empty theme
						CampaignLength: "medium",
					},
				},
			},
			expectError: true,
			errorMsg:    "theme cannot be empty",
		},
	}

	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	ng := NewNarrativeGenerator(logger)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ng.Validate(tt.params)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNarrativeGenerator_GetType(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	ng := NewNarrativeGenerator(logger)

	contentType := ng.GetType()
	assert.Equal(t, ContentTypeNarrative, contentType)
}

func TestNarrativeGenerator_GetVersion(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	ng := NewNarrativeGenerator(logger)

	version := ng.GetVersion()
	assert.Equal(t, "1.0.0", version)
}

func TestNarrativeGenerator_DeterministicGeneration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	ng1 := NewNarrativeGenerator(logger)
	ng2 := NewNarrativeGenerator(logger)

	params := GenerationParams{
		Seed:        42,
		Difficulty:  5,
		PlayerLevel: 3,
		Constraints: map[string]interface{}{
			"narrative_params": NarrativeParams{
				NarrativeType:   NarrativeLinear,
				Theme:           "classic",
				CampaignLength:  "short",
				ComplexityLevel: 2,
			},
		},
	}

	ctx := context.Background()

	// Generate with both generators using the same seed
	result1, err1 := ng1.Generate(ctx, params)
	require.NoError(t, err1)

	result2, err2 := ng2.Generate(ctx, params)
	require.NoError(t, err2)

	narrative1, ok := result1.(*CampaignNarrative)
	require.True(t, ok)

	narrative2, ok := result2.(*CampaignNarrative)
	require.True(t, ok)

	// With the same seed, key elements should be identical
	assert.Equal(t, narrative1.Title, narrative2.Title)
	assert.Equal(t, narrative1.Theme, narrative2.Theme)
	assert.Equal(t, narrative1.Setting, narrative2.Setting)
	assert.Equal(t, len(narrative1.NPCs), len(narrative2.NPCs))
	assert.Equal(t, len(narrative1.KeyLocations), len(narrative2.KeyLocations))
}

func TestNarrativeGenerator_CharacterGeneration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	ng := NewNarrativeGenerator(logger)

	params := NarrativeParams{
		NarrativeType:   NarrativeLinear,
		Theme:           "classic",
		CampaignLength:  "medium",
		ComplexityLevel: 3,
		CharacterFocus:  true,
	}

	// Test different character roles
	roles := []CharacterRole{RoleProtagonist, RoleAntagonist, RoleAlly, RoleMentor}

	for _, role := range roles {
		t.Run(string(role), func(t *testing.T) {
			character := ng.generateCharacter(role, params)

			assert.NotEmpty(t, character.ID)
			assert.NotEmpty(t, character.Name)
			assert.NotEmpty(t, character.Archetype)
			assert.Equal(t, role, character.Role)
			assert.NotEmpty(t, character.Motivation)
			assert.NotEmpty(t, character.Background)
			assert.NotEmpty(t, character.Personality)
			assert.NotNil(t, character.Arc)
			assert.NotNil(t, character.Relationships)
			assert.NotNil(t, character.Properties)
		})
	}
}

func TestNarrativeGenerator_LocationGeneration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	ng := NewNarrativeGenerator(logger)

	params := NarrativeParams{
		Theme:          "classic",
		CampaignLength: "medium",
	}

	plotline := &Plotline{ID: "test_plot"}
	locations := ng.generateKeyLocations(plotline, params)

	assert.NotEmpty(t, locations)

	for _, location := range locations {
		assert.NotEmpty(t, location.ID)
		assert.NotEmpty(t, location.Name)
		assert.NotEmpty(t, location.Type)
		assert.NotEmpty(t, location.Description)
		assert.NotEmpty(t, location.Significance)
		assert.NotEmpty(t, location.History)
		assert.NotNil(t, location.Properties)
	}

	// Test different location types are generated
	locationTypes := make(map[LocationType]bool)
	for _, location := range locations {
		locationTypes[location.Type] = true
	}
	assert.True(t, len(locationTypes) > 1, "Should generate multiple location types")
}

func TestNarrativeGenerator_PlotlineGeneration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	ng := NewNarrativeGenerator(logger)

	archetype := ng.storyArchetypes["hero_journey"]
	theme := ng.narrativeThemes["classic"]
	params := NarrativeParams{
		CampaignLength: "medium",
		Theme:          "classic",
	}

	plotline, err := ng.generateMainPlotline(archetype, theme, params)
	require.NoError(t, err)

	assert.Equal(t, "main_plot", plotline.ID)
	assert.Equal(t, PlotTypeMain, plotline.Type)
	assert.NotEmpty(t, plotline.Title)
	assert.NotEmpty(t, plotline.Acts)
	assert.NotEmpty(t, plotline.Hooks)
	assert.NotEmpty(t, plotline.Climax)
	assert.NotEmpty(t, plotline.Resolution)

	// Check acts are properly structured
	for i, act := range plotline.Acts {
		assert.NotEmpty(t, act.ID)
		assert.NotEmpty(t, act.Title)
		assert.NotEmpty(t, act.Description)
		assert.Contains(t, act.ID, fmt.Sprintf("act_%d", i+1))
	}
}

func TestNarrativeGenerator_CampaignLengthScaling(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	ng := NewNarrativeGenerator(logger)

	tests := []struct {
		length        string
		minSubplots   int
		maxSubplots   int
		expectedActs  int
		minCharacters int
	}{
		{"short", 1, 2, 3, 2},
		{"medium", 2, 4, 5, 3},
		{"long", 3, 6, 7, 4},
	}

	for _, tt := range tests {
		t.Run(tt.length, func(t *testing.T) {
			subplotCount := ng.calculateSubplotCount(tt.length)
			assert.GreaterOrEqual(t, subplotCount, tt.minSubplots)
			assert.LessOrEqual(t, subplotCount, tt.maxSubplots)

			actCount := ng.calculateActCount(tt.length)
			assert.Equal(t, tt.expectedActs, actCount)

			charCount := ng.calculateSupportingCharacterCount(tt.length)
			assert.GreaterOrEqual(t, charCount, tt.minCharacters)
		})
	}
}

func TestNarrativeGenerator_TemplateInitialization(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	ng := NewNarrativeGenerator(logger)

	// Test story archetypes
	assert.Contains(t, ng.storyArchetypes, "hero_journey")
	assert.Contains(t, ng.storyArchetypes, "tragedy")
	assert.Contains(t, ng.storyArchetypes, "mystery")

	heroJourney := ng.storyArchetypes["hero_journey"]
	assert.Equal(t, "Hero's Journey", heroJourney.Name)
	assert.NotEmpty(t, heroJourney.Structure)
	assert.NotEmpty(t, heroJourney.Themes)
	assert.NotEmpty(t, heroJourney.Conflicts)

	// Test narrative themes
	assert.Contains(t, ng.narrativeThemes, "classic")
	assert.Contains(t, ng.narrativeThemes, "grimdark")

	classic := ng.narrativeThemes["classic"]
	assert.Equal(t, "Classic Fantasy", classic.Name)
	assert.NotEmpty(t, classic.Motifs)
	assert.NotEmpty(t, classic.Symbols)
	assert.NotEmpty(t, classic.Messages)

	// Test character archetypes
	assert.Contains(t, ng.characterArchetypes, "noble_hero")
	assert.Contains(t, ng.characterArchetypes, "dark_lord")

	hero := ng.characterArchetypes["noble_hero"]
	assert.Equal(t, "Noble Hero", hero.Name)
	assert.NotEmpty(t, hero.Motivations)
	assert.NotEmpty(t, hero.Traits)
	assert.NotEmpty(t, hero.SpeechStyle)
}

// Benchmark tests
func BenchmarkNarrativeGenerator_Generate(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	ng := NewNarrativeGenerator(logger)

	params := GenerationParams{
		Seed:        12345,
		Difficulty:  5,
		PlayerLevel: 3,
		Constraints: map[string]interface{}{
			"narrative_params": NarrativeParams{
				NarrativeType:   NarrativeLinear,
				Theme:           "classic",
				CampaignLength:  "medium",
				ComplexityLevel: 3,
			},
		},
		Timeout: 30 * time.Second,
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		params.Seed = int64(i) // Vary seed for each iteration
		_, err := ng.Generate(ctx, params)
		if err != nil {
			b.Fatalf("Generation failed: %v", err)
		}
	}
}

func BenchmarkNarrativeGenerator_CharacterGeneration(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	ng := NewNarrativeGenerator(logger)

	params := NarrativeParams{
		NarrativeType:   NarrativeLinear,
		Theme:           "classic",
		CampaignLength:  "medium",
		ComplexityLevel: 3,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ng.generateCharacter(RoleProtagonist, params)
	}
}
