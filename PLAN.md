# Implementation Plan: Embedded Adventures Content

## Project Context
- **What it does**: A modern Go-based RPG engine inspired by classic SSI Gold Box games, providing turn-based combat, character management, and world interactions through JSON-RPC API with WebSocket support.
- **Current goal**: Create embedded adventure content (10 adventures, 30+ hours gameplay) to showcase the engine's capabilities
- **Estimated Scope**: Large (>15 items across content, infrastructure, and validation)

## Goal-Achievement Status
| Stated Goal | Current Status | This Plan Addresses |
|-------------|---------------|---------------------|
| Core RPG mechanics | ✅ Achieved | No |
| Combat & effects system | ✅ Achieved | No |
| WebSocket real-time | ✅ Achieved | No |
| PCG system | ✅ Achieved | Leveraged |
| Spell system (levels 0-9) | ✅ Achieved | No |
| Guild/faction systems | ✅ Achieved | No |
| Asset generation pipeline | ⚠️ Partial (252/252 placeholders) | Leveraged |
| World editor tools | ⚠️ Partial (CLI only) | No |
| Embedded adventures | ❌ Not started | **Yes** |

## Metrics Summary
- Complexity hotspots on goal-critical paths: 2 functions above threshold (map-editor, quest-builder)
- Duplication ratio: 2.17%
- Documentation coverage: 86.6%
- Package coupling: `pkg/game` is central hub with 39 files, 471 functions
- High complexity functions (cyclomatic > 9): 13 functions, mostly in CLI tools

## Research Findings
- **No open issues**: GitHub issue tracker is empty; no community backlog to prioritize
- **Dependabot PRs**: 11 open PRs for dependency updates (npm, Go, GitHub Actions)
- **Go version**: Project on Go 1.24 with known stdlib vulnerabilities requiring 1.24.12+
- **Competitive context**: No direct competitors in open-source Go RPG engine space
- **Engine maturity**: All core mechanics implemented; missing gameplay content to demonstrate capabilities

## Implementation Steps

### Step 1: Adventure Data Schema & Loader ✅
- **Deliverable**: `pkg/game/adventure.go` with schema validation; `data/adventures/` directory structure
- **Dependencies**: None
- **Goal Impact**: Foundation for all adventure content; enables content validation
- **Acceptance**: Adventure files load without error; schema validation catches malformed YAML
- **Validation**: `go test ./pkg/game/... -run TestAdventure -v`
- **Status**: COMPLETED - Adventure schema, loader, manager, and tests implemented

### Step 2: Adventure List & Load RPC Methods ✅
- **Deliverable**: `adventure.list` and `adventure.load` JSON-RPC methods in `pkg/server/handlers_adventure.go`
- **Dependencies**: Step 1
- **Goal Impact**: API surface for adventure selection; enables frontend integration
- **Acceptance**: RPC methods return valid responses; errors properly handled
- **Validation**: `go test ./pkg/server/... -run TestAdventure -v`
- **Status**: COMPLETED - RPC methods and tests implemented

### Step 3: Adventure Selection UI ✅
- **Deliverable**: Adventure browser screen in `pkg/wasmui/adventure_screen.go`
- **Dependencies**: Step 2
- **Goal Impact**: User-facing adventure selection; completes data→API→UI pipeline
- **Acceptance**: Users can browse and select adventures from WASM frontend
- **Validation**: Manual testing via `make wasm && make run`
- **Status**: COMPLETED - Adventure screen implemented with F1 key access

### Step 4: The Sunken Sanctum (Adventure 1) ✅
- **Deliverable**: Complete adventure pack in `data/adventures/sunken-sanctum/`
  - `adventure.yaml` (encounters, dialogue, objectives)
  - `maps/` (≥5 dungeon maps)
  - `items.yaml` (≥10 unique items)
  - `npcs.yaml` (NPC roster with AI profiles)
- **Dependencies**: Step 1
- **Goal Impact**: Reference implementation; proves adventure system works end-to-end
- **Acceptance**: Adventure loads, validates, and completes smoke test
- **Validation**: `go test ./pkg/game/... -run TestSunkenSanctum`
- **Status**: COMPLETED - Full adventure with 5 maps, 12 items, 10 NPCs, 5 encounters, 3 quests

