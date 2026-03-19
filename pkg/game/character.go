package game

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

func init() {
	// Configure structured logging with caller context
	logrus.SetReportCaller(true)
}

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

	// Character class
	Class CharacterClass `yaml:"char_class"` // Character's class (Fighter, Mage, etc.)

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

	// Action points for turn-based combat
	ActionPoints    int `yaml:"combat_action_points"`     // Current action points available
	MaxActionPoints int `yaml:"combat_max_action_points"` // Maximum action points per turn

	// Character progression
	Level      int   `yaml:"char_level"`      // Current character level
	Experience int64 `yaml:"char_experience"` // Experience points accumulated

	// Equipment and inventory
	Equipment map[EquipmentSlot]Item `yaml:"char_equipment"` // Equipped items by slot
	Inventory []Item                 `yaml:"char_inventory"` // Carried items
	Gold      int                    `yaml:"char_gold"`      // Currency amount

	// Effect management
	EffectManager *EffectManager `yaml:"-"` // Manages active effects on character

	// Cosmetic & biographical appearance (no gameplay impact)
	Appearance Appearance `yaml:"char_appearance" json:"appearance,omitempty"`

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
	logrus.WithFields(logrus.Fields{
		"function":     "Clone",
		"package":      "game",
		"character_id": c.ID,
	}).Debug("entering Clone")

	c.mu.RLock()
	defer c.mu.RUnlock()

	clone := &Character{
		ID:              c.ID,
		Name:            c.Name,
		Description:     c.Description,
		Position:        c.Position,
		Class:           c.Class,
		Strength:        c.Strength,
		Dexterity:       c.Dexterity,
		Constitution:    c.Constitution,
		Intelligence:    c.Intelligence,
		Wisdom:          c.Wisdom,
		Charisma:        c.Charisma,
		HP:              c.HP,
		MaxHP:           c.MaxHP,
		ArmorClass:      c.ArmorClass,
		THAC0:           c.THAC0,
		ActionPoints:    c.ActionPoints,
		MaxActionPoints: c.MaxActionPoints,
		Level:           c.Level,
		Experience:      c.Experience,
		Equipment:       make(map[EquipmentSlot]Item),
		Inventory:       make([]Item, len(c.Inventory)),
		Gold:            c.Gold,
		Appearance:      c.Appearance,
		active:          c.active,
		tags:            make([]string, len(c.tags)),
	}

	// Deep copy equipment map
	for slot, item := range c.Equipment {
		clone.Equipment[slot] = item
	}

	// Deep copy inventory slice
	copy(clone.Inventory, c.Inventory)

	// Deep copy tags slice
	copy(clone.tags, c.tags)

	// Initialize EffectManager for the clone
	clone.ensureEffectManager()

	logrus.WithFields(logrus.Fields{
		"function":     "Clone",
		"package":      "game",
		"character_id": c.ID,
		"clone_id":     clone.ID,
	}).Debug("exiting Clone")

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
	logrus.WithFields(logrus.Fields{
		"function":     "GetHealth",
		"package":      "game",
		"character_id": c.ID,
	}).Debug("entering GetHealth")

	c.mu.RLock()
	defer c.mu.RUnlock()

	health := c.HP

	logrus.WithFields(logrus.Fields{
		"function":     "GetHealth",
		"package":      "game",
		"character_id": c.ID,
		"health":       health,
	}).Debug("exiting GetHealth")

	return health
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
	logrus.WithFields(logrus.Fields{
		"function":     "IsObstacle",
		"package":      "game",
		"character_id": c.ID,
	}).Debug("entering IsObstacle")

	// Characters are considered obstacles for movement/pathing
	result := true

	logrus.WithFields(logrus.Fields{
		"function":     "IsObstacle",
		"package":      "game",
		"character_id": c.ID,
		"is_obstacle":  result,
	}).Debug("exiting IsObstacle")

	return result
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
	c.mu.Lock()
	defer c.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"function":     "SetHealth",
		"package":      "game",
		"character_id": c.ID,
		"old_health":   c.HP,
		"new_health":   health,
	}).Debug("entering SetHealth")

	oldHealth := c.HP
	c.HP = health
	// Ensure health doesn't go below 0
	if c.HP < 0 {
		c.HP = 0
	}
	// Cap health at max health
	if c.HP > c.MaxHP {
		c.HP = c.MaxHP
	}

	logrus.WithFields(logrus.Fields{
		"function":      "SetHealth",
		"package":       "game",
		"character_id":  c.ID,
		"old_health":    oldHealth,
		"final_health":  c.HP,
		"health_capped": c.HP != health,
	}).Debug("exiting SetHealth")
}

