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

// TestHandleGetFactionRelations tests getting all relations for a faction
func TestHandleGetFactionRelations(t *testing.T) {
	server := createTestServerForHandlers(t)

	// Test with valid faction ID
	params := json.RawMessage(`{"faction_id": "faction1"}`)
	result, err := server.handleGetFactionRelations(params)
	if err != nil {
		t.Errorf("handleGetFactionRelations() error = %v", err)
		return
	}

	response, ok := result.(map[string]interface{})
	if !ok {
		t.Error("Expected map response")
		return
	}
	if !response["success"].(bool) {
		t.Error("Expected success = true")
	}
}

// TestHandleOfferPeace tests peace offering
func TestHandleOfferPeace(t *testing.T) {
	server := createTestServerForHandlers(t)

	// First declare war to have something to make peace about
	_, _ = server.diplomacyManager.InitializeRelation("faction1", "faction2")
	_ = server.diplomacyManager.DeclareWar("faction1", "faction2", "test")

	params := json.RawMessage(`{"faction1_id": "faction1", "faction2_id": "faction2"}`)
	result, err := server.handleOfferPeace(params)
	if err != nil {
		t.Errorf("handleOfferPeace() error = %v", err)
		return
	}

	response, ok := result.(map[string]interface{})
	if !ok {
		t.Error("Expected map response")
		return
	}
	if !response["success"].(bool) {
		t.Error("Expected success = true")
	}
}

// TestHandleAcceptPeace tests accepting peace
func TestHandleAcceptPeace(t *testing.T) {
	server := createTestServerForHandlers(t)

	// Setup: factions at war with peace offer pending
	_, _ = server.diplomacyManager.InitializeRelation("faction1", "faction2")
	_ = server.diplomacyManager.DeclareWar("faction1", "faction2", "test")
	_ = server.diplomacyManager.OfferPeace("faction1", "faction2")

	params := json.RawMessage(`{"faction1_id": "faction2", "faction2_id": "faction1"}`)
	result, err := server.handleAcceptPeace(params)
	if err != nil {
		t.Errorf("handleAcceptPeace() error = %v", err)
		return
	}

	response, ok := result.(map[string]interface{})
	if !ok {
		t.Error("Expected map response")
		return
	}
	if !response["success"].(bool) {
		t.Error("Expected success = true")
	}
}

// TestHandleAcceptAlliance tests accepting an alliance
func TestHandleAcceptAlliance(t *testing.T) {
	server := createTestServerForHandlers(t)

	// Setup: alliance proposed
	_, _ = server.diplomacyManager.InitializeRelation("faction1", "faction2")
	_ = server.diplomacyManager.ProposeAlliance("faction1", "faction2")

	params := json.RawMessage(`{"faction1_id": "faction2", "faction2_id": "faction1"}`)
	result, err := server.handleAcceptAlliance(params)
	if err != nil {
		t.Errorf("handleAcceptAlliance() error = %v", err)
		return
	}

	response, ok := result.(map[string]interface{})
	if !ok {
		t.Error("Expected map response")
		return
	}
	if !response["success"].(bool) {
		t.Error("Expected success = true")
	}
}

// TestHandleBreakAlliance tests breaking an alliance
func TestHandleBreakAlliance(t *testing.T) {
	server := createTestServerForHandlers(t)

	// Setup: allied factions
	_, _ = server.diplomacyManager.InitializeRelation("faction1", "faction2")
	_ = server.diplomacyManager.ProposeAlliance("faction1", "faction2")
	_ = server.diplomacyManager.AcceptAlliance("faction2", "faction1")

	params := json.RawMessage(`{"faction1_id": "faction1", "faction2_id": "faction2", "reason": "test"}`)
	result, err := server.handleBreakAlliance(params)
	if err != nil {
		t.Errorf("handleBreakAlliance() error = %v", err)
		return
	}

	response, ok := result.(map[string]interface{})
	if !ok {
		t.Error("Expected map response")
		return
	}
	if !response["success"].(bool) {
		t.Error("Expected success = true")
	}
}

// TestHandleGetCharacterGuild tests getting a character's guild
func TestHandleGetCharacterGuild(t *testing.T) {
	server := createTestServerForHandlers(t)
	session := createTestSessionForHandlers(t, server)

	// First create a guild for the session's character
	createParams := map[string]interface{}{
		"session_id":  session.SessionID,
		"name":        "Test Guild",
		"description": "Test",
	}
	createData, _ := json.Marshal(createParams)
	_, _ = server.handleCreateGuild(createData)

	// Now get character guild using session_id
	params := map[string]interface{}{
		"session_id": session.SessionID,
	}
	data, _ := json.Marshal(params)
	result, err := server.handleGetCharacterGuild(data)
	if err != nil {
		t.Errorf("handleGetCharacterGuild() error = %v", err)
		return
	}

	response, ok := result.(map[string]interface{})
	if !ok {
		t.Error("Expected map response")
		return
	}
	if !response["success"].(bool) {
		t.Error("Expected success = true")
	}
}

