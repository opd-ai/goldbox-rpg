# WebSocket Library Migration Plan

**Document Version:** 2.0  
**Created:** 2026-03-13  
**Updated:** 2026-03-13  
**Status:** ✅ COMPLETE

## Overview

This document outlines the migration from `github.com/gorilla/websocket` to `nhooyr.io/websocket` for improved long-term maintainability.

**Migration Status:** The server has been successfully migrated to nhooyr.io/websocket. Test clients retain gorilla/websocket for backward compatibility (WebSocket is a standard protocol).

### Why We Migrated

| Aspect | gorilla/websocket | nhooyr.io/websocket |
|--------|-------------------|---------------------|
| Maintenance Status | ⚠️ Archived (2022) | ✅ Actively maintained |
| Context Support | Manual | ✅ Native `context.Context` |
| Compression | ✅ permessage-deflate | ✅ permessage-deflate |
| API Design | Callback-based | ✅ Modern, idiomatic Go |
| Binary Size | ~15KB | ~12KB |

## Migration Summary

### Completed Changes (2026-03-13)

1. **Created adapter interface** (`pkg/server/websocket_adapter.go`)
   - `WebSocketConn` interface abstracts WebSocket connections
   - Supports both libraries with same API

2. **Added nhooyr implementation** (`pkg/server/websocket_nhooyr.go`)
   - `nhooyrWebSocketConn` implements `WebSocketConn`
   - Full context support for cancellation/deadlines

3. **Created upgrade function** (`pkg/server/websocket_upgrade.go`)
   - `upgradeConnectionNhooyr()` handles HTTP→WebSocket upgrade
   - `CheckOriginAllowed()` for origin validation testing
   - Origin pattern extraction from URLs

4. **Updated main WebSocket handler** (`pkg/server/websocket.go`)
   - Removed gorilla import
   - Uses nhooyr via adapter interface

5. **Deleted gorilla adapter** (`pkg/server/websocket_gorilla.go`)
   - No longer needed

6. **Updated tests**
   - Origin validation tests use `CheckOriginAllowed()` method
   - All tests pass with nhooyr server + gorilla test clients

### Files Changed

| File | Change |
|------|--------|
| `pkg/server/websocket_adapter.go` | Interface for WebSocket abstraction |
| `pkg/server/websocket_nhooyr.go` | nhooyr implementation (removed build tag) |
| `pkg/server/websocket_upgrade.go` | NEW - nhooyr upgrade + origin validation |
| `pkg/server/websocket.go` | Removed gorilla import, uses adapter |
| `pkg/server/websocket_gorilla.go` | DELETED |
| `pkg/server/websocket_test.go` | Updated origin tests |
| `pkg/server/websocket_origin_validation_test.go` | Updated to use CheckOriginAllowed |
| `pkg/server/websocket_allowed_origins_fix_test.go` | Updated to use CheckOriginAllowed |

### Preserved Features

All existing WebSocket features work correctly:

1. **Delta Compression** - `LastState` tracking unchanged, 95%+ bandwidth savings
2. **Origin Validation** - `WEBSOCKET_ALLOWED_ORIGINS` environment variable
3. **Editor Protocol** - All event types (`tile_update`, `map_created`, etc.)
4. **Thread-Safe Writes** - Mutex protection via `WSWriteMu`
5. **Rate Limiting** - WebSocket requests are rate-limited

### Validation Results

- ✅ All E2E tests pass (`go test ./test/e2e/... -v`)
- ✅ All server tests pass (`go test ./pkg/server/... -v`)
- ✅ Race detector clean (`go test -race ./...`)
- ✅ go vet clean (`go vet ./...`)
- ✅ Gorilla client ↔ nhooyr server interoperability confirmed

## References

- nhooyr.io/websocket documentation: https://pkg.go.dev/nhooyr.io/websocket
- gorilla/websocket archive notice: https://github.com/gorilla/websocket
