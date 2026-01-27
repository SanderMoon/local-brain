package api

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// TodoItem represents a task in a todo.md file
type TodoItem struct {
	ID       string `json:"id"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Status   string `json:"status"` // "open" or "done"
	Content  string `json:"content"`
	Project  string `json:"project"`
	RawLine  string `json:"-"` // Original line for ID generation
}

var (
	todoOpenPattern = regexp.MustCompile(`^\s*- \[ \] (.+)$`)
	todoDonePattern = regexp.MustCompile(`^\s*- \[[xX]\] (.+)$`)
)

// ParseAllTodos scans all todo.md files in active projects
func ParseAllTodos(activeDir string, includeCompleted bool) ([]TodoItem, error) {
	var todos []TodoItem

	// Scan all project directories
	entries, err := os.ReadDir(activeDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read active directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		projectName := entry.Name()
		todoFile := filepath.Join(activeDir, projectName, "todo.md")

		// Skip if todo.md doesn't exist
		if _, err := os.Stat(todoFile); os.IsNotExist(err) {
			continue
		}

		// Parse todo.md
		projectTodos, err := parseTodoFile(todoFile, projectName, includeCompleted)
		if err != nil {
			// Log error but continue with other projects
			continue
		}

		todos = append(todos, projectTodos...)
	}

	return todos, nil
}

func parseTodoFile(filePath, projectName string, includeCompleted bool) ([]TodoItem, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Get file mtime for ID generation
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	mtime := fileInfo.ModTime().Unix()

	var todos []TodoItem
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Check for open tasks: - [ ]
		if matches := todoOpenPattern.FindStringSubmatch(line); matches != nil {
			content := matches[1]
			id := GenerateTaskID(lineNum, line, mtime)

			todos = append(todos, TodoItem{
				ID:      id,
				File:    filePath,
				Line:    lineNum,
				Status:  "open",
				Content: content,
				Project: projectName,
				RawLine: line,
			})
		} else if includeCompleted {
			// Check for completed tasks: - [x] or - [X]
			if matches := todoDonePattern.FindStringSubmatch(line); matches != nil {
				content := matches[1]
				id := GenerateTaskID(lineNum, line, mtime)

				todos = append(todos, TodoItem{
					ID:      id,
					File:    filePath,
					Line:    lineNum,
					Status:  "done",
					Content: content,
					Project: projectName,
					RawLine: line,
				})
			}
		}
	}

	return todos, scanner.Err()
}

// FindTodoByID finds a todo by its ID
func FindTodoByID(todos []TodoItem, id string) *TodoItem {
	for i := range todos {
		if todos[i].ID == id {
			return &todos[i]
		}
	}
	return nil
}

// FindTodoByPattern finds todos matching a content pattern (case-insensitive)
func FindTodoByPattern(todos []TodoItem, pattern string) []TodoItem {
	var matches []TodoItem
	lowerPattern := strings.ToLower(pattern)

	for _, todo := range todos {
		if strings.Contains(strings.ToLower(todo.Content), lowerPattern) {
			matches = append(matches, todo)
		}
	}

	return matches
}

// ToggleTodoStatus updates a todo's status in the file
func ToggleTodoStatus(todo *TodoItem, newStatus string) error {
	// Read file
	content, err := os.ReadFile(todo.File)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	// Validate line number
	if todo.Line < 1 || todo.Line > len(lines) {
		return fmt.Errorf("invalid line number: %d", todo.Line)
	}

	// Update the line (1-indexed to 0-indexed)
	line := lines[todo.Line-1]

	if newStatus == "done" {
		// Change [ ] to [x]
		line = strings.Replace(line, "- [ ]", "- [x]", 1)
	} else if newStatus == "open" {
		// Change [x] or [X] to [ ]
		line = regexp.MustCompile(`- \[[xX]\]`).ReplaceAllString(line, "- [ ]")
	}

	lines[todo.Line-1] = line

	// Write back
	newContent := strings.Join(lines, "\n")
	return os.WriteFile(todo.File, []byte(newContent), 0644)
}

// DeleteTodoLine removes a todo line from the file
func DeleteTodoLine(todo *TodoItem) error {
	// Read file
	content, err := os.ReadFile(todo.File)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	// Validate line number
	if todo.Line < 1 || todo.Line > len(lines) {
		return fmt.Errorf("invalid line number: %d", todo.Line)
	}

	// Remove the line (1-indexed to 0-indexed)
	newLines := append(lines[:todo.Line-1], lines[todo.Line:]...)

	// Write back
	newContent := strings.Join(newLines, "\n")
	return os.WriteFile(todo.File, []byte(newContent), 0644)
}
