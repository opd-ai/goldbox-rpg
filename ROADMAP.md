# Roadmap

Generated: 2026-03-20

## Gold Box–Inspired Reference Standard

This codebase targets a **retro-inspired aesthetic** grounded in the SSI Gold Box series (Pool of Radiance, Curse of the Azure Bonds, Champions of Krynn) while embracing modern enhancements. The Gold Box reference defines fixed non-overlapping panels (viewport, message log, party roster, command menu) with bold EGA-palette borders; 16-color sensibility using deep blues, magentas, purples, vivid yellows, and dungeon grays against near-black backgrounds; chunky flat-colored sprites without gradients; first-person step-and-turn dungeon exploration with brief pseudo-smooth transitions; top-down tactical combat on a tile grid with highlighted movement/attack ranges; and dense text output flowing to the message log as the primary feedback channel.

The current implementation has made substantial progress: 4:3 aspect-ratio viewport with letterboxing, EGA-inspired palette constants (`types_ui.go`), themed first-person view rendering with five dungeon palettes (classic, horror, natural, undead, magical), architectural features (pillars, altars, fountains, archways, alcoves, rubble), damage flash effects, damage number popups, spell effect animations, initiative panels with HP bars, opportunity attack zone warnings, cover/flanking indicators, compass rose, movement transition effects, party member selection (1-6 keys), and touch/mobile input support. However, gaps remain in backend system surfacing (morale, factions, guilds), effect display on combat tokens, complete attack narration, sprite asset path resolution, paper-doll equipment display, spell slot tracking, and message log interactivity.

---

## Improvement Items

Items are grouped by theme and ordered within each group by priority (highest impact first). Each item is specified in enough detail for autonomous implementation.

---

### Group: Critical Bug Fixes

#### 1. Fix Stun Effect Enforcement in Combat Handlers

**Priority:** Critical
**Complexity:** Medium
**Depends on:** None

**Current State**
The Stun effect constant exists at `pkg/game/constants.go:47` (EffectStun). The `processEffectTick()` function in `pkg/game/effectbehavior.go:513` has a case for EffectStun that defers to "checked by combat system in handlers.go". However, the combat handlers in `pkg/server/handlers.go` do not actually check for stun before allowing move, attack, or cast actions. The effect is tracked but not enforced.

**Gap**
Gold Box combat relied on crowd-control effects to create tactical depth. A stunned character losing their turn was a meaningful consequence. Currently, players and NPCs with Stun can still act normally, breaking tactical gameplay and making stun-inducing abilities worthless.

**Implementation Specification**
- Files to modify: `pkg/server/handlers.go`, `pkg/game/character.go`
- Add `HasEffect(effectType string) bool` method to Character if not present
- In `handleMove()` (around line 250), add check before movement processing:
  ```go
  if player.HasEffect(game.EffectStun) {
      return map[string]interface{}{"success": false, "message": "Cannot move while stunned"}, nil
  }
  ```
- In `handleAttack()`, add same check before attack processing
- In `handleCastSpell()`, add same check before spell processing
- When action is blocked, emit message log event: "Stunned - cannot act!"
- Consider adding `SetActionPoints(0)` in effect tick to reinforce

**Success Criteria**
- [x] Stunned characters cannot move, attack, or cast spells
- [x] Blocked action returns descriptive error message
- [x] Message log announces when action blocked by stun
- [x] Test case verifies stun blocks all action types

---

#### 2. Fix Root Effect Movement Restriction

**Priority:** Critical
**Complexity:** Small
**Depends on:** None

**Current State**
Root effect constant exists at `pkg/game/constants.go:48` (EffectRoot). The `processEffectTick()` in `effectbehavior.go:513` has an empty case that defers to handlers. Movement handlers do not check for root status.

**Gap**
Root should prevent movement while allowing attacks and spells. This is a core crowd-control mechanic. Currently rooted characters can move freely.

**Implementation Specification**
- Files to modify: `pkg/server/handlers.go`
- In `handleMove()` only (not attack/cast):
  ```go
  if player.HasEffect(game.EffectRoot) {
      return map[string]interface{}{"success": false, "message": "Cannot move while rooted"}, nil
  }
  ```
- Message log entry when movement blocked: "Rooted in place!"

**Success Criteria**
- [x] Rooted characters cannot move
- [x] Rooted characters CAN still attack and cast spells
- [x] Root message displays when movement blocked
- [x] Test case verifies root blocks only movement

---

#### 3. Fix WebSocket Race Conditions in Session Management

**Priority:** Critical
**Complexity:** Medium
**Depends on:** None

**Current State**
Per `GAPS.md`, session fields `Connected` and `WSConn` are modified at `pkg/server/websocket.go:254-267` without mutex protection, creating TOCTOU races with concurrent `broadcastToAll` operations. The broadcast iterates sessions while connection state can change.

**Gap**
Concurrent connection/disconnection during broadcast can cause panics or data races. This is a production stability issue.

**Implementation Specification**
- Files to modify: `pkg/server/websocket.go`, `pkg/server/session.go`
- Wrap session field modifications in session-level mutex:
  ```go
  session.mu.Lock()
  session.Connected = false
  session.WSConn = nil
  session.mu.Unlock()
  ```
- In `broadcastToAll()`, snapshot session list under read lock, then iterate snapshot
- Consider `sync/atomic.Bool` for `Connected` flag for lock-free reads
- Add test `TestConcurrentDisconnectDuringBroadcast` with 100 goroutines

**Success Criteria**
- [x] No races detected with `go test -race` over 100 iterations
- [x] Broadcast handles mid-iteration disconnection gracefully
- [x] No panics during stress testing with concurrent connections

---

#### 4. Fix Healing Modifier Zero Initialization

**Priority:** High
**Complexity:** Small
**Depends on:** None

**Current State**
Per `GAPS.md`, `healingModifier` in `pkg/game/effects.go` defaults to Go's float64 zero value (0.0). The condition at `effectbehavior.go:506` multiplies healing by this modifier. With default 0.0, all healing-over-time effects apply zero healing.

**Gap**
Regeneration and healing effects don't work correctly without a debuff first modifying the healing rate. This breaks baseline healing mechanics.

**Implementation Specification**
- Files to modify: `pkg/game/effects.go`
- In `NewEffectManager()`, initialize:
  ```go
  return &EffectManager{
      // ...existing fields...
      healingModifier: 1.0, // Default: no modification
  }
  ```

**Success Criteria**
- [x] healingModifier initialized to 1.0
- [x] Healing-over-time effects apply correct amounts without debuffs
- [x] Healing debuffs correctly reduce healing rate
- [x] Test verifies baseline healing works

---

#### 5. Fix Multiplicative Modifier Stacking Formula

**Priority:** High
**Complexity:** Small
**Depends on:** None

**Current State**
Per `GAPS.md`, multiplicative modifier formula at `pkg/game/effectmanager.go:341` is incorrect. Two 1.2x buffs produce 2.64x instead of correct 1.44x (1.2 × 1.2).

**Gap**
Multiple buffs produce wildly inflated stats, breaking game balance. A character with two haste effects becomes overpowered.

**Implementation Specification**
- Files to modify: `pkg/game/effectmanager.go`
- Fix accumulation starting from 1.0:
  ```go
  // Initialize multiplicative modifiers to 1.0
  multMods := make(map[string]float64)
  for stat := range affectedStats {
      multMods[stat] = 1.0
  }
  // Accumulate multiplicatively
  for _, mod := range multiplicativeModifiers {
      multMods[mod.Stat] *= mod.Value
  }
  ```
- Add test: two 1.2x buffs should produce exactly 1.44x

**Success Criteria**
- [x] Two 1.2x buffs produce 1.44x modifier
- [x] Three 1.1x buffs produce 1.331x modifier
- [x] Test case verifies correct multiplicative stacking

---

#### 6. Implement Resistance API

**Priority:** High
**Complexity:** Small
**Depends on:** None

**Current State**
Per `GAPS.md`, resistances map exists at `pkg/game/effects.go:202` but no public method to set values. `getResistanceForDamageType()` always returns 0.

**Gap**
Characters cannot gain resistance to damage types through equipment, abilities, or spells. Fire resistance armor is non-functional.

**Implementation Specification**
- Files to modify: `pkg/game/effects.go`
- Add public methods:
  ```go
  func (em *EffectManager) SetResistance(damageType string, value float64) error {
      if value < 0 || value > 1 {
          return errors.New("resistance must be 0.0-1.0")
      }
      em.mu.Lock()
      defer em.mu.Unlock()
      em.resistances[damageType] = value
      return nil
  }

  func (em *EffectManager) GetResistance(damageType string) float64 {
      em.mu.RLock()
      defer em.mu.RUnlock()
      return em.resistances[damageType]
  }
  ```
- Wire to equipment: when equipping item with fire_resistance property, call SetResistance

**Success Criteria**
- [x] Resistance can be set and retrieved via public API
- [x] Resistance reduces damage of matching type correctly
- [x] Equipment can provide resistance bonuses when equipped
- [x] Test verifies 50% fire resistance halves fire damage

---

### Group: Combat Screen

#### 7. Wire Morale System to Combat UI

**Priority:** High
**Complexity:** Medium
**Depends on:** None

**Current State**
The morale system is fully implemented in `pkg/game/morale.go` with states (Steadfast, Shaken, Broken, Panicked), modifiers for ally death/victory/leader presence, and flee calculation logic. The `InitiativeEntry` struct in `pkg/wasmui/types_game.go:96` has a `MoraleState string` field. However, the combat screen (`pkg/wasmui/combat_screen.go`) never displays morale information for enemies, and the server doesn't populate the morale_state field in initiative responses.

