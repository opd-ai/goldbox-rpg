package persistence

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// FileStore provides file-based persistence for game data using YAML serialization.
// It supports atomic writes, file locking, and automatic directory management.
//
// FileStore is thread-safe for concurrent access within a single process.
// For cross-process safety, use the file locking mechanisms.
type FileStore struct {
	dataDir string
	mu      sync.RWMutex
}

// NewFileStore creates a new FileStore instance.
//
// Parameters:
//   - dataDir: The directory where data files will be stored
//
// Returns:
//   - *FileStore: A new FileStore instance
//   - error: Any error that occurred during initialization
func NewFileStore(dataDir string) (*FileStore, error) {
	logrus.WithFields(logrus.Fields{
		"function": "NewFileStore",
		"dataDir":  dataDir,
	}).Info("creating new file store")

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return &FileStore{
		dataDir: dataDir,
	}, nil
}

// Save serializes an object to YAML and saves it to a file.
// The save operation is atomic and uses file locking to prevent corruption.
//
// Parameters:
//   - filename: The name of the file (relative to dataDir)
//   - data: The object to serialize and save
//
// Returns:
//   - error: Any error that occurred during the save operation
func (fs *FileStore) Save(filename string, data interface{}) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fullPath := filepath.Join(fs.dataDir, filename)

	logrus.WithFields(logrus.Fields{
		"function": "Save",
		"filename": filename,
		"fullPath": fullPath,
	}).Debug("saving data to file")

	// Acquire file lock
	lock, err := NewFileLock(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file lock: %w", err)
	}
	defer lock.Close()

	if err := lock.Lock(); err != nil {
		return fmt.Errorf("failed to acquire file lock: %w", err)
	}

	// Marshal data to YAML
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data to YAML: %w", err)
	}

	// Write atomically
	if err := AtomicWriteFile(fullPath, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"function": "Save",
		"filename": filename,
		"size":     len(yamlData),
	}).Info("data saved successfully")

	return nil
}

// Load reads a file and deserializes it from YAML into the provided object.
//
// Parameters:
//   - filename: The name of the file (relative to dataDir)
//   - data: A pointer to the object to deserialize into
//
// Returns:
//   - error: Any error that occurred during the load operation
func (fs *FileStore) Load(filename string, data interface{}) error {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	fullPath := filepath.Join(fs.dataDir, filename)

	logrus.WithFields(logrus.Fields{
		"function": "Load",
		"filename": filename,
		"fullPath": fullPath,
	}).Debug("loading data from file")

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", fullPath)
	}

	// Acquire read lock
	lock, err := NewFileLock(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file lock: %w", err)
	}
	defer lock.Close()

	if err := lock.Lock(); err != nil {
		return fmt.Errorf("failed to acquire file lock: %w", err)
	}

	// Read file
	yamlData, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Unmarshal YAML
	if err := yaml.Unmarshal(yamlData, data); err != nil {
		return fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"function": "Load",
		"filename": filename,
		"size":     len(yamlData),
	}).Info("data loaded successfully")

	return nil
}

// Exists checks if a file exists in the file store.
//
// Parameters:
//   - filename: The name of the file (relative to dataDir)
//
// Returns:
//   - bool: true if the file exists, false otherwise
func (fs *FileStore) Exists(filename string) bool {
	fullPath := filepath.Join(fs.dataDir, filename)
	_, err := os.Stat(fullPath)
	return err == nil
}

// Delete removes a file from the file store.
//
// Parameters:
//   - filename: The name of the file (relative to dataDir)
//
// Returns:
//   - error: Any error that occurred during deletion
func (fs *FileStore) Delete(filename string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fullPath := filepath.Join(fs.dataDir, filename)

	logrus.WithFields(logrus.Fields{
		"function": "Delete",
		"filename": filename,
		"fullPath": fullPath,
	}).Debug("deleting file")

	// Acquire file lock before deletion
	lock, err := NewFileLock(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file lock: %w", err)
	}
	defer lock.Close()

	if err := lock.Lock(); err != nil {
		return fmt.Errorf("failed to acquire file lock: %w", err)
	}

	// Delete file
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Delete lock file
	lockPath := fullPath + ".lock"
	os.Remove(lockPath) // Ignore errors

	logrus.WithFields(logrus.Fields{
		"function": "Delete",
		"filename": filename,
	}).Info("file deleted successfully")

	return nil
}

// List returns a list of all files in the data directory matching a pattern.
//
// Parameters:
//   - pattern: Glob pattern to match files (e.g., "*.yaml", "characters/*")
//
// Returns:
//   - []string: List of matching filenames (relative to dataDir)
//   - error: Any error that occurred during listing
func (fs *FileStore) List(pattern string) ([]string, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	fullPattern := filepath.Join(fs.dataDir, pattern)

	matches, err := filepath.Glob(fullPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	// Convert absolute paths to relative paths
	relPaths := make([]string, 0, len(matches))
	for _, match := range matches {
		relPath, err := filepath.Rel(fs.dataDir, match)
		if err != nil {
			continue
		}
		relPaths = append(relPaths, relPath)
	}

	return relPaths, nil
}

// GetDataDir returns the data directory path.
func (fs *FileStore) GetDataDir() string {
	return fs.dataDir
}
