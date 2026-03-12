# Implementation Plan: Spell System Completion

## Project Context
- **What it does**: A modern Go-based RPG engine inspired by SSI Gold Box games, providing character management, turn-based combat, spell casting, and multiplayer via JSON-RPC API with WebSocket support
- **Current goal**: Complete spell system content (levels 3-9 missing) to enable full magical character progression
- **Estimated Scope**: Medium (35-49 spells across 7 missing levels)

## Goal-Achievement Status
| Stated Goal | Current Status | This Plan Addresses |
|-------------|---------------|---------------------|
| Core RPG mechanics and character system | ✅ Achieved | No |
| Combat and effect systems | ✅ Achieved | No |
| WebSocket real-time communication | ✅ Achieved | No |
| Procedural Content Generation system | ✅ Achieved | No |
| Circuit breaker patterns and resilience | ✅ Achieved | No |
| Comprehensive input validation | ✅ Achieved | No |
| Health monitoring and metrics | ✅ Achieved | No |
| Asset generation pipeline (521 assets) | ⚠️ Partial (7/521) | No |
| Advanced NPC AI behaviors | ✅ Achieved | No |
| **Additional spell effects** | ⚠️ Partial (9 spells, levels 0-2 only) | **Yes** |
| World editor tools | ❌ Missing | No |
| Network optimization | ⚠️ Partial | No |
| Content creation utilities | ⚠️ Partial | No |
| Player progression persistence | ✅ Achieved | No |
| Guild and faction systems | ⚠️ Partial | No |
| Enhanced combat mechanics | ✅ Achieved | No |

**Why This Goal**: Spell system completion is **highest ROI** for current state:
1. Low complexity (primarily YAML content creation, ~20-30 hours)
2. High gameplay value (enables Mage/Cleric class progression)
3. Infrastructure already exists (`pkg/game/spell_manager.go`, `pkg/game/spell_effects.go`)
4. Unblocks magical combat testing and character viability
5. No dependencies on other incomplete features

## Metrics Summary
- Complexity hotspots on goal-critical paths: **0** functions above threshold (spell system code all <10 cyclomatic)
- Duplication ratio: **2.0%** (1,138 lines, 42 clone pairs)
- Doc coverage: **86.3%** overall (93.6% function coverage)
- Package coupling: `pkg/game` well-isolated; spell system uses existing effect manager

### Key Metrics (from go-stats-generator)
| Metric | Value | Threshold | Status |
|--------|-------|-----------|--------|
| Total functions | 517 | - | - |
| Functions cyclomatic >9 | 6 | <15 | ✅ Clean |
| Highest complexity | 14 (`addVegetation`, `refreshGameState`) | <15 | ✅ Acceptable |
| Doc coverage | 86.3% | >80% | ✅ Good |
| Duplication ratio | 2.0% | <3% | ✅ Good |

### Spell System Current State
- **Files present**: `data/spells/cantrips.yaml`, `data/spells/level1.yaml`, `data/spells/level2.yaml`
- **Total spells defined**: ~9 spells across 3 levels
- **Target**: 40-58 spells across all 10 levels (cantrips + levels 1-9)
- **Gap**: Levels 3-9 completely missing (0 spells)

## Research Findings

### Dependency Considerations
1. **Gorilla WebSocket (v1.5.3)**: Officially deprecated/archived in 2022. No new critical CVEs in 2025-2026, but recommend migration planning to `nhooyr.io/websocket` for long-term maintenance.
2. **Go 1.23.x**: 18 standard library vulnerabilities require Go 1.24.12+ to resolve (per CHANGELOG.md). Current versions are latest compatible with Go 1.23.
3. **Ebitengine v2.7.0**: Stable, well-maintained. Latest versions require Go 1.24.0+.

### Project Status
- 76% of stated goals fully achieved (13/17)
- Code quality excellent: 78% test coverage, clean architecture
- Active documentation and CI/CD in place

---

## Implementation Steps

### Step 1: Create Level 3 Spell Data File
- **Deliverable**: `data/spells/level3.yaml` with 5-7 spells
- **Dependencies**: None
- **Goal Impact**: Enables level 5-6 character magic access
- **Acceptance**: File validates against existing schema; `getSpellsByLevel(3)` returns 5+ spells
- **Validation**: 
  ```bash
  go test ./pkg/game -run TestSpellManager -v
  # Verify: go-stats-generator analyze ./data/spells --format json | jq '.documentation.coverage.overall'
  ```

