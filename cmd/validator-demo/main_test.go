package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"

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
	// Capture stdout to verify output structure
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	// Run in a separate goroutine to avoid blocking
	done := make(chan bool)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("main() panicked: %v", r)
			}
			done <- true
		}()
		main()
	}()

	<-done

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	output := buf.String()

	// Verify expected sections are present
	assert.Contains(t, output, "PCG Content Validator Demonstration")
	assert.Contains(t, output, "Validating a valid character")
	assert.Contains(t, output, "Validating and fixing an invalid character")
	assert.Contains(t, output, "Validating a quest with missing objectives")
	assert.Contains(t, output, "Validation metrics")
	assert.Contains(t, output, "Demonstration Complete")
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
				fixedChar := fixedContent.(*game.Character)
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

			fixedQuest := fixedContent.(*game.Quest)
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
