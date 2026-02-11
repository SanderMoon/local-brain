package api

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/sandermoonemans/local-brain/pkg/fileutil"
	"github.com/sandermoonemans/local-brain/pkg/markdown"
)

// TodoItem represents a task in a todo.md file
type TodoItem struct {
	ID            string   `json:"id"`
	File          string   `json:"file"`
	Line          int      `json:"line"`
	Status        string   `json:"status"` // "open", "in-progress", "blocked", or "done"
	Content       string   `json:"content"`
	Project       string   `json:"project"`
	Priority      *int     `json:"priority"`                 // 1=high, 2=medium, 3=low, nil=unprioritized
	DueDate       string   `json:"due_date"`                 // YYYY-MM-DD format, empty if no due date
	Tags          []string `json:"tags"`                     // Freeform tags (e.g., "bug", "feature", "urgent")
	RawLine       string   `json:"-"`                        // Original line for ID generation
	CapturedDate  string   `json:"captured_date,omitempty"`  // YYYY-MM-DD from #captured:
	CompletedDate string   `json:"completed_date,omitempty"` // YYYY-MM-DD from #done:
}

var (
	todoOpenPattern       = regexp.MustCompile(`^\s*- \[ \] (.+)$`)
	todoInProgressPattern = regexp.MustCompile(`^\s*- \[>\] (.+)$`)
	todoBlockedPattern    = regexp.MustCompile(`^\s*- \[-\] (.+)$`)
	todoDonePattern       = regexp.MustCompile(`^\s*- \[[xX]\] (.+)$`)
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

	var todos []TodoItem
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		var status string
		var matches []string

		// Check for all task states
		if matches = todoOpenPattern.FindStringSubmatch(line); matches != nil {
			status = "open"
		} else if matches = todoInProgressPattern.FindStringSubmatch(line); matches != nil {
			status = "in-progress"
		} else if matches = todoBlockedPattern.FindStringSubmatch(line); matches != nil {
			status = "blocked"
		} else if includeCompleted {
			if matches = todoDonePattern.FindStringSubmatch(line); matches != nil {
				status = "done"
			}
		}

		// If we found a match, create the TodoItem
		if matches != nil {
			rawContent := matches[1]

			// Extract or generate stable ID
			id := GetOrGenerateID(line)

			// Extract metadata (priority, due date, tags)
			content, priority := markdown.ExtractPriority(rawContent)
			content, dueDate := markdown.ExtractDueDate(content)
			content, tags := markdown.ExtractTags(content)

			// Extract temporal metadata
			content, capturedDate := markdown.ExtractTimestamp(content)
			var completedDate string
			if status == "done" {
				content, completedDate = markdown.ExtractCompletedDate(content)
			}

			// Remove #id: tag from content for clean display
			content = RemoveIDFromContent(content)

			todos = append(todos, TodoItem{
				ID:            id,
				File:          filePath,
				Line:          lineNum,
				Status:        status,
				Content:       content,
				Project:       projectName,
				Priority:      priority,
				DueDate:       dueDate,
				Tags:          tags,
				RawLine:       line,
				CapturedDate:  capturedDate,
				CompletedDate: completedDate,
			})
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

// SearchTodos filters todos by multiple criteria
// This is optimized for MCP server use case with flexible filtering
func SearchTodos(todos []TodoItem, query, project, status string, tags []string) []TodoItem {
	var results []TodoItem

	for _, todo := range todos {
		// Filter by content query (case-insensitive substring match)
		if query != "" {
			if !strings.Contains(strings.ToLower(todo.Content), strings.ToLower(query)) {
				continue
			}
		}

		// Filter by project name
		if project != "" {
			if todo.Project != project {
				continue
			}
		}

		// Filter by status
		if status != "" {
			if todo.Status != status {
				continue
			}
		}

		// Filter by tags (OR logic - match any of the provided tags)
		if len(tags) > 0 {
			matched := false
			for _, requiredTag := range tags {
				for _, todoTag := range todo.Tags {
					if strings.EqualFold(todoTag, requiredTag) {
						matched = true
						break
					}
				}
				if matched {
					break
				}
			}
			if !matched {
				continue
			}
		}

		results = append(results, todo)
	}

	return results
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

// SetTodoPriority sets or clears the priority tag for a todo item
// Priority should be 1 (high), 2 (medium), 3 (low), or nil (clear priority)
func SetTodoPriority(todo *TodoItem, priority *int) error {
	// Validate priority value if provided
	if priority != nil && (*priority < 1 || *priority > 3) {
		return fmt.Errorf("invalid priority: %d (must be 1-3)", *priority)
	}

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

	// Get the current line (1-indexed to 0-indexed)
	line := lines[todo.Line-1]

	// Parse the line to extract checkbox and content
	var checkbox, taskContent string
	if strings.Contains(line, "- [ ]") {
		checkbox = "- [ ]"
		taskContent = strings.TrimSpace(strings.TrimPrefix(line, "- [ ]"))
	} else if strings.Contains(line, "- [>]") {
		checkbox = "- [>]"
		taskContent = strings.TrimSpace(strings.TrimPrefix(line, "- [>]"))
	} else if strings.Contains(line, "- [-]") {
		checkbox = "- [-]"
		taskContent = strings.TrimSpace(strings.TrimPrefix(line, "- [-]"))
	} else if strings.Contains(line, "- [x]") || strings.Contains(line, "- [X]") {
		if strings.Contains(line, "- [x]") {
			checkbox = "- [x]"
			taskContent = strings.TrimSpace(strings.TrimPrefix(line, "- [x]"))
		} else {
			checkbox = "- [X]"
			taskContent = strings.TrimSpace(strings.TrimPrefix(line, "- [X]"))
		}
	} else {
		return fmt.Errorf("line is not a valid todo item")
	}

	// Remove any existing #p: tag from content
	priorityPattern := regexp.MustCompile(`\s*#p:[1-3](?:\s|$)`)
	taskContent = priorityPattern.ReplaceAllString(taskContent, "")
	taskContent = strings.TrimSpace(taskContent)

	// Add new priority tag if provided
	if priority != nil {
		taskContent = fmt.Sprintf("%s #p:%d", taskContent, *priority)
	}

	// Reconstruct the line
	newLine := fmt.Sprintf("%s %s", checkbox, taskContent)
	lines[todo.Line-1] = newLine

	// Write back
	newContent := strings.Join(lines, "\n")
	return os.WriteFile(todo.File, []byte(newContent), 0644)
}

// SetTodoStatus sets the status of a todo item by changing its checkbox
// Valid statuses: "open", "in-progress", "blocked", "done"
func SetTodoStatus(todo *TodoItem, newStatus string) error {
	// Validate status and get checkbox symbol (without brackets)
	validStatuses := map[string]string{
		"open":        " ",
		"in-progress": ">",
		"blocked":     "-",
		"done":        "x",
	}

	checkboxSymbol, ok := validStatuses[newStatus]
	if !ok {
		return fmt.Errorf("invalid status: %s (must be: open, in-progress, blocked, done)", newStatus)
	}

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

	// Get the current line (1-indexed to 0-indexed)
	line := lines[todo.Line-1]

	// Replace checkbox pattern in line
	// Match any checkbox pattern: [ ], [>], [-], [x], [X]
	checkboxPattern := regexp.MustCompile(`^(\s*)- \[[ >xX-]\]`)
	if !checkboxPattern.MatchString(line) {
		return fmt.Errorf("line is not a valid todo item")
	}

	// Replace with new checkbox
	newLine := checkboxPattern.ReplaceAllString(line, "${1}- ["+checkboxSymbol+"]")

	// Manage #done: timestamp
	doneTagPattern := regexp.MustCompile(`\s*#done:[0-9-]+`)
	if newStatus == "done" {
		if !doneTagPattern.MatchString(newLine) {
			today := time.Now().Format("2006-01-02")
			newLine = newLine + " #done:" + today
		}
	} else {
		newLine = doneTagPattern.ReplaceAllString(newLine, "")
		newLine = strings.TrimRight(newLine, " \t")
	}

	lines[todo.Line-1] = newLine

	// Write back
	newContent := strings.Join(lines, "\n")
	return os.WriteFile(todo.File, []byte(newContent), 0644)
}

// SetTodoDueDate sets or clears the due date tag for a todo item
// dueDate should be in YYYY-MM-DD format, or empty string to clear
func SetTodoDueDate(todo *TodoItem, dueDate string) error {
	// "clear" means remove due date
	if dueDate == "clear" {
		dueDate = ""
	}

	// Validate date format if provided (not empty)
	if dueDate != "" {
		// Try to parse as valid date
		if matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, dueDate); !matched {
			return fmt.Errorf("invalid date format: %s (must be YYYY-MM-DD)", dueDate)
		}
		// Actually parse the date to validate it's real
		_, err := time.Parse("2006-01-02", dueDate)
		if err != nil {
			return fmt.Errorf("invalid date: %s", dueDate)
		}
	}

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

	// Get the current line (1-indexed to 0-indexed)
	line := lines[todo.Line-1]

	// Parse the line to extract checkbox and content
	var checkbox, taskContent string
	if strings.Contains(line, "- [ ]") {
		checkbox = "- [ ]"
		taskContent = strings.TrimSpace(strings.TrimPrefix(line, "- [ ]"))
	} else if strings.Contains(line, "- [>]") {
		checkbox = "- [>]"
		taskContent = strings.TrimSpace(strings.TrimPrefix(line, "- [>]"))
	} else if strings.Contains(line, "- [-]") {
		checkbox = "- [-]"
		taskContent = strings.TrimSpace(strings.TrimPrefix(line, "- [-]"))
	} else if strings.Contains(line, "- [x]") || strings.Contains(line, "- [X]") {
		if strings.Contains(line, "- [x]") {
			checkbox = "- [x]"
			taskContent = strings.TrimSpace(strings.TrimPrefix(line, "- [x]"))
		} else {
			checkbox = "- [X]"
			taskContent = strings.TrimSpace(strings.TrimPrefix(line, "- [X]"))
		}
	} else {
		return fmt.Errorf("line is not a valid todo item")
	}

	// Remove any existing #due: tag from content
	dueDatePattern := regexp.MustCompile(`\s*#due:[^\s]+(?:\s|$)`)
	taskContent = dueDatePattern.ReplaceAllString(taskContent, "")
	taskContent = strings.TrimSpace(taskContent)

	// Add new due date tag if provided
	if dueDate != "" {
		taskContent = fmt.Sprintf("%s #due:%s", taskContent, dueDate)
	}

	// Reconstruct the line
	newLine := fmt.Sprintf("%s %s", checkbox, taskContent)
	lines[todo.Line-1] = newLine

	// Write back
	newContent := strings.Join(lines, "\n")
	return os.WriteFile(todo.File, []byte(newContent), 0644)
}

// AddTodoTags adds one or more tags to a todo item
func AddTodoTags(todo *TodoItem, newTags []string) error {
	if len(newTags) == 0 {
		return nil
	}

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

	// Get the current line (1-indexed to 0-indexed)
	line := lines[todo.Line-1]

	// Get existing tags from current todo
	existingTagsMap := make(map[string]bool)
	for _, tag := range todo.Tags {
		existingTagsMap[strings.ToLower(tag)] = true
	}

	// Filter out tags that already exist (case-insensitive)
	var tagsToAdd []string
	for _, tag := range newTags {
		if !existingTagsMap[strings.ToLower(tag)] {
			tagsToAdd = append(tagsToAdd, tag)
		}
	}

	if len(tagsToAdd) == 0 {
		return nil // No new tags to add
	}

	// Append tags to the end of the line
	for _, tag := range tagsToAdd {
		line = line + " #" + tag
	}

	lines[todo.Line-1] = line

	// Write back
	newContent := strings.Join(lines, "\n")
	return os.WriteFile(todo.File, []byte(newContent), 0644)
}

// RemoveTodoTags removes one or more tags from a todo item
func RemoveTodoTags(todo *TodoItem, tagsToRemove []string) error {
	if len(tagsToRemove) == 0 {
		return nil
	}

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

	// Get the current line (1-indexed to 0-indexed)
	line := lines[todo.Line-1]

	// Remove each tag
	for _, tag := range tagsToRemove {
		// Remove both forms: #tag and # tag (in case there's a space)
		line = strings.ReplaceAll(line, " #"+tag, "")
		line = strings.ReplaceAll(line, "#"+tag, "")
	}

	// Clean up extra spaces
	line = regexp.MustCompile(`\s+`).ReplaceAllString(line, " ")
	line = strings.TrimSpace(line)

	// Ensure checkbox remains at the start with proper spacing
	checkboxPattern := regexp.MustCompile(`^(\s*)- \[[ >xX-]\]`)
	if checkboxPattern.MatchString(line) {
		// Line is already properly formatted
	} else {
		// Something went wrong - this shouldn't happen
		return fmt.Errorf("line formatting corrupted after tag removal")
	}

	lines[todo.Line-1] = line

	// Write back
	newContent := strings.Join(lines, "\n")
	return os.WriteFile(todo.File, []byte(newContent), 0644)
}

// ListAllTags returns a map of all tags across todos with their occurrence counts
func ListAllTags(todos []TodoItem) map[string]int {
	tagCounts := make(map[string]int)

	for _, todo := range todos {
		for _, tag := range todo.Tags {
			tagCounts[tag]++
		}
	}

	return tagCounts
}

// UpdateTodoContent updates the text content of a todo while preserving all metadata
// The new content should NOT include checkbox, metadata tags, or ID - those are preserved
func UpdateTodoContent(todo *TodoItem, newContent string) error {
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

	// Get the current line (1-indexed to 0-indexed)
	line := lines[todo.Line-1]

	// Extract checkbox
	checkboxPattern := regexp.MustCompile(`^(\s*)- \[([>xX \-])\] `)
	matches := checkboxPattern.FindStringSubmatch(line)
	if matches == nil {
		return fmt.Errorf("line is not a valid todo item")
	}

	indent := matches[1]
	checkboxSymbol := matches[2]

	// Extract all metadata from original line (priority, due date, tags, ID, timestamps)
	// We do this by removing the checkbox and content, leaving just the metadata
	afterCheckbox := checkboxPattern.ReplaceAllString(line, "")

	// Remove the actual content text to isolate metadata
	// The content is everything up to the first # tag or the ID
	contentAndMetadata := strings.TrimSpace(afterCheckbox)

	// Extract all metadata tags (anything starting with #)
	metadataPattern := regexp.MustCompile(`(#[^\s]+(\s|$))+`)
	metadataMatches := metadataPattern.FindAllString(contentAndMetadata, -1)

	var allMetadata string
	if len(metadataMatches) > 0 {
		allMetadata = " " + strings.Join(metadataMatches, "")
	}

	// Reconstruct line: checkbox + new content + metadata
	newLine := fmt.Sprintf("%s- [%s] %s%s", indent, checkboxSymbol, newContent, allMetadata)
	newLine = strings.TrimRight(newLine, " \t") // Clean up trailing whitespace

	lines[todo.Line-1] = newLine

	// Write back
	newFileContent := strings.Join(lines, "\n")
	return os.WriteFile(todo.File, []byte(newFileContent), 0644)
}

// TodoCreateRequest represents a todo to be created with optional metadata
type TodoCreateRequest struct {
	Content  string
	Priority *int     // 1=high, 2=medium, 3=low, nil=no priority
	DueDate  string   // YYYY-MM-DD format, empty=no due date
	Tags     []string // Optional tags
	Status   string   // open, in-progress, blocked (default: open)
}

// AppendTodoWithMetadata adds a new task with metadata to a project's todo.md file
func AppendTodoWithMetadata(projectDir string, todo TodoCreateRequest) error {
	todoFile := filepath.Join(projectDir, "todo.md")

	// Ensure todo.md exists
	if !fileutil.FileExists(todoFile) {
		todoContent := `# Tasks

## Active

## Completed
`
		if err := os.WriteFile(todoFile, []byte(todoContent), 0644); err != nil {
			return fmt.Errorf("failed to create todo.md: %w", err)
		}
	}

	// Default status to open if not specified
	status := todo.Status
	if status == "" {
		status = "open"
	}

	// Map status to checkbox symbol
	var checkboxSymbol string
	switch status {
	case "open":
		checkboxSymbol = " "
	case "in-progress":
		checkboxSymbol = ">"
	case "blocked":
		checkboxSymbol = "-"
	case "done":
		checkboxSymbol = "x"
	default:
		return fmt.Errorf("invalid status: %s", status)
	}

	// Build content with metadata tags
	contentWithMetadata := todo.Content

	// Add priority
	if todo.Priority != nil {
		contentWithMetadata = fmt.Sprintf("%s #p:%d", contentWithMetadata, *todo.Priority)
	}

	// Add due date
	if todo.DueDate != "" {
		contentWithMetadata = fmt.Sprintf("%s #due:%s", contentWithMetadata, todo.DueDate)
	}

	// Add tags
	for _, tag := range todo.Tags {
		contentWithMetadata = fmt.Sprintf("%s #%s", contentWithMetadata, tag)
	}

	// Create task line with checkbox
	newTaskLine := fmt.Sprintf("- [%s] %s", checkboxSymbol, contentWithMetadata)

	// Add ID
	newTaskLine, _ = AddIDToLine(newTaskLine)

	// Read existing content
	fileContent, err := os.ReadFile(todoFile)
	if err != nil {
		return fmt.Errorf("failed to read todo.md: %w", err)
	}

	lines := strings.Split(string(fileContent), "\n")

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
		// No Active section found, append to end
		newLines = append(lines, newTaskLine)
	} else {
		// Insert after "## Active", after any empty lines
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

// AppendTodoToProject adds a new task to a project's todo.md file
// The task is inserted under the "## Active" section with a new ID
func AppendTodoToProject(projectDir, content string) error {
	todoFile := filepath.Join(projectDir, "todo.md")

	// Ensure todo.md exists
	if !fileutil.FileExists(todoFile) {
		todoContent := `# Tasks

## Active

## Completed
`
		if err := os.WriteFile(todoFile, []byte(todoContent), 0644); err != nil {
			return fmt.Errorf("failed to create todo.md: %w", err)
		}
	}

	// Read existing content
	fileContent, err := os.ReadFile(todoFile)
	if err != nil {
		return fmt.Errorf("failed to read todo.md: %w", err)
	}

	lines := strings.Split(string(fileContent), "\n")

	// Create new task line with ID
	newTaskLine := fmt.Sprintf("- [ ] %s", content)
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
		// No Active section found, append to end
		newLines = append(lines, newTaskLine)
	} else {
		// Insert after "## Active", after any empty lines
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
