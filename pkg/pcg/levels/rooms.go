package levels

import (
	"fmt"
	"math/rand"

	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/pcg"
)

// CombatRoomGenerator creates combat encounter rooms with tactical features.
// Generated rooms include cover positions, elevated areas, traps, and hazards
// to create interesting tactical combat scenarios. Enemy types and counts scale
// with difficulty level, and loot chances increase accordingly.
type CombatRoomGenerator struct{}

// GenerateRoom creates a combat encounter room with tactical features, enemy spawn
// configurations, and loot tables scaled by difficulty. The room includes walls,
// walkable floor tiles, and 1-3 doors positioned randomly on walls. Tactical features
// such as cover, elevation changes, and environmental hazards are placed based on
// the theme and difficulty level.
func (crg *CombatRoomGenerator) GenerateRoom(bounds pcg.Rectangle, theme pcg.LevelTheme, difficulty int, genCtx *pcg.GenerationContext) (*pcg.RoomLayout, error) {
	room := &pcg.RoomLayout{
		Type:       pcg.RoomTypeCombat,
		Bounds:     bounds,
		Tiles:      make([][]game.Tile, bounds.Height),
		Doors:      []game.Position{},
		Features:   []pcg.RoomFeature{},
		Properties: make(map[string]interface{}),
	}

	// Initialize room tiles
	for y := 0; y < bounds.Height; y++ {
		room.Tiles[y] = make([]game.Tile, bounds.Width)
		for x := 0; x < bounds.Width; x++ {
			// Create basic floor with walls on edges
			if x == 0 || x == bounds.Width-1 || y == 0 || y == bounds.Height-1 {
				room.Tiles[y][x] = game.Tile{
					Type:       game.TileWall,
					Walkable:   false,
					Properties: make(map[string]interface{}),
				}
			} else {
				room.Tiles[y][x] = game.Tile{
					Type:       game.TileFloor,
					Walkable:   true,
					Properties: make(map[string]interface{}),
				}
			}
		}
	}

	// Add tactical features based on theme and difficulty
	rng := genCtx.RNG
	featureCount := 1 + difficulty/4 + rng.Intn(3)

	for i := 0; i < featureCount; i++ {
		feature := crg.generateTacticalFeature(bounds, theme, difficulty, rng)
		room.Features = append(room.Features, feature)
	}

	// Add doors
	room.Doors = append(room.Doors, crg.generateDoorPositions(bounds, rng)...)

	// Set combat-specific properties
	room.Properties["enemy_count"] = 2 + difficulty/3
	room.Properties["enemy_types"] = crg.selectEnemyTypes(theme, difficulty)
	room.Properties["loot_chance"] = 0.3 + float64(difficulty)*0.02

	return room, nil
}

func (crg *CombatRoomGenerator) generateTacticalFeature(bounds pcg.Rectangle, theme pcg.LevelTheme, difficulty int, rng *rand.Rand) pcg.RoomFeature {
	// Generate random position inside room (not on walls)
	x := 1 + rng.Intn(bounds.Width-2)
	y := 1 + rng.Intn(bounds.Height-2)

	features := []string{"cover", "elevation", "trap", "hazard"}
	featureType := features[rng.Intn(len(features))]

	return pcg.RoomFeature{
		Type:     featureType,
		Position: game.Position{X: bounds.X + x, Y: bounds.Y + y},
		Properties: map[string]interface{}{
			"theme":      theme,
			"difficulty": difficulty,
		},
	}
}

func (crg *CombatRoomGenerator) generateDoorPositions(bounds pcg.Rectangle, rng *rand.Rand) []game.Position {
	var doors []game.Position

	// Add 1-3 doors
	doorCount := 1 + rng.Intn(3)
	for i := 0; i < doorCount; i++ {
		// Choose random wall
		wall := rng.Intn(4)
		var x, y int

		switch wall {
		case 0: // Top wall
			x = 1 + rng.Intn(bounds.Width-2)
			y = 0
		case 1: // Right wall
			x = bounds.Width - 1
			y = 1 + rng.Intn(bounds.Height-2)
		case 2: // Bottom wall
			x = 1 + rng.Intn(bounds.Width-2)
			y = bounds.Height - 1
		case 3: // Left wall
			x = 0
			y = 1 + rng.Intn(bounds.Height-2)
		}

		doors = append(doors, game.Position{X: bounds.X + x, Y: bounds.Y + y})
	}

	return doors
}

