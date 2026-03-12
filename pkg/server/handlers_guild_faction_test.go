package server

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleCreateGuild(t *testing.T) {
	server := createTestServerForHandlers(t)
	session := createTestSessionForHandlers(t, server)

	tests := []struct {
		name       string
		params     map[string]interface{}
		wantErr    bool
		wantFields []string
	}{
		{
			name: "valid guild creation",
			params: map[string]interface{}{
				"session_id":  session.SessionID,
				"name":        "Test Guild",
				"description": "A test guild for testing",
			},
			wantErr:    false,
			wantFields: []string{"success", "guild_id", "guild"},
		},
		{
			name: "missing session_id",
			params: map[string]interface{}{
				"name":        "Test Guild",
				"description": "A test guild",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, _ := json.Marshal(tt.params)
			result, err := server.handleCreateGuild(params)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			resultMap, ok := result.(map[string]interface{})
			require.True(t, ok)

			for _, field := range tt.wantFields {
				assert.Contains(t, resultMap, field)
			}
			assert.True(t, resultMap["success"].(bool))
		})
	}
}

func TestHandleGetGuild(t *testing.T) {
	server := createTestServerForHandlers(t)
	session := createTestSessionForHandlers(t, server)

	// First create a guild
	createParams, _ := json.Marshal(map[string]interface{}{
		"session_id":  session.SessionID,
		"name":        "Test Guild",
		"description": "A test guild",
	})
	createResult, err := server.handleCreateGuild(createParams)
	require.NoError(t, err)

	resultMap := createResult.(map[string]interface{})
	guildID := resultMap["guild_id"].(string)

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid guild ID",
			params: map[string]interface{}{
				"session_id": session.SessionID,
				"guild_id":   guildID,
			},
			wantErr: false,
		},
		{
			name: "invalid guild ID",
			params: map[string]interface{}{
				"session_id": session.SessionID,
				"guild_id":   "nonexistent-guild",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, _ := json.Marshal(tt.params)
			result, err := server.handleGetGuild(params)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			resultMap, ok := result.(map[string]interface{})
			require.True(t, ok)
			assert.True(t, resultMap["success"].(bool))
			assert.NotNil(t, resultMap["guild"])
		})
	}
}

func TestHandleListGuilds(t *testing.T) {
	server := createTestServerForHandlers(t)
	session := createTestSessionForHandlers(t, server)

	// Create a guild first
	createParams, _ := json.Marshal(map[string]interface{}{
		"session_id":  session.SessionID,
		"name":        "Test Guild",
		"description": "A test guild",
	})
	_, err := server.handleCreateGuild(createParams)
	require.NoError(t, err)

	// List guilds
	result, err := server.handleListGuilds(nil)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.True(t, resultMap["success"].(bool))
	assert.GreaterOrEqual(t, resultMap["count"].(int), 1)
}

func TestHandleGetFactionRelation(t *testing.T) {
	server := createTestServerForHandlers(t)

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name: "get new relation (auto-initialize)",
			params: map[string]interface{}{
				"faction1_id": "faction-a",
				"faction2_id": "faction-b",
			},
			wantErr: false,
		},
		{
			name: "self relation error",
			params: map[string]interface{}{
				"faction1_id": "faction-a",
				"faction2_id": "faction-a",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, _ := json.Marshal(tt.params)
			result, err := server.handleGetFactionRelation(params)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			resultMap, ok := result.(map[string]interface{})
			require.True(t, ok)
			assert.True(t, resultMap["success"].(bool))
			assert.NotNil(t, resultMap["relation"])
		})
	}
}

func TestHandleDeclareWar(t *testing.T) {
	server := createTestServerForHandlers(t)

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name: "declare war between factions",
			params: map[string]interface{}{
				"faction1_id": "faction-x",
				"faction2_id": "faction-y",
				"reason":      "territorial dispute",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, _ := json.Marshal(tt.params)
			result, err := server.handleDeclareWar(params)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			resultMap, ok := result.(map[string]interface{})
			require.True(t, ok)
			assert.True(t, resultMap["success"].(bool))
			assert.Equal(t, "war declared", resultMap["message"])
		})
	}
}

func TestHandleProposeAlliance(t *testing.T) {
	server := createTestServerForHandlers(t)

	params, _ := json.Marshal(map[string]interface{}{
		"faction1_id": "faction-p",
		"faction2_id": "faction-q",
	})

	result, err := server.handleProposeAlliance(params)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.True(t, resultMap["success"].(bool))
	assert.Equal(t, "alliance proposed", resultMap["message"])
}

func TestHandleSignTrade(t *testing.T) {
	server := createTestServerForHandlers(t)

	params, _ := json.Marshal(map[string]interface{}{
		"faction1_id": "faction-trade-a",
		"faction2_id": "faction-trade-b",
	})

	result, err := server.handleSignTrade(params)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.True(t, resultMap["success"].(bool))
	assert.Equal(t, "trade agreement signed", resultMap["message"])
}

func TestHandleSendDiplomaticGift(t *testing.T) {
	server := createTestServerForHandlers(t)

	params, _ := json.Marshal(map[string]interface{}{
		"sender_id":   "faction-gift-a",
		"receiver_id": "faction-gift-b",
		"value":       100,
	})

	result, err := server.handleSendDiplomaticGift(params)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.True(t, resultMap["success"].(bool))
	assert.Equal(t, "gift sent", resultMap["message"])
}
