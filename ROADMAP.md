# Production Readiness Roadmap

## Executive Summary
**Current Status**: Approaching production ready - 4 critical gaps identified, strong foundation established  
**Estimated Total Effort**: 6-8 weeks  
**Priority Issues**: 4 Critical | 8 High | 12 Medium | 9 Low

**Key Strengths**:
- Excellent observability and monitoring (health checks, Prometheus metrics, structured logging)
- Robust resilience patterns (circuit breakers, retry mechanisms, rate limiting)
- Comprehensive input validation framework
- Strong test coverage (78%, 106 test files)
- Thread-safe concurrent operations with proper mutex usage
- Good documentation (README, API docs, package documentation)

**Critical Blockers**:
- No CI/CD pipeline for automated testing and deployment
- Missing data persistence layer (in-memory only)
- No deployment manifests (Kubernetes, Docker Compose)
- Incomplete secrets management

## Assessment Overview

### Production Readiness Score: 68/100

| Category | Score | Status | Notes |
|----------|-------|--------|-------|
| Architecture & Code Quality | 12/15 | 🟢 | Well-structured, Go idioms, minor tech debt |
| Testing & Quality Assurance | 11/15 | 🟡 | 78% coverage, missing E2E tests |
| Error Handling & Resilience | 13/15 | 🟢 | Circuit breakers, retries, good patterns |
| Observability & Monitoring | 14/15 | 🟢 | Excellent metrics, health checks, logging |
| Security | 8/15 | 🔴 | Input validation strong, secrets mgmt weak |
| Documentation | 8/10 | 🟢 | Good docs, needs deployment guide |
| Deployment & Operations | 2/15 | 🔴 | No CI/CD, no K8s manifests, no persistence |

🔴 Critical gaps | 🟡 Needs improvement | 🟢 Production ready

---

## Phase 1: Critical Issues (Must Fix Before Production)
**Timeline**: Week 1-2  
**Goal**: Resolve blocking issues for production deployment

### 1.1 Implement Data Persistence Layer
**Priority**: Critical  
**Effort**: 3 days  
**Owner**: Backend Team

**Problem**: The `GameState` struct in `pkg/server/state.go` (lines 35-53) stores all game state in memory - `WorldState`, `TurnManager`, `TimeManager`, and `Sessions` map. The `Character` struct in `pkg/game/character.go` (lines 41-83) with its YAML tags is already designed for serialization but never persisted. Data is lost on restart.

**Current Code Analysis**:
- `GameState` already has YAML tags: `state_world`, `state_turns`, `state_time`, `state_sessions`
- `Character` struct has complete YAML tags for all fields (`char_id`, `attr_strength`, `combat_current_hp`, etc.)
- `game.World` in `WorldState` contains `Objects map[string]GameObject`
- Sessions stored in memory-only map: `Sessions map[string]*PlayerSession`
- No persistence hooks in `AddPlayer()` (line 56), `GetState()` (line 79), or state mutation methods

**Impact**: 
- Player characters with all attributes, equipment, inventory lost on restart
- Active sessions and world state (`WorldState.Objects` map) not persisted
- Turn order and game time tracking reset
- Blocks production as stated in current `cmd/server/main.go` graceful shutdown (lines 147-164) with no save

**Solution**:
- [ ] Create `pkg/persistence/filestore.go` implementing:
  ```go
  type FileStore struct {
      dataDir string
      mu sync.RWMutex
  }
  func (fs *FileStore) SaveGameState(gs *GameState) error
  func (fs *FileStore) LoadGameState() (*GameState, error)
  func (fs *FileStore) SaveCharacter(c *Character) error
  func (fs *FileStore) LoadCharacter(id string) (*Character, error)
  ```
- [ ] Leverage existing YAML tags - use `gopkg.in/yaml.v3` (already in go.mod) to marshal/unmarshal
- [ ] Add file locking using `syscall.Flock` in `pkg/persistence/lock.go`
- [ ] Implement atomic writes: write to `{file}.tmp` then `os.Rename()` to final location
- [ ] Hook persistence into `GameState` methods:
  - Add `SaveToFile()` method to `GameState` (after line 35 struct definition)
  - Call auto-save in state mutation methods: `AddPlayer()`, session updates
  - Add `LoadFromFile()` to restore state in `NewRPCServer()` (pkg/server/server.go:256)
- [ ] Extend `Config` struct (pkg/config/config.go:19) with:
  ```go
  DataDir string `json:"data_dir"`  // Default: "./data"
  AutoSaveInterval time.Duration `json:"auto_save_interval"`  // Default: 30s
  ```
- [ ] Update graceful shutdown (cmd/server/main.go:148) to call `gs.SaveToFile()` before exit
- [ ] Store files:
  - GameState → `data/gamestate.yaml` (includes WorldState, Sessions)
  - Individual Characters → `data/characters/{char_id}.yaml` (backup/recovery)
  - Session snapshots → `data/sessions/{session_id}.yaml` (optional for debugging)

**Success Criteria**:
- `GameState` with all fields persists across restarts via YAML
- Character data (all 20+ fields: Strength, Equipment, Inventory, Position, etc.) saves correctly
- File locking prevents corruption during concurrent writes to same character file
- Atomic writes (`tmp` + `rename`) prevent partial file corruption
- `NewRPCServer()` loads existing `GameState` if `data/gamestate.yaml` exists
- Graceful shutdown in `main.go` saves state before exit
- Health check in `pkg/server/health.go` verifies data directory writable

**Files to Create**:
- `pkg/persistence/filestore.go` - `FileStore` struct with `Save/Load` methods
- `pkg/persistence/lock.go` - `FileLock` wrapper for `syscall.Flock`
- `pkg/persistence/atomic.go` - `AtomicWriteFile()` helper

**Files to Modify**:
- `pkg/server/state.go` - Add `SaveToFile()`, `LoadFromFile()` methods to `GameState`
- `pkg/server/server.go` - Call `LoadFromFile()` in `NewRPCServer()` (line ~256)
- `pkg/config/config.go` - Add `DataDir` and `AutoSaveInterval` fields to `Config` struct (after line 42)
- `pkg/server/health.go` - Add data directory writable check to health checks
- `cmd/server/main.go` - Call `gameState.SaveToFile()` in `performGracefulShutdown()` (before line 162)
- `.gitignore` - Add `/data/gamestate.yaml`, `/data/characters/*.yaml`, `/data/sessions/*.yaml`

**Code References**:
- GameState struct: `pkg/server/state.go:35-53`
- Character struct: `pkg/game/character.go:41-83`
- NewRPCServer: `pkg/server/server.go:256`
- Config struct: `pkg/config/config.go:19-94`
- Graceful shutdown: `cmd/server/main.go:148-164`

---

### 1.2 Establish CI/CD Pipeline
**Priority**: Critical  
**Effort**: 3 days  
**Owner**: DevOps Team

**Problem**: No CI/CD exists - `.github/workflows/` directory is empty. Manual testing of 93 Go files (40k LOC) is error-prone. No automated verification of:
- `Makefile` targets (`build`, `test`, `fmt` work correctly)
- Race conditions with `go test -race ./...`
- Current test coverage 78% (no enforcement)
- Dependency vulnerabilities
- `Dockerfile` builds successfully

**Impact**:
- Cannot catch regressions in Character, GameState, or RPC handlers before merge
- No security scanning for dependencies (Prometheus v1.22.0, etc.)
- Risk deploying broken Dockerfile to production
- Slow feedback - developers must manually run `make test`, `make fmt`

**Solution**:

**A. `.github/workflows/ci.yml`** - Pull Request CI (NEW FILE, ~100 lines):
```yaml
name: CI
on: [pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      # Run existing Makefile targets
      - run: make fmt  # Uses gofumpt (line 36 in Makefile)
      - run: make build  # Builds cmd/server/main.go
      - run: go test ./... -v -race -coverprofile=coverage.out
      - run: go tool cover -func=coverage.out
      # Enforce 80% coverage (Phase 2.1 targets 85%)
      - name: Check coverage
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          if (( $(echo "$COVERAGE < 80" | bc -l) )); then
            echo "Coverage $COVERAGE% below 80% threshold"
            exit 1
          fi
      # Run security scans
      - run: go install golang.org/x/vuln/cmd/govulncheck@latest
      - run: govulncheck ./...
      - run: go list -m -u all  # Check for updates
      
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: golangci/golangci-lint-action@v6
        with:
          version: latest
          
  docker:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: docker build -t goldbox-rpg:test .
      - run: docker run -d -p 8080:8080 goldbox-rpg:test
      - run: sleep 5 && curl -f http://localhost:8080/health || exit 1
```

