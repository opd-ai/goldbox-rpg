## Objective

Make **one major improvement** to the `goldbox-rpg` codebase that makes it feel more like the best version of itself. Read the codebase first, then choose the single highest-impact change you can fully implement.

**Execution Mode:** Autonomous — implement the change completely as a pull request with no further user input.

---

## Gold Box Reference Characteristics

The SSI Gold Box series (Pool of Radiance, Curse of the Azure Bonds, Champions of Krynn, etc.) has a highly specific visual and interaction language. All changes must be evaluated against these reference traits:

**Screen Layout**
- The screen is divided into fixed, non-overlapping panels:
  - **Viewport** (top-left, ~40% of screen): first-person 3D corridor view during exploration; bird's-eye tactical grid during combat
  - **Text/Message log** (bottom): scrolling narration, combat results, DM text — *all* game feedback goes here as text
  - **Party roster** (right): vertically stacked list of party members with HP, status icons, and current selection highlight
  - **Command menu** (bottom or right): context-sensitive keyword commands (`Fight`, `Cast`, `Move`, `Look`, `Use`)
- Panels have bold, bright borders to visually separate each zone
- No floating windows — all UI is part of the fixed-panel composition

**Color and Visual Style**
- 16-color EGA palette sensibility: bold, saturated, high-contrast — deep blues, magentas, purples, vivid yellows, dungeon grays
- Dark backgrounds (near-black or deep blue) with bright foreground sprites and text
- Sprites are chunky and flat-colored, clearly readable at small sizes — no gradients, no anti-aliasing
- The color palette in `ASSET_ANALYSIS.md` (stone grays `#5A5A5A`/`#8B8B8B`, fantasy blues `#2E5090`/`#4A7DBF`, medieval reds `#8B2E2E`, gold `#BFA54A`) should be respected throughout

**Combat**
- Top-down tactical grid, tile-based movement
- Active character tile is highlighted; movement range and attack range are visually distinct zones
- Sprites represent character class and monster type (not generic tokens)
- Minimal but meaningful animation: damage flashes, spell effect overlays, hit/miss confirmation
- All combat narration flows into the text log: `"Sable attacks Orc — HIT for 7 damage"`, `"Goblin casts Sleep — FAILED"`

**Exploration**
- First-person, step-and-turn perspective — wireframe or block-filled corridor walls, doors rendered as distinct tiles
- Movement is instant (no smooth scrolling) — each step is a discrete redraw
- Encounters, examinations, and dialogue appear as text overlays or in the message log, sometimes with a static portrait

**Typography and Text**
- Dense, information-rich text panels — the message log is a primary output channel, not a secondary one
- Menu commands use single highlighted letters for keyboard navigation
- All numbers (HP, AC, damage, XP) are shown explicitly — the player always knows the exact state

**Tone**
- Functional austerity: no decorative chrome that doesn't serve gameplay
- Retro authenticity trumps visual modernization — pixel art with strong readability over smooth or polished presentation

---

## What "Best Version of Itself" Means

This is a **retro RPG homage** to the SSI Gold Box series. Every decision should serve two goals, in order:

1. **Visual / experiential quality** — Does it look, feel, and animate better? Does it represent the game world more richly? Does it more faithfully evoke the Gold Box reference characteristics above?
2. **Game system completeness** — Are existing systems (combat, spells, equipment, quests, exploration, character creation, AI, factions, guilds) fully wired into the UI and playable? Are there systems that exist in `pkg/game/` but aren't yet surfaced to the player?

This is a retro game. Authenticity to the Gold Box spirit matters more than modernization. Take every opportunity to use what's already built better — don't add new systems, make the existing ones shine.

---

## Scope of Eligible Improvements

You may work anywhere in the codebase. Non-exhaustive examples across axes:

**Visual richness / representation:**
- Screens that use placeholder fills or flat colors where sprite sheet data is already available
- UI panels, frames, or layouts that don't yet match the Gold Box split-panel convention described above
- Combat, exploration, or character creation screens that could be more detailed or evocative
- Status effects, spell animations, or combat feedback that exist in game logic but have no visual expression

**Game system wiring:**
- Rich systems in `pkg/game/` (factions, morale, guilds, AI behaviors, effect immunities, action points, spell levels) that aren't yet visible or interactive in `pkg/wasmui/`
- Overlays, panels, or screens that exist structurally but don't yet display the full game state available from the server
- RPC methods that exist in `pkg/wasmui/rpc_methods.go` but whose results aren't rendered back to the player

**Animation and feedback:**
- Turn transitions, attack resolution, or spell casting that happen silently with no visual feedback
- Movement, damage, and state change moments that could have frame-based animation using existing sprite sheets
- Combat narration that should be streaming into a visible message log but isn't

---

## Constraints

- **Read before writing.** Examine the relevant source files before making changes.
- **One thing, done deeply.** A single excellent improvement beats three shallow ones.
- **Mark improvements with clear documentation.** Comments should note ``//game improvement #: description
- **Do not break existing systems.** All screens must remain functional. `make build` must pass.
- **Respect the retro aesthetic.** The Gold Box reference characteristics above are defining constraints, not suggestions. Do not modernize the visual style.
- **Maintain test coverage** above current package thresholds (≥60% CI minimum; most packages are at 78–97%).

---

## Success Criteria

- [ ] The chosen area is measurably more complete, detailed, or faithful to Gold Box aesthetics than before.
- [ ] All existing screens and game systems remain functional.
- [ ] `make build` passes without new errors or linter warnings.
- [ ] The change is self-contained and reviewable in a single PR.
- [ ] The PR description explains: (1) what was chosen, (2) why it was the highest-impact option found, and (3) how it serves the Gold Box reference characteristics.
