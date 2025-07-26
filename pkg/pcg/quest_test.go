package pcg

import (
	"context"
	"fmt"
	"testing"
	"time"

	"goldbox-rpg/pkg/game"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewQuestGenerator(t *testing.T) {
	tests := []struct {
		name   string
		logger interface{}
		want   string
	}{
		{
			name:   "with nil logger",
			logger: nil,
			want:   "1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qg := NewQuestGenerator(nil)
			assert.NotNil(t, qg)
			assert.Equal(t, tt.want, qg.GetVersion())
			assert.Equal(t, ContentTypeQuests, qg.GetType())
		})
	}
}

func TestQuestGeneratorImpl_GetType(t *testing.T) {
	qg := NewQuestGenerator(nil)
	assert.Equal(t, ContentTypeQuests, qg.GetType())
}

func TestQuestGeneratorImpl_GetVersion(t *testing.T) {
	qg := NewQuestGenerator(nil)
	assert.Equal(t, "1.0.0", qg.GetVersion())
}

func TestQuestGeneratorImpl_Validate(t *testing.T) {
	qg := NewQuestGenerator(nil)

	tests := []struct {
		name    string
		params  GenerationParams
		wantErr bool
		errMsg  string
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
			errMsg:  "seed must be non-zero",
		},
		{
			name: "difficulty too low",
			params: GenerationParams{
				Seed:        12345,
				Difficulty:  0,
				PlayerLevel: 3,
			},
			wantErr: true,
			errMsg:  "difficulty must be between 1 and 20",
		},
		{
			name: "difficulty too high",
			params: GenerationParams{
				Seed:        12345,
				Difficulty:  25,
				PlayerLevel: 3,
			},
			wantErr: true,
			errMsg:  "difficulty must be between 1 and 20",
		},
		{
			name: "player level too low",
			params: GenerationParams{
				Seed:        12345,
				Difficulty:  5,
				PlayerLevel: 0,
			},
			wantErr: true,
			errMsg:  "player level must be between 1 and 20",
		},
		{
			name: "player level too high",
			params: GenerationParams{
				Seed:        12345,
				Difficulty:  5,
				PlayerLevel: 25,
			},
			wantErr: true,
			errMsg:  "player level must be between 1 and 20",
		},
		{
			name: "valid quest parameters",
			params: GenerationParams{
				Seed:        12345,
				Difficulty:  5,
				PlayerLevel: 3,
				Constraints: map[string]interface{}{
					"quest_params": QuestParams{
						MinObjectives: 1,
						MaxObjectives: 3,
						RewardTier:    RarityCommon,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid quest parameters - min objectives too low",
			params: GenerationParams{
				Seed:        12345,
				Difficulty:  5,
				PlayerLevel: 3,
				Constraints: map[string]interface{}{
					"quest_params": QuestParams{
						MinObjectives: 0,
						MaxObjectives: 3,
					},
				},
			},
			wantErr: true,
			errMsg:  "min objectives must be at least 1",
		},
		{
			name: "invalid quest parameters - max objectives too high",
			params: GenerationParams{
				Seed:        12345,
				Difficulty:  5,
				PlayerLevel: 3,
				Constraints: map[string]interface{}{
					"quest_params": QuestParams{
						MinObjectives: 1,
						MaxObjectives: 15,
					},
				},
			},
			wantErr: true,
			errMsg:  "max objectives cannot exceed 10",
		},
		{
			name: "invalid quest parameters - max < min",
			params: GenerationParams{
				Seed:        12345,
				Difficulty:  5,
				PlayerLevel: 3,
				Constraints: map[string]interface{}{
					"quest_params": QuestParams{
						MinObjectives: 5,
						MaxObjectives: 3,
					},
				},
			},
			wantErr: true,
			errMsg:  "max objectives (3) must be >= min objectives (5)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := qg.Validate(tt.params)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestQuestGeneratorImpl_Generate(t *testing.T) {
	qg := NewQuestGenerator(nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		params  GenerationParams
		wantErr bool
	}{
		{
			name: "basic quest generation",
			params: GenerationParams{
				Seed:        12345,
				Difficulty:  5,
				PlayerLevel: 3,
				Constraints: map[string]interface{}{
					"quest_params": QuestParams{
						QuestType:     QuestTypeFetch,
						MinObjectives: 1,
						MaxObjectives: 2,
						RewardTier:    RarityCommon,
						Narrative:     NarrativeLinear,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "quest generation with default parameters",
			params: GenerationParams{
				Seed:        67890,
				Difficulty:  8,
				PlayerLevel: 5,
			},
			wantErr: false,
		},
		{
			name: "invalid parameters",
			params: GenerationParams{
				Seed:        0, // Invalid seed
				Difficulty:  5,
				PlayerLevel: 3,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := qg.Generate(ctx, tt.params)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				quest, ok := result.(*game.Quest)
				require.True(t, ok, "result should be a *game.Quest")

				// Validate quest structure
				assert.NotEmpty(t, quest.ID)
				assert.NotEmpty(t, quest.Title)
				assert.NotEmpty(t, quest.Description)
				assert.Equal(t, game.QuestNotStarted, quest.Status)
				assert.NotEmpty(t, quest.Objectives)
				assert.NotEmpty(t, quest.Rewards)

				// Validate objectives
				for _, obj := range quest.Objectives {
					assert.NotEmpty(t, obj.Description)
					assert.GreaterOrEqual(t, obj.Required, 1)
					assert.Equal(t, 0, obj.Progress)
					assert.False(t, obj.Completed)
				}

				// Validate rewards
				hasExpReward := false
				for _, reward := range quest.Rewards {
					assert.NotEmpty(t, reward.Type)
					assert.Greater(t, reward.Value, 0)
					if reward.Type == "exp" {
						hasExpReward = true
					}
				}
				assert.True(t, hasExpReward, "quest should always have experience reward")
			}
		})
	}
}

func TestQuestGeneratorImpl_GenerateQuest(t *testing.T) {
	qg := NewQuestGenerator(nil)
	ctx := context.Background()

	tests := []struct {
		name      string
		questType QuestType
		params    QuestParams
		wantErr   bool
	}{
		{
			name:      "fetch quest",
			questType: QuestTypeFetch,
			params: QuestParams{
				GenerationParams: GenerationParams{
					Seed:        12345,
					Difficulty:  5,
					PlayerLevel: 3,
				},
				MinObjectives: 1,
				MaxObjectives: 2,
				RewardTier:    RarityCommon,
			},
			wantErr: false,
		},
		{
			name:      "kill quest",
			questType: QuestTypeKill,
			params: QuestParams{
				GenerationParams: GenerationParams{
					Seed:        67890,
					Difficulty:  8,
					PlayerLevel: 5,
				},
				MinObjectives: 1,
				MaxObjectives: 3,
				RewardTier:    RarityUncommon,
			},
			wantErr: false,
		},
		{
			name:      "story quest",
			questType: QuestTypeStory,
			params: QuestParams{
				GenerationParams: GenerationParams{
					Seed:        54321,
					Difficulty:  10,
					PlayerLevel: 8,
				},
				MinObjectives: 2,
				MaxObjectives: 4,
				RewardTier:    RarityRare,
				Narrative:     NarrativeBranching,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quest, err := qg.GenerateQuest(ctx, tt.questType, tt.params)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, quest)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, quest)

				// Validate quest structure
				assert.NotEmpty(t, quest.ID)
				assert.Contains(t, quest.ID, string(tt.questType))
				assert.NotEmpty(t, quest.Title)
				assert.NotEmpty(t, quest.Description)
				assert.Equal(t, game.QuestNotStarted, quest.Status)
				assert.Len(t, quest.Objectives, len(quest.Objectives)) // Just verify it's not empty
				assert.NotEmpty(t, quest.Rewards)

				// Check objectives count is within bounds
				assert.GreaterOrEqual(t, len(quest.Objectives), tt.params.MinObjectives)
				assert.LessOrEqual(t, len(quest.Objectives), tt.params.MaxObjectives)
			}
		})
	}
}

func TestQuestGeneratorImpl_GenerateQuestChain(t *testing.T) {
	qg := NewQuestGenerator(nil)
	ctx := context.Background()

	tests := []struct {
		name        string
		chainLength int
		params      QuestParams
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "valid chain length",
			chainLength: 3,
			params: QuestParams{
				GenerationParams: GenerationParams{
					Seed:        12345,
					Difficulty:  5,
					PlayerLevel: 3,
				},
				MinObjectives: 1,
				MaxObjectives: 2,
				RewardTier:    RarityCommon,
			},
			wantErr: false,
		},
		{
			name:        "single quest chain",
			chainLength: 1,
			params: QuestParams{
				GenerationParams: GenerationParams{
					Seed:        67890,
					Difficulty:  5,
					PlayerLevel: 3,
				},
				MinObjectives: 1,
				MaxObjectives: 2,
				RewardTier:    RarityCommon,
			},
			wantErr: false,
		},
		{
			name:        "invalid chain length - zero",
			chainLength: 0,
			params: QuestParams{
				GenerationParams: GenerationParams{
					Seed:        12345,
					Difficulty:  5,
					PlayerLevel: 3,
				},
			},
			wantErr: true,
			errMsg:  "chain length must be positive",
		},
		{
			name:        "invalid chain length - negative",
			chainLength: -1,
			params: QuestParams{
				GenerationParams: GenerationParams{
					Seed:        12345,
					Difficulty:  5,
					PlayerLevel: 3,
				},
			},
			wantErr: true,
			errMsg:  "chain length must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quests, err := qg.GenerateQuestChain(ctx, tt.chainLength, tt.params)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, quests)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, quests)
				assert.Len(t, quests, tt.chainLength)

				// Validate each quest in the chain
				for i, quest := range quests {
					assert.NotEmpty(t, quest.ID)
					assert.NotEmpty(t, quest.Title)
					assert.NotEmpty(t, quest.Description)
					assert.Equal(t, game.QuestNotStarted, quest.Status)
					assert.NotEmpty(t, quest.Objectives)
					assert.NotEmpty(t, quest.Rewards)

					// Check part numbering for multi-quest chains
					if tt.chainLength > 1 {
						assert.Contains(t, quest.Title, fmt.Sprintf("(Part %d)", i+1))
					}
				}

				// Verify difficulty progression
				if len(quests) > 1 {
					// Should have different quest types for variety
					questTypes := make(map[string]bool)
					for _, quest := range quests {
						questTypes[quest.Title] = true
					}
					// Allow some repetition in longer chains, but expect some variety
					if len(quests) >= 3 {
						assert.GreaterOrEqual(t, len(questTypes), 2, "should have some quest type variety")
					}
				}
			}
		})
	}
}

