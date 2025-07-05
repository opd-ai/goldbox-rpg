package game

import (
	"fmt"
	"math"
	"sync"
)

// World manages the game state and all game objects
// Contains the complete state of the game world including all entities and maps
type World struct {
	mu           sync.RWMutex          `yaml:"-"`                  // Protects concurrent access
	Levels       []Level               `yaml:"world_levels"`       // All game levels/maps
	CurrentTime  GameTime              `yaml:"world_current_time"` // Current game time
	Objects      map[string]GameObject `yaml:"world_objects"`      // All game objects by ID
	Players      map[string]*Player    `yaml:"world_players"`      // Active players by ID
	NPCs         map[string]*NPC       `yaml:"world_npcs"`         // Non-player characters by ID
	SpatialGrid  map[Position][]string `yaml:"world_spatial_grid"` // Legacy spatial index (for compatibility)
	SpatialIndex *SpatialIndex         `yaml:"-"`                  // Advanced spatial indexing system
	Width        int                   `yaml:"world_width"`        // Width of the world
	Height       int                   `yaml:"world_height"`       // Height of the world
}

// Update applies a set of updates to the World state
func (w *World) Update(worldUpdates map[string]interface{}) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for key, value := range worldUpdates {
		switch key {
		case "objects":
			if objects, ok := value.(map[string]GameObject); ok {
				for id, obj := range objects {
					w.Objects[id] = obj
					pos := obj.GetPosition()
					w.SpatialGrid[pos] = append(w.SpatialGrid[pos], obj.GetID())
				}
			}
		case "players":
			if players, ok := value.(map[string]*Player); ok {
				for id, player := range players {
					w.Players[id] = player
				}
			}
		case "npcs":
			if npcs, ok := value.(map[string]*NPC); ok {
				for id, npc := range npcs {
					w.NPCs[id] = npc
				}
			}
		case "current_time":
			if time, ok := value.(GameTime); ok {
				w.CurrentTime = time
			}
		default:
			return fmt.Errorf("unknown update key: %s", key)
		}
	}

	return nil
}

// Clone creates a deep copy of the World
func (w *World) Clone() *World {
	w.mu.RLock()
	defer w.mu.RUnlock()

	clone := &World{
		Levels:      make([]Level, len(w.Levels)),
		CurrentTime: w.CurrentTime,
		Objects:     make(map[string]GameObject),
		Players:     make(map[string]*Player),
		NPCs:        make(map[string]*NPC),
		SpatialGrid: make(map[Position][]string),
		Width:       w.Width,
		Height:      w.Height,
	}

	// Deep copy levels
	copy(clone.Levels, w.Levels)

	// Copy objects
	for k, v := range w.Objects {
		clone.Objects[k] = v
	}

	// Copy players
	for k, v := range w.Players {
		clone.Players[k] = v
	}

	// Copy NPCs
	for k, v := range w.NPCs {
		clone.NPCs[k] = v
	}

	// Copy spatial grid
	for k, v := range w.SpatialGrid {
		gridCopy := make([]string, len(v))
		copy(gridCopy, v)
		clone.SpatialGrid[k] = gridCopy
	}

	// Clone spatial index by rebuilding it with all objects
	if w.SpatialIndex != nil {
		clone.SpatialIndex = NewSpatialIndex(w.Width, w.Height, w.SpatialIndex.cellSize)
		for _, obj := range clone.Objects {
			clone.SpatialIndex.Insert(obj)
		}
	}

	return clone
}

// WorldState represents the serializable state of the world
// Used for saving/loading game state
type WorldState struct {
	WorldVersion string     `yaml:"world_version"`       // World data version
	LastSaved    GameTime   `yaml:"world_last_saved"`    // Last save timestamp
	ActiveLevels []string   `yaml:"world_active_levels"` // Currently active level IDs
	Statistics   WorldStats `yaml:"world_stats"`         // World statistics
}

// WorldStats tracks various world statistics
type WorldStats struct {
	TotalPlayers  int `yaml:"stat_total_players"`  // Total number of players
	ActiveNPCs    int `yaml:"stat_active_npcs"`    // Current active NPCs
	LoadedObjects int `yaml:"stat_loaded_objects"` // Total loaded objects
	ActiveQuests  int `yaml:"stat_active_quests"`  // Current active quests
	WorldAge      int `yaml:"stat_world_age"`      // Time since world creation
}

