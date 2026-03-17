package server

import (
	"encoding/json"
	"testing"
	"time"

	"goldbox-rpg/pkg/game"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleQuestEditorCreate(t *testing.T) {
	server, err := NewRPCServer(":8080")
	require.NoError(t, err)
	defer server.Stop()

	session := createTestSession(server)

	t.Run("valid quest", func(t *testing.T) {
		req := createQuestRequest{
			SessionID:   session.SessionID,
			Title:       "Rescue the Villagers",
			Description: "Save villagers from the dungeon.",
			Objectives: []questObjectiveInput{
				{Description: "Find the key", Required: 1},
				{Description: "Open the gate", Required: 1},
			},
			Rewards: []questRewardInput{
				{Type: "gold", Value: 100},
				{Type: "exp", Value: 250},
			},
		}
		params, _ := json.Marshal(req)
		result, err := server.handleQuestEditorCreate(params)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		resultMap := result.(map[string]interface{})
		assert.True(t, resultMap["success"].(bool))
		assert.NotEmpty(t, resultMap["quest_id"])
	})

	t.Run("missing title", func(t *testing.T) {
		req := createQuestRequest{
			SessionID: session.SessionID,
			Title:     "",
			Objectives: []questObjectiveInput{
				{Description: "Do something", Required: 1},
			},
		}
		params, _ := json.Marshal(req)
		_, err := server.handleQuestEditorCreate(params)
		assert.Error(t, err)
	})

	t.Run("no objectives", func(t *testing.T) {
		req := createQuestRequest{
			SessionID:  session.SessionID,
			Title:      "Empty Quest",
			Objectives: []questObjectiveInput{},
		}
		params, _ := json.Marshal(req)
		_, err := server.handleQuestEditorCreate(params)
		assert.Error(t, err)
	})

	t.Run("invalid reward type", func(t *testing.T) {
		req := createQuestRequest{
			SessionID: session.SessionID,
			Title:     "Bad Reward Quest",
			Objectives: []questObjectiveInput{
				{Description: "Task", Required: 1},
			},
			Rewards: []questRewardInput{
				{Type: "invalid", Value: 50},
			},
		}
		params, _ := json.Marshal(req)
		_, err := server.handleQuestEditorCreate(params)
		assert.Error(t, err)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		_, err := server.handleQuestEditorCreate(json.RawMessage(`{invalid}`))
		assert.Error(t, err)
	})
}

func TestHandleQuestEditorGet(t *testing.T) {
	server, err := NewRPCServer(":8080")
	require.NoError(t, err)
	defer server.Stop()

	session := createTestSession(server)

	t.Run("valid get", func(t *testing.T) {
		// First create a quest
		createReq := createQuestRequest{
			SessionID:   session.SessionID,
			Title:       "Test Quest for Get",
			Description: "Test description",
			Objectives:  []questObjectiveInput{{Description: "Objective 1", Required: 1}},
			Rewards:     []questRewardInput{{Type: "gold", Value: 100}},
		}
		createParams, _ := json.Marshal(createReq)
		createResult, err := server.handleQuestEditorCreate(createParams)
		require.NoError(t, err)
		questID := createResult.(map[string]interface{})["quest_id"].(string)

		// Now get the quest
		req := getQuestEditorRequest{
			SessionID: session.SessionID,
			QuestID:   questID,
		}
		params, _ := json.Marshal(req)
		result, err := server.handleQuestEditorGet(params)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		resultMap := result.(map[string]interface{})
		assert.Equal(t, "Test Quest for Get", resultMap["title"])
	})

	t.Run("quest not found", func(t *testing.T) {
		req := getQuestEditorRequest{
			SessionID: session.SessionID,
			QuestID:   "nonexistent-quest-123",
		}
		params, _ := json.Marshal(req)
		_, err := server.handleQuestEditorGet(params)
		assert.Error(t, err)
	})

	t.Run("missing quest ID", func(t *testing.T) {
		req := getQuestEditorRequest{
			SessionID: session.SessionID,
			QuestID:   "",
		}
		params, _ := json.Marshal(req)
		_, err := server.handleQuestEditorGet(params)
		assert.Error(t, err)
	})
}

func TestHandleQuestEditorUpdate(t *testing.T) {
	server, err := NewRPCServer(":8080")
	require.NoError(t, err)
	defer server.Stop()

	session := createTestSession(server)

	t.Run("valid update", func(t *testing.T) {
		// First create a quest
		createReq := createQuestRequest{
			SessionID:   session.SessionID,
			Title:       "Original Quest",
			Description: "Original description",
			Objectives:  []questObjectiveInput{{Description: "Objective 1", Required: 1}},
			Rewards:     []questRewardInput{{Type: "gold", Value: 100}},
		}
		createParams, _ := json.Marshal(createReq)
		createResult, err := server.handleQuestEditorCreate(createParams)
		require.NoError(t, err)
		questID := createResult.(map[string]interface{})["quest_id"].(string)

		// Now update the quest
		req := updateQuestRequest{
			SessionID:   session.SessionID,
			QuestID:     questID,
			Title:       "Updated Quest",
			Description: "New description.",
			Objectives: []questObjectiveInput{
				{Description: "New objective", Required: 3},
			},
		}
		params, _ := json.Marshal(req)
		result, err := server.handleQuestEditorUpdate(params)
		assert.NoError(t, err)
		resultMap := result.(map[string]interface{})
		assert.Equal(t, "Updated Quest", resultMap["title"])
	})

	t.Run("quest not found", func(t *testing.T) {
		req := updateQuestRequest{
			SessionID:   session.SessionID,
			QuestID:     "nonexistent-quest-123",
			Title:       "Updated Quest",
			Description: "New description.",
		}
		params, _ := json.Marshal(req)
		_, err := server.handleQuestEditorUpdate(params)
		assert.Error(t, err)
	})

	t.Run("missing quest ID", func(t *testing.T) {
		req := updateQuestRequest{
			SessionID: session.SessionID,
			QuestID:   "",
			Title:     "Test",
		}
		params, _ := json.Marshal(req)
		_, err := server.handleQuestEditorUpdate(params)
		assert.Error(t, err)
	})

	t.Run("missing title", func(t *testing.T) {
		// First create a quest
		createReq := createQuestRequest{
			SessionID:   session.SessionID,
			Title:       "Quest for Title Test",
			Description: "Description",
			Objectives:  []questObjectiveInput{{Description: "Objective 1", Required: 1}},
			Rewards:     []questRewardInput{{Type: "gold", Value: 100}},
		}
		createParams, _ := json.Marshal(createReq)
		createResult, err := server.handleQuestEditorCreate(createParams)
		require.NoError(t, err)
		questID := createResult.(map[string]interface{})["quest_id"].(string)

		req := updateQuestRequest{
			SessionID: session.SessionID,
			QuestID:   questID,
			Title:     "",
		}
		params, _ := json.Marshal(req)
		_, err = server.handleQuestEditorUpdate(params)
		assert.Error(t, err)
	})
}

func TestHandleQuestEditorDelete(t *testing.T) {
	server, err := NewRPCServer(":8080")
	require.NoError(t, err)
	defer server.Stop()

	session := createTestSession(server)

	t.Run("valid delete", func(t *testing.T) {
		// First create a quest to delete
		createReq := createQuestRequest{
			SessionID:   session.SessionID,
			Title:       "Quest to Delete",
			Description: "Description",
			Objectives:  []questObjectiveInput{{Description: "Objective 1", Required: 1}},
			Rewards:     []questRewardInput{{Type: "gold", Value: 100}},
		}
		createParams, _ := json.Marshal(createReq)
		createResult, err := server.handleQuestEditorCreate(createParams)
		require.NoError(t, err)
		questID := createResult.(map[string]interface{})["quest_id"].(string)

		req := deleteQuestRequest{
			SessionID: session.SessionID,
			QuestID:   questID,
		}
		params, _ := json.Marshal(req)
		result, err := server.handleQuestEditorDelete(params)
		assert.NoError(t, err)
		resultMap := result.(map[string]interface{})
		assert.True(t, resultMap["success"].(bool))
	})

	t.Run("quest not found", func(t *testing.T) {
		req := deleteQuestRequest{
			SessionID: session.SessionID,
			QuestID:   "nonexistent-quest-123",
		}
		params, _ := json.Marshal(req)
		_, err := server.handleQuestEditorDelete(params)
		assert.Error(t, err)
	})

	t.Run("missing quest ID", func(t *testing.T) {
		req := deleteQuestRequest{
			SessionID: session.SessionID,
			QuestID:   "",
		}
		params, _ := json.Marshal(req)
		_, err := server.handleQuestEditorDelete(params)
		assert.Error(t, err)
	})
}

func TestHandleQuestEditorList(t *testing.T) {
	server, err := NewRPCServer(":8080")
	require.NoError(t, err)
	defer server.Stop()

	session := createTestSession(server)

	t.Run("empty list", func(t *testing.T) {
		req := listQuestsRequest{
			SessionID: session.SessionID,
		}
		params, _ := json.Marshal(req)
		result, err := server.handleQuestEditorList(params)
		assert.NoError(t, err)
		resultMap := result.(map[string]interface{})
		assert.True(t, resultMap["success"].(bool))
	})

	t.Run("list with quests", func(t *testing.T) {
		// Create some quests
		for i := 0; i < 3; i++ {
			createReq := createQuestRequest{
				SessionID:   session.SessionID,
				Title:       "Test Quest " + string(rune('A'+i)),
				Description: "Description",
				Objectives:  []questObjectiveInput{{Description: "Objective 1", Required: 1}},
				Rewards:     []questRewardInput{{Type: "gold", Value: 100}},
			}
			createParams, _ := json.Marshal(createReq)
			_, err := server.handleQuestEditorCreate(createParams)
			require.NoError(t, err)
		}

		req := listQuestsRequest{
			SessionID: session.SessionID,
		}
		params, _ := json.Marshal(req)
		result, err := server.handleQuestEditorList(params)
		assert.NoError(t, err)
		resultMap := result.(map[string]interface{})
		assert.True(t, resultMap["success"].(bool))
		quests := resultMap["quests"].([]map[string]interface{})
		assert.GreaterOrEqual(t, len(quests), 3)
	})
}

func TestValidateQuestEditorInput(t *testing.T) {
	tests := []struct {
		name    string
		req     createQuestRequest
		wantErr bool
	}{
		{
			name: "valid quest",
			req: createQuestRequest{
				Title: "Test Quest",
				Objectives: []questObjectiveInput{
					{Description: "Do A", Required: 1},
				},
				Rewards: []questRewardInput{
					{Type: "gold", Value: 100},
				},
			},
			wantErr: false,
		},
		{
			name: "empty title",
			req: createQuestRequest{
				Title: "",
				Objectives: []questObjectiveInput{
					{Description: "Do A", Required: 1},
				},
			},
			wantErr: true,
		},
		{
			name: "no objectives",
			req: createQuestRequest{
				Title:      "Test",
				Objectives: []questObjectiveInput{},
			},
			wantErr: true,
		},
		{
			name: "objective with zero required",
			req: createQuestRequest{
				Title: "Test",
				Objectives: []questObjectiveInput{
					{Description: "Do A", Required: 0},
				},
			},
			wantErr: true,
		},
		{
			name: "objective with empty description",
			req: createQuestRequest{
				Title: "Test",
				Objectives: []questObjectiveInput{
					{Description: "", Required: 1},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid reward type",
			req: createQuestRequest{
				Title: "Test",
				Objectives: []questObjectiveInput{
					{Description: "Do A", Required: 1},
				},
				Rewards: []questRewardInput{
					{Type: "magic", Value: 10},
				},
			},
			wantErr: true,
		},
		{
			name: "reward with zero value",
			req: createQuestRequest{
				Title: "Test",
				Objectives: []questObjectiveInput{
					{Description: "Do A", Required: 1},
				},
				Rewards: []questRewardInput{
					{Type: "gold", Value: 0},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateQuestEditorInput(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuildQuestFromInput(t *testing.T) {
	req := createQuestRequest{
		Title:       "Dragon Hunt",
		Description: "Slay the dragon.",
		Objectives: []questObjectiveInput{
			{Description: "Find dragon lair", Required: 1},
			{Description: "Defeat dragon", Required: 1},
		},
		Rewards: []questRewardInput{
			{Type: "gold", Value: 500},
			{Type: "item", Value: 1, ItemID: "sword-of-fire"},
		},
	}

	quest := buildQuestFromInput(req)
	assert.NotEmpty(t, quest.ID)
	assert.Equal(t, "Dragon Hunt", quest.Title)
	assert.Len(t, quest.Objectives, 2)
	assert.Len(t, quest.Rewards, 2)
	assert.Equal(t, "sword-of-fire", quest.Rewards[1].ItemID)
	assert.Equal(t, 0, quest.Objectives[0].Progress)
	assert.False(t, quest.Objectives[0].Completed)
}

func TestValidateRewardType(t *testing.T) {
	tests := []struct {
		rewardType string
		wantErr    bool
	}{
		{"gold", false},
		{"item", false},
		{"exp", false},
		{"", true},
		{"magic", true},
	}

	for _, tt := range tests {
		t.Run(tt.rewardType, func(t *testing.T) {
			err := validateRewardType(tt.rewardType)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// createTestSession is a helper that creates a test session with a player.
func createTestSession(s *RPCServer) *PlayerSession {
	s.mu.Lock()
	defer s.mu.Unlock()

	testChar := game.Character{
		ID:   "test_player_quest",
		Name: "Quest Test Player",
	}
	testPlayer := &game.Player{
		Character: *testChar.Clone(),
		Level:     1,
	}

	session := &PlayerSession{
		SessionID:   "test-session-quest",
		Player:      testPlayer,
		LastActive:  time.Now(),
		CreatedAt:   time.Now(),
		Connected:   true,
		MessageChan: make(chan []byte, MessageChanBufferSize),
	}
	s.sessions[session.SessionID] = session
	return session
}
