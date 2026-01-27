package api

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sandermoonemans/local-brain/pkg/testutil"
)

func TestListProjects(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	// Create projects
	tb.AddProject("project-a")
	tb.AddProject("project-b")
	tb.AddProject("project-c")

	projects, err := ListProjects(tb.ActiveDirPath, "project-b")
	if err != nil {
		t.Fatalf("ListProjects failed: %v", err)
	}

	if len(projects) != 3 {
		t.Fatalf("Expected 3 projects, got %d", len(projects))
	}

	// Verify project-b is marked as focused
	for _, proj := range projects {
		if proj.Name == "project-b" {
			if !proj.Focused {
				t.Error("project-b should be focused")
			}
		} else {
			if proj.Focused {
				t.Errorf("%s should not be focused", proj.Name)
			}
		}
	}
}

func TestListProjects_WithTasks(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	tb.AddProject("with-tasks")
	todoFile := filepath.Join(tb.ActiveDirPath, "with-tasks", "todo.md")

	content := `# With Tasks

## Active

- [ ] Task 1
- [ ] Task 2
- [x] Completed task
`
	tb.WriteFile(todoFile, content)

	projects, err := ListProjects(tb.ActiveDirPath, "")
	if err != nil {
		t.Fatalf("ListProjects failed: %v", err)
	}

	if len(projects) != 1 {
		t.Fatalf("Expected 1 project, got %d", len(projects))
	}

	// Should count only open tasks
	if projects[0].TaskCount != 2 {
		t.Errorf("Expected 2 tasks, got %d", projects[0].TaskCount)
	}
}

func TestListProjects_WithRepos(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	tb.AddProject("with-repos")
	reposFile := filepath.Join(tb.ActiveDirPath, "with-repos", ".repos")

	reposContent := `/path/to/repo1
/path/to/repo2
# Comment line
/path/to/repo3

`
	tb.WriteFile(reposFile, reposContent)

	projects, err := ListProjects(tb.ActiveDirPath, "")
	if err != nil {
		t.Fatalf("ListProjects failed: %v", err)
	}

	if len(projects) != 1 {
		t.Fatalf("Expected 1 project, got %d", len(projects))
	}

	// Should count 3 repos (ignore comment and empty lines)
	if projects[0].RepoCount != 3 {
		t.Errorf("Expected 3 repos, got %d", projects[0].RepoCount)
	}
}

func TestListProjects_Empty(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	projects, err := ListProjects(tb.ActiveDirPath, "")
	if err != nil {
		t.Fatalf("ListProjects failed: %v", err)
	}

	if len(projects) != 0 {
		t.Errorf("Expected 0 projects, got %d", len(projects))
	}
}

func TestListProjects_SkipsHidden(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	tb.AddProject("visible")

	// Create hidden directory
	hiddenDir := filepath.Join(tb.ActiveDirPath, ".hidden")
	tb.WriteFile(filepath.Join(hiddenDir, "todo.md"), "# Hidden\n")

	projects, err := ListProjects(tb.ActiveDirPath, "")
	if err != nil {
		t.Fatalf("ListProjects failed: %v", err)
	}

	if len(projects) != 1 {
		t.Fatalf("Expected 1 project, got %d", len(projects))
	}

	if projects[0].Name != "visible" {
		t.Errorf("Expected 'visible', got '%s'", projects[0].Name)
	}
}

func TestExtractRepoName(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://github.com/user/repo.git", "repo"},
		{"https://github.com/user/repo", "repo"},
		{"git@github.com:user/repo.git", "repo"},
		{"git@github.com:user/repo", "repo"},
		{"https://gitlab.com/org/subgroup/project.git", "project"},
		{"/local/path/to/repo", "repo"},
		{"repo-name", "repo-name"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := ExtractRepoName(tt.url)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGetLinkedRepos(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	projectDir := filepath.Join(tb.ActiveDirPath, "test-project")
	reposFile := filepath.Join(projectDir, ".repos")

	reposContent := `https://github.com/user/repo1.git
https://github.com/user/repo2.git
# This is a comment
https://github.com/user/repo3.git

`
	tb.WriteFile(reposFile, reposContent)

	repos, err := GetLinkedRepos(projectDir)
	if err != nil {
		t.Fatalf("GetLinkedRepos failed: %v", err)
	}

	// GetLinkedRepos extracts repo names and creates paths in ~/dev/
	devDir := filepath.Join(os.Getenv("HOME"), "dev")
	expected := []string{
		filepath.Join(devDir, "repo1"),
		filepath.Join(devDir, "repo2"),
		filepath.Join(devDir, "repo3"),
	}

	if len(repos) != len(expected) {
		t.Fatalf("Expected %d repos, got %d", len(expected), len(repos))
	}

	for i, repo := range repos {
		if repo != expected[i] {
			t.Errorf("Repo %d: expected %q, got %q", i, expected[i], repo)
		}
	}
}

func TestGetLinkedRepos_NoFile(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	projectDir := filepath.Join(tb.ActiveDirPath, "no-repos")

	repos, err := GetLinkedRepos(projectDir)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(repos) != 0 {
		t.Errorf("Expected 0 repos, got %d", len(repos))
	}
}

func TestGetLinkedRepos_EmptyFile(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	projectDir := filepath.Join(tb.ActiveDirPath, "empty-repos")
	reposFile := filepath.Join(projectDir, ".repos")

	tb.WriteFile(reposFile, "\n\n\n")

	repos, err := GetLinkedRepos(projectDir)
	if err != nil {
		t.Fatalf("GetLinkedRepos failed: %v", err)
	}

	if len(repos) != 0 {
		t.Errorf("Expected 0 repos, got %d", len(repos))
	}
}
