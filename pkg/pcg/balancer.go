package pcg

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
)

// ContentBalancer provides difficulty scaling and balance management for procedural content
// Ensures generated content maintains appropriate challenge levels and resource distribution
type ContentBalancer struct {
	mu      sync.RWMutex
	version string
	logger  *logrus.Logger
	rng     *rand.Rand

	// Balance configuration
	config         BalanceConfig
	powerCurves    map[ContentType]PowerCurve
	scalingRules   map[ContentType]ScalingRule
	resourceLimits map[string]ResourceLimit

	// Runtime state
	metrics *BalanceMetrics
}

// BalanceConfig defines global balance parameters
type BalanceConfig struct {
	BasePlayerLevel       int     `yaml:"base_player_level"`       // Reference player level for scaling
	MaxDifficultyLevel    int     `yaml:"max_difficulty_level"`    // Maximum difficulty rating
	PowerCurveExponent    float64 `yaml:"power_curve_exponent"`    // Exponential scaling factor
	ResourceScarcityRate  float64 `yaml:"resource_scarcity_rate"`  // Rate of resource availability decay
	RewardInflationRate   float64 `yaml:"reward_inflation_rate"`   // Rate at which rewards scale with level
	VariabilityRange      float64 `yaml:"variability_range"`       // Random variation range (0.0-1.0)
	BalanceToleranceMin   float64 `yaml:"balance_tolerance_min"`   // Minimum acceptable balance ratio
	BalanceToleranceMax   float64 `yaml:"balance_tolerance_max"`   // Maximum acceptable balance ratio
	CriticalFailureThresh float64 `yaml:"critical_failure_thresh"` // Threshold for critical balance failures
}

// PowerCurve defines how content power scales with level
type PowerCurve struct {
	ContentType     ContentType `yaml:"content_type"`
	BaseValue       float64     `yaml:"base_value"`       // Starting power level
	ScalingFactor   float64     `yaml:"scaling_factor"`   // Linear scaling component
	ExponentFactor  float64     `yaml:"exponent_factor"`  // Exponential scaling component
	CapValue        float64     `yaml:"cap_value"`        // Maximum power level
	VarianceRange   float64     `yaml:"variance_range"`   // Random variance (0.0-1.0)
	BreakpointLevel int         `yaml:"breakpoint_level"` // Level where scaling changes
}

// ScalingRule defines content-specific scaling behavior
type ScalingRule struct {
	ContentType        ContentType        `yaml:"content_type"`
	ScalingType        ScalingType        `yaml:"scaling_type"`
	DifficultyFactor   float64            `yaml:"difficulty_factor"`
	RewardMultiplier   float64            `yaml:"reward_multiplier"`
	ResourceCost       float64            `yaml:"resource_cost"`
	MinLevel           int                `yaml:"min_level"`
	MaxLevel           int                `yaml:"max_level"`
	RequiredItems      []string           `yaml:"required_items"`
	ExcludedConditions []string           `yaml:"excluded_conditions"`
	CustomParameters   map[string]float64 `yaml:"custom_parameters"`
}

// ResourceLimit defines constraints on resource generation
type ResourceLimit struct {
	ResourceType    string  `yaml:"resource_type"`
	BaseAllowance   float64 `yaml:"base_allowance"`   // Base resource allocation per level
	ScalingRate     float64 `yaml:"scaling_rate"`     // How allocation scales with level
	AbsoluteMax     float64 `yaml:"absolute_max"`     // Hard cap on resource generation
	DepletionRate   float64 `yaml:"depletion_rate"`   // Rate at which resources become scarce
	ReplenishRate   float64 `yaml:"replenish_rate"`   // Rate of resource replenishment
	CriticalReserve float64 `yaml:"critical_reserve"` // Minimum reserve to maintain
}

// BalanceMetrics tracks balance performance and system health
type BalanceMetrics struct {
	mu                     sync.RWMutex
	TotalBalanceChecks     int64                       `json:"total_balance_checks"`
	SuccessfulBalances     int64                       `json:"successful_balances"`
	FailedBalances         int64                       `json:"failed_balances"`
	CriticalFailures       int64                       `json:"critical_failures"`
	AverageBalanceTime     time.Duration               `json:"average_balance_time"`
	ContentTypeMetrics     map[ContentType]TypeMetrics `json:"content_type_metrics"`
	ResourceUsageMetrics   map[string]ResourceMetrics  `json:"resource_usage_metrics"`
	DifficultyDistribution map[int]int64               `json:"difficulty_distribution"`
	LastBalanceCheck       time.Time                   `json:"last_balance_check"`
	SystemHealth           float64                     `json:"system_health"`
}

