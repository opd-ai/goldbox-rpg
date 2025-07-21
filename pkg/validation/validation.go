// Package validation provides comprehensive input validation for JSON-RPC requests
// in the GoldBox RPG Engine. It ensures all user inputs are properly sanitized
// and validated before processing to prevent security vulnerabilities and
// maintain data integrity.
package validation

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// InputValidator provides comprehensive input validation for JSON-RPC methods.
// It maintains a registry of validation functions per method and enforces
// size limits to prevent denial-of-service attacks.
type InputValidator struct {
	maxRequestSize int64
	validators     map[string]func(interface{}) error
}

// NewInputValidator creates a new InputValidator with the specified maximum request size.
// The maxRequestSize parameter limits the size of incoming requests to prevent DoS attacks.
func NewInputValidator(maxRequestSize int64) *InputValidator {
	validator := &InputValidator{
		maxRequestSize: maxRequestSize,
		validators:     make(map[string]func(interface{}) error),
	}

	// Register validators for all JSON-RPC methods
	validator.registerValidators()

	return validator
}

// ValidateRPCRequest validates a JSON-RPC request by checking method existence,
// request size limits, and running method-specific validation rules.
func (v *InputValidator) ValidateRPCRequest(method string, params interface{}, requestSize int64) error {
	// Check request size limit
	if requestSize > v.maxRequestSize {
		return fmt.Errorf("request size %d exceeds maximum allowed size %d", requestSize, v.maxRequestSize)
	}

	// Check if method has a validator
	validator, exists := v.validators[method]
	if !exists {
		return fmt.Errorf("unknown method: %s", method)
	}

	// Run method-specific validation
	return validator(params)
}

// registerValidators sets up validation rules for all JSON-RPC methods.
// Each method gets its own validation function that checks parameter types,
// ranges, and business logic constraints.
func (v *InputValidator) registerValidators() {
	// Game session methods
	v.validators["ping"] = v.validatePing
	v.validators["createPlayer"] = v.validateCreatePlayer
	v.validators["getPlayer"] = v.validateGetPlayer
	v.validators["listPlayers"] = v.validateListPlayers

	// Character management methods
	v.validators["createCharacter"] = v.validateCreateCharacter
	v.validators["getCharacter"] = v.validateGetCharacter
	v.validators["updateCharacter"] = v.validateUpdateCharacter
	v.validators["listCharacters"] = v.validateListCharacters

	// Movement and positioning methods
	v.validators["move"] = v.validateMove
	v.validators["getPosition"] = v.validateGetPosition

	// Combat methods
	v.validators["attack"] = v.validateAttack
	v.validators["castSpell"] = v.validateCastSpell
	v.validators["getSpells"] = v.validateGetSpells

	// World interaction methods
	v.validators["getWorld"] = v.validateGetWorld
	v.validators["getWorldState"] = v.validateGetWorldState

	// Equipment methods
	v.validators["equipItem"] = v.validateEquipItem
	v.validators["unequipItem"] = v.validateUnequipItem
	v.validators["getInventory"] = v.validateGetInventory

	// Additional game methods
	v.validators["useItem"] = v.validateUseItem
	v.validators["leaveGame"] = v.validateLeaveGame
}

// Validation functions for specific JSON-RPC methods

func (v *InputValidator) validatePing(params interface{}) error {
	// Ping accepts no parameters or empty parameters
	return nil
}

func (v *InputValidator) validateCreatePlayer(params interface{}) error {
	paramMap, ok := params.(map[string]interface{})
	if !ok {
		return fmt.Errorf("createPlayer expects object parameters")
	}

	// Validate player name
	name, exists := paramMap["name"]
	if !exists {
		return fmt.Errorf("createPlayer requires 'name' parameter")
	}

	nameStr, ok := name.(string)
	if !ok {
		return fmt.Errorf("player name must be a string")
	}

	return validatePlayerName(nameStr)
}

func (v *InputValidator) validateGetPlayer(params interface{}) error {
	return validateSessionID(params)
}

func (v *InputValidator) validateListPlayers(params interface{}) error {
	return validateSessionID(params)
}

func (v *InputValidator) validateCreateCharacter(params interface{}) error {
	paramMap, ok := params.(map[string]interface{})
	if !ok {
		return fmt.Errorf("createCharacter expects object parameters")
	}

	// Validate session ID
	if err := validateSessionIDFromMap(paramMap); err != nil {
		return err
	}

	// Validate character name
	name, exists := paramMap["name"]
	if !exists {
		return fmt.Errorf("createCharacter requires 'name' parameter")
	}

	nameStr, ok := name.(string)
	if !ok {
		return fmt.Errorf("character name must be a string")
	}

	if err := validateCharacterName(nameStr); err != nil {
		return err
	}

	// Validate character class
	class, exists := paramMap["class"]
	if !exists {
		return fmt.Errorf("createCharacter requires 'class' parameter")
	}

	classStr, ok := class.(string)
	if !ok {
		return fmt.Errorf("character class must be a string")
	}

	return validateCharacterClass(classStr)
}

