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
