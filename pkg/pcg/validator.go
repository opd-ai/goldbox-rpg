package pcg

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"goldbox-rpg/pkg/game"
)

// ContentValidator provides validation services for procedurally generated content
// to ensure logical consistency, game balance, and quality standards.
//
// Design Approach:
// - Validates content after generation to catch inconsistencies and edge cases
// - Provides automated validation of generated scenarios for quality assurance
// - Implements fallback mechanisms for handling edge cases and generation failures
// - Uses configurable validation rules that can be adjusted based on game requirements
//
// Quality Standards:
// - Ensures generated content follows established game mechanics
// - Validates narrative coherence and logical consistency
// - Checks for appropriate difficulty scaling and balance
// - Prevents broken or impossible game states
type ContentValidator struct {
	mu               sync.RWMutex
	logger           *logrus.Logger
	validationRules  map[ContentType][]ValidationRule
	fallbackHandlers map[ContentType]FallbackHandler
	metrics          *ValidationMetrics
}

// ValidationRule defines a single validation check for content
type ValidationRule struct {
	Name        string                           // Human-readable name for the rule
	Description string                           // Description of what the rule validates
	Severity    ValidationSeverity               // How critical this validation is
	Validator   func(content interface{}) Result // Function that performs the validation
}

// ValidationSeverity indicates how critical a validation failure is
type ValidationSeverity string

const (
	SeverityInfo     ValidationSeverity = "info"     // Informational, not a problem
	SeverityWarning  ValidationSeverity = "warning"  // Should be addressed but not critical
	SeverityError    ValidationSeverity = "error"    // Must be fixed, affects gameplay
	SeverityCritical ValidationSeverity = "critical" // Game-breaking, requires immediate fix
)

// Result represents the outcome of a validation check
type Result struct {
	Passed   bool                   // Whether the validation passed
	Severity ValidationSeverity     // Severity level of any issues found
	Message  string                 // Human-readable description of the result
	Details  map[string]interface{} // Additional context about the validation
	FixHints []string               // Suggestions for fixing validation failures
}

// FallbackHandler provides mechanisms for handling validation failures
type FallbackHandler interface {
	// CanHandle determines if this handler can fix the validation failure
	CanHandle(result Result) bool

	// Handle attempts to fix the validation failure and return corrected content
	Handle(ctx context.Context, content interface{}, result Result) (interface{}, error)

	// GetDescription returns a human-readable description of what this handler does
	GetDescription() string
}

// ValidationMetrics tracks validation statistics and performance
type ValidationMetrics struct {
	mu                  sync.RWMutex
	totalValidations    int64
	passedValidations   int64
	failedValidations   int64
	criticalFailures    int64
	fallbacksTriggered  int64
	validationDuration  time.Duration
	ruleExecutionCounts map[string]int64
}

// NewContentValidator creates a new content validator with default rules
func NewContentValidator(logger *logrus.Logger) *ContentValidator {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.WarnLevel)
	}

	validator := &ContentValidator{
		logger:           logger,
		validationRules:  make(map[ContentType][]ValidationRule),
		fallbackHandlers: make(map[ContentType]FallbackHandler),
		metrics:          NewValidationMetrics(),
	}

	// Initialize default validation rules for each content type
	validator.initializeDefaultRules()
	validator.initializeFallbackHandlers()

	return validator
}

// ValidateContent validates generated content against registered rules
func (cv *ContentValidator) ValidateContent(ctx context.Context, contentType ContentType, content interface{}) ([]Result, error) {
	cv.mu.RLock()
	rules, exists := cv.validationRules[contentType]
	cv.mu.RUnlock()

	if !exists {
		cv.logger.WithField("content_type", contentType).Warn("no validation rules found for content type")
		return []Result{}, nil
	}

	startTime := time.Now()
	results := make([]Result, 0, len(rules))

	cv.logger.WithFields(logrus.Fields{
		"content_type": contentType,
		"rule_count":   len(rules),
	}).Debug("starting content validation")

	for _, rule := range rules {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		result := rule.Validator(content)
		result.Severity = rule.Severity // Ensure severity is set from rule

		cv.metrics.recordRuleExecution(rule.Name)

		if !result.Passed {
			cv.logger.WithFields(logrus.Fields{
				"rule":     rule.Name,
				"severity": result.Severity,
				"message":  result.Message,
			}).Warn("validation rule failed")
		}

		results = append(results, result)
	}

	duration := time.Since(startTime)
	cv.metrics.recordValidation(results, duration)

	cv.logger.WithFields(logrus.Fields{
		"content_type": contentType,
		"duration":     duration,
		"results":      len(results),
	}).Debug("content validation completed")

	return results, nil
}

