package fileutil

import (
	"os"
	"os/user"
	"path/filepath"
	"testing"
)

func TestExpandPath(t *testing.T) {
	usr, err := user.Current()
	if err != nil {
		t.Fatalf("Failed to get current user: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "tilde only",
			input:    "~",
			expected: usr.HomeDir,
		},
		{
			name:     "tilde with path",
			input:    "~/Documents",
			expected: filepath.Join(usr.HomeDir, "Documents"),
		},
		{
			name:     "tilde with nested path",
			input:    "~/path/to/file.txt",
			expected: filepath.Join(usr.HomeDir, "path/to/file.txt"),
		},
		{
			name:     "absolute path (no expansion)",
			input:    "/absolute/path",
			expected: "/absolute/path",
		},
		{
			name:     "relative path (no expansion)",
			input:    "relative/path",
			expected: "relative/path",
		},
		{
			name:     "empty path",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExpandPath(tt.input)
			if err != nil {
				t.Fatalf("ExpandPath failed: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestEnsureDir(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("create new directory", func(t *testing.T) {
		newDir := filepath.Join(tmpDir, "new-dir")

		err := EnsureDir(newDir)
		if err != nil {
			t.Fatalf("EnsureDir failed: %v", err)
		}

		// Verify directory exists
		info, err := os.Stat(newDir)
		if err != nil {
			t.Fatalf("Directory not created: %v", err)
		}

		if !info.IsDir() {
			t.Error("Path is not a directory")
		}
	})

	t.Run("create nested directories", func(t *testing.T) {
		nestedDir := filepath.Join(tmpDir, "a", "b", "c", "d")

		err := EnsureDir(nestedDir)
		if err != nil {
			t.Fatalf("EnsureDir failed: %v", err)
		}

		// Verify all directories exist
		info, err := os.Stat(nestedDir)
		if err != nil {
			t.Fatalf("Nested directories not created: %v", err)
		}

		if !info.IsDir() {
			t.Error("Path is not a directory")
		}
	})

	t.Run("directory already exists", func(t *testing.T) {
		existingDir := filepath.Join(tmpDir, "existing")

		// Create directory first
		if err := os.Mkdir(existingDir, 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		// EnsureDir should not error
		err := EnsureDir(existingDir)
		if err != nil {
			t.Fatalf("EnsureDir failed on existing directory: %v", err)
		}
	})
}

func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("file exists", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "test.txt")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		if !FileExists(testFile) {
			t.Error("FileExists returned false for existing file")
		}
	})

	t.Run("directory exists", func(t *testing.T) {
		testDir := filepath.Join(tmpDir, "testdir")
		if err := os.Mkdir(testDir, 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		if !FileExists(testDir) {
			t.Error("FileExists returned false for existing directory")
		}
	})

	t.Run("file does not exist", func(t *testing.T) {
		nonExistent := filepath.Join(tmpDir, "nonexistent.txt")

		if FileExists(nonExistent) {
			t.Error("FileExists returned true for non-existent file")
		}
	})
}

func TestIsSymlink(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("is symlink", func(t *testing.T) {
		targetFile := filepath.Join(tmpDir, "target.txt")
		linkFile := filepath.Join(tmpDir, "link.txt")

		// Create target file
		if err := os.WriteFile(targetFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create target file: %v", err)
		}

		// Create symlink
		if err := os.Symlink(targetFile, linkFile); err != nil {
			t.Fatalf("Failed to create symlink: %v", err)
		}

		isLink, err := IsSymlink(linkFile)
		if err != nil {
			t.Fatalf("IsSymlink failed: %v", err)
		}

		if !isLink {
			t.Error("IsSymlink returned false for symlink")
		}
	})

	t.Run("regular file", func(t *testing.T) {
		regularFile := filepath.Join(tmpDir, "regular.txt")
		if err := os.WriteFile(regularFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create regular file: %v", err)
		}

		isLink, err := IsSymlink(regularFile)
		if err != nil {
			t.Fatalf("IsSymlink failed: %v", err)
		}

		if isLink {
			t.Error("IsSymlink returned true for regular file")
		}
	})

	t.Run("directory", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "dir")
		if err := os.Mkdir(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		isLink, err := IsSymlink(dir)
		if err != nil {
			t.Fatalf("IsSymlink failed: %v", err)
		}

		if isLink {
			t.Error("IsSymlink returned true for directory")
		}
	})

	t.Run("non-existent path", func(t *testing.T) {
		nonExistent := filepath.Join(tmpDir, "nonexistent")

		isLink, err := IsSymlink(nonExistent)
		if err != nil {
			t.Fatalf("IsSymlink failed on non-existent path: %v", err)
		}

		if isLink {
			t.Error("IsSymlink returned true for non-existent path")
		}
	})
}

func TestIsDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("is directory", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "testdir")
		if err := os.Mkdir(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		isDir, err := IsDirectory(dir)
		if err != nil {
			t.Fatalf("IsDirectory failed: %v", err)
		}

		if !isDir {
			t.Error("IsDirectory returned false for directory")
		}
	})

	t.Run("regular file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "file.txt")
		if err := os.WriteFile(file, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}

		isDir, err := IsDirectory(file)
		if err != nil {
			t.Fatalf("IsDirectory failed: %v", err)
		}

		if isDir {
			t.Error("IsDirectory returned true for regular file")
		}
	})

	t.Run("non-existent path", func(t *testing.T) {
		nonExistent := filepath.Join(tmpDir, "nonexistent")

		isDir, err := IsDirectory(nonExistent)
		if err != nil {
			t.Fatalf("IsDirectory failed on non-existent path: %v", err)
		}

		if isDir {
			t.Error("IsDirectory returned true for non-existent path")
		}
	})

	t.Run("symlink to directory", func(t *testing.T) {
		targetDir := filepath.Join(tmpDir, "targetdir")
		linkDir := filepath.Join(tmpDir, "linkdir")

		if err := os.Mkdir(targetDir, 0755); err != nil {
			t.Fatalf("Failed to create target directory: %v", err)
		}

		if err := os.Symlink(targetDir, linkDir); err != nil {
			t.Fatalf("Failed to create symlink: %v", err)
		}

		// Should follow symlink and return true
		isDir, err := IsDirectory(linkDir)
		if err != nil {
			t.Fatalf("IsDirectory failed: %v", err)
		}

		if !isDir {
			t.Error("IsDirectory returned false for symlink to directory")
		}
	})
}

