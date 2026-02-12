package api

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sandermoonemans/local-brain/pkg/fileutil"
)

// MigrateNoteResult holds the result of migrating a single note file
type MigrateNoteResult struct {
	File    string
	Changed bool
	Change  string // description of change, or "already has frontmatter", or "error: ..."
}

// MigrateTodoResult holds the result of migrating a single todo file
type MigrateTodoResult struct {
	File    string
	Changed bool
	Change  string
}

// MigrateLinkResult holds the result of migrating a notes.md index file
type MigrateLinkResult struct {
	File    string // path to notes.md
	Changed bool
	Change  string // e.g., "added 3 links"
}

// ProjectMigrateResult holds all migration results for a project
type ProjectMigrateResult struct {
	ProjectName string
	Notes       []MigrateNoteResult
	Todos       []MigrateTodoResult
	Links       []MigrateLinkResult
}

var (
	createdDatePattern = regexp.MustCompile(`Created:\s*(\d{4}-\d{2}-\d{2})`)
	inlineIDPattern    = regexp.MustCompile(`#id:([a-f0-9]{6})`)
	inlineCaptured     = regexp.MustCompile(`\s*#captured:([0-9-]+)(?:\s|$)`)
	inlineDone         = regexp.MustCompile(`\s*#done:([0-9-]+)(?:\s|$)`)
	todoLinePattern    = regexp.MustCompile(`^\s*- \[[>\-xX ]\]`)
)

// MigrateNoteToFrontmatter converts a legacy note (with # Title + Created: date body text)
// to one with YAML frontmatter. If dryRun is true, no file is written.
func MigrateNoteToFrontmatter(filePath string, dryRun bool) (MigrateNoteResult, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return MigrateNoteResult{File: filePath, Changed: false, Change: fmt.Sprintf("error: %v", err)}, err
	}

	contentStr := string(content)

	// Check if already has frontmatter
	if _, _, _, ok := parseFrontmatter(contentStr); ok {
		return MigrateNoteResult{File: filePath, Changed: false, Change: "already has frontmatter"}, nil
	}

	// Parse legacy title: first line starting with "# "
	title := ""
	lines := strings.Split(contentStr, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			title = strings.TrimPrefix(line, "# ")
			break
		}
	}

	// Parse legacy date: scan for "Created: YYYY-MM-DD"
	date := ""
	if matches := createdDatePattern.FindStringSubmatch(contentStr); matches != nil {
		date = matches[1]
	}

	// Derive project from directory structure: note is at projectDir/notes/file.md
	project := filepath.Base(filepath.Dir(filepath.Dir(filePath)))

	if dryRun {
		return MigrateNoteResult{File: filePath, Changed: true, Change: "would add frontmatter"}, nil
	}

	// Build frontmatter + existing content
	newContent := fmt.Sprintf("---\ntitle: %s\ndate: %s\nproject: %s\ntags: []\n---\n\n%s",
		title, date, project, contentStr)

	if err := fileutil.AtomicWriteFile(filePath, []byte(newContent)); err != nil {
		return MigrateNoteResult{File: filePath, Changed: false, Change: fmt.Sprintf("error: %v", err)}, err
	}

	return MigrateNoteResult{File: filePath, Changed: true, Change: "added frontmatter"}, nil
}

// MigrateTodoToHTMLComments moves inline #id:, #captured:, #done: system tags
// into HTML comments for each todo line. If dryRun is true, no file is written.
func MigrateTodoToHTMLComments(filePath string, dryRun bool) (MigrateTodoResult, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return MigrateTodoResult{File: filePath, Changed: false, Change: fmt.Sprintf("error: %v", err)}, err
	}

	lines := strings.Split(string(content), "\n")
	count := 0
	newLines := make([]string, len(lines))
	copy(newLines, lines)

	for i, line := range lines {
		// Only process todo lines
		if !todoLinePattern.MatchString(line) {
			continue
		}

		// If it already has an HTML comment with an id, it's already migrated — skip
		if ExtractIDFromComment(line) != "" {
			continue
		}

		// Check if this line has any inline metadata that needs migration
		idMatch := inlineIDPattern.FindStringSubmatch(line)
		capturedMatch := inlineCaptured.FindStringSubmatch(line)
		doneMatch := inlineDone.FindStringSubmatch(line)

		if idMatch == nil && capturedMatch == nil && doneMatch == nil {
			// Nothing to migrate on this line
			continue
		}

		// Extract values
		id := ""
		if idMatch != nil {
			id = idMatch[1]
		}
		captured := ""
		if capturedMatch != nil {
			captured = capturedMatch[1]
		}
		done := ""
		if doneMatch != nil {
			done = doneMatch[1]
		}

		// Remove inline tags from the line
		cleaned := line
		if idMatch != nil {
			cleaned = inlineIDPattern.ReplaceAllString(cleaned, "")
		}
		if capturedMatch != nil {
			cleaned = inlineCaptured.ReplaceAllString(cleaned, " ")
		}
		if doneMatch != nil {
			cleaned = inlineDone.ReplaceAllString(cleaned, " ")
		}
		// Clean up multiple trailing spaces but preserve leading whitespace structure
		cleaned = strings.TrimRight(cleaned, " \t")

		// Build HTML comment
		comment := BuildSystemComment(id, captured, done)
		if comment != "" {
			cleaned = cleaned + " " + comment
		}

		newLines[i] = cleaned
		count++
	}

	if count == 0 {
		return MigrateTodoResult{File: filePath, Changed: false, Change: "no inline metadata found"}, nil
	}

	if dryRun {
		return MigrateTodoResult{File: filePath, Changed: true, Change: fmt.Sprintf("would move %d items to HTML comments", count)}, nil
	}

	newContent := strings.Join(newLines, "\n")
	if err := fileutil.AtomicWriteFile(filePath, []byte(newContent)); err != nil {
		return MigrateTodoResult{File: filePath, Changed: false, Change: fmt.Sprintf("error: %v", err)}, err
	}

	return MigrateTodoResult{File: filePath, Changed: true, Change: fmt.Sprintf("moved %d items to HTML comments", count)}, nil
}