**Gap**
Gold Box games showed enemy morale visually and textually—when goblins fled or orcs "broke ranks." Players used this information for tactical decisions (focus weak morale enemies, protect party leader). This feedback is completely absent despite full backend support.

**Implementation Specification**
- Files to modify: `pkg/wasmui/combat_screen.go`, `pkg/server/handlers.go`, `pkg/server/websocket.go`
- Server side: In combat handlers that build InitiativeEntry, populate MoraleState:
  ```go
  entry.MoraleState = game.MoraleStateString(moraleSystem.GetMoraleState(npcID))
  ```
- Client side: In `drawInitiativeEntry()` around line 597, add morale state display after HP bar:
  ```go
  if entry.MoraleState != "" && entry.MoraleState != "Steadfast" {
      moraleColor := getMoraleColor(entry.MoraleState)
      drawColoredText(screen, entry.MoraleState, panelX+130, y, moraleColor)
  }
  ```
- Add `getMoraleColor(state string) color.RGBA` helper:
  - Shaken: ColorEffectControl (yellow)
  - Broken: color.RGBA{R: 255, G: 165, B: 0, A: 255} (orange)
  - Panicked: ColorEnemyName (red)
- Add message log entry when morale changes: "Goblin morale is SHAKEN!"

**Success Criteria**
- [x] Enemy morale state displays in initiative panel when not Steadfast
- [x] Message log announces morale state changes during combat
- [x] Morale colors match severity (yellow → orange → red)
- [x] Steadfast enemies show no morale indicator (clean UI)

---

#### 8. Display Active Effects on Combat Tokens

**Priority:** High
**Complexity:** Medium
**Depends on:** None

**Current State**
The `PlayerState.Effects` slice contains active effects with ID, Name, Type, Duration, Remaining, Magnitude (`types_game.go:40`). The `InitiativeEntry` struct has an `Effects` field (`types_game.go:98`). Combat tokens are drawn in `drawPlayerToken()` (line 461+) and `drawSingleEnemyToken()` (line 500+) but show no effect indicators. Effect display only appears in exploration mode character panel.

**Gap**
Gold Box showed status icons (burning, poisoned, held) directly on combat tokens. Players could see at a glance which enemies were affected by their DoTs or crowd control. Currently, players must remember effect states, reducing tactical clarity.

**Implementation Specification**
- Files to modify: `pkg/wasmui/combat_screen.go`
- Create `drawEffectIndicators(screen *ebiten.Image, effects []EffectData, x, y, maxWidth int)`:
  ```go
  func (g *Game) drawEffectIndicators(screen *ebiten.Image, effects []EffectData, x, y, maxWidth int) {
      iconSize := 8
      spacing := 2
      maxIcons := 4
      
      for i, effect := range effects {
          if i >= maxIcons {
              // Draw "+" indicator for overflow
              drawColoredText(screen, "+", x+i*(iconSize+spacing), y, ColorGold)
              break
          }
          
          effectColor := getEffectColor(effect.Type)
          drawRect(screen, x+i*(iconSize+spacing), y, iconSize, iconSize, effectColor)
          drawRectOutline(screen, x+i*(iconSize+spacing), y, iconSize, iconSize, ColorPanelBorder)
      }
  }
  ```
- Add `getEffectColor(effectType string) color.RGBA`:
  - Damage effects (burning, bleeding, poison): ColorEffectDebuff (red)
  - Control effects (stun, root, paralysis): ColorEffectControl (yellow)
  - Buff effects (haste, regeneration): ColorEffectBuff (green)
  - Default: ColorEffectDefault (purple)
- In `drawPlayerToken()` after token drawing, add:
  ```go
  if len(player.Effects) > 0 {
      g.drawEffectIndicators(screen, player.Effects, px-tileSize/2, py-12, tileSize)
  }
  ```
- Similarly add to `drawSingleEnemyToken()`

**Success Criteria**
- [x] Active effects display as colored icons above combat tokens
- [x] Icons use appropriate colors from Gold Box palette
- [x] Maximum 4 icons shown with "+" overflow indicator
- [x] Both player and enemy tokens show effect indicators

---

#### 9. Implement Rich Attack Roll Narration

**Priority:** High
**Complexity:** Medium
**Depends on:** None

**Current State**
Combat actions produce minimal log output. The message log (`types_ui.go:42-87`) supports typed messages with colors including MessageCombat (purple). Attack results are not narrated with Gold Box style detail. Current messages are brief like "Hit!" without naming attacker, target, or damage.

**Gap**
Gold Box combat was defined by rich textual narration. Every action was announced: "Fighter attacks Orc — HIT for 7 damage!", "Goblin casts Sleep — FAILED!", "Critical Hit! Paladin smites Skeleton for 14 damage!". This narrative style made combat feel like a tabletop RPG session.

**Implementation Specification**
- Files to modify: `pkg/wasmui/combat_screen.go`, `pkg/wasmui/types_rpc.go`, `pkg/server/handlers.go`
- Enhance `AttackResult` struct in server response (or create if not present):
  ```go
  type AttackResult struct {
      Success      bool   `json:"success"`
      Hit          bool   `json:"hit"`
      Damage       int    `json:"damage"`
      Critical     bool   `json:"critical"`
      AttackerName string `json:"attacker_name"`
      TargetName   string `json:"target_name"`
      WeaponName   string `json:"weapon_name,omitempty"`
      DamageType   string `json:"damage_type,omitempty"`
      Message      string `json:"message"`
  }
  ```
- In attack callback handler, construct rich narration:
  ```go
  var msg string
  if result.Critical {
      msg = fmt.Sprintf("CRITICAL! %s devastates %s for %d damage!",
          result.AttackerName, result.TargetName, result.Damage)
  } else if result.Hit {
      msg = fmt.Sprintf("%s hits %s for %d damage",
          result.AttackerName, result.TargetName, result.Damage)
  } else {
      msg = fmt.Sprintf("%s attacks %s — MISS",
          result.AttackerName, result.TargetName)
  }
  g.addLogMessage(msg, MessageCombat)
  ```
- Add special messages for: killing blow, damage type (fire/ice/etc.), weapon used

**Success Criteria**
- [x] Every attack produces detailed message log entry
- [x] Attacker and target names always included
- [x] Critical hits emphasized with "CRITICAL!"
- [x] Misses explicitly announced
- [x] Damage amount always shown for hits
- [x] Killing blows get special message ("Goblin is slain!")

---

#### 10. Add Opportunity Attack Visual Indicators

**Priority:** Medium
**Complexity:** Small
**Depends on:** None

**Current State**
Opportunity attacks are implemented in `pkg/game/combat_opportunity.go` with threat zone calculation and reaction tracking. The combat grid shows movement range highlighting via `drawMovementHighlights()` and already calls `drawOpportunityZones()` (line 334). The implementation exists but may need verification of visibility.

**Gap**
Gold Box games made tactical consequences visible. Moving away from an enemy was a meaningful choice communicated visually. Players need clear indication of danger zones.

**Implementation Specification**
- Files to verify/modify: `pkg/wasmui/combat_screen.go`
- Verify `drawOpportunityZones()` is rendering correctly:
  - Check that enemy positions are correctly identified
  - Verify threat zone tiles (adjacent to enemies) receive warning tint
  - Ensure warning color is distinct from movement range (orange vs blue)
- If not working, implement:
  ```go
  func (g *Game) drawOpportunityZones(screen *ebiten.Image, tilesW, tilesH, gridW, gridH int) {
      warningColor := color.RGBA{R: 255, G: 165, B: 0, A: 60} // Orange tint
      
      // Get enemy positions
      g.mu.RLock()
      combat := g.combat
      g.mu.RUnlock()
      
      if combat == nil { return }
      
      for _, entry := range combat.Initiative {
          if entry.IsPlayer { continue }
          // Draw warning on all adjacent tiles
          // ... threat zone calculation
      }
  }
  ```
- Add small "!" or sword icon on tiles where movement triggers OA

**Success Criteria**
- [x] Tiles adjacent to enemies show warning tint during move mode
- [x] Warning color (orange) distinct from movement range (blue)
- [x] Player can clearly see which moves provoke opportunity attacks
- [x] Icon or text indicator on dangerous tiles

---

#### 11. Implement Turn Order Prediction Display

**Priority:** Medium
**Complexity:** Small
**Depends on:** None

**Current State**
Initiative panel (`drawInitiativePanel()`) shows current order and highlights the current turn with a pulsing border. Combat uses D&D-style initiative. The initiative value is available in `InitiativeEntry.Initiative` but not displayed.

**Gap**
Gold Box showed clear turn ordering with visible initiative numbers. Players could plan multi-round tactics knowing who would act when.

**Implementation Specification**
- Files to modify: `pkg/wasmui/combat_screen.go`
- In `drawInitiativeList()`, add initiative score display:
  ```go
  // Show initiative value (right-aligned in panel)
  initText := fmt.Sprintf("%d", entry.Initiative)
  initX := panelX + charPanelWidth - 35
  drawColoredText(screen, initText, initX, y+2, ColorStatLabel)
  ```
- Add "NEXT" indicator on the entry following current turn:
  ```go
  if i == currentTurnIndex+1 || (currentTurnIndex == len(entries)-1 && i == 0) {
      drawColoredText(screen, "→", panelX+5, y+2, ColorGold)
  }
  ```
- Highlight current turn entry with gold background tint

**Success Criteria**
- [x] Initiative values displayed in tracker
- [x] Next combatant clearly indicated with arrow or "NEXT"
- [x] Current turn has distinct visual highlight
- [x] Turn order readable at a glance

---

#### 12. Wire Immunities Display to Combat UI

