# Audit: goldbox-rpg/pkg/validation
**Date**: 2026-02-18
**Status**: Needs Work

## Summary
The validation package provides security-critical input validation for JSON-RPC requests. Overall code quality is good with proper error handling and comprehensive test coverage. However, critical issues exist: validation is not actually invoked in request processing, documentation contradicts implementation, character class validation is misaligned with game constants, and the package has below-target test coverage at 52.1%.

## Issues Found
- [ ] **high** Stub/Incomplete — Validator is instantiated but never called in server request processing (pkg/server/server.go:loadServerConfiguration creates validator but no ValidateRPCRequest calls found in handlers.go)
- [ ] **high** API Design — Documentation-implementation mismatch: README.md documents RegisterValidator() method that doesn't exist (validation/README.md:39-42, validation.go has no RegisterValidator method, only private registerValidators)
- [ ] **high** Determinism — Character class validation misaligned with game constants: validator accepts "wizard", "magic-user", "elf", "dwarf", "halfling" but game only defines Fighter, Mage, Cleric, Thief, Ranger, Paladin (validation.go:499-502 vs pkg/game/constants.go:112-118)
- [ ] **med** Error Handling — README.md documents error constants (ErrInvalidParameterType, ErrMissingRequiredField, etc.) that don't exist in implementation (validation/README.md:140-148, validation.go uses inline fmt.Errorf without exported error constants)
- [ ] **med** Concurrency Safety — Global logrus configuration in init() affects entire process, not just validation package (validation.go:16-19 calls logrus.SetReportCaller(true) globally)
- [ ] **med** Test Coverage — Below 65% target at 52.1%, missing tests for useItem and leaveGame validators (validation_test.go has no TestValidateUseItem or TestValidateLeaveGame)
- [ ] **low** Documentation — Missing package doc.go file per coding guidelines
- [ ] **low** API Design — README.md describes ValidateEventData method that doesn't exist (validation/README.md:192-194)
- [ ] **low** API Design — Inconsistent parameter naming: "item_id" in validateUseItem but "itemId" in validateEquipItem (validation.go:566 vs validation.go:394)

## Test Coverage
52.1% (target: 65%)

Missing test coverage for:
- validateUseItem method (validation.go:554-592)
- validateLeaveGame method (validation.go:594-596)
- Edge cases in registerValidators
- Request size validation edge cases

## Dependencies
**Internal Dependencies:**
- github.com/sirupsen/logrus (logging)

**Importers (Integration Surface):**
- pkg/server (server.go, session_timeout_fix_test.go, missing_methods_test.go)

**External Dependencies:**
- Standard library only (fmt, regexp, strings, unicode/utf8)

**No circular dependencies detected.**

## Recommendations
1. **CRITICAL:** Wire up validator in server request processing - add ValidateRPCRequest call in server.go handleMethod before processing any request
2. **HIGH:** Fix character class validation to match game constants - remove "wizard", "magic-user", "elf", "dwarf", "halfling" and add "mage", or create mapping layer
3. **HIGH:** Remove or implement documented RegisterValidator API from README.md - either delete documentation or add public method
4. **MEDIUM:** Export error constants as documented - add var declarations for ErrInvalidParameterType, ErrMissingRequiredField, etc.
5. **MEDIUM:** Remove global logrus configuration from init() - use structured logger passed to validator instead
6. **MEDIUM:** Add tests for useItem and leaveGame validators to reach 65% coverage target
7. **LOW:** Add doc.go with package-level documentation
8. **LOW:** Standardize parameter naming convention (snake_case vs camelCase)
9. **LOW:** Remove undocumented ValidateEventData references from README.md
