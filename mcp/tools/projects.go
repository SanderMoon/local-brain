package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sandermoonemans/local-brain/mcp/session"
	"github.com/sandermoonemans/local-brain/mcp/validation"
	"github.com/sandermoonemans/local-brain/pkg/api"
	"github.com/sandermoonemans/local-brain/pkg/config"
)

// RegisterProjectTools registers project management tools
func RegisterProjectTools(srv *mcp.Server, sess *session.Session) error {
	// create_project
	type CreateProjectArgs struct {
		Name        string `json:"name" jsonschema:"Project name (alphanumeric, hyphens, underscores)"`
		Description string `json:"description,omitempty" jsonschema:"Project description (optional)"`
		Section     string `json:"section,omitempty" jsonschema:"PARA section: 01_active (default), 02_areas, or 03_resources"`
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
		activeDir, err := config.GetSectionPath(cfg, args.Section)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get section path: %w", err)
		}

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

	return nil
}