// ValidateAndFix validates content and attempts to fix any failures using fallback handlers
func (cv *ContentValidator) ValidateAndFix(ctx context.Context, contentType ContentType, content interface{}) (interface{}, []Result, error) {
	results, err := cv.ValidateContent(ctx, contentType, content)
	if err != nil {
		return content, results, err
	}

	fixedContent := content
	hasFailures := false

	for _, result := range results {
		if !result.Passed && (result.Severity == SeverityError || result.Severity == SeverityCritical) {
			hasFailures = true
			break
		}
	}

	if hasFailures {
		cv.mu.RLock()
		handler, exists := cv.fallbackHandlers[contentType]
		cv.mu.RUnlock()

		if exists {
			for _, result := range results {
				if !result.Passed && handler.CanHandle(result) {
					cv.logger.WithFields(logrus.Fields{
						"content_type": contentType,
						"handler":      handler.GetDescription(),
						"issue":        result.Message,
					}).Info("attempting to fix validation failure")

					fixed, err := handler.Handle(ctx, fixedContent, result)
					if err != nil {
						cv.logger.WithError(err).Warn("fallback handler failed")
						continue
					}

					fixedContent = fixed
					cv.metrics.recordFallback()

					cv.logger.WithField("content_type", contentType).Info("validation failure fixed by fallback handler")
					break // Only apply one fix at a time to avoid conflicts
				}
			}
		}
	}

	return fixedContent, results, nil
}

// RegisterValidationRule adds a custom validation rule for a content type
func (cv *ContentValidator) RegisterValidationRule(contentType ContentType, rule ValidationRule) {
	cv.mu.Lock()
	defer cv.mu.Unlock()

	cv.validationRules[contentType] = append(cv.validationRules[contentType], rule)

	cv.logger.WithFields(logrus.Fields{
		"content_type": contentType,
		"rule_name":    rule.Name,
		"severity":     rule.Severity,
	}).Debug("registered validation rule")
}

// RegisterFallbackHandler adds a fallback handler for a content type
func (cv *ContentValidator) RegisterFallbackHandler(contentType ContentType, handler FallbackHandler) {
	cv.mu.Lock()
	defer cv.mu.Unlock()

	cv.fallbackHandlers[contentType] = handler

	cv.logger.WithFields(logrus.Fields{
		"content_type": contentType,
		"handler":      handler.GetDescription(),
	}).Debug("registered fallback handler")
}

// GetValidationMetrics returns current validation statistics
func (cv *ContentValidator) GetValidationMetrics() ValidationMetrics {
	return cv.metrics.getStats()
}

// ResetMetrics clears all validation metrics
func (cv *ContentValidator) ResetMetrics() {
	cv.metrics.reset()
}

// initializeDefaultRules sets up standard validation rules for each content type
func (cv *ContentValidator) initializeDefaultRules() {
	// Character validation rules
	cv.registerCharacterRules()

	// Quest validation rules
	cv.registerQuestRules()

	// Dungeon validation rules
	cv.registerDungeonRules()

	// Dialogue validation rules
	cv.registerDialogueRules()

	// Faction validation rules
	cv.registerFactionRules()

	// World validation rules
	cv.registerWorldRules()
}

