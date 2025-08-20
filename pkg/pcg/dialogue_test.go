package pcg

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goldbox-rpg/pkg/game"
)

func TestDialogueGenerator_NewDialogueGenerator(t *testing.T) {
	generator := NewDialogueGenerator(nil)

	assert.NotNil(t, generator)
	assert.NotNil(t, generator.rng)
	assert.NotNil(t, generator.markovChains)
	assert.NotNil(t, generator.dialogTemplates)
	assert.NotNil(t, generator.logger)
	assert.Equal(t, "1.0.0", generator.GetVersion())
	assert.Equal(t, ContentTypeDialogue, generator.GetType())
}

func TestDialogueGenerator_Generate(t *testing.T) {
	tests := []struct {
		name           string
		params         GenerationParams
		expectedError  bool
		validateResult func(t *testing.T, result interface{})
	}{
		{
			name: "BasicDialogueGeneration_Success",
			params: GenerationParams{
				Seed:       12345,
				Difficulty: 5,
				Constraints: map[string]interface{}{
					"npc_id": "test_npc",
					"dialogue_params": DialogParams{
						DialogType:       DialogTypeGreeting,
						Tone:             DialogToneFriendly,
						MaxDepth:         2,
						ResponseCount:    3,
						UseMarkov:        false,
						MarkovChainOrder: 2,
						Personality: PersonalityProfile{
							Traits: []PersonalityTrait{
								{Name: "friendly", Intensity: 0.8, Description: "Very friendly"},
							},
							Speech: SpeechPattern{
								Formality:  "casual",
								Vocabulary: "moderate",
							},
						},
						Context: DialogContext{
							Location:  "tavern",
							TimeOfDay: "evening",
						},
					},
				},
			},
			expectedError: false,
			validateResult: func(t *testing.T, result interface{}) {
				dialogue, ok := result.(*GeneratedDialogue)
				require.True(t, ok)
				assert.NotNil(t, dialogue.RootEntry)
				assert.Greater(t, dialogue.TotalNodes, 0)
				assert.LessOrEqual(t, dialogue.MaxDepth, 2)
				assert.False(t, dialogue.MarkovUsed)
			},
		},
		{
			name: "QuestDialogueGeneration_Success",
			params: GenerationParams{
				Seed:       67890,
				Difficulty: 3,
				Constraints: map[string]interface{}{
					"dialogue_params": DialogParams{
						DialogType:       DialogTypeQuest,
						Tone:             DialogToneFormal,
						MaxDepth:         3,
						ResponseCount:    4,
						UseMarkov:        true,
						MarkovChainOrder: 2,
						Personality: PersonalityProfile{
							Traits: []PersonalityTrait{
								{Name: "formal", Intensity: 0.9, Description: "Very formal"},
								{Name: "mysterious", Intensity: 0.6, Description: "Somewhat mysterious"},
							},
							Speech: SpeechPattern{
								Formality:   "formal",
								Vocabulary:  "complex",
								Catchphrase: "Indeed.",
							},
						},
						Context: DialogContext{
							Location:     "castle",
							TimeOfDay:    "morning",
							ActiveQuests: []string{"main_quest_1"},
						},
					},
				},
			},
			expectedError: false,
			validateResult: func(t *testing.T, result interface{}) {
				dialogue, ok := result.(*GeneratedDialogue)
				require.True(t, ok)
				assert.NotNil(t, dialogue.RootEntry)
				assert.Greater(t, len(dialogue.AllEntries), 1)
				assert.LessOrEqual(t, dialogue.MaxDepth, 3)

				// Check that dialogue contains quest-related content
				found := false
				for _, entry := range dialogue.AllEntries {
					if strings.Contains(strings.ToLower(entry.Text), "task") ||
						strings.Contains(strings.ToLower(entry.Text), "capabilities") {
						found = true
						break
					}
				}
				assert.True(t, found, "Quest dialogue should contain task-related terms")
			},
		},
		{
			name: "MarkovEnhancedDialogue_Success",
			params: GenerationParams{
				Seed: 42,
				Constraints: map[string]interface{}{
					"dialogue_params": DialogParams{
						DialogType:       DialogTypeInformation,
						Tone:             DialogToneMysterious,
						MaxDepth:         2,
						ResponseCount:    2,
						UseMarkov:        true,
						MarkovChainOrder: 2,
						Personality: PersonalityProfile{
							Traits: []PersonalityTrait{
								{Name: "mysterious", Intensity: 0.8, Description: "Very mysterious"},
							},
							Speech: SpeechPattern{
								Formality:  "formal",
								Vocabulary: "complex",
							},
						},
					},
				},
			},
			expectedError: false,
			validateResult: func(t *testing.T, result interface{}) {
				dialogue, ok := result.(*GeneratedDialogue)
				require.True(t, ok)
				assert.NotNil(t, dialogue.RootEntry)
				assert.Greater(t, len(dialogue.AllEntries), 0)
			},
		},
		{
			name: "DefaultParameters_Success",
			params: GenerationParams{
				Seed: 1234,
				Constraints: map[string]interface{}{
					"npc_id": "default_npc",
				},
			},
			expectedError: false,
			validateResult: func(t *testing.T, result interface{}) {
				dialogue, ok := result.(*GeneratedDialogue)
				require.True(t, ok)
				assert.NotNil(t, dialogue.RootEntry)
				assert.Greater(t, len(dialogue.AllEntries), 0)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			generator := NewDialogueGenerator(logrus.New())

			result, err := generator.Generate(context.Background(), tc.params)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				tc.validateResult(t, result)
			}
		})
	}
}

