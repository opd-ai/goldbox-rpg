# Roadmap

Generated: 2026-03-19

## Gold Box Reference Standard

A "Gold Box faithful" implementation for this codebase means an 800×600 fixed-panel layout where **all game feedback flows through a scrolling, color-coded text log** as the primary output channel — not as a secondary display. The SSI Gold Box games (Pool of Radiance, Curse of the Azure Bonds, Champions of Krynn) communicated every combat result, exploration event, spell outcome, and DM narration through dense, information-rich text in the message panel. Text was color-differentiated by type: combat actions in one color, system messages in another, errors distinct from narration. The current WASM UI has the correct four-panel layout (viewport, character panel, message log, action bar) at the right resolution, and it stores message types with per-type color definitions (`MessageType.Color()` in `types_ui.go`), but renders every message as identical white text via `ebitenutil.DebugPrintAt()`. This makes the message log — the single most important UI surface in a Gold Box game — functionally monochrome and difficult to scan during combat. Fixing this is the highest-leverage change available for Gold Box authenticity.

## Recommended Improvement: Colored Text Rendering for the Message Log and UI Panels

### What It Is

Replace all `ebitenutil.DebugPrintAt()` calls in the WASM UI with a new `drawColoredText()` function that renders text in arbitrary `color.RGBA` values using Ebitengine's `ColorScale` compositing. The message log (`drawCombatLog` in `exploration.go`) will use each message's `MessageType.Color()` return value so combat messages appear purple, warnings yellow, errors red, system messages cyan, and informational text light gray. Panel headers, player names in the initiative list, turn indicators, and attribute labels will also gain distinct colors drawn from the `ASSET_ANALYSIS.md` palette (gold `#BFA54A` for headers, green for player names, red for enemy names, deep blue `#2E5090` for panel borders).

### Why This, Why Now

**The gap is immediate and pervasive.** Every text-rendering call in the frontend uses `ebitenutil.DebugPrintAt()`, which ignores the color parameter entirely — it only emits white text. The codebase already has:

- **`MessageType.Color()`** (`types_ui.go:26–38`): Returns distinct `color.RGBA` values for each of the five message types (Info, Warning, Error, Combat, System). This function is defined, tested, and **never called during rendering**.
- **`LogMessage.Type`** (`types_ui.go:8–12`): Every log message stores its type. All 30+ `addLogMessage()` call sites throughout the UI (`game.go`, `combat_screen.go`, `exploration.go`, `overlays.go`, `adventure_screen.go`, `character_creation.go`) pass an explicit `MessageType`.
- **`drawRect()`** (`game.go:764–770`): Already demonstrates the `ColorScale.ScaleWithColor()` pattern with a 1×1 white `pixelImage` — the exact same technique needed for colored text.
- **Explicit `_ = nameColor` suppressions** (`combat_screen.go:300`, `combat_screen.go:346`): The code defines color variables for initiative-list names and turn indicators, then discards them with comments like `// would use with text/v2; DebugPrintAt doesn't support color`. The intent was always to use colored text; only the rendering function is missing.

There are **192 calls** to `ebitenutil.DebugPrintAt` across the 9 WASM UI source files that render gameplay screens. The message log alone (`drawCombatLog`, `exploration.go:446–472`) renders every message at line 471 as plain white text, despite having access to `msg.Type`. This single function is the highest-impact fix point.

**Why this ranks above other candidates:**

1. **Self-contained**: Requires adding one function (`drawColoredText`) and updating call sites. No server changes, no new RPC methods, no game logic changes.
2. **Immediate visual impact**: Every screen — exploration, combat, inventory, spellcasting, character creation — benefits. The message log transforms from a wall of white text into a scannable, color-differentiated feed.
3. **Gold Box core identity**: In the reference games, the text log was the *primary* way players understood combat outcomes ("Sable attacks Orc — HIT for 7 damage" in combat color, "Save vs. Spell — FAILED" in warning color). Monochrome text undermines this completely.
4. **Low risk**: No behavioral changes to game logic, no new dependencies, no changes to server or RPC contracts. The `drawRect` pattern using `pixelImage` + `ColorScale` is already proven in the codebase.
5. **Enables future work**: Colored text is a prerequisite for many backlog items (combat narration formatting, effect damage ticks in the log, morale state indicators).

