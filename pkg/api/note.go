package api

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/sandermoonemans/local-brain/pkg/fileutil"
)

// parseFrontmatter parses a YAML frontmatter block from the start of content.
// The frontmatter must be delimited by "---" on its own line.
// Returns title, date, project fields and ok=true if frontmatter was found and parsed.
// Returns ok=false if content doesn't start with "---" or has no closing "---".
func parseFrontmatter(content string) (title, date, project string, ok bool) {
	if !strings.HasPrefix(content, "---\n") && content != "---" {
		return "", "", "", false
	}

	// Skip the opening "---\n"
	rest := content[4:]

	// Find the closing "---"
	closingIdx := strings.Index(rest, "\n---\n")
	if closingIdx == -1 {
		// Check if "---" is at end of content
		if strings.HasSuffix(rest, "\n---") {
			closingIdx = len(rest) - 4
		} else {
			return "", "", "", false
		}
	}

	// Extract the body up to (but not including) the closing "---"
	// closingIdx points to the "\n" before "---", so +1 includes that newline
	frontmatterBody := rest[:closingIdx+1]

	for _, line := range strings.Split(frontmatterBody, "\n") {
		if strings.HasPrefix(line, "title:") {
			title = strings.TrimSpace(strings.TrimPrefix(line, "title:"))
		} else if strings.HasPrefix(line, "date:") {
			date = strings.TrimSpace(strings.TrimPrefix(line, "date:"))
		} else if strings.HasPrefix(line, "project:") {
			project = strings.TrimSpace(strings.TrimPrefix(line, "project:"))
		}
	}

	return title, date, project, true
}

// NoteFile represents a note file in a project
type NoteFile struct {
	Filename string    `json:"filename"`
	Path     string    `json:"path"`
	Title    string    `json:"title"`
	Created  string    `json:"created"`
	Project  string    `json:"project"`
	ModTime  time.Time `json:"-"`
}

// ListNotes returns all notes in a project's notes directory
func ListNotes(projectDir string) ([]NoteFile, error) {
	notesDir := filepath.Join(projectDir, "notes")

	// Check if notes directory exists
	if _, err := os.Stat(notesDir); os.IsNotExist(err) {
		return []NoteFile{}, nil
	}

	// Find all .md files
	pattern := filepath.Join(notesDir, "*.md")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to glob notes: %w", err)
	}

	var notes []NoteFile
	projectName := filepath.Base(projectDir)

	for _, filePath := range files {
		note, err := parseNoteFile(filePath, projectName)
		if err != nil {
			// Skip files that can't be parsed
			continue
		}
		notes = append(notes, note)
	}

	// Sort by modification time (newest first)
	sort.Slice(notes, func(i, j int) bool {
		return notes[i].ModTime.After(notes[j].ModTime)
	})

	return notes, nil
}

func parseNoteFile(filePath, projectName string) (NoteFile, error) {
	filename := filepath.Base(filePath)

	// Get file info for modification time
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return NoteFile{}, err
	}

	// Read full file content
	rawBytes, err := os.ReadFile(filePath)
	if err != nil {
		return NoteFile{}, err
	}
	rawContent := string(rawBytes)

	// Try frontmatter first
	if fmTitle, fmDate, _, ok := parseFrontmatter(rawContent); ok {
		return NoteFile{
			Filename: filename,
			Path:     filePath,
			Title:    fmTitle,
			Created:  fmDate,
			Project:  projectName,
			ModTime:  fileInfo.ModTime(),
		}, nil
	}

	// Legacy fallback: scan line-by-line for "# Title" and "Created: YYYY-MM-DD"
	scanner := bufio.NewScanner(strings.NewReader(rawContent))

	// Read title (first line, should be "# Title")
	title := ""
	if scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "# ") {
			title = strings.TrimPrefix(line, "# ")
		} else {
			title = line
		}
	}

	// Read created date (look for "Created: YYYY-MM-DD")
	created := ""
	createdPattern := regexp.MustCompile(`Created:\s*(\d{4}-\d{2}-\d{2})`)

	for scanner.Scan() && created == "" {
		line := scanner.Text()
		if matches := createdPattern.FindStringSubmatch(line); matches != nil {
			created = matches[1]
		}
	}

	return NoteFile{
		Filename: filename,
		Path:     filePath,
		Title:    title,
		Created:  created,
		Project:  projectName,
		ModTime:  fileInfo.ModTime(),
	}, nil
}

// DeleteNote removes a note file
func DeleteNote(notePath string) error {
	return os.Remove(notePath)
}