**Recommended Spells (Level 3)**:
```yaml
spells:
  - spell_id: lightning_bolt
    spell_name: Lightning Bolt
    spell_level: 3
    spell_school: 5  # Evocation
    damage_dice: 8d6
    damage_type: lightning
    spell_range: 100
    spell_duration: 0
    spell_description: A stroke of lightning forming a line 100 feet long and 5 feet wide.
    spell_components: [0, 1, 2]
    
  - spell_id: dispel_magic
    spell_name: Dispel Magic
    spell_level: 3
    spell_school: 0  # Abjuration
    spell_range: 120
    spell_duration: 0
    spell_description: End one spell on a creature or magical effect within range.
    spell_components: [0, 1]
    
  - spell_id: haste
    spell_name: Haste
    spell_level: 3
    spell_school: 8  # Transmutation
    spell_range: 30
    spell_duration: 60  # 1 minute (10 rounds)
    spell_description: Target gains doubled speed, +2 AC, advantage on Dex saves, and extra action each turn.
    spell_components: [0, 1, 2]
    
  - spell_id: mass_cure_light_wounds
    spell_name: Mass Cure Light Wounds
    spell_level: 3
    spell_school: 2  # Conjuration (Healing)
    damage_dice: 1d8+3
    damage_type: healing
    spell_range: 60
    spell_duration: 0
    spell_description: Up to 6 creatures regain 1d8+3 hit points.
    spell_components: [0, 1]
    
  - spell_id: slow
    spell_name: Slow
    spell_level: 3
    spell_school: 8  # Transmutation
    spell_range: 120
    spell_duration: 60
    spell_description: Up to 6 creatures have halved speed, -2 AC, and can't take reactions.
    spell_components: [0, 1, 2]
```

---

### Step 2: Create Level 4 Spell Data File
- **Deliverable**: `data/spells/level4.yaml` with 5-7 spells
- **Dependencies**: Step 1 (establishes pattern)
- **Goal Impact**: Enables level 7-8 character magic access
- **Acceptance**: File validates; `getSpellsByLevel(4)` returns 5+ spells
- **Validation**: `go test ./pkg/game -run TestSpellManager -v && find data/spells -name "*.yaml" | wc -l` (should be 4)

**Recommended Spells (Level 4)**: Ice Storm, Wall of Fire, Greater Invisibility, Cure Critical Wounds, Polymorph, Dimension Door

---

### Step 3: Create Level 5 Spell Data File
- **Deliverable**: `data/spells/level5.yaml` with 5-7 spells
- **Dependencies**: Steps 1-2
- **Goal Impact**: Mid-tier magic, significant combat impact
- **Acceptance**: File validates; `getSpellsByLevel(5)` returns 5+ spells
- **Validation**: `go test ./pkg/game -run TestSpellManager -v && find data/spells -name "*.yaml" | wc -l` (should be 5)

**Recommended Spells (Level 5)**: Cone of Cold, Cloudkill, Raise Dead, Teleport, Hold Monster, Dominate Person

---

### Step 4: Create Level 6 Spell Data File
- **Deliverable**: `data/spells/level6.yaml` with 5-7 spells
- **Dependencies**: Steps 1-3
- **Goal Impact**: High-tier magic access
- **Acceptance**: File validates; `getSpellsByLevel(6)` returns 5+ spells
- **Validation**: `go test ./pkg/game -run TestSpellManager -v`

**Recommended Spells (Level 6)**: Chain Lightning, Disintegrate, Globe of Invulnerability, Heal, True Seeing, Mass Suggestion

---

### Step 5: Create Level 7 Spell Data File
- **Deliverable**: `data/spells/level7.yaml` with 5-7 spells
- **Dependencies**: Steps 1-4
- **Goal Impact**: High-level magic progression
- **Acceptance**: File validates; `getSpellsByLevel(7)` returns 5+ spells
- **Validation**: `go test ./pkg/game -run TestSpellManager -v`

**Recommended Spells (Level 7)**: Delayed Blast Fireball, Finger of Death, Resurrection, Etherealness, Prismatic Spray

---

### Step 6: Create Level 8 Spell Data File
- **Deliverable**: `data/spells/level8.yaml` with 5-7 spells
- **Dependencies**: Steps 1-5
- **Goal Impact**: Near-endgame magic
- **Acceptance**: File validates; `getSpellsByLevel(8)` returns 5+ spells
- **Validation**: `go test ./pkg/game -run TestSpellManager -v`

**Recommended Spells (Level 8)**: Incendiary Cloud, Power Word Stun, Mind Blank, Holy Aura, Earthquake

---

### Step 7: Create Level 9 Spell Data File
- **Deliverable**: `data/spells/level9.yaml` with 5-7 spells
- **Dependencies**: Steps 1-6
- **Goal Impact**: Completes spell progression, endgame magic
- **Acceptance**: File validates; `getSpellsByLevel(9)` returns 5+ spells; all 9 level files exist
- **Validation**: 
  ```bash
  find data/spells -name "level*.yaml" | wc -l  # Should be 9
  go test ./pkg/game -run TestSpellManager -v
  ```

**Recommended Spells (Level 9)**: Meteor Swarm, Power Word Kill, Gate, Wish, Time Stop, Mass Heal

---

### Step 8: Extend Spell Effects for Advanced Mechanics
- **Deliverable**: Update `pkg/game/spell_effects.go` with 3-5 new effect handlers
- **Dependencies**: Steps 1-7 (spell data files)
- **Goal Impact**: Enables summoning, polymorph, enchantment mechanics
- **Acceptance**: New effect types work with effect manager; unit tests pass
- **Validation**:
  ```bash
  go test ./pkg/game -run TestSpellEffects -v
  go-stats-generator analyze ./pkg/game/spell_effects.go --format json | jq '.functions[] | select(.complexity.cyclomatic > 10)'
  # Should return empty (all functions <10 cyclomatic)
  ```

