// Package persistence provides file-based data persistence for the GoldBox RPG Engine.
// It implements atomic file writes, file locking, and YAML serialization for game state.
package persistence

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// AtomicWriteFile writes data to a file atomically using a temporary file and rename.
// This prevents partial file corruption if the write is interrupted.
//
// The function writes to a temporary file first, then uses os.Rename to atomically
// replace the target file. This ensures the file is either fully written or not changed at all.
//
// Parameters:
//   - filename: The target file path to write to
//   - data: The byte slice containing the data to write
//   - perm: File permissions (e.g., 0644)
//
// Returns:
//   - error: Any error that occurred during the write operation
//
// Thread-safety:
// This function is safe for concurrent use, but callers should use FileLock
// to ensure exclusive access to the target file across processes.
func AtomicWriteFile(filename string, data []byte, perm os.FileMode) error {
	logrus.WithFields(logrus.Fields{
		"function": "AtomicWriteFile",
		"filename": filename,
		"size":     len(data),
		"perm":     perm,
	}).Debug("writing file atomically")

	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Create temporary file in same directory as target
	tmpFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Ensure temp file is cleaned up on error
	defer func() {
		if tmpFile != nil {
			tmpFile.Close()
			os.Remove(tmpPath)
		}
	}()

	// Write data to temp file
	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("failed to write to temp file: %w", err)
	}

	// Sync to ensure data is written to disk
	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	// Close temp file before rename
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	tmpFile = nil // Mark as closed

	// Set correct permissions
	if err := os.Chmod(tmpPath, perm); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Atomically replace target file with temp file
	if err := os.Rename(tmpPath, filename); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"function": "AtomicWriteFile",
		"filename": filename,
		"size":     len(data),
	}).Info("file written atomically")

	return nil
}
