package game

import (
	"fmt"
	"sync"
)

// World manages the game state and all game objects
// Contains the complete state of the game world including all entities and maps
type World struct {
	mu          sync.RWMutex          `yaml:"-"`                  // Protects concurrent access
	Levels      []Level               `yaml:"world_levels"`       // All game levels/maps
	CurrentTime GameTime              `yaml:"world_current_time"` // Current game time
	Objects     map[string]GameObject `yaml:"world_objects"`      // All game objects by ID
	Players     map[string]*Player    `yaml:"world_players"`      // Active players by ID
	NPCs        map[string]*NPC       `yaml:"world_npcs"`         // Non-player characters by ID
	SpatialGrid map[Position][]string `yaml:"world_spatial_grid"` // Spatial index of objects
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
		Objects:     make(map[string]GameObject),
		Players:     make(map[string]*Player),
		NPCs:        make(map[string]*NPC),
		SpatialGrid: make(map[Position][]string),
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
	pos := obj.GetPosition()
	w.SpatialGrid[pos] = append(w.SpatialGrid[pos], obj.GetID())

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