// WorldConfig represents world configuration settings
type WorldConfig struct {
	MaxPlayers      int      `yaml:"config_max_players"`      // Maximum allowed players
	MaxLevel        int      `yaml:"config_max_level"`        // Maximum character level
	StartingLevel   string   `yaml:"config_starting_level"`   // Initial player level ID
	EnabledFeatures []string `yaml:"config_enabled_features"` // Enabled world features
}

// NewWorld creates a new game world instance
func NewWorld() *World {
	return &World{
		Objects:      make(map[string]GameObject),
		Players:      make(map[string]*Player),
		NPCs:         make(map[string]*NPC),
		SpatialGrid:  make(map[Position][]string),
		SpatialIndex: nil, // Initialize as nil by default to maintain compatibility
		Width:        0,   // Default width 0 for compatibility
		Height:       0,   // Default height 0 for compatibility
	}
}

// NewWorldWithSize creates a new game world instance with specified dimensions
func NewWorldWithSize(width, height, cellSize int) *World {
	return &World{
		Objects:      make(map[string]GameObject),
		Players:      make(map[string]*Player),
		NPCs:         make(map[string]*NPC),
		SpatialGrid:  make(map[Position][]string),
		SpatialIndex: NewSpatialIndex(width, height, cellSize),
		Width:        width,
		Height:       height,
	}
}

// AddObject safely adds a GameObject to the world
func (w *World) AddObject(obj GameObject) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, exists := w.Objects[obj.GetID()]; exists {
		return fmt.Errorf("object with ID %s already exists", obj.GetID())
	}

	w.Objects[obj.GetID()] = obj

	// Update legacy spatial grid for compatibility
	pos := obj.GetPosition()
	w.SpatialGrid[pos] = append(w.SpatialGrid[pos], obj.GetID())

	// Update advanced spatial index
	if w.SpatialIndex != nil {
		if err := w.SpatialIndex.Insert(obj); err != nil {
			// If spatial index fails, we still keep the object in Objects map
			// This ensures compatibility even if spatial indexing has issues
			return fmt.Errorf("failed to add object to spatial index: %w", err)
		}
	}

	return nil
}

// GetObjectsAt returns all objects at a given position
func (w *World) GetObjectsAt(pos Position) []GameObject {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var objects []GameObject
	for _, id := range w.SpatialGrid[pos] {
		if obj, exists := w.Objects[id]; exists {
			objects = append(objects, obj)
		}
	}

	return objects
}

// ValidateMove checks if the move is valid for the given player and position
func (w *World) ValidateMove(player *Player, newPos Position) error {
	// Check if the new position is within the bounds of the world
	if !w.isPositionWithinBounds(newPos) {
		return fmt.Errorf("position out of bounds")
	}

	// Check if the new position is occupied by an obstacle
	objectsAtNewPos := w.GetObjectsAt(newPos)
	for _, obj := range objectsAtNewPos {
		if obj.IsObstacle() {
			return fmt.Errorf("position occupied by an obstacle")
		}
	}

	// Additional validation logic can be added here (e.g., checking player abilities)

	return nil
}

// isPositionWithinBounds checks if the given position is within the bounds of the world
func (w *World) isPositionWithinBounds(pos Position) bool {
	// Implement the logic to check if the position is within the bounds of the world
	return pos.X >= 0 && pos.X < w.Width && pos.Y >= 0 && pos.Y < w.Height
}

// Serialize returns a map representation of the World state
func (w *World) Serialize() map[string]interface{} {
	return map[string]interface{}{
		"objects": w.Objects,
	}
}

// GetObjectsInRange returns all objects within a rectangular area using advanced spatial indexing
func (w *World) GetObjectsInRange(rect Rectangle) []GameObject {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if w.SpatialIndex != nil {
		return w.SpatialIndex.GetObjectsInRange(rect)
	}

	// Fallback to legacy method if spatial index not available
	var objects []GameObject
	for _, obj := range w.Objects {
		pos := obj.GetPosition()
		if pos.X >= rect.MinX && pos.X <= rect.MaxX &&
			pos.Y >= rect.MinY && pos.Y <= rect.MaxY {
			objects = append(objects, obj)
		}
	}
	return objects
}

