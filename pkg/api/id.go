package api

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
)

var idPattern = regexp.MustCompile(`#id:([a-f0-9]{6})`)

// ExtractID extracts an existing #id: tag from a line
// Returns the ID if found, empty string otherwise
func ExtractID(line string) string {
	matches := idPattern.FindStringSubmatch(line)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// GenerateNewID generates a new random 6-character hex ID
func GenerateNewID() string {
	bytes := make([]byte, 3) // 3 bytes = 6 hex chars
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to a simple timestamp-based ID if random fails
		return fmt.Sprintf("%06x", (^uint32(0) >> 1))[:6]
	}
	return hex.EncodeToString(bytes)
}

// GetOrGenerateID gets existing ID from line or generates a new one
// This ensures IDs are stable and don't change when content is modified
func GetOrGenerateID(line string) string {
	// First, try to extract existing ID
	if existingID := ExtractID(line); existingID != "" {
		return existingID
	}

	// No ID found, generate a new one
	return GenerateNewID()
}

// AddIDToLine adds an #id: tag to a line if it doesn't already have one
// Returns the modified line and the ID
func AddIDToLine(line string) (string, string) {
	// Check if line already has an ID
	if existingID := ExtractID(line); existingID != "" {
		return line, existingID
	}

	// Generate new ID and append to line
	newID := GenerateNewID()
	return line + " #id:" + newID, newID
}

// RemoveIDFromContent removes the #id: tag from content for display
// This is used when showing clean content to users
func RemoveIDFromContent(content string) string {
	return idPattern.ReplaceAllString(content, "")
}
