package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/pcg"

	"github.com/sirupsen/logrus"
)

// Config holds demo configuration options.
type Config struct {
	Seed        int64
	Logger      *logrus.Logger
	Output      io.Writer
	NumTerrains int
	NumItems    int
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	return Config{
		Seed:        12345,
		Logger:      logger,
		Output:      os.Stdout,
		NumTerrains: 3,
		NumItems:    5,
	}
}

// RunDemo executes the PCG metrics demonstration with the given configuration.
// It returns an error if any critical operation fails.
func RunDemo(cfg Config) error {
	out := cfg.Output
	if out == nil {
		out = os.Stdout
	}

	fmt.Fprintln(out, "=== PCG Performance Metrics Demo ===")

	// Create a test world
	world := &game.World{
		Objects: make(map[string]game.GameObject),
		Levels:  []game.Level{},
		Players: make(map[string]*game.Player),
	}

	// Create PCG manager
	pcgManager := pcg.NewPCGManager(world, cfg.Logger)
	pcgManager.InitializeWithSeed(cfg.Seed)

	fmt.Fprintf(out, "PCG Manager initialized with seed %d\n", cfg.Seed)

	// Get initial metrics (should be empty)
	fmt.Fprintln(out, "\n=== Initial Metrics ===")
	printMetrics(out, pcgManager.GetMetrics())

	// Perform several generations to populate metrics
	ctx := context.Background()

	fmt.Fprintln(out, "\n=== Performing Content Generation ===")

	// Generate terrain multiple times
	for i := 0; i < cfg.NumTerrains; i++ {
		levelID := fmt.Sprintf("level_%d", i+1)
		fmt.Fprintf(out, "Generating terrain for %s...\n", levelID)

		_, err := pcgManager.GenerateTerrainForLevel(ctx, levelID, 20, 20, pcg.BiomeCave, 5)
		if err != nil {
			return fmt.Errorf("generating terrain for %s: %w", levelID, err)
		}
	}

	// Generate items multiple times
	for i := 0; i < cfg.NumItems; i++ {
		locationID := fmt.Sprintf("location_%d", i+1)
		fmt.Fprintf(out, "Generating items for %s...\n", locationID)

		_, err := pcgManager.GenerateItemsForLocation(ctx, locationID, 3, pcg.RarityCommon, pcg.RarityRare, 5)
		if err != nil {
			return fmt.Errorf("generating items for %s: %w", locationID, err)
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
	fmt.Fprintln(out, "\n=== Final Metrics ===")
	printMetrics(out, metrics)

	// Show generation statistics from manager
	fmt.Fprintln(out, "\n=== Manager Statistics ===")
	if err := prettyPrint(out, pcgManager.GetGenerationStatistics()); err != nil {
		return fmt.Errorf("printing statistics: %w", err)
	}

	// Demonstrate metrics reset
	fmt.Fprintln(out, "\n=== Resetting Metrics ===")
	pcgManager.ResetMetrics()
	fmt.Fprintln(out, "Metrics after reset:")
	printMetrics(out, pcgManager.GetMetrics())

	fmt.Fprintln(out, "\n=== Demo Complete ===")
	return nil
}

// printMetrics outputs generation metrics to the given writer.
func printMetrics(out io.Writer, metrics *pcg.GenerationMetrics) {
	fmt.Fprintf(out, "Total Generations: %d\n", metrics.TotalGenerations)
	fmt.Fprintf(out, "Cache Hit Ratio: %.2f%%\n", metrics.GetCacheHitRatio())

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
			fmt.Fprintf(out, "  %s: %d generations, avg time: %v, errors: %d\n",
				contentType, count, avgTime, errorCount)
		}
	}
}

// prettyPrint outputs data as formatted JSON to the given writer.
func prettyPrint(out io.Writer, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	fmt.Fprintln(out, string(jsonData))
	return nil
}

func main() {
	cfg := DefaultConfig()
	if err := RunDemo(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
