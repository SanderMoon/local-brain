package api

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sandermoonemans/local-brain/pkg/fileutil"
)

// MigrateTodoFileIDs adds #id: tags to all todos in a file that don't have one
// Returns the number of todos modified
func MigrateTodoFileIDs(todoFilePath string) (int, error) {
	content, err := os.ReadFile(todoFilePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	modified := 0
	todoPattern := regexp.MustCompile(`^\s*- \[[>\-xX ]\]`)

	for i, line := range lines {
		// Check if line is a todo
		if todoPattern.MatchString(line) {
			// Check if ID exists
			if ExtractID(line) == "" {
				// Add ID
				newLine, _ := AddIDToLine(line)
				lines[i] = newLine
				modified++
			}
		}
	}

	if modified > 0 {
		newContent := strings.Join(lines, "\n")
		err = fileutil.AtomicWriteFile(todoFilePath, []byte(newContent))
		if err != nil {
			return 0, fmt.Errorf("failed to write file: %w", err)
		}
	}

	return modified, nil
}

// MigrateDumpFileIDs adds #id: tags to all dump items
// Returns the number of items modified
func MigrateDumpFileIDs(dumpFilePath string) (int, error) {
	content, err := os.ReadFile(dumpFilePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	modified := 0
	taskPattern := regexp.MustCompile(`^\s*- \[ \]`)
	notePattern := regexp.MustCompile(`^\[Note\]`)

	for i, line := range lines {
		if taskPattern.MatchString(line) || notePattern.MatchString(line) {
			if ExtractID(line) == "" {
				newLine, _ := AddIDToLine(line)
				lines[i] = newLine
				modified++
			}
		}
	}

	if modified > 0 {
		newContent := strings.Join(lines, "\n")
		err = fileutil.AtomicWriteFile(dumpFilePath, []byte(newContent))
		if err != nil {
			return 0, fmt.Errorf("failed to write file: %w", err)
		}
	}

	return modified, nil
}

// MigrateAllProjectTodos migrates IDs for all projects
// Returns a map of project names to number of todos modified
func MigrateAllProjectTodos(activeDir string) (map[string]int, error) {
	results := make(map[string]int)

	entries, err := os.ReadDir(activeDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		todoFile := filepath.Join(activeDir, entry.Name(), "todo.md")
		if fileutil.FileExists(todoFile) {
			count, err := MigrateTodoFileIDs(todoFile)
			if err != nil {
				return results, err
			}
			if count > 0 {
				results[entry.Name()] = count
			}
		}
	}

	return results, nil
}