### Gold Box Alignment

This improvement directly addresses three Gold Box reference characteristics:

- **Typography and Text**: "Dense, information-rich text panels — the message log is a primary output channel, not secondary." Currently the log is visually flat and hard to scan. Color differentiation makes it primary.
- **Color and Visual Style**: "16-color EGA palette sensibility: bold, saturated, high-contrast." The existing `MessageType.Color()` values (purple `{200,150,255}` for combat, yellow `{255,200,0}` for warnings, red `{255,100,100}` for errors, cyan `{150,200,255}` for system) already approximate EGA-palette boldness. Panel headers using gold `#BFA54A` and borders using deep blue `#2E5090` from `ASSET_ANALYSIS.md` complete the palette.
- **Combat**: "All combat narration flows into the text log." The narration flows there today, but its visual identity is indistinguishable from system messages, errors, or movement confirmations. Color coding makes combat feedback visually distinct.

### Implementation Specification

**New function to add:**

- **File**: `pkg/wasmui/game.go` (alongside existing `drawRect`, `drawRectOutline`, `drawLine` helpers starting at line 760)
- **Function**: `drawColoredText(screen *ebiten.Image, text string, x, y int, c color.Color)`
- **Behavior**: Use Ebitengine's built-in debug font (`ebitenutil.DebugPrintAt` draws into a temporary image or uses the `text` package internally). The implementation should:
  1. Create a shared offscreen `*ebiten.Image` buffer (e.g., `textBuffer`, sized to accommodate typical text widths — 800×16 is sufficient for single-line text).
  2. Clear the buffer each call, render white text into it with `ebitenutil.DebugPrintAt(textBuffer, text, 0, 0)`.
  3. Draw the buffer onto `screen` at `(x, y)` with `DrawImageOptions.ColorScale.ScaleWithColor(c)`.
  4. The buffer should be a package-level variable (like `pixelImage`) initialized once in `init()` or lazily, to avoid per-frame allocation.
- **Alternative approach**: If Ebitengine's `text/v2` package is available in the project's Ebiten version (v2.9.x includes it), use `text.Draw()` with `text.DrawOptions` which natively supports `ColorScale`. This is cleaner but requires importing `github.com/hajimehoshi/ebiten/v2/text/v2` and loading the debug font face. Either approach is acceptable.

**UI color constants to add:**

- **File**: `pkg/wasmui/types_ui.go` (after the `MessageType.Color()` function at line 39)
- **Constants**: Define named `color.RGBA` variables for UI element colors, drawn from `ASSET_ANALYSIS.md` palette:
  ```
  ColorGold          — #BFA54A — panel headers, important labels
  ColorPanelBorder   — #2E5090 — panel border outlines
  ColorPlayerName    — {100, 200, 100} — player names in initiative/panels
  ColorEnemyName     — {200, 100, 100} — enemy names in initiative list
  ColorPanelHeader   — #BFA54A — section headers (INITIATIVE, CHARACTER, COMBAT LOG, QUESTS, MAP)
  ColorAttributeLabel— {180, 180, 200} — STR/DEX/CON/INT/WIS/CHA labels
  ColorTurnActive    — {80, 220, 80} — "YOUR TURN" indicator
  ColorTurnWaiting   — {180, 140, 60} — "Waiting..." indicator
  ```

**Files to modify and changes per file:**

1. **`pkg/wasmui/game.go`**
   - Add `drawColoredText()` function near line 783 (after `drawLine`).
   - Add a package-level `textBuffer *ebiten.Image` variable and initialize it in `init()`.
   - Update `drawRect` documentation comment to reference the shared rendering pattern if helpful.

2. **`pkg/wasmui/types_ui.go`**
   - Add the named color constants listed above after line 39.

