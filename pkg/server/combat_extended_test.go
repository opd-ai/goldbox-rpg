package server

import (
	"testing"

	"goldbox-rpg/pkg/game"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTurnManager_endTurn tests the endTurn method
func TestTurnManager_endTurn(t *testing.T) {
	tests := []struct {
		name               string
		setup              func() *TurnManager
		expectRoundInc     bool
		expectedCurrentIdx int
		expectedInitiative []string
	}{
		{
			name: "end turn with valid initiative - actor without action moved to top",
			setup: func() *TurnManager {
				return &TurnManager{
					CurrentRound:   1,
					CurrentIndex:   0,
					IsInCombat:     false, // Avoid timer start
					Initiative:     []string{"player1", "player2", "enemy1"},
					CombatGroups:   make(map[string][]string),
					DelayedActions: []DelayedAction{},
				}
			},
			expectRoundInc:     false,
			expectedCurrentIdx: 1, // After moving player1 to top and incrementing
			// player1 moved to top: [player1, player2, enemy1] -> stays same since already first
			// Then index advances: 0 -> 1
			expectedInitiative: []string{"player1", "player2", "enemy1"},
		},
		{
			name: "end turn with delayed action keeps initiative order",
			setup: func() *TurnManager {
				return &TurnManager{
					CurrentRound: 1,
					CurrentIndex: 0,
					IsInCombat:   false, // Avoid timer start
					Initiative:   []string{"player1", "player2"},
					CombatGroups: make(map[string][]string),
					DelayedActions: []DelayedAction{
						{ActorID: "player1", ActionType: "attack"},
					},
				}
			},
			expectRoundInc:     false,
			expectedCurrentIdx: 1,
			expectedInitiative: []string{"player1", "player2"},
		},
		{
			name: "end turn with empty initiative returns early",
			setup: func() *TurnManager {
				return &TurnManager{
					CurrentRound:   1,
					CurrentIndex:   0,
					IsInCombat:     false,
					Initiative:     []string{},
					CombatGroups:   make(map[string][]string),
					DelayedActions: []DelayedAction{},
				}
			},
			expectRoundInc:     false,
			expectedCurrentIdx: 0, // Unchanged since early return
			expectedInitiative: []string{},
		},
		{
			name: "end turn with index out of range returns early",
			setup: func() *TurnManager {
				return &TurnManager{
					CurrentRound:   1,
					CurrentIndex:   10,
					IsInCombat:     false,
					Initiative:     []string{"player1"},
					CombatGroups:   make(map[string][]string),
					DelayedActions: []DelayedAction{},
				}
			},
			expectRoundInc:     false,
			expectedCurrentIdx: 10, // Unchanged since early return
			expectedInitiative: []string{"player1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := tt.setup()
			initialRound := tm.CurrentRound

			tm.endTurn()

			if tt.expectRoundInc {
				assert.Equal(t, initialRound+1, tm.CurrentRound)
			}
			assert.Equal(t, tt.expectedCurrentIdx, tm.CurrentIndex)
			assert.Equal(t, tt.expectedInitiative, tm.Initiative)
		})
	}
}

// TestTurnManager_QueueAction tests the QueueAction method
func TestTurnManager_QueueAction(t *testing.T) {
	tests := []struct {
		name        string
		tm          *TurnManager
		action      DelayedAction
		expectError bool
	}{
		{
			name: "queue action on correct turn",
			tm: &TurnManager{
				CurrentIndex:   0,
				CurrentRound:   1,
				IsInCombat:     true,
				Initiative:     []string{"player1", "player2"},
				DelayedActions: []DelayedAction{},
			},
			action: DelayedAction{
				ActorID:    "player1",
				ActionType: "attack",
			},
			expectError: false,
		},
		{
			name: "queue action on wrong turn fails",
			tm: &TurnManager{
				CurrentIndex:   1,
				CurrentRound:   1,
				IsInCombat:     true,
				Initiative:     []string{"player1", "player2"},
				DelayedActions: []DelayedAction{},
			},
			action: DelayedAction{
				ActorID:    "player1",
				ActionType: "attack",
			},
			expectError: true,
		},
		{
			name: "queue action when not in combat fails",
			tm: &TurnManager{
				CurrentIndex:   0,
				CurrentRound:   1,
				IsInCombat:     false,
				Initiative:     []string{"player1"},
				DelayedActions: []DelayedAction{},
			},
			action: DelayedAction{
				ActorID:    "player1",
				ActionType: "spell",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialCount := len(tt.tm.DelayedActions)
			err := tt.tm.QueueAction(tt.action)

			if tt.expectError {
				assert.Error(t, err)
				assert.Len(t, tt.tm.DelayedActions, initialCount)
			} else {
				assert.NoError(t, err)
				assert.Len(t, tt.tm.DelayedActions, initialCount+1)
				lastAction := tt.tm.DelayedActions[len(tt.tm.DelayedActions)-1]
				assert.Equal(t, tt.action.ActorID, lastAction.ActorID)
				assert.Equal(t, tt.action.ActionType, lastAction.ActionType)
				// TriggerTime.GameTicks should be calculated from current round/index
				// Round 1 index 0 = (1*6 + 0)*10 = 60
				assert.Equal(t, int64(60), lastAction.TriggerTime.GameTicks)
			}
		})
	}
}

// TestTurnManager_moveToTopOfInitiative tests the moveToTopOfInitiative method
func TestTurnManager_moveToTopOfInitiative(t *testing.T) {
	tests := []struct {
		name              string
		tm                *TurnManager
		entityID          string
		expectError       bool
		expectedInitOrder []string
	}{
		{
			name: "move entity to top of initiative",
			tm: &TurnManager{
				CurrentRound: 1,
				CurrentIndex: 2,
				IsInCombat:   true,
				Initiative:   []string{"player1", "player2", "enemy1"},
				CombatGroups: make(map[string][]string),
			},
			entityID:          "enemy1",
			expectError:       false,
			expectedInitOrder: []string{"enemy1", "player1", "player2"},
		},
		{
			name: "move entity with group to top",
			tm: &TurnManager{
				CurrentRound: 1,
				CurrentIndex: 0,
				IsInCombat:   true,
				Initiative:   []string{"player1", "player2", "enemy1", "enemy2"},
				CombatGroups: map[string][]string{
					"enemy1": {"enemy2"},
				},
			},
			entityID:          "enemy1",
			expectError:       false,
			expectedInitOrder: []string{"enemy1", "enemy2", "player1", "player2"},
		},
		{
			name: "move first entity (no change)",
			tm: &TurnManager{
				CurrentRound: 1,
				CurrentIndex: 0,
				IsInCombat:   true,
				Initiative:   []string{"player1", "player2"},
				CombatGroups: make(map[string][]string),
			},
			entityID:          "player1",
			expectError:       false,
			expectedInitOrder: []string{"player1", "player2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tm.moveToTopOfInitiative(tt.entityID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedInitOrder, tt.tm.Initiative)
				assert.Equal(t, 0, tt.tm.CurrentIndex)
			}
		})
	}
}

