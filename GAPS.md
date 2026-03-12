# Implementation Gaps — 2026-03-12

## Gap 1: Asset Generation Pipeline Incomplete

**Stated Goal:** "Asset generation pipeline with 521 defined assets" marked as ✅ complete in roadmap (README line 399)

**Current State:** 
- YAML configuration complete: `game-assets.yaml` (1,782 lines) defines all 521 assets across 6 categories (characters, monsters, items, terrain, effects, UI)
- Generation scripts exist: `make assets`, `make assets-priority`, `make assets-verify`, `make assets-clean`
- Comprehensive documentation: ASSET_ANALYSIS.md, ASSET_INTEGRATION.md, ASSET_PIPELINE_SUMMARY.md
- **Only 6 placeholder PNG files exist** in `web/static/assets/sprites/` (verified with `find web/static/assets -type f -name "*.png" | wc -l`)
- Asset categories defined: characters (80), monsters (120), items (150), terrain (80), effects (60), UI (31)

**Impact:** 
- **User-Facing:** Game appears unpolished with placeholder graphics; professional appearance blocked
- **Development:** Developers must use generic sprites during testing; visual debugging difficult
- **Production:** Cannot deploy with production-quality assets; requires 4-6 hours of external AI tool setup (Stable Diffusion/DALL-E per ASSET_INTEGRATION.md)
- **Documentation Accuracy:** README roadmap misleading with ✅ checkmark suggesting completion

**Closing the Gap:**

**Option A: Complete Full Asset Generation** (4-6 hours runtime)
1. Install external AI image generation tool following ASSET_INTEGRATION.md
   - Stable Diffusion with automatic1111 webui, OR
   - DALL-E API access with credits, OR
   - MidJourney API integration
2. Configure tool endpoints in asset generation pipeline
3. Run `make assets` to generate all 521 assets
4. Run `make assets-verify` to confirm completeness
5. Optionally run `make assets-optimize` for production deployment (compress PNGs, create sprites sheets)
6. Update README badge to show "521/521 assets (100%)"

**Option B: Source Pre-Generated Asset Pack**
1. Contact repository maintainer (@opd-ai) for pre-generated asset pack (mentioned in README line 239)
2. Extract assets to `web/static/assets/` directory structure
3. Verify with `make assets-verify`
4. Update README to document asset pack source and licensing

**Option C: Document Placeholder Status** (Recommended for short-term)
1. Change README roadmap line 399 from ✅ to ⚠️: `- ⚠️ **Asset generation pipeline with 521 defined assets** (pipeline complete, 6/521 assets generated)`
2. Add badge to README: `![Assets](https://img.shields.io/badge/assets-6%2F521-orange)` (1% complete)
3. Create ASSETS_STATUS.md documenting:
   - Current asset count: 6 placeholders
   - Pipeline readiness: 100% (scripts, config, docs complete)
   - Generation requirements: External AI tool + 4-6 hours
   - Quick start instructions for using placeholders
4. Update installation instructions to clarify game is "fully functional with placeholder assets"

**Validation:**
```bash
# After generation
make assets-verify  # Should pass with 521 assets found
find web/static/assets -type f \( -name "*.png" -o -name "*.jpg" -o -name "*.svg" \) | wc -l  # Should output 521

# Visual inspection
ls -lh web/static/assets/sprites/characters/  # Should show 80 character sprites
ls -lh web/static/assets/sprites/monsters/   # Should show 120 monster sprites
```

**Estimated Effort:** 
- Option A: 6-8 hours (4-6 generation + 2 setup)
- Option B: 1-2 hours (depending on asset pack availability)
- Option C: 30 minutes (documentation update)

---

## Gap 2: Spatial Index Deadlock Bug

**Stated Goal:** "Advanced spatial indexing (R-tree-like structure for efficient queries)" for world management

**Current State:**
- Spatial index implemented in `pkg/game/spatial_index.go` (422 lines)
- R-tree-like quadtree structure with Rectangle bounds and SpatialNode children
- **CRITICAL BUG:** `GetNearestObjects()` (line 120-150) holds RLock and calls `GetObjectsInRadius()` which attempts to acquire another RLock
- Results in recursive lock acquisition → deadlock
- Affects: combat targeting (find nearest enemy), AOE abilities (targets in radius), pathfinding (nearest walkable tile)

**Impact:**
- **Gameplay Breaking:** Any call to `GetNearestObjects()` hangs indefinitely, freezing combat system
- **Combat System:** Cannot select nearest targets for melee attacks or spells
- **AI Pathfinding:** NPC AI cannot find nearest player to pursue
- **Spatial Queries:** All proximity-based game mechanics broken

**Closing the Gap:**

