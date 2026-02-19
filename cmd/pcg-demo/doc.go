// Package main provides a demonstration of the GoldBox RPG procedural content
// generation (PCG) performance metrics and monitoring system.
//
// This demo showcases:
//   - PCG Manager initialization and seeding
//   - Terrain generation with various biome types
//   - Item generation with rarity parameters
//   - Cache performance tracking (hits/misses)
//   - Generation statistics and timing metrics
//   - Metrics collection and reporting
//
// Usage:
//
//	go run ./cmd/pcg-demo
//
// The demo creates a test world, performs various PCG operations, simulates
// cache activity, and displays comprehensive metrics for each content type.
// All output is written to stdout with structured formatting.
//
// Example output:
//
//	=== PCG Performance Metrics Demo ===
//	PCG Manager initialized with seed 12345
//
//	=== Initial Metrics ===
//	Total Generations: 0
//	Cache Hit Ratio: 0.00%
//
//	=== Performing Content Generation ===
//	Generating terrain for level_1...
//	...
package main