// TypeMetrics tracks balance metrics for specific content types
type TypeMetrics struct {
	TotalGenerated      int64   `json:"total_generated"`
	BalanceFailures     int64   `json:"balance_failures"`
	AverageDifficulty   float64 `json:"average_difficulty"`
	AverageReward       float64 `json:"average_reward"`
	VarianceDeviation   float64 `json:"variance_deviation"`
	PowerCurveDeviation float64 `json:"power_curve_deviation"`
}

// ResourceMetrics tracks resource usage and availability
type ResourceMetrics struct {
	TotalUsage      float64 `json:"total_usage"`
	AverageUsage    float64 `json:"average_usage"`
	PeakUsage       float64 `json:"peak_usage"`
	CurrentReserve  float64 `json:"current_reserve"`
	DepletionEvents int64   `json:"depletion_events"`
	WasteEvents     int64   `json:"waste_events"`
}

// ScalingType defines how content scales with difficulty
type ScalingType string

const (
	ScalingLinear      ScalingType = "linear"
	ScalingExponential ScalingType = "exponential"
	ScalingLogarithmic ScalingType = "logarithmic"
	ScalingStepped     ScalingType = "stepped"
	ScalingCustom      ScalingType = "custom"
)

// BalanceRequest contains parameters for content balancing
type BalanceRequest struct {
	ContentType   ContentType            `json:"content_type"`
	PlayerLevel   int                    `json:"player_level"`
	Difficulty    int                    `json:"difficulty"`
	ContentValue  interface{}            `json:"content_value"`
	Context       map[string]interface{} `json:"context"`
	RequiredGaps  []string               `json:"required_gaps"`
	ExistingItems []string               `json:"existing_items"`
}

// BalanceResult contains the balanced content and metadata
type BalanceResult struct {
	BalancedContent interface{}            `json:"balanced_content"`
	AppliedScaling  float64                `json:"applied_scaling"`
	DifficultyScore float64                `json:"difficulty_score"`
	RewardScore     float64                `json:"reward_score"`
	ResourceCost    map[string]float64     `json:"resource_cost"`
	Metadata        map[string]interface{} `json:"metadata"`
	BalanceQuality  float64                `json:"balance_quality"`
	Warnings        []string               `json:"warnings"`
}

// NewContentBalancer creates a new content balancer with default configuration
func NewContentBalancer(logger *logrus.Logger) *ContentBalancer {
	if logger == nil {
		logger = logrus.New()
	}

	resourceLimits := getDefaultResourceLimits()
	resourceMetrics := make(map[string]ResourceMetrics)

	// Initialize resource metrics with starting reserves
	for resourceType, limit := range resourceLimits {
		resourceMetrics[resourceType] = ResourceMetrics{
			CurrentReserve: limit.BaseAllowance * 5, // Start with 5x base allowance
		}
	}

	balancer := &ContentBalancer{
		version:        "1.0.0",
		logger:         logger,
		rng:            rand.New(rand.NewSource(time.Now().UnixNano())),
		config:         getDefaultBalanceConfig(),
		powerCurves:    getDefaultPowerCurves(),
		scalingRules:   getDefaultScalingRules(),
		resourceLimits: resourceLimits,
		metrics: &BalanceMetrics{
			ContentTypeMetrics:     make(map[ContentType]TypeMetrics),
			ResourceUsageMetrics:   resourceMetrics,
			DifficultyDistribution: make(map[int]int64),
			LastBalanceCheck:       time.Now(),
			SystemHealth:           1.0,
		},
	}

	logger.WithField("version", balancer.version).Info("Content balancer initialized")
	return balancer
}

