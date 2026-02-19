# Audit: goldbox-rpg/pkg/persistence
**Date**: 2026-02-19
**Status**: Needs Work

## Summary
File-based persistence package providing atomic writes, advisory file locking (flock), and YAML serialization for game state storage. Implements FileStore with RWMutex protection and FileLock for cross-process safety. Contains critical deadlock risk from nested lock acquisition and missing path validation.

## Issues Found
- [ ] **high** Deadlock — Save() acquires FileLock while holding RWMutex; if another goroutine holds FileLock and tries to acquire RWMutex, deadlock occurs (`filestore.go:57-97`)
- [ ] **high** Error Handling — Delete() silently ignores lock file cleanup errors; failed lock removal causes subsequent operations to fail (`filestore.go:204-206`)
- [ ] **high** Error Handling — Exists() returns false for both "file not found" and "permission denied"; cannot distinguish real errors (`filestore.go:162-166`)
- [ ] **high** Security — No validation that filenames don't contain `../`; could write outside dataDir via path traversal (`filestore.go:60-61`)
- [ ] **med** Atomicity — AtomicWriteFile() syncs file but doesn't fsync parent directory; file can be lost on crash on some filesystems (`atomic.go:65`)
- [ ] **med** Test Coverage — Missing concurrent Save/Load stress tests, lock contention scenarios, permission error handling tests
- [ ] **med** Concurrency — FileLock `isLocked` flag is a plain bool without atomic access; race condition under heavy contention (`lock.go`)
- [ ] **low** Documentation — Missing package-level doc.go file
- [ ] **low** Documentation — README mentions distributed storage and database backends as future work but creates false expectations

## Test Coverage
77.1% (target: 65%) — ✅ ABOVE TARGET

1 test file (279 lines) covering basic Save/Load/Delete operations. Missing coverage for concurrent access patterns and error edge cases.

## Dependencies
**External:**
- `github.com/sirupsen/logrus`: Logging
- `gopkg.in/yaml.v3`: YAML serialization
- `github.com/stretchr/testify`: Testing (test only)

**Internal:** None (standalone package)

## Recommendations
1. **CRITICAL**: Fix deadlock — release RWMutex before acquiring FileLock, or use single synchronization mechanism
2. **CRITICAL**: Validate file paths with filepath.Clean() and ensure they don't escape dataDir
3. **HIGH**: Handle lock file cleanup errors in Delete(); add lock recovery mechanism
4. **HIGH**: Use sync/atomic for FileLock state tracking
5. **MEDIUM**: Add fsync of parent directory in AtomicWriteFile for full crash safety
6. **MEDIUM**: Distinguish error types in Exists() (permission denied vs not found)
7. **LOW**: Add doc.go with package documentation
