package levels

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/pcg"
)

// RoomCorridorGenerator creates levels using room-corridor approach
type RoomCorridorGenerator struct {
	version        string
	roomGenerators map[pcg.RoomType]RoomGenerator
	rng            *rand.Rand
}

// RoomGenerator interface for different room types
type RoomGenerator interface {
	GenerateRoom(bounds pcg.Rectangle, theme pcg.LevelTheme, difficulty int, genCtx *pcg.GenerationContext) (*pcg.RoomLayout, error)
}

// NewRoomCorridorGenerator creates a new room-corridor level generator with a time-based seed.
//
// Note: For reproducible level generation (e.g., in tests or replays),
// use NewRoomCorridorGeneratorWithSeed() instead. This function uses time.Now().UnixNano()
// as the seed, making it non-deterministic.
func NewRoomCorridorGenerator() *RoomCorridorGenerator {
	return NewRoomCorridorGeneratorWithSeed(time.Now().UnixNano())
}

// NewRoomCorridorGeneratorWithSeed creates a new room-corridor level generator with a specific seed.
// This enables reproducible level generation for testing, replays, and debugging.
//
// Parameters:
//   - seed: The random seed to use for deterministic generation
//
// Returns:
//   - *RoomCorridorGenerator: A fully configured level generator instance
func NewRoomCorridorGeneratorWithSeed(seed int64) *RoomCorridorGenerator {
	rcg := &RoomCorridorGenerator{
		version:        "1.0.0",
		roomGenerators: make(map[pcg.RoomType]RoomGenerator),
		rng:            rand.New(rand.NewSource(seed)),
	}

	// Register default room generators
	rcg.registerDefaultRoomGenerators()
	return rcg
}

// registerDefaultRoomGenerators registers the default room generators
func (rcg *RoomCorridorGenerator) registerDefaultRoomGenerators() {
	rcg.roomGenerators[pcg.RoomTypeCombat] = &CombatRoomGenerator{}
	rcg.roomGenerators[pcg.RoomTypeTreasure] = &TreasureRoomGenerator{}
	rcg.roomGenerators[pcg.RoomTypePuzzle] = &PuzzleRoomGenerator{}
	rcg.roomGenerators[pcg.RoomTypeBoss] = &BossRoomGenerator{}
	rcg.roomGenerators[pcg.RoomTypeEntrance] = &EntranceRoomGenerator{}
	rcg.roomGenerators[pcg.RoomTypeExit] = &ExitRoomGenerator{}
	rcg.roomGenerators[pcg.RoomTypeSecret] = &SecretRoomGenerator{}
	rcg.roomGenerators[pcg.RoomTypeShop] = &ShopRoomGenerator{}
	rcg.roomGenerators[pcg.RoomTypeRest] = &RestRoomGenerator{}
	rcg.roomGenerators[pcg.RoomTypeTrap] = &TrapRoomGenerator{}
	rcg.roomGenerators[pcg.RoomTypeStory] = &StoryRoomGenerator{}
}

// SetSeed sets the random seed for deterministic generation
func (rcg *RoomCorridorGenerator) SetSeed(seed int64) {
	rcg.rng = rand.New(rand.NewSource(seed))
}

// GetType returns the content type this generator produces
func (rcg *RoomCorridorGenerator) GetType() pcg.ContentType {
	return pcg.ContentTypeLevels
}

// GetVersion returns the generator version for compatibility checking
func (rcg *RoomCorridorGenerator) GetVersion() string {
	return rcg.version
}

// Validate checks if the provided parameters are valid for this generator
func (rcg *RoomCorridorGenerator) Validate(params pcg.GenerationParams) error {
	levelParams, ok := params.Constraints["level_params"].(pcg.LevelParams)
	if !ok {
		return fmt.Errorf("invalid level parameters provided")
	}

	if levelParams.MinRooms < 1 {
		return fmt.Errorf("minimum rooms must be at least 1")
	}

	if levelParams.MaxRooms < levelParams.MinRooms {
		return fmt.Errorf("maximum rooms must be greater than or equal to minimum rooms")
	}

	if params.Difficulty < 1 || params.Difficulty > 20 {
		return fmt.Errorf("difficulty must be between 1 and 20")
	}

	if params.PlayerLevel < 1 || params.PlayerLevel > 20 {
		return fmt.Errorf("player level must be between 1 and 20")
	}

	return nil
}