// TestTurnManager_processDelayedActions tests the processDelayedActions method
func TestTurnManager_processDelayedActions(t *testing.T) {
	tests := []struct {
		name           string
		setup          func() *TurnManager
		expectedRemain int
	}{
		{
			name: "no delayed actions",
			setup: func() *TurnManager {
				return &TurnManager{
					CurrentRound:   1,
					CurrentIndex:   0,
					DelayedActions: []DelayedAction{},
				}
			},
			expectedRemain: 0,
		},
		{
			name: "delayed action for different turn stays",
			setup: func() *TurnManager {
				return &TurnManager{
					CurrentRound: 1,
					CurrentIndex: 0,
					DelayedActions: []DelayedAction{
						{
							ActorID:     "player1",
							ActionType:  "attack",
							TriggerTime: game.GameTime{GameTicks: 9999},
						},
					},
				}
			},
			expectedRemain: 1,
		},
		{
			name: "delayed action for current turn processed",
			setup: func() *TurnManager {
				tm := &TurnManager{
					CurrentRound: 1,
					CurrentIndex: 0,
				}
				currentTicks := tm.getCurrentGameTicks()
				tm.DelayedActions = []DelayedAction{
					{
						ActorID:     "player1",
						ActionType:  "attack",
						TriggerTime: game.GameTime{GameTicks: currentTicks},
					},
				}
				return tm
			},
			expectedRemain: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := tt.setup()
			tm.processDelayedActions()
			assert.Len(t, tm.DelayedActions, tt.expectedRemain)
		})
	}
}

