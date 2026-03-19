# Roadmap

Generated: 2026-03-19

## Gold Box Reference Standard

A "Gold Box faithful" implementation for this codebase means:

**Fixed, non-overlapping panel layout** with bold bright borders separating the viewport (first-person corridors during exploration, bird's-eye tactical grid during combat), the vertically-stacked party roster on the right, the scrolling message log at the bottom receiving ALL game feedback as text, and a context-sensitive command menu with highlighted-letter keyboard navigation.

**EGA-inspired 16-color palette** sensibility: deep blues (#2E5090), magentas/purples (#7A4ABF), vivid golds (#BFA54A), dungeon grays (#5A5A5A, #8B8B8B), and medieval reds (#8B2E2E). Dark near-black backgrounds with high-contrast foreground sprites and text. No gradients, no anti-aliasing — chunky, flat-colored, clearly readable pixel art.

**All numbers explicit**: HP, AC, damage, XP, initiative order, action point costs — the player always sees exact state. Combat narration flows to the message log in prose form: `"Sable attacks Orc — HIT for 7 damage"`, `"Goblin casts Sleep — FAILED"`.

The current UI uses Gold Box–inspired panel divisions and the correct palette constants (`types_ui.go`), but lacks first-person exploration rendering, movement/attack range highlights on the combat grid, damage flash animations, morale indicators for NPCs, and full message log integration for all feedback.

---

## Improvement Items

Items are grouped by theme and ordered within each group by priority (highest impact first). Each item is specified in enough detail for autonomous implementation.

---

### Group: Exploration Screen

#### 1. First-Person Dungeon Viewport Rendering

**Priority:** High  
**Complexity:** Large  
**Depends on:** None

**Current State**  
`pkg/wasmui/exploration.go:drawViewport()` renders a bird's-eye grid of floor tiles with a "P" marker for the player at center. The function draws tiles using `DrawSpriteWithFallback()` for `floor_stone` terrain, then overlays grid lines.

**Gap**  
Gold Box exploration uses a first-person, step-and-turn perspective with wireframe or block-filled corridor walls. The current top-down grid view violates this core visual convention. Doors are not rendered as distinct tiles, and there's no sense of depth or corridor perspective.

**Implementation Specification**  
- Files to modify: `pkg/wasmui/exploration.go`
- Create new function `drawFirstPersonView(screen *ebiten.Image, playerX, playerY, facing int)` replacing current grid rendering in `drawViewport()`
- Implement raycasting or pre-rendered depth slices (3 depth levels: near, mid, far)
- Draw walls as filled rectangles using palette colors (`ColorPanelBorder` for far, `ColorStatValue` for near)
- Render doors as distinct rectangles with different color (`ColorGold` door frame)
- Use `pkg/game/map.go` data via RPC `getMapView` to determine wall/door positions relative to player facing
- Add `facing` field to `PlayerState` in `types_game.go`
- Add turn-left (Q) and turn-right (E) keybindings in `handleExplorationMovement()`
- Each step is a discrete redraw — no smooth scrolling
- Sprite references: Use `terrain/dungeon/tile_wall_stone.png`, `terrain/dungeon/tile_door_wood_closed.png` from ASSET_ANALYSIS.md
- Risk: Requires map geometry data from server; may need new RPC method `getVisibleWalls`

**Success Criteria**  
- [x] Viewport shows first-person corridor view with visible walls at 3 depth levels
- [x] Doors render as distinct tiles with different color/texture
- [x] Q/E keys rotate player facing; movement is relative to facing direction
- [x] Each movement step produces instant redraw (no animation)

---

#### 2. Encounter Text Overlay System

**Priority:** Medium  
**Complexity:** Medium  
**Depends on:** None

**Current State**  
`pkg/wasmui/exploration.go:drawCombatLog()` shows scrolling messages. Encounters and dialogue are not displayed as Gold Box–style text overlays or panels in the viewport area.

**Gap**  
Gold Box displays encounters, examinations, and dialogue as text overlays or dedicated panels, sometimes with static NPC portraits. The current implementation only uses the log panel, not viewport-area overlays.

**Implementation Specification**  
- Files to modify: `pkg/wasmui/exploration.go`, `pkg/wasmui/overlays.go`
- Add `EncounterOverlay` struct with fields: `visible bool`, `title string`, `text string`, `portraitPath string`, `choices []string`, `selectedChoice int`
- Create `drawEncounterOverlay(screen *ebiten.Image)` that renders a centered panel (400x200px) over the viewport
- Panel uses `ColorPanelBG` background with `ColorPanelBorderHi` double-border
- Title in `ColorGold`, body text in `ColorStatValue`
- Portrait (if any) rendered left side using `DrawAdventureSpriteWithFallback()`
- Handle Enter to dismiss, Up/Down for choice selection
- Wire to RPC events for `encounter_start`, `npc_dialogue` events
- Sprite references: Adventure NPC portraits via `AdventureNPCPath(adventureSlug, npcID)`

**Success Criteria**  
- [x] Encounters display as centered overlay panel with title and description
- [x] NPC portraits render when available
- [x] Multiple-choice dialogues navigable with arrow keys
- [x] Enter dismisses overlay or selects choice

---

#### 3. Minimap Fog of War

**Priority:** Low  
**Complexity:** Small  
**Depends on:** None

**Current State**  
`pkg/wasmui/exploration.go:drawMinimap()` renders a 100x80px black rectangle with the player as a green dot at center. Comment references "unexplored = black" but no actual fog-of-war tracking exists.

**Gap**  
Gold Box games reveal map areas as explored, keeping unexplored areas black. The current minimap shows nothing beyond the player position — no explored corridors or rooms.

**Implementation Specification**  
- Files to modify: `pkg/wasmui/exploration.go`, `pkg/wasmui/types_game.go`
- Add `ExploredTiles map[string]bool` to `Game` struct (key: "x,y,level")
- On each successful move, mark current tile as explored
- In `drawMinimap()`, iterate explored tiles and draw small dots/pixels for floors (gray) and walls (darker gray)
- Scale minimap to show ~10 tile radius around player
- Player dot remains bright green at center
- Constraints: Keep minimap 100x80px; use 2x2 pixel blocks per tile

**Success Criteria**  
- [x] Minimap reveals tiles as player moves through them
- [x] Unexplored areas remain black
- [x] Explored walls and floors have distinct colors
- [x] Player position clearly visible at center

---

### Group: Combat Screen

#### 4. Movement Range Highlighting

**Priority:** High  
**Complexity:** Medium  
**Depends on:** None

**Current State**  
`pkg/wasmui/combat_screen.go:drawCombatGrid()` renders floor tiles and entity positions. When move mode is active (`CombatActionMove`), the log says "Move mode - click tile or use movement keys" but no visual range indicator appears on the grid.

**Gap**  
Gold Box combat shows the movement range as visually distinct highlighted tiles. The current implementation provides no spatial feedback about how far the character can move with remaining action points.

**Implementation Specification**  
- Files to modify: `pkg/wasmui/combat_screen.go`
- Add function `getMovementRange(playerX, playerY, ap int) []Position` calculating reachable tiles based on AP
- In `drawCombatGrid()`, after drawing floor tiles and before drawing entities:
  - If `g.combatAction == CombatActionMove`, iterate movement range positions
  - Draw semi-transparent blue overlay (`color.RGBA{R: 74, G: 125, B: 191, A: 80}`) on reachable tiles
- Use Manhattan distance for simplicity: range = AP * 2 tiles (configurable)
- Query server for actual walkable tiles via `getObjectsInRange` or add `getMovementRange` RPC
- Constraints: Must not obscure entity sprites; use alpha blending

**Success Criteria**  
- [x] Entering move mode highlights all reachable tiles in distinct blue tint
- [x] Highlighted area respects AP remaining
- [x] Tiles occupied by walls/enemies are not highlighted
- [x] Overlay does not obscure entity sprites

---

#### 5. Attack Range Highlighting

**Priority:** High  
**Complexity:** Medium  
**Depends on:** 4

**Current State**  
Attack mode (`CombatActionAttack`) shows log message "Attack mode - select target" but no visual indication of which tiles are in weapon range.

**Gap**  
Gold Box combat shows attack range zones. Melee weapons highlight adjacent tiles; ranged weapons show a larger radius. The current implementation provides no visual attack range feedback.

**Implementation Specification**  
- Files to modify: `pkg/wasmui/combat_screen.go`, `pkg/wasmui/types_rpc.go`
- Query equipped weapon range from player state (add `WeaponRange int` to `PlayerState` or fetch via `getEquipment`)
- Add `getAttackRange(playerX, playerY, weaponRange int) []Position`
- In `drawCombatGrid()`:
  - If `g.combatAction == CombatActionAttack`, draw red-tinted overlay (`color.RGBA{R: 191, G: 74, B: 74, A: 80}`) on tiles in range
  - Highlight valid targets (enemies in range) with pulsing border or distinct color
- Melee range = 1 (adjacent 8 tiles), Ranged = weapon-specific (3-10 tiles)

**Success Criteria**  
- [x] Attack mode highlights tiles within weapon range in red tint
- [x] Valid enemy targets have distinct visual indicator
- [x] Range changes based on equipped weapon (melee vs ranged)

---

#### 6. Damage Flash Animation

**Priority:** High  
**Complexity:** Small  
**Depends on:** None

**Current State**  
`pkg/wasmui/combat_screen.go:executeAttack()` narrates hits/misses in the log but entity sprites show no visual feedback on hit.

**Gap**  
Gold Box combat shows damage flashes or brief visual effects on hit. The current implementation silently updates HP values without visual confirmation.

**Implementation Specification**  
- Files to modify: `pkg/wasmui/combat_screen.go`, `pkg/wasmui/types_game.go`
- Add `DamageFlash` struct: `{entityID string, startTime time.Time, duration time.Duration, color color.RGBA}`
- Add `damageFlashes []DamageFlash` to `Game` struct
- In `executeAttack()`, on successful hit add flash entry: `duration: 200ms, color: ColorEnemyName`
- In `drawCombatGrid()` entity rendering section:
  - Check if entity has active flash
  - If so, tint sprite/rect with flash color using `ColorScale`
  - Remove flash when expired
- Flash colors: red for damage, green for healing
- Constraints: Keep simple — single color tint, no complex animation

**Success Criteria**  
- [x] Entities flash red briefly when taking damage
- [x] Entities flash green briefly when healed
- [x] Flash duration is ~200ms, clearly visible but not disruptive
- [x] Multiple simultaneous flashes work correctly

---

#### 7. Spell Effect Overlay

**Priority:** Medium  
**Complexity:** Medium  
**Depends on:** 6

**Current State**  
`pkg/wasmui/overlays.go:castSelectedSpell()` narrates spell results in the log. No visual effect sprites appear on the combat grid.

**Gap**  
Gold Box combat shows spell effect overlays (fireball explosions, lightning bolts, etc.). The current implementation has effect sprites defined in ASSET_ANALYSIS.md but doesn't display them.

**Implementation Specification**  
- Files to modify: `pkg/wasmui/combat_screen.go`, `pkg/wasmui/types_game.go`, `pkg/wasmui/asset_loader.go`
- Add `SpellEffect` struct: `{spellID string, targetPos Position, startTime time.Time, frames int, currentFrame int}`
- Add `spellEffects []SpellEffect` to `Game` struct
- In `castSelectedSpell()`, on success add spell effect entry
- Add `drawSpellEffects(screen *ebiten.Image)` called from `drawCombatGrid()`
- Load effect sprites using `EffectSpritePath(spellID)` — path format: `effects/{type}/effect_{name}.png`
- Animate through frames (3-4 frames, 100ms each) then remove
- Map spell schools to effect types: Evocation→fire/lightning, Necromancy→dark, etc.
- Fallback: Draw colored expanding circle if sprite not loaded

**Success Criteria**  
- [x] Casting fireball shows flame effect sprite at target location
- [x] Effect animates through 3-4 frames over ~400ms
- [x] Different spell schools have distinct visual effects
- [x] Graceful fallback to colored overlay if sprite missing

---

#### 8. Active Character Tile Highlight

**Priority:** Medium  
**Complexity:** Small  
**Depends on:** None

**Current State**  
`pkg/wasmui/combat_screen.go:drawInitiativePanel()` marks the current turn with `"> "` prefix and highlight bar. The combat grid does not highlight the active character's tile.

**Gap**  
Gold Box combat highlights the active character's tile distinctly. The current grid treats all character tiles identically regardless of whose turn it is.

**Implementation Specification**  
- Files to modify: `pkg/wasmui/combat_screen.go`
- In `drawCombatGrid()`, after drawing floor tiles:
  - Identify current turn entity from `combat.CurrentTurn`
  - Draw pulsing gold border around that entity's tile
  - Use `ColorGoldHi` with alpha oscillating based on time (sin wave, 0.5-1.0 range)
- Apply to both player and enemy turns
- Border width: 2px, drawn before entity sprite

**Success Criteria**  
- [x] Current turn character has pulsing gold border on their tile
- [x] Border clearly visible for both player and enemy turns
- [x] Pulse rate ~1 Hz, subtle but noticeable

---

#### 9. NPC Morale Indicator

**Priority:** Medium  
**Complexity:** Medium  
**Depends on:** None

**Current State**  
`pkg/game/morale.go` implements a complete morale system with states (Steadfast, Shaken, Broken, Panicked) and modifiers. `pkg/wasmui/combat_screen.go` shows no morale information in the UI.

**Gap**  
The backend has a fully implemented morale system that affects NPC behavior (fleeing, defensive fighting), but players have no visibility into enemy morale state.

**Implementation Specification**  
- Files to modify: `pkg/wasmui/combat_screen.go`, `pkg/wasmui/types_game.go`, `pkg/wasmui/rpc_methods.go`
- Add `MoraleState string` to `InitiativeEntry` struct
- Add RPC method `getMoraleState(entityID)` or include morale in combat state response
- In `drawInitiativePanel()`, for non-player entries:
  - Show morale icon next to name based on state
  - Steadfast: no icon, Shaken: yellow "!", Broken: red "!!", Panicked: skull/flee icon
- In `drawCombatGrid()`:
  - Optionally show small morale icon above enemy sprites
- Use `ColorGold` for Steadfast, `ColorEffectControl` for Shaken, `ColorEnemyName` for Broken/Panicked

**Success Criteria**  
- [x] Initiative panel shows morale state icon for each NPC
- [x] Morale changes during combat are reflected in UI
- [x] Player can visually identify which enemies are close to fleeing
- [x] Four distinct states have four distinct visual indicators

---

#### 10. Cover/Flanking Visual Indicators

**Priority:** Low  
**Complexity:** Medium  
**Depends on:** 4, 5

**Current State**  
`pkg/game/combat_modifiers.go` implements cover (Half/ThreeQuarters/Full) and flanking calculations using line-of-sight and positioning. The combat UI shows no cover or flanking information.

**Gap**  
Players cannot see which tiles provide cover or when they have flanking bonus on an enemy. This tactical information is computed but hidden.

**Implementation Specification**  
- Files to modify: `pkg/wasmui/combat_screen.go`, `pkg/wasmui/rpc_methods.go`
- Add RPC method `getCombatModifiers(attackerPos, defenderPos)` returning cover type and flanking status
- When selecting attack target:
  - Query cover between player and target
  - Display cover icon on target tile (half-shield, full-shield)
  - If flanking applies, show "FLANK" text in gold
- In `drawCombatGrid()`:
  - Highlight tiles adjacent to obstacles with subtle cover indicator
- Cover icons use shield sprite or text: "1/2", "3/4", "FULL"

**Success Criteria**  
- [x] Attacking shows cover modifier on target tile
- [x] Flanking bonus displayed when applicable
- [x] Player can identify cover-providing terrain before moving

---

### Group: Character & Party Display

#### 11. Equipment Icons in Inventory

**Priority:** High  
**Complexity:** Small  
**Depends on:** None

**Current State**  
`pkg/wasmui/overlays.go:drawInventoryScreen()` displays inventory as text-only list: item name, type, equipped status. No icons are shown despite item sprites being defined in ASSET_ANALYSIS.md.

**Gap**  
Gold Box inventory shows item icons alongside text. The current text-only display lacks visual recognition and feels less polished.

**Implementation Specification**  
- Files to modify: `pkg/wasmui/overlays.go`
- In inventory list rendering loop:
  - Add 24x24 icon area before item name
  - Use `ItemIconPath(item.Type, item.Name)` to get sprite path
  - Call `DrawSpriteWithFallback()` for each item
  - Fallback color based on item type: weapons=gray, armor=blue, consumables=green
- Adjust list item height from 28 to 32 to accommodate icon
- Asset paths: `items/weapons/item_sword.png`, `items/armor/item_leather.png`, etc.

**Success Criteria**  
- [x] Each inventory item shows icon to left of name
- [x] Icons load asynchronously with colored rect fallback
- [x] Item types have distinct fallback colors
- [x] Layout remains readable and scannable

---

#### 12. Character Portrait in Panel

**Priority:** Medium  
**Complexity:** Small  
**Depends on:** None

**Current State**  
`pkg/wasmui/exploration.go:drawCharacterPanel()` shows text stats (name, class, HP, attributes). No character portrait is displayed despite portrait sprites existing in `web/static/assets/sprites/characters/portraits/`.

**Gap**  
Gold Box shows character portraits prominently. The current character panel is text-only without visual representation.

**Implementation Specification**  
- Files to modify: `pkg/wasmui/exploration.go`
- In `drawCharacterPanel()`, add portrait area (64x64) at top of panel below title
- Use `CharacterPortraitPath(player.Class, "human", "male")` — hardcode race/gender for now
- Call `DrawSpriteWithFallback()` with class-based fallback color
- Adjust text layout to start below portrait
- Portrait position: `panelX + 68, panelY + 30` (centered in 200px panel)
- Fallback colors per class: Fighter=red, Mage=blue, Cleric=white, Thief=gray, Ranger=green, Paladin=gold

**Success Criteria**  
- [x] Character portrait displays in panel
- [x] Portrait matches character class
- [x] Fallback color rect shows while loading or if missing
- [x] Text stats remain visible below portrait

---

#### 13. Action Point Cost Display

**Priority:** Medium  
**Complexity:** Small  
**Depends on:** None

**Current State**  
`pkg/wasmui/exploration.go:drawAPBar()` shows AP as filled/empty dots with count `(X/Y)`. Combat actions consume AP but cost is not displayed before selecting action.

**Gap**  
Players don't know AP cost of actions before selecting them. Gold Box shows explicit action costs.

**Implementation Specification**  
- Files to modify: `pkg/wasmui/combat_screen.go`
- In `drawCombatActionBar()`, add AP cost next to each button label:
  - Move: 1 AP, Attack: 1 AP, Cast: varies, UseItem: 1 AP
- Format: `[M] Move (1)` instead of `[M] Move`
- For Cast, show "(1-3)" indicating variable cost
- Dim buttons that require more AP than available
- Add `CanAfford(action CombatAction, currentAP int) bool` helper

**Success Criteria**  
- [x] Each combat action button shows AP cost in parentheses
- [x] Buttons for unaffordable actions are visually dimmed
- [x] Variable cost actions show range
- [x] Player can plan turns knowing exact costs

---

#### 14. Effect Immunity Display

**Priority:** Low  
**Complexity:** Small  
**Depends on:** None

**Current State**  
`pkg/game/effectimmunity.go` implements immunity types (None, Partial, Complete, Reflect). `pkg/wasmui/exploration.go:drawActiveEffects()` shows active effects but not immunities.

**Gap**  
Players cannot see their character's immunities in the UI. This information is computed backend-side but not exposed.

**Implementation Specification**  
- Files to modify: `pkg/wasmui/exploration.go`, `pkg/wasmui/types_game.go`, `pkg/wasmui/rpc_methods.go`
- Add `Immunities []string` to `PlayerState`
- Add RPC method or extend `getGameState` to include immunity list
- In `drawCharacterPanel()`, add "Immunities:" section below effects
- Display as abbreviated tags: "FIR", "POI", "STN" for fire/poison/stun immunity
- Use `ColorEffectBuff` for complete immunity, `ColorStatLabel` for partial
- Position: Below active effects section

**Success Criteria**  
- [x] Character panel shows immunity list
- [x] Complete vs partial immunity visually distinguished
- [x] Immunity information updates when equipment changes

---

### Group: Message Log & Feedback

#### 15. All Feedback to Message Log

**Priority:** High  
**Complexity:** Medium  
**Depends on:** None

**Current State**  
`pkg/wasmui/game.go:showError()` sets `lastError` displayed as floating text for timeout period. Some RPC failures go directly to error display instead of message log.

**Gap**  
Gold Box routes ALL feedback through the text log panel. The current implementation has floating error messages that violate the fixed-panel principle.

**Implementation Specification**  
- Files to modify: `pkg/wasmui/game.go`, all files calling `showError()`
- Replace `showError()` calls with `addLogMessage(msg, MessageError)`
- Remove floating error display from `Draw()` method
- Keep `lastError` field for programmatic access but don't render it separately
- Search all wasmui files for `showError` calls and migrate:
  - `adventure_screen.go`, `game.go`, `overlays.go`, `exploration.go`
- Ensure errors have proper color (`MessageError` → red) in log
- Add timestamp to log messages for debugging

**Success Criteria**  
- [x] No floating error messages appear outside fixed panels
- [x] All errors appear in message log with red color
- [x] Message log is the single source of all game feedback
- [x] Error messages include context (what failed)

---

#### 16. Combat Narration Enhancement

**Priority:** Medium  
**Complexity:** Small  
**Depends on:** None

**Current State**  
`pkg/wasmui/combat_screen.go:executeAttack()` produces narration like `"Fighter attacks Goblin — HIT for 7 damage!"`. The format is good but misses some details Gold Box would include.

**Gap**  
Gold Box narration includes attack roll results, AC comparisons, and critical hit/miss callouts. The current narration is functional but could be richer.

**Implementation Specification**  
- Files to modify: `pkg/wasmui/combat_screen.go`, `pkg/server/handlers.go` (if needed for more data)
- Extend `AttackResult` to include: `AttackRoll int`, `TargetAC int`, `IsCritical bool`
- Enhance narration format:
  - Normal hit: `"Sable attacks Orc (14 vs AC 12) — HIT for 7 damage!"`
  - Critical: `"Sable attacks Orc — CRITICAL HIT for 14 damage!!"` (double exclamation, gold color)
  - Miss: `"Sable attacks Orc (8 vs AC 12) — MISS"`
- Use `ColorGoldHi` for critical hits
- Add similar detail for spell saves: `"Goblin saves vs Sleep (DC 15) — FAILED!"`

**Success Criteria**  
- [x] Combat narration includes attack roll vs AC
- [x] Critical hits have distinct formatting and color
- [x] Spell narration includes save DC and result
- [x] All numeric combat values are explicit in log

---

#### 17. Turn Transition Announcement

**Priority:** Low  
**Complexity:** Small  
**Depends on:** None

**Current State**  
Initiative panel shows current turn with `"> "` marker. The action bar shows "YOUR TURN" or "Waiting..." but no explicit log announcement when turn changes.

**Gap**  
Gold Box announces each turn transition in the log. Players should see explicit "Round 2 begins" and "Orc's turn" messages.

**Implementation Specification**  
- Files to modify: `pkg/wasmui/combat_screen.go`, `pkg/wasmui/game.go`
- Track `lastAnnounced` struct: `{round int, turnID string}`
- In combat update loop, detect round/turn changes
- On round change: `addLogMessage("--- ROUND X BEGINS ---", MessageSystem)`
- On turn change: `addLogMessage("Orc's turn", MessageCombat)` for enemies, `addLogMessage("YOUR TURN", MessageSystem)` for player
- Use separator dashes for round starts to visually break up log

**Success Criteria**  
- [x] Round transitions announced with separator line
- [x] Each combatant's turn announced in log
- [x] Player turn has distinct "YOUR TURN" message
- [x] Announcements use appropriate message colors

---

### Group: UI Layout & Panels

#### 18. EGA-Style Bold Panel Borders

**Priority:** High  
**Complexity:** Small  
**Depends on:** None

**Current State**  
`pkg/wasmui/types_ui.go` defines `ColorPanelBorder` (#5A508280) and `ColorPanelBorderHi` (#8273B4FF). Panels use `drawRectOutline()` with these colors. The double-border technique is used but borders are subtle.

**Gap**  
Gold Box has BOLD, BRIGHT borders that clearly separate each zone. Current borders are visible but not bold enough for authentic EGA aesthetic.

**Implementation Specification**  
- Files to modify: `pkg/wasmui/types_ui.go`, all panel drawing functions
- Increase border brightness:
  - `ColorPanelBorder` → `color.RGBA{R: 120, G: 100, B: 180, A: 255}` (more vivid)
  - `ColorPanelBorderHi` → `color.RGBA{R: 180, G: 160, B: 255, A: 255}` (bright highlight)
- Increase border thickness: Draw 3 nested outlines instead of 2
- Apply to: character panel, log panel, action panel, combat grid boundary, overlays
- Add inner shadow line: dark color 1px inside bright border

**Success Criteria**  
- [x] Panel borders are clearly visible and bold
- [x] Borders use bright EGA-inspired colors
- [x] Each panel zone is distinctly separated
- [x] Visual style matches Gold Box screenshot references

---

#### 19. Command Menu Keyboard Hints

**Priority:** Medium  
**Complexity:** Small  
**Depends on:** None

**Current State**  
`pkg/wasmui/combat_screen.go:drawCombatActionBar()` shows buttons with `[M] Move`, `[A] Attack` etc. The key letters are part of the label but not highlighted differently.

**Gap**  
Gold Box highlights the keyboard shortcut letter in a distinct color within the menu label. Current implementation has the letter but without visual distinction.

**Implementation Specification**  
- Files to modify: `pkg/wasmui/combat_screen.go`, `pkg/wasmui/exploration.go`
- Create `drawKeyHintText(screen, text string, keyIndex int, x, y int, textColor, keyColor color.RGBA)`
- Render text in two parts: key letter in `ColorGold`, rest in `textColor`
- Apply to all action buttons and menu items
- Examples: "**M**ove", "**A**ttack", "**C**ast" where bold is gold color
- Handle bracket notation: `[M] Move` → brackets in textColor, M in keyColor

**Success Criteria**  
- [x] Keyboard shortcut letters highlighted in gold
- [x] Consistent across all menus and action bars
- [x] Highlights visible but not distracting
- [x] Pattern matches Gold Box visual style

---

#### 20. Viewport Aspect Ratio Enforcement

**Priority:** Low  
**Complexity:** Small  
**Depends on:** 1

**Current State**  
`pkg/wasmui/game.go` defines `ScreenWidth = 800`, `ScreenHeight = 600`. The viewport area is `screenWidth - charPanelWidth` by `screenHeight - logPanelHeight - actionPanelHeight`. Aspect ratio depends on panel sizes.

**Gap**  
Gold Box viewports maintain specific proportions for the first-person view. Current implementation doesn't enforce viewport aspect ratio, which could distort perspective rendering.

**Implementation Specification**  
- Files to modify: `pkg/wasmui/exploration.go`, constants in `game.go`
- Define target viewport aspect ratio: 4:3 (320x240 logical, scaled)
- In `drawViewport()`, calculate centered viewport that maintains ratio
- Add black bars if window doesn't match exactly
- Ensure first-person rendering (item 1) uses these fixed proportions
- Constants: `viewportBaseW = 320`, `viewportBaseH = 240`, `viewportScale = 2`

**Success Criteria**  
- [x] Viewport maintains 4:3 aspect ratio regardless of window size
- [x] Black bars appear if needed to preserve ratio
- [x] First-person rendering uses correct proportions
- [x] Pixel art scales cleanly at integer multiples

---

### Group: Animation & Visual Feedback

#### 21. Movement Transition Effect

**Priority:** Low  
**Complexity:** Medium  
**Depends on:** 1

**Current State**  
`pkg/wasmui/exploration.go:handleMove()` calls RPC and updates player position. The display updates instantly with no transition effect.

**Gap**  
While Gold Box movement is instant (discrete steps), there's often a brief screen redraw effect. The current instant update feels static.

**Implementation Specification**  
- Files to modify: `pkg/wasmui/exploration.go`
- Add `moveTransition` state: `{active bool, startTime time.Time, duration time.Duration}`
- On movement start, set transition active for 50ms
- During transition, apply brief screen shake or viewport offset:
  - Offset viewport by 2-4 pixels in movement direction, then snap back
- Alternative: Brief flash/darken of viewport (10% darker for 50ms)
- Keep transition subtle — no smooth scrolling, just transition feedback
- Constraint: Must complete within 100ms to feel responsive

**Success Criteria**  
- [x] Movement has brief visual feedback (shake or flash)
- [x] Transition is <100ms and doesn't delay input
- [x] Effect is subtle, not disorienting
- [x] Consecutive rapid moves queue correctly

---

#### 22. Loading State Indicators

**Priority:** Low  
**Complexity:** Small  
**Depends on:** None

**Current State**  
`pkg/wasmui/adventure_screen.go` shows "Loading adventures..." text during async operations. Other screens have no loading indicators during RPC calls.

**Gap**  
When loading inventory, spells, or quest log, the UI shows stale data until RPC completes. There's no visual indication that new data is being fetched.

**Implementation Specification**  
- Files to modify: `pkg/wasmui/overlays.go`, `pkg/wasmui/exploration.go`
- Add `loading bool` flag to relevant state structs
- In `loadInventory()`, `loadSpells()`, `loadQuestLog()`: Set loading=true before RPC, false after
- In draw functions, if loading==true:
  - Show "Loading..." text in panel
  - Or show spinner/pulsing dots
- Use simple text animation: "Loading", "Loading.", "Loading..", "Loading..." cycling every 300ms
- Apply to all data-loading overlays

**Success Criteria**  
- [x] Inventory shows loading indicator while fetching
- [x] Spellbook shows loading indicator while fetching
- [x] Quest log shows loading indicator while fetching
- [x] Indicator clearly visible, disappears when data loads

---

### Group: Game System Wiring

#### 23. Faction Relations Display Enhancement

**Priority:** Medium  
**Complexity:** Small  
**Depends on:** None

**Current State**  
`pkg/wasmui/overlays.go:drawFactionRelations()` displays faction list with state, opinion, and treaties. The display is functional but uses basic text formatting.

**Gap**  
Faction opinion (-100 to +100) and trust values are shown numerically but without visual representation. Gold Box–style would show these as bars or visual indicators.

**Implementation Specification**  
- Files to modify: `pkg/wasmui/overlays.go`
- In `drawFactionRelations()`, for each faction:
  - Draw opinion bar (100px wide): red for negative, green for positive, centered at 0
  - Draw trust bar similarly below opinion
  - Color-code state text: War=red, Hostile=orange, Neutral=gray, Friendly=green, Allied=gold
- Bar implementation: Background gray, filled portion colored based on value
- Position: After faction name, before treaty icons

**Success Criteria**  
- [x] Opinion displayed as horizontal bar
- [x] Trust displayed as horizontal bar
- [x] Diplomatic state has color-coded text
- [x] At-a-glance understanding of all faction standings

---

#### 24. Guild Treasury Deposit/Withdraw UI

**Priority:** Low  
**Complexity:** Medium  
**Depends on:** None

**Current State**  
`pkg/wasmui/rpc_methods.go` has `GuildDeposit(amount)` and `GuildWithdraw(amount)` methods. `pkg/wasmui/overlays.go:drawGuildInfo()` shows treasury balance but no deposit/withdraw interface.

**Gap**  
Players can see guild treasury but cannot interact with it from the UI. The RPC methods exist but aren't wired to UI controls.

**Implementation Specification**  
- Files to modify: `pkg/wasmui/overlays.go`
- In guild info panel, add two buttons: "Deposit" and "Withdraw"
- On click, show amount input overlay (reuse character creation name input pattern)
- Wire Enter to call `GuildDeposit()` or `GuildWithdraw()`
- Show result in log message: "Deposited 100 gold to guild treasury"
- Add error handling for insufficient funds
- Button positions: Below treasury balance display

**Success Criteria**  
- [x] Deposit button opens amount input
- [x] Withdraw button opens amount input
- [x] Successful transactions update treasury display
- [x] Error messages for insufficient funds

---

#### 25. Spell Targeting Interface

**Priority:** Medium  
**Complexity:** Medium  
**Depends on:** 5, 7

**Current State**  
`pkg/wasmui/overlays.go:castSelectedSpell()` casts with empty target ID and nil position. The spellbook doesn't support target selection for targeted spells.

**Gap**  
Many spells require targets (enemies, allies, locations). The current implementation casts all spells as untargeted, ignoring spell targeting requirements.

**Implementation Specification**  
- Files to modify: `pkg/wasmui/overlays.go`, `pkg/wasmui/combat_screen.go`
- Add `TargetType string` to `SpellData` (self, single, area, etc.)
- For targeted spells:
  - Close spellbook overlay
  - Set `pendingSpell SpellData` on Game
  - Enter targeting mode (similar to attack targeting)
  - Show spell range highlight on combat grid
  - On target confirm, call `CastSpell(spellID, targetID, position)`
- Self-target spells cast immediately
- Area spells need position targeting (click location on grid)

**Success Criteria**  
- [x] Single-target spells prompt for target selection
- [x] Area spells prompt for location selection
- [x] Self-target spells cast immediately
- [x] Spell range shown during targeting
- [x] Cancel returns to spellbook

---

---

## Preserved: Quality Maintenance Items

Carry forward from existing ROADMAP.md — these remain relevant.

---

#### QM-1. Reduce UI Complexity Hotspots

**Priority:** Low  
**Complexity:** Medium  
**Depends on:** None

**Current State**  
Three functions exceed complexity threshold 15: `drawCombatGrid` (19.4), `updateCharCreationName` (17.1), `Draw` in adventure_ui.go (15.8).

**Gap**  
High complexity functions are harder to maintain and modify safely.

**Implementation Specification**  
- Refactor `drawCombatGrid` (combat_screen.go:184-258):
  - Extract `drawCombatFloor()`, `drawCombatEntities()`, `drawCombatHighlights()`
- Simplify `updateCharCreationName` (character_creation.go:110-170):
  - Extract keyboard handling to helper
  - Use table-driven key mapping
- Split `Draw` in adventure_ui.go:
  - Extract `drawAdventureList()`, `drawAdventureDetail()`
- Validation: Run complexity analysis, ensure no functions >15

**Success Criteria**  
- [x] `drawCombatGrid` complexity <15
- [x] `updateCharCreationName` complexity <15
- [x] `Draw` (adventure_ui) complexity <15
- [x] No regressions in functionality

---

#### QM-2. Improve Server Package Test Coverage

**Priority:** Medium  
**Complexity:** Medium  
**Depends on:** None

**Current State**  
`pkg/server` has 78.3% coverage, lowest among core packages.

**Gap**  
Critical network layer has lower test coverage than game logic.

**Implementation Specification**  
- Add tests for `handleQuestEditorUpdate` (handlers_editor.go)
- Add concurrent session tests (session.go)
- Add WebSocket reconnection tests
- Target: 85% coverage
- Validation: `go test -cover ./pkg/server/...`

**Success Criteria**  
- [ ] `pkg/server` coverage ≥85%
- [ ] Concurrent session scenarios tested
- [ ] WebSocket edge cases covered

---

#### QM-3. Expand PCG Test Coverage

**Priority:** Low  
**Complexity:** Medium  
**Depends on:** None

**Current State**  
`pkg/pcg` has 78.9% coverage.

**Gap**  
Procedural generation edge cases may produce invalid content.

**Implementation Specification**  
- Test boundary conditions (1x1 maps, max sizes)
- Validate biome transitions
- Test deterministic seeding
- Target: 85% coverage

**Success Criteria**  
- [ ] `pkg/pcg` coverage ≥85%
- [ ] Seed determinism verified
- [ ] Edge cases don't produce invalid output

---

---

## Implementation Order

Recommended sequencing accounting for dependencies and impact:

1. **18. EGA-Style Bold Panel Borders** — Foundation visual fix, no dependencies, improves all screens
2. **15. All Feedback to Message Log** — Eliminates floating errors, enforces Gold Box UI paradigm
3. **6. Damage Flash Animation** — Simple high-impact combat feedback
4. **4. Movement Range Highlighting** — Core combat UX improvement
5. **5. Attack Range Highlighting** — Complements movement range
6. **11. Equipment Icons in Inventory** — Simple asset integration win
7. **12. Character Portrait in Panel** — Visual identity for player character
8. **16. Combat Narration Enhancement** — Richer tactical feedback
9. **8. Active Character Tile Highlight** — Combat clarity
10. **13. Action Point Cost Display** — Tactical planning support
11. **9. NPC Morale Indicator** — Surfaces hidden game system
12. **19. Command Menu Keyboard Hints** — Authentic Gold Box navigation
13. **7. Spell Effect Overlay** — Visual feedback for magic
14. **25. Spell Targeting Interface** — Completes spellcasting UX
15. **1. First-Person Dungeon Viewport** — Major visual overhaul, most complex
16. **2. Encounter Text Overlay System** — Supports narrative content
17. **17. Turn Transition Announcement** — Polish for combat log
18. **23. Faction Relations Display Enhancement** — Surfaces diplomacy system
19. **10. Cover/Flanking Visual Indicators** — Advanced tactical info
20. **3. Minimap Fog of War** — Exploration depth
21. **14. Effect Immunity Display** — Complete character info
22. **24. Guild Treasury Deposit/Withdraw UI** — Completes guild features
23. **20. Viewport Aspect Ratio Enforcement** — Polish for first-person view
24. **21. Movement Transition Effect** — Polish for exploration
25. **22. Loading State Indicators** — General UX polish
26. **QM-1. Reduce UI Complexity Hotspots** — Maintenance
27. **QM-2. Improve Server Package Test Coverage** — Maintenance
28. **QM-3. Expand PCG Test Coverage** — Maintenance

---

## Completion Criteria

The game will be "Gold Box faithful" when:

**Visual Layout**: The screen is divided into fixed, non-overlapping panels with bold EGA-inspired bright borders. The exploration viewport renders a first-person corridor view with visible walls and doors. The combat viewport renders a top-down tactical grid with movement/attack range highlights on the active character.

**Information Display**: All combat feedback flows through the scrolling message log in prose form (`"Sable attacks Orc — HIT for 7 damage"`). HP, AC, AP, damage, XP, morale, cover — all numeric game state is explicitly visible. The player never needs to guess about mechanical values.

**Visual Feedback**: Damage causes a brief flash on the target. Spells produce effect overlays. The current turn character has a pulsing highlight. Movement has a brief transition effect. Loading states are indicated.

**Asset Integration**: Character portraits display in the character panel. Equipment shows icons in inventory. Monster sprites load for enemies. All assets use the defined EGA-inspired palette from `types_ui.go`.

**Game System Visibility**: Morale state appears on NPCs in combat. Faction relations show visual opinion/trust bars. Guild treasury is interactive. Spell targeting supports all spell types.

**Functional Austerity**: No decorative chrome that doesn't serve gameplay. Retro authenticity over modern polish. Pixel art with strong readability. The aesthetic serves the tactical RPG experience, not visual flash.

When all items above are implemented, the game will look, feel, and play like an authentic modern successor to the SSI Gold Box series.
