package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sandermoonemans/local-brain/pkg/fileutil"
	"github.com/sandermoonemans/local-brain/pkg/testutil"
)

func TestUpdateSymlink(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	cfg, _ := Load()

	// Add another brain
	brain2Path := filepath.Join(tb.TmpDir, "brain2")
	if err := os.MkdirAll(brain2Path, 0755); err != nil {
		t.Fatalf("Failed to create brain2: %v", err)
	}
	cfg.AddBrain("brain2", brain2Path)

	// Update symlink to brain2
	err := UpdateSymlink("brain2", cfg)
	if err != nil {
		t.Fatalf("UpdateSymlink failed: %v", err)
	}

	// Verify symlink points to brain2
	linkTarget, err := os.Readlink(tb.SymlinkPath)
	if err != nil {
		t.Fatalf("Failed to read symlink: %v", err)
	}

	if linkTarget != brain2Path {
		t.Errorf("Symlink points to %q, expected %q", linkTarget, brain2Path)
	}
}

func TestUpdateSymlink_NonExistentBrain(t *testing.T) {
	_ = testutil.SetupTestBrain(t)
	cfg, _ := Load()

	err := UpdateSymlink("nonexistent", cfg)
	if err == nil {
		t.Error("Expected error for non-existent brain")
	}
}

func TestUpdateSymlink_CreateNew(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	cfg, _ := Load()

	// Remove symlink if it exists
	_ = fileutil.RemoveSymlink(tb.SymlinkPath)

	// Update symlink should create it
	err := UpdateSymlink("test", cfg)
	if err != nil {
		t.Fatalf("UpdateSymlink failed: %v", err)
	}

	// Verify symlink exists and points to correct brain
	isLink, err := fileutil.IsSymlink(tb.SymlinkPath)
	if err != nil {
		t.Fatalf("Failed to check symlink: %v", err)
	}

	if !isLink {
		t.Error("Symlink was not created")
	}

	linkTarget, err := os.Readlink(tb.SymlinkPath)
	if err != nil {
		t.Fatalf("Failed to read symlink: %v", err)
	}

	if linkTarget != tb.BrainPath {
		t.Errorf("Symlink points to %q, expected %q", linkTarget, tb.BrainPath)
	}
}

func TestGetCurrentSymlinkTarget(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	cfg, _ := Load()

	// Create symlink
	if err := UpdateSymlink("test", cfg); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	target, err := GetCurrentSymlinkTarget()
	if err != nil {
		t.Fatalf("GetCurrentSymlinkTarget failed: %v", err)
	}

	if target != tb.BrainPath {
		t.Errorf("Expected target %q, got %q", tb.BrainPath, target)
	}
}

func TestGetCurrentSymlinkTarget_NoSymlink(t *testing.T) {
	tmpDir := os.TempDir()
	fakeSymlink := filepath.Join(tmpDir, "fake-symlink-xyz")

	// Use a different symlink path
	oldSymlink := os.Getenv("BRAIN_SYMLINK")
	os.Setenv("BRAIN_SYMLINK", fakeSymlink)
	defer os.Setenv("BRAIN_SYMLINK", oldSymlink)

	// Clean up if exists
	os.Remove(fakeSymlink)

	_, err := GetCurrentSymlinkTarget()
	if err == nil {
		t.Error("Expected error when symlink doesn't exist")
	}
}
