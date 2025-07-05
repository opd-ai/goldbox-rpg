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
	c.mu.Lock()
	defer c.mu.Unlock()

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

// Equipment Management Methods

// EquipItem equips an item from the character's inventory to the specified equipment slot.
// It validates that the item can be equipped in the slot and handles slot conflicts.
//
// Parameters:
//   - itemID: The unique identifier of the item to equip
//   - slot: The equipment slot where the item should be equipped
//
// Returns:
//   - error: Returns nil on success, or an error describing why the item cannot be equipped
//
// Errors:
//   - Returns error if item is not found in inventory
//   - Returns error if item type is not valid for the specified slot
//   - Returns error if slot is already occupied (unless item is stackable)
//   - Returns error if character doesn't meet requirements for the item
//
// Thread safety: This method is thread-safe using mutex locking
func (c *Character) EquipItem(itemID string, slot EquipmentSlot) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Find the item in inventory
	itemIndex := -1
	var itemToEquip Item
	for i, item := range c.Inventory {
		if item.ID == itemID {
			itemIndex = i
			itemToEquip = item
			break
		}
	}

	if itemIndex == -1 {
		return fmt.Errorf("item not found in inventory: %s", itemID)
	}

	// Validate item can be equipped in the specified slot
	if !c.canEquipItemInSlot(itemToEquip, slot) {
		return fmt.Errorf("item %s cannot be equipped in slot %s", itemToEquip.Name, slot.String())
	}

	// Check if slot is already occupied
	if existingItem, exists := c.Equipment[slot]; exists {
		// Unequip existing item first (this will add it back to inventory)
		if _, err := c.unequipItemFromSlot(slot); err != nil {
			return fmt.Errorf("failed to unequip existing item %s: %v", existingItem.Name, err)
		}
	}

	// Equip the new item
	c.Equipment[slot] = itemToEquip

	// Remove item from inventory
	c.Inventory = append(c.Inventory[:itemIndex], c.Inventory[itemIndex+1:]...)

	return nil
}

// UnequipItem removes an item from the specified equipment slot and adds it to the character's inventory.
//
// Parameters:
//   - slot: The equipment slot to unequip
//
// Returns:
//   - *Item: Pointer to the unequipped item, or nil if slot was empty
//   - error: Returns nil on success, or an error if the operation fails
//
// Thread safety: This method is thread-safe using mutex locking
func (c *Character) UnequipItem(slot EquipmentSlot) (*Item, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.unequipItemFromSlot(slot)
}

// unequipItemFromSlot is the internal implementation of unequipping an item (requires lock to be held)
func (c *Character) unequipItemFromSlot(slot EquipmentSlot) (*Item, error) {
	// Check if there's an item equipped in this slot
	equippedItem, exists := c.Equipment[slot]
	if !exists {
		return nil, fmt.Errorf("no item equipped in slot %s", slot.String())
	}

	// Add the item back to inventory
	c.Inventory = append(c.Inventory, equippedItem)

	// Remove from equipment slot
	delete(c.Equipment, slot)

	return &equippedItem, nil
}

// CanEquipItem checks if the character can equip the specified item in the given slot.
// This performs all validation checks without actually equipping the item.
//
// Parameters:
//   - itemID: The unique identifier of the item to check
//   - slot: The equipment slot to check compatibility with
//
// Returns:
//   - bool: true if the item can be equipped, false otherwise
//   - error: Returns nil if check was successful, or an error if validation fails
//
// Thread safety: This method is thread-safe using read mutex locking
func (c *Character) CanEquipItem(itemID string, slot EquipmentSlot) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Find the item in inventory
	var itemToCheck Item
	found := false
	for _, item := range c.Inventory {
		if item.ID == itemID {
			itemToCheck = item
			found = true
			break
		}
	}

	if !found {
		return false, fmt.Errorf("item not found in inventory: %s", itemID)
	}

	return c.canEquipItemInSlot(itemToCheck, slot), nil
}

