package game

import (
	"fmt"
	"math/rand"
	"time"
)

// CharacterCreationConfig defines the parameters for creating a new character.
// It contains all the necessary information to generate a valid character
// including class selection, attribute generation method, and starting equipment.
//
// Fields:
//   - Name: The desired name for the character (must be unique and non-empty)
//   - Class: The character class selection from available CharacterClass enum
//   - AttributeMethod: Method for generating attributes ("roll", "pointbuy", "standard")
//   - CustomAttributes: Optional custom attribute values (used with "custom" method)
//   - StartingEquipment: Whether to equip character with class-appropriate gear
//   - StartingGold: Amount of starting gold (0 = use class default)
//
// Related types:
//   - CharacterClass: Enum defining available character classes
//   - Character: The resulting character struct
//   - ClassConfig: Configuration for character classes
type CharacterCreationConfig struct {
	Name              string                 `yaml:"creation_name"`               // Character name
	Class             CharacterClass         `yaml:"creation_class"`              // Character class
	AttributeMethod   string                 `yaml:"creation_attr_method"`        // Attribute generation method
	CustomAttributes  map[string]int         `yaml:"creation_custom_attrs"`       // Custom attribute values
	StartingEquipment bool                   `yaml:"creation_starting_equipment"` // Include starting equipment
	StartingGold      int                    `yaml:"creation_starting_gold"`      // Starting gold amount
	AdditionalData    map[string]interface{} `yaml:"creation_additional_data"`    // Additional character data
}

// CharacterCreationResult represents the outcome of character creation process.
// It contains the created character and any relevant metadata about the creation.
//
// Fields:
//   - Character: Pointer to the newly created Character instance
//   - Success: Boolean indicating if creation was successful
//   - Errors: Slice of error messages encountered during creation
//   - Warnings: Slice of warning messages (non-fatal issues)
//   - CreationTime: Timestamp when the character was created
//   - GeneratedStats: Map of the final generated attribute values
//
// Related types:
//   - Character: The created character instance
//   - CharacterCreationConfig: Input configuration used for creation
type CharacterCreationResult struct {
	Character      *Character     `yaml:"result_character"`       // Created character
	Success        bool           `yaml:"result_success"`         // Creation success status
	Errors         []string       `yaml:"result_errors"`          // Error messages
	Warnings       []string       `yaml:"result_warnings"`        // Warning messages
	CreationTime   time.Time      `yaml:"result_creation_time"`   // When created
	GeneratedStats map[string]int `yaml:"result_generated_stats"` // Final attribute values
	StartingItems  []Item         `yaml:"result_starting_items"`  // Starting equipment
	PlayerData     *Player        `yaml:"result_player_data"`     // Player-specific data if applicable
}

// CharacterCreator handles the creation of new characters with validation and configuration.
// It provides methods for generating characters using different attribute methods
// and ensures all created characters are valid and properly configured.
//
// Fields:
//   - classConfigs: Map of class configurations for validation and equipment
//   - itemDatabase: Map of available items for starting equipment
//   - rng: Random number generator for attribute rolling
//
// Related types:
//   - ClassConfig: Configuration data for character classes
//   - Item: Game items for starting equipment
type CharacterCreator struct {
	classConfigs map[CharacterClass]ClassConfig `yaml:"creator_class_configs"` // Class configuration data
	itemDatabase map[string]Item                `yaml:"creator_item_database"` // Available items
	rng          *rand.Rand                     `yaml:"-"`                     // Random number generator
}

