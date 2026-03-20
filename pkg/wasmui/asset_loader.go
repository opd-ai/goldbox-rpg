//go:build js && wasm

package wasmui

import (
	"fmt"
	"image"
	"image/color"
	_ "image/png" // Register PNG decoder
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// SpriteCache provides thread-safe caching and loading of sprite images for WASM.
// It loads PNG assets from the HTTP server and caches them for efficient reuse.
type SpriteCache struct {
	mu           sync.RWMutex
	cache        map[string]*ebiten.Image
	loading      map[string]bool
	fallbackImg  *ebiten.Image
	baseURL      string
	loadCallback func(path string, img *ebiten.Image)
}

// spriteCache is the global sprite cache instance.
var spriteCache *SpriteCache

// initSpriteCache initializes the global sprite cache (called from game init).
func initSpriteCache() {
	if spriteCache != nil {
		return
	}
	spriteCache = NewSpriteCache("/static/assets/sprites/")
}

// NewSpriteCache creates a new SpriteCache with the given base URL.
func NewSpriteCache(baseURL string) *SpriteCache {
	fallback := ebiten.NewImage(tileSize, tileSize)
	fallback.Fill(color.RGBA{R: 120, G: 80, B: 120, A: 255})

	return &SpriteCache{
		cache:       make(map[string]*ebiten.Image),
		loading:     make(map[string]bool),
		fallbackImg: fallback,
		baseURL:     baseURL,
	}
}

// Get retrieves a sprite from the cache or starts loading it asynchronously.
// Returns the cached sprite if available, or the fallback image while loading.
func (sc *SpriteCache) Get(path string) *ebiten.Image {
	sc.mu.RLock()
	if img, ok := sc.cache[path]; ok {
		sc.mu.RUnlock()
		return img
	}
	if sc.loading[path] {
		sc.mu.RUnlock()
		return sc.fallbackImg
	}
	sc.mu.RUnlock()

	// Start async load
	sc.mu.Lock()
	if sc.loading[path] {
		sc.mu.Unlock()
		return sc.fallbackImg
	}
	sc.loading[path] = true
	sc.mu.Unlock()

	go sc.loadAsync(path)
	return sc.fallbackImg
}

// GetSync synchronously retrieves or loads a sprite, blocking until loaded.
// Use sparingly - prefer Get for non-blocking access.
func (sc *SpriteCache) GetSync(path string) *ebiten.Image {
	sc.mu.RLock()
	if img, ok := sc.cache[path]; ok {
		sc.mu.RUnlock()
		return img
	}
	sc.mu.RUnlock()

	img, err := sc.loadFromURL(path)
	if err != nil {
		return sc.fallbackImg
	}

	sc.mu.Lock()
	sc.cache[path] = img
	sc.mu.Unlock()

	return img
}

// Preload starts loading a list of sprite paths in the background.
func (sc *SpriteCache) Preload(paths []string) {
	for _, path := range paths {
		sc.Get(path) // Triggers async load if not cached
	}
}

// loadAsync loads a sprite from the HTTP server asynchronously.
func (sc *SpriteCache) loadAsync(path string) {
	img, err := sc.loadFromURL(path)
	if err != nil {
		sc.mu.Lock()
		delete(sc.loading, path)
		sc.mu.Unlock()
		return
	}

	sc.mu.Lock()
	sc.cache[path] = img
	delete(sc.loading, path)
	callback := sc.loadCallback
	sc.mu.Unlock()

	if callback != nil {
		callback(path, img)
	}
}

// loadFromURL fetches and decodes a PNG image from the server.
func (sc *SpriteCache) loadFromURL(path string) (*ebiten.Image, error) {
	url := sc.baseURL + path
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s: status %d", url, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", url, err)
	}

	img, _, err := image.Decode(strings.NewReader(string(data)))
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", url, err)
	}

	return ebiten.NewImageFromImage(img), nil
}

// SetLoadCallback sets a callback to be invoked when a sprite finishes loading.
func (sc *SpriteCache) SetLoadCallback(cb func(path string, img *ebiten.Image)) {
	sc.mu.Lock()
	sc.loadCallback = cb
	sc.mu.Unlock()
}

// IsCached returns whether a sprite path is already cached.
func (sc *SpriteCache) IsCached(path string) bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	_, ok := sc.cache[path]
	return ok
}

// ClearCache removes all cached sprites.
func (sc *SpriteCache) ClearCache() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.cache = make(map[string]*ebiten.Image)
	sc.loading = make(map[string]bool)
}

// --- Asset Path Helpers ---