**Step 1: Refactor to Unlocked Helper Method**
```go
// pkg/game/spatial_index.go:85-118
// Extract filtering logic from GetObjectsInRadius into unlocked helper
func (si *SpatialIndex) getObjectsInRadiusUnlocked(center Position, radius float64) []GameObject {
    // Use tighter bounding box to reduce candidates
    radiusInt := int(radius)
    rect := Rectangle{
        MinX: center.X - radiusInt,
        MinY: center.Y - radiusInt,
        MaxX: center.X + radiusInt,
        MaxY: center.Y + radiusInt,
    }

    var candidates []GameObject
    si.queryNode(si.root, rect, &candidates)

    // Filter by actual circular distance
    radiusSquared := radius * radius
    result := make([]GameObject, 0, len(candidates))

    for _, obj := range candidates {
        objPos := obj.GetPosition()
        dx := float64(center.X - objPos.X)
        dy := float64(center.Y - objPos.Y)
        distanceSquared := dx*dx + dy*dy

        if distanceSquared <= radiusSquared {
            result = append(result, obj)
        }
    }

    return result
}

// Update GetObjectsInRadius to use helper
func (si *SpatialIndex) GetObjectsInRadius(center Position, radius float64) []GameObject {
    si.mu.RLock()
    defer si.mu.RUnlock()
    return si.getObjectsInRadiusUnlocked(center, radius)
}

// Update GetNearestObjects to call unlocked helper
func (si *SpatialIndex) GetNearestObjects(center Position, k int) []GameObject {
    si.mu.RLock()
    defer si.mu.RUnlock()

    radius := float64(si.cellSize)
    maxRadius := float64(maxInt(si.bounds.MaxX-si.bounds.MinX, si.bounds.MaxY-si.bounds.MinY))

    for radius <= maxRadius {
        objects := si.getObjectsInRadiusUnlocked(center, radius)  // Call unlocked version
        if len(objects) >= k {
            si.sortByDistance(objects, center)
            if len(objects) > k {
                return objects[:k]
            }
            return objects
        }
        radius *= 2
    }

    var allObjects []GameObject
    si.queryNode(si.root, si.bounds, &allObjects)
    si.sortByDistance(allObjects, center)
    if len(allObjects) > k {
        return allObjects[:k]
    }
    return allObjects
}
```

**Step 2: Add Test Coverage**
```go
// pkg/game/spatial_index_test.go
func TestSpatialIndex_GetNearestObjects_NoDeadlock(t *testing.T) {
    si := NewSpatialIndex(Rectangle{0, 0, 100, 100}, 10)
    
    // Insert 20 objects in grid pattern
    for x := 10; x <= 90; x += 10 {
        for y := 10; y <= 90; y += 10 {
            obj := &TestGameObject{
                id: fmt.Sprintf("obj_%d_%d", x, y),
                pos: Position{X: x, Y: y},
            }
            si.Insert(obj)
        }
    }
    
    // Test with timeout to detect deadlock
    done := make(chan bool)
    go func() {
        nearest := si.GetNearestObjects(Position{X: 50, Y: 50}, 5)
        assert.Len(t, nearest, 5, "Should return exactly 5 nearest objects")
        done <- true
    }()
    
    select {
    case <-done:
        // Success - no deadlock
    case <-time.After(2 * time.Second):
        t.Fatal("GetNearestObjects() deadlocked - timeout after 2 seconds")
    }
}
```

**Step 3: Verify Fix**
```bash
go test ./pkg/game -run TestSpatialIndex_GetNearestObjects -v
go test -race ./pkg/game -run TestSpatialIndex  # Check for race conditions
```

**Validation:**
- Test passes without timeout → deadlock fixed
- Race detector shows no warnings → thread-safe
- Combat system can find nearest targets → gameplay functional

**Estimated Effort:** 2-3 hours (refactoring + testing + validation)

---

## Gap 3: Effect System Compilation Errors

**Stated Goal:** "Comprehensive Effect System" with damage-over-time, healing, status effects, stacking, immunity

**Current State:**
- Effect manager implemented in `pkg/game/effectmanager.go` (392 lines)
- Effect behaviors in `pkg/game/effectbehavior.go` (502 lines)
- **CRITICAL BUG #1:** Line 368 calls `ToDamageEffect()` which doesn't exist (correct name: `AsDamageEffect()`)
- **CRITICAL BUG #2:** Line 382-385 uses `min()` builtin which may not be available in Go 1.23 toolchain
- Result: Code doesn't compile, effect ticking system broken

**Impact:**
- **Build Blocking:** `go build ./...` fails with "undefined: ToDamageEffect"
- **Combat System:** Damage-over-time effects (bleeding, poison, burning) cannot tick
- **Healing:** Healing-over-time effects cannot apply
- **Game Balance:** Effect system non-functional, combat balance broken

**Closing the Gap:**