// NewCharacterCreator initializes a new CharacterCreator with default configurations.
// It sets up class configurations, loads item database, and initializes the random number generator.
//
// Returns:
//   - *CharacterCreator: A fully configured character creator instance
//
// The creator is initialized with:
//   - Default class configurations for all 6 classes
//   - Basic starting equipment items
//   - Seeded random number generator
func NewCharacterCreator() *CharacterCreator {
	creator := &CharacterCreator{
		classConfigs: make(map[CharacterClass]ClassConfig),
		itemDatabase: make(map[string]Item),
		rng:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	// Initialize default class configurations
	creator.initializeDefaultClassConfigs()

	// Initialize basic item database
	creator.initializeItemDatabase()

	return creator
}

// CreateCharacter generates a new character based on the provided configuration.
// It validates the configuration, generates attributes, assigns starting equipment,
// and returns a complete character creation result.
//
// Parameters:
//   - config: CharacterCreationConfig containing creation parameters
//
// Returns:
//   - CharacterCreationResult: Complete result with character and metadata
//
// The creation process delegates to specialized methods for each creation step.
func (cc *CharacterCreator) CreateCharacter(config CharacterCreationConfig) CharacterCreationResult {
	result := cc.initializeCreationResult()

	if err := cc.validateCreationInput(config, &result); err != nil {
		return result
	}

	attributes, err := cc.processAttributeGeneration(config, &result)
	if err != nil {
		return result
	}

	character := cc.buildBaseCharacter(config, attributes)
	cc.calculateDerivedStats(character, config.Class)

	cc.applyStartingEquipment(config, character, &result)
	player := cc.createPlayerData(character)

	cc.finalizeCreationResult(character, player, attributes, &result)
	return result
}

// initializeCreationResult creates and returns a new character creation result with default values.
func (cc *CharacterCreator) initializeCreationResult() CharacterCreationResult {
	return CharacterCreationResult{
		Success:        false,
		Errors:         []string{},
		Warnings:       []string{},
		CreationTime:   time.Now(),
		GeneratedStats: make(map[string]int),
		StartingItems:  []Item{},
	}
}

// validateCreationInput validates the configuration and checks class requirements.
func (cc *CharacterCreator) validateCreationInput(config CharacterCreationConfig, result *CharacterCreationResult) error {
	if err := cc.validateConfig(config); err != nil {
		result.Errors = append(result.Errors, err.Error())
		return err
	}
	return nil
}

// processAttributeGeneration generates and validates character attributes.
func (cc *CharacterCreator) processAttributeGeneration(config CharacterCreationConfig, result *CharacterCreationResult) (map[string]int, error) {
	attributes, err := cc.generateAttributes(config)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("attribute generation failed: %v", err))
		return nil, err
	}

	if err := cc.validateClassRequirements(config.Class, attributes); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("class requirements not met: %v", err))
		return nil, err
	}

	return attributes, nil
}

// buildBaseCharacter creates a new Character instance with the specified configuration and attributes.
func (cc *CharacterCreator) buildBaseCharacter(config CharacterCreationConfig, attributes map[string]int) *Character {
	return &Character{
		ID:           NewUID(),
		Name:         config.Name,
		Description:  fmt.Sprintf("A %s %s", config.Class.String(), "adventurer"),
		Position:     Position{X: 0, Y: 0, Level: 0, Facing: DirectionNorth},
		Class:        config.Class,
		Level:        1, // New characters start at level 1
		Strength:     attributes["strength"],
		Dexterity:    attributes["dexterity"],
		Constitution: attributes["constitution"],
		Intelligence: attributes["intelligence"],
		Wisdom:       attributes["wisdom"],
		Charisma:     attributes["charisma"],
		Equipment:    make(map[EquipmentSlot]Item),
		Inventory:    []Item{},
		Gold:         config.StartingGold,
		active:       true,
		tags:         []string{"player_character"},
	}
}

// applyStartingEquipment assigns starting equipment to the character if requested.
func (cc *CharacterCreator) applyStartingEquipment(config CharacterCreationConfig, character *Character, result *CharacterCreationResult) {
	if config.StartingEquipment {
		startingItems := cc.getStartingEquipment(config.Class)
		result.StartingItems = startingItems
		character.Inventory = append(character.Inventory, startingItems...)
	}
}

// createPlayerData creates player-specific data associated with the character.
func (cc *CharacterCreator) createPlayerData(character *Character) *Player {
	return &Player{
		Character:   *character.Clone(),
		Level:       1,
		Experience:  0,
		QuestLog:    []Quest{},
		KnownSpells: []Spell{},
	}
}

