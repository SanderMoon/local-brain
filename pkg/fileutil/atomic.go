package fileutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// AtomicWrite writes data to a file atomically using the temp+rename pattern
// This prevents corruption during concurrent writes or crashes
func AtomicWrite(filePath string, data []byte, perm os.FileMode) error {
	// Get the directory of the target file
	dir := filepath.Dir(filePath)

	// Create a temp file in the same directory (atomic rename requires same filesystem)
	tempFile, err := os.CreateTemp(dir, ".brain-tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	tempPath := tempFile.Name()

	// Cleanup temp file on error
	cleanup := func() {
		tempFile.Close()
		os.Remove(tempPath)
	}

	// Write data to temp file
	if _, err := tempFile.Write(data); err != nil {
		cleanup()
		return fmt.Errorf("failed to write to temp file: %w", err)
	}

	// Sync to ensure data is written to disk
	if err := tempFile.Sync(); err != nil {
		cleanup()
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	// Close temp file
	if err := tempFile.Close(); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Set permissions on temp file
	if err := os.Chmod(tempPath, perm); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Atomic rename (overwrites target file if it exists)
	if err := os.Rename(tempPath, filePath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// AtomicWriteFile is a convenience wrapper around AtomicWrite
func AtomicWriteFile(filePath string, data []byte) error {
	return AtomicWrite(filePath, data, 0644)
}

// AtomicCopy copies a file atomically
func AtomicCopy(src, dst string, perm os.FileMode) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Read all data
	data, err := io.ReadAll(srcFile)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Write atomically
	return AtomicWrite(dst, data, perm)
}
