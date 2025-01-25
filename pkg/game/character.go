package game

import (
	"encoding/json"
	"fmt"
	"sync"
)

// Character represents the base attributes for both Players and NPCs
// Contains all attributes, stats, and equipment for game entities
// Character represents a playable character or NPC in the game world.
// It encapsulates all attributes, stats, and inventory management for characters.
//
// Key features:
// - Thread-safe with sync.RWMutex protection
// - Complete attribute system (Strength, Dexterity etc)
// - Combat stats tracking (HP, AC, THAC0)
// - Equipment and inventory management
// - Position tracking in game world
// - Tagging system for special attributes
//
// The Character struct uses YAML tags for persistence and serialization.
// All numeric fields use int type for simplicity and compatibility.
//
// Related types:
// - Position: Represents location in game world
// - Item: Represents equipment and inventory items
// - EquipmentSlot: Equipment slot enumeration
//
// Thread safety:
// All operations that modify Character fields should hold mu.Lock()
// Read operations should hold mu.RLock()
type Character struct {
	mu          sync.RWMutex `yaml:"-"`                // Protects concurrent access to character data
	ID          string       `yaml:"char_id"`          // Unique identifier
	Name        string       `yaml:"char_name"`        // Character's name
	Description string       `yaml:"char_description"` // Character's description
	Position    Position     `yaml:"char_position"`    // Current location in game world

	// Attributes
	Strength     int `yaml:"attr_strength"`     // Physical power
	Dexterity    int `yaml:"attr_dexterity"`    // Agility and reflexes
	Constitution int `yaml:"attr_constitution"` // Health and stamina
	Intelligence int `yaml:"attr_intelligence"` // Learning and reasoning
	Wisdom       int `yaml:"attr_wisdom"`       // Intuition and perception
	Charisma     int `yaml:"attr_charisma"`     // Leadership and personality

	// Combat stats
	HP         int `yaml:"combat_current_hp"`  // Current hit points
	MaxHP      int `yaml:"combat_max_hp"`      // Maximum hit points
	ArmorClass int `yaml:"combat_armor_class"` // Defense rating
	THAC0      int `yaml:"combat_thac0"`       // To Hit Armor Class 0

	// Equipment and inventory
	Equipment map[EquipmentSlot]Item `yaml:"char_equipment"` // Equipped items by slot
	Inventory []Item                 `yaml:"char_inventory"` // Carried items
	Gold      int                    `yaml:"char_gold"`      // Currency amount

	active bool     `yaml:"char_active"` // Whether character is active in game
	tags   []string `yaml:"char_tags"`   // Special attributes or markers
}

// Clone creates and returns a deep copy of the Character.
// This method is thread-safe and creates a completely independent copy
// of the character including all nested structures.
//
// Returns:
//   - *Character: A pointer to the new cloned Character instance
//
// Thread safety:
//   - Uses RLock to ensure safe concurrent access during cloning
func (c *Character) Clone() *Character {
	c.mu.RLock()
	defer c.mu.RUnlock()

	clone := &Character{
		ID:           c.ID,
		Name:         c.Name,
		Description:  c.Description,
		Position:     c.Position,
		Strength:     c.Strength,
		Dexterity:    c.Dexterity,
		Constitution: c.Constitution,
		Intelligence: c.Intelligence,
		Wisdom:       c.Wisdom,
		Charisma:     c.Charisma,
		HP:           c.HP,
		MaxHP:        c.MaxHP,
		ArmorClass:   c.ArmorClass,
		THAC0:        c.THAC0,
		Equipment:    make(map[EquipmentSlot]Item),
		Inventory:    make([]Item, len(c.Inventory)),
		Gold:         c.Gold,
		active:       c.active,
		tags:         make([]string, len(c.tags)),
	}

	// Deep copy equipment map
	for slot, item := range c.Equipment {
		clone.Equipment[slot] = item
	}

	// Deep copy inventory slice
	copy(clone.Inventory, c.Inventory)

	// Deep copy tags slice
	copy(clone.tags, c.tags)

	return clone
}

// GetHealth returns the current hit points (HP) of the Character.
//
// Returns:
//   - int: The current health points value
//
// Related:
//   - Character.HP field
//   - Character.SetHealth (if exists)
func (c *Character) GetHealth() int {
	return c.HP
}

// IsObstacle indicates if this Character should be treated as an obstacle for movement/pathing.
// In the current implementation, all Characters are always considered obstacles.
//
// Returns:
//   - bool: Always returns true since Characters are obstacles by default
//
// Related:
//   - Used by pathing and collision detection systems
func (c *Character) IsObstacle() bool {
	// Characters are considered obstacles for movement/pathing
	return true
}