// finalizeCreationResult populates the final result with successful creation data.
func (cc *CharacterCreator) finalizeCreationResult(character *Character, player *Player, attributes map[string]int, result *CharacterCreationResult) {
	result.Character = character
	result.PlayerData = player
	result.Success = true
	result.GeneratedStats = attributes
}

// generateAttributes creates character attributes based on the specified method.
// Supports multiple generation methods for different gameplay styles.
//
// Parameters:
//   - config: Character creation configuration
//
// Returns:
//   - map[string]int: Generated attribute values
//   - error: Error if generation fails
//
// Supported methods:
//   - "roll": 4d6 drop lowest for each attribute
//   - "pointbuy": Point-buy system with 27 points
//   - "standard": Standard array (15,14,13,12,10,8)
//   - "custom": Use provided custom values
func (cc *CharacterCreator) generateAttributes(config CharacterCreationConfig) (map[string]int, error) {
	attributes := make(map[string]int)

	switch config.AttributeMethod {
	case "roll":
		attributes["strength"] = cc.rollAttribute()
		attributes["dexterity"] = cc.rollAttribute()
		attributes["constitution"] = cc.rollAttribute()
		attributes["intelligence"] = cc.rollAttribute()
		attributes["wisdom"] = cc.rollAttribute()
		attributes["charisma"] = cc.rollAttribute()

	case "pointbuy":
		return cc.generatePointBuyAttributes()

	case "standard":
		standardArray := []int{15, 14, 13, 12, 10, 8}
		attributeNames := []string{"strength", "dexterity", "constitution", "intelligence", "wisdom", "charisma"}
		for i, name := range attributeNames {
			attributes[name] = standardArray[i]
		}

	case "custom":
		if config.CustomAttributes == nil {
			return nil, fmt.Errorf("custom attributes not provided")
		}
		for key, value := range config.CustomAttributes {
			if value < 3 || value > 18 {
				return nil, fmt.Errorf("attribute %s value %d out of range (3-18)", key, value)
			}
			attributes[key] = value
		}

	default:
		return nil, fmt.Errorf("unknown attribute method: %s", config.AttributeMethod)
	}

	return attributes, nil
}

// rollAttribute generates a single attribute using 4d6 drop lowest method.
// This is the classic D&D attribute generation method.
//
// Returns:
//   - int: Generated attribute value (3-18)
func (cc *CharacterCreator) rollAttribute() int {
	rolls := make([]int, 4)
	for i := 0; i < 4; i++ {
		rolls[i] = cc.rng.Intn(6) + 1
	}

	// Find minimum and remove it
	minValue := rolls[0]
	minIndex := 0
	for i := 1; i < 4; i++ {
		if rolls[i] < minValue {
			minValue = rolls[i]
			minIndex = i
		}
	}

	total := 0
	for i := 0; i < 4; i++ {
		if i != minIndex {
			total += rolls[i]
		}
	}

	return total
}

// generatePointBuyAttributes creates attributes using a point-buy system.
// Starts with base scores of 8 and distributes 27 points.
//
// Returns:
//   - map[string]int: Generated attributes
//   - error: Error if point allocation fails
func (cc *CharacterCreator) generatePointBuyAttributes() (map[string]int, error) {
	attributes := map[string]int{
		"strength":     8,
		"dexterity":    8,
		"constitution": 8,
		"intelligence": 8,
		"wisdom":       8,
		"charisma":     8,
	}

	// Simple random distribution for demo
	remainingPoints := 27
	attributeNames := []string{"strength", "dexterity", "constitution", "intelligence", "wisdom", "charisma"}

	for remainingPoints > 0 && len(attributeNames) > 0 {
		attrIndex := cc.rng.Intn(len(attributeNames))
		attrName := attributeNames[attrIndex]

		if attributes[attrName] < 15 {
			pointCost := 1
			if attributes[attrName] >= 13 {
				pointCost = 2
			}

			if remainingPoints >= pointCost {
				attributes[attrName]++
				remainingPoints -= pointCost
			} else {
				// Remove this attribute from consideration
				attributeNames = append(attributeNames[:attrIndex], attributeNames[attrIndex+1:]...)
			}
		} else {
			// Remove maxed attribute from consideration
			attributeNames = append(attributeNames[:attrIndex], attributeNames[attrIndex+1:]...)
		}
	}

	return attributes, nil
}

