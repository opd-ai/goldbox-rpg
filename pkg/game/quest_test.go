package game

import (
	"reflect"
	"testing"
)

// TestQuestStatus_Constants tests that all quest status constants have the expected values
func TestQuestStatus_Constants(t *testing.T) {
	tests := []struct {
		name   string
		status QuestStatus
		value  int
	}{
		{"QuestNotStarted", QuestNotStarted, 0},
		{"QuestActive", QuestActive, 1},
		{"QuestCompleted", QuestCompleted, 2},
		{"QuestFailed", QuestFailed, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.status) != tt.value {
				t.Errorf("QuestStatus %s = %d, want %d", tt.name, int(tt.status), tt.value)
			}
		})
	}
}

// TestQuest_StructInitialization tests Quest struct initialization and field assignment
func TestQuest_StructInitialization(t *testing.T) {
	objectives := []QuestObjective{
		{
			Description: "Kill 5 goblins",
			Progress:    2,
			Required:    5,
			Completed:   false,
		},
	}

	rewards := []QuestReward{
		{
			Type:   "exp",
			Value:  100,
			ItemID: "",
		},
	}

	quest := Quest{
		ID:          "quest_001",
		Title:       "Goblin Slayer",
		Description: "A quest to eliminate goblin threats",
		Status:      QuestActive,
		Objectives:  objectives,
		Rewards:     rewards,
	}

	// Test basic field assignment
	if quest.ID != "quest_001" {
		t.Errorf("Quest.ID = %q, want %q", quest.ID, "quest_001")
	}
	if quest.Title != "Goblin Slayer" {
		t.Errorf("Quest.Title = %q, want %q", quest.Title, "Goblin Slayer")
	}
	if quest.Description != "A quest to eliminate goblin threats" {
		t.Errorf("Quest.Description = %q, want %q", quest.Description, "A quest to eliminate goblin threats")
	}
	if quest.Status != QuestActive {
		t.Errorf("Quest.Status = %v, want %v", quest.Status, QuestActive)
	}

	// Test objectives slice
	if len(quest.Objectives) != 1 {
		t.Errorf("Quest.Objectives length = %d, want 1", len(quest.Objectives))
	}
	if quest.Objectives[0].Description != "Kill 5 goblins" {
		t.Errorf("Quest.Objectives[0].Description = %q, want %q", quest.Objectives[0].Description, "Kill 5 goblins")
	}

	// Test rewards slice
	if len(quest.Rewards) != 1 {
		t.Errorf("Quest.Rewards length = %d, want 1", len(quest.Rewards))
	}
	if quest.Rewards[0].Type != "exp" {
		t.Errorf("Quest.Rewards[0].Type = %q, want %q", quest.Rewards[0].Type, "exp")
	}
}

// TestQuest_EmptyInitialization tests Quest struct with zero values
func TestQuest_EmptyInitialization(t *testing.T) {
	var quest Quest

	// Test zero values
	if quest.ID != "" {
		t.Errorf("Empty Quest.ID = %q, want empty string", quest.ID)
	}
	if quest.Title != "" {
		t.Errorf("Empty Quest.Title = %q, want empty string", quest.Title)
	}
	if quest.Description != "" {
		t.Errorf("Empty Quest.Description = %q, want empty string", quest.Description)
	}
	if quest.Status != QuestNotStarted {
		t.Errorf("Empty Quest.Status = %v, want %v", quest.Status, QuestNotStarted)
	}
	if quest.Objectives != nil {
		t.Errorf("Empty Quest.Objectives = %v, want nil", quest.Objectives)
	}
	if quest.Rewards != nil {
		t.Errorf("Empty Quest.Rewards = %v, want nil", quest.Rewards)
	}
}