**Priority:** Medium
**Complexity:** Small
**Depends on:** None

**Current State**
Effect immunities are implemented in `pkg/game/effectimmunity.go` with ImmunityManager tracking permanent and temporary immunities. `PlayerState.Immunities []string` exists in `types_game.go:42`. The server can populate this field, but the UI never displays it. Character panel in exploration shows effects but not immunities.

**Gap**
Gold Box showed when characters were immune to effects (e.g., "Immune to Sleep"). Players could make informed spell choices knowing which enemies were immune.

**Implementation Specification**
- Files to modify: `pkg/wasmui/exploration.go` (character panel), `pkg/wasmui/combat_screen.go` (initiative panel)
- In `drawCharacterPanel()` after effects section (~line 600+):
  ```go
  // Draw immunities if present
  if len(player.Immunities) > 0 {
      drawColoredText(screen, "Immunities:", panelX+10, cursorY, ColorGold)
      cursorY += 15
      for _, imm := range player.Immunities {
          drawColoredText(screen, "• "+imm, panelX+15, cursorY, ColorEffectBuff)
          cursorY += 12
      }
  }
  ```
- Ensure server populates `Immunities` from `ImmunityManager.GetImmunities()`
- In initiative panel, show immunity icons (small shield) for enemies with immunities

**Success Criteria**
- [x] Character panel shows immunity list when present
- [x] Immunities displayed in green (buff color)
- [x] Server correctly populates immunity data
- [x] Enemies with immunities have visual indicator

---

#### 13. Add AI Behavior Type Display

**Priority:** Medium
**Complexity:** Small
**Depends on:** None

**Current State**
NPC AI behavior trees are implemented in `pkg/game/ai_behaviors.go` with Aggressive, Guard, Patrol, Coward archetypes. NPCs have behavior types, and `InitiativeEntry.BehaviorType` field exists in `types_game.go:97`. The UI doesn't indicate NPC behavior.

**Gap**
Gold Box showed NPC behavior through animations and text (fleeing, defending, etc.). Players could predict NPC actions and plan accordingly. A cowardly goblin behaves differently than an aggressive orc.

**Implementation Specification**
- Files to modify: `pkg/wasmui/combat_screen.go`
- In `drawInitiativeEntry()`, show behavior indicator for NPCs:
  ```go
  if !entry.IsPlayer && entry.BehaviorType != "" {
      behaviorIcons := map[string]string{
          "aggressive": "!",  // Will attack nearest
          "guard":      "G",  // Defends position
          "patrol":     "P",  // Moves in pattern
          "coward":     "~",  // Flees when hurt
      }
      if icon, ok := behaviorIcons[entry.BehaviorType]; ok {
          drawColoredText(screen, icon, panelX+10, y+2, ColorEffectControl)
      }
  }
  ```
- Log behavior when NPC acts: "Goblin (Coward) flees!" in message log

**Success Criteria**
- [x] NPC behavior type indicated in initiative tracker
- [x] Behavior-specific icons distinguishable and meaningful
- [x] Behavior affects displayed actions in message log
- [x] Aggressive/Guard/Patrol/Coward have distinct indicators

---

### Group: Exploration Screen

#### 14. Add Encounter/NPC Portrait Display

**Priority:** High
**Complexity:** Medium
**Depends on:** None

**Current State**
The `EncounterOverlay` struct exists in `types_ui.go:122-130` with `PortraitPath string` field. `drawEncounterOverlay()` in overlay rendering draws text but doesn't display portraits. Adventure NPCs have portrait paths via `AdventureNPCPath()` in `asset_loader.go`. Portrait assets exist in `web/static/assets/sprites/characters/`.

**Gap**
Gold Box encounters featured NPC portraits prominently—the mysterious stranger, the quest giver, the tavern keeper. This visual storytelling element makes encounters memorable and immersive.

**Implementation Specification**
- Files to modify: `pkg/wasmui/overlays.go` or wherever `drawEncounterOverlay()` is implemented
- In encounter overlay drawing, before text rendering:
  ```go
  if e.PortraitPath != "" {
      portraitX := overlayX + 20
      portraitY := overlayY + 40
      portraitW, portraitH := 96, 128
      
      // Draw portrait with fallback
      DrawSpriteWithFallback(screen, e.PortraitPath, 
          portraitX, portraitY, portraitW, portraitH,
          color.RGBA{R: 60, G: 50, B: 80, A: 255})
      
      // Gold Box style portrait border
      drawRectOutline(screen, portraitX-2, portraitY-2, portraitW+4, portraitH+4, ColorGold)
      
      // Offset text to right of portrait
      textX = portraitX + portraitW + 15
  }
  ```
- Ensure encounter data includes portrait path from adventure configuration

**Success Criteria**
- [x] NPC encounters display portraits when path provided
- [x] Portrait has decorative Gold Box border
- [x] Text flows beside portrait, not overlapping
- [x] Graceful fallback when portrait not found

---

#### 15. Implement Minimap Overlay

**Priority:** Medium
**Complexity:** Medium
**Depends on:** None

**Current State**
`OverlayState.ShowMinimap` exists in `types_ui.go:118`. The M key toggles minimap in `handleExplorationOverlayKeys()` (exploration.go:118-124). However, the actual minimap drawing may not be implemented or visible.

**Gap**
Gold Box games had area maps accessible via menu. Tactical awareness of dungeon layout is important for navigation and planning.

**Implementation Specification**
- Files to modify: `pkg/wasmui/exploration.go`, `pkg/wasmui/rpc_methods.go`
- Verify/implement `drawMinimapOverlay()`:
  ```go
  func (g *Game) drawMinimapOverlay(screen *ebiten.Image) {
      if !g.overlays.ShowMinimap { return }
      
      overlayW, overlayH := 200, 200
      overlayX := (g.screenWidth - overlayW) / 2
      overlayY := (g.screenHeight - overlayH) / 2
      
      // Background
      drawRect(screen, overlayX, overlayY, overlayW, overlayH, color.RGBA{R: 20, G: 20, B: 30, A: 240})
      drawRectOutline(screen, overlayX, overlayY, overlayW, overlayH, ColorGold)
      drawColoredText(screen, "MAP", overlayX+overlayW/2-12, overlayY+5, ColorGold)
      
      // Draw explored tiles (4x4 pixels each)
      tilePixelSize := 4
      // ... render explored map data
      
      // Player position (green dot)
      // ... render player marker
  }
  ```
- Add `GetExploredMap` RPC if needed to retrieve explored tile data
- Call from `drawExplorationScreen()` after other overlays

**Success Criteria**
- [x] M key toggles minimap overlay
- [x] Explored areas shown in overhead view
- [x] Player position clearly marked (green dot)
- [x] Doors and walls distinguishable by color
- [x] Overlay dismissible with M or Escape

---

#### 16. Add Touch-Friendly Directional Controls

**Priority:** Medium
**Complexity:** Medium
**Depends on:** None

**Current State**
Touch input exists via swipe gestures in `handleExplorationMovement()`. The `touchState` tracks swipes and taps. A directional pad (d-pad) area is reserved in exploration command menu (`dpadClearance=108px` per stored memory). However, tap-based directional controls for exploration may not be fully implemented.

**Gap**
Mobile players need clear touch targets for movement beyond swipe gestures. A visible d-pad provides intuitive controls.

**Implementation Specification**
- Files to modify: `pkg/wasmui/exploration.go`, `pkg/wasmui/input.go`
- Draw visible d-pad in bottom-left corner:
  ```go
  func (g *Game) drawDPad(screen *ebiten.Image, x, y int) {
      const size = 40
      const gap = 5
      
      // Up arrow
      drawDPadButton(screen, x+size+gap, y, size, "↑", g.dpadPressed == "north")
      // Down arrow
      drawDPadButton(screen, x+size+gap, y+2*(size+gap), size, "↓", g.dpadPressed == "south")
      // Left arrow
      drawDPadButton(screen, x, y+size+gap, size, "←", g.dpadPressed == "west")
      // Right arrow
      drawDPadButton(screen, x+2*(size+gap), y+size+gap, size, "→", g.dpadPressed == "east")
      // Center (turn left/right)
      drawDPadButton(screen, x+size+gap, y+size+gap, size, "○", false)
  }
  ```
- Handle taps on d-pad buttons to trigger movement
- Add turn left/right buttons or gestures

**Success Criteria**
- [x] Visible d-pad renders in exploration mode
- [x] Tap on directional button moves in that direction
- [x] Turn left/right accessible via tap or gesture
- [x] D-pad responsive without blocking other UI
- [x] Works on both portrait and landscape orientations

---

### Group: Character & Party Display

#### 17. Draw Equipment Slots with Item Sprites

**Priority:** High
**Complexity:** Medium
**Depends on:** None

**Current State**
`drawEquipmentSlots()` in `overlays.go:83-106` renders equipment as positioned slots with hover detection. Equipment slot layout is defined in `equipmentSlotLayout` map. Item sprites exist in `web/static/assets/sprites/items/`. `ItemIconPath()` function generates correct paths.

**Gap**
Gold Box showed equipment with visual icons in a paper-doll style layout. Current implementation may show text names or placeholder rectangles instead of item sprites.