// CharacterPortraitPath returns the sprite path for a character portrait.
func CharacterPortraitPath(class, race, gender string) string {
	classLower := strings.ToLower(class)
	raceLower := strings.ToLower(race)
	genderLower := strings.ToLower(gender)

	// Map class to folder
	classFolder := classLower + "s" // fighter -> fighters
	if classLower == "thief" {
		classFolder = "thieves"
	}

	return fmt.Sprintf("characters/portraits/%s/portrait_%s_%s_%s.png",
		classFolder, classLower, raceLower, genderLower)
}

// TerrainTilePath returns the sprite path for a terrain tile.
func TerrainTilePath(tileType, category string) string {
	categoryLower := strings.ToLower(category)
	if categoryLower == "" {
		categoryLower = "dungeon"
	}
	typeLower := strings.ToLower(tileType)
	return fmt.Sprintf("terrain/%s/tile_%s.png", categoryLower, typeLower)
}

// MonsterSpritePath returns the sprite path for a monster.
// Monsters are stored in monsters/<category>/monster_<type>.png.
// Categories are: undead, humanoids, dragons, beasts, demons, magical.
func MonsterSpritePath(monsterType string) string {
	// Normalize monster type to lowercase with underscores
	typeLower := strings.ToLower(strings.ReplaceAll(monsterType, " ", "_"))

	// Category classification based on monster name patterns
	categoryKeywords := map[string][]string{
		"undead":    {"skeleton", "zombie", "ghoul", "vampire", "lich", "wight", "wraith"},
		"humanoids": {"goblin", "orc", "ogre", "troll", "hobgoblin"},
		"dragons":   {"dragon"},
		"beasts":    {"wolf", "spider", "rat", "bear", "wyvern", "dire_wolf", "giant_rat", "giant_spider"},
		"demons":    {"demon", "imp", "balor"},
		"magical":   {"elemental", "sprite", "wisp"},
	}

	// Find the matching category
	for category, keywords := range categoryKeywords {
		for _, keyword := range keywords {
			if strings.Contains(typeLower, keyword) {
				return fmt.Sprintf("monsters/%s/monster_%s.png", category, typeLower)
			}
		}
	}

	// Default fallback to monsters/ root
	return fmt.Sprintf("monsters/monster_%s.png", typeLower)
}

// EffectSpritePath returns the sprite path for a combat effect.
func EffectSpritePath(effectType string) string {
	typeLower := strings.ToLower(effectType)
	return fmt.Sprintf("effects/%s/effect_%s.png", typeLower, typeLower)
}

// SpellEffectPath returns the sprite path for a spell effect.
// Spell effects are named by spell ID with underscores (e.g., "magic_missile").
func SpellEffectPath(spellID string) string {
	// Normalize spell ID to lowercase with underscores
	spellLower := strings.ToLower(strings.ReplaceAll(spellID, " ", "_"))
	return fmt.Sprintf("effects/spells/effect_spell_%s.png", spellLower)
}

// ItemIconPath returns the sprite path for an item icon.
func ItemIconPath(itemType, itemName string) string {
	typeLower := strings.ToLower(itemType)
	nameLower := strings.ToLower(strings.ReplaceAll(itemName, " ", "_"))
	return fmt.Sprintf("%s/item_%s.png", typeLower, nameLower)
}

// UIElementPath returns the sprite path for a UI element.
func UIElementPath(element string) string {
	elementLower := strings.ToLower(element)
	return fmt.Sprintf("ui/%s.png", elementLower)
}

// UIButtonPath returns the sprite path for a UI button.
// Size: "small", "medium", "large"; State: "normal", "hover"
func UIButtonPath(size, state string) string {
	sizeLower := strings.ToLower(size)
	stateLower := strings.ToLower(state)
	return fmt.Sprintf("ui/buttons/ui_button_%s_%s.png", sizeLower, stateLower)
}

// UIPanelPath returns the sprite path for a UI panel background.
// Type: "character", "combat_log", "dialog_stone", "dialog_wood", "inventory"
func UIPanelPath(panelType string) string {
	typeLower := strings.ToLower(strings.ReplaceAll(panelType, " ", "_"))
	return fmt.Sprintf("ui/panels/ui_panel_%s.png", typeLower)
}

// UIIconPath returns the sprite path for a UI icon.
// Examples: "health", "mana", "attack", "defense", "strength", etc.
func UIIconPath(iconName string) string {
	nameLower := strings.ToLower(strings.ReplaceAll(iconName, " ", "_"))
	return fmt.Sprintf("ui/icons/ui_icon_%s.png", nameLower)
}

// --- Adventure Asset Paths ---

// adventureAssetCache is a separate cache for adventure assets (different base URL).
var adventureAssetCache *SpriteCache

// initAdventureCache initializes the adventure asset cache.
func initAdventureCache() {
	if adventureAssetCache != nil {
		return
	}
	adventureAssetCache = NewSpriteCache("/static/adventures/")
}

