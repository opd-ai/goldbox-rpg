// Package pcg provides zero-configuration bootstrap capabilities for the GoldBox RPG Engine.
// The bootstrap system automatically generates a complete, playable RPG experience when no
// manual configuration files are present, enabling instant game deployment and testing.
package pcg

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// BootstrapConfig defines parameters for zero-configuration game generation
// These settings control the scope, complexity, and theme of the generated game world
type BootstrapConfig struct {
	// GameLength controls the overall scope and duration of the generated campaign
	GameLength GameLengthType `yaml:"game_length"`

	// ComplexityLevel determines the depth of systems and interconnections
	ComplexityLevel ComplexityType `yaml:"complexity_level"`

	// GenreVariant sets the thematic and mechanical variant for content generation
	GenreVariant GenreType `yaml:"genre_variant"`

	// MaxPlayers sets the expected party size for scaling encounters and rewards
	MaxPlayers int `yaml:"max_players"`

	// StartingLevel determines the initial character level for content balancing
	StartingLevel int `yaml:"starting_level"`

	// WorldSeed provides deterministic generation - 0 means random seed
	WorldSeed int64 `yaml:"world_seed"`

	// EnableQuickStart generates a minimal starting scenario for immediate play
	EnableQuickStart bool `yaml:"enable_quick_start"`

	// DataDirectory specifies where generated configuration files should be saved
	DataDirectory string `yaml:"data_directory"`
}

// GameLengthType defines the scope and duration of generated campaigns
type GameLengthType string

const (
	GameLengthShort  GameLengthType = "short"  // 3-5 hours, single location focus
	GameLengthMedium GameLengthType = "medium" // 8-12 hours, regional scope
	GameLengthLong   GameLengthType = "long"   // 20+ hours, multi-region epic
)

// ComplexityType determines the depth of generated systems and mechanics
type ComplexityType string

const (
	ComplexitySimple   ComplexityType = "simple"   // Basic mechanics, linear progression
	ComplexityStandard ComplexityType = "standard" // Full mechanics, moderate branching
	ComplexityAdvanced ComplexityType = "advanced" // All mechanics, complex interactions
)

// GenreType sets the thematic variant and mechanical emphasis
type GenreType string

const (
	GenreClassicFantasy GenreType = "classic_fantasy" // Standard D&D-style fantasy
	GenreGrimdark       GenreType = "grimdark"        // Dark, low-magic, harsh world
	GenreHighMagic      GenreType = "high_magic"      // Magic-saturated, fantastical
	GenreLowFantasy     GenreType = "low_fantasy"     // Minimal magic, realistic tone
)

// Bootstrap represents the main bootstrap system for zero-configuration game generation
type Bootstrap struct {
	config         *BootstrapConfig
	pcgManager     *PCGManager
	logger         *logrus.Logger
	world          *game.World
	generatedFiles map[string]string // Tracks generated configuration files
}

// NewBootstrap creates a new bootstrap system with the specified configuration
func NewBootstrap(config *BootstrapConfig, world *game.World, logger *logrus.Logger) *Bootstrap {
	if logger == nil {
		logger = logrus.New()
	}

	pcgManager := NewPCGManager(world, logger)

	return &Bootstrap{
		config:         config,
		pcgManager:     pcgManager,
		logger:         logger,
		world:          world,
		generatedFiles: make(map[string]string),
	}
}

