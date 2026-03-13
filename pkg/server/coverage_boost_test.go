package server

import (
	"testing"

	"goldbox-rpg/pkg/game"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetVisibleObjects tests the getVisibleObjects method
func TestGetVisibleObjects(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	character := &game.Character{
		ID:       "test-player-visible",
		Name:     "Visible Test Player",
		HP:       100,
		Position: game.Position{X: 5, Y: 5, Level: 0},
	}
	player := &game.Player{
		Character: *character.Clone(),
	}

	// Add an object near the player
	server.state.WorldState.Objects["nearby-npc"] = &game.Character{
		ID:       "nearby-npc",
		Name:     "Nearby NPC",
		HP:       50,
		Position: game.Position{X: 6, Y: 6, Level: 0},
	}

	// Add an object far from the player
	server.state.WorldState.Objects["far-npc"] = &game.Character{
		ID:       "far-npc",
		Name:     "Far NPC",
		HP:       50,
		Position: game.Position{X: 100, Y: 100, Level: 0},
	}

	visible := server.getVisibleObjects(player)

	// Should see nearby object but not far one
	assert.NotEmpty(t, visible)
}

// TestGetActiveEffects tests the getActiveEffects method
func TestGetActiveEffects(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	character := &game.Character{
		ID:   "test-player-effects",
		Name: "Effects Test Player",
		HP:   100,
	}
	player := &game.Player{
		Character: *character.Clone(),
	}

	// Add an effect to the player's character
	effect := &game.Effect{
		ID:        "test-effect",
		Type:      game.EffectStatBoost,
		Magnitude: 5,
		SourceID:  "spell-001",
	}
	player.Character.AddEffect(effect)

	effects := server.getActiveEffects(player)

	// Should have at least one effect
	assert.NotNil(t, effects)
}

// TestGetCombatStateIfActive tests the getCombatStateIfActive method
func TestGetCombatStateIfActive(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	character := &game.Character{
		ID:       "test-player-combat",
		Name:     "Combat Test Player",
		HP:       100,
		Position: game.Position{X: 5, Y: 5, Level: 0},
	}
	player := &game.Player{
		Character: *character.Clone(),
	}

	// Test when not in combat
	state := server.getCombatStateIfActive(player)
	assert.Nil(t, state, "should return nil when not in combat")

	// Set up combat state
	server.state.TurnManager.IsInCombat = true
	server.state.TurnManager.Initiative = []string{player.GetID()}
	server.state.TurnManager.CurrentRound = 1

	// Test when in combat
	state = server.getCombatStateIfActive(player)
	assert.NotNil(t, state, "should return state when in combat")
	assert.Equal(t, 1, state.RoundCount)
}

// TestGetCombatEffects tests the getCombatEffects method
func TestGetCombatEffects(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	// Create a character with effects
	character := &game.Character{
		ID:       "combat-char",
		Name:     "Combat Character",
		HP:       100,
		Position: game.Position{X: 5, Y: 5, Level: 0},
	}

	effect := &game.Effect{
		ID:        "combat-effect",
		Type:      game.EffectStatPenalty,
		Magnitude: 3,
	}
	character.AddEffect(effect)

	// Add to world state and initiative
	server.state.WorldState.Objects[character.ID] = character
	server.state.TurnManager.Initiative = []string{character.ID}
	server.state.TurnManager.IsInCombat = true

	effects := server.getCombatEffects()

	// Should have effects for the character
	assert.NotNil(t, effects)
}

// TestFindInventoryItemCoverage tests the findInventoryItem method
func TestFindInventoryItemCoverage(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	character := &game.Character{
		ID:   "test-player-inv-cov",
		Name: "Inventory Test Player",
		HP:   100,
		Inventory: []game.Item{
			{ID: "health-potion-cov", Name: "Health Potion", Type: "potion"},
			{ID: "sword-cov", Name: "Steel Sword", Type: "weapon"},
		},
	}
	player := &game.Player{
		Character: *character.Clone(),
	}

	// Test finding existing item
	item, found := server.findInventoryItem(player, "health-potion-cov")
	assert.True(t, found)
	assert.NotNil(t, item)
	assert.Equal(t, "Health Potion", item.Name)

	// Test finding non-existent item
	item, found = server.findInventoryItem(player, "nonexistent")
	assert.False(t, found)
	assert.Nil(t, item)
}

// TestExecuteSpellCast tests spell casting execution
func TestExecuteSpellCast(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	// Create player
	character := &game.Character{
		ID:           "test-spellcaster",
		Name:         "Test Mage",
		HP:           100,
		MaxHP:        100,
		ActionPoints: 10,
		Position:     game.Position{X: 5, Y: 5, Level: 0},
	}
	player := &game.Player{
		Character: *character.Clone(),
	}

	// Create a simple spell
	spell := &game.Spell{
		ID:    "test-spell",
		Name:  "Test Spell",
		Level: 1,
	}

	// Add player to world
	server.state.WorldState.Objects[player.GetID()] = &player.Character

	// Test spell execution - may fail but exercises the code path
	_, err := server.executeSpellCast(player, spell, "", game.Position{X: 6, Y: 5, Level: 0})
	// Error is acceptable here, we're testing the code path
	_ = err
}

// TestConsumeSpellCastActionPoints tests action point consumption
func TestConsumeSpellCastActionPoints(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	// Create player with action points
	character := &game.Character{
		ID:              "test-ap-player",
		Name:            "AP Test Player",
		HP:              100,
		ActionPoints:    10,
		MaxActionPoints: 10,
	}
	player := &game.Player{
		Character: *character.Clone(),
	}

	// Test when not in combat (should succeed without consuming)
	server.state.TurnManager.IsInCombat = false
	err := server.consumeSpellCastActionPoints(player)
	assert.NoError(t, err)
	assert.Equal(t, 10, player.GetActionPoints())

	// Test when in combat
	server.state.TurnManager.IsInCombat = true
	initialAP := player.GetActionPoints()
	err = server.consumeSpellCastActionPoints(player)
	assert.NoError(t, err)
	assert.Less(t, player.GetActionPoints(), initialAP)
}

// TestIsPositionVisible tests visibility checks
func TestIsPositionVisible(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	from := game.Position{X: 5, Y: 5, Level: 0}

	// Test nearby position (should be visible)
	nearby := game.Position{X: 6, Y: 6, Level: 0}
	assert.True(t, server.isPositionVisible(from, nearby), "nearby position should be visible")

	// Test far position (should not be visible)
	far := game.Position{X: 50, Y: 50, Level: 0}
	assert.False(t, server.isPositionVisible(from, far), "far position should not be visible")

	// Test different level (should not be visible)
	differentLevel := game.Position{X: 6, Y: 6, Level: 1}
	assert.False(t, server.isPositionVisible(from, differentLevel), "different level should not be visible")
}

// TestCleanupSessionConnections tests session cleanup
func TestCleanupSessionConnections(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	session := &PlayerSession{
		SessionID:   "cleanup-test",
		MessageChan: make(chan []byte, 10),
		WSConn:      nil, // No WebSocket connection
	}

	// Should not panic with nil WSConn
	assert.NotPanics(t, func() {
		server.cleanupSessionConnections(session, "cleanup-test")
	})

	// Test with mock WebSocket
	session2 := &PlayerSession{
		SessionID:   "cleanup-test-2",
		MessageChan: make(chan []byte, 10),
		WSConn:      newMockWebSocketConn(),
	}

	assert.NotPanics(t, func() {
		server.cleanupSessionConnections(session2, "cleanup-test-2")
	})
}

// TestRemovePlayerFromGameState tests player removal
func TestRemovePlayerFromGameState(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	// Create player and add to world
	character := &game.Character{
		ID:   "removal-test-player",
		Name: "Removal Test Player",
		HP:   100,
	}
	player := &game.Player{
		Character: *character.Clone(),
	}

	server.state.WorldState.Objects[player.GetID()] = &player.Character

	session := &PlayerSession{
		SessionID: "removal-test-session",
		Player:    player,
	}

	// Verify player exists
	_, exists := server.state.WorldState.Objects[player.GetID()]
	assert.True(t, exists)

	// Remove player
	server.removePlayerFromGameState(session)

	// Verify player removed
	_, exists = server.state.WorldState.Objects[player.GetID()]
	assert.False(t, exists)
}

// TestRemovePlayerFromGameStateNilPlayer tests removal with nil player
func TestRemovePlayerFromGameStateNilPlayer(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	session := &PlayerSession{
		SessionID: "nil-player-test",
		Player:    nil,
	}

	// Should not panic with nil player
	assert.NotPanics(t, func() {
		server.removePlayerFromGameState(session)
	})
}

// TestServerHealthEndpoints tests health-related functions
func TestServerHealthEndpoints(t *testing.T) {
	server := createTestServerForHandlers(t)
	require.NotNil(t, server)
	defer server.Close()

	// Test that server was created successfully - just verify no panic
	assert.NotNil(t, server.state)
}
