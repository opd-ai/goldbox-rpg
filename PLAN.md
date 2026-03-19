# Implementation Plan: Gold Box UI Authenticity & Gap Remediation

## Project Context
- **What it does**: A modern Go-based RPG engine inspired by SSI Gold Box games, providing turn-based combat, character management, and world interactions via JSON-RPC/WebSocket API with Ebitengine/WASM frontend.
- **Current goal**: Achieve "Gold Box faithful" visual and gameplay experience while closing critical implementation gaps identified in GAPS.md.
- **Estimated Scope**: Large (202 functions above complexity threshold 9.0, 10 implementation gaps, 25 roadmap items)

## Goal-Achievement Status

| Stated Goal | Current Status | This Plan Addresses |
|-------------|---------------|---------------------|
| Combat conditions (Stun, Root) affect gameplay | ❌ Empty implementations | Yes — Step 1 |
| HTTP request size limiting for DoS prevention | ⚠️ Validates after full read | Yes — Step 2 |
| Quest Builder saves to backend | ❌ Console log only | Yes — Step 3 |
| WASM Quest Editor persistence | ❌ Placeholder function | Yes — Step 4 |
| Custom character creation validates all attributes | ⚠️ Allows missing attributes | Yes — Step 5 |
| Line-of-sight respects obstacles | ⚠️ Distance check only | Yes — Step 6 |
| Frost/Lightning damage resistance mapping | ❌ Missing | Yes — Step 7 |
| WebSocket connection metrics recorded | ⚠️ Functions exist, never called | Yes — Step 8 |
| EGA-style bold panel borders | ⚠️ Subtle borders | Yes — Step 9 |
| Movement/attack range highlighting | ❌ Not implemented | Yes — Step 10 |
| Damage flash animation | ✅ Implemented | No |
| First-person dungeon viewport | ⚠️ Partial | Yes — Step 11 |
| Spatial index sort performance | ⚠️ O(n²) bubble sort | Yes — Step 12 |

## Metrics Summary

Based on `go-stats-generator analyze . --skip-tests`:

- **Total lines of code**: 37,020
- **Complexity hotspots on goal-critical paths**: 202 functions above threshold (9.0)
  - `handleAttack` (14.0) — combat condition enforcement
  - `handleApplyEffect` (13.5) — effect system
  - `processWebSocketRequest` (13.2) — metrics integration point
  - `drawCombatGrid` (inferred ~19) — range highlighting target
- **Duplication ratio**: 1.26% (918 duplicated lines across 44 clone pairs)
- **Doc coverage**: 100% (all 1,162 exported functions documented)
- **Package coupling**: game (2.98 cohesion, 1.5 coupling), server (estimated similar), wasmui (high internal coupling for UI state)

## Implementation Steps

### Step 1: Enforce Combat Condition Effects (Stun/Root)
- **Deliverable**: Modified `pkg/server/handlers.go` and `pkg/game/effectbehavior.go` to prevent actions when stunned and prevent movement when rooted
- **Dependencies**: None
- **Goal Impact**: Closes "Combat Condition Enforcement Gap" (CRITICAL severity per GAPS.md) — enables tactical combat depth
- **Acceptance**: Characters with EffectStun cannot perform any actions; characters with EffectRoot cannot move but can attack/cast
- **Validation**: 
  ```bash
  go test -run 'TestStunPreventsAction|TestRootPreventsMovement' ./pkg/game/... ./pkg/server/...
  ```

### Step 2: Apply HTTP Request Size Limit at Transport Layer
- **Deliverable**: Modified `pkg/server/server.go` to wrap `r.Body` with `io.LimitReader` before JSON decoding
- **Dependencies**: None
- **Goal Impact**: Closes "HTTP Request Size DoS Vulnerability" (HIGH severity) — prevents memory exhaustion attacks
- **Acceptance**: POST body larger than configured limit (default 1MB) rejected before full read
- **Validation**:
  ```bash
  # Server should reject immediately without allocating full buffer
  dd if=/dev/zero bs=1M count=10 | curl -X POST -H "Content-Type: application/json" --data-binary @- http://localhost:8080/rpc
  # Should return error within milliseconds, not after reading 10MB
  ```