// TestTurnManager_getCurrentGameTicks tests the getCurrentGameTicks method
func TestTurnManager_getCurrentGameTicks(t *testing.T) {
	tests := []struct {
		name         string
		round        int
		index        int
		expectedTick int64
	}{
		{
			name:         "first round first index",
			round:        0,
			index:        0,
			expectedTick: 0,
		},
		{
			name:         "first round second index",
			round:        0,
			index:        1,
			expectedTick: 10,
		},
		{
			name:         "round 1 index 0",
			round:        1,
			index:        0,
			expectedTick: 60,
		},
		{
			name:         "round 5 index 3",
			round:        5,
			index:        3,
			expectedTick: 330, // (5*6 + 3) * 10
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := &TurnManager{
				CurrentRound: tt.round,
				CurrentIndex: tt.index,
			}
			result := tm.getCurrentGameTicks()
			assert.Equal(t, tt.expectedTick, result)
		})
	}
}

// TestCalculateWeaponDamage tests the calculateWeaponDamage function
func TestCalculateWeaponDamage(t *testing.T) {
	tests := []struct {
		name           string
		weapon         *game.Item
		attackerStr    int
		expectedDamage int
	}{
		{
			name:           "unarmed attack with average strength",
			weapon:         nil,
			attackerStr:    10,
			expectedDamage: 1, // 1 + (10-10)/2 = 1
		},
		{
			name:           "unarmed attack with high strength",
			weapon:         nil,
			attackerStr:    18,
			expectedDamage: 5, // 1 + (18-10)/2 = 5
		},
		{
			name:           "unarmed attack with low strength (minimum 1)",
			weapon:         nil,
			attackerStr:    6,
			expectedDamage: 1, // Would be -1, clamped to 1
		},
		{
			name:           "weapon attack with strength bonus",
			weapon:         &game.Item{ID: "sword", Damage: "1d8"},
			attackerStr:    14,
			expectedDamage: 6, // base 4 (avg 1d8) + 2 str bonus
		},
		{
			name:           "weapon attack with no strength bonus",
			weapon:         &game.Item{ID: "dagger", Damage: "1d4"},
			attackerStr:    10,
			expectedDamage: 2, // base 2 (avg 1d4) + 0 str bonus
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attacker := &game.Player{
				Character: game.Character{
					Strength: tt.attackerStr,
				},
			}
			attacker.ID = "test-attacker"

			damage := calculateWeaponDamage(tt.weapon, attacker)
			assert.Equal(t, tt.expectedDamage, damage)
		})
	}
}

// TestCreateItemDrop tests the CreateItemDrop function
func TestCreateItemDrop(t *testing.T) {
	item := game.Item{
		ID:         "longsword",
		Name:       "Longsword +1",
		Type:       game.ItemTypeWeapon,
		Damage:     "1d8+1",
		Weight:     3,
		Value:      100,
		Properties: []string{"enchanted"},
	}

	char := &game.Character{
		Name: "TestHero",
	}
	char.ID = "hero-123"

	dropPos := game.Position{X: 5, Y: 10}

	result := CreateItemDrop(item, char, dropPos)

	// Verify result is an Item
	droppedItem, ok := result.(*game.Item)
	require.True(t, ok, "CreateItemDrop should return *game.Item")

	// Verify properties
	assert.Equal(t, "drop_longsword_TestHero", droppedItem.ID)
	assert.Equal(t, item.Name, droppedItem.Name)
	assert.Equal(t, item.Type, droppedItem.Type)
	assert.Equal(t, item.Damage, droppedItem.Damage)
	assert.Equal(t, item.Weight, droppedItem.Weight)
	assert.Equal(t, item.Value, droppedItem.Value)
	assert.Equal(t, dropPos, droppedItem.Position)
	assert.Contains(t, droppedItem.Properties, "enchanted")
}