**B. `.github/workflows/build.yml`** - Main Branch Build (NEW FILE):
- Build and push Docker image to GitHub Container Registry
- Tag with commit SHA and `latest`
- Store as artifact for deployment

**C. `.golangci.yml`** - Linter Configuration (NEW FILE, ~50 lines):
```yaml
linters:
  enable:
    - gofmt
    - govet
    - staticcheck
    - gosimple
    - unused
    - errcheck
    - ineffassign
linters-settings:
  errcheck:
    check-blank: true  # Catch ignored errors
run:
  timeout: 5m
  skip-dirs:
    - vendor
```

**D. Update `README.md`** - Add Status Badges:
```markdown
![Build Status](https://github.com/opd-ai/goldbox-rpg/workflows/CI/badge.svg)
![Coverage](https://img.shields.io/badge/coverage-78%25-yellow)
```

- [ ] Create `.github/workflows/ci.yml` - runs on every PR
- [ ] Create `.github/workflows/build.yml` - runs on main branch
- [ ] Create `.golangci.yml` - linter config
- [ ] Configure branch protection in GitHub repo settings:
  - Require PR reviews before merge
  - Require CI checks to pass
  - Prevent direct pushes to `main`
- [ ] Set up GitHub Container Registry (ghcr.io)
- [ ] Update `README.md` with build badges (line ~3)
- [ ] Test CI by opening a test PR

**Success Criteria**:
- CI runs automatically on every PR
- Tests with `-race` flag catch concurrency issues in Character/GameState
- Coverage enforcement fails PRs below 80% (later increase to 85% after Phase 2.1)
- `govulncheck` scans dependencies (Prometheus, Gorilla WebSocket, etc.)
- Dockerfile builds successfully
- Build badges show status in README.md
- CI pipeline completes in <5 minutes
- Branch protection prevents bypassing checks

**Files to Create**:
- `.github/workflows/ci.yml` - CI pipeline for PRs, ~100 lines
- `.github/workflows/build.yml` - Build/push Docker images, ~60 lines
- `.golangci.yml` - Linter configuration, ~50 lines

**Files to Modify**:
- `README.md` - Add status badges near top (after line 3)
- GitHub repo settings - Enable branch protection (not in code)

**Code References**:
- Existing Makefile targets: Lines 1-10 (`build`, `test`, `fmt`)
- Dockerfile: Root directory, line 14 builds with CGO_ENABLED=0
- Test coverage script: `scripts/analyze_test_coverage.sh`
- All PRs require passing CI checks
- Test coverage tracked and enforced
- Security vulnerabilities detected automatically
- Docker images built and pushed automatically
- Deployment process documented and automated
- Build status visible in README

**Files to Create**:
- `.github/workflows/ci.yml` - CI pipeline
- `.github/workflows/cd.yml` - CD pipeline
- `.github/workflows/security.yml` - Security scanning
- `.github/dependabot.yml` - Dependency updates
- `.golangci.yml` - Linter configuration

---

### 1.3 Implement Secrets Management
**Priority**: Critical  
**Effort**: 2 days  
**Owner**: Security Team

**Problem**: No secrets management system implemented. Configuration uses environment variables, but there's no secure way to manage API keys, encryption keys, or other secrets in production.

**Impact**:
- Risk of hardcoded secrets (checked for, none found currently)
- No secret rotation capability
- Secrets visible in environment variables
- Compliance issues (SOC2, GDPR requirements)
- Cannot securely store sensitive configuration

**Solution**:
- [ ] Add secrets management library (HashiCorp Vault SDK or AWS Secrets Manager)
- [ ] Create secrets loading abstraction in `pkg/secrets/`
- [ ] Implement secret provider interface:
  ```go
  type SecretProvider interface {
      GetSecret(ctx context.Context, key string) (string, error)
      SetSecret(ctx context.Context, key, value string) error
      RotateSecret(ctx context.Context, key string) error
  }
  ```
- [ ] Add environment-based provider for development
- [ ] Add Vault/AWS provider for production
- [ ] Update Config to load sensitive values via secrets provider
- [ ] Document secret naming conventions
- [ ] Add secret rotation procedures to operations guide
- [ ] Implement secret health checks
- [ ] Update Dockerfile to support secrets mounting

**Success Criteria**:
- No secrets in environment variables or config files
- Secrets loaded from secure backend (Vault/AWS)
- Secret rotation documented and tested
- Audit log for secret access
- Development mode works with environment variables
- Production mode requires secrets backend

**Files to Create**:
- `pkg/secrets/provider.go` - Secret provider interface
- `pkg/secrets/vault.go` - Vault implementation
- `pkg/secrets/env.go` - Environment fallback
- `docs/SECRETS_MANAGEMENT.md` - Operations guide

**Files to Modify**:
- `pkg/config/config.go` - Integrate secrets loading
- `cmd/server/main.go` - Initialize secrets provider
- `Dockerfile` - Support secrets mounting

---

### 1.4 Create Deployment Manifests
**Priority**: Critical  
**Effort**: 3 days  
**Owner**: DevOps Team

**Problem**: No Kubernetes manifests, Docker Compose files, or deployment configurations exist. The application cannot be easily deployed to staging or production environments. The basic Dockerfile exists but lacks production hardening.

**Impact**:
- Cannot deploy to Kubernetes clusters
- No local development environment with all services
- Manual deployment is error-prone
- No rollback capability
- No load balancing or scaling configuration

**Solution**:
- [ ] Improve Dockerfile for production:
  - Multi-stage build to reduce image size
  - Use scratch or distroless base image
  - Run as non-root user (already done)
  - Add security scanning
  - Proper COPY order for layer caching
  - Add HEALTHCHECK (already exists, verify configuration)
- [ ] Create Docker Compose for local development:
  - Game server with mounted data volumes
  - Prometheus (metrics collection)
  - Grafana (metrics visualization)
  - Health check dependencies
- [ ] Create Kubernetes manifests in `deploy/k8s/`:
  - Deployment for game server (with resource limits)
  - Service (ClusterIP and LoadBalancer)
  - ConfigMap for configuration
  - Secret for sensitive data
  - HorizontalPodAutoscaler (HPA)
  - PodDisruptionBudget (PDB)
  - Ingress for HTTPS
  - NetworkPolicy for security
- [ ] Create Helm chart in `deploy/helm/`:
  - Chart.yaml with version management
  - values.yaml with configurable parameters
  - Templates for all K8s resources
  - README with installation instructions
- [ ] Document deployment procedures
- [ ] Create environment-specific configurations (dev, staging, prod)

**Success Criteria**:
- Docker image builds successfully with multi-stage
- Docker Compose starts full stack locally
- Kubernetes manifests deploy successfully
- Helm chart installs and upgrades cleanly
- Resource limits prevent runaway processes
- Health checks integrated with K8s probes
- Horizontal scaling works under load
- Deployment documented with examples

**Files to Create**:
- `Dockerfile.production` - Optimized production Dockerfile
- `docker-compose.yml` - Local development stack
- `docker-compose.prod.yml` - Production stack example
- `deploy/k8s/deployment.yaml` - K8s deployment
- `deploy/k8s/service.yaml` - K8s service
- `deploy/k8s/configmap.yaml` - Configuration
- `deploy/k8s/ingress.yaml` - Ingress rules
- `deploy/k8s/hpa.yaml` - Autoscaling
- `deploy/helm/goldbox-rpg/` - Helm chart
- `docs/DEPLOYMENT.md` - Deployment guide

**Files to Modify**:
- `Dockerfile` - Enhance for production
- `README.md` - Add deployment instructions

---

## Phase 2: High Priority Issues
**Timeline**: Week 3-4  
**Goal**: Improve stability, security, and operational readiness

### 2.1 Complete Test Coverage for Core Files
**Priority**: High  
**Effort**: 5 days  
**Owner**: Development Team

