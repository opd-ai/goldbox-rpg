# Implementation Gaps — 2026-03-11

This document identifies discrepancies between the GoldBox RPG Engine's stated goals (from README.md and documentation) and the actual implementation as verified through code analysis.

---

## Gap 1: Asset Generation Pipeline Incomplete

### Stated Goal
README.md lines 87-102, 199-240, and roadmap line 399 claim:
- "Complete pipeline for generating 521 game assets"
- "Character portraits, monster sprites, item icons, terrain tiles, combat effects, UI elements"
- "YAML-based configuration with hierarchical structure"
- "Reproducible generation with seed-based randomization"
- Roadmap checkbox: "✅ Asset generation pipeline with 521 defined assets"

### Current State
**Analysis Evidence:**
- `game-assets.yaml`: 1,782 lines defining 248 asset entries (verified with `grep -E "^\s*-\s*id:" game-assets.yaml | wc -l`)
- `web/static/assets/sprites/`: Only **7 files** present (characters.png, effects.png, mindmaze1-1024x701.jpg, terrain.png, README.md, terrain.jpg, ui.png)
- Asset completion: **7/521 = 1.3%**
- Scripts exist (`make assets`, `make assets-priority`, `make assets-verify`) but require external AI image generation tool
- Documentation comprehensive (ASSET_ANALYSIS.md, ASSET_INTEGRATION.md, ASSET_PIPELINE_SUMMARY.md)

**Verification Commands:**
```bash
$ find web/static/assets/sprites -type f | wc -l
7
$ grep -E "^\s*-\s*id:" game-assets.yaml | wc -l
248
```

### Impact
- **User Experience**: Visual frontend lacks polish; game appears incomplete
- **Developer Experience**: Cannot demonstrate visual assets without 4-6 hour generation process requiring Stable Diffusion/DALL-E setup
- **Project Credibility**: README marks feature as complete (✅) despite 1.3% completion, creating false expectations
- **Documentation Mismatch**: Installation section implies assets are ready to use ("Quick Start Option: Placeholder assets are included")

### Closing the Gap

**Option A: Complete Asset Generation (High Effort)**
1. Set up Stable Diffusion or DALL-E per ASSET_INTEGRATION.md instructions
2. Run `make assets` (estimated 4-6 hours for 521 assets)
3. Verify with `make assets-verify`
4. Commit generated assets to repository or provide downloadable asset pack

**Option B: Adjust Documentation (Low Effort - Recommended)**
1. Update README.md line 399 roadmap: Change "✅ Asset generation pipeline with 521 defined assets" to "⚠️ Asset generation pipeline defined (7/521 assets generated)"
2. Add prominent warning in Installation section (after line 110):
   ```markdown
   **Asset Status**: The game includes 7 placeholder sprite files for development. Full asset generation requires external AI image generation tools (Stable Diffusion/DALL-E) and takes 4-6 hours. See [ASSET_INTEGRATION.md](./ASSET_INTEGRATION.md) for setup instructions or contact maintainers for pre-generated asset packs.
   ```
3. Update game-assets.yaml header with actual vs target asset counts

**Option C: Reduce Scope (Medium Effort)**
1. Identify minimum viable asset set for playable demo (e.g., 50 essential assets: 10 characters, 10 monsters, 15 items, 10 terrain, 5 effects)
2. Generate priority assets with `make assets-priority`
3. Update documentation to reflect realistic asset goals
4. Mark full 521 assets as "stretch goal" for community contributions

**Validation:**
```bash
# After completion:
find web/static/assets/sprites -type f | wc -l  # Should be 521
make assets-verify  # Should pass without errors
```

---

## Gap 2: Advanced NPC AI Behaviors Not Implemented

### Stated Goal
README.md lines 32, 36, 400 claim:
- "Object and NPC management with procedural generation"
- "Combat positioning and line-of-sight calculations"
- "Advanced NPC AI behaviors" (roadmap item marked for future implementation)

### Current State
**Analysis Evidence:**
- `pkg/game/world_types.go`: NPCBehavior enum exists with 4 types (Idle=0, Patrol=1, Guard=2, Aggressive=3)
- **No pathfinding implementation**: `grep -r "pathfinding\|AStar\|A\*" pkg/game` returns zero results
- **No tactical AI**: No files named `ai*.go`, `behavior*.go`, or `decision*.go` in pkg/game/
- **No behavior trees**: NPCs have behavior enum but no execution logic
- Spatial index exists (pkg/game/spatial_index.go) but unused for NPC navigation
- NPCs can be generated via PCG (pkg/pcg/character.go) with personalities but no behavioral implementation

**Verification Commands:**
```bash
$ find pkg/game -name "*ai*.go" -o -name "*pathfind*.go"
(no output - files don't exist)
$ grep -r "pathfinding\|AStar\|behavior tree" pkg/game
(no output - functionality not implemented)
```