// GetActionPoints returns the character's current action points.
// This method is thread-safe.
//
// Returns:
//   - int: The character's current action points
func (c *Character) GetActionPoints() int {
	logrus.WithFields(logrus.Fields{
		"function":     "GetActionPoints",
		"package":      "game",
		"character_id": c.ID,
	}).Debug("entering GetActionPoints")

	c.mu.RLock()
	defer c.mu.RUnlock()

	actionPoints := c.ActionPoints

	logrus.WithFields(logrus.Fields{
		"function":      "GetActionPoints",
		"package":       "game",
		"character_id":  c.ID,
		"action_points": actionPoints,
	}).Debug("exiting GetActionPoints")

	return actionPoints
}

// SetActionPoints sets the character's current action points.
// This method is thread-safe and ensures action points don't exceed MaxActionPoints or go below 0.
//
// Parameters:
//   - actionPoints: The new action points value to set
func (c *Character) SetActionPoints(actionPoints int) {
	logrus.WithFields(logrus.Fields{
		"function":     "SetActionPoints",
		"package":      "game",
		"character_id": c.ID,
		"new_value":    actionPoints,
	}).Debug("entering SetActionPoints")

	c.mu.Lock()
	defer c.mu.Unlock()

	oldActionPoints := c.ActionPoints
	c.ActionPoints = actionPoints
	// Ensure action points don't go below 0
	if c.ActionPoints < 0 {
		logrus.WithFields(logrus.Fields{
			"function":     "SetActionPoints",
			"package":      "game",
			"character_id": c.ID,
			"old_value":    oldActionPoints,
			"new_value":    actionPoints,
		}).Debug("capping action points at 0")
		c.ActionPoints = 0
	}
	// Cap action points at max action points
	if c.ActionPoints > c.MaxActionPoints {
		logrus.WithFields(logrus.Fields{
			"function":     "SetActionPoints",
			"package":      "game",
			"character_id": c.ID,
			"old_value":    oldActionPoints,
			"new_value":    actionPoints,
			"max_value":    c.MaxActionPoints,
		}).Debug("capping action points at maximum")
		c.ActionPoints = c.MaxActionPoints
	}

	logrus.WithFields(logrus.Fields{
		"function":     "SetActionPoints",
		"package":      "game",
		"character_id": c.ID,
		"old_value":    oldActionPoints,
		"final_value":  c.ActionPoints,
		"was_modified": c.ActionPoints != actionPoints,
	}).Debug("exiting SetActionPoints")
}

// GetMaxActionPoints returns the character's maximum action points.
// This method is thread-safe.
//
// Returns:
//   - int: The character's maximum action points
func (c *Character) GetMaxActionPoints() int {
	logrus.WithFields(logrus.Fields{
		"function":     "GetMaxActionPoints",
		"package":      "game",
		"character_id": c.ID,
	}).Debug("entering GetMaxActionPoints")

	c.mu.RLock()
	defer c.mu.RUnlock()

	maxActionPoints := c.MaxActionPoints

	logrus.WithFields(logrus.Fields{
		"function":          "GetMaxActionPoints",
		"package":           "game",
		"character_id":      c.ID,
		"max_action_points": maxActionPoints,
	}).Debug("exiting GetMaxActionPoints")

	return maxActionPoints
}

