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

	"github.com/sirupsen/logrus"
)

// Note: logrus configuration (SetReportCaller, log level, etc.) should be
// done at the application level, not in library packages, to avoid
// affecting the entire process when the package is imported.

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
	logrus.WithFields(logrus.Fields{
		"function":         "NewInputValidator",
		"package":          "validation",
		"max_request_size": maxRequestSize,
	}).Debug("entering NewInputValidator")

	validator := &InputValidator{
		maxRequestSize: maxRequestSize,
		validators:     make(map[string]func(interface{}) error),
	}

	// Register validators for all JSON-RPC methods
	validator.registerValidators()

	logrus.WithFields(logrus.Fields{
		"function":         "NewInputValidator",
		"package":          "validation",
		"max_request_size": maxRequestSize,
		"validators_count": len(validator.validators),
	}).Debug("exiting NewInputValidator")

	return validator
}

// ValidateRPCRequest validates a JSON-RPC request by checking method existence,
// request size limits, and running method-specific validation rules.
func (v *InputValidator) ValidateRPCRequest(method string, params interface{}, requestSize int64) error {
	logrus.WithFields(logrus.Fields{
		"function":     "ValidateRPCRequest",
		"package":      "validation",
		"method":       method,
		"request_size": requestSize,
	}).Debug("entering ValidateRPCRequest")

	// Check request size limit
	if requestSize > v.maxRequestSize {
		logrus.WithFields(logrus.Fields{
			"function":     "ValidateRPCRequest",
			"package":      "validation",
			"method":       method,
			"request_size": requestSize,
			"max_size":     v.maxRequestSize,
		}).Error("request size exceeds maximum allowed")
		return fmt.Errorf("request size %d exceeds maximum allowed size %d", requestSize, v.maxRequestSize)
	}

	// Check if method has a validator
	validator, exists := v.validators[method]
	if !exists {
		logrus.WithFields(logrus.Fields{
			"function":       "ValidateRPCRequest",
			"package":        "validation",
			"unknown_method": method,
		}).Error("unknown method")
		return fmt.Errorf("unknown method: %s", method)
	}

	// Run method-specific validation
	err := validator(params)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "ValidateRPCRequest",
			"package":  "validation",
			"method":   method,
			"error":    err,
		}).Error("method validation failed")
	} else {
		logrus.WithFields(logrus.Fields{
			"function": "ValidateRPCRequest",
			"package":  "validation",
			"method":   method,
		}).Debug("exiting ValidateRPCRequest - validation successful")
	}
	return err
}

// registerValidators sets up validation rules for all JSON-RPC methods.
// Each method gets its own validation function that checks parameter types,
// ranges, and business logic constraints.
func (v *InputValidator) registerValidators() {
	logrus.WithFields(logrus.Fields{
		"function": "registerValidators",
		"package":  "validation",
	}).Debug("entering registerValidators")

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

	// Game session methods (server-side)
	v.validators["joinGame"] = v.validateJoinGame
	v.validators["applyEffect"] = v.validateApplyEffect
	v.validators["startCombat"] = v.validateStartCombat
	v.validators["endTurn"] = v.validateEndTurn
	v.validators["getGameState"] = v.validateGetGameState
	v.validators["getEquipment"] = v.validateGetEquipment

	// Quest management methods
	v.validators["startQuest"] = v.validateQuestSessionAndID
	v.validators["completeQuest"] = v.validateQuestSessionAndID
	v.validators["updateObjective"] = v.validateUpdateObjective
	v.validators["failQuest"] = v.validateQuestSessionAndID
	v.validators["getQuest"] = v.validateQuestSessionAndID
	v.validators["getActiveQuests"] = v.validateSessionOnly
	v.validators["getCompletedQuests"] = v.validateSessionOnly
	v.validators["getQuestLog"] = v.validateSessionOnly

	// Spell query methods
	v.validators["getSpell"] = v.validateGetSpell
	v.validators["getSpellsByLevel"] = v.validateGetSpellsByLevel
	v.validators["getSpellsBySchool"] = v.validateGetSpellsBySchool
	v.validators["getAllSpells"] = v.validateNoParams
	v.validators["searchSpells"] = v.validateSearchSpells
	v.validators["getSpells"] = v.validateGetSpells

	// Spatial query methods
	v.validators["getObjectsInRange"] = v.validateSpatialRange
	v.validators["getObjectsInRadius"] = v.validateSpatialRadius
	v.validators["getNearestObjects"] = v.validateNearestObjects

	// PCG methods
	v.validators["generateContent"] = v.validateGenerateContent
	v.validators["regenerateTerrain"] = v.validateSessionOnly
	v.validators["generateItems"] = v.validateSessionOnly
	v.validators["generateLevel"] = v.validateSessionOnly
	v.validators["generateQuest"] = v.validateSessionOnly
	v.validators["getPCGStats"] = v.validateNoParams
	v.validators["validateContent"] = v.validateSessionOnly
}