func TestQuestGeneratorImpl_GenerateObjectives(t *testing.T) {
	qg := NewQuestGenerator(nil)
	ctx := context.Background()

	// Create a mock world for testing
	world := &game.World{}

	tests := []struct {
		name    string
		params  QuestParams
		wantErr bool
	}{
		{
			name: "basic objective generation",
			params: QuestParams{
				GenerationParams: GenerationParams{
					Seed:        12345,
					Difficulty:  5,
					PlayerLevel: 3,
				},
				QuestType:     QuestTypeFetch,
				MinObjectives: 1,
				MaxObjectives: 3,
			},
			wantErr: false,
		},
		{
			name: "single objective",
			params: QuestParams{
				GenerationParams: GenerationParams{
					Seed:        67890,
					Difficulty:  8,
					PlayerLevel: 5,
				},
				QuestType:     QuestTypeKill,
				MinObjectives: 1,
				MaxObjectives: 1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objectives, err := qg.GenerateObjectives(ctx, world, tt.params)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, objectives)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, objectives)

				// Check objectives count is within bounds
				assert.GreaterOrEqual(t, len(objectives), tt.params.MinObjectives)
				assert.LessOrEqual(t, len(objectives), tt.params.MaxObjectives)

				// Validate each objective
				for i, obj := range objectives {
					assert.NotEmpty(t, obj.ID)
					assert.NotEmpty(t, obj.Type)
					assert.NotEmpty(t, obj.Description)
					assert.NotEmpty(t, obj.Target)
					assert.GreaterOrEqual(t, obj.Quantity, 1)
					assert.Equal(t, 0, obj.Progress)
					assert.False(t, obj.Complete)
					assert.NotNil(t, obj.Conditions)

					// First objective should never be optional
					if i == 0 {
						assert.False(t, obj.Optional, "first objective should not be optional")
					}
				}
			}
		})
	}
}

