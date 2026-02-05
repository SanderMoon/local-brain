package main

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sandermoonemans/local-brain/mcp/session"
	"github.com/sandermoonemans/local-brain/mcp/tools"
	"github.com/sandermoonemans/local-brain/pkg/config"
)

// MCPServer wraps the MCP server and session
type MCPServer struct {
	server  *mcp.Server
	session *session.Session
}

// NewMCPServer creates a new MCP server instance
func NewMCPServer() (*MCPServer, error) {
	// Load brain configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load brain config: %w", err)
	}

	// Create session manager
	sess := session.NewSession(cfg)

	// Create MCP server
	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "local-brain-mcp",
		Version: version,
	}, nil)

	// Register tool handlers
	if err := tools.RegisterContextTools(srv, sess); err != nil {
		return nil, fmt.Errorf("failed to register context tools: %w", err)
	}

	if err := tools.RegisterDumpTools(srv, sess); err != nil {
		return nil, fmt.Errorf("failed to register dump tools: %w", err)
	}

	if err := tools.RegisterTodoTools(srv, sess); err != nil {
		return nil, fmt.Errorf("failed to register todo tools: %w", err)
	}

	if err := tools.RegisterNoteTools(srv, sess); err != nil {
		return nil, fmt.Errorf("failed to register note tools: %w", err)
	}

	if err := tools.RegisterProjectTools(srv, sess); err != nil {
		return nil, fmt.Errorf("failed to register project tools: %w", err)
	}

	return &MCPServer{
		server:  srv,
		session: sess,
	}, nil
}

// Run starts the MCP server with stdio transport
func (s *MCPServer) Run(ctx context.Context) error {
	// Use stdio transport for local communication
	return s.server.Run(ctx, &mcp.StdioTransport{})
}
