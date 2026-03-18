//go:build !js || !wasm

package wasmui

import (
	"image/color"
)

// SpriteCache provides thread-safe caching and loading of sprite images.
// This is a stub implementation for non-WASM builds (native Go testing).
type SpriteCache struct{}

// spriteCache is the global sprite cache instance.
var spriteCache *SpriteCache

// initSpriteCache initializes the global sprite cache.
func initSpriteCache() {
	if spriteCache != nil {
		return
	}
	spriteCache = &SpriteCache{}
}

// NewSpriteCache creates a new SpriteCache.
func NewSpriteCache(_ string) *SpriteCache {
	return &SpriteCache{}
}

// Get retrieves a sprite (stub returns nil).
func (sc *SpriteCache) Get(_ string) interface{} {
	return nil
}

// GetSync synchronously retrieves a sprite (stub returns nil).
func (sc *SpriteCache) GetSync(_ string) interface{} {
	return nil
}

// Preload starts loading sprites in the background (stub is no-op).
func (sc *SpriteCache) Preload(_ []string) {}

// SetLoadCallback sets a callback (stub is no-op).
func (sc *SpriteCache) SetLoadCallback(_ func(string, interface{})) {}

// IsCached returns whether a sprite is cached (stub always returns false).
func (sc *SpriteCache) IsCached(_ string) bool {
	return false
}

// ClearCache clears the cache (stub is no-op).
func (sc *SpriteCache) ClearCache() {}

// --- Asset Path Helpers (same as WASM version) ---

// CharacterPortraitPath returns the sprite path for a character portrait.
func CharacterPortraitPath(class, race, gender string) string {
	classLower := toLower(class)
	raceLower := toLower(race)
	genderLower := toLower(gender)

	classFolder := classLower + "s"
	if classLower == "thief" {
		classFolder = "thieves"
	}

	return "characters/portraits/" + classFolder + "/portrait_" + classLower + "_" + raceLower + "_" + genderLower + ".png"
}

// TerrainTilePath returns the sprite path for a terrain tile.
func TerrainTilePath(tileType, category string) string {
	categoryLower := toLower(category)
	if categoryLower == "" {
		categoryLower = "dungeon"
	}
	typeLower := toLower(tileType)
	return "terrain/" + categoryLower + "/tile_" + typeLower + ".png"
}

// MonsterSpritePath returns the sprite path for a monster.
func MonsterSpritePath(monsterType string) string {
	typeLower := toLower(monsterType)
	return "monsters/monster_" + typeLower + ".png"
}

// EffectSpritePath returns the sprite path for a combat effect.
func EffectSpritePath(effectType string) string {
	typeLower := toLower(effectType)
	return "effects/" + typeLower + "/effect_" + typeLower + ".png"
}

// ItemIconPath returns the sprite path for an item icon.
func ItemIconPath(itemType, itemName string) string {
	typeLower := toLower(itemType)
	nameLower := toLower(replaceSpaces(itemName))
	return typeLower + "/item_" + nameLower + ".png"
}

// UIElementPath returns the sprite path for a UI element.
func UIElementPath(element string) string {
	return "ui/" + toLower(element) + ".png"
}

// --- Adventure Asset Paths ---

// adventureAssetCache is a stub for the adventure asset cache.
var adventureAssetCache *SpriteCache

// initAdventureCache initializes the adventure asset cache (stub).
func initAdventureCache() {
	if adventureAssetCache != nil {
		return
	}
	adventureAssetCache = &SpriteCache{}
}

// AdventureBannerPath returns the path for an adventure's banner image.
func AdventureBannerPath(adventureSlug string) string {
	return adventureSlug + "/banner.png"
}

// AdventureNPCPath returns the path for an adventure NPC portrait.
func AdventureNPCPath(adventureSlug, npcID string) string {
	return adventureSlug + "/npc-" + npcID + ".png"
}

// AdventureItemPath returns the path for an adventure item icon.
func AdventureItemPath(adventureSlug, itemID string) string {
	return adventureSlug + "/item-" + itemID + ".png"
}

// AdventureMapPath returns the path for an adventure map background.
func AdventureMapPath(adventureSlug, mapID string) string {
	return adventureSlug + "/map-" + mapID + ".png"
}

// DrawAdventureSprite is a no-op in native builds.
func DrawAdventureSprite(_ interface{}, _ string, _, _ int) {}

// DrawAdventureSpriteScaled is a no-op in native builds.
func DrawAdventureSpriteScaled(_ interface{}, _ string, _, _, _, _ int) {}

// DrawAdventureSpriteWithFallback is a no-op in native builds.
func DrawAdventureSpriteWithFallback(_ interface{}, _ string, _, _, _, _ int, _ color.RGBA) {}

// IsAdventureAssetCached always returns false in native builds.
func IsAdventureAssetCached(_ string) bool {
	return false
}

// toLower is a simple lowercase helper for native stub.
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

// replaceSpaces replaces spaces with underscores.
func replaceSpaces(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' {
			result[i] = '_'
		} else {
			result[i] = s[i]
		}
	}
	return string(result)
}

// --- Drawing Helpers (stubs for native builds) ---

// DrawSprite is a no-op in native builds.
func DrawSprite(_ interface{}, _ string, _, _ int) {}

// DrawSpriteScaled is a no-op in native builds.
func DrawSpriteScaled(_ interface{}, _ string, _, _, _, _ int) {}

// DrawSpriteWithFallback is a no-op in native builds.
func DrawSpriteWithFallback(_ interface{}, _ string, _, _, _, _ int, _ color.RGBA) {}

// PreloadCharacterSprites is a no-op in native builds.
func PreloadCharacterSprites() {}

// PreloadTerrainSprites is a no-op in native builds.
func PreloadTerrainSprites() {}
