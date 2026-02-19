# Audit: goldbox-rpg/pkg/pcg/levels
**Date**: 2026-02-19
**Status**: Needs Work

## Summary
Room-corridor based dungeon level generation using BSP partitioning. Well-architected but contains multiple stub implementations where connectivity enforcement, cave features, and dungeon furniture are placeholder no-ops.

## Issues Found
- [ ] **high** Stub/Incomplete — `findWalkableRegions()`, `findLargestRegion()`, `connectRegions()` are simplified stubs returning empty results; connectivity enforcement doesn't actually work (`generator.go:410-433`)
- [ ] **high** Stub/Incomplete — `addDungeonDoors()`, `addTorchPositions()` are empty function implementations; features never actually added (`generator.go:394-407`)
- [ ] **med** Determinism — Constructor creates RNG with fixed seed `rand.NewSource(1)` that is overridden later; confusing initialization pattern (`generator.go:29`)
- [ ] **med** API Design — Connectivity level functions (moderate, high, complete) all call identical implementation with no functional differentiation (`generator.go:273-323`)
- [ ] **med** Documentation — Missing package-level doc.go file
- [ ] **low** Concurrency — Modifies room slices during iteration without synchronization; not thread-safe (`generator.go:373-380`)

## Test Coverage
85.1% (target: 65%) — ✅ EXCELLENT

2 test files with ~10 test functions and 1 integration test. Missing tests for BSP algorithm correctness and room distribution validation.

## Dependencies
**Internal:**
- `goldbox-rpg/pkg/game`: Game types
- `goldbox-rpg/pkg/pcg`: PCG interfaces

## Recommendations
1. **HIGH**: Complete stub implementations for connectivity enforcement, doors, and torch placement
2. **MEDIUM**: Differentiate connectivity level implementations or remove unused variants
3. **MEDIUM**: Fix constructor to not use fixed seed that gets immediately overridden
4. **LOW**: Add doc.go file
