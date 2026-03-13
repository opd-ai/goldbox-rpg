# Implementation Gaps — 2026-03-13

## Summary

This document identifies gaps between the GoldBox RPG Engine's stated goals and current implementation, based on functional audit against README claims and codebase analysis using go-stats-generator.

**Gap Status:**
- ✅ Fully Achieved: 27 goals
- ⚠️ Partial: 3 goals
- ❌ Missing: 0 goals

---

## Active Gaps

### Asset Generation Requires External AI Tool

- **Stated Goal**: README claims "Automated Asset Creation — Complete pipeline for generating 521 game assets" with `make assets` command.
- **Current State**: 
  - Pipeline code is complete and functional (`game-assets.yaml`, `scripts/generate-*.sh`)
  - 513/521 image files exist in `web/static/assets/sprites/`
  - All existing assets are colored rectangle placeholders, not AI-generated art
  - `make assets-verify` passes with "PARTIAL" status
  - Pipeline requires external AI tool (Stable Diffusion, DALL-E, or similar)
- **Impact**: 
  - Game is fully playable with placeholders — visual experience degraded but functional
  - Users expecting production-ready art must configure external AI tool
  - Full generation takes 4-6 hours with proper setup
- **Closing the Gap**:
  1. Install external AI image generation tool per `ASSET_INTEGRATION.md`
  2. Configure API keys and model settings
  3. Run `make assets` (4-6 hours for 521 assets)
  4. Verify with `make assets-verify` expecting 521/521
  5. Alternative: Contact maintainer for pre-generated asset packs
  6. **Validation:** `find web/static/assets -name "*.png" | wc -l` returns 521; `make assets-verify` reports 521/521

---

### GUI World Editor Not Implemented

- **Stated Goal**: README roadmap notes "World editor tools" with ⚠️ warning indicating partial implementation.
- **Current State**: 
  - CLI tools exist and are functional:
    - `cmd/map-editor/` — Terminal-based map creation (64.6% test coverage)
    - `cmd/quest-builder/` — Quest chain builder with validation (27.6% coverage)
    - `cmd/content-creator/` — Content generation tool (37.0% coverage)
  - WebSocket editor protocol exists (`pkg/server/websocket_editor.go`)
  - WASM editor foundation exists (`pkg/wasmui/editor.go`)
  - No browser-based graphical editors completed
- **Impact**: 
  - Content creators must use command-line interfaces
  - Higher barrier to entry for non-technical users
  - Existing infrastructure provides 70%+ foundation for GUI
- **Closing the Gap**:
  1. Extend `pkg/wasmui/editor.go` with visual tile placement using Ebitengine
  2. Connect WebSocket editor protocol to canvas for real-time preview
  3. Add visual quest chain builder using existing quest schema
  4. Add save/load workflow for editor state
  5. **Validation:** User can create and save a map at `http://localhost:8080/editor` without CLI

---

### Content Creation Utilities CLI-Only

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

### CLI Tool Test Coverage Below Standards

- **Stated Goal**: README states "Include tests for new features" in development guidelines.
- **Current State**: 
  | Package | Coverage | Target |
  |---------|----------|--------|
  | `pkg/cliutil` | 12.2% | 70% |
  | `cmd/quest-builder` | 27.6% | 60% |
  | `cmd/content-creator` | 37.0% | 60% |
- **Impact**: 
  - Reduced confidence in CLI tool reliability
  - Bug risk for content creators using these tools
  - Lower threshold for regressions
- **Closing the Gap**:
  1. Add unit tests for `pkg/cliutil/preview.go` WebSocket handlers
  2. Add integration tests for quest-builder validation workflows
  3. Add error path tests for content-creator
  4. **Validation:** `go test ./pkg/cliutil/... ./cmd/quest-builder/... ./cmd/content-creator/... -cover` shows ≥60% for each

---

### Server Package Coverage Below 80%

- **Stated Goal**: Project targets ≥60% coverage (achieved at 79.1% overall), but core infrastructure should exceed this.
- **Current State**: 
  - `pkg/server` at 70.5% coverage
  - `pkg/game` at 87.8% coverage (exemplary)
  - Session timeout and WebSocket error recovery paths undertested
- **Impact**: 
  - Production server code has less test coverage than game mechanics
  - Edge cases in session management may contain latent bugs
- **Closing the Gap**:
  1. Add tests for session timeout cleanup paths
  2. Add WebSocket reconnection and error recovery tests
  3. Add boundary tests for RPC parameter validation
  4. **Validation:** `go test ./pkg/server/... -cover` shows ≥80%

---

## Code Quality Observations

### Code Duplication (1.52% ratio)

- **Stated Goal**: Maintainable codebase following Go best practices.
- **Current State**: 
  - 37 clone pairs detected (938 duplicated lines)
  - Largest clone: 35 lines in `pkg/server/server.go`
  - Primary duplication in RPC handler registration patterns
- **Impact**: 
  - Maintenance burden when updating duplicated logic
  - Risk of drift between clone instances
  - Increased codebase size
- **Closing the Gap**:
  1. Extract handler registration into helper accepting `map[string]HandlerFunc`
  2. Use table-driven patterns for repetitive RPC dispatch
  3. Target <500 duplicated lines (0.8% ratio)
  4. **Validation:** `go-stats-generator analyze . --sections duplication | grep "Duplicated Lines"` shows <500

---

### Unreferenced Functions (252 detected)

- **Stated Goal**: Clean codebase without dead code.
- **Current State**: 
  - 252 functions detected as unreferenced by go-stats-generator
  - Many are exported constants/types that form public API
  - Some may be genuinely unused code
- **Impact**: 
  - Increased codebase size
  - Confusion about which APIs are actively used
  - Potential for stale code to accumulate
- **Closing the Gap**:
  1. Review each unreferenced function for usage
  2. Remove genuinely dead code
  3. Add `//nolint:unused` with justification for intentional public API
  4. **Validation:** `go-stats-generator analyze . | grep "Dead Code"` shows <50 unreferenced after cleanup

---

## Resolved Since Prior Audits

| Gap | Resolution | Evidence |
|-----|------------|----------|
| gorilla/websocket in production | Migrated to nhooyr.io/websocket | `pkg/server/websocket_nhooyr.go` |
| Spell system incomplete (0-2) | Levels 3-9 implemented | `data/spells/level*.yaml` (11 files) |
| Guild mechanics missing | Full implementation | `pkg/game/guild.go` (685 lines) |
| Network delta compression | Implemented | `pkg/server/websocket_delta.go` |
| Adventures incomplete | 10 adventures complete | `data/adventures/` (10 directories) |
| Coverage below 60% | Coverage at 79.1% | CI pipeline passing |
| Go version security | Updated to Go 1.25.6 | `go.mod:3` |

---

## Verification Commands

```bash
# Verify gap status
make assets-verify                              # Asset pipeline status
go test ./pkg/cliutil/... -cover               # CLI util coverage
go test ./cmd/quest-builder/... -cover         # Quest builder coverage
go test ./cmd/content-creator/... -cover       # Content creator coverage
go test ./pkg/server/... -cover                # Server coverage

# Check overall health
go test -race ./...
go vet ./...
go-stats-generator analyze . --skip-tests

# Verify no blocking issues
make adventures-verify                          # 10/10 adventures valid
curl http://localhost:8080/health              # Health check passes
```

---

*Last Updated: 2026-03-13*
*Based on go-stats-generator v1.0.0 analysis*