// Generate implements the Generator interface
func (rcg *RoomCorridorGenerator) Generate(ctx context.Context, params pcg.GenerationParams) (interface{}, error) {
	levelParams, ok := params.Constraints["level_params"].(pcg.LevelParams)
	if !ok {
		return nil, fmt.Errorf("invalid level parameters provided")
	}

	// Set the seed for deterministic generation
	rcg.SetSeed(params.Seed)

	return rcg.GenerateLevel(ctx, levelParams)
}

// GenerateLevel creates a complete dungeon level
func (rcg *RoomCorridorGenerator) GenerateLevel(ctx context.Context, params pcg.LevelParams) (*game.Level, error) {
	// Create generation context
	seedMgr := pcg.NewSeedManager(params.Seed)
	genCtx := pcg.NewGenerationContext(seedMgr, pcg.ContentTypeLevels, "level_generation", params.GenerationParams)

	// Calculate level dimensions based on room count
	width, height := rcg.calculateLevelDimensions(params)

	// 1. Plan room layout using space partitioning
	roomLayouts, err := rcg.generateRoomLayout(width, height, params, genCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate room layout: %w", err)
	}

	// 2. Generate individual rooms
	err = rcg.generateRooms(roomLayouts, params, genCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate rooms: %w", err)
	}

	// 3. Create corridor connections
	corridors, err := rcg.ConnectRooms(ctx, roomLayouts, params)
	if err != nil {
		return nil, fmt.Errorf("failed to connect rooms: %w", err)
	}

	// 4. Add special features and encounters
	err = rcg.addSpecialFeatures(roomLayouts, params, genCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to add special features: %w", err)
	}

	// 5. Validate connectivity and balance
	err = rcg.validateLevel(roomLayouts, corridors)
	if err != nil {
		return nil, fmt.Errorf("level validation failed: %w", err)
	}

	// 6. Convert to game.Level format
	level, err := rcg.convertToGameLevel(roomLayouts, corridors, width, height, params)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to game level: %w", err)
	}

	return level, nil
}

// calculateLevelDimensions calculates appropriate dimensions based on room count
func (rcg *RoomCorridorGenerator) calculateLevelDimensions(params pcg.LevelParams) (width, height int) {
	roomCount := params.MinRooms + rcg.rng.Intn(params.MaxRooms-params.MinRooms+1)

	// Estimate dimensions based on room count and theme
	baseSize := 40 + roomCount*8 // Base size scales with room count

	// Theme-specific adjustments
	switch params.LevelTheme {
	case pcg.ThemeClassic:
		width, height = baseSize, baseSize
	case pcg.ThemeHorror:
		// Longer, narrower levels for tension
		width, height = baseSize+20, baseSize-10
	case pcg.ThemeNatural:
		// More organic, irregular dimensions
		width, height = baseSize+rcg.rng.Intn(20)-10, baseSize+rcg.rng.Intn(20)-10
	case pcg.ThemeMechanical:
		// Perfect squares for mechanical precision
		width, height = baseSize, baseSize
	default:
		width, height = baseSize, baseSize
	}

	// Ensure minimum dimensions
	if width < 30 {
		width = 30
	}
	if height < 30 {
		height = 30
	}

	return width, height
}

// generateRoomLayout creates the spatial layout of rooms using BSP
func (rcg *RoomCorridorGenerator) generateRoomLayout(width, height int, params pcg.LevelParams, genCtx *pcg.GenerationContext) ([]*pcg.RoomLayout, error) {
	roomCount := params.MinRooms + rcg.rng.Intn(params.MaxRooms-params.MinRooms+1)

	// Create BSP tree for room placement
	rootArea := pcg.Rectangle{X: 5, Y: 5, Width: width - 10, Height: height - 10}
	bspAreas := rcg.createBSPAreas(rootArea, roomCount)

	var roomLayouts []*pcg.RoomLayout

	// Create rooms from BSP areas
	for i, area := range bspAreas {
		roomType := rcg.selectRoomType(i, len(bspAreas), params)

		roomLayout := &pcg.RoomLayout{
			ID:         fmt.Sprintf("room_%d", i),
			Type:       roomType,
			Bounds:     area,
			Difficulty: params.Difficulty,
			Properties: make(map[string]interface{}),
			Connected:  []string{},
		}

		roomLayouts = append(roomLayouts, roomLayout)
	}

	// Ensure we have required special rooms
	rcg.ensureSpecialRooms(roomLayouts, params)

	return roomLayouts, nil
}