// registerCharacterRules adds validation rules for character content
func (cv *ContentValidator) registerCharacterRules() {
	rules := []ValidationRule{
		{
			Name:        "character_attributes_valid",
			Description: "Ensures character attributes are within valid ranges",
			Severity:    SeverityError,
			Validator: func(content interface{}) Result {
				char, ok := content.(*game.Character)
				if !ok {
					return Result{
						Passed:  false,
						Message: "content is not a valid character",
						Details: map[string]interface{}{"type": fmt.Sprintf("%T", content)},
					}
				}

				// Check attribute ranges (D&D standard: 3-18 for base stats)
				attributes := map[string]int{
					"strength":     char.Strength,
					"dexterity":    char.Dexterity,
					"constitution": char.Constitution,
					"intelligence": char.Intelligence,
					"wisdom":       char.Wisdom,
					"charisma":     char.Charisma,
				}

				invalidAttrs := make([]string, 0)
				for name, value := range attributes {
					if value < 3 || value > 25 { // Allow some buffer for magical enhancement
						invalidAttrs = append(invalidAttrs, fmt.Sprintf("%s: %d", name, value))
					}
				}

				if len(invalidAttrs) > 0 {
					return Result{
						Passed:  false,
						Message: fmt.Sprintf("character has invalid attributes: %s", strings.Join(invalidAttrs, ", ")),
						Details: map[string]interface{}{
							"invalid_attributes": invalidAttrs,
							"character_id":       char.ID,
						},
						FixHints: []string{
							"Clamp attribute values to valid ranges (3-25)",
							"Check character generation algorithm for balance issues",
						},
					}
				}

				return Result{Passed: true, Message: "character attributes are valid"}
			},
		},
		{
			Name:        "character_name_not_empty",
			Description: "Ensures character has a non-empty name",
			Severity:    SeverityError,
			Validator: func(content interface{}) Result {
				char, ok := content.(*game.Character)
				if !ok {
					return Result{Passed: false, Message: "content is not a valid character"}
				}

				if strings.TrimSpace(char.Name) == "" {
					return Result{
						Passed:  false,
						Message: "character has empty or whitespace-only name",
						Details: map[string]interface{}{"character_id": char.ID},
						FixHints: []string{
							"Generate a default name using name generation system",
							"Use character ID as fallback name",
						},
					}
				}

				return Result{Passed: true, Message: "character has valid name"}
			},
		},
	}

	for _, rule := range rules {
		cv.RegisterValidationRule(ContentTypeCharacters, rule)
	}
}

// registerQuestRules adds validation rules for quest content
func (cv *ContentValidator) registerQuestRules() {
	rules := []ValidationRule{
		{
			Name:        "quest_has_objectives",
			Description: "Ensures quest has at least one objective",
			Severity:    SeverityError,
			Validator: func(content interface{}) Result {
				quest, ok := content.(*game.Quest)
				if !ok {
					return Result{Passed: false, Message: "content is not a valid quest"}
				}

				if len(quest.Objectives) == 0 {
					return Result{
						Passed:  false,
						Message: "quest has no objectives",
						Details: map[string]interface{}{"quest_id": quest.ID},
						FixHints: []string{
							"Add at least one objective to the quest",
							"Use default fetch objective as fallback",
						},
					}
				}

				return Result{Passed: true, Message: "quest has valid objectives"}
			},
		},
		{
			Name:        "quest_has_title",
			Description: "Ensures quest has a valid title",
			Severity:    SeverityWarning,
			Validator: func(content interface{}) Result {
				quest, ok := content.(*game.Quest)
				if !ok {
					return Result{Passed: false, Message: "content is not a valid quest"}
				}

				if strings.TrimSpace(quest.Title) == "" {
					return Result{
						Passed:  false,
						Message: "quest has empty or whitespace-only title",
						Details: map[string]interface{}{"quest_id": quest.ID},
						FixHints: []string{
							"Generate a default title based on quest type",
							"Use quest ID as fallback title",
						},
					}
				}

				return Result{Passed: true, Message: "quest has valid title"}
			},
		},
	}

	for _, rule := range rules {
		cv.RegisterValidationRule(ContentTypeQuests, rule)
	}
}

