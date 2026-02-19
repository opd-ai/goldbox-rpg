package main

import (
	"errors"
	"flag"
	"fmt"
	"time"

	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/pcg"

	"github.com/sirupsen/logrus"
)

// Config holds the command-line configuration for the demo.
type Config struct {
	Seed int64
}

// parseFlags parses command-line flags and returns the configuration.
// This function is exported for testing purposes.
func parseFlags() *Config {
	seed := flag.Int64("seed", 42, "Random seed for deterministic demo (default: 42)")
	flag.Parse()
	return &Config{Seed: *seed}
}

// ErrNilWorld is returned when attempting to initialize PCG with a nil world.
var ErrNilWorld = errors.New("world cannot be nil for PCG initialization")

// demoContext holds shared state for demo execution.
type demoContext struct {
	logger         *logrus.Logger
	pcgManager     *pcg.PCGManager
	qualityMetrics *pcg.ContentQualityMetrics
}

// initializePCG creates and initializes the PCG manager with the world.
// Returns an error if the world is nil.
func initializePCG(world *game.World, logger *logrus.Logger, seed int64) (*demoContext, error) {
	if world == nil {
		return nil, ErrNilWorld
	}

	pcgManager := pcg.NewPCGManager(world, logger)
	pcgManager.InitializeWithSeed(seed)
	qualityMetrics := pcgManager.GetQualityMetrics()

	return &demoContext{
		logger:         logger,
		pcgManager:     pcgManager,
		qualityMetrics: qualityMetrics,
	}, nil
}

// demonstrateTerrainGeneration simulates terrain generation with quality tracking.
func demonstrateTerrainGeneration(ctx *demoContext) {
	for i := 0; i < 5; i++ {
		terrainContent := fmt.Sprintf("terrain_level_%d", i)
		duration := time.Duration(50+i*10) * time.Millisecond
		ctx.qualityMetrics.RecordContentGeneration(pcg.ContentTypeTerrain, terrainContent, duration, nil)
		fmt.Printf("   Generated terrain level %d in %v\n", i+1, duration)
	}
}

// demonstrateQuestGeneration simulates quest generation with some failures.
func demonstrateQuestGeneration(ctx *demoContext) {
	for i := 0; i < 8; i++ {
		questContent := fmt.Sprintf("quest_%d", i)
		duration := time.Duration(80+i*15) * time.Millisecond

		var err error
		if i == 3 || i == 6 { // Simulate some generation failures
			err = fmt.Errorf("generation failed for quest %d", i)
		}

		ctx.qualityMetrics.RecordContentGeneration(pcg.ContentTypeQuests, questContent, duration, err)

		if err != nil {
			fmt.Printf("   Quest %d generation failed: %v\n", i+1, err)
		} else {
			fmt.Printf("   Generated quest %d in %v\n", i+1, duration)
		}
	}
}

// demonstrateItemGeneration simulates item set generation.
func demonstrateItemGeneration(ctx *demoContext) {
	for i := 0; i < 3; i++ {
		itemContent := fmt.Sprintf("item_set_%d", i)
		duration := time.Duration(30+i*5) * time.Millisecond
		ctx.qualityMetrics.RecordContentGeneration(pcg.ContentTypeItems, itemContent, duration, nil)
		fmt.Printf("   Generated item set %d in %v\n", i+1, duration)
	}
}

