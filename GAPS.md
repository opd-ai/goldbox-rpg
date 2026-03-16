# Implementation Gaps — 2026-03-16

This document identifies gaps between stated goals in project documentation and the current implementation.

## Executive Summary

**Overall Status: Excellent**

The GoldBox RPG Engine demonstrates strong alignment between stated goals and implementation. No critical gaps were identified. All 17 stated features from the README are fully implemented and functional. The gaps documented below are primarily maintenance and quality improvements rather than missing functionality.

---

## Test Coverage Gaps

### cmd/quest-builder Coverage

- **Stated Goal**: Maintain ≥60% test coverage across all packages
- **Current State**: 71.6% coverage — lowest among CLI tools, though above threshold
- **Impact**: Reduced confidence in quest chain validation edge cases; potential for undiscovered bugs in quest prerequisite logic
- **Closing the Gap**: 
  1. Add table-driven tests for `run()` function edge cases in `cmd/quest-builder/main.go:46`
  2. Test invalid YAML input handling paths
  3. Test quest chain validation with circular dependencies
  4. Target: 80%+ coverage
  5. Validation: `go test ./cmd/quest-builder/... -cover`

### cmd/server Coverage

- **Stated Goal**: Maintain ≥60% test coverage across all packages
- **Current State**: 71.8% coverage — main server entry point
- **Impact**: Bootstrap and initialization paths have less test coverage; potential for startup configuration bugs
- **Closing the Gap**:
  1. Add integration tests for `initializeBootstrapGame()` in `cmd/server/main.go:47`
  2. Test configuration loading with various environment variable combinations
  3. Test graceful shutdown paths
  4. Target: 80%+ coverage
  5. Validation: `go test ./cmd/server/... -cover`

### scripts Package Coverage

- **Stated Goal**: Maintain ≥60% test coverage across all packages
- **Current State**: 68.8% coverage — utility scripts
- **Impact**: Minimal; scripts are build-time utilities not runtime code
- **Closing the Gap**: 
  1. Add tests for untested script functions if they're used in CI/CD pipelines
  2. Low priority unless scripts are modified
  3. Validation: `go test ./scripts/... -cover`

---

## Code Quality Gaps

### Handler Registration Pattern

- **Stated Goal**: Follow Go best practices and coding standards (from Contributing Guidelines)
- **Current State**: Handler registration in `pkg/server/server.go:1026-1100` uses 70+ lines of repetitive assignment statements
- **Impact**: Maintenance burden when adding new RPC methods; increased risk of copy-paste errors
- **Closing the Gap**:
  1. Refactor to table-driven registration pattern:
     ```go
     var handlers = []struct {
         method  string
         handler func(json.RawMessage) (interface{}, error)
     }{
         {MethodJoinGame, s.handleJoinGame},
         {MethodCreateCharacter, s.handleCreateCharacter},
         // ...
     }
     for _, h := range handlers {
         s.methodRegistry[h.method] = h.handler
     }
     ```
  2. Create new file `pkg/server/handlers_registration.go`
  3. Validation: `go-stats-generator analyze . --skip-tests | grep -A5 "Duplication"`

### Oversized Files

- **Stated Goal**: Maintainable, well-organized codebase
- **Current State**: Several files exceed 500 lines:
  - `pkg/server/handlers.go`: 1,187 lines, 56 functions
  - `pkg/pcg/metrics.go`: 687 lines
  - `pkg/pcg/world.go`: 834 lines (with tests: 652 additional)
  - `pkg/pcg/validator.go`: 1,020 lines
- **Impact**: Longer files are harder to navigate and understand; may slow down code reviews
- **Closing the Gap**:
  1. Consider splitting `pkg/server/handlers.go` by RPC category when making related changes:
     - `handlers_character.go` (character management)
     - `handlers_combat.go` (combat actions)
     - `handlers_quest.go` (quest system)
     - `handlers_pcg.go` (procedural generation) — already exists
  2. Low priority; current organization is functional
  3. Validation: `wc -l pkg/server/handlers*.go`

---

## Documentation Gaps

### Content Duration Claim

