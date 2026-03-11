# Error Handling Guide

## Overview

The GoldBox RPG Engine implements structured error handling using domain-specific error types with proper error wrapping. This guide describes the error handling patterns and best practices used throughout the codebase.

## Error Handling Principles

1. **Use Sentinel Errors**: Define sentinel errors as package-level variables for common error conditions
2. **Wrap Errors**: Always wrap errors with context using `fmt.Errorf` with `%w` verb
3. **Custom Error Types**: Use custom error types for domain-specific errors with structured context
4. **Error Inspection**: Use `errors.Is()` and `errors.As()` for error inspection
5. **Clear Messages**: Provide descriptive error messages with relevant context

## Package-Specific Errors

### Game Package (`pkg/game/errors.go`)

The game package defines errors for core game mechanics:

#### Sentinel Errors

```go
// Character errors
ErrCharacterNotFound    // Character not found
ErrInvalidCharacterName // Invalid character name
ErrInsufficientHP       // Insufficient hit points
ErrInsufficientMP       // Insufficient mana points
ErrInsufficientAP       // Insufficient action points
ErrExperienceOverflow   // Experience overflow
ErrNegativeExperience   // Negative experience not allowed

// Combat errors
ErrInvalidTarget      // Invalid target
ErrTargetOutOfRange   // Target out of range
ErrCannotReceiveDamage // Target cannot receive damage
ErrEffectReflected    // Effect reflected

// Inventory errors
ErrItemNotFound      // Item not found in inventory
ErrEmptyItemID       // Item ID cannot be empty
ErrCarryingCapacity  // Carrying capacity exceeded
ErrCannotEquipItem   // Cannot equip item
ErrInvalidSlot       // Invalid equipment slot

// Effect errors
ErrEffectNotFound    // Effect not found
ErrWeakerEffect      // Cannot apply weaker effect of same type
ErrInvalidEffectType // Invalid effect type
ErrEffectImmunity    // Target is immune to effect

// Spell errors
ErrCannotCastSpells      // Class cannot cast spells
ErrInvalidSpellLevel     // Invalid spell level
ErrSpellNotKnown         // Spell not known by character
ErrInvalidDiceExpression // Invalid dice expression

// Spatial errors
ErrObjectNotFound  // Object not found
ErrOutOfBounds     // Position outside bounds
ErrInvalidPosition // Invalid position

// World errors
ErrWorldNotInitialized // World not initialized
ErrInvalidWorldState   // Invalid world state
```

#### Custom Error Types

**CharacterError**: Errors related to character operations
```go
err := game.NewCharacterError(charID, "addExperience", game.ErrExperienceOverflow)
// Output: character abc123: addExperience: experience overflow
```

**InventoryError**: Errors related to inventory operations
```go
err := game.NewInventoryError(itemID, "addItem", game.ErrCarryingCapacity)
// Output: inventory: addItem: item sword-01: carrying capacity exceeded
```

**CombatError**: Errors during combat operations
```go
err := game.NewCombatError(attackerID, targetID, "attack", game.ErrTargetOutOfRange)
// Output: combat: attack: attacker player-1 -> target enemy-5: target out of range
```

**EffectError**: Errors related to effect management
```go
err := game.NewEffectError(effectID, targetID, "stun", "apply", game.ErrEffectImmunity)
// Output: effect: apply: effect stun-01 on target player-1 (type: stun): target is immune to effect
```

**SpatialError**: Errors in spatial indexing operations
```go
err := game.NewSpatialError(objectID, &position, "insert", game.ErrOutOfBounds)
// Output: spatial: insert: object obj-123 at position (150,200): position outside bounds
```

**SpellError**: Errors related to spell operations
```go
err := game.NewSpellError(spellID, casterID, targetID, "cast", game.ErrInsufficientMP)
// Output: spell: cast: spell fireball cast by wizard-1: insufficient mana points
```

### Server Package (`pkg/server/errors.go`)

The server package defines errors for server operations:

#### Sentinel Errors