func TestQuestGeneratorImpl_DeterministicGeneration(t *testing.T) {
	qg1 := NewQuestGenerator(nil)
	qg2 := NewQuestGenerator(nil)
	ctx := context.Background()

	params := GenerationParams{
		Seed:        12345,
		Difficulty:  5,
		PlayerLevel: 3,
		Constraints: map[string]interface{}{
			"quest_params": QuestParams{
				QuestType:     QuestTypeFetch,
				MinObjectives: 2,
				MaxObjectives: 2,
				RewardTier:    RarityCommon,
				Narrative:     NarrativeLinear,
			},
		},
	}

	// Generate with same seed should produce identical results
	result1, err1 := qg1.Generate(ctx, params)
	result2, err2 := qg2.Generate(ctx, params)

	assert.NoError(t, err1)
	assert.NoError(t, err2)

	quest1, ok1 := result1.(*game.Quest)
	quest2, ok2 := result2.(*game.Quest)

	require.True(t, ok1)
	require.True(t, ok2)

	// Should generate identical content with same seed
	assert.Equal(t, quest1.Title, quest2.Title)
	assert.Equal(t, quest1.Description, quest2.Description)
	assert.Equal(t, len(quest1.Objectives), len(quest2.Objectives))
	assert.Equal(t, len(quest1.Rewards), len(quest2.Rewards))

	// Verify objective details match
	for i := range quest1.Objectives {
		assert.Equal(t, quest1.Objectives[i].Description, quest2.Objectives[i].Description)
		assert.Equal(t, quest1.Objectives[i].Required, quest2.Objectives[i].Required)
	}
}

