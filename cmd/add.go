package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sandermoonemans/local-brain/pkg/config"
	"github.com/sandermoonemans/local-brain/pkg/external"
	"github.com/sandermoonemans/local-brain/pkg/fileutil"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [text]",
	Short: "Capture items to dump",
	Long: `Capture tasks or notes to the dump file.

Quick capture (with text):
  Appends a task to the dump: - [ ] your text #captured:YYYY-MM-DD

Editor mode (no text):
  Opens editor for writing notes. On close, the content is appended
  to dump as an indented note block:
    [Note] Title #captured:YYYY-MM-DD
        Your note content here...`,
	Example: `  brain add "Fix the authentication bug"
  brain add "Email Sarah about proposal"
  brain add                              # Opens editor for meeting notes`,
	RunE: runAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	brainPath, err := cfg.GetCurrentBrainPath()
	if err != nil {
		return fmt.Errorf("failed to get brain path: %w", err)
	}

	dumpPath := filepath.Join(brainPath, "00_dump.md")

	// Ensure dump exists
	if !fileutil.FileExists(dumpPath) {
		return fmt.Errorf("dump not found at %s. Run 'brain new' first", dumpPath)
	}

	timestamp := time.Now().Format("2006-01-02")

	// Quick capture mode: brain add "text"
	if len(args) > 0 {
		text := strings.Join(args, " ")

		// Acquire lock and append
		err := fileutil.WithLock(dumpPath, func() error {
			f, err := os.OpenFile(dumpPath, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			defer f.Close()

			line := fmt.Sprintf("- [ ] %s #captured:%s\n", text, timestamp)
			_, err = f.WriteString(line)
			return err
		})

		if err != nil {
			return fmt.Errorf("failed to append to dump: %w", err)
		}

		fmt.Println("OK: Added task to dump")
		return nil
	}

	// Editor mode: brain add (no args)
	return addNoteMode(dumpPath, timestamp)
}

func addNoteMode(dumpPath, timestamp string) error {
	// Prompt for title
	fmt.Print("Note title: ")
	reader := bufio.NewReader(os.Stdin)
	title, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read title: %w", err)
	}

	title = strings.TrimSpace(title)
	if title == "" {
		fmt.Println("Aborted (no title)")
		return nil
	}

	// Create temp file with helpful header
	initialContent := `# Write your note below. This line will be removed.
# Save and close the editor when done.

`

	// Open editor
	editor, err := external.DetectEditor()
	if err != nil {
		return err
	}

	content, err := editor.EditTemp(initialContent)
	if err != nil {
		return fmt.Errorf("editor failed: %w", err)
	}

	// Process content: strip comment lines and empty lines
	lines := strings.Split(content, "\n")
	var cleanLines []string
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.TrimSpace(line) != "" || len(cleanLines) > 0 {
			cleanLines = append(cleanLines, line)
		}
	}

	// Remove trailing empty lines
	for len(cleanLines) > 0 && strings.TrimSpace(cleanLines[len(cleanLines)-1]) == "" {
		cleanLines = cleanLines[:len(cleanLines)-1]
	}

	if len(cleanLines) == 0 {
		fmt.Println("Aborted (empty note)")
		return nil
	}

	// Append note to dump
	err = fileutil.WithLock(dumpPath, func() error {
		f, err := os.OpenFile(dumpPath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()

		// Write note header
		header := fmt.Sprintf("[Note] %s #captured:%s\n", title, timestamp)
		if _, err := f.WriteString(header); err != nil {
			return err
		}

		// Write indented content
		for _, line := range cleanLines {
			indentedLine := fmt.Sprintf("    %s\n", line)
			if _, err := f.WriteString(indentedLine); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to append note to dump: %w", err)
	}

	fmt.Println("OK: Added note to dump")
	return nil
}
