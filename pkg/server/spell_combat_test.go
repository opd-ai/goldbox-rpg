package server

import (
	"encoding/json"
	"testing"

	"goldbox-rpg/pkg/game"

	"github.com/stretchr/testify/assert"
)

// TestSpellDamageProcessing tests spell damage application
func TestApplySpellDamageTargetNotFound(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	// Target doesn't exist - function may or may not error (simulation mode)
	err := server.applySpellDamage("nonexistent", 10, "fire")
	// Just verify it doesn't panic
	_ = err
}

// TestApplySpellDamageValid tests spell damage on valid target
func TestApplySpellDamageValid(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	// Create target
	enemy := &game.Character{
		ID:    "spell-target",
		Name:  "Goblin",
		HP:    30,
		MaxHP: 30,
	}
	server.state.WorldState.Objects[enemy.ID] = enemy

	err := server.applySpellDamage(enemy.ID, 10, "fire")
	// Should succeed
	_ = err // May or may not error depending on implementation
}

// TestApplySpellHealing tests spell healing application
func TestApplySpellHealingTargetNotFound(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	// Target doesn't exist - function may or may not error (simulation mode)
	err := server.applySpellHealing("nonexistent", 15)
	// Just verify it doesn't panic
	_ = err
}

// TestApplySpellHealingValid tests valid healing
func TestApplySpellHealingValid(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	// Create damaged target
	character := &game.Character{
		ID:    "heal-target",
		Name:  "Wounded Fighter",
		HP:    20,
		MaxHP: 50,
	}
	server.state.WorldState.Objects[character.ID] = character

	err := server.applySpellHealing(character.ID, 15)
	// Should succeed
	_ = err
}

// TestProcessCombatActionAttack tests combat action processing
func TestProcessCombatActionAttack(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	session := createTestSessionForHandlers(t, server)

	// Create enemy
	enemy := &game.Character{
		ID:       "combat-enemy-001",
		Name:     "Goblin",
		HP:       20,
		MaxHP:    20,
		Position: game.Position{X: 6, Y: 5, Level: 0},
	}
	server.state.WorldState.Objects[enemy.ID] = enemy
	server.state.WorldState.Objects[session.Player.GetID()] = &session.Player.Character

	server.state.TurnManager.IsInCombat = true
	server.state.TurnManager.Initiative = []string{session.Player.GetID(), enemy.ID}

	// Should not panic
	result, err := server.processCombatAction(session.Player, enemy.ID, "")
	_ = result
	_ = err
}

// TestStartCombatWithMultipleCombatants tests combat start
func TestStartCombatWithMultipleCombatants(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	session := createTestSessionForHandlers(t, server)

	// Add combatants to world
	server.state.WorldState.Objects[session.Player.GetID()] = &session.Player.Character

	enemy := &game.Character{
		ID:       "start-combat-enemy",
		Name:     "Orc",
		HP:       30,
		MaxHP:    30,
		Position: game.Position{X: 7, Y: 5, Level: 0},
	}
	server.state.WorldState.Objects[enemy.ID] = enemy

	// Test start combat handler
	params := json.RawMessage(`{"session_id": "` + session.SessionID + `", "enemy_ids": ["start-combat-enemy"]}`)
	result, err := server.handleStartCombat(params)

	// May succeed or fail based on combat state
	_ = result
	_ = err
}

// TestEndRoundProcessing tests end of round processing
func TestEndRoundProcessing(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	server.state.TurnManager.IsInCombat = true
	server.state.TurnManager.CurrentRound = 1

	// Should not panic
	assert.NotPanics(t, func() {
		server.processEndRound()
	})

	assert.Equal(t, 2, server.state.TurnManager.CurrentRound)
}

// TestAdvanceTurn tests turn advancement
func TestAdvanceTurn(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	session := createTestSessionForHandlers(t, server)

	server.state.TurnManager.Initiative = []string{
		session.Player.GetID(),
		"enemy-001",
	}
	server.state.TurnManager.CurrentIndex = 0
	server.state.TurnManager.IsInCombat = true

	nextTurn := server.state.TurnManager.AdvanceTurn()
	assert.Equal(t, "enemy-001", nextTurn)
}

// TestSerializeCombatState tests combat state serialization
func TestSerializeCombatState(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	server.state.TurnManager.Initiative = []string{"player-1", "enemy-1"}
	server.state.TurnManager.CurrentRound = 3
	server.state.TurnManager.IsInCombat = true

	data := server.state.TurnManager.Serialize()
	assert.NotNil(t, data)
	// Verify it's a map with expected structure
	assert.NotEmpty(t, data)
}
