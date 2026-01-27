package markdown

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

// ItemType represents the type of dump item
type ItemType string

const (
	ItemTypeTodo ItemType = "todo"
	ItemTypeNote ItemType = "note"
)

// DumpItem represents a parsed item from the dump file
type DumpItem struct {
	StartLine int
	EndLine   int
	Type      ItemType
	Content   string // Full content including any metadata
	RawLine   string // For tasks: the complete line; For notes: the title
}

var (
	taskPattern   = regexp.MustCompile(`^- \[ \] (.+)$`)
	notePattern   = regexp.MustCompile(`^\[Note\] (.+)$`)
	headerPattern = regexp.MustCompile(`^#+`)
	indentPattern = regexp.MustCompile(`^    `) // 4 spaces
)

// ParseDumpFile parses a dump file and returns all tasks and notes
// This replicates the parse_dump_items function from brain-api.sh (lines 33-83)
func ParseDumpFile(filePath string) ([]DumpItem, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var items []DumpItem
	scanner := bufio.NewScanner(file)

	lineNum := 0
	inNote := false
	noteStart := 0
	noteTitle := ""
	noteRawLine := ""

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Check if line is indented (part of note content)
		if indentPattern.MatchString(line) && inNote {
			// Continue accumulating note content
			continue
		}

		// If we were in a note and hit non-indented line, close the note
		if inNote {
			items = append(items, DumpItem{
				StartLine: noteStart,
				EndLine:   lineNum - 1,
				Type:      ItemTypeNote,
				Content:   noteTitle,
				RawLine:   noteRawLine,
			})
			inNote = false
		}

		// Skip empty lines (only whitespace)
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Skip markdown headers
		if headerPattern.MatchString(line) {
			continue
		}

		// Detect task
		if matches := taskPattern.FindStringSubmatch(line); matches != nil {
			taskContent := matches[1]
			items = append(items, DumpItem{
				StartLine: lineNum,
				EndLine:   lineNum,
				Type:      ItemTypeTodo,
				Content:   taskContent,
				RawLine:   line, // Full line including "- [ ] "
			})
		} else if matches := notePattern.FindStringSubmatch(line); matches != nil {
			// Detect note header
			inNote = true
			noteStart = lineNum
			noteTitle = matches[1]
			noteRawLine = noteTitle // For notes, we use the title for ID generation
		}
	}

	// Close any remaining note at end of file
	if inNote {
		items = append(items, DumpItem{
			StartLine: noteStart,
			EndLine:   lineNum,
			Type:      ItemTypeNote,
			Content:   noteTitle,
			RawLine:   noteRawLine,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

// ExtractTimestamp extracts the #captured:YYYY-MM-DD timestamp from content
// Returns the content without timestamp and the timestamp string
func ExtractTimestamp(content string) (string, string) {
	timestampPattern := regexp.MustCompile(`\s*#captured:([0-9-]+)$`)
	matches := timestampPattern.FindStringSubmatch(content)

	if matches == nil {
		return content, ""
	}

	timestamp := matches[1]
	cleanContent := timestampPattern.ReplaceAllString(content, "")

	return cleanContent, timestamp
}

// IsEmptyOrWhitespace checks if a string is empty or contains only whitespace
func IsEmptyOrWhitespace(s string) bool {
	return strings.TrimSpace(s) == ""
}
