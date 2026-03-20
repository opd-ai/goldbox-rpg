# Roadmap

Generated: 2026-03-19

## Gold Box Reference Standard

For this codebase to achieve Gold Box faithfulness, the UI must embody the distinctive SSI visual language: fixed, non-overlapping panels with bold bright borders (viewport top-left, message log bottom, party roster right, command menu bottom); a 16-color EGA palette sensibility with deep blues, magentas, purples, vivid yellows, and dungeon grays against near-black backgrounds; chunky flat-colored sprites without gradients or anti-aliasing; first-person corridor exploration with instant step-and-turn movement; top-down tactical combat on a tile grid with highlighted movement/attack ranges; dense text output flowing to the message log as the primary feedback channel; and functional austerity where every UI element serves gameplay. The current codebase has made significant progress—4:3 viewport letterboxing, EGA-inspired palette constants in `types_ui.go`, damage/spell flash effects, initiative panels, first-person corridor rendering—but gaps remain in sprite utilization, message log integration, backend system surfacing, and panel layout refinement.

---

## Improvement Items

Items are grouped by theme and ordered within each group by priority (highest impact first). Each item is specified in enough detail for autonomous implementation.

---

### Group: Combat Screen

#### 1. Wire Morale System to Combat UI

**Priority:** High
**Complexity:** Medium
**Depends on:** None

**Current State**
The morale system is fully implemented in `pkg/game/morale.go` with states (Steadfast, Shaken, Broken, Panicked), modifiers for ally death/victory/leader presence, and flee calculation logic. The `InitiativeEntry` struct in `pkg/wasmui/types_game.go:94` has a `MoraleState string` field. However, the combat screen (`pkg/wasmui/combat_screen.go`) never displays morale information for enemies, and the server doesn't populate the morale_state field in initiative responses.

**Gap**
Gold Box games showed enemy morale visually and textually—when goblins fled or orcs "broke ranks." This feedback is completely absent despite the backend support. Players cannot make tactical decisions based on enemy morale.

**Implementation Specification**
- Files to modify: `pkg/wasmui/combat_screen.go`, `pkg/server/handlers.go`, `pkg/server/websocket.go`
- In `drawInitiativeEntry()` around line 597, add morale state display after HP bar:
  ```go
  if entry.MoraleState != "" && entry.MoraleState != "Steadfast" {
      moraleColor := getMoraleColor(entry.MoraleState)
      drawColoredText(screen, entry.MoraleState, panelX+130, y, moraleColor)
  }
  ```
- Add `getMoraleColor(state string) color.RGBA` helper returning ColorEffectControl for Shaken, ColorEnemyName for Broken/Panicked
- In server combat handlers, populate InitiativeEntry.MoraleState from MoraleSystem.GetMoraleState()
- Add message log entry when morale changes: "Goblin morale is SHAKEN!" in gold

**Success Criteria**
- [x] Enemy morale state displays in initiative panel when not Steadfast
- [x] Message log announces morale state changes during combat
- [x] Morale colors: Shaken=yellow, Broken=orange, Panicked=red

---

#### 2. Display Active Effects on Combat Tokens

**Priority:** High
**Complexity:** Medium
**Depends on:** None

**Current State**
The `PlayerState.Effects` slice contains active effects (`pkg/wasmui/types_game.go:40`). Effect data includes ID, Name, Type, Duration, Remaining, Magnitude. The combat screen draws player and enemy tokens but shows no effect indicators. The character panel (`exploration.go:600+`) draws effects list when in exploration mode.

**Gap**
Gold Box games showed status icons (burning, poisoned, held, etc.) directly on combat tokens. Players could see at a glance which enemies were burning or held. Currently, players must remember effect states.

**Implementation Specification**
- Files to modify: `pkg/wasmui/combat_screen.go`
- In `drawPlayerToken()` after line 461, add effect indicator drawing:
  ```go
  if len(player.Effects) > 0 {
      g.drawEffectIndicators(screen, player.Effects, px, py-12, tileSize)
  }
  ```
- In `drawSingleEnemyToken()` after line 500, similarly add effect indicators
- Create `drawEffectIndicators(screen, effects []EffectData, x, y, maxWidth int)`:
  - Draw small colored squares (8x8) above token for each effect
  - Use ColorEffectDebuff for damage effects, ColorEffectControl for CC, ColorEffectBuff for buffs
  - Limit to 4 icons; show "+" if more
- Map effect types to colors using `types_ui.go` constants

**Success Criteria**
- [x] Active effects display as colored icons above combat tokens
- [x] Icons use appropriate colors from Gold Box palette
- [x] Maximum 4 icons shown with overflow indicator

---

#### 3. Implement Attack Roll Narration in Message Log

**Priority:** High
**Complexity:** Small
**Depends on:** None

**Current State**
Combat actions produce minimal log output. The message log (`pkg/wasmui/types_ui.go:42-74`) supports typed messages with colors (MessageCombat = purple). Attack results are not narrated with Gold Box style detail ("Fighter attacks Orc — HIT for 7 damage!").

**Gap**
Gold Box combat was defined by rich textual narration in the message log. Every attack, miss, critical, and spell result was announced. The current UI provides minimal feedback.

**Implementation Specification**
- Files to modify: `pkg/wasmui/combat_screen.go`, `pkg/wasmui/rpc_methods.go`, `pkg/server/handlers.go`
- Create `AttackResult` struct in `types_rpc.go`:
  ```go
  type AttackResult struct {
      Success     bool   `json:"success"`
      Hit         bool   `json:"hit"`
      Damage      int    `json:"damage"`
      Critical    bool   `json:"critical"`
      AttackerName string `json:"attacker_name"`
      TargetName  string `json:"target_name"`
      WeaponName  string `json:"weapon_name"`
      Message     string `json:"message"`
  }
  ```
- Update attack RPC to return rich result
- In attack callback, construct narration:
  ```go
  if result.Critical {
      g.addLogMessage(fmt.Sprintf("%s CRITICAL HIT on %s for %d damage!",
          result.AttackerName, result.TargetName, result.Damage), MessageCombat)
  } else if result.Hit {
      g.addLogMessage(fmt.Sprintf("%s hits %s for %d damage",
          result.AttackerName, result.TargetName, result.Damage), MessageCombat)
  } else {
      g.addLogMessage(fmt.Sprintf("%s attacks %s — MISS",
          result.AttackerName, result.TargetName), MessageCombat)
  }
  ```

**Success Criteria**
- [x] Every attack produces detailed message log entry
- [x] Critical hits are highlighted with emphasis
- [x] Misses are explicitly announced
- [x] Attacker, target, and damage are always named

---

#### 4. Add Opportunity Attack Visual Indicator

**Priority:** Medium
**Complexity:** Small
**Depends on:** None

**Current State**
Opportunity attacks are implemented in `pkg/game/combat_opportunity.go` with threat zone calculation and reaction tracking. The combat grid shows movement range highlighting (`drawMovementHighlights`) but doesn't indicate which tiles would provoke opportunity attacks.