func TestDialogueGenerator_Validate(t *testing.T) {
	tests := []struct {
		name          string
		params        GenerationParams
		expectedError bool
		errorContains string
	}{
		{
			name: "ValidParameters_Success",
			params: GenerationParams{
				Seed: 12345,
				Constraints: map[string]interface{}{
					"dialogue_params": DialogParams{
						MaxDepth:         3,
						ResponseCount:    4,
						MarkovChainOrder: 2,
					},
				},
			},
			expectedError: false,
		},
		{
			name: "NegativeSeed_Error",
			params: GenerationParams{
				Seed: -1,
			},
			expectedError: true,
			errorContains: "seed must be non-negative",
		},
		{
			name: "MaxDepthTooLarge_Error",
			params: GenerationParams{
				Seed: 12345,
				Constraints: map[string]interface{}{
					"dialogue_params": DialogParams{
						MaxDepth: 15,
					},
				},
			},
			expectedError: true,
			errorContains: "max_depth must be between 1 and 10",
		},
		{
			name: "ResponseCountTooHigh_Error",
			params: GenerationParams{
				Seed: 12345,
				Constraints: map[string]interface{}{
					"dialogue_params": DialogParams{
						MaxDepth:      3,
						ResponseCount: 12,
					},
				},
			},
			expectedError: true,
			errorContains: "response_count must be between 1 and 8",
		},
		{
			name: "MarkovOrderInvalid_Error",
			params: GenerationParams{
				Seed: 12345,
				Constraints: map[string]interface{}{
					"dialogue_params": DialogParams{
						MaxDepth:         3,
						ResponseCount:    3,
						MarkovChainOrder: 0,
					},
				},
			},
			expectedError: true,
			errorContains: "markov_chain_order must be between 1 and 4",
		},
	}

	generator := NewDialogueGenerator(nil)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := generator.Validate(tc.params)

			if tc.expectedError {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDialogueGenerator_GenerateDialogueTree(t *testing.T) {
	generator := NewDialogueGenerator(logrus.New())

	params := DialogParams{
		DialogType:       DialogTypeGreeting,
		Tone:             DialogToneFriendly,
		MaxDepth:         2,
		ResponseCount:    3,
		UseMarkov:        false,
		MarkovChainOrder: 2,
		Personality: PersonalityProfile{
			Traits: []PersonalityTrait{
				{Name: "cheerful", Intensity: 0.8, Description: "Very cheerful"},
			},
			Speech: SpeechPattern{
				Formality:   "casual",
				Catchphrase: "Cheers!",
			},
		},
		Context: DialogContext{
			Location:  "marketplace",
			TimeOfDay: "afternoon",
		},
	}

	dialogue, err := generator.generateDialogueTree(context.Background(), params)

	assert.NoError(t, err)
	assert.NotNil(t, dialogue)
	assert.NotNil(t, dialogue.RootEntry)
	assert.Greater(t, len(dialogue.AllEntries), 0)
	assert.LessOrEqual(t, dialogue.MaxDepth, params.MaxDepth)
	assert.Equal(t, len(dialogue.AllEntries), dialogue.TotalNodes)
	assert.Equal(t, params.UseMarkov, dialogue.MarkovUsed)
}

func TestDialogueGenerator_PersonalityIntegration(t *testing.T) {
	generator := NewDialogueGenerator(logrus.New())

	tests := []struct {
		name        string
		personality PersonalityProfile
		expectedKey string
	}{
		{
			name: "FriendlyPersonality_FriendlyKey",
			personality: PersonalityProfile{
				Traits: []PersonalityTrait{
					{Name: "friendly", Intensity: 0.9, Description: "Very friendly"},
				},
			},
			expectedKey: "friendly",
		},
		{
			name: "AggressivePersonality_HostileKey",
			personality: PersonalityProfile{
				Traits: []PersonalityTrait{
					{Name: "aggressive", Intensity: 0.8, Description: "Quite aggressive"},
				},
			},
			expectedKey: "hostile",
		},
		{
			name: "FormalPersonality_FormalKey",
			personality: PersonalityProfile{
				Traits: []PersonalityTrait{
					{Name: "formal", Intensity: 0.9, Description: "Very formal"},
				},
			},
			expectedKey: "formal",
		},
		{
			name: "MysteriousPersonality_MysteriousKey",
			personality: PersonalityProfile{
				Traits: []PersonalityTrait{
					{Name: "mysterious", Intensity: 0.5, Description: "Somewhat mysterious"},
				},
			},
			expectedKey: "mysterious",
		},
		{
			name: "NoSpecificTraits_CasualKey",
			personality: PersonalityProfile{
				Traits: []PersonalityTrait{
					{Name: "ordinary", Intensity: 0.5, Description: "Just ordinary"},
				},
			},
			expectedKey: "casual",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			key := generator.getPersonalityKey(tc.personality)
			assert.Equal(t, tc.expectedKey, key)
		})
	}
}

func TestDialogueGenerator_TextTransformations(t *testing.T) {
	generator := NewDialogueGenerator(logrus.New())

	tests := []struct {
		name        string
		inputText   string
		personality PersonalityProfile
		contains    string
	}{
		{
			name:      "FormalSpeech_ContracionsExpanded",
			inputText: "I'm sure you're ready.",
			personality: PersonalityProfile{
				Speech: SpeechPattern{Formality: "formal"},
			},
			contains: "I am sure you are ready",
		},
		{
			name:      "CasualSpeech_ContractionsUsed",
			inputText: "I am sure you are ready.",
			personality: PersonalityProfile{
				Speech: SpeechPattern{Formality: "crude"},
			},
			contains: "I'm sure you're ready",
		},
		{
			name:      "CheerfulTrait_ExclamationAdded",
			inputText: "Hello there",
			personality: PersonalityProfile{
				Traits: []PersonalityTrait{
					{Name: "cheerful", Intensity: 0.8, Description: "Very cheerful"},
				},
			},
			contains: "!",
		},
		{
			name:      "ArrogantTrait_ObviouslyAdded",
			inputText: "The answer is simple.",
			personality: PersonalityProfile{
				Traits: []PersonalityTrait{
					{Name: "arrogant", Intensity: 0.9, Description: "Very arrogant"},
				},
			},
			contains: "Obviously",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := generator.applyPersonalityToText(tc.inputText, tc.personality)
			assert.Contains(t, result, tc.contains)
		})
	}
}

