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

// ExtractPriority extracts the #p:[1-3] priority tag from content
// Returns the content without priority tag and the priority value (1=high, 2=medium, 3=low)
// Returns nil priority if no valid tag is found
func ExtractPriority(content string) (string, *int) {
	priorityPattern := regexp.MustCompile(`\s*#p:([1-3])(?:\s|$)`)
	matches := priorityPattern.FindStringSubmatch(content)

	if matches == nil {
		return content, nil
	}

	priorityStr := matches[1]
	priority := int(priorityStr[0] - '0')
	cleanContent := priorityPattern.ReplaceAllString(content, " ")
	cleanContent = strings.TrimSpace(cleanContent)

	return cleanContent, &priority
}

// ExtractDueDate extracts the #due:YYYY-MM-DD tag from content
// Returns the content without due date tag and the due date string (YYYY-MM-DD or whatever was in the tag)
// Returns empty string if no valid tag is found
func ExtractDueDate(content string) (string, string) {
	dueDatePattern := regexp.MustCompile(`\s*#due:([^\s]+)(?:\s|$)`)
	matches := dueDatePattern.FindStringSubmatch(content)

	if matches == nil {
		return content, ""
	}

	dueDate := matches[1]
	cleanContent := dueDatePattern.ReplaceAllString(content, " ")
	cleanContent = strings.TrimSpace(cleanContent)

	return cleanContent, dueDate
}

// ExtractTags extracts all freeform #tag markers from content
// Returns the content without tags and a slice of tag names
// Freeform tags are hashtags WITHOUT colons (e.g., #bug, #feature)
// Metadata tags WITH colons are NOT extracted (#p:1, #due:2026-02-15, #captured:2024-01-21)
func ExtractTags(content string) (string, []string) {
	// Pattern matches: #word followed by either : or non-word character or end of string
	tagPattern := regexp.MustCompile(`#([a-zA-Z0-9_-]+)(:?)`)

	// Find all matches
	matches := tagPattern.FindAllStringSubmatch(content, -1)

	var tags []string
	var freeformTags []string
	seen := make(map[string]bool)

	for _, match := range matches {
		tag := match[1]
		hasColon := match[2] == ":"

		// Only collect tags without colons (freeform tags)
		if !hasColon && !seen[tag] {
			tags = append(tags, tag)
			freeformTags = append(freeformTags, "#"+tag)
			seen[tag] = true
		}
	}

	// Remove only freeform tags from content
	cleanContent := content
	for _, freeformTag := range freeformTags {
		// Use word boundary to ensure we match whole tags
		cleanContent = strings.ReplaceAll(cleanContent, freeformTag, "")
	}

	// Clean up extra spaces
	cleanContent = regexp.MustCompile(`\s+`).ReplaceAllString(cleanContent, " ")
	cleanContent = strings.TrimSpace(cleanContent)

	return cleanContent, tags
}

// IsEmptyOrWhitespace checks if a string is empty or contains only whitespace
func IsEmptyOrWhitespace(s string) bool {
	return strings.TrimSpace(s) == ""
}
