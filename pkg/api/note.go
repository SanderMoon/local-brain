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

	// Read first few lines to get title and created date
	file, err := os.Open(filePath)
	if err != nil {
		return NoteFile{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

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

	// Format note content
	noteContent := fmt.Sprintf("# %s\n\nCreated: %s\n\n%s\n", title, timestamp, content)

	// Write note file
	if err := os.WriteFile(filePath, []byte(noteContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write note file: %w", err)
	}

	return filePath, nil
}

// ReadNoteFile reads the complete content of a note file
func ReadNoteFile(notePath string) (string, error) {
	content, err := os.ReadFile(notePath)
	if err != nil {
		return "", fmt.Errorf("failed to read note file: %w", err)
	}

	return string(content), nil
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
