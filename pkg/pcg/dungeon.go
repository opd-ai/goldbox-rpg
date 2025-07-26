package pcg

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
)

// DungeonComplex represents a multi-level dungeon structure
// Manages interconnected levels with proper progression and connectivity
type DungeonComplex struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Levels      map[int]*DungeonLevel  `json:"levels"`
	Connections []LevelConnection      `json:"connections"`
	Theme       LevelTheme             `json:"theme"`
	Difficulty  DifficultyProgression  `json:"difficulty"`
	Metadata    map[string]interface{} `json:"metadata"`
	Generated   time.Time              `json:"generated"`
}

// DungeonLevel represents a single level within a dungeon complex
type DungeonLevel struct {
	Level       int                    `json:"level"`
	Map         *game.GameMap          `json:"map"`
	Rooms       []*RoomLayout          `json:"rooms"`
	Connections []ConnectionPoint      `json:"connections"`
	Theme       LevelTheme             `json:"theme"`
	Difficulty  int                    `json:"difficulty"`
	Properties  map[string]interface{} `json:"properties"`
}

// LevelConnection represents connections between dungeon levels
type LevelConnection struct {
	FromLevel    int                    `json:"from_level"`
	ToLevel      int                    `json:"to_level"`
	FromPosition game.Position          `json:"from_position"`
	ToPosition   game.Position          `json:"to_position"`
	Type         ConnectionType         `json:"type"`
	Properties   map[string]interface{} `json:"properties"`
}

// ConnectionPoint represents a connection point within a level
type ConnectionPoint struct {
	Position    game.Position          `json:"position"`
	Type        ConnectionType         `json:"type"`
	TargetLevel int                    `json:"target_level"`
	Properties  map[string]interface{} `json:"properties"`
}

// ConnectionType represents different types of level connections
type ConnectionType string

const (
	ConnectionStairs   ConnectionType = "stairs"
	ConnectionElevator ConnectionType = "elevator"
	ConnectionPortal   ConnectionType = "portal"
	ConnectionPit      ConnectionType = "pit"
	ConnectionLadder   ConnectionType = "ladder"
	ConnectionTunnel   ConnectionType = "tunnel"
)

// DifficultyProgression defines how difficulty scales across levels
type DifficultyProgression struct {
	BaseDifficulty  int     `json:"base_difficulty"`
	ScalingFactor   float64 `json:"scaling_factor"`
	MaxDifficulty   int     `json:"max_difficulty"`
	ProgressionType string  `json:"progression_type"`
}

// DungeonGenerator creates multi-level dungeon complexes
// Uses existing level generation as foundation for complex structures
type DungeonGenerator struct {
	version string
	logger  *logrus.Logger
	rng     *rand.Rand
}

