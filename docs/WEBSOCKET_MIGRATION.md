# WebSocket Library Migration Plan

**Document Version:** 1.0  
**Created:** 2026-03-13  
**Status:** Planning

## Overview

This document outlines the migration plan from `github.com/gorilla/websocket` to `nhooyr.io/websocket` for improved long-term maintainability.

### Why Migrate?

| Aspect | gorilla/websocket | nhooyr.io/websocket |
|--------|-------------------|---------------------|
| Maintenance Status | ⚠️ Archived (2022) | ✅ Actively maintained |
| Context Support | Manual | ✅ Native `context.Context` |
| Compression | ✅ permessage-deflate | ✅ permessage-deflate |
| API Design | Callback-based | ✅ Modern, idiomatic Go |
| Binary Size | ~15KB | ~12KB |

### Risk Assessment

- **Risk Level:** Medium
- **Blast Radius:** 9 files across server, test, and editor packages
- **Estimated Effort:** 4-8 hours

## Affected Files

| File | Role | Complexity |
|------|------|------------|
| `pkg/server/websocket.go` | Primary WebSocket handling, delta compression | High |
| `pkg/server/websocket_editor.go` | Map editor WebSocket | Medium |
| `pkg/server/types.go` | Type imports | Low |
| `pkg/server/boundary_test.go` | Test utilities | Low |
| `pkg/server/benchmark_test.go` | Performance tests | Low |
| `pkg/server/handlers_test.go` | Handler tests | Low |
| `pkg/server/websocket_test.go` | WebSocket tests | Medium |
| `pkg/server/turn_based_combat_test.go` | Combat tests | Low |
| `test/e2e/client.go` | E2E test client | Medium |

## API Differences

### Connection Upgrade

**gorilla/websocket:**
```go
upgrader := websocket.Upgrader{
    ReadBufferSize:    1024,
    WriteBufferSize:   1024,
    EnableCompression: true,
    CheckOrigin: func(r *http.Request) bool {
        return isOriginAllowed(r.Header.Get("Origin"))
    },
}
conn, err := upgrader.Upgrade(w, r, nil)
```

**nhooyr.io/websocket:**
```go
conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
    CompressionMode: websocket.CompressionContextTakeover,
    OriginPatterns:  []string{"localhost:*", "127.0.0.1:*"},
})
```

### Reading Messages

**gorilla/websocket:**
```go
messageType, p, err := conn.ReadMessage()
if err != nil {
    if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
        log.Error(err)
    }
    return
}
```

**nhooyr.io/websocket:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

messageType, p, err := conn.Read(ctx)
if err != nil {
    if websocket.CloseStatus(err) != websocket.StatusNormalClosure {
        log.Error(err)
    }
    return
}
```

### Writing Messages

**gorilla/websocket:**
```go
conn.SetWriteDeadline(time.Now().Add(writeWait))
err := conn.WriteMessage(websocket.TextMessage, data)
```

**nhooyr.io/websocket:**
```go
ctx, cancel := context.WithTimeout(context.Background(), writeWait)
defer cancel()

