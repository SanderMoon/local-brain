package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FileLock represents a directory-based file lock
type FileLock struct {
	lockDir    string
	maxRetries int
	retryDelay time.Duration
}

// NewFileLock creates a new file lock for the given file path
// The lock directory is created in the same directory as the file
func NewFileLock(filePath string) *FileLock {
	dir := filepath.Dir(filePath)
	base := filepath.Base(filePath)
	lockDir := filepath.Join(dir, "."+base+".lock")

	return &FileLock{
		lockDir:    lockDir,
		maxRetries: 5,
		retryDelay: 1 * time.Second,
	}
}

// NewLockWithRetries creates a lock with custom retry settings
func NewLockWithRetries(filePath string, maxRetries int, retryDelay time.Duration) *FileLock {
	lock := NewFileLock(filePath)
	lock.maxRetries = maxRetries
	lock.retryDelay = retryDelay
	return lock
}

// Acquire attempts to acquire the lock
// Returns error if lock cannot be acquired after max retries
func (l *FileLock) Acquire() error {
	for i := 0; i < l.maxRetries; i++ {
		// Try to create lock directory (atomic operation)
		err := os.Mkdir(l.lockDir, 0755)
		if err == nil {
			// Successfully acquired lock
			return nil
		}

		// Lock exists, check if it's a stale lock
		if os.IsExist(err) {
			// Wait and retry
			if i < l.maxRetries-1 {
				time.Sleep(l.retryDelay)
				continue
			}
		}

		// Other error or max retries exceeded
		return fmt.Errorf("failed to acquire lock after %d retries: %w", l.maxRetries, err)
	}

	return fmt.Errorf("could not acquire lock (is another process writing?)")
}

// Release releases the lock by removing the lock directory
func (l *FileLock) Release() error {
	err := os.Remove(l.lockDir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to release lock: %w", err)
	}
	return nil
}

// WithLock executes a function while holding the lock
// Automatically acquires and releases the lock
func WithLock(filePath string, fn func() error) error {
	lock := NewFileLock(filePath)

	if err := lock.Acquire(); err != nil {
		return err
	}

	defer func() {
		_ = lock.Release() // Ignore release errors in defer
	}()

	return fn()
}