// demonstratePlayerFeedback records sample player feedback.
func demonstratePlayerFeedback(ctx *demoContext) {
	feedbacks := []pcg.PlayerFeedback{
		{
			Timestamp:   time.Now(),
			ContentType: pcg.ContentTypeQuests,
			ContentID:   "quest_0",
			Rating:      5,
			Difficulty:  3,
			Enjoyment:   5,
			Comments:    "Excellent quest design!",
			SessionID:   "session_1",
		},
		{
			Timestamp:   time.Now(),
			ContentType: pcg.ContentTypeQuests,
			ContentID:   "quest_1",
			Rating:      4,
			Difficulty:  4,
			Enjoyment:   4,
			Comments:    "Good challenge level",
			SessionID:   "session_1",
		},
		{
			Timestamp:   time.Now(),
			ContentType: pcg.ContentTypeTerrain,
			ContentID:   "terrain_level_0",
			Rating:      3,
			Difficulty:  2,
			Enjoyment:   3,
			Comments:    "Terrain was okay, could be more varied",
			SessionID:   "session_2",
		},
		{
			Timestamp:   time.Now(),
			ContentType: pcg.ContentTypeQuests,
			ContentID:   "quest_4",
			Rating:      2,
			Difficulty:  5,
			Enjoyment:   2,
			Comments:    "Too difficult, frustrating",
			SessionID:   "session_3",
		},
	}

	for _, feedback := range feedbacks {
		ctx.pcgManager.RecordPlayerFeedback(feedback)
		fmt.Printf("   Recorded feedback for %s: Rating %d/5, Enjoyment %d/5\n",
			feedback.ContentID, feedback.Rating, feedback.Enjoyment)
	}
}

// questCompletion represents quest completion tracking data.
type questCompletion struct {
	questID        string
	completionTime time.Duration
	completed      bool
}

// demonstrateQuestCompletions records quest completion tracking data.
func demonstrateQuestCompletions(ctx *demoContext) {
	completions := []questCompletion{
		{"quest_0", 15 * time.Minute, true},
		{"quest_1", 22 * time.Minute, true},
		{"quest_2", 8 * time.Minute, false}, // Abandoned
		{"quest_4", 45 * time.Minute, true},
		{"quest_5", 12 * time.Minute, false}, // Abandoned
	}

	for _, c := range completions {
		ctx.pcgManager.RecordQuestCompletion(c.questID, c.completionTime, c.completed)
		if c.completed {
			fmt.Printf("   Quest %s completed in %v\n", c.questID, c.completionTime)
		} else {
			fmt.Printf("   Quest %s abandoned after %v\n", c.questID, c.completionTime)
		}
	}
}