**Fix 1: Correct Function Name (Line 368)**
```go
// pkg/game/effectbehavior.go:365-371
// Change:
func (em *EffectManager) processEffectTick(effect *Effect) {
    if damageEffect, ok := ToDamageEffect(effect); ok {  // ❌ WRONG
        em.processDamageEffect(damageEffect, time.Now())
        return
    }
    // ...
}

// To:
func (em *EffectManager) processEffectTick(effect *Effect) {
    if damageEffect, ok := AsDamageEffect(effect); ok {  // ✅ CORRECT
        em.processDamageEffect(damageEffect, time.Now())
        return
    }
    // ...
}
```

**Fix 2: Replace min() Builtin (Lines 382-385)**
```go
// Change:
em.currentStats.Health = min(
    em.currentStats.Health+healing,
    em.currentStats.MaxHealth,
)

// To (explicit conditional):
healedHealth := em.currentStats.Health + healing
if healedHealth > em.currentStats.MaxHealth {
    em.currentStats.Health = em.currentStats.MaxHealth
} else {
    em.currentStats.Health = healedHealth
}

// OR use existing utility (pkg/game/utils.go:105):
em.currentStats.Health = minFloat(
    em.currentStats.Health+healing,
    em.currentStats.MaxHealth,
)
```

**Validation:**
```bash
# Verify compilation
go build ./...  # Should succeed without errors

# Run effect system tests
go test ./pkg/game -run TestEffectManager -v
go test ./pkg/game -run TestDamageEffect -v

# Verify race-free
go test -race ./pkg/game -run TestEffect
```

**Regression Test:**
```go
// pkg/game/effectbehavior_test.go
func TestEffectManager_ProcessDamageEffect_Compilation(t *testing.T) {
    em := NewEffectManager()
    em.currentStats = &CharacterStats{Health: 100, MaxHealth: 100}
    
    // Create burning effect (damage over time)
    burning := CreateBurningEffect(5.0, 10*time.Second)
    em.AddEffect(burning.GetEffect())
    
    // Tick should process without panic
    em.ProcessEffects(time.Now())
    
    // Health should decrease from DoT
    assert.Less(t, em.currentStats.Health, 100.0, "Damage effect should reduce health")
}

func TestEffectManager_HealingEffect_MinClamping(t *testing.T) {
    em := NewEffectManager()
    em.currentStats = &CharacterStats{Health: 80, MaxHealth: 100}
    
    // Create healing effect that would overheal
    healing := &Effect{
        Type: EffectHealOverTime,
        Magnitude: 30.0,
        Stacks: 1,
        NextTick: time.Now(),
        TickInterval: time.Second,
    }
    em.AddEffect(healing)
    
    em.ProcessEffects(time.Now().Add(time.Second))
    
    // Health should clamp to MaxHealth
    assert.Equal(t, 100.0, em.currentStats.Health, "Healing should not exceed MaxHealth")
}
```

**Estimated Effort:** 30 minutes (2 simple fixes + test verification)

---

## Gap 4: Spell Content Insufficient for Full Gameplay

**Stated Goal:** "Spell System" with spell casting, 9 spell schools, levels 0-9, magical gameplay

**Current State:**
- Spell files exist: `data/spells/cantrips.yaml`, `level1.yaml` through `level9.yaml` (11 files total)
- Spell manager implemented: `pkg/game/spell_manager.go` (195 lines)
- **Gap:** Each level file contains only 3-5 spells (verified with `wc -l`: 839-2653 bytes per file)
- D&D Basic/OSR reference suggests 50-90 spells minimum for full progression
- Only basic damage types covered: fire, cold, lightning, acid
- Missing advanced spell effects: summoning, polymorph, illusions, teleportation, enchantments, divination

**Impact:**
- **Class Viability:** Mage and Cleric classes have limited spell selection, reducing replayability
- **Tactical Depth:** Combat lacks spell variety for strategic choices
- **Character Progression:** Leveling up doesn't provide meaningful spell unlocks
- **Magical Identity:** Each spell school (9 schools) has 1-2 spells max, no school specialization
- **Content Depth:** Advertised "spell system" feels incomplete compared to tabletop RPG inspiration

**Closing the Gap:**

**Phase 1: Expand Spell Data Files** (YAML content, no code changes)

Target: 8-12 spells per level × 9 levels = 72-108 spells

**Cantrips (Level 0)** - Expand to 12 spells
```yaml
# data/spells/cantrips.yaml
spells:
  # Existing: Fire Bolt, Ray of Frost, Shocking Grasp
  # Add:
  - spell_id: "acid_splash"
    name: "Acid Splash"
    spell_level: 0
    spell_school: 6  # Conjuration
    damage_type: "acid"
    base_damage: 4
    range: 60
    description: "Hurl a bubble of acid"
    
  - spell_id: "dancing_lights"
    name: "Dancing Lights"
    spell_level: 0
    spell_school: 5  # Evocation
    damage_type: ""
    range: 120
    duration: 60  # 1 minute
    description: "Create up to 4 lights that move"
    
  - spell_id: "mage_hand"
    name: "Mage Hand"
    spell_level: 0
    spell_school: 6  # Conjuration
    range: 30
    duration: 60
    description: "Spectral hand manipulates objects"
    
  # ... continue to 12 total
```

