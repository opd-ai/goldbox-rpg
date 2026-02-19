package game

// Direction represents a cardinal direction in 2D space.
// It is implemented as an integer type to allow for efficient
// direction comparisons and calculations.
type Direction int

// Position represents the location and orientation of an entity in the game world.
// It tracks both the 2D grid coordinates and vertical level for 3D positioning,
// as well as which direction the entity is facing.
//
// Fields:
//   - X: Horizontal position on the map grid (integer)
//   - Y: Vertical position on the map grid (integer)
//   - Level: Current depth/floor number in the dungeon (integer)
//   - Facing: Direction the entity is oriented (Direction enum)
//
// Related types:
//   - Direction: Used for the Facing field to indicate orientation
//
// The Position struct uses YAML tags for serialization/deserialization
type Position struct {
	X      int       `yaml:"position_x"`      // X coordinate on the map grid
	Y      int       `yaml:"position_y"`      // Y coordinate on the map grid
	Level  int       `yaml:"position_level"`  // Current dungeon/map level
	Facing Direction `yaml:"position_facing"` // Direction the entity is facing
}

// DirectionConfig represents the configuration for a directional value in the game system.
// It encapsulates direction-related properties including numeric values, names and angular measurements.
//
// Fields:
//   - Value: Direction type representing the numeric/enum value of the direction
//   - Name: String name of the direction (e.g. "North", "East")
//   - DegreeAngle: Integer angle in degrees, must be one of: 0, 90, 180, 270
//
// The DirectionConfig struct is typically loaded from YAML configuration files
// and used to define cardinal directions in the game world.
//
// Related types:
//   - Direction (enum type)
type DirectionConfig struct {
	Value       Direction `yaml:"direction_value"` // Numeric value of the direction
	Name        string    `yaml:"direction_name"`  // String representation (North, East, etc.)
	DegreeAngle int       `yaml:"direction_angle"` // Angle in degrees (0, 90, 180, 270)
}

// Identifiable represents entities with unique identification in the game world.
// This interface provides the minimum contract for identifying game entities.
type Identifiable interface {
	GetID() string
	GetName() string
	GetDescription() string
}

// Positionable represents entities that have a position in the game world.
// This interface is used by the spatial index for efficient spatial queries.
type Positionable interface {
	GetPosition() Position
	SetPosition(Position) error
}

// Damageable represents entities that have health and can take damage.
// Not all game objects need health (e.g., items, terrain), so this is
// a separate interface following the Interface Segregation Principle.
type Damageable interface {
	GetHealth() int
	SetHealth(int)
}

// Serializable represents entities that can be serialized to/from JSON.
type Serializable interface {
	ToJSON() ([]byte, error)
	FromJSON([]byte) error
}

// GameObject represents a base interface for all game objects in the RPG system.
// It composes smaller interfaces following the Interface Segregation Principle:
//   - Identifiable: ID, name, description
//   - Positionable: position management
//   - Damageable: health management (implemented by characters, not items)
//   - Serializable: JSON serialization
//
// Core capabilities include:
// - Unique identification (GetID)
// - Basic properties (name, description, position)
// - State management (active status, health)
// - Tag-based classification
// - JSON serialization/deserialization
// - Collision detection (obstacle status)
//
// Related types:
// - Position: Represents the object's location in the game world
//
// Implementation note:
// All game objects should implement this interface to ensure consistent behavior
// across the game system. This enables uniform handling of different object types
// in the game loop and collision detection systems.
//
// The interface is designed to be extensible - additional specialized interfaces
// can embed GameObject to add more specific functionality while maintaining
// compatibility with base game systems.
type GameObject interface {
	Identifiable
	Positionable
	Damageable
	Serializable
	IsActive() bool
	GetTags() []string
	IsObstacle() bool
}

// EffectHolder represents entities that can have effects applied to them.
// It defines the core functionality for effect management and stat modification.
// - Stats: Contains the actual stat values
// - EffectType: Enumeration of possible effect types
// Moved from: effectmanager.go
type EffectHolder interface {
	// Effect management
	AddEffect(effect *Effect) error
	RemoveEffect(effectID string) error
	HasEffect(effectType EffectType) bool
	GetEffects() []*Effect

	// Stats that can be modified by effects
	GetStats() *Stats
	SetStats(*Stats)

	// Base stats before effects
	GetBaseStats() *Stats
}

// EffectTyper defines the interface for objects that have an effect type.
// Related types:
//   - EffectType: The enumeration of possible effect types
//
// Moved from: effects.go
type EffectTyper interface {
	GetEffectType() EffectType
}