// validateConfig checks if the character creation configuration is valid.
//
// Parameters:
//   - config: Configuration to validate
//
// Returns:
//   - error: Validation error if configuration is invalid
func (cc *CharacterCreator) validateConfig(config CharacterCreationConfig) error {
	if config.Name == "" {
		return fmt.Errorf("character name cannot be empty")
	}

	if len(config.Name) > 50 {
		return fmt.Errorf("character name too long (max 50 characters)")
	}

	validMethods := map[string]bool{
		"roll":     true,
		"pointbuy": true,
		"standard": true,
		"custom":   true,
	}

	if !validMethods[config.AttributeMethod] {
		return fmt.Errorf("invalid attribute method: %s", config.AttributeMethod)
	}

	return nil
}

// validateClassRequirements checks if generated attributes meet class requirements.
//
// Parameters:
//   - class: Character class to validate against
//   - attributes: Generated attribute map
//
// Returns:
//   - error: Error if requirements not met
func (cc *CharacterCreator) validateClassRequirements(class CharacterClass, attributes map[string]int) error {
	classConfig, exists := cc.classConfigs[class]
	if !exists {
		return fmt.Errorf("unknown character class: %v", class)
	}

	if attributes["strength"] < classConfig.Requirements.MinStr {
		return fmt.Errorf("insufficient strength for %s (need %d, have %d)",
			class.String(), classConfig.Requirements.MinStr, attributes["strength"])
	}

	if attributes["dexterity"] < classConfig.Requirements.MinDex {
		return fmt.Errorf("insufficient dexterity for %s (need %d, have %d)",
			class.String(), classConfig.Requirements.MinDex, attributes["dexterity"])
	}

	if attributes["constitution"] < classConfig.Requirements.MinCon {
		return fmt.Errorf("insufficient constitution for %s (need %d, have %d)",
			class.String(), classConfig.Requirements.MinCon, attributes["constitution"])
	}

	if attributes["intelligence"] < classConfig.Requirements.MinInt {
		return fmt.Errorf("insufficient intelligence for %s (need %d, have %d)",
			class.String(), classConfig.Requirements.MinInt, attributes["intelligence"])
	}

	if attributes["wisdom"] < classConfig.Requirements.MinWis {
		return fmt.Errorf("insufficient wisdom for %s (need %d, have %d)",
			class.String(), classConfig.Requirements.MinWis, attributes["wisdom"])
	}

	if attributes["charisma"] < classConfig.Requirements.MinCha {
		return fmt.Errorf("insufficient charisma for %s (need %d, have %d)",
			class.String(), classConfig.Requirements.MinCha, attributes["charisma"])
	}

	return nil
}

// calculateDerivedStats computes secondary stats like HP and AC based on class and attributes.
//
// Parameters:
//   - character: Character to calculate stats for
//   - class: Character's class
func (cc *CharacterCreator) calculateDerivedStats(character *Character, class CharacterClass) {
	// Calculate hit points based on class and constitution
	baseHP := map[CharacterClass]int{
		ClassFighter: 10,
		ClassMage:    4,
		ClassCleric:  8,
		ClassThief:   6,
		ClassRanger:  8,
		ClassPaladin: 10,
	}

	conBonus := (character.Constitution - 10) / 2
	character.MaxHP = baseHP[class] + conBonus
	character.HP = character.MaxHP

	// Calculate armor class (base 10 + dex modifier)
	dexBonus := (character.Dexterity - 10) / 2
	character.ArmorClass = 10 + dexBonus

	// Calculate THAC0 (simplified)
	character.THAC0 = 20 // Base for level 1 character

	// Initialize action points based on level and dexterity (level 1 for new characters)
	character.MaxActionPoints = calculateMaxActionPoints(character.Level, character.Dexterity)
	character.ActionPoints = character.MaxActionPoints // Start with full action points
}