### Impact
- **Gameplay Depth**: NPCs stand idle regardless of behavior enum setting; no tactical challenge
- **Combat Quality**: Enemies don't pursue players, seek cover, or make tactical decisions
- **World Dynamics**: Patrol and Guard behaviors non-functional; world feels static
- **Developer Expectations**: README implies NPC management exists but only data structures present

### Closing the Gap

**Phase 1: Pathfinding (~200-300 lines, 20-30 developer hours)**
1. Create `pkg/game/pathfinding.go` with A* algorithm
2. Leverage existing `pkg/game/spatial_index.go` for efficient neighbor queries
3. Integrate with `pkg/game/world_types.go` NPC struct
4. Support obstacle avoidance using tile passability from `pkg/game/map.go`
5. Add unit tests with table-driven pattern (test cases: straight path, obstacles, unreachable target)

**Implementation Pattern:**
```go
// pkg/game/pathfinding.go
type Pathfinder struct {
    world *World
    spatialIndex *SpatialIndex
}

func (pf *Pathfinder) FindPath(start, goal Position, maxDistance int) ([]Position, error) {
    // A* implementation leveraging spatial index
    // Return path or error if unreachable
}
```

**Phase 2: Combat AI (~400-500 lines, 40-50 developer hours)**
1. Create `pkg/game/ai_combat.go` with tactical decision-making
2. Implement target selection logic (prioritize low HP, casters, player character)
3. Add ability usage decisions (heal when HP <30%, use area attack when 3+ enemies clustered)
4. Integrate with existing `pkg/server/combat.go` CombatState and TurnManager
5. Support difficulty tiers (Easy: random actions, Medium: basic tactics, Hard: optimal decisions)
6. Add E2E tests in `test/e2e/combat_test.go` demonstrating NPC tactical behavior

**Phase 3: Behavior Trees (~300-400 lines, 30-40 developer hours)**
1. Create `pkg/game/ai_behaviors.go` with composable behavior tree nodes
2. Support conditions (health threshold, distance check, ally count)
3. Support actions (move to position, attack target, flee, patrol route)
4. Enable YAML-based behavior definitions in `data/ai/behaviors.yaml`
5. Link to existing NPCBehavior enum as behavior tree selector
6. Add validation via `pkg/validation/` for behavior definitions

**Example Behavior YAML:**
```yaml
behaviors:
  - id: aggressive
    root:
      type: selector
      children:
        - type: sequence  # Attack if enemy in range
          children:
            - condition: enemy_in_range(10)
            - action: attack_nearest_enemy
        - type: sequence  # Otherwise move toward enemy
          children:
            - condition: enemy_visible
            - action: move_toward_enemy
```

**Validation:**
```bash
# After implementation:
go test ./pkg/game -run TestPathfinding -v
go test ./test/e2e -run TestNPCTacticalCombat -v
# Should demonstrate NPC using pathfinding to navigate around obstacles,
# selecting optimal targets, and executing behavior tree actions
```

**Success Criteria:**
- NPC navigates around obstacles using A* pathfinding
- Guard NPC patrols defined route and engages approaching enemies
- Aggressive NPC pursues player, prioritizes low-HP targets, flees when outnumbered
- E2E test shows 5+ NPCs with different behaviors executing tactical combat

---

## Gap 3: Spell System Content Incomplete (Levels 3-9 Missing)

### Stated Goal
README.md lines 48-50 claim:
- "Spell casting"
- "Spell system" (listed under "Event System" as implemented feature)
- Spell queries: `getSpell`, `getSpellsByLevel`, `getSpellsBySchool` in pkg/README-RPC.md

### Current State
**Analysis Evidence:**
- Spell manager implemented: `pkg/game/spell_manager.go` supports 9 spell levels and 9 schools
- Spell data files: **Only 3 files** exist in data/spells/
  - `cantrips.yaml`: 31 lines (~3 spells: Light, Mage Hand, Prestidigitation)
  - `level1.yaml`: 34 lines (~3 spells: Magic Missile, Cure Light Wounds, Shield)
  - `level2.yaml`: 37 lines (~3 spells: Fireball, Cure Moderate Wounds, Invisibility)
- **Levels 3-9 completely missing**: No level3.yaml through level9.yaml files
- Total spell content: ~9 spells across 3 levels vs expected 50-90 spells for basic D&D-style gameplay
- Spell effect system exists (`pkg/game/spell_effects.go`) but underutilized

**Verification Commands:**
```bash
$ find data/spells -name "*.yaml" -type f
data/spells/level2.yaml
data/spells/cantrips.yaml
data/spells/level1.yaml

$ wc -l data/spells/*.yaml
  31 data/spells/cantrips.yaml
  34 data/spells/level1.yaml
  37 data/spells/level2.yaml
 102 total
```

