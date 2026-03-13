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

// HasValidator returns true if a validator is registered for the given method.
func (v *InputValidator) HasValidator(method string) bool {
	_, exists := v.validators[method]
	return exists
}

// registerValidators sets up validation rules for all JSON-RPC methods.
// Uses a data-driven approach with method groupings to reduce repetition.
func (v *InputValidator) registerValidators() {
	logrus.WithFields(logrus.Fields{
		"function": "registerValidators",
		"package":  "validation",
	}).Debug("entering registerValidators")

	v.registerCoreGameValidators()
	v.registerCombatValidators()
	v.registerQuestValidators()
	v.registerSpellValidators()
	v.registerPCGValidators()
	v.registerEditorValidators()
	v.registerGuildValidators()
	v.registerDiplomacyValidators()
	v.registerAdventureValidators()
}

// registerCoreGameValidators registers validators for game session, character, and movement methods.
func (v *InputValidator) registerCoreGameValidators() {
	// Methods requiring custom validation
	v.validators["ping"] = v.validatePing
	v.validators["createPlayer"] = v.validateCreatePlayer
	v.validators["createCharacter"] = v.validateCreateCharacter
	v.validators["getCharacter"] = v.validateGetCharacter
	v.validators["updateCharacter"] = v.validateUpdateCharacter
	v.validators["move"] = v.validateMove
	v.validators["joinGame"] = v.validateJoinGame
	v.validators["equipItem"] = v.validateEquipItem
	v.validators["unequipItem"] = v.validateUnequipItem
	v.validators["useItem"] = v.validateUseItem

	// Methods requiring only session validation
	sessionOnlyMethods := []string{
		"getPlayer", "listPlayers", "listCharacters", "getPosition",
		"getWorld", "getWorldState", "getInventory", "leaveGame", "getEquipment",
	}
	for _, method := range sessionOnlyMethods {
		v.validators[method] = sessionRequiredValidatorFunc()
	}

	// Methods with optional session
	optionalSessionMethods := []string{"endTurn", "getGameState"}
	for _, method := range optionalSessionMethods {
		v.validators[method] = optionalSessionValidatorFunc()
	}

	// Methods with session and extract validation
	sessionExtractMethods := []string{"applyEffect", "startCombat"}
	for _, method := range sessionExtractMethods {
		v.validators[method] = sessionAndExtractValidatorFunc(method)
	}
}

// registerCombatValidators registers validators for combat-related methods.
func (v *InputValidator) registerCombatValidators() {
	v.validators["attack"] = v.validateAttack
	v.validators["castSpell"] = v.validateCastSpell
	v.validators["getSpells"] = sessionRequiredValidatorFunc()
}

// registerQuestValidators registers validators for quest management methods.
func (v *InputValidator) registerQuestValidators() {
	// Methods with quest ID validation
	questIDMethods := []string{"startQuest", "completeQuest", "failQuest", "getQuest"}
	for _, method := range questIDMethods {
		v.validators[method] = v.validateQuestSessionAndID
	}
	v.validators["updateObjective"] = v.validateUpdateObjective

	// Methods with optional session
	optionalQuestMethods := []string{"getActiveQuests", "getCompletedQuests", "getQuestLog"}
	for _, method := range optionalQuestMethods {
		v.validators[method] = optionalSessionValidatorFunc()
	}
}

// registerSpellValidators registers validators for spell query methods.
func (v *InputValidator) registerSpellValidators() {
	v.validators["getSpell"] = v.validateGetSpell
	v.validators["getSpellsByLevel"] = v.validateGetSpellsByLevel
	v.validators["getSpellsBySchool"] = v.validateGetSpellsBySchool
	v.validators["getAllSpells"] = v.validateNoParams
	v.validators["searchSpells"] = v.validateSearchSpells

	// Spatial query methods
	spatialMethods := []string{"getObjectsInRange", "getObjectsInRadius", "getNearestObjects", "findPath"}
	for _, method := range spatialMethods {
		v.validators[method] = sessionAndExtractValidatorFunc(method)
	}
}

// registerPCGValidators registers validators for procedural content generation methods.
func (v *InputValidator) registerPCGValidators() {
	v.validators["generateContent"] = v.validateGenerateContent
	v.validators["getPCGStats"] = v.validateNoParams

	optionalPCGMethods := []string{"regenerateTerrain", "generateItems", "generateLevel", "generateQuest", "validateContent"}
	for _, method := range optionalPCGMethods {
		v.validators[method] = optionalSessionValidatorFunc()
	}
}

// registerEditorValidators registers validators for map and quest editor methods.
func (v *InputValidator) registerEditorValidators() {
	editorMethods := []string{
		"editor.createMap", "editor.updateTile", "editor.saveMap", "editor.loadMap",
		"questEditor.create", "questEditor.get", "questEditor.update", "questEditor.delete", "questEditor.list",
	}
	for _, method := range editorMethods {
		v.validators[method] = sessionAndExtractValidatorFunc(method)
	}
}

