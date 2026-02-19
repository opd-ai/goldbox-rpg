package server

import (
	"fmt"
	"sync"
	"testing"

	"goldbox-rpg/pkg/game"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGameState_AddPlayer tests adding a player to game state
func TestGameState_AddPlayer(t *testing.T) {
	tests := []struct {
		name      string
		session   *PlayerSession
		checkFunc func(t *testing.T, gs *GameState)
	}{
		{
			name: "add valid player",
			session: &PlayerSession{
				SessionID: "test-session-1",
				Player: &game.Player{
					Character: game.Character{
						ID:   "player-1",
						Name: "Test Player",
					},
				},
			},
			checkFunc: func(t *testing.T, gs *GameState) {
				assert.NotNil(t, gs.WorldState)
				assert.NotNil(t, gs.WorldState.Objects)
				assert.Contains(t, gs.WorldState.Objects, "player-1")
			},
		},
		{
			name:    "add nil session",
			session: nil,
			checkFunc: func(t *testing.T, gs *GameState) {
				// Should not panic, state should remain unchanged
				if gs.WorldState != nil {
					assert.Empty(t, gs.WorldState.Objects)
				}
			},
		},
		{
			name: "add session with nil player",
			session: &PlayerSession{
				SessionID: "test-session-2",
				Player:    nil,
			},
			checkFunc: func(t *testing.T, gs *GameState) {
				// Should not panic, state should remain unchanged
				if gs.WorldState != nil {
					assert.Empty(t, gs.WorldState.Objects)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := &GameState{}
			gs.AddPlayer(tt.session)
			tt.checkFunc(t, gs)
		})
	}
}

// TestGameState_AddPlayer_Concurrent tests thread-safety of AddPlayer
func TestGameState_AddPlayer_Concurrent(t *testing.T) {
	gs := &GameState{
		WorldState: &game.World{
			Objects: make(map[string]game.GameObject),
		},
	}

	numGoroutines := 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			session := &PlayerSession{
				SessionID: "test-session",
				Player: &game.Player{
					Character: game.Character{
						ID:   fmt.Sprintf("player-%d", id),
						Name: fmt.Sprintf("Player %d", id),
					},
				},
			}
			gs.AddPlayer(session)
		}(i)
	}

	wg.Wait()

	// Verify all players were added
	assert.Equal(t, numGoroutines, len(gs.WorldState.Objects))
}

// TestGameState_GetState tests retrieving game state
func TestGameState_GetState(t *testing.T) {
	// Initialize game state with required components
	gs := &GameState{
		WorldState:  game.NewWorld(),
		TimeManager: NewTimeManager(),
		TurnManager: &TurnManager{
			Initiative: []string{},
			IsInCombat: false,
		},
		Sessions: make(map[string]*PlayerSession),
		Version:  1,
	}

	// Get state
	state := gs.GetState()

	// Verify structure
	assert.NotNil(t, state)
	assert.Contains(t, state, "world")
	assert.Contains(t, state, "time")
	assert.Contains(t, state, "turns")
	assert.Contains(t, state, "sessions")
	assert.Contains(t, state, "version")
	assert.Equal(t, 1, state["version"])
}

// TestGameState_GetState_Caching tests the caching mechanism
func TestGameState_GetState_Caching(t *testing.T) {
	gs := &GameState{
		WorldState:  game.NewWorld(),
		TimeManager: NewTimeManager(),
		TurnManager: &TurnManager{
			Initiative: []string{},
			IsInCombat: false,
		},
		Sessions: make(map[string]*PlayerSession),
		Version:  1,
	}

	// First call should generate cache
	state1 := gs.GetState()
	require.NotNil(t, state1)

	// Second call should use cache
	state2 := gs.GetState()
	require.NotNil(t, state2)

	// Verify cache is consistent
	assert.Equal(t, state1["version"], state2["version"])
}