```go
// Session errors
ErrInvalidSession   // Invalid session
ErrSessionNotFound  // Session not found
ErrSessionExpired   // Session expired
ErrSessionCreation  // Session creation failed

// Request validation errors
ErrInvalidRequest // Invalid request
ErrInvalidParams  // Invalid parameters
ErrMissingParams  // Missing required parameters
ErrInvalidMethod  // Invalid method

// Game state errors
ErrGameStateNotFound  // Game state not found
ErrGameStateCorrupted // Game state corrupted
ErrGameStateNil       // Game state is nil

// Server health errors
ErrServerShuttingDown // Server is shutting down
ErrServerNotReady     // Server not ready
ErrServerNil          // Server instance is nil

// Component initialization errors
ErrWorldNil          // World state is nil
ErrSpellManagerNil   // Spell manager not initialized
ErrEventSystemNil    // Event system not initialized
ErrPCGManagerNil     // PCG manager not initialized
ErrCircuitBreakerNil // Circuit breaker manager not initialized
ErrConfigNil         // Configuration not initialized

// Persistence errors
ErrPersistenceFailed // Persistence operation failed
ErrLoadFailed        // Load operation failed
ErrSaveFailed        // Save operation failed

// WebSocket errors
ErrWebSocketClosed  // Websocket connection closed
ErrWebSocketUpgrade // Websocket upgrade failed
ErrInvalidOrigin    // Invalid websocket origin
```

#### Custom Error Types

**SessionError**: Errors related to session management
```go
err := server.NewSessionError(sessionID, "validate", server.ErrSessionExpired)
// Output: session sess-456: validate: session expired
```

**ValidationError**: Request validation errors
```go
err := server.NewValidationError("handleMove", "targetX", -5, server.ErrInvalidParams)
// Output: validation error: method handleMove: parameter targetX (value: -5): invalid parameters
```

**PersistenceError**: Errors during persistence operations
```go
err := server.NewPersistenceError(sessionID, filePath, "save", server.ErrSaveFailed)
// Output: persistence: save: session sess-789: file /data/session.yaml: save operation failed
```

**HealthCheckError**: Errors during health check operations
```go
err := server.NewHealthCheckError("spell_manager", "initialization", server.ErrSpellManagerNil)
// Output: health check: spell_manager: initialization: spell manager not initialized
```

**RPCError**: Errors in RPC request processing
```go
err := server.NewRPCError("getState", requestID, server.ErrInvalidSession)
// Output: RPC: method getState: request 42: invalid session
```

**WebSocketError**: Errors in WebSocket operations
```go
err := server.NewWebSocketError(clientID, "broadcast", server.ErrWebSocketClosed)
// Output: websocket: client ws-001: broadcast: websocket connection closed
```

## Usage Patterns

### Creating and Returning Errors

#### Simple Sentinel Errors
```go
if characterID == "" {
    return game.ErrCharacterNotFound
}
```

#### Wrapped Errors with Context
```go
if err := saveCharacter(char); err != nil {
    return fmt.Errorf("failed to save character %s: %w", char.ID, err)
}
```

#### Custom Error Types
```go
if item == nil {
    return game.NewInventoryError(itemID, "equip", game.ErrItemNotFound)
}
```

### Checking Errors

#### Using errors.Is for Sentinel Errors
```go
if errors.Is(err, game.ErrCharacterNotFound) {
    // Handle character not found
    return http.StatusNotFound
}
```

#### Using errors.As for Custom Error Types
```go
var invErr *game.InventoryError
if errors.As(err, &invErr) {
    log.WithFields(logrus.Fields{
        "item_id":      invErr.ItemID,
        "operation":    invErr.Operation,
        "current_load": invErr.CurrentLoad,
        "max_load":     invErr.MaxLoad,
    }).Error("Inventory operation failed")
}
```

#### Checking Multiple Error Types
```go
if errors.Is(err, game.ErrInsufficientMP) || errors.Is(err, game.ErrInsufficientAP) {
    return fmt.Errorf("insufficient resources: %w", err)
}
```

### Logging with Error Context