// TestRPCServer_checkCombatEnd tests the checkCombatEnd method
func TestRPCServer_checkCombatEnd(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() *RPCServer
		expectEnded  bool
		expectCombat bool
	}{
		{
			name: "not in combat returns false",
			setup: func() *RPCServer {
				world := game.NewWorld()
				eventSys := game.NewEventSystem()
				return &RPCServer{
					state: &GameState{
						TurnManager: &TurnManager{
							IsInCombat:   false,
							CombatGroups: make(map[string][]string),
						},
						WorldState: world,
					},
					eventSys: eventSys,
				}
			},
			expectEnded:  false,
			expectCombat: false,
		},
		{
			name: "one hostile group ends combat",
			setup: func() *RPCServer {
				world := game.NewWorld()
				eventSys := game.NewEventSystem()
				return &RPCServer{
					state: &GameState{
						TurnManager: &TurnManager{
							IsInCombat: true,
							CombatGroups: map[string][]string{
								"team_a": {"player1", "player2"},
							},
						},
						WorldState: world,
					},
					eventSys: eventSys,
				}
			},
			expectEnded:  true,
			expectCombat: false,
		},
		{
			name: "two hostile groups continues combat",
			setup: func() *RPCServer {
				world := game.NewWorld()
				eventSys := game.NewEventSystem()
				return &RPCServer{
					state: &GameState{
						TurnManager: &TurnManager{
							IsInCombat: true,
							CombatGroups: map[string][]string{
								"team_a": {"player1"},
								"team_b": {"enemy1"},
							},
						},
						WorldState: world,
					},
					eventSys: eventSys,
				}
			},
			expectEnded:  false,
			expectCombat: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.setup()
			result := server.checkCombatEnd()

			assert.Equal(t, tt.expectEnded, result)
			assert.Equal(t, tt.expectCombat, server.state.TurnManager.IsInCombat)
		})
	}
}

// TestRPCServer_getHostileGroups tests the getHostileGroups method
func TestRPCServer_getHostileGroups(t *testing.T) {
	tests := []struct {
		name          string
		combatGroups  map[string][]string
		expectedCount int
	}{
		{
			name:          "no combat groups",
			combatGroups:  make(map[string][]string),
			expectedCount: 0,
		},
		{
			name: "one group",
			combatGroups: map[string][]string{
				"team_a": {"player1", "player2"},
			},
			expectedCount: 1,
		},
		{
			name: "two groups",
			combatGroups: map[string][]string{
				"team_a": {"player1"},
				"team_b": {"enemy1"},
			},
			expectedCount: 2,
		},
		{
			name: "three groups",
			combatGroups: map[string][]string{
				"players":  {"player1", "player2"},
				"enemies":  {"enemy1", "enemy2"},
				"neutrals": {"npc1"},
			},
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &RPCServer{
				state: &GameState{
					TurnManager: &TurnManager{
						CombatGroups: tt.combatGroups,
					},
				},
			}

			groups := server.getHostileGroups()
			assert.Len(t, groups, tt.expectedCount)
		})
	}
}

// TestRPCServer_endCombat tests the endCombat method
func TestRPCServer_endCombat(t *testing.T) {
	world := game.NewWorld()
	eventSys := game.NewEventSystem()

	server := &RPCServer{
		state: &GameState{
			TurnManager: &TurnManager{
				IsInCombat:   true,
				Initiative:   []string{"player1", "enemy1"},
				CurrentIndex: 1,
				CurrentRound: 5,
				CombatGroups: map[string][]string{
					"players": {"player1"},
					"enemies": {"enemy1"},
				},
			},
			WorldState: world,
		},
		eventSys: eventSys,
	}

	server.endCombat()

	// Verify combat ended
	assert.False(t, server.state.TurnManager.IsInCombat)
	assert.Nil(t, server.state.TurnManager.Initiative)
	assert.Equal(t, 0, server.state.TurnManager.CurrentIndex)
	// Note: Event is emitted asynchronously, state verification is sufficient
}