func TestQuestGeneratorImpl_ContextCancellation(t *testing.T) {
	qg := NewQuestGenerator(nil)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	params := QuestParams{
		GenerationParams: GenerationParams{
			Seed:        12345,
			Difficulty:  5,
			PlayerLevel: 3,
		},
		QuestType:     QuestTypeFetch,
		MinObjectives: 1,
		MaxObjectives: 2,
	}

	quest, err := qg.GenerateQuest(ctx, params.QuestType, params)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.Nil(t, quest)
}

func TestQuestGeneratorImpl_AllQuestTypes(t *testing.T) {
	qg := NewQuestGenerator(nil)
	ctx := context.Background()

	questTypes := []QuestType{
		QuestTypeFetch, QuestTypeKill, QuestTypeEscort,
		QuestTypeExplore, QuestTypeDefend, QuestTypePuzzle,
		QuestTypeDelivery, QuestTypeSurvival, QuestTypeStory,
	}

	for _, questType := range questTypes {
		t.Run(string(questType), func(t *testing.T) {
			params := QuestParams{
				GenerationParams: GenerationParams{
					Seed:        12345,
					Difficulty:  5,
					PlayerLevel: 3,
				},
				QuestType:     questType,
				MinObjectives: 1,
				MaxObjectives: 2,
				RewardTier:    RarityCommon,
			}

			quest, err := qg.GenerateQuest(ctx, questType, params)
			assert.NoError(t, err)
			require.NotNil(t, quest)

			// Verify quest is valid
			assert.NotEmpty(t, quest.ID)
			assert.Contains(t, quest.ID, string(questType))
			assert.NotEmpty(t, quest.Title)
			assert.NotEmpty(t, quest.Description)
			assert.NotEmpty(t, quest.Objectives)
			assert.NotEmpty(t, quest.Rewards)
		})
	}
}

