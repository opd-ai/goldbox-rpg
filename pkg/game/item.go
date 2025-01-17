package game

import (
	"encoding/json"
	"fmt"
)

// Item represents a game item with its properties
// Contains all attributes that define an item's behavior and characteristics
// Item represents a game item with various attributes and properties.
// It is used to define objects that players can interact with in the game world.
//
// Fields:
//   - ID (string): Unique identifier used to reference the item in the game
//   - Name (string): Human-readable display name of the item
//   - Type (string): Category classification (e.g. "weapon", "armor", "potion")
//   - Damage (string): Optional damage specification for weapons (e.g. "1d6")
//   - AC (int): Optional armor class value for defensive equipment
//   - Weight (int): Weight of the item in game units
//   - Value (int): Worth of the item in game currency
//   - Properties ([]string): Optional list of special effects or attributes
//   - Position (Position): Optional current location in the game world
//
// The Item struct is serializable to/from YAML format using the specified tags.
// Related types:
//   - Position: Represents location coordinates in the game world
type Item struct {
	ID         string   `yaml:"item_id"`                    // Unique identifier for the item
	Name       string   `yaml:"item_name"`                  // Display name of the item
	Type       string   `yaml:"item_type"`                  // Category of item (weapon, armor, etc.)
	Damage     string   `yaml:"item_damage,omitempty"`      // Damage specification for weapons
	AC         int      `yaml:"item_armor_class,omitempty"` // Armor class for defensive items
	Weight     int      `yaml:"item_weight"`                // Weight in game units
	Value      int      `yaml:"item_value"`                 // Monetary value in game currency
	Properties []string `yaml:"item_properties,omitempty"`  // Special properties or effects
	Position   Position `yaml:"item_position,omitempty"`    // Current location in game world
}

// FromJSON implements GameObject.
// FromJSON deserializes JSON data into an Item struct.
//
// Parameters:
//   - data []byte: Raw JSON bytes to deserialize
//
// Returns:
//   - error: Returns an error if JSON unmarshaling fails
//
// Related:
//   - Item.ToJSON() for the inverse serialization operation
func (i *Item) FromJSON(data []byte) error {
	return json.Unmarshal(data, i)
}

// GetDescription implements GameObject.
// GetDescription returns a formatted string representation of the item
// combining its Name and Type properties.
//
// Returns a string in the format "Name (Type)"
//
// Related types:
// - Item struct
func (i *Item) GetDescription() string {
	return fmt.Sprintf("%s (%s)", i.Name, i.Type)
}

// GetHealth implements GameObject.
// GetHealth returns the health value of an Item.
// Since items don't inherently have health in this implementation, it always returns 0.
// This method satisfies an interface but has no practical effect for basic Item objects.
// Returns:
//   - int: Always returns 0 for base items
//
// Related types:
//   - Item struct
func (i *Item) GetHealth() int {
	return 0 // Items don't have health
}

// GetID implements GameObject.
// GetID returns the unique identifier string for this Item.
// This method provides access to the private ID field.
// Returns a string representing the item's unique identifier.
// Related: Item struct
func (i *Item) GetID() string {
	return i.ID
}

// GetName implements GameObject.
// GetName returns the name of the item
//
// Returns:
//   - string: The name property of the Item struct
func (i *Item) GetName() string {
	return i.Name
}

// GetPosition implements GameObject.
// GetPosition returns the current position of this item in the game world.
// If the item's position has not been explicitly set, returns an empty Position struct.
// Returns:
//   - Position: The x,y coordinates of the item
//
// Related types:
//   - Position struct
func (i *Item) GetPosition() Position {
	return Position{} // Default position if not set
}

// GetTags implements GameObject.
// GetTags returns the Properties field of an Item, which contains string tags/attributes
// associated with this item. The returned slice can be empty if no properties are set.
//
// Returns:
//   - []string: A slice of strings representing the item's properties/tags
//
// Related:
//   - Item struct
//   - Properties field
func (i *Item) GetTags() []string {
	return i.Properties
}

// IsActive implements GameObject.
func (i *Item) IsActive() bool {
	return true // Items are always active
}

// IsObstacle implements GameObject.
func (i *Item) IsObstacle() bool {
	return false // Items are not obstacles
}

// SetHealth implements GameObject.
func (i *Item) SetHealth(health int) {
	// Items don't have health, no-op
}

// SetPosition implements GameObject.
func (i *Item) SetPosition(pos Position) error {
	return nil // Items don't track position
}

// ToJSON implements GameObject.
func (i *Item) ToJSON() ([]byte, error) {
	return json.Marshal(i)
}

// ItemType constants
// ItemTypeWeapon represents a weapon item type constant used for categorizing items
// in the game inventory and equipment system. This type is used when creating or
// identifying weapon items.
const (
	ItemTypeWeapon = "weapon"
	ItemTypeArmor  = "armor"
)
