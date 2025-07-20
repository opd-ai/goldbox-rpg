package game

import (
	"encoding/json"
	"fmt"
	"strings"
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

// GetActionPoints returns the character's current action points.
// This method is thread-safe.
//
// Returns:
//   - int: The character's current action points
func (c *Character) GetActionPoints() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ActionPoints
}

// SetActionPoints sets the character's current action points.
// This method is thread-safe and ensures action points don't exceed MaxActionPoints or go below 0.
//
// Parameters:
//   - actionPoints: The new action points value to set
func (c *Character) SetActionPoints(actionPoints int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ActionPoints = actionPoints
	// Ensure action points don't go below 0
	if c.ActionPoints < 0 {
		c.ActionPoints = 0
	}
	// Cap action points at max action points
	if c.ActionPoints > c.MaxActionPoints {
		c.ActionPoints = c.MaxActionPoints
	}
}

// GetMaxActionPoints returns the character's maximum action points.
// This method is thread-safe.
//
// Returns:
//   - int: The character's maximum action points
func (c *Character) GetMaxActionPoints() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.MaxActionPoints
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
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ActionPoints < cost {
		return false
	}

	c.ActionPoints -= cost
	return true
}

// RestoreActionPoints restores the character's action points to their maximum value.
// This is typically called at the start of a new turn.
// This method is thread-safe.
func (c *Character) RestoreActionPoints() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ActionPoints = c.MaxActionPoints
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
func (c *Character) SetPosition(pos Position) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Validate position before setting
	if !isValidPosition(pos, 100, 100, 10) {
		return fmt.Errorf("invalid position: %v", pos)
	}

	c.Position = pos
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
	c.mu.Lock()
	defer c.mu.Unlock()

	if !isValidPosition(pos, width, height, maxLevel) {
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

	itemIndex, itemToEquip, err := c.findItemInInventory(itemID)
	if err != nil {
		return err
	}

	if err := c.validateItemCanBeEquipped(itemToEquip, slot); err != nil {
		return err
	}

	if err := c.handleSlotConflict(slot); err != nil {
		return err
	}

	c.equipItemToSlot(itemToEquip, slot)
	c.removeItemFromInventoryByIndex(itemIndex)

	return nil
}

// findItemInInventory searches for the item by ID and returns its index and value.
// Returns an error if not found.
func (c *Character) findItemInInventory(itemID string) (int, Item, error) {
	for i, item := range c.Inventory {
		if item.ID == itemID {
			return i, item, nil
		}
	}
	return -1, Item{}, fmt.Errorf("item not found in inventory: %s", itemID)
}

// validateItemCanBeEquipped checks if the item can be equipped in the specified slot.
// Returns an error if validation fails.
func (c *Character) validateItemCanBeEquipped(item Item, slot EquipmentSlot) error {
	if !c.canEquipItemInSlot(item, slot) {
		return fmt.Errorf("item %s cannot be equipped in slot %s", item.Name, slot.String())
	}
	return nil
}

// handleSlotConflict unequips any existing item in the slot, if present.
// Returns an error if unequipping fails.
func (c *Character) handleSlotConflict(slot EquipmentSlot) error {
	if existingItem, exists := c.Equipment[slot]; exists {
		if _, err := c.unequipItemFromSlot(slot); err != nil {
			return fmt.Errorf("failed to unequip existing item %s: %v", existingItem.Name, err)
		}
	}
	return nil
}

// equipItemToSlot assigns the item to the specified equipment slot.
func (c *Character) equipItemToSlot(item Item, slot EquipmentSlot) {
	c.Equipment[slot] = item
}

// removeItemFromInventoryByIndex removes the item at the specified index from inventory.
func (c *Character) removeItemFromInventoryByIndex(index int) {
	c.Inventory = append(c.Inventory[:index], c.Inventory[index+1:]...)
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
	if !c.isItemTypeValidForSlot(item, slot) {
		return false
	}

	proficiencies := GetClassProficiencies(c.Class)

	if c.isWeaponSlot(slot) {
		return c.canEquipWeaponInSlot(item, slot, proficiencies)
	}

	if c.isArmorSlot(slot) {
		return c.canEquipArmorInSlot(item, proficiencies)
	}

	return true
}

