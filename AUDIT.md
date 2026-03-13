# AUDIT — 2026-03-13

## Project Goals

GoldBox RPG Engine claims to be a modern, Go-based RPG engine inspired by the classic SSI Gold Box series. Per the README, it promises:

**Core Systems:**
- Character management with 6 core attributes (STR, DEX, CON, INT, WIS, CHA)
- Class-based system (Fighter, Mage, Cleric, Thief, Ranger, Paladin)
- Multiple character creation methods (roll, standard array, point-buy, custom)
- Equipment and inventory management with class proficiency restrictions
- Experience and level progression

**Combat & Effects:**
- Comprehensive effect system (DoT, HoT, Stun, Root, Burning, Bleeding, Poison)
- Effect stacking and priority management
- Immunity and resistance handling

**World Management:**
- Tile-based environments with multiple terrain types
- Advanced spatial indexing (R-tree-like structure)
- Object/NPC management with procedural generation
- Combat positioning and line-of-sight

**Infrastructure:**
- JSON-RPC 2.0 API over HTTP and WebSockets
- Real-time event broadcasting
- Health monitoring endpoints (/health, /ready, /live, /metrics)
- Circuit breaker patterns and retry mechanisms
- Input validation framework

**Content:**
- 60 spells across levels 0-9
- 10 embedded adventure packs (51 maps, 37 quests)
- Procedural content generation (terrain, items, quests, NPCs)
- Asset generation pipeline for 521 game assets

**Target Audience:** Game developers building web-based RPG experiences with classical tabletop RPG mechanics.

---

## Goal-Achievement Summary

| Goal | Status | Evidence |
|------|--------|----------|
| 6 Core Attributes | ✅ Achieved | `pkg/game/character.go:41-46` — Strength, Dexterity, Constitution, Intelligence, Wisdom, Charisma |
| 6 Character Classes | ✅ Achieved | `pkg/game/classes.go` — Fighter, Mage, Cleric, Thief, Ranger, Paladin |
| Character Creation Methods | ✅ Achieved | `pkg/game/character_creation.go` — 4 methods implemented |
| Equipment with Proficiency | ✅ Achieved | `pkg/game/character_equipment.go`, `pkg/game/classes.go` — proficiency validation |
| Comprehensive Effect System | ✅ Achieved | `pkg/game/effects.go`, `pkg/game/effectmanager.go` — 5 DoT types, stacking |
| Spatial Indexing | ✅ Achieved | `pkg/game/spatial_index.go` — R-tree-like SpatialIndex |
| A* Pathfinding | ✅ Achieved | `pkg/game/pathfinding.go` — full implementation |
| JSON-RPC API | ✅ Achieved | `pkg/server/` — 71 RPC methods defined in `constants.go` |
| WebSocket Communication | ✅ Achieved | `pkg/server/websocket_nhooyr.go` — nhooyr.io/websocket |
| Health Endpoints | ✅ Achieved | `pkg/server/health.go` — /health, /ready, /live implemented |
| Prometheus Metrics | ✅ Achieved | `pkg/server/metrics.go` — /metrics endpoint |
| Circuit Breaker | ✅ Achieved | `pkg/resilience/circuitbreaker.go` — full implementation |
| Retry Mechanisms | ✅ Achieved | `pkg/retry/retry.go` — exponential backoff |
| Input Validation | ✅ Achieved | `pkg/validation/` — 71+ method validators |
| 60 Spells (Levels 0-9) | ✅ Achieved | `data/spells/*.yaml` — 10 files, 60 spells verified |
| 10 Adventure Packs | ✅ Achieved | `data/adventures/` — 10 directories |
| PCG System | ✅ Achieved | `pkg/pcg/` — terrain, items, quests, NPC generation |
| Guild System | ✅ Achieved | `pkg/game/guild.go` — 685 lines, 5 ranks, treasury |
| Faction System | ✅ Achieved | `pkg/game/faction_relations.go` — relations, diplomacy |
| Asset Pipeline (521 assets) | ⚠️ Partial | 252/521 assets exist; requires external AI tool |
| World Editor Tools | ⚠️ Partial | CLI tools only (`cmd/map-editor/`, `cmd/quest-builder/`) |
| Content Creation Utilities | ⚠️ Partial | CLI-only, no visual editors |

**Summary: 20/23 goals fully achieved, 3/23 partial**

---

## Findings

