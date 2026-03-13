# Implementation Gaps — 2026-03-13

## Summary

This document identifies gaps between the GoldBox RPG Engine's stated goals and current implementation. The project achieves 17 of 18 stated goals, with one partial implementation requiring external tooling.

**Gap Status:**
- ✅ Fully Achieved: 17 goals
- ⚠️ Partial: 1 goal (asset generation)
- ❌ Missing: 0 goals

---

## Active Gaps

### Asset Generation Requires External AI Tool

- **Stated Goal**: README claims "Asset Generation Pipeline for 521 game assets across 6 categories" with automated creation via `make assets`.
- **Current State**: 
  - Pipeline code is complete and functional (`game-assets.yaml`, `scripts/generate-*.sh`)
  - 252/521 assets exist (48.4% coverage)
  - All existing assets are colored rectangle placeholders, not AI-generated art
  - `make assets-verify` passes with "PARTIAL" status
  - Pipeline requires external AI tool (Stable Diffusion, DALL-E) for full generation
- **Impact**: 
  - Game is fully playable with placeholders — visual experience degraded but functional
  - Users expecting production-ready art must run 4-6 hour AI generation process
  - External tool setup required per `ASSET_INTEGRATION.md` before full generation
- **Closing the Gap**:
  1. Install external AI image generation tool per `ASSET_INTEGRATION.md`
  2. Configure API keys and model settings for chosen tool
  3. Run `make assets` (4-6 hours for 521 assets)
  4. Verify with `make assets-verify` expecting 521/521
  5. Alternative: Contact maintainer for pre-generated asset packs
  6. **Validation:** `make assets-verify` reports 521/521 assets present

---

## Infrastructure Considerations

### Go Toolchain Security Updates

- **Stated Goal**: Production security for game servers handling untrusted input
- **Current State**: 
  - Project requires Go 1.24.0 (`go.mod:3`)
  - Go 1.24 reached EOL on 2026-02-11
  - Six CVEs patched in Go 1.24.12+/1.25+:
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
  1. Update `go.mod` to require Go 1.24.12 or Go 1.25+
  2. Run `go mod edit -go=1.24.12 && go mod tidy`
  3. Update CI workflows (`.github/workflows/ci.yml`) to use Go 1.24.12+
  4. **Validation:** `go version && go test -race ./...`

---

### gorilla/websocket Test Dependency

- **Stated Goal**: Maintain modern, actively supported dependencies
- **Current State**: 
  - Production WebSocket code migrated to nhooyr.io/websocket (2026-03-13)
  - gorilla/websocket v1.5.3 remains in `go.mod` for E2E test compatibility
  - gorilla/websocket archived since September 2022 — no security patches
- **Impact**: 
  - No production impact — server uses nhooyr.io/websocket
  - Test code uses gorilla for WebSocket client compatibility (standard protocol)
  - Dependency scanner may flag gorilla as abandoned
- **Closing the Gap**:
  1. Evaluate if E2E tests can use nhooyr.io/websocket client API
  2. If gorilla required for compatibility, document rationale in `go.mod` comment
  3. Add explicit `// indirect` comment explaining test-only usage
  4. **Validation:** `go mod graph | grep gorilla` shows test path only

---

## Enhancement Opportunities

### GUI World Editor (Optional Enhancement)

- **Stated Goal**: README roadmap notes "World editor tools" with ⚠️ indicating "CLI tools only, no GUI editors"
- **Current State**: 
  - CLI tools exist and function:
    - `cmd/map-editor/` — Terminal-based map creation with interactive editing
    - `cmd/quest-builder/` — Quest chain builder with validation
    - `cmd/content-creator/` — Content generation tool
  - WebSocket editor protocol exists (`pkg/server/websocket_editor.go`)
  - WASM editor foundation exists (`pkg/wasmui/editor.go`)
  - No browser-based graphical editors
- **Impact**: 
  - Content creators must use command-line interfaces
  - Barrier to entry for non-technical users creating adventures
  - Existing infrastructure provides foundation for GUI
- **Closing the Gap** (optional):
  1. Extend `pkg/wasmui/editor.go` with visual map editing UI
  2. Connect WebSocket editor protocol to Ebitengine canvas
  3. Add visual quest builder using existing quest schema
  4. **Validation:** User can create and save a map without command-line interaction

---

## Resolved Gaps

The following gaps were identified in prior audits but have since been resolved:

| Previous Gap | Resolution | Evidence |
|--------------|------------|----------|
| Gorilla WebSocket (Production) | Migrated to nhooyr.io/websocket | `pkg/server/websocket_nhooyr.go`, `docs/WEBSOCKET_MIGRATION.md` |
| Spell System Incomplete (levels 0-2) | Levels 3-9 implemented | `data/spells/` contains 11 YAML files with 60 spells |
| Guild Mechanics Missing | Full implementation exists | `pkg/game/guild.go` (685 lines, 5 ranks, treasury, perks) |
| Network Delta Compression Missing | Implemented | `pkg/server/websocket_delta.go` with 95% bandwidth savings |
| Adventure System Incomplete | All 10 adventures complete | `make adventures-verify` reports 10/10 valid, 51 maps, 37 quests |
| E2E Test Failures | All tests now pass | `go test ./test/e2e/... -v` shows 100% pass rate |
| Coverage Below Threshold | Coverage at 79.1% | CI threshold is 60%, well exceeded |
| WebSocket Rate Limiting Missing | Added to processWebSocketRequest | `pkg/server/websocket.go:297-304` rate limiting check |
| High Complexity Functions | Refactored | All critical functions under complexity 10 |

---

## Verification Commands

```bash
# Verify all goals are achieved
go test -race ./...                    # All tests pass
make adventures-verify                  # 10/10 adventures valid
make assets-verify                      # 252/521 assets (partial)

# Check coverage meets threshold
go test ./... -coverprofile=/tmp/c.out && go tool cover -func=/tmp/c.out | grep total
# Expected: ~79.1% coverage (above 60% threshold)

# Verify no race conditions
go test -race ./pkg/game/... ./pkg/server/...

# Check Go version requirements
go version
# Should be Go 1.24.12+ for security patches

# Verify dependency health
go mod graph | grep -E "(gorilla|nhooyr)"
```

---

*Last Updated: 2026-03-13*