// ConsumeActionPoints deducts the specified amount from the character's current action points.
// This method is thread-safe and ensures action points don't go below 0.
//
// Parameters:
//   - cost: The amount of action points to consume
//
// Returns:
//   - bool: true if the action points were successfully consumed, false if insufficient
func (c *Character) ConsumeActionPoints(cost int) bool {
	logrus.WithFields(logrus.Fields{
		"function":     "ConsumeActionPoints",
		"package":      "game",
		"character_id": c.ID,
		"cost":         cost,
	}).Debug("entering ConsumeActionPoints")

	c.mu.Lock()
	defer c.mu.Unlock()

	currentActionPoints := c.ActionPoints
	if c.ActionPoints < cost {
		logrus.WithFields(logrus.Fields{
			"function":     "ConsumeActionPoints",
			"package":      "game",
			"character_id": c.ID,
			"cost":         cost,
			"current":      currentActionPoints,
		}).Warn("insufficient action points for consumption")
		return false
	}

	c.ActionPoints -= cost

	logrus.WithFields(logrus.Fields{
		"function":     "ConsumeActionPoints",
		"package":      "game",
		"character_id": c.ID,
		"cost":         cost,
		"old_value":    currentActionPoints,
		"new_value":    c.ActionPoints,
		"consumed":     true,
	}).Debug("exiting ConsumeActionPoints")
	return true
}

// RestoreActionPoints restores the character's action points to their maximum value.
// This is typically called at the start of a new turn.
// This method is thread-safe.
func (c *Character) RestoreActionPoints() {
	logrus.WithFields(logrus.Fields{
		"function":     "RestoreActionPoints",
		"package":      "game",
		"character_id": c.ID,
	}).Debug("entering RestoreActionPoints")

	c.mu.Lock()
	defer c.mu.Unlock()

	oldActionPoints := c.ActionPoints
	c.ActionPoints = c.MaxActionPoints

	logrus.WithFields(logrus.Fields{
		"function":       "RestoreActionPoints",
		"package":        "game",
		"character_id":   c.ID,
		"old_value":      oldActionPoints,
		"restored_value": c.ActionPoints,
	}).Debug("exiting RestoreActionPoints")
}

// Implement GameObject interface methods

// GetID returns the unique identifier string for this Character instance.
// It uses a read lock to safely access the ID field in a concurrent context.
// Returns the character's unique ID string.
// Related: Character struct, ID field
func (c *Character) GetID() string {
	logrus.WithFields(logrus.Fields{
		"function":     "GetID",
		"package":      "game",
		"character_id": c.ID,
	}).Debug("entering GetID")

	c.mu.RLock()
	defer c.mu.RUnlock()

	id := c.ID

	logrus.WithFields(logrus.Fields{
		"function":     "GetID",
		"package":      "game",
		"character_id": id,
	}).Debug("exiting GetID")

	return id
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
	logrus.WithFields(logrus.Fields{
		"function":     "GetName",
		"package":      "game",
		"character_id": c.ID,
	}).Debug("entering GetName")

	c.mu.RLock()
	defer c.mu.RUnlock()

	name := c.Name

	logrus.WithFields(logrus.Fields{
		"function":     "GetName",
		"package":      "game",
		"character_id": c.ID,
		"name":         name,
	}).Debug("exiting GetName")

	return name
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
	logrus.WithFields(logrus.Fields{
		"function":     "GetDescription",
		"package":      "game",
		"character_id": c.ID,
	}).Debug("entering GetDescription")

	c.mu.RLock()
	defer c.mu.RUnlock()

	description := c.Description

	logrus.WithFields(logrus.Fields{
		"function":     "GetDescription",
		"package":      "game",
		"character_id": c.ID,
		"description":  description,
	}).Debug("exiting GetDescription")

	return description
}

// GetPosition returns the current position of the Character.
// This method is thread-safe and uses read locking to protect concurrent access.
// Returns a Position struct containing the character's x,y coordinates.
// Related types:
// - Position struct
func (c *Character) GetPosition() Position {
	logrus.WithFields(logrus.Fields{
		"function":     "GetPosition",
		"package":      "game",
		"character_id": c.ID,
	}).Debug("entering GetPosition")

	c.mu.RLock()
	defer c.mu.RUnlock()

	position := c.Position

	logrus.WithFields(logrus.Fields{
		"function":     "GetPosition",
		"package":      "game",
		"character_id": c.ID,
		"position":     position,
	}).Debug("exiting GetPosition")

	return position
}

