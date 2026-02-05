package tools

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sandermoonemans/local-brain/mcp/session"
	"github.com/sandermoonemans/local-brain/mcp/validation"
	"github.com/sandermoonemans/local-brain/pkg/api"
	"github.com/sandermoonemans/local-brain/pkg/markdown"
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

	// refile_item - supports single or batch refiling
	type RefileRequest struct {
		ItemID      string `json:"item_id" jsonschema:"6-character hex ID of the item"`
		ProjectName string `json:"project_name" jsonschema:"Target project name"`
	}
	type RefileItemArgs struct {
		Refiles []RefileRequest `json:"refiles" jsonschema:"Array of refile operations (supports single or multiple)"`
	}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "refile_item",
		Description: "Move one or more items from the inbox to projects (always pass refiles as array, even for single item)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args RefileItemArgs) (*mcp.CallToolResult, any, error) {
		if len(args.Refiles) == 0 {
			return nil, nil, fmt.Errorf("refiles array cannot be empty")
		}

		// Validate all inputs first
		for i, refile := range args.Refiles {
			if err := validation.ValidateTodoID(refile.ItemID); err != nil {
				return nil, nil, fmt.Errorf("refile[%d]: %w", i, err)
			}
			if err := validation.ValidateProjectName(refile.ProjectName); err != nil {
				return nil, nil, fmt.Errorf("refile[%d]: %w", i, err)
			}
		}

		cfg := sess.GetConfig()
		brainPath, err := cfg.GetCurrentBrainPath()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get brain path: %w", err)
		}

		dumpPath := filepath.Join(brainPath, "00_dump.md")

		// Find all items first (to validate they exist before any mutations)
		type itemWithDest struct {
			item       *markdown.DumpItem
			projectDir string
		}
		var itemsToRefile []itemWithDest

		for _, refile := range args.Refiles {
			item, err := api.FindDumpItemByID(dumpPath, refile.ItemID)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to find item %s: %w", refile.ItemID, err)
			}
			if item == nil {
				return nil, nil, validation.NewItemNotFoundError("dump item", refile.ItemID)
			}

			projectDir := filepath.Join(brainPath, "01_active", refile.ProjectName)
			itemsToRefile = append(itemsToRefile, itemWithDest{item, projectDir})
		}

		// Sort items by StartLine descending (process from bottom to top)
		// This prevents line number shifts from affecting subsequent operations
		sort.Slice(itemsToRefile, func(i, j int) bool {
			return itemsToRefile[i].item.StartLine > itemsToRefile[j].item.StartLine
		})

		// Process all refiles
		var results []string
		for _, iwd := range itemsToRefile {
			item := iwd.item
			projectDir := iwd.projectDir

			// Extract ID for display
			itemID := api.ExtractID(item.RawLine)
			if itemID == "" {
				itemID = "unknown"
			}

			// Refile based on type
			if item.Type == "todo" {
				if err := api.RefileTaskToProject(projectDir, item); err != nil {
					return nil, nil, fmt.Errorf("failed to refile task %s: %w", itemID, err)
				}
			} else if item.Type == "note" {
				_, err := api.RefileNoteToProject(dumpPath, projectDir, item)
				if err != nil {
					return nil, nil, fmt.Errorf("failed to refile note %s: %w", itemID, err)
				}
			} else {
				return nil, nil, fmt.Errorf("unknown item type: %s", item.Type)
			}

			// Remove from dump
			if err := api.RemoveItemFromDump(dumpPath, item.StartLine, item.EndLine); err != nil {
				return nil, nil, fmt.Errorf("failed to remove item %s from dump: %w", itemID, err)
			}

			results = append(results, fmt.Sprintf("%s → %s", itemID, filepath.Base(projectDir)))
		}

		// Invalidate cache after mutation
		sess.Invalidate()

		// Format response
		var message string
		if len(results) == 1 {
			message = fmt.Sprintf("Refiled 1 item: %s", results[0])
		} else {
			message = fmt.Sprintf("Refiled %d items:\n- %s", len(results), strings.Join(results, "\n- "))
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: message},
			},
		}, nil, nil
	})

	return nil
}
