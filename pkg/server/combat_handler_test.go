package server

import (
	"testing"

	"goldbox-rpg/pkg/game"

	"github.com/stretchr/testify/assert"
)

// TestValidateCombatConstraintsNotInCombat tests validation when not in combat
func TestValidateCombatConstraintsNotInCombat(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	character := &game.Character{
		ID:           "test-combat-constraints",
		Name:         "Test Player",
		HP:           100,
		ActionPoints: 10,
	}
	player := &game.Player{
		Character: *character.Clone(),
	}

	server.state.TurnManager.IsInCombat = false
	err := server.validateCombatConstraints(player)
	assert.NoError(t, err, "should succeed when not in combat")
}

// TestValidateCombatConstraintsNotYourTurn tests validation when not player's turn
func TestValidateCombatConstraintsNotYourTurn(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	character := &game.Character{
		ID:           "test-wrong-turn",
		Name:         "Test Player",
		HP:           100,
		ActionPoints: 10,
	}
	player := &game.Player{
		Character: *character.Clone(),
	}

	server.state.TurnManager.IsInCombat = true
	server.state.TurnManager.Initiative = []string{"other-player"}
	server.state.TurnManager.CurrentIndex = 0

	err := server.validateCombatConstraints(player)
	assert.Error(t, err, "should fail when not player's turn")
	assert.Contains(t, err.Error(), "turn_order")
}

// TestValidateCombatConstraintsInsufficientAP tests validation with low action points
func TestValidateCombatConstraintsInsufficientAP(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	character := &game.Character{
		ID:           "test-low-ap",
		Name:         "Test Player",
		HP:           100,
		ActionPoints: 0, // No action points
	}
	player := &game.Player{
		Character: *character.Clone(),
	}

	server.state.TurnManager.IsInCombat = true
	server.state.TurnManager.Initiative = []string{player.GetID()}
	server.state.TurnManager.CurrentIndex = 0

	err := server.validateCombatConstraints(player)
	assert.Error(t, err, "should fail with insufficient action points")
	assert.Contains(t, err.Error(), "action_points")
}

// TestConsumeMovementActionPointsNotInCombat tests movement AP consumption outside combat
func TestConsumeMovementActionPointsNotInCombat(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	character := &game.Character{
		ID:           "test-move-ap",
		Name:         "Test Player",
		HP:           100,
		ActionPoints: 10,
	}
	player := &game.Player{
		Character: *character.Clone(),
	}

	server.state.TurnManager.IsInCombat = false
	initialAP := player.GetActionPoints()

	err := server.consumeMovementActionPoints(player)
	assert.NoError(t, err)
	assert.Equal(t, initialAP, player.GetActionPoints(), "AP should not change when not in combat")
}

// TestConsumeMovementActionPointsInCombat tests movement AP consumption in combat
func TestConsumeMovementActionPointsInCombat(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	character := &game.Character{
		ID:              "test-move-ap-combat",
		Name:            "Test Player",
		HP:              100,
		ActionPoints:    10,
		MaxActionPoints: 10,
	}
	player := &game.Player{
		Character: *character.Clone(),
	}

	server.state.TurnManager.IsInCombat = true
	initialAP := player.GetActionPoints()

	err := server.consumeMovementActionPoints(player)
	assert.NoError(t, err)
	assert.Less(t, player.GetActionPoints(), initialAP, "AP should decrease in combat")
}

// TestValidateEndTurnCombatStateNotInCombat tests end turn validation outside combat
func TestValidateEndTurnCombatStateNotInCombat(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	session := createTestSessionForHandlers(t, server)
	server.state.TurnManager.IsInCombat = false

	err := server.validateEndTurnCombatState(session)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in combat")
}

// TestValidateEndTurnCombatStateNotYourTurn tests end turn validation when not player's turn
func TestValidateEndTurnCombatStateNotYourTurn(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	session := createTestSessionForHandlers(t, server)
	server.state.TurnManager.IsInCombat = true
	server.state.TurnManager.Initiative = []string{"other-player"}
	server.state.TurnManager.CurrentIndex = 0

	err := server.validateEndTurnCombatState(session)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not your turn")
}

// TestValidateEndTurnCombatStateValid tests valid end turn state
func TestValidateEndTurnCombatStateValid(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	session := createTestSessionForHandlers(t, server)
	server.state.TurnManager.IsInCombat = true
	server.state.TurnManager.Initiative = []string{session.Player.GetID()}
	server.state.TurnManager.CurrentIndex = 0

	err := server.validateEndTurnCombatState(session)
	assert.NoError(t, err)
}