// BalanceContent applies difficulty scaling and balance rules to generated content
func (cb *ContentBalancer) BalanceContent(ctx context.Context, request BalanceRequest) (*BalanceResult, error) {
	startTime := time.Now()
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Validate request
	if err := cb.validateBalanceRequest(request); err != nil {
		cb.metrics.FailedBalances++
		return nil, fmt.Errorf("invalid balance request: %w", err)
	}

	// Check context for cancellation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Get scaling rule for content type
	scalingRule, exists := cb.scalingRules[request.ContentType]
	if !exists {
		cb.logger.WithField("content_type", request.ContentType).Warn("No scaling rule found, using defaults")
		scalingRule = cb.getDefaultScalingRule(request.ContentType)
	}

	// Calculate power curve scaling
	powerCurve := cb.powerCurves[request.ContentType]
	scaling := cb.calculatePowerScaling(request.PlayerLevel, request.Difficulty, powerCurve)

	// Apply content-specific balancing
	balancedContent, err := cb.applyContentBalancing(request, scalingRule, scaling)
	if err != nil {
		cb.metrics.FailedBalances++
		return nil, fmt.Errorf("failed to apply content balancing: %w", err)
	}

	// Calculate resource costs
	resourceCost := cb.calculateResourceCost(request, scalingRule)

	// Validate resource availability
	if err := cb.validateResourceAvailability(resourceCost); err != nil {
		cb.metrics.FailedBalances++
		return nil, fmt.Errorf("resource constraints violated: %w", err)
	}

	// Calculate balance quality metrics
	balanceQuality := cb.assessBalanceQuality(request, scaling, resourceCost)

	// Collect warnings
	warnings := cb.collectBalanceWarnings(request, scaling, balanceQuality)

	// Update metrics
	cb.updateBalanceMetrics(request, scaling, time.Since(startTime), balanceQuality >= cb.config.BalanceToleranceMin)

	// Consume resources
	cb.consumeResources(resourceCost)

	result := &BalanceResult{
		BalancedContent: balancedContent,
		AppliedScaling:  scaling,
		DifficultyScore: cb.calculateDifficultyScore(request, scaling),
		RewardScore:     cb.calculateRewardScore(request, scaling),
		ResourceCost:    resourceCost,
		Metadata:        cb.collectMetadata(request, scalingRule),
		BalanceQuality:  balanceQuality,
		Warnings:        warnings,
	}

	cb.metrics.SuccessfulBalances++
	cb.logger.WithFields(logrus.Fields{
		"content_type":    request.ContentType,
		"player_level":    request.PlayerLevel,
		"applied_scaling": scaling,
		"balance_quality": balanceQuality,
		"processing_time": time.Since(startTime),
	}).Debug("Content balanced successfully")

	return result, nil
}

// calculatePowerScaling determines the scaling factor based on level and difficulty
func (cb *ContentBalancer) calculatePowerScaling(playerLevel, difficulty int, curve PowerCurve) float64 {
	// Base scaling from power curve
	levelScaling := curve.BaseValue +
		(float64(playerLevel) * curve.ScalingFactor) +
		(math.Pow(float64(playerLevel), curve.ExponentFactor))

	// Apply difficulty modifier
	difficultyScaling := 1.0 + (float64(difficulty) * 0.1)

	// Apply breakpoint modifications
	if playerLevel > curve.BreakpointLevel {
		excess := float64(playerLevel - curve.BreakpointLevel)
		levelScaling *= (1.0 + excess*0.05) // Reduced scaling after breakpoint
	}

	scaling := levelScaling * difficultyScaling

	// Apply variance
	if curve.VarianceRange > 0 {
		variance := (cb.rng.Float64()*2 - 1) * curve.VarianceRange * scaling
		scaling += variance
	}

	// Apply cap
	if curve.CapValue > 0 && scaling > curve.CapValue {
		scaling = curve.CapValue
	}

	return math.Max(scaling, curve.BaseValue)
}

// applyContentBalancing applies scaling to specific content types
func (cb *ContentBalancer) applyContentBalancing(request BalanceRequest, rule ScalingRule, scaling float64) (interface{}, error) {
	switch request.ContentType {
	case ContentTypeQuests:
		return cb.balanceQuest(request.ContentValue, rule, scaling)
	case ContentTypeCharacters:
		return cb.balanceCharacter(request.ContentValue, rule, scaling)
	case ContentTypeDungeon:
		return cb.balanceDungeon(request.ContentValue, rule, scaling)
	case ContentTypeItems:
		return cb.balanceItem(request.ContentValue, rule, scaling)
	case ContentTypeTerrain:
		return cb.balanceTerrain(request.ContentValue, rule, scaling)
	default:
		return cb.balanceGenericContent(request.ContentValue, rule, scaling)
	}
}

