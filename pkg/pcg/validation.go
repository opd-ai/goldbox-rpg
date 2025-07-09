package pcg

import (
	"fmt"
	"strings"

	"goldbox-rpg/pkg/game"
)

// ValidationResult represents the result of content validation
type ValidationResult struct {
	Valid    bool     `yaml:"valid"`
	Errors   []string `yaml:"errors"`
	Warnings []string `yaml:"warnings"`
}

// IsValid returns true if validation passed without errors
func (vr *ValidationResult) IsValid() bool {
	return vr.Valid && len(vr.Errors) == 0
}

// HasWarnings returns true if there are validation warnings
func (vr *ValidationResult) HasWarnings() bool {
	return len(vr.Warnings) > 0
}

// AddError adds an error to the validation result
func (vr *ValidationResult) AddError(message string) {
	vr.Errors = append(vr.Errors, message)
	vr.Valid = false
}

// AddWarning adds a warning to the validation result
func (vr *ValidationResult) AddWarning(message string) {
	vr.Warnings = append(vr.Warnings, message)
}

// Merge combines another validation result into this one
func (vr *ValidationResult) Merge(other *ValidationResult) {
	vr.Errors = append(vr.Errors, other.Errors...)
	vr.Warnings = append(vr.Warnings, other.Warnings...)
	if !other.Valid {
		vr.Valid = false
	}
}

// Validator provides validation for generated content
type Validator struct {
	strictMode bool
}

// NewValidator creates a new content validator
func NewValidator(strictMode bool) *Validator {
	return &Validator{
		strictMode: strictMode,
	}
}

// ValidateGenerationParams validates common generation parameters
func (v *Validator) ValidateGenerationParams(params GenerationParams) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Validate seed (any int64 value is acceptable)

	// Validate difficulty range
	if params.Difficulty < 1 || params.Difficulty > 20 {
		result.AddError(fmt.Sprintf("difficulty must be between 1 and 20, got %d", params.Difficulty))
	}

	// Validate player level
	if params.PlayerLevel < 1 || params.PlayerLevel > 20 {
		result.AddError(fmt.Sprintf("player level must be between 1 and 20, got %d", params.PlayerLevel))
	}

	// Validate timeout
	if params.Timeout <= 0 {
		result.AddWarning("timeout not specified or invalid, generation may run indefinitely")
	}

	return result
}

// ValidateTerrainParams validates terrain-specific parameters
func (v *Validator) ValidateTerrainParams(params TerrainParams) *ValidationResult {
	result := v.ValidateGenerationParams(params.GenerationParams)

	// Validate biome type
	validBiomes := []BiomeType{
		BiomeForest, BiomeMountain, BiomeDesert, BiomeSwamp,
		BiomeCave, BiomeDungeon, BiomeCoastal, BiomeUrban, BiomeWasteland,
	}

	valid := false
	for _, validBiome := range validBiomes {
		if params.BiomeType == validBiome {
			valid = true
			break
		}
	}
	if !valid {
		result.AddError(fmt.Sprintf("invalid biome type: %s", params.BiomeType))
	}

	// Validate density
	if params.Density < 0.0 || params.Density > 1.0 {
		result.AddError(fmt.Sprintf("density must be between 0.0 and 1.0, got %f", params.Density))
	}

	// Validate water level
	if params.WaterLevel < 0.0 || params.WaterLevel > 1.0 {
		result.AddError(fmt.Sprintf("water level must be between 0.0 and 1.0, got %f", params.WaterLevel))
	}

	// Validate roughness
	if params.Roughness < 0.0 || params.Roughness > 1.0 {
		result.AddError(fmt.Sprintf("roughness must be between 0.0 and 1.0, got %f", params.Roughness))
	}

	return result
}

// ValidateItemParams validates item-specific parameters
func (v *Validator) ValidateItemParams(params ItemParams) *ValidationResult {
	result := v.ValidateGenerationParams(params.GenerationParams)

	// Validate rarity tiers
	validRarities := []RarityTier{
		RarityCommon, RarityUncommon, RarityRare,
		RarityEpic, RarityLegendary, RarityArtifact,
	}

	// Check minimum rarity
	minValid := false
	for _, validRarity := range validRarities {
		if params.MinRarity == validRarity {
			minValid = true
			break
		}
	}
	if !minValid {
		result.AddError(fmt.Sprintf("invalid minimum rarity: %s", params.MinRarity))
	}

	// Check maximum rarity
	maxValid := false
	for _, validRarity := range validRarities {
		if params.MaxRarity == validRarity {
			maxValid = true
			break
		}
	}
	if !maxValid {
		result.AddError(fmt.Sprintf("invalid maximum rarity: %s", params.MaxRarity))
	}

	// Validate enchantment rate
	if params.EnchantmentRate < 0.0 || params.EnchantmentRate > 1.0 {
		result.AddError(fmt.Sprintf("enchantment rate must be between 0.0 and 1.0, got %f", params.EnchantmentRate))
	}

	// Validate unique chance
	if params.UniqueChance < 0.0 || params.UniqueChance > 1.0 {
		result.AddError(fmt.Sprintf("unique chance must be between 0.0 and 1.0, got %f", params.UniqueChance))
	}

	return result
}