// TestCheckAndProcessEndRound tests round end processing
func TestCheckAndProcessEndRound(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	// When CurrentIndex is 0, it should process end of round
	server.state.TurnManager.CurrentIndex = 0
	server.state.TurnManager.CurrentRound = 1

	// Should not panic
	assert.NotPanics(t, func() {
		server.checkAndProcessEndRound()
	})
}

// TestRestoreNextPlayerActionPoints tests AP restoration for next player
func TestRestoreNextPlayerActionPoints(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	session := createTestSessionForHandlers(t, server)
	session.Player.Character.ActionPoints = 0 // Deplete AP

	server.restoreNextPlayerActionPoints(session.Player.GetID())

	// AP should be restored
	assert.Greater(t, session.Player.GetActionPoints(), 0, "action points should be restored")
}

// TestRestoreNextPlayerActionPointsEmptyID tests with empty player ID
func TestRestoreNextPlayerActionPointsEmptyID(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	// Should not panic with empty ID
	assert.NotPanics(t, func() {
		server.restoreNextPlayerActionPoints("")
	})
}

// TestExecutePlayerMovement tests player movement execution
func TestExecutePlayerMovement(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	character := &game.Character{
		ID:       "test-movement-exec",
		Name:     "Movement Test",
		HP:       100,
		Position: game.Position{X: 5, Y: 5, Level: 0},
	}
	player := &game.Player{
		Character: *character.Clone(),
	}

	// Add player to world state
	server.state.WorldState.Objects[player.GetID()] = &player.Character

	newPos := game.Position{X: 6, Y: 5, Level: 0}
	err := server.executePlayerMovement(player, newPos)
	assert.NoError(t, err)

	actualPos := player.GetPosition()
	assert.Equal(t, newPos.X, actualPos.X)
	assert.Equal(t, newPos.Y, actualPos.Y)
}

// TestCalculateNewPosition tests position calculation for all directions
func TestCalculateNewPosition(t *testing.T) {
	testCases := []struct {
		name      string
		direction game.Direction
		start     game.Position
		width     int
		height    int
		expected  game.Position
	}{
		{
			name:      "move north",
			direction: game.DirectionNorth,
			start:     game.Position{X: 5, Y: 5, Level: 0},
			width:     10,
			height:    10,
			expected:  game.Position{X: 5, Y: 4, Level: 0},
		},
		{
			name:      "move south",
			direction: game.DirectionSouth,
			start:     game.Position{X: 5, Y: 5, Level: 0},
			width:     10,
			height:    10,
			expected:  game.Position{X: 5, Y: 6, Level: 0},
		},
		{
			name:      "move east",
			direction: game.DirectionEast,
			start:     game.Position{X: 5, Y: 5, Level: 0},
			width:     10,
			height:    10,
			expected:  game.Position{X: 6, Y: 5, Level: 0},
		},
		{
			name:      "move west",
			direction: game.DirectionWest,
			start:     game.Position{X: 5, Y: 5, Level: 0},
			width:     10,
			height:    10,
			expected:  game.Position{X: 4, Y: 5, Level: 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calculateNewPosition(tc.start, tc.direction, tc.width, tc.height)
			assert.Equal(t, tc.expected.X, result.X)
			assert.Equal(t, tc.expected.Y, result.Y)
		})
	}
}

// TestCalculateNewPositionBoundary tests boundary handling
func TestCalculateNewPositionBoundary(t *testing.T) {
	// Test north boundary (should clamp at 0)
	result := calculateNewPosition(game.Position{X: 5, Y: 0, Level: 0}, game.DirectionNorth, 10, 10)
	assert.GreaterOrEqual(t, result.Y, 0, "Y should not go negative")

	// Test west boundary (should clamp at 0)
	result = calculateNewPosition(game.Position{X: 0, Y: 5, Level: 0}, game.DirectionWest, 10, 10)
	assert.GreaterOrEqual(t, result.X, 0, "X should not go negative")
}

// TestEmitTurnEndEvent tests turn end event emission
func TestEmitTurnEndEvent(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	// Should not panic
	assert.NotPanics(t, func() {
		server.emitTurnEndEvent("player-1", "player-2")
	})
}

// TestProcessEndTurnEffects tests end turn effect processing
func TestProcessEndTurnEffects(t *testing.T) {
	server := createTestServerForHandlers(t)
	defer server.Close()

	character := &game.Character{
		ID:   "effect-test-player",
		Name: "Effect Test",
		HP:   100,
	}
	player := &game.Player{
		Character: *character.Clone(),
	}

	// Add an effect
	effect := &game.Effect{
		ID:        "test-dot-effect",
		Type:      game.EffectDamageOverTime,
		Magnitude: 5,
		Stacks:    1,
	}
	player.Character.AddEffect(effect)

	// Should not panic
	assert.NotPanics(t, func() {
		server.processEndTurnEffects(player)
	})
}