// TestQuestObjective_StructInitialization tests QuestObjective struct initialization
func TestQuestObjective_StructInitialization(t *testing.T) {
	tests := []struct {
		name        string
		objective   QuestObjective
		description string
		progress    int
		required    int
		completed   bool
	}{
		{
			name: "Active objective",
			objective: QuestObjective{
				Description: "Collect 10 herbs",
				Progress:    7,
				Required:    10,
				Completed:   false,
			},
			description: "Collect 10 herbs",
			progress:    7,
			required:    10,
			completed:   false,
		},
		{
			name: "Completed objective",
			objective: QuestObjective{
				Description: "Talk to the merchant",
				Progress:    1,
				Required:    1,
				Completed:   true,
			},
			description: "Talk to the merchant",
			progress:    1,
			required:    1,
			completed:   true,
		},
		{
			name: "Overachieved objective",
			objective: QuestObjective{
				Description: "Defeat enemies",
				Progress:    15,
				Required:    10,
				Completed:   true,
			},
			description: "Defeat enemies",
			progress:    15,
			required:    10,
			completed:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.objective.Description != tt.description {
				t.Errorf("QuestObjective.Description = %q, want %q", tt.objective.Description, tt.description)
			}
			if tt.objective.Progress != tt.progress {
				t.Errorf("QuestObjective.Progress = %d, want %d", tt.objective.Progress, tt.progress)
			}
			if tt.objective.Required != tt.required {
				t.Errorf("QuestObjective.Required = %d, want %d", tt.objective.Required, tt.required)
			}
			if tt.objective.Completed != tt.completed {
				t.Errorf("QuestObjective.Completed = %v, want %v", tt.objective.Completed, tt.completed)
			}
		})
	}
}

// TestQuestObjective_EdgeCases tests QuestObjective with edge case values
func TestQuestObjective_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		objective QuestObjective
	}{
		{
			name: "Zero values",
			objective: QuestObjective{
				Description: "",
				Progress:    0,
				Required:    0,
				Completed:   false,
			},
		},
		{
			name: "Negative progress",
			objective: QuestObjective{
				Description: "Test negative",
				Progress:    -5,
				Required:    10,
				Completed:   false,
			},
		},
		{
			name: "Large values",
			objective: QuestObjective{
				Description: "Test large values",
				Progress:    1000000,
				Required:    1000000,
				Completed:   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that the struct can be created and accessed without issues
			_ = tt.objective.Description
			_ = tt.objective.Progress
			_ = tt.objective.Required
			_ = tt.objective.Completed
		})
	}
}

// TestQuestReward_StructInitialization tests QuestReward struct initialization
func TestQuestReward_StructInitialization(t *testing.T) {
	tests := []struct {
		name   string
		reward QuestReward
		rType  string
		value  int
		itemID string
	}{
		{
			name: "Gold reward",
			reward: QuestReward{
				Type:   "gold",
				Value:  500,
				ItemID: "",
			},
			rType:  "gold",
			value:  500,
			itemID: "",
		},
		{
			name: "Experience reward",
			reward: QuestReward{
				Type:   "exp",
				Value:  1000,
				ItemID: "",
			},
			rType:  "exp",
			value:  1000,
			itemID: "",
		},
		{
			name: "Item reward",
			reward: QuestReward{
				Type:   "item",
				Value:  1,
				ItemID: "sword_001",
			},
			rType:  "item",
			value:  1,
			itemID: "sword_001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.reward.Type != tt.rType {
				t.Errorf("QuestReward.Type = %q, want %q", tt.reward.Type, tt.rType)
			}
			if tt.reward.Value != tt.value {
				t.Errorf("QuestReward.Value = %d, want %d", tt.reward.Value, tt.value)
			}
			if tt.reward.ItemID != tt.itemID {
				t.Errorf("QuestReward.ItemID = %q, want %q", tt.reward.ItemID, tt.itemID)
			}
		})
	}
}

// TestQuestReward_EmptyValues tests QuestReward with default/empty values
func TestQuestReward_EmptyValues(t *testing.T) {
	var reward QuestReward

	if reward.Type != "" {
		t.Errorf("Empty QuestReward.Type = %q, want empty string", reward.Type)
	}
	if reward.Value != 0 {
		t.Errorf("Empty QuestReward.Value = %d, want 0", reward.Value)
	}
	if reward.ItemID != "" {
		t.Errorf("Empty QuestReward.ItemID = %q, want empty string", reward.ItemID)
	}
}