// TestRPCServer_applyDamage tests the applyDamage method
func TestRPCServer_applyDamage(t *testing.T) {
	tests := []struct {
		name        string
		setupTarget func() game.GameObject
		damage      int
		expectError bool
		expectedHP  int
		expectDeath bool
	}{
		{
			name: "apply damage to player",
			setupTarget: func() game.GameObject {
				p := &game.Player{
					Character: game.Character{
						HP:    50,
						MaxHP: 100,
					},
				}
				p.ID = "player1"
				return p
			},
			damage:      20,
			expectError: false,
			expectedHP:  30,
			expectDeath: false,
		},
		{
			name: "apply lethal damage to player",
			setupTarget: func() game.GameObject {
				p := &game.Player{
					Character: game.Character{
						HP:    10,
						MaxHP: 100,
					},
				}
				p.ID = "player2"
				return p
			},
			damage:      20,
			expectError: false,
			expectedHP:  0,
			expectDeath: true,
		},
		{
			name: "apply damage to character",
			setupTarget: func() game.GameObject {
				c := &game.Character{
					HP:    40,
					MaxHP: 80,
				}
				c.ID = "npc1"
				return c
			},
			damage:      15,
			expectError: false,
			expectedHP:  25,
			expectDeath: false,
		},
		{
			name: "apply damage to item fails",
			setupTarget: func() game.GameObject {
				return &game.Item{ID: "sword"}
			},
			damage:      10,
			expectError: true,
			expectedHP:  0,
			expectDeath: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := game.NewWorld()
			eventSys := game.NewEventSystem()

			server := &RPCServer{
				state: &GameState{
					WorldState: world,
					TurnManager: &TurnManager{
						CombatGroups: make(map[string][]string),
					},
				},
				eventSys: eventSys,
			}

			target := tt.setupTarget()
			err := server.applyDamage(target, tt.damage)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Check HP on appropriate type
				if player, ok := target.(*game.Player); ok {
					assert.Equal(t, tt.expectedHP, player.HP)
				} else if char, ok := target.(*game.Character); ok {
					assert.Equal(t, tt.expectedHP, char.HP)
				}

				// For death cases, verify character was set inactive
				if tt.expectDeath {
					if player, ok := target.(*game.Player); ok {
						assert.False(t, player.Character.IsActive())
					} else if char, ok := target.(*game.Character); ok {
						assert.False(t, char.IsActive())
					}
				}
			}
		})
	}
}

// TestRPCServer_processDelayedActions tests the RPCServer processDelayedActions method
func TestRPCServer_processDelayedActions(t *testing.T) {
	world := game.NewWorld()
	eventSys := game.NewEventSystem()

	server := &RPCServer{
		state: &GameState{
			WorldState: world,
			TurnManager: &TurnManager{
				CurrentRound: 1,
				CurrentIndex: 0,
				DelayedActions: []DelayedAction{
					{ActorID: "player1", ActionType: "attack", TriggerTime: game.GameTime{GameTicks: 60}},
					{ActorID: "player2", ActionType: "spell", TriggerTime: game.GameTime{GameTicks: 9999}},
				},
			},
			TimeManager: &TimeManager{
				CurrentTime: game.GameTime{GameTicks: 60},
			},
		},
		eventSys: eventSys,
	}

	server.processDelayedActions()

	// One action should be processed (matching time), one should remain
	assert.Len(t, server.state.TurnManager.DelayedActions, 1)
	assert.Equal(t, "player2", server.state.TurnManager.DelayedActions[0].ActorID)
}

// TestRPCServer_executeDelayedAction tests the executeDelayedAction method
func TestRPCServer_executeDelayedAction(t *testing.T) {
	world := game.NewWorld()
	eventSys := game.NewEventSystem()

	server := &RPCServer{
		state: &GameState{
			WorldState: world,
			TurnManager: &TurnManager{
				CurrentRound: 1,
			},
		},
		eventSys: eventSys,
	}

	action := DelayedAction{
		ActorID:    "player1",
		ActionType: "attack",
		Target:     game.Position{X: 5, Y: 5},
	}

	// Should not panic - just verify it runs
	server.executeDelayedAction(action)
}
