package api

import (
	"encoding/json"
	"os"

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
	// Get file modification time for ID generation
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	mtime := fileInfo.ModTime().Unix()

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

		// Generate ID based on item type
		var id string
		if item.Type == markdown.ItemTypeTodo {
			// For tasks, use full line content for ID (including "- [ ] ")
			id = GenerateTaskID(item.StartLine, item.RawLine, mtime)
		} else {
			// For notes, use start line and title
			id = GenerateNoteID(item.StartLine, item.RawLine, mtime)
		}

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
