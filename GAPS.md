# Implementation Gaps — 2026-03-13

## Deprecated WebSocket Library

- **Stated Goal**: Production-ready WebSocket real-time communication with security and ongoing maintenance support.
- **Current State**: The project uses `github.com/gorilla/websocket v1.5.3`, which has been archived and deprecated by its maintainers. While v1.5.3 has no known critical vulnerabilities, future security issues will not receive patches.
- **Impact**: Long-term security risk. Any newly discovered vulnerabilities in Gorilla WebSocket will remain unpatched. This affects all real-time game features including state synchronization, event broadcasting, and multiplayer sessions.
- **Closing the Gap**:
  1. Evaluate replacement libraries: `nhooyr.io/websocket` (idiomatic, context-based), `github.com/gobwas/ws` (high performance), or `github.com/coder/websocket`
  2. Create migration branch and update `pkg/server/websocket.go`, `pkg/server/websocket_delta.go`
  3. Update upgrader configuration and connection handling to match new API
  4. Run full E2E test suite: `go test ./test/e2e/... -v`
  5. Timeline: Complete migration within 6 months

---

## Vault Secret Provider Non-Functional

- **Stated Goal**: `pkg/secrets/` package suggests multiple secret providers including HashiCorp Vault for enterprise deployments.
- **Current State**: `pkg/secrets/vault_provider.go:99-116` returns `ErrNotImplemented` for all operations (Get, Set, Delete, HealthCheck). The implementation is a documented stub with no actual Vault connectivity.
- **Impact**: Users expecting Vault integration for production secrets management cannot use this feature. The secrets package only works with environment variables (`EnvSecretProvider`) and local files.
- **Closing the Gap**:
  1. **Option A (Full Implementation):**
     - Add `github.com/hashicorp/vault/api` dependency
     - Implement token authentication in `vault_provider.go`
     - Add support for AppRole and Kubernetes auth methods
     - Test with actual Vault instance
     - Effort: 4-8 hours
  2. **Option B (Remove Stub):**
     - Remove `vault_provider.go` or mark it clearly as "planned future feature"
     - Update package documentation to list only env/file providers
     - Effort: 30 minutes
  3. Validation: `go test ./pkg/secrets/... -v` passes; documentation matches reality

---

## Asset Pipeline Produces Placeholders Only

- **Stated Goal**: README claims "521 total assets across 6 categories" with a "complete asset generation pipeline" and states users can run `make assets` to generate all game assets.
- **Current State**: 252 PNG files exist in `web/static/assets/sprites/`. ALL are minimal 245-byte placeholder images (likely 1x1 transparent PNGs). The README badge "252/521 (48%)" is technically accurate but misleading since these are placeholders, not usable game art. Real asset generation requires external AI tools (Stable Diffusion/DALL-E) with 4-6 hour setup.
- **Impact**: The game runs but has no visual fidelity. Players see blank or minimal placeholder graphics for all characters, monsters, items, terrain, and UI elements. This significantly impacts user experience and makes the game unsuitable for demonstration without custom art.
- **Closing the Gap**:
  1. Update README to clarify: "252 placeholder assets / 0 real art assets generated"
  2. Provide one of:
     - Pre-generated asset pack download link
     - CC0/public domain pixel art bundle for basic playability
     - Detailed step-by-step AI tool integration guide with expected outputs
  3. Consider separate "Asset Generation" section in README with honest time/effort estimates
  4. Validation: `find web/static/assets -name "*.png" -size +1k | wc -l` shows actual art asset count

---

## GUI World Editor Tools Not Available

- **Stated Goal**: README roadmap mentions "World editor tools" with note "(CLI tools only, no GUI editors)".
- **Current State**: CLI tools exist and are functional:
  - `cmd/map-editor/` — Interactive terminal-based map creation
  - `cmd/quest-builder/` — Interactive quest chain builder
  - `cmd/content-creator/` — Interactive content generation tool
  
  No web-based or graphical editors exist. The existing Ebitengine/WASM infrastructure in `pkg/wasmui/editor.go` provides basic functionality but is not a full GUI editor.
- **Impact**: Content creators must use command-line tools, limiting adoption to technically proficient users. Non-technical game designers cannot easily create custom content.
- **Closing the Gap**:
  1. This is noted as a known limitation in README, not a discrepancy
  2. Enhancement path:
     - Extend `pkg/wasmui/editor.go` to full browser-based map editing
     - Add WebSocket-based preview for real-time editing feedback
     - Create visual quest builder using existing quest schema
  3. Effort estimate: 2-4 weeks for basic GUI editor
  4. Validation: User can create and save a map without command-line interaction

---

## Go Toolchain Version Requirements

- **Stated Goal**: Project requires Go 1.24.0+ as stated in `go.mod` and README badge.
- **Current State**: The project builds and tests successfully with Go 1.24.0+. However, `govulncheck` (when run with Go 1.23) reports version mismatch errors for many files. The `vendor/` directory includes files that strictly require Go 1.24.
- **Impact**: Developers or CI systems using Go 1.23 will encounter confusing errors. Some security scanning tools may fail to analyze the codebase.
- **Closing the Gap**:
  1. Ensure all CI/CD workflows explicitly use Go 1.24.2+ toolchain
  2. Add Go version check to Makefile: `@go version | grep -q "1.24" || (echo "Go 1.24+ required"; exit 1)`
  3. Update CONTRIBUTING.md to specify exact Go version requirements
  4. Validation: `go version` shows 1.24.0+; all tools run without version errors

---

## Summary Table

| Gap | Severity | Effort to Close | Priority |
|-----|----------|-----------------|----------|
| Gorilla WebSocket Deprecated | HIGH | 2-4 days (migration) | P1 — Security |
| Vault Provider Stub | HIGH | 4-8h (implement) or 30m (remove) | P2 — Correctness |
| Asset Placeholders Only | MEDIUM | 1h (docs) or 4-6h (real assets) | P2 — UX |
| CLI-Only Editors | LOW | 2-4 weeks (enhancement) | P3 — Enhancement |
| Go Version Clarity | LOW | 1h (docs + CI update) | P3 — DX |

---

## Previously Reported Gaps Now Resolved

The following gaps from the previous GAPS.md have been verified as resolved:

| Previous Gap | Status | Evidence |
|--------------|--------|----------|
| Adventure System Non-Functional | ✅ RESOLVED | `pkg/validation/validation.go:237-238` registers adventure validators; all 11 E2E tests pass |
| Validation Layer Missing Adventure Methods | ✅ RESOLVED | `adventure.list` and `adventure.load` validators registered at lines 237-238 |
| Network Delta Compression | ✅ RESOLVED | `pkg/server/websocket_delta.go` implements state diffing with documented 95.6% bandwidth reduction |
| README Roadmap Accuracy | ✅ RESOLVED | README updated to reflect spell system completion and guild mechanics |
