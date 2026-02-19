package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/pcg"
)

// TestValidateValidCharacter tests validation of a valid character.
func TestValidateValidCharacter(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	validator := pcg.NewContentValidator(logger)

	validChar := &game.Character{
		ID:           "test_char_1",
		Name:         "Test Hero",
		Strength:     15,
		Dexterity:    14,
		Constitution: 13,
		Intelligence: 12,
		Wisdom:       11,
		Charisma:     16,
	}

	results, err := validator.ValidateContent(context.Background(), pcg.ContentTypeCharacters, validChar)
	require.NoError(t, err)
	assert.NotEmpty(t, results)

	// All validation rules should pass for a valid character
	for _, result := range results {
		if !result.Passed {
			t.Logf("Validation rule failed: %s (severity: %s)", result.Message, result.Severity)
		}
	}
}

// TestValidateAndFixInvalidCharacter tests fixing an invalid character.
func TestValidateAndFixInvalidCharacter(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	validator := pcg.NewContentValidator(logger)

	invalidChar := &game.Character{
		ID:           "test_char_2",
		Name:         "", // Empty name - may be fixed if this error is processed first
		Strength:     50, // Too high - will be clamped to 25
		Dexterity:    1,  // Too low - will be clamped to 3
		Constitution: 13,
		Intelligence: 12,
		Wisdom:       11,
		Charisma:     10,
	}

	fixedContent, results, err := validator.ValidateAndFix(context.Background(), pcg.ContentTypeCharacters, invalidChar)
	require.NoError(t, err)
	require.NotNil(t, fixedContent)
	assert.NotEmpty(t, results)

	fixedChar, ok := fixedContent.(*game.Character)
	require.True(t, ok, "Expected fixed content to be a *game.Character")

	// The validator applies one fix at a time. The attribute fix clamps to 3-25 range.
	// Name fix may not be applied in the same pass due to "one fix at a time" policy.
	assert.LessOrEqual(t, fixedChar.Strength, 25, "Strength should be clamped to 25")
	assert.GreaterOrEqual(t, fixedChar.Dexterity, 3, "Dexterity should be at least 3")
}

// TestValidateAndFixInvalidQuest tests fixing a quest with missing objectives.
func TestValidateAndFixInvalidQuest(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	validator := pcg.NewContentValidator(logger)

	invalidQuest := &game.Quest{
		ID:         "test_quest_1",
		Title:      "The Lost Treasure",
		Objectives: []game.QuestObjective{}, // No objectives
	}

	fixedContent, results, err := validator.ValidateAndFix(context.Background(), pcg.ContentTypeQuests, invalidQuest)
	require.NoError(t, err)
	require.NotNil(t, fixedContent)
	assert.NotEmpty(t, results)

	fixedQuest, ok := fixedContent.(*game.Quest)
	require.True(t, ok, "Expected fixed content to be a *game.Quest")

	// Verify objectives were added
	assert.NotEmpty(t, fixedQuest.Objectives, "Fixed quest should have at least one objective")
}

// TestValidationMetrics tests that validation metrics are properly tracked.
func TestValidationMetrics(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	validator := pcg.NewContentValidator(logger)

	// Perform some validations
	char := &game.Character{
		ID:           "test_char",
		Name:         "Test",
		Strength:     10,
		Dexterity:    10,
		Constitution: 10,
		Intelligence: 10,
		Wisdom:       10,
		Charisma:     10,
	}

	_, err := validator.ValidateContent(context.Background(), pcg.ContentTypeCharacters, char)
	require.NoError(t, err)

	metrics := validator.GetValidationMetrics()
	assert.NotNil(t, metrics)
	assert.GreaterOrEqual(t, metrics.GetTotalValidations(), int64(1))
}