// TestHandleGetCharacterGuild_NoGuild tests getting guild for character not in any guild
func TestHandleGetCharacterGuild_NoGuild(t *testing.T) {
	server := createTestServerForHandlers(t)
	session := createTestSessionForHandlers(t, server)

	// Get character guild without creating one
	params := map[string]interface{}{
		"session_id": session.SessionID,
	}
	data, _ := json.Marshal(params)
	result, err := server.handleGetCharacterGuild(data)
	if err != nil {
		t.Errorf("handleGetCharacterGuild() error = %v", err)
		return
	}

	response, ok := result.(map[string]interface{})
	if !ok {
		t.Error("Expected map response")
		return
	}
	if !response["success"].(bool) {
		t.Error("Expected success = true")
	}
	// Should return nil guild
	if response["guild"] != nil {
		t.Error("Expected nil guild for character not in guild")
	}
}

// TestHandleLeaveGuild tests leaving a guild - uses direct guild manager
func TestHandleLeaveGuild(t *testing.T) {
	server := createTestServerForHandlers(t)
	session := createTestSessionForHandlers(t, server)

	// First create a guild for the session's character
	createParams := map[string]interface{}{
		"session_id":  session.SessionID,
		"name":        "Test Guild",
		"description": "Test",
	}
	createData, _ := json.Marshal(createParams)
	createResult, _ := server.handleCreateGuild(createData)
	guildID := createResult.(map[string]interface{})["guild_id"].(string)

	// Add another member directly to guild manager (bypassing session)
	otherCharID := "other-char-123"
	_ = server.guildManager.JoinGuild(guildID, otherCharID, session.Player.GetID())

	// Now the founder can leave (since there's another member)
	leaveParams := map[string]interface{}{
		"session_id": session.SessionID,
	}
	data, _ := json.Marshal(leaveParams)
	result, err := server.handleLeaveGuild(data)
	if err != nil {
		t.Errorf("handleLeaveGuild() error = %v", err)
		return
	}

	response, ok := result.(map[string]interface{})
	if !ok {
		t.Error("Expected map response")
		return
	}
	if !response["success"].(bool) {
		t.Error("Expected success = true")
	}
}

// TestHandleGuildDeposit tests depositing gold to guild treasury
func TestHandleGuildDeposit(t *testing.T) {
	server := createTestServerForHandlers(t)
	session := createTestSessionForHandlers(t, server)

	// First create a guild for the session's character
	createParams := map[string]interface{}{
		"session_id":  session.SessionID,
		"name":        "Test Guild",
		"description": "Test",
	}
	createData, _ := json.Marshal(createParams)
	createResult, _ := server.handleCreateGuild(createData)
	guildID := createResult.(map[string]interface{})["guild_id"].(string)

	// Deposit to treasury
	depositParams := map[string]interface{}{
		"session_id": session.SessionID,
		"guild_id":   guildID,
		"amount":     100,
	}
	data, _ := json.Marshal(depositParams)
	result, err := server.handleGuildDeposit(data)
	if err != nil {
		t.Errorf("handleGuildDeposit() error = %v", err)
		return
	}

	response, ok := result.(map[string]interface{})
	if !ok {
		t.Error("Expected map response")
		return
	}
	if !response["success"].(bool) {
		t.Error("Expected success = true")
	}
}

// TestHandleGuildWithdraw tests withdrawing gold from guild treasury
func TestHandleGuildWithdraw(t *testing.T) {
	server := createTestServerForHandlers(t)
	session := createTestSessionForHandlers(t, server)

	// First create a guild for the session's character
	createParams := map[string]interface{}{
		"session_id":  session.SessionID,
		"name":        "Test Guild",
		"description": "Test",
	}
	createData, _ := json.Marshal(createParams)
	createResult, _ := server.handleCreateGuild(createData)
	guildID := createResult.(map[string]interface{})["guild_id"].(string)

	// Deposit first
	depositParams := map[string]interface{}{
		"session_id": session.SessionID,
		"guild_id":   guildID,
		"amount":     200,
	}
	depositData, _ := json.Marshal(depositParams)
	_, _ = server.handleGuildDeposit(depositData)

	// Now withdraw
	withdrawParams := map[string]interface{}{
		"session_id": session.SessionID,
		"guild_id":   guildID,
		"amount":     50,
	}
	data, _ := json.Marshal(withdrawParams)
	result, err := server.handleGuildWithdraw(data)
	if err != nil {
		t.Errorf("handleGuildWithdraw() error = %v", err)
		return
	}

	response, ok := result.(map[string]interface{})
	if !ok {
		t.Error("Expected map response")
		return
	}
	if !response["success"].(bool) {
		t.Error("Expected success = true")
	}
}
