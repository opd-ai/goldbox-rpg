# Audit: goldbox-rpg/pkg/pcg/items
**Date**: 2026-02-19
**Status**: Needs Work

## Summary
Core PCG item generation system with template-based generation, enchantments, and YAML configuration loading. Overall health is good with 83.9% test coverage, but has determinism issues in ID generation, performance problems from repeated registry instantiation, and missing package documentation.

## Issues Found
- [x] high determinism — Global rand used in generateItemID breaks deterministic generation (`generator.go:314`)
- [x] high performance — GenerateItemName creates new registry on every call causing memory/CPU waste (`templates.go:258`)
- [x] high performance — ApplyEnchantments creates new registry on every call causing memory/CPU waste (`enchantments.go:46-47`)
- [x] med documentation — Package missing doc.go file for package-level documentation (package items)
- [x] med error-handling — GenerateItemSet silently skips failed items without logging (`generator.go:150-156`)
- [x] med test — TestDeterministicEnchantments fails with inconsistent value changes (`enchantments_test.go:485`)
- [x] low determinism — Test files use time.Now() for RNG seeding affecting reproducibility (`enchantments_test.go:49`, `templates_test.go:165`)
- [x] low api-design — applyRarityModifications always returns nil, should be void (`generator.go:197`)
- [x] low error-handling — NewTemplateBasedGenerator silently ignores LoadDefaultTemplates error (`generator.go:31-33`)

## Test Coverage
83.9% (target: 65%)

## Dependencies
**Internal**: goldbox-rpg/pkg/game, goldbox-rpg/pkg/pcg, goldbox-rpg/pkg/resilience
**External**: gopkg.in/yaml.v3 (YAML parsing), math/rand (RNG), context (cancellation)
**Integration**: Template system loads from YAML files, enchantment system modifies game.Item structs

## Recommendations
1. Fix generateItemID to use tbg.rng for deterministic ID generation
2. Cache ItemTemplateRegistry instance in GenerateItemName and ApplyEnchantments
3. Add package doc.go with overview of item generation system
4. Log skipped items in GenerateItemSet with structured logging
5. Fix or document TestDeterministicEnchantments flaky behavior
