# Roadmap

Generated: 2026-03-20

## Gold Box–Inspired Reference Standard

This codebase targets a *retro-inspired* aesthetic rooted in the classic SSI Gold Box series (Pool of Radiance, Curse of the Azure Bonds, Champions of Krynn), while embracing modern enhancements that improve the gameplay experience. The visual foundation uses an EGA-inspired 16-color palette sensibility—bold, saturated, high-contrast colors with deep blues, magentas, purples, vivid yellows, and dungeon grays—rendered through fixed, non-overlapping panel layouts. The first-person exploration view features depth-sliced corridor rendering with themed palettes (classic, horror, natural, undead, magical), architectural features, and atmospheric effects driven by procedural content generation. Combat displays a top-down tactical grid with sprite-based entity representation and rich information density (attack rolls, damage, effects) flowing into the message log. Modern UX improvements include touch/mobile support with directional controls, pseudo-smooth transitions, and responsive input. All game numbers (HP, AC, damage, XP, gold, attribute modifiers) are explicitly shown—the player always knows the exact game state.

---

## Improvement Items

Items are grouped by theme and ordered within each group by priority (highest impact first). Each item is specified in enough detail for autonomous implementation.

---

### Group: Critical Bug Fixes

#### 1. Add Paralysis Effect Enforcement in Combat Handlers

**Priority:** High
**Complexity:** Small
**Depends on:** None

**Current State**
The `EffectParalysis` type is defined in `pkg/game/constants.go:58` as "an enhanced stun effect," but combat handlers only check for `EffectStun` and `EffectRoot`. The paralysis effect exists but provides no actual gameplay restriction.

- `handlers.go:195` checks `EffectStun` for movement blocking
- `handlers.go:204` checks `EffectRoot` for movement blocking  
- `handlers.go:430` checks `EffectStun` for attack blocking
- `handlers.go:641` checks `EffectStun` for spell blocking
- No handler checks `EffectParalysis`

**Gap**
Paralysis should be treated as a complete action block (like stun but stronger). Currently, paralyzed characters can move, attack, and cast spells freely. This violates Gold Box reference behavior where paralysis is a serious condition.

**Implementation Specification**
- Files to modify: `pkg/server/handlers.go`
- In `validateCombatConstraints()` (line 192), add paralysis check after stun check:
  ```go
  if player.HasEffect(game.EffectParalysis) {
      return fmt.Errorf("cannot act while paralyzed")
  }
  ```
- In `handleAttack()` (line 428), add paralysis check after stun check
- In `validateCombatConstraintsForSpell()` (line 639), add paralysis check after stun check
- Add test case in `pkg/server/handlers_test.go` following the stun test pattern (lines 851-885)
- Message log should announce: "Cannot act while paralyzed"

**Success Criteria**
- [ ] Paralyzed characters cannot move
- [ ] Paralyzed characters cannot attack
- [ ] Paralyzed characters cannot cast spells
- [ ] Test case verifies paralysis blocks all action types
- [ ] Message log announces paralysis block

---

### Group: Game System Wiring

#### 2. Integrate Opportunity Attack System into Combat

**Priority:** High
**Complexity:** Medium
**Depends on:** None

**Current State**
The opportunity attack system is fully implemented in `pkg/game/combat_opportunity.go` with:
- `OpportunityAttackManager` struct (line 12)
- `RegisterEntity()`, `UnregisterEntity()` methods (lines 39-57)
- `CheckMovement()` that detects opportunity attacks (line 64)
- Threatened square tracking (lines 89-120)

However, the system is **completely unused**:
- No handlers in `pkg/server/` reference `combat_opportunity.go`
- `grep -r "OpportunityAttack" pkg/server/` returns no matches
- Movement handler (`handlers.go:44`) does not check for opportunity attacks

The UI already shows opportunity attack warnings (ROADMAP item 10 marked complete), but server-side enforcement is missing.

**Gap**
Gold Box games feature opportunity attacks when characters disengage from melee combat. The backend system exists but is not integrated, meaning movement near enemies has no tactical consequence.

**Implementation Specification**
- Files to modify: `pkg/server/handlers.go`, `pkg/server/state.go`
- Add `OpportunityManager *game.OpportunityAttackManager` field to `GameState` in `state.go`
- Initialize in `NewGameState()`: `OpportunityManager: game.NewOpportunityAttackManager()`
- In `handleMove()` after successful movement (around line 280):
  ```go
  // Check for opportunity attacks from adjacent enemies
  opportunityAttacks := s.state.OpportunityManager.CheckMovement(
      player.ID, oldPos, newPos, isDisengage,
  )
  for _, oa := range opportunityAttacks {
      // Execute opportunity attack and log result
      result := s.executeOpportunityAttack(oa.AttackerID, player.ID)
      s.eventSys.Emit(EventOpportunityAttack, map[string]interface{}{
          "attacker": oa.AttackerID,
          "target":   player.ID,
          "result":   result,
      })
  }
  ```
