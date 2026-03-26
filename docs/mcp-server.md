# Local Brain MCP Server

MCP server exposing Local Brain to AI assistants (Claude Desktop, Claude Code, etc.).

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

### Claude Code (Project-Specific)

To add the Local Brain MCP server for a specific project in Claude Code:

```bash
claude mcp add --transport stdio local-brain /opt/homebrew/bin/brain-mcp
```

Run this from your project directory. The server is registered at the project level (stored in `.claude/mcp.json`) and Claude Code will have access to all Local Brain tools when working in that project.

**Path examples** (adjust to your installation):
- Homebrew (macOS ARM): `/opt/homebrew/bin/brain-mcp`
- Homebrew (macOS Intel): `/usr/local/bin/brain-mcp`
- User install: `/Users/<you>/.local/bin/brain-mcp`

## Available Tools (16)

### Context Retrieval & Search (6 tools)

**`get_brain_overview`** - Complete workspace overview: brain name, focused project, all projects with stats, inbox count

**`get_dump_items`** - All inbox items with IDs, content, and timestamps

**`get_project_context`** - Full project details: description, todos, repos, notes
- `project_name` (string) - Project to retrieve
- `include_completed` (bool) - Include done tasks
- `include_note_content` (string, optional) - Note content depth: none|preview|full (default: preview)
- `section` (string, optional) - PARA section: 01_active (default), 02_areas, or 03_resources

**`search`** - Search across todos and notes with filtering by status, priority, tags, and dates
- `query` (string, optional) - Search term
- `project` (string, optional) - Project filter
- `include_todos` (bool, optional) - Include todos (default: true)
- `include_notes` (bool, optional) - Include notes (default: true)
- `search_note_content` (bool, optional) - Search note body text, not just titles
- `status` (string, optional) - Filter todos by status (backlog, open, in-progress, blocked, done)
- `priority` (int, optional) - Filter todos by priority (1=high, 2=medium, 3=low)
- `tags` ([]string, optional) - Filter todos by tags (OR logic)
- `include_completed` (bool, optional) - Include completed tasks (default: false)
- `created_after`, `created_before` (strings, optional) - Filter by created date (YYYY-MM-DD)
- `completed_after`, `completed_before` (strings, optional) - Filter by completed date (YYYY-MM-DD)
- `due_after`, `due_before` (strings, optional) - Filter by due date (YYYY-MM-DD)
- `section` (string, optional) - PARA section: 01_active (default), 02_areas, or 03_resources

**`set_context`** - Switch brain and/or set focused project (at least one required)
- `brain_name` (string, optional) - Brain to switch to
- `project_name` (string, optional) - Project to focus

**`get_daily_briefing`** - Daily briefing: overdue/due items, high-priority tasks, in-progress work, blocked items, recent completions, inbox
- `section` (string, optional) - PARA section: 01_active (default), 02_areas, or 03_resources

### Quick Capture (2 tools)

**`add_to_dump`** - Capture a task or note to the inbox
- `type` (string) - Item type: "task" or "note"
- `content` (string) - Task or note content (multi-line supported for notes)
- `title` (string, optional) - Note title (required when type="note", ignored for tasks)

**`refile_item`** - Move items from inbox to projects (batch supported)
- `refiles` (array) - Refile operations, each with:
  - `item_id` (string) - 6-character hex ID
  - `project_name` (string) - Target project

### Todo Management (2 tools)

**`update_todo`** - Update todos: content, status, priority, due date, tags, or delete (batch supported)
- `updates` (array) - Todo updates, each with:
  - `todo_id` (string) - 6-character hex ID
  - `content` (string, optional) - New task text
  - `status` (string, optional) - backlog | open | in-progress | blocked | done
  - `priority` (int, optional) - 1 (high), 2 (medium), 3 (low), null to clear
  - `due_date` (string, optional) - YYYY-MM-DD format, or "" to clear
  - `add_tags` ([]string, optional) - Tags to add
  - `remove_tags` ([]string, optional) - Tags to remove
  - `delete` (bool, optional) - Permanently delete this todo
