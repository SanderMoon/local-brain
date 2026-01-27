package fileutil

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestAtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	data := []byte("hello world")
	err := AtomicWrite(testFile, data, 0644)
	if err != nil {
		t.Fatalf("AtomicWrite failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("File was not created")
	}

	// Verify contents
	got, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(got) != string(data) {
		t.Errorf("Expected %q, got %q", data, got)
	}

	// Verify permissions
	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	expectedPerm := os.FileMode(0644)
	if info.Mode().Perm() != expectedPerm {
		t.Errorf("Expected permissions %v, got %v", expectedPerm, info.Mode().Perm())
	}
}

func TestAtomicWrite_Overwrite(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Write initial content
	initial := []byte("initial content")
	if err := os.WriteFile(testFile, initial, 0644); err != nil {
		t.Fatalf("Failed to create initial file: %v", err)
	}

	// Overwrite with AtomicWrite
	updated := []byte("updated content")
	err := AtomicWrite(testFile, updated, 0644)
	if err != nil {
		t.Fatalf("AtomicWrite failed: %v", err)
	}

	// Verify contents were updated
	got, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(got) != string(updated) {
		t.Errorf("Expected %q, got %q", updated, got)
	}
}

func TestAtomicWrite_NonExistentDir(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "nonexistent", "test.txt")

	data := []byte("test")
	err := AtomicWrite(testFile, data, 0644)
	if err == nil {
		t.Error("Expected error when writing to non-existent directory")
	}
}

func TestAtomicWriteFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	data := []byte("hello world")
	err := AtomicWriteFile(testFile, data)
	if err != nil {
		t.Fatalf("AtomicWriteFile failed: %v", err)
	}

	// Verify file exists and has correct content
	got, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(got) != string(data) {
		t.Errorf("Expected %q, got %q", data, got)
	}
}

func TestAtomicCopy(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "src.txt")
	dstFile := filepath.Join(tmpDir, "dst.txt")

	// Create source file
	srcData := []byte("source content")
	if err := os.WriteFile(srcFile, srcData, 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Copy atomically
	err := AtomicCopy(srcFile, dstFile, 0644)
	if err != nil {
		t.Fatalf("AtomicCopy failed: %v", err)
	}

	// Verify destination file
	got, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(got) != string(srcData) {
		t.Errorf("Expected %q, got %q", srcData, got)
	}
}

func TestAtomicCopy_NonExistentSource(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "nonexistent.txt")
	dstFile := filepath.Join(tmpDir, "dst.txt")

	err := AtomicCopy(srcFile, dstFile, 0644)
	if err == nil {
		t.Error("Expected error when copying non-existent file")
	}
}

func TestAtomicWrite_ConcurrentWrites(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "concurrent.txt")

	numGoroutines := 10
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()

			data := []byte("goroutine " + string(rune('0'+n)))
			if err := AtomicWrite(testFile, data, 0644); err != nil {
				t.Errorf("AtomicWrite failed in goroutine %d: %v", n, err)
			}
		}(i)
	}

	wg.Wait()

	// Verify file exists and is not corrupted
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file after concurrent writes: %v", err)
	}

	// Should contain data from one of the goroutines
	if len(data) == 0 {
		t.Error("File is empty after concurrent writes")
	}
}

func TestAtomicWrite_NoTempFileLeftBehind(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	data := []byte("test data")
	err := AtomicWrite(testFile, data, 0644)
	if err != nil {
		t.Fatalf("AtomicWrite failed: %v", err)
	}

	// Check for any temp files
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".tmp" ||
		   filepath.Base(entry.Name())[:5] == ".brain-tmp-" {
			t.Errorf("Temp file left behind: %s", entry.Name())
		}
	}
}

func TestAtomicWrite_CustomPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	data := []byte("test")
	customPerm := os.FileMode(0600)

	err := AtomicWrite(testFile, data, customPerm)
	if err != nil {
		t.Fatalf("AtomicWrite failed: %v", err)
	}

	// Verify permissions
	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if info.Mode().Perm() != customPerm {
		t.Errorf("Expected permissions %v, got %v", customPerm, info.Mode().Perm())
	}
}

func TestAtomicWrite_EmptyData(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty.txt")

	err := AtomicWrite(testFile, []byte{}, 0644)
	if err != nil {
		t.Fatalf("AtomicWrite failed on empty data: %v", err)
	}

	// Verify file exists and is empty
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if len(data) != 0 {
		t.Errorf("Expected empty file, got %d bytes", len(data))
	}
}

func TestAtomicWrite_LargeData(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "large.txt")

	// Create 1MB of data
	data := make([]byte, 1024*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	err := AtomicWrite(testFile, data, 0644)
	if err != nil {
		t.Fatalf("AtomicWrite failed on large data: %v", err)
	}

	// Verify size
	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if info.Size() != int64(len(data)) {
		t.Errorf("Expected size %d, got %d", len(data), info.Size())
	}
}
