# Implementation Plan: Asset Pipeline Completion & Dependency Modernization

## Project Context
- **What it does**: A modern, Go-based RPG engine inspired by SSI Gold Box games providing character management, turn-based combat, procedural content generation, and real-time multiplayer via JSON-RPC and WebSocket APIs with an Ebitengine/WASM frontend.
- **Current goal**: Complete the visual asset generation pipeline and merge pending dependency updates to maintain security posture and framework compatibility.
- **Estimated Scope**: Small (~10 items requiring attention across 7 Dependabot PRs + asset distribution infrastructure)

## Goal-Achievement Status
| Stated Goal | Current Status | This Plan Addresses |
|-------------|---------------|---------------------|
| Asset generation pipeline (521 assets) | ⚠️ Partial - Pipeline complete, 500/521 placeholder assets exist | Yes - Asset distribution |
| WebSocket real-time communication | ✅ Achieved - Migrated to nhooyr.io/websocket | No - Already complete |
| Procedural Content Generation | ✅ Achieved - Full system operational | No - Already complete |
| Complete spell system (levels 0-9) | ✅ Achieved - 60 spells across 10 YAML files | No - Already complete |
| World editor tools | ⚠️ Partial - CLI tools exist, browser editor at `/editor` | No - Not critical path |
| Dependency security | ⚠️ Needs attention - 7 open Dependabot PRs | Yes - Dependency updates |
| Go version compatibility | ⚠️ Using Go 1.25.6, Docker base on 1.23 | Yes - Go 1.26 upgrade |

## Metrics Summary
- **Complexity hotspots on goal-critical paths**: 0 functions above threshold (max overall: 14.5, threshold: 15.0)
- **Duplication ratio**: 1.52% (956 lines duplicated, well below 5% threshold)
- **Doc coverage**: Not measured (focus is on functional completion, not documentation)
- **Package coupling**: Server package has expected high coupling (11 dependencies) as primary API layer
- **Test coverage**: 82.5% (exceeds 60% threshold by 22.5 percentage points)
- **Static analysis**: Zero `go vet` issues, zero race conditions detected
- **Code health**: Excellent - all complexity metrics within acceptable ranges

### Top Complexity Functions (All Below Threshold)
| Function | Package | Overall | Lines | Location |
|----------|---------|---------|-------|----------|
| Stop | e2e | 14.5 | 42 | test/e2e/server.go |
| ValidateAndFix | pcg | 14.2 | 46 | pkg/pcg/validator.go |
| parseConstDeclaration | main | 14.2 | 33 | cmd/openapi-gen/main.go |
| RunCellularAutomata | terrain | 14.0 | 40 | pkg/pcg/terrain/cellular_automata.go |
| run | main | 14.0 | 46 | cmd/quest-builder/main.go |

All high-complexity functions are in test/demo code or utilities, not production game logic.

## Implementation Steps

### Step 1: Merge Dependency Updates (Go Dependencies)
- **Deliverable**: Merge PR #39 (Go dependencies: kin-openapi v0.133.0→v0.134.0, ebiten v2.7.0→v2.9.9, logrus v1.9.3→v1.9.4, golang.org/x/time v0.12.0→v0.15.0)
- **Dependencies**: None
- **Goal Impact**: Addresses security and compatibility gap; ebiten v2.9.9 includes critical bug fixes for fullscreening/vector rendering
- **Acceptance**: PR #39 merged, all CI checks pass on master branch
- **Validation**: `go mod tidy && go test -race ./... && go vet ./...` all succeed

**Rationale**: Ebiten v2.9.9 includes critical fixes:
- Bug fix: app frozen by fullscreening (fixed in e5c8863)
- Bug fix: vector atlas size insufficient for antialias (fixed in 8c76fc9)
- Bug fix: CAMetalDisplayLink deadlock after sleep (fixed in 3c95370)
These directly impact the WASM frontend experience on macOS and with vector graphics.

### Step 2: Upgrade Docker Base Image to Go 1.26
- **Deliverable**: Merge PR #38 (Docker base image golang:1.23-bookworm → golang:1.26-bookworm)
- **Dependencies**: Step 1 (Go dependencies must support Go 1.26)
- **Goal Impact**: Aligns Docker environment with latest Go toolchain, ensures consistent build environment
- **Acceptance**: PR #38 merged, Docker build succeeds, health checks pass
- **Validation**: `docker build -t goldbox-rpg . && docker run --rm -d -p 8080:8080 goldbox-rpg && sleep 10 && curl http://localhost:8080/health` returns HTTP 200

