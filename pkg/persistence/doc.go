// Package persistence provides file-based data persistence for the GoldBox RPG Engine.
//
// This package handles game state storage with atomic writes, file locking, and
// YAML serialization to ensure data integrity and protection against corruption
// from concurrent access or crashes.
//
// # FileStore
//
// FileStore is the primary interface for persisting game data:
//
//	store := persistence.NewFileStore("/path/to/data")
//
//	// Save game state
//	err := store.Save("game.yaml", gameState)
//
//	// Load game state
//	var loaded GameState
//	err := store.Load("game.yaml", &loaded)
//
// # Atomic Writes
//
// All write operations use atomic file replacement to prevent corruption:
//
//  1. Data is written to a temporary file
//  2. Temporary file is synced to disk
//  3. Temporary file is renamed to target (atomic operation)
//
// This ensures that even if a crash occurs during save, the original file
// remains intact.
//
// # File Locking
//
// FileLock provides cross-process synchronization using flock syscalls:
//
//	lock := persistence.NewFileLock("/path/to/lockfile")
//
//	// Blocking lock acquisition
//	if err := lock.Lock(); err != nil {
//	    return err
//	}
//	defer lock.Unlock()
//
//	// Non-blocking lock attempt
//	acquired, err := lock.TryLock()
//	if !acquired {
//	    return errors.New("resource busy")
//	}
//
// # File Operations
//
// Additional file management methods:
//
//	// Check existence
//	if store.Exists("character.yaml") {
//	    // File exists
//	}
//
//	// Delete file and associated lock
//	err := store.Delete("old-save.yaml")
//
//	// List files matching pattern
//	files, err := store.List("saves/*.yaml")
//
// # YAML Serialization
//
// Data is serialized using YAML for human-readable storage. Types should
// use yaml struct tags for field mapping:
//
//	type Character struct {
//	    Name  string `yaml:"name"`
//	    Level int    `yaml:"level"`
//	}
//
// # Thread Safety
//
// FileStore operations are protected by internal mutexes for safe concurrent
// access within a single process. FileLock extends protection across processes.
//
// # Platform Support
//
// File locking uses Unix flock syscalls. The package includes build tags
// for platform-specific implementations.
package persistence
