# Implementation Gaps — 2026-03-12

## GUI World Editor

- **Stated Goal**: README mentions "World editor tools" in roadmap with ⚠️ indicating partial status, noting "CLI tools only, no GUI editors"
- **Current State**: `cmd/map-editor/` provides CLI-based ASCII map editing with template support. `cmd/quest-builder/` offers interactive quest creation. Both functional but terminal-only.
- **Impact**: Content creators without CLI experience face higher barrier to entry. Visual feedback for map design requires manual JSON inspection or running the game.
- **Closing the Gap**: 
  1. Create `cmd/map-editor-web/` using existing WASM infrastructure from `pkg/wasmui/`
  2. Extend `pkg/server/` with map editor RPC methods (`saveMapDraft`, `loadMapDraft`, `validateMap`)
  3. Add preview rendering using Ebitengine's tile drawing already in `pkg/wasmui/game.go`
  4. Target completion: Browser-based editor accessible at `http://localhost:8080/editor`

---

## Visual Content Creation Tools

- **Stated Goal**: README notes "Content creation utilities" as ⚠️ partial with "CLI tools only, no visual editors"
- **Current State**: `cmd/content-creator/` generates items, NPCs, and encounters via terminal prompts. `cmd/quest-builder/` creates quest YAML files. No visual preview of generated content.
- **Impact**: Designers cannot see visual representation of items/NPCs until they appear in-game. Iteration cycle is slower than visual tools.
- **Closing the Gap**:
  1. Add sprite preview to content-creator by outputting HTML file with embedded images
  2. Create `web/tools/item-preview.html` that renders item sprites from asset pipeline
  3. Integrate with existing asset references in `game-assets.yaml`
  4. Consider future Electron or Tauri wrapper for desktop app experience

---

## Advanced Network Optimization

- **Stated Goal**: README notes "Network optimization" as ⚠️ partial with "basic pooling/rate limiting, no delta compression"
- **Current State**: 
  - Rate limiting implemented in `pkg/server/ratelimit.go` with token bucket algorithm
  - Delta compression **is** implemented in `pkg/server/websocket_delta.go` (contradicts README claim)
  - WebSocket connections use per-connection goroutines with proper cleanup
- **Impact**: README is out-of-date — delta compression exists and shows 95% bandwidth savings per ROADMAP.md benchmarks
- **Closing the Gap**:
  1. Update README.md roadmap to change network optimization from ⚠️ to ✅
  2. Add note: "Delta compression implemented with 95% bandwidth reduction"
  3. Document compression toggle via `GOLDBOX_ENABLE_DELTA_COMPRESSION` env var if configurable

---

## Embedded Adventures (Content Gap)

- **Stated Goal**: Engine provides complete combat, quests, spells, PCG, and editor infrastructure for RPG gameplay
- **Current State**: Engine ships with no bundled adventures. Demo commands generate random content but no curated playable experience exists.
- **Impact**: New users cannot experience the engine's capabilities without creating content first. Reduces immediate engagement and makes capability assessment harder.
- **Closing the Gap** (from existing ROADMAP.md Priority 8):
  1. Create `data/adventures/` directory with adventure YAML schema
  2. Implement adventure loader in `pkg/game/adventure.go`
  3. Add `adventure.list` and `adventure.load` JSON-RPC methods
  4. Build 10 adventures (3-6 hours each) as outlined in ROADMAP.md
  5. Generate placeholder assets via `scripts/generate-placeholders.sh`
  6. Add adventure selection to WASM UI

---

## Documentation Gaps

### README Roadmap Accuracy

- **Stated Goal**: Roadmap should accurately reflect implementation status
- **Current State**: 
  - Guild system marked as "faction generation only, no guild mechanics" but `pkg/game/guild.go` has 686 lines implementing ranks, permissions, treasury, perks
  - Network optimization marked as partial but delta compression is implemented
- **Impact**: Developers may not discover features that exist; potential contributors may duplicate work
- **Closing the Gap**:
  1. Update README.md line 410: Change "⚠️ Guild and faction systems" to "✅ Guild and faction systems with full mechanics (ranks, permissions, treasury, perks, leadership transfer)"
  2. Update README.md line 407: Change network optimization note to acknowledge delta compression exists

### Go Version Documentation

- **Stated Goal**: Consistent Go version requirements across documentation
- **Current State**: 
  - `go.mod` specifies `go 1.24.0` with toolchain `go1.24.2`
  - README badge shows `Go >= 1.23.0`
  - System Overview mentions Go 1.23.0
- **Impact**: Users on Go 1.23.x may encounter issues; CI uses different version than documented
- **Closing the Gap**:
  1. Test project builds with Go 1.23.x to verify backwards compatibility
  2. If compatible: Keep README badge as-is
  3. If not: Update README badge to `go >= 1.24.0`

---

## Test Coverage Gaps

### E2E Test Reliability

- **Stated Goal**: Tests should pass reliably in CI
- **Current State**: E2E tests pass locally but ROADMAP.md mentioned previous failures that were fixed
- **Impact**: Regression risk if E2E tests become flaky
- **Closing the Gap**:
  1. Monitor CI runs for intermittent failures
  2. Add retry logic to flaky network tests
  3. Consider separate E2E test workflow with longer timeouts

---

## Dependency Update Gap

- **Stated Goal**: Keep dependencies current for security and features
- **Current State**: 11 Dependabot PRs open (oldest from 2025-10-29):
  - `#14`: Go dependencies bump (prometheus, testify, time)
  - `#15`: Dockerfile golang 1.22 → 1.25
  - `#17-24`: npm dev dependencies (esbuild, TypeScript, eslint)
- **Impact**: Potential security vulnerabilities in older dependencies; missing features from newer versions
- **Closing the Gap**:
  1. Merge `#14` (Go dependencies) after CI passes
  2. Merge `#15` (Dockerfile) to align with go.mod toolchain
  3. Review npm updates for breaking changes before merging
  4. Set up monthly dependency review cadence

---

## Gap Summary

| Gap | Severity | Effort | Status in README |
|-----|----------|--------|------------------|
| GUI World Editor | Medium | High | Acknowledged (⚠️) |
| Visual Content Tools | Low | Medium | Acknowledged (⚠️) |
| Network Optimization Docs | Low | Trivial | **Outdated** — feature exists |
| Embedded Adventures | Medium | Very High | Not mentioned |
| Roadmap Accuracy | Low | Trivial | **Outdated** |
| Go Version Docs | Low | Trivial | Minor inconsistency |
| Dependency Updates | Low | Low | N/A |

---

## Prioritized Remediation

1. **Immediate (< 1 hour)**: Update README roadmap accuracy — guild system ✅, network compression ✅
2. **Short-term (< 1 week)**: Merge Dependabot PRs for Go dependencies and Dockerfile
3. **Medium-term (< 1 month)**: Design browser-based map editor architecture
4. **Long-term (> 1 month)**: Build first embedded adventure as proof-of-concept
