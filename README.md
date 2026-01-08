# Local Brain 🧠

> A minimalist, local-first project management system for developers who live in the terminal.

**Local Brain** is more than just a note-taking tool; it's a context manager for your workflow. It stitches together your notes, tasks, and code repositories into a cohesive, keyboard-driven environment.

## ✨ Features

- **Project Context**: seamlessly link Git repositories to your project notes.
- **Dev Mode Automation**: `brain go` launches a full development environment:
    - Auto-detects `tmux` and creates a named session for the project.
    - **Window 1 (Code)**: Opens repo, activates `venv`, and launches your editor.
    - **Window 2 (Brain)**: Opens project notes and todos.
- **Local-First**: Plain text Markdown, grep-able, and version-controllable.
- **Cloud-Free Sync**: Built-in support for Syncthing (or use iCloud/Dropbox).
- **Zero-Friction Capture**: Rapidly add tasks to your inbox without context switching.

## 🚀 Quick Start

### Option 1: The "Easy" Way (Interactive)

Installs dependencies, configures your shell, and sets up your first brain.

```bash
git clone https://github.com/YOUR_USERNAME/local-brain.git
cd local-brain
./install-brain.sh
```

### Option 2: The "Pro" Way (Make)

For package maintainers or system-wide installation.

```bash
sudo make install
# Then initialize manually:
brain init ~/brain
```

##  workflow

### 1. Capture Everything
Don't break your flow. Quickly dump thoughts into your inbox.

```bash
brain add "Review PR #42 for the api-service"
```

### 2. Manage Projects
Create a project and link your code.

```bash
# Create a new workspace
brain project new api-service

# Link the git repository (clones to ~/dev/api-service)
brain project link https://github.com/company/api-service.git
```

### 3. Enter "Dev Mode" ⚡️
This is where the magic happens. Run `brain go` and select a project.

- **If you have Tmux**: It detects linked repos, creates a session, activates your Python `venv`, and opens Neovim in the code *and* notes directories simultaneously.
- **If you don't**: It simply jumps you to the project directory in a new shell.

### 4. Navigation
We install a smart shell function `pg` ("Project Go") for rapid directory switching.

```bash
pg  # Fuzzy find and cd into any active project
```

## 🛠 Command Reference

| Command | Description |
|---------|-------------|
| `brain add "text"` | Quick capture to inbox. |
| `brain todo` | Fuzzy search all unchecked tasks across projects. |
| `brain go` | Enter project context (Shell or Tmux + Venv). |
| `brain refile` | Interactive inbox processing (GTD style). |
| `brain project new`| Create a new project structure. |
| `brain project link`| Link a Git URL to the current project. |
| `brain project pull`| Update/Clone all linked repositories. |
| `brain switch` | Switch between different brains (work/personal). |

## 📦 Directory Structure

Each brain follows a modified PARA method:

```
~/brain/
├── 00_inbox.md          # Quick captures
├── 01_active/           # Active projects
│   └── api-service/
│       ├── notes.md     # Documentation
│       ├── todo.md      # Tasks
│       └── .repos       # Linked git URLs
├── 02_areas/            # Ongoing responsibilities
├── 03_resources/        # Reference material
└── 99_archive/          # Completed projects
```

 Linked repositories are cloned to `~/dev/` by default.

## 🔧 Configuration

Stored in `~/.config/brain/config.json`.

**Dependencies (Installed automatically by script):**
- `fzf` (Fuzzy finding)
- `ripgrep` (Fast searching)
- `jq` (JSON processing)
- `tmux` (Optional, for Dev Mode)
- `neovim` / `vim`

## 🤝 Contributing

1. Fork it
2. Create your feature branch (`git checkout -b feature/cool-feature`)
3. Commit your changes (`git commit -am 'Add cool feature'`)
4. Push to the branch (`git push origin feature/cool-feature`)
5. Create a Pull Request

## License

MIT