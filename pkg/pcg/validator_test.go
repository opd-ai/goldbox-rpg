package pcg

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goldbox-rpg/pkg/game"
)

func TestNewContentValidator(t *testing.T) {
	tests := []struct {
		name   string
		logger *logrus.Logger
	}{
		{
			name:   "with_logger",
			logger: logrus.New(),
		},
		{
			name:   "with_nil_logger",
			logger: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewContentValidator(tt.logger)

			assert.NotNil(t, validator)
			assert.NotNil(t, validator.logger)
			assert.NotNil(t, validator.validationRules)
			assert.NotNil(t, validator.fallbackHandlers)
			assert.NotNil(t, validator.metrics)

			// Should have rules for different content types
			assert.Greater(t, len(validator.validationRules[ContentTypeCharacters]), 0)
			assert.Greater(t, len(validator.validationRules[ContentTypeQuests]), 0)
			assert.Greater(t, len(validator.validationRules[ContentTypeDungeon]), 0)
			assert.Greater(t, len(validator.validationRules[ContentTypeDialogue]), 0)
			assert.Greater(t, len(validator.validationRules[ContentTypeFactions]), 0)
		})
	}
}

func TestContentValidator_ValidateContent(t *testing.T) {
	validator := NewContentValidator(nil)

	tests := []struct {
		name        string
		contentType ContentType
		content     interface{}
		expectError bool
		expectRules int
	}{
		{
			name:        "valid_character",
			contentType: ContentTypeCharacters,
			content: &game.Character{
				ID:           "test_char_1",
				Name:         "Test Character",
				Strength:     15,
				Dexterity:    14,
				Constitution: 13,
				Intelligence: 12,
				Wisdom:       11,
				Charisma:     10,
			},
			expectError: false,
			expectRules: 2, // character_attributes_valid, character_name_not_empty
		},
		{
			name:        "invalid_character_attributes",
			contentType: ContentTypeCharacters,
			content: &game.Character{
				ID:           "test_char_2",
				Name:         "Invalid Character",
				Strength:     50, // Too high
				Dexterity:    1,  // Too low
				Constitution: 13,
				Intelligence: 12,
				Wisdom:       11,
				Charisma:     10,
			},
			expectError: false,
			expectRules: 2,
		},
		{
			name:        "character_empty_name",
			contentType: ContentTypeCharacters,
			content: &game.Character{
				ID:           "test_char_3",
				Name:         "",
				Strength:     15,
				Dexterity:    14,
				Constitution: 13,
				Intelligence: 12,
				Wisdom:       11,
				Charisma:     10,
			},
			expectError: false,
			expectRules: 2,
		},
		{
			name:        "valid_quest",
			contentType: ContentTypeQuests,
			content: &game.Quest{
				ID:    "test_quest_1",
				Title: "Test Quest",
				Objectives: []game.QuestObjective{
					{
						Description: "Find the treasure",
						Progress:    0,
						Required:    1,
						Completed:   false,
					},
				},
			},
			expectError: false,
			expectRules: 2, // quest_has_objectives, quest_has_title
		},
		{
			name:        "quest_no_objectives",
			contentType: ContentTypeQuests,
			content: &game.Quest{
				ID:         "test_quest_2",
				Title:      "Empty Quest",
				Objectives: []game.QuestObjective{},
			},
			expectError: false,
			expectRules: 2,
		},
		{
			name:        "unknown_content_type",
			contentType: "unknown",
			content:     &game.Character{},
			expectError: false,
			expectRules: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			results, err := validator.ValidateContent(ctx, tt.contentType, tt.content)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, results, tt.expectRules)
			}
		})
	}
}