### Impact
- **Character Viability**: Mage and Cleric classes severely limited; cannot progress beyond level 2 spells
- **Gameplay Depth**: Magical combat lacks variety and strategic options
- **Level Progression**: Character advancement hollow without new spell access
- **API Completeness**: `getSpellsByLevel(3-9)` and `getSpellsBySchool` queries return empty results

### Closing the Gap

**Phase 1: Create Spell Data Files (~20-30 hours total)**

Create `data/spells/level3.yaml` through `data/spells/level9.yaml` following existing structure:

**Recommended Spell Distribution:**
- Level 3: 5-7 spells (Lightning Bolt, Dispel Magic, Fireball upgrade, Mass Cure Light Wounds, Haste, Slow, Fly)
- Level 4: 5-7 spells (Ice Storm, Wall of Fire, Greater Invisibility, Cure Critical Wounds, Polymorph, Dimension Door)
- Level 5: 5-7 spells (Cone of Cold, Cloudkill, Raise Dead, Teleport, Hold Monster, Dominate Person)
- Level 6: 5-7 spells (Chain Lightning, Disintegrate, Globe of Invulnerability, Heal, True Seeing, Mass Suggestion)
- Level 7: 5-7 spells (Delayed Blast Fireball, Finger of Death, Resurrection, Etherealness, Prismatic Spray)
- Level 8: 5-7 spells (Incendiary Cloud, Power Word Stun, Mind Blank, Holy Aura, Earthquake)
- Level 9: 5-7 spells (Meteor Swarm, Power Word Kill, Gate, Wish, Time Stop, Mass Heal)

**Total**: 35-49 new spells (minimum viable spell set for 9-level progression)

**YAML Template (following cantrips.yaml pattern):**
```yaml
spells:
    - damage_dice: 8d6
      damage_type: lightning
      spell_components:
        - 0  # Verbal
        - 1  # Somatic
        - 2  # Material
      spell_description: A stroke of lightning forming a line 100 feet long and 5 feet wide.
      spell_duration: 0  # Instantaneous
      spell_id: lightning_bolt
      spell_level: 3
      spell_name: Lightning Bolt
      spell_range: 100
      spell_school: 5  # Evocation
```

**Phase 2: Extend Spell Effects (~200-300 lines)**

Update `pkg/game/spell_effects.go` to support advanced spell mechanics:
1. **Summoning**: Create temporary NPC entities (Conjuration school)
2. **Polymorph**: Transform character stats temporarily
3. **Illusions**: Apply concealment/advantage effects (Illusion school)
4. **Enchantments**: Mind control and charm mechanics (Enchantment school)
5. **Teleportation**: Position manipulation (Conjuration school)

Integrate with existing effect manager (`pkg/game/effectmanager.go`) for duration, stacking, and dispelling.

**Phase 3: Spell Resistance & Saving Throws (~100-150 lines)**

Add spell resistance mechanics in `pkg/game/spell_resistance.go`:
1. Calculate save DC based on caster Intelligence/Wisdom/Charisma (from `pkg/game/character.go`)
2. Target rolls saving throw using appropriate attribute (Dexterity for area spells, Wisdom for mind effects)
3. Determine success/failure and half-damage on save where applicable
4. Integrate with combat system in `pkg/server/combat.go`

**Phase 4: Testing & Integration**

Create E2E tests in `test/e2e/spell_test.go`:
```go
func TestSpellProgression(t *testing.T) {
    // Test casting spells from levels 1-9
    // Verify damage calculation, saving throws, spell effects
}

func TestSpellSchools(t *testing.T) {
    // Test representative spell from each of 9 schools
    // Verify school-specific mechanics (Evocation damage, Conjuration summons, etc.)
}
```

**Validation:**
```bash
# After implementation:
find data/spells -name "level*.yaml" | wc -l  # Should be 9
go test ./test/e2e -run TestSpellProgression -v
go test ./pkg/game -run TestSpellEffects -v

# API queries should return results:
curl -X POST http://localhost:8080/rpc \
  -d '{"jsonrpc":"2.0","method":"getSpellsByLevel","params":{"level":5},"id":1}'
# Should return 5-7 level 5 spells
```

**Success Criteria:**
- All 9 spell level files present (cantrips + levels 1-9)
- Minimum 40 total spells (average 4-5 per level)
- Advanced spell effects functional (summoning, polymorph, teleportation)
- E2E tests demonstrate spell casting for all levels
- RPC endpoints return complete spell data for all queries

---

## Gap 4: Enhanced Combat Mechanics Not Implemented