**Problem**: 20 critical Go files have NO test files (per `scripts/find_untested_files.sh`). Key missing tests:
- `pkg/game/character.go` (1,400+ lines) - NO `character_test.go` exists
- `pkg/game/effects.go` - Effect system untested  
- `pkg/server/handlers.go` (2,800+ lines) - NO `handlers_test.go`, RPC methods like `handleMove()` (line 45), `handleAttack()` (line 247) untested
- `pkg/server/server.go` - Server lifecycle untested
- `pkg/server/health.go` - Health check functions untested
- `pkg/pcg/manager.go` - PCG coordination untested

**Current Coverage**: 78% overall (73 files tested, 20 untested)

**Impact**:
- Cannot verify `Character` mutex safety (lines 42, 95 use RWMutex but no race tests)
- RPC handlers (`handleMove`, `handleAttack`, etc.) error paths untested
- Health checks (`checkServer()`, `checkGameState()`) in `pkg/server/health.go` never verified
- Risk of breaking existing functionality when adding persistence layer

**Solution**:

**A. `pkg/game/character_test.go`** (NEW FILE, ~500 lines):
```go
func TestCharacter_Clone(t *testing.T) {
    // Test Clone() method (line 95) with all fields
    // Verify mutex independence, Equipment map copy, Inventory slice copy
}
func TestCharacter_ConcurrentAccess(t *testing.T) {
    // Test with go test -race
    // Concurrent SetPosition(), AddItem(), EquipItem()
}
func TestCharacter_SetPosition(t *testing.T) {
    // Test line ~400 SetPosition() method
}
func TestCharacter_EquipItem(t *testing.T) {
    // Test Equipment map, slot conflicts, class restrictions
}
```

**B. `pkg/server/handlers_test.go`** (NEW FILE, ~800 lines):
```go
func TestHandleMove(t *testing.T) {
    // Test handleMove() (line 45) with valid/invalid directions
    // Test parseMoveRequest() (line 89)
    // Test session validation via getSessionForMove() (line 55)
}
func TestHandleAttack(t *testing.T) {
    // Test handleAttack() (line 247) combat calculations
    // Test validateAttackParams(), executeAttack()
}
func TestHandleCastSpell(t *testing.T) {
    // Test spell casting RPC method
}
func TestHandleEquipItem(t *testing.T) {
    // Test equipment RPC method
}
// Add tests for all 20+ RPC handler functions
```

**C. `pkg/server/health_test.go`** (NEW FILE, ~300 lines):
```go
func TestHealthChecker_CheckServer(t *testing.T) {
    // Test checkServer() function
}
func TestHealthChecker_CheckGameState(t *testing.T) {
    // Test checkGameState() validation
}
func TestHealthChecker_RunHealthChecks(t *testing.T) {
    // Test full health check aggregation
}
```

**D. `pkg/game/effects_test.go`** (NEW FILE, ~400 lines):
- Test effect stacking, priority, expiration
- Test immunity checks in `EffectManager`

**E. `pkg/server/server_test.go`** (NEW FILE, ~300 lines):
- Test `NewRPCServer()` initialization (line 256)
- Test graceful shutdown hooks
- Test session management

**F. `pkg/pcg/manager_test.go`** (NEW FILE, ~200 lines):
- Test PCG generator registry
- Test content validation integration

- [ ] Create 6 new test files (2,500+ total lines of tests)
- [ ] Use table-driven tests following existing pattern in `pkg/game/character_creation_test.go`
- [ ] Run `go test -race ./...` to verify no race conditions
- [ ] Run `make test-coverage` (uses `scripts/analyze_test_coverage.sh`)
- [ ] Target 85%+ coverage (current 78%)
- [ ] Add coverage requirement to Phase 1 CI/CD (section 1.2)

**Success Criteria**:
- `pkg/game/character_test.go` exists with 20+ test functions
- `pkg/server/handlers_test.go` tests all RPC methods (`handleMove`, `handleAttack`, etc.)
- `pkg/server/health_test.go` tests all health check functions
- Coverage increases from 78% to 85%+
- `go test -race ./...` passes (verifies mutex usage in Character, GameState)
- CI fails if coverage drops below 85%

**Files to Create**:
- `pkg/game/character_test.go` - 20+ tests for Character methods, ~500 lines
- `pkg/game/effects_test.go` - Effect system tests, ~400 lines
- `pkg/server/handlers_test.go` - All RPC handler tests, ~800 lines
- `pkg/server/server_test.go` - Server lifecycle tests, ~300 lines
- `pkg/server/health_test.go` - Health check tests, ~300 lines
- `pkg/pcg/manager_test.go` - PCG tests, ~200 lines

**Code References**:
- Untested files list: Run `scripts/find_untested_files.sh`
- Character methods: `pkg/game/character.go:95` (Clone), line ~400 (SetPosition)
- RPC handlers: `pkg/server/handlers.go:45` (handleMove), line 247 (handleAttack)
- Health checks: `pkg/server/health.go` check functions
- Existing test pattern: `pkg/game/character_creation_test.go` for table-driven style

---

### 2.2 Update Dependencies and Fix Vulnerabilities
**Priority**: High  
**Effort**: 2 days  
**Owner**: Security Team

**Problem**: Dependencies in `go.mod` are outdated (verified with `go list -m -u all`):
- `prometheus/client_golang` v1.22.0 → v1.23.2 (used in `pkg/server/metrics.go:9`)
- `prometheus/common` v0.62.0 → v0.67.1
- `prometheus/procfs` v0.15.1 → v0.19.1
- `github.com/golang/protobuf` v1.5.0 (DEPRECATED, used indirectly)
- Other minor updates: `creack/pty`, `klauspost/compress`, `montanaflynn/stats`

**Current Dependencies** (from `go.mod:7-20`):
```
github.com/google/uuid v1.6.0
github.com/gorilla/websocket v1.5.3
github.com/mb-14/gomarkov v0.0.0-20231120193207-9cbdc8df67a8
github.com/prometheus/client_golang v1.22.0  // NEEDS UPDATE
github.com/sirupsen/logrus v1.9.3
github.com/stretchr/testify v1.10.0
golang.org/x/exp v0.0.0-20250106191152-7588d65b2ba8
golang.org/x/time v0.12.0
gopkg.in/yaml.v3 v3.0.1
```

**Impact**:
- Potential security vulnerabilities in Prometheus client (used for all metrics)
- Missing bug fixes in WebSocket library (used for real-time game updates)
- Deprecated protobuf package may have compatibility issues
- Cannot pass security audits with outdated deps

**Solution**:
- [ ] Run `govulncheck` on current dependencies:
  ```bash
  go install golang.org/x/vuln/cmd/govulncheck@latest
  govulncheck ./...
  ```
- [ ] Update all dependencies:
  ```bash
  go get -u ./...
  go mod tidy
  ```
- [ ] Verify updates don't break Prometheus metrics:
  - Test `pkg/server/metrics.go` NewMetrics() function
  - Verify `/metrics` endpoint still works
  - Check metric registration (lines 53-144 in metrics.go)
- [ ] Verify WebSocket compatibility:
  - Test `pkg/server/websocket.go` broadcast functionality
  - Verify real-time event streaming works
- [ ] Replace deprecated `github.com/golang/protobuf`:
  ```bash
  go get google.golang.org/protobuf@latest
  ```
- [ ] Run full test suite: `go test ./...` (must pass)
- [ ] Run with race detector: `go test -race ./...` (must pass)
- [ ] Update `go.mod` and `go.sum` files
- [ ] Document any breaking changes in commit message

**Success Criteria**:
- `govulncheck ./...` reports zero vulnerabilities
- All dependencies at latest stable versions
- `prometheus/client_golang` >= v1.23.2
- No deprecated packages in `go.mod`
- All tests pass after updates (`go test ./...`)
- No race conditions (`go test -race ./...`)
- Prometheus metrics endpoint `/metrics` works correctly
- WebSocket connections in `pkg/server/websocket.go` function properly

**Files to Modify**:
- `go.mod` - Update all dependency versions (lines 7-20)
- `go.sum` - Regenerated checksums (automatic)

**Verification Commands**:
```bash
go list -m -u all  # Check for updates
govulncheck ./...  # Scan vulnerabilities
go test ./...      # Verify functionality
go test -race ./...  # Check concurrency safety
curl http://localhost:8080/metrics  # Test Prometheus endpoint
```