- Register/unregister entities in combat start/end handlers
- Add `EventOpportunityAttack` constant to `pkg/game/constants.go`

**Success Criteria**
- [ ] Movement past enemies triggers opportunity attack check
- [ ] Opportunity attacks execute and deal damage
- [ ] Disengage action prevents opportunity attacks
- [ ] Event emitted for WebSocket clients
- [ ] Message log announces opportunity attack results
- [ ] Test case verifies opportunity attack trigger conditions

---

#### 3. Expose Morale Score in Combat State

**Priority:** Medium
**Complexity:** Small
**Depends on:** None

**Current State**
The morale system is fully implemented in `pkg/game/morale.go` with:
- `MoraleState` enum (Steadfast, Shaken, Broken, Panicked) at lines 13-22
- `GetMoraleScore(npcID)` returning numeric 0-100 score at line 175
- `GetMoraleState(npcID)` returning enum state at line 162

The combat UI receives morale state strings (`handlers.go:1002-1003`):
```go
entry["morale_state"] = game.MoraleStateString(moraleState)
```

But the numeric score is not exposed.

**Gap**
UI can only display state labels (Steadfast, Shaken, etc.), not progress bars or percentage indicators. Gold Box games showed morale as a more granular system. The numeric score exists but isn't transmitted.

**Implementation Specification**
- Files to modify: `pkg/server/handlers.go`
- In `buildInitiativeEntries()` (line 1010), add morale score after morale_state:
  ```go
  entry["morale_state"] = game.MoraleStateString(moraleState)
  entry["morale_score"] = s.state.MoraleSystem.GetMoraleScore(entityID)
  ```
- Update `types_rpc.go` documentation to note the new field
- UI can then display: "Shaken (42%)" or a progress bar

**Success Criteria**
- [ ] `morale_score` field (0-100) included in initiative entries
- [ ] Only included for NPCs (not player characters)
- [ ] UI can render morale progress bar
- [ ] Documentation updated

---

#### 4. Add Equipment Resistance Breakdown API

**Priority:** Medium
**Complexity:** Small
**Depends on:** None

**Current State**
The resistance system exists in `pkg/game/character_equipment.go`:
- `CalculateEquipmentResistances()` at line 443
- `GetResistance(damageType)` at line 489
- Equipment can provide resistance bonuses

The `getEquipment` handler (`handlers_equipment.go:239`) returns `equipment_bonuses` but not detailed resistance breakdown:
```go
"equipment_bonuses": char.CalculateEquipmentBonuses()
```

`CalculateEquipmentBonuses()` returns aggregate stats, not per-damage-type resistances.

**Gap**
Players cannot see their character's fire resistance, poison resistance, etc. This information is computed server-side but not transmitted. Gold Box games displayed resistances clearly.

**Implementation Specification**
- Files to modify: `pkg/server/handlers.go` (or new `handlers_character.go`)
- Add new RPC method `getCharacterResistances`:
  ```go
  func (s *RPCServer) handleGetCharacterResistances(params json.RawMessage) (interface{}, error) {
      // Parse session_id from params
      session, err := s.getSessionForMove(sessionID)
      if err != nil { return nil, err }
      defer s.releaseSession(session)
      
      resistances := make(map[string]int)
      for _, dt := range []game.DamageType{
          game.DamagePhysical, game.DamageFire, game.DamagePoison,
          game.DamageFrost, game.DamageLightning,
      } {
          resistances[string(dt)] = session.Player.GetResistance(dt)
      }
      return map[string]interface{}{"resistances": resistances}, nil
  }
  ```
- Register method in `registerMethods()`
- Optionally add to PlayerState in `buildPlayerStateData()`:
  ```go
  "resistances": player.Character.GetAllResistances(),
  ```

**Success Criteria**
- [ ] RPC method `getCharacterResistances` returns per-type resistance percentages
- [ ] Character panel can display resistance icons/values
- [ ] Equipment changes reflect in resistance values
- [ ] Documentation updated in `pkg/README-RPC.md`

---

#### 5. Expose AI Behavior Type in Combat State

**Priority:** Low
**Complexity:** Small
**Depends on:** None

**Current State**
AI behavior trees are implemented in `pkg/game/ai_behaviors.go`:
- `AggressiveTree()`, `GuardTree()`, `PatrolTree()`, `CowardTree()` at lines 547-646
- AI difficulty levels in `ai_combat.go:18-28`

The ROADMAP item 13 ("Add AI Behavior Type Display") is marked complete, and there's a placeholder for behavior type in initiative entries. However, the actual AI behavior assignment and retrieval isn't connected.

**Gap**
Verify that NPC behavior type (Aggressive, Guard, Patrol, Coward) is actually transmitted in combat state and displayed in UI. If the wiring is incomplete, complete it.

**Implementation Specification**
- Verify `handlers.go` includes behavior type in initiative entries
- If missing, add to `buildInitiativeEntries()`:
  ```go
  if npc := s.state.GetNPC(entityID); npc != nil {
      entry["behavior_type"] = npc.BehaviorType // "aggressive", "guard", "patrol", "coward"
  }
  ```
