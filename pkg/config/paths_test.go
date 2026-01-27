package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sandermoonemans/local-brain/pkg/testutil"
)

func TestGetDumpPath(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	cfg, _ := Load()

	dumpPath, err := GetDumpPath(cfg)
	if err != nil {
		t.Fatalf("GetDumpPath failed: %v", err)
	}

	expected := filepath.Join(tb.BrainPath, "00_dump.md")
	if dumpPath != expected {
		t.Errorf("Expected %q, got %q", expected, dumpPath)
	}
}

func TestGetProjectsPath(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	cfg, _ := Load()

	projectsPath, err := GetProjectsPath(cfg)
	if err != nil {
		t.Fatalf("GetProjectsPath failed: %v", err)
	}

	expected := filepath.Join(tb.BrainPath, "01_active")
	if projectsPath != expected {
		t.Errorf("Expected %q, got %q", expected, projectsPath)
	}
}

func TestGetProjectPath(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	cfg, _ := Load()

	projectPath, err := GetProjectPath(cfg, "my-project")
	if err != nil {
		t.Fatalf("GetProjectPath failed: %v", err)
	}

	expected := filepath.Join(tb.BrainPath, "01_active", "my-project")
	if projectPath != expected {
		t.Errorf("Expected %q, got %q", expected, projectPath)
	}
}

func TestGetArchivePath(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	cfg, _ := Load()

	archivePath, err := GetArchivePath(cfg)
	if err != nil {
		t.Fatalf("GetArchivePath failed: %v", err)
	}

	expected := filepath.Join(tb.BrainPath, "02_archive")
	if archivePath != expected {
		t.Errorf("Expected %q, got %q", expected, archivePath)
	}
}

func TestGetLinkedRepos_Config(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	cfg, _ := Load()

	// Create a project with .repos file
	tb.AddProject("test-project")
	reposFile := filepath.Join(tb.ActiveDirPath, "test-project", ".repos")

	reposContent := `https://github.com/user/repo1.git
https://github.com/user/repo2.git
`
	tb.WriteFile(reposFile, reposContent)

	repos, err := GetLinkedRepos(cfg, "test-project")
	if err != nil {
		t.Fatalf("GetLinkedRepos failed: %v", err)
	}

	// Repos are expanded to ~/dev/reponame
	devDir := filepath.Join(os.Getenv("HOME"), "dev")
	if len(repos) != 2 {
		t.Errorf("Expected 2 repos, got %d", len(repos))
	}

	expected := []string{
		filepath.Join(devDir, "repo1"),
		filepath.Join(devDir, "repo2"),
	}

	for i, repo := range repos {
		if repo != expected[i] {
			t.Errorf("Repo %d: expected %q, got %q", i, expected[i], repo)
		}
	}
}

func TestExtractRepoName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://github.com/user/myrepo.git", "myrepo"},
		{"https://github.com/user/myrepo", "myrepo"},
		{"git@github.com:user/myrepo.git", "myrepo"},
		{"https://gitlab.com/org/sub/project.git", "project"},
		{"/local/path/to/repo", "repo"},
		{"invalid", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractRepoName(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
