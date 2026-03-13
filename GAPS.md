# Implementation Gaps — 2026-03-13

## Summary

This document identifies gaps between the GoldBox RPG Engine's stated goals and current implementation. The project achieves 15 of 17 stated goals, with two partial implementations.

**Gap Status:**
- ✅ Fully Achieved: 15 goals
- ⚠️ Partial: 2 goals
- ❌ Missing: 0 goals

---

## Active Gaps

### Asset Generation Requires External AI Tool

- **Stated Goal**: README claims "Asset Generation Pipeline for 521 game assets across 6 categories" with automated creation.
- **Current State**: 
  - Pipeline code is complete and functional (`game-assets.yaml`, `scripts/generate-*.sh`)
  - 252/521 assets exist (48.4% coverage)
  - All existing assets are placeholder PNGs (colored rectangles), not AI-generated art
  - `make assets-verify` passes but reports "PARTIAL"
- **Impact**: 
  - Game is fully playable with placeholders (visual experience degraded but functional)
  - Users expecting production-ready art will need to run 4-6 hour AI generation process
  - External tool setup (Stable Diffusion, DALL-E) required before full generation
- **Closing the Gap**:
  1. Install external AI image generation tool per `ASSET_INTEGRATION.md`
  2. Configure API keys and model settings
  3. Run `make assets` (4-6 hours for 521 assets)
  4. Verify with `make assets-verify` expecting 521/521
  5. Alternative: Contact maintainer for pre-generated asset packs

---

### GUI World Editor Tools Not Implemented

- **Stated Goal**: README roadmap lists "World editor tools" though marked with ⚠️ noting "CLI tools only, no GUI editors"
- **Current State**: 
  - CLI tools exist and function:
    - `cmd/map-editor/` — Terminal-based map creation with interactive editing (complexity 16.3 in `interactiveEdit`)
    - `cmd/quest-builder/` — Quest chain builder with validation
    - `cmd/content-creator/` — Content generation tool
  - No browser-based or graphical editors
  - WebSocket editor protocol exists (`pkg/server/websocket_editor.go`) but lacks frontend
- **Impact**: 
  - Content creators must use command-line interfaces
  - Barrier to entry for non-technical users creating adventures
  - Existing `pkg/wasmui/editor.go` provides foundation but incomplete
- **Closing the Gap**:
  1. Extend `pkg/wasmui/editor.go` with visual map editing UI
  2. Connect WebSocket editor protocol to Ebitengine canvas
  3. Add visual quest builder using existing quest schema
  4. Test with `go test ./pkg/wasmui/... -v -run Editor`

---

## Security Considerations

### WebSocket Rate Limiting Gap

- **Stated Goal**: README claims "Input Validation — Security against injection attacks, DoS prevention"
- **Current State**: 
  - HTTP requests are rate-limited (`pkg/server/server.go:807` via `checkRateLimit()`)
  - WebSocket RPC messages bypass rate limiting entirely (`pkg/server/websocket.go:340-376`)
  - Token bucket algorithm exists (`pkg/server/ratelimit.go`) but not applied to WebSocket path
- **Impact**: 
  - Clients can send unlimited WebSocket RPC requests per second
  - Expensive operations (spell casting, combat, movement) can be spammed
  - Potential DoS vector through WebSocket connection abuse
- **Closing the Gap**:
  1. Add rate limit check in `processWebSocketRequest()` at line 354
  2. Apply per-session rate limiting using existing `RateLimiter` infrastructure
  3. Test with: `go test ./pkg/server/... -v -run TestWebSocket`
  4. Benchmark to ensure rate limiting doesn't impact legitimate gameplay

---

### Go Toolchain Security Updates

- **Stated Goal**: Production security for game servers handling untrusted input
- **Current State**: 
  - Project requires Go 1.24.0 (`go.mod:3`)
  - Go 1.24 reached EOL on 2026-02-11
  - Six critical CVEs patched in Go 1.24.12/1.25.6:
    - CVE-2025-61728: archive/zip DoS via malicious filenames
    - CVE-2025-61726: net/http memory exhaustion via large forms
    - CVE-2025-68121: crypto/tls session key leak
    - CVE-2025-61731: cmd/go code execution via pkg-config
    - CVE-2025-68119: VCS toolchain code execution
    - CVE-2025-61730: crypto/tls handshake information disclosure
- **Impact**: 
  - Servers running on Go 1.24.0-1.24.11 are vulnerable to these CVEs
  - Web applications handling untrusted input most affected
  - TLS security weakened by session resumption bugs
- **Closing the Gap**:
  1. Update `go.mod` to require Go 1.25.6 or later
  2. Run `go mod edit -go=1.25.6 && go mod tidy`
  3. Update CI workflows to use Go 1.25.6+
  4. Validate with: `go version && go test -race ./...`

---

## Dependency Considerations

### Gorilla WebSocket Archived Status

- **Current State**: Using gorilla/websocket v1.5.3, archived since September 2022
- **Risk Level**: LOW — No CVEs in 2024-2026, library functions correctly
- **Known Resolved Issues**: CVE-2020-27813 (integer overflow) patched in v1.4.1
- **Migration Options**:
  1. Continue using v1.5.3 (functional, well-tested)
  2. Migrate to `nhooyr.io/websocket` (active development, modern API)
  3. Migrate to `golang.org/x/net/websocket` (stdlib, minimal features)
- **Recommendation**: Plan migration per `docs/WEBSOCKET_MIGRATION.md` when time permits; not urgent

---

## Verification Commands

```bash
# Verify all goals are achieved
go test -race ./...                    # All tests pass
make adventures-verify                  # 10/10 adventures valid
make assets-verify                      # 252/521 assets (partial)
go test ./test/e2e/... -v              # E2E tests pass

# Check coverage meets threshold
go test ./... -coverprofile=/tmp/c.out && go tool cover -func=/tmp/c.out | grep total
# Expected: ~79.3% coverage (above 60% threshold)

# Verify no race conditions
go test -race ./pkg/game/... ./pkg/server/...

# Check for security vulnerabilities (requires Go 1.25+)
govulncheck ./...
```

---

## Resolved Gaps

The following gaps were identified in prior audits but have since been resolved:

| Previous Gap | Resolution | Evidence |
|--------------|------------|----------|
| Spell System Incomplete (levels 0-2 only) | Levels 3-9 implemented | `data/spells/` contains 11 YAML files with 60 spells |
| Guild Mechanics Missing | Full implementation exists | `pkg/game/guild.go` (686 lines, 5 ranks, treasury, perks) |
| Network Delta Compression Missing | Implemented | `pkg/server/websocket_delta.go` with 95% bandwidth savings |
| Adventure System Incomplete | All 10 adventures complete | `make adventures-verify` reports 10/10 valid |
| E2E Test Failures | All tests now pass | `go test ./test/e2e/... -v` shows 100% pass rate |
| Coverage Below Threshold | Coverage at 79.3% | CI threshold is 60%, well exceeded |

---

*Last Updated: 2026-03-13*
