# Production Readiness Quick Reference

**Status:** ✅ CONDITIONALLY READY (5/7 gates)  
**Launch Blocker:** Phase 1 tasks (6.5 hours)  
**Generated:** 2026-03-11

---

## TL;DR

**Can we ship?** YES, after fixing 5 documentation bugs and verifying goroutine cleanup (1 day effort).

**Why not 7/7?** Function length (278 violations) and minor complexity (6 violations) are maintainability issues, not runtime risks. With 78% test coverage + race detector in CI, we're operationally sound.

**What's next?** Split `handlers.go` within 30 days of launch to prevent tech debt accumulation.

---

## Gate Scores (5/7 Passing)

| Gate | Status | Details |
|------|--------|---------|
| Complexity | ⚠️ 6 violations | Highest: `ModifyReputation` (cyclomatic 14, 116 lines) |
| **Function Length** | ❌ 278 >30 lines | Worst: `run` (122 lines, demo code) |
| Documentation | ✅ 83.0% | Exceeds 80% threshold |
| Duplication | ✅ 1.25% | Far below 5% threshold |
| Circular Deps | ✅ 0 detected | Clean architecture |
| Naming | ✅ 99.4% | 38 minor violations (all low severity) |
| Concurrency | ✅ 2 medium-risk | No high-risk patterns; race detector passes |

---

## Pre-Launch Checklist (Phase 1)

### 1. Fix Documentation Bugs (90 min)
```bash
# Edit these files to resolve BUG annotations:
pkg/server/session.go:160    # Clarify cleanup logging
pkg/server/server.go:489     # Document profiling behavior
pkg/pcg/doc.go:52            # Complete reproduction guide
cmd/server/doc.go:46,48      # Fix logging config examples
```

### 2. Verify Concurrency (2 hours)
```bash
# Run race detector
go test ./... -race -timeout 10m

# Audit these goroutines for graceful shutdown:
# - server:232 (infinite loop - verify context.Done() handling)
# - server:290 (pipeline - confirm channel closure)

# Add goroutine leak detection (optional but recommended)
go get -u go.uber.org/goleak
```

### 3. Production Config (3 hours)
```bash
# Set WebSocket origin validation
export WEBSOCKET_ALLOWED_ORIGINS="https://yourdomain.com"

# Test health checks under load
ab -n 10000 -c 100 http://localhost:8080/health
ab -n 10000 -c 100 http://localhost:8080/ready

# Verify metrics endpoint
curl http://localhost:8080/metrics | grep goldbox

# Test graceful shutdown (SIGTERM)
docker run -d goldbox-rpg:latest
docker stop goldbox-rpg  # Should exit cleanly in <30s
```

**Total Phase 1 Effort:** 6.5 hours → **Ready for production** ✅

---

## Post-Launch Priorities (Weeks 2-3)

### Priority 1: Split handlers.go (8 hours)
```bash
# Current: pkg/server/handlers.go (2,431 lines, 94 functions)
# Target: 5 domain-specific files

pkg/server/
├── handlers_combat.go      # Attack, Move, CastSpell (15 methods)
├── handlers_character.go   # CreateCharacter, EquipItem (20 methods)
├── handlers_world.go       # GetWorld, QueryObjects (12 methods)
├── handlers_spell.go       # SearchSpells, GetSpells (8 methods)
└── handlers_misc.go        # Remaining utilities (39 methods)

# Reduces burden score: 4.32 → <1.5
```

### Priority 2: Refactor Top Complexity (7 hours)
```bash
# 1. ModifyReputation (pkg/pcg/reputation.go)
#    - Current: 116 lines, cyclomatic 14
#    - Extract: calculateReputationDelta(), updateRelationshipStatus()

# 2. Roll function (pkg/game/dice.go)
#    - Current: 86 lines
#    - Extract: applyDiceModifiers(), validateDiceExpression()

# 3. refreshGameState (pkg/wasmui/game.go)
#    - Current: 51 lines, cyclomatic 14
#    - Extract state update logic into smaller functions
```

---

## Metrics to Monitor

### Runtime Metrics (Prometheus `/metrics`)
```bash
# Goroutine leak detection
go_goroutines{job="goldbox-rpg"}  # Should stabilize after startup

# Memory usage
go_memstats_heap_inuse_bytes

# Request latency
goldbox_http_request_duration_seconds{endpoint="/health"}
```