// Validation functions for specific JSON-RPC methods

func (v *InputValidator) validatePing(params interface{}) error {
	// Ping accepts no parameters or empty parameters
	return nil
}

func (v *InputValidator) validateCreatePlayer(params interface{}) error {
	paramMap, err := extractParamMap(params, "createPlayer")
	if err != nil {
		return err
	}
	return validateRequiredStringParam(paramMap, "name", "createPlayer", validatePlayerName)
}

func (v *InputValidator) validateGetPlayer(params interface{}) error {
	return validateSessionID(params)
}

func (v *InputValidator) validateListPlayers(params interface{}) error {
	return validateSessionID(params)
}

func (v *InputValidator) validateCreateCharacter(params interface{}) error {
	paramMap, err := extractParamMap(params, "createCharacter")
	if err != nil {
		return err
	}
	// Note: session_id is optional for createCharacter since the handler creates a new session
	if err := validateRequiredStringParam(paramMap, "name", "createCharacter", validateCharacterName); err != nil {
		return err
	}
	return validateRequiredStringParam(paramMap, "class", "createCharacter", validateCharacterClass)
}

func (v *InputValidator) validateGetCharacter(params interface{}) error {
	paramMap, err := validateSessionAndExtract(params, "getCharacter")
	if err != nil {
		return err
	}
	_, _, err = validateOptionalStringParam(paramMap, "characterId", validateUUID)
	return err
}

func (v *InputValidator) validateUpdateCharacter(params interface{}) error {
	paramMap, err := validateSessionAndExtract(params, "updateCharacter")
	if err != nil {
		return err
	}
	return validateRequiredStringParam(paramMap, "characterId", "updateCharacter", validateUUID)
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

	// The move handler accepts a "direction" parameter (integer Direction enum).
	// It can also accept "x"/"y" coordinates for absolute positioning.
	_, dirExists := paramMap["direction"]
	x, xExists := paramMap["x"]
	y, yExists := paramMap["y"]

	if !dirExists && (!xExists || !yExists) {
		return fmt.Errorf("move requires 'direction' or 'x' and 'y' coordinates")
	}

	// Validate coordinate ranges if present
	if xExists && yExists {
		xFloat, ok := x.(float64)
		if !ok {
			return fmt.Errorf("x coordinate must be a number")
		}

		yFloat, ok := y.(float64)
		if !ok {
			return fmt.Errorf("y coordinate must be a number")
		}

		if xFloat < -10000 || xFloat > 10000 || yFloat < -10000 || yFloat > 10000 {
			return fmt.Errorf("coordinates out of valid range (-10000 to 10000)")
		}
	}

	return nil
}

func (v *InputValidator) validateGetPosition(params interface{}) error {
	return validateSessionID(params)
}