func (v *InputValidator) validateGetCharacter(params interface{}) error {
	paramMap, ok := params.(map[string]interface{})
	if !ok {
		return fmt.Errorf("getCharacter expects object parameters")
	}

	// Validate session ID
	if err := validateSessionIDFromMap(paramMap); err != nil {
		return err
	}

	// Validate character ID (optional)
	if charID, exists := paramMap["characterId"]; exists {
		charIDStr, ok := charID.(string)
		if !ok {
			return fmt.Errorf("character ID must be a string")
		}
		return validateUUID(charIDStr)
	}

	return nil
}

func (v *InputValidator) validateUpdateCharacter(params interface{}) error {
	paramMap, ok := params.(map[string]interface{})
	if !ok {
		return fmt.Errorf("updateCharacter expects object parameters")
	}

	// Validate session ID and character ID
	if err := validateSessionIDFromMap(paramMap); err != nil {
		return err
	}

	charID, exists := paramMap["characterId"]
	if !exists {
		return fmt.Errorf("updateCharacter requires 'characterId' parameter")
	}

	charIDStr, ok := charID.(string)
	if !ok {
		return fmt.Errorf("character ID must be a string")
	}

	return validateUUID(charIDStr)
}

func (v *InputValidator) validateListCharacters(params interface{}) error {
	return validateSessionID(params)
}

func (v *InputValidator) validateMove(params interface{}) error {
	paramMap, ok := params.(map[string]interface{})
	if !ok {
		return fmt.Errorf("move expects object parameters")
	}

	// Validate session ID
	if err := validateSessionIDFromMap(paramMap); err != nil {
		return err
	}

	// Validate coordinates
	x, xExists := paramMap["x"]
	y, yExists := paramMap["y"]

	if !xExists || !yExists {
		return fmt.Errorf("move requires 'x' and 'y' coordinates")
	}

	// Convert to float64 for validation (JSON numbers)
	xFloat, ok := x.(float64)
	if !ok {
		return fmt.Errorf("x coordinate must be a number")
	}

	yFloat, ok := y.(float64)
	if !ok {
		return fmt.Errorf("y coordinate must be a number")
	}

	// Validate coordinate ranges (assuming reasonable world bounds)
	if xFloat < -10000 || xFloat > 10000 || yFloat < -10000 || yFloat > 10000 {
		return fmt.Errorf("coordinates out of valid range (-10000 to 10000)")
	}

	return nil
}

func (v *InputValidator) validateGetPosition(params interface{}) error {
	return validateSessionID(params)
}

func (v *InputValidator) validateAttack(params interface{}) error {
	paramMap, ok := params.(map[string]interface{})
	if !ok {
		return fmt.Errorf("attack expects object parameters")
	}

	// Validate session ID
	if err := validateSessionIDFromMap(paramMap); err != nil {
		return err
	}

	// Validate target ID
	target, exists := paramMap["targetId"]
	if !exists {
		return fmt.Errorf("attack requires 'targetId' parameter")
	}

	targetStr, ok := target.(string)
	if !ok {
		return fmt.Errorf("target ID must be a string")
	}

	return validateUUID(targetStr)
}

func (v *InputValidator) validateCastSpell(params interface{}) error {
	paramMap, ok := params.(map[string]interface{})
	if !ok {
		return fmt.Errorf("castSpell expects object parameters")
	}

	// Validate session ID
	if err := validateSessionIDFromMap(paramMap); err != nil {
		return err
	}

	// Validate spell ID
	spellID, exists := paramMap["spellId"]
	if !exists {
		return fmt.Errorf("castSpell requires 'spellId' parameter")
	}

	spellIDStr, ok := spellID.(string)
	if !ok {
		return fmt.Errorf("spell ID must be a string")
	}

	return validateSpellID(spellIDStr)
}

func (v *InputValidator) validateGetSpells(params interface{}) error {
	return validateSessionID(params)
}

func (v *InputValidator) validateGetWorld(params interface{}) error {
	return validateSessionID(params)
}

func (v *InputValidator) validateGetWorldState(params interface{}) error {
	return validateSessionID(params)
}

func (v *InputValidator) validateEquipItem(params interface{}) error {
	paramMap, ok := params.(map[string]interface{})
	if !ok {
		return fmt.Errorf("equipItem expects object parameters")
	}

	// Validate session ID
	if err := validateSessionIDFromMap(paramMap); err != nil {
		return err
	}

	// Validate item ID
	itemID, exists := paramMap["itemId"]
	if !exists {
		return fmt.Errorf("equipItem requires 'itemId' parameter")
	}

	itemIDStr, ok := itemID.(string)
	if !ok {
		return fmt.Errorf("item ID must be a string")
	}

	return validateUUID(itemIDStr)
}