// LoadBootstrapTemplate loads a named template from the bootstrap_templates.yaml file
// If the template file doesn't exist or the template name isn't found, returns the default config
func LoadBootstrapTemplate(templateName, dataDir string) (*BootstrapConfig, error) {
	templatesPath := filepath.Join(dataDir, "pcg", "bootstrap_templates.yaml")
	
	// If template file doesn't exist, return default config
	if _, err := os.Stat(templatesPath); os.IsNotExist(err) {
		return DefaultBootstrapConfig(), nil
	}

	// Read template file
	data, err := os.ReadFile(templatesPath)
	if err != nil {
		return DefaultBootstrapConfig(), fmt.Errorf("failed to read templates file: %w", err)
	}

	// Parse templates
	templates := make(map[string]*BootstrapConfig)
	if err := yaml.Unmarshal(data, templates); err != nil {
		return DefaultBootstrapConfig(), fmt.Errorf("failed to parse templates file: %w", err)
	}

	// Get requested template or fall back to default
	if config, exists := templates[templateName]; exists {
		return config, nil
	}

	// Template not found, try "default" template
	if config, exists := templates["default"]; exists {
		return config, nil
	}

	// No templates found, return hardcoded default
	return DefaultBootstrapConfig(), nil
}

// ListAvailableTemplates returns all template names from the bootstrap_templates.yaml file
func ListAvailableTemplates(dataDir string) ([]string, error) {
	templatesPath := filepath.Join(dataDir, "pcg", "bootstrap_templates.yaml")
	
	// If template file doesn't exist, return empty list
	if _, err := os.Stat(templatesPath); os.IsNotExist(err) {
		return []string{}, nil
	}

	// Read template file
	data, err := os.ReadFile(templatesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read templates file: %w", err)
	}

	// Parse templates to get keys
	templates := make(map[string]*BootstrapConfig)
	if err := yaml.Unmarshal(data, templates); err != nil {
		return nil, fmt.Errorf("failed to parse templates file: %w", err)
	}

	// Extract template names
	names := make([]string, 0, len(templates))
	for name := range templates {
		names = append(names, name)
	}

	return names, nil
}

// DetectConfigurationPresence checks if manual configuration files exist
// Returns true if sufficient configuration is present, false if bootstrap is needed
func DetectConfigurationPresence(dataDir string) bool {
	requiredFiles := []string{
		"spells/cantrips.yaml",
		"spells/level1.yaml",
		"items/items.yaml",
	}

	for _, file := range requiredFiles {
		fullPath := filepath.Join(dataDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return false
		}
	}

	return true
}

// DefaultBootstrapConfig returns sensible defaults for zero-configuration games
func DefaultBootstrapConfig() *BootstrapConfig {
	return &BootstrapConfig{
		GameLength:       GameLengthMedium,
		ComplexityLevel:  ComplexityStandard,
		GenreVariant:     GenreClassicFantasy,
		MaxPlayers:       4,
		StartingLevel:    1,
		WorldSeed:        0, // Will use time-based seed
		EnableQuickStart: true,
		DataDirectory:    "data",
	}
}