func (crg *CombatRoomGenerator) selectEnemyTypes(theme pcg.LevelTheme, difficulty int) []string {
	var enemies []string

	switch theme {
	case pcg.ThemeClassic:
		enemies = []string{"goblin", "orc", "skeleton"}
	case pcg.ThemeHorror:
		enemies = []string{"zombie", "wraith", "shadow"}
	case pcg.ThemeNatural:
		enemies = []string{"wolf", "bear", "spider"}
	case pcg.ThemeMechanical:
		enemies = []string{"construct", "golem", "automaton"}
	case pcg.ThemeMagical:
		enemies = []string{"elemental", "sprite", "wisp"}
	case pcg.ThemeUndead:
		enemies = []string{"skeleton", "zombie", "lich"}
	case pcg.ThemeElemental:
		enemies = []string{"fire_elemental", "water_elemental", "earth_elemental"}
	default:
		enemies = []string{"goblin", "orc", "bandit"}
	}

	// Scale with difficulty
	if difficulty > 10 {
		enemies = append(enemies, "elite_"+enemies[0])
	}

	return enemies
}

// TreasureRoomGenerator creates treasure and loot rooms with valuable contents.
// Generated rooms feature ornate decorations, treasure containers with rarity
// scaled by difficulty, and optional guardians for high-value rooms.
type TreasureRoomGenerator struct{}

// GenerateRoom creates a treasure room with valuable contents scaled by difficulty.
// Higher difficulty rooms may have locked/trapped chests, rare loot, and guardian
// creatures. Rooms feature decorated walls and polished floors with single secure
// entry points. Treasure value and locking requirements increase with difficulty.
func (trg *TreasureRoomGenerator) GenerateRoom(bounds pcg.Rectangle, theme pcg.LevelTheme, difficulty int, genCtx *pcg.GenerationContext) (*pcg.RoomLayout, error) {
	room := &pcg.RoomLayout{
		Type:       pcg.RoomTypeTreasure,
		Bounds:     bounds,
		Tiles:      make([][]game.Tile, bounds.Height),
		Doors:      []game.Position{},
		Features:   []pcg.RoomFeature{},
		Properties: make(map[string]interface{}),
	}

	// Initialize room tiles with ornate decoration
	for y := 0; y < bounds.Height; y++ {
		room.Tiles[y] = make([]game.Tile, bounds.Width)
		for x := 0; x < bounds.Width; x++ {
			if x == 0 || x == bounds.Width-1 || y == 0 || y == bounds.Height-1 {
				room.Tiles[y][x] = game.Tile{
					Type:       game.TileWall,
					Walkable:   false,
					Properties: map[string]interface{}{"decorated": true},
				}
			} else {
				room.Tiles[y][x] = game.Tile{
					Type:       game.TileFloor,
					Walkable:   true,
					Properties: map[string]interface{}{"polished": true},
				}
			}
		}
	}

	rng := genCtx.RNG

	// Add treasure containers
	treasureCount := 1 + difficulty/5 + rng.Intn(2)
	for i := 0; i < treasureCount; i++ {
		x := 1 + rng.Intn(bounds.Width-2)
		y := 1 + rng.Intn(bounds.Height-2)

		treasure := pcg.RoomFeature{
			Type:     "treasure_chest",
			Position: game.Position{X: bounds.X + x, Y: bounds.Y + y},
			Properties: map[string]interface{}{
				"rarity":   trg.getTreasureRarity(difficulty, rng),
				"locked":   difficulty > 5,
				"trapped":  difficulty > 8 && rng.Float64() < 0.3,
				"contents": trg.generateTreasureContents(difficulty, rng),
			},
		}
		room.Features = append(room.Features, treasure)
	}

	// Add guardian if valuable enough
	if difficulty > 7 {
		room.Features = append(room.Features, pcg.RoomFeature{
			Type:     "guardian",
			Position: game.Position{X: bounds.X + bounds.Width/2, Y: bounds.Y + bounds.Height/2},
			Properties: map[string]interface{}{
				"type":       "treasure_guardian",
				"difficulty": difficulty - 2,
			},
		})
	}

	// Add single door (secure access)
	room.Doors = []game.Position{
		{X: bounds.X + bounds.Width/2, Y: bounds.Y},
	}

	room.Properties["treasure_value"] = difficulty * 100
	room.Properties["requires_key"] = difficulty > 10

	return room, nil
}

