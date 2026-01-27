package config

import (
	"fmt"
	"os"

	"github.com/sandermoonemans/local-brain/pkg/fileutil"
)

// UpdateSymlink updates the ~/brain symlink to point to the specified brain
func UpdateSymlink(brainName string, cfg *Config) error {
	// Get brain path
	brainPath, err := cfg.GetBrainPath(brainName)
	if err != nil {
		return err
	}

	symlinkPath := GetSymlinkPath()

	// Check if symlink exists and handle it
	isLink, err := fileutil.IsSymlink(symlinkPath)
	if err != nil {
		return fmt.Errorf("failed to check symlink: %w", err)
	}

	if isLink {
		// Remove existing symlink
		if err := os.Remove(symlinkPath); err != nil {
			return fmt.Errorf("failed to remove old symlink: %w", err)
		}
	} else if fileutil.FileExists(symlinkPath) {
		// Path exists but is not a symlink
		// Check if it's a directory
		isDir, err := fileutil.IsDirectory(symlinkPath)
		if err != nil {
			return fmt.Errorf("failed to check if path is directory: %w", err)
		}

		if !isDir {
			// It's a regular file, warn and don't replace
			return fmt.Errorf("warning: %s exists and is not a symlink", symlinkPath)
		}
		// If it's a directory, don't touch it (user might have a real ~/brain directory)
	}

	// Create new symlink
	if err := os.Symlink(brainPath, symlinkPath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}

// GetCurrentSymlinkTarget returns the target of the ~/brain symlink
func GetCurrentSymlinkTarget() (string, error) {
	symlinkPath := GetSymlinkPath()

	isLink, err := fileutil.IsSymlink(symlinkPath)
	if err != nil {
		return "", err
	}

	if !isLink {
		return "", fmt.Errorf("path is not a symlink")
	}

	target, err := os.Readlink(symlinkPath)
	if err != nil {
		return "", fmt.Errorf("failed to read symlink: %w", err)
	}

	return target, nil
}