**Code References**:
- go.mod dependencies: Lines 7-20
- Prometheus usage: `pkg/server/metrics.go:9` (import), lines 53-144 (metric definitions)
- WebSocket usage: `pkg/server/websocket.go` throughout
- Current Go version: `go.mod:3` (1.23.0 with toolchain 1.23.2)
- [ ] Update all dependencies to latest stable versions:
  ```bash
  go get -u ./...
  go mod tidy
  ```
- [ ] Replace deprecated `github.com/golang/protobuf` with `google.golang.org/protobuf`
- [ ] Run full test suite to verify compatibility
- [ ] Run benchmarks to verify no performance regressions
- [ ] Update go.mod and go.sum
- [ ] Document any breaking changes in CHANGELOG.md
- [ ] Add Dependabot configuration for automated updates
- [ ] Set up automated vulnerability scanning in CI

**Success Criteria**:
- All dependencies updated to latest stable versions
- Zero known vulnerabilities reported by govulncheck
- All tests pass after updates
- No performance regressions
- Dependabot configured and monitoring
- Automated vulnerability scanning in CI

**Files to Modify**:
- `go.mod` - Update dependency versions
- `go.sum` - Regenerated checksums
- `.github/dependabot.yml` - Configure auto-updates
- `.github/workflows/security.yml` - Add govulncheck

---

### 2.3 Implement Structured Error Handling with Wrapping
**Priority**: High  
**Effort**: 3 days  
**Owner**: Development Team

**Problem**: Limited use of modern Go error handling patterns. Only 8 instances of `errors.Is` or `errors.As` found in the codebase. Error context is often lost when propagating errors up the call stack.

**Impact**:
- Difficult to debug production issues
- Lost error context makes root cause analysis harder
- Cannot distinguish between error types for retry logic
- Poor error messages for API consumers

**Solution**:
- [ ] Define custom error types for common scenarios:
  ```go
  var (
      ErrNotFound = errors.New("resource not found")
      ErrInvalidInput = errors.New("invalid input")
      ErrUnauthorized = errors.New("unauthorized")
      ErrInternalServer = errors.New("internal server error")
      ErrFileSystemUnavailable = errors.New("file system unavailable")
  )
  ```
- [ ] Create typed error structs for detailed context:
  ```go
  type ValidationError struct {
      Field string
      Value interface{}
      Err   error
  }
  ```
- [ ] Update error handling to use wrapping:
  - Replace `return err` with `return fmt.Errorf("operation failed: %w", err)`
  - Use `errors.Is` for error checking
  - Use `errors.As` for extracting error details
- [ ] Update error responses in RPC handlers to provide better context
- [ ] Add error categorization for monitoring and alerting
- [ ] Update tests to verify error types and wrapping
- [ ] Document error handling patterns in style guide

**Success Criteria**:
- All packages define domain-specific error types
- Errors properly wrapped with context throughout codebase
- RPC handlers return appropriate error codes
- Error logs include full context chain
- Tests verify error types and messages
- Style guide documents error patterns

**Files to Create**:
- `pkg/game/errors.go` - Game-specific errors
- `pkg/server/errors.go` - Server-specific errors
- `docs/ERROR_HANDLING.md` - Error handling guide

**Files to Modify**:
- All files returning errors (systematic refactoring)
- Test files to verify error handling

---

### 2.4 Add End-to-End Integration Tests
**Priority**: High  
**Effort**: 4 days  
**Owner**: QA Team

**Problem**: No end-to-end tests exist. Current tests are unit and component tests. Cannot verify that the full system works together correctly, especially RPC/WebSocket communication.

**Impact**:
- Cannot verify full user workflows
- Integration issues not caught until manual testing
- Risk of breaking changes to API contracts
- Difficult to test deployment configurations

**Solution**:
- [ ] Create E2E test framework in `test/e2e/`:
  - HTTP client for JSON-RPC calls
  - WebSocket client for event streaming
  - Test fixtures for game data
  - Assertions for game state
- [ ] Implement E2E test scenarios:
  - Complete game session flow (join → move → combat → leave)
  - Character creation and progression
  - Spell casting workflow
  - Item equipment and inventory management
  - Quest start and completion
  - PCG content generation
  - Multi-player session interaction
  - WebSocket event broadcasting
- [ ] Add E2E tests for error scenarios:
  - Invalid session handling
  - Network failures and reconnection
  - Concurrent access conflicts
- [ ] Create test data seeding utilities
- [ ] Add E2E tests to CI pipeline
- [ ] Document E2E test writing guidelines

**Success Criteria**:
- E2E tests cover all major user workflows
- Tests run automatically in CI
- Tests verify RPC and WebSocket functionality
- Test data management automated
- E2E test failures provide clear diagnostics
- All tests pass consistently

**Files to Create**:
- `test/e2e/client.go` - E2E test client
- `test/e2e/session_test.go` - Session workflows
- `test/e2e/combat_test.go` - Combat workflows
- `test/e2e/pcg_test.go` - PCG workflows
- `test/e2e/fixtures/` - Test data
- `docs/E2E_TESTING.md` - Testing guide

---

### 2.5 Implement Request Tracing and Correlation IDs
**Priority**: High  
**Effort**: 2 days  
**Owner**: Observability Team

**Problem**: While the codebase has excellent structured logging, there's no request tracing or correlation ID propagation. Cannot trace a request through the entire system for debugging.

**Impact**:
- Difficult to debug issues in production
- Cannot correlate logs across services (future microservices)
- Cannot identify slow request paths
- Limited observability for distributed operations

**Solution**:
- [ ] Add correlation ID middleware for HTTP requests
- [ ] Generate unique correlation ID for each request
- [ ] Propagate correlation ID in context.Context
- [ ] Add correlation ID to all log statements
- [ ] Include correlation ID in error responses
- [ ] Add correlation ID to Prometheus metrics labels
- [ ] Add correlation ID to WebSocket connections
- [ ] Document correlation ID usage for clients
- [ ] Add OpenTelemetry for distributed tracing (optional)

**Success Criteria**:
- Every request has a unique correlation ID
- Correlation ID appears in all related logs
- Correlation ID included in error responses
- Easy to trace request flow through logs
- Metrics can be filtered by correlation ID
- Documentation updated with examples

**Files to Create**:
- `pkg/server/correlation.go` - Correlation ID middleware
- `pkg/server/tracing.go` - Tracing utilities

**Files to Modify**:
- `pkg/server/server.go` - Add middleware
- All handlers to log correlation ID
- Error responses to include correlation ID

---

### 2.6 Enhance Dockerfile for Production
**Priority**: High  
**Effort**: 1 day  
**Owner**: DevOps Team

**Problem**: Current Dockerfile uses `golang:1.22-bookworm` as base (large image), doesn't use multi-stage builds efficiently, and could be optimized for security and size.

**Impact**:
- Large image size (slower deployments)
- Unnecessary tools in production image (attack surface)
- Longer build times
- Higher storage costs

**Solution**:
- [ ] Implement multi-stage build:
  - Stage 1: Build with full golang image
  - Stage 2: Run with distroless or scratch base
- [ ] Optimize layer caching:
  - Copy go.mod/go.sum first
  - Run go mod download
  - Then copy source code
- [ ] Use Go 1.23 to match go.mod
- [ ] Add security scanning (trivy or snyk)
- [ ] Minimize final image size (aim for <50MB)
- [ ] Add metadata labels (version, build date, commit SHA)
- [ ] Document build process and image tags

**Success Criteria**:
- Production image under 50MB
- No unnecessary tools in production image
- Security scanner shows no high/critical issues
- Build time reduced by layer caching
- Image tags include version and commit SHA
- Multi-arch support (amd64, arm64)

**Files to Modify**:
- `Dockerfile` - Refactor with multi-stage build
- `.dockerignore` - Optimize build context

**Files to Create**:
- `Dockerfile.debug` - Debug image with tools
- `docs/DOCKER.md` - Docker usage guide (update existing)

---

### 2.7 Add Performance Benchmarks for Critical Paths
**Priority**: High  
**Effort**: 3 days  
**Owner**: Performance Team

**Problem**: No benchmark tests exist for performance-critical code paths. Cannot measure performance improvements or detect regressions.

**Impact**:
- Cannot establish performance baselines
- Risk of performance regressions
- Difficult to optimize without measurements
- Cannot verify scalability claims