// getStartingEquipment returns appropriate starting items for a character class.
//
// Parameters:
//   - class: Character class
//
// Returns:
//   - []Item: List of starting equipment items
func (cc *CharacterCreator) getStartingEquipment(class CharacterClass) []Item {
	equipment := []Item{}

	switch class {
	case ClassFighter:
		equipment = append(equipment, cc.itemDatabase["weapon_shortsword"])
		equipment = append(equipment, cc.itemDatabase["armor_leather"])
	case ClassMage:
		// Mages get minimal equipment
		break
	case ClassCleric:
		equipment = append(equipment, cc.itemDatabase["armor_leather"])
	case ClassThief:
		equipment = append(equipment, cc.itemDatabase["weapon_shortsword"])
	case ClassRanger:
		equipment = append(equipment, cc.itemDatabase["weapon_shortsword"])
		equipment = append(equipment, cc.itemDatabase["armor_leather"])
	case ClassPaladin:
		equipment = append(equipment, cc.itemDatabase["weapon_shortsword"])
		equipment = append(equipment, cc.itemDatabase["armor_leather"])
	}

	return equipment
}

// initializeDefaultClassConfigs sets up default configurations for all character classes.
func (cc *CharacterCreator) initializeDefaultClassConfigs() {
	classDefinitions := cc.createClassDefinitionMap()
	cc.populateClassConfigs(classDefinitions)
}

// createClassDefinitionMap returns a complete map of all character class definitions.
func (cc *CharacterCreator) createClassDefinitionMap() map[CharacterClass]classDefinition {
	return map[CharacterClass]classDefinition{
		ClassFighter: cc.createFighterDefinition(),
		ClassMage:    cc.createMageDefinition(),
		ClassCleric:  cc.createClericDefinition(),
		ClassThief:   cc.createThiefDefinition(),
		ClassRanger:  cc.createRangerDefinition(),
		ClassPaladin: cc.createPaladinDefinition(),
	}
}

// populateClassConfigs converts class definitions to class configs and stores them.
func (cc *CharacterCreator) populateClassConfigs(definitions map[CharacterClass]classDefinition) {
	for classType, definition := range definitions {
		cc.classConfigs[classType] = cc.buildClassConfig(classType, definition)
	}
}

// createFighterDefinition builds the Fighter class definition.
func (cc *CharacterCreator) createFighterDefinition() classDefinition {
	return classDefinition{
		name:         "Fighter",
		description:  "A warrior skilled in combat and tactics",
		hitDice:      "1d10",
		baseSkills:   []string{"Weapon Mastery", "Combat Tactics"},
		abilities:    []string{"Second Wind", "Action Surge"},
		requirements: classRequirements{str: 13},
	}
}

// createMageDefinition builds the Mage class definition.
func (cc *CharacterCreator) createMageDefinition() classDefinition {
	return classDefinition{
		name:         "Mage",
		description:  "A spellcaster who manipulates arcane forces",
		hitDice:      "1d4",
		baseSkills:   []string{"Spellcraft", "Arcane Knowledge"},
		abilities:    []string{"Cantrips", "Spell Casting"},
		requirements: classRequirements{int: 13},
	}
}

// createClericDefinition builds the Cleric class definition.
func (cc *CharacterCreator) createClericDefinition() classDefinition {
	return classDefinition{
		name:         "Cleric",
		description:  "A divine spellcaster and healer",
		hitDice:      "1d8",
		baseSkills:   []string{"Divine Magic", "Healing"},
		abilities:    []string{"Turn Undead", "Divine Casting"},
		requirements: classRequirements{wis: 13},
	}
}

