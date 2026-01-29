# Local Brain

> A minimalist, local-first project management system for developers who live in the terminal.

Local Brain is a context manager for your workflow. It stitches together your notes, tasks, and code repositories into a cohesive, keyboard-driven environment.

## Install

**Recommended: Homebrew (macOS/Linux)**

```bash
brew tap SanderMoon/tap
brew install brain
```

**Or via Go:**

```bash
go install github.com/SanderMoon/local-brain@latest
```

See [INSTALL.md](INSTALL.md) for more installation options and details.

## Concepts

Before getting started, it is helpful to understand the hierarchy of Local Brain:

- **Brains**: These are your high-level workspaces (e.g., "Work", "Personal", "Research"). Each brain is a completely separate directory tree. Only one brain is active at a time, which is symlinked to your home directory for easy access.
- **Projects**: These are specific areas of focus within a Brain. Every project automatically gets its own directory with `notes.md`, `todo.md`, and optional links to external code repositories.

## Core Philosophy: Capture Fast, Curate Later

Local Brain is built around a two-phase workflow that keeps you in flow:

### Phase 1: Capture (Speed is King)

When ideas strike, capture them instantly without breaking focus:

```bash
brain add "Fix authentication bug"
brain add "Call client about proposal"
brain add -n "Idea: Use Redis for session storage"
brain add # Write a note an editor like nvim
```

**Key principles:**

- **< 1 second:** Zero friction brain dumping
- **No metadata required:** Just the thought, nothing else
- **No interruptions:** No prompts, no decisions, no context switching

Everything goes to your **dump** (`00_dump.md`) - an inbox for raw thoughts.

### Phase 2: Curate (Organization is King)

Later, when you have dedicated time, enrich and organize:

```bash
brain refile    # Move items from dump to projects
brain plan      # Add metadata: priorities, due dates, tags, states
```

**Key principles:**

- **Batch processing:** Handle multiple items in one session
- **Interactive workflows:** FZF-powered selection and prompting
- **Rich metadata:** Priorities, due dates, tags, task states

### The Three Commands

Local Brain's entire workflow revolves around three commands:

| Command | Phase | Purpose | When to Use |
|---------|-------|---------|-------------|
| `brain add` | Capture | Instant brain dump | Anytime an idea strikes |
| `brain refile` | Curate | Sort dump → projects | End of day / start of day |
| `brain plan` | Curate | Enrich tasks with metadata | Weekly planning sessions |

### Understanding the Data Model

**Brains** (Workspaces)

- Top-level contexts (e.g., "Work", "Personal", "Research")
- Only one brain active at a time (`~/brain` symlink)
- Completely separate directory trees
- Stored in `~/brains/`

**Projects** (Focus Areas)

- Specific areas within a brain (e.g., "website-redesign", "blog-posts")
- Each project has:
  - `notes.md` - Free-form notes and documentation
  - `todo.md` - Action items and task lists
  - `.repos` - Optional links to code repositories

**Todos** (Action Items)

- Markdown checkboxes in `todo.md` files
- Four states: `[ ]` open, `[>]` in-progress, `[-]` blocked, `[x]` done
- Optional metadata: `#p:1` priority, `#due:2026-02-15` deadline, `#bug` tags
- Example:

  ```markdown
  - [>] Fix auth bug #p:1 #due:2026-02-15 #bug #security
  ```

**Notes** (Knowledge)

- Free-form markdown in `notes.md` files
- Meeting notes, ideas, documentation, references
- Tagged with `#captured:YYYY-MM-DD` when refiled from dump

### Workflow Example

**Morning (Capture):**

```bash
brain add "Fix auth bug in login"
brain add "Review Sarah's PR"
brain add "Update deployment docs"
brain add -n "Meeting notes: discussed new architecture"
```

**End of Day (Curate - Refile):**

```bash
brain refile
# Interactively move:
# - "Fix auth bug" → backend-api project (as todo)
# - "Review Sarah's PR" → frontend project (as todo)
# - "Update deployment docs" → backend-api project (as todo)
# - "Meeting notes..." → architecture project (as note)
```

**Friday (Curate - Plan):**

```bash
brain plan
# Interactive prompts for each task:
# - Priority? 1 (high)
# - Due date? next-friday
# - Tags? bug security
# - State? in-progress
```

**Throughout the Week:**

```bash
brain todo ls --status in-progress --priority 1
brain todo ls --overdue
brain todo done <id>
brain todo start <id>
```