**Solution**:
- [ ] Add benchmark tests for critical operations:
  - Character operations (creation, equipment, effects)
  - Combat calculations (damage, hit chance, effect processing)
  - Spatial index queries (range, radius, nearest)
  - RPC request handling (full request lifecycle)
  - PCG content generation (terrain, items, quests)
  - Event system dispatch (broadcast to N subscribers)
- [ ] Benchmark concurrent operations with different goroutine counts
- [ ] Add memory allocation profiling
- [ ] Create performance regression CI job
- [ ] Document performance targets and SLOs
- [ ] Create performance dashboard (Grafana)

**Success Criteria**:
- Benchmarks for all critical paths
- Performance baselines documented
- CI fails on significant regressions (>10%)
- Memory allocation tracked
- Performance targets met:
  - RPC request: <10ms p95
  - Spatial query: <1ms p95
  - Combat round: <50ms p95
  - PCG generation: <500ms per level

**Files to Create**:
- `pkg/game/*_bench_test.go` - Game benchmarks
- `pkg/server/*_bench_test.go` - Server benchmarks
- `pkg/pcg/*_bench_test.go` - PCG benchmarks
- `docs/PERFORMANCE.md` - Performance guide

---

### 2.8 Implement Session Persistence
**Priority**: High  
**Effort**: 2 days  
**Owner**: Backend Team

**Problem**: Sessions are stored in memory only. While Phase 1 adds file-based persistence for game state, sessions also need durability across server restarts.

**Impact**:
- Sessions lost on server restart
- Cannot maintain session continuity during deployments
- No session-based analytics or monitoring

**Solution**:
- [ ] Extend file-based persistence to sessions
- [ ] Create session store interface:
  ```go
  type SessionStore interface {
      Set(ctx context.Context, id string, session *PlayerSession, ttl time.Duration) error
      Get(ctx context.Context, id string) (*PlayerSession, error)
      Delete(ctx context.Context, id string) error
      Exists(ctx context.Context, id string) (bool, error)
  }
  ```
- [ ] Implement file-backed session store (`data/sessions/{id}.yaml`)
- [ ] Implement memory-backed store for development/testing
- [ ] Add session serialization (YAML or JSON)
- [ ] Use file locking for concurrent session access
- [ ] Implement session cleanup/expiration background task
- [ ] Update session management in RPCServer
- [ ] Document session management approach

**Success Criteria**:
- Sessions persist across server restarts
- Session TTL handled automatically with cleanup
- Fallback to memory store for development
- File locking prevents corruption
- Performance tested with 1k+ concurrent sessions

**Files to Create**:
- `pkg/persistence/session_store.go` - Session store interface & file implementation
- `pkg/persistence/memory_store.go` - Memory implementation

**Files to Modify**:
- `pkg/server/server.go` - Use session store
- `pkg/config/config.go` - Add session persistence config

**Note**: For horizontal scaling with multiple server instances, consider BoltDB with shared file system or external solution. Current file-based approach works for single-instance deployments.

---

## Phase 3: Medium Priority Improvements
**Timeline**: Week 5-6  
**Goal**: Enhance developer experience and operational maturity

### 3.1 Add OpenAPI/Swagger Documentation for RPC API
**Priority**: Medium  
**Effort**: 3 days  
**Owner**: Documentation Team

**Problem**: While good RPC documentation exists in `pkg/README-RPC.md`, there's no machine-readable API specification (OpenAPI/Swagger). Difficult for clients to auto-generate SDKs.

**Solution**:
- [ ] Create OpenAPI 3.0 specification for JSON-RPC API
- [ ] Document all RPC methods, parameters, and responses
- [ ] Add example requests and responses
- [ ] Host Swagger UI at `/api/docs`
- [ ] Generate TypeScript client from OpenAPI spec
- [ ] Add schema validation using OpenAPI spec
- [ ] Keep OpenAPI spec in sync with code (use go-swagger or similar)

**Success Criteria**:
- Complete OpenAPI 3.0 specification
- Swagger UI accessible and functional
- Client SDK auto-generated from spec
- API changes automatically update spec

**Files to Create**:
- `api/openapi.yaml` - OpenAPI specification
- `pkg/server/swagger.go` - Swagger UI handler

---

### 3.2 Implement Graceful Degradation for Dependencies
**Priority**: Medium  
**Effort**: 2 days  
**Owner**: Resilience Team

**Problem**: While circuit breakers exist for some operations, not all external dependencies have degradation strategies. Server may become unavailable if non-critical services fail.

**Solution**:
- [ ] Identify non-critical dependencies (metrics, PCG, validation)
- [ ] Implement graceful degradation strategies:
  - Continue serving without metrics if Prometheus fails
  - Use cached PCG content if generation fails
  - Log validation failures but allow requests
- [ ] Add fallback behaviors for each subsystem
- [ ] Add health status levels (healthy, degraded, unhealthy)
- [ ] Update health endpoints to report degradation
- [ ] Document degradation modes

**Success Criteria**:
- Server stays available during partial failures
- Health endpoints report degradation status
- Degraded mode documented
- Alert when running in degraded mode

**Files to Modify**:
- `pkg/server/health.go` - Add degradation status
- Circuit breaker implementations

---

### 3.3 Add Data Structure Versioning
**Priority**: Medium  
**Effort**: 2 days  
**Owner**: Backend Team

**Problem**: No versioning system for persisted data structures (prerequisite from Phase 1.1). Need systematic way to evolve file formats and handle backward compatibility.

**Solution**:
- [ ] Add version field to all persisted data structures
- [ ] Implement data format migration utilities
- [ ] Create migration functions for schema changes
- [ ] Add version compatibility checking on load
- [ ] Document migration procedures
- [ ] Implement automatic backups before migrations
- [ ] Add migration CLI tool for offline data migration

**Success Criteria**:
- All data files include version information
- Migrations run automatically on version mismatch
- Backward compatibility maintained
- Migration rollback possible via backups
- Migration CLI tool functional

**Files to Create**:
- `pkg/persistence/versioning.go` - Version management
- `pkg/persistence/migrations.go` - Migration utilities
- `cmd/migrate-data/main.go` - Migration CLI

**Note**: This replaces traditional database migrations with file format evolution.

---

### 3.4 Implement Rate Limiting per User/Session
**Priority**: Medium  
**Effort**: 2 days  
**Owner**: Security Team

**Problem**: Rate limiting exists per IP (`pkg/server/ratelimit.go`), but not per user session. Authenticated users behind same IP share rate limits.

**Solution**:
- [ ] Add session-based rate limiting
- [ ] Add per-user rate limit tiers (free, premium)
- [ ] Implement token bucket algorithm per session
- [ ] Add rate limit headers to responses (X-RateLimit-*)
- [ ] Return 429 Too Many Requests with retry-after
- [ ] Monitor rate limit violations
- [ ] Document rate limits in API docs

**Success Criteria**:
- Sessions rate limited independently
- Rate limits configurable per tier
- Clear rate limit feedback to clients
- Abuse patterns detected and alerted

**Files to Create**:
- `pkg/server/ratelimit_session.go` - Session rate limiter

**Files to Modify**:
- `pkg/server/server.go` - Add session rate limiting

---

### 3.5 Add Structured Logging with Log Levels
**Priority**: Medium  
**Effort**: 2 days  
**Owner**: Observability Team

**Problem**: Logging is well-structured with logrus, but could benefit from more consistent log level usage and context propagation. Some areas over-log or under-log.

**Solution**:
- [ ] Audit all logging calls for appropriate levels:
  - DEBUG: Detailed diagnostics
  - INFO: Normal operations
  - WARN: Recoverable issues
  - ERROR: Failures requiring attention
- [ ] Add log sampling for high-frequency events
- [ ] Implement context-aware logging (with correlation ID)
- [ ] Add log aggregation configuration (for ELK/Loki)
- [ ] Document logging standards
- [ ] Add log volume monitoring

**Success Criteria**:
- Consistent log levels across codebase
- High-frequency logs sampled appropriately
- Log volume within acceptable limits
- Logging standards documented

**Files to Modify**:
- All files with logging (systematic review)

---

### 3.6 Add Configuration Validation at Startup
**Priority**: Medium  
**Effort**: 1 day  
**Owner**: Configuration Team

