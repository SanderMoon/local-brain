# Local Brain

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

**[📖 Full Documentation →](https://sandermoon.github.io/local-brain/)**

---

## Core Philosophy: Capture Fast, Curate Later

Local Brain follows a two-phase workflow:

**Phase 1: Capture** (< 1 second)
- No metadata, no decisions, no interruptions
- Everything goes to your dump (`00_dump.md`)

**Phase 2: Curate** (dedicated time blocks)
- Batch process items with `brain refile`
- Enrich tasks with `brain plan` (priorities, due dates, tags)

This keeps you in flow while maintaining organized projects.

---

## Features

- **Zero-Friction Capture** - Add tasks in < 1 second without context switching
- **Batch Curation** - Process and organize during dedicated time blocks
- **Local-First** - Plain text Markdown files, grep-able, version-controllable
- **Developer-Friendly** - Integrates with git repos, supports JSON API for scripts
- **Privacy-First** - Everything lives locally in `~/brains/`, syncable via Syncthing/Dropbox

---

## Documentation

- **[🚀 Quickstart Guide](https://sandermoon.github.io/local-brain/)** - Get started in 3 minutes
- **[📦 Installation](https://sandermoon.github.io/local-brain/installation/)** - All installation methods
- **[📖 Command Reference](https://sandermoon.github.io/local-brain/commands/)** - Complete command documentation
- **[💻 Development Guide](https://sandermoon.github.io/local-brain/development/)** - Contributing and architecture

---

## Daily Workflow Example

**Morning** (Capture):
```bash
brain add "Fix auth bug in login"
brain add "Review Sarah's PR"
brain add "Update deployment docs"
```

**End of Day** (Curate - Refile):
```bash
brain refile
# Interactive prompts move items to projects:
# - "Fix auth bug" → backend-api
# - "Review Sarah's PR" → frontend
# - "Update deployment docs" → backend-api
```

**Friday** (Curate - Plan):
```bash
brain plan
# Add priorities, due dates, tags, states
```

**Throughout the Week**:
```bash
brain todo ls --status in-progress --priority 1
brain todo ls --overdue
brain todo done <id>
```

---

## Key Concepts

**Brains**: Top-level workspaces (e.g., "Work", "Personal"). Only one active at a time, symlinked to `~/brain`.

**Projects**: Focus areas within a brain (e.g., "website-redesign"). Each has `notes.md`, `todo.md`, and optional code repo links.

**Dump**: Your inbox (`00_dump.md`) for rapid capture. Process it regularly with `brain refile`.

**[Learn more →](https://sandermoon.github.io/local-brain/)**

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

- **Issues**: [GitHub Issues](https://github.com/SanderMoon/local-brain/issues)
- **Discussions**: [GitHub Discussions](https://github.com/SanderMoon/local-brain/discussions)
- **Documentation**: [https://sandermoon.github.io/local-brain/](https://sandermoon.github.io/local-brain/)

---

## License

MIT License - See [LICENSE](LICENSE) file for details.

---

**[📖 Read the Full Documentation](https://sandermoon.github.io/local-brain/)**
