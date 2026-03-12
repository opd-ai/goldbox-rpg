# Implementation Gaps — 2026-03-12

This document identifies gaps between the GoldBox RPG Engine's stated goals and current implementation, with actionable steps to close each gap.

---

## Asset Generation Automation

- **Stated Goal**: README claims "complete pipeline for generating 521 game assets" with "Automated Asset Creation"
- **Current State**: Pipeline defined in `game-assets.yaml` has 248 asset definitions. 252 placeholder files exist in `web/static/assets/sprites/`. Generation requires manual external AI tool setup (Stable Diffusion or DALL-E) with 4-6 hour processing time.
- **Impact**: New users cannot run `make assets` out of the box. Development works with placeholders, but production visuals require significant manual setup effort documented in `ASSET_INTEGRATION.md`.
- **Closing the Gap**:
  1. Update README to clarify "248 assets defined" instead of "521 assets"
  2. Consider bundling pre-generated placeholder pack in releases
  3. Add automated fallback that generates colored rectangles for missing assets
  4. Document exact Stable Diffusion API configuration in CI-friendly format

---

## GUI World Editor Tools

- **Stated Goal**: README Roadmap mentions "World editor tools" as planned feature
- **Current State**: CLI tools exist (`cmd/map-editor/`, `cmd/quest-builder/`, `cmd/content-creator/`) but no graphical interface. README correctly notes "CLI tools only, no GUI editors" in roadmap section.
- **Impact**: Content creators must use command-line tools, limiting accessibility for non-technical game designers. Map editing requires manual coordinate entry.
- **Closing the Gap**:
  1. Design browser-based map editor leveraging existing WASM infrastructure in `pkg/wasmui/`
  2. Add WebSocket-based live preview to `cmd/map-editor/`
  3. Create visual drag-and-drop quest builder using existing quest schema
  4. Integrate with existing JSON-RPC API for real-time validation
  5. Estimated effort: 40-60 hours for basic GUI map editor

---

## Advanced Network Optimization

- **Stated Goal**: README mentions "Network optimization" in roadmap
- **Current State**: Basic rate limiting implemented in `pkg/server/ratelimit.go`. Delta compression added in `pkg/server/websocket_delta.go` (95% bandwidth reduction benchmarked). No advanced connection pooling or message batching.
- **Impact**: Adequate for small-scale deployments. May experience performance issues with 100+ concurrent WebSocket connections due to per-connection goroutine model without pooling.
- **Closing the Gap**:
  1. Implement connection pooling for WebSocket upgrade handling
  2. Add message batching for high-frequency state updates
  3. Enable permessage-deflate compression on WebSocket upgrader (partially implemented)
  4. Add connection health monitoring with automatic reconnection hints
  5. Benchmark under load: `go test ./pkg/server/... -bench=. -benchmem`

---

## CLI Tool Test Coverage

- **Stated Goal**: Project maintains 78% test coverage threshold enforced in CI
- **Current State**: Three CLI tools have 0% coverage:
  - `cmd/content-creator/main.go` (0.0%)
  - `cmd/map-editor/main.go` (0.0%)  
  - `cmd/quest-builder/main.go` (0.0%)
- **Impact**: Changes to CLI tools may introduce regressions undetected by CI. Combined 400+ lines of untested validation and I/O logic.
- **Closing the Gap**:
  1. Create `cmd/content-creator/main_test.go`:
     ```go
     func TestRun_CreateMode(t *testing.T) { ... }
     func TestRun_ValidateMode(t *testing.T) { ... }
     func TestParseArgs(t *testing.T) { ... }
     ```
  2. Create `cmd/map-editor/main_test.go`:
     ```go
     func TestDrawRect(t *testing.T) { ... }
     func TestInteractiveEdit(t *testing.T) { ... }
     ```
  3. Create `cmd/quest-builder/main_test.go`:
     ```go
     func TestValidateQuest(t *testing.T) { ... }
     func TestPromptObjectives(t *testing.T) { ... }
     ```
  4. Target: 70%+ coverage for each CLI tool

---

## Go Runtime Security Updates

- **Stated Goal**: SECURITY.md and CI pipeline emphasize security scanning
- **Current State**: `go.mod` specifies Go 1.23.0 which reached EOL August 2025. Multiple CVEs affect the runtime:
  - CVE-2025-4673: net/http sensitive header leak on redirect
  - CVE-2025-22874: crypto/x509 policy validation bypass
  - CVE-2025-68121: crypto/tls session resumption issue
  - CVE-2025-22871: net/http request smuggling via bare LF
- **Impact**: Production deployments may have unpatched security vulnerabilities. `govulncheck` will report findings.
- **Closing the Gap**:
  1. Update `go.mod` when Go 1.24+ is available:
     ```
     go 1.24
     toolchain go1.24.x
     ```
  2. Update CI workflow to use `go-version: '1.24'`
  3. Run `govulncheck ./...` to verify zero findings
  4. Monitor Go security announcements for point releases

---

## pkg/server Test Coverage