**Level 1** - Expand to 10 spells
```yaml
# data/spells/level1.yaml
spells:
  # Existing: Magic Missile, Cure Wounds, Shield
  # Add:
  - spell_id: "burning_hands"
    name: "Burning Hands"
    spell_level: 1
    spell_school: 5  # Evocation
    damage_type: "fire"
    base_damage: 15
    range: 15  # 15-foot cone
    area_of_effect: "cone"
    description: "Flames shoot from fingertips"
    
  - spell_id: "detect_magic"
    name: "Detect Magic"
    spell_level: 1
    spell_school: 2  # Divination
    range: 30
    duration: 600  # 10 minutes
    description: "Sense magical auras within range"
    
  - spell_id: "sleep"
    name: "Sleep"
    spell_level: 1
    spell_school: 4  # Enchantment
    range: 90
    area_of_effect: "20-foot radius"
    description: "Creatures fall unconscious"
    effect_type: "control"  # New effect type
    
  # ... continue to 10 total
```

**Levels 2-9** - Follow same pattern
- Level 2: Mirror Image, Scorching Ray, Invisibility, Knock, Darkness, Hold Person
- Level 3: Fireball, Lightning Bolt, Dispel Magic, Haste, Slow, Fly
- Level 4: Wall of Fire, Ice Storm, Polymorph, Dimension Door, Banishment
- Level 5: Cone of Cold, Cloudkill, Teleportation Circle, Wall of Stone
- Level 6: Chain Lightning, Disintegrate, True Seeing, Globe of Invulnerability
- Level 7: Delayed Blast Fireball, Teleport, Etherealness, Plane Shift
- Level 8: Sunburst, Earthquake, Maze, Mind Blank
- Level 9: Meteor Swarm, Wish, Time Stop, Gate

**Phase 2: Implement Advanced Spell Effects** (Code: ~200-300 lines)

Extend `pkg/game/spell_effects.go` with new effect types:

```go
// pkg/game/spell_effects.go

// EffectType constants (add to existing)
const (
    // ... existing: EffectDamageOverTime, EffectHealOverTime, etc.
    EffectSummon       EffectType = 50  // Summon creature
    EffectPolymorph    EffectType = 51  // Transform creature
    EffectIllusion     EffectType = 52  // Create illusion
    EffectTeleport     EffectType = 53  // Teleport position
    EffectCharm        EffectType = 54  // Enchantment/charm
    EffectSleep        EffectType = 55  // Put to sleep
    EffectFear         EffectType = 56  // Cause fear
    EffectInvisibility EffectType = 57  // Turn invisible
    EffectDetection    EffectType = 58  // Detect magic/enemies
)

// SummonEffect creates creatures under caster control
type SummonEffect struct {
    Effect      *Effect
    CreatureID  string
    HP          float64
    Duration    time.Duration
}

// PolymorphEffect transforms target's form
type PolymorphEffect struct {
    Effect        *Effect
    TargetForm    string  // "sheep", "rat", etc.
    HPMultiplier  float64
    Duration      time.Duration
}

// TeleportEffect moves character instantly
type TeleportEffect struct {
    Effect      *Effect
    TargetPos   Position
    Range       int
}

// Implement handlers in spell_manager.go
func (sm *SpellManager) ApplySummonSpell(spell *Spell, caster *Character, targetPos Position) error {
    // Create NPC at target position
    // Set to friendly to caster's faction
    // Add to world state with duration timer
}

func (sm *SpellManager) ApplyPolymorphSpell(spell *Spell, target *Character) error {
    // Save original stats
    // Apply form transformation
    // Set duration timer for revert
}

func (sm *SpellManager) ApplyTeleportSpell(spell *Spell, caster *Character, targetPos Position) error {
    // Validate target position (line of sight, not occupied)
    // Move character instantly
    // Trigger movement event
}
```

**Phase 3: Spell School Balance** (~100 lines)

Ensure each of 9 spell schools has 8-12 spells:
1. **Abjuration** (protection): Shield, Counterspell, Dispel Magic, Antimagic Field
2. **Divination** (knowledge): Detect Magic, Identify, Scrying, True Seeing
3. **Enchantment** (mind): Charm Person, Sleep, Hold Person, Dominate
4. **Evocation** (energy): Magic Missile, Fireball, Lightning Bolt, Cone of Cold
5. **Illusion** (deception): Disguise Self, Mirror Image, Invisibility, Phantasmal Killer
6. **Conjuration** (summoning): Mage Hand, Summon Monster, Teleport, Gate
7. **Necromancy** (death): Chill Touch, Animate Dead, Blight, Finger of Death
8. **Transmutation** (change): Enlarge/Reduce, Haste, Polymorph, Time Stop
9. **Universal** (all schools): Wish, Limited Wish, Read Magic