// GenerateCompleteGame creates a full game configuration from scratch
// This is the main entry point for zero-configuration game generation
func (b *Bootstrap) GenerateCompleteGame(ctx context.Context) (*game.World, error) {
	logrus.WithFields(logrus.Fields{
		"function":       "GenerateCompleteGame",
		"package":        "pcg",
		"game_length":    b.config.GameLength,
		"complexity":     b.config.ComplexityLevel,
		"genre":          b.config.GenreVariant,
		"max_players":    b.config.MaxPlayers,
		"starting_level": b.config.StartingLevel,
	}).Info("entering GenerateCompleteGame")

	b.logger.WithFields(logrus.Fields{
		"game_length":    b.config.GameLength,
		"complexity":     b.config.ComplexityLevel,
		"genre":          b.config.GenreVariant,
		"max_players":    b.config.MaxPlayers,
		"starting_level": b.config.StartingLevel,
	}).Info("Starting zero-configuration game generation")

	startTime := time.Now()

	// Set deterministic seed if specified, otherwise use current time
	worldSeed := b.config.WorldSeed
	if worldSeed == 0 {
		worldSeed = time.Now().UnixNano()
	}
	logrus.WithFields(logrus.Fields{
		"function":   "GenerateCompleteGame",
		"package":    "pcg",
		"world_seed": worldSeed,
	}).Debug("initializing PCG manager with seed")
	b.pcgManager.InitializeWithSeed(worldSeed)

	// Generate core game components with simple placeholder data
	// In a full implementation, these would use the PCG generators
	logrus.WithFields(logrus.Fields{
		"function": "GenerateCompleteGame",
		"package":  "pcg",
	}).Debug("generating simple game content")
	if err := b.generateSimpleGameContent(); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "GenerateCompleteGame",
			"package":  "pcg",
			"error":    err,
		}).Error("failed to generate game content")
		return nil, fmt.Errorf("failed to generate game content: %w", err)
	}

	// Create starting scenario if quick start is enabled
	if b.config.EnableQuickStart {
		logrus.WithFields(logrus.Fields{
			"function": "GenerateCompleteGame",
			"package":  "pcg",
		}).Debug("generating starting scenario")
		if err := b.generateStartingScenario(ctx); err != nil {
			logrus.WithFields(logrus.Fields{
				"function": "GenerateCompleteGame",
				"package":  "pcg",
				"error":    err,
			}).Error("failed to generate starting scenario")
			return nil, fmt.Errorf("failed to generate starting scenario: %w", err)
		}
	}

	// Save generated configuration files
	logrus.WithFields(logrus.Fields{
		"function": "GenerateCompleteGame",
		"package":  "pcg",
	}).Debug("saving generated configuration")
	if err := b.saveGeneratedConfiguration(); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "GenerateCompleteGame",
			"package":  "pcg",
			"error":    err,
		}).Error("failed to save generated configuration")
		return nil, fmt.Errorf("failed to save generated configuration: %w", err)
	}

	duration := time.Since(startTime)
	b.logger.WithFields(logrus.Fields{
		"duration":        duration,
		"files_generated": len(b.generatedFiles),
		"world_seed":      worldSeed,
	}).Info("Zero-configuration game generation completed successfully")

	logrus.WithFields(logrus.Fields{
		"function":        "GenerateCompleteGame",
		"package":         "pcg",
		"duration":        duration,
		"files_generated": len(b.generatedFiles),
		"world_seed":      worldSeed,
	}).Debug("exiting GenerateCompleteGame")

	return b.world, nil
}

// generateSimpleGameContent creates basic game content for immediate play
func (b *Bootstrap) generateSimpleGameContent() error {
	b.logger.Debug("Generating simple game content for immediate play")

	// Generate basic world structure
	worldData := b.createBasicWorld()
	b.storeGeneratedContent("world", worldData)

	// Generate basic faction system
	factionData := b.createBasicFactions()
	b.storeGeneratedContent("factions", factionData)

	// Generate basic NPCs
	characterData := b.createBasicCharacters()
	b.storeGeneratedContent("characters", characterData)

	// Generate basic quests
	questData := b.createBasicQuests()
	b.storeGeneratedContent("quests", questData)

	// Generate basic dialogue
	dialogueData := b.createBasicDialogue()
	b.storeGeneratedContent("dialogue", dialogueData)

	// Note: Spells and items are generated and written to YAML files
	// in the main Run() method for immediate server compatibility
	// but we still track them for testing
	spellData := b.generateBasicSpells()
	b.storeGeneratedContent("spells", spellData)

	itemData := b.generateBasicItems()
	b.storeGeneratedContent("items", itemData)

	b.logger.Debug("Simple game content generation completed")
	return nil
}

// createBasicWorld generates a simple world structure
func (b *Bootstrap) createBasicWorld() interface{} {
	regionCount := b.getRegionCountForLength()

	world := map[string]interface{}{
		"name":        "The Generated Realm",
		"description": "A procedurally generated fantasy world ready for adventure",
		"regions":     regionCount,
		"settlements": regionCount * 2,
		"climate":     "temperate",
		"magic_level": 5,
	}

	return world
}

