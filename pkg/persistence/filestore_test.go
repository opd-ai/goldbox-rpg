package persistence

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAtomicWriteFile(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "atomic-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	t.Run("writes file successfully", func(t *testing.T) {
		filename := filepath.Join(tmpDir, "test.txt")
		data := []byte("test data")

		err := AtomicWriteFile(filename, data, 0644)
		assert.NoError(t, err)

		// Verify file exists and has correct content
		content, err := os.ReadFile(filename)
		assert.NoError(t, err)
		assert.Equal(t, data, content)

		// Verify permissions
		info, err := os.Stat(filename)
		assert.NoError(t, err)
		assert.Equal(t, os.FileMode(0644), info.Mode().Perm())
	})

	t.Run("creates parent directory if missing", func(t *testing.T) {
		filename := filepath.Join(tmpDir, "subdir", "test.txt")
		data := []byte("test data")

		err := AtomicWriteFile(filename, data, 0644)
		assert.NoError(t, err)

		// Verify file exists
		_, err = os.Stat(filename)
		assert.NoError(t, err)
	})

	t.Run("overwrites existing file", func(t *testing.T) {
		filename := filepath.Join(tmpDir, "overwrite.txt")
		
		// Write initial data
		err := AtomicWriteFile(filename, []byte("original"), 0644)
		assert.NoError(t, err)

		// Overwrite with new data
		err = AtomicWriteFile(filename, []byte("updated"), 0644)
		assert.NoError(t, err)

		// Verify new content
		content, err := os.ReadFile(filename)
		assert.NoError(t, err)
		assert.Equal(t, []byte("updated"), content)
	})

	t.Run("handles large files", func(t *testing.T) {
		filename := filepath.Join(tmpDir, "large.txt")
		data := make([]byte, 1024*1024) // 1MB
		for i := range data {
			data[i] = byte(i % 256)
		}

		err := AtomicWriteFile(filename, data, 0644)
		assert.NoError(t, err)

		// Verify file size
		info, err := os.Stat(filename)
		assert.NoError(t, err)
		assert.Equal(t, int64(len(data)), info.Size())
	})
}

func TestFileLock(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lock-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	t.Run("creates and acquires lock", func(t *testing.T) {
		lockPath := filepath.Join(tmpDir, "test.lock")

		lock, err := NewFileLock(lockPath)
		require.NoError(t, err)
		defer lock.Close()

		err = lock.Lock()
		assert.NoError(t, err)
		assert.True(t, lock.isLocked)

		err = lock.Unlock()
		assert.NoError(t, err)
		assert.False(t, lock.isLocked)
	})

	t.Run("try lock succeeds when unlocked", func(t *testing.T) {
		lockPath := filepath.Join(tmpDir, "trylock.lock")

		lock, err := NewFileLock(lockPath)
		require.NoError(t, err)
		defer lock.Close()

		acquired, err := lock.TryLock()
		assert.NoError(t, err)
		assert.True(t, acquired)
		assert.True(t, lock.isLocked)
	})

	t.Run("prevents double locking", func(t *testing.T) {
		lockPath := filepath.Join(tmpDir, "double.lock")

		lock, err := NewFileLock(lockPath)
		require.NoError(t, err)
		defer lock.Close()

		err = lock.Lock()
		assert.NoError(t, err)

		err = lock.Lock()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already held")
	})

	t.Run("close releases lock", func(t *testing.T) {
		lockPath := filepath.Join(tmpDir, "close.lock")

		lock, err := NewFileLock(lockPath)
		require.NoError(t, err)

		err = lock.Lock()
		assert.NoError(t, err)

		err = lock.Close()
		assert.NoError(t, err)
		assert.False(t, lock.isLocked)
	})

	t.Run("unlock is idempotent", func(t *testing.T) {
		lockPath := filepath.Join(tmpDir, "idempotent.lock")

		lock, err := NewFileLock(lockPath)
		require.NoError(t, err)
		defer lock.Close()

		err = lock.Lock()
		assert.NoError(t, err)

		err = lock.Unlock()
		assert.NoError(t, err)

		err = lock.Unlock()
		assert.NoError(t, err) // Should not error on double unlock
	})
}

func TestFileStore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filestore-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	type TestData struct {
		Name  string `yaml:"name"`
		Value int    `yaml:"value"`
	}

	t.Run("creates file store", func(t *testing.T) {
		fs, err := NewFileStore(tmpDir)
		assert.NoError(t, err)
		assert.NotNil(t, fs)
		assert.Equal(t, tmpDir, fs.GetDataDir())
	})

	t.Run("saves and loads data", func(t *testing.T) {
		fs, err := NewFileStore(tmpDir)
		require.NoError(t, err)

		original := TestData{Name: "test", Value: 42}
		
		err = fs.Save("test.yaml", &original)
		assert.NoError(t, err)

		var loaded TestData
		err = fs.Load("test.yaml", &loaded)
		assert.NoError(t, err)
		assert.Equal(t, original.Name, loaded.Name)
		assert.Equal(t, original.Value, loaded.Value)
	})

	t.Run("checks file existence", func(t *testing.T) {
		fs, err := NewFileStore(tmpDir)
		require.NoError(t, err)

		data := TestData{Name: "exists", Value: 1}
		err = fs.Save("exists.yaml", &data)
		assert.NoError(t, err)

		assert.True(t, fs.Exists("exists.yaml"))
		assert.False(t, fs.Exists("nonexistent.yaml"))
	})

	t.Run("deletes files", func(t *testing.T) {
		fs, err := NewFileStore(tmpDir)
		require.NoError(t, err)

		data := TestData{Name: "delete", Value: 1}
		err = fs.Save("delete.yaml", &data)
		assert.NoError(t, err)

		assert.True(t, fs.Exists("delete.yaml"))

		err = fs.Delete("delete.yaml")
		assert.NoError(t, err)

		assert.False(t, fs.Exists("delete.yaml"))
	})

	t.Run("lists files with pattern", func(t *testing.T) {
		fs, err := NewFileStore(tmpDir)
		require.NoError(t, err)

		// Create multiple files
		for i := 0; i < 3; i++ {
			data := TestData{Name: "list", Value: i}
			err = fs.Save(filepath.Join("list", string(rune('a'+i))+".yaml"), &data)
			assert.NoError(t, err)
		}

		files, err := fs.List("list/*.yaml")
		assert.NoError(t, err)
		assert.Len(t, files, 3)
	})

	t.Run("handles nested directories", func(t *testing.T) {
		fs, err := NewFileStore(tmpDir)
		require.NoError(t, err)

		data := TestData{Name: "nested", Value: 123}
		err = fs.Save("deep/nested/path/test.yaml", &data)
		assert.NoError(t, err)

		var loaded TestData
		err = fs.Load("deep/nested/path/test.yaml", &loaded)
		assert.NoError(t, err)
		assert.Equal(t, data.Value, loaded.Value)
	})

	t.Run("returns error for missing file", func(t *testing.T) {
		fs, err := NewFileStore(tmpDir)
		require.NoError(t, err)

		var data TestData
		err = fs.Load("nonexistent.yaml", &data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not exist")
	})

	t.Run("handles invalid YAML", func(t *testing.T) {
		fs, err := NewFileStore(tmpDir)
		require.NoError(t, err)

		// Write invalid YAML directly
		invalidPath := filepath.Join(tmpDir, "invalid.yaml")
		err = os.WriteFile(invalidPath, []byte("{ invalid yaml ["), 0644)
		require.NoError(t, err)

		var data TestData
		err = fs.Load("invalid.yaml", &data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unmarshal")
	})
}
