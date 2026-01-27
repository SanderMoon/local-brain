package fileutil

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// ExpandPath expands ~ to the user's home directory
func ExpandPath(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	if path == "~" {
		return usr.HomeDir, nil
	}

	return filepath.Join(usr.HomeDir, path[2:]), nil
}

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsSymlink checks if a path is a symbolic link
func IsSymlink(path string) (bool, error) {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.Mode()&os.ModeSymlink != 0, nil
}

// IsDirectory checks if a path is a directory
func IsDirectory(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}

// RemoveSymlink removes a symbolic link or returns error if it's not a symlink
func RemoveSymlink(path string) error {
	// Check if path exists first
	if !FileExists(path) {
		return nil // Nothing to remove
	}

	isLink, err := IsSymlink(path)
	if err != nil {
		return err
	}

	if !isLink {
		// Check if it's a directory
		isDir, err := IsDirectory(path)
		if err != nil {
			return err
		}
		if isDir {
			// Don't remove directories
			return nil
		}
		// It's a regular file, warn but don't remove
		return os.ErrExist // Signal that it exists but is not a symlink
	}

	return os.Remove(path)
}

// CreateSymlink creates a symbolic link, removing old one if it exists
func CreateSymlink(oldname, newname string) error {
	// Try to remove old symlink
	err := RemoveSymlink(newname)
	if err != nil && err != os.ErrExist {
		return err
	}

	if err == os.ErrExist {
		// Path exists but is not a symlink
		return os.ErrExist
	}

	return os.Symlink(oldname, newname)
}