// SetHealth updates the character's current health points (HP) with the provided value.
// The health value will be constrained between 0 and the character's maximum HP.
//
// Parameters:
//   - health: The new health value to set (integer)
//
// Edge cases handled:
//   - Health below 0 is capped at 0
//   - Health above MaxHP is capped at MaxHP
//
// Related fields:
//   - Character.HP
//   - Character.MaxHP
func (c *Character) SetHealth(health int) {
	c.HP = health
	// Ensure health doesn't go below 0
	if c.HP < 0 {
		c.HP = 0
	}
	// Cap health at max health
	if c.HP > c.MaxHP {
		c.HP = c.MaxHP
	}
}

// Implement GameObject interface methods
// GetID returns the unique identifier string for this Character instance.
// It uses a read lock to safely access the ID field in a concurrent context.
// Returns the character's unique ID string.
// Related: Character struct, ID field
func (c *Character) GetID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ID
}

// GetName returns the name of the Character.
//
// This method is thread-safe and uses a read lock to safely access the character's name.
//
// Returns:
//   - string: The name of the character
//
// Related:
//   - Character struct
func (c *Character) GetName() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Name
}

// GetDescription returns the character's description as a string.
// This method is thread-safe as it uses a read lock when accessing the description field.
// Returns:
//   - string: The character's description text
//
// Related:
//   - Character struct
//   - Character.SetDescription()
func (c *Character) GetDescription() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Description
}

// GetPosition returns the current position of the Character.
// This method is thread-safe and uses read locking to protect concurrent access.
// Returns a Position struct containing the character's x,y coordinates.
// Related types:
// - Position struct
func (c *Character) GetPosition() Position {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Position
}

// SetPosition updates the character's position to the specified coordinates after validation.
//
// Parameters:
//   - pos Position: The new position coordinates to set
//
// Returns:
//   - error: nil if successful, error if position is invalid
//
// Errors:
//   - Returns error if position fails validation check
//
// Thread Safety:
//   - Method is thread-safe using mutex locking
//
// Related:
//   - isValidPosition() - Helper function that validates position coordinates
func (c *Character) SetPosition(pos Position) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Validate position before setting
	if !isValidPosition(pos) {
		return fmt.Errorf("invalid position: %v", pos)
	}

	c.Position = pos
	return nil
}

// IsActive returns the current active state of the Character.
// This method is concurrent-safe through use of a read lock.
//
// Returns:
//   - bool: true if the character is active, false otherwise
//
// Thread-safety: This method uses RLock/RUnlock for concurrent access
func (c *Character) IsActive() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.active
}

// SetActive sets the active state of the character.
// Thread-safe method that controls whether the character is active in the game.
//
// Parameters:
//   - active: bool - The desired active state for the character
//
// Thread safety:
//
//	Uses mutex locking to ensure thread-safe access to the active state
//
// Related:
//   - Character struct (contains the active field being modified)
func (c *Character) SetActive(active bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.active = active
}

// GetTags returns a copy of the character's tags list.
//
// This method provides thread-safe access to the character's tags by using a read lock.
// A new slice containing copies of all tags is returned to prevent external modifications
// to the character's internal state.
//
// Returns:
//
//	[]string - A new slice containing copies of all the character's tags
//
// Related:
//
//	Character.AddTag() - For adding new tags
//	Character.RemoveTag() - For removing existing tags
func (c *Character) GetTags() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return append([]string{}, c.tags...) // Return copy to prevent modification
}

// ToJSON serializes the Character struct to JSON format with thread safety.
//
// This method acquires a read lock on the character to ensure safe concurrent access
// during serialization.
//
// Returns:
//   - []byte: The JSON encoded representation of the Character
//   - error: Any error that occurred during marshaling
//
// Related:
//   - FromJSON() for deserialization
//   - json.Marshal() from encoding/json
func (c *Character) ToJSON() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return json.Marshal(c)
}

// FromJSON unmarshals a JSON byte array into the Character struct.
// This method is thread-safe as it uses a mutex lock.
//
// Parameters:
//   - data []byte: JSON encoded byte array containing character data
//
// Returns:
//   - error: Returns any error that occurred during unmarshaling
//
// Related:
//   - Character.ToJSON() for serialization
//   - json.Unmarshal() from encoding/json package
func (c *Character) FromJSON(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return json.Unmarshal(data, c)
}
