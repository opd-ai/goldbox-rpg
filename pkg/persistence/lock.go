package persistence

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/sirupsen/logrus"
)

// FileLock provides file-based locking to prevent concurrent writes.
// It uses flock system calls to ensure exclusive access to files.
//
// This is important for preventing corruption when multiple processes
// or goroutines attempt to write to the same file simultaneously.
type FileLock struct {
	file     *os.File
	path     string
	isLocked bool
}

// NewFileLock creates a new file lock for the given path.
// The lock file is created in the same directory with a .lock extension.
//
// Parameters:
//   - path: The file path to create a lock for
//
// Returns:
//   - *FileLock: A new file lock instance
//   - error: Any error that occurred during lock file creation
func NewFileLock(path string) (*FileLock, error) {
	lockPath := path + ".lock"

	logrus.WithFields(logrus.Fields{
		"function": "NewFileLock",
		"path":     path,
		"lockPath": lockPath,
	}).Debug("creating file lock")

	// Ensure directory exists for lock file
	lockDir := filepath.Dir(lockPath)
	if err := os.MkdirAll(lockDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create lock directory: %w", err)
	}

	// Create lock file if it doesn't exist
	file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to create lock file: %w", err)
	}

	return &FileLock{
		file:     file,
		path:     lockPath,
		isLocked: false,
	}, nil
}

// Lock acquires an exclusive lock on the file.
// This call will block until the lock is acquired.
//
// Returns:
//   - error: Any error that occurred while acquiring the lock
func (fl *FileLock) Lock() error {
	if fl.isLocked {
		return fmt.Errorf("lock already held")
	}

	logrus.WithFields(logrus.Fields{
		"function": "Lock",
		"path":     fl.path,
	}).Debug("acquiring file lock")

	// Acquire exclusive lock (blocking)
	if err := syscall.Flock(int(fl.file.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	fl.isLocked = true

	logrus.WithFields(logrus.Fields{
		"function": "Lock",
		"path":     fl.path,
	}).Debug("file lock acquired")

	return nil
}

// TryLock attempts to acquire an exclusive lock without blocking.
// Returns immediately with an error if the lock is held by another process.
//
// Returns:
//   - bool: true if lock was acquired, false if already locked
//   - error: Any error that occurred while attempting to lock
func (fl *FileLock) TryLock() (bool, error) {
	if fl.isLocked {
		return false, fmt.Errorf("lock already held by this instance")
	}

	// Try to acquire exclusive lock (non-blocking)
	err := syscall.Flock(int(fl.file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		if err == syscall.EWOULDBLOCK {
			return false, nil // Lock is held by another process
		}
		return false, fmt.Errorf("failed to try lock: %w", err)
	}

	fl.isLocked = true
	return true, nil
}

// Unlock releases the exclusive lock on the file.
//
// Returns:
//   - error: Any error that occurred while releasing the lock
func (fl *FileLock) Unlock() error {
	if !fl.isLocked {
		return nil // Already unlocked
	}

	logrus.WithFields(logrus.Fields{
		"function": "Unlock",
		"path":     fl.path,
	}).Debug("releasing file lock")

	// Release lock
	if err := syscall.Flock(int(fl.file.Fd()), syscall.LOCK_UN); err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	fl.isLocked = false

	logrus.WithFields(logrus.Fields{
		"function": "Unlock",
		"path":     fl.path,
	}).Debug("file lock released")

	return nil
}

// Close closes the lock file and releases the lock if held.
// This should be called when done with the lock to clean up resources.
//
// Returns:
//   - error: Any error that occurred during cleanup
func (fl *FileLock) Close() error {
	// Release lock if held
	if fl.isLocked {
		if err := fl.Unlock(); err != nil {
			return err
		}
	}

	// Close file
	if fl.file != nil {
		if err := fl.file.Close(); err != nil {
			return fmt.Errorf("failed to close lock file: %w", err)
		}
		fl.file = nil
	}

	return nil
}