func TestDialogueGenerator_DialogueTypeInference(t *testing.T) {
	generator := NewDialogueGenerator(logrus.New())

	tests := []struct {
		name         string
		responseText string
		expectedType DialogType
	}{
		{
			name:         "QuestResponse_QuestType",
			responseText: "I need a quest to complete.",
			expectedType: DialogTypeQuest,
		},
		{
			name:         "TradeResponse_TradeType",
			responseText: "I want to buy some supplies.",
			expectedType: DialogTypeTrade,
		},
		{
			name:         "NewsResponse_InformationType",
			responseText: "What's the latest news?",
			expectedType: DialogTypeInformation,
		},
		{
			name:         "RumorResponse_InformationType",
			responseText: "Any interesting rumors?",
			expectedType: DialogTypeInformation,
		},
		{
			name:         "GeneralResponse_InformationType",
			responseText: "Tell me more about yourself.",
			expectedType: DialogTypeInformation,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := generator.inferDialogTypeFromResponse(tc.responseText)
			assert.Equal(t, tc.expectedType, result)
		})
	}
}

func TestDialogueGenerator_TemplateSystem(t *testing.T) {
	generator := NewDialogueGenerator(logrus.New())

	// Verify that all dialog types have templates
	for dialogType := range map[DialogType]bool{
		DialogTypeGreeting:    true,
		DialogTypeQuest:       true,
		DialogTypeTrade:       true,
		DialogTypeInformation: true,
	} {
		templates, exists := generator.dialogTemplates[dialogType]
		assert.True(t, exists, "Dialog type %s should have templates", dialogType)
		assert.Greater(t, len(templates), 0, "Dialog type %s should have at least one template", dialogType)
	}

	// Test template filling
	params := DialogParams{
		NPC: &game.NPC{
			Character: game.Character{
				Name: "Bob the Merchant",
			},
		},
		PlayerCharacter: &game.Character{
			Name: "Hero",
		},
		Context: DialogContext{
			Location:  "shop",
			TimeOfDay: "morning",
		},
		Personality: PersonalityProfile{
			Speech: SpeechPattern{Formality: "casual"},
		},
	}

	templateText := "Hello {player_name}, welcome to my {location}!"
	result := generator.fillTemplate(templateText, params)

	assert.Contains(t, result, "Hero")
	assert.Contains(t, result, "shop")
	assert.NotContains(t, result, "{player_name}")
	assert.NotContains(t, result, "{location}")
}

