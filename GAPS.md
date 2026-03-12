# Implementation Gaps — 2026-03-12

## E2E Test Suite API Mismatch

- **Stated Goal**: README claims "Comprehensive input validation" and working E2E tests (CI badge shows passing).
- **Current State**: 33 of 68 E2E subtests fail due to API contract mismatches between tests and handlers.
  - Equipment tests pass string item names (`"longsword"`, `"chainmail"`) but handlers expect UUIDs
  - PCG tests pass `type` field but handlers expect `content_type` and require `location_id`
  - Character tests expect `player.character` in GetGameState but it returns only session data
  - Attack/combat tests depend on character creation working correctly
- **Impact**: CI pipeline shows FAIL for E2E tests. New contributors cannot verify their changes work correctly. Production deployment confidence is undermined.
- **Closing the Gap**:
  1. Update `test/e2e/inventory_test.go` to use valid UUIDs by first creating items via RPC or using fixture UUIDs
  2. Update `test/e2e/pcg_test.go` to use correct parameter names (`content_type` not `type`) and include `location_id`
  3. Modify `pkg/server/state.go:GetState()` to include character data from session player object
  4. Alternatively, add item name resolution in handlers to convert friendly names to UUIDs
  5. **Validation**: `go test ./test/e2e/... -v 2>&1 | grep -E '^(ok|FAIL)' | grep FAIL` should return nothing

## Asset Generation Completion

- **Stated Goal**: README claims "Asset generation pipeline with 521 defined assets" and "Full asset generation (~4-6 hours)".
- **Current State**: Only 4 placeholder PNG files exist in `web/static/assets/sprites/`:
  - `characters.png` (245 bytes, placeholder)
  - `effects.png` (245 bytes, placeholder)
  - `terrain.png` (245 bytes, placeholder)
  - `ui.png` (245 bytes, placeholder)
  - 1 legacy JPEG (`terrain.jpg`, 5.3KB)
  - 1 reference image (`mindmaze1-1024x701.jpg`, 115KB)
  - Total: 4 actual game assets out of 521 defined (0.8%)
- **Impact**: Users cannot run the game with full visual fidelity. New contributors must read ASSET_INTEGRATION.md and set up external AI tools (Stable Diffusion, DALL-E) before generating assets. This is a 4-6 hour barrier to entry.
- **Closing the Gap**:
  1. Create `scripts/generate-placeholders.sh` that generates all 521 assets as simple colored rectangles with text labels using ImageMagick (no external AI required)
  2. Add `make assets-placeholders` target that runs in <5 minutes
  3. Include pre-generated placeholder pack in releases or provide download link
  4. Document that placeholders are sufficient for development/testing
  5. **Validation**: `find web/static/assets -type f \( -name "*.png" -o -name "*.svg" \) | wc -l` shows 521+

## Test Coverage Below CI Threshold

- **Stated Goal**: README badge shows "78% coverage" and CI enforces this threshold.
- **Current State**: Actual coverage is 74.8% (3.2 percentage points below threshold).
  - Low-coverage packages:
    - `pkg/persistence/` - cohesion score 1.1, core persistence functions
    - `pkg/secrets/` - cohesion score 0.8, 12 functions
    - `pkg/integration/` - cohesion score 1.4, 13 functions
  - Many exported functions have 0% coverage
- **Impact**: CI would fail if coverage check is enforced. Regressions in untested code go undetected.
- **Closing the Gap**:
  1. Add unit tests for `pkg/persistence/store.go` covering save, load, atomic operations
  2. Add unit tests for `pkg/secrets/` provider implementations
  3. Add integration tests for `pkg/integration/` utilities
  4. Target: 2-3% coverage increase from these 3 packages
  5. **Validation**: `go test ./... -coverprofile=c.out && go tool cover -func=c.out | grep total` shows ≥78%

## README Roadmap Accuracy

- **Stated Goal**: README roadmap reflects actual implementation status.
- **Current State**: Two roadmap items are marked incomplete but are actually implemented:
  - Line 405: "⚠️ Additional spell effects (cantrips + levels 1-2 only, levels 3-9 needed)"
    - Reality: `data/spells/` contains 11 YAML files with levels 0-9 (60 total spells)
  - Line 410: "⚠️ Guild and faction systems (faction generation only, no guild mechanics)"
    - Reality: `pkg/game/guild.go` (686 lines) implements full guild system with 5 ranks, permissions, treasury, leveling, perks, leadership transfer
