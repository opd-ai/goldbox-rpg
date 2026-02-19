# Audit: goldbox-rpg/pkg/pcg/items
**Date**: 2026-02-19
**Status**: Needs Work

## Summary
Template-based item generation system with enchantment support and YAML template loading. Well-documented with good test coverage, but has a critical performance issue where a new ItemTemplateRegistry is created on every enchantment application.

## Issues Found
- [ ] **high** Performance — Creates new `ItemTemplateRegistry()` and calls `LoadDefaultTemplates()` on every enchantment application; should inject registry as dependency (`enchantments.go:46-47`)
- [ ] **high** Error Handling — `GenerateItemSet()` silently skips errors without logging; if all items fail, returns generic error with no diagnostics (`generator.go:150,155`)
- [ ] **high** Determinism — `generateItemID()` uses `rand.Int63()` (global/unseeded) instead of seeded RNG; inconsistent with deterministic generation pattern (`generator.go:314`)
- [ ] **med** Test Coverage — No tests for error paths in GenerateItemSet() silent skipping
- [ ] **med** Documentation — Missing package-level doc.go file
- [ ] **low** Code Quality — Damage scaling `"+ 1"` hardcoded instead of using rarity modifier (`generator.go:246`)

## Test Coverage
84.3% (target: 65%) — ✅ EXCELLENT

7 test files with ~34 test functions including integration tests and YAML loading edge cases.

## Dependencies
**External:**
- `gopkg.in/yaml.v3`: YAML parsing

**Internal:**
- `goldbox-rpg/pkg/game`: Item types
- `goldbox-rpg/pkg/pcg`: PCG interfaces

## Recommendations
1. **CRITICAL**: Inject or cache ItemTemplateRegistry in enchantment system instead of recreating per call
2. **HIGH**: Log or aggregate errors in GenerateItemSet() for diagnostics
3. **HIGH**: Use seeded RNG for generateItemID() to maintain determinism
4. **LOW**: Add doc.go file
