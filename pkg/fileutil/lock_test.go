package fileutil

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestNewFileLock(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	lock := NewFileLock(testFile)

	expectedLockDir := filepath.Join(tmpDir, ".test.txt.lock")
	if lock.lockDir != expectedLockDir {
		t.Errorf("Expected lock dir %s, got %s", expectedLockDir, lock.lockDir)
	}

	if lock.maxRetries != 5 {
		t.Errorf("Expected maxRetries 5, got %d", lock.maxRetries)
	}

	if lock.retryDelay != 1*time.Second {
		t.Errorf("Expected retryDelay 1s, got %v", lock.retryDelay)
	}
}

func TestNewLockWithRetries(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	lock := NewLockWithRetries(testFile, 3, 500*time.Millisecond)

	if lock.maxRetries != 3 {
		t.Errorf("Expected maxRetries 3, got %d", lock.maxRetries)
	}

	if lock.retryDelay != 500*time.Millisecond {
		t.Errorf("Expected retryDelay 500ms, got %v", lock.retryDelay)
	}
}

func TestFileLock_AcquireAndRelease(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	lock := NewFileLock(testFile)

	// Acquire lock
	err := lock.Acquire()
	if err != nil {
		t.Fatalf("Failed to acquire lock: %v", err)
	}

	// Verify lock directory exists
	if _, err := os.Stat(lock.lockDir); os.IsNotExist(err) {
		t.Error("Lock directory was not created")
	}

	// Release lock
	err = lock.Release()
	if err != nil {
		t.Fatalf("Failed to release lock: %v", err)
	}

	// Verify lock directory is removed
	if _, err := os.Stat(lock.lockDir); !os.IsNotExist(err) {
		t.Error("Lock directory was not removed")
	}
}

func TestFileLock_AcquireTwice(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	lock1 := NewLockWithRetries(testFile, 2, 100*time.Millisecond)
	lock2 := NewLockWithRetries(testFile, 2, 100*time.Millisecond)

	// Acquire first lock
	err := lock1.Acquire()
	if err != nil {
		t.Fatalf("Failed to acquire first lock: %v", err)
	}
	defer lock1.Release()

	// Try to acquire second lock (should fail)
	err = lock2.Acquire()
	if err == nil {
		t.Error("Expected error when acquiring lock twice")
		lock2.Release()
	}
}

func TestFileLock_ReleaseUnacquired(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	lock := NewFileLock(testFile)

	// Release without acquiring should not error
	err := lock.Release()
	if err != nil {
		t.Errorf("Release of unacquired lock failed: %v", err)
	}
}

func TestFileLock_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	numGoroutines := 5
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	// Multiple goroutines trying to acquire lock
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()

			lock := NewLockWithRetries(testFile, 3, 50*time.Millisecond)
			err := lock.Acquire()
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()

				// Hold lock briefly
				time.Sleep(10 * time.Millisecond)

				lock.Release()
			}
		}(i)
	}

	wg.Wait()

	// At least one should have succeeded
	if successCount == 0 {
		t.Error("No goroutine acquired lock")
	}

	// Should be fewer successes than attempts (some should have failed due to contention)
	t.Logf("Success count: %d/%d", successCount, numGoroutines)
}

func TestWithLock(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	lockFile := filepath.Join(tmpDir, ".test.txt.lock")

	executed := false
	err := WithLock(testFile, func() error {
		executed = true

		// Lock should be held during function execution
		if _, err := os.Stat(lockFile); os.IsNotExist(err) {
			t.Error("Lock directory does not exist during function execution")
		}

		return nil
	})

	if err != nil {
		t.Fatalf("WithLock failed: %v", err)
	}

	if !executed {
		t.Error("Function was not executed")
	}

	// Lock should be released after function completes
	if _, err := os.Stat(lockFile); !os.IsNotExist(err) {
		t.Error("Lock directory was not removed after function completed")
	}
}

func TestWithLock_FunctionError(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	lockFile := filepath.Join(tmpDir, ".test.txt.lock")

	expectedErr := os.ErrNotExist
	err := WithLock(testFile, func() error {
		return expectedErr
	})

	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}

	// Lock should be released even on error
	if _, err := os.Stat(lockFile); !os.IsNotExist(err) {
		t.Error("Lock directory was not removed after function error")
	}
}

func TestWithLock_Nested(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Nested locks on same file should fail
	err := WithLock(testFile, func() error {
		// Try to acquire same lock again
		lock := NewLockWithRetries(testFile, 2, 50*time.Millisecond)
		err := lock.Acquire()
		if err == nil {
			t.Error("Expected error when acquiring nested lock")
			lock.Release()
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Outer WithLock failed: %v", err)
	}
}

func TestFileLock_Retry(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Acquire first lock
	lock1 := NewFileLock(testFile)
	if err := lock1.Acquire(); err != nil {
		t.Fatalf("Failed to acquire first lock: %v", err)
	}

	// Try to acquire second lock in background, release first after delay
	go func() {
		time.Sleep(200 * time.Millisecond)
		lock1.Release()
	}()

	// Second lock with retries should eventually succeed
	lock2 := NewLockWithRetries(testFile, 5, 100*time.Millisecond)
	start := time.Now()
	err := lock2.Acquire()
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to acquire lock after retry: %v", err)
	}
	defer lock2.Release()

	// Should have taken at least one retry period
	if elapsed < 100*time.Millisecond {
		t.Errorf("Lock acquired too quickly (%v), expected retry delay", elapsed)
	}

	t.Logf("Lock acquired after %v", elapsed)
}

func TestFileLock_Sequential(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Acquire and release multiple times
	for i := 0; i < 3; i++ {
		lock := NewFileLock(testFile)

		err := lock.Acquire()
		if err != nil {
			t.Fatalf("Iteration %d: Failed to acquire lock: %v", i, err)
		}

		err = lock.Release()
		if err != nil {
			t.Fatalf("Iteration %d: Failed to release lock: %v", i, err)
		}
	}
}

func TestFileLock_CrossGoroutineRelease(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	lock1 := NewFileLock(testFile)
	lock2 := NewFileLock(testFile)

	// Acquire in one "goroutine"
	if err := lock1.Acquire(); err != nil {
		t.Fatalf("Failed to acquire lock: %v", err)
	}

	// Release using different lock object (same lock directory)
	if err := lock2.Release(); err != nil {
		t.Fatalf("Failed to release lock from different object: %v", err)
	}

	// Should be able to acquire again
	if err := lock1.Acquire(); err != nil {
		t.Errorf("Failed to reacquire after release: %v", err)
	}
	lock1.Release()
}
