//go:build windows

package persistence

import (
	"syscall"
	"unsafe"
)

const (
	lockfileExclusiveLock   = 0x00000002
	lockfileFailImmediately = 0x00000001

	// errLockViolation corresponds to Windows ERROR_LOCK_VIOLATION (33).
	errLockViolation = syscall.Errno(33)
)

var (
	modkernel32      = syscall.NewLazyDLL("kernel32.dll")
	procLockFileEx   = modkernel32.NewProc("LockFileEx")
	procUnlockFileEx = modkernel32.NewProc("UnlockFileEx")
)

// lockFile acquires an exclusive blocking lock on the file handle using LockFileEx.
func lockFile(fd uintptr) error {
	var ol syscall.Overlapped
	// LockFileEx(hFile, dwFlags, dwReserved, nNumberOfBytesToLockLow, nNumberOfBytesToLockHigh, lpOverlapped)
	r1, _, err := procLockFileEx.Call(
		fd,                           // hFile
		lockfileExclusiveLock,        // dwFlags - exclusive lock
		0,                            // dwReserved - must be zero
		1,                            // nNumberOfBytesToLockLow - lock 1 byte
		0,                            // nNumberOfBytesToLockHigh
		uintptr(unsafe.Pointer(&ol)), // lpOverlapped
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
	// LockFileEx(hFile, dwFlags, dwReserved, nNumberOfBytesToLockLow, nNumberOfBytesToLockHigh, lpOverlapped)
	r1, _, err := procLockFileEx.Call(
		fd, // hFile
		lockfileExclusiveLock|lockfileFailImmediately, // dwFlags - exclusive + non-blocking
		0,                            // dwReserved - must be zero
		1,                            // nNumberOfBytesToLockLow - lock 1 byte
		0,                            // nNumberOfBytesToLockHigh
		uintptr(unsafe.Pointer(&ol)), // lpOverlapped
	)
	if r1 == 0 {
		if err == errLockViolation {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// unlockFile releases the lock on the file handle using UnlockFileEx.
func unlockFile(fd uintptr) error {
	var ol syscall.Overlapped
	// UnlockFileEx(hFile, dwReserved, nNumberOfBytesToUnlockLow, nNumberOfBytesToUnlockHigh, lpOverlapped)
	r1, _, err := procUnlockFileEx.Call(
		fd,                           // hFile
		0,                            // dwReserved - must be zero
		1,                            // nNumberOfBytesToUnlockLow - unlock 1 byte
		0,                            // nNumberOfBytesToUnlockHigh
		uintptr(unsafe.Pointer(&ol)), // lpOverlapped
	)
	if r1 == 0 {
		return err
	}
	return nil
}
