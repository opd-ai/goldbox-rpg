# Implementation Gaps — 2026-03-13

## Summary

This document identifies gaps between the GoldBox RPG Engine's stated goals (per README.md) and current implementation, based on functional audit using go-stats-generator and manual code verification.

**Gap Status:**
- ✅ Fully Achieved: 24 goals
- ⚠️ Partial: 3 goals
- ❌ Missing: 0 goals

---

## Active Gaps

### Gap 1: Asset Generation Requires External AI Tool

- **Stated Goal**: README claims "Automated Asset Creation — Complete pipeline for generating 521 game assets" with `make assets` command.
- **Current State**: 
  - Pipeline code is complete and functional (`game-assets.yaml`, `scripts/generate-*.sh`)
  - 252/521 image files exist in `web/static/assets/sprites/`
  - All existing assets are colored rectangle placeholders, not AI-generated art
  - Pipeline requires external AI tool (Stable Diffusion, DALL-E, or similar) not included in repository
- **Impact**: 
  - Game is fully playable with placeholders — visual experience degraded but functional
  - Users expecting production-ready art must configure external AI tool (4-6 hour setup + generation)
  - README accurately notes this limitation with ⚠️ warning
- **Closing the Gap**:
  1. Install external AI image generation tool per `ASSET_INTEGRATION.md`
  2. Configure API keys and model settings
  3. Run `make assets` (4-6 hours for 521 assets)
  4. Verify with `make assets-verify` expecting 521/521
  5. Alternative: Provide pre-generated asset packs via GitHub Releases
  6. **Validation:** `find web/static/assets -name "*.png" | wc -l` returns 521

---

### Gap 2: GUI World Editor Not Implemented

- **Stated Goal**: README roadmap notes "World editor tools" with ⚠️ warning indicating partial implementation.
- **Current State**: 
  - CLI tools exist and are functional:
    - `cmd/map-editor/` — Terminal-based map creation (64.6% test coverage)
    - `cmd/quest-builder/` — Quest chain builder with validation (71.6% coverage)
    - `cmd/content-creator/` — Content generation tool (60.8% coverage)
  - WebSocket editor protocol exists (`pkg/server/websocket_editor.go`)
  - WASM editor foundation exists (`pkg/wasmui/editor.go`)
  - No browser-based graphical editors completed
- **Impact**: 
  - Content creators must use command-line interfaces
  - Higher barrier to entry for non-technical users
  - Existing infrastructure provides ~70% foundation for GUI implementation
- **Closing the Gap**:
  1. Extend `pkg/wasmui/editor.go` with visual tile placement using Ebitengine
  2. Connect WebSocket editor protocol to canvas for real-time preview
  3. Add visual quest chain builder using existing quest schema
  4. Add save/load workflow for editor state
  5. **Validation:** User can create and save a map at `http://localhost:8080/editor` without CLI

---

### Gap 3: Content Creation Utilities CLI-Only

- **Stated Goal**: README roadmap lists "Content creation utilities" with ⚠️ warning.
- **Current State**: 
  - `cmd/content-creator/` provides CLI for generating game content
  - `cmd/quest-builder/` enables quest chain creation via terminal
  - `pkg/cliutil/preview.go` implements live preview server for content
  - No integrated visual content authoring environment
- **Impact**: 
  - Adventure authors need command-line proficiency
  - Content iteration requires manual file editing and CLI validation
  - Live preview exists but requires CLI to trigger updates
- **Closing the Gap**:
  1. Build on `pkg/cliutil/preview.go` WebSocket server for real-time editing
  2. Create web UI that connects to preview server
  3. Add drag-and-drop content placement in browser
  4. **Validation:** Content can be created and previewed entirely in browser

---

## Infrastructure Considerations

### Server Package Test Coverage

- **Stated Goal**: Project targets ≥60% coverage (achieved at 80.3% overall), but core server infrastructure should match game package standards.
- **Current State**: 
  - `pkg/server` at 70.6% coverage
  - `pkg/game` at 87.8% coverage (exemplary)
  - Session timeout and WebSocket error recovery paths undertested
- **Impact**: 
  - Production server code has less test coverage than game mechanics
  - Edge cases in session management may contain latent bugs
- **Closing the Gap**:
  1. Add tests for `pkg/server/session.go` timeout cleanup paths
  2. Add WebSocket reconnection and error recovery tests in `pkg/server/websocket_nhooyr.go`
  3. Add boundary tests for RPC parameter validation edge cases
  4. **Validation:** `go test ./pkg/server/... -cover` shows ≥80%

---

### Code Duplication (1.51% ratio)

- **Stated Goal**: Maintainable codebase following Go best practices.
- **Current State**: 
  - 938 duplicated lines across codebase
  - Largest duplication in RPC handler registration (`pkg/server/server.go:1026` — 81 lines)
  - Primary pattern: repetitive handler registration calls
- **Impact**: 
  - Maintenance burden when updating duplicated logic
  - Risk of drift between clone instances
