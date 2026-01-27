package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/sandermoonemans/local-brain/pkg/api"
	"github.com/sandermoonemans/local-brain/pkg/config"
	"github.com/sandermoonemans/local-brain/pkg/external"
	"github.com/sandermoonemans/local-brain/pkg/fileutil"
	"github.com/sandermoonemans/local-brain/pkg/markdown"
	"github.com/spf13/cobra"
)

var refileCmd = &cobra.Command{
	Use:   "refile [ID] [project]",
	Short: "Process dump items to projects",
	Long: `Process dump items to projects.

Interactive mode (no arguments):
  Process items one by one with fuzzy project selection

Direct mode (with ID and project):
  Refile specific item to a specific project

Tasks go to project's todo.md
Notes go to project's notes/ directory as separate files`,
	Example: `  brain refile              # Start interactive refiling
  brain refile a1b2c3 work  # Refile item a1b2c3 to 'work' project`,
	Args: cobra.MaximumNArgs(2),
	RunE: runRefile,
}

func init() {
	rootCmd.AddCommand(refileCmd)
}

func runRefile(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	brainPath, err := cfg.GetCurrentBrainPath()
	if err != nil {
		return fmt.Errorf("failed to get brain path: %w", err)
	}

	dumpPath := filepath.Join(brainPath, "00_dump.md")
	activeDir := filepath.Join(brainPath, "01_active")

	if !fileutil.FileExists(dumpPath) {
		return fmt.Errorf("dump not found at %s", dumpPath)
	}

	if !fileutil.FileExists(activeDir) {
		return fmt.Errorf("active projects directory not found: %s", activeDir)
	}

	// Direct mode: brain refile <ID> <project>
	if len(args) == 2 {
		itemID := args[0]
		projectName := args[1]
		return refileDirect(dumpPath, activeDir, itemID, projectName)
	}

	// Interactive mode
	return refileInteractive(dumpPath, activeDir)
}

func refileDirect(dumpPath, activeDir, itemID, projectName string) error {
	// Parse dump
	items, err := markdown.ParseDumpFile(dumpPath)
	if err != nil {
		return fmt.Errorf("failed to parse dump: %w", err)
	}

	// Get file mtime for ID calculation
	fileInfo, err := os.Stat(dumpPath)
	if err != nil {
		return fmt.Errorf("failed to stat dump: %w", err)
	}
	mtime := fileInfo.ModTime().Unix()

	// Find item by ID
	var targetItem *markdown.DumpItem
	for i := range items {
		item := &items[i]
		var id string
		if item.Type == markdown.ItemTypeTodo {
			id = api.GenerateTaskID(item.StartLine, item.RawLine, mtime)
		} else {
			id = api.GenerateNoteID(item.StartLine, item.RawLine, mtime)
		}

		if id == itemID {
			targetItem = item
			break
		}
	}

	if targetItem == nil {
		return fmt.Errorf("item with ID '%s' not found", itemID)
	}

	// Verify project exists
	projectDir := filepath.Join(activeDir, projectName)
	if !fileutil.FileExists(projectDir) {
		return fmt.Errorf("project '%s' not found", projectName)
	}

	// Refile item
	if err := refileItem(targetItem, projectDir, dumpPath, mtime); err != nil {
		return err
	}

	// Remove item from dump
	if err := removeItemFromDump(dumpPath, targetItem.StartLine, targetItem.EndLine); err != nil {
		return fmt.Errorf("failed to remove item from dump: %w", err)
	}

	fmt.Printf("OK: Refiled item %s to %s\n", itemID, projectName)
	return nil
}

func refileInteractive(dumpPath, activeDir string) error {
	// Check for fzf
	if !external.IsFZFAvailable() {
		return fmt.Errorf("fzf not found (required for interactive mode)")
	}

	// Get list of projects
	projects, err := listProjects(activeDir)
	if err != nil {
		return err
	}

	if len(projects) == 0 {
		return fmt.Errorf("no projects found. Create one with: brain project new <name>")
	}

	// Add special options
	options := append([]string{"[SKIP]", "[TRASH]"}, projects...)

	// Parse dump
	items, err := markdown.ParseDumpFile(dumpPath)
	if err != nil {
		return fmt.Errorf("failed to parse dump: %w", err)
	}

	if len(items) == 0 {
		fmt.Println("No items in dump")
		return nil
	}

	// Get file mtime
	fileInfo, err := os.Stat(dumpPath)
	if err != nil {
		return fmt.Errorf("failed to stat dump: %w", err)
	}
	mtime := fileInfo.ModTime().Unix()

	processed := 0

	// Process each item
	for i := range items {
		item := &items[i]

		// Show item
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Printf("Item %d of %d\n", i+1, len(items))
		fmt.Printf("Type: %s\n", item.Type)

		// Extract clean content
		content, _ := markdown.ExtractTimestamp(item.Content)
		fmt.Printf("Content: %s\n", content)

		if item.Type == markdown.ItemTypeNote {
			lines := item.EndLine - item.StartLine + 1
			fmt.Printf("Lines: %d\n", lines)
		}

		fmt.Println(strings.Repeat("=", 60))

		// Select project
		selected, err := external.SelectOne(options, external.FZFOptions{
			Header: "Select destination project",
			Prompt: "Project> ",
			Height: "40%",
		})

		if err != nil {
			if err.Error() == "cancelled" {
				fmt.Println("\nRefile cancelled")
				break
			}
			return err
		}

		// Handle selection
		if selected == "[SKIP]" {
			fmt.Println("Skipped")
			continue
		}

		if selected == "[TRASH]" {
			// Remove from dump
			if err := removeItemFromDump(dumpPath, item.StartLine, item.EndLine); err != nil {
				return fmt.Errorf("failed to remove item: %w", err)
			}
			fmt.Println("Deleted")
			processed++
			continue
		}

		// Refile to project
		projectDir := filepath.Join(activeDir, selected)
		if err := refileItem(item, projectDir, dumpPath, mtime); err != nil {
			return err
		}

		// Remove from dump
		if err := removeItemFromDump(dumpPath, item.StartLine, item.EndLine); err != nil {
			return fmt.Errorf("failed to remove item: %w", err)
		}

		fmt.Printf("Refiled to %s\n", selected)
		processed++
	}

	fmt.Printf("\nProcessed %d items\n", processed)
	return nil
}