// registerGuildValidators registers validators for guild management methods.
func (v *InputValidator) registerGuildValidators() {
	guildMethods := []string{
		"createGuild", "getGuild", "getCharacterGuild", "joinGuild", "leaveGuild",
		"kickGuildMember", "promoteGuildMember", "demoteGuildMember",
		"guildDeposit", "guildWithdraw", "listGuilds", "transferGuildLeader",
	}
	for _, method := range guildMethods {
		v.validators[method] = sessionAndExtractValidatorFunc(method)
	}
}

// registerDiplomacyValidators registers validators for faction diplomacy methods.
func (v *InputValidator) registerDiplomacyValidators() {
	diplomacyMethods := []string{
		"getFactionRelation", "getFactionRelations", "declareWar", "offerPeace", "acceptPeace",
		"proposeAlliance", "acceptAlliance", "breakAlliance", "signTrade", "sendDiplomaticGift",
	}
	for _, method := range diplomacyMethods {
		v.validators[method] = sessionAndExtractValidatorFunc(method)
	}
}

// registerAdventureValidators registers validators for adventure management methods.
func (v *InputValidator) registerAdventureValidators() {
	v.validators["adventure.list"] = v.validateNoParams
	v.validators["adventure.load"] = v.validateAdventureLoad
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

func (v *InputValidator) validateAttack(params interface{}) error {
	paramMap, err := validateSessionAndExtract(params, "attack")
	if err != nil {
		return err
	}
	// Target IDs can be friendly names (e.g., "enemy1") or UUIDs
	// so we don't enforce UUID format - just require non-empty string
	return validateTargetIDFromMap(paramMap, "attack")
}

// validateTargetIDFromMap extracts and validates a target_id parameter from a map.
// Target IDs can be friendly names or UUIDs, so no UUID format is enforced.
func validateTargetIDFromMap(paramMap map[string]interface{}, methodName string) error {
	targetID, exists := paramMap["target_id"]
	if !exists {
		return fmt.Errorf("%s requires 'target_id' parameter", methodName)
	}

	targetIDStr, ok := targetID.(string)
	if !ok {
		return fmt.Errorf("target ID must be a string")
	}

	if strings.TrimSpace(targetIDStr) == "" {
		return fmt.Errorf("target ID cannot be empty")
	}

	return nil
}

func (v *InputValidator) validateCastSpell(params interface{}) error {
	paramMap, err := validateSessionAndExtract(params, "castSpell")
	if err != nil {
		return err
	}
	return validateRequiredStringParam(paramMap, "spell_id", "castSpell", validateSpellID)
}

func (v *InputValidator) validateEquipItem(params interface{}) error {
	paramMap, err := validateSessionAndExtract(params, "equipItem")
	if err != nil {
		return err
	}
	// Item IDs can be friendly names (e.g., "weapon_shortsword") or UUIDs
	// so we don't enforce UUID format
	return validateItemIDFromMap(paramMap, "equipItem", false)
}

func (v *InputValidator) validateUnequipItem(params interface{}) error {
	paramMap, err := validateSessionAndExtract(params, "unequipItem")
	if err != nil {
		return err
	}
	_, _, err = validateOptionalStringParam(paramMap, "slot", validateEquipmentSlot)
	return err
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
	// Support both hyphen and underscore variants for compatibility
	validSlots := []string{
		"head", "neck", "shoulders", "chest", "waist", "legs", "feet",
		"hands", "wrists", "ring1", "ring2", "rings",
		"main-hand", "off-hand", "main_hand", "off_hand",
		"weapon_main", "weapon_off",
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

// validateJoinGame validates joinGame parameters (player_name required).
func (v *InputValidator) validateJoinGame(params interface{}) error {
	paramMap, err := extractParamMap(params, "joinGame")
	if err != nil {
		return err
	}
	return validateRequiredStringParam(paramMap, "player_name", "joinGame", validatePlayerName)
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

// validateGenerateContent validates generateContent parameters.
func (v *InputValidator) validateGenerateContent(params interface{}) error {
	paramMap, err := extractParamMap(params, "generateContent")
	if err != nil {
		return err
	}
	return validateRequiredStringParam(paramMap, "content_type", "generateContent", nil)
}

// validateAdventureLoad validates adventure.load parameters.
func (v *InputValidator) validateAdventureLoad(params interface{}) error {
	paramMap, err := extractParamMap(params, "adventure.load")
	if err != nil {
		return err
	}
	return validateRequiredStringParam(paramMap, "slug", "adventure.load", validateNonEmpty)
}
