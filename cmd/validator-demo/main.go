package main

import (
	"context"
	"flag"
	"fmt"
	"io"
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
	// Verbose enables verbose logging to demonstrate validation logging behavior.
	Verbose bool
}

// parseFlags parses command-line flags and returns the configuration.
func parseFlags() *Config {
	cfg := &Config{}
	flag.DurationVar(&cfg.Timeout, "timeout", 30*time.Second, "timeout for validation operations")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "enable verbose logging to demonstrate validation logging")
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
	if err := run(cfg, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// printSection prints a formatted section header.
func printSection(w io.Writer, number int, title string) {
	fmt.Fprintf(w, "\n%d. %s\n", number, title)
}

// printResult prints a single validation result with consistent formatting.
func printResult(w io.Writer, result pcg.Result) {
	if result.Passed {
		fmt.Fprintf(w, "   ✓ PASS: %s\n", result.Message)
	} else {
		fmt.Fprintf(w, "   ✗ FAIL: %s (Severity: %s)\n", result.Message, result.Severity)
	}
}

// printKV prints a key-value pair with consistent formatting.
func printKV(w io.Writer, key string, value interface{}) {
	fmt.Fprintf(w, "   %s: %v\n", key, value)
}

// run executes the validator demo and returns any errors encountered.
// It uses the provided configuration to set up context timeout for all
// validation operations.
func run(cfg *Config, w io.Writer) error {
	// Create context with timeout for all validation operations
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	// Set up a logger for demonstration
	logger := logrus.New()
	if cfg.Verbose {
		logger.SetLevel(logrus.DebugLevel)
		logger.SetOutput(w)
		logger.SetFormatter(&logrus.TextFormatter{
			ForceColors:   false,
			FullTimestamp: false,
		})
	} else {
		logger.SetLevel(logrus.WarnLevel)
		logger.SetOutput(w)
	}

	// Create a content validator
	validator := pcg.NewContentValidator(logger)

	fmt.Fprintln(w, "=== PCG Content Validator Demonstration ===")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "This demonstration showcases the content validation system for")
	fmt.Fprintln(w, "procedural content generation (PCG) in the Gold Box RPG engine.")
	fmt.Fprintln(w, "")
	printKV(w, "Timeout", cfg.Timeout)
	printKV(w, "Verbose logging", cfg.Verbose)

	if cfg.Verbose {
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "   Note: Verbose mode enabled - validation logs will appear below.")
	}

	// Test 1: Valid character
	printSection(w, 1, "Validating a valid character...")
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

	printKV(w, "Rules checked", len(results))
	for _, result := range results {
		printResult(w, result)
	}

	// Test 2: Invalid character that gets fixed
	printSection(w, 2, "Validating and fixing an invalid character...")
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

	printKV(w, "Original Name", fmt.Sprintf("'%s'", invalidChar.Name))
	printKV(w, "Original Strength", invalidChar.Strength)
	printKV(w, "Original Dexterity", invalidChar.Dexterity)

	fixedCharTyped, ok := fixedChar.(*game.Character)
	if !ok {
		return fmt.Errorf("unexpected type returned from ValidateAndFix: expected *game.Character, got %T", fixedChar)
	}
	printKV(w, "Fixed Name", fmt.Sprintf("'%s'", fixedCharTyped.Name))
	printKV(w, "Fixed Strength", fixedCharTyped.Strength)
	printKV(w, "Fixed Dexterity", fixedCharTyped.Dexterity)

	// Test 3: Quest validation
	printSection(w, 3, "Validating a quest with missing objectives...")
	invalidQuest := &game.Quest{
		ID:         "demo_quest_1",
		Title:      "The Lost Treasure",
		Objectives: []game.QuestObjective{}, // No objectives - will be fixed
	}

	fixedQuest, results, err := validator.ValidateAndFix(ctx, pcg.ContentTypeQuests, invalidQuest)
	if err != nil {
		return fmt.Errorf("validating and fixing quest: %w", err)
	}

	printKV(w, "Original objectives", len(invalidQuest.Objectives))
	fixedQuestTyped, ok := fixedQuest.(*game.Quest)
	if !ok {
		return fmt.Errorf("unexpected type returned from ValidateAndFix: expected *game.Quest, got %T", fixedQuest)
	}
	printKV(w, "Fixed objectives", len(fixedQuestTyped.Objectives))
	if len(fixedQuestTyped.Objectives) > 0 {
		printKV(w, "Default objective", fixedQuestTyped.Objectives[0].Description)
	}

	// Test 4: Validation metrics
	printSection(w, 4, "Validation metrics...")
	metrics := validator.GetValidationMetrics()
	printKV(w, "Total validations", metrics.GetTotalValidations())
	printKV(w, "Success rate", fmt.Sprintf("%.1f%%", metrics.GetSuccessRate()))
	printKV(w, "Average validation time", metrics.GetAverageValidationTime())
	printKV(w, "Critical failure rate", fmt.Sprintf("%.1f%%", metrics.GetCriticalFailureRate()))

	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "=== Demonstration Complete ===")

	return nil
}
