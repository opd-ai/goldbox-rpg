package pcg

import (
	"context"
	"fmt"
	"time"

	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
)

// PCGManager is the main coordinator for procedural content generation
// Integrates with existing game systems and manages the generation lifecycle
type PCGManager struct {
	registry    *Registry
	factory     *Factory
	validator   *Validator
	logger      *logrus.Logger
	world       *game.World
	seedManager *SeedManager
}

// NewPCGManager creates a new PCG manager instance
func NewPCGManager(world *game.World, logger *logrus.Logger) *PCGManager {
	if logger == nil {
		logger = logrus.New()
	}

	registry := NewRegistry(logger)
	factory := NewFactory(registry, logger)
	validator := NewValidator(false)
	seedManager := NewSeedManager(0) // Will be set by game initialization

	return &PCGManager{
		registry:    registry,
		factory:     factory,
		validator:   validator,
		logger:      logger,
		world:       world,
		seedManager: seedManager,
	}
}

// InitializeWithSeed sets the base seed for all generation
func (pcg *PCGManager) InitializeWithSeed(seed int64) {
	pcg.seedManager = NewSeedManager(seed)
	pcg.logger.WithField("seed", seed).Info("PCG manager initialized with seed")
}

// RegisterDefaultGenerators registers the built-in generators
func (pcg *PCGManager) RegisterDefaultGenerators() error {
	// This would be called during game initialization to register all available generators
	// Implementation would register generators from each sub-package

	pcg.logger.Info("Registering default PCG generators")

	// Example registration - in real implementation, these would come from their packages
	// terrainGen := terrain.NewCellularAutomataGenerator()
	// if err := pcg.registry.RegisterGenerator("cellular_automata", terrainGen); err != nil {
	//     return fmt.Errorf("failed to register cellular automata generator: %w", err)
	// }

	return nil
}

// GenerateTerrainForLevel generates terrain for a specific game level
func (pcg *PCGManager) GenerateTerrainForLevel(ctx context.Context, levelID string, width, height int, biome BiomeType, difficulty int) (*game.GameMap, error) {
	params := TerrainParams{
		GenerationParams: GenerationParams{
			Seed:        pcg.seedManager.DeriveContextSeed(ContentTypeTerrain, levelID),
			Difficulty:  difficulty,
			PlayerLevel: 1, // Could be derived from world state
			WorldState:  pcg.world,
			Timeout:     30 * time.Second,
			Constraints: make(map[string]interface{}),
		},
		BiomeType:    biome,
		Density:      0.45,
		Connectivity: ConnectivityModerate,
		WaterLevel:   0.1,
		Roughness:    0.5,
	}

	// Add terrain-specific constraints
	params.Constraints["width"] = width
	params.Constraints["height"] = height
	params.Constraints["terrain_params"] = params

	return pcg.factory.GenerateTerrain(ctx, "cellular_automata", params)
}

// GenerateItemsForLocation generates items appropriate for a specific location
func (pcg *PCGManager) GenerateItemsForLocation(ctx context.Context, locationID string, itemCount int, minRarity, maxRarity RarityTier, playerLevel int) ([]*game.Item, error) {
	params := ItemParams{
		GenerationParams: GenerationParams{
			Seed:        pcg.seedManager.DeriveContextSeed(ContentTypeItems, locationID),
			Difficulty:  pcg.calculateLocationDifficulty(locationID),
			PlayerLevel: playerLevel,
			WorldState:  pcg.world,
			Timeout:     10 * time.Second,
			Constraints: make(map[string]interface{}),
		},
		MinRarity:       minRarity,
		MaxRarity:       maxRarity,
		EnchantmentRate: 0.2,
		UniqueChance:    0.05,
		LevelScaling:    true,
	}

	// Add item count constraint
	params.Constraints["item_count"] = itemCount

	return pcg.factory.GenerateItems(ctx, "template_based", params)
}

