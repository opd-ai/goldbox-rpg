# Audit: goldbox-rpg/pkg/pcg/utils
**Date**: 2026-02-19
**Status**: Complete

## Summary
PCG utilities package providing Perlin/Simplex noise generation and A* pathfinding. Package demonstrates solid mathematical implementations with deterministic seeding, comprehensive test coverage (92.9%), and good algorithmic correctness. Minor documentation and API design improvements recommended.

## Issues Found
- [ ] **low** Documentation — Missing package doc.go file to explain overall purpose and usage patterns
- [ ] **low** Documentation — AStarPathfind function missing godoc comment (`pathfinding.go:59`)
- [ ] **low** Documentation — SimplexNoise type missing exported godoc comment (`noise.go:105`)
- [ ] **med** API Design — Node struct exposes all fields including Index which is internal to priority queue (`pathfinding.go:18-25`)
- [ ] **low** API Design — Helper functions (fade, lerp, grad2d, dot2d) are unexported but may be useful for extending noise algorithms
- [ ] **low** Test Coverage — FractalNoise method lacks dedicated tests beyond basic functionality check
- [ ] **low** Dependencies — Package currently unused by other PCG modules, suggesting integration gap

## Test Coverage
92.9% (target: 65%)

## Dependencies
**Standard Library:**
- `container/heap` - Priority queue for A* pathfinding
- `math` - Mathematical operations for noise generation

**Internal:**
- `goldbox-rpg/pkg/game` - Position and GameMap types for pathfinding

**External:**
- None (all standard library)

**Integration Surface:**
- Currently NOT imported by any other PCG modules (terrain, items, quests, levels)
- Suggests utilities are built but not yet integrated into PCG pipeline
- Good candidate for future terrain generation integration

## Recommendations
1. **High Priority**: Add package doc.go explaining noise generation and pathfinding utilities with usage examples
2. **Medium Priority**: Refactor Node struct to hide Index field (make internal NodeInternal type or use unexported field pattern)
3. **Low Priority**: Add comprehensive FractalNoise tests with various octave/persistence/scale combinations
4. **Low Priority**: Add godoc comments to AStarPathfind and SimplexNoise exported API
5. **Low Priority**: Integrate noise generators into pkg/pcg/terrain for procedural terrain generation