### CRITICAL

*None identified.* All core game mechanics function as documented.

### HIGH

- [x] **WebSocket library deprecated** — `go.mod:23` — Production code uses `nhooyr.io/websocket v1.8.17` which was deprecated in mid-2024 and transferred to Coder. The new official package is `github.com/coder/websocket`. While the library remains functional, no new security patches will be released to the original namespace. — **Remediation:** Update import path from `nhooyr.io/websocket` to `github.com/coder/websocket` across all production files in `pkg/server/websocket_*.go`. Verify with `go test ./pkg/server/... -v`. — **COMPLETED: 2026-03-13** — Migrated to `github.com/coder/websocket v1.8.14`.

- [ ] **Server package coupling** — `pkg/server/` — Package has 11 dependencies (coupling score: 5.5), highest in codebase. High coupling increases change propagation risk and testing complexity. — **Remediation:** Extract session management into dedicated `pkg/session/` package. Extract WebSocket handling into `pkg/websocket/`. Verify coupling reduction with `go-stats-generator analyze . --skip-tests --sections packages`.

### MEDIUM

- [ ] **Asset count gap** — `web/static/assets/sprites/` — README claims 521 assets with `make assets` but only 252 placeholder PNG files exist. Pipeline code complete but requires external AI tool not included in repository. README correctly notes this with ⚠️ warning but badge shows "252 placeholders/521 defined" which may confuse users. — **Remediation:** Create `scripts/download-assets.sh` to fetch pre-generated asset packs from GitHub Releases. Add `make assets-download` target. Update badge to clearly indicate "252 ready / 521 with AI tool".

- [x] **CLI tool test coverage below project mean** — `cmd/map-editor/main.go` — Coverage at 64.6% vs project mean of 81.1%. Core content creation tool with lower-than-average test coverage. — **Remediation:** Add table-driven tests for command parsing and map validation in `cmd/map-editor/main_test.go`. Target 70% coverage. Verify with `go test ./cmd/map-editor/... -cover`. — **COMPLETED: 2026-03-13** — Coverage improved to 79.9%.

- [x] **CLI tool test coverage below project mean** — `cmd/content-creator/main.go` — Coverage at 60.8% vs project mean of 81.1%. Second content creation tool with lowest coverage among CLI tools. — **Remediation:** Add integration tests for output validation in `cmd/content-creator/main_test.go`. Target 70% coverage. Verify with `go test ./cmd/content-creator/... -cover`. — **PARTIAL: 2026-03-13** — Coverage improved to 61.9%. Remaining gap (to 70%) blocked by interactive functions requiring stdin refactoring.

- [ ] **Server package test coverage gap** — `pkg/server/` — Coverage at 73.8% vs 87.8% for `pkg/game`. Core server infrastructure has lower coverage than game mechanics. — **Remediation:** Add tests for session timeout paths in `pkg/server/session.go`, WebSocket error recovery in `pkg/server/websocket_nhooyr.go`, and RPC edge cases in `pkg/server/handlers.go`. Target 80% coverage. Verify with `go test ./pkg/server/... -cover`.

- [ ] **Code duplication in server handlers** — `pkg/server/server.go:1027` — 35-line code clone repeated for handler registration (37 clone pairs total, 938 duplicated lines, 1.51% ratio). — **Remediation:** Extract handler registration into table-driven `registerMethodHandlers(map[string]HandlerFunc)` function. Target <500 duplicated lines. Verify with `go-stats-generator analyze . --skip-tests --sections duplication`.

- [ ] **Low cohesion packages** — `pkg/cliutil/`, `pkg/secrets/` — Cohesion scores 0.8 and 0.7 respectively, indicating functions may be misplaced or package boundaries unclear. — **Remediation:** Review function placement in low-cohesion packages. Consider merging `pkg/secrets/` into `pkg/config/` if related. Verify improvement with `go-stats-generator analyze . --skip-tests --sections packages`.

### LOW

- [ ] **gorilla/websocket in test dependencies** — `go.mod:13-14` — Archived library (gorilla/websocket v1.5.3) used in E2E test client. Comment documents rationale but archived packages may accumulate unpatched CVEs. — **Remediation:** Monitor CVE-2020-27813 and related advisories. Consider migrating test client to `github.com/coder/websocket` when convenient. Low priority as test-only usage.

