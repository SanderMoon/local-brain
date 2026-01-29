# Local Brain

> A minimalist, local-first project management system for developers who live in the terminal.

Local Brain is a context manager for your workflow. It stitches together your notes, tasks, and code repositories into a cohesive, keyboard-driven environment.

## 30-Second Install

```bash
brew tap SanderMoon/tap
brew install brain
brain init
```

That's it. Start capturing immediately:

```bash
brain add "Fix authentication bug"
brain add -n "Meeting notes from standup"
```

## Core Philosophy: Capture Fast, Curate Later

Local Brain follows a two-phase workflow designed to keep you in flow state:

### Phase 1: Capture (Speed is King)

When ideas strike, capture them instantly without breaking focus. No prompts, no decisions, no context switching.

```bash
brain add "Fix auth bug"
brain add "Call client"
brain add -n "Use Redis for sessions"
```

Everything goes to your **dump** (`00_dump.md`) - an inbox for raw thoughts. Capturing takes < 1 second.

### Phase 2: Curate (Organization is King)

Later, when you have dedicated time, organize and enrich your captured items:

```bash
brain refile    # Move items from dump to projects
brain plan      # Add priorities, due dates, tags
```

Batch process multiple items in one focused session using interactive FZF-powered workflows.

## Daily Practice Example

**Morning** (Capture throughout the day):

```bash
brain add "Fix auth bug in login"
brain add "Review Sarah's PR"
brain add -n "Architecture discussion notes"
```

**End of Day** (Refile):

```bash
brain refile
# Interactive prompts move items to appropriate projects
# "Fix auth bug" → backend-api
# "Review Sarah's PR" → frontend
# "Architecture notes" → architecture (as note)
```

**Friday** (Weekly planning):

```bash
brain plan
# Set priorities, due dates, tags, states for tasks
# Priority? 1 (high)
# Due date? next-friday
# Tags? bug security
```

**Throughout the Week**:

```bash
brain todo ls --status in-progress --priority 1
brain todo ls --overdue
brain todo done <id>
```

This workflow ensures you **never lose an idea** while maintaining **organized, actionable projects**.

## Key Concepts

**Brains**: Top-level workspaces (e.g., "Work", "Personal"). Only one brain active at a time, symlinked to `~/brain`.

**Projects**: Focus areas within a brain (e.g., "website-redesign"). Each has `notes.md`, `todo.md`, and optional code repo links.

**Dump**: Your inbox (`00_dump.md`) for rapid capture. Process it regularly with `brain refile`.

## Key Features

- **Local-First**: Plain text Markdown files, grep-able, version-controllable
- **Zero-Friction Capture**: Add tasks in < 1 second without context switching
- **Batch Curation**: Process and organize during dedicated time blocks
- **Developer-Friendly**: Integrates with git repos, supports JSON API for scripts
- **Privacy-First**: Everything lives locally in `~/brains/`, syncable via Syncthing/Dropbox

## Next Steps

- **[Installation Guide](installation.md)** - Complete installation instructions for all platforms
- **[Command Reference](commands.md)** - Full documentation of all 23 commands
- **[Development Guide](development.md)** - Architecture, contributing, and API docs

## Quick Reference: The Three Commands

| Command | Phase | Purpose | When to Use |
|---------|-------|---------|-------------|
| `brain add` | Capture | Instant brain dump | Anytime an idea strikes |
| `brain refile` | Curate | Sort dump → projects | End of day / start of day |
| `brain plan` | Curate | Enrich with metadata | Weekly planning sessions |

---

**Ready to get started?** Head to the [Installation Guide](installation.md) to set up Local Brain in under 5 minutes.
