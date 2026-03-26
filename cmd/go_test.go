package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSelectRepo_NoRepos(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	devDir := t.TempDir()

	// No .repos file
	_, err := selectRepo(projectDir, devDir)
	if err == nil {
		t.Fatal("expected error for project with no repos")
	}
}

func TestSelectRepo_EmptyReposFile(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	devDir := t.TempDir()

	// Empty .repos file
	if err := os.WriteFile(filepath.Join(projectDir, ".repos"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := selectRepo(projectDir, devDir)
	if err == nil {
		t.Fatal("expected error for project with empty .repos")
	}
}

func TestSelectRepo_SingleRepo(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Use a temp dir as the dev directory
	devDir := t.TempDir()

	// Create the repo directory that .repos will point to
	repoDir := filepath.Join(devDir, "my-repo")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatal(err)
	}

	// .repos with single entry
	if err := os.WriteFile(filepath.Join(projectDir, ".repos"), []byte("https://github.com/user/my-repo.git\n"), 0644); err != nil {
		t.Fatal(err)
	}

	selected, err := selectRepo(projectDir, devDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if selected != repoDir {
		t.Fatalf("expected %s, got %s", repoDir, selected)
	}
}

func TestSelectRepo_SingleRepo_NotCloned(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	devDir := t.TempDir()

	// .repos pointing to non-existent directory
	if err := os.WriteFile(filepath.Join(projectDir, ".repos"), []byte("https://github.com/user/nonexistent-repo-12345.git\n"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := selectRepo(projectDir, devDir)
	if err == nil {
		t.Fatal("expected error for uncloned repo")
	}
}

func TestSelectProject_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := selectProject(tmpDir)
	// Should fail: either no fzf or no projects
	if err == nil {
		t.Fatal("expected error for empty directory")
	}
}