func refileItem(item *markdown.DumpItem, projectDir, dumpPath string, mtime int64) error {
	if item.Type == markdown.ItemTypeTodo {
		return refileTask(item, projectDir)
	}
	return refileNote(item, projectDir, dumpPath)
}

func refileTask(item *markdown.DumpItem, projectDir string) error {
	todoFile := filepath.Join(projectDir, "todo.md")

	// Ensure todo.md exists
	if !fileutil.FileExists(todoFile) {
		content := `# Tasks

## Active

## Completed
`
		if err := os.WriteFile(todoFile, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create todo.md: %w", err)
		}
	}

	// Append task (preserve the original format with captured timestamp)
	err := fileutil.WithLock(todoFile, func() error {
		f, err := os.OpenFile(todoFile, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = fmt.Fprintf(f, "- [ ] %s\n", item.Content)
		return err
	})

	return err
}

func refileNote(item *markdown.DumpItem, projectDir, dumpPath string) error {
	notesDir := filepath.Join(projectDir, "notes")
	if err := fileutil.EnsureDir(notesDir); err != nil {
		return fmt.Errorf("failed to create notes directory: %w", err)
	}

	// Extract date and clean title
	capturedDate := time.Now().Format("2006-01-02")
	cleanTitle := item.Content

	timestampPattern := regexp.MustCompile(`\s*#captured:([0-9-]+)`)
	if matches := timestampPattern.FindStringSubmatch(item.Content); matches != nil {
		capturedDate = matches[1]
		cleanTitle = timestampPattern.ReplaceAllString(item.Content, "")
	}

	cleanTitle = strings.TrimSpace(cleanTitle)

	// Create slug
	slug := slugify(cleanTitle)
	if slug == "" {
		slug = "note"
	}

	// Create filename
	filename := fmt.Sprintf("%s-%s.md", capturedDate, slug)
	filePath := filepath.Join(notesDir, filename)

	// Handle duplicates
	counter := 1
	for fileutil.FileExists(filePath) {
		filename = fmt.Sprintf("%s-%s-%d.md", capturedDate, slug, counter)
		filePath = filepath.Join(notesDir, filename)
		counter++
	}

	// Get note content from dump
	content, err := readNoteContent(dumpPath, item.StartLine, item.EndLine)
	if err != nil {
		return fmt.Errorf("failed to read note content: %w", err)
	}

	// Create note file
	noteContent := fmt.Sprintf("# %s\n\nCreated: %s\n\n%s\n", cleanTitle, capturedDate, content)
	if err := os.WriteFile(filePath, []byte(noteContent), 0644); err != nil {
		return fmt.Errorf("failed to create note file: %w", err)
	}

	return nil
}

func readNoteContent(dumpPath string, startLine, endLine int) (string, error) {
	file, err := os.Open(dumpPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	var contentLines []string

	for scanner.Scan() {
		lineNum++

		// Skip until start line
		if lineNum < startLine {
			continue
		}

		// Stop after end line
		if lineNum > endLine {
			break
		}

		line := scanner.Text()

		// Skip the note header line
		if lineNum == startLine {
			continue
		}

		// Remove 4-space indent from content lines
		if strings.HasPrefix(line, "    ") {
			contentLines = append(contentLines, line[4:])
		}
	}

	return strings.Join(contentLines, "\n"), scanner.Err()
}

func removeItemFromDump(dumpPath string, startLine, endLine int) error {
	// Read entire file
	content, err := os.ReadFile(dumpPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")

	// Remove lines (1-indexed to 0-indexed conversion)
	var newLines []string
	for i, line := range lines {
		lineNum := i + 1
		if lineNum < startLine || lineNum > endLine {
			newLines = append(newLines, line)
		}
	}

	// Write back
	newContent := strings.Join(newLines, "\n")
	return fileutil.AtomicWriteFile(dumpPath, []byte(newContent))
}

func listProjects(activeDir string) ([]string, error) {
	entries, err := os.ReadDir(activeDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read active directory: %w", err)
	}

	var projects []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			projects = append(projects, entry.Name())
		}
	}

	return projects, nil
}

func slugify(text string) string {
	// Take first 40 characters
	if len(text) > 40 {
		text = text[:40]
	}

	// Convert to lowercase
	text = strings.ToLower(text)

	// Replace spaces with hyphens
	text = strings.ReplaceAll(text, " ", "-")

	// Keep only alphanumeric and hyphens
	reg := regexp.MustCompile(`[^a-z0-9-]+`)
	text = reg.ReplaceAllString(text, "")

	// Replace multiple hyphens with single hyphen
	reg = regexp.MustCompile(`-+`)
	text = reg.ReplaceAllString(text, "-")

	// Trim hyphens from edges
	text = strings.Trim(text, "-")

	return text
}