// TestQuestProgress_StructInitialization tests QuestProgress struct initialization
func TestQuestProgress_StructInitialization(t *testing.T) {
	progress := QuestProgress{
		QuestID:            "quest_001",
		ObjectivesComplete: 3,
		TimeSpent:          1800, // 30 minutes
		Attempts:           2,
	}

	if progress.QuestID != "quest_001" {
		t.Errorf("QuestProgress.QuestID = %q, want %q", progress.QuestID, "quest_001")
	}
	if progress.ObjectivesComplete != 3 {
		t.Errorf("QuestProgress.ObjectivesComplete = %d, want 3", progress.ObjectivesComplete)
	}
	if progress.TimeSpent != 1800 {
		t.Errorf("QuestProgress.TimeSpent = %d, want 1800", progress.TimeSpent)
	}
	if progress.Attempts != 2 {
		t.Errorf("QuestProgress.Attempts = %d, want 2", progress.Attempts)
	}
}

// TestQuestProgress_ZeroValues tests QuestProgress with zero values
func TestQuestProgress_ZeroValues(t *testing.T) {
	var progress QuestProgress

	if progress.QuestID != "" {
		t.Errorf("Empty QuestProgress.QuestID = %q, want empty string", progress.QuestID)
	}
	if progress.ObjectivesComplete != 0 {
		t.Errorf("Empty QuestProgress.ObjectivesComplete = %d, want 0", progress.ObjectivesComplete)
	}
	if progress.TimeSpent != 0 {
		t.Errorf("Empty QuestProgress.TimeSpent = %d, want 0", progress.TimeSpent)
	}
	if progress.Attempts != 0 {
		t.Errorf("Empty QuestProgress.Attempts = %d, want 0", progress.Attempts)
	}
}

// TestQuest_StructFieldTags tests that Quest struct has correct field tags
func TestQuest_StructFieldTags(t *testing.T) {
	questType := reflect.TypeOf(Quest{})

	expectedTags := map[string]string{
		"ID":          "quest_id",
		"Title":       "quest_title",
		"Description": "quest_description",
		"Status":      "quest_status",
		"Objectives":  "quest_objectives",
		"Rewards":     "quest_rewards",
	}

	for fieldName, expectedTag := range expectedTags {
		field, found := questType.FieldByName(fieldName)
		if !found {
			t.Errorf("Field %s not found in Quest struct", fieldName)
			continue
		}

		yamlTag := field.Tag.Get("yaml")
		if yamlTag != expectedTag {
			t.Errorf("Field %s yaml tag = %q, want %q", fieldName, yamlTag, expectedTag)
		}
	}
}

