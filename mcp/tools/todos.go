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
	// update_todo - unified tool for updating todo status and metadata
	type UpdateTodoArgs struct {
		TodoID     string   `json:"todo_id" jsonschema:"6-character hex ID of the task"`
		Content    string   `json:"content,omitempty" jsonschema:"New task text (updates the content while preserving metadata)"`
		Status     string   `json:"status,omitempty" jsonschema:"New status (open, in-progress, blocked, done)"`
		Priority   *int     `json:"priority,omitempty" jsonschema:"Priority (1=high, 2=medium, 3=low, null to clear)"`
		DueDate    string   `json:"due_date,omitempty" jsonschema:"Due date in YYYY-MM-DD format, or empty string to clear"`
		AddTags    []string `json:"add_tags,omitempty" jsonschema:"Tags to add"`
		RemoveTags []string `json:"remove_tags,omitempty" jsonschema:"Tags to remove"`
	}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "update_todo",
		Description: "Update todo content, status, priority, due date, or tags in one call",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args UpdateTodoArgs) (*mcp.CallToolResult, any, error) {
		// Validate inputs
		if err := validation.ValidateTodoID(args.TodoID); err != nil {
			return nil, nil, err
		}
		if args.Status != "" {
			if err := validation.ValidateTodoStatus(args.Status); err != nil {
				return nil, nil, err
			}
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

		var updates []string

		// Update content first (if provided), as it changes the base text
		if args.Content != "" {
			if err := api.UpdateTodoContent(todo, args.Content); err != nil {
				return nil, nil, fmt.Errorf("failed to update content: %w", err)
			}
			updates = append(updates, fmt.Sprintf("content=\"%s\"", args.Content))
			// Reload after content change
			todos, _ = api.ParseAllTodos(activeDir, true)
			todo = api.FindTodoByID(todos, args.TodoID)
		}

		// Update status
		if args.Status != "" {
			if err := api.SetTodoStatus(todo, args.Status); err != nil {
				return nil, nil, fmt.Errorf("failed to set status: %w", err)
			}
			updates = append(updates, fmt.Sprintf("status=%s", args.Status))
			// Reload after status change
			todos, _ = api.ParseAllTodos(activeDir, true)
			todo = api.FindTodoByID(todos, args.TodoID)
		}

		// Update priority
		if args.Priority != nil {
			if err := api.SetTodoPriority(todo, args.Priority); err != nil {
				return nil, nil, fmt.Errorf("failed to set priority: %w", err)
			}
			updates = append(updates, fmt.Sprintf("priority=%d", *args.Priority))
			todos, _ = api.ParseAllTodos(activeDir, true)
			todo = api.FindTodoByID(todos, args.TodoID)
		}

		// Update due date
		if args.DueDate != "" {
			if err := api.SetTodoDueDate(todo, args.DueDate); err != nil {
				return nil, nil, fmt.Errorf("failed to set due date: %w", err)
			}
			updates = append(updates, fmt.Sprintf("due_date=%s", args.DueDate))
			todos, _ = api.ParseAllTodos(activeDir, true)
			todo = api.FindTodoByID(todos, args.TodoID)
		}

		// Add tags
		if len(args.AddTags) > 0 {
			if err := api.AddTodoTags(todo, args.AddTags); err != nil {
				return nil, nil, fmt.Errorf("failed to add tags: %w", err)
			}
			updates = append(updates, fmt.Sprintf("added_tags=%v", args.AddTags))
			todos, _ = api.ParseAllTodos(activeDir, true)
			todo = api.FindTodoByID(todos, args.TodoID)
		}

		// Remove tags
		if len(args.RemoveTags) > 0 {
			if err := api.RemoveTodoTags(todo, args.RemoveTags); err != nil {
				return nil, nil, fmt.Errorf("failed to remove tags: %w", err)
			}
			updates = append(updates, fmt.Sprintf("removed_tags=%v", args.RemoveTags))
		}

		sess.Invalidate()

		var message string
		if len(updates) > 0 {
			message = fmt.Sprintf("Updated task: %s\nChanges: %v", todo.Content, updates)
		} else {
			message = fmt.Sprintf("No changes made to task: %s", todo.Content)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: message},
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