**Implementation Specification**
- Files to modify: `pkg/wasmui/overlays.go`
- Update `drawEquipmentSlots()` to use item sprites:
  ```go
  const slotSize = 36
  for slotName, pos := range equipmentSlotLayout {
      slotX := baseX + pos.x
      slotY := baseY + pos.y
      
      // Draw slot background
      drawRect(screen, slotX, slotY, slotSize, slotSize, color.RGBA{R: 40, G: 38, B: 50, A: 255})
      drawRectOutline(screen, slotX, slotY, slotSize, slotSize, ColorPanelBorder)
      
      // Draw item sprite if equipped
      if item := getEquippedItem(slotName); item != nil {
          spritePath := ItemIconPath(item.Type, item.Name)
          DrawSpriteWithFallback(screen, spritePath, slotX+2, slotY+2, slotSize-4, slotSize-4,
              color.RGBA{R: 100, G: 80, B: 60, A: 255})
      }
      
      // Hover tooltip
      if g.hoveredEquipSlot == slotName && item != nil {
          drawTooltip(screen, item.Name, slotX, slotY-20)
      }
  }
  ```
- Paper-doll layout positions:
  - Head: center-top
  - Neck: below head
  - Chest: center
  - Hands: left of chest
  - Rings: right of chest
  - Legs: below chest
  - Feet: bottom center
  - WeaponMain: far left
  - WeaponOff: far right

**Success Criteria**
- [x] Equipment shown in paper-doll visual layout
- [x] Item sprites display in appropriate slots
- [x] Empty slots clearly indicated (border only)
- [x] Hover shows item name tooltip
- [x] Item icons match equipped item types

---

#### 18. Add Character Portrait to Character Panel

**Priority:** Medium
**Complexity:** Small
**Depends on:** None

**Current State**
Character panel in exploration shows stats as text. `CharacterPortraitPath()` in `asset_loader.go:182-195` generates correct paths based on class, race, and gender. `PreloadCharacterSprites()` exists. Portrait assets exist in `web/static/assets/sprites/characters/portraits/`. No portrait renders in the character panel.

**Gap**
Gold Box character sheets featured character portraits prominently. The portrait provided visual identity for the character.

**Implementation Specification**
- Files to modify: `pkg/wasmui/exploration.go`
- In `drawCharacterPanel()` at top of panel after title:
  ```go
  if player != nil {
      // Get appearance data for correct portrait
      race := "human"
      gender := "male"
      if player.Appearance != nil {
          if player.Appearance.GenderExpression != "" {
              gender = strings.ToLower(player.Appearance.GenderExpression)
          }
          // Race would come from appearance if tracked
      }
      
      portraitPath := CharacterPortraitPath(player.Class, race, gender)
      portraitX := panelX + (charPanelWidth - 96) / 2
      portraitY := panelY + 35
      
      DrawSpriteWithFallback(screen, portraitPath, portraitX, portraitY, 96, 96,
          color.RGBA{R: 60, G: 50, B: 80, A: 255})
      drawRectOutline(screen, portraitX-2, portraitY-2, 100, 100, ColorGold)
      
      cursorY = portraitY + 110 // Start stats below portrait
  }
  ```
- Adjust stat layout to flow below portrait

**Success Criteria**
- [x] Character portrait displays at top of character panel
- [x] Portrait has decorative Gold Box border
- [x] Falls back gracefully if sprite not loaded
- [x] Stats render below portrait without overlap

---

#### 19. Display Attribute Modifiers

**Priority:** Medium
**Complexity:** Small
**Depends on:** None

**Current State**
Character panel shows raw attribute scores (STR: 16, DEX: 14, etc.). `AttributeModifier()` function exists in `types_game.go:152-154` calculating D&D-style modifier ((score-10)/2). Modifiers aren't displayed in the UI.

**Gap**
Gold Box showed both score and modifier. Players need to see that 16 STR = +3 modifier for quick tactical decisions.

**Implementation Specification**
- Files to modify: `pkg/wasmui/exploration.go`
- In attribute display section of `drawCharacterPanel()`:
  ```go
  attrs := []struct{name string; val int}{
      {"STR", player.Attributes.Strength},
      {"DEX", player.Attributes.Dexterity},
      {"CON", player.Attributes.Constitution},
      {"INT", player.Attributes.Intelligence},
      {"WIS", player.Attributes.Wisdom},
      {"CHA", player.Attributes.Charisma},
  }
  for i, attr := range attrs {
      mod := AttributeModifier(attr.val)
      modStr := fmt.Sprintf("%+d", mod) // "+3" or "-1"
      modColor := ColorStatValue
      if mod > 0 {
          modColor = ColorEffectBuff // Green for positive
      } else if mod < 0 {
          modColor = ColorEffectDebuff // Red for negative
      }
      
      text := fmt.Sprintf("%s: %d", attr.name, attr.val)
      drawColoredText(screen, text, panelX+10, cursorY, ColorStatValue)
      drawColoredText(screen, modStr, panelX+70, cursorY, modColor)
      cursorY += 15
  }
  ```

**Success Criteria**
- [x] Attribute modifiers shown next to scores
- [x] Positive modifiers in green with "+" prefix
- [x] Negative modifiers in red
- [x] Zero modifiers in neutral color
- [x] Format matches Gold Box style

---

#### 20. Show Spell Slots / Spell Preparation

**Priority:** Medium
**Complexity:** Medium
**Depends on:** None

**Current State**
`PlayerState.SpellSlots` and `PlayerState.UsedSlots` fields exist in `types_game.go:43-44` as `map[int]int`. The spellbook overlay lists spells by level. D&D-style spell slots (spells per day) aren't tracked or displayed in the UI. Backend spell system in `pkg/game/spell_manager.go` may not enforce slot limits.

**Gap**
Gold Box had memorized spell tracking. Mages prepared spells and had limited casts per day. This resource management was core to tactical gameplay.

**Implementation Specification**
- Files to modify: `pkg/wasmui/overlays.go`, `pkg/server/handlers.go`, `pkg/game/spell_manager.go`
- Server side: Track spell slots per level, decrement on cast, restore on rest
- Client side: In spellbook header, show slots per level:
  ```go
  func (g *Game) drawSpellSlotSummary(screen *ebiten.Image, x, y int) {
      g.mu.RLock()
      slots := g.player.SpellSlots
      used := g.player.UsedSlots
      g.mu.RUnlock()
      
      drawColoredText(screen, "Spell Slots:", x, y, ColorGold)
      slotY := y + 15
      for level := 1; level <= 9; level++ {
          total := slots[level]
          if total == 0 { continue }
          remaining := total - used[level]
          text := fmt.Sprintf("Lv%d: %d/%d", level, remaining, total)
          color := ColorStatValue
          if remaining == 0 {
              color = ColorAPDepleted
          }
          drawColoredText(screen, text, x, slotY, color)
          slotY += 12
      }
  }
  ```
- Gray out spells where slot exhausted
- Add "Rest" action to restore slots (via exploration command menu)

**Success Criteria**
- [x] Spell slots displayed per level in spellbook
- [x] Used/available clearly shown (e.g., "3/4")
- [x] Depleted levels shown in red
- [x] Cannot cast when slots exhausted (server enforced)
- [x] Rest action restores spell slots

---

### Group: Game System Wiring

#### 21. Surface Faction Relations in UI

**Priority:** High
**Complexity:** Medium
**Depends on:** None

**Current State**
Faction diplomacy fully implemented in `pkg/game/faction_relations.go` with DiplomacyManager, diplomatic states (War, Hostile, Tense, Neutral, Friendly, Allied), and diplomatic actions (declare war, offer peace, propose alliance, etc.). RPC methods exist in `rpc_methods.go:390-469` for all faction operations. F key toggles faction panel per `handleExplorationOverlayKeys()` (exploration.go:107-116). However, no faction data is displayed.

**Gap**
Gold Box games had faction mechanics affecting dialogue options and combat. Players have no visibility into faction standings despite full backend support.

**Implementation Specification**
- Files to modify: `pkg/wasmui/overlays.go`, `pkg/wasmui/exploration.go`
- Implement faction panel drawing when `g.guildTab == 2` (Factions tab):
  ```go
  func (g *Game) drawFactionPanel(screen *ebiten.Image) {
      overlayX, overlayY := 100, 80
      overlayW, overlayH := 440, 400
      
      drawRect(screen, overlayX, overlayY, overlayW, overlayH, ColorPanelBG)
      drawBoldPanelBorder(screen, overlayX, overlayY, overlayW, overlayH)
      drawColoredText(screen, "FACTION RELATIONS", overlayX+overlayW/2-60, overlayY+10, ColorGold)
      
      // List known factions with relation state
      factionY := overlayY + 40
      for _, relation := range g.factionRelations {
          stateColor := getFactionStateColor(relation.State)
          drawColoredText(screen, relation.FactionName, overlayX+20, factionY, ColorStatValue)
          drawColoredText(screen, string(relation.State), overlayX+200, factionY, stateColor)
          factionY += 20
      }
  }
  
  func getFactionStateColor(state string) color.RGBA {
      switch state {
      case "war": return ColorEnemyName // Red
      case "hostile": return color.RGBA{R: 255, G: 100, B: 50, A: 255} // Orange-red
      case "tense": return ColorEffectControl // Yellow
      case "neutral": return ColorStatValue // White
      case "friendly": return ColorEffectBuff // Green
      case "allied": return ColorGold // Gold
      default: return ColorStatLabel
      }
  }
  ```
- Add `loadFactionData()` RPC call when faction panel opened
- Show faction reputation changes in message log

**Success Criteria**
- [x] F key opens faction relations panel
- [x] Faction standings clearly displayed with names
- [x] Diplomatic states color-coded (War=red to Allied=gold)
- [x] Reputation changes logged in message log
- [x] Panel dismissible with Escape

---

#### 22. Complete Guild Panel Implementation

**Priority:** Medium
**Complexity:** Medium
**Depends on:** None

**Current State**
Guild system implemented in `pkg/game/guild.go` with membership, ranks, treasury, perks. RPC methods exist in `rpc_methods.go:308-388` for all guild operations. G key toggles guild panel per `handleExplorationOverlayKeys()` (exploration.go:98-105). `loadGuildData()` is called but panel rendering may be incomplete.