// isItemTypeValidForSlot checks if the item type is valid for the specified equipment slot.
// It returns true if the item can be placed in the slot based on type compatibility.
func (c *Character) isItemTypeValidForSlot(item Item, slot EquipmentSlot) bool {
	slotValidTypes := c.getSlotValidTypes()

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

// getSlotValidTypes returns the mapping of equipment slots to their valid item types.
// This defines which item types can be equipped in each slot.
func (c *Character) getSlotValidTypes() map[EquipmentSlot][]string {
	return map[EquipmentSlot][]string{
		SlotHead:       {"helmet", "hat", "crown", "circlet"},
		SlotNeck:       {"amulet", "necklace", "pendant"},
		SlotChest:      {"armor", "robe", "shirt", "vest"},
		SlotHands:      {"gloves", "gauntlets", "bracers"},
		SlotRings:      {"ring"},
		SlotLegs:       {"pants", "leggings", "greaves"},
		SlotFeet:       {"boots", "shoes", "sandals"},
		SlotWeaponMain: {"weapon", "sword", "axe", "staff", "bow", "dagger", "mace", "spear", "hammer", "wand"},
		SlotWeaponOff:  {"shield", "weapon", "dagger", "orb"},
	}
}

// isWeaponSlot checks if the given slot is a weapon slot.
// It returns true for main hand and off-hand weapon slots.
func (c *Character) isWeaponSlot(slot EquipmentSlot) bool {
	return slot == SlotWeaponMain || slot == SlotWeaponOff
}

// isArmorSlot checks if the given slot is an armor slot.
// It returns true for head, chest, hands, legs, and feet slots.
func (c *Character) isArmorSlot(slot EquipmentSlot) bool {
	return slot == SlotHead || slot == SlotChest || slot == SlotHands || slot == SlotLegs || slot == SlotFeet
}

// canEquipWeaponInSlot validates if a character can equip a weapon in the specified slot.
// It checks weapon proficiencies and special shield handling for off-hand slots.
func (c *Character) canEquipWeaponInSlot(item Item, slot EquipmentSlot, proficiencies ClassProficiencies) bool {
	// Special handling for shields in off-hand slot
	if slot == SlotWeaponOff && item.Type == "shield" {
		return proficiencies.ShieldProficient
	}

	// Allow generic "weapon" type if character has any weapon proficiencies
	if item.Type == "weapon" && len(proficiencies.WeaponTypes) > 0 {
		return true
	}

	// Check for specific weapon type match
	for _, allowedWeapon := range proficiencies.WeaponTypes {
		if item.Type == allowedWeapon {
			return true
		}
	}

	return false
}

// canEquipArmorInSlot validates if a character can equip armor in the specified slot.
// It checks armor proficiencies and determines armor type based on item properties.
func (c *Character) canEquipArmorInSlot(item Item, proficiencies ClassProficiencies) bool {
	if !c.isArmorItem(item) {
		return true // Non-armor items don't require armor proficiency
	}

	armorType := determineArmorType(item)

	for _, allowedArmor := range proficiencies.ArmorTypes {
		if armorType == allowedArmor {
			return true
		}
	}
	return false
}

// isArmorItem checks if the item is considered armor that requires proficiency.
// It returns true for items that are classified as armor types.
func (c *Character) isArmorItem(item Item) bool {
	return item.Type == "armor" || item.Type == "helmet" || item.Type == "gauntlets" || item.Type == "greaves"
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
		c.applyPropertyBonuses(item, bonuses)
		c.applyArmorClassBonus(item, bonuses)
	}

	return bonuses
}

// applyPropertyBonuses parses item properties for stat bonuses and updates the bonuses map.
func (c *Character) applyPropertyBonuses(item Item, bonuses map[string]int) {
	for _, property := range item.Properties {
		if len(property) > 2 {
			stat, value, ok := parseStatProperty(property)
			if ok && stat != "" {
				bonuses[stat] += value
			}
		}
	}
}

// applyArmorClassBonus adds AC bonus from armor items to the bonuses map.
func (c *Character) applyArmorClassBonus(item Item, bonuses map[string]int) {
	if item.Type == "armor" && item.AC > 0 {
		bonuses["armor_class"] += item.AC - 10 // Base AC is 10
	}
}

// parseStatProperty parses a property string like "strength+2" or "dexterity-10".
// Returns the stat name, the signed value, and true if parsing succeeded.
func parseStatProperty(property string) (string, int, bool) {
	signPos := -1
	for i := len(property) - 1; i >= 0; i-- {
		if property[i] == '+' || property[i] == '-' {
			signPos = i
			break
		}
	}
	if signPos > 0 && signPos < len(property)-1 {
		stat := property[:signPos]
		sign := 1
		if property[signPos] == '-' {
			sign = -1
		}
		var modifier int
		_, err := fmt.Sscanf(property[signPos+1:], "%d", &modifier)
		if err == nil {
			return stat, sign * modifier, true
		}
	}
	return "", 0, false
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

// determineArmorType determines the armor type (light, medium, heavy) based on item properties
func determineArmorType(item Item) string {
	// Check item properties for armor type indicators
	for _, property := range item.Properties {
		switch property {
		case "light", "light_armor":
			return "light"
		case "medium", "medium_armor":
			return "medium"
		case "heavy", "heavy_armor":
			return "heavy"
		}
	}

	// Default classification based on item type and name
	itemName := strings.ToLower(item.Name)
	switch {
	case strings.Contains(itemName, "leather") || strings.Contains(itemName, "cloth") || strings.Contains(itemName, "robe"):
		return "light"
	case strings.Contains(itemName, "chain") || strings.Contains(itemName, "scale") || strings.Contains(itemName, "studded"):
		return "medium"
	case strings.Contains(itemName, "plate") || strings.Contains(itemName, "full") || strings.Contains(itemName, "heavy"):
		return "heavy"
	default:
		// Default to light for unspecified armor
		return "light"
	}
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

// GetStats returns the current stats (with effects applied)
func (c *Character) GetStats() *Stats {
	c.mu.RLock()
	if c.EffectManager != nil {
		defer c.mu.RUnlock()
		return c.EffectManager.GetStats()
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	c.ensureEffectManager()
	return c.EffectManager.GetStats()
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
	c.mu.RLock()
	if c.EffectManager != nil {
		defer c.mu.RUnlock()
		return c.EffectManager.GetBaseStats()
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	c.ensureEffectManager()
	return c.EffectManager.GetBaseStats()
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
		return false, fmt.Errorf("experience points cannot be negative: %d", xp)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	oldLevel := c.Level
	c.Experience += xp

	// Check for level up
	newLevel := c.calculateLevelFromExperience()
	if newLevel > oldLevel {
		c.Level = newLevel
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