**Validation:**
```bash
# Count spells per file
wc -l data/spells/*.yaml  # Should show 2000-4000 bytes per file (10+ spells)

# Count total spells
grep "spell_id:" data/spells/*.yaml | wc -l  # Should output 72-108

# Verify spell loading
go test ./pkg/game -run TestSpellManager_LoadSpells -v

# Test spell effects
go test ./pkg/game -run TestSpellEffects -v
```

**Content Sources:**
- D&D 5e SRD (Open Game License - free to reference)
- Pathfinder SRD (Open Game License)
- Original Gold Box game spell lists
- OSR retroclones (OSRIC, Labyrinth Lord)

**Estimated Effort:**
- Phase 1 (YAML content): 8-12 hours (research + writing 72-108 spell entries)
- Phase 2 (spell effects code): 6-8 hours (implementation + testing)
- Phase 3 (school balance): 2-3 hours (distribution check + missing spells)
- **Total:** 16-23 hours

---

## Gap 5: No Visual/GUI Content Editors

**Stated Goal:** "World editor tools" and "Content creation utilities" in roadmap

**Current State:**
- **CLI tools exist:**
  - `cmd/map-editor/main.go` (545 lines) - ASCII tile map editor with templates
  - `cmd/quest-builder/main.go` (439 lines) - Interactive quest creation prompts
  - `cmd/content-creator/main.go` (511 lines) - Spell/item YAML generator
- **No GUI frameworks found:** No Fyne, Gio, web-based React/Vue editors
- **Documentation:** CLI tools documented in `docs/CONTENT_CREATION.md`
- **Barrier to entry:** Requires Go/YAML programming knowledge, not artist/designer-friendly

**Impact:**
- **Non-Programmer Exclusion:** Artists, game designers, writers cannot create content without coding
- **Workflow Inefficiency:** Manual YAML editing error-prone, no visual preview
- **Quality Assurance:** No live preview of maps, quests, items before testing in game
- **Onboarding Friction:** New contributors face steep learning curve
- **Competitive Gap:** Modern game engines (Unity, Unreal, Godot) have visual editors

**Closing the Gap:**

**Option A: Web-Based Visual Editor** (Recommended - leverages existing JSON-RPC API)

**Architecture:**
- Frontend: React + TypeScript + Canvas API
- Backend: Existing `/rpc` JSON-RPC endpoint
- Real-time preview: WebSocket connection for live testing

**Features:**
1. **Map Editor UI**
   - Drag-and-drop tile palette (floor, wall, water, door, etc.)
   - Grid-based canvas with zoom/pan
   - Multi-layer support (terrain, objects, NPCs)
   - Template library (dungeon, outdoor, cave, town)
   - Export to JSON format compatible with game engine

2. **Quest Builder UI**
   - Node-based graph editor for quest flow
   - Objective types: fetch, kill, escort, explore, puzzle
   - Reward calculator (gold, XP, items)
   - Condition builder (quest prerequisites, reputation requirements)
   - YAML export with validation

3. **Spell/Item Creator UI**
   - Form-based editor with dropdowns (spell schools, damage types)
   - Preview panel showing spell/item stats
   - Template gallery (damage, healing, buff, debuff, utility)
   - Validation against game schema
   - YAML output

**Implementation Plan:**
```bash
# Project structure
web/editor/
├── package.json
├── src/
│   ├── components/
│   │   ├── MapEditor.tsx
│   │   ├── QuestBuilder.tsx
│   │   ├── ContentCreator.tsx
│   │   └── TilePalette.tsx
│   ├── api/
│   │   └── rpcClient.ts  # JSON-RPC wrapper
│   ├── types/
│   │   └── gameTypes.ts  # TypeScript definitions
│   └── App.tsx
└── public/
    └── index.html
```

**JSON-RPC Integration:**
```typescript
// web/editor/src/api/rpcClient.ts
class RPCClient {
    async call(method: string, params: object): Promise<any> {
        const response = await fetch('/rpc', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({
                jsonrpc: '2.0',
                method,
                params,
                id: Date.now()
            })
        });
        return response.json();
    }
    
    // Editor-specific methods
    async saveMap(mapData: MapData): Promise<void> {
        await this.call('saveMap', {map_data: mapData});
    }
    
    async validateQuest(questData: QuestData): Promise<ValidationResult> {
        return await this.call('validateContent', {content: questData});
    }
}
```