- NPCs need a `BehaviorType` field (string) that maps to their behavior tree
- UI should display behavior icon/text in initiative tracker

**Success Criteria**
- [ ] NPC behavior type included in initiative entries
- [ ] Behavior type displays in combat UI
- [ ] Different behaviors are distinguishable
- [ ] Test verifies behavior type transmission

---

### Group: Asset Integration

#### 6. Call Sprite Preload Functions on Game Start

**Priority:** High
**Complexity:** Small
**Depends on:** None

**Current State**
Sprite preload functions exist but are **never called** during game initialization:
- `PreloadCharacterSprites()` defined at `asset_loader.go:421` — loads 48 character portraits
- `PreloadTerrainSprites()` defined at `asset_loader.go:439` — loads 10 terrain tiles
- Both functions are only called in `asset_loader_test.go:234-235` for tests

Current behavior: Sprites load on-demand via `DrawSpriteWithFallback()`, causing purple placeholder rectangles until HTTP fetch completes.

**Gap**
Players see purple placeholder rectangles during initial gameplay while sprites load asynchronously. Gold Box games had all assets loaded before gameplay began—no pop-in.

**Implementation Specification**
- Files to modify: `pkg/wasmui/game.go`
- In `NewGame()` or game initialization, add calls:
  ```go
  // Preload critical sprites before gameplay
  PreloadCharacterSprites()
  PreloadTerrainSprites()
  preloadMonsterSprites() // New function to add
  ```
- Add new function `PreloadMonsterSprites()` in `asset_loader.go`:
  ```go
  func PreloadMonsterSprites() {
      initSpriteCache()
      monsters := []string{
          "skeleton", "zombie", "goblin", "orc", "dragon",
          "spider", "rat", "wolf", "elemental", "demon",
      }
      var paths []string
      for _, m := range monsters {
          paths = append(paths, MonsterSpritePath(m))
      }
      spriteCache.Preload(paths)
  }
  ```
- Consider adding a loading screen that waits for preload completion

**Success Criteria**
- [ ] Character portraits preloaded at game start
- [ ] Terrain tiles preloaded at game start
- [ ] Common monster sprites preloaded
- [ ] No purple placeholder rectangles during normal gameplay
- [ ] Loading indicator during preload phase (optional)

---

#### 7. Add UI Panel Sprite Backgrounds

**Priority:** Medium
**Complexity:** Medium
**Depends on:** 6

**Current State**
All UI panels use colored rectangles via `drawRect()`:
- Character panel: `exploration.go:1289` uses `drawBoldPanelBorder()` + `drawRect()` for background
- Combat log: `exploration.go:1944` uses `drawRect()` for background
- Initiative panel: `combat_screen.go:761` uses `drawRect()` for background
- All overlays: `overlays.go` uses `drawRect()` throughout

UI panel sprites are defined in `game-assets.yaml` and path functions exist:
- `UIPanelPath(panelType)` at `asset_loader.go:274` returns paths like `ui/panels/ui_panel_character.png`

**Gap**
Gold Box games had decorative panel backgrounds (parchment texture, stone borders, etc.). Current implementation uses flat colored rectangles—functional but visually bland.

**Implementation Specification**
- Files to modify: `pkg/wasmui/exploration.go`, `pkg/wasmui/combat_screen.go`, `pkg/wasmui/overlays.go`
- Create helper function in `game.go`:
  ```go
  func drawPanelBackground(screen *ebiten.Image, x, y, w, h int, panelType string) {
      path := UIPanelPath(panelType)
      if spriteCache.IsCached(path) {
          DrawSpriteScaled(screen, path, x, y, w, h)
      } else {
          spriteCache.Get(path) // Start loading
          drawRect(screen, x, y, w, h, ColorPanelBG) // Fallback
      }
  }
  ```
- Replace `drawRect(screen, panelX, panelY, charPanelWidth, panelHeight, ColorPanelBG)` calls with `drawPanelBackground()`
- Panel types: "character", "combat_log", "initiative", "inventory", "dialog_stone"
- Ensure panel sprites exist in `web/static/assets/sprites/ui/panels/`

**Success Criteria**
- [ ] Character panel uses sprite background
- [ ] Combat log panel uses sprite background
- [ ] Initiative panel uses sprite background
- [ ] Inventory overlay uses sprite background
- [ ] Graceful fallback to colored rectangles if sprites missing

---

#### 8. Load Spell Effect Sprites for Combat Animations

**Priority:** Medium
**Complexity:** Medium
**Depends on:** 6

**Current State**
Spell effects use procedural rendering with expanding circles:
- `combat_screen.go:1231-1270` draws spell effects as colored circles
- `SpellEffectPath()` defined at `asset_loader.go:245` returns `effects/spells/effect_spell_{spellID}.png`
- `SpellEffect` struct in `types_ui.go:389-413` has `SpellID` field

Current rendering (`combat_screen.go:1254-1260`):
```go
// Draw expanding circle effect (procedural fallback)
ebitenutil.DrawCircle(screen, float64(x), float64(y), radius, effectColor)
```