- **Impact**: Users and contributors underestimate project completeness. May lead to duplicate implementation efforts.
- **Closing the Gap**:
  1. Update README.md line 405: `- [x] Complete spell system (levels 0-9, 60 spells)`
  2. Update README.md line 410: `- [x] Guild and faction systems with full mechanics`
  3. Review other roadmap items for similar discrepancies
  4. **Validation**: `grep -E '✅.*spell|✅.*Guild' README.md` shows updated lines

## Network Delta Compression

- **Stated Goal**: README roadmap notes "⚠️ Network optimization (basic pooling/rate limiting, no delta compression)".
- **Current State**: This is accurately marked as incomplete.
  - `pkg/server/ratelimit.go` implements token bucket rate limiting
  - `pkg/server/websocket.go` sends full state on each broadcast
  - No state diffing or delta compression implemented
- **Impact**: High bandwidth usage for real-time gameplay. Mobile users and players with slow connections experience latency.
- **Closing the Gap**:
  1. Add `LastState` map per WebSocket connection in `pkg/server/websocket.go`
  2. Implement `calculateDelta(old, new map[string]interface{})` function
  3. Enable `permessage-deflate` compression on WebSocket upgrader
  4. Add benchmark tests comparing full state vs delta message sizes
  5. Target: ≥50% reduction in typical state update message size
  6. **Validation**: `go test ./pkg/server/... -bench=BenchmarkWebSocketDelta -benchmem` shows improvement

## PCG RPC Handler Parameter Inconsistency

- **Stated Goal**: README claims "Procedural Content Generation system" with "Deterministic seeding for reproducible content".
- **Current State**: PCG works internally, but RPC handlers have stricter requirements than tests expect:
  - `handleGenerateContent` requires both `content_type` and `location_id`
  - `handleRegenerateTerrain` requires `session_id`, `biome`, and has defaults for size
  - E2E tests don't provide these required parameters
- **Impact**: PCG features appear broken to users testing via RPC. The 10 PCG-related E2E tests all fail.
- **Closing the Gap**:
  1. Option A: Relax handler validation to make `location_id` optional (generate random ID)
  2. Option B: Update E2E tests to provide all required parameters
  3. Add documentation for PCG RPC methods in `pkg/README-RPC.md`
  4. **Validation**: `go test ./test/e2e/... -run 'TestGenerate|TestTerrain|TestPCG' -v` shows all PASS

## World Editor: CLI Only

- **Stated Goal**: README roadmap notes "⚠️ World editor tools (CLI tools only, no GUI editors)".
- **Current State**: This is accurately marked as incomplete.
  - `cmd/map-editor/main.go` - CLI interactive map editor
  - `cmd/quest-builder/main.go` - CLI quest creation tool
  - `cmd/content-creator/main.go` - CLI content generation tool
  - No browser-based or GUI editors exist
- **Impact**: Content creators must learn CLI tools. Higher barrier for non-technical users.
- **Closing the Gap**:
  1. Leverage existing WASM frontend infrastructure in `pkg/wasmui/`
  2. Add browser-based map editor mode that reuses `cmd/map-editor` logic
  3. Add WebSocket-based live preview for content creation
  4. Target: Basic drag-and-drop map editing in browser
  5. **Validation**: Browser navigation to `/editor` shows functional map editor

---

## Summary

| Gap | Severity | Effort to Close |
|-----|----------|-----------------|
| E2E Test Suite API Mismatch | Critical | Medium (1-2 days) |
| Asset Generation Completion | High | Low (create placeholder script) |
| Test Coverage Below CI Threshold | High | Medium (add ~50 test cases) |
| README Roadmap Accuracy | Medium | Low (documentation update) |
| Network Delta Compression | Medium | High (new feature) |
| PCG Handler Parameter Inconsistency | Medium | Low (fix tests or handlers) |
| World Editor GUI | Low | High (new feature) |

*Generated from functional audit comparing README.md stated goals against actual implementation.*
