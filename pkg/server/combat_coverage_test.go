package server

import (
	"testing"

	"goldbox-rpg/pkg/game"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTurnManager_Update tests the TurnManager Update method
func TestTurnManager_Update(t *testing.T) {
	tests := []struct {
		name        string
		initial     *TurnManager
		updates     map[string]interface{}
		expectError bool
		validate    func(t *testing.T, tm *TurnManager)
	}{
		{
			name: "update current round",
			initial: &TurnManager{
				CurrentRound: 1,
				Initiative:   []string{"player1"},
				IsInCombat:   true,
			},
			updates: map[string]interface{}{
				"current_round": 5,
			},
			expectError: false,
			validate: func(t *testing.T, tm *TurnManager) {
				assert.Equal(t, 5, tm.CurrentRound)
			},
		},
		{
			name: "update current index",
			initial: &TurnManager{
				CurrentIndex: 0,
				Initiative:   []string{"player1", "player2"},
				IsInCombat:   true,
			},
			updates: map[string]interface{}{
				"current_index": 1,
			},
			expectError: false,
			validate: func(t *testing.T, tm *TurnManager) {
				assert.Equal(t, 1, tm.CurrentIndex)
			},
		},
		{
			name: "update in combat flag",
			initial: &TurnManager{
				IsInCombat: false,
				Initiative: []string{},
			},
			updates: map[string]interface{}{
				"in_combat": true,
			},
			expectError: false,
			validate: func(t *testing.T, tm *TurnManager) {
				assert.True(t, tm.IsInCombat)
			},
		},
		{
			name: "update initiative order",
			initial: &TurnManager{
				Initiative: []string{"player1"},
				IsInCombat: true,
			},
			updates: map[string]interface{}{
				"initiative_order": []string{"player2", "player1"},
			},
			expectError: false,
			validate: func(t *testing.T, tm *TurnManager) {
				assert.Equal(t, []string{"player2", "player1"}, tm.Initiative)
			},
		},
		{
			name: "update combat groups",
			initial: &TurnManager{
				Initiative:   []string{},
				CombatGroups: make(map[string][]string),
			},
			updates: map[string]interface{}{
				"combat_groups": map[string][]string{
					"team_a": {"player1", "player2"},
					"team_b": {"enemy1"},
				},
			},
			expectError: false,
			validate: func(t *testing.T, tm *TurnManager) {
				assert.Equal(t, []string{"player1", "player2"}, tm.CombatGroups["team_a"])
				assert.Equal(t, []string{"enemy1"}, tm.CombatGroups["team_b"])
			},
		},
		{
			name: "update delayed actions",
			initial: &TurnManager{
				Initiative:     []string{},
				DelayedActions: []DelayedAction{},
			},
			updates: map[string]interface{}{
				"delayed_actions": []DelayedAction{
					{ActorID: "player1", ActionType: "spell"},
				},
			},
			expectError: false,
			validate: func(t *testing.T, tm *TurnManager) {
				require.Len(t, tm.DelayedActions, 1)
				assert.Equal(t, "player1", tm.DelayedActions[0].ActorID)
			},
		},
		{
			name: "update multiple fields at once",
			initial: &TurnManager{
				CurrentRound: 1,
				CurrentIndex: 0,
				IsInCombat:   false,
				Initiative:   []string{},
			},
			updates: map[string]interface{}{
				"current_round": 3,
				"current_index": 2,
				"in_combat":     true,
			},
			expectError: false,
			validate: func(t *testing.T, tm *TurnManager) {
				assert.Equal(t, 3, tm.CurrentRound)
				assert.Equal(t, 2, tm.CurrentIndex)
				assert.True(t, tm.IsInCombat)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.initial.Update(tt.updates)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, tt.initial)
				}
			}
		})
	}
}

