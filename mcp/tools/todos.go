package tools

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sandermoonemans/local-brain/mcp/session"
	"github.com/sandermoonemans/local-brain/mcp/validation"
	"github.com/sandermoonemans/local-brain/pkg/api"
)

// RegisterTodoTools registers todo management tools
func RegisterTodoTools(srv *mcp.Server, sess *session.Session) error {
	// update_todo_status
	type UpdateTodoStatusArgs struct {
		TodoID string `json:"todo_id" jsonschema:"6-character hex ID of the task"`
		Status string `json:"status" jsonschema:"New status (open, in-progress, blocked, done)"`
	}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "update_todo_status",
		Description: "Change a task's status",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args UpdateTodoStatusArgs) (*mcp.CallToolResult, any, error) {
		// Validate inputs
		if err := validation.ValidateTodoID(args.TodoID); err != nil {
			return nil, nil, err
		}
		if err := validation.ValidateTodoStatus(args.Status); err != nil {
			return nil, nil, err
		}

		cfg := sess.GetConfig()
		brainPath, err := cfg.GetCurrentBrainPath()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get brain path: %w", err)
		}

		activeDir := filepath.Join(brainPath, "01_active")
		todos, err := api.ParseAllTodos(activeDir, true)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse todos: %w", err)
		}

		todo := api.FindTodoByID(todos, args.TodoID)
		if todo == nil {
			return nil, nil, validation.NewItemNotFoundError("todo", args.TodoID)
		}

		if err := api.SetTodoStatus(todo, args.Status); err != nil {
			return nil, nil, fmt.Errorf("failed to update status: %w", err)
		}

		sess.Invalidate()

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Updated task status to %s: %s", args.Status, todo.Content)},
			},
		}, nil, nil
	})

	// update_todo_metadata
	type UpdateTodoMetadataArgs struct {
		TodoID     string   `json:"todo_id" jsonschema:"6-character hex ID of the task"`
		Priority   *int     `json:"priority,omitempty" jsonschema:"Priority (1=high, 2=medium, 3=low, null to clear)"`
		DueDate    string   `json:"due_date,omitempty" jsonschema:"Due date in YYYY-MM-DD format, or empty string to clear"`
		AddTags    []string `json:"add_tags,omitempty" jsonschema:"Tags to add"`
		RemoveTags []string `json:"remove_tags,omitempty" jsonschema:"Tags to remove"`
	}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "update_todo_metadata",
		Description: "Batch update task metadata (priority, due date, tags)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args UpdateTodoMetadataArgs) (*mcp.CallToolResult, any, error) {
		// Validate inputs
		if err := validation.ValidateTodoID(args.TodoID); err != nil {
			return nil, nil, err
		}
		if err := validation.ValidatePriority(args.Priority); err != nil {
			return nil, nil, err
		}
		if err := validation.ValidateDueDate(args.DueDate); err != nil {
			return nil, nil, err
		}

		cfg := sess.GetConfig()
		brainPath, err := cfg.GetCurrentBrainPath()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get brain path: %w", err)
		}

		activeDir := filepath.Join(brainPath, "01_active")
		todos, err := api.ParseAllTodos(activeDir, true)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse todos: %w", err)
		}

		todo := api.FindTodoByID(todos, args.TodoID)
		if todo == nil {
			return nil, nil, validation.NewItemNotFoundError("todo", args.TodoID)
		}

		// Update priority if specified
		if args.Priority != nil || req.Params.Arguments != nil {
			if err := api.SetTodoPriority(todo, args.Priority); err != nil {
				return nil, nil, fmt.Errorf("failed to set priority: %w", err)
			}
		}

		// Update due date if specified
		if args.DueDate != "" {
			if err := api.SetTodoDueDate(todo, args.DueDate); err != nil {
				return nil, nil, fmt.Errorf("failed to set due date: %w", err)
			}
		}

		// Add tags if specified
		if len(args.AddTags) > 0 {
			if err := api.AddTodoTags(todo, args.AddTags); err != nil {
				return nil, nil, fmt.Errorf("failed to add tags: %w", err)
			}
		}

		// Remove tags if specified
		if len(args.RemoveTags) > 0 {
			if err := api.RemoveTodoTags(todo, args.RemoveTags); err != nil {
				return nil, nil, fmt.Errorf("failed to remove tags: %w", err)
			}
		}

		sess.Invalidate()

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Updated task metadata: %s", todo.Content)},
			},
		}, nil, nil
	})

	// create_todo_in_project
	type CreateTodoInProjectArgs struct {
		ProjectName string `json:"project_name" jsonschema:"Project name"`
		Content     string `json:"content" jsonschema:"Task content"`
	}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_todo_in_project",
		Description: "Add a new task directly to a project",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args CreateTodoInProjectArgs) (*mcp.CallToolResult, any, error) {
		// Validate inputs
		if err := validation.ValidateProjectName(args.ProjectName); err != nil {
			return nil, nil, err
		}
		if err := validation.ValidateNonEmpty("content", args.Content); err != nil {
			return nil, nil, err
		}

		cfg := sess.GetConfig()
		brainPath, err := cfg.GetCurrentBrainPath()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get brain path: %w", err)
		}

		projectDir := filepath.Join(brainPath, "01_active", args.ProjectName)

		if err := api.AppendTodoToProject(projectDir, args.Content); err != nil {
			return nil, nil, fmt.Errorf("failed to create todo: %w", err)
		}

		sess.Invalidate()

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Created task in %s: %s", args.ProjectName, args.Content)},
			},
		}, nil, nil
	})

	// delete_todo (requires user confirmation via MCP)
	type DeleteTodoArgs struct {
		TodoID string `json:"todo_id" jsonschema:"6-character hex ID of the task to delete"`
	}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "delete_todo",
		Description: "Delete a task permanently (requires user confirmation)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args DeleteTodoArgs) (*mcp.CallToolResult, any, error) {
		// Validate inputs
		if err := validation.ValidateTodoID(args.TodoID); err != nil {
			return nil, nil, err
		}

		cfg := sess.GetConfig()
		brainPath, err := cfg.GetCurrentBrainPath()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get brain path: %w", err)
		}

		activeDir := filepath.Join(brainPath, "01_active")
		todos, err := api.ParseAllTodos(activeDir, true)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse todos: %w", err)
		}

		todo := api.FindTodoByID(todos, args.TodoID)
		if todo == nil {
			return nil, nil, validation.NewItemNotFoundError("todo", args.TodoID)
		}

		if err := api.DeleteTodoLine(todo); err != nil {
			return nil, nil, fmt.Errorf("failed to delete todo: %w", err)
		}

		sess.Invalidate()

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Deleted task: %s", todo.Content)},
			},
		}, nil, nil
	})

	return nil
}
