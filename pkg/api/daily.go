package api

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sandermoonemans/local-brain/pkg/fileutil"
)

// DailyNote represents a created or existing daily note
type DailyNote struct {
	Path  string `json:"path"`
	Date  string `json:"date"`
	IsNew bool   `json:"is_new"`
}

// CreateOrOpenDailyNote creates {brainPath}/00_daily/YYYY-MM-DD.md if it doesn't exist.
// If the file already exists, returns it unchanged with IsNew=false.
// overdueTodos are pre-fetched and included in the briefing section.
func CreateOrOpenDailyNote(brainPath, date string, overdueTodos []TodoItem) (DailyNote, error) {
	dailyDir := filepath.Join(brainPath, "00_daily")

	if err := fileutil.EnsureDir(dailyDir); err != nil {
		return DailyNote{}, fmt.Errorf("failed to create daily directory: %w", err)
	}

	filePath := filepath.Join(dailyDir, date+".md")

	if fileutil.FileExists(filePath) {
		return DailyNote{Path: filePath, Date: date, IsNew: false}, nil
	}

	// Build the overdue todos section
	var overdueLines strings.Builder
	if len(overdueTodos) == 0 {
		overdueLines.WriteString("(no overdue items)")
	} else {
		for _, todo := range overdueTodos {
			overdueLines.WriteString(fmt.Sprintf("- [ ] [%s] %s (due: %s)\n", todo.Project, todo.Content, todo.DueDate))
		}
	}

	content := fmt.Sprintf(`---
title: %s
date: %s
---

# %s

## Daily Briefing

%s

## Today's Focus

-

## Notes

`, date, date, date, strings.TrimRight(overdueLines.String(), "\n"))

	if err := fileutil.AtomicWriteFile(filePath, []byte(content)); err != nil {
		return DailyNote{}, fmt.Errorf("failed to write daily note: %w", err)
	}

	return DailyNote{Path: filePath, Date: date, IsNew: true}, nil
}