**Problem**: Configuration validation exists (`pkg/config/config.go`) but could be more comprehensive. Some invalid configurations only discovered at runtime.

**Solution**:
- [ ] Enhance configuration validation:
  - Verify required files exist (web directory, data files)
  - Verify data directory is writable
  - Validate port availability
  - Check resource limits are reasonable
  - Verify secrets backend connectivity
  - Test file system access and permissions
- [ ] Add --validate flag to check config without starting server
- [ ] Fail fast on invalid configuration
- [ ] Provide actionable error messages
- [ ] Document all configuration options

**Success Criteria**:
- All configuration errors caught at startup
- Validation CLI command works
- Clear error messages for invalid config
- Configuration fully documented

**Files to Modify**:
- `pkg/config/config.go` - Enhanced validation
- `cmd/server/main.go` - Add validation mode

---

### 3.7 Implement Audit Logging for Security Events
**Priority**: Medium  
**Effort**: 2 days  
**Owner**: Security Team

**Problem**: No dedicated audit logging for security-relevant events (authentication, authorization, data access, configuration changes).

**Solution**:
- [ ] Create audit log system separate from application logs
- [ ] Log security events:
  - Session creation/destruction
  - Failed authentication attempts
  - Permission denied errors
  - Configuration changes
  - Data modifications
- [ ] Use immutable log storage
- [ ] Add audit log retention policy
- [ ] Implement audit log querying API
- [ ] Document compliance requirements (GDPR, SOC2)

**Success Criteria**:
- All security events audited
- Audit logs tamper-proof
- Audit logs queryable
- Retention policy enforced
- Compliance documented

**Files to Create**:
- `pkg/audit/logger.go` - Audit logging
- `pkg/audit/events.go` - Audit event types

---

### 3.8 Add Metrics Dashboard Templates
**Priority**: Medium  
**Effort**: 2 days  
**Owner**: Observability Team

**Problem**: Prometheus metrics exist but no pre-built dashboards for operators. Must manually create visualizations.

**Solution**:
- [ ] Create Grafana dashboard JSON templates:
  - System overview (requests, errors, latency)
  - Game metrics (active sessions, player actions)
  - Performance (memory, CPU, goroutines)
  - Error rates and types
  - PCG generation metrics
- [ ] Create alerting rules for Prometheus
- [ ] Document dashboard installation
- [ ] Add dashboard screenshots to docs

**Success Criteria**:
- Grafana dashboards importable
- All key metrics visualized
- Alert rules configured
- Dashboard docs with screenshots

**Files to Create**:
- `deploy/grafana/dashboards/*.json` - Dashboard templates
- `deploy/prometheus/alerts.yml` - Alert rules
- `docs/MONITORING.md` - Monitoring guide

---

### 3.9 Improve Error Messages for Common User Mistakes
**Priority**: Medium  
**Effort**: 2 days  
**Owner**: UX Team

**Problem**: Error messages are technically accurate but may not be user-friendly or provide guidance for resolution.

**Solution**:
- [ ] Audit all user-facing error messages
- [ ] Improve error message clarity:
  - Explain what went wrong
  - Suggest how to fix it
  - Include relevant context
  - Use consistent formatting
- [ ] Add error codes for programmatic handling
- [ ] Create error message catalog
- [ ] Add helpful links to documentation
- [ ] Support error message localization

**Success Criteria**:
- All errors have user-friendly messages
- Error codes assigned consistently
- Documentation links in errors
- User testing validates improvements

**Files to Modify**:
- `pkg/server/errors.go` - Error messages
- `pkg/validation/*.go` - Validation errors

---

### 3.10 Add Smoke Tests for Deployment Verification
**Priority**: Medium  
**Effort**: 2 days  
**Owner**: QA Team

**Problem**: No automated smoke tests to verify deployment success. Must manually test each deployment.

**Solution**:
- [ ] Create smoke test suite:
  - Server responds to health checks
  - File system persistence works
  - RPC methods respond correctly
  - WebSocket connections work
  - PCG generation functions
  - Metrics endpoint accessible
- [ ] Run smoke tests after deployment in CD pipeline
- [ ] Alert on smoke test failures
- [ ] Document smoke test requirements

**Success Criteria**:
- Smoke tests cover critical paths
- Tests run automatically post-deployment
- Deployment failures detected quickly
- Rollback triggered on failures

**Files to Create**:
- `test/smoke/smoke_test.go` - Smoke tests
- `scripts/run_smoke_tests.sh` - Test runner

---

### 3.11 Document Troubleshooting Procedures
**Priority**: Medium  
**Effort**: 2 days  
**Owner**: Operations Team

**Problem**: Limited operational documentation for troubleshooting common issues in production.

**Solution**:
- [ ] Create troubleshooting guide:
  - Common error scenarios
  - Log analysis procedures
  - Performance debugging
  - File system issues
  - Memory leaks
  - Goroutine leaks
  - Network problems
- [ ] Document escalation procedures
- [ ] Create runbooks for common tasks
- [ ] Add decision trees for issue diagnosis

**Success Criteria**:
- Troubleshooting guide comprehensive
- Common issues documented with solutions
- Runbooks tested and verified
- Team trained on procedures

**Files to Create**:
- `docs/TROUBLESHOOTING.md` - Troubleshooting guide
- `docs/RUNBOOKS.md` - Operational runbooks

---

### 3.12 Add Request/Response Logging Middleware
**Priority**: Medium  
**Effort**: 1 day  
**Owner**: Observability Team

**Problem**: No comprehensive request/response logging for debugging. Must add logging manually to investigate issues.

**Solution**:
- [ ] Add request/response logging middleware
- [ ] Log request details (method, params, size)
- [ ] Log response details (status, size, duration)
- [ ] Add configurable sampling rate
- [ ] Sanitize sensitive data from logs
- [ ] Add request replay capability for debugging

**Success Criteria**:
- All requests logged with details
- Sensitive data properly sanitized
- Log volume manageable with sampling
- Request replay possible from logs

**Files to Create**:
- `pkg/server/request_logger.go` - Logging middleware

---

## Phase 4: Low Priority Enhancements
**Timeline**: Week 7-8  
**Goal**: Polish and nice-to-have improvements

### 4.1 Add API Versioning Support
**Priority**: Low  
**Effort**: 2 days  
**Owner**: API Team

**Problem**: No API versioning strategy. Breaking changes will affect all clients simultaneously.

**Solution**:
- [ ] Add version to RPC method names (`v1.move`, `v2.move`)
- [ ] Implement version negotiation in handlers
- [ ] Support multiple versions concurrently
- [ ] Document deprecation policy
- [ ] Add version to metrics labels

**Success Criteria**:
- Multiple API versions supported
- Version negotiation works
- Deprecation policy documented
- No breaking changes without version bump

---

### 4.2 Implement Caching Layer for Read-Heavy Operations
**Priority**: Low  
**Effort**: 3 days  
**Owner**: Performance Team

**Problem**: Some operations query static data repeatedly (spells, items, terrain). Could benefit from caching.

**Solution**:
- [ ] Add caching layer with TTL
- [ ] Cache spell definitions
- [ ] Cache item templates
- [ ] Cache PCG templates
- [ ] Add cache hit/miss metrics
- [ ] Implement cache invalidation strategy

**Success Criteria**:
- Cache hit rate >80% for static data
- Cache properly invalidated on updates
- Performance improved for cached operations
- Cache metrics monitored

---

### 4.3 Add GraphQL API Alternative to JSON-RPC
**Priority**: Low  
**Effort**: 5 days  
**Owner**: API Team

**Problem**: JSON-RPC is functional but GraphQL might provide better flexibility for complex queries.

**Solution**:
- [ ] Add GraphQL server (gqlgen)
- [ ] Define GraphQL schema
- [ ] Implement resolvers for game data
- [ ] Add GraphQL playground
- [ ] Document GraphQL API
- [ ] Compare performance with JSON-RPC

**Success Criteria**:
- GraphQL API functional
- Complex queries work efficiently
- Documentation complete
- Client examples provided

---

### 4.4 Implement Server-Sent Events (SSE) Alternative to WebSocket
**Priority**: Low  
**Effort**: 2 days  
**Owner**: API Team

**Problem**: WebSocket requires bidirectional communication, but some clients only need server-to-client events. SSE is simpler for this use case.