// balanceQuest applies balance rules to quest content
func (cb *ContentBalancer) balanceQuest(content interface{}, rule ScalingRule, scaling float64) (interface{}, error) {
	quest, ok := content.(*game.Quest)
	if !ok {
		return nil, fmt.Errorf("invalid quest content type")
	}

	// Create a copy to avoid modifying original
	balancedQuest := *quest

	// Scale experience rewards
	for i := range balancedQuest.Rewards {
		if balancedQuest.Rewards[i].Type == "exp" {
			originalExp := float64(balancedQuest.Rewards[i].Value)
			scaledExp := originalExp * scaling * rule.RewardMultiplier
			balancedQuest.Rewards[i].Value = int(scaledExp)
		}

		// Scale gold rewards
		if balancedQuest.Rewards[i].Type == "gold" {
			originalGold := float64(balancedQuest.Rewards[i].Value)
			scaledGold := originalGold * scaling * rule.RewardMultiplier
			balancedQuest.Rewards[i].Value = int(scaledGold)
		}
	}

	// Scale objective requirements
	for i := range balancedQuest.Objectives {
		if balancedQuest.Objectives[i].Required > 0 {
			originalRequired := float64(balancedQuest.Objectives[i].Required)
			scaledRequired := originalRequired * math.Sqrt(scaling) // Less aggressive scaling for objectives
			balancedQuest.Objectives[i].Required = int(math.Max(scaledRequired, 1))
		}
	}

	return &balancedQuest, nil
}

// balanceCharacter applies balance rules to character content
func (cb *ContentBalancer) balanceCharacter(content interface{}, rule ScalingRule, scaling float64) (interface{}, error) {
	character, ok := content.(*game.Character)
	if !ok {
		return nil, fmt.Errorf("invalid character content type")
	}

	// Create a copy to avoid modifying original
	balancedChar := character.Clone()

	// Scale hit points
	originalHP := float64(balancedChar.MaxHP)
	scaledHP := originalHP * scaling * rule.DifficultyFactor
	balancedChar.MaxHP = int(scaledHP)
	balancedChar.HP = balancedChar.MaxHP

	// Scale armor class (inverse scaling - higher level means better AC)
	if scaling > 1.0 {
		acImprovement := int((scaling - 1.0) * 2)
		balancedChar.ArmorClass = int(math.Max(float64(balancedChar.ArmorClass-acImprovement), -10)) // AC cap at -10
	}

	// Scale THAC0 (inverse scaling - lower is better)
	if scaling > 1.0 {
		thac0Improvement := int((scaling - 1.0) * 1.5)
		balancedChar.THAC0 = int(math.Max(float64(balancedChar.THAC0-thac0Improvement), 1)) // THAC0 cap at 1
	}

	return balancedChar, nil
}

// balanceDungeon applies balance rules to dungeon content
func (cb *ContentBalancer) balanceDungeon(content interface{}, rule ScalingRule, scaling float64) (interface{}, error) {
	dungeon, ok := content.(*DungeonComplex)
	if !ok {
		return nil, fmt.Errorf("invalid dungeon content type")
	}

	// Create a copy to avoid modifying original
	balancedDungeon := *dungeon

	// Scale difficulty progression
	balancedDungeon.Difficulty.BaseDifficulty = int(float64(balancedDungeon.Difficulty.BaseDifficulty) * scaling)
	balancedDungeon.Difficulty.ScalingFactor *= scaling
	balancedDungeon.Difficulty.MaxDifficulty = int(float64(balancedDungeon.Difficulty.MaxDifficulty) * scaling)

	return &balancedDungeon, nil
}

// balanceItem applies balance rules to item content
func (cb *ContentBalancer) balanceItem(content interface{}, rule ScalingRule, scaling float64) (interface{}, error) {
	item, ok := content.(*game.Item)
	if !ok {
		return nil, fmt.Errorf("invalid item content type")
	}

	// Create a copy to avoid modifying original
	balancedItem := *item

	// Scale item value based on scaling factor
	if balancedItem.Value > 0 {
		originalValue := float64(balancedItem.Value)
		scaledValue := originalValue * scaling * rule.RewardMultiplier
		balancedItem.Value = int(scaledValue)
	}

	return &balancedItem, nil
}

