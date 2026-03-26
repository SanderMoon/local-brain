package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sandermoonemans/local-brain/mcp/session"
	"github.com/sandermoonemans/local-brain/mcp/validation"
	"github.com/sandermoonemans/local-brain/pkg/api"
	"github.com/sandermoonemans/local-brain/pkg/config"
)

// RegisterContextTools registers context retrieval tools
func RegisterContextTools(srv *mcp.Server, sess *session.Session) error {
	// get_brain_overview - no args needed
	type GetBrainOverviewArgs struct{}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_brain_overview",
		Description: "Get complete overview of Local Brain workspace including current brain, focused project, all projects with stats, and dump item count",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetBrainOverviewArgs) (*mcp.CallToolResult, any, error) {
		overview, err := sess.GetBrainOverview()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get brain overview: %w", err)
		}

		// Convert to JSON string for text content
		data, err := json.MarshalIndent(overview, "", "  ")
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal overview: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(data)},
			},
		}, nil, nil
	})

	// get_dump_items
	type GetDumpItemsArgs struct{}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_dump_items",
		Description: "Get all items (tasks and notes) from the inbox (00_dump.md) with IDs, content, and timestamps",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetDumpItemsArgs) (*mcp.CallToolResult, any, error) {
		cfg := sess.GetConfig()
		brainPath, err := cfg.GetCurrentBrainPath()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get brain path: %w", err)
		}

		dumpPath := filepath.Join(brainPath, "00_dump.md")
		items, err := api.ParseDumpToJSON(dumpPath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse dump: %w", err)
		}

		data, err := json.MarshalIndent(items, "", "  ")
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal items: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(data)},
			},
		}, nil, nil
	})

	// get_project_context
	type GetProjectContextArgs struct {
		ProjectName        string `json:"project_name" jsonschema:"Name of the project"`
		IncludeCompleted   bool   `json:"include_completed" jsonschema:"Whether to include completed tasks (default: false)"`
		IncludeNoteContent string `json:"include_note_content,omitempty" jsonschema:"Note content: none|preview|full (default: preview)"`
		Section            string `json:"section,omitempty" jsonschema:"PARA section: 01_active (default), 02_areas, or 03_resources"`
	}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_project_context",
		Description: "Get complete project details including description, todos, linked repos, and note files (with optional note content)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetProjectContextArgs) (*mcp.CallToolResult, any, error) {
		// Validate inputs
		if err := validation.ValidateProjectName(args.ProjectName); err != nil {
			return nil, nil, err
		}

		// Default to preview if not specified
		if args.IncludeNoteContent == "" {
			args.IncludeNoteContent = "preview"
		}

		cfg := sess.GetConfig()
		sectionDir, err := config.GetSectionPath(cfg, args.Section)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get section path: %w", err)
		}

		projectPath := filepath.Join(sectionDir, args.ProjectName)
		focusedProject := cfg.GetFocusedProject()

		devDir := cfg.GetDevDir()
		projectCtx, err := api.GetProjectContext(projectPath, args.ProjectName, focusedProject, args.IncludeCompleted, args.IncludeNoteContent, devDir)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get project context: %w", err)
		}

		data, err := json.MarshalIndent(projectCtx, "", "  ")
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal project context: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(data)},
			},
		}, nil, nil
	})


	// search - unified search across todos and notes
	type SearchArgs struct {
		Query             string   `json:"query,omitempty" jsonschema:"Search query (optional)"`
		Project           string   `json:"project,omitempty" jsonschema:"Filter by project name (optional)"`
		IncludeTodos      bool     `json:"include_todos,omitempty" jsonschema:"Include todos in search (default: true)"`
		IncludeNotes      bool     `json:"include_notes,omitempty" jsonschema:"Include notes in search (default: true)"`
		SearchNoteContent bool     `json:"search_note_content,omitempty" jsonschema:"Search note content (not just titles)"`
		Status            string   `json:"status,omitempty" jsonschema:"Filter todos by status: open, in-progress, blocked, done (optional)"`
		Priority          *int     `json:"priority,omitempty" jsonschema:"Filter todos by priority: 1=high, 2=medium, 3=low (optional)"`
		Tags              []string `json:"tags,omitempty" jsonschema:"Filter todos by tags - OR logic, match any tag (optional)"`
		IncludeCompleted  bool     `json:"include_completed,omitempty" jsonschema:"Whether to include completed tasks (default: false)"`
		CreatedAfter      string   `json:"created_after,omitempty" jsonschema:"Filter by created date >= YYYY-MM-DD (optional)"`
		CreatedBefore     string   `json:"created_before,omitempty" jsonschema:"Filter by created date <= YYYY-MM-DD (optional)"`
		CompletedAfter    string   `json:"completed_after,omitempty" jsonschema:"Filter by completed date >= YYYY-MM-DD (optional)"`
		CompletedBefore   string   `json:"completed_before,omitempty" jsonschema:"Filter by completed date <= YYYY-MM-DD (optional)"`
		DueAfter          string   `json:"due_after,omitempty" jsonschema:"Filter by due date >= YYYY-MM-DD (optional)"`
		DueBefore         string   `json:"due_before,omitempty" jsonschema:"Filter by due date <= YYYY-MM-DD (optional)"`
		Section           string   `json:"section,omitempty" jsonschema:"PARA section: 01_active (default), 02_areas, or 03_resources"`
	}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "search",
		Description: "Unified search across todos and notes with optional temporal filtering",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args SearchArgs) (*mcp.CallToolResult, any, error) {
		cfg := sess.GetConfig()
		activeDir, err := config.GetSectionPath(cfg, args.Section)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get section path: %w", err)
		}

		results, err := api.UnifiedSearch(activeDir, api.SearchOptions{
			Query:             args.Query,
			Project:           args.Project,
			IncludeTodos:      args.IncludeTodos,
			IncludeNotes:      args.IncludeNotes,
			SearchNoteContent: args.SearchNoteContent,
			Status:            args.Status,
			Priority:          args.Priority,
			Tags:              args.Tags,
			IncludeCompleted:  args.IncludeCompleted,
			CreatedAfter:      args.CreatedAfter,
			CreatedBefore:     args.CreatedBefore,
			CompletedAfter:    args.CompletedAfter,
			CompletedBefore:   args.CompletedBefore,
			DueAfter:          args.DueAfter,
			DueBefore:         args.DueBefore,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("search failed: %w", err)
		}

		data, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal results: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(data)},
			},
		}, nil, nil
	})

	// set_context - unified tool for switching brain and/or project
	type SetContextArgs struct {
		BrainName   string `json:"brain_name,omitempty" jsonschema:"Name of the brain to switch to (optional)"`
		ProjectName string `json:"project_name,omitempty" jsonschema:"Name of the project to focus (optional)"`
	}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "set_context",
		Description: "Switch brain and/or set focused project in one call",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args SetContextArgs) (*mcp.CallToolResult, any, error) {
		// At least one must be provided
		if args.BrainName == "" && args.ProjectName == "" {
			return nil, nil, fmt.Errorf("must provide at least brain_name or project_name")
		}

		// Validate inputs
		if args.BrainName != "" {
			if err := validation.ValidateBrainName(args.BrainName); err != nil {
				return nil, nil, err
			}
		}
		if args.ProjectName != "" {
			if err := validation.ValidateProjectName(args.ProjectName); err != nil {
				return nil, nil, err
			}
		}

		cfg := sess.GetConfig()
		var updates []string

		// Switch brain if specified
		if args.BrainName != "" {
			if err := cfg.SetCurrentBrain(args.BrainName); err != nil {
				return nil, nil, fmt.Errorf("failed to switch brain: %w", err)
			}
			updates = append(updates, fmt.Sprintf("brain=%s", args.BrainName))
		}

		// Set focused project if specified
		if args.ProjectName != "" {
			if err := cfg.SetFocusedProject(args.ProjectName); err != nil {
				return nil, nil, fmt.Errorf("failed to set focused project: %w", err)
			}
			updates = append(updates, fmt.Sprintf("project=%s", args.ProjectName))
		}

		// Save config
		if err := cfg.Save(); err != nil {
			return nil, nil, fmt.Errorf("failed to save config: %w", err)
		}

		// Invalidate cache
		sess.Invalidate()

		// Refresh config in session
		if err := sess.RefreshConfig(); err != nil {
			return nil, nil, fmt.Errorf("failed to refresh config: %w", err)
		}

		message := fmt.Sprintf("Context updated: %v", updates)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: message},
			},
		}, nil, nil
	})

	// get_daily_briefing - comprehensive start-of-day overview
	type GetDailyBriefingArgs struct {
		Section string `json:"section,omitempty" jsonschema:"PARA section: 01_active (default), 02_areas, or 03_resources"`
	}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_daily_briefing",
		Description: "Get comprehensive daily briefing: overdue/due items, high-priority tasks, in-progress work, blocked items, recent completions, and inbox items. Optimized for starting the day.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetDailyBriefingArgs) (*mcp.CallToolResult, any, error) {
		cfg := sess.GetConfig()
		activeDir, err := config.GetSectionPath(cfg, args.Section)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get section path: %w", err)
		}

		brainPath, err := cfg.GetCurrentBrainPath()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get brain path: %w", err)
		}

		dumpPath := filepath.Join(brainPath, "00_dump.md")

		briefing, err := api.GetDailyBriefing(activeDir, dumpPath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate briefing: %w", err)
		}

		data, err := json.MarshalIndent(briefing, "", "  ")
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal briefing: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(data)},
			},
		}, nil, nil
	})

	return nil
}