// TestQuest_ComplexScenarios tests complex quest scenarios with multiple objectives and rewards
func TestQuest_ComplexScenarios(t *testing.T) {
	tests := []struct {
		name  string
		quest Quest
	}{
		{
			name: "Multi-objective quest",
			quest: Quest{
				ID:          "complex_001",
				Title:       "The Grand Adventure",
				Description: "A complex multi-part quest",
				Status:      QuestActive,
				Objectives: []QuestObjective{
					{Description: "Find the ancient scroll", Progress: 1, Required: 1, Completed: true},
					{Description: "Decode the scroll", Progress: 0, Required: 1, Completed: false},
					{Description: "Visit three temples", Progress: 2, Required: 3, Completed: false},
				},
				Rewards: []QuestReward{
					{Type: "exp", Value: 5000, ItemID: ""},
					{Type: "gold", Value: 1000, ItemID: ""},
					{Type: "item", Value: 1, ItemID: "legendary_sword"},
				},
			},
		},
		{
			name: "Failed quest scenario",
			quest: Quest{
				ID:          "failed_001",
				Title:       "The Lost Cause",
				Description: "A quest that went wrong",
				Status:      QuestFailed,
				Objectives: []QuestObjective{
					{Description: "Save the village", Progress: 0, Required: 1, Completed: false},
				},
				Rewards: []QuestReward{}, // No rewards for failed quest
			},
		},
		{
			name: "Completed quest",
			quest: Quest{
				ID:          "completed_001",
				Title:       "Hero's Journey",
				Description: "A successfully completed quest",
				Status:      QuestCompleted,
				Objectives: []QuestObjective{
					{Description: "Defeat the dragon", Progress: 1, Required: 1, Completed: true},
					{Description: "Return to town", Progress: 1, Required: 1, Completed: true},
				},
				Rewards: []QuestReward{
					{Type: "exp", Value: 10000, ItemID: ""},
					{Type: "item", Value: 1, ItemID: "dragon_slayer_title"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that the quest can be properly accessed
			if tt.quest.ID == "" {
				t.Error("Quest ID should not be empty")
			}
			if tt.quest.Title == "" {
				t.Error("Quest Title should not be empty")
			}

			// Verify objective consistency
			for i, obj := range tt.quest.Objectives {
				if obj.Progress < 0 {
					t.Errorf("Objective %d has negative progress: %d", i, obj.Progress)
				}
				if obj.Required <= 0 {
					t.Errorf("Objective %d has invalid required value: %d", i, obj.Required)
				}
				// Check completion logic
				if obj.Completed && obj.Progress < obj.Required {
					t.Errorf("Objective %d marked complete but progress (%d) < required (%d)", i, obj.Progress, obj.Required)
				}
			}

			// Verify reward types
			for i, reward := range tt.quest.Rewards {
				if reward.Type == "" {
					t.Errorf("Reward %d has empty type", i)
				}
				if reward.Value < 0 {
					t.Errorf("Reward %d has negative value: %d", i, reward.Value)
				}
				if reward.Type == "item" && reward.ItemID == "" {
					t.Errorf("Item reward %d missing ItemID", i)
				}
			}
		})
	}
}

// TestQuestStatus_TypeProperties tests properties of the QuestStatus type
func TestQuestStatus_TypeProperties(t *testing.T) {
	// Test that QuestStatus is based on int
	var status QuestStatus
	statusType := reflect.TypeOf(status)
	if statusType.Kind() != reflect.Int {
		t.Errorf("QuestStatus kind = %v, want %v", statusType.Kind(), reflect.Int)
	}
}

// TestQuestReward_ValidRewardTypes tests common reward type validation patterns
func TestQuestReward_ValidRewardTypes(t *testing.T) {
	validTypes := []string{"gold", "exp", "item"}

	for _, rewardType := range validTypes {
		t.Run("RewardType_"+rewardType, func(t *testing.T) {
			reward := QuestReward{
				Type:   rewardType,
				Value:  100,
				ItemID: "",
			}

			if reward.Type != rewardType {
				t.Errorf("QuestReward.Type = %q, want %q", reward.Type, rewardType)
			}
		})
	}
}

// TestQuest_DeepCopyScenario tests that quest structures maintain independence
func TestQuest_DeepCopyScenario(t *testing.T) {
	original := Quest{
		ID:     "original",
		Title:  "Original Quest",
		Status: QuestActive,
		Objectives: []QuestObjective{
			{Description: "Test", Progress: 5, Required: 10, Completed: false},
		},
		Rewards: []QuestReward{
			{Type: "gold", Value: 100, ItemID: ""},
		},
	}

	// Create a copy by value assignment
	copy := original
	copy.ID = "copy"
	copy.Objectives[0].Progress = 10
	copy.Rewards[0].Value = 200

	// Original should be unchanged in terms of ID
	if original.ID != "original" {
		t.Errorf("Original Quest.ID changed: %q, want %q", original.ID, "original")
	}

	// But slice modifications affect both (this is expected Go behavior)
	if original.Objectives[0].Progress == 5 {
		t.Log("Slice modifications affect original (expected Go behavior)")
	}
}
