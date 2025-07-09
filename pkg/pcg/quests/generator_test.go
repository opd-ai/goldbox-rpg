package quests

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/pcg"
)

func TestNewObjectiveBasedGenerator(t *testing.T) {
	generator := NewObjectiveBasedGenerator()

	if generator == nil {
		t.Fatal("expected generator to be created")
	}

	if generator.GetType() != pcg.ContentTypeQuests {
		t.Errorf("expected content type %s, got %s", pcg.ContentTypeQuests, generator.GetType())
	}

	if generator.GetVersion() != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", generator.GetVersion())
	}

	// Check that templates are initialized
	if len(generator.objectiveTemplates) == 0 {
		t.Error("expected objective templates to be initialized")
	}

	// Verify narrative engine is created
	if generator.narrativeEngine == nil {
		t.Error("expected narrative engine to be initialized")
	}
}

func TestObjectiveBasedGenerator_Validate(t *testing.T) {
	generator := NewObjectiveBasedGenerator()

	tests := []struct {
		name    string
		params  pcg.GenerationParams
		wantErr bool
	}{
		{
			name: "valid basic parameters",
			params: pcg.GenerationParams{
				Seed:        12345,
				Difficulty:  5,
				PlayerLevel: 3,
				Constraints: map[string]interface{}{
					"min_objectives": 1,
					"max_objectives": 3,
				},
				Timeout: 30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "difficulty too low",
			params: pcg.GenerationParams{
				Seed:        12345,
				Difficulty:  0,
				PlayerLevel: 3,
				Constraints: map[string]interface{}{},
				Timeout:     30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "difficulty too high",
			params: pcg.GenerationParams{
				Seed:        12345,
				Difficulty:  21,
				PlayerLevel: 3,
				Constraints: map[string]interface{}{},
				Timeout:     30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid min objectives",
			params: pcg.GenerationParams{
				Seed:        12345,
				Difficulty:  5,
				PlayerLevel: 3,
				Constraints: map[string]interface{}{
					"min_objectives": 0,
				},
				Timeout: 30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "max objectives less than min",
			params: pcg.GenerationParams{
				Seed:        12345,
				Difficulty:  5,
				PlayerLevel: 3,
				Constraints: map[string]interface{}{
					"min_objectives": 3,
					"max_objectives": 2,
				},
				Timeout: 30 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := generator.Validate(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestObjectiveBasedGenerator_Generate(t *testing.T) {
	generator := NewObjectiveBasedGenerator()
	ctx := context.Background()

	params := pcg.GenerationParams{
		Seed:        12345,
		Difficulty:  5,
		PlayerLevel: 3,
		Constraints: map[string]interface{}{
			"quest_type":     "fetch",
			"min_objectives": 1,
			"max_objectives": 2,
			"reward_tier":    "common",
		},
		Timeout: 30 * time.Second,
	}

	result, err := generator.Generate(ctx, params)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	quest, ok := result.(*game.Quest)
	if !ok {
		t.Fatalf("expected *game.Quest, got %T", result)
	}

	// Verify quest properties
	if quest.ID == "" {
		t.Error("expected quest ID to be set")
	}

	if quest.Title == "" {
		t.Error("expected quest title to be set")
	}

	if quest.Description == "" {
		t.Error("expected quest description to be set")
	}

	if quest.Status != game.QuestNotStarted {
		t.Errorf("expected quest status %v, got %v", game.QuestNotStarted, quest.Status)
	}

	if len(quest.Objectives) == 0 {
		t.Error("expected quest to have objectives")
	}

	if len(quest.Rewards) == 0 {
		t.Error("expected quest to have rewards")
	}

	// Verify objective properties
	for i, obj := range quest.Objectives {
		if obj.Description == "" {
			t.Errorf("objective %d: expected description to be set", i)
		}
		if obj.Progress != 0 {
			t.Errorf("objective %d: expected progress to be 0, got %d", i, obj.Progress)
		}
		if obj.Required <= 0 {
			t.Errorf("objective %d: expected required > 0, got %d", i, obj.Required)
		}
		if obj.Completed {
			t.Errorf("objective %d: expected completed to be false", i)
		}
	}

	// Verify reward properties
	hasExpReward := false
	for i, reward := range quest.Rewards {
		if reward.Type == "" {
			t.Errorf("reward %d: expected type to be set", i)
		}
		if reward.Value <= 0 {
			t.Errorf("reward %d: expected value > 0, got %d", i, reward.Value)
		}
		if reward.Type == "exp" {
			hasExpReward = true
		}
	}

	if !hasExpReward {
		t.Error("expected quest to have experience reward")
	}
}

func TestObjectiveBasedGenerator_GenerateQuest(t *testing.T) {
	generator := NewObjectiveBasedGenerator()
	ctx := context.Background()

	tests := []struct {
		name      string
		questType pcg.QuestType
		params    pcg.QuestParams
		wantErr   bool
	}{
		{
			name:      "kill quest",
			questType: pcg.QuestTypeKill,
			params: pcg.QuestParams{
				GenerationParams: pcg.GenerationParams{
					Seed:        12345,
					Difficulty:  5,
					PlayerLevel: 3,
					Timeout:     30 * time.Second,
				},
				MinObjectives: 1,
				MaxObjectives: 2,
				RewardTier:    pcg.RarityCommon,
				Narrative:     pcg.NarrativeLinear,
			},
			wantErr: false,
		},
		{
			name:      "fetch quest",
			questType: pcg.QuestTypeFetch,
			params: pcg.QuestParams{
				GenerationParams: pcg.GenerationParams{
					Seed:        54321,
					Difficulty:  8,
					PlayerLevel: 5,
					Timeout:     30 * time.Second,
				},
				MinObjectives: 2,
				MaxObjectives: 3,
				RewardTier:    pcg.RarityUncommon,
				Narrative:     pcg.NarrativeLinear,
			},
			wantErr: false,
		},
		{
			name:      "explore quest",
			questType: pcg.QuestTypeExplore,
			params: pcg.QuestParams{
				GenerationParams: pcg.GenerationParams{
					Seed:        98765,
					Difficulty:  3,
					PlayerLevel: 2,
					Timeout:     30 * time.Second,
				},
				MinObjectives: 1,
				MaxObjectives: 1,
				RewardTier:    pcg.RarityCommon,
				Narrative:     pcg.NarrativeLinear,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quest, err := generator.GenerateQuest(ctx, tt.questType, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateQuest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify quest type is reflected in objectives
				if len(quest.Objectives) < tt.params.MinObjectives {
					t.Errorf("expected at least %d objectives, got %d", tt.params.MinObjectives, len(quest.Objectives))
				}
				if len(quest.Objectives) > tt.params.MaxObjectives {
					t.Errorf("expected at most %d objectives, got %d", tt.params.MaxObjectives, len(quest.Objectives))
				}
			}
		})
	}
}

func TestObjectiveBasedGenerator_GenerateQuestChain(t *testing.T) {
	generator := NewObjectiveBasedGenerator()
	ctx := context.Background()

	params := pcg.QuestParams{
		GenerationParams: pcg.GenerationParams{
			Seed:        12345,
			Difficulty:  5,
			PlayerLevel: 3,
			Timeout:     30 * time.Second,
		},
		QuestType:     pcg.QuestTypeFetch,
		MinObjectives: 1,
		MaxObjectives: 2,
		RewardTier:    pcg.RarityCommon,
		Narrative:     pcg.NarrativeLinear,
	}

	tests := []struct {
		name        string
		chainLength int
		wantErr     bool
	}{
		{
			name:        "single quest chain",
			chainLength: 1,
			wantErr:     false,
		},
		{
			name:        "three quest chain",
			chainLength: 3,
			wantErr:     false,
		},
		{
			name:        "invalid chain length",
			chainLength: 0,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quests, err := generator.GenerateQuestChain(ctx, tt.chainLength, params)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateQuestChain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(quests) != tt.chainLength {
					t.Errorf("expected %d quests, got %d", tt.chainLength, len(quests))
				}

				// Verify each quest in chain
				for i, quest := range quests {
					if quest == nil {
						t.Errorf("quest %d is nil", i)
						continue
					}

					if tt.chainLength > 1 {
						expectedSuffix := fmt.Sprintf("(Part %d)", i+1)
						if !strings.Contains(quest.Title, expectedSuffix) {
							t.Errorf("quest %d title should contain '%s', got '%s'", i, expectedSuffix, quest.Title)
						}
					}
				}
			}
		})
	}
}

func TestObjectiveBasedGenerator_GenerateObjectives(t *testing.T) {
	generator := NewObjectiveBasedGenerator()
	ctx := context.Background()

	// Create a simple world for testing
	world := &game.World{}

	params := pcg.QuestParams{
		GenerationParams: pcg.GenerationParams{
			Seed:        12345,
			Difficulty:  5,
			PlayerLevel: 3,
			WorldState:  world,
			Timeout:     30 * time.Second,
		},
		QuestType:     pcg.QuestTypeKill,
		MinObjectives: 2,
		MaxObjectives: 4,
		RewardTier:    pcg.RarityCommon,
		Narrative:     pcg.NarrativeLinear,
	}

	objectives, err := generator.GenerateObjectives(ctx, world, params)
	if err != nil {
		t.Fatalf("GenerateObjectives() error = %v", err)
	}

	if len(objectives) < params.MinObjectives {
		t.Errorf("expected at least %d objectives, got %d", params.MinObjectives, len(objectives))
	}

	if len(objectives) > params.MaxObjectives {
		t.Errorf("expected at most %d objectives, got %d", params.MaxObjectives, len(objectives))
	}

	for i, obj := range objectives {
		if obj.ID == "" {
			t.Errorf("objective %d: expected ID to be set", i)
		}
		if obj.Type == "" {
			t.Errorf("objective %d: expected type to be set", i)
		}
		if obj.Description == "" {
			t.Errorf("objective %d: expected description to be set", i)
		}
		if obj.Quantity <= 0 {
			t.Errorf("objective %d: expected quantity > 0, got %d", i, obj.Quantity)
		}
		if obj.Progress != 0 {
			t.Errorf("objective %d: expected progress to be 0, got %d", i, obj.Progress)
		}
		if obj.Complete {
			t.Errorf("objective %d: expected complete to be false", i)
		}
	}
}

// TestDeterministicGeneration verifies that the same seed produces the same quest
func TestDeterministicGeneration(t *testing.T) {
	generator := NewObjectiveBasedGenerator()
	ctx := context.Background()

	params := pcg.GenerationParams{
		Seed:        12345,
		Difficulty:  5,
		PlayerLevel: 3,
		Constraints: map[string]interface{}{
			"quest_type":     "kill",
			"min_objectives": 2,
			"max_objectives": 2,
		},
		Timeout: 30 * time.Second,
	}

	// Generate quest twice with same seed
	result1, err1 := generator.Generate(ctx, params)
	if err1 != nil {
		t.Fatalf("first generation error = %v", err1)
	}

	result2, err2 := generator.Generate(ctx, params)
	if err2 != nil {
		t.Fatalf("second generation error = %v", err2)
	}

	quest1 := result1.(*game.Quest)
	quest2 := result2.(*game.Quest)

	// Verify deterministic generation
	if quest1.Title != quest2.Title {
		t.Errorf("expected same title, got '%s' and '%s'", quest1.Title, quest2.Title)
	}

	if len(quest1.Objectives) != len(quest2.Objectives) {
		t.Errorf("expected same number of objectives, got %d and %d", len(quest1.Objectives), len(quest2.Objectives))
	}

	if len(quest1.Rewards) != len(quest2.Rewards) {
		t.Errorf("expected same number of rewards, got %d and %d", len(quest1.Rewards), len(quest2.Rewards))
	}
}