**Solution**:
- [ ] Add SSE endpoint for event streaming
- [ ] Support same events as WebSocket
- [ ] Add SSE reconnection logic
- [ ] Document SSE usage
- [ ] Compare with WebSocket performance

**Success Criteria**:
- SSE endpoint functional
- Events delivered reliably
- Reconnection works
- Documentation complete

---

### 4.5 Add Admin API for Operational Tasks
**Priority**: Low  
**Effort**: 3 days  
**Owner**: Operations Team

**Problem**: No admin API for operational tasks. Must manually edit files or restart server.

**Solution**:
- [ ] Create admin API endpoints:
  - Session management (list, terminate)
  - User management (ban, unban)
  - Configuration reload
  - Cache clearing
  - Metrics reset
- [ ] Add authentication for admin API
- [ ] Document admin API
- [ ] Add admin CLI tool

**Success Criteria**:
- Admin API functional
- Authentication required
- Common operations automated
- CLI tool usable

---

### 4.6 Improve TypeScript Frontend Build Process
**Priority**: Low  
**Effort**: 2 days  
**Owner**: Frontend Team

**Problem**: TypeScript build is basic. Could benefit from optimization and better tooling.

**Solution**:
- [ ] Add webpack or vite for better bundling
- [ ] Add code splitting for faster loads
- [ ] Add source maps for debugging
- [ ] Add tree shaking for smaller bundles
- [ ] Add hot module replacement for development

**Success Criteria**:
- Build time reduced
- Bundle size reduced
- Development experience improved
- Source maps working

---

### 4.7 Add Load Testing Suite
**Priority**: Low  
**Effort**: 3 days  
**Owner**: Performance Team

**Problem**: No load testing infrastructure. Cannot verify performance under realistic load.

**Solution**:
- [ ] Create load testing scenarios (k6 or locust)
- [ ] Simulate realistic user behavior
- [ ] Test various load levels (100, 1k, 10k users)
- [ ] Identify performance bottlenecks
- [ ] Document performance limits
- [ ] Add load tests to CI (optional, expensive)

**Success Criteria**:
- Load tests cover major scenarios
- Performance baselines established
- Bottlenecks identified
- Capacity planning data available

---

### 4.8 Add Chaos Engineering Tests
**Priority**: Low  
**Effort**: 3 days  
**Owner**: Resilience Team

**Problem**: Resilience patterns exist but not tested under realistic failure conditions.

**Solution**:
- [ ] Add chaos engineering framework (chaos-mesh or similar)
- [ ] Test failure scenarios:
  - Network latency/partition
  - Pod restarts
  - File system failures
  - Memory pressure
  - CPU throttling
- [ ] Verify graceful degradation
- [ ] Document failure behavior

**Success Criteria**:
- Chaos tests automated
- System resilient to common failures
- Degradation behavior verified
- Recovery time measured

---

### 4.9 Add Multi-Language Support (i18n)
**Priority**: Low  
**Effort**: 4 days  
**Owner**: Localization Team

**Problem**: All messages in English only. Cannot serve international audience effectively.

**Solution**:
- [ ] Add i18n library (go-i18n)
- [ ] Extract all user-facing strings
- [ ] Create translation files
- [ ] Add language negotiation
- [ ] Support language parameter in API
- [ ] Test with multiple languages

**Success Criteria**:
- i18n framework integrated
- English and one other language supported
- Translation workflow documented
- Language switching works

---

## Detailed Findings

### Architecture & Code Quality

**Strengths**:
- Clean package structure following Go conventions (`pkg/`, `cmd/`, `internal/` pattern)
- Clear separation of concerns (game logic, server, validation, resilience)
- Thread-safe concurrent operations with proper mutex usage
- Good use of interfaces for abstraction (SecretProvider, SessionStore patterns)
- Strong resilience patterns (circuit breakers, retry mechanisms)

**Issues**:
- Limited use of `errors.Is` and `errors.As` for error type checking (only 8 instances)
- Some packages lack clear boundaries (server package is large at 4000+ lines)
- No architectural decision records (ADRs) documenting key design choices
- 3 TODO/FIXME comments in codebase (minimal technical debt)

**Recommendations**:
1. Adopt systematic error wrapping with `fmt.Errorf("context: %w", err)`
2. Consider splitting large packages (server) into sub-packages
3. Document architectural decisions in `docs/architecture/` directory
4. Establish code review checklist for consistency

---

### Testing & Quality Assurance

**Current Coverage**: 78% (73 files with tests, 20 without)

**Strengths**:
- 106 test files with comprehensive table-driven tests
- Race detector runs clean (no race conditions detected)
- Good use of testify for assertions
- Test helper functions for common setup

**Gaps**:
- No test coverage for 20 critical files (character.go, handlers.go, server.go)
- No end-to-end integration tests
- No performance benchmark tests
- No load testing or stress testing
- No mutation testing to verify test quality
- Test coverage not enforced in CI (no CI exists)

**Recommendations**:
1. Prioritize tests for uncovered files (see Phase 2.1)
2. Implement E2E test framework (see Phase 2.4)
3. Add benchmark tests for critical paths (see Phase 2.7)
4. Set up CI with coverage enforcement (see Phase 1.2)
5. Consider property-based testing for complex logic (gopter or similar)

---

### Error Handling & Resilience

**Strengths**:
- Comprehensive circuit breaker implementation (`pkg/resilience/`)
- Retry mechanisms with exponential backoff (`pkg/retry/`)
- Rate limiting per IP address
- Graceful shutdown implemented in main.go
- Custom error types defined (ErrInvalidSession)

**Issues**:
- Limited error wrapping (context often lost)
- No error categorization for metrics/alerting
- Some errors lack user-friendly messages
- No systematic error code system for API
- Context cancellation not always checked in long-running operations

**Recommendations**:
1. Implement systematic error wrapping (see Phase 2.3)
2. Add error codes for API responses
3. Create error catalog with solutions
4. Add context deadline monitoring
5. Document error handling patterns

---

### Observability & Monitoring

**Strengths**:
- Excellent Prometheus metrics coverage (`pkg/server/metrics.go`)
- Comprehensive health check system (`pkg/server/health.go`)
- Structured logging with logrus throughout
- Multiple health endpoints (/health, /ready, /live)
- Performance monitoring system (`pkg/server/performance.go`)
- WebSocket event broadcasting for real-time updates

**Issues**:
- No request tracing or correlation IDs
- No pre-built Grafana dashboards
- No alerting rules defined
- Log volume not monitored or limited
- No distributed tracing (OpenTelemetry)
- No log aggregation configuration

**Recommendations**:
1. Add correlation ID propagation (see Phase 2.5)
2. Create Grafana dashboard templates (see Phase 3.8)
3. Define Prometheus alerting rules
4. Add log sampling for high-frequency events
5. Consider OpenTelemetry for future distributed tracing

---

### Security

**Strengths**:
- Comprehensive input validation framework (`pkg/validation/`)
- Rate limiting implemented
- Security headers in HTTP responses
- No hardcoded secrets found in codebase
- WebSocket origin validation (currently permissive for dev)
- User runs as non-root in Docker

**Weaknesses**:
- No secrets management system (relies on environment variables)
- No authentication/authorization system
- No audit logging for security events
- No security scanning in CI/CD
- WebSocket allows all origins in development mode
- No HTTPS/TLS termination configured
- No protection against CSRF attacks
- Dependencies have known vulnerabilities (not checked automatically)

**Recommendations**:
1. Implement secrets management (see Phase 1.3)
2. Add authentication/authorization middleware
3. Implement audit logging (see Phase 3.7)
4. Add security scanning to CI (govulncheck, gosec)
5. Configure proper CORS policies for production
6. Add HTTPS support with TLS configuration
7. Implement CSRF protection for state-changing operations
8. Set up automated dependency vulnerability scanning

**Security Checklist**:
- [ ] Secrets management implemented
- [ ] Authentication required for admin operations
- [ ] Authorization checks on all endpoints
- [ ] Audit logging for security events
- [ ] Security scanning in CI
- [ ] HTTPS enabled in production
- [ ] CORS properly configured
- [ ] Rate limiting per user
- [ ] Input validation on all inputs
- [ ] SQL injection protection (when DB added)
- [ ] XSS protection in responses
- [ ] CSRF protection enabled

---

### Documentation

