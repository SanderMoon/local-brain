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

// RegisterProjectTools registers project management tools
func RegisterProjectTools(srv *mcp.Server, sess *session.Session) error {
	// create_project
	type CreateProjectArgs struct {
		Name        string `json:"name" jsonschema:"Project name (alphanumeric, hyphens, underscores)"`
		Description string `json:"description,omitempty" jsonschema:"Project description (optional)"`
	}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_project",
		Description: "Create a new project",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args CreateProjectArgs) (*mcp.CallToolResult, any, error) {
		// Validate inputs
		if err := validation.ValidateProjectName(args.Name); err != nil {
			return nil, nil, err
		}

		cfg := sess.GetConfig()
		brainPath, err := cfg.GetCurrentBrainPath()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get brain path: %w", err)
		}

		activeDir := filepath.Join(brainPath, "01_active")
		projectPath, err := api.CreateProject(activeDir, args.Name)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create project: %w", err)
		}

		// Set description if provided
		if args.Description != "" {
			if err := api.WriteProjectDescription(projectPath, args.Description); err != nil {
				return nil, nil, fmt.Errorf("failed to write description: %w", err)
			}
		}

		sess.Invalidate()

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Created project: %s", args.Name)},
			},
		}, nil, nil
	})

	// set_focused_project
	type SetFocusedProjectArgs struct {
		ProjectName string `json:"project_name" jsonschema:"Project name to focus"`
	}
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "set_focused_project",
		Description: "Set the focused project",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args SetFocusedProjectArgs) (*mcp.CallToolResult, any, error) {
		// Validate inputs
		if err := validation.ValidateProjectName(args.ProjectName); err != nil {
			return nil, nil, err
		}

		cfg := sess.GetConfig()

		if err := cfg.SetFocusedProject(args.ProjectName); err != nil {
			return nil, nil, fmt.Errorf("failed to set focused project: %w", err)
		}

		if err := cfg.Save(); err != nil {
			return nil, nil, fmt.Errorf("failed to save config: %w", err)
		}

		sess.Invalidate()

		// Refresh config in session
		if err := sess.RefreshConfig(); err != nil {
			return nil, nil, fmt.Errorf("failed to refresh config: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Set focused project: %s", args.ProjectName)},
			},
		}, nil, nil
	})

	return nil
}