// TestTurnManager_Clone tests the TurnManager Clone method
func TestTurnManager_Clone(t *testing.T) {
	original := &TurnManager{
		CurrentRound:   5,
		CurrentIndex:   2,
		IsInCombat:     true,
		Initiative:     []string{"player1", "player2", "enemy1"},
		CombatGroups:   map[string][]string{"allies": {"player1", "player2"}, "enemies": {"enemy1"}},
		DelayedActions: []DelayedAction{{ActorID: "player1", ActionType: "spell", Target: game.Position{X: 5, Y: 5}}},
	}

	clone := original.Clone()

	// Verify clone has same values
	assert.Equal(t, original.CurrentRound, clone.CurrentRound)
	assert.Equal(t, original.CurrentIndex, clone.CurrentIndex)
	assert.Equal(t, original.IsInCombat, clone.IsInCombat)
	assert.Equal(t, original.Initiative, clone.Initiative)
	assert.Equal(t, original.CombatGroups, clone.CombatGroups)
	assert.Equal(t, len(original.DelayedActions), len(clone.DelayedActions))

	// Verify modifications to clone don't affect original
	clone.CurrentRound = 10
	clone.Initiative[0] = "modified"
	clone.CombatGroups["allies"] = []string{"modified"}

	assert.Equal(t, 5, original.CurrentRound)
	assert.Equal(t, "player1", original.Initiative[0])
	assert.Equal(t, []string{"player1", "player2"}, original.CombatGroups["allies"])
}

// TestTurnManager_ValidateInitiativeOrder tests initiative validation
func TestTurnManager_ValidateInitiativeOrder(t *testing.T) {
	tests := []struct {
		name        string
		tm          *TurnManager
		initiative  []string
		expectError bool
	}{
		{
			name: "valid initiative - outside combat",
			tm: &TurnManager{
				IsInCombat: false,
			},
			initiative:  []string{"player1", "player2"},
			expectError: false,
		},
		{
			name: "nil initiative",
			tm: &TurnManager{
				IsInCombat: false,
			},
			initiative:  nil,
			expectError: true,
		},
		{
			name: "duplicate entries",
			tm: &TurnManager{
				IsInCombat: true,
			},
			initiative:  []string{"player1", "player1"},
			expectError: true,
		},
		{
			name: "empty string in initiative",
			tm: &TurnManager{
				IsInCombat: true,
			},
			initiative:  []string{"player1", ""},
			expectError: true,
		},
		{
			name: "valid unique entries",
			tm: &TurnManager{
				IsInCombat: true,
			},
			initiative:  []string{"player1", "player2", "enemy1"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tm.validateInitiativeOrder(tt.initiative)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestEndCombat tests the EndCombat method
func TestEndCombat(t *testing.T) {
	tm := &TurnManager{
		CurrentRound:   5,
		CurrentIndex:   2,
		IsInCombat:     true,
		Initiative:     []string{"player1", "enemy1"},
		CombatGroups:   map[string][]string{"allies": {"player1"}, "enemies": {"enemy1"}},
		DelayedActions: []DelayedAction{{ActorID: "player1"}},
	}

	tm.EndCombat()

	// EndCombat sets IsInCombat to false and clears Initiative and CurrentIndex
	assert.False(t, tm.IsInCombat)
	assert.Equal(t, 0, tm.CurrentIndex)
	assert.Nil(t, tm.Initiative)
	// Note: CurrentRound, CombatGroups, and DelayedActions are NOT cleared by EndCombat
}

// TestIsCurrentTurn tests the IsCurrentTurn method
func TestIsCurrentTurn(t *testing.T) {
	tests := []struct {
		name       string
		tm         *TurnManager
		playerID   string
		expectTrue bool
	}{
		{
			name: "player's turn",
			tm: &TurnManager{
				Initiative:   []string{"player1", "player2"},
				CurrentIndex: 0,
				IsInCombat:   true,
			},
			playerID:   "player1",
			expectTrue: true,
		},
		{
			name: "not player's turn",
			tm: &TurnManager{
				Initiative:   []string{"player1", "player2"},
				CurrentIndex: 1,
				IsInCombat:   true,
			},
			playerID:   "player1",
			expectTrue: false,
		},
		{
			name: "not in combat returns false",
			tm: &TurnManager{
				Initiative:   []string{"player1", "player2"},
				CurrentIndex: 0,
				IsInCombat:   false,
			},
			playerID:   "player1",
			expectTrue: false, // When not in combat, IsCurrentTurn returns false
		},
		{
			name: "empty initiative",
			tm: &TurnManager{
				Initiative:   []string{},
				CurrentIndex: 0,
				IsInCombat:   true,
			},
			playerID:   "player1",
			expectTrue: false,
		},
		{
			name: "index out of range",
			tm: &TurnManager{
				Initiative:   []string{"player1"},
				CurrentIndex: 5,
				IsInCombat:   true,
			},
			playerID:   "player1",
			expectTrue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.tm.IsCurrentTurn(tt.playerID)
			assert.Equal(t, tt.expectTrue, result)
		})
	}
}