**Gap**
Guild mechanics exist but may not be fully visible to players.

**Implementation Specification**
- Files to modify: `pkg/wasmui/overlays.go`, `pkg/wasmui/exploration.go`
- Implement complete `drawGuildPanel()`:
  ```go
  func (g *Game) drawGuildPanel(screen *ebiten.Image) {
      if !g.overlays.ShowGuildPanel { return }
      
      // Panel dimensions and background
      overlayX, overlayY := 100, 80
      overlayW, overlayH := 440, 400
      
      drawRect(screen, overlayX, overlayY, overlayW, overlayH, ColorPanelBG)
      drawBoldPanelBorder(screen, overlayX, overlayY, overlayW, overlayH)
      
      // Tab buttons: Guild Info | Members | Factions
      g.drawGuildTabs(screen, overlayX, overlayY)
      
      switch g.guildTab {
      case 0: g.drawGuildInfoTab(screen, overlayX, overlayY+40, overlayW, overlayH-40)
      case 1: g.drawGuildMembersTab(screen, overlayX, overlayY+40, overlayW, overlayH-40)
      case 2: g.drawFactionPanel(screen) // Reuse faction panel
      }
  }
  ```
- Guild Info tab: Guild name, level, treasury balance, available perks
- Members tab: Scrollable member list with ranks
- Wire `loadGuildData()` to populate `g.guildData`

**Success Criteria**
- [x] G key opens guild panel
- [x] Guild info displays correctly (name, level, treasury)
- [x] Member list shown with ranks
- [x] Available perks visible
- [x] Tab switching works

---

#### 23. Emit Missing Event Types

**Priority:** Medium
**Complexity:** Medium
**Depends on:** None

**Current State**
Per `GAPS.md`, eight event type constants are defined in `pkg/game/constants.go:169-178`, but only three are actually emitted (EventLevelUp, EventDeath, EventMovement). Five event types are defined but never broadcast: EventDamage, EventItemPickup, EventItemDrop, EventSpellCast, EventQuestUpdate.

**Gap**
WebSocket clients subscribed to these event types never receive notifications. Real-time feedback for item pickups, spell casts, and quest updates is missing.

**Implementation Specification**
- Files to modify: `pkg/server/handlers.go`, `pkg/server/handlers_equipment.go`, `pkg/server/handlers_quest.go`, `pkg/server/combat.go`
- Add event emissions:
  1. EventDamage: In attack handlers after damage dealt
     ```go
     s.eventSys.Emit(game.NewGameEvent(game.EventDamage, attackerID, targetID, 
         map[string]interface{}{"damage": damage, "type": damageType}))
     ```
  2. EventItemPickup: In handleEquipItem and inventory add
  3. EventItemDrop: In handleUnequipItem and inventory remove
  4. EventSpellCast: In handleCastSpell after spell execution
  5. EventQuestUpdate: In handleStartQuest, handleCompleteQuest, handleFailQuest

**Success Criteria**
- [x] EventDamage emitted on every damage dealt
- [x] EventItemPickup emitted when items added to inventory
- [x] EventItemDrop emitted when items removed
- [x] EventSpellCast emitted on spell execution
- [x] EventQuestUpdate emitted on quest state changes
- [x] WebSocket clients receive all event types

---

### Group: Animation & Visual Feedback

#### 24. Add Damage Number Popups

**Priority:** High
**Complexity:** Medium
**Depends on:** None

**Current State**
`DamagePopup` struct exists in `types_ui.go:344-385` with Position, Amount, IsHeal, IsCrit, StartTime, Duration, and animation methods (Progress, Alpha, YOffset). `damagePopups []DamagePopup` likely exists in Game struct. `drawDamagePopups()` is called from `drawCombatGrid()` (combat_screen.go:261). Need to verify popups are created when damage occurs.

**Gap**
Verify damage popups are being created and displayed. If not, implement the creation logic.

**Implementation Specification**
- Files to verify/modify: `pkg/wasmui/combat_screen.go`, `pkg/wasmui/game.go`
- Verify `addDamagePopup()` is called when damage is dealt:
  ```go
  func (g *Game) addDamagePopup(entityID string, x, y, amount int, isHeal, isCrit bool) {
      popup := DamagePopup{
          EntityID:  entityID,
          X:         x,
          Y:         y,
          Amount:    amount,
          IsHeal:    isHeal,
          IsCrit:    isCrit,
          StartTime: time.Now(),
          Duration:  800 * time.Millisecond,
      }
      g.mu.Lock()
      g.damagePopups = append(g.damagePopups, popup)
      g.mu.Unlock()
  }
  ```
- Verify `drawDamagePopups()` renders correctly:
  ```go
  func (g *Game) drawDamagePopups(screen *ebiten.Image) {
      g.mu.RLock()
      popups := g.damagePopups
      g.mu.RUnlock()
      
      for _, popup := range popups {
          if !popup.IsActive() { continue }
          
          y := popup.Y - popup.YOffset()
          alpha := popup.Alpha()
          
          text := fmt.Sprintf("%d", popup.Amount)
          textColor := ColorEnemyName // Red for damage
          if popup.IsHeal {
              textColor = ColorEffectBuff // Green for healing
          }
          textColor.A = uint8(float32(textColor.A) * alpha)
          
          if popup.IsCrit {
              text = "!" + text + "!"
              // Draw slightly larger or with gold border
          }
          
          drawColoredText(screen, text, popup.X, y, textColor)
      }
  }
  ```
- Wire damage callback to create popups

**Success Criteria**
- [x] Damage numbers appear on hit
- [x] Numbers float upward and fade out
- [x] Healing shown in green, damage in red
- [x] Critical hits emphasized (larger or bordered)
- [x] Numbers visible but not intrusive

---

#### 25. Implement Spell Cast Animations

**Priority:** Medium
**Complexity:** Medium
**Depends on:** None

**Current State**
`SpellEffect` struct exists in `types_ui.go:387-433` with SpellID, SpellSchool, TargetPos, animation fields, and methods (IsActive, GetFrame, GetRadius, GetAlpha). `SpellSchoolColor()` returns appropriate colors. `drawSpellEffects()` is called from `drawCombatGrid()`. Spell effect sprites may exist in `web/static/assets/sprites/effects/spells/`.

**Gap**
Spell effects should use actual spell sprites when available, falling back to procedural expanding circles.

**Implementation Specification**
- Files to modify: `pkg/wasmui/combat_screen.go`, `pkg/wasmui/asset_loader.go`
- Verify `SpellEffectPath()` exists in asset_loader.go (line 245-248)
- In `drawSpellEffects()`, check for sprite before procedural:
  ```go
  func (g *Game) drawSpellEffects(screen *ebiten.Image, offsetX, offsetY, tileSize int) {
      g.mu.RLock()
      effects := g.spellEffects
      g.mu.RUnlock()
      
      for _, effect := range effects {
          if !effect.IsActive() { continue }
          
          x := offsetX + effect.TargetPos.X*tileSize + tileSize/2
          y := offsetY + effect.TargetPos.Y*tileSize + tileSize/2
          
          spritePath := SpellEffectPath(effect.SpellID)
          if spriteCache.IsCached(spritePath) {
              // Draw sprite animation
              frame := effect.GetFrame()
              alpha := effect.GetAlpha()
              DrawSpriteWithAlpha(screen, spritePath, x-tileSize/2, y-tileSize/2, tileSize, tileSize, alpha)
          } else {
              // Trigger async load
              spriteCache.Get(spritePath)
              // Procedural fallback: expanding circle
              g.drawProceduralSpellEffect(screen, effect, x, y)
          }
      }
  }
  ```

**Success Criteria**
- [x] Spell effects use sprites when available
- [x] Graceful fallback to procedural effects
- [x] Effects match spell school colors
- [x] Animations play at correct speed

---

#### 26. Enhance Movement Animation Feedback

**Priority:** Medium
**Complexity:** Small
**Depends on:** None

**Current State**
Movement transition exists with `moveTransitionStart`, `moveTransitionDir`, `moveTransitionDur` fields. `calculateMoveTransitionOffset()` returns viewport offset. `getMoveTransitionFlashAlpha()` returns flash intensity (peaks at 50 alpha). `drawMoveDirectionIndicator()` draws arrow during transition. Duration is 50ms for instant step feel.

**Gap**
Current transitions are subtle. Gold Box had distinct step-by-step movement feel. Could be more noticeable while maintaining instant feel.

**Implementation Specification**
- Files to modify: `pkg/wasmui/exploration.go`
- Increase flash visibility:
  ```go
  // In getMoveTransitionFlashAlpha(), change peak alpha from 50 to 70
  if progress < 0.25 {
      alpha = progress * 4 * 70 // Was 50
  } else {
      alpha = (1.0 - (progress-0.25)/0.75) * 70
  }
  ```
- Add footstep-style visual element:
  ```go
  // In drawMoveDirectionIndicator(), add brief "step" marker
  func (g *Game) drawMoveDirectionIndicator(screen *ebiten.Image, cx, cy int) {
      // Existing arrow code...
      
      // Add brief footstep marker at bottom of viewport
      stepColor := color.RGBA{R: 150, G: 150, B: 180, A: 150}
      drawRect(screen, cx-3, cy+5, 6, 3, stepColor)
  }
  ```
- Keep transitions at 50ms for Gold Box instant feel

**Success Criteria**
- [x] Movement feedback more noticeable
- [x] Still maintains instant step feel (50ms)
- [x] Direction of movement clear
- [x] Flash doesn't obscure viewport content

---