// registerDungeonRules adds validation rules for dungeon content
func (cv *ContentValidator) registerDungeonRules() {
	rules := []ValidationRule{
		{
			Name:        "dungeon_has_levels",
			Description: "Ensures dungeon has at least one level",
			Severity:    SeverityError,
			Validator: func(content interface{}) Result {
				dungeon, ok := content.(*DungeonComplex)
				if !ok {
					return Result{Passed: false, Message: "content is not a valid dungeon complex"}
				}

				if len(dungeon.Levels) == 0 {
					return Result{
						Passed:  false,
						Message: "dungeon has no levels",
						Details: map[string]interface{}{"dungeon_id": dungeon.ID},
						FixHints: []string{
							"Add at least one level to the dungeon",
							"Check dungeon generation parameters",
						},
					}
				}

				return Result{Passed: true, Message: "dungeon has valid levels"}
			},
		},
		{
			Name:        "dungeon_connectivity",
			Description: "Ensures dungeon levels are properly connected",
			Severity:    SeverityWarning,
			Validator: func(content interface{}) Result {
				dungeon, ok := content.(*DungeonComplex)
				if !ok {
					return Result{Passed: false, Message: "content is not a valid dungeon complex"}
				}

				if len(dungeon.Levels) <= 1 {
					return Result{Passed: true, Message: "single-level dungeon does not need connectivity validation"}
				}

				// Check that levels are connected
				connectedLevels := make(map[int]bool)
				for _, connection := range dungeon.Connections {
					connectedLevels[connection.FromLevel] = true
					connectedLevels[connection.ToLevel] = true
				}

				unconnectedLevels := make([]int, 0)
				for i := range dungeon.Levels {
					if !connectedLevels[i] {
						unconnectedLevels = append(unconnectedLevels, i)
					}
				}

				if len(unconnectedLevels) > 0 {
					return Result{
						Passed:  false,
						Message: fmt.Sprintf("dungeon has unconnected levels: %v", unconnectedLevels),
						Details: map[string]interface{}{
							"dungeon_id":         dungeon.ID,
							"unconnected_levels": unconnectedLevels,
							"total_levels":       len(dungeon.Levels),
						},
						FixHints: []string{
							"Add connections between all dungeon levels",
							"Ensure level 0 is accessible from entrance",
						},
					}
				}

				return Result{Passed: true, Message: "dungeon levels are properly connected"}
			},
		},
	}

	for _, rule := range rules {
		cv.RegisterValidationRule(ContentTypeDungeon, rule)
	}
}

// registerDialogueRules adds validation rules for dialogue content
func (cv *ContentValidator) registerDialogueRules() {
	rules := []ValidationRule{
		{
			Name:        "dialogue_tree_not_empty",
			Description: "Ensures dialogue tree has at least one node",
			Severity:    SeverityError,
			Validator: func(content interface{}) Result {
				tree, ok := content.(*GeneratedDialogue)
				if !ok {
					return Result{Passed: false, Message: "content is not a valid dialogue tree"}
				}

				if tree.TotalNodes == 0 {
					return Result{
						Passed:  false,
						Message: "dialogue tree is empty",
						Details: map[string]interface{}{"total_nodes": tree.TotalNodes},
						FixHints: []string{
							"Add at least one dialogue node",
							"Check dialogue generation parameters",
						},
					}
				}

				return Result{Passed: true, Message: "dialogue tree has content"}
			},
		},
	}

	for _, rule := range rules {
		cv.RegisterValidationRule(ContentTypeDialogue, rule)
	}
}