func TestContentValidator_ValidateAndFix(t *testing.T) {
	validator := NewContentValidator(nil)

	tests := []struct {
		name            string
		contentType     ContentType
		content         interface{}
		expectFixed     bool
		expectError     bool
		validateResults func(t *testing.T, fixedContent interface{}, results []Result)
	}{
		{
			name:        "fix_character_attributes",
			contentType: ContentTypeCharacters,
			content: &game.Character{
				ID:           "test_char_fix",
				Name:         "Test Character",
				Strength:     50, // Too high
				Dexterity:    1,  // Too low
				Constitution: 13,
				Intelligence: 12,
				Wisdom:       11,
				Charisma:     10,
			},
			expectFixed: true,
			expectError: false,
			validateResults: func(t *testing.T, fixedContent interface{}, results []Result) {
				char, ok := fixedContent.(*game.Character)
				require.True(t, ok)
				assert.LessOrEqual(t, char.Strength, 25)
				assert.GreaterOrEqual(t, char.Dexterity, 3)
			},
		},
		{
			name:        "fix_character_empty_name",
			contentType: ContentTypeCharacters,
			content: &game.Character{
				ID:           "test_char_name",
				Name:         "",
				Strength:     15,
				Dexterity:    14,
				Constitution: 13,
				Intelligence: 12,
				Wisdom:       11,
				Charisma:     10,
			},
			expectFixed: true,
			expectError: false,
			validateResults: func(t *testing.T, fixedContent interface{}, results []Result) {
				char, ok := fixedContent.(*game.Character)
				require.True(t, ok)
				assert.NotEmpty(t, char.Name)
				assert.Contains(t, char.Name, "Character_")
			},
		},
		{
			name:        "fix_quest_no_objectives",
			contentType: ContentTypeQuests,
			content: &game.Quest{
				ID:         "test_quest_fix",
				Title:      "Test Quest",
				Objectives: []game.QuestObjective{},
			},
			expectFixed: true,
			expectError: false,
			validateResults: func(t *testing.T, fixedContent interface{}, results []Result) {
				quest, ok := fixedContent.(*game.Quest)
				require.True(t, ok)
				assert.Len(t, quest.Objectives, 1)
				assert.Equal(t, "Complete the quest", quest.Objectives[0].Description)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			fixedContent, results, err := validator.ValidateAndFix(ctx, tt.contentType, tt.content)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, fixedContent)

			if tt.validateResults != nil {
				tt.validateResults(t, fixedContent, results)
			}
		})
	}
}

func TestContentValidator_RegisterValidationRule(t *testing.T) {
	validator := NewContentValidator(nil)

	customRule := ValidationRule{
		Name:        "custom_test_rule",
		Description: "Test rule for validation",
		Severity:    SeverityWarning,
		Validator: func(content interface{}) Result {
			return Result{Passed: true, Message: "custom rule passed"}
		},
	}

	initialRules := len(validator.validationRules[ContentTypeCharacters])
	validator.RegisterValidationRule(ContentTypeCharacters, customRule)

	assert.Len(t, validator.validationRules[ContentTypeCharacters], initialRules+1)

	// Test the custom rule is actually used
	ctx := context.Background()
	char := &game.Character{
		ID:           "test_char",
		Name:         "Test Character",
		Strength:     15,
		Dexterity:    14,
		Constitution: 13,
		Intelligence: 12,
		Wisdom:       11,
		Charisma:     10,
	}

	results, err := validator.ValidateContent(ctx, ContentTypeCharacters, char)
	assert.NoError(t, err)
	assert.Len(t, results, initialRules+1)

	// Find our custom rule result
	found := false
	for _, result := range results {
		if strings.Contains(result.Message, "custom rule passed") {
			found = true
			break
		}
	}
	assert.True(t, found, "Custom rule should have been executed")
}

func TestContentValidator_RegisterFallbackHandler(t *testing.T) {
	validator := NewContentValidator(nil)

	// Create a mock fallback handler
	mockHandler := &mockFallbackHandler{}

	validator.RegisterFallbackHandler(ContentTypeCharacters, mockHandler)

	// Verify the handler was registered
	handler, exists := validator.fallbackHandlers[ContentTypeCharacters]
	assert.True(t, exists)
	assert.Equal(t, mockHandler, handler)
}

func TestValidationMetrics(t *testing.T) {
	metrics := NewValidationMetrics()

	// Test initial state
	assert.Equal(t, int64(0), metrics.totalValidations)
	assert.Equal(t, 0.0, metrics.GetSuccessRate())
	assert.Equal(t, time.Duration(0), metrics.GetAverageValidationTime())

	// Record some validations
	passedResult := []Result{{Passed: true, Severity: SeverityInfo}}
	failedResult := []Result{{Passed: false, Severity: SeverityError}}
	criticalResult := []Result{{Passed: false, Severity: SeverityCritical}}

	metrics.recordValidation(passedResult, time.Millisecond*10)
	metrics.recordValidation(failedResult, time.Millisecond*15)
	metrics.recordValidation(criticalResult, time.Millisecond*20)

	// Test metrics
	assert.Equal(t, int64(3), metrics.totalValidations)
	assert.Equal(t, int64(1), metrics.passedValidations)
	assert.Equal(t, int64(2), metrics.failedValidations)
	assert.Equal(t, int64(1), metrics.criticalFailures)

	// Test calculated metrics
	successRate := metrics.GetSuccessRate()
	assert.InDelta(t, 33.33, successRate, 0.1)

	criticalRate := metrics.GetCriticalFailureRate()
	assert.InDelta(t, 33.33, criticalRate, 0.1)

	avgTime := metrics.GetAverageValidationTime()
	assert.Equal(t, time.Millisecond*15, avgTime)

	// Test rule execution tracking
	metrics.recordRuleExecution("test_rule")
	metrics.recordRuleExecution("test_rule")
	metrics.recordRuleExecution("other_rule")

	stats := metrics.getStats()
	assert.Equal(t, int64(2), stats.ruleExecutionCounts["test_rule"])
	assert.Equal(t, int64(1), stats.ruleExecutionCounts["other_rule"])

	// Test fallback tracking
	metrics.recordFallback()
	assert.Equal(t, int64(1), metrics.fallbacksTriggered)

	// Test reset
	metrics.reset()
	assert.Equal(t, int64(0), metrics.totalValidations)
	assert.Equal(t, 0.0, metrics.GetSuccessRate())
}