### Step 3: Wire Quest Builder HTML Save to Backend RPC
- **Deliverable**: Modified `web/quest-builder.html` to call `questEditor.create` RPC instead of console.log
- **Dependencies**: None
- **Goal Impact**: Closes "Quest Builder Browser Save Functionality" gap — enables content creation workflow
- **Acceptance**: Quests created in browser persist to filesystem via RPC
- **Validation**:
  ```bash
  # Create quest in browser, verify file exists
  ls -la data/quests/*.yaml | grep -c 'yaml'
  ```

### Step 4: Implement WASM Quest Editor Persistence
- **Deliverable**: Modified `pkg/wasmui/quest_editor.go` to serialize quest data and call WebSocket RPC
- **Dependencies**: Step 3 (backend handler already exists)
- **Goal Impact**: Closes "WASM Quest Editor Persistence" gap — completes visual editor workflow
- **Acceptance**: Ctrl+S in WASM quest editor persists quest to server
- **Validation**:
  ```bash
  GOOS=js GOARCH=wasm go build -o /dev/null ./cmd/wasm-ui
  # Manual browser test: create quest, save, reload, verify persisted
  ```

### Step 5: Validate All Six Attributes in Custom Character Creation
- **Deliverable**: Modified `pkg/game/character_creation.go` to require all six attributes in custom mode
- **Dependencies**: None
- **Goal Impact**: Closes "Custom Character Creation Validation" gap — ensures data integrity
- **Acceptance**: Custom creation with missing attributes returns error
- **Validation**:
  ```bash
  go test -run TestCustomAttributeValidation ./pkg/game/...
  go-stats-generator analyze ./pkg/game/character_creation.go --skip-tests --format json | grep -A5 '"name": "CreateCharacter"'
  ```

### Step 6: Implement Bresenham Line-of-Sight with Obstacle Detection
- **Deliverable**: Modified `pkg/server/util.go` `isPositionVisible()` to trace tiles and check for blocking terrain
- **Dependencies**: None
- **Goal Impact**: Closes "Line-of-Sight Obstacle Detection" gap (MEDIUM severity) — enables tactical cover/positioning
- **Acceptance**: Visibility blocked by walls, not just distance
- **Validation**:
  ```bash
  go test -run TestLineOfSightBlockedByWall ./pkg/server/...
  ```

### Step 7: Add Frost and Lightning Resistance Mappings
- **Deliverable**: Modified `pkg/game/effectbehavior.go` `getResistanceForDamageType()` to include DamageFrost and DamageLightning
- **Dependencies**: None
- **Goal Impact**: Closes "Frost and Lightning Resistance Mapping" gap — completes damage type system
- **Acceptance**: Frost/Lightning damage reduced by corresponding resistance
- **Validation**:
  ```bash
  go test -run 'TestFrostResistance|TestLightningResistance' ./pkg/game/...
  ```

### Step 8: Record WebSocket Connection Metrics
- **Deliverable**: Modified `pkg/server/websocket.go` to call `metrics.RecordWebSocketConnection()` and `metrics.RecordWebSocketMessage()` at appropriate points
- **Dependencies**: None
- **Goal Impact**: Closes "WebSocket Connection Metrics" gap — enables operational monitoring
- **Acceptance**: Prometheus metrics include non-zero WebSocket connection/message counts
- **Validation**:
  ```bash
  # Start server, connect WebSocket client, then:
  curl -s localhost:8080/metrics | grep -E 'websocket_(connections|messages)_total'
  # Should show counts > 0
  ```

### Step 9: Enhance EGA-Style Bold Panel Borders
- **Deliverable**: Modified `pkg/wasmui/types_ui.go` color constants and panel drawing functions for bolder, brighter borders
- **Dependencies**: None
- **Goal Impact**: Addresses Roadmap item #18 — foundational Gold Box visual authenticity
- **Acceptance**: Panel borders clearly visible, bright EGA-inspired colors, 3-pixel thickness
- **Validation**:
  ```bash
  GOOS=js GOARCH=wasm go build -o /dev/null ./cmd/wasm-ui
  # Visual inspection via browser test
  make test-browser  # Screenshots capture border appearance
  ```