// createBSPAreas uses Binary Space Partitioning to create room areas
func (rcg *RoomCorridorGenerator) createBSPAreas(area pcg.Rectangle, targetRooms int) []pcg.Rectangle {
	if targetRooms <= 1 {
		return []pcg.Rectangle{area}
	}

	// Split the area
	var areas []pcg.Rectangle
	splitVertical := rcg.rng.Float64() < 0.5

	if splitVertical && area.Width > 20 {
		// Vertical split
		splitPos := area.Width/3 + rcg.rng.Intn(area.Width/3)
		left := pcg.Rectangle{X: area.X, Y: area.Y, Width: splitPos, Height: area.Height}
		right := pcg.Rectangle{X: area.X + splitPos, Y: area.Y, Width: area.Width - splitPos, Height: area.Height}

		leftRooms := targetRooms / 2
		rightRooms := targetRooms - leftRooms

		areas = append(areas, rcg.createBSPAreas(left, leftRooms)...)
		areas = append(areas, rcg.createBSPAreas(right, rightRooms)...)
	} else if !splitVertical && area.Height > 20 {
		// Horizontal split
		splitPos := area.Height/3 + rcg.rng.Intn(area.Height/3)
		top := pcg.Rectangle{X: area.X, Y: area.Y, Width: area.Width, Height: splitPos}
		bottom := pcg.Rectangle{X: area.X, Y: area.Y + splitPos, Width: area.Width, Height: area.Height - splitPos}

		topRooms := targetRooms / 2
		bottomRooms := targetRooms - topRooms

		areas = append(areas, rcg.createBSPAreas(top, topRooms)...)
		areas = append(areas, rcg.createBSPAreas(bottom, bottomRooms)...)
	} else {
		// Can't split further, return the area
		return []pcg.Rectangle{area}
	}

	return areas
}

// selectRoomType determines the appropriate room type for a position
func (rcg *RoomCorridorGenerator) selectRoomType(index, total int, params pcg.LevelParams) pcg.RoomType {
	// First room is entrance
	if index == 0 {
		return pcg.RoomTypeEntrance
	}

	// Last room is exit
	if index == total-1 {
		return pcg.RoomTypeExit
	}

	// Boss room if specified
	if params.HasBoss && index == total-2 {
		return pcg.RoomTypeBoss
	}

	// Select from allowed room types
	if len(params.RoomTypes) > 0 {
		return params.RoomTypes[rcg.rng.Intn(len(params.RoomTypes))]
	}

	// Default room type distribution
	roll := rcg.rng.Float64()
	switch {
	case roll < 0.4:
		return pcg.RoomTypeCombat
	case roll < 0.6:
		return pcg.RoomTypeTreasure
	case roll < 0.75:
		return pcg.RoomTypePuzzle
	case roll < 0.85:
		return pcg.RoomTypeRest
	case roll < 0.95:
		return pcg.RoomTypeTrap
	default:
		return pcg.RoomTypeSecret
	}
}

// ensureSpecialRooms makes sure required special rooms are present
func (rcg *RoomCorridorGenerator) ensureSpecialRooms(roomLayouts []*pcg.RoomLayout, params pcg.LevelParams) {
	hasEntrance := false
	hasExit := false

	for _, room := range roomLayouts {
		if room.Type == pcg.RoomTypeEntrance {
			hasEntrance = true
		}
		if room.Type == pcg.RoomTypeExit {
			hasExit = true
		}
	}

	// Ensure entrance exists
	if !hasEntrance && len(roomLayouts) > 0 {
		roomLayouts[0].Type = pcg.RoomTypeEntrance
	}

	// Ensure exit exists
	if !hasExit && len(roomLayouts) > 1 {
		roomLayouts[len(roomLayouts)-1].Type = pcg.RoomTypeExit
	}
}

// generateRooms creates the actual room content
func (rcg *RoomCorridorGenerator) generateRooms(roomLayouts []*pcg.RoomLayout, params pcg.LevelParams, genCtx *pcg.GenerationContext) error {
	for _, roomLayout := range roomLayouts {
		generator, exists := rcg.roomGenerators[roomLayout.Type]
		if !exists {
			// Use default combat room generator
			generator = rcg.roomGenerators[pcg.RoomTypeCombat]
		}

		generatedRoom, err := generator.GenerateRoom(roomLayout.Bounds, params.LevelTheme, params.Difficulty, genCtx)
		if err != nil {
			return fmt.Errorf("failed to generate room %s: %w", roomLayout.ID, err)
		}

		// Copy generated content back to layout
		roomLayout.Tiles = generatedRoom.Tiles
		roomLayout.Doors = generatedRoom.Doors
		roomLayout.Features = generatedRoom.Features
		roomLayout.Properties = generatedRoom.Properties
	}

	return nil
}

