package game

// Direction represents a cardinal direction in 2D space.
// It is implemented as an integer type to allow for efficient
// direction comparisons and calculations.
type Direction int

// Direction constants represent the four cardinal directions.
// These values are used throughout the game for movement, facing, and orientation.
// The values increment clockwise starting from North (0).
const (
	North Direction = iota // North direction (0 degrees)
	East                   // East direction (90 degrees)
	South                  // South direction (180 degrees)
	West                   // West direction (270 degrees)
)

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

// GameObject represents a base interface for all game objects in the RPG system.
// It defines the core functionality that every game object must implement.
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
	GetID() string
	GetName() string
	GetDescription() string
	GetPosition() Position
	SetPosition(Position) error
	IsActive() bool
	GetTags() []string
	ToJSON() ([]byte, error)
	FromJSON([]byte) error
	GetHealth() int
	SetHealth(int)
	IsObstacle() bool
}