**Gap**
Gold Box games had distinctive spell effect sprites (fireballs, lightning bolts, etc.). Current implementation shows generic colored circles for all spells.

**Implementation Specification**
- Files to modify: `pkg/wasmui/combat_screen.go`
- In `drawSpellEffects()`, attempt sprite loading before procedural fallback:
  ```go
  func (g *Game) drawSpellEffect(screen *ebiten.Image, effect *SpellEffect, x, y int) {
      path := SpellEffectPath(effect.SpellID)
      if spriteCache.IsCached(path) {
          // Draw sprite frame based on effect.GetFrame()
          DrawSpriteScaled(screen, path, x-32, y-32, 64, 64)
          return
      }
      spriteCache.Get(path) // Start loading
      // Procedural fallback
      g.drawProceduralSpellEffect(screen, effect, x, y)
  }
  ```
- Support sprite sheets with multiple frames for animation
- Add spell effect sprites to preload list
- Ensure sprites exist in `web/static/assets/sprites/effects/spells/`

**Success Criteria**
- [ ] Fireball shows fire explosion sprite
- [ ] Magic Missile shows projectile sprite
- [ ] Healing spells show healing glow sprite
- [ ] Graceful fallback to colored circles if sprites missing
- [ ] Animation frames cycle correctly

---

### Group: Message Log Enhancement

#### 9. Add Message Log Event Filtering

**Priority:** Medium
**Complexity:** Medium
**Depends on:** None

**Current State**
The message log displays all messages chronologically:
- `LogMessage` struct in `types_ui.go:42-48` has `Type MessageType` field
- `MessageType` enum (lines 51-63): Info, Warning, Error, Combat, System, Loot, Quest, LevelUp, Interact
- No filtering mechanism exists in the UI

Messages accumulate with no way to filter by category.

**Gap**
Gold Box games had contextual message display—combat log showed combat-relevant messages, exploration showed exploration-relevant messages. Players need ability to filter noise.

**Implementation Specification**
- Files to modify: `pkg/wasmui/exploration.go`, `pkg/wasmui/types_ui.go`, `pkg/wasmui/game.go`
- Add filter state to Game struct:
  ```go
  type Game struct {
      // ...existing fields...
      logFilter   MessageType   // Currently active filter (0 = all)
      logFilters  []MessageType // Enabled filter types
  }
  ```
- Add filter toggle hotkeys in exploration mode:
  - `Ctrl+L` → Cycle through filter modes
  - Filter modes: All, Combat Only, Quest Only, Loot Only
- Modify `drawCombatLog()` to respect filter:
  ```go
  for _, msg := range g.logMessages {
      if g.logFilter != 0 && msg.Type != g.logFilter {
          continue // Skip filtered messages
      }
      // Draw message
  }
  ```
- Add filter indicator in log header: "MESSAGE LOG [Combat]"

**Success Criteria**
- [ ] Ctrl+L cycles through filter modes
- [ ] Filter mode displayed in log header
- [ ] Combat filter shows only combat messages
- [ ] Quest filter shows only quest messages
- [ ] "All" mode shows all messages

---

#### 10. Add Missing Event Types to Message Log

**Priority:** Medium
**Complexity:** Small
**Depends on:** None

**Current State**
`MessageType` enum in `types_ui.go:51-63` covers:
- Info, Warning, Error, Combat, System, Loot, Quest, LevelUp, Interact

Missing from enum:
- Trap triggers
- Critical hits (mentioned in code but no dedicated type)
- Saving throw results
- Condition applications (curse, bless, etc.)
- Rest completion
- Spell failure/fizzle

**Gap**
Not all game events have appropriate message types for styling. Trap triggers should be red/warning, critical hits should be emphasized, etc.

**Implementation Specification**
- Files to modify: `pkg/wasmui/types_ui.go`
- Add new MessageType constants:
  ```go
  const (
      // ...existing types...
      MessageTrap      // Trap triggers (red/warning)
      MessageCritical  // Critical hits (gold/emphasized)
      MessageSave      // Saving throws (blue)
      MessageCondition // Status effect applications (purple)
      MessageRest      // Rest and recovery (green)
      MessageSpellFail // Spell failure/fizzle (gray)
  )
  ```
- Add color mappings in `MessageType.Color()`:
  ```go
  case MessageTrap:
      return color.RGBA{R: 255, G: 80, B: 80, A: 255} // Bright red
  case MessageCritical:
      return color.RGBA{R: 255, G: 215, B: 0, A: 255} // Gold
  case MessageSave:
      return color.RGBA{R: 100, G: 150, B: 255, A: 255} // Light blue
  case MessageCondition:
      return color.RGBA{R: 200, G: 150, B: 255, A: 255} // Purple
  case MessageRest:
      return color.RGBA{R: 100, G: 200, B: 100, A: 255} // Green
  case MessageSpellFail:
      return color.RGBA{R: 150, G: 150, B: 150, A: 255} // Gray
  ```