// displayQualityReport generates and displays the comprehensive quality report.
func displayQualityReport(ctx *demoContext) {
	report := ctx.pcgManager.GenerateQualityReport()

	fmt.Printf("\n=== CONTENT QUALITY REPORT ===\n")
	fmt.Printf("Generated at: %v\n", report.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("Overall Quality Score: %.3f\n", report.OverallScore)
	fmt.Printf("Quality Grade: %s\n", report.QualityGrade)

	fmt.Printf("\nComponent Scores:\n")
	for component, score := range report.ComponentScores {
		fmt.Printf("  %s: %.3f\n", component, score)
	}

	fmt.Printf("\nThreshold Compliance:\n")
	for threshold, passed := range report.ThresholdStatus {
		status := "✓ PASS"
		if !passed {
			status = "✗ FAIL"
		}
		fmt.Printf("  %s: %s\n", threshold, status)
	}

	if len(report.Recommendations) > 0 {
		fmt.Printf("\nRecommendations:\n")
		for i, rec := range report.Recommendations {
			fmt.Printf("  %d. %s\n", i+1, rec)
		}
	}

	if len(report.CriticalIssues) > 0 {
		fmt.Printf("\nCritical Issues:\n")
		for i, issue := range report.CriticalIssues {
			fmt.Printf("  %d. %s\n", i+1, issue)
		}
	}

	fmt.Printf("\nSystem Summary:\n")
	for key, value := range report.SystemSummary {
		fmt.Printf("  %s: %v\n", key, value)
	}
}

// displayMetricsComponents shows individual metrics component details.
func displayMetricsComponents(ctx *demoContext) {
	performanceStats := ctx.qualityMetrics.GetPerformanceMetrics().GetStats()
	fmt.Printf("\nPerformance Metrics:\n")
	fmt.Printf("  Total Generations: %v\n", performanceStats["total_generations"])
	fmt.Printf("  Cache Hit Ratio: %.1f%%\n", ctx.qualityMetrics.GetPerformanceMetrics().GetCacheHitRatio())

	fmt.Printf("\nValidation Metrics:\n")
	fmt.Printf("  Validation system initialized and monitoring content\n")

	balanceMetrics := ctx.qualityMetrics.GetBalanceMetrics()
	fmt.Printf("\nBalance Metrics:\n")
	fmt.Printf("  System Health: %.3f\n", balanceMetrics.SystemHealth)
	fmt.Printf("  Total Balance Checks: %d\n", balanceMetrics.TotalBalanceChecks)
}

// displayFinalAssessment shows the overall quality assessment.
func displayFinalAssessment(ctx *demoContext) {
	overallScore := ctx.pcgManager.GetOverallQualityScore()
	fmt.Printf("\n=== FINAL QUALITY ASSESSMENT ===\n")
	fmt.Printf("Overall Quality Score: %.3f\n", overallScore)

	switch {
	case overallScore >= 0.9:
		fmt.Printf("Quality Status: EXCELLENT - Content generation is performing exceptionally well\n")
	case overallScore >= 0.8:
		fmt.Printf("Quality Status: GOOD - Content generation is performing well with minor areas for improvement\n")
	case overallScore >= 0.7:
		fmt.Printf("Quality Status: ACCEPTABLE - Content generation is adequate but has room for improvement\n")
	case overallScore >= 0.6:
		fmt.Printf("Quality Status: NEEDS IMPROVEMENT - Content generation requires attention\n")
	default:
		fmt.Printf("Quality Status: CRITICAL - Content generation requires immediate attention\n")
	}
}

// displayDemoSummary shows what the metrics system tracks.
func displayDemoSummary() {
	fmt.Printf("\nDemo completed successfully! The metrics system is tracking:\n")
	fmt.Printf("- Content generation performance and errors\n")
	fmt.Printf("- Content variety and uniqueness\n")
	fmt.Printf("- Logical consistency validation\n")
	fmt.Printf("- Player engagement and satisfaction\n")
	fmt.Printf("- System stability and reliability\n")
	fmt.Printf("- Comprehensive quality reporting with actionable insights\n")
}

// run executes the metrics demo with the provided configuration and returns any error.
// If cfg is nil, it parses command-line flags to get the configuration.
func run(cfg *Config) error {
	if cfg == nil {
		cfg = parseFlags()
	}

	fmt.Println("=== GoldBox RPG - Content Quality Metrics System Demo ===")
	fmt.Printf("Using seed: %d\n", cfg.Seed)
	fmt.Println()

	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create a minimal world for demonstration
	world := &game.World{
		Levels: make([]game.Level, 0),
	}

	// Initialize PCG Manager with quality metrics and error checking
	ctx, err := initializePCG(world, logger, cfg.Seed)
	if err != nil {
		return fmt.Errorf("failed to initialize PCG system: %w", err)
	}

	fmt.Println("1. Initializing Content Quality Metrics System...")

	// Demonstrate content generation with quality tracking
	fmt.Println("\n2. Generating Content with Quality Tracking...")
	demonstrateTerrainGeneration(ctx)
	demonstrateQuestGeneration(ctx)
	demonstrateItemGeneration(ctx)

	// Demonstrate player feedback recording
	fmt.Println("\n3. Recording Player Feedback...")
	demonstratePlayerFeedback(ctx)

	// Demonstrate quest completion tracking
	fmt.Println("\n4. Recording Quest Completions...")
	demonstrateQuestCompletions(ctx)

	// Generate comprehensive quality report
	fmt.Println("\n5. Generating Quality Report...")
	displayQualityReport(ctx)

	// Show individual metrics components
	fmt.Printf("\n6. Individual Metrics Components:\n")
	displayMetricsComponents(ctx)

	// Final assessment
	displayFinalAssessment(ctx)
	displayDemoSummary()

	return nil
}

// main is the entry point for the metrics demo application.
// It parses command-line flags and runs the demo.
func main() {
	if err := run(nil); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
