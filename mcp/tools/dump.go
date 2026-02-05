package tools

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sandermoonemans/local-brain/mcp/session"
	"github.com/sandermoonemans/local-brain/mcp/validation"
	"github.com/sandermoonemans/local-brain/pkg/api"
)

// RegisterDumpTools registers dump/capture tools
func RegisterDumpTools(srv *mcp.Server, sess *session.Session) error {
	// add_to_dump - unified tool for adding tasks or notes to inbox
	type AddToDumpArgs struct {
		Type    string `json:"type" jsonschema:"Item type: 'task' or 'note'"`
		Content string `json:"content" jsonschema:"Task content or note content (can be multi-line for notes)"`
		Title   string `json:"title,omitempty" jsonschema:"Note title (required when type is 'note', ignored for tasks)"`
	}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "add_to_dump",
		Description: "Quick capture a task or note to the inbox",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args AddToDumpArgs) (*mcp.CallToolResult, any, error) {
		// Validate type
		if args.Type != "task" && args.Type != "note" {
			return nil, nil, fmt.Errorf("invalid type: %s (must be 'task' or 'note')", args.Type)
		}

		// Validate inputs
		if err := validation.ValidateNonEmpty("content", args.Content); err != nil {
			return nil, nil, err
		}

		// For notes, title is required
		if args.Type == "note" {
			if err := validation.ValidateNonEmpty("title", args.Title); err != nil {
				return nil, nil, err
			}
		}

		cfg := sess.GetConfig()
		brainPath, err := cfg.GetCurrentBrainPath()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get brain path: %w", err)
		}

		dumpPath := filepath.Join(brainPath, "00_dump.md")
		timestamp := time.Now().Format("2006-01-02")

		var responseMsg string

		if args.Type == "task" {
			if err := api.AddTaskToDump(dumpPath, args.Content, timestamp); err != nil {
				return nil, nil, fmt.Errorf("failed to add task: %w", err)
			}
			responseMsg = fmt.Sprintf("Added task to inbox: %s", args.Content)
		} else {
			// Split content into lines for notes
			contentLines := strings.Split(args.Content, "\n")
			if err := api.AddNoteToDump(dumpPath, args.Title, contentLines, timestamp); err != nil {
				return nil, nil, fmt.Errorf("failed to add note: %w", err)
			}
			responseMsg = fmt.Sprintf("Added note to inbox: %s", args.Title)
		}

		// Invalidate cache after mutation
		sess.Invalidate()

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: responseMsg},
			},
		}, nil, nil
	})

	// refile_item
	type RefileItemArgs struct {
		ItemID      string `json:"item_id" jsonschema:"6-character hex ID of the item"`
		ProjectName string `json:"project_name" jsonschema:"Target project name"`
	}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "refile_item",
		Description: "Move an item from the inbox to a project",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args RefileItemArgs) (*mcp.CallToolResult, any, error) {
		// Validate inputs
		if err := validation.ValidateTodoID(args.ItemID); err != nil {
			return nil, nil, err
		}
		if err := validation.ValidateProjectName(args.ProjectName); err != nil {
			return nil, nil, err
		}

		cfg := sess.GetConfig()
		brainPath, err := cfg.GetCurrentBrainPath()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get brain path: %w", err)
		}

		dumpPath := filepath.Join(brainPath, "00_dump.md")
		projectDir := filepath.Join(brainPath, "01_active", args.ProjectName)

		// Find the item in dump
		item, err := api.FindDumpItemByID(dumpPath, args.ItemID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to find item: %w", err)
		}
		if item == nil {
			return nil, nil, validation.NewItemNotFoundError("dump item", args.ItemID)
		}

		// Refile based on type
		if item.Type == "todo" {
			if err := api.RefileTaskToProject(projectDir, item); err != nil {
				return nil, nil, fmt.Errorf("failed to refile task: %w", err)
			}
		} else if item.Type == "note" {
			_, err := api.RefileNoteToProject(dumpPath, projectDir, item)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to refile note: %w", err)
			}
		} else {
			return nil, nil, fmt.Errorf("unknown item type: %s", item.Type)
		}

		// Remove from dump
		if err := api.RemoveItemFromDump(dumpPath, item.StartLine, item.EndLine); err != nil {
			return nil, nil, fmt.Errorf("failed to remove item from dump: %w", err)
		}

		// Invalidate cache after mutation
		sess.Invalidate()

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Refiled %s to project: %s", item.Type, args.ProjectName)},
			},
		}, nil, nil
	})

	return nil
}