**Gap**
Gold Box games made tactical consequences visible. Moving away from an enemy was a meaningful choice communicated visually.

**Implementation Specification**
- Files to modify: `pkg/wasmui/combat_screen.go`
- Add `drawOpportunityZones(screen, gridWidth, gridHeight int)` after `drawAttackHighlights`:
  - Get enemy positions from combat.Initiative
  - For each enemy, calculate adjacent tiles (threat zone)
  - Draw subtle warning tint (orange, alpha 40) on tiles in threat zones
- Add small "!" icon on tiles where movement would trigger OA
- Call from `drawCombatHighlights()` when in move mode

**Success Criteria**
- [x] Tiles adjacent to enemies show warning tint during move mode
- [x] Player can see which moves would provoke opportunity attacks
- [x] Warning uses distinct color from movement range highlight

---

#### 5. Implement Turn Order Prediction Display

**Priority:** Medium
**Complexity:** Small
**Depends on:** None

**Current State**
Initiative panel (`drawInitiativePanel`) shows current order. Combat uses standard D&D initiative. Players cannot see how their speed/dexterity affects future turn order.

**Gap**
Gold Box showed clear turn ordering. Players could plan multi-round tactics knowing initiative order.

**Implementation Specification**
- Files to modify: `pkg/wasmui/combat_screen.go`
- In `drawInitiativeList()`, add initiative score display:
  ```go
  initText := fmt.Sprintf("%d", entry.Initiative)
  drawColoredText(screen, initText, panelX+charPanelWidth-30, y+2, ColorStatLabel)
  ```
- Add "NEXT" indicator on the entry following current turn
- Highlight current turn entry with gold background tint

**Success Criteria**
- [x] Initiative values displayed in tracker
- [x] Next combatant clearly indicated
- [x] Current turn has visual highlight

---

#### 6. Wire Immunities Display to UI

**Priority:** Medium
**Complexity:** Small
**Depends on:** None

**Current State**
Effect immunities are implemented in `pkg/game/effectimmunity.go` with ImmunityManager tracking permanent and temporary immunities. `PlayerState.Immunities []string` exists in `types_game.go:43`. The server can populate this field, but the UI never displays it.

**Gap**
Gold Box showed when characters were immune to effects (e.g., "Immune to Sleep"). Players could make informed spell choices.

**Implementation Specification**
- Files to modify: `pkg/wasmui/exploration.go` (character panel), `pkg/wasmui/combat_screen.go`
- In `drawCharacterPanel()` after effects section, add:
  ```go
  if len(player.Immunities) > 0 {
      drawColoredText(screen, "Immunities:", panelX+10, y, ColorGold)
      y += 15
      for _, imm := range player.Immunities {
          drawColoredText(screen, "• "+imm, panelX+15, y, ColorEffectBuff)
          y += 12
      }
  }
  ```
- Ensure server populates Immunities from ImmunityManager

**Success Criteria**
- [x] Character panel shows immunity list when present
- [x] Immunities displayed in green (buff color)
- [x] Server correctly populates immunity data

---

### Group: Exploration Screen

#### 7. Implement Actual Map Data in First-Person View

**Priority:** High
**Complexity:** Large
**Depends on:** None

**Current State**
The first-person view (`drawFirstPersonViewAt` in `exploration.go:356-431`) renders a static placeholder corridor. It draws hardcoded depth slices with a door, but never queries the actual map. Comment at line 472: "TODO: Query server for visible walls via getVisibleWalls RPC".

**Gap**
Gold Box exploration showed actual dungeon geometry—walls, doors, openings based on real map data and player position/facing. Currently every location looks identical.

**Implementation Specification**
- Files to modify: `pkg/wasmui/exploration.go`, `pkg/wasmui/rpc_methods.go`, `pkg/server/handlers.go`
- Add `GetVisibleTiles` RPC method returning 3-deep view cone data:
  ```go
  type VisibleTile struct {
      RelativeX int    `json:"rel_x"` // -1, 0, 1 (left, center, right)
      Depth     int    `json:"depth"` // 0=near, 1=mid, 2=far
      TileType  string `json:"type"`  // wall, floor, door_open, door_closed
      Facing    int    `json:"facing"`
  }
  ```
- In `drawFirstPersonViewAt()`, call RPC to get visible tiles
- Cache result; refresh on movement
- Replace hardcoded wall drawing with map-driven logic:
  - Check left/right walls at each depth
  - Render doors where door tiles exist
  - Show openings where floor tiles exist

**Success Criteria**
- [x] First-person view reflects actual map geometry
- [x] Walls appear where map has walls
- [x] Doors visible at correct positions
- [x] View updates on player movement

---

#### 8. Add Encounter/NPC Portrait Display

**Priority:** High
**Complexity:** Medium
**Depends on:** None

**Current State**
The `EncounterOverlay` struct exists (`types_ui.go:107-116`) with `PortraitPath` field. `drawEncounterOverlay()` in `overlay_helpers.go` renders text but doesn't display portraits. Adventure NPCs have portrait paths (`AdventureNPCPath` in `asset_loader.go:256`).

**Gap**
Gold Box encounters featured NPC portraits prominently—the innkeeper, the mysterious stranger, the quest giver. This visual storytelling element is completely missing.

**Implementation Specification**
- Files to modify: `pkg/wasmui/overlay_helpers.go`, `pkg/wasmui/asset_loader.go`
- In `drawEncounterOverlay()`, before text rendering:
  ```go
  if e.PortraitPath != "" {
      portraitX := overlayX + 20
      portraitY := overlayY + 40
      portraitW, portraitH := 96, 128
      DrawAdventureSpriteWithFallback(screen, e.PortraitPath, 
          portraitX, portraitY, portraitW, portraitH,
          color.RGBA{R: 60, G: 50, B: 80, A: 255})
      // Offset text to right of portrait
      textX = portraitX + portraitW + 15
  }
  ```
- Add portrait border using Gold Box colors
- Support NPC portraits from adventure assets

**Success Criteria**
- [x] NPC encounters display portraits when path provided
- [x] Portrait has decorative border matching Gold Box style
- [x] Text flows beside portrait, not behind

---

#### 9. Implement Party Roster Panel (Multi-Character Support)

**Priority:** High
**Complexity:** Large
**Depends on:** None

**Current State**
The character panel (`drawCharacterPanel` in `exploration.go:584`) shows a single player. Gold Box games featured party-based gameplay with 4-6 characters. The backend has session management that could support multiple characters, but UI is single-character.

**Gap**
Gold Box's right panel showed the entire party roster with HP bars and status. This is core to the authentic experience.

**Implementation Specification**
- Files to modify: `pkg/wasmui/types_game.go`, `pkg/wasmui/exploration.go`, `pkg/server/handlers.go`
- Add `PartyMembers []PlayerState` to `GameStateData`
- Create `drawPartyRoster(screen, partyMembers []PlayerState, panelX, panelY, panelH int)`:
  - Vertical list of party members
  - Each entry: Name, Class, HP bar, status icons
  - Selected member highlighted with gold border
  - 50px height per member