func TestDialogueGenerator_MarkovChainInitialization(t *testing.T) {
	generator := NewDialogueGenerator(logrus.New())

	// Verify Markov chains are created for all personality types
	expectedPersonalities := []string{"friendly", "hostile", "mysterious", "formal", "casual"}

	for _, personality := range expectedPersonalities {
		chain, exists := generator.markovChains[personality]
		assert.True(t, exists, "Markov chain should exist for personality: %s", personality)
		assert.NotNil(t, chain, "Markov chain should not be nil for personality: %s", personality)
	}
}

func TestDialogueGenerator_ConcurrentGeneration(t *testing.T) {
	generator := NewDialogueGenerator(logrus.New())

	// Test concurrent dialogue generation
	const numGoroutines = 10
	results := make(chan interface{}, numGoroutines)
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(seed int64) {
			params := GenerationParams{
				Seed: seed,
				Constraints: map[string]interface{}{
					"dialogue_params": DialogParams{
						DialogType:    DialogTypeGreeting,
						Tone:          DialogToneFriendly,
						MaxDepth:      2,
						ResponseCount: 2,
						UseMarkov:     false,
					},
				},
			}

			result, err := generator.Generate(context.Background(), params)
			if err != nil {
				errors <- err
			} else {
				results <- result
			}
		}(int64(i + 1))
	}

	// Collect results
	var successCount int
	var errorCount int

	for i := 0; i < numGoroutines; i++ {
		select {
		case <-results:
			successCount++
		case err := <-errors:
			t.Logf("Generation error: %v", err)
			errorCount++
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out")
		}
	}

	assert.Equal(t, numGoroutines, successCount)
	assert.Equal(t, 0, errorCount)
}

// Benchmark tests
func BenchmarkDialogueGenerator_Generate(b *testing.B) {
	generator := NewDialogueGenerator(logrus.New())

	params := GenerationParams{
		Seed: 12345,
		Constraints: map[string]interface{}{
			"dialogue_params": DialogParams{
				DialogType:       DialogTypeGreeting,
				Tone:             DialogToneFriendly,
				MaxDepth:         3,
				ResponseCount:    3,
				UseMarkov:        false,
				MarkovChainOrder: 2,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := generator.Generate(context.Background(), params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDialogueGenerator_GenerateWithMarkov(b *testing.B) {
	generator := NewDialogueGenerator(logrus.New())

	params := GenerationParams{
		Seed: 12345,
		Constraints: map[string]interface{}{
			"dialogue_params": DialogParams{
				DialogType:       DialogTypeInformation,
				Tone:             DialogToneMysterious,
				MaxDepth:         2,
				ResponseCount:    3,
				UseMarkov:        true,
				MarkovChainOrder: 2,
				Personality: PersonalityProfile{
					Traits: []PersonalityTrait{
						{Name: "mysterious", Intensity: 0.8, Description: "Very mysterious"},
					},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := generator.Generate(context.Background(), params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDialogueGenerator_ComplexGeneration(b *testing.B) {
	generator := NewDialogueGenerator(logrus.New())

	params := GenerationParams{
		Seed: 67890,
		Constraints: map[string]interface{}{
			"dialogue_params": DialogParams{
				DialogType:       DialogTypeQuest,
				Tone:             DialogToneFormal,
				MaxDepth:         5,
				ResponseCount:    4,
				UseMarkov:        true,
				MarkovChainOrder: 3,
				Personality: PersonalityProfile{
					Traits: []PersonalityTrait{
						{Name: "formal", Intensity: 0.9, Description: "Very formal"},
						{Name: "mysterious", Intensity: 0.6, Description: "Somewhat mysterious"},
					},
					Speech: SpeechPattern{
						Formality:   "formal",
						Vocabulary:  "complex",
						Catchphrase: "Indeed.",
					},
				},
				Context: DialogContext{
					Location:        "castle",
					TimeOfDay:       "morning",
					ActiveQuests:    []string{"main_quest_1", "side_quest_2"},
					CompletedQuests: []string{"tutorial_quest"},
					Reputation:      75,
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := generator.Generate(context.Background(), params)
		if err != nil {
			b.Fatal(err)
		}
	}
}
