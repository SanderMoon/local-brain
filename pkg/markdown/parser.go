package markdown

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

// ExtractHTMLComment parses an HTML comment at the end of a line in the format
// <!-- key:val key2:val2 -->
// Returns a map of key->value and the line with the comment removed.
// Returns empty map and original line if no comment found.
func ExtractHTMLComment(line string) (map[string]string, string) {
	// pattern: <!-- ... -->
	// The comment may appear anywhere in the line but typically at the end
	commentPattern := regexp.MustCompile(`<!--\s*(.*?)\s*-->`)
	match := commentPattern.FindStringSubmatchIndex(line)
	if match == nil {
		return map[string]string{}, line
	}

	commentContent := line[match[2]:match[3]]
	lineWithoutComment := strings.TrimSpace(line[:match[0]] + line[match[1]:])

	// Parse key:value pairs from comment content
	fields := make(map[string]string)
	kvPattern := regexp.MustCompile(`(\w+):(\S+)`)
	for _, kv := range kvPattern.FindAllStringSubmatch(commentContent, -1) {
		fields[kv[1]] = kv[2]
	}
	return fields, lineWithoutComment
}

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

// ExtractTimestamp extracts the captured timestamp from content.
// First tries HTML comment format <!-- captured:YYYY-MM-DD -->, then falls back to inline #captured:YYYY-MM-DD.
// Returns the content without timestamp and the timestamp string.
func ExtractTimestamp(content string) (string, string) {
	// Try HTML comment format first
	fields, lineWithoutComment := ExtractHTMLComment(content)
	if capturedDate, ok := fields["captured"]; ok {
		return lineWithoutComment, capturedDate
	}

	// Fall back to inline #captured: format
	timestampPattern := regexp.MustCompile(`\s*#captured:([0-9-]+)(?:\s|$)`)
	matches := timestampPattern.FindStringSubmatch(content)

	if matches == nil {
		return content, ""
	}

	timestamp := matches[1]
	cleanContent := timestampPattern.ReplaceAllString(content, "")
	cleanContent = strings.TrimSpace(cleanContent)

	return cleanContent, timestamp
}

// ExtractCompletedDate extracts the done (completion) date from content.
// First tries HTML comment format <!-- done:YYYY-MM-DD -->, then falls back to inline #done:YYYY-MM-DD.
// Returns the content without the done date and the date string.
func ExtractCompletedDate(content string) (string, string) {
	// Try HTML comment format first
	fields, lineWithoutComment := ExtractHTMLComment(content)
	if doneDate, ok := fields["done"]; ok {
		return lineWithoutComment, doneDate
	}

	// Fall back to inline #done: format
	donePattern := regexp.MustCompile(`\s*#done:([0-9-]+)(?:\s|$)`)
	matches := donePattern.FindStringSubmatch(content)

	if matches == nil {
		return content, ""
	}

	doneDate := matches[1]
	cleanContent := donePattern.ReplaceAllString(content, "")
	cleanContent = strings.TrimSpace(cleanContent)

	return cleanContent, doneDate
}

// ExtractPriority extracts the priority tag from content.
// Tries new format p:[1-3] (no hash) first, then falls back to legacy #p:[1-3].
// The no-hash format requires p: to appear at start of string or after whitespace (not after #).
// Returns the content without priority tag and the priority value (1=high, 2=medium, 3=low)
// Returns nil priority if no valid tag is found
func ExtractPriority(content string) (string, *int) {
	// Try new format p:[1-3] first (no hash prefix).
	// Require p: to be at start of string or preceded by whitespace to avoid matching #p:.
	priorityPattern := regexp.MustCompile(`(^|\s)p:([1-3])(\s|$)`)
	matches := priorityPattern.FindStringSubmatch(content)
	if matches != nil {
		priorityStr := matches[2]
		priority := int(priorityStr[0] - '0')
		// Replace the full match but preserve any leading/trailing whitespace context.
		cleanContent := priorityPattern.ReplaceAllStringFunc(content, func(m string) string {
			// Preserve any surrounding whitespace that was part of the match.
			leading := ""
			trailing := ""
			if len(m) > 0 && (m[0] == ' ' || m[0] == '\t' || m[0] == '\n') {
				leading = string(m[0])
			}
			if len(m) > 0 && (m[len(m)-1] == ' ' || m[len(m)-1] == '\t' || m[len(m)-1] == '\n') {
				trailing = string(m[len(m)-1])
			}
			return leading + trailing
		})
		cleanContent = strings.TrimSpace(cleanContent)
		// Clean up multiple consecutive spaces that might have been introduced.
		cleanContent = regexp.MustCompile(`  +`).ReplaceAllString(cleanContent, " ")
		return cleanContent, &priority
	}

	// Fall back to legacy #p:[1-3] format
	legacyPriorityPattern := regexp.MustCompile(`\s*#p:([1-3])(?:\s|$)`)
	matches = legacyPriorityPattern.FindStringSubmatch(content)
	if matches == nil {
		return content, nil
	}

	priorityStr := matches[1]
	priority := int(priorityStr[0] - '0')
	cleanContent := legacyPriorityPattern.ReplaceAllString(content, " ")
	cleanContent = strings.TrimSpace(cleanContent)

	return cleanContent, &priority
}

// ExtractDueDate extracts the due date tag from content.
// Tries new format due:YYYY-MM-DD (no hash) first, then falls back to legacy #due:YYYY-MM-DD.
// The no-hash format requires due: to appear at start of string or after whitespace (not after #).
// Returns the content without due date tag and the due date string (YYYY-MM-DD or whatever was in the tag)
// Returns empty string if no valid tag is found
func ExtractDueDate(content string) (string, string) {
	// Try new format due:... first (no hash prefix).
	// Require due: to be at start of string or preceded by whitespace to avoid matching #due:.
	dueDatePattern := regexp.MustCompile(`(^|\s)due:([^\s]+)(\s|$)`)
	matches := dueDatePattern.FindStringSubmatch(content)
	if matches != nil {
		dueDate := matches[2]
		// Replace the full match but preserve any surrounding whitespace context.
		cleanContent := dueDatePattern.ReplaceAllStringFunc(content, func(m string) string {
			leading := ""
			trailing := ""
			if len(m) > 0 && (m[0] == ' ' || m[0] == '\t' || m[0] == '\n') {
				leading = string(m[0])
			}
			if len(m) > 0 && (m[len(m)-1] == ' ' || m[len(m)-1] == '\t' || m[len(m)-1] == '\n') {
				trailing = string(m[len(m)-1])
			}
			return leading + trailing
		})
		cleanContent = strings.TrimSpace(cleanContent)
		// Clean up multiple consecutive spaces that might have been introduced.
		cleanContent = regexp.MustCompile(`  +`).ReplaceAllString(cleanContent, " ")
		return cleanContent, dueDate
	}

	// Fall back to legacy #due:... format
	legacyDueDatePattern := regexp.MustCompile(`\s*#due:([^\s]+)(?:\s|$)`)
	matches = legacyDueDatePattern.FindStringSubmatch(content)
	if matches == nil {
		return content, ""
	}

	dueDate := matches[1]
	cleanContent := legacyDueDatePattern.ReplaceAllString(content, " ")
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