// registerFactionRules adds validation rules for faction content
func (cv *ContentValidator) registerFactionRules() {
	rules := []ValidationRule{
		{
			Name:        "faction_relationships_balanced",
			Description: "Ensures faction relationships are not all hostile or all friendly",
			Severity:    SeverityWarning,
			Validator: func(content interface{}) Result {
				system, ok := content.(*GeneratedFactionSystem)
				if !ok {
					return Result{Passed: false, Message: "content is not a valid faction system"}
				}

				if len(system.Relationships) == 0 {
					return Result{Passed: true, Message: "no relationships to validate"}
				}

				hostileCount := 0
				friendlyCount := 0
				for _, rel := range system.Relationships {
					if rel.Status == RelationStatusHostile || rel.Status == RelationStatusWar {
						hostileCount++
					} else if rel.Status == RelationStatusAllied || rel.Status == RelationStatusFriendly {
						friendlyCount++
					}
				}

				totalRels := len(system.Relationships)
				hostileRatio := float64(hostileCount) / float64(totalRels)
				friendlyRatio := float64(friendlyCount) / float64(totalRels)

				if hostileRatio > 0.8 {
					return Result{
						Passed:  false,
						Message: fmt.Sprintf("faction system is too hostile (%.1f%% hostile relationships)", hostileRatio*100),
						Details: map[string]interface{}{
							"hostile_count": hostileCount,
							"total_count":   totalRels,
							"hostile_ratio": hostileRatio,
						},
						FixHints: []string{
							"Add more neutral or friendly relationships",
							"Reduce conflict level in generation parameters",
						},
					}
				}

				if friendlyRatio > 0.8 {
					return Result{
						Passed:  false,
						Message: fmt.Sprintf("faction system is too friendly (%.1f%% friendly relationships)", friendlyRatio*100),
						Details: map[string]interface{}{
							"friendly_count": friendlyCount,
							"total_count":    totalRels,
							"friendly_ratio": friendlyRatio,
						},
						FixHints: []string{
							"Add more neutral or hostile relationships",
							"Increase conflict level in generation parameters",
						},
					}
				}

				return Result{Passed: true, Message: "faction relationships are balanced"}
			},
		},
	}

	for _, rule := range rules {
		cv.RegisterValidationRule(ContentTypeFactions, rule)
	}
}

// registerWorldRules adds validation rules for world content
func (cv *ContentValidator) registerWorldRules() {
	rules := []ValidationRule{
		{
			Name:        "world_has_settlements",
			Description: "Ensures world has at least one settlement",
			Severity:    SeverityWarning,
			Validator: func(content interface{}) Result {
				world, ok := content.(*GeneratedWorld)
				if !ok {
					return Result{Passed: false, Message: "content is not a valid generated world"}
				}

				if len(world.Settlements) == 0 {
					return Result{
						Passed:  false,
						Message: "world has no settlements",
						Details: map[string]interface{}{"world_id": world.ID},
						FixHints: []string{
							"Add at least one settlement to provide player services",
							"Check population density parameters",
						},
					}
				}

				return Result{Passed: true, Message: "world has settlements"}
			},
		},
	}

	for _, rule := range rules {
		cv.RegisterValidationRule(ContentTypeNarrative, rule) // Using narrative as world type placeholder
	}
}

// initializeFallbackHandlers sets up default fallback handlers
func (cv *ContentValidator) initializeFallbackHandlers() {
	// Character fallback handler
	cv.RegisterFallbackHandler(ContentTypeCharacters, &characterFallbackHandler{logger: cv.logger})

	// Quest fallback handler
	cv.RegisterFallbackHandler(ContentTypeQuests, &questFallbackHandler{logger: cv.logger})

	// Dungeon fallback handler
	cv.RegisterFallbackHandler(ContentTypeDungeon, &dungeonFallbackHandler{logger: cv.logger})
}

// NewValidationMetrics creates a new validation metrics tracker
func NewValidationMetrics() *ValidationMetrics {
	return &ValidationMetrics{
		ruleExecutionCounts: make(map[string]int64),
	}
}

// recordValidation records the results of a validation session
func (vm *ValidationMetrics) recordValidation(results []Result, duration time.Duration) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	vm.totalValidations++
	vm.validationDuration += duration

	passed := true
	hasCritical := false

	for _, result := range results {
		if !result.Passed {
			passed = false
			if result.Severity == SeverityCritical {
				hasCritical = true
			}
		}
	}

	if passed {
		vm.passedValidations++
	} else {
		vm.failedValidations++
		if hasCritical {
			vm.criticalFailures++
		}
	}
}