- [ ] **Naming convention violations** — Various files — 28 identifier violations (stuttering: `AdventureManager`, `GameEvent`, `GameMap`) and 14 file name violations (generic names: `constants.go`, `types.go`, `errors.go`). — **Remediation:** Rename identified types to remove package prefix (e.g., `game.AdventureManager` → `game.Manager`). Rename generic files to descriptive names (e.g., `game/constants.go` → `game/attribute_constants.go`). Verify with `go-stats-generator analyze . --skip-tests --sections naming`.

- [ ] **Function placement suggestions** — Various files — 566 misplaced functions identified by affinity analysis. Average file cohesion 0.33. — **Remediation:** Review top misplaced functions (e.g., `NewTimeManager` in `combat.go` → `server.go`). Relocate based on usage patterns. Verify with `go-stats-generator analyze . --skip-tests --sections placement`.

---

## Metrics Snapshot

| Metric | Value | Threshold | Status |
|--------|-------|-----------|--------|
| Total Lines of Code | 30,955 | — | — |
| Total Functions | 612 | — | — |
| Total Methods | 1,720 | — | — |
| Total Structs | 408 | — | — |
| Total Interfaces | 22 | — | — |
| Total Packages | 19 | — | — |
| Total Files | 197 | — | — |
| Test Coverage | 81.1% | ≥60% | ✅ Exceeds |
| Max Cyclomatic Complexity | 10 | ≤15 | ✅ Healthy |
| High Complexity Functions (CC>15) | 0 | <5 | ✅ Excellent |
| Functions at CC=10 | 4 | — | ⚠️ Monitor |
| Duplication Ratio | 1.51% | <3% | ✅ Acceptable |
| Duplicated Lines | 938 | <1000 | ✅ Acceptable |
| Clone Pairs | 37 | — | — |
| Circular Dependencies | 0 | 0 | ✅ None |
| Race Conditions | 0 | 0 | ✅ Clean |
| go vet Issues | 0 | 0 | ✅ Clean |

### Complexity Distribution

| Complexity Range | Count | Percentage |
|------------------|-------|------------|
| CC = 10 | 4 | 0.2% |
| CC = 9 | 6 | 0.3% |
| CC ≤ 8 | 2,322 | 99.5% |

**Top 5 Complex Functions:**
| Function | File | CC | Lines |
|----------|------|----|----|
| Stop | test/e2e/server.go:164 | 10 | 42 |
| run | cmd/quest-builder/main.go | 10 | 46 |
| Validate | cmd/bootstrap-demo/main.go | 10 | 44 |
| RunCellularAutomata | pkg/pcg/terrain/cellular_automata.go | 10 | 40 |
| ValidateAndFix | pkg/pcg/validator.go | 9 | 46 |

### Package Coverage

| Package | Coverage | Status |
|---------|----------|--------|
| pkg/game | 87.8% | ✅ Excellent |
| pkg/wasmui | 94.1% | ✅ Excellent |
| pkg/validation | 92.0% | ✅ Excellent |
| pkg/resilience | 94.5% | ✅ Excellent |
| pkg/config | 94.0% | ✅ Excellent |
| pkg/server | 73.8% | ⚠️ Below mean |
| cmd/map-editor | 64.6% | ⚠️ Below mean |
| cmd/content-creator | 60.8% | ⚠️ Below mean |

---

## Verification Commands

```bash
# Verify goal achievement
go test -race ./...                              # All tests pass, no races
go vet ./...                                     # No static analysis issues
find data/adventures -name "adventure.yaml" | wc -l  # Expected: 10
cat data/spells/*.yaml | grep -c "spell_id:"    # Expected: 60
find web/static/assets -name "*.png" | wc -l    # Expected: 252 (gap: 521)

# Verify coverage
go test ./... -coverprofile=/tmp/c.out && go tool cover -func=/tmp/c.out | grep total
# Expected: ~81.1%

# Verify metrics
go-stats-generator analyze . --skip-tests --sections functions,duplication
# Expected: max CC ≤10, duplication <3%

# Verify health endpoints (with server running)
curl http://localhost:8080/health | jq .status   # Expected: "healthy"
curl http://localhost:8080/ready                 # Expected: 200 OK
curl http://localhost:8080/live                  # Expected: 200 OK
```

---

*Generated by functional audit against README.md stated goals*
*Analysis tool: go-stats-generator v1.0.0*
*Last Updated: 2026-03-13*