- **Stated Goal**: Project enforces 78% minimum coverage in CI
- **Current State**: `pkg/server/` has 70.5% coverage, below threshold. Specific gaps:
  - `handlers_pcg.go`: PCG handler error paths undertested
  - `handlers_spatial.go`: Spatial query edge cases missing
  - `websocket.go`: Reconnection and error recovery paths
- **Impact**: Core server package is 7.5% below coverage threshold. CI may fail if threshold enforcement is strictly applied to individual packages.
- **Closing the Gap**:
  1. Add tests to `pkg/server/pcg_handlers_test.go`:
     - Test `handleGenerateContent` with invalid content types
     - Test `handleGenerateLevel` with out-of-bounds dimensions
  2. Add tests to `pkg/server/spatial_integration_test.go`:
     - Test `getObjectsInRange` with empty world
     - Test radius queries at world boundaries
  3. Extend `pkg/server/websocket_test.go`:
     - Test connection drop and reconnect scenarios
     - Test message ordering under high load
  4. Target: 78%+ coverage for `pkg/server/`

---

## Documentation Accuracy (README Roadmap)

- **Stated Goal**: README provides accurate project status
- **Current State**: README roadmap section has outdated status indicators:
  - Line 405: Says "levels 0-2 only, levels 3-9 needed" but spells exist for all levels
  - Line 411: Says "faction generation only, no guild mechanics" but guild system is fully implemented
- **Impact**: Users may underestimate project completeness. Contributors may work on features that already exist.
- **Closing the Gap**:
  1. Update README.md line 405:
     ```markdown
     - [x] Complete spell system (levels 0-9, 60 spells across 10 YAML files)
     ```
  2. Update README.md line 411:
     ```markdown
     - [x] Guild and faction systems with full mechanics (ranks, permissions, treasury, perks)
     ```
  3. Add last-reviewed date to roadmap section
  4. Create automated check comparing code features to README claims

---

## Code Duplication in Core Packages

- **Stated Goal**: Maintainable, DRY codebase
- **Current State**: 2.51% duplication (1,456 lines). Notable clusters:
  - `pkg/game/faction_relations.go`: 11-line pattern × 10 = 110 lines
  - `pkg/game/guild.go`: 14-line pattern × 4 = 56 lines
  - `pkg/server/handlers_guild.go`: 14-line pattern × 7 = 98 lines
- **Impact**: Bug fixes must be applied to multiple locations. Increased maintenance burden and risk of drift between copies.
- **Closing the Gap**:
  1. Extract `pkg/game/faction_relations.go` helper:
     ```go
     func (fm *FactionManager) modifyRelationState(
         faction1, faction2 string, 
         newState RelationState,
         opinionDelta, trustDelta int,
     ) error
     ```
  2. Extract `pkg/game/guild.go` helper:
     ```go
     func (gm *GuildManager) validateAndExecuteMemberOp(
         guildID, actorID, targetID string,
         requiredPerm GuildPermission,
         operation func(*Guild, *GuildMember) error,
     ) error
     ```
  3. Extract `pkg/server/handlers_guild.go` response helper:
     ```go
     func guildSuccessResponse(message string) map[string]interface{}
     func guildErrorResponse(code int, message string) *JSONRPCError
     ```
  4. Validation: `go-stats-generator analyze . --skip-tests --sections duplication`

---

## Elevation-Based Combat (Incomplete Feature)

- **Stated Goal**: `pkg/game/combat_modifiers.go:285` contains TODO for elevation bonuses
- **Current State**: Comment exists but no implementation. Terrain system doesn't track elevation data.
- **Impact**: Combat system is missing a tactical depth feature that was planned. May cause confusion if documentation mentions elevation.
- **Closing the Gap**:
  1. **Option A - Implement:**
     - Add `Elevation int` field to `Tile` struct in `pkg/game/tile.go`
     - Implement `getElevationBonus(attackerTile, defenderTile *Tile) int` in `combat_modifiers.go`
     - Add +2 attack bonus for height advantage, -2 for disadvantage
     - Update terrain generation to assign elevation values
  2. **Option B - Defer:**
     - Remove TODO comment to avoid confusion
     - Add to ROADMAP.md as future enhancement
     - Document as "not implemented" in combat mechanics docs

---

## Summary

| Gap | Severity | Effort | Priority |
|-----|----------|--------|----------|
| Go Runtime Security | HIGH | Low (config change) | 1 |
| CLI Tool Test Coverage | HIGH | Medium (12-20 hrs) | 2 |
| pkg/server Coverage | MEDIUM | Medium (8-12 hrs) | 3 |
| Documentation Accuracy | LOW | Low (30 min) | 4 |
| Code Duplication | MEDIUM | Medium (4-8 hrs) | 5 |
| Asset Generation | MEDIUM | High (ongoing) | 6 |
| Network Optimization | LOW | High (20-40 hrs) | 7 |
| GUI World Editor | LOW | Very High (40-60 hrs) | 8 |
| Elevation Combat | LOW | Medium (4-8 hrs) | 9 |

**Recommended order**: Address security (#1), then coverage gaps (#2-3), then documentation (#4), then incremental improvements (#5-9) as time permits.