// GetObjectsInRadius returns all objects within a circular area using advanced spatial indexing
func (w *World) GetObjectsInRadius(center Position, radius float64) []GameObject {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if w.SpatialIndex != nil {
		return w.SpatialIndex.GetObjectsInRadius(center, radius)
	}

	// Fallback to legacy method if spatial index not available
	var objects []GameObject
	for _, obj := range w.Objects {
		pos := obj.GetPosition()
		dx := float64(center.X - pos.X)
		dy := float64(center.Y - pos.Y)
		distance := math.Sqrt(dx*dx + dy*dy)
		if distance <= radius {
			objects = append(objects, obj)
		}
	}
	return objects
}

// GetNearestObjects returns the k nearest objects to a given position using advanced spatial indexing
func (w *World) GetNearestObjects(center Position, k int) []GameObject {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if w.SpatialIndex != nil {
		return w.SpatialIndex.GetNearestObjects(center, k)
	}

	// Fallback to legacy method if spatial index not available
	type objectDistance struct {
		obj      GameObject
		distance float64
	}

	var candidates []objectDistance
	for _, obj := range w.Objects {
		pos := obj.GetPosition()
		dx := float64(center.X - pos.X)
		dy := float64(center.Y - pos.Y)
		distance := math.Sqrt(dx*dx + dy*dy)
		candidates = append(candidates, objectDistance{obj, distance})
	}

	// Simple bubble sort for small k values
	for i := 0; i < len(candidates)-1; i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[i].distance > candidates[j].distance {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	var result []GameObject
	limit := k
	if len(candidates) < k {
		limit = len(candidates)
	}
	for i := 0; i < limit; i++ {
		result = append(result, candidates[i].obj)
	}
	return result
}

// UpdateObjectPosition updates an object's position in both legacy and advanced spatial indexes
func (w *World) UpdateObjectPosition(objectID string, newPos Position) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	obj, exists := w.Objects[objectID]
	if !exists {
		return fmt.Errorf("object with ID %s not found", objectID)
	}

	oldPos := obj.GetPosition()

	// Update the object's position
	if err := obj.SetPosition(newPos); err != nil {
		return fmt.Errorf("failed to set object position: %w", err)
	}

	// Update legacy spatial grid
	// Remove from old position
	if oldObjects, exists := w.SpatialGrid[oldPos]; exists {
		for i, id := range oldObjects {
			if id == objectID {
				w.SpatialGrid[oldPos] = append(oldObjects[:i], oldObjects[i+1:]...)
				break
			}
		}
		if len(w.SpatialGrid[oldPos]) == 0 {
			delete(w.SpatialGrid, oldPos)
		}
	}
	// Add to new position
	w.SpatialGrid[newPos] = append(w.SpatialGrid[newPos], objectID)

	// Update advanced spatial index
	if w.SpatialIndex != nil {
		if err := w.SpatialIndex.Update(objectID, newPos); err != nil {
			return fmt.Errorf("failed to update object in spatial index: %w", err)
		}
	}

	return nil
}

// RemoveObject safely removes a GameObject from the world and all spatial indexes
func (w *World) RemoveObject(objectID string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	obj, exists := w.Objects[objectID]
	if !exists {
		return fmt.Errorf("object with ID %s not found", objectID)
	}

	pos := obj.GetPosition()

	// Remove from objects map
	delete(w.Objects, objectID)

	// Remove from legacy spatial grid
	if objects, exists := w.SpatialGrid[pos]; exists {
		for i, id := range objects {
			if id == objectID {
				w.SpatialGrid[pos] = append(objects[:i], objects[i+1:]...)
				break
			}
		}
		if len(w.SpatialGrid[pos]) == 0 {
			delete(w.SpatialGrid, pos)
		}
	}

	// Remove from advanced spatial index
	if w.SpatialIndex != nil {
		if err := w.SpatialIndex.Remove(objectID); err != nil {
			return fmt.Errorf("failed to remove object from spatial index: %w", err)
		}
	}

	return nil
}

// GetSpatialIndexStats returns performance statistics for the spatial indexing system
func (w *World) GetSpatialIndexStats() *SpatialIndexStats {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if w.SpatialIndex != nil {
		stats := w.SpatialIndex.GetStats()
		return &stats
	}
	return nil
}