### Step 10: Implement Movement and Attack Range Highlighting
- **Deliverable**: Modified `pkg/wasmui/combat_screen.go` to draw colored overlays for reachable tiles in move/attack modes
- **Dependencies**: Step 9 (uses color constants)
- **Goal Impact**: Addresses Roadmap items #4, #5 — core combat UX improvement
- **Acceptance**: Move mode shows blue-tinted reachable tiles; attack mode shows red-tinted range
- **Validation**:
  ```bash
  GOOS=js GOARCH=wasm go build -o /dev/null ./cmd/wasm-ui
  # Visual inspection: enter combat, press M, verify blue tiles visible
  make test-browser
  ```

### Step 11: First-Person Dungeon Viewport Door Rendering
- **Deliverable**: Modified `pkg/wasmui/exploration.go` to render doors as distinct tiles in first-person view
- **Dependencies**: Steps 9, 10 (UI foundation)
- **Goal Impact**: Addresses Roadmap item #1 partial — doors distinguishable from walls
- **Acceptance**: Doors render with different color/texture than walls in first-person view
- **Validation**:
  ```bash
  GOOS=js GOARCH=wasm go build -o /dev/null ./cmd/wasm-ui
  # Visual inspection: navigate to door, verify distinct rendering
  ```

### Step 12: Replace Spatial Index Bubble Sort with QuickSort
- **Deliverable**: Modified `pkg/game/spatial_index.go` `sortByDistance()` to use `sort.Slice`
- **Dependencies**: None
- **Goal Impact**: Closes "Spatial Indexing Sort Performance" gap — O(n²) → O(n log n)
- **Acceptance**: k-nearest-neighbor queries complete in <10ms for 1000 objects
- **Validation**:
  ```bash
  go test -bench=BenchmarkGetNearestObjects -benchmem ./pkg/game/...
  # Should show <10ms for 1000 objects
  go-stats-generator analyze ./pkg/game/spatial_index.go --skip-tests --format json | grep -A3 sortByDistance
  ```

---

## Scope Assessment Calibration

| Metric | Measured Value | Severity |
|--------|----------------|----------|
| Functions above complexity 9.0 | 202 | Large (>15) |
| Duplication ratio | 1.26% | Small (<3%) |
| Doc coverage gap | 0% | None |
| Implementation gaps (GAPS.md) | 10 | Medium |
| Roadmap items incomplete | 25 | Large |

**Overall Scope**: Large — significant work across multiple packages, but well-defined acceptance criteria make each step independently testable.

---

## Notes

### Go Toolchain Security
The project currently uses Go 1.25.6 (toolchain 1.25.8). Per CHANGELOG.md, 18 known vulnerabilities exist in Go standard library affecting `crypto/tls`, `crypto/x509`, `net/http`, `net/url`, `html/template`, `os`. Recommend upgrading to Go 1.26+ when dependency compatibility allows. This is tracked but not addressed in this plan (infrastructure, not feature work).

### Complexity Hotspots on Critical Paths
Functions requiring modification have moderate-to-high complexity:
- `handleAttack` (14.0) — needs condition checks added
- `handleApplyEffect` (13.5) — already complex, changes should be minimal
- `processWebSocketRequest` (13.2) — single-line metrics calls

New code should not increase complexity. Consider extracting helpers if modifications push functions above threshold 15.

### Test Strategy
All steps include validation commands. Run existing test suite after each step:
```bash
go test -race ./pkg/game/... ./pkg/server/...
make test-coverage
```
Target: Maintain ≥65% coverage, no test regressions.

---

## Cleanup

```bash
rm -f /tmp/metrics.json /tmp/metrics_clean.json
```

---

*Generated: 2026-03-19 by go-stats-generator analysis*
