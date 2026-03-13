# Implementation Plan: Embedded Adventures Content

## Project Context
- **What it does**: A modern Go-based RPG engine inspired by the classic SSI Gold Box series, providing character management, combat systems, and world interactions through JSON-RPC with WebSocket support.
- **Current goal**: Create bundled adventure content — the engine has complete mechanics but ships with no playable adventures
- **Estimated Scope**: Large (>15 items; 10 adventures × ~10 deliverables each)

## Goal-Achievement Status
| Stated Goal | Current Status | This Plan Addresses |
|-------------|---------------|---------------------|
| Core RPG mechanics and character system | ✅ Achieved | No |
| Combat and effect systems | ✅ Achieved | No |
| WebSocket real-time communication | ✅ Achieved | No |
| Procedural Content Generation system | ✅ Achieved | No |
| Circuit breaker patterns and resilience | ✅ Achieved | No |
| Comprehensive input validation | ✅ Achieved | No |
| Health monitoring and metrics | ✅ Achieved | No |
| Asset generation pipeline | ⚠️ Partial (248 placeholders) | No |
| Advanced NPC AI behaviors | ✅ Achieved | No |
| Enhanced combat mechanics | ✅ Achieved | No |
| Complete spell system (levels 0-9) | ✅ Achieved | No |
| World editor tools | ✅ Achieved | No |
| Network optimization (delta compression) | ✅ Achieved | No |
| Content creation utilities | ✅ Achieved | No |
| Player progression persistence | ✅ Achieved | No |
| Guild and faction systems | ✅ Achieved | No |
| **Embedded Adventures** | ❌ Not started | **Yes** |

## Metrics Summary
- **Complexity hotspots on goal-critical paths**: 25 functions above threshold (complexity >10)
  - Key functions: `validateQuest` (15.3), `handleStartCombat` (13.2), `AStarPathfind` (13.7)
- **Duplication ratio**: 2.3% (1,395 lines duplicated) ✅ Acceptable
- **Documentation coverage**: 86.4% ✅ Good
- **Package coupling**: Notable duplication in `pkg/server/server.go` lines 1026-1105 (RPC handler pattern repeated 3x)

## Research Findings
- **Open PRs**: 14 Dependabot PRs pending (Go deps, npm deps, Docker image updates)
- **Go toolchain**: 18 stdlib vulnerabilities require Go 1.24.12+ (documented in CHANGELOG.md)
- **Competitive landscape**: Project fills unique niche as Go-native Gold Box-style engine with WASM frontend
- **Community activity**: No open issues; project appears well-maintained with recent commits

## Implementation Steps

### Step 1: Create Adventure Schema and Directory Structure
- **Deliverable**: `data/adventures/schema.yaml` defining adventure data format; `data/adventures/` directory with subdirectories for maps, items, NPCs
- **Dependencies**: None
- **Goal Impact**: Provides foundation for all adventure content
- **Acceptance**: Schema validates against existing quest/NPC/item structures in `data/`
- **Validation**: `go run ./cmd/pcg-demo/ --validate-schema data/adventures/schema.yaml`

### Step 2: Implement Adventure Loader in pkg/game
- **Deliverable**: `pkg/game/adventure_loader.go` with `LoadAdventure(path string) (*Adventure, error)` function
- **Dependencies**: Step 1 (schema)
- **Goal Impact**: Enables runtime loading of adventure packs
- **Acceptance**: Function parses YAML, returns populated Adventure struct, 80%+ test coverage
- **Validation**: `go test ./pkg/game/... -run TestAdventureLoader -coverprofile=cov.out && go tool cover -func=cov.out | grep adventure_loader`

### Step 3: Add Adventure JSON-RPC Methods
- **Deliverable**: `adventure.list` and `adventure.load` methods in `pkg/server/handlers_adventure.go`
- **Dependencies**: Step 2 (loader)
- **Goal Impact**: Exposes adventures to WASM frontend
- **Acceptance**: Methods registered in `pkg/server/constants.go`, documented in `pkg/README-RPC.md`
- **Validation**: `curl -X POST http://localhost:8080/rpc -d '{"method":"adventure.list","params":{},"id":1}'` returns empty list

### Step 4: Build "The Sunken Sanctum" Reference Adventure
- **Deliverable**: Complete adventure in `data/adventures/sunken-sanctum/` with 5+ maps, 10+ unique items, 15+ NPCs, quest chain
- **Dependencies**: Steps 1-3
- **Goal Impact**: Proves end-to-end adventure system; serves as template for remaining 9
- **Acceptance**: Adventure loads without errors; party can complete main quest line
- **Validation**: `go test ./test/e2e/... -run TestSunkenSanctumSmokeRun -v`

