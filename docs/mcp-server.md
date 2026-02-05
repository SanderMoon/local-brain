# Local Brain MCP Server

Model Context Protocol server that exposes Local Brain functionality to AI assistants like Claude Desktop.

## Quick Start

### Installation

**Via Homebrew (macOS, recommended):**
```bash
# Install Local Brain (includes both CLI and MCP server)
brew install sandermoon/tap/brain

# brain-mcp is automatically installed to $(brew --prefix)/bin/brain-mcp
```

**Via Makefile (all platforms):**
```bash
# Install to ~/.local/bin (no sudo required, recommended)
make dev-install-mcp

# Or install system-wide to /usr/local/bin
sudo make install-mcp

# Or specify custom location
make install-mcp INSTALL_DIR=~/.local/bin

# Or install both CLI and MCP server
make install-all
```

**Via Go (if you have Go installed):**
```bash
# Clone the repository
git clone https://github.com/SanderMoon/local-brain.git
cd local-brain

# Build and install
make build-mcp
cp brain-mcp /usr/local/bin/  # or ~/.local/bin/
```

### Configuration

Claude Desktop config location depends on your OS:
- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

Add this configuration (adjust path based on your installation method):

```json
{
  "mcpServers": {
    "local-brain": {
      "command": "/usr/local/bin/brain-mcp"
    }
  }
}
```

**Path examples:**
- Homebrew (macOS): `/opt/homebrew/bin/brain-mcp` or `/usr/local/bin/brain-mcp`
- User install: `~/.local/bin/brain-mcp` (expand ~ to full path)
- System install: `/usr/local/bin/brain-mcp`

Restart Claude Desktop to activate.

## Available Tools (17)

### Context Retrieval & Search (7 tools)

**`get_brain_overview`** - Complete workspace overview (brain name, focused project, all projects with stats, inbox count)

**`get_dump_items`** - All inbox items with IDs, content, and timestamps

**`get_all_todos`** - All tasks across projects with metadata (status, priority, due date, tags, capture/complete dates)
- `include_completed` (bool) - Include done tasks
- `project` (string, optional) - Filter by project name
- `created_after` (string, optional) - Filter by capture date >= YYYY-MM-DD
- `created_before` (string, optional) - Filter by capture date <= YYYY-MM-DD
- `completed_after` (string, optional) - Filter by completion date >= YYYY-MM-DD
- `completed_before` (string, optional) - Filter by completion date <= YYYY-MM-DD

**`get_project_context`** - Complete project details (description, todos, repos, notes with optional content)
- `project_name` (string) - Project to retrieve
- `include_completed` (bool) - Include done tasks
- `include_note_content` (string, optional) - Note content: none|preview|full (default: preview)

**`search_todos`** - Filter todos by query, project, status, tags, and dates
- `query` (string, optional) - Search term
- `project` (string, optional) - Project filter
- `status` (string, optional) - Status filter (open, in-progress, blocked, done)
- `tags` ([]string, optional) - Tag filter (OR logic)
- `created_after` (string, optional) - Filter by capture date >= YYYY-MM-DD
- `created_before` (string, optional) - Filter by capture date <= YYYY-MM-DD
- `completed_after` (string, optional) - Filter by completion date >= YYYY-MM-DD
- `completed_before` (string, optional) - Filter by completion date <= YYYY-MM-DD

**`search`** - Unified search across todos and notes with temporal filtering
- `query` (string, optional) - Search term
- `project` (string, optional) - Project filter
- `include_todos` (bool, optional) - Include todos (default: true)
- `include_notes` (bool, optional) - Include notes (default: true)
- `search_note_content` (bool, optional) - Search note content not just titles
- `created_after`, `created_before`, `completed_after`, `completed_before` (strings, optional) - Temporal filters

**`switch_brain`** - Switch to different brain workspace
- `brain_name` (string) - Brain to activate

### Quick Capture (3 tools)