- Update message logging calls throughout codebase to use appropriate types

**Success Criteria**
- [ ] Trap messages display in warning red
- [ ] Critical hits display in gold
- [ ] Saving throws display in light blue
- [ ] Condition applications display in purple
- [ ] All message types have distinct colors

---

#### 11. Display Message Timestamp on Hover

**Priority:** Low
**Complexity:** Small
**Depends on:** None

**Current State**
`LogMessage` struct has `Timestamp int64` field (`types_ui.go:46`) that is set when messages are created (`game.go:847`) but never displayed.

**Gap**
Players cannot see when messages occurred. Useful for reviewing combat timeline or debugging.

**Implementation Specification**
- Files to modify: `pkg/wasmui/exploration.go`
- In `drawCombatLog()`, track hovered message via mouse position
- When mouse hovers over a message line, draw tooltip with timestamp:
  ```go
  // Check hover
  if mx >= logX && mx < logX+logWidth && my >= msgY && my < msgY+lineHeight {
      hoveredMsg = msg
  }
  // Draw tooltip
  if hoveredMsg != nil {
      timeStr := time.Unix(0, hoveredMsg.Timestamp).Format("15:04:05")
      drawTooltip(screen, mx+10, my, timeStr)
  }
  ```
- Format timestamp as HH:MM:SS

**Success Criteria**
- [ ] Hovering message shows timestamp tooltip
- [ ] Timestamp format is readable (HH:MM:SS)
- [ ] Tooltip doesn't overlap other UI
- [ ] Touch users can tap to see timestamp

---

### Group: Combat Screen Enhancement

#### 12. Add Attack Animation Sprites

**Priority:** Medium
**Complexity:** Large
**Depends on:** 6

**Current State**
Combat attacks have damage flash and damage popup (`types_ui.go:319-386`) but no attack animation:
- `DamageFlash` provides color overlay on hit
- `DamagePopup` shows floating damage numbers
- No sword swing, arrow flight, or melee contact animation

**Gap**
Gold Box games showed brief weapon swing animations when attacking. Current implementation is damage number + flash only—attack action is visually instantaneous.

**Implementation Specification**
- Files to modify: `pkg/wasmui/combat_screen.go`, `pkg/wasmui/types_ui.go`
- Add `AttackAnimation` struct to `types_ui.go`:
  ```go
  type AttackAnimation struct {
      AttackerID  string
      TargetID    string
      WeaponType  string        // "sword", "bow", "spell", "fist"
      StartTime   time.Time
      Duration    time.Duration // ~300ms
      StartPos    Position      // Attacker grid position
      EndPos      Position      // Target grid position
  }
  ```
- Add `attackAnimations []*AttackAnimation` to Game struct
- In `combat_screen.go`, add drawing function:
  ```go
  func (g *Game) drawAttackAnimations(screen *ebiten.Image) {
      for _, anim := range g.attackAnimations {
          if !anim.IsActive() { continue }
          progress := anim.Progress()
          // Interpolate position from attacker to target
          x := lerp(anim.StartPos.X, anim.EndPos.X, progress)
          y := lerp(anim.StartPos.Y, anim.EndPos.Y, progress)
          // Draw weapon sprite at interpolated position
          DrawSprite(screen, WeaponAnimPath(anim.WeaponType), x, y)
      }
  }
  ```
- Trigger animation when attack action executes
- Support: melee swing (0→target→0), ranged projectile (attacker→target), spell effect (expands at target)

**Success Criteria**
- [ ] Melee attacks show swing animation toward target
- [ ] Ranged attacks show projectile traveling to target
- [ ] Animation duration ~300ms
- [ ] Works with damage flash and popup
- [ ] Graceful fallback (no animation) if sprites missing

---

#### 13. Add Cover Indicator Sprites to Combat Grid

**Priority:** Medium
**Complexity:** Medium
**Depends on:** 6

**Current State**
Cover system is implemented in `pkg/game/combat_modifiers.go`:
- `CoverNone`, `CoverHalf`, `CoverThreeQuarters`, `CoverFull` at lines 11-37
- `CalculateCover()` computes cover based on obstacles
- `getCombatModifiers` RPC returns cover_type and cover_bonus

UI currently shows no visual indication of cover.

**Gap**
Players cannot see which tiles provide cover bonuses. Gold Box games showed obstacles clearly affecting tactical positioning.

**Implementation Specification**
- Files to modify: `pkg/wasmui/combat_screen.go`
- When character is selected and in attack mode, highlight tiles with cover:
  ```go
  func (g *Game) drawCoverIndicators(screen *ebiten.Image, attackerPos Position) {
      for _, enemy := range g.combatState.Enemies {
          coverType := g.getCoverType(attackerPos, enemy.Position)
          if coverType == "none" { continue }
          // Draw cover icon at enemy position
          iconPath := fmt.Sprintf("ui/icons/ui_icon_cover_%s.png", coverType)
          DrawSpriteWithFallback(screen, iconPath, ...)
      }
  }
  ```
