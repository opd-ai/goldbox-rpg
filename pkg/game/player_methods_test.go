package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPlayer_GetHP tests the GetHP method
func TestPlayer_GetHP(t *testing.T) {
	tests := []struct {
		name     string
		hp       int
		expected int
	}{
		{name: "normal HP", hp: 50, expected: 50},
		{name: "zero HP", hp: 0, expected: 0},
		{name: "full HP", hp: 100, expected: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := &Player{
				Character: Character{
					HP:    tt.hp,
					MaxHP: 100,
				},
			}
			assert.Equal(t, tt.expected, player.GetHP())
		})
	}
}

// TestPlayer_SetHP tests the SetHP method
func TestPlayer_SetHP(t *testing.T) {
	tests := []struct {
		name       string
		initialHP  int
		maxHP      int
		newHP      int
		expectedHP int
	}{
		{name: "set normal value", initialHP: 50, maxHP: 100, newHP: 75, expectedHP: 75},
		{name: "set to max", initialHP: 50, maxHP: 100, newHP: 100, expectedHP: 100},
		{name: "clamp above max", initialHP: 50, maxHP: 100, newHP: 150, expectedHP: 100},
		{name: "clamp below zero", initialHP: 50, maxHP: 100, newHP: -10, expectedHP: 0},
		{name: "set to zero", initialHP: 50, maxHP: 100, newHP: 0, expectedHP: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := &Player{
				Character: Character{
					HP:    tt.initialHP,
					MaxHP: tt.maxHP,
				},
			}
			player.SetHP(tt.newHP)
			assert.Equal(t, tt.expectedHP, player.HP)
		})
	}
}

// TestPlayer_GetMaxHP tests the GetMaxHP method
func TestPlayer_GetMaxHP(t *testing.T) {
	tests := []struct {
		name     string
		maxHP    int
		expected int
	}{
		{name: "normal max HP", maxHP: 100, expected: 100},
		{name: "low max HP", maxHP: 20, expected: 20},
		{name: "high max HP", maxHP: 500, expected: 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := &Player{
				Character: Character{
					MaxHP: tt.maxHP,
				},
			}
			assert.Equal(t, tt.expected, player.GetMaxHP())
		})
	}
}

// TestPlayer_StartQuest tests the StartQuest method
func TestPlayer_StartQuest(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *Player
		quest       Quest
		expectError bool
		errorMsg    string
	}{
		{
			name: "start valid quest",
			setup: func() *Player {
				return &Player{
					Character: Character{ID: "player1", Name: "Hero"},
					QuestLog:  []Quest{},
				}
			},
			quest: Quest{
				ID:          "quest1",
				Title:       "Rescue the Princess",
				Description: "Save the princess from the dragon",
				Objectives: []QuestObjective{
					{Description: "Find the dragon", Required: 1},
				},
			},
			expectError: false,
		},
		{
			name: "start quest with empty ID fails",
			setup: func() *Player {
				return &Player{
					Character: Character{ID: "player1"},
					QuestLog:  []Quest{},
				}
			},
			quest:       Quest{ID: "", Title: "Invalid Quest"},
			expectError: true,
			errorMsg:    "quest ID cannot be empty",
		},
		{
			name: "start duplicate quest fails",
			setup: func() *Player {
				return &Player{
					Character: Character{ID: "player1"},
					QuestLog: []Quest{
						{ID: "quest1", Title: "Existing Quest", Status: QuestActive},
					},
				}
			},
			quest:       Quest{ID: "quest1", Title: "Duplicate Quest"},
			expectError: true,
			errorMsg:    "quest quest1 already exists in quest log",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := tt.setup()
			err := player.StartQuest(tt.quest)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Len(t, player.QuestLog, 1)
				assert.Equal(t, QuestActive, player.QuestLog[0].Status)
			}
		})
	}
}