- Replace `drawCharacterPanel()` single-player view with roster
- Add number keys 1-6 to select party member
- Selected member's full stats shown in lower portion of panel

**Success Criteria**
- [x] Party roster displays multiple characters vertically
- [x] HP bars visible for each party member
- [x] Number keys select party members
- [x] Selected member details shown below roster

---

#### 10. Add Compass Rose to Viewport

**Priority:** Medium
**Complexity:** Small
**Depends on:** None

**Current State**
Facing direction shown as text "Facing: North" at bottom of viewport (`exploration.go:241-243`). Gold Box games had a graphical compass rose.

**Gap**
Text is functional but lacks the visual polish of a compass indicator.

**Implementation Specification**
- Files to modify: `pkg/wasmui/exploration.go`
- Create `drawCompassRose(screen, x, y, facing int)`:
  - Draw 50x50 compass in corner of viewport
  - Use Gold Box colors: gold cardinal points, gray background
  - Highlight current facing direction
  - N/S/E/W letters at cardinal positions
- Replace text facing indicator with compass rose at `vpX+vpWidth-60, vpY+10`

**Success Criteria**
- [x] Compass rose displays in viewport corner
- [x] Current facing direction highlighted
- [x] Matches Gold Box visual style

---

#### 11. Implement Minimap Toggle

**Priority:** Medium
**Complexity:** Medium
**Depends on:** None

**Current State**
No minimap exists. Players have no overhead view of explored areas.

**Gap**
Gold Box games had area maps accessible via menu. Tactical awareness of dungeon layout is important.

**Implementation Specification**
- Files to modify: `pkg/wasmui/exploration.go`, `pkg/wasmui/overlays.go`, `pkg/wasmui/rpc_methods.go`
- Add `ShowMinimap bool` to `OverlayState`
- Add M key binding to toggle minimap in `handleExplorationOverlayKeys()`
- Create `drawMinimapOverlay(screen)`:
  - 200x200 overlay in center of viewport
  - Request explored tile data from server
  - Draw 4x4 pixel squares for each explored tile
  - Wall=gray, floor=dark, door=gold, player=green dot
  - Current position centered
- Add `GetExploredMap` RPC returning explored tile grid

**Success Criteria**
- [x] M key toggles minimap overlay
- [x] Explored areas shown in overhead view
- [x] Player position clearly marked
- [x] Doors and walls distinguishable

---

### Group: Character & Party Display

#### 12. Draw Equipment Slots with Item Sprites

**Priority:** High
**Complexity:** Medium
**Depends on:** None

**Current State**
`drawEquipmentSlots()` in `overlays.go:238-265` renders equipment as text list: "Head: (empty)", "Chest: Chainmail", etc. Item sprites exist at paths like `ItemIconPath(type, name)` in `asset_loader.go:225-229`.

**Gap**
Gold Box showed equipment with visual icons. The paper-doll inventory display was a key interface element.

**Implementation Specification**
- Files to modify: `pkg/wasmui/overlays.go`
- Redesign `drawEquipmentSlots()` with visual layout:
  ```go
  // Paper-doll style layout
  slots := map[string]struct{x, y int}{
      "Head":       {x: 110, y: 60},
      "Neck":       {x: 110, y: 100},
      "Chest":      {x: 110, y: 140},
      "Hands":      {x: 50, y: 140},
      "Rings":      {x: 170, y: 140},
      "Legs":       {x: 110, y: 200},
      "Feet":       {x: 110, y: 250},
      "WeaponMain": {x: 30, y: 200},
      "WeaponOff":  {x: 190, y: 200},
  }
  ```
- Draw 32x32 slot background for each position
- If item equipped, draw item sprite via `DrawSpriteWithFallback()`
- Show item name on hover/selection
- Use Gold Box panel border style for slot frames

**Success Criteria**
- [x] Equipment shown in paper-doll visual layout
- [x] Item sprites display in appropriate slots
- [x] Empty slots clearly indicated
- [x] Hover shows item name

---

#### 13. Add Character Portrait to Character Panel

**Priority:** Medium
**Complexity:** Small
**Depends on:** None

**Current State**
Character panel shows stats as text. `CharacterPortraitPath()` in `asset_loader.go:182-195` generates correct paths. `PreloadCharacterSprites()` exists. But no portrait renders in the character panel.

**Gap**
Gold Box character sheets featured character portraits prominently.

**Implementation Specification**
- Files to modify: `pkg/wasmui/exploration.go`
- In `drawCharacterPanel()` at top of panel after title:
  ```go
  if player != nil {
      portraitPath := CharacterPortraitPath(player.Class, "human", "male")
      // If appearance has gender/race, use those instead
      DrawSpriteWithFallback(screen, portraitPath, 
          panelX+50, panelY+35, 96, 96,
          color.RGBA{R: 60, G: 50, B: 80, A: 255})
      drawRectOutline(screen, panelX+48, panelY+33, 100, 100, ColorGold)
  }
  ```
- Adjust stat layout to flow below portrait
- Use player's appearance data if available (race, gender)

**Success Criteria**
- [x] Character portrait displays at top of character panel
- [x] Portrait has decorative Gold Box border
- [x] Falls back gracefully if sprite not loaded

---

#### 14. Display Attribute Modifiers

**Priority:** Medium
**Complexity:** Small
**Depends on:** None

**Current State**
Character panel shows raw attribute scores (STR: 16, DEX: 14, etc.). `AttributeModifier()` function exists in `types_game.go:147-149`. Modifiers aren't displayed.

**Gap**
Gold Box showed both score and modifier. Players need to see that 16 STR = +3 modifier.

**Implementation Specification**
- Files to modify: `pkg/wasmui/exploration.go`
- In attribute display section, format with modifier:
  ```go
  attrs := []struct{name string; val int}{
      {"STR", player.Attributes.Strength},
      {"DEX", player.Attributes.Dexterity},
      // ...
  }
  for i, attr := range attrs {
      mod := AttributeModifier(attr.val)
      modStr := fmt.Sprintf("%+d", mod)
      text := fmt.Sprintf("%s: %d (%s)", attr.name, attr.val, modStr)
      drawColoredText(screen, text, panelX+10, y+i*15, ColorStatValue)
  }
  ```
- Use green for positive, red for negative modifiers

**Success Criteria**
- [x] Attribute modifiers shown in parentheses
- [x] Positive modifiers in green, negative in red
- [x] Format matches Gold Box style

---

#### 15. Show Spell Slots / Spell Preparation

**Priority:** Medium
**Complexity:** Medium
**Depends on:** None

**Current State**
Spellbook overlay lists all spells by level. D&D-style spell slots (spells per day) aren't tracked or displayed. The backend spell system (`pkg/game/spell.go`) doesn't enforce slot limits.

