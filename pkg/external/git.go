package external

import (
	"fmt"
	"os/exec"
	"strings"
)

// VerifyRemote checks if a git remote URL is valid
// Equivalent to: git ls-remote <url> HEAD
func VerifyRemote(url string) error {
	cmd := exec.Command("git", "ls-remote", url, "HEAD")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("invalid git remote: %w: %s", err, string(output))
	}

	return nil
}

// Clone clones a git repository
// Equivalent to: git clone <url> <dest>
func Clone(url, dest string) error {
	cmd := exec.Command("git", "clone", url, dest)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %w: %s", err, string(output))
	}

	return nil
}

// Pull pulls changes from the remote repository
// Equivalent to: git -C <repoPath> pull
func Pull(repoPath string) error {
	cmd := exec.Command("git", "-C", repoPath, "pull")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull failed: %w: %s", err, string(output))
	}

	return nil
}

// Status returns the git status output
// Equivalent to: git -C <repoPath> status --porcelain
func Status(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "status", "--porcelain")

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git status failed: %w", err)
	}

	return string(output), nil
}

// IsClean checks if the repository has no uncommitted changes
func IsClean(repoPath string) (bool, error) {
	status, err := Status(repoPath)
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(status) == "", nil
}

// GetCurrentBranch returns the current git branch name
// Equivalent to: git -C <repoPath> branch --show-current
func GetCurrentBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "branch", "--show-current")

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git branch failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// IsGitRepo checks if a path is a git repository
func IsGitRepo(path string) bool {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil
}

// ExtractRepoName extracts the repository name from a git URL
// Handles various URL formats:
//   - https://github.com/user/repo.git -> repo
//   - git@github.com:user/repo.git -> repo
//   - https://github.com/user/repo -> repo
func ExtractRepoName(gitURL string) string {
	// Remove trailing .git if present
	url := strings.TrimSuffix(gitURL, ".git")

	// Split by / or :
	var parts []string
	if strings.Contains(url, "/") {
		parts = strings.Split(url, "/")
	} else if strings.Contains(url, ":") {
		parts = strings.Split(url, ":")
	}

	if len(parts) == 0 {
		return ""
	}

	// Return the last part
	return parts[len(parts)-1]
}
