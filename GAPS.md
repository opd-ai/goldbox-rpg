# Implementation Gaps — 2026-03-13

## Adventure System Non-Functional (CRITICAL)

- **Stated Goal**: README and ROADMAP indicate an adventure system with 10 complete adventures, adventure selection UI, and `adventure.list`/`adventure.load` RPC methods. The codebase includes `pkg/game/adventure.go`, `pkg/server/handlers_adventure.go`, and 10 adventure data packs in `data/adventures/`.

- **Current State**: The adventure RPC handlers are correctly implemented and registered in `pkg/server/server.go:1105-1106`. However, the input validation layer in `pkg/validation/validation.go` does not include `adventure.list` or `adventure.load` in its `registerValidators()` function (lines 110-229). When any client calls these methods, the validation layer rejects them with "unknown method" before they reach the handler.

- **Impact**: Users cannot access any adventure content through the API. The 10 complete adventure data packs (`data/adventures/sunken-sanctum/`, `crimson-coast/`, etc.) are unreachable. All 11 E2E adventure tests fail. This effectively means the advertised adventure feature does not work at all.

- **Closing the Gap**:
  1. Add validation registrations to `pkg/validation/validation.go` in `registerValidators()`:
     ```go
     // Adventure methods
     v.validators["adventure.list"] = v.validateNoParams
     v.validators["adventure.load"] = sessionAndExtractValidatorFunc("adventure.load")
     ```
  2. Run `go test ./test/e2e/... -run TestAdventure -v` to verify fix
  3. Update `pkg/validation/validation_test.go` with adventure method tests

---

## Asset Generation Pipeline (Partial)

- **Stated Goal**: README claims "521 total assets across 6 categories" with a "complete asset generation pipeline" and states users can run `make assets` to generate all game assets.

- **Current State**: The asset pipeline configuration exists (`game-assets.yaml`), placeholder generation scripts work (`scripts/generate-placeholders.sh`), and 513 PNG files exist in `web/static/assets/sprites/`. However, ALL 513 images are minimal placeholders (~245 bytes each, likely 1x1 transparent PNGs). Zero actual game art assets exist. The README notes this requires external AI tools (Stable Diffusion/DALL-E) but the badge claims "6/521 (1%)" assets, which is misleading since even those 6 are placeholders.

- **Impact**: The game runs but has no visual fidelity. Players see blank or minimal placeholder graphics for all characters, monsters, items, terrain, and UI elements. This significantly impacts the user experience and makes the game unsuitable for any real play testing or demonstration.

- **Closing the Gap**:
  1. Update README badge to accurately reflect "0% real assets" or "placeholder assets only"
  2. Provide pre-generated asset packs for download (contact maintainer option mentioned in README)
  3. Complete the external AI tool integration documentation with step-by-step setup
  4. Consider bundling a minimal set of real CC0/public domain pixel art for basic playability

---

## Vault Secrets Provider (Stub Only)

- **Stated Goal**: `pkg/secrets/` package suggests support for multiple secret providers including HashiCorp Vault integration.

- **Current State**: `pkg/secrets/vault_provider.go:105` contains `TODO: Future implementation will use:` comment. The Vault provider is a stub that doesn't actually connect to or retrieve secrets from HashiCorp Vault.

- **Impact**: Users requiring Vault integration for production deployments cannot use this feature. The secrets package works only with environment variables and local file providers.

- **Closing the Gap**:
  1. Either implement the Vault provider fully using the HashiCorp Vault Go client
  2. Or remove the stub and document that only env/file providers are supported
  3. Update package documentation to reflect actual capabilities

---

## GUI World Editor Tools (CLI Only)

- **Stated Goal**: README roadmap mentions "World editor tools" with a note "(CLI tools only, no GUI editors)".

- **Current State**: CLI tools exist: `cmd/map-editor/`, `cmd/quest-builder/`, `cmd/content-creator/`. These are terminal-based interactive tools. No web-based or graphical editors exist.

- **Impact**: Content creators must use command-line tools, which limits adoption to technically proficient users. The existing Ebitengine/WASM infrastructure could support browser-based editors.

- **Closing the Gap**:
  1. Extend `pkg/wasmui/editor.go` to provide browser-based map editing
  2. Add WebSocket-based preview for real-time map editing feedback
  3. Create visual quest builder using existing quest schema
  4. This is marked as "Enhancement" priority in GAPS.md since CLI tools work

---

## Network Delta Compression (Partial)

- **Stated Goal**: README mentions "Network optimization" with a note "(basic pooling/rate limiting, no delta compression)".

- **Current State**: The GAPS.md in the repo claims "95.6% bandwidth reduction" from delta compression implementation in `pkg/server/websocket_delta.go`. However, the README still notes "no delta compression". One of these is outdated.

- **Impact**: If delta compression IS implemented, the README is misleading users about capabilities. If it's NOT fully implemented, the internal GAPS.md is inaccurate.

- **Closing the Gap**:
  1. Verify actual delta compression implementation status in `pkg/server/websocket_delta.go`
  2. Update README to reflect true capability (either add ✅ or remove the caveat)
  3. Add benchmark tests to CI to verify ongoing bandwidth optimization

---

## Documentation Accuracy (Multiple Inconsistencies)

- **Stated Goal**: README should accurately reflect implementation status.

- **Current State**: Several inconsistencies exist:
  - README claims "60% coverage" badge but actual coverage is 79.4%
  - README claims "6/521 (1%)" assets but 513 placeholder PNGs exist (not 6)
  - ROADMAP claims adventures are "not started" but 10 adventure packs exist
  - README notes guild system as "faction generation only" but full guild mechanics exist

- **Impact**: Users get incorrect expectations about project maturity and capabilities. Contributors may duplicate work that's already complete.

- **Closing the Gap**:
  1. Update coverage badge to reflect actual 79.4%
  2. Clarify asset badge: "513 placeholders / 0 real assets"
  3. Update ROADMAP adventure status to "⚠️ Data complete, RPC broken"
  4. Update guild system status to "✅ Full guild mechanics implemented"

---

## Validation Layer Completeness

- **Stated Goal**: The validation layer (`pkg/validation/`) should validate all RPC methods before processing.

- **Current State**: The `registerValidators()` function in `pkg/validation/validation.go` is missing validators for:
  - `adventure.list` (CRITICAL - blocks adventure feature)
  - `adventure.load` (CRITICAL - blocks adventure feature)

  Additionally, there may be other methods in `pkg/server/constants.go` that are registered in the method registry but not in the validation layer.

- **Impact**: Any RPC method without a validator will fail with "unknown method" even if the handler exists. This creates a silent failure mode where features appear broken.

- **Closing the Gap**:
  1. Audit `pkg/server/constants.go` against `pkg/validation/validation.go` to find all missing validators
  2. Add a CI test that verifies every registered method has a corresponding validator
  3. Create `pkg/server/handler_coverage_test.go` to enforce parity

---

## Summary Table

| Gap | Severity | Effort to Close |
|-----|----------|-----------------|
| Adventure RPC Validation | CRITICAL | 10 min (2 lines of code) |
| Asset Generation | HIGH | 4-6 hours (external tool setup) |
| Vault Provider | MEDIUM | 4-8 hours (implementation) or 30 min (removal) |
| GUI Editors | LOW | 2-4 weeks (new feature development) |
| Documentation Accuracy | LOW | 1 hour (README updates) |
| Validation Layer Audit | MEDIUM | 2-4 hours (audit + tests) |