func TestContextCancellation(t *testing.T) {
	validator := NewContentValidator(nil)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	cancel() // Cancel immediately

	char := &game.Character{
		ID:           "test_char",
		Name:         "Test Character",
		Strength:     15,
		Dexterity:    14,
		Constitution: 13,
		Intelligence: 12,
		Wisdom:       11,
		Charisma:     10,
	}

	results, err := validator.ValidateContent(ctx, ContentTypeCharacters, char)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.Nil(t, results)
}

func TestFallbackHandlers(t *testing.T) {
	t.Run("characterFallbackHandler", func(t *testing.T) {
		handler := &characterFallbackHandler{logger: logrus.New()}

		// Test CanHandle
		attributeResult := Result{Message: "character has invalid attributes"}
		nameResult := Result{Message: "character has empty name"}
		otherResult := Result{Message: "some other issue"}

		assert.True(t, handler.CanHandle(attributeResult))
		assert.True(t, handler.CanHandle(nameResult))
		assert.False(t, handler.CanHandle(otherResult))

		// Test Handle with attribute fix
		char := &game.Character{
			ID:           "test",
			Name:         "Test",
			Strength:     50, // Too high
			Dexterity:    1,  // Too low
			Constitution: 13,
			Intelligence: 12,
			Wisdom:       11,
			Charisma:     10,
		}

		fixed, err := handler.Handle(context.Background(), char, attributeResult)
		assert.NoError(t, err)

		fixedChar, ok := fixed.(*game.Character)
		require.True(t, ok)
		assert.LessOrEqual(t, fixedChar.Strength, 25)
		assert.GreaterOrEqual(t, fixedChar.Dexterity, 3)

		// Test Handle with name fix
		char.Name = ""
		fixed, err = handler.Handle(context.Background(), char, nameResult)
		assert.NoError(t, err)

		fixedChar, ok = fixed.(*game.Character)
		require.True(t, ok)
		assert.NotEmpty(t, fixedChar.Name)
	})

	t.Run("questFallbackHandler", func(t *testing.T) {
		handler := &questFallbackHandler{logger: logrus.New()}

		objectiveResult := Result{Message: "quest has no objectives"}
		titleResult := Result{Message: "quest has empty title"}

		assert.True(t, handler.CanHandle(objectiveResult))
		assert.True(t, handler.CanHandle(titleResult))

		quest := &game.Quest{
			ID:         "test_quest",
			Title:      "",
			Objectives: []game.QuestObjective{},
		}

		// Test objective fix
		fixed, err := handler.Handle(context.Background(), quest, objectiveResult)
		assert.NoError(t, err)

		fixedQuest, ok := fixed.(*game.Quest)
		require.True(t, ok)
		assert.Len(t, fixedQuest.Objectives, 1)

		// Test title fix
		fixed, err = handler.Handle(context.Background(), quest, titleResult)
		assert.NoError(t, err)

		fixedQuest, ok = fixed.(*game.Quest)
		require.True(t, ok)
		assert.NotEmpty(t, fixedQuest.Title)
	})
}

func TestClampAttribute(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{input: 1, expected: 3},   // Too low
		{input: 3, expected: 3},   // Min valid
		{input: 15, expected: 15}, // Normal
		{input: 25, expected: 25}, // Max valid
		{input: 30, expected: 25}, // Too high
	}

	for _, tt := range tests {
		result := clampAttribute(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

// Mock fallback handler for testing
type mockFallbackHandler struct{}

func (m *mockFallbackHandler) CanHandle(result Result) bool {
	return true
}

func (m *mockFallbackHandler) Handle(ctx context.Context, content interface{}, result Result) (interface{}, error) {
	return content, nil
}

func (m *mockFallbackHandler) GetDescription() string {
	return "Mock fallback handler for testing"
}