// recordRuleExecution records that a specific rule was executed
func (vm *ValidationMetrics) recordRuleExecution(ruleName string) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	vm.ruleExecutionCounts[ruleName]++
}

// recordFallback records that a fallback handler was triggered
func (vm *ValidationMetrics) recordFallback() {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	vm.fallbacksTriggered++
}

// getStats returns a copy of current metrics
func (vm *ValidationMetrics) getStats() ValidationMetrics {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	ruleCounts := make(map[string]int64)
	for k, v := range vm.ruleExecutionCounts {
		ruleCounts[k] = v
	}

	return ValidationMetrics{
		totalValidations:    vm.totalValidations,
		passedValidations:   vm.passedValidations,
		failedValidations:   vm.failedValidations,
		criticalFailures:    vm.criticalFailures,
		fallbacksTriggered:  vm.fallbacksTriggered,
		validationDuration:  vm.validationDuration,
		ruleExecutionCounts: ruleCounts,
	}
}

// reset clears all metrics
func (vm *ValidationMetrics) reset() {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	vm.totalValidations = 0
	vm.passedValidations = 0
	vm.failedValidations = 0
	vm.criticalFailures = 0
	vm.fallbacksTriggered = 0
	vm.validationDuration = 0
	vm.ruleExecutionCounts = make(map[string]int64)
}

// GetSuccessRate returns the percentage of validations that passed
func (vm *ValidationMetrics) GetSuccessRate() float64 {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	if vm.totalValidations == 0 {
		return 0.0
	}

	return float64(vm.passedValidations) / float64(vm.totalValidations) * 100.0
}

// GetAverageValidationTime returns the average time taken per validation
func (vm *ValidationMetrics) GetAverageValidationTime() time.Duration {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	if vm.totalValidations == 0 {
		return 0
	}

	return vm.validationDuration / time.Duration(vm.totalValidations)
}

// GetCriticalFailureRate returns the percentage of validations with critical failures
func (vm *ValidationMetrics) GetCriticalFailureRate() float64 {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	if vm.totalValidations == 0 {
		return 0.0
	}

	return float64(vm.criticalFailures) / float64(vm.totalValidations) * 100.0
}

// GetTotalValidations returns the total number of validations performed
func (vm *ValidationMetrics) GetTotalValidations() int64 {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.totalValidations
}

// Fallback handler implementations

// characterFallbackHandler provides fallback fixes for character validation failures
type characterFallbackHandler struct {
	logger *logrus.Logger
}

func (h *characterFallbackHandler) CanHandle(result Result) bool {
	return strings.Contains(result.Message, "attribute") ||
		strings.Contains(result.Message, "name")
}

func (h *characterFallbackHandler) Handle(ctx context.Context, content interface{}, result Result) (interface{}, error) {
	char, ok := content.(*game.Character)
	if !ok {
		return content, fmt.Errorf("content is not a character")
	}

	// Fix invalid attributes by clamping them to valid ranges
	if strings.Contains(result.Message, "attribute") {
		char.Strength = clampAttribute(char.Strength)
		char.Dexterity = clampAttribute(char.Dexterity)
		char.Constitution = clampAttribute(char.Constitution)
		char.Intelligence = clampAttribute(char.Intelligence)
		char.Wisdom = clampAttribute(char.Wisdom)
		char.Charisma = clampAttribute(char.Charisma)

		h.logger.WithField("character_id", char.ID).Info("fixed character attributes")
	}

	// Fix empty names
	if strings.Contains(result.Message, "name") && strings.TrimSpace(char.Name) == "" {
		idSuffix := char.ID
		if len(idSuffix) > 8 {
			idSuffix = idSuffix[:8]
		}
		char.Name = fmt.Sprintf("Character_%s", idSuffix)
		h.logger.WithField("character_id", char.ID).Info("generated fallback character name")
	}

	return char, nil
}

func (h *characterFallbackHandler) GetDescription() string {
	return "Character attribute and name fallback handler"
}