// SetPosition updates the character's position to the specified coordinates after validation.
//
// Parameters:
//   - pos Position: The new position coordinates to set
//
// Returns:
//   - error: nil if successful, error if position is invalid
func (c *Character) SetPosition(pos Position) error {
	logrus.WithFields(logrus.Fields{
		"function":     "SetPosition",
		"package":      "game",
		"character_id": c.ID,
		"new_position": pos,
	}).Debug("entering SetPosition")

	c.mu.Lock()
	defer c.mu.Unlock()

	oldPosition := c.Position

	// Validate position before setting
	if !isValidPosition(pos, 100, 100, 10) {
		logrus.WithFields(logrus.Fields{
			"function":     "SetPosition",
			"package":      "game",
			"character_id": c.ID,
			"new_position": pos,
			"old_position": oldPosition,
		}).Error("position validation failed")
		return fmt.Errorf("invalid position: %v", pos)
	}

	c.Position = pos

	logrus.WithFields(logrus.Fields{
		"function":     "SetPosition",
		"package":      "game",
		"character_id": c.ID,
		"old_position": oldPosition,
		"new_position": pos,
	}).Debug("exiting SetPosition")

	return nil
}

// SetPositionWithBounds updates the character's position with map bounds validation.
//
// Parameters:
//   - pos Position: The new position coordinates to set
//   - width, height, maxLevel: map bounds for validation
//
// Returns:
//   - error: nil if successful, error if position is invalid
func (c *Character) SetPositionWithBounds(pos Position, width, height, maxLevel int) error {
	logrus.WithFields(logrus.Fields{
		"function":     "SetPositionWithBounds",
		"package":      "game",
		"character_id": c.ID,
		"new_position": pos,
		"width":        width,
		"height":       height,
		"max_level":    maxLevel,
	}).Debug("entering SetPositionWithBounds")

	c.mu.Lock()
	defer c.mu.Unlock()

	oldPosition := c.Position

	if !isValidPosition(pos, width, height, maxLevel) {
		logrus.WithFields(logrus.Fields{
			"function":     "SetPositionWithBounds",
			"package":      "game",
			"character_id": c.ID,
			"new_position": pos,
			"old_position": oldPosition,
			"width":        width,
			"height":       height,
			"max_level":    maxLevel,
		}).Error("position bounds validation failed")
		return fmt.Errorf("invalid position: %v", pos)
	}

	c.Position = pos

	logrus.WithFields(logrus.Fields{
		"function":     "SetPositionWithBounds",
		"package":      "game",
		"character_id": c.ID,
		"old_position": oldPosition,
		"new_position": pos,
	}).Debug("exiting SetPositionWithBounds")

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
	logrus.WithFields(logrus.Fields{
		"function":     "IsActive",
		"package":      "game",
		"character_id": c.ID,
	}).Debug("entering IsActive")

	c.mu.RLock()
	defer c.mu.RUnlock()

	active := c.active

	logrus.WithFields(logrus.Fields{
		"function":     "IsActive",
		"package":      "game",
		"character_id": c.ID,
		"active":       active,
	}).Debug("exiting IsActive")

	return active
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
	logrus.WithFields(logrus.Fields{
		"function":     "SetActive",
		"package":      "game",
		"character_id": c.ID,
		"new_active":   active,
	}).Debug("entering SetActive")

	c.mu.Lock()
	defer c.mu.Unlock()

	oldActive := c.active
	c.active = active

	logrus.WithFields(logrus.Fields{
		"function":     "SetActive",
		"package":      "game",
		"character_id": c.ID,
		"old_active":   oldActive,
		"new_active":   active,
	}).Debug("exiting SetActive")
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

// Inventory Management Methods

// AddItemToInventory adds an item to the character's inventory with weight and capacity checking.
//
// Parameters:
//   - item: The Item to add to the inventory
//
// Returns:
//   - error: Returns nil on success, or an error if the item cannot be added
//
// Errors:
//   - Returns error if adding the item would exceed carrying capacity
//   - Returns error if item is invalid
//
// Thread safety: This method is thread-safe using mutex locking
func (c *Character) AddItemToInventory(item Item) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Validate item
	if item.ID == "" {
		return NewInventoryError("", "add", ErrEmptyItemID)
	}

	// Check carrying capacity (simplified - could be enhanced with strength-based limits)
	currentWeight := c.calculateTotalWeight()
	maxWeight := c.calculateMaxCarryingCapacity()

	if currentWeight+item.Weight > maxWeight {
		invErr := NewInventoryError(item.ID, "add", ErrCarryingCapacity)
		invErr.CurrentLoad = currentWeight + item.Weight
		invErr.MaxLoad = maxWeight
		return invErr
	}

	// Add item to inventory
	c.Inventory = append(c.Inventory, item)
	return nil
}

