//go:build windows

package persistence

import (
	"syscall"
	"unsafe"
)

const (
	lockfileExclusiveLock   = 0x00000002
	lockfileFailImmediately = 0x00000001
)

var (
	modkernel32      = syscall.NewLazyDLL("kernel32.dll")
	procLockFileEx   = modkernel32.NewProc("LockFileEx")
	procUnlockFileEx = modkernel32.NewProc("UnlockFileEx")
)

// lockFile acquires an exclusive blocking lock on the file handle using LockFileEx.
func lockFile(fd uintptr) error {
	var ol syscall.Overlapped
	r1, _, err := procLockFileEx.Call(
		fd,
		lockfileExclusiveLock,
		0,
		1,
		0,
		uintptr(unsafe.Pointer(&ol)),
	)
	if r1 == 0 {
		return err
	}
	return nil
}

// tryLockFile attempts to acquire an exclusive non-blocking lock on the file handle.
// Returns (true, nil) if the lock was acquired, (false, nil) if it is held by another process.
func tryLockFile(fd uintptr) (bool, error) {
	var ol syscall.Overlapped
	r1, _, err := procLockFileEx.Call(
		fd,
		lockfileExclusiveLock|lockfileFailImmediately,
		0,
		1,
		0,
		uintptr(unsafe.Pointer(&ol)),
	)
	if r1 == 0 {
		// ERROR_LOCK_VIOLATION = 33
		if err == syscall.Errno(33) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// unlockFile releases the lock on the file handle using UnlockFileEx.
func unlockFile(fd uintptr) error {
	var ol syscall.Overlapped
	r1, _, err := procUnlockFileEx.Call(
		fd,
		0,
		1,
		0,
		uintptr(unsafe.Pointer(&ol)),
	)
	if r1 == 0 {
		return err
	}
	return nil
}
