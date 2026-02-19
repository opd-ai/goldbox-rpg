package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/pcg"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// run executes the validator demo and returns any errors encountered.
func run() error {
	// Set up a logger for demonstration
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create a content validator
	validator := pcg.NewContentValidator(logger)

	fmt.Println("=== PCG Content Validator Demonstration ===")
	fmt.Println("")
	fmt.Println("This demonstration showcases the content validation system for procedural content generation (PCG) in the Gold Box RPG engine.")

	// Test 1: Valid character
	fmt.Println("1. Validating a valid character...")
	validChar := &game.Character{
		ID:           "demo_char_1",
		Name:         "Aria the Brave",
		Strength:     15,
		Dexterity:    14,
		Constitution: 13,
		Intelligence: 12,
		Wisdom:       11,
		Charisma:     16,
	}

	results, err := validator.ValidateContent(context.Background(), pcg.ContentTypeCharacters, validChar)
	if err != nil {
		return fmt.Errorf("validating character: %w", err)
	}

	fmt.Printf("Validation results: %d rules checked\n", len(results))
	for _, result := range results {
		if result.Passed {
			fmt.Printf("✓ PASS: %s\n", result.Message)
		} else {
			fmt.Printf("✗ FAIL: %s (Severity: %s)\n", result.Message, result.Severity)
		}
	}

	// Test 2: Invalid character that gets fixed
	fmt.Println("\n2. Validating and fixing an invalid character...")
	invalidChar := &game.Character{
		ID:           "demo_char_2",
		Name:         "", // Empty name - will be fixed
		Strength:     50, // Too high - will be fixed
		Dexterity:    1,  // Too low - will be fixed
		Constitution: 13,
		Intelligence: 12,
		Wisdom:       11,
		Charisma:     10,
	}

	fixedChar, results, err := validator.ValidateAndFix(context.Background(), pcg.ContentTypeCharacters, invalidChar)
	if err != nil {
		return fmt.Errorf("validating and fixing character: %w", err)
	}

	fmt.Printf("Original character: Name='%s', Strength=%d, Dexterity=%d\n",
		invalidChar.Name, invalidChar.Strength, invalidChar.Dexterity)

	fixedCharTyped := fixedChar.(*game.Character)
	fmt.Printf("Fixed character: Name='%s', Strength=%d, Dexterity=%d\n",
		fixedCharTyped.Name, fixedCharTyped.Strength, fixedCharTyped.Dexterity)

	// Test 3: Quest validation
	fmt.Println("\n3. Validating a quest with missing objectives...")
	invalidQuest := &game.Quest{
		ID:         "demo_quest_1",
		Title:      "The Lost Treasure",
		Objectives: []game.QuestObjective{}, // No objectives - will be fixed
	}

	fixedQuest, results, err := validator.ValidateAndFix(context.Background(), pcg.ContentTypeQuests, invalidQuest)
	if err != nil {
		return fmt.Errorf("validating and fixing quest: %w", err)
	}

	fmt.Printf("Original quest objectives: %d\n", len(invalidQuest.Objectives))
	fixedQuestTyped := fixedQuest.(*game.Quest)
	fmt.Printf("Fixed quest objectives: %d\n", len(fixedQuestTyped.Objectives))
	if len(fixedQuestTyped.Objectives) > 0 {
		fmt.Printf("Default objective: %s\n", fixedQuestTyped.Objectives[0].Description)
	}

	// Test 4: Validation metrics
	fmt.Println("\n4. Validation metrics...")
	metrics := validator.GetValidationMetrics()
	fmt.Printf("Total validations: %d\n", metrics.GetTotalValidations())
	fmt.Printf("Success rate: %.1f%%\n", metrics.GetSuccessRate())
	fmt.Printf("Average validation time: %v\n", metrics.GetAverageValidationTime())
	fmt.Printf("Critical failure rate: %.1f%%\n", metrics.GetCriticalFailureRate())

	fmt.Println("\n=== Demonstration Complete ===")

	return nil
}