- **Closing the Gap**:
  1. Extract `registerMethodHandlers` into table-driven pattern accepting `map[string]HandlerFunc`
  2. Use reflection or code generation for repetitive RPC dispatch
  3. Target <500 duplicated lines (0.8% ratio)
  4. **Validation:** `go-stats-generator analyze . --sections duplication | grep "Duplicated Lines"` shows <500

---

### CLI Tool Test Coverage Variance

- **Stated Goal**: README states "Include tests for new features" in development guidelines.
- **Current State**: 
  | Package | Coverage |
  |---------|----------|
  | `cmd/map-editor` | 64.6% |
  | `cmd/content-creator` | 60.8% |
  | `cmd/quest-builder` | 71.6% |
- **Impact**: 
  - Content creation tools have lower coverage than core packages
  - Reduced confidence in CLI tool reliability for content creators
- **Closing the Gap**:
  1. Add unit tests for command parsing in `cmd/map-editor/main.go`
  2. Add integration tests for `cmd/content-creator` output validation
  3. Add error path tests for all CLI tools
  4. **Validation:** `go test ./cmd/map-editor/... ./cmd/content-creator/... ./cmd/quest-builder/... -cover` shows ≥70% each

---

## Resolved Since Prior State

| Previous Gap | Resolution | Evidence |
|--------------|------------|----------|
| gorilla/websocket in production | Migrated to nhooyr.io/websocket | `pkg/server/websocket_nhooyr.go`, `go.mod:23` — gorilla only in tests |
| Spell system incomplete (0-2 only) | Levels 0-9 implemented | `data/spells/` (10 files: cantrips.yaml, level1-9.yaml) |
| Guild mechanics missing | Full implementation | `pkg/game/guild.go` (685 lines) |
| Network delta compression | Implemented | `pkg/server/websocket_delta.go` |
| Adventures incomplete | 10 adventures complete | `data/adventures/` (10 directories) |
| Test coverage below 60% | Coverage at 80.3% | `go test ./... -cover` |
| Go version outdated | Updated to Go 1.25.6 | `go.mod:3` |

---

## Positive Findings (No Gaps)

The following claimed features are fully implemented with no gaps identified:

1. **Character System**: 6 attributes, 6 classes, 4 creation methods — `pkg/game/character.go`, `pkg/game/character_creation.go`
2. **Combat System**: Turn-based, action points, opportunity attacks — `pkg/game/combat_opportunity.go`
3. **Effect System**: 5 DoT types, stacking, immunity — `pkg/game/effectmanager.go`
4. **Spatial Indexing**: R-tree-like structure for efficient queries — `pkg/game/spatial_index.go`
5. **A* Pathfinding**: Full implementation with priority queue — `pkg/game/pathfinding.go`
6. **AI Behaviors**: Tactical combat AI, behavior trees — `pkg/game/ai_combat.go`, `pkg/game/ai_behaviors.go`
7. **PCG System**: Terrain, items, quests, NPCs with deterministic seeding — `pkg/pcg/`
8. **Resilience Patterns**: Circuit breakers, retry with backoff — `pkg/resilience/`, `pkg/retry/`
9. **Input Validation**: 72 RPC methods validated — `pkg/validation/`
10. **Health Monitoring**: `/health`, `/ready`, `/live`, `/metrics` — `pkg/server/health.go`, `pkg/server/metrics.go`
11. **WebSocket Communication**: nhooyr.io/websocket with delta compression — `pkg/server/websocket_nhooyr.go`
12. **Guild System**: 5 ranks, treasury, diplomacy — `pkg/game/guild.go`
13. **Faction System**: Relations, war/peace, alliances — `pkg/game/faction_relations.go`
14. **Persistence**: File-based save/load — `pkg/persistence/`
15. **Embedded Adventures**: 10 complete adventure packs — `data/adventures/`
16. **Spell System**: 60 spells across levels 0-9 — `data/spells/`

---

## Verification Commands

```bash
# Verify gap status
find web/static/assets/sprites -name "*.png" | wc -l
# Expected: 252 (gap: should be 521)

# Verify CLI tools exist
ls cmd/map-editor/main.go cmd/quest-builder/main.go cmd/content-creator/main.go
# Expected: All files exist

# Check server coverage
go test ./pkg/server/... -cover | grep "coverage:"
# Expected: ~70.6% (gap: should be ≥80%)

# Check CLI tool coverage
go test ./cmd/map-editor/... ./cmd/quest-builder/... ./cmd/content-creator/... -cover

# Verify overall health
go test -race ./...
go vet ./...
go-stats-generator analyze . --skip-tests

# Verify no blocking issues for stated goals
ls data/adventures/*/adventure.yaml | wc -l    # Expected: 10
cat data/spells/*.yaml | grep -c "spell_id:"   # Expected: 60
curl http://localhost:8080/health              # Expected: 200 OK (when running)
```

---

*Last Updated: 2026-03-13*
*Based on go-stats-generator v1.0.0 analysis*