// TestGameState_Validate tests validation of game state
func TestGameState_Validate(t *testing.T) {
	tests := []struct {
		name      string
		gameState *GameState
		wantErr   bool
	}{
		{
			name: "valid game state",
			gameState: &GameState{
				WorldState:  &game.World{},
				TimeManager: &TimeManager{},
				TurnManager: &TurnManager{},
				Sessions:    make(map[string]*PlayerSession),
			},
			wantErr: false,
		},
		{
			name: "missing WorldState",
			gameState: &GameState{
				TimeManager: &TimeManager{},
				TurnManager: &TurnManager{},
				Sessions:    make(map[string]*PlayerSession),
			},
			wantErr: true,
		},
		{
			name: "missing TimeManager",
			gameState: &GameState{
				WorldState:  &game.World{},
				TurnManager: &TurnManager{},
				Sessions:    make(map[string]*PlayerSession),
			},
			wantErr: true,
		},
		{
			name: "missing TurnManager",
			gameState: &GameState{
				WorldState:  &game.World{},
				TimeManager: &TimeManager{},
				Sessions:    make(map[string]*PlayerSession),
			},
			wantErr: true,
		},
		{
			name: "missing Sessions",
			gameState: &GameState{
				WorldState:  &game.World{},
				TimeManager: &TimeManager{},
				TurnManager: &TurnManager{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.gameState.validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// MockFileStore is a mock implementation of the FileStore interface for testing
type MockFileStore struct {
	data   map[string]interface{}
	mu     sync.RWMutex
	exists map[string]bool
}

func NewMockFileStore() *MockFileStore {
	return &MockFileStore{
		data:   make(map[string]interface{}),
		exists: make(map[string]bool),
	}
}

func (m *MockFileStore) Save(filename string, data interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[filename] = data
	m.exists[filename] = true
	return nil
}

func (m *MockFileStore) Load(filename string, dest interface{}) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if !m.exists[filename] {
		return fmt.Errorf("file not found: %s", filename)
	}
	return nil
}

func (m *MockFileStore) Exists(filename string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.exists[filename]
}

// TestGameState_SaveToFile tests saving game state to file
func TestGameState_SaveToFile(t *testing.T) {
	gs := &GameState{
		WorldState:  game.NewWorld(),
		TimeManager: NewTimeManager(),
		TurnManager: &TurnManager{
			Initiative: []string{},
			IsInCombat: false,
		},
		Sessions: make(map[string]*PlayerSession),
		Version:  1,
	}

	mockStore := NewMockFileStore()

	err := gs.SaveToFile(mockStore)
	assert.NoError(t, err)

	// Verify file was created
	assert.True(t, mockStore.Exists("gamestate.yaml"))
}

// TestGameState_LoadFromFile tests loading game state from file
func TestGameState_LoadFromFile(t *testing.T) {
	t.Run("no existing file", func(t *testing.T) {
		gs := &GameState{}
		mockStore := NewMockFileStore()

		err := gs.LoadFromFile(mockStore)
		assert.NoError(t, err) // Should not error when file doesn't exist
	})

	t.Run("file exists", func(t *testing.T) {
		gs := &GameState{}
		mockStore := NewMockFileStore()

		// Create a file
		mockStore.exists["gamestate.yaml"] = true

		err := gs.LoadFromFile(mockStore)
		assert.NoError(t, err)
	})
}

// TestTimeManager tests the TimeManager functionality
func TestTimeManager_Creation(t *testing.T) {
	tm := NewTimeManager()
	assert.NotNil(t, tm)
	assert.False(t, tm.LastTick.IsZero())
}

// TestTimeManager_Serialize tests TimeManager serialization
func TestTimeManager_Serialize(t *testing.T) {
	tm := NewTimeManager()
	serialized := tm.Serialize()
	assert.NotNil(t, serialized)
	assert.Contains(t, serialized, "current_time")
	assert.Contains(t, serialized, "time_scale")
	assert.Contains(t, serialized, "last_tick")
}

// TestTurnManager_Serialize tests TurnManager serialization
func TestTurnManager_Serialize(t *testing.T) {
	tm := &TurnManager{
		Initiative:   []string{"player1", "player2"},
		CurrentIndex: 0,
		IsInCombat:   true,
	}
	serialized := tm.Serialize()
	assert.NotNil(t, serialized)
	assert.Contains(t, serialized, "initiative_order")
	assert.Contains(t, serialized, "current_index")
	assert.Contains(t, serialized, "in_combat")
	assert.Contains(t, serialized, "current_round")
}

// TestGameState_ConcurrentAccess tests thread-safety of multiple operations
func TestGameState_ConcurrentAccess(t *testing.T) {
	gs := &GameState{
		WorldState:  game.NewWorld(),
		TimeManager: NewTimeManager(),
		TurnManager: &TurnManager{
			Initiative: []string{},
			IsInCombat: false,
		},
		Sessions: make(map[string]*PlayerSession),
		Version:  1,
	}

	numGoroutines := 20
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Half add players, half get state
	for i := 0; i < numGoroutines; i++ {
		if i%2 == 0 {
			go func(id int) {
				defer wg.Done()
				session := &PlayerSession{
					SessionID: fmt.Sprintf("session-%d", id),
					Player: &game.Player{
						Character: game.Character{
							ID:   fmt.Sprintf("player-%d", id),
							Name: fmt.Sprintf("Player %d", id),
						},
					},
				}
				gs.AddPlayer(session)
			}(i)
		} else {
			go func() {
				defer wg.Done()
				state := gs.GetState()
				assert.NotNil(t, state)
			}()
		}
	}

	wg.Wait()

	// Verify state is consistent
	finalState := gs.GetState()
	assert.NotNil(t, finalState)
}