### Step 5: Add Adventure Selection to WASM UI
- **Deliverable**: Adventure browser screen in `pkg/wasmui/adventure_screen.go` integrated into game flow
- **Dependencies**: Steps 3-4
- **Goal Impact**: Users can discover and launch adventures from UI
- **Acceptance**: Screen renders adventure list, selecting one starts new game with that content
- **Validation**: Manual verification via `make wasm && make run` then navigate to adventure selection

### Step 6: Generate Placeholder Assets for Adventures
- **Deliverable**: Extended `scripts/generate-placeholders.sh` to include adventure-specific assets
- **Dependencies**: Step 4 (reference adventure defines asset needs)
- **Goal Impact**: Adventures render with consistent placeholder art
- **Acceptance**: `make assets-verify` reports 100% coverage including adventure assets
- **Validation**: `make assets-verify | grep -E '^(OK|MISSING)'`

### Step 7: Build Adventures 2-5
- **Deliverable**: Four complete adventures:
  - `data/adventures/crimson-coast/` (Slavers theme)
  - `data/adventures/frost-king/` (Undead wilderness)
  - `data/adventures/forbidden-spire/` (Wizard tower)
  - `data/adventures/ember-caverns/` (Underworld mega-dungeon)
- **Dependencies**: Steps 1-4 (established patterns)
- **Goal Impact**: Provides ~16 hours of content
- **Acceptance**: Each adventure loads, validates, passes smoke test
- **Validation**: `go test ./test/e2e/... -run 'TestAdventure(CrimsonCoast|FrostKing|ForbiddenSpire|EmberCaverns)' -v`

### Step 8: Build Adventures 6-10
- **Deliverable**: Five complete adventures:
  - `data/adventures/giant-clans/` (Giant-slaying saga)
  - `data/adventures/emerald-swamp/` (Hex-crawl exploration)
  - `data/adventures/iron-colosseum/` (Gladiatorial arena)
  - `data/adventures/dreaming-pharaoh/` (Trap/puzzle tomb)
  - `data/adventures/void-tyrant/` (Capstone planar campaign)
- **Dependencies**: Step 7 (iteration velocity)
- **Goal Impact**: Completes 30+ hour content library
- **Acceptance**: Each adventure loads, validates, passes smoke test
- **Validation**: `go test ./test/e2e/... -run 'TestAdventure(GiantClans|EmeraldSwamp|IronColosseum|DreamingPharaoh|VoidTyrant)' -v`

### Step 9: Add Adventure Verification Make Target
- **Deliverable**: `make adventures-verify` target in Makefile that validates all 10 adventures
- **Dependencies**: Steps 4, 7, 8
- **Goal Impact**: CI can gate on adventure integrity
- **Acceptance**: Target exits 0 when all adventures valid, non-zero with descriptive errors otherwise
- **Validation**: `make adventures-verify && echo "All adventures valid"`

### Step 10: Update Documentation
- **Deliverable**: Updated README.md with adventure list, ROADMAP.md marked complete, adventure authoring guide
- **Dependencies**: Steps 1-9
- **Goal Impact**: Users and contributors understand adventure system
- **Acceptance**: README roadmap shows "Embedded Adventures ✅", authoring guide explains schema and tooling
- **Validation**: `grep -E '^\- \[x\] .*(Embedded|Adventure)' README.md`

## Complexity Reduction Opportunities (Post-Adventure)
These functions exceed complexity 10 and would benefit from refactoring after adventure delivery:

| Function | File | Complexity | Suggested Action |
|----------|------|------------|------------------|
| `validateQuest` | `cmd/quest-builder/main.go` | 15.3 | Extract validation rules to separate functions |
| `handleStartCombat` | `pkg/server/handlers.go` | 13.2 | Split into setup/validation/execution phases |
| `AStarPathfind` | `pkg/game/pathfinding.go` | 13.7 | Extract heuristic and neighbor functions |
| `Connect` | `pkg/wasmui/rpc_client_wasm.go` | 13.5 | Extract retry logic to utility |
| Server RPC pattern | `pkg/server/server.go:1026-1105` | N/A | Extract common handler wrapper (3x duplication) |

## Dependency Updates (Low Priority)
14 Dependabot PRs are open. Recommend merging after adventure milestone:
- `#14`: Go dependencies (prometheus, testify, x/time)
- `#15`: Dockerfile golang 1.22 → 1.25
- `#9-13`: GitHub Actions (checkout, setup-go, upload-artifact, golangci-lint, attest-build-provenance)
- `#17-20,24`: npm dev dependencies

---

**Generated**: 2026-03-13  
**Analysis Tool**: go-stats-generator v1.0.0  
**Files Analyzed**: 184 Go files, 30,524 lines of code