// ConnectRooms generates corridors and passages between rooms
func (rcg *RoomCorridorGenerator) ConnectRooms(ctx context.Context, rooms []*pcg.RoomLayout, params pcg.LevelParams) ([]pcg.Corridor, error) {
	var corridors []pcg.Corridor
	planner := NewCorridorPlanner(params.CorridorStyle, rcg.rng)

	// Create minimum spanning tree for basic connectivity
	connections := rcg.createMinimumConnections(rooms)

	// Generate corridors for each connection
	for i, connection := range connections {
		corridor, err := planner.CreateCorridor(
			fmt.Sprintf("corridor_%d", i),
			connection.Start,
			connection.End,
			params.LevelTheme,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create corridor %d: %w", i, err)
		}

		corridors = append(corridors, *corridor)

		// Update room connections
		rooms[connection.StartRoomIndex].Connected = append(
			rooms[connection.StartRoomIndex].Connected,
			rooms[connection.EndRoomIndex].ID,
		)
		rooms[connection.EndRoomIndex].Connected = append(
			rooms[connection.EndRoomIndex].Connected,
			rooms[connection.StartRoomIndex].ID,
		)
	}

	return corridors, nil
}

// RoomConnection represents a connection between two rooms
type RoomConnection struct {
	Start          game.Position
	End            game.Position
	StartRoomIndex int
	EndRoomIndex   int
}

// createMinimumConnections creates a minimum set of connections for level connectivity
func (rcg *RoomCorridorGenerator) createMinimumConnections(rooms []*pcg.RoomLayout) []RoomConnection {
	var connections []RoomConnection

	// Connect adjacent rooms in a simple chain
	for i := 0; i < len(rooms)-1; i++ {
		startPos := rcg.findConnectionPoint(rooms[i])
		endPos := rcg.findConnectionPoint(rooms[i+1])

		connections = append(connections, RoomConnection{
			Start:          startPos,
			End:            endPos,
			StartRoomIndex: i,
			EndRoomIndex:   i + 1,
		})
	}

	// Add some additional connections for interesting layouts
	if len(rooms) > 3 {
		// Connect first and last room for loop
		startPos := rcg.findConnectionPoint(rooms[0])
		endPos := rcg.findConnectionPoint(rooms[len(rooms)-1])

		connections = append(connections, RoomConnection{
			Start:          startPos,
			End:            endPos,
			StartRoomIndex: 0,
			EndRoomIndex:   len(rooms) - 1,
		})
	}

	return connections
}

// findConnectionPoint finds a good door position on a room
func (rcg *RoomCorridorGenerator) findConnectionPoint(room *pcg.RoomLayout) game.Position {
	// For now, return center of a random wall
	switch rcg.rng.Intn(4) {
	case 0: // Top wall
		return game.Position{
			X: room.Bounds.X + room.Bounds.Width/2,
			Y: room.Bounds.Y,
		}
	case 1: // Right wall
		return game.Position{
			X: room.Bounds.X + room.Bounds.Width - 1,
			Y: room.Bounds.Y + room.Bounds.Height/2,
		}
	case 2: // Bottom wall
		return game.Position{
			X: room.Bounds.X + room.Bounds.Width/2,
			Y: room.Bounds.Y + room.Bounds.Height - 1,
		}
	default: // Left wall
		return game.Position{
			X: room.Bounds.X,
			Y: room.Bounds.Y + room.Bounds.Height/2,
		}
	}
}

// addSpecialFeatures adds special features and encounters to the level
func (rcg *RoomCorridorGenerator) addSpecialFeatures(roomLayouts []*pcg.RoomLayout, params pcg.LevelParams, genCtx *pcg.GenerationContext) error {
	// Add secret rooms if specified
	secretRoomsAdded := 0
	for _, room := range roomLayouts {
		if room.Type == pcg.RoomTypeSecret {
			secretRoomsAdded++
		}
	}

	// Generate additional secret rooms if needed
	for secretRoomsAdded < params.SecretRooms && len(roomLayouts) > 0 {
		// Find a suitable room to add a secret connection to
		targetRoom := roomLayouts[rcg.rng.Intn(len(roomLayouts))]
		if targetRoom.Type != pcg.RoomTypeSecret {
			// Add secret room feature
			targetRoom.Features = append(targetRoom.Features, pcg.RoomFeature{
				Type:     "secret_door",
				Position: rcg.findConnectionPoint(targetRoom),
				Properties: map[string]interface{}{
					"hidden":     true,
					"difficulty": params.Difficulty + 2,
				},
			})
			secretRoomsAdded++
		}
	}

	return nil
}