- Cover icons: half shield (half cover), three-quarter shield (3/4 cover), full shield (full cover)
- Show cover percentage as text: "+2", "+5", "+10" AC
- Color: Blue for cover that benefits player, red for cover benefiting enemy

**Success Criteria**
- [ ] Cover indicators display in attack mode
- [ ] Half/3Q/Full cover have distinct icons
- [ ] AC bonus shown as text overlay
- [ ] Indicators visible but not intrusive
- [ ] Works for both player and enemy analysis

---

#### 14. Add Flanking Indicator to Combat Grid

**Priority:** Medium
**Complexity:** Small
**Depends on:** 6, 13

**Current State**
Flanking system exists in `pkg/game/combat_modifiers.go:227-294`:
- `CalculateFlanking()` detects flanking bonus
- Returns +2 attack when ally is opposite defender
- `getCombatModifiers` RPC returns `is_flanking` and `flanking_bonus`

UI shows no flanking indicator.

**Gap**
Players cannot see when they have flanking advantage. Important tactical information is hidden.

**Implementation Specification**
- Files to modify: `pkg/wasmui/combat_screen.go`
- When in attack mode with target selected, show flanking indicator:
  ```go
  func (g *Game) drawFlankingIndicator(screen *ebiten.Image, attackerPos, targetPos Position) {
      isFlanking, bonus := g.checkFlanking(attackerPos, targetPos)
      if !isFlanking { return }
      // Draw crossed swords icon at target position
      DrawSpriteWithFallback(screen, "ui/icons/ui_icon_flanking.png", ...)
      // Draw "+2" bonus text
      drawColoredText(screen, fmt.Sprintf("+%d", bonus), x, y, ColorGoldHi)
  }
  ```
- Show line connecting player and flanking ally (optional)
- Highlight the ally providing flank

**Success Criteria**
- [ ] Flanking indicator shows when attack has flanking bonus
- [ ] Flanking ally highlighted
- [ ] Attack bonus displayed
- [ ] Works for multi-ally flanking

---

### Group: Exploration Screen Enhancement

#### 15. Add Enemy Position Markers to Minimap

**Priority:** Medium
**Complexity:** Small
**Depends on:** None

**Current State**
Minimap shows:
- Explored tiles (`exploration.go:1773-1785`)
- Player position as green dot (`exploration.go:1786-1800`)
- Compass indicator (`overlays.go:1595-1626`)

Enemy positions are NOT shown on minimap.

**Gap**
Gold Box games showed enemy positions on tactical displays. Current minimap is purely navigational—players must find enemies visually in the 3D view.

**Implementation Specification**
- Files to modify: `pkg/wasmui/exploration.go`, `pkg/wasmui/overlays.go`
- Add enemy drawing to `drawMinimapTiles()`:
  ```go
  func (g *Game) drawMinimapTiles(screen *ebiten.Image, offsetX, offsetY, scale int) {
      // ...existing tile drawing...
      
      // Draw enemy positions in red
      for _, enemy := range g.visibleEnemies {
          ex := offsetX + (enemy.Position.X - g.playerPos.X + centerOffset) * scale
          ey := offsetY + (enemy.Position.Y - g.playerPos.Y + centerOffset) * scale
          drawRect(screen, ex, ey, scale, scale, ColorEnemyName)
      }
  }
  ```
- Only show enemies within visible/detected range
- Use red color for enemies, yellow for neutral NPCs

**Success Criteria**
- [ ] Enemy positions shown as red dots on minimap
- [ ] Neutral NPCs shown as yellow dots
- [ ] Only visible/detected enemies shown
- [ ] Updates in real-time as enemies move

---

#### 16. Add Quest Markers to Minimap

**Priority:** Low
**Complexity:** Small
**Depends on:** 15

**Current State**
Quest system is fully implemented with objectives (`pkg/game/quest.go`), but minimap shows no quest-related markers.

**Gap**
Players have no visual indication of quest objective locations on the map. Must rely on text descriptions.

**Implementation Specification**
- Files to modify: `pkg/wasmui/exploration.go`
- Quest objectives need location data (may require server-side addition)
- Add quest marker drawing after enemy positions:
  ```go
  // Draw quest objective markers in gold
  for _, quest := range g.activeQuests {
      for _, obj := range quest.Objectives {
          if obj.Position.Level != g.playerPos.Level { continue }
          qx := offsetX + (obj.Position.X - g.playerPos.X + centerOffset) * scale
          qy := offsetY + (obj.Position.Y - g.playerPos.Y + centerOffset) * scale
          drawRect(screen, qx, qy, scale+2, scale+2, ColorGold)
      }
  }
  ```
- Quest markers should be gold/yellow stars or exclamation points

**Success Criteria**
- [ ] Active quest objectives shown on minimap
- [ ] Quest markers distinct from enemy markers
- [ ] Only current-level objectives shown
- [ ] Markers disappear when objective completed

---

### Group: UI Polish

#### 17. Add Tooltip System for UI Elements