**Note**: Current codebase uses Go 1.25.6 with toolchain 1.25.8 locally. After this step, consider bumping `go.mod` to require Go 1.26 minimum for consistency.

### Step 3: Update GitHub Actions to Node 24 Runtime
- **Deliverable**: Merge PRs #37, #36, #35, #34, #33 (actions/attest-build-provenance v2→v4, golangci/golangci-lint-action v6→v9, actions/upload-artifact v4→v7, docker/setup-buildx-action v3→v4, docker/build-push-action v5→v7)
- **Dependencies**: None (independent of Go version)
- **Goal Impact**: GitHub Actions runners are deprecating Node 20 runtime; this prevents future CI breakage
- **Acceptance**: All 5 PRs merged, CI workflow completes successfully on master
- **Validation**: Monitor GitHub Actions workflow runs for at least 2 successful builds post-merge

**Context**: GitHub announced Node 20 deprecation on 2025-09-19. These updates migrate to Node 24 runtime (requires Actions Runner v2.327.1+).

### Step 4: Create Pre-Generated Asset Distribution Package
- **Deliverable**: GitHub Release with `goldbox-rpg-assets-v1.0.0.tar.gz` containing all 521 final assets
- **Dependencies**: None (asset generation is external to code pipeline)
- **Goal Impact**: Closes the gap between 500 placeholder assets and 521 required assets; provides turnkey experience for users without AI image generation tools
- **Acceptance**: Release published with SHA256 checksum, asset pack extracts to `web/static/assets/` with correct directory structure
- **Validation**: `make assets-download && find web/static/assets -type f -name "*.png" | wc -l` returns 521

**Asset Pack Structure**:
```
goldbox-rpg-assets-v1.0.0.tar.gz
├── characters/       # Character portraits (60 assets)
├── monsters/         # Monster sprites (120 assets)
├── items/            # Item icons (150 assets)
├── terrain/          # Terrain tiles (80 assets)
├── effects/          # Combat effects (60 assets)
├── ui/               # UI elements (51 assets)
└── SHA256SUMS        # Checksums for verification
```

### Step 5: Implement Assets Download Script
- **Deliverable**: `scripts/download-assets.sh` script that fetches pre-generated asset pack from GitHub Releases
- **Dependencies**: Step 4 (release must exist to download from)
- **Goal Impact**: Enables `make assets-download` to pull production-ready assets without 4-6 hour AI generation step
- **Acceptance**: Script downloads, verifies SHA256 checksum, extracts to correct location, sets file permissions
- **Validation**: 
```bash
make assets-clean && make assets-download && make assets-verify
# Expected output: "All 521 required assets present and valid"
```

**Script Requirements**:
1. Download from latest GitHub Release matching pattern `goldbox-rpg-assets-v*.tar.gz`
2. Verify SHA256 checksum before extraction
3. Extract to `web/static/assets/` preserving directory structure
4. Set read-only permissions on extracted assets (prevent accidental modification)
5. Exit with error code on checksum mismatch or download failure

### Step 6: Update README Asset Status Badge
- **Deliverable**: Update README.md badge from "252 placeholders / 521 defined" to "521 ready (download) / 521 defined"
- **Dependencies**: Step 5 (download mechanism must exist)
- **Goal Impact**: Clarifies to users that assets are ready via `make assets-download`, reducing confusion about asset generation requirements
- **Acceptance**: README badge reflects new status, asset generation section documents three workflows (download, priority, full)
- **Validation**: Visual inspection of README.md rendering on GitHub

**New Asset Section**:
```markdown
### Asset Availability
- **Quick Start**: `make assets-download` - Download 521 pre-generated assets (~50MB, 30 seconds)
- **Priority Assets**: `make assets-priority` - Generate 50 high-priority assets (~30 minutes, requires AI tool)
- **Full Generation**: `make assets` - Generate all 521 assets from scratch (~4-6 hours, requires AI tool setup per [ASSET_INTEGRATION.md](./ASSET_INTEGRATION.md))
```

### Step 7: Add Asset Verification to CI Pipeline
- **Deliverable**: Extend `.github/workflows/ci.yml` to run `make assets-download && make assets-verify` in CI
- **Dependencies**: Step 5 (download script must be functional)
- **Goal Impact**: Ensures asset pack downloads remain functional as releases evolve; catches broken download URLs early
- **Acceptance**: CI job "asset-verification" succeeds, verifies all 521 assets present and valid
- **Validation**: Monitor CI run on PR that modifies asset configuration

