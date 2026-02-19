package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/pcg"
)

// Config holds the command-line configuration for the validator demo.
type Config struct {
	// Timeout specifies the maximum duration for validation operations.
	Timeout time.Duration
}

// parseFlags parses command-line flags and returns the configuration.
func parseFlags() *Config {
	cfg := &Config{}
	flag.DurationVar(&cfg.Timeout, "timeout", 30*time.Second, "timeout for validation operations")
	flag.Parse()
	return cfg
}

// main is the entry point for the validator demo application.
// It executes the run() function which demonstrates the PCG content
// validation system including character validation, automatic fixing
// of invalid content, quest validation, and metrics collection.
// On any error, it prints to stderr and exits with status code 1.
func main() {
	cfg := parseFlags()
	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// run executes the validator demo and returns any errors encountered.
// It uses the provided configuration to set up context timeout for all
// validation operations.
func run(cfg *Config) error {
	// Create context with timeout for all validation operations
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	// Set up a logger for demonstration
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create a content validator
	validator := pcg.NewContentValidator(logger)

	fmt.Println("=== PCG Content Validator Demonstration ===")
	fmt.Println("")
	fmt.Printf("This demonstration showcases the content validation system for procedural content generation (PCG) in the Gold Box RPG engine.\n")
	fmt.Printf("Using timeout: %v\n", cfg.Timeout)

	// Test 1: Valid character
	fmt.Println("\n1. Validating a valid character...")
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

	results, err := validator.ValidateContent(ctx, pcg.ContentTypeCharacters, validChar)
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

	fixedChar, results, err := validator.ValidateAndFix(ctx, pcg.ContentTypeCharacters, invalidChar)
	if err != nil {
		return fmt.Errorf("validating and fixing character: %w", err)
	}

	fmt.Printf("Original character: Name='%s', Strength=%d, Dexterity=%d\n",
		invalidChar.Name, invalidChar.Strength, invalidChar.Dexterity)

	fixedCharTyped, ok := fixedChar.(*game.Character)
	if !ok {
		return fmt.Errorf("unexpected type returned from ValidateAndFix: expected *game.Character, got %T", fixedChar)
	}
	fmt.Printf("Fixed character: Name='%s', Strength=%d, Dexterity=%d\n",
		fixedCharTyped.Name, fixedCharTyped.Strength, fixedCharTyped.Dexterity)

	// Test 3: Quest validation
	fmt.Println("\n3. Validating a quest with missing objectives...")
	invalidQuest := &game.Quest{
		ID:         "demo_quest_1",
		Title:      "The Lost Treasure",
		Objectives: []game.QuestObjective{}, // No objectives - will be fixed
	}

	fixedQuest, results, err := validator.ValidateAndFix(ctx, pcg.ContentTypeQuests, invalidQuest)
	if err != nil {
		return fmt.Errorf("validating and fixing quest: %w", err)
	}

	fmt.Printf("Original quest objectives: %d\n", len(invalidQuest.Objectives))
	fixedQuestTyped, ok := fixedQuest.(*game.Quest)
	if !ok {
		return fmt.Errorf("unexpected type returned from ValidateAndFix: expected *game.Quest, got %T", fixedQuest)
	}
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