// GenerateDungeonLevel generates a complete dungeon level
func (pcg *PCGManager) GenerateDungeonLevel(ctx context.Context, levelID string, minRooms, maxRooms int, theme LevelTheme, difficulty int) (*game.Level, error) {
	params := LevelParams{
		GenerationParams: GenerationParams{
			Seed:        pcg.seedManager.DeriveContextSeed(ContentTypeLevels, levelID),
			Difficulty:  difficulty,
			PlayerLevel: pcg.getAveragePartyLevel(),
			WorldState:  pcg.world,
			Timeout:     60 * time.Second,
			Constraints: make(map[string]interface{}),
		},
		MinRooms:      minRooms,
		MaxRooms:      maxRooms,
		RoomTypes:     []RoomType{RoomTypeEntrance, RoomTypeExit, RoomTypeCombat, RoomTypeTreasure},
		CorridorStyle: CorridorWindy,
		LevelTheme:    theme,
		HasBoss:       difficulty >= 10,
		SecretRooms:   maxRooms / 10,
	}

	return pcg.factory.GenerateLevel(ctx, "room_corridor", params)
}

// GenerateQuestForArea generates a quest appropriate for a specific area
func (pcg *PCGManager) GenerateQuestForArea(ctx context.Context, areaID string, questType QuestType, playerLevel int) (*game.Quest, error) {
	params := QuestParams{
		GenerationParams: GenerationParams{
			Seed:        pcg.seedManager.DeriveContextSeed(ContentTypeQuests, areaID),
			Difficulty:  pcg.calculateAreaDifficulty(areaID),
			PlayerLevel: playerLevel,
			WorldState:  pcg.world,
			Timeout:     15 * time.Second,
			Constraints: make(map[string]interface{}),
		},
		QuestType:     questType,
		MinObjectives: 1,
		MaxObjectives: 3,
		RewardTier:    RarityRare,
		Narrative:     NarrativeLinear,
	}

	return pcg.factory.GenerateQuest(ctx, "objective_based", params)
}

// ValidateGeneratedContent validates content before integration into the world
func (pcg *PCGManager) ValidateGeneratedContent(content interface{}) (*ValidationResult, error) {
	switch v := content.(type) {
	case *game.GameMap:
		return pcg.validator.ValidateGameMap(v), nil
	case *game.Item:
		return pcg.validator.ValidateItem(v), nil
	case *game.Level:
		return pcg.validator.ValidateLevel(v), nil
	default:
		return nil, fmt.Errorf("unsupported content type for validation: %T", content)
	}
}