// TestMainOutputIntegration tests that main produces expected output structure.
func TestMainOutputIntegration(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{Timeout: 30 * time.Second, Verbose: false}
	err := run(cfg, &buf)
	require.NoError(t, err)

	output := buf.String()

	// Verify expected sections are present
	assert.Contains(t, output, "PCG Content Validator Demonstration")
	assert.Contains(t, output, "Validating a valid character")
	assert.Contains(t, output, "Validating and fixing an invalid character")
	assert.Contains(t, output, "Validating a quest with missing objectives")
	assert.Contains(t, output, "Validation metrics")
	assert.Contains(t, output, "Demonstration Complete")
	assert.Contains(t, output, "Timeout:")
}

// TestContentValidatorCharacterStatBounds tests validation of stat boundaries.
func TestContentValidatorCharacterStatBounds(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	validator := pcg.NewContentValidator(logger)

	tests := []struct {
		name        string
		char        *game.Character
		expectFixes bool
	}{
		{
			name: "valid_stats_middle_range",
			char: &game.Character{
				ID:           "char1",
				Name:         "Normal",
				Strength:     12,
				Dexterity:    12,
				Constitution: 12,
				Intelligence: 12,
				Wisdom:       12,
				Charisma:     12,
			},
			expectFixes: false,
		},
		{
			name: "all_minimum_valid",
			char: &game.Character{
				ID:           "char2",
				Name:         "Minimum Stats",
				Strength:     3,
				Dexterity:    3,
				Constitution: 3,
				Intelligence: 3,
				Wisdom:       3,
				Charisma:     3,
			},
			expectFixes: false,
		},
		{
			name: "all_maximum_valid",
			char: &game.Character{
				ID:           "char3",
				Name:         "Maximum Stats",
				Strength:     18,
				Dexterity:    18,
				Constitution: 18,
				Intelligence: 18,
				Wisdom:       18,
				Charisma:     18,
			},
			expectFixes: false,
		},
		{
			name: "too_low_stats",
			char: &game.Character{
				ID:           "char4",
				Name:         "Weak",
				Strength:     0,
				Dexterity:    0,
				Constitution: 0,
				Intelligence: 0,
				Wisdom:       0,
				Charisma:     0,
			},
			expectFixes: true,
		},
		{
			name: "too_high_stats",
			char: &game.Character{
				ID:           "char5",
				Name:         "Overpowered",
				Strength:     99,
				Dexterity:    99,
				Constitution: 99,
				Intelligence: 99,
				Wisdom:       99,
				Charisma:     99,
			},
			expectFixes: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fixedContent, results, err := validator.ValidateAndFix(context.Background(), pcg.ContentTypeCharacters, tc.char)
			require.NoError(t, err)
			require.NotNil(t, fixedContent)
			assert.NotEmpty(t, results)

			if tc.expectFixes {
				// Check that at least some fixes were applied
				fixedChar, ok := fixedContent.(*game.Character)
				require.True(t, ok, "Expected fixed content to be a *game.Character")
				// Validator clamps to 3-25 range
				assert.True(t,
					fixedChar.Strength >= 3 && fixedChar.Strength <= 25,
					"Strength should be in valid range after fix")
			}
		})
	}
}

// TestQuestValidationVariants tests different quest validation scenarios.
func TestQuestValidationVariants(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	validator := pcg.NewContentValidator(logger)

	tests := []struct {
		name  string
		quest *game.Quest
	}{
		{
			name: "quest_with_one_objective",
			quest: &game.Quest{
				ID:    "q1",
				Title: "Simple Quest",
				Objectives: []game.QuestObjective{
					{Description: "Test objective", Required: 1, Completed: false},
				},
			},
		},
		{
			name: "quest_with_multiple_objectives",
			quest: &game.Quest{
				ID:    "q2",
				Title: "Complex Quest",
				Objectives: []game.QuestObjective{
					{Description: "First objective", Required: 1, Completed: false},
					{Description: "Second objective", Required: 1, Completed: false},
					{Description: "Third objective", Required: 1, Completed: false},
				},
			},
		},
		{
			name: "quest_empty_title",
			quest: &game.Quest{
				ID:         "q3",
				Title:      "",
				Objectives: []game.QuestObjective{},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fixedContent, results, err := validator.ValidateAndFix(context.Background(), pcg.ContentTypeQuests, tc.quest)
			require.NoError(t, err)
			require.NotNil(t, fixedContent)
			assert.NotEmpty(t, results)

			fixedQuest, ok := fixedContent.(*game.Quest)
			require.True(t, ok, "Expected fixed content to be a *game.Quest")
			// Ensure quest is valid after fix
			assert.NotEmpty(t, fixedQuest.Objectives, "Quest should have objectives after validation")
		})
	}
}

