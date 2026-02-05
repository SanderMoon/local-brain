package api

import (
	"fmt"
	"path/filepath"

	"github.com/sandermoonemans/local-brain/pkg/config"
)

// BrainOverview provides complete context about the current brain state
type BrainOverview struct {
	CurrentBrain   string        `json:"current_brain"`
	BrainPath      string        `json:"brain_path"`
	FocusedProject string        `json:"focused_project"`
	Projects       []ProjectInfo `json:"projects"`
	DumpItemCount  int           `json:"dump_item_count"`
}

// GetBrainOverview returns comprehensive brain context in a single call
// This is optimized for MCP server to minimize round-trips
func GetBrainOverview(cfg *config.Config) (*BrainOverview, error) {
	// Get current brain name
	currentBrain := cfg.GetCurrentBrain()
	if currentBrain == "" {
		return nil, fmt.Errorf("no active brain")
	}

	// Get brain path
	brainPath, err := cfg.GetCurrentBrainPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get brain path: %w", err)
	}

	// Get focused project
	focusedProject := cfg.GetFocusedProject()

	// Get all projects
	activeDir := filepath.Join(brainPath, "01_active")
	projects, err := ListProjects(activeDir, focusedProject)
	if err != nil {
		// If directory doesn't exist yet, return empty list
		projects = []ProjectInfo{}
	}

	// Count dump items
	dumpPath := filepath.Join(brainPath, "00_dump.md")
	dumpItems, err := ParseDumpToJSON(dumpPath)
	dumpItemCount := 0
	if err == nil {
		dumpItemCount = len(dumpItems)
	}

	return &BrainOverview{
		CurrentBrain:   currentBrain,
		BrainPath:      brainPath,
		FocusedProject: focusedProject,
		Projects:       projects,
		DumpItemCount:  dumpItemCount,
	}, nil
}