// MigrateNotesIndex appends missing relative markdown links in notes.md for note
// files in the notes/ directory that are not yet linked. If dryRun is true, no file is written.
func MigrateNotesIndex(projectDir string, dryRun bool) (MigrateLinkResult, error) {
	notesDir := filepath.Join(projectDir, "notes")
	notesIndexPath := filepath.Join(projectDir, "notes.md")

	// Find all *.md files in notes/
	pattern := filepath.Join(notesDir, "*.md")
	noteFiles, err := filepath.Glob(pattern)
	if err != nil {
		return MigrateLinkResult{File: notesIndexPath}, fmt.Errorf("failed to glob notes: %w", err)
	}

	if len(noteFiles) == 0 {
		return MigrateLinkResult{File: notesIndexPath, Changed: false, Change: "no note files found"}, nil
	}

	// Read notes.md (treat missing file as empty)
	indexContent := ""
	rawBytes, err := os.ReadFile(notesIndexPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return MigrateLinkResult{File: notesIndexPath}, fmt.Errorf("failed to read notes.md: %w", err)
		}
	} else {
		indexContent = string(rawBytes)
	}

	// Count how many note files are not yet linked
	var unlinked []string
	for _, filePath := range noteFiles {
		filename := filepath.Base(filePath)
		if !strings.Contains(indexContent, "notes/"+filename) {
			unlinked = append(unlinked, filePath)
		}
	}

	if len(unlinked) == 0 {
		return MigrateLinkResult{File: notesIndexPath, Changed: false, Change: "all notes already linked"}, nil
	}

	if dryRun {
		return MigrateLinkResult{File: notesIndexPath, Changed: true, Change: fmt.Sprintf("would add %d links", len(unlinked))}, nil
	}

	// Append links for each unlinked note
	projectName := filepath.Base(projectDir)
	for _, filePath := range unlinked {
		note, err := parseNoteFile(filePath, projectName)
		if err != nil {
			// Best effort: use filename as fallback
			note = NoteFile{
				Filename: filepath.Base(filePath),
				Title:    filepath.Base(filePath),
				Created:  "",
			}
		}
		if appendErr := AppendNoteLink(projectDir, note.Created, note.Title, note.Filename); appendErr != nil {
			return MigrateLinkResult{File: notesIndexPath}, fmt.Errorf("failed to append note link: %w", appendErr)
		}
	}

	return MigrateLinkResult{File: notesIndexPath, Changed: true, Change: fmt.Sprintf("added %d links", len(unlinked))}, nil
}

// MigrateProject runs all three migration operations on a single project directory.
func MigrateProject(projectDir string, dryRun bool) (ProjectMigrateResult, error) {
	projectName := filepath.Base(projectDir)
	result := ProjectMigrateResult{ProjectName: projectName}

	// Migrate notes in notes/ subdirectory
	notesDir := filepath.Join(projectDir, "notes")
	if _, err := os.Stat(notesDir); err == nil {
		pattern := filepath.Join(notesDir, "*.md")
		noteFiles, err := filepath.Glob(pattern)
		if err != nil {
			return result, fmt.Errorf("failed to glob notes: %w", err)
		}
		for _, filePath := range noteFiles {
			noteResult, err := MigrateNoteToFrontmatter(filePath, dryRun)
			if err != nil {
				noteResult.Change = fmt.Sprintf("error: %v", err)
			}
			result.Notes = append(result.Notes, noteResult)
		}
	}

	// Migrate todo.md if it exists
	todoFile := filepath.Join(projectDir, "todo.md")
	if fileutil.FileExists(todoFile) {
		todoResult, err := MigrateTodoToHTMLComments(todoFile, dryRun)
		if err != nil {
			todoResult.Change = fmt.Sprintf("error: %v", err)
		}
		result.Todos = append(result.Todos, todoResult)
	}

	// Migrate notes.md index
	linkResult, err := MigrateNotesIndex(projectDir, dryRun)
	if err != nil {
		linkResult.Change = fmt.Sprintf("error: %v", err)
	}
	result.Links = append(result.Links, linkResult)

	return result, nil
}

// MigrateAllProjects runs migration on all project directories in a section.
func MigrateAllProjects(activeDir string, dryRun bool) ([]ProjectMigrateResult, error) {
	entries, err := os.ReadDir(activeDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var results []ProjectMigrateResult
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		projectDir := filepath.Join(activeDir, entry.Name())
		projectResult, err := MigrateProject(projectDir, dryRun)
		if err != nil {
			return results, fmt.Errorf("failed to migrate project %s: %w", entry.Name(), err)
		}
		results = append(results, projectResult)
	}

	return results, nil
}
