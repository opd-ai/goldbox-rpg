// Package main provides a demonstration application for the PCG Event System
// integration in the GoldBox RPG engine.
//
// The events-demo application showcases the runtime adjustment capabilities of
// the procedural content generation (PCG) system, demonstrating real-time
// quality monitoring, player feedback integration, and system health tracking.
//
// # Usage
//
// Run the demo directly:
//
//	go run ./cmd/events-demo
//
// Or build and execute:
//
//	go build -o events-demo ./cmd/events-demo
//	./events-demo
//
// # Event System Features
//
// The demo demonstrates the following PCG event types:
//
//   - Content Generation Events: Tracks when PCG content is created with quality metrics
//   - Quality Assessment Events: Monitors overall content quality and triggers adjustments
//   - Player Feedback Events: Integrates player ratings, difficulty feedback, and enjoyment
//   - System Health Events: Monitors memory usage and error rates for adaptive performance
//
// # Runtime Adjustment System
//
// The PCG event manager provides runtime adjustment capabilities:
//
//   - Quality Thresholds: Configurable thresholds for triggering automatic adjustments
//   - Adjustment Types: Difficulty, variety, complexity, and performance adjustments
//   - Feedback Integration: Adjusts content generation based on player feedback
//   - Health Monitoring: Responds to system resource constraints
//
// # Configuration
//
// The runtime adjustment system uses configurable parameters:
//
//   - EnableRuntimeAdjustments: Toggle automatic adjustments on/off
//   - QualityThresholds: Minimum scores for overall, performance, variety, etc.
//   - AdjustmentRates: Step sizes for difficulty, variety boost, complexity reduction
//   - MonitoringInterval: Frequency of quality checks (default: 30 seconds)
//   - MaxAdjustments: Maximum adjustments per session (default: 10)
//
// # Demo Scenarios
//
// The demo simulates several scenarios:
//
//  1. Content Generation: Generates quests and items with quality tracking
//  2. Quality Assessment: Creates quality reports and triggers adjustments
//  3. Player Feedback: Simulates various feedback scenarios (too easy, too hard, low enjoyment)
//  4. System Health: Monitors memory usage and error rates
//
// # Output
//
// The demo outputs:
//
//   - Event system initialization confirmation
//   - Runtime adjustment configuration details
//   - Content generation events with quality scores
//   - Player feedback processing results
//   - System health monitoring events
//   - Final adjustment history and quality assessment
//
// # Integration Example
//
// The event manager can be integrated into game sessions:
//
//	eventSystem := game.NewEventSystem()
//	pcgManager := pcg.NewPCGManager(world, logger)
//	eventManager := pcg.NewPCGEventManager(logger, eventSystem, pcgManager)
//
//	// Configure adjustments
//	config := pcg.DefaultRuntimeAdjustmentConfig()
//	config.QualityThresholds.MinOverallScore = 0.6
//	eventManager.SetAdjustmentConfig(config)
//
//	// Start monitoring
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	eventManager.StartMonitoring(ctx)
//
//	// Generate content with quality tracking
//	eventManager.EmitContentGenerated(pcg.ContentTypeQuests, quest, duration, quality)
package main