**Priority:** Medium
**Complexity:** Medium
**Depends on:** None

**Current State**
Equipment slots have hover detection (`overlays.go:83-106`) with `hoveredEquipSlot` tracking, but no tooltip is rendered. Various UI elements could benefit from tooltips.

**Gap**
Players must click/interact to see item details. Gold Box games showed item stats on highlight.

**Implementation Specification**
- Files to modify: `pkg/wasmui/game.go`, `pkg/wasmui/overlays.go`
- Add tooltip state to Game:
  ```go
  type Tooltip struct {
      Text     string
      X, Y     int
      Visible  bool
      ShowTime time.Time
  }
  ```
- Add global tooltip drawing function:
  ```go
  func (g *Game) drawTooltip(screen *ebiten.Image) {
      if !g.tooltip.Visible { return }
      if time.Since(g.tooltip.ShowTime) < 300*time.Millisecond { return } // Delay
      
      // Measure text
      lines := strings.Split(g.tooltip.Text, "\n")
      maxWidth := 0
      for _, line := range lines {
          if len(line)*7 > maxWidth { maxWidth = len(line) * 7 }
      }
      height := len(lines) * 14
      
      // Draw background
      drawRect(screen, g.tooltip.X, g.tooltip.Y, maxWidth+8, height+8, ColorPanelBG)
      drawBoldPanelBorder(screen, g.tooltip.X, g.tooltip.Y, maxWidth+8, height+8)
      
      // Draw text
      for i, line := range lines {
          drawColoredText(screen, line, g.tooltip.X+4, g.tooltip.Y+4+i*14, ColorStatValue)
      }
  }
  ```
- Set tooltip content on hover:
  ```go
  g.tooltip = Tooltip{
      Text:     fmt.Sprintf("%s\nAC Bonus: +%d\nWeight: %d", item.Name, item.ACBonus, item.Weight),
      X:        mx + 10,
      Y:        my + 10,
      Visible:  true,
      ShowTime: time.Now(),
  }
  ```
- Clear tooltip when mouse moves away

**Success Criteria**
- [ ] Tooltips appear after 300ms hover delay
- [ ] Item tooltips show name, stats, description
- [ ] Tooltips have panel border styling
- [ ] Tooltips don't extend off screen
- [ ] Touch users can tap-and-hold for tooltip

---

#### 18. Add Keyboard Navigation for Overlays

**Priority:** Low
**Complexity:** Small
**Depends on:** None

**Current State**
Overlays (inventory, quest log, guild panel) use mouse/touch for navigation. Keyboard support is limited to Escape for close.

**Gap**
Gold Box games were fully keyboard-navigable. Power users expect arrow key navigation.

**Implementation Specification**
- Files to modify: `pkg/wasmui/overlays.go`
- Add selection state to overlays:
  ```go
  type InventoryState struct {
      selectedIndex int
      scrollOffset  int
  }
  ```
- Handle arrow keys in `updateInventory()`:
  ```go
  if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
      g.inventoryState.selectedIndex--
  }
  if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
      g.inventoryState.selectedIndex++
  }
  if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
      g.activateSelectedItem()
  }
  ```
- Highlight selected item row
- Support Tab to cycle between sections

**Success Criteria**
- [ ] Up/Down arrows navigate item list
- [ ] Enter activates selected item
- [ ] Tab cycles between overlay sections
- [ ] Page Up/Down scrolls faster
- [ ] Selection wraps at list boundaries

---

### Group: Quality Maintenance

These items are carried forward from quality maintenance concerns.

#### 19. Reduce Cyclomatic Complexity in Large Functions

**Priority:** Low
**Complexity:** Medium
**Depends on:** None

**Current State**
Several handler functions in `pkg/server/handlers.go` exceed 50 lines:
- `handleApplyEffect()`: 97 lines (line 1725)
- `handleGetCombatModifiers()`: 72 lines (line 1473)
- `handleRest()`: 62 lines (line 1368)
- `buildCharacterConfig()`: 61 lines (line 2085)
- `executeCastSpellAction()`: 58 lines (line 591)
- `buildCharacterData()`: 57 lines (line 2009)
- `handleAttack()`: 54 lines (line 386)
- `buildInitiativeEntries()`: 51 lines (line 1010)

**Gap**
High complexity reduces maintainability and makes testing harder. Target: no function >50 lines, cyclomatic complexity ≤10.

**Implementation Specification**
- Files to modify: `pkg/server/handlers.go`
- Extract logical blocks into helper functions:
  ```go
  // Before
  func (s *RPCServer) handleApplyEffect(params json.RawMessage) (interface{}, error) {
      // 97 lines with nested ifs
  }
  
  // After
  func (s *RPCServer) handleApplyEffect(params json.RawMessage) (interface{}, error) {
      req, err := parseApplyEffectRequest(params)
      if err != nil { return nil, err }
      
      session, err := s.getSessionForEffect(req.SessionID)
      if err != nil { return nil, err }
      defer s.releaseSession(session)
      
      effect := buildEffectFromRequest(req)
      result, err := s.applyEffectToTarget(session, effect, req.TargetID)
      if err != nil { return nil, err }
      
      return formatEffectResult(result), nil
  }
  ```
