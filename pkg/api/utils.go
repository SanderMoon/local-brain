package api

import (
	"regexp"
	"strings"
)

// Slugify converts text to a filename-safe slug
// Rules:
// - Converts to lowercase
// - Replaces spaces with hyphens
// - Keeps only alphanumeric characters and hyphens
// - Removes multiple consecutive hyphens
// - Trims hyphens from edges
// - Limits to 40 characters
func Slugify(text string) string {
	// Take first 40 characters
	if len(text) > 40 {
		text = text[:40]
	}

	// Convert to lowercase
	text = strings.ToLower(text)

	// Replace spaces with hyphens
	text = strings.ReplaceAll(text, " ", "-")

	// Keep only alphanumeric and hyphens
	reg := regexp.MustCompile(`[^a-z0-9-]+`)
	text = reg.ReplaceAllString(text, "")

	// Replace multiple hyphens with single hyphen
	reg = regexp.MustCompile(`-+`)
	text = reg.ReplaceAllString(text, "-")

	// Trim hyphens from edges
	text = strings.Trim(text, "-")

	return text
}

// ExtractCapturedDate extracts the #captured:YYYY-MM-DD tag from content
// Returns clean content (without tag) and the captured date string
// If no captured date found, returns original content and empty string
func ExtractCapturedDate(content string) (string, string) {
	pattern := regexp.MustCompile(`\s*#captured:([0-9-]+)`)
	matches := pattern.FindStringSubmatch(content)

	if len(matches) > 1 {
		cleanContent := pattern.ReplaceAllString(content, "")
		return strings.TrimSpace(cleanContent), matches[1]
	}

	return content, ""
}