func (trg *TreasureRoomGenerator) getTreasureRarity(difficulty int, rng *rand.Rand) string {
	switch {
	case difficulty < 5:
		return "common"
	case difficulty < 10:
		return "uncommon"
	case difficulty < 15:
		return "rare"
	default:
		return "epic"
	}
}

func (trg *TreasureRoomGenerator) generateTreasureContents(difficulty int, rng *rand.Rand) []string {
	contents := []string{"gold"}

	if difficulty > 3 {
		contents = append(contents, "gems")
	}
	if difficulty > 7 {
		contents = append(contents, "magic_item")
	}
	if difficulty > 12 {
		contents = append(contents, "artifact")
	}

	return contents
}

// PuzzleRoomGenerator creates puzzle and challenge rooms with interactive elements.
// Generated rooms contain themed puzzle mechanics such as lever sequences, pressure
// plates, rune puzzles, or mechanical challenges depending on the level theme.
type PuzzleRoomGenerator struct{}

// GenerateRoom creates a puzzle room with interactive elements that must be solved
// to progress. Puzzle type is selected based on the level theme (classic, mechanical,
// magical, etc.). The room includes entrance and exit doors, with the exit potentially
// locked until the puzzle is solved. Element count and complexity scale with difficulty.
func (prg *PuzzleRoomGenerator) GenerateRoom(bounds pcg.Rectangle, theme pcg.LevelTheme, difficulty int, genCtx *pcg.GenerationContext) (*pcg.RoomLayout, error) {
	room := &pcg.RoomLayout{
		Type:       pcg.RoomTypePuzzle,
		Bounds:     bounds,
		Tiles:      make([][]game.Tile, bounds.Height),
		Doors:      []game.Position{},
		Features:   []pcg.RoomFeature{},
		Properties: make(map[string]interface{}),
	}

	// Initialize room tiles
	for y := 0; y < bounds.Height; y++ {
		room.Tiles[y] = make([]game.Tile, bounds.Width)
		for x := 0; x < bounds.Width; x++ {
			if x == 0 || x == bounds.Width-1 || y == 0 || y == bounds.Height-1 {
				room.Tiles[y][x] = game.Tile{
					Type:       game.TileWall,
					Walkable:   false,
					Properties: make(map[string]interface{}),
				}
			} else {
				room.Tiles[y][x] = game.Tile{
					Type:       game.TileFloor,
					Walkable:   true,
					Properties: make(map[string]interface{}),
				}
			}
		}
	}

	rng := genCtx.RNG

	// Generate puzzle type based on theme
	puzzleType := prg.selectPuzzleType(theme, difficulty, rng)

	// Add puzzle elements
	room.Features = append(room.Features, prg.generatePuzzleElements(bounds, puzzleType, difficulty, rng)...)

	// Add entrance door
	room.Doors = []game.Position{
		{X: bounds.X + bounds.Width/2, Y: bounds.Y},
	}

	// Add exit door (may be locked until puzzle solved)
	exitDoor := game.Position{X: bounds.X + bounds.Width/2, Y: bounds.Y + bounds.Height - 1}
	room.Doors = append(room.Doors, exitDoor)

	room.Properties["puzzle_type"] = puzzleType
	room.Properties["difficulty"] = difficulty
	room.Properties["requires_solution"] = true

	return room, nil
}

func (prg *PuzzleRoomGenerator) selectPuzzleType(theme pcg.LevelTheme, difficulty int, rng *rand.Rand) string {
	var puzzles []string

	switch theme {
	case pcg.ThemeClassic:
		puzzles = []string{"lever_sequence", "pressure_plates", "riddle"}
	case pcg.ThemeMechanical:
		puzzles = []string{"gear_puzzle", "circuit_puzzle", "weight_balance"}
	case pcg.ThemeMagical:
		puzzles = []string{"rune_sequence", "elemental_matching", "spell_focus"}
	default:
		puzzles = []string{"lever_sequence", "pressure_plates", "riddle"}
	}

	return puzzles[rng.Intn(len(puzzles))]
}

