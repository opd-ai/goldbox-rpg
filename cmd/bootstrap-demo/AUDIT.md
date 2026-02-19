# Audit: goldbox-rpg/cmd/bootstrap-demo
**Date**: 2026-02-19
**Status**: Complete

## Summary
Feature-rich CLI demo showcasing procedural game generation with template support and configuration options. Good input validation for enum values.

## Issues Found
- [ ] **med** Security — Output directory cleanup via `os.RemoveAll()` without confirmation could delete user data (`main.go:152`)
- [ ] **med** Configuration — Data directory path `"data"` hardcoded (`main.go:127,133`)
- [ ] **low** Validation — No bounds checking for MaxPlayers/StartingLevel parameters (`main.go:100-101`)
- [ ] **low** Security — No validation of file paths in `-output` flag; could allow path traversal

## Test Coverage
N/A — Demo application (no test files, appropriate for cmd package)

## Dependencies
- `goldbox-rpg/pkg/game`, `goldbox-rpg/pkg/pcg`
- `github.com/sirupsen/logrus`