// AdventureBannerPath returns the path for an adventure's banner image.
func AdventureBannerPath(adventureSlug string) string {
	return fmt.Sprintf("%s/banner.png", adventureSlug)
}

// AdventureNPCPath returns the path for an adventure NPC portrait.
func AdventureNPCPath(adventureSlug, npcID string) string {
	return fmt.Sprintf("%s/npc-%s.png", adventureSlug, npcID)
}

// AdventureItemPath returns the path for an adventure item icon.
func AdventureItemPath(adventureSlug, itemID string) string {
	return fmt.Sprintf("%s/item-%s.png", adventureSlug, itemID)
}

// AdventureMapPath returns the path for an adventure map background.
func AdventureMapPath(adventureSlug, mapID string) string {
	return fmt.Sprintf("%s/map-%s.png", adventureSlug, mapID)
}

// DrawAdventureSprite draws an adventure asset at the given position.
func DrawAdventureSprite(screen *ebiten.Image, path string, x, y int) {
	initAdventureCache()
	img := adventureAssetCache.Get(path)
	if img == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(img, op)
}

// DrawAdventureSpriteScaled draws an adventure asset scaled to the given size.
func DrawAdventureSpriteScaled(screen *ebiten.Image, path string, x, y, w, h int) {
	initAdventureCache()
	img := adventureAssetCache.Get(path)
	if img == nil {
		return
	}

	bounds := img.Bounds()
	srcW := float64(bounds.Dx())
	srcH := float64(bounds.Dy())
	if srcW == 0 || srcH == 0 {
		return
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(w)/srcW, float64(h)/srcH)
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(img, op)
}

// DrawAdventureSpriteWithFallback draws an adventure sprite if cached, otherwise a colored rect.
func DrawAdventureSpriteWithFallback(screen *ebiten.Image, path string, x, y, w, h int, fallbackColor color.RGBA) {
	initAdventureCache()
	if adventureAssetCache.IsCached(path) {
		DrawAdventureSpriteScaled(screen, path, x, y, w, h)
	} else {
		adventureAssetCache.Get(path)
		drawRect(screen, x, y, w, h, fallbackColor)
	}
}

// IsAdventureAssetCached checks if an adventure asset is already cached.
func IsAdventureAssetCached(path string) bool {
	initAdventureCache()
	return adventureAssetCache.IsCached(path)
}

// --- Drawing Helpers ---

// DrawSprite draws a sprite at the given position on the screen.
// Uses the global sprite cache.
func DrawSprite(screen *ebiten.Image, path string, x, y int) {
	initSpriteCache()
	img := spriteCache.Get(path)
	if img == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(img, op)
}

// DrawSpriteScaled draws a sprite scaled to the given size.
func DrawSpriteScaled(screen *ebiten.Image, path string, x, y, w, h int) {
	initSpriteCache()
	img := spriteCache.Get(path)
	if img == nil {
		return
	}

	bounds := img.Bounds()
	srcW := float64(bounds.Dx())
	srcH := float64(bounds.Dy())

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(w)/srcW, float64(h)/srcH)
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(img, op)
}

// DrawSpriteWithFallback draws a sprite if cached, otherwise draws a colored rect.
func DrawSpriteWithFallback(screen *ebiten.Image, path string, x, y, w, h int, fallbackColor color.RGBA) {
	initSpriteCache()
	if spriteCache.IsCached(path) {
		DrawSpriteScaled(screen, path, x, y, w, h)
	} else {
		// Request load and draw fallback
		spriteCache.Get(path)
		drawRect(screen, x, y, w, h, fallbackColor)
	}
}

// PreloadCharacterSprites preloads common character sprites.
func PreloadCharacterSprites() {
	initSpriteCache()
	classes := []string{"fighter", "mage", "cleric", "thief", "ranger", "paladin"}
	races := []string{"human", "elf", "dwarf", "halfling"}
	genders := []string{"male", "female"}

	var paths []string
	for _, class := range classes {
		for _, race := range races {
			for _, gender := range genders {
				paths = append(paths, CharacterPortraitPath(class, race, gender))
			}
		}
	}
	spriteCache.Preload(paths)
}

// PreloadTerrainSprites preloads common terrain tiles.
func PreloadTerrainSprites() {
	initSpriteCache()
	terrainTypes := []string{
		"floor_stone", "floor_dirt", "floor_marble",
		"wall_stone", "wall_brick",
		"door_wood_closed", "door_wood_open", "door_iron",
		"chest", "barrel", "crate",
	}

	var paths []string
	for _, t := range terrainTypes {
		paths = append(paths, TerrainTilePath(t, "dungeon"))
	}
	spriteCache.Preload(paths)
}