**`add_task_to_dump`** - Add task to inbox
- `content` (string) - Task description

**`add_note_to_dump`** - Add note to inbox
- `title` (string) - Note title
- `content` (string) - Note content (can be multi-line)

**`refile_item`** - Move item from inbox to project
- `item_id` (string) - 6-character hex ID
- `project_name` (string) - Target project

### Todo Management (3 tools)

**`update_todo`** - Update todo status and/or metadata in one call
- `todo_id` (string) - 6-character hex ID
- `status` (string, optional) - open | in-progress | blocked | done
- `priority` (int, optional) - 1 (high), 2 (medium), 3 (low), null to clear
- `due_date` (string, optional) - YYYY-MM-DD format or empty to clear
- `add_tags` ([]string, optional) - Tags to add
- `remove_tags` ([]string, optional) - Tags to remove

**`create_todo_in_project`** - Add task directly to project
- `project_name` (string) - Target project
- `content` (string) - Task description

**`delete_todo`** - Delete task permanently (requires user confirmation)
- `todo_id` (string) - 6-character hex ID

### Note Management (2 tools)

**`create_project_note`** - Create timestamped note file in project
- `project_name` (string) - Target project
- `title` (string) - Note title
- `content` (string) - Note content

**`get_note_content`** - Read note file content
- `note_path` (string) - Full path to note file

### Project Management (2 tools)

**`create_project`** - Create new project
- `name` (string) - Project name (alphanumeric, hyphens, underscores)
- `description` (string, optional) - Project description

**`set_focused_project`** - Set the focused project
- `project_name` (string) - Project to focus

## Usage Examples

### Get started
```
Ask Claude: "What am I working on?"
→ Calls get_brain_overview
```

### Process inbox
```
Ask Claude: "Help me process my inbox"
→ Calls get_dump_items
→ Asks where each item should go
→ Calls refile_item for each
```

### Manage tasks
```
Ask Claude: "Show me all high priority tasks"
→ Calls get_all_todos
→ Filters by priority

Ask Claude: "Mark task abc123 as in-progress with high priority"
→ Calls update_todo(todo_id: "abc123", status: "in-progress", priority: 1)

Ask Claude: "What's overdue?"
→ Calls get_all_todos
→ Filters by due date
```

### Create content
```
Ask Claude: "Add a task to my backend project: Fix auth bug"
→ Calls create_todo_in_project(project_name: "backend", content: "Fix auth bug")

Ask Claude: "Add a note about today's standup to project X"
→ Calls create_project_note with meeting notes
```

## Design Principles

**Efficiency** - Batch operations minimize round-trips (e.g., `get_brain_overview` replaces 5+ separate calls)

**Safety** - Destructive operations (delete_todo) require user confirmation. No project/brain deletion exposed.

**Privacy** - Local stdio transport only, no network access

**Validation** - All inputs validated before execution with clear error messages

## Troubleshooting

**Tools not appearing in Claude Desktop**
- Verify config file location: `~/Library/Application Support/Claude/claude_desktop_config.json`
- Check binary path is correct: `which brain-mcp`
- Restart Claude Desktop completely

**Permission errors**
- Ensure binary is executable: `chmod +x /usr/local/bin/brain-mcp`
- Check brain directory permissions

**Tool calls failing**
- Verify you have an active brain configured: `brain config show`
- Check item IDs are valid 6-character hex (use get_dump_items or get_all_todos to find IDs)
- Ensure project names match exactly (use list_projects to see available projects)

**Validation errors**
- Todo IDs must be 6-character lowercase hex (e.g., "abc123" not "ABC123")
- Dates must be YYYY-MM-DD format (e.g., "2026-02-15")
- Priority must be 1, 2, or 3
- Status must be: open, in-progress, blocked, or done

## Technical Details

**Transport**: stdio (local only)
**Cache**: 30-second TTL for read operations, invalidated on mutations
**Session**: Thread-safe with automatic config refresh
**Version**: Uses Local Brain CLI version