### Step 5: Adventure Placeholder Assets
- **Deliverable**: Generated placeholder sprites for Adventure 1 via `scripts/generate-adventure-placeholders.sh`
- **Dependencies**: Step 4
- **Goal Impact**: Visual completeness for first adventure; establishes asset pattern
- **Acceptance**: All required sprites exist in `web/static/adventures/sunken-sanctum/`
- **Validation**: `ls web/static/adventures/sunken-sanctum/*.png | wc -l` ≥ 20

### Step 6: Adventures 2-4 (Slavers, Barrow, Spire)
- **Deliverable**: Three adventure packs in `data/adventures/{crimson-coast,frost-barrow,forbidden-spire}/`
- **Dependencies**: Step 4 (pattern established)
- **Goal Impact**: Content expansion; 12-15 hours additional gameplay
- **Acceptance**: Each adventure loads and passes validation
- **Validation**: `go test ./pkg/game/... -run 'TestAdventure/(CrimsonCoast|FrostBarrow|ForbiddenSpire)'`

### Step 7: Adventures 5-7 (Ember Caverns, Giant Clans, Emerald Swamp)
- **Deliverable**: Three adventure packs in `data/adventures/{ember-caverns,giant-clans,emerald-swamp}/`
- **Dependencies**: Step 6
- **Goal Impact**: Mid-tier content; 12-15 hours additional gameplay
- **Acceptance**: Each adventure loads and passes validation
- **Validation**: `go test ./pkg/game/... -run 'TestAdventure/(EmberCaverns|GiantClans|EmeraldSwamp)'`

### Step 8: Adventures 8-10 (Colosseum, Pharaoh, Void Tyrant)
- **Deliverable**: Three adventure packs in `data/adventures/{iron-colosseum,dreaming-pharaoh,void-tyrant}/`
- **Dependencies**: Step 7
- **Goal Impact**: Capstone content; 12-17 hours additional gameplay; completes 10-adventure goal
- **Acceptance**: Each adventure loads and passes validation
- **Validation**: `go test ./pkg/game/... -run 'TestAdventure/(IronColosseum|DreamingPharaoh|VoidTyrant)'`

### Step 9: Adventure Verification Target
- **Deliverable**: `make adventures-verify` Makefile target
- **Dependencies**: Steps 4-8
- **Goal Impact**: CI integration; prevents adventure regressions
- **Acceptance**: Target reports 10/10 adventures valid
- **Validation**: `make adventures-verify` exits 0

### Step 10: Adventure Integration Tests
- **Deliverable**: `test/e2e/adventure_test.go` with smoke tests for each adventure
- **Dependencies**: Steps 1-9
- **Goal Impact**: CI protection; ensures adventures remain playable
- **Acceptance**: All 10 adventures pass load→start→complete cycle
- **Validation**: `go test ./test/e2e/... -run TestAdventure -v`

## Scope Assessment Calibration
| Metric | This Project | Assessment |
|--------|--------------|------------|
| Functions above complexity 9.0 | 13 | Medium |
| Duplication ratio | 2.17% | Small |
| Doc coverage gap | 13.4% | Medium |
| Adventure content items | 10 adventures × 4 files = 40+ | Large |

## Dependencies and Risk Mitigation
- **Risk**: Adventure content creation is labor-intensive
  - **Mitigation**: Use PCG system to generate initial content; refine manually
- **Risk**: Map/NPC balance may require iteration
  - **Mitigation**: Start with Adventure 1 as reference; apply patterns to others
- **Risk**: Asset generation may slow progress
  - **Mitigation**: Use placeholder generation script; defer AI-generated assets

## Notes
- This plan addresses the highest-impact unachieved goal per ROADMAP.md Priority 8
- All core engine features are complete; adventure content is the primary remaining work
- Each step is independently testable with clear validation commands
- Total estimated effort: 80-120 hours across all 10 adventures
- Adventure 1 (Step 4) is the critical path; subsequent adventures parallelize
