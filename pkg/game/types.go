package game

// Direction represents cardinal directions in the game world
type Direction int

const (
	North Direction = iota
	East
	South
	West
)

// Position represents a location in the game world
// Contains coordinates and facing direction for precise positioning
type Position struct {
	X      int       `yaml:"position_x"`      // X coordinate on the map grid
	Y      int       `yaml:"position_y"`      // Y coordinate on the map grid
	Level  int       `yaml:"position_level"`  // Current dungeon/map level
	Facing Direction `yaml:"position_facing"` // Direction the entity is facing
}

// DirectionConfig represents a serializable direction configuration
type DirectionConfig struct {
	Value       Direction `yaml:"direction_value"` // Numeric value of the direction
	Name        string    `yaml:"direction_name"`  // String representation (North, East, etc.)
	DegreeAngle int       `yaml:"direction_angle"` // Angle in degrees (0, 90, 180, 270)
}

// GameObject defines the interface for all interactive entities
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
