package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sirupsen/logrus"

	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/pcg"
)

// timeNow is the function used to get the current time.
// It defaults to time.Now but can be overridden in tests for reproducibility.
var timeNow = time.Now

func main() {
	fmt.Println("=== PCG Event System Integration Demo ===")
	fmt.Println()

	// Set up logging
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create a test world
	world := &game.World{
		Width:       100,
		Height:      100,
		Levels:      []game.Level{},
		Objects:     make(map[string]game.GameObject),
		Players:     make(map[string]*game.Player),
		NPCs:        make(map[string]*game.NPC),
		SpatialGrid: make(map[game.Position][]string),
	}

	// Initialize PCG Manager
	pcgManager := pcg.NewPCGManager(world, logger)
	fmt.Println("✓ PCG Manager initialized")

	// Initialize Event System
	eventSystem := game.NewEventSystem()
	eventManager := pcg.NewPCGEventManager(logger, eventSystem, pcgManager)
	fmt.Println("✓ PCG Event Manager initialized")

	// Display initial configuration
	config := eventManager.GetAdjustmentConfig()
	fmt.Printf("Runtime Adjustments Enabled: %t\n", config.EnableRuntimeAdjustments)
	fmt.Printf("Quality Threshold (Overall): %.2f\n", config.QualityThresholds.MinOverallScore)
	fmt.Printf("Monitoring Interval: %v\n", config.MonitoringInterval)
	fmt.Printf("Max Adjustments: %d\n\n", config.MaxAdjustments)

	// Start runtime monitoring
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	eventManager.StartMonitoring(ctx)
	fmt.Println("✓ Runtime monitoring started")

	// Simulate content generation with quality tracking
	fmt.Println("\n=== Simulating Content Generation Events ===")

	// Generate some content with varying quality
	fmt.Println("1. Generating quests with normal quality...")

	// Generate a quest using the available method
	quest, err := pcgManager.GenerateQuestForArea(context.Background(), "demo_area", pcg.QuestTypeFetch, 5)
	if err != nil {
		log.Printf("Warning: Failed to generate quest: %v", err)
		// Create a mock quest for demonstration
		quest = &game.Quest{
			ID:          "demo_quest_1",
			Title:       "Find the Lost Artifact",
			Description: "A precious artifact has gone missing from the local temple.",
			Status:      game.QuestActive,
		}
	}

	// Simulate content generation event
	eventManager.EmitContentGenerated(pcg.ContentTypeQuests, quest, 15*time.Millisecond, 0.85)
	fmt.Printf("   Generated quest: %s (Quality: 0.85)\n", quest.Title)

	// Generate some items
	items, err := pcgManager.GenerateItemsForLocation(context.Background(), "demo_location", 3, pcg.RarityCommon, pcg.RarityRare, 5)
	if err != nil {
		log.Printf("Warning: Failed to generate items: %v", err)
	} else {
		for _, item := range items {
			eventManager.EmitContentGenerated(pcg.ContentTypeItems, item, 10*time.Millisecond, 0.80)
			fmt.Printf("   Generated item: %s (Quality: 0.80)\n", item.Name)
		}
	}

	// Generate quality report
	fmt.Println("\n2. Generating initial quality report...")
	qualityReport := pcgManager.GenerateQualityReport()
	if qualityReport != nil {
		eventManager.EmitQualityAssessment(qualityReport)
		fmt.Printf("   Overall Quality Score: %.3f (%s)\n", qualityReport.OverallScore, qualityReport.QualityGrade)

		for component, score := range qualityReport.ComponentScores {
			fmt.Printf("   %s: %.3f\n", component, score)
		}
	}

	// Simulate low-quality content generation that should trigger adjustments
	fmt.Println("\n3. Simulating low-quality content generation...")
	eventManager.EmitContentGenerated(pcg.ContentTypeCharacters, "low-quality-npc", 50*time.Millisecond, 0.4) // Very low quality
	fmt.Println("   Generated low-quality content (Quality: 0.4)")

	// Simulate player feedback
	fmt.Println("\n=== Simulating Player Feedback ===")

	feedbackScenarios := []struct {
		description string
		feedback    pcg.PlayerFeedback
	}{
		{
			description: "Player finds content too easy",
			feedback: pcg.PlayerFeedback{
				ContentType: pcg.ContentTypeQuests,
				ContentID:   "quest_001",
				Rating:      3,
				Difficulty:  2, // Too easy
				Enjoyment:   6,
				Comments:    "Too easy, need more challenge",
				SessionID:   "demo_session_1",
				Timestamp:   timeNow(),
			},
		},
		{
			description: "Player finds content too difficult",
			feedback: pcg.PlayerFeedback{
				ContentType: pcg.ContentTypeQuests,
				ContentID:   "quest_002",
				Rating:      4,
				Difficulty:  8, // Too hard
				Enjoyment:   4,
				Comments:    "Very challenging, maybe too hard",
				SessionID:   "demo_session_2",
				Timestamp:   timeNow(),
			},
		},
		{
			description: "Player has low enjoyment",
			feedback: pcg.PlayerFeedback{
				ContentType: pcg.ContentTypeDungeon, // Fixed: ContentTypeDungeon not ContentTypeDungeons
				ContentID:   "dungeon_001",
				Rating:      2,
				Difficulty:  5,
				Enjoyment:   2, // Low enjoyment
				Comments:    "Boring and repetitive",
				SessionID:   "demo_session_3",
				Timestamp:   timeNow(),
			},
		},
		{
			description: "Player is satisfied",
			feedback: pcg.PlayerFeedback{
				ContentType: pcg.ContentTypeQuests,
				ContentID:   "quest_003",
				Rating:      5,
				Difficulty:  5, // Just right
				Enjoyment:   8, // High enjoyment
				Comments:    "Perfect balance and very engaging!",
				SessionID:   "demo_session_4",
				Timestamp:   timeNow(),
			},
		},
	}

	for i, scenario := range feedbackScenarios {
		fmt.Printf("%d. %s\n", i+1, scenario.description)
		eventManager.EmitPlayerFeedback(&scenario.feedback)
		pcgManager.RecordPlayerFeedback(scenario.feedback)
		fmt.Printf("   Difficulty: %d/10, Enjoyment: %d/10, Rating: %d/5\n",
			scenario.feedback.Difficulty, scenario.feedback.Enjoyment, scenario.feedback.Rating)
	}

	// Simulate system health events
	fmt.Println("\n=== Simulating System Health Monitoring ===")

	healthScenarios := []struct {
		description string
		healthData  map[string]interface{}
	}{
		{
			description: "High memory usage detected",
			healthData: map[string]interface{}{
				"memory_usage": 0.85, // High memory usage
				"error_rate":   0.02,
			},
		},
		{
			description: "High error rate detected",
			healthData: map[string]interface{}{
				"memory_usage": 0.4,
				"error_rate":   0.08, // High error rate
			},
		},
		{
			description: "System running normally",
			healthData: map[string]interface{}{
				"memory_usage": 0.3,
				"error_rate":   0.01,
			},
		},
	}

	for i, scenario := range healthScenarios {
		fmt.Printf("%d. %s\n", i+1, scenario.description)

		healthEvent := game.GameEvent{
			Type:      pcg.EventPCGSystemHealth,
			SourceID:  "system_monitor",
			TargetID:  "pcg_system",
			Data:      map[string]interface{}{"health_data": scenario.healthData},
			Timestamp: timeNow().Unix(),
		}

		eventSystem.Emit(healthEvent)
		fmt.Printf("   Memory: %.1f%%, Error Rate: %.1f%%\n",
			scenario.healthData["memory_usage"].(float64)*100,
			scenario.healthData["error_rate"].(float64)*100)
	}

	// Allow time for event processing
	time.Sleep(100 * time.Millisecond)

	// Display adjustment results
	fmt.Println("\n=== Runtime Adjustment Results ===")
	adjustmentCount := eventManager.GetAdjustmentCount()
	fmt.Printf("Total Adjustments Made: %d\n", adjustmentCount)

	history := eventManager.GetAdjustmentHistory()
	if len(history) > 0 {
		fmt.Println("\nAdjustment History:")
		for i, record := range history {
			fmt.Printf("%d. %s - %s (%s) - Success: %t\n",
				i+1, record.Timestamp.Format("15:04:05"), record.Trigger, record.AdjustmentType, record.Success)
			if record.QualityBefore > 0 {
				fmt.Printf("   Quality Before: %.3f\n", record.QualityBefore)
			}
		}
	} else {
		fmt.Println("No adjustments were made during this demo.")
	}

	// Generate final quality report
	fmt.Println("\n=== Final Quality Assessment ===")
	finalReport := pcgManager.GenerateQualityReport()
	if finalReport != nil {
		fmt.Printf("Final Overall Score: %.3f (%s)\n", finalReport.OverallScore, finalReport.QualityGrade)

		fmt.Println("\nFinal Component Scores:")
		for component, score := range finalReport.ComponentScores {
			fmt.Printf("  %s: %.3f\n", component, score)
		}

		if len(finalReport.Recommendations) > 0 {
			fmt.Println("\nRecommendations:")
			for i, rec := range finalReport.Recommendations {
				fmt.Printf("  %d. %s\n", i+1, rec)
			}
		}

		if len(finalReport.CriticalIssues) > 0 {
			fmt.Println("\nCritical Issues:")
			for i, issue := range finalReport.CriticalIssues {
				fmt.Printf("  %d. %s\n", i+1, issue)
			}
		}
	}

	// Stop monitoring
	eventManager.StopMonitoring()
	fmt.Println("\n✓ Runtime monitoring stopped")

	// Display event system statistics
	fmt.Println("\n=== Event System Statistics ===")
	fmt.Printf("Monitoring Duration: %v\n", 30*time.Second)
	fmt.Printf("Events Processed: Multiple PCG events\n")
	fmt.Printf("Adjustments Triggered: %d\n", adjustmentCount)
	fmt.Printf("Final System Status: %s\n", func() string {
		if finalReport != nil && finalReport.OverallScore >= 0.7 {
			return "HEALTHY"
		}
		return "NEEDS ATTENTION"
	}())

	fmt.Println("\n=== Demo Complete ===")
	fmt.Println("\nThe PCG Event System Integration demonstrates:")
	fmt.Println("• Real-time quality monitoring and adjustment")
	fmt.Println("• Player feedback integration for runtime tuning")
	fmt.Println("• System health monitoring and response")
	fmt.Println("• Automatic content quality assessment")
	fmt.Println("• Event-driven runtime parameter adjustment")
}
