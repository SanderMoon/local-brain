package tools

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sandermoonemans/local-brain/mcp/session"
	"github.com/sandermoonemans/local-brain/mcp/validation"
	"github.com/sandermoonemans/local-brain/pkg/api"
	"github.com/sandermoonemans/local-brain/pkg/config"
)

// RegisterNoteTools registers note management tools
func RegisterNoteTools(srv *mcp.Server, sess *session.Session) error {
	// create_project_note
	type CreateProjectNoteArgs struct {
		ProjectName string `json:"project_name" jsonschema:"Project name"`
		Title       string `json:"title" jsonschema:"Note title"`
		Content     string `json:"content" jsonschema:"Note content"`
		Section     string `json:"section,omitempty" jsonschema:"description=PARA section: 01_active (default), 02_areas, or 03_resources"`
	}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_project_note",
		Description: "Create a timestamped note file in a project",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args CreateProjectNoteArgs) (*mcp.CallToolResult, any, error) {
		// Validate inputs
		if err := validation.ValidateProjectName(args.ProjectName); err != nil {
			return nil, nil, err
		}
		if err := validation.ValidateNonEmpty("title", args.Title); err != nil {
			return nil, nil, err
		}
		if err := validation.ValidateNonEmpty("content", args.Content); err != nil {
			return nil, nil, err
		}

		cfg := sess.GetConfig()
		sectionDir, err := config.GetSectionPath(cfg, args.Section)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get section path: %w", err)
		}

		projectDir := filepath.Join(sectionDir, args.ProjectName)
		timestamp := time.Now().Format("2006-01-02")

		notePath, err := api.CreateNoteFile(projectDir, args.Title, args.Content, timestamp)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create note: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Created note: %s", filepath.Base(notePath))},
			},
		}, nil, nil
	})

	// get_note_content
	type GetNoteContentArgs struct {
		NotePath string `json:"note_path" jsonschema:"Full path to the note file"`
	}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_note_content",
		Description: "Read the content of a note file",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetNoteContentArgs) (*mcp.CallToolResult, any, error) {
		// Validate inputs
		if err := validation.ValidateNonEmpty("note_path", args.NotePath); err != nil {
			return nil, nil, err
		}

		content, err := api.ReadNoteFile(args.NotePath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read note: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: content},
			},
		}, nil, nil
	})

	// update_note
	type UpdateNoteArgs struct {
		NotePath string `json:"note_path" jsonschema:"Full path to the note file"`
		Content  string `json:"content" jsonschema:"New full content for the note file"`
	}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "update_note",
		Description: "Replace the full content of an existing note file",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args UpdateNoteArgs) (*mcp.CallToolResult, any, error) {
		if err := validation.ValidateNonEmpty("note_path", args.NotePath); err != nil {
			return nil, nil, err
		}
		if err := validation.ValidateNonEmpty("content", args.Content); err != nil {
			return nil, nil, err
		}

		if err := api.UpdateNoteFile(args.NotePath, args.Content); err != nil {
			return nil, nil, fmt.Errorf("failed to update note: %w", err)
		}

		sess.Invalidate()

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Updated note: %s", filepath.Base(args.NotePath))},
			},
		}, nil, nil
	})

	return nil
}
