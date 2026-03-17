//go:build !windows

package persistence

import "syscall"

// lockFile acquires an exclusive blocking lock on the file descriptor using flock.
func lockFile(fd uintptr) error {
	return syscall.Flock(int(fd), syscall.LOCK_EX)
}

// tryLockFile attempts to acquire an exclusive non-blocking lock on the file descriptor.
// Returns (true, nil) if the lock was acquired, (false, nil) if it is held by another process.
func tryLockFile(fd uintptr) (bool, error) {
	err := syscall.Flock(int(fd), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		if err == syscall.EWOULDBLOCK {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// unlockFile releases the lock on the file descriptor using flock.
func unlockFile(fd uintptr) error {
	return syscall.Flock(int(fd), syscall.LOCK_UN)
}