// createBasicFactions generates a simple faction system
func (b *Bootstrap) createBasicFactions() interface{} {
	factionCount := b.getFactionCountForLength()

	factions := map[string]interface{}{
		"faction_count": factionCount,
		"relationships": "balanced",
		"conflicts":     factionCount / 2,
		"trade_deals":   factionCount,
	}

	return factions
}

// createBasicCharacters generates simple NPC data
func (b *Bootstrap) createBasicCharacters() interface{} {
	npcCount := b.getNPCCountForComplexity()

	characters := map[string]interface{}{
		"npc_count":     npcCount,
		"variety":       "high",
		"personalities": []string{"friendly", "neutral", "hostile", "mysterious"},
		"backgrounds":   []string{"noble", "merchant", "peasant", "soldier", "scholar"},
	}

	return characters
}

// createBasicQuests generates simple quest data
func (b *Bootstrap) createBasicQuests() interface{} {
	questCount := b.getQuestCountForLength()

	quests := map[string]interface{}{
		"quest_count": questCount,
		"types":       []string{"fetch", "escort", "elimination", "discovery"},
		"difficulty":  "scaled",
		"rewards":     "balanced",
	}

	return quests
}

// createBasicDialogue generates simple dialogue data
func (b *Bootstrap) createBasicDialogue() interface{} {
	dialogue := map[string]interface{}{
		"templates":     []string{"greeting", "quest", "shop", "tavern", "combat"},
		"personality":   true,
		"context_aware": b.config.ComplexityLevel != ComplexitySimple,
		"markov_chains": b.config.ComplexityLevel == ComplexityAdvanced,
	}

	return dialogue
}

// generateStartingScenario creates an immediate play scenario
func (b *Bootstrap) generateStartingScenario(ctx context.Context) error {
	b.logger.Debug("Generating quick start scenario")

	scenario := &StartingScenario{
		Title:            "The Adventure Begins",
		Description:      "A perfect starting point for new adventurers seeking glory and gold.",
		StartingLocation: "Crossroads Tavern",
		InitialQuests:    3,
		RecommendedLevel: b.config.StartingLevel,
		MaxPartySize:     b.config.MaxPlayers,
	}

	b.storeGeneratedContent("starting_scenario", scenario)

	b.logger.Debug("Quick start scenario generation completed")

	return nil
}

// Helper methods for parameter calculation based on bootstrap configuration

func (b *Bootstrap) getRegionCountForLength() int {
	switch b.config.GameLength {
	case GameLengthShort:
		return 1
	case GameLengthMedium:
		return 3
	case GameLengthLong:
		return 5
	default:
		return 3
	}
}

func (b *Bootstrap) getFactionCountForLength() int {
	switch b.config.GameLength {
	case GameLengthShort:
		return 2
	case GameLengthMedium:
		return 4
	case GameLengthLong:
		return 6
	default:
		return 4
	}
}

func (b *Bootstrap) getNPCCountForComplexity() int {
	base := 10
	switch b.config.ComplexityLevel {
	case ComplexitySimple:
		return base
	case ComplexityStandard:
		return base * 2
	case ComplexityAdvanced:
		return base * 3
	default:
		return base * 2
	}
}

func (b *Bootstrap) getQuestCountForLength() int {
	switch b.config.GameLength {
	case GameLengthShort:
		return 5
	case GameLengthMedium:
		return 12
	case GameLengthLong:
		return 25
	default:
		return 12
	}
}