**New RPC Endpoints Required:**
```go
// pkg/server/editor_handlers.go
func (s *RPCServer) handleSaveMap(params map[string]interface{}) (interface{}, error) {
    // Parse map data from params
    // Validate against schema
    // Write to data/maps/{map_id}.json
    // Return success/error
}

func (s *RPCServer) handleLoadMap(params map[string]interface{}) (interface{}, error) {
    // Read map file
    // Return map data + metadata
}

func (s *RPCServer) handleValidateContent(params map[string]interface{}) (interface{}, error) {
    // Use existing pkg/validation/ framework
    // Return validation errors + warnings
}
```

**Estimated Effort:** 80-120 hours (full web-based editor suite)

---

**Option B: Desktop GUI with Fyne** (Native Go toolkit)

```go
// cmd/gui-editor/main.go
package main

import (
    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
)

func main() {
    myApp := app.New()
    myWindow := myApp.NewWindow("GoldBox Content Editor")
    
    // Map editor tab
    mapCanvas := canvas.NewRaster(renderMapGrid)
    
    // Quest builder tab
    questForm := widget.NewForm(
        widget.NewFormItem("Quest Name", widget.NewEntry()),
        widget.NewFormItem("Type", widget.NewSelect(
            []string{"fetch", "kill", "escort", "explore", "puzzle"},
            func(value string) {})),
    )
    
    tabs := container.NewAppTabs(
        container.NewTabItem("Maps", mapCanvas),
        container.NewTabItem("Quests", questForm),
        container.NewTabItem("Items", itemEditor),
    )
    
    myWindow.SetContent(tabs)
    myWindow.ShowAndRun()
}
```

**Estimated Effort:** 60-80 hours (Fyne desktop editor)

---

**Option C: Document CLI-Only Status** (Minimal effort)

Update documentation to clarify current limitations:

**README.md Changes:**
```markdown
## 🛠️ Content Creation Tools

The GoldBox RPG Engine includes **CLI-based content creation tools** for maps, quests, and items:

### Available Tools
- **Map Editor** (`cmd/map-editor`) - ASCII tile-based map creator with templates
- **Quest Builder** (`cmd/quest-builder`) - Interactive quest creation with prompts
- **Content Creator** (`cmd/content-creator`) - Spell and item YAML generator

### Usage Examples
```bash
# Create a dungeon map
./map-editor -w 30 -h 20 -t dungeon -o maps/dungeon1.json

# Build a quest interactively
./quest-builder -i -o quests/main_quest.yaml

# Generate a spell
./content-creator -c spell -t damage -o spells/custom_fireball.yaml
```

### Limitations
⚠️ **Current tools are CLI-based and require command-line familiarity.** Visual/GUI editors are planned for future releases (see roadmap).

For detailed usage, see [CONTENT_CREATION.md](docs/CONTENT_CREATION.md).
```

**ROADMAP.md Changes:**
```markdown
- ⚠️ **Content creation utilities** (CLI tools complete, GUI editors planned)
- ⚠️ **World editor tools** (ASCII map editor available, visual editor planned)
```

**Estimated Effort:** 1 hour (documentation updates)

---

## Gap 6: Guild and Faction System Incomplete

**Stated Goal:** "Guild and faction systems" in roadmap

**Current State:**
- Faction generation exists: `pkg/pcg/faction.go` (213 lines)
- Reputation system implemented: `pkg/pcg/reputation.go` (227 lines) with dynamic effects and decay
- **Explicit TODO:** Line 31 in `faction.go` marks "territory control" as unimplemented
- **Missing:**
  - Guild membership mechanics (join, leave, rank progression)
  - Faction territory control (zones owned by factions)
  - Guild quests (faction-specific objectives)
  - Faction wars (inter-faction conflict)
  - Player-created guilds
  - Inter-faction diplomacy (allied/neutral/hostile relationships)

**Impact:**
- **Social Gameplay:** No multiplayer guild system for cooperation
- **End-Game Content:** High-level players lack faction progression goals
- **World Dynamics:** Factions exist but don't interact or control territory
- **Reputation System:** Player-to-faction reputation works, but no faction-to-faction relationships
- **Strategic Depth:** No territorial conquest or faction alliance mechanics

**Closing the Gap:**