### Code Quality Metrics (Monthly)
```bash
# Re-run analysis
go-stats-generator analyze . --skip-tests --format json --output metrics.json

# Track trends:
# - Duplication ratio (target: <1.0%)
# - Complexity violations (target: ≤3)
# - Function length violations (target: <200)
```

---

## Top 10 Files to Watch

| File | Lines | Burden | Action |
|------|-------|--------|--------|
| pkg/server/handlers.go | 2,431 | 4.32 | Split in Phase 2 |
| pkg/pcg/types.go | 485 | 3.44 | OK (type definitions) |
| pkg/game/character.go | 1,009 | 2.45 | OK (core domain) |
| pkg/pcg/metrics.go | 683 | 2.37 | OK (metrics tracking) |
| pkg/pcg/world.go | 630 | 2.27 | Consider splitting |
| pkg/pcg/narrative.go | 573 | 2.02 | OK (coherent module) |
| pkg/pcg/validator.go | 739 | 1.82 | OK (validation logic) |
| pkg/pcg/balancer.go | 679 | 1.80 | OK (balancing system) |
| pkg/pcg/levels/rooms.go | 508 | 1.56 | OK (room generation) |
| pkg/validation/validation.go | 671 | 1.52 | Extract helpers (Phase 3) |

---

## Common Questions

**Q: Why allow functions >30 lines if it fails the gate?**  
A: Context matters. Game logic (dice rolls, procedural generation) often requires coherent 50-80 line algorithms. Our violations are in PCG systems and demo code, not critical API handlers.

**Q: Are the goroutine leaks a problem?**  
A: Medium-risk, not high-risk. CI runs race detector successfully. We need to audit graceful shutdown, but no data races detected in 78% test coverage.

**Q: Can we rename `game.GameEvent` to `game.Event`?**  
A: Not worth it. Breaking public API for minor naming improvement. Apply strict naming to *new* code only.

**Q: What about the 12,287 magic numbers?**  
A: Most are game constants (dice faces, attribute ranges). Extract gameplay values in Phase 3; ignore import paths and low-level constants.

**Q: How do we prevent regression?**  
A: Run `go-stats-generator` monthly, enforce golangci-lint in CI, maintain 78%+ coverage threshold.

---

## Emergency Contacts

- **CI Failure:** Check `.github/workflows/ci.yml` for thresholds
- **Docker Issues:** See `Dockerfile` and `DOCKER.md`
- **API Documentation:** `pkg/README-RPC.md`
- **Full Roadmap:** `PRODUCTION_READINESS_REMEDIATION_ROADMAP.md`
- **Owner:** [@opd-ai](https://github.com/opd-ai)

---

## Quick Commands

```bash
# Run full test suite
make test && make test-e2e

# Check test coverage
make test-coverage
# Expect: ≥78% (current baseline)

# Build production binary
make build
# Output: bin/server

# Build Docker image
docker build -t goldbox-rpg:latest .

# Run with health checks
docker run -d -p 8080:8080 \
  -e WEBSOCKET_ALLOWED_ORIGINS="https://example.com" \
  goldbox-rpg:latest

# Verify health
curl http://localhost:8080/health | jq .

# Re-run production readiness analysis
go-stats-generator analyze . --skip-tests
```

---

## Approval Sign-Off

**Technical Lead:**  
- [ ] Phase 1 tasks completed (6.5 hours)
- [ ] CI pipeline passing (test, lint, format, security)
- [ ] Load testing completed (health endpoints <100ms)
- [ ] Graceful shutdown verified (SIGTERM <30s)
- [ ] Production config documented

**Product Owner:**  
- [ ] Phase 2 commitment (30 days post-launch)
- [ ] Monitoring dashboards configured
- [ ] Rollback plan documented

**Security:**  
- [ ] WebSocket origin validation enabled
- [ ] Rate limiting configured
- [ ] Input validation audited (pkg/validation/)
- [ ] govulncheck passing (no critical CVEs)

**Date Approved:** ________________  
**Production Deployment Date:** ________________

---

_Analysis powered by go-stats-generator v1.0.0_  
_Last updated: 2026-03-11_
