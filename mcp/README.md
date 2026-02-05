# Local Brain MCP Server

Go implementation of the Model Context Protocol server for Local Brain.

## Overview

Exposes 18 tools across 5 categories:
- **Context** (6) - Brain/project/todo retrieval
- **Capture** (3) - Quick inbox operations
- **Todos** (4) - Task management
- **Notes** (2) - Note creation and reading
- **Projects** (3) - Project management

## Building

```bash
go build -o brain-mcp .
```

## Architecture

```
mcp/
├── main.go                 # Entry point, stdio transport
├── server.go               # MCP server setup, tool registration
├── tools/                  # Tool implementations (5 files)
│   ├── context.go          # Context retrieval tools
│   ├── dump.go             # Quick capture tools
│   ├── todos.go            # Todo management tools
│   ├── notes.go            # Note management tools
│   └── projects.go         # Project management tools
├── validation/             # Input validation
│   ├── validation.go       # Validators for all input types
│   ├── errors.go           # Structured error types
│   └── validation_test.go  # Validation tests (47 cases)
├── session/                # Session management
│   └── session.go          # Config caching, cache invalidation
└── testutil/               # Test helpers
    └── testutil.go         # Test brain setup utilities
```

## Key Design Decisions

**Direct API Import** - Imports `pkg/api` directly rather than shelling out to CLI (fast, no subprocess overhead)

**Validation First** - All tools validate inputs before calling API functions

**Cache Strategy** - 30s TTL for reads, immediate invalidation on writes

**Error Handling** - Structured errors with helpful context (e.g., available options for not-found errors)

**Thread Safety** - Session uses RWMutex for concurrent access

## Testing

**What's tested:**
- Validation layer (47 unit tests in `validation/validation_test.go`)
- API functions (comprehensive tests in `pkg/api/*_test.go`)

**What's not tested:**
- MCP tool handlers (thin wrappers, tested manually via Claude Desktop)

Rationale: Tool handlers are 5-10 line wrappers that validate + call API + format response. Testing them would duplicate API test coverage.

## Development

**Adding a new tool:**

1. Add tool registration in appropriate `tools/*.go` file
2. Define args struct with jsonschema tags
3. Add validation calls at start of handler
4. Call API function(s)
5. Call `sess.Invalidate()` if mutation
6. Return formatted result

Example:
```go
type MyToolArgs struct {
    TodoID string `json:"todo_id" jsonschema:"6-character hex ID"`
}
mcp.AddTool(srv, &mcp.Tool{
    Name:        "my_tool",
    Description: "Does something useful",
}, func(ctx context.Context, req *mcp.CallToolRequest, args MyToolArgs) (*mcp.CallToolResult, any, error) {
    // Validate
    if err := validation.ValidateTodoID(args.TodoID); err != nil {
        return nil, nil, err
    }

    // Call API
    result, err := api.DoSomething(args.TodoID)
    if err != nil {
        return nil, nil, fmt.Errorf("failed: %w", err)
    }

    // Invalidate cache if mutation
    sess.Invalidate()

    // Return result
    return &mcp.CallToolResult{
        Content: []mcp.Content{
            &mcp.TextContent{Text: fmt.Sprintf("Success: %v", result)},
        },
    }, nil, nil
})
```

## Dependencies

- `github.com/modelcontextprotocol/go-sdk` v1.2.0 - Official MCP SDK
- `github.com/sandermoonemans/local-brain/pkg/api` - Local Brain API
- `github.com/sandermoonemans/local-brain/pkg/config` - Config management

## See Also

- [User Documentation](../docs/mcp-server.md) - Setup and usage guide
- [Implementation Plan](../.claude/plans/happy-knitting-lark.md) - Full design document
- [MCP Specification](https://modelcontextprotocol.io/) - Protocol specification