// canEquipItemInSlot is the internal validation logic for equipment compatibility
func (c *Character) canEquipItemInSlot(item Item, slot EquipmentSlot) bool {
	// Define valid item types for each slot
	slotValidTypes := map[EquipmentSlot][]string{
		SlotHead:       {"helmet", "hat", "crown", "circlet"},
		SlotNeck:       {"amulet", "necklace", "pendant"},
		SlotChest:      {"armor", "robe", "shirt", "vest"},
		SlotHands:      {"gloves", "gauntlets", "bracers"},
		SlotRings:      {"ring"},
		SlotLegs:       {"pants", "leggings", "greaves"},
		SlotFeet:       {"boots", "shoes", "sandals"},
		SlotWeaponMain: {"weapon", "sword", "axe", "staff", "bow", "dagger"},
		SlotWeaponOff:  {"shield", "weapon", "dagger", "orb"},
	}

	// Check if item type is valid for this slot
	validTypes, exists := slotValidTypes[slot]
	if !exists {
		return false
	}

	for _, validType := range validTypes {
		if item.Type == validType {
			return true
		}
	}

	return false
}

// GetEquippedItem returns the item equipped in the specified slot.
//
// Parameters:
//   - slot: The equipment slot to check
//
// Returns:
//   - *Item: Pointer to the equipped item, or nil if slot is empty
//   - bool: true if an item is equipped in the slot, false otherwise
//
// Thread safety: This method is thread-safe using read mutex locking
func (c *Character) GetEquippedItem(slot EquipmentSlot) (*Item, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if item, exists := c.Equipment[slot]; exists {
		return &item, true
	}
	return nil, false
}

// GetAllEquippedItems returns a copy of all currently equipped items.
//
// Returns:
//   - map[EquipmentSlot]Item: A map containing all equipped items by slot
//
// Thread safety: This method is thread-safe using read mutex locking
func (c *Character) GetAllEquippedItems() map[EquipmentSlot]Item {
	c.mu.RLock()
	defer c.mu.RUnlock()

	equippedItems := make(map[EquipmentSlot]Item)
	for slot, item := range c.Equipment {
		equippedItems[slot] = item
	}
	return equippedItems
}

// GetEquipmentSlots returns all available equipment slots for this character.
//
// Returns:
//   - []EquipmentSlot: Slice containing all valid equipment slot types
func (c *Character) GetEquipmentSlots() []EquipmentSlot {
	return []EquipmentSlot{
		SlotHead, SlotNeck, SlotChest, SlotHands, SlotRings,
		SlotLegs, SlotFeet, SlotWeaponMain, SlotWeaponOff,
	}
}

// CalculateEquipmentBonuses calculates the total stat bonuses from all equipped items.
// This examines item properties for stat modifiers and returns the cumulative effect.
//
// Returns:
//   - map[string]int: Map of stat names to their total bonus values
//
// Thread safety: This method is thread-safe using read mutex locking
func (c *Character) CalculateEquipmentBonuses() map[string]int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	bonuses := make(map[string]int)

	for _, item := range c.Equipment {
		// Parse item properties for stat bonuses
		for _, property := range item.Properties {
			// Handle properties like "strength+2", "dexterity-1", etc.
			if len(property) > 1 {
				var stat string
				var modifier int
				var sign int

				if property[len(property)-2] == '+' {
					stat = property[:len(property)-2]
					sign = 1
					fmt.Sscanf(property[len(property)-1:], "%d", &modifier)
				} else if property[len(property)-2] == '-' {
					stat = property[:len(property)-2]
					sign = -1
					fmt.Sscanf(property[len(property)-1:], "%d", &modifier)
				}

				if stat != "" {
					bonuses[stat] += sign * modifier
				}
			}
		}

		// Handle AC bonus from armor (outside the properties loop)
		if item.Type == "armor" && item.AC > 0 {
			bonuses["armor_class"] += item.AC - 10 // Base AC is 10
		}
	}

	return bonuses
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
		return fmt.Errorf("cannot add item with empty ID")
	}

	// Check carrying capacity (simplified - could be enhanced with strength-based limits)
	currentWeight := c.calculateTotalWeight()
	maxWeight := c.calculateMaxCarryingCapacity()

	if currentWeight+item.Weight > maxWeight {
		return fmt.Errorf("adding item %s would exceed carrying capacity (%d/%d weight)",
			item.Name, currentWeight+item.Weight, maxWeight)
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

	return nil, fmt.Errorf("item not found in inventory: %s", itemID)
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