### Stated Goal
README.md line 400 roadmap:
- "Enhanced combat mechanics" (marked for future implementation)
- Line 36: "Combat positioning and line-of-sight calculations" (implies tactical mechanics)

### Current State
**Analysis Evidence:**
- Turn-based combat exists: `pkg/server/combat.go` (30,170 bytes, 800+ lines)
- Basic functionality implemented:
  - Initiative tracking (TurnManager struct)
  - Combat rounds and turn sequencing
  - Active combatants list
  - Status effects tracking
  - Turn timers (10-second default)
- **Missing tactical mechanics**: `grep -i "opportunity\|flanking\|cover\|morale" pkg/server/combat.go` returns zero results
- Line-of-sight calculations mentioned but no evidence of cover/concealment mechanics
- Positioning tracked but not used for tactical advantages

**Verification Commands:**
```bash
$ wc -l pkg/server/combat.go
800+ lines

$ grep -i "opportunity\|flanking\|cover\|morale" pkg/server/combat.go
(no output - mechanics not implemented)
```

### Impact
- **Tactical Depth**: Combat lacks strategic positioning elements expected in Gold Box-inspired games
- **Feature Parity**: Original Gold Box games (Pool of Radiance, etc.) had facing, flanking, and morale
- **Player Engagement**: Combat becomes repetitive without tactical decision-making
- **Positioning Wasted**: Spatial index and positioning system underutilized

### Closing the Gap

**Phase 1: Opportunity Attacks (~150 lines, 15-20 hours)**

Create `pkg/game/combat_opportunity.go`:

```go
type OpportunityAttackManager struct {
    world *World
    combatState *CombatState
}

// TriggerOpportunityAttacks checks if movement provokes attacks
func (oam *OpportunityAttackManager) TriggerOpportunityAttacks(mover *game.Character, from, to game.Position) []OpportunityAttack {
    // 1. Find all enemies adjacent to 'from' position
    // 2. Check if movement leaves their reach without Disengage action
    // 3. Roll opportunity attacks for each eligible enemy
    // 4. Apply damage and effects
    // Return list of triggered attacks for logging
}
```

Integration points:
- Modify `pkg/server/handlers.go` handleMove() to call TriggerOpportunityAttacks before moving
- Add "Disengage" action to movement options (prevents opportunity attacks but costs action)
- Use existing spatial index (`pkg/game/spatial_index.go`) to find adjacent enemies efficiently

**Phase 2: Cover & Flanking (~200 lines, 20-25 hours)**

Create `pkg/game/combat_modifiers.go`:

```go
type CombatModifierCalculator struct {
    world *World
    spatialIndex *SpatialIndex
}

// CalculateCoverBonus determines AC bonus from cover
func (cmc *CombatModifierCalculator) CalculateCoverBonus(attacker, defender game.Position) int {
    // 1. Trace line from attacker to defender
    // 2. Check terrain tiles along path for cover-providing tiles (walls, trees)
    // 3. Return AC bonus: +2 for half cover, +5 for three-quarters cover
}

// CalculateFlankingBonus determines attack bonus from flanking
func (cmc *CombatModifierCalculator) CalculateFlankingBonus(attacker, target game.Position, allies []game.Position) int {
    // 1. Check if 2+ allies adjacent to target on opposite sides
    // 2. Calculate angle between attacker and allies relative to target
    // 3. Return +2 attack bonus if flanking condition met
}
```

Integration points:
- Modify `pkg/server/combat.go` attack calculations to apply CalculateCoverBonus() to AC
- Add CalculateFlankingBonus() to attack rolls when multiple allies present
- Use existing terrain tile data from `pkg/game/map.go` for cover determination

**Phase 3: Morale System (~250 lines, 25-30 hours)**

Create `pkg/game/morale.go`:

```go
type MoraleManager struct {
    characters map[string]*MoraleState
}

type MoraleState struct {
    CurrentMorale int  // 0-100 scale
    BreakThreshold int  // Morale level triggering flee
    IsBroken bool
}

// UpdateMorale modifies morale based on combat events
func (mm *MoraleManager) UpdateMorale(characterID string, event MoraleEvent) {
    // Events:
    // - AllyDeath: -15 morale
    // - TakeDamage: -(damage / max_hp * 20) morale
    // - EnemyDeath: +10 morale
    // - Outnumbered: -5 morale per turn
    // - Victory: +20 morale
}

// CheckMoraleBreak determines if character flees
func (mm *MoraleManager) CheckMoraleBreak(characterID string) bool {
    state := mm.characters[characterID]
    if state.CurrentMorale < state.BreakThreshold {
        state.IsBroken = true
        return true
    }
    return false
}
```