func (v *InputValidator) validateUnequipItem(params interface{}) error {
	paramMap, ok := params.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unequipItem expects object parameters")
	}

	// Validate session ID
	if err := validateSessionIDFromMap(paramMap); err != nil {
		return err
	}

	// Validate slot (optional parameter)
	if slot, exists := paramMap["slot"]; exists {
		slotStr, ok := slot.(string)
		if !ok {
			return fmt.Errorf("equipment slot must be a string")
		}
		return validateEquipmentSlot(slotStr)
	}

	return nil
}

func (v *InputValidator) validateGetInventory(params interface{}) error {
	return validateSessionID(params)
}

// Helper validation functions

func validateSessionID(params interface{}) error {
	paramMap, ok := params.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid parameters: expected object")
	}

	return validateSessionIDFromMap(paramMap)
}

func validateSessionIDFromMap(paramMap map[string]interface{}) error {
	sessionID, exists := paramMap["session_id"]
	if !exists {
		return fmt.Errorf("missing required parameter: session_id")
	}

	sessionIDStr, ok := sessionID.(string)
	if !ok {
		return fmt.Errorf("session_id must be a string")
	}

	return validateUUID(sessionIDStr)
}

func validateUUID(id string) error {
	// Basic UUID format validation (8-4-4-4-12 hex digits)
	uuidRegex := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	if !uuidRegex.MatchString(id) {
		return fmt.Errorf("invalid UUID format: %s", id)
	}
	return nil
}

func validatePlayerName(name string) error {
	name = strings.TrimSpace(name)

	if len(name) == 0 {
		return fmt.Errorf("player name cannot be empty")
	}

	if len(name) > 50 {
		return fmt.Errorf("player name cannot exceed 50 characters")
	}

	if !utf8.ValidString(name) {
		return fmt.Errorf("player name contains invalid UTF-8 characters")
	}

	// Check for reasonable character set (letters, numbers, spaces, common punctuation)
	nameRegex := regexp.MustCompile(`^[a-zA-Z0-9\s\-_'\.]+$`)
	if !nameRegex.MatchString(name) {
		return fmt.Errorf("player name contains invalid characters")
	}

	return nil
}

func validateCharacterName(name string) error {
	// Character names have similar rules to player names
	return validatePlayerName(name)
}

func validateCharacterClass(class string) error {
	// Define valid character classes
	validClasses := []string{
		"fighter", "wizard", "cleric", "thief", "ranger", "paladin",
		"magic-user", "elf", "dwarf", "halfling",
	}

	class = strings.ToLower(strings.TrimSpace(class))

	for _, validClass := range validClasses {
		if class == validClass {
			return nil
		}
	}

	return fmt.Errorf("invalid character class: %s", class)
}

func validateSpellID(spellID string) error {
	// Spell IDs should be valid identifiers (lowercase with dashes/underscores)
	spellID = strings.TrimSpace(spellID)

	if len(spellID) == 0 {
		return fmt.Errorf("spell ID cannot be empty")
	}

	if len(spellID) > 100 {
		return fmt.Errorf("spell ID cannot exceed 100 characters")
	}

	spellRegex := regexp.MustCompile(`^[a-z0-9\-_]+$`)
	if !spellRegex.MatchString(spellID) {
		return fmt.Errorf("spell ID contains invalid characters (use lowercase letters, numbers, hyphens, underscores)")
	}

	return nil
}

func validateEquipmentSlot(slot string) error {
	// Define valid equipment slots
	validSlots := []string{
		"head", "neck", "shoulders", "chest", "waist", "legs", "feet",
		"hands", "wrists", "ring1", "ring2", "main-hand", "off-hand",
		"two-hand", "ranged", "ammo",
	}

	slot = strings.ToLower(strings.TrimSpace(slot))

	for _, validSlot := range validSlots {
		if slot == validSlot {
			return nil
		}
	}

	return fmt.Errorf("invalid equipment slot: %s", slot)
}

func (v *InputValidator) validateUseItem(params interface{}) error {
	paramMap, ok := params.(map[string]interface{})
	if !ok {
		return fmt.Errorf("useItem expects object parameters")
	}

	// Validate session ID
	if err := validateSessionIDFromMap(paramMap); err != nil {
		return err
	}

	// Validate item ID
	itemID, exists := paramMap["item_id"]
	if !exists {
		return fmt.Errorf("useItem requires 'item_id' parameter")
	}

	itemIDStr, ok := itemID.(string)
	if !ok {
		return fmt.Errorf("item ID must be a string")
	}

	if strings.TrimSpace(itemIDStr) == "" {
		return fmt.Errorf("item ID cannot be empty")
	}

	// Optional target ID validation
	if target, exists := paramMap["target_id"]; exists {
		targetStr, ok := target.(string)
		if !ok {
			return fmt.Errorf("target ID must be a string")
		}
		if strings.TrimSpace(targetStr) == "" {
			return fmt.Errorf("target ID cannot be empty")
		}
	}

	return nil
}

func (v *InputValidator) validateLeaveGame(params interface{}) error {
	return validateSessionID(params)
}
