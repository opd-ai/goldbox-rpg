package game

import (
	"encoding/json"
	"fmt"
)

// Item represents a game item with its properties
// Contains all attributes that define an item's behavior and characteristics
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
func (i *Item) FromJSON(data []byte) error {
	return json.Unmarshal(data, i)
}

// GetDescription implements GameObject.
func (i *Item) GetDescription() string {
	return fmt.Sprintf("%s (%s)", i.Name, i.Type)
}

// GetHealth implements GameObject.
func (i *Item) GetHealth() int {
	return 0 // Items don't have health
}

// GetID implements GameObject.
func (i *Item) GetID() string {
	return i.ID
}

// GetName implements GameObject.
func (i *Item) GetName() string {
	return i.Name
}

// GetPosition implements GameObject.
func (i *Item) GetPosition() Position {
	return Position{} // Default position if not set
}

// GetTags implements GameObject.
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
const (
	ItemTypeWeapon = "weapon"
	ItemTypeArmor  = "armor"
)