func (prg *PuzzleRoomGenerator) generatePuzzleElements(bounds pcg.Rectangle, puzzleType string, difficulty int, rng *rand.Rand) []pcg.RoomFeature {
	var features []pcg.RoomFeature

	elementCount := 2 + difficulty/3

	switch puzzleType {
	case "lever_sequence":
		for i := 0; i < elementCount; i++ {
			x := 1 + rng.Intn(bounds.Width-2)
			y := 1 + rng.Intn(bounds.Height-2)
			features = append(features, pcg.RoomFeature{
				Type:     "lever",
				Position: game.Position{X: bounds.X + x, Y: bounds.Y + y},
				Properties: map[string]interface{}{
					"sequence_number": i + 1,
					"state":           "off",
				},
			})
		}
	case "pressure_plates":
		for i := 0; i < elementCount; i++ {
			x := 1 + rng.Intn(bounds.Width-2)
			y := 1 + rng.Intn(bounds.Height-2)
			features = append(features, pcg.RoomFeature{
				Type:     "pressure_plate",
				Position: game.Position{X: bounds.X + x, Y: bounds.Y + y},
				Properties: map[string]interface{}{
					"activated": false,
					"weight":    10 + rng.Intn(50),
				},
			})
		}
	default:
		// Generic interactive element
		x := bounds.Width / 2
		y := bounds.Height / 2
		features = append(features, pcg.RoomFeature{
			Type:     "puzzle_element",
			Position: game.Position{X: bounds.X + x, Y: bounds.Y + y},
			Properties: map[string]interface{}{
				"type": puzzleType,
			},
		})
	}

	return features
}

// BossRoomGenerator creates climactic boss encounter rooms with arena-style layouts.
// Generated rooms are larger than standard rooms with reinforced walls, boss spawn
// points, and environmental hazards that activate during different fight phases.
type BossRoomGenerator struct{}

// GenerateRoom creates a boss encounter arena with phase-based environmental hazards.
// The room features a central boss spawn point, reinforced walls, and escape routes
// for tactical retreats. Environmental hazards trigger based on boss health thresholds
// (75%, 50%, 25%) for multi-phase encounters. Boss type is selected based on theme.
func (brg *BossRoomGenerator) GenerateRoom(bounds pcg.Rectangle, theme pcg.LevelTheme, difficulty int, genCtx *pcg.GenerationContext) (*pcg.RoomLayout, error) {
	room := &pcg.RoomLayout{
		Type:       pcg.RoomTypeBoss,
		Bounds:     bounds,
		Tiles:      make([][]game.Tile, bounds.Height),
		Doors:      []game.Position{},
		Features:   []pcg.RoomFeature{},
		Properties: make(map[string]interface{}),
	}

	// Initialize larger room tiles
	for y := 0; y < bounds.Height; y++ {
		room.Tiles[y] = make([]game.Tile, bounds.Width)
		for x := 0; x < bounds.Width; x++ {
			if x == 0 || x == bounds.Width-1 || y == 0 || y == bounds.Height-1 {
				room.Tiles[y][x] = game.Tile{
					Type:       game.TileWall,
					Walkable:   false,
					Properties: map[string]interface{}{"reinforced": true},
				}
			} else {
				room.Tiles[y][x] = game.Tile{
					Type:       game.TileFloor,
					Walkable:   true,
					Properties: map[string]interface{}{"arena": true},
				}
			}
		}
	}

	rng := genCtx.RNG

	// Add boss spawn point (center)
	boss := pcg.RoomFeature{
		Type:     "boss_spawn",
		Position: game.Position{X: bounds.X + bounds.Width/2, Y: bounds.Y + bounds.Height/2},
		Properties: map[string]interface{}{
			"boss_type":  brg.selectBossType(theme, difficulty),
			"difficulty": difficulty + 2,
			"phases":     1 + difficulty/8,
		},
	}
	room.Features = append(room.Features, boss)

	// Add environmental features for multi-phase encounters
	phaseCount := 1 + difficulty/8
	for i := 0; i < phaseCount; i++ {
		x := 2 + rng.Intn(bounds.Width-4)
		y := 2 + rng.Intn(bounds.Height-4)

		feature := pcg.RoomFeature{
			Type:     "environmental_hazard",
			Position: game.Position{X: bounds.X + x, Y: bounds.Y + y},
			Properties: map[string]interface{}{
				"phase":   i + 1,
				"type":    brg.selectHazardType(theme),
				"trigger": "boss_health_" + fmt.Sprintf("%d", 75-(i*25)),
			},
		}
		room.Features = append(room.Features, feature)
	}

	// Add single entrance (dramatic entry)
	room.Doors = []game.Position{
		{X: bounds.X + bounds.Width/2, Y: bounds.Y},
	}

	room.Properties["boss_encounter"] = true
	room.Properties["arena_size"] = "large"
	room.Properties["escape_routes"] = brg.generateEscapeRoutes(bounds)

	return room, nil
}