Integration points:
- Call UpdateMorale() from `pkg/server/combat.go` on combat events (damage dealt, character death)
- Check CheckMoraleBreak() each turn; if broken, NPC enters flee state
- Use Wisdom and Charisma attributes for BreakThreshold calculation (higher mental stats = higher morale resistance)
- Integrate with Gap 2 (NPC AI) pathfinding for flee behavior

**Phase 4: Testing & Documentation**

Create E2E tests in `test/e2e/combat_test.go`:

```go
func TestOpportunityAttacks(t *testing.T) {
    // Setup: NPC adjacent to player
    // Action: Player moves away without Disengage
    // Verify: NPC executes opportunity attack
    // Verify: Player takes damage
}

func TestCoverBonus(t *testing.T) {
    // Setup: Defender behind wall terrain
    // Action: Attacker attempts ranged attack
    // Verify: Defender receives +2 AC bonus from half cover
}

func TestFlankingBonus(t *testing.T) {
    // Setup: 2 allies on opposite sides of enemy
    // Action: Ally attacks flanked enemy
    // Verify: +2 attack bonus applied
}

func TestMoraleBreak(t *testing.T) {
    // Setup: NPC in combat, allies dying
    // Action: Trigger morale events (ally deaths, damage)
    // Verify: NPC morale drops below threshold and flees
}
```

Update `pkg/README-RPC.md` to document new combat mechanics and their effects on RPC methods.

**Validation:**
```bash
# After implementation:
go test ./test/e2e -run TestOpportunityAttacks -v
go test ./test/e2e -run TestCoverBonus -v
go test ./test/e2e -run TestFlankingBonus -v
go test ./test/e2e -run TestMoraleBreak -v

# All tests should pass demonstrating tactical combat features
```

**Success Criteria:**
- Opportunity attacks trigger when enemy moves through threatened area without Disengage
- Cover bonuses apply based on terrain tiles (+2 for half cover, +5 for three-quarters cover)
- Flanking bonuses apply when allies positioned on opposite sides (+2 attack bonus)
- Morale system causes NPCs to flee when morale drops below threshold (based on Wisdom/Charisma)
- E2E tests validate all three systems working together in tactical combat scenario

---

## Gap 5: World Editor Tools Missing

### Stated Goal
README.md line 401 roadmap:
- "World editor tools" (marked for future implementation)
- Line 403: "Content creation utilities"

### Current State
**Analysis Evidence:**
- No editor code in `cmd/` directory (only server, demos: dungeon-demo, events-demo, metrics-demo, validator-demo, pcg-demo, bootstrap-demo)
- No GUI applications found in repository
- No CLI tools for map creation, quest authoring, or content management
- Content creation requires:
  - Manual YAML editing in `data/` directory (spells, items, PCG templates)
  - Go programming for complex content (PCG systems, custom mechanics)
- PCG system exists (`pkg/pcg/`: 20 files, 503 functions) but no user-facing tools

**Verification Commands:**
```bash
$ ls cmd/
bootstrap-demo  dungeon-demo  events-demo  metrics-demo  openapi-gen  pcg-demo  server  validator-demo  wasm-ui

$ find cmd -name "*editor*" -o -name "*builder*" -o -name "*creator*"
(no output - no editor tools exist)
```

### Impact
- **Accessibility**: Game designers and modders without Go knowledge cannot create content
- **Development Speed**: Manual YAML editing is slow and error-prone
- **Community Growth**: High barrier to entry for content contributions
- **Feature Claims**: README claims "content creation utilities" but only provides programmatic APIs

### Closing the Gap

**Option A: CLI Tools (Recommended - 40-60 hours)**

Create three command-line tools following existing `cmd/` patterns:

**1. Quest Builder (`cmd/quest-builder/main.go`, ~500 lines)**
```go
// Interactive CLI for creating quest YAML files
// Example usage:
// $ go run cmd/quest-builder/main.go
// > Quest Title: Retrieve the Lost Amulet
// > Quest Type: [fetch/kill/escort/explore]: fetch
// > Objective 1: Find the Ancient Amulet in the Dungeon
// > Objective 2: (Enter to finish)
// > Reward Gold: 100
// > Reward XP: 500
// > Prerequisites: (quest IDs, comma-separated):
// > Saving to data/quests/retrieve_amulet.yaml...
```

Features:
- Guided prompts for quest objectives, rewards, prerequisites, narrative text
- Validation against quest schema before saving
- Template selection (fetch, kill, escort, explore quests)
- Auto-generate quest IDs from title
- List existing quests for prerequisite selection

**2. Map Editor (`cmd/map-editor/main.go`, ~600 lines)**
```go
// ASCII-based tile placement for creating custom maps
// Example usage:
// $ go run cmd/map-editor/main.go --new dungeon_level1
// Map Name: Goblin Cave Level 1
// Width (in tiles): 20
// Height (in tiles): 15
// 
// [Tile palette: . = floor, # = wall, ~ = water, ^ = mountain, T = tree]
// Row 1: ####################
// Row 2: #..................#
// Row 3: #.....###..........#
// ...
// [Commands: s=save, l=load, p=place objects, q=quit]
```