// TestPlayer_CompleteQuest tests the CompleteQuest method
func TestPlayer_CompleteQuest(t *testing.T) {
	tests := []struct {
		name          string
		setup         func() *Player
		questID       string
		expectError   bool
		errorContains string
		expectRewards bool
	}{
		{
			name: "complete quest with all objectives done",
			setup: func() *Player {
				return &Player{
					Character: Character{ID: "player1"},
					QuestLog: []Quest{
						{
							ID:     "quest1",
							Title:  "Test Quest",
							Status: QuestActive,
							Objectives: []QuestObjective{
								{Description: "Objective 1", Required: 1, Progress: 1, Completed: true},
							},
							Rewards: []QuestReward{
								{Type: "gold", Value: 100},
							},
						},
					},
				}
			},
			questID:       "quest1",
			expectError:   false,
			expectRewards: true,
		},
		{
			name: "complete quest with incomplete objectives fails",
			setup: func() *Player {
				return &Player{
					Character: Character{ID: "player1"},
					QuestLog: []Quest{
						{
							ID:     "quest1",
							Title:  "Test Quest",
							Status: QuestActive,
							Objectives: []QuestObjective{
								{Description: "Incomplete", Required: 5, Progress: 2, Completed: false},
							},
						},
					},
				}
			},
			questID:       "quest1",
			expectError:   true,
			errorContains: "cannot be completed",
		},
		{
			name: "complete already completed quest fails",
			setup: func() *Player {
				return &Player{
					Character: Character{ID: "player1"},
					QuestLog: []Quest{
						{ID: "quest1", Status: QuestCompleted},
					},
				}
			},
			questID:       "quest1",
			expectError:   true,
			errorContains: "already completed",
		},
		{
			name: "complete non-existent quest fails",
			setup: func() *Player {
				return &Player{
					Character: Character{ID: "player1"},
					QuestLog:  []Quest{},
				}
			},
			questID:       "nonexistent",
			expectError:   true,
			errorContains: "not found",
		},
		{
			name: "complete quest with empty ID fails",
			setup: func() *Player {
				return &Player{
					Character: Character{ID: "player1"},
					QuestLog:  []Quest{},
				}
			},
			questID:       "",
			expectError:   true,
			errorContains: "cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := tt.setup()
			rewards, err := player.CompleteQuest(tt.questID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
				if tt.expectRewards {
					assert.NotEmpty(t, rewards)
				}
			}
		})
	}
}

// TestPlayer_UpdateQuestObjective tests the UpdateQuestObjective method
func TestPlayer_UpdateQuestObjective(t *testing.T) {
	tests := []struct {
		name           string
		setup          func() *Player
		questID        string
		objectiveIndex int
		progress       int
		expectError    bool
		errorContains  string
		expectComplete bool
	}{
		{
			name: "update objective progress",
			setup: func() *Player {
				return &Player{
					Character: Character{ID: "player1"},
					QuestLog: []Quest{
						{
							ID:     "quest1",
							Status: QuestActive,
							Objectives: []QuestObjective{
								{Description: "Kill enemies", Required: 10, Progress: 0},
							},
						},
					},
				}
			},
			questID:        "quest1",
			objectiveIndex: 0,
			progress:       5,
			expectError:    false,
			expectComplete: false,
		},
		{
			name: "complete objective when progress meets requirement",
			setup: func() *Player {
				return &Player{
					Character: Character{ID: "player1"},
					QuestLog: []Quest{
						{
							ID:     "quest1",
							Status: QuestActive,
							Objectives: []QuestObjective{
								{Description: "Collect items", Required: 5, Progress: 0},
							},
						},
					},
				}
			},
			questID:        "quest1",
			objectiveIndex: 0,
			progress:       5,
			expectError:    false,
			expectComplete: true,
		},
		{
			name: "clamp progress to requirement",
			setup: func() *Player {
				return &Player{
					Character: Character{ID: "player1"},
					QuestLog: []Quest{
						{
							ID:     "quest1",
							Status: QuestActive,
							Objectives: []QuestObjective{
								{Description: "Test", Required: 10},
							},
						},
					},
				}
			},
			questID:        "quest1",
			objectiveIndex: 0,
			progress:       999, // Exceeds required
			expectError:    false,
			expectComplete: true,
		},
		{
			name: "negative progress fails",
			setup: func() *Player {
				return &Player{
					Character: Character{ID: "player1"},
					QuestLog:  []Quest{{ID: "quest1", Status: QuestActive}},
				}
			},
			questID:       "quest1",
			progress:      -5,
			expectError:   true,
			errorContains: "negative",
		},
		{
			name: "invalid objective index fails",
			setup: func() *Player {
				return &Player{
					Character: Character{ID: "player1"},
					QuestLog: []Quest{
						{
							ID:         "quest1",
							Status:     QuestActive,
							Objectives: []QuestObjective{},
						},
					},
				}
			},
			questID:        "quest1",
			objectiveIndex: 5,
			progress:       1,
			expectError:    true,
			errorContains:  "out of bounds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := tt.setup()
			err := player.UpdateQuestObjective(tt.questID, tt.objectiveIndex, tt.progress)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
				objective := &player.QuestLog[0].Objectives[tt.objectiveIndex]
				assert.Equal(t, tt.expectComplete, objective.Completed)
			}
		})
	}
}