**New Effects to Implement**:
1. **EffectTypeHaste**: Apply speed boost via existing effect system
2. **EffectTypeSlow**: Apply speed penalty via existing effect system  
3. **EffectTypeDispelMagic**: Remove active effects from target
4. **EffectTypeRegeneration**: Extend HoT effect with higher values
5. **EffectTypeParalysis**: Extend Stun effect with longer duration

---

### Step 9: Create E2E Spell Progression Tests
- **Deliverable**: `test/e2e/spell_progression_test.go` with comprehensive spell casting tests
- **Dependencies**: Steps 1-8
- **Goal Impact**: Validates spell system works end-to-end
- **Acceptance**: Tests demonstrate casting spells from levels 1-9; damage/effects applied correctly
- **Validation**:
  ```bash
  go test ./test/e2e -run TestSpellProgression -v
  go test ./test/e2e -run TestSpellSchools -v
  ```

**Test Cases**:
1. Cast spell from each level (1-9), verify damage/effect
2. Verify spell saves against appropriate attributes
3. Test spell duration and effect expiration
4. Verify spell slots consumed and recovered

---

### Step 10: Update Documentation
- **Deliverable**: Update `pkg/README-RPC.md` spell section with new spell examples
- **Dependencies**: Steps 1-9
- **Goal Impact**: Developer experience, API documentation completeness
- **Acceptance**: Documentation includes examples for levels 3-9 spells; spell endpoints documented
- **Validation**: Manual review of documentation; spell API examples work as documented

---

## Quality Gates

All steps must satisfy these criteria before completion:

| Gate | Requirement | Validation Command |
|------|-------------|-------------------|
| Tests Pass | All existing tests continue passing | `go test ./... -race` |
| Coverage | Maintain ≥78% coverage | `go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out | grep total` |
| Complexity | No new functions >10 cyclomatic | `go-stats-generator analyze ./pkg/game --format json | jq '.functions[] | select(.complexity.cyclomatic > 10)'` |
| Lint | No new linter warnings | `golangci-lint run` |
| YAML Valid | All spell files parse correctly | `go run cmd/validator-demo/main.go` |

---

## Scope Assessment Calibration

Based on `go-stats-generator` metrics:

| Metric | This Project Baseline | Plan Impact | Assessment |
|--------|----------------------|-------------|------------|
| Functions above complexity 9.0 | 6 | +0 (YAML content) | ✅ Small |
| Duplication ratio | 2.0% | No change | ✅ Acceptable |
| Doc coverage gap | 13.7% (86.3% covered) | Slight improvement | ✅ Small |
| New files | 0 | +7 YAML, +1 test, +1 Go file | Medium |
| New functions | 0 | +5-10 effect handlers | Small |

**Overall Scope**: **Medium** - Primarily content creation (YAML) with small code additions.

---

## Alternative Approaches Considered

1. **Procedurally Generate Spells**: Use PCG system to generate spell data
   - Rejected: Would produce generic spells lacking D&D flavor; existing cantrips show hand-crafted quality matters

2. **Import from External Source**: Parse D&D SRD spell data
   - Rejected: Copyright concerns; existing spell format differs from SRD structure

3. **Minimal Viable Set (3 spells per level)**: Create smallest possible spell set
   - Partially adopted: 5-7 spells per level balances variety with manageable scope

---

## Next Steps After This Plan

1. **Priority 1**: Complete Asset Generation Pipeline (visual polish)
2. **Priority 4**: Build Content Creation CLI Tools (quest-builder, map-editor)
3. **Priority 6**: Implement Guild & Faction Territory (endgame content)
4. **Maintenance**: Plan WebSocket library migration (Gorilla → nhooyr.io/websocket)
5. **Maintenance**: Prepare for Go 1.24.x upgrade to resolve standard library CVEs

---

## Appendix: Existing Spell Schema Reference

From `data/spells/cantrips.yaml`:
```yaml
spells:
    - damage_dice: ""
      damage_type: ""
      spell_components:
        - 0  # Verbal
        - 1  # Somatic
      spell_description: You create a sound or image...
      spell_duration: 60
      spell_id: prestidigitation
      spell_level: 0
      spell_name: Prestidigitation
      spell_range: 10
      spell_school: 8  # Transmutation
```

**Spell Schools (pkg/game/spell_types.go)**:
- 0: Abjuration
- 1: Conjuration
- 2: Divination
- 3: Enchantment
- 4: Evocation (was 5 in existing data)
- 5: Illusion
- 6: Necromancy
- 7: Transmutation (was 8 in existing data)

**Spell Components**:
- 0: Verbal
- 1: Somatic
- 2: Material

---

*Generated: 2026-03-12 by go-stats-generator analysis*
*Project: goldbox-rpg (github.com/opd-ai/goldbox-rpg)*
