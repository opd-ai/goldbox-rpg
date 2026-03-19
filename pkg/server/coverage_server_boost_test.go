package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"goldbox-rpg/pkg/config"
	"goldbox-rpg/pkg/game"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGameState_UpdateState tests the UpdateState method which had 0% coverage
func TestGameState_UpdateState(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *GameState
		updates map[string]interface{}
		wantErr bool
	}{
		{
			name: "empty updates",
			setup: func() *GameState {
				world := game.NewWorld()
				world.Width = 10
				world.Height = 10
				return &GameState{
					WorldState:  world,
					TimeManager: NewTimeManager(),
					TurnManager: &TurnManager{
						Initiative: []string{},
						IsInCombat: false,
					},
					Sessions: make(map[string]*PlayerSession),
					Version:  1,
				}
			},
			updates: map[string]interface{}{},
			wantErr: false,
		},
		{
			name: "time scale update",
			setup: func() *GameState {
				world := game.NewWorld()
				world.Width = 10
				world.Height = 10
				return &GameState{
					WorldState:  world,
					TimeManager: NewTimeManager(),
					TurnManager: &TurnManager{
						Initiative: []string{},
						IsInCombat: false,
					},
					Sessions: make(map[string]*PlayerSession),
					Version:  1,
				}
			},
			updates: map[string]interface{}{
				"time": map[string]interface{}{
					"current_time": map[string]interface{}{
						"time_scale": 2.0,
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := tt.setup()
			err := gs.UpdateState(tt.updates)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestGameState_applyWorldUpdates tests the world update application
func TestGameState_applyWorldUpdates(t *testing.T) {
	tests := []struct {
		name    string
		updates map[string]interface{}
		wantErr bool
	}{
		{
			name:    "no world key",
			updates: map[string]interface{}{},
			wantErr: false,
		},
		{
			name: "nil world value",
			updates: map[string]interface{}{
				"world": nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := game.NewWorld()
			world.Width = 10
			world.Height = 10
			gs := &GameState{
				WorldState:  world,
				TimeManager: NewTimeManager(),
				TurnManager: &TurnManager{},
				Sessions:    make(map[string]*PlayerSession),
			}
			err := gs.applyWorldUpdates(tt.updates)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestGameState_applyTimeUpdates tests the time update application
func TestGameState_applyTimeUpdates(t *testing.T) {
	tests := []struct {
		name           string
		updates        map[string]interface{}
		expectedScale  float64
		checkTimeScale bool
	}{
		{
			name:           "no time key",
			updates:        map[string]interface{}{},
			checkTimeScale: false,
		},
		{
			name: "with time scale",
			updates: map[string]interface{}{
				"time": map[string]interface{}{
					"current_time": map[string]interface{}{
						"time_scale": 3.5,
					},
				},
			},
			expectedScale:  3.5,
			checkTimeScale: true,
		},
		{
			name: "nil time value",
			updates: map[string]interface{}{
				"time": nil,
			},
			checkTimeScale: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := &GameState{
				TimeManager: NewTimeManager(),
			}
			err := gs.applyTimeUpdates(tt.updates)
			assert.NoError(t, err)
			if tt.checkTimeScale {
				assert.Equal(t, tt.expectedScale, gs.TimeManager.TimeScale)
			}
		})
	}
}

// TestGameState_applyTurnUpdates tests the turn update application
func TestGameState_applyTurnUpdates(t *testing.T) {
	tests := []struct {
		name    string
		updates map[string]interface{}
		wantErr bool
	}{
		{
			name:    "no turns key",
			updates: map[string]interface{}{},
			wantErr: false,
		},
		{
			name: "nil turns value",
			updates: map[string]interface{}{
				"turns": nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := &GameState{
				TurnManager: &TurnManager{
					Initiative: []string{},
				},
			}
			err := gs.applyTurnUpdates(tt.updates)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestGameState_applySessionUpdates tests the session update application
func TestGameState_applySessionUpdates(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *GameState
		updates map[string]interface{}
		wantErr bool
	}{
		{
			name: "no sessions key",
			setup: func() *GameState {
				return &GameState{
					Sessions: make(map[string]*PlayerSession),
				}
			},
			updates: map[string]interface{}{},
			wantErr: false,
		},
		{
			name: "nil sessions value",
			setup: func() *GameState {
				return &GameState{
					Sessions: make(map[string]*PlayerSession),
				}
			},
			updates: map[string]interface{}{
				"sessions": nil,
			},
			wantErr: false,
		},
		{
			name: "session not found",
			setup: func() *GameState {
				return &GameState{
					Sessions: make(map[string]*PlayerSession),
				}
			},
			updates: map[string]interface{}{
				"sessions": map[string]interface{}{
					"nonexistent": map[string]interface{}{},
				},
			},
			wantErr: false, // Should not error for non-existent session
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := tt.setup()
			err := gs.applySessionUpdates(tt.updates)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestGameState_rollback tests the rollback functionality
func TestGameState_rollback(t *testing.T) {
	t.Run("valid snapshot", func(t *testing.T) {
		world := game.NewWorld()
		world.Width = 10
		world.Height = 10
		gs := &GameState{
			WorldState:  world,
			TimeManager: NewTimeManager(),
			TurnManager: &TurnManager{
				Initiative: []string{},
			},
			Sessions: make(map[string]*PlayerSession),
			Version:  5,
		}

		// Create snapshot
		snapshot := &GameState{
			WorldState:  game.NewWorld(),
			TimeManager: NewTimeManager(),
			TurnManager: &TurnManager{
				Initiative: []string{"old"},
			},
			Sessions: make(map[string]*PlayerSession),
			Version:  3,
		}

		gs.rollback(snapshot)
		assert.Equal(t, 3, gs.Version)
	})

	t.Run("invalid snapshot type", func(t *testing.T) {
		world := game.NewWorld()
		world.Width = 10
		world.Height = 10
		gs := &GameState{
			WorldState:  world,
			TimeManager: NewTimeManager(),
			TurnManager: &TurnManager{},
			Sessions:    make(map[string]*PlayerSession),
			Version:     5,
		}

		// Should not panic with invalid snapshot type
		gs.rollback("invalid")
		assert.Equal(t, 5, gs.Version) // Version should remain unchanged
	})
}

// TestGameState_createSnapshot tests snapshot creation
func TestGameState_createSnapshot(t *testing.T) {
	world := game.NewWorld()
	world.Width = 10
	world.Height = 10
	gs := &GameState{
		WorldState:  world,
		TimeManager: NewTimeManager(),
		TurnManager: &TurnManager{
			Initiative: []string{"player1", "player2"},
		},
		Sessions: make(map[string]*PlayerSession),
		Version:  10,
	}
	gs.Sessions["test"] = &PlayerSession{
		SessionID: "test",
		Player: &game.Player{
			Character: game.Character{
				ID:   "player1",
				Name: "Test",
			},
		},
	}

	snapshot := gs.createSnapshot()
	require.NotNil(t, snapshot)

	snapshotState, ok := snapshot.(*GameState)
	require.True(t, ok)
	assert.NotNil(t, snapshotState.WorldState)
	assert.NotNil(t, snapshotState.TimeManager)
	assert.NotNil(t, snapshotState.TurnManager)
}

// TestRPCServer_applySpellDamage tests spell damage application
func TestRPCServer_applySpellDamage(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() (*RPCServer, string)
		damage     int
		damageType string
		wantErr    bool
		checkHP    bool
		expectedHP int
	}{
		{
			name: "damage to player",
			setup: func() (*RPCServer, string) {
				server := &RPCServer{
					sessions: make(map[string]*PlayerSession),
					mu:       sync.RWMutex{},
				}
				session := &PlayerSession{
					SessionID: "test-session",
					Player: &game.Player{
						Character: game.Character{
							ID:    "test-player",
							HP:    100,
							MaxHP: 100,
						},
					},
				}
				server.sessions["test-session"] = session
				return server, "test-player"
			},
			damage:     25,
			damageType: "fire",
			wantErr:    false,
			checkHP:    true,
			expectedHP: 75,
		},
		{
			name: "damage exceeds HP",
			setup: func() (*RPCServer, string) {
				server := &RPCServer{
					sessions: make(map[string]*PlayerSession),
					mu:       sync.RWMutex{},
				}
				session := &PlayerSession{
					SessionID: "test-session",
					Player: &game.Player{
						Character: game.Character{
							ID:    "test-player",
							HP:    10,
							MaxHP: 100,
						},
					},
				}
				server.sessions["test-session"] = session
				return server, "test-player"
			},
			damage:     50,
			damageType: "ice",
			wantErr:    false,
			checkHP:    true,
			expectedHP: 0, // HP capped at 0
		},
		{
			name: "NPC target (simulated)",
			setup: func() (*RPCServer, string) {
				server := &RPCServer{
					sessions: make(map[string]*PlayerSession),
					mu:       sync.RWMutex{},
				}
				return server, "npc-target"
			},
			damage:     30,
			damageType: "lightning",
			wantErr:    false,
			checkHP:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, targetID := tt.setup()
			err := server.applySpellDamage(targetID, tt.damage, tt.damageType)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.checkHP {
				// Find player and check HP
				server.mu.RLock()
				for _, session := range server.sessions {
					if session.Player.GetID() == targetID {
						assert.Equal(t, tt.expectedHP, session.Player.GetHP())
					}
				}
				server.mu.RUnlock()
			}
		})
	}
}

// TestRPCServer_applySpellHealing tests spell healing application
func TestRPCServer_applySpellHealing(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() (*RPCServer, string)
		healing    int
		wantErr    bool
		checkHP    bool
		expectedHP int
	}{
		{
			name: "heal player",
			setup: func() (*RPCServer, string) {
				server := &RPCServer{
					sessions: make(map[string]*PlayerSession),
					mu:       sync.RWMutex{},
				}
				session := &PlayerSession{
					SessionID: "test-session",
					Player: &game.Player{
						Character: game.Character{
							ID:    "test-player",
							HP:    50,
							MaxHP: 100,
						},
					},
				}
				server.sessions["test-session"] = session
				return server, "test-player"
			},
			healing:    30,
			wantErr:    false,
			checkHP:    true,
			expectedHP: 80,
		},
		{
			name: "heal capped at max HP",
			setup: func() (*RPCServer, string) {
				server := &RPCServer{
					sessions: make(map[string]*PlayerSession),
					mu:       sync.RWMutex{},
				}
				session := &PlayerSession{
					SessionID: "test-session",
					Player: &game.Player{
						Character: game.Character{
							ID:    "test-player",
							HP:    90,
							MaxHP: 100,
						},
					},
				}
				server.sessions["test-session"] = session
				return server, "test-player"
			},
			healing:    50,
			wantErr:    false,
			checkHP:    true,
			expectedHP: 100, // Capped at MaxHP
		},
		{
			name: "NPC target (simulated)",
			setup: func() (*RPCServer, string) {
				server := &RPCServer{
					sessions: make(map[string]*PlayerSession),
					mu:       sync.RWMutex{},
				}
				return server, "npc-target"
			},
			healing: 25,
			wantErr: false,
			checkHP: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, targetID := tt.setup()
			err := server.applySpellHealing(targetID, tt.healing)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.checkHP {
				// Find player and check HP
				server.mu.RLock()
				for _, session := range server.sessions {
					if session.Player.GetID() == targetID {
						assert.Equal(t, tt.expectedHP, session.Player.GetHP())
					}
				}
				server.mu.RUnlock()
			}
		})
	}
}

// TestRPCServer_executeDelayedAction_Coverage tests the delayed action execution stub
func TestRPCServer_executeDelayedAction_Coverage(t *testing.T) {
	server := &RPCServer{}

	// The function is a stub, so we just verify it doesn't panic
	action := DelayedAction{
		ActorID:    "test-entity",
		ActionType: "attack",
		Target:     game.Position{X: 5, Y: 5},
	}

	// Should not panic
	server.executeDelayedAction(action)
}

// TestRPCServer_handleApplyEffect tests the effect application handler
func TestRPCServer_handleApplyEffect(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *RPCServer
		params      string
		expectError bool
		errorMsg    string
	}{
		{
			name: "invalid JSON",
			setup: func() *RPCServer {
				return &RPCServer{
					sessions: make(map[string]*PlayerSession),
				}
			},
			params:      `{invalid json}`,
			expectError: true,
		},
		{
			name: "missing effect_type",
			setup: func() *RPCServer {
				return &RPCServer{
					sessions: make(map[string]*PlayerSession),
				}
			},
			params:      `{"session_id": "test", "target_id": "target", "magnitude": 10}`,
			expectError: true,
		},
		{
			name: "missing target_id",
			setup: func() *RPCServer {
				return &RPCServer{
					sessions: make(map[string]*PlayerSession),
				}
			},
			params:      `{"session_id": "test", "effect_type": "burning", "magnitude": 10}`,
			expectError: true,
		},
		{
			name: "invalid session",
			setup: func() *RPCServer {
				return &RPCServer{
					sessions: make(map[string]*PlayerSession),
					mu:       sync.RWMutex{},
				}
			},
			params:      `{"session_id": "nonexistent", "effect_type": "burning", "target_id": "target", "magnitude": 10}`,
			expectError: true,
		},
		{
			name: "session with nil player",
			setup: func() *RPCServer {
				s := &RPCServer{
					sessions: make(map[string]*PlayerSession),
					mu:       sync.RWMutex{},
				}
				session := &PlayerSession{
					SessionID: "test-session",
					Player:    nil,
				}
				session.addRef() // Use the proper method instead of direct field access
				s.sessions["test-session"] = session
				return s
			},
			params:      `{"session_id": "test-session", "effect_type": "burning", "target_id": "target", "magnitude": 10}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.setup()
			result, err := server.handleApplyEffect([]byte(tt.params))
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

// TestTurnManager_Update_Coverage tests the TurnManager Update method
func TestTurnManager_Update_Coverage(t *testing.T) {
	tests := []struct {
		name    string
		tm      *TurnManager
		updates map[string]interface{}
		wantErr bool
	}{
		{
			name: "update initiative order",
			tm: &TurnManager{
				Initiative: []string{"old1", "old2"},
			},
			updates: map[string]interface{}{
				"initiative_order": []interface{}{"new1", "new2", "new3"},
			},
			wantErr: false,
		},
		{
			name: "update current index",
			tm: &TurnManager{
				Initiative:   []string{"p1", "p2"},
				CurrentIndex: 0,
			},
			updates: map[string]interface{}{
				"current_index": float64(1),
			},
			wantErr: false,
		},
		{
			name: "update combat state",
			tm: &TurnManager{
				IsInCombat: false,
			},
			updates: map[string]interface{}{
				"in_combat": true,
			},
			wantErr: false,
		},
		{
			name:    "empty updates",
			tm:      &TurnManager{},
			updates: map[string]interface{}{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tm.Update(tt.updates)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSessionTimeout_EdgeCases tests session timeout edge cases
func TestSessionTimeout_EdgeCases(t *testing.T) {
	t.Run("session exactly at timeout boundary", func(t *testing.T) {
		server := &RPCServer{
			sessions: make(map[string]*PlayerSession),
			mu:       sync.RWMutex{},
			config: &config.Config{
				SessionTimeout: 1 * time.Millisecond,
			},
		}

		session := &PlayerSession{
			SessionID:  "boundary-session",
			LastActive: time.Now().Add(-2 * time.Millisecond),
			CreatedAt:  time.Now().Add(-1 * time.Hour),
		}
		server.sessions["boundary-session"] = session

		// Small delay to ensure we're past the timeout
		time.Sleep(5 * time.Millisecond)

		// Session should be considered expired
		server.mu.RLock()
		lastActive := session.LastActive
		server.mu.RUnlock()

		elapsed := time.Since(lastActive)
		assert.True(t, elapsed > server.config.SessionTimeout)
	})

	t.Run("session just before timeout", func(t *testing.T) {
		server := &RPCServer{
			sessions: make(map[string]*PlayerSession),
			mu:       sync.RWMutex{},
			config: &config.Config{
				SessionTimeout: 1 * time.Hour,
			},
		}

		session := &PlayerSession{
			SessionID:  "active-session",
			LastActive: time.Now(),
			CreatedAt:  time.Now(),
		}
		server.sessions["active-session"] = session

		// Session should not be expired
		server.mu.RLock()
		lastActive := session.LastActive
		server.mu.RUnlock()

		elapsed := time.Since(lastActive)
		assert.True(t, elapsed < server.config.SessionTimeout)
	})
}

// TestConcurrentSessionAccess tests concurrent session operations
func TestConcurrentSessionAccess(t *testing.T) {
	server := &RPCServer{
		sessions: make(map[string]*PlayerSession),
		mu:       sync.RWMutex{},
		config: &config.Config{
			SessionTimeout: 30 * time.Minute,
		},
	}

	// Pre-populate some sessions
	for i := 0; i < 5; i++ {
		sessionID := string(rune('a' + i))
		server.sessions[sessionID] = &PlayerSession{
			SessionID:   sessionID,
			LastActive:  time.Now(),
			MessageChan: make(chan []byte, 100),
			Player: &game.Player{
				Character: game.Character{
					ID:   sessionID + "-player",
					Name: "Player " + sessionID,
				},
			},
		}
	}

	var wg sync.WaitGroup
	numOps := 50

	// Concurrent reads
	for i := 0; i < numOps; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			server.mu.RLock()
			_ = len(server.sessions)
			server.mu.RUnlock()
		}()
	}

	// Concurrent session lookups
	for i := 0; i < numOps; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sessionID := string(rune('a' + (idx % 5)))
			_, _ = server.getSession(sessionID)
		}(i)
	}

	wg.Wait()

	// Verify sessions are still intact
	server.mu.RLock()
	assert.Equal(t, 5, len(server.sessions))
	server.mu.RUnlock()
}

// TestPersistenceError_Error tests the PersistenceError.Error() method variations
func TestPersistenceError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *PersistenceError
		contains []string
	}{
		{
			name: "with session ID and file path",
			err: &PersistenceError{
				SessionID: "session-123",
				FilePath:  "/path/to/file",
				Operation: "save",
				Err:       fmt.Errorf("disk full"),
			},
			contains: []string{"session-123", "/path/to/file", "save", "disk full"},
		},
		{
			name: "with session ID only",
			err: &PersistenceError{
				SessionID: "session-456",
				FilePath:  "",
				Operation: "load",
				Err:       fmt.Errorf("not found"),
			},
			contains: []string{"session-456", "load", "not found"},
		},
		{
			name: "with operation only",
			err: &PersistenceError{
				SessionID: "",
				FilePath:  "",
				Operation: "delete",
				Err:       fmt.Errorf("permission denied"),
			},
			contains: []string{"delete", "permission denied"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errStr := tt.err.Error()
			for _, substr := range tt.contains {
				assert.Contains(t, errStr, substr)
			}
		})
	}
}

// TestRPCServer_logSpellProcessingError tests the spell error logging
func TestRPCServer_logSpellProcessingError(t *testing.T) {
	server := &RPCServer{}
	// Should not panic - just logs the error
	server.logSpellProcessingError(fmt.Errorf("spell casting failed"))
}

// TestHealthChecker_RegisterCheck tests the RegisterCheck method
func TestHealthChecker_RegisterCheck(t *testing.T) {
	server := &RPCServer{}
	hc := NewHealthChecker(server)
	require.NotNil(t, hc)

	// Register a check
	checkFunc := func(ctx context.Context) error {
		return nil
	}
	hc.RegisterCheck("test-check", checkFunc)

	// Verify check was registered
	assert.NotNil(t, hc.checks["test-check"])
}

// TestHealthChecker_RunHealthChecks tests the health check execution
func TestHealthChecker_RunHealthChecks(t *testing.T) {
	server := &RPCServer{}
	hc := NewHealthChecker(server)
	require.NotNil(t, hc)

	// Register a healthy check
	hc.RegisterCheck("healthy-check", func(ctx context.Context) error {
		return nil
	})

	// Register an unhealthy check
	hc.RegisterCheck("unhealthy-check", func(ctx context.Context) error {
		return fmt.Errorf("service unavailable")
	})

	ctx := context.Background()
	response := hc.RunHealthChecks(ctx)

	assert.NotEmpty(t, response.Checks)
	assert.False(t, response.Timestamp.IsZero())
}

// TestUpdateSingleSession tests the updateSingleSession method
func TestUpdateSingleSession(t *testing.T) {
	gs := &GameState{
		Sessions: make(map[string]*PlayerSession),
	}

	// Add a session
	gs.Sessions["test-session"] = &PlayerSession{
		SessionID:  "test-session",
		Connected:  false,
		LastActive: time.Now().Add(-1 * time.Hour),
	}

	// Update session with connected status
	err := gs.updateSingleSession("test-session", map[string]interface{}{
		"connected": true,
	})
	assert.NoError(t, err)
	assert.True(t, gs.Sessions["test-session"].Connected)

	// Update non-existent session (should not error)
	err = gs.updateSingleSession("nonexistent", map[string]interface{}{})
	assert.NoError(t, err)

	// Update with non-map value (should not error)
	err = gs.updateSingleSession("test-session", "invalid")
	assert.NoError(t, err)
}

// TestExecuteDelayedAction_Coverage tests the executeDelayedAction stub
func TestExecuteDelayedAction_Coverage(t *testing.T) {
	server := &RPCServer{}

	// Create various action types
	actions := []DelayedAction{
		{
			ActorID:    "player-1",
			ActionType: "attack",
			Target:     game.Position{X: 1, Y: 1},
		},
		{
			ActorID:    "player-2",
			ActionType: "spell",
			Target:     game.Position{X: 2, Y: 2},
			Parameters: []string{"fireball", "target"},
		},
		{
			ActorID:    "npc-1",
			ActionType: "move",
			Target:     game.Position{X: 3, Y: 3},
		},
	}

	// Should not panic for any action type
	for _, action := range actions {
		server.executeDelayedAction(action)
	}
}

// TestDeepCopyMap tests the deepCopyMap function with 50% coverage
func TestDeepCopyMap(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]interface{}
		want  map[string]interface{}
	}{
		{
			name:  "nil map",
			input: nil,
			want:  nil,
		},
		{
			name:  "empty map",
			input: map[string]interface{}{},
			want:  map[string]interface{}{},
		},
		{
			name: "simple map",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			},
			want: map[string]interface{}{
				"key1": "value1",
				"key2": float64(42), // JSON unmarshals numbers as float64
			},
		},
		{
			name: "nested map",
			input: map[string]interface{}{
				"outer": map[string]interface{}{
					"inner": "value",
				},
			},
			want: map[string]interface{}{
				"outer": map[string]interface{}{
					"inner": "value",
				},
			},
		},
		{
			name: "map with array",
			input: map[string]interface{}{
				"items": []interface{}{"a", "b", "c"},
			},
			want: map[string]interface{}{
				"items": []interface{}{"a", "b", "c"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deepCopyMap(tt.input)
			assert.Equal(t, tt.want, got)

			// Verify deep copy - modifications to original don't affect copy
			if tt.input != nil && got != nil {
				tt.input["newKey"] = "newValue"
				_, exists := got["newKey"]
				assert.False(t, exists, "deep copy should be independent of original")
			}
		})
	}
}

// TestStripPort tests the stripPort function with 50% coverage
func TestStripPort(t *testing.T) {
	tests := []struct {
		name string
		host string
		want string
	}{
		{
			name: "no port",
			host: "example.com",
			want: "example.com",
		},
		{
			name: "with port",
			host: "example.com:8080",
			want: "example.com",
		},
		{
			name: "localhost with port",
			host: "localhost:3000",
			want: "localhost",
		},
		{
			name: "IPv4 with port",
			host: "192.168.1.1:8080",
			want: "192.168.1.1",
		},
		{
			name: "IPv6 no port",
			host: "[::1]",
			want: "[::1]",
		},
		{
			name: "IPv6 with port",
			host: "[::1]:8080",
			want: "[::1]",
		},
		{
			name: "IPv6 full",
			host: "[2001:db8::1]:8080",
			want: "[2001:db8::1]",
		},
		{
			name: "empty string",
			host: "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripPort(tt.host)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestExtractHost tests the extractHost function with 60% coverage
func TestExtractHost(t *testing.T) {
	tests := []struct {
		name   string
		origin string
		want   string
	}{
		{
			name:   "empty origin",
			origin: "",
			want:   "",
		},
		{
			name:   "http URL",
			origin: "http://example.com",
			want:   "example.com",
		},
		{
			name:   "https URL",
			origin: "https://example.com",
			want:   "example.com",
		},
		{
			name:   "http URL with port",
			origin: "http://example.com:8080",
			want:   "example.com",
		},
		{
			name:   "https URL with port",
			origin: "https://example.com:443",
			want:   "example.com",
		},
		{
			name:   "http URL with path",
			origin: "http://example.com/path",
			want:   "example.com",
		},
		{
			name:   "host:port pattern",
			origin: "example.com:8080",
			want:   "example.com",
		},
		{
			name:   "plain host",
			origin: "example.com",
			want:   "example.com",
		},
		{
			name:   "localhost URL",
			origin: "http://localhost:3000",
			want:   "localhost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractHost(tt.origin)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestApplyTerrainRegenerationDefaults tests default application at 50% coverage
func TestApplyTerrainRegenerationDefaults(t *testing.T) {
	server := &RPCServer{}

	tests := []struct {
		name    string
		req     *terrainRegenerationRequest
		wantReq *terrainRegenerationRequest
	}{
		{
			name: "all empty - apply all defaults",
			req:  &terrainRegenerationRequest{},
			wantReq: &terrainRegenerationRequest{
				Width:        50,
				Height:       50,
				BiomeType:    "forest",
				Density:      0.5,
				WaterLevel:   0.3,
				Connectivity: "moderate",
			},
		},
		{
			name: "partial values - apply only missing",
			req: &terrainRegenerationRequest{
				Width:     100,
				BiomeType: "desert",
			},
			wantReq: &terrainRegenerationRequest{
				Width:        100,
				Height:       50,
				BiomeType:    "desert",
				Density:      0.5,
				WaterLevel:   0.3,
				Connectivity: "moderate",
			},
		},
		{
			name: "all values set - no changes",
			req: &terrainRegenerationRequest{
				Width:        80,
				Height:       60,
				BiomeType:    "swamp",
				Density:      0.7,
				WaterLevel:   0.6,
				Connectivity: "high",
			},
			wantReq: &terrainRegenerationRequest{
				Width:        80,
				Height:       60,
				BiomeType:    "swamp",
				Density:      0.7,
				WaterLevel:   0.6,
				Connectivity: "high",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server.applyTerrainRegenerationDefaults(tt.req)
			assert.Equal(t, tt.wantReq.Width, tt.req.Width)
			assert.Equal(t, tt.wantReq.Height, tt.req.Height)
			assert.Equal(t, tt.wantReq.BiomeType, tt.req.BiomeType)
			assert.Equal(t, tt.wantReq.Density, tt.req.Density)
			assert.Equal(t, tt.wantReq.WaterLevel, tt.req.WaterLevel)
			assert.Equal(t, tt.wantReq.Connectivity, tt.req.Connectivity)
		})
	}
}

// TestApplyContentGenerationDefaults tests default application at 50% coverage
func TestApplyContentGenerationDefaults(t *testing.T) {
	server := &RPCServer{}

	tests := []struct {
		name     string
		req      *contentGenerationRequest
		wantDiff int
		wantCons bool
	}{
		{
			name:     "empty - apply defaults",
			req:      &contentGenerationRequest{},
			wantDiff: 5,
			wantCons: true,
		},
		{
			name: "difficulty set, no constraints",
			req: &contentGenerationRequest{
				Difficulty: 8,
			},
			wantDiff: 8,
			wantCons: true,
		},
		{
			name: "constraints set, no difficulty",
			req: &contentGenerationRequest{
				Constraints: map[string]interface{}{"key": "value"},
			},
			wantDiff: 5,
			wantCons: true,
		},
		{
			name: "both set",
			req: &contentGenerationRequest{
				Difficulty:  3,
				Constraints: map[string]interface{}{"key": "value"},
			},
			wantDiff: 3,
			wantCons: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server.applyContentGenerationDefaults(tt.req)
			assert.Equal(t, tt.wantDiff, tt.req.Difficulty)
			assert.True(t, tt.req.Constraints != nil == tt.wantCons)
		})
	}
}

// TestApplyCreateMapDefaults tests default application at 50% coverage
func TestApplyCreateMapDefaults(t *testing.T) {
	server := &RPCServer{}

	tests := []struct {
		name       string
		req        *createMapRequest
		wantWidth  int
		wantHeight int
	}{
		{
			name:       "empty - apply defaults",
			req:        &createMapRequest{},
			wantWidth:  20,
			wantHeight: 15,
		},
		{
			name: "width set only",
			req: &createMapRequest{
				Width: 30,
			},
			wantWidth:  30,
			wantHeight: 15,
		},
		{
			name: "height set only",
			req: &createMapRequest{
				Height: 25,
			},
			wantWidth:  20,
			wantHeight: 25,
		},
		{
			name: "both set",
			req: &createMapRequest{
				Width:  40,
				Height: 50,
			},
			wantWidth:  40,
			wantHeight: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server.applyCreateMapDefaults(tt.req)
			assert.Equal(t, tt.wantWidth, tt.req.Width)
			assert.Equal(t, tt.wantHeight, tt.req.Height)
		})
	}
}

// TestTimeManagerSerialize tests the Serialize method at 50% coverage
func TestTimeManagerSerialize(t *testing.T) {
	tests := []struct {
		name  string
		setup func() *TimeManager
	}{
		{
			name: "empty time manager",
			setup: func() *TimeManager {
				return &TimeManager{
					CurrentTime: &game.GameTime{
						RealTime:  time.Now(),
						GameTicks: 0,
						TimeScale: 1.0,
					},
					TimeScale:       1.0,
					LastTick:        time.Now(),
					ScheduledEvents: []ScheduledEvent{},
				}
			},
		},
		{
			name: "with scheduled events",
			setup: func() *TimeManager {
				return &TimeManager{
					CurrentTime: &game.GameTime{
						RealTime:  time.Now(),
						GameTicks: 100,
						TimeScale: 2.0,
					},
					TimeScale: 2.0,
					LastTick:  time.Now(),
					ScheduledEvents: []ScheduledEvent{
						{
							EventID:   "event1",
							EventType: "spawn",
							TriggerTime: game.GameTime{
								GameTicks: 200,
							},
							Parameters: []string{"monster", "goblin"},
							Repeating:  false,
						},
						{
							EventID:   "event2",
							EventType: "weather",
							TriggerTime: game.GameTime{
								GameTicks: 300,
							},
							Parameters: []string{"rain"},
							Repeating:  true,
						},
					},
				}
			},
		},
		{
			name: "with empty parameters",
			setup: func() *TimeManager {
				return &TimeManager{
					CurrentTime: &game.GameTime{
						RealTime:  time.Now(),
						GameTicks: 50,
						TimeScale: 0.5,
					},
					TimeScale: 0.5,
					LastTick:  time.Now(),
					ScheduledEvents: []ScheduledEvent{
						{
							EventID:     "event3",
							EventType:   "notification",
							TriggerTime: game.GameTime{GameTicks: 60},
							Parameters:  []string{},
							Repeating:   false,
						},
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := tt.setup()
			result := tm.Serialize()

			assert.NotNil(t, result)
			assert.Contains(t, result, "current_time")
			assert.Contains(t, result, "time_scale")
			assert.Contains(t, result, "last_tick")
			assert.Contains(t, result, "events")

			// Verify events are properly serialized
			events, ok := result["events"].([]map[string]interface{})
			assert.True(t, ok)
			assert.Equal(t, len(tm.ScheduledEvents), len(events))

			// Verify current_time structure
			currentTime, ok := result["current_time"].(map[string]interface{})
			assert.True(t, ok)
			assert.Contains(t, currentTime, "real_time")
			assert.Contains(t, currentTime, "game_ticks")
			assert.Contains(t, currentTime, "time_scale")
		})
	}
}

// TestWithRecovery tests the panic recovery middleware at 40% coverage
func TestWithRecovery(t *testing.T) {
	cfg := &config.Config{
		EnableDevMode: true,
	}
	server := &RPCServer{
		config: cfg,
	}

	t.Run("normal request - no panic", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		})

		wrapped := server.withRecovery(handler)
		req := httptest.NewRequest(http.MethodPost, "/rpc", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "success", rec.Body.String())
	})

	t.Run("panic - recovers and returns error", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		})

		wrapped := server.withRecovery(handler)
		req := httptest.NewRequest(http.MethodPost, "/rpc", nil)
		rec := httptest.NewRecorder()

		// Should not panic
		wrapped.ServeHTTP(rec, req)

		// Check correlation ID was set
		assert.NotEmpty(t, rec.Header().Get("X-Correlation-ID"))
	})

	t.Run("panic with correlation ID in context", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("another test panic")
		})

		wrapped := server.withRecovery(handler)

		// Create request with correlation ID in context
		req := httptest.NewRequest(http.MethodPost, "/rpc", nil)
		ctx := WithCorrelationID(req.Context(), "test-correlation-123")
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		// Check same correlation ID was used
		assert.Equal(t, "test-correlation-123", rec.Header().Get("X-Correlation-ID"))
	})
}

// TestExtractHostFromOrigin tests extractHostFromOrigin at 90% coverage
func TestExtractHostFromOrigin(t *testing.T) {
	tests := []struct {
		name   string
		origin string
		want   string
	}{
		{
			name:   "empty origin",
			origin: "",
			want:   "",
		},
		{
			name:   "http origin",
			origin: "http://example.com",
			want:   "example.com",
		},
		{
			name:   "https origin",
			origin: "https://secure.example.com",
			want:   "secure.example.com",
		},
		{
			name:   "origin with port",
			origin: "https://example.com:8443",
			want:   "example.com",
		},
		{
			name:   "origin with path",
			origin: "https://example.com/path/to/resource",
			want:   "example.com",
		},
		{
			name:   "localhost",
			origin: "http://localhost:3000",
			want:   "localhost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractHostFromOrigin(tt.origin)
			assert.Equal(t, tt.want, got)
		})
	}
}