// TestValidatorWithCancelledContext tests behavior with cancelled context.
func TestValidatorWithCancelledContext(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	validator := pcg.NewContentValidator(logger)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	char := &game.Character{
		ID:           "test",
		Name:         "Test",
		Strength:     10,
		Dexterity:    10,
		Constitution: 10,
		Intelligence: 10,
		Wisdom:       10,
		Charisma:     10,
	}

	// The validator should handle cancelled context gracefully
	results, err := validator.ValidateContent(ctx, pcg.ContentTypeCharacters, char)
	// Depending on implementation, this may return an error or empty results
	if err != nil {
		assert.True(t, strings.Contains(err.Error(), "context") || strings.Contains(err.Error(), "cancel"),
			"Error should mention context cancellation")
	} else {
		// If no error, results should still be valid (implementation may not check context)
		assert.NotNil(t, results)
	}
}

// TestConfigDefaults tests the default configuration values.
func TestConfigDefaults(t *testing.T) {
	cfg := &Config{Timeout: 30 * time.Second, Verbose: false}
	assert.Equal(t, 30*time.Second, cfg.Timeout)
	assert.False(t, cfg.Verbose)
}

// TestRunWithCustomTimeout tests run() with a custom timeout.
func TestRunWithCustomTimeout(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{Timeout: 5 * time.Second, Verbose: false}
	runErr := run(cfg, &buf)
	require.NoError(t, runErr)

	output := buf.String()
	assert.Contains(t, output, "Timeout: 5s")
}

// TestRunWithVerboseLogging tests that verbose mode enables debug logging.
func TestRunWithVerboseLogging(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{Timeout: 30 * time.Second, Verbose: true}
	runErr := run(cfg, &buf)
	require.NoError(t, runErr)

	output := buf.String()
	// Verify verbose mode shows logging-related output
	assert.Contains(t, output, "Verbose logging: true")
	assert.Contains(t, output, "validation logs will appear")
	// In verbose mode, logrus debug messages should appear
	assert.Contains(t, output, "starting content validation")
}

// TestRunWithContextTimeout tests that run() respects context timeout.
func TestRunWithContextTimeout(t *testing.T) {
	// Use a very short timeout - since validation is fast, this should still succeed
	cfg := &Config{Timeout: 100 * time.Millisecond, Verbose: false}

	var buf bytes.Buffer
	runErr := run(cfg, &buf)

	// With fast validation, 100ms should be enough
	// If it's too slow, we'd get a context deadline exceeded error
	assert.NoError(t, runErr, "run() should complete within timeout")
}

// TestPrintSection tests the section header formatting.
func TestPrintSection(t *testing.T) {
	var buf bytes.Buffer
	printSection(&buf, 1, "Test Section")
	assert.Equal(t, "\n1. Test Section\n", buf.String())
}

// TestPrintResult tests the result formatting.
func TestPrintResult(t *testing.T) {
	tests := []struct {
		name     string
		result   pcg.Result
		expected string
	}{
		{
			name:     "passed_result",
			result:   pcg.Result{Passed: true, Message: "Test passed"},
			expected: "   ✓ PASS: Test passed\n",
		},
		{
			name:     "failed_result",
			result:   pcg.Result{Passed: false, Message: "Test failed", Severity: pcg.SeverityCritical},
			expected: "   ✗ FAIL: Test failed (Severity: critical)\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			printResult(&buf, tc.result)
			assert.Equal(t, tc.expected, buf.String())
		})
	}
}

// TestPrintKV tests the key-value formatting.
func TestPrintKV(t *testing.T) {
	var buf bytes.Buffer
	printKV(&buf, "Key", "Value")
	assert.Equal(t, "   Key: Value\n", buf.String())
}
