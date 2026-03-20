## Objective
Analyze the `goldbox-rpg` codebase and produce an updated `ROADMAP.md` that comprehensively catalogs every meaningful improvement opportunity — evaluated against a retro Gold Box–inspired aesthetic combined with enhanced environmental detail, modern UX improvements, and overall gameplay quality — with each item specified in enough detail for autonomous implementation. **Execution Mode:** Report generation — produce a revised `ROADMAP.md` file as output. Do not implement any changes.
---
## Gold Box Reference Characteristics
The SSI Gold Box series (Pool of Radiance, Curse of the Azure Bonds, Champions of Krynn, etc.) has a distinctive visual and interaction language that serves as a stylistic foundation. Improvement candidates should be evaluated against these reference traits while allowing enhancements that improve gameplay experience, accessibility, and environmental detail: **Screen Layout** - Fixed, non-overlapping panels:
  - **Viewport** (top-left, ~40% of screen): first-person 3D corridor view during exploration; bird's-eye tactical grid during combat - **Text/Message log** (bottom): 
  scrolling narration, combat results, DM text — *all* game feedback goes here as text - **Party roster** (right): vertically stacked list of party members with HP, status 
  icons, and current selection highlight - **Command menu** (bottom or right): context-sensitive keyword commands (`Fight`, `Cast`, `Move`, `Look`, `Use`)
- Panels have bold, bright borders to visually separate each zone - No floating windows — all UI is part of the fixed-panel composition **Color and Visual Style** - 16-color 
EGA palette sensibility: bold, saturated, high-contrast — deep blues, magentas, purples, vivid yellows, dungeon grays - Dark backgrounds (near-black or deep blue) with bright 
foreground sprites and text - Sprites are chunky and flat-colored, clearly readable at small sizes — no gradients, no anti-aliasing - The color palette in `ASSET_ANALYSIS.md` 
(stone grays `#5A5A5A`/`#8B8B8B`, fantasy blues `#2E5090`/`#4A7DBF`, medieval reds `#8B2E2E`, gold `#BFA54A`) must be respected throughout **Combat** - Top-down tactical 
grid, tile-based movement - Active character tile is highlighted; movement range and attack range are visually distinct zones - Sprites represent character class and monster 
type — not generic tokens - Minimal but meaningful animation: damage flashes, spell effect overlays, hit/miss confirmation - All combat narration flows into the text log: 
`"Sable attacks Orc — HIT for 7 damage"`, `"Goblin casts Sleep — FAILED"` **Exploration** - First-person, step-and-turn perspective — walls may use textured or block-filled rendering with environmental detail beyond simple wireframes (e.g., stone textures, themed decorations, atmospheric effects via PCG) - Movement uses pseudo-smooth scrolling (brief interpolated transitions between discrete grid positions) to maintain the tile-based feel while improving visual fluidity - Encounters, examinations, and dialogue appear as text 
overlays or in the message log, sometimes with a static portrait **Typography and Text** - Dense, information-rich text panels — the message log is a primary output channel, 
not secondary - Menu commands use single highlighted letters for keyboard navigation - All numbers (HP, AC, damage, XP) are shown explicitly — the player always knows the 
exact state **Tone** - Functional austerity: no decorative chrome that doesn't serve gameplay - Retro Gold Box aesthetic as a stylistic foundation, combined with enhanced environmental detail and UX improvements (e.g., touch/mobile support, responsive input) where they serve gameplay — pixel art with strong readability remains the priority ---
## Analysis Instructions
Read the codebase thoroughly before writing. Pay particular attention to: - `pkg/wasmui/` — all screen and rendering code (combat, exploration, adventure, overlays, character 
creation) - `pkg/game/` — all game systems that may or may not be surfaced to the player - `pkg/wasmui/rpc_methods.go` — RPC calls available to the frontend - 
`ASSET_ANALYSIS.md` — existing sprite assets and their organization - The existing `ROADMAP.md` — retain any still-relevant content; this is an update, not a replacement For 
**every** file and system examined, ask: - Does the current rendering match Gold Box panel layout and visual conventions? - Are game systems implemented in `pkg/game/` but 
absent from the UI? - Are sprite assets available but unused or fallen back to placeholder fills? - Is the message log receiving all feedback it should? - Are there animation 
or visual feedback moments that are silent when they shouldn't be? Evaluate every candidate across these axes, in priority order:

1. **Visual / Retro-inspired fidelity** — layout correctness, palette adherence, sprite utilization, message log usage, animation, environmental detail, and PCG-driven content
2. **UX and accessibility** — touch/mobile support, responsive input, pseudo-smooth transitions, and usability improvements that enhance the gameplay experience
3. **Game system completeness** — systems in `pkg/game/` (factions, morale, guilds, AI behaviors, effect immunities, action points, spell levels, equipment) not yet visible or interactive in the UI For each candidate found, assess and record: - Current state 
(what exists today) - The gap relative to Gold Box reference behavior - Files and functions involved - Implementation complexity (Small / Medium / Large) - Dependencies on 
other items in the list ---
## Output Format
Produce a complete replacement `ROADMAP.md` with the following structure: ```markdown
# Roadmap
Generated: {date}
## Gold Box Reference Standard
[One paragraph summarizing what "Gold Box–inspired" means for this codebase specifically: a retro aesthetic foundation combined with enhanced environmental detail, modern UX improvements, and PCG-driven content, grounded in the reference characteristics and the current state of the UI.] ---
## Improvement Items
Items are grouped by theme and ordered within each group by priority (highest impact first). Each item is specified in enough detail for autonomous implementation. ---
### Group: {Group Name}
<!-- Groups should reflect natural themes, e.g.: Combat Screen, Exploration Screen, Character & Party Display, Game System Wiring, Animation & Feedback, Message Log, UI 
     Layout & Panels -->
#### {N}. {Item Title}
**Priority:** High / Medium / Low **Complexity:** Small / Medium / Large **Depends on:** {item numbers this should follow, or "None"} **Current State** [What exists today — 
specific files, functions, and observed behavior.] **Gap** [What is missing or wrong relative to Gold Box reference behavior. Cite the specific reference characteristic being 
violated.] **Implementation Specification** [Enough detail for autonomous implementation without further clarification:] - Files to modify - Functions to add or refactor - 
Behavioral description of the change - Sprite/asset references to use (from ASSET_ANALYSIS.md where applicable) - Any constraints or risks **Success Criteria** - [ ] 
Criterion 1 - [ ] Criterion 2 --- <!-- Repeat for every item found. Do not omit items because they seem minor.
     A Low priority item documented well is more useful than an undocumented gap. --> ---
## Preserved: Quality Maintenance Items
[Carry forward any still-relevant non-visual items from the existing ROADMAP.md — e.g., test coverage targets, complexity refactoring, duplication reduction. Do not repeat 
items already completed. Use the same item format above.] ---
## Implementation Order
[A recommended sequencing of ALL items above, accounting for dependencies and risk. Format as a simple numbered list: "{N}. {Item Title} — {one-sentence rationale for its 
position}"] ---
## Completion Criteria
[Define what "done" looks like for the roadmap as a whole — i.e., what state the game will be in when all items above are implemented. Describe it in terms of the retro-inspired design standard with enhanced UX and environmental detail, not a checklist.] ``` ---
## Constraints
- **Do not implement any code.** Output is documentation only. - **Document every candidate found.** Do not filter items out because they seem small or hard — include them 
all. An incomplete roadmap is worse than a long one. - **Every item gets a full specification.** The implementation specification for each item must be detailed enough that 
an agent could execute it autonomously without asking follow-up questions. - **Ground every claim in the codebase.** Do not speculate — read the files and cite specific 
functions, files, and observed behavior. - **Embrace retro-inspired design with enhancements.** Items should use the Gold Box aesthetic as a stylistic foundation while welcoming UX improvements (touch/mobile, smooth transitions), enhanced environmental detail, and PCG-driven content that improve the gameplay experience.
- **Preserve maintenance items.** Existing roadmap content that is still relevant and not yet complete must be carried forward, not discarded.