// RemoveItemFromInventory removes an item from the character's inventory by ID.
//
// Parameters:
//   - itemID: The unique identifier of the item to remove
//
// Returns:
//   - *Item: Pointer to the removed item, or nil if not found
//   - error: Returns nil on success, or an error if the item cannot be removed
//
// Thread safety: This method is thread-safe using mutex locking
func (c *Character) RemoveItemFromInventory(itemID string) (*Item, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, item := range c.Inventory {
		if item.ID == itemID {
			// Remove item from inventory
			removedItem := item
			c.Inventory = append(c.Inventory[:i], c.Inventory[i+1:]...)
			return &removedItem, nil
		}
	}

	return nil, NewInventoryError(itemID, "remove", ErrItemNotFound)
}

// FindItemInInventory searches for an item in the character's inventory by ID.
//
// Parameters:
//   - itemID: The unique identifier of the item to find
//
// Returns:
//   - *Item: Pointer to the found item, or nil if not found
//   - int: Index of the item in the inventory, or -1 if not found
//
// Thread safety: This method is thread-safe using read mutex locking
func (c *Character) FindItemInInventory(itemID string) (*Item, int) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for i, item := range c.Inventory {
		if item.ID == itemID {
			return &item, i
		}
	}
	return nil, -1
}

// GetInventory returns a copy of the character's inventory.
//
// Returns:
//   - []Item: A slice containing copies of all inventory items
//
// Thread safety: This method is thread-safe using read mutex locking
func (c *Character) GetInventory() []Item {
	c.mu.RLock()
	defer c.mu.RUnlock()

	inventory := make([]Item, len(c.Inventory))
	copy(inventory, c.Inventory)
	return inventory
}

// GetInventoryWeight calculates the total weight of all items in the character's inventory.
//
// Returns:
//   - int: Total weight of all inventory items
//
// Thread safety: This method is thread-safe using read mutex locking
func (c *Character) GetInventoryWeight() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.calculateTotalWeight()
}

// calculateTotalWeight calculates total weight including inventory and equipped items (requires lock)
func (c *Character) calculateTotalWeight() int {
	totalWeight := 0

	// Add inventory weight
	for _, item := range c.Inventory {
		totalWeight += item.Weight
	}

	// Add equipped items weight
	for _, item := range c.Equipment {
		totalWeight += item.Weight
	}

	return totalWeight
}

// calculateMaxCarryingCapacity determines maximum weight this character can carry
func (c *Character) calculateMaxCarryingCapacity() int {
	// Base carrying capacity + strength modifier
	baseCapacity := 50
	strengthBonus := (c.Strength - 10) / 2 * 10 // +10 per strength modifier point
	return baseCapacity + strengthBonus
}

// HasItem checks if the character has a specific item in their inventory.
//
// Parameters:
//   - itemID: The unique identifier of the item to check for
//
// Returns:
//   - bool: true if the item is found in inventory, false otherwise
//
// Thread safety: This method is thread-safe using read mutex locking
func (c *Character) HasItem(itemID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, item := range c.Inventory {
		if item.ID == itemID {
			return true
		}
	}
	return false
}

// CountItems counts how many items of a specific type the character has in inventory.
//
// Parameters:
//   - itemType: The type of items to count (e.g. "weapon", "potion")
//
// Returns:
//   - int: Number of items of the specified type
//
// Thread safety: This method is thread-safe using read mutex locking
func (c *Character) CountItems(itemType string) int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	count := 0
	for _, item := range c.Inventory {
		if item.Type == itemType {
			count++
		}
	}
	return count
}