#### 27. Implement Turn Change Visual Effect

**Priority:** Low
**Complexity:** Small
**Depends on:** None

**Current State**
Turn changes update `combat.CurrentTurn`. Pulsing border shows current turn entity. `drawTurnChangeFlash()` is called from `drawCombatGrid()` (combat_screen.go:264). Need to verify turn change detection and flash implementation.

**Gap**
Gold Box had clear turn transitions. Players should notice immediately when it's their turn.

**Implementation Specification**
- Files to modify: `pkg/wasmui/combat_screen.go`
- Add turn change tracking:
  ```go
  // In Game struct
  lastCurrentTurn string
  turnChangeFlash time.Time
  
  // In combat update
  if combat.CurrentTurn != g.lastCurrentTurn {
      g.lastCurrentTurn = combat.CurrentTurn
      if combat.IsPlayerTurn {
          g.addLogMessage("-- YOUR TURN --", MessageSystem)
          g.turnChangeFlash = time.Now()
      }
  }
  ```
- Implement `drawTurnChangeFlash()`:
  ```go
  func (g *Game) drawTurnChangeFlash(screen *ebiten.Image) {
      if time.Since(g.turnChangeFlash) > 500*time.Millisecond { return }
      
      progress := float32(time.Since(g.turnChangeFlash)) / float32(500*time.Millisecond)
      alpha := uint8((1.0 - progress) * 40)
      
      // Gold border pulse around combat grid
      flashColor := color.RGBA{R: 191, G: 165, B: 74, A: alpha}
      gridW := g.screenWidth - charPanelWidth
      gridH := g.screenHeight - logPanelHeight - actionPanelHeight
      drawRectOutline(screen, 2, 2, gridW-4, gridH-4, flashColor)
      drawRectOutline(screen, 4, 4, gridW-8, gridH-8, flashColor)
  }
  ```

**Success Criteria**
- [x] "YOUR TURN" message in log on player turn
- [x] Visual flash (gold border pulse) indicates turn change
- [x] Easy to notice even when distracted
- [x] Doesn't trigger on enemy turns

---

### Group: Message Log

#### 28. Route All Game Events to Message Log

**Priority:** High
**Complexity:** Medium
**Depends on:** None

**Current State**
Message log receives some events (combat messages, spell casting). MessageType enum includes MessageLoot, MessageQuest, MessageLevelUp, MessageInteract for appropriate coloring. Many events are silent: item pickups, door opening, trap triggers, quest updates, exploration events.

**Gap**
Gold Box message log was the primary feedback channel. ALL game events appeared as text. The log was how players understood what happened.

**Implementation Specification**
- Files to modify: `pkg/wasmui/exploration.go`, `pkg/wasmui/combat_screen.go`, `pkg/wasmui/overlays.go`
- Audit all RPC callbacks and add log messages:
  - **Item pickup**: `g.addLogMessage(fmt.Sprintf("Found: %s", item.Name), MessageLoot)`
  - **Item equip/unequip**: Already exists in `overlays.go:127,134`
  - **Door open**: `g.addLogMessage("The door creaks open...", MessageInteract)`
  - **Trap trigger**: `g.addLogMessage("A poison dart trap triggers!", MessageWarning)`
  - **Quest start**: `g.addLogMessage(fmt.Sprintf("Quest Started: %s", quest.Title), MessageQuest)`
  - **Quest update**: `g.addLogMessage(fmt.Sprintf("Quest Updated: %s (%d/%d)", ...), MessageQuest)`
  - **Quest complete**: `g.addLogMessage(fmt.Sprintf("QUEST COMPLETE: %s", quest.Title), MessageQuest)`
  - **Level up**: `g.addLogMessage(fmt.Sprintf("LEVEL UP! %s is now level %d!", name, level), MessageLevelUp)`
  - **Gold found**: `g.addLogMessage(fmt.Sprintf("Found %d gold pieces", amount), MessageLoot)`
  - **Spell learned**: `g.addLogMessage(fmt.Sprintf("Learned spell: %s", spell.Name), MessageSystem)`
- Ensure no player action is silent

**Success Criteria**
- [x] Every item pickup logged
- [x] Every interaction (doors, chests, levers) logged
- [x] Quest progress logged
- [x] Level ups logged with emphasis
- [x] Gold/treasure finds logged
- [x] All message types use correct colors

---

#### 29. Implement Message Log Scrolling

**Priority:** Medium
**Complexity:** Small
**Depends on:** None

**Current State**
Message log shows recent messages via `g.logMessages` slice. `maxLogMessages` (likely 8-10) limits visible count. No scrollback mechanism exists to review history.

**Gap**
Gold Box let players review combat history. Current log is ephemeral—important information scrolls away.

**Implementation Specification**
- Files to modify: `pkg/wasmui/exploration.go`, `pkg/wasmui/combat_screen.go`, `pkg/wasmui/game.go`
- Increase message retention:
  ```go
  const maxLogMessages = 100 // Was ~8-10
  ```
- Add scroll state:
  ```go
  // In Game struct
  logScrollOffset int
  ```
- Add scroll controls:
  ```go
  // In update functions
  if ebiten.IsKeyPressed(ebiten.KeyPageUp) {
      g.logScrollOffset = min(g.logScrollOffset+1, len(g.logMessages)-visibleCount)
  }
  if ebiten.IsKeyPressed(ebiten.KeyPageDown) {
      g.logScrollOffset = max(0, g.logScrollOffset-1)
  }
  // Mouse wheel over log area
  ```
- Update `drawCombatLog()`:
  ```go
  startIdx := max(0, len(messages)-visibleCount-g.logScrollOffset)
  endIdx := min(len(messages), startIdx+visibleCount)
  for i, msg := range messages[startIdx:endIdx] {
      drawColoredText(screen, msg.Text, logX+5, logY+i*lineHeight, msg.Type.Color())
  }
  // Show scroll indicators
  if g.logScrollOffset > 0 {
      drawColoredText(screen, "↓ more", logX+logW-50, logY+logH-15, ColorStatLabel)
  }
  if startIdx > 0 {
      drawColoredText(screen, "↑ more", logX+logW-50, logY+5, ColorStatLabel)
  }
  ```
- Reset scroll to bottom on new message

**Success Criteria**
- [x] Message log stores 100+ messages
- [x] Page Up/Down scrolls log
- [x] Mouse wheel scrolls when over log area
- [x] Scroll position indicators visible when more content
- [x] Auto-scrolls to bottom on new message

---

#### 30. Add Combat Round/Turn Prefix to Log Messages

**Priority:** Low
**Complexity:** Small
**Depends on:** None

**Current State**
`LogMessage` has `CombatRound int` field (types_ui.go:47) but it's not displayed. Combat messages show without timing context.

**Gap**
Combat round/turn context helps players track sequence of events and understand timing.

**Implementation Specification**
- Files to modify: `pkg/wasmui/exploration.go`, `pkg/wasmui/combat_screen.go`
- In log message display during combat:
  ```go
  func (g *Game) drawCombatLogMessage(screen *ebiten.Image, msg LogMessage, x, y int) {
      prefix := ""
      if g.mode == ModeCombat && msg.CombatRound > 0 {
          prefix = fmt.Sprintf("[R%d] ", msg.CombatRound)
      }
      
      // Draw prefix in muted color
      if prefix != "" {
          drawColoredText(screen, prefix, x, y, ColorStatLabel)
          x += len(prefix) * 6 // Approximate character width
      }
      
      // Draw message in appropriate color
      drawColoredText(screen, msg.Text, x, y, msg.Type.Color())
  }
  ```
- Set `CombatRound` when creating messages during combat