```go
var sessErr *server.SessionError
if errors.As(err, &sessErr) {
    logrus.WithFields(logrus.Fields{
        "session_id": sessErr.SessionID,
        "operation":  sessErr.Operation,
        "error":      sessErr.Err,
    }).Error("Session operation failed")
}
```

## Migration Guide

### Replacing fmt.Errorf Strings

**Before:**
```go
return fmt.Errorf("character not found: %s", characterID)
```

**After:**
```go
return game.NewCharacterError(characterID, "lookup", game.ErrCharacterNotFound)
```

### Replacing errors.New

**Before:**
```go
if sessionID == "" {
    return errors.New("invalid session")
}
```

**After:**
```go
if sessionID == "" {
    return server.ErrInvalidSession
}
```

### Adding Error Context

**Before:**
```go
return fmt.Errorf("failed to apply damage: %w", err)
```

**After:**
```go
return game.NewCombatError(attackerID, targetID, "applyDamage", err)
```

## Best Practices

1. **Always Use %w for Wrapping**: Use `%w` verb in `fmt.Errorf` to preserve error chain
   ```go
   return fmt.Errorf("operation failed: %w", err)
   ```

2. **Prefer Sentinel Errors**: Use sentinel errors for well-known error conditions
   ```go
   if hp <= 0 {
       return game.ErrInsufficientHP
   }
   ```

3. **Add Context with Custom Types**: Use custom error types when you need structured context
   ```go
   return game.NewInventoryError(itemID, "remove", game.ErrItemNotFound)
   ```

4. **Check Errors Properly**: Use `errors.Is()` for sentinel errors and `errors.As()` for custom types
   ```go
   if errors.Is(err, game.ErrCharacterNotFound) { ... }
   
   var charErr *game.CharacterError
   if errors.As(err, &charErr) { ... }
   ```

5. **Log with Structured Fields**: Extract error context for structured logging
   ```go
   var rpcErr *server.RPCError
   if errors.As(err, &rpcErr) {
       logrus.WithFields(logrus.Fields{
           "method": rpcErr.Method,
           "request_id": rpcErr.RequestID,
       }).Error("RPC call failed")
   }
   ```

6. **Don't Lose Error Information**: Always preserve the original error
   ```go
   // Bad
   return fmt.Errorf("operation failed")
   
   // Good
   return fmt.Errorf("operation failed: %w", originalErr)
   ```

7. **Be Specific with Error Messages**: Provide actionable context
   ```go
   // Bad
   return errors.New("validation failed")
   
   // Good
   return server.NewValidationError("handleMove", "targetX", -5, server.ErrInvalidParams)
   ```

## Testing with Errors

### Testing for Specific Errors
```go
func TestCharacterOperation(t *testing.T) {
    err := character.AddExperience(-100)
    assert.ErrorIs(t, err, game.ErrNegativeExperience)
}
```

### Testing Custom Error Types
```go
func TestInventoryOperation(t *testing.T) {
    err := inventory.AddItem(heavyItem)
    
    var invErr *game.InventoryError
    require.ErrorAs(t, err, &invErr)
    assert.Equal(t, heavyItem.ID, invErr.ItemID)
    assert.Equal(t, "addItem", invErr.Operation)
    assert.ErrorIs(t, invErr.Err, game.ErrCarryingCapacity)
}
```

## Performance Considerations

1. **Sentinel Errors are Cheap**: Sentinel errors are pre-allocated and have minimal overhead
2. **Custom Types Add Context**: Custom error types add small overhead but provide valuable context
3. **Error Wrapping is Fast**: Error wrapping with `%w` is optimized by the Go runtime
4. **Avoid Deep Chains**: Keep error chains reasonably shallow (3-5 levels maximum)

## See Also

- [Go Error Handling Best Practices](https://golang.org/blog/error-handling-and-go)
- [Working with Errors in Go 1.13](https://golang.org/blog/go1.13-errors)
- `pkg/game/errors.go` - Game package error definitions
- `pkg/server/errors.go` - Server package error definitions
