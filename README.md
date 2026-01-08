# Local Brain 🧠

> A minimalist, local-first project management system for developers who live in the terminal.

**Local Brain** is more than just a note-taking tool; it's a context manager for your workflow. It stitches together your notes, tasks, and code repositories into a cohesive, keyboard-driven environment.

## ✨ Features

- **One-Command Import**: `brain clone <url>` handles project creation, git linking, and cloning in one go.
- **Project Context**: Seamlessly link Git repositories to your project notes.
- **Dev Mode Automation**: `brain go` launches a full development environment:
    - Auto-detects `tmux` and creates a named session for the project.
    - **Window 1 (Code)**: Opens repo, activates `venv`, and launches your editor.
    - **Window 2 (Brain)**: Opens project notes and todos.
- **Local-First**: Plain text Markdown, grep-able, and version-controllable.
- **Zero-Friction Capture**: Rapidly add tasks to your inbox without context switching.

## 🚀 Quick Start

### 1. Install
Installs dependencies (`fzf`, `rg`, `jq`, `syncthing`, `make`), configures your shell, and sets up your first brain.

```bash
git clone https://github.com/YOUR_USERNAME/local-brain.git
cd local-brain
./install-brain.sh
```

### 2. Import Your First Project
The fastest way to get started is to import an existing Git repository.

```bash
brain clone https://github.com/yourname/your-project.git
```

### 3. Start Working
```bash
brain go
```
This will open your dev environment (Tmux + Editor + Venv) for that project.

##  Workflow

### Capture Everything
Don't break your flow. Quickly dump thoughts into your inbox.
```bash
brain add "Review PR #42 for the api-service"
```

### Navigation
We install a smart shell function `pg` ("Project Go") for rapid directory switching.
```bash
pg  # Fuzzy find and cd into any active project
```

### Remote Management
Manage your projects from anywhere in your terminal without leaving your home folder.
```bash
brain project list              # See all active projects
brain project select my-app     # Focus on a project
brain project link <url>        # Links to the focused project
brain project pull              # Clones/Updates focused project repos
```

## 🛠 Command Reference

| Command | Description |
|---------|-------------|
| `brain clone <url>` | **[New]** One-command project import. |
| `brain add "text"` | Quick capture to inbox. |
| `brain todo` | Fuzzy search all unchecked tasks across projects. |
| `brain go` | Enter project context (Shell or Tmux + Venv). |
| `brain refile` | Interactive inbox processing (GTD style). |
| `brain project list`| List active projects. |
| `brain project select`| Focus on a project for remote commands. |
| `brain switch` | Switch between different brains (work/personal). |

## 📦 Directory Structure

Each brain follows a modified PARA method:

```
~/brain/
├── 00_inbox.md          # Quick captures
├── 01_active/           # Active projects
│   └── project-name/
│       ├── notes.md     # Documentation
│       ├── todo.md      # Tasks
│       └── .repos       # Linked git URLs
├── 02_areas/            # Ongoing responsibilities
├── 03_resources/        # Reference material
└── 99_archive/          # Completed projects
```

## 🔧 Configuration

Stored in `~/.config/brain/config.json`.

**Dependencies:**
`fzf`, `ripgrep`, `jq`, `make`. (Optional: `tmux` for Dev Mode).

## License

MIT
