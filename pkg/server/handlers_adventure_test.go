package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goldbox-rpg/pkg/game"
)

func TestHandleAdventureList(t *testing.T) {
	// Create a temporary directory with a test adventure
	tmpDir := t.TempDir()
	advDir := filepath.Join(tmpDir, "test-adventure")
	require.NoError(t, os.MkdirAll(advDir, 0o755))

	advYAML := `adventure_id: test-adv-1
adventure_slug: test-adventure
adventure_title: The Test Dungeon
adventure_description: A test adventure
adventure_theme: dungeon
adventure_min_level: 1
adventure_max_level: 3
adventure_est_hours: "1-2"
adventure_author: Test Author
adventure_version: "1.0.0"
`
	require.NoError(t, os.WriteFile(filepath.Join(advDir, "adventure.yaml"), []byte(advYAML), 0o644))

	// Create adventure manager and load adventures
	advMgr := game.NewAdventureManager(tmpDir)
	require.NoError(t, advMgr.LoadAll())

	// Create a mock server with the adventure manager
	server := &RPCServer{
		adventureManager: advMgr,
	}

	// Call the handler
	result, err := server.handleAdventureList(nil)
	require.NoError(t, err)

	// Verify result
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.True(t, resultMap["success"].(bool))
	assert.Equal(t, 1, resultMap["count"].(int))

	adventures, ok := resultMap["adventures"].([]game.AdventureSummary)
	require.True(t, ok)
	require.Len(t, adventures, 1)
	assert.Equal(t, "test-adventure", adventures[0].Slug)
	assert.Equal(t, "The Test Dungeon", adventures[0].Title)
}

func TestHandleAdventureListEmpty(t *testing.T) {
	// Create adventure manager with empty directory
	tmpDir := t.TempDir()
	advMgr := game.NewAdventureManager(tmpDir)
	require.NoError(t, advMgr.LoadAll())

	server := &RPCServer{
		adventureManager: advMgr,
	}

	result, err := server.handleAdventureList(nil)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.True(t, resultMap["success"].(bool))
	assert.Equal(t, 0, resultMap["count"].(int))
}

func TestHandleAdventureListNoManager(t *testing.T) {
	server := &RPCServer{
		adventureManager: nil,
	}

	result, err := server.handleAdventureList(nil)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.True(t, resultMap["success"].(bool))
	assert.Equal(t, 0, resultMap["count"].(int))
	assert.Equal(t, "No adventures available", resultMap["message"].(string))
}

func TestHandleAdventureLoad(t *testing.T) {
	// Create a temporary directory with a test adventure
	tmpDir := t.TempDir()
	advDir := filepath.Join(tmpDir, "test-adventure")
	require.NoError(t, os.MkdirAll(advDir, 0o755))

	advYAML := `adventure_id: test-adv-1
adventure_slug: test-adventure
adventure_title: The Test Dungeon
adventure_description: A test adventure
adventure_theme: dungeon
adventure_min_level: 1
adventure_max_level: 3
adventure_est_hours: "1-2"
adventure_author: Test Author
adventure_version: "1.0.0"
adventure_maps:
  - map_id: entrance
    map_name: Dungeon Entrance
    map_width: 20
    map_height: 20
    map_file: maps/entrance.yaml
adventure_npcs:
  - npc_id: guard
    npc_name: Guard Captain
    npc_role: enemy
    npc_description: A vigilant guard
    npc_level: 2
    npc_hp: 20
    npc_hostile: true
`
	require.NoError(t, os.WriteFile(filepath.Join(advDir, "adventure.yaml"), []byte(advYAML), 0o644))

	// Create adventure manager and load adventures
	advMgr := game.NewAdventureManager(tmpDir)
	require.NoError(t, advMgr.LoadAll())

	server := &RPCServer{
		adventureManager: advMgr,
	}

	// Test successful load
	params := json.RawMessage(`{"slug": "test-adventure"}`)
	result, err := server.handleAdventureLoad(params)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.True(t, resultMap["success"].(bool))

	adventure, ok := resultMap["adventure"].(*game.Adventure)
	require.True(t, ok)
	assert.Equal(t, "test-adv-1", adventure.ID)
	assert.Equal(t, "The Test Dungeon", adventure.Title)
	assert.Len(t, adventure.Maps, 1)
	assert.Len(t, adventure.NPCs, 1)
}

func TestHandleAdventureLoadNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	advMgr := game.NewAdventureManager(tmpDir)
	require.NoError(t, advMgr.LoadAll())

	server := &RPCServer{
		adventureManager: advMgr,
	}

	params := json.RawMessage(`{"slug": "non-existent"}`)
	_, err := server.handleAdventureLoad(params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "adventure not found")
}

func TestHandleAdventureLoadMissingSlug(t *testing.T) {
	tmpDir := t.TempDir()
	advMgr := game.NewAdventureManager(tmpDir)
	require.NoError(t, advMgr.LoadAll())

	server := &RPCServer{
		adventureManager: advMgr,
	}

	params := json.RawMessage(`{}`)
	_, err := server.handleAdventureLoad(params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "slug is required")
}

func TestHandleAdventureLoadInvalidParams(t *testing.T) {
	server := &RPCServer{
		adventureManager: game.NewAdventureManager(t.TempDir()),
	}

	params := json.RawMessage(`{invalid json}`)
	_, err := server.handleAdventureLoad(params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid request parameters")
}

func TestHandleAdventureLoadNoManager(t *testing.T) {
	server := &RPCServer{
		adventureManager: nil,
	}

	params := json.RawMessage(`{"slug": "test"}`)
	_, err := server.handleAdventureLoad(params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "adventure manager not initialized")
}