**CI Job Definition**:
```yaml
asset-verification:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - name: Download assets
      run: make assets-download
    - name: Verify asset integrity
      run: make assets-verify
    - name: Check asset count
      run: |
        count=$(find web/static/assets -type f -name "*.png" | wc -l)
        if [ "$count" -ne 521 ]; then
          echo "Expected 521 assets, found $count"
          exit 1
        fi
```

### Step 8: Document go.mod Go Version Alignment
- **Deliverable**: Update `go.mod` to require Go 1.26.0 minimum (align with Docker base image from Step 2)
- **Dependencies**: Step 2 (Docker base must be on Go 1.26)
- **Goal Impact**: Ensures local development, Docker builds, and CI all use consistent Go version; prevents "works on my machine" issues
- **Acceptance**: `go.mod` specifies `go 1.26.0`, all builds succeed with Go 1.26 toolchain
- **Validation**: `go version` returns 1.26+, `go build ./cmd/server` succeeds, `go test -race ./...` passes

**Change**:
```diff
-go 1.25.6
+go 1.26.0

-toolchain go1.25.8
+toolchain go1.26.2
```

**Impact**: Developers using Go 1.25.x will need to upgrade. Document in CHANGELOG.md as breaking change requiring Go 1.26+ for builds.

### Step 9: Reduce Server Handler Registration Duplication
- **Deliverable**: Extract handler registration pattern in `pkg/server/server.go:1027-1106` into table-driven registration
- **Dependencies**: None (internal refactoring)
- **Goal Impact**: Reduces duplication from 1.52% to <1.0% (target: 956 → <500 duplicated lines); simplifies addition of new RPC handlers
- **Acceptance**: Duplication ratio drops below 1.0%, all existing handler tests pass unchanged
- **Validation**: `go-stats-generator analyze . --skip-tests --format json --sections duplication | jq '.duplication.summary.ratio'` returns <1.0

**Current Pattern** (35 lines duplicated 2x):
```go
s.server.HandleFunc(constants.MethodCastSpell, func(w http.ResponseWriter, r *http.Request) {
    var req CastSpellRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        // error handling...
    }
    // ... validation and business logic ...
})
```

**Proposed Pattern**:
```go
type handlerRegistration struct {
    method  string
    handler func(http.ResponseWriter, *http.Request)
}

var handlers = []handlerRegistration{
    {constants.MethodCastSpell, s.handleCastSpell},
    {constants.MethodMove, s.handleMove},
    // ... all other handlers ...
}

for _, h := range handlers {
    s.server.HandleFunc(h.method, h.handler)
}
```

**Files Modified**: `pkg/server/server.go`, `pkg/server/handlers_registration.go` (new)

### Step 10: Update CHANGELOG.md with Release Notes
- **Deliverable**: Add comprehensive changelog entry documenting all changes from Steps 1-9
- **Dependencies**: Steps 1-8 complete (all work finished)
- **Goal Impact**: Provides users and contributors clear understanding of what changed and why; documents breaking change (Go 1.26 requirement)
- **Acceptance**: CHANGELOG.md includes version entry with all changes categorized (Added, Changed, Removed, Security, Breaking)
- **Validation**: Manual review for completeness and clarity

**Changelog Entry Structure**:
```markdown
## [Unreleased]

### Breaking Changes
- **Minimum Go Version**: Now requires Go 1.26.0+ for builds (was Go 1.25.6)

### Added
- Asset distribution via GitHub Releases (521 pre-generated assets available via `make assets-download`)
- CI verification for asset download integrity
- Table-driven handler registration in server package

### Changed
- Upgraded Ebiten from v2.7.0 to v2.9.9 (includes critical fullscreen/vector rendering fixes)
- Upgraded kin-openapi from v0.133.0 to v0.134.0
- Upgraded logrus from v1.9.3 to v1.9.4
- Upgraded golang.org/x/time from v0.12.0 to v0.15.0
- Docker base image from golang:1.23-bookworm to golang:1.26-bookworm
- GitHub Actions runtime from Node 20 to Node 24
- Reduced code duplication from 1.52% to <1.0% (handler registration refactor)

### Removed
- Placeholder asset generation step (replaced with downloadable asset pack)

### Security
- All dependencies updated to latest versions addressing known vulnerabilities
```

## Technical Debt Addressed (Secondary Benefits)

1. **Code Duplication**: Handler registration refactor (Step 9) reduces duplication by ~456 lines
2. **Runtime Consistency**: Go 1.26 alignment (Step 8) ensures Docker, CI, and local dev use same toolchain
3. **CI Resilience**: Node 24 migration (Step 3) prevents future GitHub Actions deprecation breakage
4. **User Experience**: Asset download (Steps 4-5) removes 4-6 hour barrier for new developers

