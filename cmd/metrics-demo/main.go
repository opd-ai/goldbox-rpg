package main

import (
	"fmt"
	"time"

	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/pcg"

	"github.com/sirupsen/logrus"
)

// MetricsDemo demonstrates the comprehensive content quality metrics system
func main() {
	fmt.Println("=== GoldBox RPG - Content Quality Metrics System Demo ===")
	fmt.Println()

	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create a minimal world for demonstration
	world := &game.World{
		Levels: make([]game.Level, 0),
	}

	// Initialize PCG Manager with quality metrics
	pcgManager := pcg.NewPCGManager(world, logger)
	pcgManager.InitializeWithSeed(42) // Use fixed seed for deterministic demo

	fmt.Println("1. Initializing Content Quality Metrics System...")
	qualityMetrics := pcgManager.GetQualityMetrics()

	// Demonstrate content generation with quality tracking
	fmt.Println("\n2. Generating Content with Quality Tracking...")

	// Simulate terrain generation
	for i := 0; i < 5; i++ {
		terrainContent := fmt.Sprintf("terrain_level_%d", i)
		duration := time.Duration(50+i*10) * time.Millisecond

		// Record successful generation
		qualityMetrics.RecordContentGeneration(pcg.ContentTypeTerrain, terrainContent, duration, nil)

		fmt.Printf("   Generated terrain level %d in %v\n", i+1, duration)
	}

	// Simulate quest generation with some failures
	for i := 0; i < 8; i++ {
		questContent := fmt.Sprintf("quest_%d", i)
		duration := time.Duration(80+i*15) * time.Millisecond

		var err error
		if i == 3 || i == 6 { // Simulate some generation failures
			err = fmt.Errorf("generation failed for quest %d", i)
		}

		qualityMetrics.RecordContentGeneration(pcg.ContentTypeQuests, questContent, duration, err)

		if err != nil {
			fmt.Printf("   Quest %d generation failed: %v\n", i+1, err)
		} else {
			fmt.Printf("   Generated quest %d in %v\n", i+1, duration)
		}
	}

	// Simulate item generation
	for i := 0; i < 3; i++ {
		itemContent := fmt.Sprintf("item_set_%d", i)
		duration := time.Duration(30+i*5) * time.Millisecond
		qualityMetrics.RecordContentGeneration(pcg.ContentTypeItems, itemContent, duration, nil)
		fmt.Printf("   Generated item set %d in %v\n", i+1, duration)
	}

	// Demonstrate player feedback recording
	fmt.Println("\n3. Recording Player Feedback...")

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
		pcgManager.RecordPlayerFeedback(feedback)
		fmt.Printf("   Recorded feedback for %s: Rating %d/5, Enjoyment %d/5\n",
			feedback.ContentID, feedback.Rating, feedback.Enjoyment)
	}

	// Demonstrate quest completion tracking
	fmt.Println("\n4. Recording Quest Completions...")

	completions := []struct {
		questID        string
		completionTime time.Duration
		completed      bool
	}{
		{"quest_0", 15 * time.Minute, true},
		{"quest_1", 22 * time.Minute, true},
		{"quest_2", 8 * time.Minute, false}, // Abandoned
		{"quest_4", 45 * time.Minute, true},
		{"quest_5", 12 * time.Minute, false}, // Abandoned
	}

	for _, completion := range completions {
		pcgManager.RecordQuestCompletion(completion.questID, completion.completionTime, completion.completed)
		if completion.completed {
			fmt.Printf("   Quest %s completed in %v\n", completion.questID, completion.completionTime)
		} else {
			fmt.Printf("   Quest %s abandoned after %v\n", completion.questID, completion.completionTime)
		}
	}

	// Generate comprehensive quality report
	fmt.Println("\n5. Generating Quality Report...")

	report := pcgManager.GenerateQualityReport()

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

	// Display system summary
	fmt.Printf("\nSystem Summary:\n")
	for key, value := range report.SystemSummary {
		fmt.Printf("  %s: %v\n", key, value)
	}

	// Show individual metrics components
	fmt.Printf("\n6. Individual Metrics Components:\n")

	// Performance Metrics
	performanceStats := qualityMetrics.GetPerformanceMetrics().GetStats()
	fmt.Printf("\nPerformance Metrics:\n")
	fmt.Printf("  Total Generations: %v\n", performanceStats["total_generations"])
	fmt.Printf("  Cache Hit Ratio: %.1f%%\n", qualityMetrics.GetPerformanceMetrics().GetCacheHitRatio())

	// Validation Metrics
	fmt.Printf("\nValidation Metrics:\n")
	fmt.Printf("  Validation system initialized and monitoring content\n")

	// Balance Metrics
	balanceMetrics := qualityMetrics.GetBalanceMetrics()
	fmt.Printf("\nBalance Metrics:\n")
	fmt.Printf("  System Health: %.3f\n", balanceMetrics.SystemHealth)
	fmt.Printf("  Total Balance Checks: %d\n", balanceMetrics.TotalBalanceChecks)

	// Overall quality score
	overallScore := pcgManager.GetOverallQualityScore()
	fmt.Printf("\n=== FINAL QUALITY ASSESSMENT ===\n")
	fmt.Printf("Overall Quality Score: %.3f\n", overallScore)

	if overallScore >= 0.9 {
		fmt.Printf("Quality Status: EXCELLENT - Content generation is performing exceptionally well\n")
	} else if overallScore >= 0.8 {
		fmt.Printf("Quality Status: GOOD - Content generation is performing well with minor areas for improvement\n")
	} else if overallScore >= 0.7 {
		fmt.Printf("Quality Status: ACCEPTABLE - Content generation is adequate but has room for improvement\n")
	} else if overallScore >= 0.6 {
		fmt.Printf("Quality Status: NEEDS IMPROVEMENT - Content generation requires attention\n")
	} else {
		fmt.Printf("Quality Status: CRITICAL - Content generation requires immediate attention\n")
	}

	fmt.Printf("\nDemo completed successfully! The metrics system is tracking:\n")
	fmt.Printf("- Content generation performance and errors\n")
	fmt.Printf("- Content variety and uniqueness\n")
	fmt.Printf("- Logical consistency validation\n")
	fmt.Printf("- Player engagement and satisfaction\n")
	fmt.Printf("- System stability and reliability\n")
	fmt.Printf("- Comprehensive quality reporting with actionable insights\n")
}