- Use early returns to reduce nesting
- Group related validation into dedicated functions

**Success Criteria**
- [ ] No function exceeds 50 lines
- [ ] Cyclomatic complexity ≤10 for all functions
- [ ] Existing tests still pass
- [ ] Code review finds no significant maintainability issues

---

#### 20. Add Tests for Untested Source Files

**Priority:** Low
**Complexity:** Large
**Depends on:** None

**Current State**
Coverage analysis shows 92 Go source files without test files, including:
- `pkg/game/effects.go` - effect system
- `pkg/game/faction_helpers.go` - faction utilities
- `pkg/server/combat.go` - combat logic
- `pkg/server/alerting.go` - alerting system
- `pkg/pcg/` - multiple PCG files
- `pkg/persistence/` - persistence layer
- `pkg/resilience/` - resilience patterns

Current coverage is 65-96% depending on package.

**Gap**
Target coverage is ≥70% for all critical packages. Untested files represent risk of regressions.

**Implementation Specification**
- Run `make find-untested` to identify gaps
- Priority order for test creation:
  1. `pkg/game/effects.go` - Core game mechanic
  2. `pkg/server/combat.go` - Combat system
  3. `pkg/game/faction_helpers.go` - Faction utilities
  4. `pkg/persistence/atomic.go` - Data safety
- Use table-driven tests following existing patterns:
  ```go
  func TestEffectApplication(t *testing.T) {
      tests := []struct {
          name     string
          effect   Effect
          expected int
      }{
          {"poison tick", Effect{Type: EffectPoison, Magnitude: 3}, -3},
          // ...
      }
      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              // ...
          })
      }
  }
  ```
- Focus on:
  - Edge cases (nil inputs, boundary values)
  - Error handling paths
  - Concurrent access (use `go test -race`)

**Success Criteria**
- [ ] Overall coverage ≥70%
- [ ] Critical game logic paths covered
- [ ] No untested error handling paths
- [ ] Race detector passes on all tests

---

## Implementation Order

A recommended sequencing of ALL items above, accounting for dependencies and risk:

1. **1. Add Paralysis Effect Enforcement** — Critical bug fix, simple implementation
2. **2. Integrate Opportunity Attack System** — High-value tactical system, enables meaningful movement decisions
3. **6. Call Sprite Preload Functions** — Foundation for all asset improvements, quick win
4. **3. Expose Morale Score in Combat** — Simple API addition with UI value
5. **4. Add Equipment Resistance API** — Important character information surfacing
6. **10. Add Missing Message Types** — Improves all message logging, foundational
7. **9. Add Message Log Filtering** — UX improvement building on message types
8. **15. Add Enemy Markers to Minimap** — High-impact tactical improvement
9. **7. Add UI Panel Sprites** — Visual polish building on preloading
10. **8. Load Spell Effect Sprites** — Combat visual polish
11. **12. Add Attack Animation Sprites** — Combat feedback improvement
12. **13. Add Cover Indicators** — Tactical information display
13. **14. Add Flanking Indicator** — Tactical information display
14. **17. Add Tooltip System** — UX polish enabling detailed information
15. **5. Expose AI Behavior Type** — Combat information completeness
16. **11. Display Message Timestamp** — Minor UX improvement
17. **16. Add Quest Markers to Minimap** — Navigation improvement
18. **18. Add Keyboard Navigation** — Accessibility improvement
19. **19. Reduce Cyclomatic Complexity** — Code quality maintenance
20. **20. Add Tests for Untested Files** — Long-term quality investment

---

## Completion Criteria

When all items above are implemented, the GoldBox RPG Engine will achieve a comprehensive Gold Box–inspired experience with modern enhancements:

**Tactical Depth**: The combat system will enforce all status effect restrictions (stun, root, paralysis), execute opportunity attacks when characters disengage improperly, display cover and flanking bonuses visually on the tactical grid, and show AI behavior types so players can predict enemy tactics. Morale will be visible as a progress bar, not just a state label.

**Visual Polish**: Sprites will preload before gameplay begins, eliminating placeholder pop-in. UI panels will use decorative sprite backgrounds instead of flat rectangles. Spell effects will use animated sprites matching their schools. Attack animations will show weapon swings and projectile flights. The minimap will display enemy positions and quest objectives.

**Information Density**: The message log will use distinct colors for traps, critical hits, saving throws, and status conditions. Players can filter the log by category. Tooltips will appear on hover, showing detailed item stats. Timestamps will be visible for combat replay analysis.

**Code Quality**: All handler functions will be under 50 lines with clear separation of concerns. Test coverage will meet 70% minimum. The codebase will maintain its established patterns while reducing complexity hotspots.

The result will be a game that captures the tactical richness and information density of classic Gold Box titles while leveraging modern browser technology, touch support, and procedural content generation for enhanced accessibility and replayability.