**Gap**
Gold Box had memorized spell tracking. Mages prepared spells and had limited casts per day.

**Implementation Specification**
- Files to modify: `pkg/wasmui/types_game.go`, `pkg/wasmui/overlays.go`, `pkg/game/spell_manager.go`
- Add `SpellSlots map[int]int` and `UsedSlots map[int]int` to PlayerState
- In spellbook header, show slots: "Level 1: 3/4  Level 2: 2/2"
- Gray out spells where slot exhausted
- Add "Rest" action to restore slots
- Backend: implement spell slot tracking in SpellManager

**Success Criteria**
- [x] Spell slots displayed per level
- [x] Used/available clearly shown
- [x] Cannot cast when slots exhausted
- [x] Rest restores spell slots

---

### Group: Game System Wiring

#### 16. Surface Faction Relations in UI

**Priority:** High
**Complexity:** Medium
**Depends on:** None

**Current State**
Faction diplomacy fully implemented in `pkg/game/faction_relations.go` with DiplomacyManager, diplomatic states (War, Hostile, Tense, Neutral, Friendly, Allied), and diplomatic actions. RPC methods exist in `rpc_methods.go:362-440`. No UI displays faction information.

**Gap**
Gold Box games had faction mechanics affecting dialogue options and combat. Players have no visibility into faction standings.

**Implementation Specification**
- Files to modify: `pkg/wasmui/overlays.go`, `pkg/wasmui/types_rpc.go`
- Add "Faction" overlay toggled by F key
- Create `drawFactionPanel(screen)`:
  - Header: "FACTION RELATIONS"
  - List known factions with relation state
  - Color code: War=red, Hostile=orange, Tense=yellow, Neutral=white, Friendly=green, Allied=gold
- Add `GetPlayerFactions` RPC to retrieve faction standings
- Show faction reputation changes in message log

**Success Criteria**
- [x] F key opens faction relations panel
- [x] Faction standings clearly displayed
- [x] Diplomatic states color-coded
- [x] Reputation changes logged

---

#### 17. Wire Guild System to UI

**Priority:** Medium
**Complexity:** Medium
**Depends on:** None

**Current State**
Guild system implemented in `pkg/game/guild.go` with membership, ranks, treasury, perks. RPC methods exist in `rpc_methods.go:279-359`. Guild panel toggled by G key (`exploration.go:80-86`) but `loadGuildData()` and rendering not implemented.

**Gap**
Guild mechanics exist but are invisible to players.

**Implementation Specification**
- Files to modify: `pkg/wasmui/overlays.go`, `pkg/wasmui/exploration.go`
- Implement `loadGuildData()`:
  ```go
  func (g *Game) loadGuildData() {
      guild, err := g.rpcClient.GetCharacterGuild()
      if err != nil { return }
      g.mu.Lock()
      g.guildData = guild
      g.mu.Unlock()
  }
  ```
- Create `drawGuildPanel(screen)`:
  - Guild name and level
  - Player's rank and permissions
  - Treasury balance
  - Member list (scrollable)
  - Available perks
- Add GuildData struct to types_rpc.go

**Success Criteria**
- [x] G key opens guild panel
- [x] Guild info displays correctly
- [x] Member list shown
- [x] Treasury and perks visible

---

#### 18. Implement AI Behavior Display for NPCs

**Priority:** Medium
**Complexity:** Small
**Depends on:** None

**Current State**
NPC AI behavior trees implemented in `pkg/game/ai_behaviors.go` with Aggressive, Guard, Patrol, Coward archetypes. NPCs have behavior types but UI doesn't indicate them.

**Gap**
Gold Box showed NPC behavior through animations and text (fleeing, defending, etc.). Players can't predict NPC actions.

**Implementation Specification**
- Files to modify: `pkg/wasmui/combat_screen.go`, `pkg/wasmui/types_game.go`
- Add `BehaviorType string` to InitiativeEntry
- In `drawInitiativeEntry()`, show behavior indicator:
  ```go
  behaviorIcons := map[string]string{
      "aggressive": "!",
      "guard": "G",
      "patrol": "P",
      "coward": "F",
  }
  if icon, ok := behaviorIcons[entry.BehaviorType]; ok {
      drawColoredText(screen, icon, panelX+10, y, ColorEffectControl)
  }
  ```
- Log behavior when NPC acts: "Goblin (Coward) flees!"

**Success Criteria**
- [x] NPC behavior type indicated in initiative tracker
- [x] Behavior affects displayed actions in log
- [x] Icons distinguishable and meaningful

---

#### 19. Fix Stun Effect Implementation

**Priority:** Critical
**Complexity:** Medium
**Depends on:** None

**Current State**
Per `GAPS.md`, Stun effect constant exists at `pkg/game/constants.go:47` but the processing code at `pkg/game/effectbehavior.go:497` has an empty case. Combat handlers don't check for stun before allowing actions.

**Gap**
Players and NPCs with Stun can still act. This breaks tactical combat.

**Implementation Specification**
- Files to modify: `pkg/server/handlers.go`, `pkg/game/effectbehavior.go`
- In `handleMove()`, `handleAttack()`, `handleCastSpell()`:
  ```go
  if player.HasEffect(game.EffectStun) {
      return nil, errors.New("cannot act while stunned")
  }
  ```
- In `processEffectTick()` case EffectStun:
  ```go
  case EffectStun:
      target.SetActionPoints(0) // Stunned entities have no AP
      // Log stun effect
  ```
- Add UI feedback when action blocked by stun

**Success Criteria**
- [x] Stunned characters cannot move, attack, or cast
- [x] Stun message displays when action blocked
- [x] Stun effect properly processes each tick

---

#### 20. Fix Root Effect Implementation

**Priority:** Critical
**Complexity:** Small
**Depends on:** None

**Current State**
Per `GAPS.md`, Root effect constant exists at `pkg/game/constants.go:48` but processing at `effectbehavior.go:494` is empty. Movement handlers don't check for root.

**Gap**
Rooted characters can still move freely, breaking crowd-control mechanics.

**Implementation Specification**
- Files to modify: `pkg/server/handlers.go`
- In `handleMove()`:
  ```go
  if player.HasEffect(game.EffectRoot) {
      return nil, errors.New("cannot move while rooted")
  }
  ```
- Allow other actions (attack, cast) while rooted
- Add message log feedback when movement blocked

**Success Criteria**
- [x] Rooted characters cannot move
- [x] Rooted characters can still attack and cast
- [x] Root message displays when movement blocked

---

#### 21. Implement Resistance API

**Priority:** High
**Complexity:** Small
**Depends on:** None

**Current State**
Per `GAPS.md`, resistances map exists at `pkg/game/effects.go:202` but no public method to set values. `getResistanceForDamageType()` always returns 0.

**Gap**
Characters cannot gain resistance to damage types through equipment or abilities.

