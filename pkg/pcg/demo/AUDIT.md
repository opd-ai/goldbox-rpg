# Audit: goldbox-rpg/pkg/pcg/demo
**Date**: 2026-02-19
**Status**: Needs Work

## Summary
Single-file demonstration package showcasing PCG performance metrics system. Package has zero test coverage and is marked as package `main`, which is architecturally questionable for a library sub-package under `pkg/`. Multiple Go best practice violations including missing error handling, hard-coded values, and inappropriate package declaration.

## Issues Found
- [ ] high api-design — Package declared as `main` in library tree (`metrics_demo.go:1`)
- [ ] high test-coverage — Zero test coverage (0.0%, target: 65%) - no test files exist
- [ ] med error-handling — Error logging uses Printf instead of returning errors (`metrics_demo.go:51`, `metrics_demo.go:62`)
- [ ] med determinism — Hard-coded seed value (12345) prevents demonstrating seed flexibility (`metrics_demo.go:31`)
- [ ] med documentation — No godoc comments for exported functions (`metrics_demo.go:15`, `metrics_demo.go:94`, `metrics_demo.go:117`)
- [ ] low documentation — Package lacks doc.go file explaining purpose and usage
- [ ] low stub-code — Metrics simulation loop uses arbitrary modulo logic without explanation (`metrics_demo.go:69`)
- [ ] low error-handling — MarshalIndent error only prints message without proper context (`metrics_demo.go:120`)

## Test Coverage
0.0% (target: 65%)

## Dependencies
**External:**
- github.com/sirupsen/logrus (logging)

**Internal:**
- goldbox-rpg/pkg/game (World, GameObject, Level, Player types)
- goldbox-rpg/pkg/pcg (PCGManager, metrics types, biome/rarity constants)

**Importers:** None (demo package not imported by other code)

## Recommendations
1. Move to `cmd/pcg-metrics-demo/main.go` - demo/example code should not be under `pkg/`
2. Add table-driven tests demonstrating metrics collection edge cases
3. Accept seed as command-line flag instead of hard-coding value
4. Convert error handling to return errors with context wrapping
5. Add godoc comments for all exported functions
6. Create doc.go explaining demonstration scope and expected output
