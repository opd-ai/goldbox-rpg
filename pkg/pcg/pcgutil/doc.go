// Package pcgutil provides utility functions for procedural content generation.
//
// This package contains common algorithms and data structures used across
// the PCG subsystem, including noise generation, pathfinding, and random
// number utilities.
//
// # Perlin Noise
//
// PerlinNoise generates coherent noise for terrain and texture generation:
//
//	noise := pcgutil.NewPerlinNoise(seed)
//	value := noise.Noise2D(x, y)        // Single point
//	value = noise.OctaveNoise2D(x, y, octaves, persistence)  // Layered detail
//
// The noise generator produces values in [-1, 1] range with smooth interpolation
// between sample points. Octave layering adds detail at multiple frequencies.
//
// # A* Pathfinding
//
// AStarPathfind finds optimal paths on game maps:
//
//	result := pcgutil.AStarPathfind(gameMap, start, goal)
//	if result.Found {
//	    for _, pos := range result.Path {
//	        // Follow path
//	    }
//	}
//
// The implementation uses Manhattan distance heuristic and supports
// configurable movement costs per tile type.
//
// # Priority Queue
//
// PriorityQueue provides a min-heap implementation for A* pathfinding:
//
//	pq := &pcgutil.PriorityQueue{}
//	heap.Init(pq)
//	heap.Push(pq, node)
//	lowest := heap.Pop(pq).(*pcgutil.Node)
//
// # Random Utilities
//
// Helper functions for seeded random number generation ensure deterministic
// content generation across sessions and multiplayer environments.
//
// # Mathematical Helpers
//
// Dot2D computes 2D dot products for gradient noise calculations:
//
//	result := pcgutil.Dot2D(gradient, x, y)
package pcgutil