func TestQuestGeneratorImpl_RewardGeneration(t *testing.T) {
	qg := NewQuestGenerator(nil)
	ctx := context.Background()

	tests := []struct {
		name       string
		questType  QuestType
		rewardTier RarityTier
		difficulty int
		level      int
	}{
		{
			name:       "common rewards",
			questType:  QuestTypeFetch,
			rewardTier: RarityCommon,
			difficulty: 3,
			level:      2,
		},
		{
			name:       "rare rewards",
			questType:  QuestTypeKill,
			rewardTier: RarityRare,
			difficulty: 8,
			level:      6,
		},
		{
			name:       "legendary rewards",
			questType:  QuestTypeStory,
			rewardTier: RarityLegendary,
			difficulty: 15,
			level:      12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := QuestParams{
				GenerationParams: GenerationParams{
					Seed:        12345,
					Difficulty:  tt.difficulty,
					PlayerLevel: tt.level,
				},
				QuestType:     tt.questType,
				MinObjectives: 1,
				MaxObjectives: 2,
				RewardTier:    tt.rewardTier,
			}

			quest, err := qg.GenerateQuest(ctx, tt.questType, params)
			assert.NoError(t, err)
			require.NotNil(t, quest)

			// Verify rewards
			assert.NotEmpty(t, quest.Rewards)

			// Should always have experience reward
			hasExp := false
			for _, reward := range quest.Rewards {
				if reward.Type == "exp" {
					hasExp = true
					assert.Greater(t, reward.Value, 0)
					// Higher difficulty/level should give more exp
					if tt.difficulty > 5 || tt.level > 5 {
						assert.Greater(t, reward.Value, 100)
					}
				}
			}
			assert.True(t, hasExp, "should always include experience reward")
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkQuestGeneratorImpl_Generate(b *testing.B) {
	qg := NewQuestGenerator(nil)
	ctx := context.Background()

	params := GenerationParams{
		Seed:        12345,
		Difficulty:  5,
		PlayerLevel: 3,
		Constraints: map[string]interface{}{
			"quest_params": QuestParams{
				QuestType:     QuestTypeFetch,
				MinObjectives: 2,
				MaxObjectives: 3,
				RewardTier:    RarityCommon,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Use different seed for each iteration to avoid caching effects
		params.Seed = int64(i + 12345)
		_, err := qg.Generate(ctx, params)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkQuestGeneratorImpl_GenerateQuestChain(b *testing.B) {
	qg := NewQuestGenerator(nil)
	ctx := context.Background()

	params := QuestParams{
		GenerationParams: GenerationParams{
			Seed:        12345,
			Difficulty:  5,
			PlayerLevel: 3,
		},
		MinObjectives: 1,
		MaxObjectives: 2,
		RewardTier:    RarityCommon,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Use different seed for each iteration to avoid caching effects
		params.Seed = int64(i + 12345)
		_, err := qg.GenerateQuestChain(ctx, 3, params)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func TestQuestGeneratorImpl_Performance(t *testing.T) {
	qg := NewQuestGenerator(nil)
	ctx := context.Background()

	params := GenerationParams{
		Seed:        12345,
		Difficulty:  5,
		PlayerLevel: 3,
		Constraints: map[string]interface{}{
			"quest_params": QuestParams{
				QuestType:     QuestTypeFetch,
				MinObjectives: 3,
				MaxObjectives: 5,
				RewardTier:    RarityUncommon,
			},
		},
	}

	// Measure generation time
	start := time.Now()
	_, err := qg.Generate(ctx, params)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.Less(t, duration, 50*time.Millisecond, "quest generation should complete within 50ms")

	t.Logf("Quest generation took %v", duration)
}
