package api

import (
	"crypto/md5"
	"fmt"
)

// GenerateItemID generates a stable 6-character hex ID for a dump item
// This must match the bash implementation exactly for backward compatibility
//
// Algorithm from brain-api.sh lines 20-27:
//   1. Create hash input: "${line_num}:${content}:${mtime}"
//   2. Compute MD5 hash
//   3. Take first 6 hex characters
//
// Args:
//   - lineNum: Line number in the file (1-indexed)
//   - content: The full line content (for tasks) or title (for notes)
//   - mtime: Unix timestamp of file modification time (seconds)
//
// Returns: 6-character hex ID
func GenerateItemID(lineNum int, content string, mtime int64) string {
	// Format: line_num:content:mtime
	hashInput := fmt.Sprintf("%d:%s:%d", lineNum, content, mtime)

	// Compute MD5 hash
	hash := md5.Sum([]byte(hashInput))

	// Convert to hex string and take first 6 characters
	hexHash := fmt.Sprintf("%x", hash)
	return hexHash[:6]
}

// GenerateTaskID is a convenience wrapper for generating task IDs
// For tasks, the content is the entire line (including "- [ ] " prefix)
func GenerateTaskID(lineNum int, line string, mtime int64) string {
	return GenerateItemID(lineNum, line, mtime)
}

// GenerateNoteID is a convenience wrapper for generating note IDs
// For notes, the content is just the note title (from the [Note] header)
func GenerateNoteID(startLine int, title string, mtime int64) string {
	return GenerateItemID(startLine, title, mtime)
}
