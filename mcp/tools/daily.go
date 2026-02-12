package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sandermoonemans/local-brain/mcp/session"
	"github.com/sandermoonemans/local-brain/pkg/api"
)

// RegisterDailyTools registers daily note tools
func RegisterDailyTools(srv *mcp.Server, sess *session.Session) error {
	type CreateDailyNoteArgs struct {
		Date string `json:"date,omitempty" jsonschema:"Optional date in YYYY-MM-DD format (defaults to today)"`
	}

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_daily_note",
		Description: "Create or open a daily note in {brain}/00_daily/YYYY-MM-DD.md. Includes a briefing section with overdue todos. Returns the path, date, and whether the note was newly created.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args CreateDailyNoteArgs) (*mcp.CallToolResult, any, error) {
		// Determine date
		date := args.Date
		if date == "" {
			date = time.Now().Format("2006-01-02")
		} else {
			// Validate date format
			if _, err := time.Parse("2006-01-02", date); err != nil {
				return nil, nil, fmt.Errorf("invalid date format %q: must be YYYY-MM-DD", date)
			}
		}

		cfg := sess.GetConfig()
		brainPath, err := cfg.GetCurrentBrainPath()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get brain path: %w", err)
		}

		activeDir := filepath.Join(brainPath, "01_active")
		today := time.Now().Format("2006-01-02")

		allTodos, err := api.ParseAllTodos(activeDir, false)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse todos: %w", err)
		}

		var overdueTodos []api.TodoItem
		for _, todo := range allTodos {
			if todo.DueDate != "" && todo.DueDate < today {
				overdueTodos = append(overdueTodos, todo)
			}
		}

		result, err := api.CreateOrOpenDailyNote(brainPath, date, overdueTodos)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create daily note: %w", err)
		}

		data, err := json.MarshalIndent(map[string]any{
			"path":   result.Path,
			"date":   result.Date,
			"is_new": result.IsNew,
		}, "", "  ")
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal result: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(data)},
			},
		}, nil, nil
	})

	return nil
}
