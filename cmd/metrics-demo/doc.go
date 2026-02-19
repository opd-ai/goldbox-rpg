// Package main provides a demonstration application for the Content Quality
// Metrics System in the GoldBox RPG engine.
//
// The metrics-demo application showcases the comprehensive quality tracking
// capabilities of the procedural content generation (PCG) system, demonstrating
// performance metrics, player feedback integration, balance tracking, and
// quality reporting.
//
// # Usage
//
// Run the demo directly:
//
//	go run ./cmd/metrics-demo
//
// Or build and execute:
//
//	go build -o metrics-demo ./cmd/metrics-demo
//	./metrics-demo
//
// # Quality Metrics Features
//
// The demo demonstrates the following quality metrics components:
//
//   - Content Generation Tracking: Records generation duration, success/failure rates
//   - Player Feedback Integration: Captures ratings, difficulty, and enjoyment scores
//   - Quest Completion Analysis: Tracks completion times and abandonment rates
//   - Balance Metrics: Monitors system health and balance check results
//   - Performance Metrics: Measures cache hit ratios and generation statistics
//
// # Quality Report System
//
// The quality report provides comprehensive assessment:
//
//   - Overall Quality Score: Weighted aggregate of all quality components (0.0-1.0)
//   - Quality Grade: Letter grade (A, B, C, D, F) based on overall score
//   - Component Scores: Individual scores for performance, variety, consistency, etc.
//   - Threshold Compliance: Pass/fail status for each quality threshold
//   - Recommendations: Actionable suggestions for improving content quality
//   - Critical Issues: Urgent problems requiring immediate attention
//
// # Demo Scenarios
//
// The demo simulates several scenarios:
//
//  1. Terrain Generation: Generates multiple terrain levels with performance tracking
//  2. Quest Generation: Creates quests including simulated failures for testing
//  3. Item Generation: Produces item sets with generation time measurement
//  4. Player Feedback: Records various feedback patterns for different content types
//  5. Quest Completion: Tracks completion and abandonment rates
//  6. Quality Reporting: Generates comprehensive quality assessment report
//
// # Output
//
// The demo outputs:
//
//   - Initialization confirmation with seed value
//   - Generation events for each content type
//   - Recorded player feedback summaries
//   - Quest completion status updates
//   - Comprehensive quality report with all metrics
//   - Performance statistics including cache efficiency
//   - Final quality assessment with status determination
//
// # Deterministic Behavior
//
// The demo uses a fixed seed (42) for deterministic output, ensuring consistent
// results across runs. This facilitates debugging and validation of the metrics
// system behavior.
//
// # Integration Example
//
// The quality metrics system can be integrated into game sessions:
//
//	pcgManager := pcg.NewPCGManager(world, logger)
//	pcgManager.InitializeWithSeed(seed)
//
//	qualityMetrics := pcgManager.GetQualityMetrics()
//
//	// Record content generation
//	qualityMetrics.RecordContentGeneration(pcg.ContentTypeTerrain, contentID, duration, err)
//
//	// Record player feedback
//	pcgManager.RecordPlayerFeedback(feedback)
//
//	// Generate quality report
//	report := pcgManager.GenerateQualityReport()
//	fmt.Printf("Quality Score: %.3f (%s)\n", report.OverallScore, report.QualityGrade)
package main