// balanceTerrain applies balance rules to terrain content
func (cb *ContentBalancer) balanceTerrain(content interface{}, rule ScalingRule, scaling float64) (interface{}, error) {
	// For terrain, scaling mainly affects encounter density and resource distribution
	// Return content as-is since terrain structure doesn't change much with level
	return content, nil
}

// balanceGenericContent applies basic balance rules to unknown content types
func (cb *ContentBalancer) balanceGenericContent(content interface{}, rule ScalingRule, scaling float64) (interface{}, error) {
	// For unknown content types, we can't apply specific scaling
	// Log the issue and return content unchanged
	cb.logger.WithField("scaling", scaling).Warn("Generic content balancing applied - no specific rules available")
	return content, nil
}

// Helper functions for balance calculation and management
func (cb *ContentBalancer) calculateResourceCost(request BalanceRequest, rule ScalingRule) map[string]float64 {
	resourceCost := make(map[string]float64)

	// Base resource cost
	baseCost := rule.ResourceCost * float64(request.PlayerLevel)

	resourceCost["generation_budget"] = baseCost
	resourceCost["complexity_budget"] = baseCost * 0.5
	resourceCost["balance_budget"] = baseCost * 0.3

	return resourceCost
}

// getDefaultBalanceConfig returns the default balance configuration
func getDefaultBalanceConfig() BalanceConfig {
	return BalanceConfig{
		BasePlayerLevel:       1,
		MaxDifficultyLevel:    20,
		PowerCurveExponent:    1.2,
		ResourceScarcityRate:  0.1,
		RewardInflationRate:   0.15,
		VariabilityRange:      0.2,
		BalanceToleranceMin:   0.7,
		BalanceToleranceMax:   1.3,
		CriticalFailureThresh: 0.5,
	}
}

// getDefaultPowerCurves returns default power curves for all content types
func getDefaultPowerCurves() map[ContentType]PowerCurve {
	curves := make(map[ContentType]PowerCurve)

	curves[ContentTypeQuests] = PowerCurve{
		ContentType:     ContentTypeQuests,
		BaseValue:       1.0,
		ScalingFactor:   0.1,
		ExponentFactor:  1.15,
		CapValue:        10.0,
		VarianceRange:   0.2,
		BreakpointLevel: 10,
	}

	curves[ContentTypeCharacters] = PowerCurve{
		ContentType:     ContentTypeCharacters,
		BaseValue:       1.0,
		ScalingFactor:   0.15,
		ExponentFactor:  1.1,
		CapValue:        8.0,
		VarianceRange:   0.15,
		BreakpointLevel: 12,
	}

	curves[ContentTypeDungeon] = PowerCurve{
		ContentType:     ContentTypeDungeon,
		BaseValue:       1.0,
		ScalingFactor:   0.12,
		ExponentFactor:  1.2,
		CapValue:        12.0,
		VarianceRange:   0.25,
		BreakpointLevel: 8,
	}

	curves[ContentTypeItems] = PowerCurve{
		ContentType:     ContentTypeItems,
		BaseValue:       1.0,
		ScalingFactor:   0.08,
		ExponentFactor:  1.25,
		CapValue:        15.0,
		VarianceRange:   0.3,
		BreakpointLevel: 15,
	}

	curves[ContentTypeTerrain] = PowerCurve{
		ContentType:     ContentTypeTerrain,
		BaseValue:       1.0,
		ScalingFactor:   0.05,
		ExponentFactor:  1.05,
		CapValue:        3.0,
		VarianceRange:   0.1,
		BreakpointLevel: 20,
	}

	return curves
}