func (brg *BossRoomGenerator) selectBossType(theme pcg.LevelTheme, difficulty int) string {
	switch theme {
	case pcg.ThemeClassic:
		return "dragon"
	case pcg.ThemeHorror:
		return "abomination"
	case pcg.ThemeUndead:
		return "lich"
	case pcg.ThemeMechanical:
		return "war_machine"
	case pcg.ThemeMagical:
		return "archmage"
	case pcg.ThemeElemental:
		return "elemental_lord"
	default:
		return "champion"
	}
}

func (brg *BossRoomGenerator) selectHazardType(theme pcg.LevelTheme) string {
	switch theme {
	case pcg.ThemeClassic:
		return "falling_rocks"
	case pcg.ThemeHorror:
		return "blood_pools"
	case pcg.ThemeMagical:
		return "magic_storm"
	case pcg.ThemeElemental:
		return "elemental_eruption"
	default:
		return "debris"
	}
}

func (brg *BossRoomGenerator) generateEscapeRoutes(bounds pcg.Rectangle) []game.Position {
	// Add potential escape routes for tactical retreats
	return []game.Position{
		{X: bounds.X + 1, Y: bounds.Y + bounds.Height/2},
		{X: bounds.X + bounds.Width - 2, Y: bounds.Y + bounds.Height/2},
	}
}

// Define other room generators with basic implementations

// EntranceRoomGenerator creates level entrance rooms that serve as safe starting points.
// Generated rooms are designated as safe zones with healing capabilities.
type EntranceRoomGenerator struct{}

// GenerateRoom creates a safe entrance room where players begin the level.
// The room provides a healing safe zone before players venture into dangerous areas.
func (erg *EntranceRoomGenerator) GenerateRoom(bounds pcg.Rectangle, theme pcg.LevelTheme, difficulty int, genCtx *pcg.GenerationContext) (*pcg.RoomLayout, error) {
	return generateBasicRoom(bounds, "entrance", map[string]interface{}{
		"safe_zone": true,
		"healing":   true,
	})
}

// ExitRoomGenerator creates level exit rooms with portals to the next area.
// Generated rooms are designated as safe zones with exit portal mechanics.
type ExitRoomGenerator struct{}

// GenerateRoom creates a safe exit room with a portal to progress to the next level.
// The room provides a brief respite before players move to the next challenge.
func (erg *ExitRoomGenerator) GenerateRoom(bounds pcg.Rectangle, theme pcg.LevelTheme, difficulty int, genCtx *pcg.GenerationContext) (*pcg.RoomLayout, error) {
	return generateBasicRoom(bounds, "exit", map[string]interface{}{
		"exit_portal": true,
		"safe_zone":   true,
	})
}

// SecretRoomGenerator creates hidden secret rooms with special rewards.
// Generated rooms are marked as hidden and contain special loot plus discovery XP
// that scales with difficulty level.
type SecretRoomGenerator struct{}

// GenerateRoom creates a hidden secret room with special loot rewards.
// Players gain discovery XP (difficulty * 10) upon finding these rooms.
func (srg *SecretRoomGenerator) GenerateRoom(bounds pcg.Rectangle, theme pcg.LevelTheme, difficulty int, genCtx *pcg.GenerationContext) (*pcg.RoomLayout, error) {
	return generateBasicRoom(bounds, "secret", map[string]interface{}{
		"hidden":       true,
		"special_loot": true,
		"discovery_xp": difficulty * 10,
	})
}

// ShopRoomGenerator creates merchant shop rooms where players can buy and sell items.
// Generated rooms are safe zones with merchant NPCs and configured buy/sell price ratios.
type ShopRoomGenerator struct{}

// GenerateRoom creates a safe shop room with a merchant NPC.
// Default buy prices are at 100% and sell prices at 50% of item value.
func (srg *ShopRoomGenerator) GenerateRoom(bounds pcg.Rectangle, theme pcg.LevelTheme, difficulty int, genCtx *pcg.GenerationContext) (*pcg.RoomLayout, error) {
	return generateBasicRoom(bounds, "shop", map[string]interface{}{
		"merchant":    true,
		"safe_zone":   true,
		"buy_prices":  1.0,
		"sell_prices": 0.5,
	})
}

