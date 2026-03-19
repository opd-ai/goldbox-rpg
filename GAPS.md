# Implementation Gaps — 2026-03-19

This document identifies gaps between the GoldBox RPG Engine's stated goals (per README.md and documentation) and its actual implementation.

---

## Stun Effect Behavioral Implementation

- **Stated Goal**: Combat conditions include "Stun" which should prevent entities from taking actions (README: "Combat conditions (Stun, Root, Burning, Bleeding, Poison)")
- **Current State**: The `EffectStun` constant is defined at `pkg/game/constants.go:47`. Effects can be applied via `ApplyEffect()`. However, the processing code at `pkg/game/effectbehavior.go:497` contains an empty case statement. No combat handler checks for stun status before allowing actions.
- **Impact**: Players and NPCs with the Stun condition can still move, attack, and cast spells. This defeats the purpose of stun-based abilities and spells, breaking tactical combat expectations.
- **Closing the Gap**: 
  1. Add `player.HasEffect(game.EffectStun)` check at the start of `handleMove()`, `handleAttack()`, and `handleCastSpell()` in `pkg/server/handlers.go`
  2. Return error "cannot act while stunned" if check passes
  3. Add behavioral effect in `processEffectTick()` (e.g., set action points to 0)
  4. Add test: `TestStunPreventsActions` in `pkg/game/effectbehavior_test.go`
  5. Verify: `go test -run TestStun ./pkg/game/... ./pkg/server/...`

---

## Root Effect Behavioral Implementation

- **Stated Goal**: Combat conditions include "Root" which should prevent movement (README: "Combat conditions (Stun, Root, Burning, Bleeding, Poison)")
- **Current State**: The `EffectRoot` constant is defined at `pkg/game/constants.go:48`. The processing code at `pkg/game/effectbehavior.go:494` contains an empty case statement. Movement handlers do not check for root status.
- **Impact**: Players and NPCs with the Root condition can still move freely. This breaks crowd-control mechanics and tactical positioning strategies.
- **Closing the Gap**:
  1. Add `player.HasEffect(game.EffectRoot)` check in `handleMove()` at `pkg/server/handlers.go:356`
  2. Return error "cannot move while rooted" if check passes
  3. Allow other actions (attack, cast) while rooted
  4. Add test: `TestRootPreventsMovement` in `pkg/server/handlers_test.go`
  5. Verify: `go test -run TestRoot ./pkg/server/...`

---

## WebSocket Session Thread Safety

- **Stated Goal**: "Concurrent player management" with "Session-based multiplayer support" (README: Real-time Communication section)
- **Current State**: Session fields `Connected` and `WSConn` are modified at `pkg/server/websocket.go:254-267` without holding the server mutex. Broadcast operations at `pkg/server/websocket.go:691-724` read these fields after releasing the lock, creating TOCTOU (Time-of-Check-Time-of-Use) race conditions.
- **Impact**: Concurrent connection/disconnection with broadcast can cause null pointer dereferences, data corruption, or panic. While panic recovery exists, this masks real bugs and can cause lost messages.
- **Closing the Gap**:
  1. Wrap `session.Connected = false; session.WSConn = nil` in mutex: `s.mu.Lock(); defer s.mu.Unlock()` at lines 254-255 and 266-267
  2. Add reference counting in broadcast snapshot: `session.addRef()` when adding to snapshot, `session.releaseRef()` after write
  3. Consider using `atomic.Bool` for `Connected` and `atomic.Value` for `WSConn`
  4. Add test: `TestConcurrentDisconnectDuringBroadcast` in `pkg/server/websocket_test.go`
  5. Verify: `go test -race -count=100 ./pkg/server/...`

---

## Effect System Resistance API

- **Stated Goal**: "Immunity and resistance handling" (README: Combat & Effects section)
- **Current State**: The `resistances` map is declared at `pkg/game/effects.go:202` and created in `NewEffectManager()`, but there is no public method to set resistance values. The `getResistanceForDamageType()` function at `pkg/game/effectbehavior.go:395-409` always returns 0 because the map is never populated.
- **Impact**: Characters cannot gain resistance to damage types through equipment, buffs, or racial abilities. The resistance system exists structurally but is non-functional.
- **Closing the Gap**:
  1. Add `SetResistance(effectType EffectType, value float64) error` method to EffectManager
  2. Add `GetResistance(effectType EffectType) float64` for reading
  3. Validate resistance value 0.0-1.0
  4. Wire to equipment bonuses and buff effects
  5. Add test: `TestResistanceReducesDamage` in `pkg/game/effectbehavior_test.go`
  6. Verify: `go test -run TestResistance ./pkg/game/...`

---

## Healing Modifier Initialization

- **Stated Goal**: "Effect stacking and priority management" with healing-over-time effects (README: Combat & Effects section)
- **Current State**: At `pkg/game/effectbehavior.go:484-485`, the healing modifier is checked with `if em.healingModifier != 0`. However, Go initializes float64 to 0.0, so an unset modifier and a "no healing" modifier are indistinguishable. The Bleeding effect's healing debuff at line 316-317 sets `healingModifier = 0.5`, but if this effect isn't active, the modifier remains 0.0 and the check fails.
- **Impact**: Healing-over-time effects may not apply correctly when no healing debuff is present, as the uninitialized state (0.0) causes the modifier path to be skipped.
- **Closing the Gap**:
  1. Initialize `healingModifier = 1.0` in `NewEffectManager()` at `pkg/game/effects.go:236`
  2. Change condition to always apply: `healing *= em.healingModifier` (no if-check needed when default is 1.0)
  3. Add test: `TestHealOverTimeWithoutDebuff` verifying full healing applies
  4. Verify: `go test -run TestHeal ./pkg/game/...`