// IntegrateContentIntoWorld integrates generated content into the game world
func (pcg *PCGManager) IntegrateContentIntoWorld(content interface{}, locationID string) error {
	// Validate content before integration
	validationResult, err := pcg.ValidateGeneratedContent(content)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if !validationResult.IsValid() {
		return fmt.Errorf("content validation failed: %v", validationResult.Errors)
	}

	// Log warnings if present
	if validationResult.HasWarnings() {
		pcg.logger.WithFields(logrus.Fields{
			"location": locationID,
			"warnings": validationResult.Warnings,
		}).Warn("Generated content has validation warnings")
	}

	// Integrate based on content type
	switch v := content.(type) {
	case *game.Level:
		return pcg.integrateLevelIntoWorld(v, locationID)
	case *game.Item:
		return pcg.integrateItemIntoWorld(v, locationID)
	case []*game.Item:
		for _, item := range v {
			if err := pcg.integrateItemIntoWorld(item, locationID); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported content type for integration: %T", content)
	}
}

// RegenerateContentForLocation regenerates content for a specific location
func (pcg *PCGManager) RegenerateContentForLocation(ctx context.Context, locationID string, contentType ContentType) (interface{}, error) {
	pcg.logger.WithFields(logrus.Fields{
		"location":     locationID,
		"content_type": contentType,
	}).Info("Regenerating content for location")

	// Get current world state for context
	difficulty := pcg.calculateLocationDifficulty(locationID)
	playerLevel := pcg.getAveragePartyLevel()

	switch contentType {
	case ContentTypeTerrain:
		return pcg.GenerateTerrainForLevel(ctx, locationID, 50, 50, BiomeDungeon, difficulty)
	case ContentTypeItems:
		return pcg.GenerateItemsForLocation(ctx, locationID, 3, RarityCommon, RarityRare, playerLevel)
	case ContentTypeLevels:
		return pcg.GenerateDungeonLevel(ctx, locationID, 5, 15, ThemeClassic, difficulty)
	case ContentTypeQuests:
		return pcg.GenerateQuestForArea(ctx, locationID, QuestTypeFetch, playerLevel)
	default:
		return nil, fmt.Errorf("unsupported content type for regeneration: %s", contentType)
	}
}

// GetGenerationStatistics returns statistics about generation activity
func (pcg *PCGManager) GetGenerationStatistics() map[string]interface{} {
	stats := make(map[string]interface{})

	// Get available generators
	stats["available_generators"] = pcg.registry.ListAllGenerators()

	// Get seed information
	stats["base_seed"] = pcg.seedManager.GetBaseSeed()

	// Could include generation counters, performance metrics, etc.

	return stats
}

// Helper methods for integration

func (pcg *PCGManager) integrateLevelIntoWorld(level *game.Level, locationID string) error {
	// Add level to world - World should provide thread-safe methods for this
	// For now, we'll use a direct approach assuming World has proper synchronization
	pcg.world.Levels = append(pcg.world.Levels, *level)

	pcg.logger.WithFields(logrus.Fields{
		"level_id": level.ID,
		"location": locationID,
		"width":    level.Width,
		"height":   level.Height,
	}).Info("Integrated generated level into world")

	return nil
}

func (pcg *PCGManager) integrateItemIntoWorld(item *game.Item, locationID string) error {
	// Add item to world objects - World should provide thread-safe methods for this
	if pcg.world.Objects == nil {
		pcg.world.Objects = make(map[string]game.GameObject)
	}

	pcg.world.Objects[item.ID] = item

	// Update spatial index if available
	if pcg.world.SpatialIndex != nil {
		if err := pcg.world.SpatialIndex.Insert(item); err != nil {
			pcg.logger.WithFields(logrus.Fields{
				"item_id": item.ID,
				"error":   err.Error(),
			}).Warn("Failed to add item to spatial index")
		}
	}

	pcg.logger.WithFields(logrus.Fields{
		"item_id":  item.ID,
		"location": locationID,
		"type":     item.Type,
		"value":    item.Value,
	}).Info("Integrated generated item into world")

	return nil
}

// Helper methods for world state analysis

func (pcg *PCGManager) calculateLocationDifficulty(locationID string) int {
	// Analyze world state to determine appropriate difficulty
	// This would examine factors like:
	// - Player party levels
	// - Location depth/progression
	// - Existing challenges in the area
	// - World difficulty curve

	// Simplified implementation
	return 5 // Default moderate difficulty
}

func (pcg *PCGManager) calculateAreaDifficulty(areaID string) int {
	// Similar to location difficulty but for larger areas
	return pcg.calculateLocationDifficulty(areaID)
}

func (pcg *PCGManager) getAveragePartyLevel() int {
	if pcg.world == nil {
		return 1
	}

	// Note: In a real implementation, World should provide thread-safe accessors
	if len(pcg.world.Players) == 0 {
		return 1
	}

	totalLevel := 0
	count := 0

	for _, player := range pcg.world.Players {
		// Note: Character is a struct, not a pointer, so we check Level directly
		if player.Character.Level > 0 {
			totalLevel += player.Character.Level
			count++
		}
	}

	if count == 0 {
		return 1
	}

	return totalLevel / count
}
