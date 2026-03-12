package server

import (
	"testing"
	"time"

	"goldbox-rpg/pkg/game"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWebSocketEditorBroadcaster tests the editor broadcaster functionality
func TestWebSocketEditorBroadcaster(t *testing.T) {
	// Create a test server
	server, err := NewRPCServer(":8080")
	require.NoError(t, err, "Failed to create server")
	defer server.Stop()

	t.Run("NewEditorBroadcaster", func(t *testing.T) {
		broadcaster := NewEditorBroadcaster(server)
		assert.NotNil(t, broadcaster)
		assert.False(t, broadcaster.active)
		assert.NotNil(t, broadcaster.sessions)
	})

	t.Run("StartStop", func(t *testing.T) {
		broadcaster := NewEditorBroadcaster(server)
		
		// Start
		broadcaster.Start()
		assert.True(t, broadcaster.active)
		
		// Start again (should be idempotent)
		broadcaster.Start()
		assert.True(t, broadcaster.active)
		
		// Stop
		broadcaster.Stop()
		assert.False(t, broadcaster.active)
	})

	t.Run("RegisterUnregisterSession", func(t *testing.T) {
		broadcaster := NewEditorBroadcaster(server)
		broadcaster.Start()
		defer broadcaster.Stop()

		session := &EditorSession{
			SessionID: "test-session-1",
			MapID:     "test-map-1",
		}

		broadcaster.RegisterSession(session)
		
		broadcaster.mu.RLock()
		_, exists := broadcaster.sessions["test-session-1"]
		broadcaster.mu.RUnlock()
		assert.True(t, exists, "Session should be registered")

		broadcaster.UnregisterSession("test-session-1")
		
		broadcaster.mu.RLock()
		_, exists = broadcaster.sessions["test-session-1"]
		broadcaster.mu.RUnlock()
		assert.False(t, exists, "Session should be unregistered")
	})

	t.Run("BroadcastTileUpdate", func(t *testing.T) {
		broadcaster := NewEditorBroadcaster(server)
		broadcaster.Start()
		defer broadcaster.Stop()

		// Create sessions (without actual WebSocket connections)
		session1 := &EditorSession{
			SessionID: "session-1",
			MapID:     "map-1",
		}
		session2 := &EditorSession{
			SessionID: "session-2",
			MapID:     "map-1",
		}

		broadcaster.RegisterSession(session1)
		broadcaster.RegisterSession(session2)

		// This should not panic even without WebSocket connections
		data := TileUpdateData{
			X:           5,
			Y:           10,
			SpriteX:     1,
			SpriteY:     0,
			Walkable:    false,
			Transparent: false,
		}
		
		// Should not panic
		broadcaster.BroadcastTileUpdate("map-1", "session-1", data)
	})

	t.Run("BroadcastWhenInactive", func(t *testing.T) {
		broadcaster := NewEditorBroadcaster(server)
		// Don't start - should be inactive

		data := TileUpdateData{
			X: 0,
			Y: 0,
		}

		// Should not panic when inactive
		broadcaster.BroadcastTileUpdate("map-1", "session-1", data)
	})
}

// TestEditorSession tests the editor session functionality
func TestEditorSession(t *testing.T) {
	t.Run("CreateEditorSession", func(t *testing.T) {
		session := &EditorSession{
			SessionID:   "test-session",
			MapID:       "test-map",
			subscribers: make(map[string]*EditorSession),
		}

		assert.Equal(t, "test-session", session.SessionID)
		assert.Equal(t, "test-map", session.MapID)
		assert.Nil(t, session.CurrentMap)
	})

	t.Run("SessionWithMap", func(t *testing.T) {
		gameMap := &game.GameMap{
			Width:  20,
			Height: 15,
			Tiles:  make([][]game.MapTile, 15),
		}
		for y := 0; y < 15; y++ {
			gameMap.Tiles[y] = make([]game.MapTile, 20)
		}

		session := &EditorSession{
			SessionID:  "test-session",
			MapID:      "test-map",
			CurrentMap: gameMap,
		}

		assert.NotNil(t, session.CurrentMap)
		assert.Equal(t, 20, session.CurrentMap.Width)
		assert.Equal(t, 15, session.CurrentMap.Height)
	})

	t.Run("SendEditorUpdateWithoutConnection", func(t *testing.T) {
		session := &EditorSession{
			SessionID: "test-session",
			MapID:     "test-map",
			WSConn:    nil, // No connection
		}

		// Should not error without connection
		err := session.SendEditorUpdate(EditorEventTileUpdate, map[string]interface{}{
			"x": 5,
			"y": 10,
		})
		assert.NoError(t, err)
	})
}

// TestEditorMessage tests the editor message structures
func TestEditorMessage(t *testing.T) {
	t.Run("CreateEditorMessage", func(t *testing.T) {
		msg := EditorMessage{
			Type:      EditorEventTileUpdate,
			MapID:     "map-123",
			SessionID: "session-456",
			Data: map[string]interface{}{
				"x":           5,
				"y":           10,
				"sprite_x":    1,
				"sprite_y":    0,
				"walkable":    false,
				"transparent": false,
			},
		}

		assert.Equal(t, EditorEventTileUpdate, msg.Type)
		assert.Equal(t, "map-123", msg.MapID)
		assert.Equal(t, "session-456", msg.SessionID)
		assert.Equal(t, float64(5), msg.Data["x"])
	})

	t.Run("TileUpdateData", func(t *testing.T) {
		data := TileUpdateData{
			X:           10,
			Y:           20,
			SpriteX:     2,
			SpriteY:     1,
			Walkable:    true,
			Transparent: false,
		}

		assert.Equal(t, 10, data.X)
		assert.Equal(t, 20, data.Y)
		assert.Equal(t, 2, data.SpriteX)
		assert.Equal(t, 1, data.SpriteY)
		assert.True(t, data.Walkable)
		assert.False(t, data.Transparent)
	})
}

// TestEditorEventTypes tests the editor event type constants
func TestEditorEventTypes(t *testing.T) {
	eventTypes := []EditorEventType{
		EditorEventTileUpdate,
		EditorEventMapCreated,
		EditorEventMapLoaded,
		EditorEventMapSaved,
		EditorEventCursorMove,
		EditorEventSelectTool,
		EditorEventUndoRedo,
	}

	for _, et := range eventTypes {
		assert.NotEmpty(t, string(et), "Event type should have a string value")
	}

	// Verify specific values
	assert.Equal(t, EditorEventType("tile_update"), EditorEventTileUpdate)
	assert.Equal(t, EditorEventType("map_created"), EditorEventMapCreated)
}

// TestEditorIntegrationWithServer tests that the editor broadcaster is properly initialized
func TestEditorIntegrationWithServer(t *testing.T) {
	server, err := NewRPCServer(":8080")
	require.NoError(t, err, "Failed to create server")
	defer server.Stop()

	// Wait for server initialization
	time.Sleep(100 * time.Millisecond)

	// Editor broadcaster should be initialized
	assert.NotNil(t, server.editorBroadcaster, "Editor broadcaster should be initialized")
	assert.True(t, server.editorBroadcaster.active, "Editor broadcaster should be active")
}
