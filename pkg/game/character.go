package game

import (
	"encoding/json"
	"fmt"
	"sync"
)

// Character represents the base attributes for both Players and NPCs
// Contains all attributes, stats, and equipment for game entities
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

func (c *Character) GetHealth() int {
	return c.HP
}

func (c *Character) IsObstacle() bool {
	// Characters are considered obstacles for movement/pathing
	return true
}

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
func (c *Character) GetID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ID
}

func (c *Character) GetName() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Name
}

func (c *Character) GetDescription() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Description
}

func (c *Character) GetPosition() Position {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Position
}

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

func (c *Character) IsActive() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.active
}

func (c *Character) GetTags() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return append([]string{}, c.tags...) // Return copy to prevent modification
}

func (c *Character) ToJSON() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return json.Marshal(c)
}

func (c *Character) FromJSON(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return json.Unmarshal(data, c)
}