**Phase 1: Guild Data Structures** (~200 lines)
```go
// pkg/game/guild.go
package game

import (
    "sync"
    "time"
)

type Guild struct {
    mu sync.RWMutex
    
    ID          string
    Name        string
    FactionID   string  // Associated faction (optional)
    LeaderID    string  // Character ID of guild leader
    
    Members     map[string]*GuildMember  // Character ID -> Member
    Ranks       []GuildRank
    Treasury    int  // Guild gold
    
    CreatedAt   time.Time
    Level       int  // Guild level (unlocks perks)
    Reputation  int  // Guild-wide reputation
    
    Territory   []string  // Zone IDs controlled by guild
}

type GuildMember struct {
    CharacterID  string
    RankID       int
    JoinedAt     time.Time
    Contribution int  // Activity points
}

type GuildRank struct {
    ID          int
    Name        string
    Permissions GuildPermissions
}

type GuildPermissions struct {
    CanInvite      bool
    CanKick        bool
    CanPromote     bool
    CanAccessBank  bool
    CanDeclareWar  bool
}

// Guild operations
func (g *Guild) AddMember(characterID string) error {
    g.mu.Lock()
    defer g.mu.Unlock()
    
    if len(g.Members) >= g.getMaxMembers() {
        return ErrGuildFull
    }
    
    g.Members[characterID] = &GuildMember{
        CharacterID: characterID,
        RankID:      0,  // Lowest rank
        JoinedAt:    time.Now(),
    }
    return nil
}

func (g *Guild) PromoteMember(characterID string, newRank int) error {
    g.mu.Lock()
    defer g.mu.Unlock()
    
    member, exists := g.Members[characterID]
    if !exists {
        return ErrNotGuildMember
    }
    
    if newRank >= len(g.Ranks) {
        return ErrInvalidRank
    }
    
    member.RankID = newRank
    return nil
}

func (g *Guild) getMaxMembers() int {
    return 50 + (g.Level * 10)  // Base 50, +10 per level
}
```

**Phase 2: Faction Territory Control** (~300 lines)
```go
// pkg/game/territory.go
package game

type Territory struct {
    ZoneID      string
    OwnerGuildID string
    OwnerFactionID string
    
    DefensePoints int  // Strength of control
    Resources     map[string]int  // Resources generated
    
    ContestedBy   map[string]int  // Guild ID -> contest points
    ControlledSince time.Time
}

// World method for territory management
func (w *World) ClaimTerritory(zoneID string, guildID string) error {
    w.mu.Lock()
    defer w.mu.Unlock()
    
    territory, exists := w.Territories[zoneID]
    if !exists {
        return ErrZoneNotFound
    }
    
    // Check if zone is already owned
    if territory.OwnerGuildID != "" && territory.OwnerGuildID != guildID {
        // Must contest ownership first
        return ErrTerritoryContested
    }
    
    territory.OwnerGuildID = guildID
    territory.ControlledSince = time.Now()
    territory.DefensePoints = 100  // Base defense
    
    return nil
}

func (w *World) ContestTerritory(zoneID string, attackerGuildID string, contestPoints int) error {
    territory, exists := w.Territories[zoneID]
    if !exists {
        return ErrZoneNotFound
    }
    
    if territory.ContestedBy == nil {
        territory.ContestedBy = make(map[string]int)
    }
    
    territory.ContestedBy[attackerGuildID] += contestPoints
    
    // If contest points exceed defense, transfer ownership
    if territory.ContestedBy[attackerGuildID] >= territory.DefensePoints {
        territory.OwnerGuildID = attackerGuildID
        territory.ContestedBy = make(map[string]int)
        territory.ControlledSince = time.Now()
        
        // Emit territory captured event
        w.EmitEvent(GameEvent{
            Type: EventTerritoryCaptured,
            SourceID: attackerGuildID,
            Data: map[string]interface{}{
                "zone_id": zoneID,
            },
        })
    }
    
    return nil
}
```

**Phase 3: Inter-Faction Diplomacy** (~150 lines)
```go
// pkg/pcg/faction_diplomacy.go
package pcg

type FactionRelationship struct {
    Faction1ID  string
    Faction2ID  string
    
    Status      RelationshipStatus  // Allied, Neutral, Hostile, War
    Opinion     int  // -100 to +100
    
    TradeAgreement  bool
    DefensePact     bool
    AtWar           bool
}

type RelationshipStatus int
const (
    RelationshipAllied RelationshipStatus = iota
    RelationshipFriendly
    RelationshipNeutral
    RelationshipUnfriendly
    RelationshipHostile
    RelationshipWar
)

func (fg *FactionGenerator) InitializeDiplomacy(factions []*Faction) map[string]map[string]*FactionRelationship {
    relationships := make(map[string]map[string]*FactionRelationship)
    
    for i, f1 := range factions {
        relationships[f1.ID] = make(map[string]*FactionRelationship)
        
        for j, f2 := range factions {
            if i == j {
                continue  // Skip self-relationship
            }
            
            // Determine initial relationship based on faction types
            opinion := fg.calculateInitialOpinion(f1, f2)
            status := fg.opinionToStatus(opinion)
            
            relationships[f1.ID][f2.ID] = &FactionRelationship{
                Faction1ID: f1.ID,
                Faction2ID: f2.ID,
                Status:     status,
                Opinion:    opinion,
            }
        }
    }
    
    return relationships
}

func (fg *FactionGenerator) calculateInitialOpinion(f1, f2 *Faction) int {
    opinion := 0
    
    // Law vs Chaos alignment
    if abs(f1.Alignment.Law - f2.Alignment.Law) < 30 {
        opinion += 20  // Similar legal views
    } else {
        opinion -= 30  // Opposing legal views
    }
    
    // Good vs Evil alignment
    if abs(f1.Alignment.Good - f2.Alignment.Good) < 30 {
        opinion += 20
    } else {
        opinion -= 30
    }
    
    // Historical grudges (random chance)
    if fg.rng.Float64() < 0.1 {
        opinion -= 50  // 10% chance of historical conflict
    }
    
    return clamp(opinion, -100, 100)
}

func (fg *FactionGenerator) opinionToStatus(opinion int) RelationshipStatus {
    switch {
    case opinion >= 75:
        return RelationshipAllied
    case opinion >= 25:
        return RelationshipFriendly
    case opinion >= -25:
        return RelationshipNeutral
    case opinion >= -75:
        return RelationshipUnfriendly
    default:
        return RelationshipHostile
    }
}
```