**Implementation Specification**
- Files to modify: `pkg/game/effects.go`
- Add public methods:
  ```go
  func (em *EffectManager) SetResistance(effectType EffectType, value float64) error {
      if value < 0 || value > 1 {
          return errors.New("resistance must be 0.0-1.0")
      }
      em.mu.Lock()
      defer em.mu.Unlock()
      em.resistances[effectType] = value
      return nil
  }
  
  func (em *EffectManager) GetResistance(effectType EffectType) float64 {
      em.mu.RLock()
      defer em.mu.RUnlock()
      return em.resistances[effectType]
  }
  ```
- Wire to equipment bonuses when equipping items

**Success Criteria**
- [x] Resistance can be set and retrieved
- [x] Resistance reduces damage correctly
- [x] Equipment can provide resistance bonuses

---

#### 22. Fix Healing Modifier Initialization

**Priority:** High
**Complexity:** Small
**Depends on:** None

**Current State**
Per `GAPS.md`, healingModifier defaults to 0.0 (Go's float64 zero value). Check at `effectbehavior.go:484-485` skips healing modification when modifier is 0.

**Gap**
Healing-over-time effects don't apply correctly when no debuff present.

**Implementation Specification**
- Files to modify: `pkg/game/effects.go`
- In `NewEffectManager()`:
  ```go
  return &EffectManager{
      // ...existing fields...
      healingModifier: 1.0, // Default: no modification
  }
  ```
- Change condition to always apply modifier:
  ```go
  healing := em.healingModifier * baseHealing
  ```

**Success Criteria**
- [x] healingModifier initialized to 1.0
- [x] Healing applies correctly without debuffs
- [x] Healing debuffs correctly reduce healing

---

#### 23. Fix Multiplicative Modifier Stacking

**Priority:** High
**Complexity:** Small
**Depends on:** None

**Current State**
Per `GAPS.md`, multiplicative modifier formula at `effectmanager.go:341` is incorrect. Two 1.2x buffs yield 2.64x instead of 1.44x.

**Gap**
Multiple buffs produce wildly inflated stats, breaking game balance.

**Implementation Specification**
- Files to modify: `pkg/game/effectmanager.go`
- Fix accumulation:
  ```go
  // Initialize multiplicative modifiers to 1.0
  for stat := range multMods {
      multMods[stat] = 1.0
  }
  // Accumulate multiplicatively
  for _, mod := range multiplicativeModifiers {
      multMods[mod.Stat] *= mod.Value
  }
  ```
- Add test verifying two 1.2x buffs = 1.44x

**Success Criteria**
- [x] Multiplicative stacking mathematically correct
- [x] Two 1.2x buffs produce 1.44x result
- [x] Test verifies correct behavior

---

### Group: Animation & Feedback

#### 24. Add Damage Number Popups

**Priority:** High
**Complexity:** Medium
**Depends on:** None

**Current State**
Damage flashes exist (`DamageFlash` in `types_ui.go:305-328`) showing colored overlay when hit. Damage amount only appears in message log.

**Gap**
Gold Box showed damage numbers visually. Modern players expect floating damage numbers.

**Implementation Specification**
- Files to modify: `pkg/wasmui/types_ui.go`, `pkg/wasmui/combat_screen.go`
- Create `DamagePopup` struct:
  ```go
  type DamagePopup struct {
      EntityID  string
      Amount    int
      IsHeal    bool
      StartTime time.Time
      Duration  time.Duration
  }
  ```
- Add `damagePopups []DamagePopup` to Game struct
- In damage callback, create popup alongside flash
- Draw popups floating upward:
  ```go
  func (g *Game) drawDamagePopups(screen *ebiten.Image, gridWidth int) {
      for _, popup := range g.damagePopups {
          if !popup.IsActive() { continue }
          // Calculate position based on entity position
          y := baseY - int(progress * 30) // Float upward
          text := fmt.Sprintf("%d", popup.Amount)
          color := ColorEnemyName
          if popup.IsHeal { color = ColorEffectBuff }
          drawColoredText(screen, text, x, y, color)
      }
  }
  ```

**Success Criteria**
- [x] Damage numbers appear on hit
- [x] Numbers float upward and fade
- [x] Healing shown in green, damage in red
- [x] Numbers visible but not intrusive

---

#### 25. Implement Spell Cast Animations

**Priority:** Medium
**Complexity:** Medium
**Depends on:** None

**Current State**
`SpellEffect` struct exists (`types_ui.go:330-400`) with frame animation support. `addSpellEffect()` creates effects. Effects render as colored expanding circles.

**Gap**
Spell effects should use actual spell sprites when available, falling back to procedural.

**Implementation Specification**
- Files to modify: `pkg/wasmui/combat_screen.go`, `pkg/wasmui/asset_loader.go`
- Create `SpellEffectPath(spellID, school string) string`:
  ```go
  func SpellEffectPath(spellID, school string) string {
      return fmt.Sprintf("effects/spells/%s_%s.png", school, spellID)
  }
  ```
- In `drawSpellEffects()`, check for sprite:
  ```go
  spritePath := SpellEffectPath(effect.SpellID, effect.SpellSchool)
  if spriteCache.IsCached(spritePath) {
      DrawSpriteScaled(screen, spritePath, x, y, tileSize, tileSize)
  } else {
      // Existing procedural fallback
      spriteCache.Get(spritePath) // Trigger load
      g.drawProceduralSpellEffect(screen, effect, x, y)
  }
  ```

**Success Criteria**
- [x] Spell effects use sprites when available
- [x] Graceful fallback to procedural effects
- [x] Effects match spell school colors

---

#### 26. Add Movement Animation Feedback

**Priority:** Medium
**Complexity:** Small
**Depends on:** None

**Current State**
Movement transition exists (`moveTransitionStart`, `moveTransitionDir` in exploration.go) with subtle viewport offset and flash. Very subtle—easy to miss.

**Gap**
Gold Box had distinct step-by-step movement feel. Current transitions could be more pronounced.

**Implementation Specification**
- Files to modify: `pkg/wasmui/exploration.go`
- Increase flash intensity in `getMoveTransitionFlashAlpha()`:
  ```go
  // Peak alpha from 30 to 50
  alpha = progress * 4 * 50 // was 30
  ```
- Add footstep-style visual at bottom of viewport during transition
- Consider brief direction indicator arrow
- Keep transitions instant (50ms) per Gold Box style

**Success Criteria**
- [x] Movement feedback more noticeable
- [x] Still maintains instant step feel
- [x] Direction of movement clear

---

#### 27. Implement Turn Change Visual Effect

**Priority:** Low
**Complexity:** Small
**Depends on:** None

**Current State**
Turn changes update `combat.CurrentTurn`. Pulsing border shows current turn entity. No additional feedback when turn changes.

**Gap**
Gold Box had clear turn transitions. Players should notice immediately when it's their turn.

**Implementation Specification**
- Files to modify: `pkg/wasmui/combat_screen.go`, `pkg/wasmui/types_ui.go`
- Add `turnChangeFlash` timer to Game
- When turn changes to player, trigger flash + message:
  ```go
  if newTurn.IsPlayer && !previousTurn.IsPlayer {
      g.addLogMessage("-- YOUR TURN --", MessageSystem)
      g.turnChangeFlash = time.Now()
  }
  ```
- Draw brief screen border pulse on turn change
- Add audio cue placeholder (when audio implemented)

**Success Criteria**
- [x] "YOUR TURN" message in log on player turn
- [x] Visual flash indicates turn change
- [x] Easy to notice even when distracted

---

### Group: Message Log

#### 28. Route All Game Events to Message Log

**Priority:** High
**Complexity:** Medium
**Depends on:** None

**Current State**
Message log receives some events (combat messages, spell casting). Many events are silent: item pickups, door opening, trap triggers, quest updates, level up.

**Gap**
Gold Box message log was the primary feedback channel. ALL game events appeared as text.

**Implementation Specification**
- Files to modify: `pkg/wasmui/exploration.go`, `pkg/wasmui/combat_screen.go`, `pkg/wasmui/overlays.go`
- Audit all RPC callbacks to add log messages:
  - Item pickup: "Found: Iron Sword"
  - Door open: "The door creaks open..."
  - Trap trigger: "A poison dart trap triggers!"
  - Quest update: "Quest Updated: Slay the Dragon (2/3)"
  - Level up: "LEVEL UP! Fighter is now level 5"
  - Gold found: "Found 50 gold pieces"
- Add message type for each category with appropriate color
- Ensure no player action is silent

**Success Criteria**
- [ ] Every item pickup logged
- [ ] Every interaction logged
- [ ] Quest progress logged
- [ ] Level ups logged with emphasis

---

#### 29. Implement Message Log Scrolling

**Priority:** Medium
**Complexity:** Small
**Depends on:** None

**Current State**
Message log shows recent messages. `maxLogMessages` limits visible count. No scrollback to review history.

**Gap**
Gold Box let players review combat history. Current log is ephemeral.

**Implementation Specification**
- Files to modify: `pkg/wasmui/exploration.go`, `pkg/wasmui/combat_screen.go`
- Increase `maxLogMessages` to 100 (from ~8)
- Add `logScrollOffset int` to Game
- Add scroll controls:
  - Page Up/Down to scroll log
  - Mouse wheel over log area
- In `drawCombatLog()`:
  ```go
  startIdx := len(messages) - visibleCount - g.logScrollOffset
  endIdx := startIdx + visibleCount
  for i, msg := range messages[startIdx:endIdx] {
      drawColoredText(screen, msg.Text, logX+5, logY+i*lineHeight, msg.Type.Color())
  }
  ```
- Show scroll indicators when more content exists

**Success Criteria**
- [x] Message log stores more history
- [x] Page Up/Down scrolls log
- [x] Scroll position indicators visible
- [x] Returns to bottom on new message

---

#### 30. Add Timestamp or Turn Number to Log Messages

**Priority:** Low
**Complexity:** Small
**Depends on:** None

**Current State**
LogMessage has `Timestamp int64` field but it's not displayed.

**Gap**
Combat round/turn context helps players track sequence of events.

**Implementation Specification**
- Files to modify: `pkg/wasmui/exploration.go`, `pkg/wasmui/combat_screen.go`
- In combat mode, prefix messages with round number:
  ```go
  prefix := ""
  if g.mode == ModeCombat && g.combat != nil {
      prefix = fmt.Sprintf("[R%d] ", g.combat.Round)
  }
  drawColoredText(screen, prefix+msg.Text, ...)
  ```
- Use dimmer color for prefix
- In exploration, optionally show time-of-day if implemented

**Success Criteria**
- [ ] Combat messages show round number
- [ ] Round prefix in muted color
- [ ] Helps track combat timeline

---

### Group: UI Layout & Panels

#### 31. Implement Gold Box-Style Panel Borders

**Priority:** High
**Complexity:** Small
**Depends on:** None

**Current State**
`drawBoldPanelBorder()` exists (referenced but implementation varies). Borders are functional but could be more authentic to EGA-era double-line style.

**Gap**
Gold Box had distinctive thick bright borders separating panels. Current borders are subtle.

**Implementation Specification**
- Files to modify: `pkg/wasmui/exploration.go`, `pkg/wasmui/overlay_helpers.go`
- Standardize `drawBoldPanelBorder(screen, x, y, w, h int)`:
  ```go
  func drawBoldPanelBorder(screen *ebiten.Image, x, y, w, h int) {
      // Outer bright border
      drawRectOutline(screen, x, y, w, h, ColorPanelBorderHi)
      // Inner darker border
      drawRectOutline(screen, x+2, y+2, w-4, h-4, ColorPanelBorder)
      // Optional shadow line
      drawLine(screen, x+3, y+h-1, x+w-3, y+h-1, ColorPanelShadow)
  }
  ```
- Apply consistently to all panels: character, log, viewport, overlays
- Increase border thickness to 3px

**Success Criteria**
- [ ] All panels have consistent bold borders
- [ ] Borders match EGA color palette
- [ ] Panels clearly visually separated

---

#### 32. Standardize Action Panel Layout

**Priority:** Medium
**Complexity:** Small
**Depends on:** None

**Current State**
Action panel in exploration (`drawActionPanel`) and combat (`drawCombatActionBar`) have different layouts. Key hints inconsistent.

**Gap**
Gold Box had consistent command menu style across modes.

**Implementation Specification**
- Files to modify: `pkg/wasmui/exploration.go`, `pkg/wasmui/combat_screen.go`
- Create unified action button style:
  ```go
  func drawActionButton(screen *ebiten.Image, x, y int, label, key string, selected bool) {
      bgColor := color.RGBA{R: 50, G: 50, B: 70, A: 255}
      if selected { bgColor = color.RGBA{R: 70, G: 70, B: 100, A: 255} }
      drawRect(screen, x, y, 100, 30, bgColor)
      drawRectOutline(screen, x, y, 100, 30, ColorPanelBorder)
      drawKeyHintText(screen, fmt.Sprintf("[%s] %s", key, label), x+5, y+8, ColorStatValue, ColorGold)
  }
  ```
- Use for both exploration and combat actions
- Consistent 100x30 button size
- Key letter always highlighted in gold

**Success Criteria**
- [ ] Action buttons consistent across modes
- [ ] Key hints clearly visible
- [ ] Selected state obvious

---

#### 33. Add Panel Title Headers

**Priority:** Low
**Complexity:** Small
**Depends on:** None

**Current State**
Some panels have titles (CHARACTER, INITIATIVE), others don't. Inconsistent styling.

**Gap**
Gold Box panels had clear header bars with titles.

**Implementation Specification**
- Files to modify: `pkg/wasmui/exploration.go`, `pkg/wasmui/overlays.go`, `pkg/wasmui/combat_screen.go`
- Create `drawPanelHeader(screen, x, y, w int, title string)`:
  ```go
  func drawPanelHeader(screen *ebiten.Image, x, y, w int, title string) {
      // Header bar background
      drawRect(screen, x, y, w, 25, color.RGBA{R: 40, G: 38, B: 55, A: 255})
      // Title centered
      textX := x + (w - len(title)*6) / 2
      drawColoredText(screen, title, textX, y+5, ColorGold)
      // Separator line
      drawLine(screen, x, y+24, x+w, y+24, ColorPanelBorder)
  }
  ```
- Apply to: CHARACTER, COMBAT LOG, INITIATIVE, INVENTORY, SPELLBOOK, etc.

**Success Criteria**
- [ ] All panels have consistent header bars
- [ ] Titles centered in gold
- [ ] Headers visually distinguished from content

---

### Group: Asset Integration

#### 34. Implement Monster Sprite Loading

**Priority:** High
**Complexity:** Medium
**Depends on:** None

**Current State**
`MonsterSpritePath()` in `asset_loader.go:208-216` generates paths. Monster sprites exist in `web/static/assets/sprites/` subdirectories (beasts, demons, dragons, humanoids, undead). `drawSingleEnemyToken()` calls `DrawSpriteWithFallback()` but falls back to red squares because paths don't match actual file structure.

**Gap**
Monster sprites exist but aren't displayed. All enemies appear as red squares with "E".

**Implementation Specification**
- Files to modify: `pkg/wasmui/asset_loader.go`
- Fix `MonsterSpritePath()` to match actual asset structure:
  ```go
  func MonsterSpritePath(monsterType string) string {
      typeLower := strings.ToLower(strings.ReplaceAll(monsterType, " ", "_"))
      // Check category folders
      categories := map[string][]string{
          "undead":    {"skeleton", "zombie", "ghoul", "vampire", "lich"},
          "humanoids": {"goblin", "orc", "ogre", "troll", "hobgoblin"},
          "dragons":   {"dragon"},
          "beasts":    {"wolf", "spider", "rat", "bear"},
          "demons":    {"demon", "imp"},
      }
      for cat, monsters := range categories {
          for _, m := range monsters {
              if strings.Contains(typeLower, m) {
                  return fmt.Sprintf("%s/monster_%s.png", cat, typeLower)
              }
          }
      }
      return fmt.Sprintf("monsters/monster_%s.png", typeLower)
  }
  ```
- Verify paths match actual files in web/static/assets/sprites/

**Success Criteria**
- [ ] Monster sprites display in combat
- [ ] Paths resolve to actual asset files
- [ ] Fallback only when sprite truly missing

---

#### 35. Load Terrain Sprites for Combat Grid

**Priority:** Medium
**Complexity:** Small
**Depends on:** None

**Current State**
`drawCombatFloor()` calls `DrawSpriteWithFallback()` for floor tiles but uses generic path. Terrain sprites exist in `terrain/dungeon/`, `terrain/outdoor/`.

**Gap**
Combat grid shows fallback colors instead of terrain sprites.

**Implementation Specification**
- Files to modify: `pkg/wasmui/combat_screen.go`
- In `drawCombatFloor()`, use correct terrain paths:
  ```go
  floorPath := TerrainTilePath("floor_stone", "dungeon")
  // Verify this matches actual file: terrain/dungeon/tile_floor_stone.png
  ```
- Add variety: alternate between floor tiles for visual interest
- Preload terrain sprites at combat start

**Success Criteria**
- [ ] Terrain sprites display in combat
- [ ] Floor tiles vary for visual interest
- [ ] No placeholder rectangles for terrain

---

#### 36. Implement UI Element Sprite Loading

**Priority:** Medium
**Complexity:** Medium
**Depends on:** None

**Current State**
UI sprites exist in `web/static/assets/sprites/ui/`, `buttons/`, `panels/`, `icons/`. `UIElementPath()` in `asset_loader.go:232-235` generates paths. UI rendering uses `drawRect()` and `drawColoredText()` instead of sprites.

**Gap**
UI could use sprite-based buttons, panels, and icons for more authentic look.

**Implementation Specification**
- Files to modify: `pkg/wasmui/asset_loader.go`, `pkg/wasmui/overlays.go`
- Create `drawUIButton(screen, x, y, w, h int, state string)`:
  ```go
  func drawUIButton(screen *ebiten.Image, x, y, w, h int, state string) {
      path := fmt.Sprintf("buttons/button_%s.png", state) // normal, hover, pressed
      if spriteCache.IsCached(path) {
          DrawSpriteScaled(screen, path, x, y, w, h)
      } else {
          spriteCache.Get(path)
          // Fallback to drawRect
      }
  }
  ```
- Create similar helpers for panels, icons
- Gradually replace procedural UI with sprites

**Success Criteria**
- [ ] Buttons use sprite assets when available
- [ ] Panel frames use sprite assets
- [ ] Graceful fallback to procedural rendering

---

#### 37. Add Effect Status Icons

**Priority:** Medium
**Complexity:** Small
**Depends on:** None

**Current State**
Effect sprites exist in `effects/status/` directory. Effects displayed as colored squares or text.

**Gap**
Status effect icons would improve combat readability.

**Implementation Specification**
- Files to modify: `pkg/wasmui/combat_screen.go`, `pkg/wasmui/asset_loader.go`
- Create `StatusEffectIconPath(effectType string) string`:
  ```go
  func StatusEffectIconPath(effectType string) string {
      return fmt.Sprintf("effects/status/effect_status_%s.png", strings.ToLower(effectType))
  }
  ```
- In `drawEffectIndicators()`, use sprites:
  ```go
  iconPath := StatusEffectIconPath(effect.Type)
  DrawSpriteWithFallback(screen, iconPath, x+i*10, y, 8, 8, effectColor)
  ```

**Success Criteria**
- [ ] Status effects use icon sprites
- [ ] Icons clearly represent effect type
- [ ] Fallback to colored squares if no sprite

---

### Group: Quality Maintenance Items

These items are carried forward from general code quality concerns.

#### 38. Increase Test Coverage to 70%

**Priority:** Medium
**Complexity:** Large
**Depends on:** None

**Current State**
Test coverage is 65-96% depending on package. Some files in `pkg/wasmui/` have limited coverage due to WASM constraints.

**Gap**
Target coverage is ≥70% for all critical packages.

**Implementation Specification**
- Run `make find-untested` to identify coverage gaps
- Focus on: `pkg/game/effectbehavior.go`, `pkg/server/handlers.go`
- Add table-driven tests for uncovered functions
- Use mocks for external dependencies

**Success Criteria**
- [ ] Overall coverage ≥70%
- [ ] All critical game logic covered
- [ ] No untested error paths

---

#### 39. Add WebSocket Race Condition Fixes

**Priority:** Critical
**Complexity:** Medium
**Depends on:** None

**Current State**
Per `GAPS.md`, session fields `Connected` and `WSConn` modified at `websocket.go:254-267` without mutex, creating TOCTOU races with broadcast operations.

**Gap**
Concurrent connection/disconnection with broadcast can cause panics.

**Implementation Specification**
- Files to modify: `pkg/server/websocket.go`, `pkg/server/session.go`
- Wrap session field modifications in mutex:
  ```go
  s.mu.Lock()
  session.Connected = false
  session.WSConn = nil
  s.mu.Unlock()
  ```
- Add reference counting in broadcast snapshot
- Consider `atomic.Bool` for Connected flag
- Add `TestConcurrentDisconnectDuringBroadcast` test

**Success Criteria**
- [ ] No races with `-race` flag on 100 iterations
- [ ] Broadcast handles disconnection gracefully
- [ ] Test reproduces and verifies fix

---

#### 40. Document RPC API Completely

**Priority:** Low
**Complexity:** Medium
**Depends on:** None

**Current State**
`pkg/README-RPC.md` exists with some method documentation. Many new RPC methods undocumented.

**Gap**
Developers need complete API reference.

**Implementation Specification**
- Files to modify: `pkg/README-RPC.md`
- Document all methods in `rpc_methods.go`:
  - Method name
  - Parameters with types
  - Return value structure
  - Example request/response
- Generate from code comments where possible

**Success Criteria**
- [ ] All RPC methods documented
- [ ] Examples for each method
- [ ] Types clearly defined

---

## Preserved: Quality Maintenance Items

The following items are carried forward from previous quality maintenance efforts:

#### 41. Reduce Cyclomatic Complexity in Large Functions

**Priority:** Low
**Complexity:** Medium
**Depends on:** None

**Current State**
Some handler functions in `pkg/server/handlers.go` exceed 50 lines with multiple nested conditionals.

**Gap**
High complexity reduces maintainability and test coverage.

**Implementation Specification**
- Identify functions with >10 cyclomatic complexity via `go-critic` or similar
- Extract helper functions for distinct logical blocks
- Reduce nesting with early returns
- Target: no function >50 lines, no complexity >10

**Success Criteria**
- [ ] No function exceeds 50 lines
- [ ] Cyclomatic complexity ≤10 for all functions
- [ ] Existing tests still pass

---

## Implementation Order

A recommended sequencing of ALL items above, accounting for dependencies and risk:

1. **19. Fix Stun Effect Implementation** — Critical bug affecting combat; blocks tactical gameplay
2. **20. Fix Root Effect Implementation** — Critical bug affecting combat; simple fix
3. **39. Add WebSocket Race Condition Fixes** — Critical stability issue; prevents production deployment
4. **21. Implement Resistance API** — High-priority backend fix enabling character progression
5. **22. Fix Healing Modifier Initialization** — High-priority balance fix; simple change
6. **23. Fix Multiplicative Modifier Stacking** — High-priority balance fix; simple change
7. **3. Implement Attack Roll Narration** — High-impact UX improvement; defines Gold Box feedback style
8. **31. Implement Gold Box-Style Panel Borders** — Visual foundation for all other UI work
9. **28. Route All Game Events to Message Log** — Core Gold Box interaction pattern
10. **1. Wire Morale System to Combat UI** — High-value backend surfacing
11. **2. Display Active Effects on Combat Tokens** — Important tactical feedback
12. **34. Implement Monster Sprite Loading** — Critical for visual fidelity
13. **7. Implement Actual Map Data in First-Person View** — Large but essential for exploration
14. **9. Implement Party Roster Panel** — Core Gold Box feature; large but foundational
15. **8. Add Encounter/NPC Portrait Display** — Important narrative element
16. **24. Add Damage Number Popups** — High-impact visual feedback
17. **12. Draw Equipment Slots with Item Sprites** — Important inventory UX
18. **13. Add Character Portrait to Character Panel** — Visual polish
19. **16. Surface Faction Relations in UI** — Backend system visibility
20. **17. Wire Guild System to UI** — Backend system visibility
21. **4. Add Opportunity Attack Visual Indicator** — Tactical feedback
22. **5. Implement Turn Order Prediction Display** — Tactical information
23. **6. Wire Immunities Display to UI** — Combat information
24. **10. Add Compass Rose to Viewport** — Visual polish
25. **11. Implement Minimap Toggle** — Navigation aid
26. **14. Display Attribute Modifiers** — Character information
27. **15. Show Spell Slots / Spell Preparation** — Spellcasting depth
28. **18. Implement AI Behavior Display** — Combat information
29. **25. Implement Spell Cast Animations** — Visual polish
30. **26. Add Movement Animation Feedback** — UX polish
31. **35. Load Terrain Sprites for Combat Grid** — Visual polish
32. **36. Implement UI Element Sprite Loading** — Visual polish
33. **37. Add Effect Status Icons** — Visual polish
34. **29. Implement Message Log Scrolling** — UX improvement
35. **32. Standardize Action Panel Layout** — Consistency
36. **33. Add Panel Title Headers** — Consistency
37. **27. Implement Turn Change Visual Effect** — UX polish
38. **30. Add Timestamp to Log Messages** — UX detail
39. **38. Increase Test Coverage to 70%** — Quality maintenance
40. **40. Document RPC API Completely** — Documentation
41. **41. Reduce Cyclomatic Complexity** — Code quality

---

## Completion Criteria

When all items above are implemented, the GoldBox RPG Engine will embody authentic Gold Box presentation and interaction:

**Visual Fidelity**: The screen layout will feature fixed, non-overlapping panels with bold EGA-palette borders. The viewport will show first-person dungeon corridors rendered from actual map data with proper depth and door rendering. Combat will display a tactical grid where player and monster sprites from the asset library replace placeholder rectangles. Status effects, morale states, and immunities will be visible on combat tokens. Equipment will appear in a paper-doll display with item sprites.

**Information Density**: The message log will receive every game event—attacks with hit/miss/damage detail, item pickups, door interactions, trap triggers, quest updates, level ups—in Gold Box narrative style. All numbers (HP, AC, damage, XP, gold) will be explicitly shown. Attribute scores will display with modifiers. Turn order, round numbers, and initiative values will be clearly visible.

**Backend Surfacing**: Every implemented game system will be accessible through the UI—factions with diplomatic standings, guilds with membership and treasury, morale affecting NPC behavior with visible state, resistances and immunities from equipment and abilities, spell slots limiting caster resources.

**Interaction Authenticity**: Movement will feel instant (step-and-turn) with brief visual feedback. Combat will follow turn-based tactical flow with clear current-turn indication, movement/attack range highlighting, and opportunity attack warnings. All actions will have keyboard shortcuts with highlighted letters in command menus.

**Technical Quality**: Critical bugs (Stun, Root, WebSocket races, modifier stacking) will be fixed. Test coverage will meet the 70% target. RPC API will be fully documented. Code complexity will remain manageable.

The result will be a game that feels immediately familiar to anyone who played Pool of Radiance or Curse of the Azure Bonds, while leveraging modern browser technology for accessibility.