// getDefaultScalingRules returns default scaling rules for all content types
func getDefaultScalingRules() map[ContentType]ScalingRule {
	rules := make(map[ContentType]ScalingRule)

	rules[ContentTypeQuests] = ScalingRule{
		ContentType:      ContentTypeQuests,
		ScalingType:      ScalingExponential,
		DifficultyFactor: 1.2,
		RewardMultiplier: 1.0,
		ResourceCost:     1.0,
		MinLevel:         1,
		MaxLevel:         20,
		CustomParameters: map[string]float64{
			"objective_scaling": 0.8,
			"time_scaling":      1.1,
		},
	}

	rules[ContentTypeCharacters] = ScalingRule{
		ContentType:      ContentTypeCharacters,
		ScalingType:      ScalingLinear,
		DifficultyFactor: 1.5,
		RewardMultiplier: 0.8,
		ResourceCost:     1.2,
		MinLevel:         1,
		MaxLevel:         20,
		CustomParameters: map[string]float64{
			"stat_scaling":   1.0,
			"health_scaling": 1.3,
		},
	}

	rules[ContentTypeDungeon] = ScalingRule{
		ContentType:      ContentTypeDungeon,
		ScalingType:      ScalingExponential,
		DifficultyFactor: 1.8,
		RewardMultiplier: 1.2,
		ResourceCost:     2.0,
		MinLevel:         1,
		MaxLevel:         20,
		CustomParameters: map[string]float64{
			"room_scaling":      1.1,
			"encounter_scaling": 1.4,
		},
	}

	rules[ContentTypeItems] = ScalingRule{
		ContentType:      ContentTypeItems,
		ScalingType:      ScalingLogarithmic,
		DifficultyFactor: 0.8,
		RewardMultiplier: 1.5,
		ResourceCost:     0.8,
		MinLevel:         1,
		MaxLevel:         20,
		CustomParameters: map[string]float64{
			"value_scaling":  1.2,
			"rarity_scaling": 1.0,
		},
	}

	rules[ContentTypeTerrain] = ScalingRule{
		ContentType:      ContentTypeTerrain,
		ScalingType:      ScalingLinear,
		DifficultyFactor: 0.5,
		RewardMultiplier: 0.3,
		ResourceCost:     0.5,
		MinLevel:         1,
		MaxLevel:         20,
		CustomParameters: map[string]float64{
			"feature_scaling":    0.8,
			"complexity_scaling": 0.6,
		},
	}

	return rules
}

// getDefaultResourceLimits returns default resource limits
func getDefaultResourceLimits() map[string]ResourceLimit {
	limits := make(map[string]ResourceLimit)

	limits["generation_budget"] = ResourceLimit{
		ResourceType:    "generation_budget",
		BaseAllowance:   100.0,
		ScalingRate:     1.1,
		AbsoluteMax:     1000.0,
		DepletionRate:   0.05,
		ReplenishRate:   0.1,
		CriticalReserve: 10.0,
	}

	limits["complexity_budget"] = ResourceLimit{
		ResourceType:    "complexity_budget",
		BaseAllowance:   50.0,
		ScalingRate:     1.05,
		AbsoluteMax:     500.0,
		DepletionRate:   0.03,
		ReplenishRate:   0.08,
		CriticalReserve: 5.0,
	}

	limits["balance_budget"] = ResourceLimit{
		ResourceType:    "balance_budget",
		BaseAllowance:   30.0,
		ScalingRate:     1.02,
		AbsoluteMax:     300.0,
		DepletionRate:   0.02,
		ReplenishRate:   0.05,
		CriticalReserve: 3.0,
	}

	return limits
}

// Additional helper methods would continue here...
// (Validation, metrics, resource management, etc.)

// validateBalanceRequest validates the balance request parameters
func (cb *ContentBalancer) validateBalanceRequest(request BalanceRequest) error {
	if request.PlayerLevel < 1 {
		return fmt.Errorf("player level must be positive, got %d", request.PlayerLevel)
	}

	if request.Difficulty < 0 {
		return fmt.Errorf("difficulty must be non-negative, got %d", request.Difficulty)
	}

	if request.ContentValue == nil {
		return fmt.Errorf("content value cannot be nil")
	}

	return nil
}

// getDefaultScalingRule returns a default scaling rule for unknown content types
func (cb *ContentBalancer) getDefaultScalingRule(contentType ContentType) ScalingRule {
	return ScalingRule{
		ContentType:      contentType,
		ScalingType:      ScalingLinear,
		DifficultyFactor: 1.0,
		RewardMultiplier: 1.0,
		ResourceCost:     1.0,
		MinLevel:         1,
		MaxLevel:         20,
		CustomParameters: make(map[string]float64),
	}
}