// ValidateLevelParams validates level-specific parameters
func (v *Validator) ValidateLevelParams(params LevelParams) *ValidationResult {
	result := v.ValidateGenerationParams(params.GenerationParams)

	// Validate room counts
	if params.MinRooms < 1 {
		result.AddError("minimum rooms must be at least 1")
	}

	if params.MaxRooms < params.MinRooms {
		result.AddError("maximum rooms must be greater than or equal to minimum rooms")
	}

	if params.MaxRooms > 100 {
		result.AddWarning("maximum rooms is very high, generation may be slow")
	}

	// Validate secret rooms
	if params.SecretRooms < 0 {
		result.AddError("secret rooms cannot be negative")
	}

	if params.SecretRooms > params.MaxRooms/2 {
		result.AddWarning("high number of secret rooms relative to total rooms")
	}

	return result
}

// ValidateGameMap validates a generated game map
func (v *Validator) ValidateGameMap(gameMap *game.GameMap) *ValidationResult {
	result := &ValidationResult{Valid: true}

	if gameMap == nil {
		result.AddError("game map is nil")
		return result
	}

	// Validate dimensions
	if gameMap.Width <= 0 || gameMap.Height <= 0 {
		result.AddError(fmt.Sprintf("invalid map dimensions: %dx%d", gameMap.Width, gameMap.Height))
	}

	// Validate tiles array
	if len(gameMap.Tiles) != gameMap.Height {
		result.AddError(fmt.Sprintf("tiles array height mismatch: expected %d, got %d", gameMap.Height, len(gameMap.Tiles)))
	}

	for y, row := range gameMap.Tiles {
		if len(row) != gameMap.Width {
			result.AddError(fmt.Sprintf("tiles array width mismatch at row %d: expected %d, got %d", y, gameMap.Width, len(row)))
		}
	}

	// Check for walkable path connectivity if in strict mode
	if v.strictMode {
		if !v.validateMapConnectivity(gameMap) {
			result.AddError("map lacks proper connectivity between walkable areas")
		}
	}

	return result
}

// ValidateItem validates a generated item
func (v *Validator) ValidateItem(item *game.Item) *ValidationResult {
	result := &ValidationResult{Valid: true}

	if item == nil {
		result.AddError("item is nil")
		return result
	}

	// Validate required fields
	if strings.TrimSpace(item.ID) == "" {
		result.AddError("item ID cannot be empty")
	}

	if strings.TrimSpace(item.Name) == "" {
		result.AddError("item name cannot be empty")
	}

	if strings.TrimSpace(item.Type) == "" {
		result.AddError("item type cannot be empty")
	}

	// Validate value
	if item.Value < 0 {
		result.AddError("item value cannot be negative")
	}

	// Validate weight
	if item.Weight < 0 {
		result.AddError("item weight cannot be negative")
	}

	// Validate armor class for armor items
	if item.Type == "armor" && item.AC <= 0 {
		result.AddWarning("armor item has zero or negative AC")
	}

	// Validate damage for weapon items
	if item.Type == "weapon" && strings.TrimSpace(item.Damage) == "" {
		result.AddWarning("weapon item has no damage specification")
	}

	return result
}

// ValidateLevel validates a generated level
func (v *Validator) ValidateLevel(level *game.Level) *ValidationResult {
	result := &ValidationResult{Valid: true}

	if level == nil {
		result.AddError("level is nil")
		return result
	}

	// Validate required fields
	if strings.TrimSpace(level.ID) == "" {
		result.AddError("level ID cannot be empty")
	}

	if strings.TrimSpace(level.Name) == "" {
		result.AddError("level name cannot be empty")
	}

	// Validate dimensions
	if level.Width <= 0 || level.Height <= 0 {
		result.AddError(fmt.Sprintf("invalid level dimensions: %dx%d", level.Width, level.Height))
	}

	// Validate tiles array
	if len(level.Tiles) != level.Height {
		result.AddError(fmt.Sprintf("tiles array height mismatch: expected %d, got %d", level.Height, len(level.Tiles)))
	}

	for y, row := range level.Tiles {
		if len(row) != level.Width {
			result.AddError(fmt.Sprintf("tiles array width mismatch at row %d: expected %d, got %d", y, level.Width, len(row)))
		}
	}

	return result
}

// validateMapConnectivity checks if walkable areas in the map are properly connected
func (v *Validator) validateMapConnectivity(gameMap *game.GameMap) bool {
	if gameMap.Width == 0 || gameMap.Height == 0 {
		return false
	}

	// Find all walkable tiles
	walkableTiles := make([]game.Position, 0)
	for y := 0; y < gameMap.Height; y++ {
		for x := 0; x < gameMap.Width; x++ {
			if gameMap.Tiles[y][x].Walkable {
				walkableTiles = append(walkableTiles, game.Position{X: x, Y: y})
			}
		}
	}

	if len(walkableTiles) == 0 {
		return false // No walkable areas
	}

	// Use flood fill to check connectivity from the first walkable tile
	visited := make(map[game.Position]bool)
	queue := []game.Position{walkableTiles[0]}
	visited[walkableTiles[0]] = true
	reachableCount := 1

	directions := []game.Position{
		{X: 0, Y: -1}, // North
		{X: 1, Y: 0},  // East
		{X: 0, Y: 1},  // South
		{X: -1, Y: 0}, // West
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// Check all adjacent tiles
		for _, dir := range directions {
			next := game.Position{
				X: current.X + dir.X,
				Y: current.Y + dir.Y,
			}

			// Check bounds
			if next.X < 0 || next.X >= gameMap.Width || next.Y < 0 || next.Y >= gameMap.Height {
				continue
			}

			// Check if already visited
			if visited[next] {
				continue
			}

			// Check if walkable
			if gameMap.Tiles[next.Y][next.X].Walkable {
				visited[next] = true
				queue = append(queue, next)
				reachableCount++
			}
		}
	}

	// All walkable tiles should be reachable
	return reachableCount == len(walkableTiles)
}
