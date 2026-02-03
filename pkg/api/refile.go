package api

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sandermoonemans/local-brain/pkg/fileutil"
	"github.com/sandermoonemans/local-brain/pkg/markdown"
)

// RefileTaskToProject moves a task from dump to a project's todo.md file
// The task is inserted under the "## Active" section
// ID is preserved from the original task
func RefileTaskToProject(projectDir string, item *markdown.DumpItem) error {
	if item.Type != markdown.ItemTypeTodo {
		return fmt.Errorf("item is not a task")
	}

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

	// Read existing content
	content, err := os.ReadFile(todoFile)
	if err != nil {
		return fmt.Errorf("failed to read todo.md: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	// Create task with stable ID preserved from original
	newTaskLine := fmt.Sprintf("- [ ] %s", item.Content)
	newTaskLine, _ = AddIDToLine(newTaskLine)

	// Find the "## Active" section and insert after it
	activeIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "## Active" {
			activeIdx = i
			break
		}
	}

	var newLines []string
	if activeIdx == -1 {
		// No Active section found, append to end (fallback)
		newLines = append(lines, newTaskLine)
	} else {
		// Insert after "## Active" line, after any empty lines
		insertIdx := activeIdx + 1

		// Skip empty lines immediately after "## Active"
		for insertIdx < len(lines) && strings.TrimSpace(lines[insertIdx]) == "" {
			insertIdx++
		}

		// Insert the new task
		newLines = make([]string, 0, len(lines)+1)
		newLines = append(newLines, lines[:insertIdx]...)
		newLines = append(newLines, newTaskLine)
		newLines = append(newLines, lines[insertIdx:]...)
	}

	// Write back atomically
	newContent := strings.Join(newLines, "\n")
	return fileutil.AtomicWriteFile(todoFile, []byte(newContent))
}

// RefileNoteToProject moves a note from dump to a project's notes/ directory
// Creates a timestamped markdown file with the note content
// Returns the created file path
func RefileNoteToProject(dumpPath, projectDir string, item *markdown.DumpItem) (string, error) {
	if item.Type != markdown.ItemTypeNote {
		return "", fmt.Errorf("item is not a note")
	}

	notesDir := filepath.Join(projectDir, "notes")
	if err := fileutil.EnsureDir(notesDir); err != nil {
		return "", fmt.Errorf("failed to create notes directory: %w", err)
	}

	// Extract date and clean title
	cleanTitle, capturedDate := ExtractCapturedDate(item.Content)
	if capturedDate == "" {
		// Fallback to today if no captured date
		capturedDate = strings.Split(strings.TrimSpace(strings.Split(item.Content, "#")[0]), " ")[0]
		if !regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`).MatchString(capturedDate) {
			capturedDate = ""
		}
	}

	cleanTitle = strings.TrimSpace(cleanTitle)

	// Create slug
	slug := Slugify(cleanTitle)
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
	content, err := readNoteContentFromDump(dumpPath, item.StartLine, item.EndLine)
	if err != nil {
		return "", fmt.Errorf("failed to read note content: %w", err)
	}

	// Create note file
	noteContent := fmt.Sprintf("# %s\n\nCreated: %s\n\n%s\n", cleanTitle, capturedDate, content)
	if err := os.WriteFile(filePath, []byte(noteContent), 0644); err != nil {
		return "", fmt.Errorf("failed to create note file: %w", err)
	}

	return filePath, nil
}

// readNoteContentFromDump extracts the indented content from a note in the dump file
// Removes the 4-space indentation from content lines
func readNoteContentFromDump(dumpPath string, startLine, endLine int) (string, error) {
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
		} else if strings.TrimSpace(line) != "" {
			// Non-indented, non-empty line (shouldn't happen in valid notes)
			contentLines = append(contentLines, line)
		}
	}

	return strings.Join(contentLines, "\n"), scanner.Err()
}