**Strengths**:
- Excellent README.md with comprehensive feature list
- Detailed RPC API documentation (`pkg/README-RPC.md`)
- Package documentation in doc.md files
- Good inline code comments
- Multiple specialized guides (ASSET_ANALYSIS.md, AUDIT.md, etc.)

**Gaps**:
- No deployment guide
- No operations/runbook documentation
- No troubleshooting guide
- No architecture documentation
- No API versioning policy
- No contribution guidelines
- No security policy (SECURITY.md)
- No changelog (CHANGELOG.md)

**Recommendations**:
1. Create deployment guide (see Phase 1.4)
2. Document troubleshooting procedures (see Phase 3.11)
3. Add architecture decision records
4. Create CONTRIBUTING.md
5. Add SECURITY.md with vulnerability reporting
6. Maintain CHANGELOG.md for releases
7. Add API versioning documentation

---

### Deployment & Operations

**Strengths**:
- Basic Dockerfile exists
- Graceful shutdown implemented
- Health checks for readiness probes
- Configuration via environment variables
- Makefile for common tasks

**Critical Gaps**:
- No CI/CD pipeline (`.github/workflows/` empty)
- No Kubernetes manifests
- No Docker Compose for local development
- No deployment automation
- No rollback procedures
- No monitoring dashboards
- No alerting configuration
- No backup/restore procedures
- No disaster recovery plan

**Operational Concerns**:
- Data persistence missing (Phase 1.1 prerequisite)
- Session persistence missing (Phase 2.8)
- No data format versioning system
- No zero-downtime deployment strategy
- No capacity planning data
- No performance SLOs defined
- No incident response procedures

**Recommendations**:
1. Implement data persistence (see Phase 1.1)
2. Create CI/CD pipeline (see Phase 1.2)
3. Build deployment manifests (see Phase 1.4)
4. Document operations procedures (see Phase 3.11)
5. Set up monitoring and alerting
6. Define SLOs and SLIs
7. Create disaster recovery plan
8. Document capacity planning

---

## Dependencies & Blockers

### Phase 1 Dependencies
- None (all can start immediately)

### Phase 2 Dependencies
- **2.1 (Test Coverage)**: Requires working build from Phase 1.2 CI setup
- **2.8 (Session Persistence)**: Requires persistence layer from Phase 1.1

### Phase 3 Dependencies
- **3.3 (Data Versioning)**: Requires persistence layer from Phase 1.1
- **3.8 (Metrics Dashboards)**: Benefits from request tracing from Phase 2.5

### Phase 4 Dependencies
- **4.5 (Admin API)**: Requires authentication system (not in roadmap)

### External Dependencies
- Secrets backend selection (HashiCorp Vault or AWS Secrets Manager)
- Container registry (GitHub Container Registry recommended)
- Kubernetes cluster for deployment
- Monitoring infrastructure (Prometheus + Grafana)

---

## Recommended Tools & Libraries

### Infrastructure
- **BoltDB**: Embedded key-value database (if flat files insufficient) - `go.etcd.io/bbolt`
- **Docker**: Container runtime
- **Kubernetes**: Container orchestration
- **Helm**: Package management for K8s

### Security & Secrets
- **HashiCorp Vault**: Secrets management (or AWS Secrets Manager)
- **govulncheck**: Vulnerability scanning
- **gosec**: Security code analysis
- **trivy**: Container image scanning

### Monitoring & Observability
- **Prometheus**: Metrics collection (already integrated)
- **Grafana**: Metrics visualization
- **Loki**: Log aggregation (optional)
- **Jaeger**: Distributed tracing (optional)

### Testing
- **testify**: Assertion library (already used)
- **gomock**: Mock generation (optional)
- **k6**: Load testing
- **testcontainers-go**: Integration testing with real dependencies

### Development
- **golangci-lint**: Comprehensive linting
- **gofumpt**: Strict formatting (already used)
- **air**: Hot reload for development

### CI/CD
- **GitHub Actions**: CI/CD platform
- **Dependabot**: Automated dependency updates
- **Codecov**: Test coverage tracking

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Data loss during file operations | Medium | High | Implement atomic writes, thorough backup/restore testing, file locking, test rollback procedures |
| Performance with file-based persistence | Medium | Medium | Benchmark file I/O operations, implement caching, use efficient serialization, consider BoltDB if needed |
| Breaking API changes affect clients | Low | High | Implement API versioning early, maintain backwards compatibility, document deprecations |
| Security vulnerability discovered | Medium | High | Implement security scanning in CI, regular dependency updates, security audit before launch |
| Scaling issues under load | Medium | Medium | Load test before launch, monitor performance metrics, consider horizontal scaling approach |
| Data format migration failures | Medium | High | Test migrations in staging, implement automatic backups, maintain format version history |
| Secrets management misconfiguration | Low | Critical | Use infrastructure-as-code, test secret rotation, implement least-privilege access |
| CI/CD pipeline failures block releases | Low | Medium | Implement pipeline testing, maintain fast feedback loops, allow emergency deploys |
| Incomplete test coverage hides bugs | High | Medium | Enforce coverage thresholds, add E2E tests, perform exploratory testing |
| Documentation drift from implementation | Medium | Low | Keep docs in code repo, review in PRs, automate API doc generation |

---

## Next Steps

### Immediate Actions (Week 1)
1. **Review and approve this roadmap** with stakeholders
2. **Assign owners** to Phase 1 tasks
3. **Set up project tracking** in GitHub Projects or Jira
4. **Begin Phase 1.1** (Data Persistence) and **Phase 1.2** (CI/CD) in parallel
5. **Schedule weekly roadmap reviews** to track progress

### Success Metrics
- **Week 2**: Phase 1 tasks 50% complete
- **Week 4**: Phase 1 complete, Phase 2 started
- **Week 6**: Phase 2 complete, Phase 3 started
- **Week 8**: Phase 3 complete, production deployment ready
- **Ongoing**: Phase 4 nice-to-have improvements

### Decision Points
- **Week 2**: Database technology confirmed and approved
- **Week 2**: Secrets management solution selected
- **Week 4**: CI/CD pipeline validated and automated
- **Week 6**: Staging environment deployed and tested
- **Week 8**: Production readiness review and go/no-go decision

---

## Maintenance Plan (Post-Launch)

### Daily Operations
- Monitor health check endpoints
- Review error rates and alert notifications
- Check Grafana dashboards for anomalies
- Verify backup completion
- Review audit logs for security events

### Weekly Operations
- Review and triage new issues
- Update dependencies (automated via Dependabot)
- Analyze performance trends
- Review capacity planning metrics
- Update documentation as needed

### Monthly Operations
- Security audit and vulnerability scan
- Performance optimization review
- Disaster recovery drill
- Capacity planning review
- Documentation completeness review

### Quarterly Operations
- Major version planning
- Architecture review
- Tech debt assessment and planning
- Load testing with traffic projections
- Team training on new features

### Ongoing Improvements
- Respond to user feedback
- Implement high-value feature requests
- Optimize performance bottlenecks
- Enhance monitoring and observability
- Update documentation and guides

---

## Conclusion

The GoldBox RPG Engine demonstrates a strong foundation with excellent observability, resilience patterns, and code quality. The roadmap addresses critical gaps in deployment, persistence, and CI/CD while maintaining the existing strengths.

**Key Takeaways**:
1. **Solid Foundation**: 68/100 production readiness score is strong for a greenfield project
2. **Clear Path Forward**: Phased approach provides actionable steps with clear priorities
3. **Reasonable Timeline**: 6-8 weeks to production readiness is achievable
4. **Low Risk**: Most issues are infrastructure and process, not code quality
5. **Strong Team**: Existing documentation and code quality suggest capable team

**Production Readiness Timeline**:
- **Week 2**: Critical blockers resolved (persistence, CI/CD, secrets)
- **Week 4**: High-priority improvements complete (tests, dependencies)
- **Week 6**: Medium-priority enhancements done (operations, monitoring)
- **Week 8**: Production deployment ready with full operations support

**Success Factors**:
1. Executive sponsorship and resource allocation
2. Clear ownership and accountability
3. Regular progress reviews and course correction
4. Quality over speed (no shortcuts on security)
5. Team training and knowledge sharing

This roadmap provides a comprehensive path to production readiness while maintaining the high quality standards evident in the existing codebase.