// questFallbackHandler provides fallback fixes for quest validation failures
type questFallbackHandler struct {
	logger *logrus.Logger
}

func (h *questFallbackHandler) CanHandle(result Result) bool {
	return strings.Contains(result.Message, "objective") ||
		strings.Contains(result.Message, "title")
}

func (h *questFallbackHandler) Handle(ctx context.Context, content interface{}, result Result) (interface{}, error) {
	quest, ok := content.(*game.Quest)
	if !ok {
		return content, fmt.Errorf("content is not a quest")
	}

	// Fix missing objectives
	if strings.Contains(result.Message, "objective") && len(quest.Objectives) == 0 {
		defaultObjective := game.QuestObjective{
			Description: "Complete the quest",
			Progress:    0,
			Required:    1,
			Completed:   false,
		}
		quest.Objectives = append(quest.Objectives, defaultObjective)
		h.logger.WithField("quest_id", quest.ID).Info("added default quest objective")
	}

	// Fix empty titles
	if strings.Contains(result.Message, "title") && strings.TrimSpace(quest.Title) == "" {
		idSuffix := quest.ID
		if len(idSuffix) > 8 {
			idSuffix = idSuffix[:8]
		}
		quest.Title = fmt.Sprintf("Quest_%s", idSuffix)
		h.logger.WithField("quest_id", quest.ID).Info("generated fallback quest title")
	}

	return quest, nil
}

func (h *questFallbackHandler) GetDescription() string {
	return "Quest objective and title fallback handler"
}

// dungeonFallbackHandler provides fallback fixes for dungeon validation failures
type dungeonFallbackHandler struct {
	logger *logrus.Logger
}

func (h *dungeonFallbackHandler) CanHandle(result Result) bool {
	return strings.Contains(result.Message, "level") ||
		strings.Contains(result.Message, "connect")
}

func (h *dungeonFallbackHandler) Handle(ctx context.Context, content interface{}, result Result) (interface{}, error) {
	dungeon, ok := content.(*DungeonComplex)
	if !ok {
		return content, fmt.Errorf("content is not a dungeon complex")
	}

	// Fix missing levels
	if strings.Contains(result.Message, "no levels") && len(dungeon.Levels) == 0 {
		defaultLevel := &DungeonLevel{
			Level:       0,
			Map:         nil, // Will be generated later
			Rooms:       make([]*RoomLayout, 0),
			Connections: make([]ConnectionPoint, 0),
			Theme:       ThemeClassic,
			Difficulty:  1,
			Properties:  make(map[string]interface{}),
		}
		dungeon.Levels[0] = defaultLevel
		h.logger.WithField("dungeon_id", dungeon.ID).Info("added default dungeon level")
	}

	// Fix connectivity issues by adding basic connections
	if strings.Contains(result.Message, "unconnected") && len(dungeon.Levels) > 1 {
		for i := 0; i < len(dungeon.Levels)-1; i++ {
			connectionExists := false
			for _, conn := range dungeon.Connections {
				if (conn.FromLevel == i && conn.ToLevel == i+1) ||
					(conn.FromLevel == i+1 && conn.ToLevel == i) {
					connectionExists = true
					break
				}
			}

			if !connectionExists {
				connection := LevelConnection{
					FromLevel:    i,
					ToLevel:      i + 1,
					FromPosition: game.Position{X: 10, Y: 10},
					ToPosition:   game.Position{X: 10, Y: 10},
					Type:         ConnectionStairs,
					Properties:   make(map[string]interface{}),
				}
				dungeon.Connections = append(dungeon.Connections, connection)
			}
		}
		h.logger.WithField("dungeon_id", dungeon.ID).Info("added missing level connections")
	}

	return dungeon, nil
}

func (h *dungeonFallbackHandler) GetDescription() string {
	return "Dungeon level and connectivity fallback handler"
}

// Helper functions

// clampAttribute ensures an attribute value is within valid range (3-25)
func clampAttribute(value int) int {
	if value < 3 {
		return 3
	}
	if value > 25 {
		return 25
	}
	return value
}
