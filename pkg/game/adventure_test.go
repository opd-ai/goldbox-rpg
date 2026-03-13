package game

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdventureValidate(t *testing.T) {
	tests := []struct {
		name    string
		adv     Adventure
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid adventure",
			adv: Adventure{
				ID:       "test-adv",
				Slug:     "test-adventure",
				Title:    "Test Adventure",
				MinLevel: 1,
				MaxLevel: 5,
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			adv: Adventure{
				Slug:     "test",
				Title:    "Test",
				MinLevel: 1,
				MaxLevel: 5,
			},
			wantErr: true,
			errMsg:  "missing ID",
		},
		{
			name: "missing slug",
			adv: Adventure{
				ID:       "test",
				Title:    "Test",
				MinLevel: 1,
				MaxLevel: 5,
			},
			wantErr: true,
			errMsg:  "missing slug",
		},
		{
			name: "missing title",
			adv: Adventure{
				ID:       "test",
				Slug:     "test",
				MinLevel: 1,
				MaxLevel: 5,
			},
			wantErr: true,
			errMsg:  "missing title",
		},
		{
			name: "invalid min_level",
			adv: Adventure{
				ID:       "test",
				Slug:     "test",
				Title:    "Test",
				MinLevel: 0,
				MaxLevel: 5,
			},
			wantErr: true,
			errMsg:  "min_level must be >= 1",
		},
		{
			name: "max_level less than min_level",
			adv: Adventure{
				ID:       "test",
				Slug:     "test",
				Title:    "Test",
				MinLevel: 5,
				MaxLevel: 3,
			},
			wantErr: true,
			errMsg:  "max_level must be >= min_level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.adv.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAdventureSummary(t *testing.T) {
	adv := Adventure{
		ID:          "test-id",
		Slug:        "test-slug",
		Title:       "Test Adventure",
		Description: "A test adventure",
		Theme:       "dungeon",
		MinLevel:    1,
		MaxLevel:    5,
		EstHours:    "3-4",
		Maps: []AdventureMap{
			{ID: "map1", Name: "Map 1"},
			{ID: "map2", Name: "Map 2"},
		},
		QuestChain: []AdventureQuest{
			{ID: "q1", Title: "Quest 1"},
		},
	}

	summary := adv.Summary()

	assert.Equal(t, "test-id", summary.ID)
	assert.Equal(t, "test-slug", summary.Slug)
	assert.Equal(t, "Test Adventure", summary.Title)
	assert.Equal(t, "A test adventure", summary.Description)
	assert.Equal(t, "dungeon", summary.Theme)
	assert.Equal(t, 1, summary.MinLevel)
	assert.Equal(t, 5, summary.MaxLevel)
	assert.Equal(t, "3-4", summary.EstHours)
	assert.Equal(t, 2, summary.MapCount)
	assert.Equal(t, 1, summary.QuestCount)
}

func TestAdventureManager(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()
	advDir := filepath.Join(tmpDir, "test-adventure")
	require.NoError(t, os.MkdirAll(advDir, 0o755))

	// Create a valid adventure YAML file
	advYAML := `adventure_id: test-adv-1
adventure_slug: test-adventure
adventure_title: The Test Dungeon
adventure_description: A test adventure for unit testing
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
adventure_items:
  - item_id: rusty-sword
    item_name: Rusty Sword
    item_type: weapon
    item_description: An old sword
    item_rarity: common
    item_value: 5
adventure_encounters:
  - encounter_id: first-fight
    encounter_name: First Battle
    encounter_description: Initial encounter
    encounter_map_id: entrance
    encounter_min_level: 1
    encounter_max_level: 2
    encounter_enemies:
      - guard
    encounter_rewards:
      - rusty-sword
adventure_quests:
  - quest_id: explore
    quest_title: Explore the Dungeon
    quest_description: Find the entrance
    quest_order: 1
    quest_objectives:
      - objective_id: enter
        objective_description: Enter the dungeon
        objective_type: location
        objective_target: entrance
        objective_current: 0
        objective_required: 1
    quest_rewards:
      - reward_type: experience
        reward_value: 100
`
	require.NoError(t, os.WriteFile(filepath.Join(advDir, "adventure.yaml"), []byte(advYAML), 0o644))

	// Test NewAdventureManager
	mgr := NewAdventureManager(tmpDir)
	assert.NotNil(t, mgr)
	assert.Equal(t, 0, mgr.Count())

	// Test LoadAll
	err := mgr.LoadAll()
	require.NoError(t, err)
	assert.Equal(t, 1, mgr.Count())

	// Test List
	summaries := mgr.List()
	require.Len(t, summaries, 1)
	assert.Equal(t, "test-adventure", summaries[0].Slug)
	assert.Equal(t, "The Test Dungeon", summaries[0].Title)
	assert.Equal(t, 1, summaries[0].MapCount)
	assert.Equal(t, 1, summaries[0].QuestCount)

	// Test Get
	adv, err := mgr.Get("test-adventure")
	require.NoError(t, err)
	assert.Equal(t, "test-adv-1", adv.ID)
	assert.Equal(t, "The Test Dungeon", adv.Title)
	assert.Len(t, adv.Maps, 1)
	assert.Len(t, adv.NPCs, 1)
	assert.Len(t, adv.Items, 1)
	assert.Len(t, adv.Encounters, 1)
	assert.Len(t, adv.QuestChain, 1)

	// Test Get with non-existent slug
	_, err = mgr.Get("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestAdventureManagerEmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	mgr := NewAdventureManager(tmpDir)
	err := mgr.LoadAll()
	require.NoError(t, err)
	assert.Equal(t, 0, mgr.Count())
}

func TestAdventureManagerNonExistentDirectory(t *testing.T) {
	mgr := NewAdventureManager("/non/existent/path")
	err := mgr.LoadAll()
	require.NoError(t, err) // Should not error, just skip
	assert.Equal(t, 0, mgr.Count())
}

func TestAdventureManagerInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	advDir := filepath.Join(tmpDir, "bad-adventure")
	require.NoError(t, os.MkdirAll(advDir, 0o755))

	// Create invalid YAML
	require.NoError(t, os.WriteFile(filepath.Join(advDir, "adventure.yaml"), []byte("invalid: [yaml: content"), 0o644))

	mgr := NewAdventureManager(tmpDir)
	err := mgr.LoadAll()
	require.NoError(t, err) // Should not error, just warn and skip
	assert.Equal(t, 0, mgr.Count())
}

func TestAdventureManagerValidationFailure(t *testing.T) {
	tmpDir := t.TempDir()
	advDir := filepath.Join(tmpDir, "invalid-adventure")
	require.NoError(t, os.MkdirAll(advDir, 0o755))

	// Create adventure with missing required fields
	advYAML := `adventure_id: ""
adventure_title: Missing ID
adventure_min_level: 1
adventure_max_level: 5
`
	require.NoError(t, os.WriteFile(filepath.Join(advDir, "adventure.yaml"), []byte(advYAML), 0o644))

	mgr := NewAdventureManager(tmpDir)
	err := mgr.LoadAll()
	require.NoError(t, err) // Should not error, just warn and skip invalid
	assert.Equal(t, 0, mgr.Count())
}

// TestSunkenSanctum validates the reference adventure "The Sunken Sanctum"
func TestSunkenSanctum(t *testing.T) {
	// Find the data/adventures directory relative to the test
	advDir := "../../data/adventures"
	if _, err := os.Stat(advDir); os.IsNotExist(err) {
		advDir = "data/adventures"
		if _, err := os.Stat(advDir); os.IsNotExist(err) {
			t.Skip("Could not find data/adventures directory")
		}
	}

	mgr := NewAdventureManager(advDir)
	err := mgr.LoadAll()
	require.NoError(t, err)

	// Check that Sunken Sanctum loaded
	adv, err := mgr.Get("sunken-sanctum")
	require.NoError(t, err, "Sunken Sanctum adventure should load")

	// Validate adventure metadata
	assert.Equal(t, "sunken-sanctum-001", adv.ID)
	assert.Equal(t, "The Sunken Sanctum", adv.Title)
	assert.Equal(t, 1, adv.MinLevel)
	assert.Equal(t, 3, adv.MaxLevel)
	assert.Equal(t, "dungeon", adv.Theme)
	assert.Equal(t, "3-4", adv.EstHours)

	// Validate maps (requirement: ≥5 maps)
	assert.GreaterOrEqual(t, len(adv.Maps), 5, "Adventure should have at least 5 maps")

	// Validate NPCs
	assert.Greater(t, len(adv.NPCs), 0, "Adventure should have NPCs")

	// Count hostile and non-hostile NPCs
	hostileCount := 0
	friendlyCount := 0
	for _, npc := range adv.NPCs {
		if npc.Hostile {
			hostileCount++
		} else {
			friendlyCount++
		}
	}
	assert.Greater(t, hostileCount, 0, "Adventure should have hostile NPCs")
	assert.Greater(t, friendlyCount, 0, "Adventure should have friendly NPCs")

	// Validate items (requirement: ≥10 unique items)
	assert.GreaterOrEqual(t, len(adv.Items), 10, "Adventure should have at least 10 items")

	// Validate encounters
	assert.Greater(t, len(adv.Encounters), 0, "Adventure should have encounters")

	// Validate quest chain
	assert.Greater(t, len(adv.QuestChain), 0, "Adventure should have quests")

	// Validate quest order and linking
	for i, quest := range adv.QuestChain {
		assert.Equal(t, i+1, quest.Order, "Quest order should be sequential starting from 1")
		assert.NotEmpty(t, quest.Objectives, "Quest %s should have objectives", quest.ID)
	}

	// Validate the adventure passes overall validation
	err = adv.Validate()
	require.NoError(t, err, "Adventure should pass validation")
}

// TestSunkenSanctumContent validates the specific content of Sunken Sanctum
func TestSunkenSanctumContent(t *testing.T) {
	advDir := "../../data/adventures"
	if _, err := os.Stat(advDir); os.IsNotExist(err) {
		advDir = "data/adventures"
		if _, err := os.Stat(advDir); os.IsNotExist(err) {
			t.Skip("Could not find data/adventures directory")
		}
	}

	mgr := NewAdventureManager(advDir)
	err := mgr.LoadAll()
	require.NoError(t, err)

	adv, err := mgr.Get("sunken-sanctum")
	require.NoError(t, err)

	// Check for boss NPC
	hasBoss := false
	for _, npc := range adv.NPCs {
		if npc.Role == "boss" {
			hasBoss = true
			assert.Equal(t, "Deep Priest Vothus", npc.Name)
			assert.GreaterOrEqual(t, npc.HP, 40, "Boss should have significant HP")
			break
		}
	}
	assert.True(t, hasBoss, "Adventure should have a boss NPC")

	// Check for quest giver
	hasQuestGiver := false
	for _, npc := range adv.NPCs {
		if npc.Role == "quest_giver" {
			hasQuestGiver = true
			break
		}
	}
	assert.True(t, hasQuestGiver, "Adventure should have a quest giver NPC")

	// Check for keys (progression items)
	keyCount := 0
	for _, item := range adv.Items {
		if item.Type == "key" {
			keyCount++
		}
	}
	assert.GreaterOrEqual(t, keyCount, 2, "Adventure should have key items for progression")

	// Check for boss encounter
	hasBossEncounter := false
	for _, enc := range adv.Encounters {
		if enc.ID == "boss-vothus" {
			hasBossEncounter = true
			assert.Contains(t, enc.Enemies, "deep-priest-vothus")
			break
		}
	}
	assert.True(t, hasBossEncounter, "Adventure should have a boss encounter")
}