Features:
- ASCII tile placement with visual preview
- Terrain type selection (floor, wall, water, mountain, forest, etc.)
- Object placement (NPCs, items, spawn points)
- Import/edit existing maps from `data/maps/`
- Export to YAML format compatible with `pkg/game/map.go`
- Validation of map connectivity (all floor tiles reachable)

**3. Content Creator (`cmd/content-creator/main.go`, ~400 lines)**
```go
// Template-driven creation of spell/item YAML files
// Example usage:
// $ go run cmd/content-creator/main.go --type spell
// Content Type: [spell/item]: spell
// Spell Name: Acid Arrow
// Spell Level [0-9]: 2
// School [Abjuration/Conjuration/.../Transmutation]: Evocation
// Damage Type [physical/fire/frost/poison/lightning/force]: poison
// Damage Dice: 2d4
// Range (feet): 90
// Duration (rounds, 0=instant): 0
// Components [V]erbal [S]omatic [M]aterial: VSM
// Description: A shimmering green arrow streaks toward a target...
// Saving to data/spells/level2.yaml...
```

Features:
- Template-driven creation for spells, items, NPCs
- Dropdown/enum selection for schools, damage types, rarity using existing constants
- Validation via `pkg/validation/` before saving
- Append to existing YAML files or create new
- Preview generated YAML before saving

**Implementation Pattern:**
- Use `github.com/spf13/cobra` for CLI framework (common in Go projects)
- Leverage existing types from `pkg/game/`, `pkg/pcg/quests/`, `pkg/game/map.go`
- Reuse validation logic from `pkg/validation/`
- Follow `cmd/dungeon-demo/` and `cmd/validator-demo/` as structural examples
- Add smoke tests in `.github/workflows/ci.yml`: `go run cmd/quest-builder/main.go --help`

**Documentation:**
Create `docs/CONTENT_CREATION.md`:
- Tool usage guide with screenshots/examples
- YAML schema reference for manual editing
- Examples from `data/spells/cantrips.yaml`, `data/items/`, `data/pcg/`
- Best practices for quest design, map layout, spell balancing

**Option B: Visual GUI Tools (Deferred - High Effort)**
- Web-based map editor using React/Vue + Canvas API
- Visual quest graph editor with node-based workflow
- Sprite asset editor with preview and metadata
- **Recommendation**: Defer until CLI tools demonstrate demand and feasibility; GUI adds significant complexity

**Validation:**
```bash
# After implementation:
$ ls cmd/{quest-builder,map-editor,content-creator}/main.go
cmd/quest-builder/main.go
cmd/map-editor/main.go
cmd/content-creator/main.go

$ go run cmd/quest-builder/main.go --help
Quest Builder - Create YAML quest files interactively
Usage: quest-builder [flags]
...

$ go run cmd/quest-builder/main.go
(interactive session creating data/quests/test_quest.yaml)

$ ls data/quests/test_quest.yaml
data/quests/test_quest.yaml  # New quest file created

$ go test ./... | grep "content-creator"
ok  	goldbox-rpg/cmd/content-creator	0.123s  # Smoke tests pass
```

**Success Criteria:**
- Three CLI tools exist in `cmd/` directory with `--help` documentation
- Quest builder creates valid YAML files loadable by `pkg/pcg/quests/`
- Map editor generates maps compatible with `pkg/game/map.go`
- Content creator produces spell/item YAML matching `data/spells/` schema
- docs/CONTENT_CREATION.md provides comprehensive usage guide
- CI smoke tests verify tools compile and display help

---

## Gap 6: Guild Membership and Faction Territory Control Incomplete

### Stated Goal
README.md line 404 roadmap:
- "Guild and faction systems" (marked for future implementation)

### Current State
**Analysis Evidence:**
- Faction generation exists: `pkg/pcg/faction.go` (303 lines, GenerateFaction function)
- Reputation system implemented: `pkg/pcg/reputation.go` (345 lines, dynamic reputation effects, decay mechanics)
- **TODO at line 31**: `pkg/pcg/faction.go:31` states "TODO: Implement territory generation based on faction power and world geography"
- No guild membership mechanics found (grep "guild\|Guild" in pkg/game returned only reputation references)
- No faction territory control system
- Reputation is player-to-faction only (no inter-faction relationships)
- No guild quests, guild halls, or rank progression

**Verification Commands:**
```bash
$ grep -n "TODO" pkg/pcg/faction.go
31:	// TODO: Implement territory generation based on faction power and world geography

$ grep -r "GuildMembership\|guild_rank" pkg/game
(no output - guild membership not implemented)

$ grep -r "TerritoryControl\|faction_territory" pkg/pcg
(no output - territory control not implemented)
```