// validateResourceAvailability checks if resources are available for the request
func (cb *ContentBalancer) validateResourceAvailability(resourceCost map[string]float64) error {
	for resourceType, cost := range resourceCost {
		limit, exists := cb.resourceLimits[resourceType]
		if !exists {
			continue // Skip validation for unknown resource types
		}

		// Get current resource usage from metrics
		resourceMetrics, exists := cb.metrics.ResourceUsageMetrics[resourceType]
		if !exists {
			continue // No usage recorded yet
		}

		// Check if request would exceed absolute maximum
		if resourceMetrics.CurrentReserve+cost > limit.AbsoluteMax {
			return fmt.Errorf("resource %s would exceed absolute maximum (%.2f + %.2f > %.2f)",
				resourceType, resourceMetrics.CurrentReserve, cost, limit.AbsoluteMax)
		}

		// Check if reserve would fall below critical level
		if resourceMetrics.CurrentReserve-cost < limit.CriticalReserve {
			return fmt.Errorf("resource %s would fall below critical reserve (%.2f - %.2f < %.2f)",
				resourceType, resourceMetrics.CurrentReserve, cost, limit.CriticalReserve)
		}
	}

	return nil
}

// assessBalanceQuality calculates a quality score for the balance operation
func (cb *ContentBalancer) assessBalanceQuality(request BalanceRequest, scaling float64, resourceCost map[string]float64) float64 {
	quality := 1.0

	// Check if scaling is within reasonable bounds
	if scaling < cb.config.BalanceToleranceMin || scaling > cb.config.BalanceToleranceMax {
		quality *= 0.7 // Reduce quality for extreme scaling
	}

	// Check resource efficiency
	totalCost := 0.0
	for _, cost := range resourceCost {
		totalCost += cost
	}
	expectedCost := float64(request.PlayerLevel) * 1.5

	if totalCost > expectedCost*2 {
		quality *= 0.8 // Reduce quality for excessive resource usage
	}

	// Add some randomization based on complexity
	complexityFactor := 1.0 - (cb.rng.Float64() * 0.1) // 0-10% reduction
	quality *= complexityFactor

	return math.Max(quality, 0.0)
}

// collectBalanceWarnings collects warnings about the balance operation
func (cb *ContentBalancer) collectBalanceWarnings(request BalanceRequest, scaling float64, balanceQuality float64) []string {
	var warnings []string

	if scaling > cb.config.BalanceToleranceMax {
		warnings = append(warnings, fmt.Sprintf("Scaling factor %.2f exceeds maximum tolerance %.2f",
			scaling, cb.config.BalanceToleranceMax))
	}

	if scaling < cb.config.BalanceToleranceMin {
		warnings = append(warnings, fmt.Sprintf("Scaling factor %.2f below minimum tolerance %.2f",
			scaling, cb.config.BalanceToleranceMin))
	}

	if balanceQuality < cb.config.CriticalFailureThresh {
		warnings = append(warnings, fmt.Sprintf("Balance quality %.2f below critical threshold %.2f",
			balanceQuality, cb.config.CriticalFailureThresh))
	}

	if request.PlayerLevel > cb.config.MaxDifficultyLevel {
		warnings = append(warnings, fmt.Sprintf("Player level %d exceeds maximum configured level %d",
			request.PlayerLevel, cb.config.MaxDifficultyLevel))
	}

	return warnings
}

// updateBalanceMetrics updates the balance performance metrics
func (cb *ContentBalancer) updateBalanceMetrics(request BalanceRequest, scaling float64, duration time.Duration, success bool) {
	cb.metrics.TotalBalanceChecks++
	cb.metrics.LastBalanceCheck = time.Now()

	// Update difficulty distribution
	difficultyKey := request.Difficulty
	if difficultyKey > 20 {
		difficultyKey = 20 // Cap at 20 for histogram
	}
	cb.metrics.DifficultyDistribution[difficultyKey]++

	// Update content type metrics
	typeMetrics := cb.metrics.ContentTypeMetrics[request.ContentType]
	typeMetrics.TotalGenerated++

	if !success {
		typeMetrics.BalanceFailures++
	}

	// Update average difficulty (running average)
	if typeMetrics.TotalGenerated == 1 {
		typeMetrics.AverageDifficulty = float64(request.Difficulty)
	} else {
		typeMetrics.AverageDifficulty = (typeMetrics.AverageDifficulty*float64(typeMetrics.TotalGenerated-1) + float64(request.Difficulty)) / float64(typeMetrics.TotalGenerated)
	}

	cb.metrics.ContentTypeMetrics[request.ContentType] = typeMetrics

	// Update average balance time (running average)
	if cb.metrics.TotalBalanceChecks == 1 {
		cb.metrics.AverageBalanceTime = duration
	} else {
		cb.metrics.AverageBalanceTime = time.Duration((int64(cb.metrics.AverageBalanceTime)*int64(cb.metrics.TotalBalanceChecks-1) + int64(duration)) / int64(cb.metrics.TotalBalanceChecks))
	}

	// Calculate system health based on success rate
	successRate := float64(cb.metrics.SuccessfulBalances) / float64(cb.metrics.TotalBalanceChecks)
	cb.metrics.SystemHealth = successRate
}

