# Local Brain

```
 ______     ______     ______     __     __   __
/\  == \   /\  == \   /\  __ \   /\ \   /\ "-.\ \
\ \  __<   \ \  __<   \ \  __ \  \ \ \  \ \ \-.  \  
 \ \_____\  \ \_\ \_\  \ \_\ \_\  \ \_\  \ \_\\"\_\
  \/_____/   \/_/ /_/   \/_/\/_/   \/_/   \/_/ \/_/
```

> A minimalist, local-first project management system for developers who live in the terminal.

[![Documentation](https://img.shields.io/badge/docs-sandermoon.github.io-blue)](https://sandermoon.github.io/local-brain/)
[![Go Version](https://img.shields.io/github/go-mod/go-version/SanderMoon/local-brain)](https://github.com/SanderMoon/local-brain)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Local Brain is a context manager for your workflow. It stitches together your notes, tasks, and code repositories into a cohesive, keyboard-driven environment.

---

## Quick Start

### Install

```bash
# macOS/Linux (Homebrew)
brew tap SanderMoon/tap
brew install brain

# Or via Go
go install github.com/SanderMoon/local-brain@latest
```

### Initialize

```bash
brain init
```

### Start Using

```bash
# Capture thoughts instantly
brain add "Fix authentication bug"
brain add "Review Sarah's PR"

# Process and organize later
brain refile

# Weekly planning
brain plan
```

__[📖 Full Documentation →](https://sandermoon.github.io/local-brain/)__

---

## Core Philosophy: Capture Fast, Curate Later

Local Brain follows a two-phase workflow:

__Phase 1: Capture__ (< 1 second)

- No metadata, no decisions, no interruptions
- Everything goes to your dump (`00_dump.md`)

__Phase 2: Curate__ (dedicated time blocks)

- Batch process items with `brain refile`
- Enrich tasks with `brain plan` (priorities, due dates, tags)

This keeps you in flow while maintaining organized projects.

---

## Features

- __Zero-Friction Capture__ - Add tasks in < 1 second without context switching
- __Batch Curation__ - Process and organize during dedicated time blocks
- __Local-First__ - Plain text Markdown files, grep-able, version-controllable
- __Developer-Friendly__ - Integrates with git repos, supports JSON API for scripts
- __Privacy-First__ - Everything lives locally in `~/brains/`, syncable via Syncthing/Dropbox

---

## Documentation

- __[🚀 Quickstart Guide](https://sandermoon.github.io/local-brain/)__ - Get started in 3 minutes
- __[📦 Installation](https://sandermoon.github.io/local-brain/installation/)__ - All installation methods
- __[📖 Command Reference](https://sandermoon.github.io/local-brain/commands/)__ - Complete command documentation
- __[💻 Development Guide](https://sandermoon.github.io/local-brain/development/)__ - Contributing and architecture

---

## Daily Workflow Example

__Morning__ (Capture):

```bash
brain add "Fix auth bug in login"
brain add "Review Sarah's PR"
brain add "Update deployment docs"
```

__End of Day__ (Curate - Refile):

```bash
brain refile
# Interactive prompts move items to projects:
# - "Fix auth bug" → backend-api
# - "Review Sarah's PR" → frontend
# - "Update deployment docs" → backend-api
```

__Friday__ (Curate - Plan):

```bash
brain plan
# Add priorities, due dates, tags, states
```

__Throughout the Week__:

```bash
brain todo ls --status in-progress --priority 1
brain todo ls --overdue
brain todo done <id>
```

---

## Key Concepts

__Brains__: Top-level workspaces (e.g., "Work", "Personal"). Only one active at a time, symlinked to `~/brain`.

__Projects__: Focus areas within a brain (e.g., "website-redesign"). Each has `notes.md`, `todo.md`, and optional code repo links.

__Dump__: Your inbox (`00_dump.md`) for rapid capture. Process it regularly with `brain refile`.

__[Learn more →](https://sandermoon.github.io/local-brain/)__

---

## Contributing

Contributions are welcome! See the [Development Guide](https://sandermoon.github.io/local-brain/development/) for setup, architecture, and contributing guidelines.

```bash
# Quick setup for contributors
git clone https://github.com/SanderMoon/local-brain.git
cd local-brain
make build
make test
```

---

## Community

- __Issues__: [GitHub Issues](https://github.com/SanderMoon/local-brain/issues)
- __Discussions__: [GitHub Discussions](https://github.com/SanderMoon/local-brain/discussions)
- __Documentation__: [https://sandermoon.github.io/local-brain/](https://sandermoon.github.io/local-brain/)

---

## License

MIT License - See [LICENSE](LICENSE) file for details.

---

__[📖 Read the Full Documentation](https://sandermoon.github.io/local-brain/)__