### Impact
- **Endgame Content**: Players lack long-term progression goals (guild ranks, faction wars)
- **World Dynamics**: Factions exist but don't control territory or interact with each other
- **Multiplayer Engagement**: No guild-based cooperative gameplay
- **Feature Completeness**: Reputation system exists but lacks context without territory/guilds

### Closing the Gap

**Phase 1: Complete Faction Territory Generation (~400 lines, 40 hours)**

Complete TODO at `pkg/pcg/faction.go:31` by creating `pkg/pcg/faction_territory.go`:

```go
type TerritoryGenerator struct {
    world *game.World
    factionGen *FactionGenerator
    terrainGen *terrain.Generator
}

type FactionTerritory struct {
    FactionID string
    Borders []game.Position  // Territory boundary coordinates
    ControlPoints []ControlPoint  // Strategic locations (cities, fortresses)
    Power int  // Faction's territorial influence (0-100)
    ContestedWith []string  // IDs of factions contesting borders
}

func (tg *TerritoryGenerator) GenerateTerritories(factions []GeneratedFaction, worldMap *game.World) (map[string]*FactionTerritory, error) {
    // 1. Assign starting positions based on faction power and world geography
    // 2. Expand territories using flood-fill weighted by faction power
    // 3. Create natural borders along rivers, mountains (from terrain biomes)
    // 4. Mark contested zones where territories overlap
    // 5. Generate control points (cities, fortresses) within territories
    // 6. Validate territory connectivity and playability
}
```

Integration:
- Call from `pkg/pcg/world.go` GenerateWorld() after faction generation
- Use terrain biomes from `pkg/pcg/terrain/` as natural barriers
- Store territory data in WorldState for persistence
- Add RPC endpoint `getFactionTerritories` in `pkg/server/handlers.go`

**Phase 2: Guild Membership System (~300 lines, 30 hours)**

Create `pkg/game/guild.go`:

```go
type Guild struct {
    ID string
    Name string
    FactionID string  // Guild's faction allegiance
    Ranks []GuildRank
    Members map[string]*GuildMember  // PlayerID -> Member data
    GuildHallLocation game.Position
    Treasury int  // Shared guild resources
}

type GuildRank struct {
    Level int
    Title string  // "Initiate", "Member", "Officer", "Leader"
    RequiredReputation int
    Permissions []string  // "invite", "promote", "access_treasury"
}

type GuildMember struct {
    PlayerID string
    RankLevel int
    JoinDate time.Time
    Contributions int  // Quest completions, donations for rank progression
}

func (g *Guild) Join(playerID string) error {
    // Add player as Initiate rank (level 0)
}

func (g *Guild) PromoteRank(playerID string) error {
    // Check contributions and reputation
    // Advance to next rank if eligible
}

func (g *Guild) GetGuildQuests() []*game.Quest {
    // Return guild-specific quests from faction allegiance
}
```

Integration:
- Add guild membership to `pkg/game/character.go` Character struct
- Create guild quests in `pkg/pcg/quests/` linked to faction allegiance
- Add RPC endpoints: `joinGuild`, `leaveGuild`, `promoteGuildMember`, `getGuildQuests`
- Store guild data in persistence layer (`pkg/persistence/`)

**Phase 3: Inter-Faction Diplomacy (~250 lines, 25 hours)**

Create `pkg/game/faction_relations.go`:

```go
type FactionRelationsManager struct {
    relations map[string]map[string]*DiplomaticState
}

type DiplomaticState struct {
    FactionA string
    FactionB string
    Relationship int  // -100 (war) to 100 (alliance)
    State DiplomaticStatus  // War, Neutral, Trade, Alliance
    TreatiesSigned []Treaty
}

type DiplomaticStatus int
const (
    War DiplomaticStatus = iota
    Hostile
    Neutral
    Friendly
    Trade
    Alliance
)

func (frm *FactionRelationsManager) UpdateRelation(factionA, factionB string, delta int, reason string) {
    // Modify relationship score based on events:
    // - Player actions (quest completion for faction A hurts faction B)
    // - Territory conflicts (border skirmishes)
    // - Trade agreements (+10 for trade treaty)
    // - War declarations (-100, sets State=War)
}

func (frm *FactionRelationsManager) GetAlliedFactions(factionID string) []string {
    // Return factions with Alliance state
}

func (frm *FactionRelationsManager) GetEnemyFactions(factionID string) []string {
    // Return factions with War or Hostile state
}
```