- `section` (string, optional) - PARA section: 01_active (default), 02_areas, or 03_resources

**`create_todo_in_project`** - Add tasks directly to a project (batch supported)
- `project_name` (string) - Target project
- `todos` (array) - Todos to create, each with:
  - `content` (string) - Task description
  - `priority` (int, optional) - 1 (high), 2 (medium), 3 (low)
  - `due_date` (string, optional) - YYYY-MM-DD format
  - `tags` ([]string, optional) - Tags
  - `status` (string, optional) - Initial status: backlog (default), open, in-progress, blocked
- `section` (string, optional) - PARA section: 01_active (default), 02_areas, or 03_resources

### Note Management (3 tools)

**`create_project_note`** - Create a timestamped note file in a project
- `project_name` (string) - Target project
- `title` (string) - Note title
- `content` (string) - Note content
- `section` (string, optional) - PARA section: 01_active (default), 02_areas, or 03_resources

**`get_note_content`** - Read a note file's content
- `note_path` (string) - Full path to note file

**`update_note`** - Replace full content of an existing note file
- `note_path` (string) - Full path to note file
- `content` (string) - New full content for the note file

### Project Management (2 tools)

**`create_project`** - Create new project
- `name` (string) - Project name (alphanumeric, hyphens, underscores)
- `description` (string, optional) - Project description
- `section` (string, optional) - PARA section: 01_active (default), 02_areas, or 03_resources

**`archive_project`** - Move a project from active to archive (timestamped)
- `name` (string) - Name of the project to archive

### Daily Notes (1 tool)

**`create_daily_note`** - Create or open a daily note in `{brain}/00_daily/YYYY-MM-DD.md` with overdue todo briefing
- `date` (string, optional) - Date in YYYY-MM-DD format (defaults to today)

## Usage Examples

### Get started
```
Ask Claude: "What am I working on?"
→ Calls get_brain_overview

Ask Claude: "Give me my daily briefing"
→ Calls get_daily_briefing
```

### Process inbox
```
Ask Claude: "Help me process my inbox"
→ Calls get_dump_items
→ Asks where each item should go
→ Calls refile_item(refiles: [{item_id: "abc123", project_name: "backend"}, ...])
```

### Manage tasks
```
Ask Claude: "Show me all high priority tasks"
→ Calls search(priority: 1)

Ask Claude: "Mark task abc123 as in-progress with high priority"
→ Calls update_todo(updates: [{todo_id: "abc123", status: "in-progress", priority: 1}])

Ask Claude: "What's overdue?"
→ Calls search(due_before: "2026-03-25")
```

### Create content
```
Ask Claude: "Add a task to my backend project: Fix auth bug"
→ Calls create_todo_in_project(project_name: "backend", todos: [{content: "Fix auth bug"}])

Ask Claude: "Write up notes from today's standup in project X"
→ Calls create_project_note(project_name: "X", title: "Standup notes", content: "...")
```

## Design Principles

- **Efficiency** - Batch operations minimize round-trips (`get_brain_overview` replaces 5+ separate calls)
- **Safety** - Destructive operations (`update_todo` with `delete` flag) are explicit. No brain deletion exposed.
- **Privacy** - Local stdio transport only, no network access.
- **Validation** - All inputs validated before execution with clear error messages.

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
- Check item IDs are valid 6-character hex (use get_dump_items or search to find IDs)
- Ensure project names match exactly (use get_brain_overview to see available projects)

**Validation errors**
- Todo IDs must be 6-character lowercase hex (e.g., "abc123" not "ABC123")
- Dates must be YYYY-MM-DD format (e.g., "2026-02-15")
- Priority must be 1, 2, or 3
- Status must be: backlog, open, in-progress, blocked, or done

## Technical Details

**Transport**: stdio (local only)
**Cache**: 30-second TTL for read operations, invalidated on mutations
**Session**: Thread-safe with automatic config refresh
**Version**: Uses Local Brain CLI version
