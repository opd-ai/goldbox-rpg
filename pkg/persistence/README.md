# Persistence Package

The persistence package provides file-based data persistence for the GoldBox RPG Engine using YAML serialization, atomic file writes, and file locking.

## Features

- **Atomic File Writes**: Uses temporary files and atomic rename operations to prevent partial file corruption
- **File Locking**: Implements flock-based locking to prevent concurrent write conflicts
- **YAML Serialization**: Leverages existing YAML tags on game structs for persistence
- **Thread-Safe**: Safe for concurrent access within a single process
- **Nested Directories**: Automatically creates parent directories as needed

## Components

### FileStore

The main persistence interface providing Save/Load/Delete/List operations.

```go
// Create a file store
fs, err := persistence.NewFileStore("./data")
if err != nil {
    log.Fatal(err)
}

// Save data
err = fs.Save("gamestate.yaml", &gameState)

// Load data
var loaded GameState
err = fs.Load("gamestate.yaml", &loaded)

// Check existence
if fs.Exists("gamestate.yaml") {
    // File exists
}

// Delete file
err = fs.Delete("old-save.yaml")

// List files
files, err := fs.List("characters/*.yaml")
```

### AtomicWriteFile

Low-level atomic file writing function.

```go
data := []byte("content")
err := persistence.AtomicWriteFile("/path/to/file.yaml", data, 0644)
```

### FileLock

File-based locking mechanism using flock system calls.

```go
lock, err := persistence.NewFileLock("/path/to/file.yaml")
if err != nil {
    log.Fatal(err)
}
defer lock.Close()

// Blocking lock
err = lock.Lock()

// Non-blocking try lock
acquired, err := lock.TryLock()
if !acquired {
    // Lock held by another process
}

// Release lock
err = lock.Unlock()
```

## Integration with Game State

The persistence package is designed to work seamlessly with the existing YAML tags on game structures:

```go
// GameState struct already has YAML tags
type GameState struct {
    WorldState  *game.World               `yaml:"state_world"`
    TurnManager *TurnManager              `yaml:"state_turns"`
    TimeManager *TimeManager              `yaml:"state_time"`
    Sessions    map[string]*PlayerSession `yaml:"state_sessions"`
    Version     int                       `yaml:"state_version"`
}

// Save game state
fs := persistence.NewFileStore("./data")
err := fs.Save("gamestate.yaml", gameState)

// Load game state
var gs GameState
err := fs.Load("gamestate.yaml", &gs)
```

## File Structure

The persistence layer creates the following file structure:

```
data/
├── gamestate.yaml          # Main game state
├── gamestate.yaml.lock     # Lock file for gamestate
├── characters/             # Character saves
│   ├── char-123.yaml
│   ├── char-123.yaml.lock
│   ├── char-456.yaml
│   └── char-456.yaml.lock
└── sessions/               # Session snapshots (optional)
    ├── session-abc.yaml
    └── session-abc.yaml.lock
```

## Error Handling

All functions return descriptive errors:

```go
err := fs.Load("missing.yaml", &data)
// Returns: file does not exist: /data/missing.yaml

err := fs.Save("protected.yaml", &data)
// Returns: failed to acquire file lock: ...

err := fs.Load("corrupt.yaml", &data)
// Returns: failed to unmarshal YAML: ...
```

## Performance Considerations

- **Atomic writes** involve temporary file creation and rename, adding minimal overhead
- **File locking** uses efficient flock system calls with minimal blocking
- **YAML serialization** is suitable for game state but may be slower than binary formats
- **Auto-save** should use appropriate intervals (default: 30 seconds) to balance durability and performance

## Thread Safety

- FileStore methods use internal mutexes for thread-safe access
- FileLock provides process-level synchronization via flock
- Multiple FileStore instances can safely access the same data directory
- Concurrent Save/Load operations on different files are efficient

## Testing

Comprehensive test suite covering:

- Atomic file writes with various scenarios
- File locking (blocking and non-blocking)
- FileStore operations (Save/Load/Delete/List)
- Nested directory handling
- Error conditions (missing files, invalid YAML, etc.)

Run tests:
```bash
go test ./pkg/persistence/...
```

## Future Enhancements

Potential future improvements:

- Compression support for large save files
- Backup rotation and cleanup
- Database backend option
- Distributed storage integration
- Migration tools for schema changes