// consumeResources updates resource usage metrics
func (cb *ContentBalancer) consumeResources(resourceCost map[string]float64) {
	for resourceType, cost := range resourceCost {
		resourceMetrics := cb.metrics.ResourceUsageMetrics[resourceType]

		resourceMetrics.TotalUsage += cost
		resourceMetrics.CurrentReserve -= cost

		// Update running average
		if resourceMetrics.TotalUsage == cost {
			resourceMetrics.AverageUsage = cost
		} else {
			// Approximate running average (simplified)
			resourceMetrics.AverageUsage = (resourceMetrics.AverageUsage + cost) / 2
		}

		// Update peak usage
		if cost > resourceMetrics.PeakUsage {
			resourceMetrics.PeakUsage = cost
		}

		// Check for depletion
		limit := cb.resourceLimits[resourceType]
		if resourceMetrics.CurrentReserve < limit.CriticalReserve {
			resourceMetrics.DepletionEvents++
		}

		cb.metrics.ResourceUsageMetrics[resourceType] = resourceMetrics
	}
}

// calculateDifficultyScore calculates a difficulty score for the content
func (cb *ContentBalancer) calculateDifficultyScore(request BalanceRequest, scaling float64) float64 {
	baseScore := float64(request.Difficulty) * 10 // Scale to 0-200 range
	scaledScore := baseScore * scaling

	// Add level-based component
	levelComponent := float64(request.PlayerLevel) * 5

	totalScore := scaledScore + levelComponent

	// Normalize to 0-100 range
	normalizedScore := math.Min(totalScore/3, 100)

	return normalizedScore
}

// calculateRewardScore calculates a reward score for the content
func (cb *ContentBalancer) calculateRewardScore(request BalanceRequest, scaling float64) float64 {
	rule := cb.scalingRules[request.ContentType]

	baseScore := 50.0 // Base reward score
	scaledScore := baseScore * scaling * rule.RewardMultiplier

	// Add level-based component
	levelComponent := float64(request.PlayerLevel) * 2

	totalScore := scaledScore + levelComponent

	// Normalize to 0-100 range
	normalizedScore := math.Min(totalScore/2, 100)

	return normalizedScore
}

// collectMetadata collects metadata about the balance operation
func (cb *ContentBalancer) collectMetadata(request BalanceRequest, rule ScalingRule) map[string]interface{} {
	metadata := make(map[string]interface{})

	metadata["content_type"] = string(request.ContentType)
	metadata["player_level"] = request.PlayerLevel
	metadata["difficulty"] = request.Difficulty
	metadata["scaling_type"] = string(rule.ScalingType)
	metadata["balance_version"] = cb.version
	metadata["timestamp"] = time.Now().Unix()

	// Add power curve info
	if curve, exists := cb.powerCurves[request.ContentType]; exists {
		metadata["power_curve"] = map[string]interface{}{
			"base_value":       curve.BaseValue,
			"scaling_factor":   curve.ScalingFactor,
			"exponent_factor":  curve.ExponentFactor,
			"breakpoint_level": curve.BreakpointLevel,
		}
	}

	return metadata
}

// GetVersion returns the balancer version
func (cb *ContentBalancer) GetVersion() string {
	return cb.version
}

// GetMetrics returns current balance metrics
func (cb *ContentBalancer) GetMetrics() *BalanceMetrics {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	// Return a copy to prevent external modification
	metricsCopy := *cb.metrics
	return &metricsCopy
}