3. **`pkg/wasmui/exploration.go`**
   - `drawCombatLog()` (line 446–472): Replace `ebitenutil.DebugPrintAt(screen, msg.Text, ...)` at line 471 with `drawColoredText(screen, msg.Text, ..., msg.Type.Color())`. Replace the "COMBAT LOG" header at line 453 with `drawColoredText` using `ColorPanelHeader`.
   - `drawCharacterPanel()` (line 240): Replace "CHARACTER" header with `drawColoredText` using `ColorPanelHeader`.
   - `drawPlayerStats()` (line 276): Render player name with `ColorPlayerName`, class/level with `ColorAttributeLabel`.
   - `drawAttributes()` (line 409): Render attribute labels with `ColorAttributeLabel`.
   - `drawMinimap()` (line 346): Render "MAP" header with `ColorPanelHeader`.
   - `drawQuestTracker()` (line 374): Render "QUESTS" header with `ColorPanelHeader`.
   - `drawActiveEffects()` (line 332): Render "Effects:" label with `ColorPanelHeader`.

4. **`pkg/wasmui/combat_screen.go`**
   - `drawInitiativePanel()` (line 260–324): Replace "INITIATIVE" header with `drawColoredText` using `ColorPanelHeader`. Use the `nameColor` variable at line 296–298 (currently suppressed with `_ = nameColor`) by passing it to `drawColoredText` for each initiative entry. Remove the `_ = nameColor` suppression.
   - `drawCombatActionBar()` (line 326): Use the `turnColor` variable at line 341–345 (currently suppressed with `_ = turnColor`) to render "YOUR TURN" / "Waiting..." with `drawColoredText`. Remove the `_ = turnColor` suppression. Render action button labels with appropriate emphasis colors.
   - `drawCombatGrid()` (line 184): Render "Round N" with `ColorPanelHeader`.

5. **`pkg/wasmui/screens.go`** — Update main menu labels and splash text to use `ColorGold` for titles.

6. **`pkg/wasmui/adventure_screen.go`** — Update "ADVENTURE SELECT" header and adventure titles to use `ColorGold` / `ColorPanelHeader`.

7. **`pkg/wasmui/character_creation.go`** — Update step headers and class names to use palette colors.

8. **`pkg/wasmui/overlays.go`** — Update overlay titles ("QUEST LOG", "INVENTORY", "SPELLBOOK", "GUILD") to use `ColorPanelHeader`. Update spell details, item names, and quest titles with appropriate named colors.

**Sprite/asset references**: None required. This is a text-rendering change only. The color palette values come from `ASSET_ANALYSIS.md` (Section "Color Palette": Stone `#5A5A5A`/`#8B8B8B`, Fantasy Blue `#2E5090`/`#4A7DBF`, Gold `#BFA54A`/`#E0C57A`, Medieval Red `#8B2E2E`).

**Constraints and risks**:
- `ebitenutil.DebugPrintAt` uses a fixed 6×16 pixel monospace font. The colored text function must use the same font metrics so existing layouts do not shift.
- The `textBuffer` approach introduces one offscreen image allocation. Size it to `ScreenWidth × 16` (800×16 = 12,800 pixels) — negligible memory impact.
- If using Ebitengine's `text/v2` package, verify it is available in the project's current Ebiten dependency version (v2.9.x should include it). Check `go.mod` for the exact version.
- All 192 `DebugPrintAt` call sites across 9 files should eventually be converted, but the implementation can be phased: start with `drawCombatLog` and `drawInitiativePanel` (highest player-facing impact), then fan out to remaining files.
- Thread safety: `drawColoredText` must not share mutable state across concurrent goroutines. Since `Draw()` in Ebitengine is single-threaded (only called from the main goroutine), the shared `textBuffer` is safe without additional locking.

### Success Criteria

- [ ] A `drawColoredText(screen, text, x, y, color)` function exists in `pkg/wasmui/game.go` and renders text in the specified color
- [ ] `drawCombatLog()` renders each message using `msg.Type.Color()` — combat messages appear purple, warnings yellow, errors red, system messages cyan, info messages light gray
- [ ] The `_ = nameColor` suppression at `combat_screen.go:300` is removed and initiative list entries render player names in green and enemy names in gray
- [ ] The `_ = turnColor` suppression at `combat_screen.go:346` is removed and "YOUR TURN" renders in green while "Waiting..." renders in amber
- [ ] Panel headers ("INITIATIVE", "CHARACTER", "COMBAT LOG", "QUESTS", "MAP") render in gold (`#BFA54A`)
- [ ] Named color constants (`ColorGold`, `ColorPanelBorder`, `ColorPlayerName`, etc.) are defined in `types_ui.go`
- [ ] All existing layout positions remain unchanged — no panel shifts or text alignment regressions
- [ ] `go test ./pkg/wasmui/...` passes with no regressions
- [ ] `GOOS=js GOARCH=wasm go build ./cmd/wasm-ui` compiles without errors