err := conn.Write(ctx, websocket.MessageText, data)
```

### Close Handling

**gorilla/websocket:**
```go
conn.WriteMessage(websocket.CloseMessage, 
    websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
conn.Close()
```

**nhooyr.io/websocket:**
```go
conn.Close(websocket.StatusNormalClosure, "normal closure")
```

### Ping/Pong

**gorilla/websocket:**
```go
conn.SetReadDeadline(time.Now().Add(pongWait))
conn.SetPongHandler(func(string) error {
    conn.SetReadDeadline(time.Now().Add(pongWait))
    return nil
})

// In write goroutine:
ticker := time.NewTicker(pingPeriod)
for {
    select {
    case <-ticker.C:
        conn.SetWriteDeadline(time.Now().Add(writeWait))
        conn.WriteMessage(websocket.PingMessage, nil)
    }
}
```

**nhooyr.io/websocket:**
```go
// Automatic ping/pong handled by library
// Configure with:
ctx, cancel := context.WithCancel(context.Background())
conn.CloseRead(ctx) // Enables automatic ping/pong
```

## Migration Steps

### Phase 1: Preparation (No Code Changes)

1. ✅ Create this migration document
2. Add `nhooyr.io/websocket` to `go.mod` alongside existing gorilla
3. Create feature flag for gradual rollout:
   ```go
   var UseNhooyrWebSocket = os.Getenv("USE_NHOOYR_WEBSOCKET") == "true"
   ```

### Phase 2: Adapter Layer

Create `pkg/server/websocket_adapter.go` with interface:

```go
// WebSocketConn abstracts WebSocket connections for library independence
type WebSocketConn interface {
    ReadMessage(ctx context.Context) (messageType int, p []byte, err error)
    WriteMessage(ctx context.Context, messageType int, data []byte) error
    Close(code int, reason string) error
    SetReadDeadline(t time.Time) error
    RemoteAddr() net.Addr
}
```

### Phase 3: Incremental Migration

Migrate files in order of complexity (low to high):

1. `pkg/server/types.go` - Update imports
2. `pkg/server/boundary_test.go` - Test utilities
3. `pkg/server/benchmark_test.go` - Performance baseline
4. `pkg/server/handlers_test.go` - Handler tests
5. `pkg/server/turn_based_combat_test.go` - Combat tests
6. `test/e2e/client.go` - E2E client
7. `pkg/server/websocket_test.go` - WebSocket tests
8. `pkg/server/websocket_editor.go` - Editor WebSocket
9. `pkg/server/websocket.go` - Primary handler (last)

### Phase 4: Cleanup

1. Remove gorilla/websocket from `go.mod`
2. Delete adapter layer if no longer needed
3. Update vendor directory
4. Remove feature flag

## Rollback Plan

If issues are discovered post-migration:

1. Revert `go.mod` to restore gorilla/websocket
2. Revert affected source files
3. Run `go mod tidy && go mod vendor`
4. Verify all tests pass: `go test -race ./...`

Git commands for quick rollback:
```bash
# Tag pre-migration state before starting
git tag pre-websocket-migration

# Rollback if needed
git checkout pre-websocket-migration -- go.mod go.sum pkg/server/ test/e2e/client.go
go mod vendor
```

## Validation Checklist

### Functional Requirements

- [ ] All existing E2E tests pass (`go test ./test/e2e/... -v`)
- [ ] WebSocket connections establish successfully
- [ ] Delta compression works (95%+ bandwidth reduction)
- [ ] Origin validation enforced in production mode
- [ ] Editor WebSocket protocol unchanged
- [ ] Graceful connection closure

### Performance Requirements

- [ ] Benchmark within 10% of gorilla baseline
- [ ] No increase in connection latency
- [ ] Memory usage comparable or lower

### Security Requirements

- [ ] WEBSOCKET_ALLOWED_ORIGINS respected
- [ ] Dev mode origin bypass works
- [ ] No new security warnings from golangci-lint

## Current WebSocket Features to Preserve

### 1. Delta Compression (`websocket.go`)
The current implementation tracks `LastState` per connection and sends only changed fields. This must be preserved as-is.

### 2. Origin Validation
```go
func (s *RPCServer) getAllowedOrigins() []string
func (s *RPCServer) isOriginAllowed(origin string, allowedOrigins []string) bool
```

### 3. Editor Protocol (`websocket_editor.go`)
Event types: `tile_update`, `map_created`, `map_loaded`, `map_saved`, `cursor_move`, `select_tool`, `undo_redo`

### 4. Thread-Safe Writes
```go
type wsConnection struct {
    conn *websocket.Conn
    mu   sync.Mutex
}
```

## References

- nhooyr.io/websocket documentation: https://pkg.go.dev/nhooyr.io/websocket
- gorilla/websocket archive notice: https://github.com/gorilla/websocket
- Migration examples: https://github.com/nhooyr/websocket/tree/master/examples