---

## Multiplicative Modifier Stacking

- **Stated Goal**: "Stat modifications (Boosts and Penalties)" with "Effect stacking" (README: Combat & Effects section)
- **Current State**: At `pkg/game/effectmanager.go:341`, multiplicative modifiers are accumulated with formula `(multMods[mod.Stat] + 1) * (mod.Value * magnitude)`. This is mathematically incorrect for multiplicative stacking. Two 20% boosts (1.2x each) should yield 1.44x total, but this formula yields 2.64x.
- **Impact**: Multiple multiplicative buffs produce vastly inflated stat values, breaking game balance. A character with two haste effects would have 264% speed instead of 144%.
- **Closing the Gap**:
  1. Initialize `multMods[stat] = 1.0` for each stat before accumulation
  2. Change formula to `multMods[mod.Stat] *= mod.Value` (remove magnitude factor if already incorporated)
  3. Verify final application: `stat = (base + addMods[stat]) * multMods[stat]`
  4. Add test: `TestMultiplicativeStackingCorrect` with two 1.2x buffs → 1.44x
  5. Verify: `go test -run TestMultiplicative ./pkg/game/...`

---

## NPC Dialogue in Quest Builder UI

- **Stated Goal**: Quest Builder includes "NPC dialogue" (README: "Quest objective creation, reward configuration, prerequisite chains, NPC dialogue")
- **Current State**: The browser-based quest builder at `web/quest-builder.html` has sections for quest metadata, objectives, and rewards, but no dedicated NPC dialogue editing interface. The WASM component at `pkg/wasmui/quest_editor.go` supports description fields but lacks explicit dialogue trees.
- **Impact**: Content creators must manually edit YAML files to add NPC dialogue, reducing the effectiveness of the visual quest builder for complete quest authoring.
- **Closing the Gap**:
  1. Add "NPC Dialogue" section to `web/quest-builder.html` between objectives and rewards
  2. Include fields for: NPC ID, dialogue text, response options
  3. Wire to `questEditor.create` RPC call with dialogue array
  4. Update WASM quest editor to render dialogue nodes
  5. Verify: Manual testing of dialogue creation flow

---

## Liveness Probe Verification

- **Stated Goal**: "/live - Basic liveness probe for load balancers" (README: Health Check Endpoints section)
- **Current State**: The `/live` endpoint at `pkg/server/health.go:237-241` immediately returns HTTP 200 "Alive" without performing any verification. It doesn't check if the server can actually process requests.
- **Impact**: Kubernetes liveness probes may report the server as alive even if the HTTP handler goroutine is blocked or the server is in a degraded state. This could delay restarts of unhealthy pods.
- **Closing the Gap**:
  1. Option A (Minimal): Keep current implementation, document that liveness = handler responsiveness
  2. Option B (Enhanced): Add simple self-check like allocating small memory or incrementing atomic counter
  3. Document the design decision in code comments
  4. No test needed if Option A; add `TestLivenessProbeResponds` if Option B

---

## Cleric Weapon Restriction Documentation vs Implementation

- **Stated Goal**: Class proficiency system with restrictions (implied by "no edged weapons" comment in `pkg/game/classes.go:141`)
- **Current State**: The code comment mentions clerics cannot use edged weapons, but `WeaponProficiencies` for Cleric at `pkg/game/classes.go:137` simply lists `["mace", "staff", "dagger"]`. There's no "edged" property check—the restriction relies on weapon type names only.
- **Impact**: Minor documentation inconsistency. A dagger is arguably edged, yet it's in the cleric's allowed list. The system works but the comment is misleading.
- **Closing the Gap**:
  1. Option A: Remove "no edged weapons" comment since restriction is by weapon type, not property
  2. Option B: Add `IsEdged bool` property to weapons and check in `canEquipWeaponInSlot()`
  3. Verify existing tests pass: `go test -run TestClassProficiency ./pkg/game/...`

---

## Summary Table

| Gap | Severity | Effort | Priority |
|-----|----------|--------|----------|
| Stun Effect No-Op | CRITICAL | Medium | P0 |
| Root Effect No-Op | CRITICAL | Medium | P0 |
| WebSocket Race Conditions | CRITICAL | Medium | P0 |
| Resistance API Missing | HIGH | Low | P1 |
| Healing Modifier Init | HIGH | Low | P1 |
| Multiplicative Stacking Bug | HIGH | Low | P1 |
| NPC Dialogue UI | MEDIUM | Medium | P2 |
| Liveness Probe Check | LOW | Low | P3 |
| Cleric Weapon Comment | LOW | Trivial | P3 |

**Legend:**
- P0: Fix before production deployment
- P1: Fix in next sprint
- P2: Fix when touching related code
- P3: Document and defer