This workflow ensures you **never lose an idea** while also maintaining **organized, actionable projects**.

## Additional Features

- **Local-First**: Plain text Markdown, grep-able, and version-controllable.
- **Centralized Storage**: All data lives in `~/brains/`, easily syncable via Syncthing/Dropbox.
- **Zero-Friction Capture**: Rapidly add tasks to your inbox without context switching.
- **Dev Mode Automation** (under development): `brain go` launches a full development environment (Tmux + Venv).
- **Programmatic API**: JSON output and non-interactive commands for AI agents and scripts.

## Quick Start

### 1. Install

**Recommended: Homebrew (macOS/Linux)**

```bash
brew tap SanderMoon/tap
brew install brain
```

**Or via Go:**

```bash
go install github.com/SanderMoon/local-brain@latest
```

See [INSTALL.md](INSTALL.md) for more installation options and details.

### 2. Create a Project

Start by creating a workspace for your current focus.

```bash
brain project new my-idea
```

### 3. Start Working

Jump into your project context.

```bash
brain go
```

*Opens a new shell in the project directory (or a Tmux session if code is linked).*

## Daily Workflow

### Capture Everything

Don't break your flow. Quickly dump thoughts into your dump from anywhere.

```bash
brain add "Schedule dentist appointment"
brain add "Review architecture for my-idea"
```

### Process & Refile

Later, review your dump and move tasks to specific projects.

```bash
brain refile
```

### Pro Tip: `brain project clone`

If you starting a project from a single existing repo, use this shortcut:

```bash
# Creates project, links repo, and clones code in one command
brain project clone https://github.com/me/new-tool.git
```

## API Access (for AI Agents & Scripts)

Brain supports programmatic access via JSON output and non-interactive commands:

```bash
# List dump items with stable IDs
brain dump ls --json

# Discover projects
brain project list --json

# Refile items by ID (non-interactive)
brain refile <id> <project-name> [--as todo|note]
```

**Example workflow for AI agents:**

```bash
# Get item ID from dump
ID=$(brain dump ls --json | jq -r '.[0].id')

# Refile to focused project
brain refile "$ID" my-project
```

## Command Reference

| Command | Description |
|---------|-------------|
| `brain project new <name>`| Create a new project. |
| `brain project list [--json]` | List all active projects (Alias: `brain projects`). |
| `brain add "text"` | Quick capture to dump. |
| `brain add -n "text"` | Quick capture note (vs task) to dump. |
| `brain dump ls [--json]` | List dump items with IDs. |
| `brain refile [<id> <project>]` | Process dump items (interactive or by ID). |
| `brain plan` | Interactive batch task planning. |
| `brain todo` | Fuzzy search open tasks. |
| `brain go` | Enter project context (Shell or Tmux). |
| `brain new [name]` | Create a new brain in `~/brains/<name>`. |
| `brain switch [name]` | Switch active brain. |
| `brain rename <old> <new>`| Rename a brain folder and config. |
| `brain delete [name]` | Permanently delete a brain. |
| `brain project link <url>`| Link a git repository to the current project. |
| `brain project pull`| Clone/Update linked repositories. |
| `brain project clone <url>` | Import a git repo as a new project. |
| `brain import [path]` | Scan folder and register existing brains. |

## Customization

You can customize the storage and symlink locations by setting environment variables in your `.zshrc` or `.bashrc`:

```bash
# Root directory for all brains (default: ~/brains)
export BRAIN_ROOT="$HOME/Dropbox/Brains"

# Location of the active brain symlink (default: ~/brain)
export BRAIN_SYMLINK="$HOME/Desktop/ActiveBrain"
```

## Directory Structure

Data is centralized in `~/brains/` (or your `$BRAIN_ROOT`). The active brain is symlinked to `~/brain` (or your `$BRAIN_SYMLINK`) for easy access.

```
~/brains/
└── default/             # Your default brain
    ├── 00_dump.md
    ├── 01_active/
    │   └── project-name/
    │       ├── notes.md
    │       ├── todo.md
    │       └── .repos
    └── 99_archive/
```

## Configuration

Stored in `~/.config/brain/config.json`.

**Optional Dependencies:**

- `fzf` - Fuzzy finding
- `ripgrep` - Fast search
- `jq` - JSON processing
- `tmux` - Dev mode workspaces

Install with: `brew install fzf ripgrep jq tmux`

## Contributing

Contributions are welcome! See [docs/](docs/) for developer documentation.

## License

MIT

