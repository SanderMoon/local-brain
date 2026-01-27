# Local Brain

> A minimalist, local-first project management system for developers who live in the terminal.

Local Brain is a context manager for your workflow. It stitches together your notes, tasks, and code repositories into a cohesive, keyboard-driven environment.

## Concepts

Before getting started, it is helpful to understand the hierarchy of Local Brain:

- **Brains**: These are your high-level workspaces (e.g., "Work", "Personal", "Research"). Each brain is a completely separate directory tree. Only one brain is active at a time, which is symlinked to your home directory for easy access.
- **Projects**: These are specific areas of focus within a Brain. Every project automatically gets its own directory with `notes.md`, `todo.md`, and optional links to external code repositories.

## Features

- **Context Management**: Keep your notes, tasks, and related code in one place.
- **Local-First**: Plain text Markdown, grep-able, and version-controllable.
- **Centralized Storage**: All data lives in `~/brains/`, easily syncable via Syncthing/Dropbox.
- **Zero-Friction Capture**: Rapidly add tasks to your inbox without context switching.
- **Dev Mode Automation**: `brain go` launches a full development environment (Tmux + Venv).
- **Programmatic API**: JSON output and non-interactive commands for AI agents and scripts.

## Quick Start

### 1. Install

**Recommended: One-line install script**

```bash
curl -fsSL https://raw.githubusercontent.com/sandermoonemans/local-brain/main/install.sh | bash
```

**Or install via Homebrew (macOS):**

```bash
brew tap sandermoonemans/tap
brew install brain
```

**Or install via Go:**

```bash
go install github.com/sandermoonemans/local-brain@latest
```

See [INSTALL.md](INSTALL.md) for detailed installation options.

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

**Dependencies:**
`fzf`, `ripgrep`, `jq`, `make`. (Optional: `tmux` for Dev Mode).

## License

MIT