// CreateNoteFile creates a timestamped note file in a project's notes/ directory
// Returns the path to the created file
func CreateNoteFile(projectDir, title, content, timestamp string) (string, error) {
	notesDir := filepath.Join(projectDir, "notes")

	// Ensure notes/ directory exists
	if err := fileutil.EnsureDir(notesDir); err != nil {
		return "", fmt.Errorf("failed to create notes directory: %w", err)
	}

	// Generate slug from title
	slug := Slugify(title)
	if slug == "" {
		slug = "note"
	}

	// Create filename
	filename := fmt.Sprintf("%s-%s.md", timestamp, slug)
	filePath := filepath.Join(notesDir, filename)

	// Handle duplicate filenames
	counter := 1
	for fileutil.FileExists(filePath) {
		filename = fmt.Sprintf("%s-%s-%d.md", timestamp, slug, counter)
		filePath = filepath.Join(notesDir, filename)
		counter++
	}

	// Format note content with YAML frontmatter
	noteContent := fmt.Sprintf("---\ntitle: %s\ndate: %s\nproject: %s\ntags: []\n---\n\n# %s\n\n%s\n",
		title, timestamp, filepath.Base(projectDir), title, content)

	// Write note file
	if err := os.WriteFile(filePath, []byte(noteContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write note file: %w", err)
	}

	// Best-effort: try to update notes.md index, but don't fail if it errors
	_ = AppendNoteLink(projectDir, timestamp, title, filepath.Base(filePath))

	return filePath, nil
}

// AppendNoteLink appends a relative markdown link to the project's notes.md index file.
// It is idempotent: if the link already exists, the file is left unchanged.
// The link is inserted as the first item inside the "## Notes" section, or the
// section is created at the end of the file if it does not yet exist.
func AppendNoteLink(projectDir, timestamp, title, filename string) error {
	notesIndexPath := filepath.Join(projectDir, "notes.md")

	// Read existing content; treat a missing file as empty.
	var content string
	rawBytes, err := os.ReadFile(notesIndexPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read notes.md: %w", err)
		}
		content = ""
	} else {
		content = string(rawBytes)
	}

	// Idempotency check.
	if strings.Contains(content, "notes/"+filename) {
		return nil
	}

	linkLine := fmt.Sprintf("- [%s %s](notes/%s)", timestamp, title, filename)

	var newContent string
	const section = "## Notes"
	if idx := strings.Index(content, section+"\n"); idx != -1 {
		// Insert the link right after "## Notes\n", before any existing entries.
		insertAt := idx + len(section) + 1 // position right after the "\n"
		// Skip a single blank line that immediately follows the header.
		if insertAt < len(content) && content[insertAt] == '\n' {
			insertAt++
		}
		newContent = content[:insertAt] + linkLine + "\n" + content[insertAt:]
	} else {
		// No "## Notes" section found: append it.
		newContent = content + "\n" + section + "\n\n" + linkLine + "\n"
	}

	return fileutil.AtomicWriteFile(notesIndexPath, []byte(newContent))
}

// ReadNoteFile reads the complete content of a note file
func ReadNoteFile(notePath string) (string, error) {
	content, err := os.ReadFile(notePath)
	if err != nil {
		return "", fmt.Errorf("failed to read note file: %w", err)
	}

	return string(content), nil
}

// UpdateNoteFile replaces the content of an existing note file atomically
func UpdateNoteFile(notePath, content string) error {
	if _, err := os.Stat(notePath); os.IsNotExist(err) {
		return fmt.Errorf("note file not found: %s", notePath)
	}
	return fileutil.AtomicWriteFile(notePath, []byte(content))
}

// NoteSearchResult represents a note matching search criteria
type NoteSearchResult struct {
	NoteFile
	MatchedContent string `json:"matched_content,omitempty"` // Preview of matching content
}

// SearchNotes searches notes by content query and project filter
func SearchNotes(activeDir string, query string, project string, searchContent bool) ([]NoteSearchResult, error) {
	var results []NoteSearchResult

	// Determine which projects to search
	var projectDirs []string
	if project != "" {
		projectDirs = []string{filepath.Join(activeDir, project)}
	} else {
		// Search all projects
		entries, err := os.ReadDir(activeDir)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				projectDirs = append(projectDirs, filepath.Join(activeDir, entry.Name()))
			}
		}
	}

	for _, pDir := range projectDirs {
		notes, err := ListNotes(pDir)
		if err != nil {
			continue // Skip projects with errors
		}

		for _, note := range notes {
			matched := false
			matchedContent := ""

			// Search in title
			if query == "" || strings.Contains(strings.ToLower(note.Title), strings.ToLower(query)) {
				matched = true
			}

			// Search in content if requested
			if searchContent && !matched {
				content, err := ReadNoteFile(note.Path)
				if err == nil {
					if strings.Contains(strings.ToLower(content), strings.ToLower(query)) {
						matched = true
						// Extract preview around match
						matchedContent = extractMatchPreview(content, query, 200)
					}
				}
			}

			if matched {
				results = append(results, NoteSearchResult{
					NoteFile:       note,
					MatchedContent: matchedContent,
				})
			}
		}
	}

	return results, nil
}

// extractMatchPreview extracts text around the query match
func extractMatchPreview(content, query string, maxLen int) string {
	lowerContent := strings.ToLower(content)
	lowerQuery := strings.ToLower(query)

	idx := strings.Index(lowerContent, lowerQuery)
	if idx == -1 {
		if len(content) > maxLen {
			return content[:maxLen] + "..."
		}
		return content
	}

	start := idx - 50
	if start < 0 {
		start = 0
	}

	end := idx + len(query) + 150
	if end > len(content) {
		end = len(content)
	}

	preview := content[start:end]
	if start > 0 {
		preview = "..." + preview
	}
	if end < len(content) {
		preview = preview + "..."
	}

	return preview
}