func (v *InputValidator) validateAttack(params interface{}) error {
	paramMap, err := validateSessionAndExtract(params, "attack")
	if err != nil {
		return err
	}
	return validateRequiredStringParam(paramMap, "target_id", "attack", validateUUID)
}

func (v *InputValidator) validateCastSpell(params interface{}) error {
	paramMap, err := validateSessionAndExtract(params, "castSpell")
	if err != nil {
		return err
	}
	return validateRequiredStringParam(paramMap, "spell_id", "castSpell", validateSpellID)
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
	paramMap, err := validateSessionAndExtract(params, "equipItem")
	if err != nil {
		return err
	}
	return validateItemIDFromMap(paramMap, "equipItem", true)
}

func (v *InputValidator) validateUnequipItem(params interface{}) error {
	paramMap, err := validateSessionAndExtract(params, "unequipItem")
	if err != nil {
		return err
	}
	_, _, err = validateOptionalStringParam(paramMap, "slot", validateEquipmentSlot)
	return err
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

// validateItemIDFromMap extracts and validates an item_id parameter from a map.
// If requireUUID is true, validates the item_id as a UUID format.
func validateItemIDFromMap(paramMap map[string]interface{}, methodName string, requireUUID bool) error {
	itemID, exists := paramMap["item_id"]
	if !exists {
		return fmt.Errorf("%s requires 'item_id' parameter", methodName)
	}

	itemIDStr, ok := itemID.(string)
	if !ok {
		return fmt.Errorf("item ID must be a string")
	}

	if strings.TrimSpace(itemIDStr) == "" {
		return fmt.Errorf("item ID cannot be empty")
	}

	// Validate UUID format if required
	if requireUUID {
		return validateUUID(itemIDStr)
	}

	return nil
}

func validateCharacterClass(class string) error {
	// Define valid character classes - must match game.CharacterClass constants
	// See pkg/game/constants.go: ClassFighter, ClassMage, ClassCleric, ClassThief, ClassRanger, ClassPaladin
	validClasses := []string{
		"fighter", "mage", "cleric", "thief", "ranger", "paladin",
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
	paramMap, err := validateSessionAndExtract(params, "useItem")
	if err != nil {
		return err
	}
	if err := validateItemIDFromMap(paramMap, "useItem", false); err != nil {
		return err
	}
	_, _, err = validateOptionalStringParam(paramMap, "target_id", validateNonEmpty)
	return err
}

func (v *InputValidator) validateLeaveGame(params interface{}) error {
	return validateSessionID(params)
}

// validateJoinGame validates joinGame parameters (player_name required).
func (v *InputValidator) validateJoinGame(params interface{}) error {
	paramMap, err := extractParamMap(params, "joinGame")
	if err != nil {
		return err
	}
	return validateRequiredStringParam(paramMap, "player_name", "joinGame", validatePlayerName)
}

// validateApplyEffect validates applyEffect parameters.
func (v *InputValidator) validateApplyEffect(params interface{}) error {
	_, err := validateSessionAndExtract(params, "applyEffect")
	return err
}

// validateStartCombat validates startCombat parameters.
func (v *InputValidator) validateStartCombat(params interface{}) error {
	_, err := validateSessionAndExtract(params, "startCombat")
	return err
}

// validateEndTurn validates endTurn parameters.
func (v *InputValidator) validateEndTurn(params interface{}) error {
	if params == nil {
		return nil
	}
	paramMap, ok := params.(map[string]interface{})
	if !ok {
		return nil // endTurn can accept no params
	}
	if _, exists := paramMap["session_id"]; exists {
		return validateSessionIDFromMap(paramMap)
	}
	return nil
}

// validateGetGameState validates getGameState parameters.
func (v *InputValidator) validateGetGameState(params interface{}) error {
	if params == nil {
		return nil
	}
	paramMap, ok := params.(map[string]interface{})
	if !ok {
		return nil // getGameState can accept no params
	}
	if _, exists := paramMap["session_id"]; exists {
		return validateSessionIDFromMap(paramMap)
	}
	return nil
}

// validateGetEquipment validates getEquipment parameters.
func (v *InputValidator) validateGetEquipment(params interface{}) error {
	return validateSessionID(params)
}

// validateSessionOnly validates that only a session_id is present.
func (v *InputValidator) validateSessionOnly(params interface{}) error {
	if params == nil {
		return nil
	}
	paramMap, ok := params.(map[string]interface{})
	if !ok {
		return nil
	}
	if _, exists := paramMap["session_id"]; exists {
		return validateSessionIDFromMap(paramMap)
	}
	return nil
}

// validateNoParams validates that no parameters are required.
func (v *InputValidator) validateNoParams(params interface{}) error {
	return nil
}

// validateQuestSessionAndID validates quest methods that need session_id and quest_id.
func (v *InputValidator) validateQuestSessionAndID(params interface{}) error {
	paramMap, err := validateSessionAndExtract(params, "quest method")
	if err != nil {
		return err
	}
	return validateRequiredStringParam(paramMap, "quest_id", "quest method", nil)
}

// validateUpdateObjective validates updateObjective parameters.
func (v *InputValidator) validateUpdateObjective(params interface{}) error {
	paramMap, err := validateSessionAndExtract(params, "updateObjective")
	if err != nil {
		return err
	}
	if err := validateRequiredStringParam(paramMap, "quest_id", "updateObjective", nil); err != nil {
		return err
	}
	return validateRequiredStringParam(paramMap, "objective_id", "updateObjective", nil)
}

// validateGetSpell validates getSpell parameters.
func (v *InputValidator) validateGetSpell(params interface{}) error {
	paramMap, err := extractParamMap(params, "getSpell")
	if err != nil {
		return err
	}
	return validateRequiredStringParam(paramMap, "spell_id", "getSpell", validateSpellID)
}

// validateGetSpellsByLevel validates getSpellsByLevel parameters.
func (v *InputValidator) validateGetSpellsByLevel(params interface{}) error {
	paramMap, err := extractParamMap(params, "getSpellsByLevel")
	if err != nil {
		return err
	}
	_, err = validateRequiredNumericParam(paramMap, "level", "getSpellsByLevel", 0, 20)
	return err
}

// validateGetSpellsBySchool validates getSpellsBySchool parameters.
func (v *InputValidator) validateGetSpellsBySchool(params interface{}) error {
	paramMap, err := extractParamMap(params, "getSpellsBySchool")
	if err != nil {
		return err
	}
	return validateRequiredStringParam(paramMap, "school", "getSpellsBySchool", nil)
}

// validateSearchSpells validates searchSpells parameters.
func (v *InputValidator) validateSearchSpells(params interface{}) error {
	paramMap, err := extractParamMap(params, "searchSpells")
	if err != nil {
		return err
	}
	return validateRequiredStringParam(paramMap, "query", "searchSpells", validateNonEmpty)
}

// validateSpatialRange validates getObjectsInRange parameters.
func (v *InputValidator) validateSpatialRange(params interface{}) error {
	_, err := validateSessionAndExtract(params, "getObjectsInRange")
	return err
}

// validateSpatialRadius validates getObjectsInRadius parameters.
func (v *InputValidator) validateSpatialRadius(params interface{}) error {
	_, err := validateSessionAndExtract(params, "getObjectsInRadius")
	return err
}

// validateNearestObjects validates getNearestObjects parameters.
func (v *InputValidator) validateNearestObjects(params interface{}) error {
	_, err := validateSessionAndExtract(params, "getNearestObjects")
	return err
}

// validateGenerateContent validates generateContent parameters.
func (v *InputValidator) validateGenerateContent(params interface{}) error {
	paramMap, err := extractParamMap(params, "generateContent")
	if err != nil {
		return err
	}
	return validateRequiredStringParam(paramMap, "content_type", "generateContent", nil)
}