Integration:
- Integrate with reputation system in `pkg/pcg/reputation.go`
- Player reputation changes with faction A affect relations with allied/enemy factions
- NPC behavior (from Gap 2 AI) considers faction diplomacy (attack enemy factions on sight)
- Add diplomatic events to `pkg/game/events.go` (WarDeclared, AllianceFormed)
- RPC endpoints: `getFactionRelations`, `getDiplomaticState`

**Phase 4: Testing & Documentation**

Create E2E tests in `test/e2e/faction_test.go`:

```go
func TestFactionTerritoryGeneration(t *testing.T) {
    // Generate world with 3 factions
    // Verify each faction has territory assigned
    // Verify territories don't overlap excessively (contested zones <20%)
    // Verify natural borders follow terrain (rivers, mountains)
}

func TestGuildMembership(t *testing.T) {
    // Player joins guild
    // Complete guild quests
    // Verify rank progression (Initiate -> Member -> Officer)
    // Verify permissions change with rank
}

func TestFactionDiplomacy(t *testing.T) {
    // Establish alliance between factions A and B
    // Player increases reputation with faction A
    // Verify reputation with faction B also increases (allied bonus)
    // Declare war between factions B and C
    // Verify faction B NPCs attack faction C NPCs on sight
}
```

Update `pkg/README-RPC.md` with new faction/guild RPC endpoints.

**Validation:**
```bash
# After implementation:
go test ./test/e2e -run TestFactionTerritoryGeneration -v
go test ./test/e2e -run TestGuildMembership -v
go test ./test/e2e -run TestFactionDiplomacy -v

# RPC queries should work:
curl -X POST http://localhost:8080/rpc \
  -d '{"jsonrpc":"2.0","method":"getFactionTerritories","params":{},"id":1}'
# Should return territory data for all factions

curl -X POST http://localhost:8080/rpc \
  -d '{"jsonrpc":"2.0","method":"joinGuild","params":{"guild_id":"merchants_guild","player_id":"p1"},"id":1}'
# Should add player to guild with Initiate rank
```

**Success Criteria:**
- TODO at `pkg/pcg/faction.go:31` resolved with working territory generation
- Faction territories generated with natural borders (rivers, mountains) and control points
- Guild membership system allows join/leave, rank progression, and guild quests
- Inter-faction diplomacy affects player reputation and NPC behavior
- E2E tests demonstrate territory control, guild operations, and diplomatic state changes
- RPC endpoints provide access to faction/guild data

---

## Summary Table

| Gap | Current Completion | Estimated Effort | Priority | Blocking? |
|-----|-------------------|------------------|----------|-----------|
| Asset Generation Pipeline | 1.3% (7/521) | 4-40 hrs (depends on approach) | HIGH | User experience |
| NPC AI Behaviors | 0% (enum only) | 60-80 hrs | HIGH | Gameplay depth |
| Spell Content (Levels 3-9) | 33% (3/9 levels) | 20-30 hrs | HIGH | Character progression |
| Enhanced Combat Mechanics | 0% (basic only) | 30-40 hrs | MEDIUM | Tactical gameplay |
| World Editor Tools | 0% | 40-60 hrs | MEDIUM | Developer experience |
| Guild & Faction Territory | 30% (reputation only) | 40-50 hrs | MEDIUM | Endgame content |

**Total Estimated Effort to Close All Gaps**: 194-300 developer hours (approximately 5-8 weeks full-time)

---

## Recommendations

### Immediate Actions (Week 1)
1. **Update README.md** to accurately reflect asset status (change ✅ to ⚠️ in line 399)
2. **Document YAML-based content creation** in docs/CONTENT_CREATION.md for manual spell/item creation
3. **Prioritize spell content** (Gap 3) - low complexity, high gameplay value, can be completed in 20-30 hours

### Short-Term (Weeks 2-4)
4. **Implement NPC AI** (Gap 2) - unlocks dynamic gameplay, highest impact on user experience
5. **Create CLI content tools** (Gap 5) - lowers barrier to entry for community contributions
6. **Add enhanced combat mechanics** (Gap 4) - builds on NPC AI, increases tactical depth

### Medium-Term (Weeks 5-8)
7. **Complete guild/faction systems** (Gap 6) - requires NPC AI for faction NPCs
8. **Address asset generation** (Gap 1) - generate priority assets or provide downloadable pack
9. **Optimize network** (only if benchmarks justify) - defer until scale requires

### Community Engagement
- Mark incomplete features clearly in roadmap to set accurate expectations
- Create GitHub issues for each gap with "help wanted" labels
- Provide contribution guidelines for spell data, maps, and content
- Consider pre-generated asset pack distribution to bypass generation requirement

---

**Document Version**: 2026-03-11  
**Analysis Tool**: go-stats-generator v1.0.0, manual code inspection, grep/find analysis  
**Codebase Version**: Last updated 2025-10-29 per README.md line 409
