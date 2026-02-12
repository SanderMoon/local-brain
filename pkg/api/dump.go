package api

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sandermoonemans/local-brain/pkg/fileutil"
	"github.com/sandermoonemans/local-brain/pkg/markdown"
)

// DumpItemJSON represents a dump item in JSON format
// This matches the JSON schema from brain-api.sh dump_to_json (lines 98-105)
type DumpItemJSON struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

// ParseDumpToJSON parses a dump file and returns JSON array of items
// This replicates the combination of parse_dump_items + dump_to_json from brain-api.sh
func ParseDumpToJSON(filePath string) ([]DumpItemJSON, error) {
	// Parse dump file
	items, err := markdown.ParseDumpFile(filePath)
	if err != nil {
		return nil, err
	}

	// Convert to JSON format
	jsonItems := make([]DumpItemJSON, 0, len(items))

	for _, item := range items {
		// Extract timestamp from content
		cleanContent, timestamp := markdown.ExtractTimestamp(item.Content)

		// Get or generate stable ID from the raw line
		id := GetOrGenerateID(item.RawLine)

		// Remove #id: tag from content for display
		cleanContent = RemoveIDFromContent(cleanContent)

		jsonItems = append(jsonItems, DumpItemJSON{
			ID:        id,
			Content:   cleanContent,
			Type:      string(item.Type),
			Timestamp: timestamp,
			StartLine: item.StartLine,
			EndLine:   item.EndLine,
		})
	}

	return jsonItems, nil
}

// ParseDumpToJSONBytes returns JSON array as bytes
func ParseDumpToJSONBytes(filePath string) ([]byte, error) {
	items, err := ParseDumpToJSON(filePath)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(items, "", "  ")
}

// ParseDumpToJSONString returns JSON array as string
func ParseDumpToJSONString(filePath string) (string, error) {
	bytes, err := ParseDumpToJSONBytes(filePath)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// AddTaskToDump appends a task to the dump file with timestamp and ID
// Uses file locking for thread-safe append operations
// Format: - [ ] {content} #id:{newID} #captured:{timestamp}
// Note: timestamp must be at end for ExtractTimestamp to work correctly
func AddTaskToDump(dumpPath, content, timestamp string) error {
	// Acquire lock and append
	err := fileutil.WithLock(dumpPath, func() error {
		f, err := os.OpenFile(dumpPath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open dump file: %w", err)
		}
		defer f.Close()

		// Create task line with HTML comment containing id and captured date
		newID := GenerateNewID()
		comment := BuildSystemComment(newID, timestamp, "")
		line := fmt.Sprintf("- [ ] %s %s\n", content, comment)

		if _, err := f.WriteString(line); err != nil {
			return fmt.Errorf("failed to write to dump: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to add task to dump: %w", err)
	}

	return nil
}

// AddNoteToDump appends a note with header and indented content to the dump file
// Uses file locking for thread-safe append operations
// Format:
//
//	[Note] {title} #id:{newID} #captured:{timestamp}
//	    {line1}
//	    {line2}
//
// Note: timestamp must be at end for ExtractTimestamp to work correctly
func AddNoteToDump(dumpPath, title string, contentLines []string, timestamp string) error {
	// Acquire lock and append
	err := fileutil.WithLock(dumpPath, func() error {
		f, err := os.OpenFile(dumpPath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open dump file: %w", err)
		}
		defer f.Close()

		// Write note header with HTML comment containing id and captured date
		newID := GenerateNewID()
		comment := BuildSystemComment(newID, timestamp, "")
		header := fmt.Sprintf("[Note] %s %s\n", title, comment)

		if _, err := f.WriteString(header); err != nil {
			return fmt.Errorf("failed to write note header: %w", err)
		}

		// Write indented content
		for _, line := range contentLines {
			indentedLine := fmt.Sprintf("    %s\n", line)
			if _, err := f.WriteString(indentedLine); err != nil {
				return fmt.Errorf("failed to write note content: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to add note to dump: %w", err)
	}

	return nil
}

// RemoveItemFromDump removes lines from the dump file by line range
// Uses atomic file write to ensure file integrity
// Line numbers are 1-indexed (startLine and endLine inclusive)
func RemoveItemFromDump(dumpPath string, startLine, endLine int) error {
	// Read entire file
	content, err := os.ReadFile(dumpPath)
	if err != nil {
		return fmt.Errorf("failed to read dump file: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	// Validate line numbers
	if startLine < 1 || endLine < startLine || startLine > len(lines) {
		return fmt.Errorf("invalid line range: start=%d, end=%d, total lines=%d", startLine, endLine, len(lines))
	}

	// Remove lines (1-indexed to 0-indexed conversion)
	var newLines []string
	for i, line := range lines {
		lineNum := i + 1
		if lineNum < startLine || lineNum > endLine {
			newLines = append(newLines, line)
		}
	}

	// Write back atomically
	newContent := strings.Join(newLines, "\n")
	if err := fileutil.AtomicWriteFile(dumpPath, []byte(newContent)); err != nil {
		return fmt.Errorf("failed to write updated dump: %w", err)
	}

	return nil
}

// FindDumpItemByID searches for a dump item by its ID
// Returns pointer to the matching item or nil if not found
func FindDumpItemByID(dumpPath, itemID string) (*markdown.DumpItem, error) {
	// Parse dump file
	items, err := markdown.ParseDumpFile(dumpPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dump: %w", err)
	}

	// Search for item with matching ID
	for i := range items {
		item := &items[i]
		// Extract ID from raw line
		id := GetOrGenerateID(item.RawLine)

		if id == itemID {
			return item, nil
		}
	}

	return nil, nil
}