## Testing Strategy

### Unit Tests
- All existing tests must pass after each step: `go test -race ./...`
- No new business logic introduced, so no new unit tests required

### Integration Tests
- Docker build verification (Step 2): Full container build + health check
- Asset download verification (Step 5): End-to-end asset fetch + checksum validation
- CI pipeline verification (Step 7): GitHub Actions workflow success

### Regression Testing
- After Step 9 (handler refactor): Run full E2E test suite `go test ./test/e2e/...`
- Verify all 72 RPC methods still route correctly via manual smoke test of frontend

## Rollback Strategy

- **Steps 1-3** (Dependencies/CI): Revert merged PRs if CI failures detected post-merge
- **Step 4** (Asset release): Delete GitHub Release if asset pack corrupted/invalid
- **Step 5** (Download script): Revert commit if download mechanism fails in CI
- **Step 9** (Handler refactor): Revert commit, use git bisect to identify breakage if tests fail

## Success Metrics

| Metric | Baseline | Target | Validation Command |
|--------|----------|--------|-------------------|
| Duplication Ratio | 1.52% | <1.0% | `go-stats-generator analyze . --skip-tests \| jq '.duplication.summary.ratio'` |
| Asset Count | 500 placeholders | 521 final | `find web/static/assets -name "*.png" \| wc -l` |
| CI Success Rate | N/A (Node 20) | 100% (Node 24) | Monitor GitHub Actions dashboard |
| Docker Build Time | Baseline | <5% increase | `time docker build .` |
| Test Coverage | 82.5% | ≥80% maintained | `go test ./... -coverprofile=c.out && go tool cover -func=c.out \| grep total` |

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Ebiten v2.9.9 breaks WASM frontend | Low | High | Rollback to v2.7.0 if rendering issues detected; extensive testing on macOS/browser |
| Go 1.26 incompatibility | Low | Medium | All dependencies already compatible; gradual rollout via Docker first |
| Asset download bandwidth costs | Low | Low | Serve from GitHub Releases (free for open source); consider CDN if needed |
| Handler refactor introduces routing bug | Medium | High | Comprehensive E2E testing before merge; gradual rollout with feature flag |

## Timeline Estimate

- **Steps 1-3** (Dependency merges): 1-2 hours (PR review + CI monitoring)
- **Step 4** (Asset pack creation): 4-6 hours (one-time AI generation + packaging)
- **Step 5** (Download script): 2-3 hours (scripting + error handling)
- **Steps 6-7** (Documentation/CI): 1-2 hours (README updates + workflow changes)
- **Step 8** (Go version alignment): 30 minutes (go.mod edit + verification)
- **Step 9** (Handler refactor): 3-4 hours (refactoring + testing)
- **Step 10** (Changelog): 30 minutes (documentation)

**Total Estimated Effort**: 12-18 hours of developer time

## Post-Completion State

After completing this plan:
- ✅ All 521 visual assets available via `make assets-download` (~30 seconds vs 4-6 hours)
- ✅ All dependencies current as of 2026-03-16
- ✅ GitHub Actions using Node 24 runtime (future-proof until ~2027)
- ✅ Docker, CI, and local dev all on Go 1.26 (consistent toolchain)
- ✅ Code duplication <1.0% (improved maintainability)
- ✅ No impact on 82.5% test coverage (maintained)
- ✅ Zero complexity regressions (all functions still <15.0)

The project will be in a state where:
1. New contributors can get started in <5 minutes (`git clone && make assets-download && make run`)
2. All security patches applied (dependencies current)
3. CI pipeline resilient to GitHub platform changes (Node 24)
4. Codebase remains clean and maintainable (low duplication)

## Notes

- **Asset Generation Still Supported**: The full `make assets` workflow remains for users who want to customize visual style via their own AI models
- **No Breaking API Changes**: All RPC methods, WebSocket protocol, and game data schemas remain unchanged
- **Test Coverage Maintained**: Target ≥80% coverage (currently 82.5%) must be preserved throughout
- **Backward Compatibility**: Existing save games, adventures, and configuration files require no migration

---

*Generated: 2026-03-16 using go-stats-generator v1.0.0*  
*Metrics Baseline: 31,409 LOC, 621 functions, 1744 methods, 19 packages*  
*Complexity: Max 14.5 (well below 15.0 threshold)*  
*Duplication: 1.52% (956 lines, target <1.0%)*  
*Test Coverage: 82.5% (exceeds 60% requirement)*