// TestPlayer_FailQuest tests the FailQuest method
func TestPlayer_FailQuest(t *testing.T) {
	tests := []struct {
		name          string
		setup         func() *Player
		questID       string
		expectError   bool
		errorContains string
	}{
		{
			name: "fail active quest succeeds",
			setup: func() *Player {
				return &Player{
					Character: Character{ID: "player1"},
					QuestLog: []Quest{
						{ID: "quest1", Status: QuestActive},
					},
				}
			},
			questID:     "quest1",
			expectError: false,
		},
		{
			name: "fail completed quest fails",
			setup: func() *Player {
				return &Player{
					Character: Character{ID: "player1"},
					QuestLog: []Quest{
						{ID: "quest1", Status: QuestCompleted},
					},
				}
			},
			questID:       "quest1",
			expectError:   true,
			errorContains: "already completed",
		},
		{
			name: "fail already failed quest fails",
			setup: func() *Player {
				return &Player{
					Character: Character{ID: "player1"},
					QuestLog: []Quest{
						{ID: "quest1", Status: QuestFailed},
					},
				}
			},
			questID:       "quest1",
			expectError:   true,
			errorContains: "already failed",
		},
		{
			name: "fail non-existent quest fails",
			setup: func() *Player {
				return &Player{
					Character: Character{ID: "player1"},
					QuestLog:  []Quest{},
				}
			},
			questID:       "nonexistent",
			expectError:   true,
			errorContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := tt.setup()
			err := player.FailQuest(tt.questID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
				quest, _ := player.GetQuest(tt.questID)
				assert.Equal(t, QuestFailed, quest.Status)
			}
		})
	}
}

// TestPlayer_GetQuest tests the GetQuest method
func TestPlayer_GetQuest(t *testing.T) {
	tests := []struct {
		name          string
		setup         func() *Player
		questID       string
		expectError   bool
		errorContains string
	}{
		{
			name: "get existing quest",
			setup: func() *Player {
				return &Player{
					Character: Character{ID: "player1"},
					QuestLog: []Quest{
						{ID: "quest1", Title: "Test Quest", Status: QuestActive},
					},
				}
			},
			questID:     "quest1",
			expectError: false,
		},
		{
			name: "get non-existent quest fails",
			setup: func() *Player {
				return &Player{
					Character: Character{ID: "player1"},
					QuestLog:  []Quest{},
				}
			},
			questID:       "nonexistent",
			expectError:   true,
			errorContains: "not found",
		},
		{
			name: "empty quest ID fails",
			setup: func() *Player {
				return &Player{
					Character: Character{ID: "player1"},
					QuestLog:  []Quest{},
				}
			},
			questID:       "",
			expectError:   true,
			errorContains: "cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := tt.setup()
			quest, err := player.GetQuest(tt.questID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, quest)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, quest)
				assert.Equal(t, tt.questID, quest.ID)
			}
		})
	}
}

// TestPlayer_GetActiveQuests tests the GetActiveQuests method
func TestPlayer_GetActiveQuests(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *Player
		expected int
	}{
		{
			name: "get active quests from mixed log",
			setup: func() *Player {
				return &Player{
					Character: Character{ID: "player1"},
					QuestLog: []Quest{
						{ID: "quest1", Status: QuestActive},
						{ID: "quest2", Status: QuestCompleted},
						{ID: "quest3", Status: QuestActive},
						{ID: "quest4", Status: QuestFailed},
					},
				}
			},
			expected: 2,
		},
		{
			name: "no active quests",
			setup: func() *Player {
				return &Player{
					Character: Character{ID: "player1"},
					QuestLog: []Quest{
						{ID: "quest1", Status: QuestCompleted},
						{ID: "quest2", Status: QuestFailed},
					},
				}
			},
			expected: 0,
		},
		{
			name: "empty quest log",
			setup: func() *Player {
				return &Player{
					Character: Character{ID: "player1"},
					QuestLog:  []Quest{},
				}
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := tt.setup()
			activeQuests := player.GetActiveQuests()
			assert.Len(t, activeQuests, tt.expected)

			// Verify all returned quests are active
			for _, quest := range activeQuests {
				assert.Equal(t, QuestActive, quest.Status)
			}
		})
	}
}

// TestPlayer_GetCompletedQuests tests the GetCompletedQuests method
func TestPlayer_GetCompletedQuests(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *Player
		expected int
	}{
		{
			name: "get completed quests from mixed log",
			setup: func() *Player {
				return &Player{
					Character: Character{ID: "player1"},
					QuestLog: []Quest{
						{ID: "quest1", Status: QuestActive},
						{ID: "quest2", Status: QuestCompleted},
						{ID: "quest3", Status: QuestCompleted},
						{ID: "quest4", Status: QuestFailed},
					},
				}
			},
			expected: 2,
		},
		{
			name: "no completed quests",
			setup: func() *Player {
				return &Player{
					Character: Character{ID: "player1"},
					QuestLog: []Quest{
						{ID: "quest1", Status: QuestActive},
					},
				}
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := tt.setup()
			completedQuests := player.GetCompletedQuests()
			assert.Len(t, completedQuests, tt.expected)

			// Verify all returned quests are completed
			for _, quest := range completedQuests {
				assert.Equal(t, QuestCompleted, quest.Status)
			}
		})
	}
}