// validateLevel ensures the level meets quality standards
func (rcg *RoomCorridorGenerator) validateLevel(rooms []*pcg.RoomLayout, corridors []pcg.Corridor) error {
	// Check that all rooms are reachable
	reachable := make(map[string]bool)
	rcg.markReachableRooms(rooms[0].ID, rooms, reachable)

	for _, room := range rooms {
		if !reachable[room.ID] {
			return fmt.Errorf("room %s is not reachable", room.ID)
		}
	}

	// Check that we have at least one entrance and exit
	hasEntrance := false
	hasExit := false
	for _, room := range rooms {
		if room.Type == pcg.RoomTypeEntrance {
			hasEntrance = true
		}
		if room.Type == pcg.RoomTypeExit {
			hasExit = true
		}
	}

	if !hasEntrance {
		return fmt.Errorf("level has no entrance room")
	}
	if !hasExit {
		return fmt.Errorf("level has no exit room")
	}

	return nil
}

// markReachableRooms recursively marks all reachable rooms
func (rcg *RoomCorridorGenerator) markReachableRooms(roomID string, rooms []*pcg.RoomLayout, reachable map[string]bool) {
	if reachable[roomID] {
		return
	}

	reachable[roomID] = true

	// Find the room
	var currentRoom *pcg.RoomLayout
	for _, room := range rooms {
		if room.ID == roomID {
			currentRoom = room
			break
		}
	}

	if currentRoom == nil {
		return
	}

	// Mark connected rooms as reachable
	for _, connectedID := range currentRoom.Connected {
		rcg.markReachableRooms(connectedID, rooms, reachable)
	}
}

// convertToGameLevel converts the generated rooms and corridors to a game.Level
func (rcg *RoomCorridorGenerator) convertToGameLevel(rooms []*pcg.RoomLayout, corridors []pcg.Corridor, width, height int, params pcg.LevelParams) (*game.Level, error) {
	// Create level with basic info
	level := &game.Level{
		ID:         fmt.Sprintf("generated_level_%d", params.Seed),
		Name:       fmt.Sprintf("Generated %s Level", params.LevelTheme),
		Width:      width,
		Height:     height,
		Tiles:      make([][]game.Tile, height),
		Properties: make(map[string]interface{}),
	}

	// Initialize tiles with walls
	for y := 0; y < height; y++ {
		level.Tiles[y] = make([]game.Tile, width)
		for x := 0; x < width; x++ {
			level.Tiles[y][x] = game.Tile{
				Type:       game.TileWall,
				Walkable:   false,
				Properties: make(map[string]interface{}),
			}
		}
	}

	// Place room tiles
	for _, room := range rooms {
		for y := 0; y < len(room.Tiles) && room.Bounds.Y+y < height; y++ {
			for x := 0; x < len(room.Tiles[y]) && room.Bounds.X+x < width; x++ {
				level.Tiles[room.Bounds.Y+y][room.Bounds.X+x] = room.Tiles[y][x]
			}
		}
	}

	// Place corridor tiles
	for _, corridor := range corridors {
		for _, pos := range corridor.Path {
			if pos.X >= 0 && pos.X < width && pos.Y >= 0 && pos.Y < height {
				level.Tiles[pos.Y][pos.X] = game.Tile{
					Type:       game.TileFloor,
					Walkable:   true,
					Properties: make(map[string]interface{}),
				}
			}
		}
	}

	// Add level metadata
	level.Properties["theme"] = params.LevelTheme
	level.Properties["difficulty"] = params.Difficulty
	level.Properties["room_count"] = len(rooms)
	level.Properties["corridor_count"] = len(corridors)
	level.Properties["generator"] = "room_corridor"
	level.Properties["version"] = rcg.version

	return level, nil
}

// GenerateRoom creates a single room with specified constraints
func (rcg *RoomCorridorGenerator) GenerateRoom(ctx context.Context, bounds pcg.Rectangle, roomType pcg.RoomType, params pcg.LevelParams) (*pcg.RoomLayout, error) {
	generator, exists := rcg.roomGenerators[roomType]
	if !exists {
		return nil, fmt.Errorf("no generator available for room type: %s", roomType)
	}

	seedMgr := pcg.NewSeedManager(params.Seed)
	genCtx := pcg.NewGenerationContext(seedMgr, pcg.ContentTypeLevels, "room_generation", params.GenerationParams)

	return generator.GenerateRoom(bounds, params.LevelTheme, params.Difficulty, genCtx)
}