- **Stated Goal**: README claims "30+ hours of content" in embedded adventures
- **Current State**: 10 adventure packs exist; actual playable duration unverified
- **Impact**: Users may have incorrect expectations about content volume
- **Closing the Gap**:
  1. Audit adventure packs for estimated playtime (quest count × average completion time)
  2. Update README with verified duration or remove specific hour claim
  3. Validation: Manual review of `data/adventures/*/adventure.yaml` quest counts

### Adventure Count Terminology

- **Stated Goal**: README states "10 complete adventure packs"
- **Current State**: Exactly 10 adventures exist (verified), not "10+"
- **Impact**: Minimal; current count matches claim exactly
- **Closing the Gap**: No action needed; documentation is accurate

---

## Dependency Gaps

### Test Infrastructure Dependency

- **Stated Goal**: Use actively maintained dependencies
- **Current State**: `gorilla/websocket v1.5.3` retained for E2E tests despite being archived since 2022
- **Impact**: No security impact (test-only); potential future compatibility issues
- **Closing the Gap**:
  1. Migrate E2E test client in `test/e2e/client.go` to use `github.com/coder/websocket`
  2. Remove gorilla/websocket from go.mod
  3. Low priority; documented in go.mod comments
  4. Validation: `go mod graph | grep gorilla`

### govulncheck Compatibility

- **Stated Goal**: CI includes govulncheck security scanning
- **Current State**: Local govulncheck fails due to Go version mismatch (project: 1.25.6, local: 1.23)
- **Impact**: Cannot verify security scan locally; must rely on CI
- **Closing the Gap**: 
  1. Update local Go installation to 1.25.6+ 
  2. Or run via Docker: `docker run -v $(pwd):/app -w /app golang:1.25 govulncheck ./...`
  3. Validation: `govulncheck ./...` exits cleanly

---

## Missing Feature Gaps

### Keyboard Shortcuts in Editors

- **Stated Goal**: Visual editors documented in `docs/EDITOR_GUIDE.md` mention keyboard shortcuts
- **Current State**: `docs/EDITOR_GUIDE.md:173-183` lists shortcuts as "To be implemented in future versions"
- **Impact**: Reduced usability for power users; documentation describes non-functional feature
- **Closing the Gap**:
  1. Implement keyboard shortcuts in `pkg/wasmui/editor.go` and `pkg/wasmui/quest_editor.go`
  2. Or update documentation to remove planned feature section
  3. Low priority; core editing functionality works via mouse/forms
  4. Validation: Manual testing of editor pages

---

## Gap Summary Table

| Gap ID | Category | Severity | Effort | Recommendation |
|--------|----------|----------|--------|----------------|
| TC-1 | Test Coverage | HIGH | Medium | Add quest-builder edge case tests |
| TC-2 | Test Coverage | HIGH | Medium | Add server startup tests |
| TC-3 | Test Coverage | LOW | Low | Add scripts tests (optional) |
| CQ-1 | Code Quality | MEDIUM | Low | Refactor handler registration |
| CQ-2 | Code Quality | LOW | Medium | Split oversized files (optional) |
| DOC-1 | Documentation | LOW | Low | Verify content duration claim |
| DEP-1 | Dependencies | LOW | Medium | Migrate test WebSocket client |
| DEP-2 | Dependencies | LOW | Low | Update local Go version |
| FTR-1 | Features | LOW | Medium | Implement editor shortcuts (or remove docs) |

---

## Conclusion

The GoldBox RPG Engine has **no critical implementation gaps**. All documented features are functional and implemented. The gaps identified are:

1. **Test coverage improvements** (2 HIGH priority items targeting 80%+ coverage)
2. **Code quality refinements** (optional refactoring for maintainability)
3. **Minor documentation accuracy** (content duration verification)
4. **Optional dependency cleanup** (test infrastructure modernization)
5. **One documented-but-unimplemented feature** (keyboard shortcuts in editors)

The project exceeds its stated 60% coverage threshold with 82.9% overall coverage. All 17 README feature claims are verified as implemented. The codebase demonstrates good engineering practices with zero `go vet` issues, zero race conditions, zero circular dependencies, and comprehensive documentation coverage (89.9%).

**Recommended Priority Order:**
1. TC-1: Quest-builder test coverage (HIGH)
2. TC-2: Server startup tests (HIGH)
3. CQ-1: Handler registration refactor (MEDIUM)
4. All other items: Address opportunistically during related work

---

*Generated: 2026-03-16*