// TransferItemTo transfers an item from this character's inventory to another character's inventory.
//
// Parameters:
//   - itemID: The unique identifier of the item to transfer
//   - targetCharacter: The character to transfer the item to
//
// Returns:
//   - error: Returns nil on success, or an error if the transfer fails
//
// Thread safety: This method is thread-safe using mutex locking on both characters
func (c *Character) TransferItemTo(itemID string, targetCharacter *Character) error {
	// Lock both characters in consistent order to prevent deadlock
	if c.ID < targetCharacter.ID {
		c.mu.Lock()
		defer c.mu.Unlock()
		targetCharacter.mu.Lock()
		defer targetCharacter.mu.Unlock()
	} else {
		targetCharacter.mu.Lock()
		defer targetCharacter.mu.Unlock()
		c.mu.Lock()
		defer c.mu.Unlock()
	}

	// Find and remove item from source inventory
	var transferItem Item
	itemIndex := -1
	for i, item := range c.Inventory {
		if item.ID == itemID {
			transferItem = item
			itemIndex = i
			break
		}
	}

	if itemIndex == -1 {
		return fmt.Errorf("item not found in source inventory: %s", itemID)
	}

	// Check if target can carry the item
	targetCurrentWeight := targetCharacter.calculateTotalWeight()
	targetMaxWeight := targetCharacter.calculateMaxCarryingCapacity()

	if targetCurrentWeight+transferItem.Weight > targetMaxWeight {
		return fmt.Errorf("target character cannot carry item %s - would exceed capacity", transferItem.Name)
	}

	// Remove from source
	c.Inventory = append(c.Inventory[:itemIndex], c.Inventory[itemIndex+1:]...)

	// Add to target
	targetCharacter.Inventory = append(targetCharacter.Inventory, transferItem)

	return nil
}

// ensureEffectManager initializes the EffectManager if it's nil
// Note: Caller must hold the mutex lock
func (c *Character) ensureEffectManager() {
	if c.EffectManager == nil {
		baseStats := c.toStats()
		c.EffectManager = NewEffectManager(baseStats)
	}
}

// toStats converts the Character's attributes to a Stats struct
func (c *Character) toStats() *Stats {
	return &Stats{
		Health:       float64(c.HP),
		MaxHealth:    float64(c.MaxHP),
		Strength:     float64(c.Strength),
		Dexterity:    float64(c.Dexterity),
		Intelligence: float64(c.Intelligence),
		// Note: Character doesn't have Mana field, so default to 0
		Mana:    0,
		MaxMana: 0,
		Defense: float64(c.ArmorClass),
		Speed:   10, // Default speed value
	}
}

// GetEffectManager returns the character's effect manager, initializing it if necessary
func (c *Character) GetEffectManager() *EffectManager {
	c.mu.RLock()
	if c.EffectManager != nil {
		defer c.mu.RUnlock()
		return c.EffectManager
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	c.ensureEffectManager()
	return c.EffectManager
}

// EffectHolder interface implementation - delegates to EffectManager

// AddEffect applies an effect to this character
func (c *Character) AddEffect(effect *Effect) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ensureEffectManager()
	return c.EffectManager.AddEffect(effect)
}

// RemoveEffect removes an effect from this character
func (c *Character) RemoveEffect(effectID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ensureEffectManager()
	return c.EffectManager.RemoveEffect(effectID)
}

// HasEffect checks if this character has an active effect of the specified type
func (c *Character) HasEffect(effectType EffectType) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.EffectManager == nil {
		return false
	}
	return c.EffectManager.HasEffect(effectType)
}

// GetEffects returns all active effects on this character
func (c *Character) GetEffects() []*Effect {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.EffectManager == nil {
		return []*Effect{}
	}
	return c.EffectManager.GetEffects()
}

// GetImmunities returns a list of all active immunities for this character.
func (c *Character) GetImmunities() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.EffectManager == nil {
		return []string{}
	}
	return c.EffectManager.GetImmunities()
}

// ensureEffectManagerAndGet ensures EffectManager exists and calls a getter function
func (c *Character) ensureEffectManagerAndGet(getter func() *Stats) *Stats {
	c.mu.RLock()
	if c.EffectManager != nil {
		defer c.mu.RUnlock()
		return getter()
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	c.ensureEffectManager()
	return getter()
}

// GetStats returns the current stats (with effects applied)
func (c *Character) GetStats() *Stats {
	return c.ensureEffectManagerAndGet(func() *Stats {
		return c.EffectManager.GetStats()
	})
}

// SetStats updates the current stats
func (c *Character) SetStats(stats *Stats) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ensureEffectManager()
	c.EffectManager.SetStats(stats)
}

// GetBaseStats returns the base stats (without effects)
func (c *Character) GetBaseStats() *Stats {
	return c.ensureEffectManagerAndGet(func() *Stats {
		return c.EffectManager.GetBaseStats()
	})
}

