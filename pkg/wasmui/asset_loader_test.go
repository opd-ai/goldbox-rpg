package wasmui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCharacterPortraitPath(t *testing.T) {
	tests := []struct {
		name     string
		class    string
		race     string
		gender   string
		expected string
	}{
		{
			name:     "fighter human male",
			class:    "Fighter",
			race:     "Human",
			gender:   "Male",
			expected: "characters/portraits/fighters/portrait_fighter_human_male.png",
		},
		{
			name:     "mage elf female",
			class:    "Mage",
			race:     "Elf",
			gender:   "Female",
			expected: "characters/portraits/mages/portrait_mage_elf_female.png",
		},
		{
			name:     "thief dwarf male uses thieves folder",
			class:    "Thief",
			race:     "Dwarf",
			gender:   "Male",
			expected: "characters/portraits/thieves/portrait_thief_dwarf_male.png",
		},
		{
			name:     "cleric halfling female",
			class:    "Cleric",
			race:     "Halfling",
			gender:   "Female",
			expected: "characters/portraits/clerics/portrait_cleric_halfling_female.png",
		},
		{
			name:     "lowercase input",
			class:    "ranger",
			race:     "human",
			gender:   "male",
			expected: "characters/portraits/rangers/portrait_ranger_human_male.png",
		},
		{
			name:     "paladin uppercase",
			class:    "PALADIN",
			race:     "ELF",
			gender:   "FEMALE",
			expected: "characters/portraits/paladins/portrait_paladin_elf_female.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CharacterPortraitPath(tt.class, tt.race, tt.gender)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTerrainTilePath(t *testing.T) {
	tests := []struct {
		name     string
		tileType string
		category string
		expected string
	}{
		{
			name:     "dungeon stone floor",
			tileType: "floor_stone",
			category: "dungeon",
			expected: "terrain/dungeon/tile_floor_stone.png",
		},
		{
			name:     "outdoor grass",
			tileType: "grass",
			category: "outdoor",
			expected: "terrain/outdoor/tile_grass.png",
		},
		{
			name:     "empty category defaults to dungeon",
			tileType: "door_wood_closed",
			category: "",
			expected: "terrain/dungeon/tile_door_wood_closed.png",
		},
		{
			name:     "special terrain",
			tileType: "portal",
			category: "special",
			expected: "terrain/special/tile_portal.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TerrainTilePath(tt.tileType, tt.category)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMonsterSpritePath(t *testing.T) {
	tests := []struct {
		name        string
		monsterType string
		expected    string
	}{
		{
			name:        "goblin",
			monsterType: "Goblin",
			expected:    "monsters/monster_goblin.png",
		},
		{
			name:        "dragon uppercase",
			monsterType: "DRAGON",
			expected:    "monsters/monster_dragon.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MonsterSpritePath(tt.monsterType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEffectSpritePath(t *testing.T) {
	tests := []struct {
		name       string
		effectType string
		expected   string
	}{
		{
			name:       "fire effect",
			effectType: "fire",
			expected:   "effects/fire/effect_fire.png",
		},
		{
			name:       "poison uppercase",
			effectType: "POISON",
			expected:   "effects/poison/effect_poison.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EffectSpritePath(tt.effectType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestItemIconPath(t *testing.T) {
	tests := []struct {
		name     string
		itemType string
		itemName string
		expected string
	}{
		{
			name:     "sword",
			itemType: "weapons",
			itemName: "longsword",
			expected: "weapons/item_longsword.png",
		},
		{
			name:     "item with spaces",
			itemType: "armor",
			itemName: "chain mail",
			expected: "armor/item_chain_mail.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ItemIconPath(tt.itemType, tt.itemName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUIElementPath(t *testing.T) {
	tests := []struct {
		name     string
		element  string
		expected string
	}{
		{
			name:     "button",
			element:  "button",
			expected: "ui/button.png",
		},
		{
			name:     "panel uppercase",
			element:  "PANEL",
			expected: "ui/panel.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UIElementPath(tt.element)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewSpriteCache(t *testing.T) {
	sc := NewSpriteCache("/static/")
	assert.NotNil(t, sc, "SpriteCache should be created")
}

func TestSpriteCacheIsCached(t *testing.T) {
	sc := NewSpriteCache("/static/")
	assert.False(t, sc.IsCached("nonexistent.png"), "uncached path should return false")
}

func TestSpriteCacheClearCache(t *testing.T) {
	sc := NewSpriteCache("/static/")
	sc.ClearCache() // Should not panic
}

func TestPreloadFunctions(t *testing.T) {
	// These should not panic in native builds (no-op stubs)
	PreloadCharacterSprites()
	PreloadTerrainSprites()
}