// NewDungeonGenerator creates a new dungeon complex generator
func NewDungeonGenerator(logger *logrus.Logger) *DungeonGenerator {
	if logger == nil {
		logger = logrus.New()
	}

	return &DungeonGenerator{
		version: "1.0.0",
		logger:  logger,
		rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Generate creates a complete multi-level dungeon complex
// Implements Generator interface for PCG system integration
func (dg *DungeonGenerator) Generate(ctx context.Context, params GenerationParams) (interface{}, error) {
	dungeonParams, ok := params.Constraints["dungeon_params"].(DungeonParams)
	if !ok {
		return nil, fmt.Errorf("invalid parameters for dungeon generation: expected dungeon_params in constraints")
	}

	// Validate parameters before generation
	if err := dg.Validate(params); err != nil {
		return nil, fmt.Errorf("parameter validation failed: %w", err)
	}

	// Initialize RNG with provided seed for deterministic generation
	dg.rng = rand.New(rand.NewSource(params.Seed))

	dg.logger.WithFields(logrus.Fields{
		"levels":     dungeonParams.LevelCount,
		"theme":      dungeonParams.Theme,
		"difficulty": dungeonParams.Difficulty.BaseDifficulty,
	}).Info("generating multi-level dungeon complex")

	dungeon, err := dg.generateDungeonComplex(ctx, params, dungeonParams)
	if err != nil {
		return nil, fmt.Errorf("dungeon generation failed: %w", err)
	}

	dg.logger.WithField("dungeon_id", dungeon.ID).Info("dungeon complex generation completed")
	return dungeon, nil
}

// generateDungeonComplex creates the complete dungeon structure
func (dg *DungeonGenerator) generateDungeonComplex(ctx context.Context, params GenerationParams, dungeonParams DungeonParams) (*DungeonComplex, error) {
	dungeon := &DungeonComplex{
		ID:          fmt.Sprintf("dungeon_%d", dg.rng.Int63()),
		Name:        dg.generateDungeonName(dungeonParams.Theme),
		Levels:      make(map[int]*DungeonLevel),
		Connections: make([]LevelConnection, 0),
		Theme:       dungeonParams.Theme,
		Difficulty:  dungeonParams.Difficulty,
		Metadata:    make(map[string]interface{}),
		Generated:   time.Now(),
	}

	// Generate individual levels
	for level := 1; level <= dungeonParams.LevelCount; level++ {
		levelDifficulty := dg.calculateLevelDifficulty(level, dungeonParams.Difficulty)

		dungeonLevel, err := dg.generateDungeonLevel(ctx, level, levelDifficulty, params, dungeonParams)
		if err != nil {
			return nil, fmt.Errorf("failed to generate level %d: %w", level, err)
		}

		dungeon.Levels[level] = dungeonLevel
	}

	// Create connections between levels
	if err := dg.createLevelConnections(dungeon, dungeonParams); err != nil {
		return nil, fmt.Errorf("failed to create level connections: %w", err)
	}

	// Add metadata for debugging and validation
	dungeon.Metadata["total_rooms"] = dg.countTotalRooms(dungeon)
	dungeon.Metadata["connection_count"] = len(dungeon.Connections)
	dungeon.Metadata["generation_seed"] = params.Seed

	return dungeon, nil
}

// generateDungeonLevel creates a single level with basic room layout
func (dg *DungeonGenerator) generateDungeonLevel(ctx context.Context, levelNum, difficulty int, params GenerationParams, dungeonParams DungeonParams) (*DungeonLevel, error) {
	// Create a basic game map for this level
	gameMap := &game.GameMap{
		Width:  dungeonParams.LevelWidth,
		Height: dungeonParams.LevelHeight,
		Tiles:  make([][]game.MapTile, dungeonParams.LevelHeight),
	}

	// Initialize map tiles
	for y := 0; y < dungeonParams.LevelHeight; y++ {
		gameMap.Tiles[y] = make([]game.MapTile, dungeonParams.LevelWidth)
		for x := 0; x < dungeonParams.LevelWidth; x++ {
			gameMap.Tiles[y][x] = game.MapTile{
				SpriteX:     1, // Wall sprite
				SpriteY:     1,
				Walkable:    false,
				Transparent: false,
			}
		}
	}

	// Generate rooms using a simplified approach to avoid import cycle
	rooms := dg.generateRoomsForLevel(gameMap, dungeonParams, difficulty)

	// Connect rooms with basic corridors
	dg.connectRoomsWithCorridors(gameMap, rooms)

	// Convert to DungeonLevel structure
	dungeonLevel := &DungeonLevel{
		Level:       levelNum,
		Map:         gameMap,
		Rooms:       rooms,
		Connections: make([]ConnectionPoint, 0),
		Theme:       dungeonParams.Theme,
		Difficulty:  difficulty,
		Properties:  make(map[string]interface{}),
	}

	// Add level-specific properties
	dungeonLevel.Properties["room_count"] = len(rooms)
	dungeonLevel.Properties["generated_at"] = time.Now()
	dungeonLevel.Properties["level_seed"] = params.Seed + int64(levelNum)

	return dungeonLevel, nil
}

// generateRoomsForLevel creates rooms for a dungeon level
func (dg *DungeonGenerator) generateRoomsForLevel(gameMap *game.GameMap, dungeonParams DungeonParams, difficulty int) []*RoomLayout {
	rooms := make([]*RoomLayout, 0, dungeonParams.RoomsPerLevel)

	// Calculate room sizes based on map dimensions
	minRoomSize := 5
	maxRoomSize := 12

	// Place rooms with basic non-overlapping algorithm
	attempts := 0
	maxAttempts := dungeonParams.RoomsPerLevel * 10

	for len(rooms) < dungeonParams.RoomsPerLevel && attempts < maxAttempts {
		attempts++

		// Random room dimensions
		roomWidth := minRoomSize + dg.rng.Intn(maxRoomSize-minRoomSize+1)
		roomHeight := minRoomSize + dg.rng.Intn(maxRoomSize-minRoomSize+1)

		// Random position with padding
		x := 2 + dg.rng.Intn(gameMap.Width-roomWidth-4)
		y := 2 + dg.rng.Intn(gameMap.Height-roomHeight-4)

		newRoom := Rectangle{
			X:      x,
			Y:      y,
			Width:  roomWidth,
			Height: roomHeight,
		}

		// Check for overlaps with existing rooms
		overlaps := false
		for _, existingRoom := range rooms {
			if newRoom.Intersects(existingRoom.Bounds) {
				overlaps = true
				break
			}
		}

		if !overlaps {
			room := dg.createRoom(newRoom, len(rooms), dungeonParams.Theme, difficulty)
			rooms = append(rooms, room)
			dg.carveRoom(gameMap, room)
		}
	}

	return rooms
}

// createRoom creates a room layout with the specified bounds
func (dg *DungeonGenerator) createRoom(bounds Rectangle, index int, theme LevelTheme, difficulty int) *RoomLayout {
	// Determine room type based on index and theme
	roomType := dg.determineRoomType(index, theme)

	room := &RoomLayout{
		ID:         fmt.Sprintf("room_%d", index),
		Type:       roomType,
		Bounds:     bounds,
		Tiles:      make([][]game.Tile, bounds.Height),
		Doors:      make([]game.Position, 0),
		Features:   make([]RoomFeature, 0),
		Difficulty: difficulty,
		Properties: make(map[string]interface{}),
		Connected:  make([]string, 0),
	}

	// Initialize room tiles as floor
	for y := 0; y < bounds.Height; y++ {
		room.Tiles[y] = make([]game.Tile, bounds.Width)
		for x := 0; x < bounds.Width; x++ {
			room.Tiles[y][x] = game.Tile{
				Type:        game.TileFloor,
				Walkable:    true,
				Transparent: true,
				Properties:  make(map[string]interface{}),
				Sprite:      "floor",
				Color:       game.RGB{R: 128, G: 128, B: 128},
			}
		}
	}

	return room
}

// carveRoom carves the room into the game map
func (dg *DungeonGenerator) carveRoom(gameMap *game.GameMap, room *RoomLayout) {
	for y := room.Bounds.Y; y < room.Bounds.Y+room.Bounds.Height; y++ {
		for x := room.Bounds.X; x < room.Bounds.X+room.Bounds.Width; x++ {
			if y >= 0 && y < gameMap.Height && x >= 0 && x < gameMap.Width {
				gameMap.Tiles[y][x] = game.MapTile{
					SpriteX:     0, // Floor sprite
					SpriteY:     0,
					Walkable:    true,
					Transparent: true,
				}
			}
		}
	}
}

// connectRoomsWithCorridors creates simple L-shaped corridors between rooms
func (dg *DungeonGenerator) connectRoomsWithCorridors(gameMap *game.GameMap, rooms []*RoomLayout) {
	for i := 0; i < len(rooms)-1; i++ {
		room1 := rooms[i]
		room2 := rooms[i+1]

		// Get center points of rooms
		x1 := room1.Bounds.X + room1.Bounds.Width/2
		y1 := room1.Bounds.Y + room1.Bounds.Height/2
		x2 := room2.Bounds.X + room2.Bounds.Width/2
		y2 := room2.Bounds.Y + room2.Bounds.Height/2

		// Create L-shaped corridor
		dg.carveCorridor(gameMap, x1, y1, x2, y1) // Horizontal
		dg.carveCorridor(gameMap, x2, y1, x2, y2) // Vertical

		// Update connections
		room1.Connected = append(room1.Connected, room2.ID)
		room2.Connected = append(room2.Connected, room1.ID)
	}
}

// carveCorridor carves a corridor between two points
func (dg *DungeonGenerator) carveCorridor(gameMap *game.GameMap, x1, y1, x2, y2 int) {
	// Ensure we're carving in the right direction
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}

	for x := x1; x <= x2; x++ {
		if x >= 0 && x < gameMap.Width && y1 >= 0 && y1 < gameMap.Height {
			gameMap.Tiles[y1][x] = game.MapTile{
				SpriteX:     0, // Floor sprite
				SpriteY:     0,
				Walkable:    true,
				Transparent: true,
			}
		}
	}

	for y := y1; y <= y2; y++ {
		if x2 >= 0 && x2 < gameMap.Width && y >= 0 && y < gameMap.Height {
			gameMap.Tiles[y][x2] = game.MapTile{
				SpriteX:     0, // Floor sprite
				SpriteY:     0,
				Walkable:    true,
				Transparent: true,
			}
		}
	}
}

// determineRoomType determines the type of room based on index and theme
func (dg *DungeonGenerator) determineRoomType(index int, theme LevelTheme) RoomType {
	if index == 0 {
		return RoomTypeEntrance
	}

	// Use weighted random selection for other room types
	weights := map[RoomType]int{
		RoomTypeCombat:   40,
		RoomTypeTreasure: 15,
		RoomTypePuzzle:   10,
		RoomTypeShop:     5,
		RoomTypeRest:     5,
		RoomTypeTrap:     10,
		RoomTypeStory:    10,
		RoomTypeSecret:   5,
	}

	// Adjust weights based on theme
	switch theme {
	case ThemeHorror:
		weights[RoomTypeTrap] = 20
		weights[RoomTypeCombat] = 50
	case ThemeMagical:
		weights[RoomTypePuzzle] = 20
		weights[RoomTypeShop] = 10
	}

	return dg.weightedRandomRoomType(weights)
}

// weightedRandomRoomType selects a room type using weighted random selection
func (dg *DungeonGenerator) weightedRandomRoomType(weights map[RoomType]int) RoomType {
	totalWeight := 0
	for _, weight := range weights {
		totalWeight += weight
	}

	randomValue := dg.rng.Intn(totalWeight)
	currentWeight := 0

	for roomType, weight := range weights {
		currentWeight += weight
		if randomValue < currentWeight {
			return roomType
		}
	}

	return RoomTypeCombat // fallback
}

// createLevelConnections establishes connections between dungeon levels
func (dg *DungeonGenerator) createLevelConnections(dungeon *DungeonComplex, params DungeonParams) error {
	levels := make([]int, 0, len(dungeon.Levels))
	for levelNum := range dungeon.Levels {
		levels = append(levels, levelNum)
	}
	sort.Ints(levels)

	// Create connections between adjacent levels
	for i := 0; i < len(levels)-1; i++ {
		fromLevel := levels[i]
		toLevel := levels[i+1]

		connection, err := dg.createLevelConnection(dungeon.Levels[fromLevel], dungeon.Levels[toLevel], params)
		if err != nil {
			return fmt.Errorf("failed to create connection from level %d to %d: %w", fromLevel, toLevel, err)
		}

		dungeon.Connections = append(dungeon.Connections, *connection)
	}

	// Add occasional skip connections for complexity (e.g., level 1 to level 3)
	if len(levels) > 3 && dg.rng.Float64() < 0.3 {
		skipConnection, err := dg.createSkipConnection(dungeon, levels, params)
		if err == nil {
			dungeon.Connections = append(dungeon.Connections, *skipConnection)
		}
	}

	return nil
}

// createLevelConnection creates a connection between two adjacent levels
func (dg *DungeonGenerator) createLevelConnection(fromLevel, toLevel *DungeonLevel, params DungeonParams) (*LevelConnection, error) {
	// Find suitable positions for connections in both levels
	fromPos, err := dg.findConnectionPosition(fromLevel, ConnectionStairs)
	if err != nil {
		return nil, fmt.Errorf("failed to find connection position in level %d: %w", fromLevel.Level, err)
	}

	toPos, err := dg.findConnectionPosition(toLevel, ConnectionStairs)
	if err != nil {
		return nil, fmt.Errorf("failed to find connection position in level %d: %w", toLevel.Level, err)
	}

	// Choose connection type based on theme and level difference
	connectionType := dg.chooseConnectionType(fromLevel, toLevel, params.Theme)

	connection := &LevelConnection{
		FromLevel:    fromLevel.Level,
		ToLevel:      toLevel.Level,
		FromPosition: fromPos,
		ToPosition:   toPos,
		Type:         connectionType,
		Properties:   make(map[string]interface{}),
	}

	// Add connection points to the levels
	fromLevel.Connections = append(fromLevel.Connections, ConnectionPoint{
		Position:    fromPos,
		Type:        connectionType,
		TargetLevel: toLevel.Level,
		Properties:  make(map[string]interface{}),
	})

	toLevel.Connections = append(toLevel.Connections, ConnectionPoint{
		Position:    toPos,
		Type:        connectionType,
		TargetLevel: fromLevel.Level,
		Properties:  make(map[string]interface{}),
	})

	return connection, nil
}

// findConnectionPosition finds a suitable position for a level connection
func (dg *DungeonGenerator) findConnectionPosition(level *DungeonLevel, connType ConnectionType) (game.Position, error) {
	// Look for rooms that can accommodate connections
	suitableRooms := make([]*RoomLayout, 0)

	for _, room := range level.Rooms {
		// Avoid boss rooms and secret rooms for main connections
		if room.Type != RoomTypeBoss && room.Type != RoomTypeSecret {
			suitableRooms = append(suitableRooms, room)
		}
	}

	if len(suitableRooms) == 0 {
		return game.Position{}, fmt.Errorf("no suitable rooms found for connection")
	}

	// Choose a random suitable room
	room := suitableRooms[dg.rng.Intn(len(suitableRooms))]

	// Find a floor tile within the room for the connection
	attempts := 0
	maxAttempts := 50

	for attempts < maxAttempts {
		// Use relative coordinates within the room bounds
		relativeX := 1 + dg.rng.Intn(room.Bounds.Width-2)
		relativeY := 1 + dg.rng.Intn(room.Bounds.Height-2)

		// Convert to absolute map coordinates
		absoluteX := room.Bounds.X + relativeX
		absoluteY := room.Bounds.Y + relativeY

		// Check if the relative coordinates are valid for the room tiles array
		if relativeY < len(room.Tiles) && relativeX < len(room.Tiles[relativeY]) {
			if room.Tiles[relativeY][relativeX].Type == game.TileFloor && room.Tiles[relativeY][relativeX].Walkable {
				return game.Position{X: absoluteX, Y: absoluteY}, nil
			}
		}
		attempts++
	}

	return game.Position{}, fmt.Errorf("failed to find valid connection position after %d attempts", maxAttempts)
}

// Helper methods

// generateDungeonName creates a thematic name for the dungeon
func (dg *DungeonGenerator) generateDungeonName(theme LevelTheme) string {
	prefixes := map[LevelTheme][]string{
		ThemeClassic:    {"Ancient", "Forgotten", "Lost", "Hidden"},
		ThemeHorror:     {"Cursed", "Haunted", "Nightmare", "Shadow"},
		ThemeNatural:    {"Living", "Verdant", "Root", "Grove"},
		ThemeMechanical: {"Clockwork", "Steam", "Gear", "Iron"},
		ThemeMagical:    {"Arcane", "Crystal", "Ethereal", "Mystic"},
		ThemeUndead:     {"Bone", "Death", "Tomb", "Crypt"},
		ThemeElemental:  {"Elemental", "Primal", "Storm", "Flame"},
	}

	suffixes := []string{"Depths", "Chambers", "Caverns", "Halls", "Passages", "Tunnels", "Labyrinth", "Dungeon"}

	prefix := prefixes[theme][dg.rng.Intn(len(prefixes[theme]))]
	suffix := suffixes[dg.rng.Intn(len(suffixes))]

	return fmt.Sprintf("%s %s", prefix, suffix)
}

// calculateLevelDifficulty computes difficulty for a specific level
func (dg *DungeonGenerator) calculateLevelDifficulty(level int, progression DifficultyProgression) int {
	baseDiff := float64(progression.BaseDifficulty)
	scaledDiff := baseDiff + (float64(level-1) * progression.ScalingFactor)

	difficulty := int(scaledDiff)
	if difficulty > progression.MaxDifficulty {
		difficulty = progression.MaxDifficulty
	}

	return difficulty
}

// chooseConnectionType selects appropriate connection type based on context
func (dg *DungeonGenerator) chooseConnectionType(fromLevel, toLevel *DungeonLevel, theme LevelTheme) ConnectionType {
	// Default to stairs for most themes
	weights := map[ConnectionType]int{
		ConnectionStairs: 60,
		ConnectionLadder: 20,
		ConnectionPit:    10,
		ConnectionTunnel: 5,
	}

	// Adjust weights based on theme
	switch theme {
	case ThemeMechanical:
		weights[ConnectionElevator] = 30
		weights[ConnectionStairs] = 40
	case ThemeMagical:
		weights[ConnectionPortal] = 25
		weights[ConnectionStairs] = 45
	case ThemeNatural:
		weights[ConnectionLadder] = 40
		weights[ConnectionTunnel] = 20
	}

	return dg.weightedRandomConnection(weights)
}

// weightedRandomConnection selects a connection type using weighted random selection
func (dg *DungeonGenerator) weightedRandomConnection(weights map[ConnectionType]int) ConnectionType {
	totalWeight := 0
	for _, weight := range weights {
		totalWeight += weight
	}

	randomValue := dg.rng.Intn(totalWeight)
	currentWeight := 0

	for connType, weight := range weights {
		currentWeight += weight
		if randomValue < currentWeight {
			return connType
		}
	}

	return ConnectionStairs // fallback
}

// createSkipConnection creates optional skip connections between non-adjacent levels
func (dg *DungeonGenerator) createSkipConnection(dungeon *DungeonComplex, levels []int, params DungeonParams) (*LevelConnection, error) {
	if len(levels) < 3 {
		return nil, fmt.Errorf("insufficient levels for skip connection")
	}

	// Create connection from level 1 to level 3 (or similar pattern)
	fromLevel := levels[0]
	toLevel := levels[2]

	return dg.createLevelConnection(dungeon.Levels[fromLevel], dungeon.Levels[toLevel], params)
}

// countTotalRooms counts total rooms across all levels
func (dg *DungeonGenerator) countTotalRooms(dungeon *DungeonComplex) int {
	total := 0
	for _, level := range dungeon.Levels {
		total += len(level.Rooms)
	}
	return total
}

// Interface compliance methods

// GetType returns the content type this generator produces
func (dg *DungeonGenerator) GetType() ContentType {
	return ContentTypeDungeon
}

// GetVersion returns the generator version for compatibility
func (dg *DungeonGenerator) GetVersion() string {
	return dg.version
}

// Validate checks if the provided parameters are valid
func (dg *DungeonGenerator) Validate(params GenerationParams) error {
	dungeonParams, ok := params.Constraints["dungeon_params"].(DungeonParams)
	if !ok {
		return fmt.Errorf("invalid parameters: expected dungeon_params in constraints")
	}

	if dungeonParams.LevelCount < 1 || dungeonParams.LevelCount > 20 {
		return fmt.Errorf("level count must be between 1 and 20, got %d", dungeonParams.LevelCount)
	}

	if dungeonParams.LevelWidth < 20 || dungeonParams.LevelWidth > 200 {
		return fmt.Errorf("level width must be between 20 and 200, got %d", dungeonParams.LevelWidth)
	}

	if dungeonParams.LevelHeight < 20 || dungeonParams.LevelHeight > 200 {
		return fmt.Errorf("level height must be between 20 and 200, got %d", dungeonParams.LevelHeight)
	}

	if dungeonParams.RoomsPerLevel < 3 || dungeonParams.RoomsPerLevel > 50 {
		return fmt.Errorf("rooms per level must be between 3 and 50, got %d", dungeonParams.RoomsPerLevel)
	}

	if dungeonParams.Difficulty.ScalingFactor < 0 || dungeonParams.Difficulty.ScalingFactor > 10 {
		return fmt.Errorf("scaling factor must be between 0 and 10, got %f", dungeonParams.Difficulty.ScalingFactor)
	}

	return nil
}
