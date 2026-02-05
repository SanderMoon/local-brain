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

	// get_all_todos
	type GetAllTodosArgs struct {
		IncludeCompleted bool   `json:"include_completed" jsonschema:"Whether to include completed tasks (default: false)"`
		Project          string `json:"project,omitempty" jsonschema:"Filter by specific project name (optional)"`
		CreatedAfter     string `json:"created_after,omitempty" jsonschema:"Filter by captured date >= YYYY-MM-DD (optional)"`
		CreatedBefore    string `json:"created_before,omitempty" jsonschema:"Filter by captured date <= YYYY-MM-DD (optional)"`
		CompletedAfter   string `json:"completed_after,omitempty" jsonschema:"Filter by done date >= YYYY-MM-DD (optional)"`
		CompletedBefore  string `json:"completed_before,omitempty" jsonschema:"Filter by done date <= YYYY-MM-DD (optional)"`
	}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_all_todos",
		Description: "Get all tasks across all projects with rich metadata (status, priority, due date, tags) and optional temporal filtering",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetAllTodosArgs) (*mcp.CallToolResult, any, error) {
		cfg := sess.GetConfig()
		brainPath, err := cfg.GetCurrentBrainPath()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get brain path: %w", err)
		}

		activeDir := filepath.Join(brainPath, "01_active")
		todos, err := api.ParseAllTodos(activeDir, args.IncludeCompleted)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse todos: %w", err)
		}

		// Filter by project if specified
		if args.Project != "" {
			var filtered []api.TodoItem
			for _, todo := range todos {
				if todo.Project == args.Project {
					filtered = append(filtered, todo)
				}
			}
			todos = filtered
		}

		// Apply temporal filters
		if args.CreatedAfter != "" || args.CreatedBefore != "" || args.CompletedAfter != "" || args.CompletedBefore != "" {
			todos = api.FilterTodosByTemporal(todos, args.CreatedAfter, args.CreatedBefore, args.CompletedAfter, args.CompletedBefore)
		}

		data, err := json.MarshalIndent(todos, "", "  ")
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal todos: %w", err)
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
		brainPath, err := cfg.GetCurrentBrainPath()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get brain path: %w", err)
		}

		projectPath := filepath.Join(brainPath, "01_active", args.ProjectName)
		focusedProject := cfg.GetFocusedProject()

		projectCtx, err := api.GetProjectContext(projectPath, args.ProjectName, focusedProject, args.IncludeCompleted, args.IncludeNoteContent)
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

	// search_todos
	type SearchTodosArgs struct {
		Query           string   `json:"query,omitempty" jsonschema:"Search query for todo content (case-insensitive)"`
		Project         string   `json:"project,omitempty" jsonschema:"Filter by project name (optional)"`
		Status          string   `json:"status,omitempty" jsonschema:"Filter by status: open, in-progress, blocked, done (optional)"`
		Tags            []string `json:"tags,omitempty" jsonschema:"Filter by tags (OR logic - match any tag)"`
		CreatedAfter    string   `json:"created_after,omitempty" jsonschema:"Filter by captured date >= YYYY-MM-DD (optional)"`
		CreatedBefore   string   `json:"created_before,omitempty" jsonschema:"Filter by captured date <= YYYY-MM-DD (optional)"`
		CompletedAfter  string   `json:"completed_after,omitempty" jsonschema:"Filter by done date >= YYYY-MM-DD (optional)"`
		CompletedBefore string   `json:"completed_before,omitempty" jsonschema:"Filter by done date <= YYYY-MM-DD (optional)"`
	}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "search_todos",
		Description: "Search and filter todos by query, project, status, tags, and dates",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args SearchTodosArgs) (*mcp.CallToolResult, any, error) {
		cfg := sess.GetConfig()
		brainPath, err := cfg.GetCurrentBrainPath()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get brain path: %w", err)
		}

		activeDir := filepath.Join(brainPath, "01_active")
		allTodos, err := api.ParseAllTodos(activeDir, true)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse todos: %w", err)
		}

		results := api.SearchTodos(allTodos, args.Query, args.Project, args.Status, args.Tags)

		// Apply temporal filters
		if args.CreatedAfter != "" || args.CreatedBefore != "" || args.CompletedAfter != "" || args.CompletedBefore != "" {
			results = api.FilterTodosByTemporal(results, args.CreatedAfter, args.CreatedBefore, args.CompletedAfter, args.CompletedBefore)
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

	// search - unified search for todos and notes
	type SearchArgs struct {
		Query             string `json:"query,omitempty" jsonschema:"Search query (optional)"`
		Project           string `json:"project,omitempty" jsonschema:"Filter by project name (optional)"`
		IncludeTodos      bool   `json:"include_todos,omitempty" jsonschema:"Include todos in search (default: true)"`
		IncludeNotes      bool   `json:"include_notes,omitempty" jsonschema:"Include notes in search (default: true)"`
		SearchNoteContent bool   `json:"search_note_content,omitempty" jsonschema:"Search note content (not just titles)"`
		CreatedAfter      string `json:"created_after,omitempty" jsonschema:"Filter by created date >= YYYY-MM-DD (optional)"`
		CreatedBefore     string `json:"created_before,omitempty" jsonschema:"Filter by created date <= YYYY-MM-DD (optional)"`
		CompletedAfter    string `json:"completed_after,omitempty" jsonschema:"Filter by completed date >= YYYY-MM-DD (optional)"`
		CompletedBefore   string `json:"completed_before,omitempty" jsonschema:"Filter by completed date <= YYYY-MM-DD (optional)"`
	}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "search",
		Description: "Unified search across todos and notes with optional temporal filtering",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args SearchArgs) (*mcp.CallToolResult, any, error) {
		cfg := sess.GetConfig()
		brainPath, err := cfg.GetCurrentBrainPath()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get brain path: %w", err)
		}

		activeDir := filepath.Join(brainPath, "01_active")

		results, err := api.UnifiedSearch(
			activeDir,
			args.Query,
			args.Project,
			args.IncludeTodos,
			args.IncludeNotes,
			args.SearchNoteContent,
			args.CreatedAfter,
			args.CreatedBefore,
			args.CompletedAfter,
			args.CompletedBefore,
		)
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

	// switch_brain
	type SwitchBrainArgs struct {
		BrainName string `json:"brain_name" jsonschema:"Name of the brain to switch to"`
	}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "switch_brain",
		Description: "Switch to a different brain workspace",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args SwitchBrainArgs) (*mcp.CallToolResult, any, error) {
		// Validate inputs
		if err := validation.ValidateBrainName(args.BrainName); err != nil {
			return nil, nil, err
		}

		cfg := sess.GetConfig()

		if err := cfg.SetCurrentBrain(args.BrainName); err != nil {
			return nil, nil, fmt.Errorf("failed to switch brain: %w", err)
		}

		if err := cfg.Save(); err != nil {
			return nil, nil, fmt.Errorf("failed to save config: %w", err)
		}

		// Invalidate cache after brain switch
		sess.Invalidate()

		// Refresh config in session
		if err := sess.RefreshConfig(); err != nil {
			return nil, nil, fmt.Errorf("failed to refresh config: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Switched to brain: %s", args.BrainName)},
			},
		}, nil, nil
	})

	return nil
}