// createThiefDefinition builds the Thief class definition.
func (cc *CharacterCreator) createThiefDefinition() classDefinition {
	return classDefinition{
		name:         "Thief",
		description:  "A stealthy character skilled in subterfuge",
		hitDice:      "1d6",
		baseSkills:   []string{"Stealth", "Lockpicking", "Trap Detection"},
		abilities:    []string{"Sneak Attack", "Thieves Tools"},
		requirements: classRequirements{dex: 13},
	}
}

// createRangerDefinition builds the Ranger class definition.
func (cc *CharacterCreator) createRangerDefinition() classDefinition {
	return classDefinition{
		name:         "Ranger",
		description:  "A wilderness warrior and tracker",
		hitDice:      "1d8",
		baseSkills:   []string{"Tracking", "Survival", "Archery"},
		abilities:    []string{"Favored Enemy", "Natural Magic"},
		requirements: classRequirements{dex: 13, wis: 13},
	}
}

// createPaladinDefinition builds the Paladin class definition.
func (cc *CharacterCreator) createPaladinDefinition() classDefinition {
	return classDefinition{
		name:         "Paladin",
		description:  "A holy warrior dedicated to justice",
		hitDice:      "1d10",
		baseSkills:   []string{"Divine Magic", "Combat", "Leadership"},
		abilities:    []string{"Lay on Hands", "Divine Smite"},
		requirements: classRequirements{str: 13, cha: 13},
	}
}

// buildClassConfig constructs a ClassConfig from a class type and definition.
func (cc *CharacterCreator) buildClassConfig(classType CharacterClass, definition classDefinition) ClassConfig {
	return ClassConfig{
		Type:         classType,
		Name:         definition.name,
		Description:  definition.description,
		HitDice:      definition.hitDice,
		BaseSkills:   definition.baseSkills,
		Abilities:    definition.abilities,
		Requirements: cc.buildClassRequirements(definition.requirements),
	}
}

// buildClassRequirements converts simple requirements to the full requirements struct.
func (cc *CharacterCreator) buildClassRequirements(reqs classRequirements) struct {
	MinStr int `yaml:"min_strength"`
	MinDex int `yaml:"min_dexterity"`
	MinCon int `yaml:"min_constitution"`
	MinInt int `yaml:"min_intelligence"`
	MinWis int `yaml:"min_wisdom"`
	MinCha int `yaml:"min_charisma"`
} {
	return struct {
		MinStr int `yaml:"min_strength"`
		MinDex int `yaml:"min_dexterity"`
		MinCon int `yaml:"min_constitution"`
		MinInt int `yaml:"min_intelligence"`
		MinWis int `yaml:"min_wisdom"`
		MinCha int `yaml:"min_charisma"`
	}{
		MinStr: reqs.str,
		MinDex: reqs.dex,
		MinCon: reqs.con,
		MinInt: reqs.int,
		MinWis: reqs.wis,
		MinCha: reqs.cha,
	}
}

// classDefinition holds the configuration data for a character class.
type classDefinition struct {
	name         string
	description  string
	hitDice      string
	baseSkills   []string
	abilities    []string
	requirements classRequirements
}

// classRequirements holds the attribute requirements for a character class.
type classRequirements struct {
	str int // Strength requirement
	dex int // Dexterity requirement
	con int // Constitution requirement
	int int // Intelligence requirement
	wis int // Wisdom requirement
	cha int // Charisma requirement
}

// initializeItemDatabase sets up the basic item database for starting equipment.
func (cc *CharacterCreator) initializeItemDatabase() {
	cc.itemDatabase["weapon_shortsword"] = Item{
		ID:         "weapon_shortsword",
		Name:       "Short Sword",
		Type:       "weapon",
		Damage:     "1d6",
		Weight:     2,
		Value:      10,
		Properties: []string{"finesse", "light"},
	}

	cc.itemDatabase["armor_leather"] = Item{
		ID:     "armor_leather",
		Name:   "Leather Armor",
		Type:   "armor",
		AC:     11,
		Weight: 10,
		Value:  10,
	}
}