func TestRemoveSymlink(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("remove symlink", func(t *testing.T) {
		targetFile := filepath.Join(tmpDir, "target1.txt")
		linkFile := filepath.Join(tmpDir, "link1.txt")

		if err := os.WriteFile(targetFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create target file: %v", err)
		}

		if err := os.Symlink(targetFile, linkFile); err != nil {
			t.Fatalf("Failed to create symlink: %v", err)
		}

		// Remove symlink
		err := RemoveSymlink(linkFile)
		if err != nil {
			t.Fatalf("RemoveSymlink failed: %v", err)
		}

		// Verify symlink is removed
		if FileExists(linkFile) {
			t.Error("Symlink still exists after removal")
		}

		// Target should still exist
		if !FileExists(targetFile) {
			t.Error("Target file was removed")
		}
	})

	t.Run("regular file (should not remove)", func(t *testing.T) {
		regularFile := filepath.Join(tmpDir, "regular2.txt")
		if err := os.WriteFile(regularFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create regular file: %v", err)
		}

		// Should return error (os.ErrExist) and not remove
		err := RemoveSymlink(regularFile)
		if err != os.ErrExist {
			t.Errorf("Expected os.ErrExist, got %v", err)
		}

		// File should still exist
		if !FileExists(regularFile) {
			t.Error("Regular file was removed")
		}
	})

	t.Run("directory (should not remove)", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "dir2")
		if err := os.Mkdir(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		// Should not remove directory
		err := RemoveSymlink(dir)
		if err != nil {
			t.Errorf("RemoveSymlink returned error for directory: %v", err)
		}

		// Directory should still exist
		if !FileExists(dir) {
			t.Error("Directory was removed")
		}
	})

	t.Run("non-existent path", func(t *testing.T) {
		nonExistent := filepath.Join(tmpDir, "nonexistent2")

		// Should not error on non-existent path (nothing to remove)
		err := RemoveSymlink(nonExistent)
		if err != nil {
			t.Errorf("RemoveSymlink failed on non-existent path: %v", err)
		}
	})
}

func TestCreateSymlink(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("create new symlink", func(t *testing.T) {
		targetFile := filepath.Join(tmpDir, "target1.txt")
		linkFile := filepath.Join(tmpDir, "link1.txt")

		if err := os.WriteFile(targetFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create target file: %v", err)
		}

		err := CreateSymlink(targetFile, linkFile)
		if err != nil {
			t.Fatalf("CreateSymlink failed: %v", err)
		}

		// Verify symlink exists
		isLink, err := IsSymlink(linkFile)
		if err != nil {
			t.Fatalf("Failed to check symlink: %v", err)
		}

		if !isLink {
			t.Error("Link was not created")
		}
	})

	t.Run("replace existing symlink", func(t *testing.T) {
		target1 := filepath.Join(tmpDir, "target2.txt")
		target2 := filepath.Join(tmpDir, "target3.txt")
		linkFile := filepath.Join(tmpDir, "link2.txt")

		if err := os.WriteFile(target1, []byte("test1"), 0644); err != nil {
			t.Fatalf("Failed to create target1: %v", err)
		}
		if err := os.WriteFile(target2, []byte("test2"), 0644); err != nil {
			t.Fatalf("Failed to create target2: %v", err)
		}

		// Create initial symlink
		if err := os.Symlink(target1, linkFile); err != nil {
			t.Fatalf("Failed to create initial symlink: %v", err)
		}

		// Replace with new symlink
		err := CreateSymlink(target2, linkFile)
		if err != nil {
			t.Fatalf("CreateSymlink failed on replace: %v", err)
		}

		// Verify it points to target2
		linkTarget, err := os.Readlink(linkFile)
		if err != nil {
			t.Fatalf("Failed to read link: %v", err)
		}

		if linkTarget != target2 {
			t.Errorf("Link points to %q, expected %q", linkTarget, target2)
		}
	})

	t.Run("fail on existing file", func(t *testing.T) {
		targetFile := filepath.Join(tmpDir, "target4.txt")
		existingFile := filepath.Join(tmpDir, "existing.txt")

		if err := os.WriteFile(targetFile, []byte("target"), 0644); err != nil {
			t.Fatalf("Failed to create target: %v", err)
		}
		if err := os.WriteFile(existingFile, []byte("existing"), 0644); err != nil {
			t.Fatalf("Failed to create existing file: %v", err)
		}

		// Should fail because existing file is not a symlink
		err := CreateSymlink(targetFile, existingFile)
		if err != os.ErrExist {
			t.Errorf("Expected os.ErrExist, got %v", err)
		}

		// Existing file should not be modified
		data, err := os.ReadFile(existingFile)
		if err != nil {
			t.Fatalf("Failed to read existing file: %v", err)
		}
		if string(data) != "existing" {
			t.Error("Existing file was modified")
		}
	})
}