**Phase 4: Guild Quests** (~200 lines)
```go
// pkg/pcg/quests/guild_quests.go
package quests

type GuildQuestGenerator struct {
    questGen *QuestGenerator
}

func (gqg *GuildQuestGenerator) GenerateGuildQuest(guild *Guild, faction *Faction) (*Quest, error) {
    // Generate quest types specific to guilds
    questTypes := []string{
        "guild_territory_defense",
        "guild_resource_gathering",
        "guild_rival_sabotage",
        "guild_alliance_mission",
        "guild_rank_advancement",
    }
    
    questType := questTypes[gqg.questGen.rng.Intn(len(questTypes))]
    
    switch questType {
    case "guild_territory_defense":
        return gqg.generateTerritoryDefense(guild)
    case "guild_resource_gathering":
        return gqg.generateResourceGathering(guild)
    case "guild_rival_sabotage":
        return gqg.generateRivalSabotage(guild)
    default:
        return gqg.questGen.Generate(questType)
    }
}

func (gqg *GuildQuestGenerator) generateTerritoryDefense(guild *Guild) (*Quest, error) {
    // Create quest to defend guild territory from NPC attackers
    return &Quest{
        ID:          generateID(),
        Title:       fmt.Sprintf("Defend %s Territory", guild.Name),
        Description: "Enemy forces are attacking our territory. Repel the invaders!",
        Type:        "defense",
        
        Objectives: []Objective{
            {
                Type:        "kill",
                Target:      "enemy_soldier",
                Count:       10,
                Description: "Defeat 10 enemy soldiers",
            },
            {
                Type:        "survive",
                Duration:    300,  // 5 minutes
                Description: "Hold the territory for 5 minutes",
            },
        },
        
        Rewards: Rewards{
            GuildReputation: 500,
            GuildGold:       1000,
            TerritoryDefense: 50,  // Increase territory defense points
        },
        
        TimeLimit: 3600,  // 1 hour
    }, nil
}
```

**Validation:**
```bash
# Test guild operations
go test ./pkg/game -run TestGuild -v

# Test territory control
go test ./pkg/game -run TestTerritory -v

# Test faction diplomacy
go test ./pkg/pcg -run TestFactionDiplomacy -v

# E2E guild scenario
go test ./test/e2e -run TestGuildTerritory -v
```

**Estimated Effort:**
- Phase 1 (guild structures): 8-10 hours
- Phase 2 (territory control): 12-15 hours
- Phase 3 (diplomacy): 6-8 hours
- Phase 4 (guild quests): 8-10 hours
- Testing + integration: 6-8 hours
- **Total:** 40-51 hours

---

## Summary Table

| Gap # | Category | Severity | Effort | Impact |
|-------|----------|----------|--------|--------|
| 1 | Asset Pipeline | MEDIUM | 6-8h (Option A) / 30min (Option C) | User-facing polish |
| 2 | Spatial Deadlock | CRITICAL | 2-3h | Game-breaking bug |
| 3 | Effect Compilation | CRITICAL | 30min | Build-blocking bug |
| 4 | Spell Content | MEDIUM | 16-23h | Gameplay depth |
| 5 | GUI Editors | LOW | 80-120h (web) / 1h (docs) | Workflow efficiency |
| 6 | Guild/Faction | MEDIUM | 40-51h | Social/end-game |

**Recommended Priority:**
1. **Gap 3** (30 min) - Fix compilation errors immediately
2. **Gap 2** (2-3h) - Fix deadlock bug (game-breaking)
3. **Gap 1** (30 min) - Document asset status (Option C)
4. **Gap 4** (16-23h) - Expand spell content for gameplay depth
5. **Gap 6** (40-51h) - Complete guild/faction systems
6. **Gap 5** (80-120h) - Build visual editors (long-term)

---

**Generated:** 2026-03-12T03:45:47Z  
**Based On:** README.md, ROADMAP.md, codebase analysis, go-stats-generator metrics
