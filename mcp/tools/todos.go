package tools

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sandermoonemans/local-brain/mcp/session"
	"github.com/sandermoonemans/local-brain/mcp/validation"
	"github.com/sandermoonemans/local-brain/pkg/api"
	"github.com/sandermoonemans/local-brain/pkg/config"
)

// RegisterTodoTools registers todo management tools
func RegisterTodoTools(srv *mcp.Server, sess *session.Session) error {
	// update_todo - supports single or batch updates
	type TodoUpdate struct {
		TodoID     string   `json:"todo_id" jsonschema:"6-character hex ID of the task"`
		Content    string   `json:"content,omitempty" jsonschema:"New task text (updates the content while preserving metadata)"`
		Status     string   `json:"status,omitempty" jsonschema:"New status (open, in-progress, blocked, done)"`
		Priority   *int     `json:"priority,omitempty" jsonschema:"Priority (1=high, 2=medium, 3=low, null to clear)"`
		DueDate    string   `json:"due_date,omitempty" jsonschema:"Due date in YYYY-MM-DD format, or empty string to clear"`
		AddTags    []string `json:"add_tags,omitempty" jsonschema:"Tags to add"`
		RemoveTags []string `json:"remove_tags,omitempty" jsonschema:"Tags to remove"`
	}
	type UpdateTodoArgs struct {
		Updates []TodoUpdate `json:"updates" jsonschema:"Array of todo updates (supports single or multiple)"`
		Section string       `json:"section,omitempty" jsonschema:"description=PARA section: 01_active (default), 02_areas, or 03_resources"`
	}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "update_todo",
		Description: "Update one or more todos - supports content, status, priority, due date, and tags (always pass updates as array, even for single todo)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args UpdateTodoArgs) (*mcp.CallToolResult, any, error) {
		if len(args.Updates) == 0 {
			return nil, nil, fmt.Errorf("updates array cannot be empty")
		}

		// Validate all inputs first
		for i, update := range args.Updates {
			if err := validation.ValidateTodoID(update.TodoID); err != nil {
				return nil, nil, fmt.Errorf("update[%d]: %w", i, err)
			}
			if update.Status != "" {
				if err := validation.ValidateTodoStatus(update.Status); err != nil {
					return nil, nil, fmt.Errorf("update[%d]: %w", i, err)
				}
			}
			if err := validation.ValidatePriority(update.Priority); err != nil {
				return nil, nil, fmt.Errorf("update[%d]: %w", i, err)
			}
			if err := validation.ValidateDueDate(update.DueDate); err != nil {
				return nil, nil, fmt.Errorf("update[%d]: %w", i, err)
			}
		}

		cfg := sess.GetConfig()
		activeDir, err := config.GetSectionPath(cfg, args.Section)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get section path: %w", err)
		}

		// Process all updates
		var results []string
		for _, update := range args.Updates {
			// Reload todos for each update to get fresh line numbers
			todos, err := api.ParseAllTodos(activeDir, true)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse todos: %w", err)
			}

			todo := api.FindTodoByID(todos, update.TodoID)
			if todo == nil {
				return nil, nil, validation.NewItemNotFoundError("todo", update.TodoID)
			}

			var changes []string

			// Update content first (if provided), as it changes the base text
			if update.Content != "" {
				if err := api.UpdateTodoContent(todo, update.Content); err != nil {
					return nil, nil, fmt.Errorf("failed to update content for %s: %w", update.TodoID, err)
				}
				changes = append(changes, fmt.Sprintf("content=\"%s\"", update.Content))
				// Reload after content change
				todos, _ = api.ParseAllTodos(activeDir, true)
				todo = api.FindTodoByID(todos, update.TodoID)
			}

			// Update status
			if update.Status != "" {
				if err := api.SetTodoStatus(todo, update.Status); err != nil {
					return nil, nil, fmt.Errorf("failed to set status for %s: %w", update.TodoID, err)
				}
				changes = append(changes, fmt.Sprintf("status=%s", update.Status))
				todos, _ = api.ParseAllTodos(activeDir, true)
				todo = api.FindTodoByID(todos, update.TodoID)
			}

			// Update priority
			if update.Priority != nil {
				if err := api.SetTodoPriority(todo, update.Priority); err != nil {
					return nil, nil, fmt.Errorf("failed to set priority for %s: %w", update.TodoID, err)
				}
				changes = append(changes, fmt.Sprintf("priority=%d", *update.Priority))
				todos, _ = api.ParseAllTodos(activeDir, true)
				todo = api.FindTodoByID(todos, update.TodoID)
			}

			// Update due date
			if update.DueDate != "" {
				if err := api.SetTodoDueDate(todo, update.DueDate); err != nil {
					return nil, nil, fmt.Errorf("failed to set due date for %s: %w", update.TodoID, err)
				}
				changes = append(changes, fmt.Sprintf("due_date=%s", update.DueDate))
				todos, _ = api.ParseAllTodos(activeDir, true)
				todo = api.FindTodoByID(todos, update.TodoID)
			}

			// Add tags
			if len(update.AddTags) > 0 {
				if err := api.AddTodoTags(todo, update.AddTags); err != nil {
					return nil, nil, fmt.Errorf("failed to add tags for %s: %w", update.TodoID, err)
				}
				changes = append(changes, fmt.Sprintf("+tags=%v", update.AddTags))
				todos, _ = api.ParseAllTodos(activeDir, true)
				todo = api.FindTodoByID(todos, update.TodoID)
			}

			// Remove tags
			if len(update.RemoveTags) > 0 {
				if err := api.RemoveTodoTags(todo, update.RemoveTags); err != nil {
					return nil, nil, fmt.Errorf("failed to remove tags for %s: %w", update.TodoID, err)
				}
				changes = append(changes, fmt.Sprintf("-tags=%v", update.RemoveTags))
			}

			if len(changes) > 0 {
				results = append(results, fmt.Sprintf("%s: %s", update.TodoID, strings.Join(changes, ", ")))
			}
		}

		sess.Invalidate()

		// Format response
		var message string
		if len(results) == 1 {
			message = fmt.Sprintf("Updated 1 todo:\n%s", results[0])
		} else {
			message = fmt.Sprintf("Updated %d todos:\n- %s", len(results), strings.Join(results, "\n- "))
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: message},
			},
		}, nil, nil
	})

	// create_todo_in_project - supports single or batch creation with full metadata
	type TodoCreateRequest struct {
		Content  string   `json:"content" jsonschema:"Task content (required)"`
		Priority *int     `json:"priority,omitempty" jsonschema:"Priority: 1=high, 2=medium, 3=low (optional)"`
		DueDate  string   `json:"due_date,omitempty" jsonschema:"Due date in YYYY-MM-DD format (optional)"`
		Tags     []string `json:"tags,omitempty" jsonschema:"Tags (optional)"`
		Status   string   `json:"status,omitempty" jsonschema:"Initial status: open, in-progress, blocked (optional, defaults to open)"`
	}
	type CreateTodoInProjectArgs struct {
		ProjectName string              `json:"project_name" jsonschema:"Project name"`
		Todos       []TodoCreateRequest `json:"todos" jsonschema:"Array of todos to create (supports single or multiple)"`
		Section     string              `json:"section,omitempty" jsonschema:"description=PARA section: 01_active (default), 02_areas, or 03_resources"`
	}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_todo_in_project",
		Description: "Add one or more tasks directly to a project with optional metadata (priority, due_date, tags, status). Always pass todos as array, even for single task.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args CreateTodoInProjectArgs) (*mcp.CallToolResult, any, error) {
		// Validate inputs
		if err := validation.ValidateProjectName(args.ProjectName); err != nil {
			return nil, nil, err
		}
		if len(args.Todos) == 0 {
			return nil, nil, fmt.Errorf("todos array cannot be empty")
		}

		// Validate all todos first
		for i, todo := range args.Todos {
			if strings.TrimSpace(todo.Content) == "" {
				return nil, nil, fmt.Errorf("todo[%d]: content cannot be empty", i)
			}
			if todo.Priority != nil && (*todo.Priority < 1 || *todo.Priority > 3) {
				return nil, nil, fmt.Errorf("todo[%d]: priority must be 1-3", i)
			}
			if todo.DueDate != "" {
				if err := validation.ValidateDueDate(todo.DueDate); err != nil {
					return nil, nil, fmt.Errorf("todo[%d]: %w", i, err)
				}
			}
			if todo.Status != "" && todo.Status != "open" && todo.Status != "in-progress" && todo.Status != "blocked" {
				return nil, nil, fmt.Errorf("todo[%d]: invalid status '%s' (must be: open, in-progress, blocked)", i, todo.Status)
			}
		}

		cfg := sess.GetConfig()
		sectionDir, err := config.GetSectionPath(cfg, args.Section)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get section path: %w", err)
		}

		projectDir := filepath.Join(sectionDir, args.ProjectName)

		// Create all tasks
		var created []string
		for _, todo := range args.Todos {
			createReq := api.TodoCreateRequest{
				Content:  todo.Content,
				Priority: todo.Priority,
				DueDate:  todo.DueDate,
				Tags:     todo.Tags,
				Status:   todo.Status,
			}

			if err := api.AppendTodoWithMetadata(projectDir, createReq); err != nil {
				return nil, nil, fmt.Errorf("failed to create todo '%s': %w", todo.Content, err)
			}

			// Build summary string
			summary := todo.Content
			if todo.Priority != nil {
				summary = fmt.Sprintf("%s (p:%d)", summary, *todo.Priority)
			}
			if todo.DueDate != "" {
				summary = fmt.Sprintf("%s [due:%s]", summary, todo.DueDate)
			}
			if len(todo.Tags) > 0 {
				summary = fmt.Sprintf("%s #%s", summary, strings.Join(todo.Tags, " #"))
			}

			created = append(created, summary)
		}

		sess.Invalidate()

		// Format response
		var message string
		if len(created) == 1 {
			message = fmt.Sprintf("Created task in %s: %s", args.ProjectName, created[0])
		} else {
			message = fmt.Sprintf("Created %d tasks in %s:\n- %s", len(created), args.ProjectName, strings.Join(created, "\n- "))
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: message},
			},
		}, nil, nil
	})

	// delete_todo (requires user confirmation via MCP)
	type DeleteTodoArgs struct {
		TodoID  string `json:"todo_id" jsonschema:"6-character hex ID of the task to delete"`
		Section string `json:"section,omitempty" jsonschema:"description=PARA section: 01_active (default), 02_areas, or 03_resources"`
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
		activeDir, err := config.GetSectionPath(cfg, args.Section)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get section path: %w", err)
		}

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
