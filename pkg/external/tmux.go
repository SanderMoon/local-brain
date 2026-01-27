package external

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// HasSession checks if a tmux session exists
// Equivalent to: tmux has-session -t <name>
func HasSession(name string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", name)
	err := cmd.Run()
	return err == nil
}

// CreateSession creates a new detached tmux session
// Equivalent to: tmux new-session -d -s <name> -c <dir>
func CreateSession(name, dir string) error {
	cmd := exec.Command("tmux", "new-session", "-d", "-s", name, "-c", dir)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create tmux session: %w: %s", err, string(output))
	}

	return nil
}

// SendKeys sends keys to a tmux session/window
// Equivalent to: tmux send-keys -t <target> <keys> C-m
func SendKeys(target string, keys string) error {
	cmd := exec.Command("tmux", "send-keys", "-t", target, keys, "C-m")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to send keys to tmux: %w: %s", err, string(output))
	}

	return nil
}

// AttachSession attaches to a tmux session
// If already in tmux, switches to the session instead
// Equivalent to: tmux attach -t <name> or tmux switch-client -t <name>
func AttachSession(name string) error {
	// Check if we're already in tmux
	inTmux := os.Getenv("TMUX") != ""

	var cmd *exec.Cmd
	if inTmux {
		// Switch to session
		cmd = exec.Command("tmux", "switch-client", "-t", name)
	} else {
		// Attach to session
		cmd = exec.Command("tmux", "attach", "-t", name)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// KillSession kills a tmux session
// Equivalent to: tmux kill-session -t <name>
func KillSession(name string) error {
	cmd := exec.Command("tmux", "kill-session", "-t", name)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to kill tmux session: %w: %s", err, string(output))
	}

	return nil
}

// ListSessions returns a list of tmux session names
// Equivalent to: tmux list-sessions -F "#{session_name}"
func ListSessions() ([]string, error) {
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}")

	output, err := cmd.Output()
	if err != nil {
		// No sessions is not an error
		if strings.Contains(err.Error(), "no server running") {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to list tmux sessions: %w", err)
	}

	sessions := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(sessions) == 1 && sessions[0] == "" {
		return []string{}, nil
	}

	return sessions, nil
}

// IsTmuxAvailable checks if tmux is installed
func IsTmuxAvailable() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

// IsInTmux checks if the current shell is running inside tmux
func IsInTmux() bool {
	return os.Getenv("TMUX") != ""
}

// NewWindow creates a new window in a tmux session
// Equivalent to: tmux new-window -t <session>:<window> -n <name> -c <dir>
func NewWindow(session string, window int, name string, dir string) error {
	target := fmt.Sprintf("%s:%d", session, window)
	cmd := exec.Command("tmux", "new-window", "-t", target, "-n", name, "-c", dir)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create tmux window: %w: %s", err, string(output))
	}

	return nil
}

// SelectWindow selects a specific window in a tmux session
// Equivalent to: tmux select-window -t <session>:<window>
func SelectWindow(session string, window int) error {
	target := fmt.Sprintf("%s:%d", session, window)
	cmd := exec.Command("tmux", "select-window", "-t", target)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to select tmux window: %w: %s", err, string(output))
	}

	return nil
}