### What Not To Do

- **Do not change the debug font to a custom font.** Stick with Ebitengine's built-in debug font for this improvement. Custom font loading is a separate, larger effort.
- **Do not add new message types.** The five existing types (Info, Warning, Error, Combat, System) are sufficient. Adding types like "DM narration" or "loot" is future work.
- **Do not change server-side message generation.** The server already returns appropriate message strings; this improvement is purely client-side rendering.
- **Do not implement rich text formatting (bold, italic, mixed colors in one line).** Single-color-per-message is the Gold Box standard and the current data model.
- **Do not refactor the screen layout or panel sizes.** This improvement colors existing text in its current position; layout changes are a separate improvement.
- **Do not add the `text/v2` package if it requires an Ebitengine version bump.** Use the `textBuffer` + `ColorScale` approach if the current Ebiten version does not include `text/v2`.

---

## Backlog

### 1. First-Person Dungeon Viewport for Exploration

Replace the flat 2D tile grid in `drawViewport()` (`exploration.go:173–215`) with a simplified first-person perspective rendering walls, doors, and corridors. This is the defining visual identity of Gold Box games, but it requires implementing a raycasting or block-rendering system, loading wall/door sprites, and integrating with the server's map data. Ranked below colored text because it is a much larger implementation effort (estimated 500+ new lines), depends on map data the client does not yet receive, and has higher risk of layout regressions.

### 2. Combat Result Narration in Message Log

Enrich combat feedback so the message log shows Gold Box-style narration: `"Sable attacks Orc — HIT for 7 damage"`, `"Goblin casts Sleep — Sable SAVES"`, `"Orc is slain!"`. The server already returns `Message` strings in `AttackResult` and `CastSpellResult`, but the client does not always propagate them with the right `MessageType`. Ranked below colored text because the messages already appear in the log — they just lack color differentiation, which the recommended improvement fixes first.

### 3. Surfacing Combat Modifiers (Cover, Flanking) in the UI

The game engine calculates cover bonuses (+2/+5/+10 AC) and flanking bonuses (+2 attack) in `combat_modifiers.go`, but these are never displayed. Gold Box games showed positional advantages in combat narration. This requires new RPC response fields and UI indicators. Ranked below colored text because it depends on server-side changes and is not purely a rendering fix.

### 4. Morale State Indicators for Enemies

The morale system (`morale.go`) tracks four states (Steadfast, Shaken, Broken, Panicked) that affect NPC behavior, but no UI element shows enemy morale. Gold Box games showed when enemies broke and fled. Requires adding morale data to the combat state broadcast and rendering morale icons or text in the initiative panel. Ranked below colored text due to cross-cutting server+client changes.

### 5. Panel Border Styling with Gold Box Palette

Replace the single-pixel rectangle outlines (drawn by `drawRectOutline` at `game.go:773`) with thicker, more visible borders using the `ASSET_ANALYSIS.md` palette (deep blue `#2E5090` outer border, medium blue `#4A7DBF` inner highlight). Gold Box panels had bold, bright borders. This is a visual polish item that would benefit from the named color constants introduced by the recommended improvement. Ranked below colored text because text readability has more gameplay impact than border styling.

---

## Preserved: Quality Maintenance Items

The following items from the previous roadmap remain relevant and are not superseded by the recommended improvement:

### Reduce UI Complexity Hotspots

Three functions in `pkg/wasmui/` exceed complexity threshold 15:

- [ ] Refactor `drawCombatGrid` (73 lines, complexity 19.4 in `combat_screen.go`)
- [ ] Simplify `updateCharCreationName` (59 lines, complexity 17.1 in `character_creation.go`)
- [ ] Extract helpers from `Draw` in `adventure_ui.go` (99 lines, complexity 15.8)

### Improve Server Package Test Coverage

- [ ] Raise `pkg/server` coverage from 78.3% to 85%
- [ ] Add edge case tests for session management and complex handlers

### Expand PCG Test Coverage

- [ ] Raise `pkg/pcg` coverage from 78.9% to 85%
- [ ] Add deterministic seeding verification and boundary condition tests