// Experience and Level Progression Methods

// GetLevel returns the character's current level.
//
// Returns:
//   - int: The current level of the character
//
// Thread safety: This method is thread-safe using read mutex locking
func (c *Character) GetLevel() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Level
}

// GetExperience returns the character's current experience points.
//
// Returns:
//   - int64: The current experience points of the character
//
// Thread safety: This method is thread-safe using read mutex locking
func (c *Character) GetExperience() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Experience
}

// AddExperience adds experience points to the character and handles level ups.
//
// Parameters:
//   - xp: The amount of experience points to add
//
// Returns:
//   - bool: true if the character leveled up, false otherwise
//   - error: Returns nil on success, or an error if the operation fails
//
// Thread safety: This method is thread-safe using mutex locking
func (c *Character) AddExperience(xp int64) (bool, error) {
	if xp < 0 {
		return false, NewCharacterError(c.ID, "addExperience", ErrNegativeExperience)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	oldLevel := c.Level
	c.Experience += xp

	// Check for level up
	newLevel := c.calculateLevelFromExperience()
	if newLevel > oldLevel {
		c.Level = newLevel
		// Recalculate THAC0 based on class and new level
		c.THAC0 = calculateTHAC0(c.Class, newLevel)
		// Emit level up event using the existing event system
		if defaultEventSystem != nil {
			emitLevelUpEvent(c.ID, oldLevel, newLevel)
		}
		return true, nil
	}

	return false, nil
}

// SetLevel directly sets the character's level (typically used during character creation).
//
// Parameters:
//   - level: The level to set
//
// Returns:
//   - error: Returns nil on success, or an error if the level is invalid
//
// Thread safety: This method is thread-safe using mutex locking
func (c *Character) SetLevel(level int) error {
	if level < 1 {
		return fmt.Errorf("level must be at least 1: %d", level)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.Level = level
	return nil
}

// SetExperience directly sets the character's experience points.
//
// Parameters:
//   - xp: The experience points to set
//
// Returns:
//   - error: Returns nil on success, or an error if the experience is invalid
//
// Thread safety: This method is thread-safe using mutex locking
func (c *Character) SetExperience(xp int64) error {
	if xp < 0 {
		return fmt.Errorf("experience points cannot be negative: %d", xp)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.Experience = xp
	return nil
}

// GetExperienceToNextLevel returns the experience points needed to reach the next level.
//
// Returns:
//   - int64: Experience points needed for next level, or 0 if at max level
//
// Thread safety: This method is thread-safe using read mutex locking
func (c *Character) GetExperienceToNextLevel() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	nextLevelXP := c.getExperienceRequiredForLevel(c.Level + 1)
	if nextLevelXP == -1 {
		return 0 // Max level reached
	}

	remaining := nextLevelXP - c.Experience
	if remaining < 0 {
		return 0
	}
	return remaining
}

// calculateLevelFromExperience determines the appropriate level for current experience
// Note: Caller must hold the mutex lock
func (c *Character) calculateLevelFromExperience() int {
	level := 1
	for {
		requiredXP := c.getExperienceRequiredForLevel(level + 1)
		if requiredXP == -1 || c.Experience < requiredXP {
			break
		}
		level++
	}
	return level
}

// getExperienceRequiredForLevel returns the total experience needed for a given level
// Returns -1 if level is beyond maximum
func (c *Character) getExperienceRequiredForLevel(level int) int64 {
	if level <= 1 {
		return 0
	}
	if level > 20 { // Max level cap
		return -1
	}

	// Simple experience table - can be enhanced with class-specific tables
	// Uses a standard D&D-style progression: 1000 XP for level 2, then roughly doubles
	switch level {
	case 2:
		return 1000
	case 3:
		return 2000
	case 4:
		return 4000
	case 5:
		return 8000
	case 6:
		return 16000
	case 7:
		return 32000
	case 8:
		return 64000
	case 9:
		return 120000
	case 10:
		return 200000
	default:
		// For levels 11-20, use geometric progression
		baseXP := int64(200000)
		for i := 10; i < level; i++ {
			baseXP = int64(float64(baseXP) * 1.5)
		}
		return baseXP
	}
}