// saveGeneratedConfiguration saves all generated content as actual YAML files
func (b *Bootstrap) saveGeneratedConfiguration() error {
	dataDir := b.config.DataDirectory

	// Ensure data directory structure exists
	dirs := []string{
		filepath.Join(dataDir, "spells"),
		filepath.Join(dataDir, "items"),
		filepath.Join(dataDir, "pcg"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Generate and save actual game files
	if err := b.saveSpellFiles(); err != nil {
		return fmt.Errorf("failed to save spell files: %w", err)
	}

	if err := b.saveItemFiles(); err != nil {
		return fmt.Errorf("failed to save item files: %w", err)
	}

	// Save bootstrap configuration itself
	if err := b.saveBootstrapConfig(); err != nil {
		return fmt.Errorf("failed to save bootstrap config: %w", err)
	}

	b.logger.WithFields(logrus.Fields{
		"data_directory": dataDir,
		"files_saved":    len(b.generatedFiles),
	}).Info("Generated configuration files saved successfully")

	return nil
}

// saveSpellFiles creates actual spell YAML files
func (b *Bootstrap) saveSpellFiles() error {
	spellsDir := filepath.Join(b.config.DataDirectory, "spells")

	// Create cantrips.yaml
	cantrips := map[string]interface{}{
		"spells": []map[string]interface{}{
			{
				"spell_id":          "light",
				"spell_name":        "Light",
				"spell_level":       0,
				"spell_school":      5,           // Evocation
				"spell_components":  []int{0, 1}, // Verbal, Somatic
				"spell_range":       5,           // 5 feet for touch
				"spell_duration":    60,          // 1 hour = 60 minutes
				"spell_description": "Creates a bright light that illuminates a 20-foot radius.",
			},
			{
				"spell_id":          "mage_hand",
				"spell_name":        "Mage Hand",
				"spell_level":       0,
				"spell_school":      8,           // Transmutation
				"spell_components":  []int{0, 1}, // Verbal, Somatic
				"spell_range":       30,          // 30 feet
				"spell_duration":    1,           // 1 minute
				"spell_description": "Creates a spectral hand that can manipulate objects at a distance.",
			},
			{
				"spell_id":          "prestidigitation",
				"spell_name":        "Prestidigitation",
				"spell_level":       0,
				"spell_school":      8,           // Transmutation
				"spell_components":  []int{0, 1}, // Verbal, Somatic
				"spell_range":       10,          // 10 feet
				"spell_duration":    60,          // up to 1 hour
				"spell_description": "A simple magical trick that creates minor effects.",
			},
		},
	}

	cantripData, err := yaml.Marshal(cantrips)
	if err != nil {
		return fmt.Errorf("failed to marshal cantrips: %w", err)
	}

	if err := os.WriteFile(filepath.Join(spellsDir, "cantrips.yaml"), cantripData, 0644); err != nil {
		return fmt.Errorf("failed to write cantrips.yaml: %w", err)
	}

	// Create level1.yaml
	level1Spells := map[string]interface{}{
		"spells": []map[string]interface{}{
			{
				"spell_id":          "magic_missile",
				"spell_name":        "Magic Missile",
				"spell_level":       1,
				"spell_school":      5,           // Evocation
				"spell_components":  []int{0, 1}, // Verbal, Somatic
				"spell_range":       120,         // 120 feet
				"spell_duration":    0,           // instantaneous
				"spell_description": "Three darts of magical force strike their target unerringly.",
				"damage_dice":       "1d4+1",
				"damage_type":       "force",
			},
			{
				"spell_id":          "cure_light_wounds",
				"spell_name":        "Cure Light Wounds",
				"spell_level":       1,
				"spell_school":      7,           // Conjuration (Healing)
				"spell_components":  []int{0, 1}, // Verbal, Somatic
				"spell_range":       5,           // touch
				"spell_duration":    0,           // instantaneous
				"spell_description": "Heals minor wounds and injuries.",
				"healing_dice":      "1d8+1",
			},
			{
				"spell_id":          "shield",
				"spell_name":        "Shield",
				"spell_level":       1,
				"spell_school":      1,           // Abjuration
				"spell_components":  []int{0, 1}, // Verbal, Somatic
				"spell_range":       0,           // self
				"spell_duration":    1,           // 1 minute
				"spell_description": "Creates an invisible barrier that protects against attacks.",
			},
		},
	}

	level1Data, err := yaml.Marshal(level1Spells)
	if err != nil {
		return fmt.Errorf("failed to marshal level1 spells: %w", err)
	}

	if err := os.WriteFile(filepath.Join(spellsDir, "level1.yaml"), level1Data, 0644); err != nil {
		return fmt.Errorf("failed to write level1.yaml: %w", err)
	}

	// Create level2.yaml
	level2Spells := map[string]interface{}{
		"spells": []map[string]interface{}{
			{
				"spell_id":          "fireball",
				"spell_name":        "Fireball",
				"spell_level":       2,
				"spell_school":      5,              // Evocation
				"spell_components":  []int{0, 1, 2}, // Verbal, Somatic, Material
				"spell_range":       150,            // 150 feet
				"spell_duration":    0,              // instantaneous
				"spell_description": "A bright streak flashes to a point and blossoms into an explosion of flame.",
				"damage_dice":       "3d6",
				"damage_type":       "fire",
				"area_effect":       true,
			},
			{
				"spell_id":          "cure_moderate_wounds",
				"spell_name":        "Cure Moderate Wounds",
				"spell_level":       2,
				"spell_school":      7,           // Conjuration (Healing)
				"spell_components":  []int{0, 1}, // Verbal, Somatic
				"spell_range":       5,           // touch
				"spell_duration":    0,           // instantaneous
				"spell_description": "Heals moderate wounds and injuries.",
				"healing_dice":      "2d8+2",
			},
			{
				"spell_id":          "invisibility",
				"spell_name":        "Invisibility",
				"spell_level":       2,
				"spell_school":      4,              // Illusion
				"spell_components":  []int{0, 1, 2}, // Verbal, Somatic, Material
				"spell_range":       5,              // touch
				"spell_duration":    60,             // 1 hour
				"spell_description": "Makes a creature invisible until it attacks or casts a spell.",
			},
		},
	}

	level2Data, err := yaml.Marshal(level2Spells)
	if err != nil {
		return fmt.Errorf("failed to marshal level2 spells: %w", err)
	}

	if err := os.WriteFile(filepath.Join(spellsDir, "level2.yaml"), level2Data, 0644); err != nil {
		return fmt.Errorf("failed to write level2.yaml: %w", err)
	}

	return nil
}

// saveItemFiles creates actual item YAML files
func (b *Bootstrap) saveItemFiles() error {
	itemsDir := filepath.Join(b.config.DataDirectory, "items")

	items := []map[string]interface{}{
		{
			"item_id":     "sword",
			"name":        "Sword",
			"type":        "weapon",
			"weapon_type": "melee",
			"damage":      "1d8",
			"weight":      3,
			"value":       15,
			"description": "A well-balanced steel sword.",
		},
		{
			"item_id":     "bow",
			"name":        "Bow",
			"type":        "weapon",
			"weapon_type": "ranged",
			"damage":      "1d6",
			"range":       150,
			"weight":      2,
			"value":       30,
			"description": "A sturdy longbow made of yew wood.",
		},
		{
			"item_id":     "dagger",
			"name":        "Dagger",
			"type":        "weapon",
			"weapon_type": "melee",
			"damage":      "1d4",
			"weight":      1,
			"value":       2,
			"description": "A small, sharp blade suitable for close combat.",
		},
		{
			"item_id":     "staff",
			"name":        "Staff",
			"type":        "weapon",
			"weapon_type": "melee",
			"damage":      "1d6",
			"weight":      4,
			"value":       5,
			"description": "A long wooden staff, useful for walking and combat.",
		},
		{
			"item_id":     "leather_armor",
			"name":        "Leather Armor",
			"type":        "armor",
			"armor_class": 11,
			"weight":      10,
			"value":       45,
			"description": "Flexible leather armor that provides basic protection.",
		},
		{
			"item_id":     "chain_mail",
			"name":        "Chain Mail",
			"type":        "armor",
			"armor_class": 16,
			"weight":      55,
			"value":       750,
			"description": "Heavy armor made of interlocking metal rings.",
		},
		{
			"item_id":     "shield",
			"name":        "Shield",
			"type":        "shield",
			"ac_bonus":    2,
			"weight":      6,
			"value":       10,
			"description": "A sturdy wooden shield reinforced with metal.",
		},
		{
			"item_id":     "healing_potion",
			"name":        "Healing Potion",
			"type":        "consumable",
			"healing":     "2d4+2",
			"weight":      0.5,
			"value":       50,
			"description": "A magical potion that heals wounds when consumed.",
		},
		{
			"item_id":     "rope",
			"name":        "Rope (50 feet)",
			"type":        "equipment",
			"weight":      10,
			"value":       2,
			"description": "Sturdy hemp rope, useful for climbing and binding.",
		},
		{
			"item_id":      "torch",
			"name":         "Torch",
			"type":         "equipment",
			"light_radius": 20,
			"duration":     "1 hour",
			"weight":       1,
			"value":        0.01,
			"description":  "A wooden torch that provides light and can be used as a weapon.",
		},
		{
			"item_id":     "rations",
			"name":        "Trail Rations (1 day)",
			"type":        "equipment",
			"weight":      2,
			"value":       0.5,
			"description": "Dried food suitable for long journeys.",
		},
	}

	itemData, err := yaml.Marshal(items)
	if err != nil {
		return fmt.Errorf("failed to marshal items: %w", err)
	}

	if err := os.WriteFile(filepath.Join(itemsDir, "items.yaml"), itemData, 0644); err != nil {
		return fmt.Errorf("failed to write items.yaml: %w", err)
	}

	return nil
}

func (b *Bootstrap) saveBootstrapConfig() error {
	configPath := filepath.Join(b.config.DataDirectory, "pcg", "bootstrap_config.yaml")

	// Ensure PCG directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create PCG directory: %w", err)
	}

	data, err := yaml.Marshal(b.config)
	if err != nil {
		return fmt.Errorf("failed to marshal bootstrap config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write bootstrap config: %w", err)
	}

	return nil
}

// generateBasicSpells creates a minimal spell set for immediate gameplay
func (b *Bootstrap) generateBasicSpells() interface{} {
	// Return basic spell data structure - this would be expanded in full implementation
	return map[string]interface{}{
		"cantrips": []string{"Light", "Mage Hand", "Prestidigitation"},
		"level1":   []string{"Magic Missile", "Cure Light Wounds", "Shield"},
		"level2":   []string{"Fireball", "Cure Moderate Wounds", "Invisibility"},
	}
}

// generateBasicItems creates a minimal item set for immediate gameplay
func (b *Bootstrap) generateBasicItems() interface{} {
	// Return basic item data structure - this would be expanded in full implementation
	return map[string]interface{}{
		"weapons": []string{"Sword", "Bow", "Dagger", "Staff"},
		"armor":   []string{"Leather Armor", "Chain Mail", "Shield"},
		"items":   []string{"Healing Potion", "Rope", "Torch", "Rations"},
	}
}

// StartingScenario represents a quick-start gameplay scenario
type StartingScenario struct {
	Title            string `yaml:"title"`
	Description      string `yaml:"description"`
	StartingLocation string `yaml:"starting_location"`
	InitialQuests    int    `yaml:"initial_quests"`
	RecommendedLevel int    `yaml:"recommended_level"`
	MaxPartySize     int    `yaml:"max_party_size"`
}

// storeGeneratedContent tracks generated content for testing purposes
func (b *Bootstrap) storeGeneratedContent(contentType string, data interface{}) {
	if b.generatedFiles == nil {
		b.generatedFiles = make(map[string]string)
	}

	// Store a simple marker that this content type was generated
	b.generatedFiles[contentType] = fmt.Sprintf("%T", data)
}
