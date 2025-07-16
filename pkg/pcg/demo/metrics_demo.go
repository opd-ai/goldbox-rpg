package main

import (
	"context"
	"encoding/json"
	"fmt"

	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/pcg"

	"github.com/sirupsen/logrus"
)

// Demo script to showcase the Performance Metrics and Monitoring system
func main() {
	fmt.Println("=== PCG Performance Metrics Demo ===")

	// Create a test world
	world := &game.World{
		Objects: make(map[string]game.GameObject),
		Levels:  []game.Level{},
		Players: make(map[string]*game.Player),
	}

	// Create logger with info level for demo
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create PCG manager
	pcgManager := pcg.NewPCGManager(world, logger)
	pcgManager.InitializeWithSeed(12345)

	fmt.Println("PCG Manager initialized with seed 12345")

	// Get initial metrics (should be empty)
	fmt.Println("\n=== Initial Metrics ===")
	printMetrics(pcgManager.GetMetrics())

	// Perform several generations to populate metrics
	ctx := context.Background()

	fmt.Println("\n=== Performing Content Generation ===")

	// Generate terrain multiple times
	for i := 0; i < 3; i++ {
		levelID := fmt.Sprintf("level_%d", i+1)
		fmt.Printf("Generating terrain for %s...\n", levelID)

		_, err := pcgManager.GenerateTerrainForLevel(ctx, levelID, 20, 20, pcg.BiomeCave, 5)
		if err != nil {
			fmt.Printf("Error generating terrain: %v\n", err)
		}
	}

	// Generate items multiple times
	for i := 0; i < 5; i++ {
		locationID := fmt.Sprintf("location_%d", i+1)
		fmt.Printf("Generating items for %s...\n", locationID)

		_, err := pcgManager.GenerateItemsForLocation(ctx, locationID, 3, pcg.RarityCommon, pcg.RarityRare, 5)
		if err != nil {
			fmt.Printf("Error generating items: %v\n", err)
		}
	}

	// Simulate cache hits and misses
	metrics := pcgManager.GetMetrics()
	for i := 0; i < 10; i++ {
		if i%3 == 0 {
			metrics.RecordCacheMiss()
		} else {
			metrics.RecordCacheHit()
		}
	}

	// Show final metrics
	fmt.Println("\n=== Final Metrics ===")
	printMetrics(metrics)

	// Show generation statistics from manager
	fmt.Println("\n=== Manager Statistics ===")
	stats := pcgManager.GetGenerationStatistics()
	prettyPrint(stats)

	// Demonstrate metrics reset
	fmt.Println("\n=== Resetting Metrics ===")
	pcgManager.ResetMetrics()
	fmt.Println("Metrics after reset:")
	printMetrics(pcgManager.GetMetrics())

	fmt.Println("\n=== Demo Complete ===")
}

func printMetrics(metrics *pcg.GenerationMetrics) {
	fmt.Printf("Total Generations: %d\n", metrics.TotalGenerations)
	fmt.Printf("Cache Hit Ratio: %.2f%%\n", metrics.GetCacheHitRatio())

	// Print per-content-type metrics
	contentTypes := []pcg.ContentType{
		pcg.ContentTypeTerrain,
		pcg.ContentTypeItems,
		pcg.ContentTypeLevels,
		pcg.ContentTypeQuests,
	}

	for _, contentType := range contentTypes {
		count := metrics.GetGenerationCount(contentType)
		if count > 0 {
			avgTime := metrics.GetAverageTiming(contentType)
			errorCount := metrics.GetErrorCount(contentType)
			fmt.Printf("  %s: %d generations, avg time: %v, errors: %d\n",
				contentType, count, avgTime, errorCount)
		}
	}
}

func prettyPrint(data interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling data: %v\n", err)
		return
	}
	fmt.Println(string(jsonData))
}