**Success Criteria**
- [x] Combat messages show round number prefix
- [x] Round prefix in muted color (doesn't distract)
- [x] Helps track combat timeline
- [x] Non-combat messages show no prefix

---

### Group: UI Layout & Panels

#### 31. Standardize Gold Box-Style Panel Borders

**Priority:** High
**Complexity:** Small
**Depends on:** None

**Current State**
`drawBoldPanelBorder()` exists and is used throughout. Border colors defined in `types_ui.go:17-20` (ColorPanelBorder, ColorPanelBorderHi, ColorPanelShadow). BorderThickness = 2. Borders may vary slightly across different panels.

**Gap**
Gold Box had distinctive thick bright borders separating panels consistently. Need to ensure all panels use the same border style.

**Implementation Specification**
- Files to modify: `pkg/wasmui/exploration.go`, `pkg/wasmui/combat_screen.go`, `pkg/wasmui/overlays.go`
- Standardize `drawBoldPanelBorder()`:
  ```go
  func drawBoldPanelBorder(screen *ebiten.Image, x, y, w, h int) {
      // Outer bright border (3px thick)
      for i := 0; i < 3; i++ {
          drawRectOutline(screen, x+i, y+i, w-2*i, h-2*i, ColorPanelBorderHi)
      }
      // Inner darker border
      drawRectOutline(screen, x+3, y+3, w-6, h-6, ColorPanelBorder)
      // Shadow line (bottom and right)
      drawLine(screen, x+4, y+h-2, x+w-4, y+h-2, ColorPanelShadow)
      drawLine(screen, x+w-2, y+4, x+w-2, y+h-4, ColorPanelShadow)
  }
  ```
- Apply consistently to all panels:
  - Character panel
  - Combat log
  - Viewport frame
  - Initiative panel
  - Overlay panels (inventory, spellbook, quests, guild)
  - Action panel

**Success Criteria**
- [x] All panels have consistent bold borders
- [x] Borders match EGA color palette
- [x] Panels clearly visually separated
- [x] Border thickness uniform (3px outer, 1px inner)

---

#### 32. Standardize Command Menu Layout

**Priority:** Medium
**Complexity:** Small
**Depends on:** None

**Current State**
Action panel in exploration (`drawActionPanel()`) and combat (`drawCombatActionBar()`) have different layouts. Command definitions in `command_menu_defs.go`. `drawCommandMenu()` renders commands. Key hints may be inconsistent.

**Gap**
Gold Box had consistent command menu style across modes with highlighted key letters.

**Implementation Specification**
- Files to modify: `pkg/wasmui/exploration.go`, `pkg/wasmui/combat_screen.go`, `pkg/wasmui/command_menu.go`
- Create unified action button style:
  ```go
  func drawActionButton(screen *ebiten.Image, x, y, w, h int, label, key string, selected, available bool) {
      bgColor := color.RGBA{R: 50, G: 50, B: 70, A: 255}
      if selected {
          bgColor = color.RGBA{R: 70, G: 70, B: 100, A: 255}
      }
      if !available {
          bgColor = color.RGBA{R: 40, G: 40, B: 50, A: 255}
      }
      
      drawRect(screen, x, y, w, h, bgColor)
      drawRectOutline(screen, x, y, w, h, ColorPanelBorder)
      
      // Draw key in gold, rest in white
      keyText := fmt.Sprintf("[%s]", key)
      drawColoredText(screen, keyText, x+5, y+(h-12)/2, ColorGold)
      drawColoredText(screen, label, x+5+len(keyText)*6+5, y+(h-12)/2, 
          conditionalColor(available, ColorStatValue, ColorStatLabel))
  }
  ```
- Use for both exploration and combat command menus
- Ensure consistent button sizing (width based on available space, height 28-32px)

**Success Criteria**
- [x] Action buttons consistent across modes
- [x] Key hints clearly visible in gold
- [x] Selected state has distinct background
- [x] Disabled state visually muted
- [x] Touch targets adequate size

---

#### 33. Add Panel Title Headers

**Priority:** Low
**Complexity:** Small
**Depends on:** None

**Current State**
Some panels have titles (CHARACTER, INITIATIVE), others don't. Styling varies.

**Gap**
Gold Box panels had clear header bars with centered titles.

**Implementation Specification**
- Files to modify: `pkg/wasmui/exploration.go`, `pkg/wasmui/combat_screen.go`, `pkg/wasmui/overlays.go`
- Create `drawPanelHeader()`:
  ```go
  func drawPanelHeader(screen *ebiten.Image, x, y, w int, title string) {
      headerH := 25
      
      // Header bar background (slightly lighter than panel)
      drawRect(screen, x, y, w, headerH, color.RGBA{R: 40, G: 38, B: 55, A: 255})
      
      // Title centered in gold
      textW := len(title) * 6 // Approximate
      textX := x + (w-textW)/2
      drawColoredText(screen, title, textX, y+6, ColorGold)
      
      // Separator line below header
      drawLine(screen, x, y+headerH-1, x+w, y+headerH-1, ColorPanelBorder)
  }
  ```
- Apply to all major panels:
  - "CHARACTER" on character panel
  - "COMBAT LOG" on message log
  - "INITIATIVE" on initiative panel
  - "INVENTORY", "SPELLBOOK", "QUESTS", "GUILD" on overlays

**Success Criteria**
- [x] All panels have consistent header bars
- [x] Titles centered in gold
- [x] Headers visually distinguished from content
- [x] Separator line below headers

---

### Group: Asset Integration

#### 34. Verify Monster Sprite Loading

**Priority:** High
**Complexity:** Medium
**Depends on:** None

**Current State**
`MonsterSpritePath()` in `asset_loader.go:210-234` generates paths based on monster type with category classification. Monster sprites exist in `web/static/assets/sprites/monsters/` subdirectories (beasts, demons, dragons, humanoids, undead, magical). `drawSingleEnemyToken()` calls `DrawSpriteWithFallback()`.

**Gap**
Need to verify paths match actual file structure. If mismatched, enemies appear as red fallback squares with "E".

**Implementation Specification**
- Files to verify: `pkg/wasmui/asset_loader.go`, `web/static/assets/sprites/monsters/`
- Verify path format matches:
  ```
  Expected: monsters/humanoids/monster_goblin.png
  Actual files: Check ls output
  ```
- Common issues:
  - Case sensitivity (Goblin vs goblin)
  - Underscore vs dash (giant_rat vs giant-rat)
  - Category misclassification
- Update `MonsterSpritePath()` if needed:
  ```go
  func MonsterSpritePath(monsterType string) string {
      // Normalize: lowercase, replace spaces with underscores
      typeLower := strings.ToLower(strings.ReplaceAll(monsterType, " ", "_"))
      
      // Check actual directory structure and match
      // ...
  }
  ```
- Add logging when sprite not found for debugging

**Success Criteria**
- [x] Monster sprites display in combat
- [x] Paths resolve to actual asset files
- [x] Fallback only when sprite truly missing
- [x] Common monster types all have working sprites

---

#### 35. Load Terrain Sprites for Combat Grid

**Priority:** Medium
**Complexity:** Small
**Depends on:** None

**Current State**
`drawCombatFloor()` in `combat_screen.go:283-315` calls `DrawSpriteWithFallback()` for floor tiles using `TerrainTilePath("floor_stone", "dungeon")`. Terrain sprites exist in `terrain/dungeon/`, `terrain/outdoor/`.

**Gap**
Need to verify paths match actual files. Combat grid may show fallback colors instead of terrain sprites.

**Implementation Specification**
- Files to verify: `pkg/wasmui/combat_screen.go`, `pkg/wasmui/asset_loader.go`, `web/static/assets/sprites/terrain/`
- Verify `TerrainTilePath()` generates correct paths:
  ```go
  // Expected: terrain/dungeon/tile_floor_stone.png
  TerrainTilePath("floor_stone", "dungeon")
  ```
- Check actual file structure:
  ```bash
  ls web/static/assets/sprites/terrain/dungeon/
  ls web/static/assets/sprites/dungeon/  # Alternate location
  ```
- Update path generation if needed
- Add floor tile variety:
  ```go
  floorTiles := []string{
      TerrainTilePath("floor_stone", "dungeon"),
      TerrainTilePath("floor_stone_alt", "dungeon"),
      TerrainTilePath("floor_cobble", "dungeon"),
  }
  ```

**Success Criteria**
- [x] Terrain sprites display in combat grid
- [x] Floor tiles show variety (not uniform)
- [x] No placeholder rectangles for standard terrain
- [x] Different dungeon themes use appropriate tiles

---

#### 36. Implement UI Element Sprite Loading

**Priority:** Low
**Complexity:** Medium
**Depends on:** None

**Current State**
UI sprites exist in `web/static/assets/sprites/ui/`, `buttons/`, `panels/`, `icons/`. `UIElementPath()` likely exists. UI rendering uses procedural `drawRect()` and `drawColoredText()` instead of sprites.

**Gap**
UI could use sprite-based buttons, panels, and icons for more authentic Gold Box look.

**Implementation Specification**
- Files to modify: `pkg/wasmui/asset_loader.go`, `pkg/wasmui/overlays.go`
- Add UI sprite helpers:
  ```go
  func UIButtonPath(state string) string {
      return fmt.Sprintf("buttons/button_%s.png", state) // normal, hover, pressed
  }
  
  func UIPanelPath(style string) string {
      return fmt.Sprintf("panels/panel_%s.png", style)
  }
  
  func UIIconPath(name string) string {
      return fmt.Sprintf("icons/icon_%s.png", name)
  }
  ```
- Create hybrid rendering (sprite with fallback):
  ```go
  func drawUIButton(screen *ebiten.Image, x, y, w, h int, state string, fallbackColor color.RGBA) {
      path := UIButtonPath(state)
      if spriteCache.IsCached(path) {
          DrawSpriteScaled(screen, path, x, y, w, h)
      } else {
          spriteCache.Get(path) // Trigger load
          drawRect(screen, x, y, w, h, fallbackColor) // Procedural fallback
      }
  }
  ```
- Gradually replace procedural UI with sprites where available

**Success Criteria**
- [x] Buttons use sprite assets when available
- [x] Graceful fallback to procedural rendering
- [x] Panel frames use sprite assets when available
- [x] Icons use sprite assets when available

---

#### 37. Add Status Effect Icons

**Priority:** Medium
**Complexity:** Small
**Depends on:** None

**Current State**
Effect sprites likely exist in `web/static/assets/sprites/effects/status/`. Effects currently displayed as colored squares in `drawEffectIndicators()`. `StatusEffectIconPath()` may not exist.

**Gap**
Status effect icons would improve combat readability with recognizable symbols for each effect type.

**Implementation Specification**
- Files to modify: `pkg/wasmui/asset_loader.go`, `pkg/wasmui/combat_screen.go`
- Add path helper:
  ```go
  func StatusEffectIconPath(effectType string) string {
      typeLower := strings.ToLower(strings.ReplaceAll(effectType, " ", "_"))
      return fmt.Sprintf("effects/status/effect_status_%s.png", typeLower)
  }
  ```
- Update `drawEffectIndicators()` to use sprites:
  ```go
  for i, effect := range effects {
      if i >= maxIcons { break }
      
      iconPath := StatusEffectIconPath(effect.Type)
      ix := x + i*(iconSize+spacing)
      
      if spriteCache.IsCached(iconPath) {
          DrawSpriteWithFallback(screen, iconPath, ix, y, iconSize, iconSize, getEffectColor(effect.Type))
      } else {
          spriteCache.Get(iconPath)
          // Fallback to colored square
          drawRect(screen, ix, y, iconSize, iconSize, getEffectColor(effect.Type))
      }
  }
  ```

**Success Criteria**
- [x] Status effects use icon sprites when available
- [x] Icons clearly represent effect type (flame for burning, etc.)
- [x] Fallback to colored squares if no sprite
- [x] Icons scale appropriately (8x8 on tokens)

---

### Group: Quality Maintenance

These items are carried forward from quality maintenance concerns.

#### 38. Increase Test Coverage to 70%

**Priority:** Medium
**Complexity:** Large
**Depends on:** None

**Current State**
Test coverage is 65-96% depending on package. Some files in `pkg/wasmui/` have limited coverage due to WASM build constraints (tests run natively). `pkg/game/` and `pkg/server/` have good coverage.

**Gap**
Target coverage is ≥70% for all critical packages to ensure reliability.

**Implementation Specification**
- Run `make find-untested` to identify coverage gaps
- Run `make test-coverage` to see current state
- Focus on:
  - `pkg/game/effectbehavior.go` - effect tick processing
  - `pkg/server/handlers.go` - RPC handler edge cases
  - `pkg/game/combat_*.go` - combat system functions
- Add table-driven tests:
  ```go
  func TestEffectProcessing(t *testing.T) {
      tests := []struct {
          name     string
          effect   Effect
          expected int // expected health change
      }{
          {"healing tick", Effect{Type: EffectHealOverTime, Magnitude: 5}, 5},
          {"damage tick", Effect{Type: EffectDamageOverTime, Magnitude: 3}, -3},
          // ...
      }
      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              // ...
          })
      }
  }
  ```
- Use mocks for external dependencies (WebSocket, HTTP)

**Success Criteria**
- [ ] Overall coverage ≥70%
- [ ] All critical game logic paths covered
- [ ] No untested error handling paths
- [ ] Race detector passes on all tests

---

#### 39. Document RPC API Completely

**Priority:** Low
**Complexity:** Medium
**Depends on:** None

**Current State**
`pkg/README-RPC.md` exists with some method documentation. `rpc_methods.go` has 60+ methods. Many methods may be undocumented or have incomplete examples.

**Gap**
Developers need complete API reference for building clients or understanding server capabilities.

**Implementation Specification**
- Files to modify: `pkg/README-RPC.md`
- Document all methods in `rpc_methods.go`:
  - Method name
  - Parameters with types
  - Return value structure
  - Example request/response JSON
  - Error conditions
- Format:
  ```markdown
  ### castSpell
  
  Casts a spell on a target.
  
  **Parameters:**
  - `spell_id` (string, required): ID of the spell to cast
  - `target_id` (string, required): ID of the target entity
  - `position` (object, optional): Target position for area spells
  
  **Returns:**
  - `success` (boolean): Whether the spell was cast
  - `damage` (int): Damage dealt (for damage spells)
  - `healing` (int): Health restored (for healing spells)
  - `message` (string): Result description
  
  **Example:**
  Request: `{"jsonrpc":"2.0","method":"castSpell","params":{"spell_id":"fireball","target_id":"goblin_1"},"id":1}`
  Response: `{"jsonrpc":"2.0","result":{"success":true,"damage":18,"message":"Fireball hits Goblin for 18 fire damage!"},"id":1}`
  ```
- Generate from code comments where possible

**Success Criteria**
- [ ] All RPC methods documented
- [ ] Parameters and return types specified
- [ ] Example request/response for each method
- [ ] Error conditions described

---

#### 40. Reduce Cyclomatic Complexity in Large Functions

**Priority:** Low
**Complexity:** Medium
**Depends on:** None

**Current State**
Some handler functions in `pkg/server/handlers.go` may exceed 50 lines with multiple nested conditionals. Complex switch statements exist.

**Gap**
High complexity reduces maintainability and makes testing harder.

**Implementation Specification**
- Identify functions with >10 cyclomatic complexity using static analysis
- Extract helper functions for distinct logical blocks:
  ```go
  // Before
  func handleCombatAction(params) {
      // 80 lines with nested ifs
  }
  
  // After
  func handleCombatAction(params) {
      if err := validateCombatAction(params); err != nil {
          return err
      }
      target := resolveCombatTarget(params)
      result := executeCombatAction(target, params)
      return formatCombatResult(result)
  }
  ```
- Reduce nesting with early returns
- Target: no function >50 lines, no complexity >10

**Success Criteria**
- [ ] No function exceeds 50 lines
- [ ] Cyclomatic complexity ≤10 for all functions
- [ ] Existing tests still pass
- [ ] Code review finds no significant maintainability issues

---

## Implementation Order

A recommended sequencing of ALL items above, accounting for dependencies and risk:

1. **1. Fix Stun Effect Enforcement** — Critical bug blocking tactical gameplay
2. **2. Fix Root Effect Movement Restriction** — Critical bug, simple fix
3. **3. Fix WebSocket Race Conditions** — Critical stability issue
4. **4. Fix Healing Modifier Zero Initialization** — High-priority balance fix
5. **5. Fix Multiplicative Modifier Stacking** — High-priority balance fix
6. **6. Implement Resistance API** — Enables equipment progression
7. **9. Implement Rich Attack Roll Narration** — Defines Gold Box feedback style
8. **31. Standardize Gold Box-Style Panel Borders** — Visual foundation for all UI
9. **28. Route All Game Events to Message Log** — Core Gold Box interaction pattern
10. **7. Wire Morale System to Combat UI** — High-value backend surfacing
11. **8. Display Active Effects on Combat Tokens** — Important tactical feedback
12. **34. Verify Monster Sprite Loading** — Critical for visual fidelity
13. **14. Add Encounter/NPC Portrait Display** — Important narrative element
14. **24. Add Damage Number Popups** — High-impact visual feedback
15. **17. Draw Equipment Slots with Item Sprites** — Important inventory UX
16. **18. Add Character Portrait to Character Panel** — Visual polish
17. **21. Surface Faction Relations in UI** — Backend system visibility
18. **22. Complete Guild Panel Implementation** — Backend system visibility
19. **23. Emit Missing Event Types** — WebSocket functionality completeness
20. **10. Add Opportunity Attack Visual Indicators** — Verify existing implementation
21. **11. Implement Turn Order Prediction Display** — Tactical information
22. **12. Wire Immunities Display to Combat UI** — Combat information
23. **13. Add AI Behavior Type Display** — Combat information
24. **15. Implement Minimap Overlay** — Navigation aid
25. **16. Add Touch-Friendly Directional Controls** — Mobile UX
26. **19. Display Attribute Modifiers** — Character information
27. **20. Show Spell Slots / Spell Preparation** — Spellcasting depth
28. **25. Implement Spell Cast Animations** — Visual polish
29. **26. Enhance Movement Animation Feedback** — UX polish
30. **35. Load Terrain Sprites for Combat Grid** — Visual polish
31. **36. Implement UI Element Sprite Loading** — Visual polish
32. **37. Add Status Effect Icons** — Visual polish
33. **29. Implement Message Log Scrolling** — UX improvement
34. **32. Standardize Command Menu Layout** — Consistency
35. **33. Add Panel Title Headers** — Consistency
36. **27. Implement Turn Change Visual Effect** — UX polish
37. **30. Add Combat Round Prefix to Log Messages** — UX detail
38. **38. Increase Test Coverage to 70%** — Quality maintenance
39. **39. Document RPC API Completely** — Documentation
40. **40. Reduce Cyclomatic Complexity** — Code quality

---

## Completion Criteria

When all items above are implemented, the GoldBox RPG Engine will embody authentic Gold Box presentation and interaction with modern enhancements:

**Visual Fidelity**: The screen layout will feature fixed, non-overlapping panels with bold EGA-palette borders (3px bright outer, 1px inner, shadow lines). The viewport will show first-person dungeon corridors rendered from actual map data with themed palettes (classic, horror, natural, undead, magical), architectural features (pillars, altars, fountains), and atmospheric effects. Combat will display a tactical grid where player and monster sprites from the asset library are clearly visible. Status effects, morale states, and immunities will appear on combat tokens. Equipment will show in a paper-doll display with item sprites. Character portraits will appear in panels and encounters.

**Information Density**: The message log will receive every game event—attacks with hit/miss/damage detail (naming attacker, target, weapon), item pickups, door interactions, trap triggers, quest updates, level ups—in Gold Box narrative style with round number prefixes during combat. All numbers (HP, AC, damage, XP, gold, attribute modifiers) will be explicitly shown. Turn order, initiative values, and next combatant will be clearly visible. The log will support scrollback for reviewing history.

**Backend Surfacing**: Every implemented game system will be accessible through the UI—factions with diplomatic standings (color-coded from War to Allied), guilds with membership and treasury, morale affecting NPC behavior with visible state indicators, resistances and immunities from equipment and abilities, spell slots limiting caster resources, AI behavior types showing NPC tactics.

**Interaction Authenticity**: Movement will feel instant (step-and-turn with 50ms transitions) with noticeable visual feedback. Combat will follow turn-based tactical flow with clear current-turn indication, movement/attack range highlighting in distinct colors, opportunity attack warnings, cover/flanking indicators, and initiative tracking. All actions will have keyboard shortcuts with highlighted key letters in command menus. Touch controls will provide adequate targets for mobile play.

**Technical Quality**: Critical bugs (Stun/Root enforcement, WebSocket races, healing modifier, multiplicative stacking) will be fixed. All event types will emit to WebSocket clients. Test coverage will meet the 70% target. RPC API will be fully documented. Code complexity will remain manageable with functions under 50 lines and complexity under 10.

The result will be a game that feels immediately familiar to anyone who played Pool of Radiance or Curse of the Azure Bonds, while leveraging modern browser technology and touch support for accessibility, and procedural content generation for replayability.