// RestRoomGenerator creates rest area rooms for party recovery.
// Generated rooms are safe zones with healing and spell recharge capabilities.
type RestRoomGenerator struct{}

// GenerateRoom creates a safe rest room where players can heal and recharge spells.
// These rooms provide respite between dangerous encounters.
func (rrg *RestRoomGenerator) GenerateRoom(bounds pcg.Rectangle, theme pcg.LevelTheme, difficulty int, genCtx *pcg.GenerationContext) (*pcg.RoomLayout, error) {
	return generateBasicRoom(bounds, "rest", map[string]interface{}{
		"safe_zone":      true,
		"healing":        true,
		"spell_recharge": true,
	})
}

// TrapRoomGenerator creates dangerous trap-filled rooms requiring careful navigation.
// Generated rooms contain hidden traps with density scaling by difficulty level.
type TrapRoomGenerator struct{}

// GenerateRoom creates a dangerous room filled with hidden traps.
// Trap density scales with difficulty, making higher-level rooms more hazardous.
func (trg *TrapRoomGenerator) GenerateRoom(bounds pcg.Rectangle, theme pcg.LevelTheme, difficulty int, genCtx *pcg.GenerationContext) (*pcg.RoomLayout, error) {
	return generateBasicRoom(bounds, "trap", map[string]interface{}{
		"trap_density": difficulty,
		"hidden_traps": true,
		"danger_level": "high",
	})
}

// StoryRoomGenerator creates narrative-focused rooms with lore and story elements.
// Generated rooms are safe zones with narrative content and lore points that scale
// with difficulty level for progression-based storytelling.
type StoryRoomGenerator struct{}

// GenerateRoom creates a safe story room with narrative elements and lore content.
// Lore points scale with difficulty, rewarding exploration at higher levels.
func (srg *StoryRoomGenerator) GenerateRoom(bounds pcg.Rectangle, theme pcg.LevelTheme, difficulty int, genCtx *pcg.GenerationContext) (*pcg.RoomLayout, error) {
	return generateBasicRoom(bounds, "story", map[string]interface{}{
		"narrative":   true,
		"safe_zone":   true,
		"lore_points": difficulty,
	})
}

// generateBasicRoom creates a simple room with standard layout
func generateBasicRoom(bounds pcg.Rectangle, roomType string, properties map[string]interface{}) (*pcg.RoomLayout, error) {
	var roomTypeEnum pcg.RoomType
	switch roomType {
	case "entrance":
		roomTypeEnum = pcg.RoomTypeEntrance
	case "exit":
		roomTypeEnum = pcg.RoomTypeExit
	case "secret":
		roomTypeEnum = pcg.RoomTypeSecret
	case "shop":
		roomTypeEnum = pcg.RoomTypeShop
	case "rest":
		roomTypeEnum = pcg.RoomTypeRest
	case "trap":
		roomTypeEnum = pcg.RoomTypeTrap
	case "story":
		roomTypeEnum = pcg.RoomTypeStory
	default:
		roomTypeEnum = pcg.RoomTypeCombat
	}

	room := &pcg.RoomLayout{
		Type:       roomTypeEnum,
		Bounds:     bounds,
		Tiles:      make([][]game.Tile, bounds.Height),
		Doors:      []game.Position{},
		Features:   []pcg.RoomFeature{},
		Properties: properties,
	}

	// Initialize basic room tiles
	for y := 0; y < bounds.Height; y++ {
		room.Tiles[y] = make([]game.Tile, bounds.Width)
		for x := 0; x < bounds.Width; x++ {
			if x == 0 || x == bounds.Width-1 || y == 0 || y == bounds.Height-1 {
				room.Tiles[y][x] = game.Tile{
					Type:       game.TileWall,
					Walkable:   false,
					Properties: make(map[string]interface{}),
				}
			} else {
				room.Tiles[y][x] = game.Tile{
					Type:       game.TileFloor,
					Walkable:   true,
					Properties: make(map[string]interface{}),
				}
			}
		}
	}

	// Add basic door
	room.Doors = []game.Position{
		{X: bounds.X + bounds.Width/2, Y: bounds.Y},
	}

	// Add room type specific feature
	if roomType != "entrance" && roomType != "exit" {
		room.Features = append(room.Features, pcg.RoomFeature{
			Type:       roomType + "_feature",
			Position:   game.Position{X: bounds.X + bounds.Width/2, Y: bounds.Y + bounds.Height/2},
			Properties: properties,
		})
	}

	return room, nil
}
