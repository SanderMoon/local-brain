# Local Brain

```
 ______     ______     ______     __     __   __
/\  == \   /\  == \   /\  __ \   /\ \   /\ "-.\ \
\ \  __<   \ \  __<   \ \  __ \  \ \ \  \ \ \-.  \
 \ \_____\  \ \_\ \_\  \ \_\ \_\  \ \_\  \ \_\\"\_\
  \/_____/   \/_/ /_/   \/_/\/_/   \/_/   \/_/ \/_/
```

> A local-first project management tool — with a CLI/TUI for humans and an MCP server for AI agents.

[![Documentation](https://img.shields.io/badge/docs-sandermoon.github.io-blue)](https://sandermoon.github.io/local-brain/)
[![Go Version](https://img.shields.io/github/go-mod/go-version/SanderMoon/local-brain)](https://github.com/SanderMoon/local-brain)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Local Brain keeps your projects, tasks, and notes in plain Markdown files on your own machine. Use it yourself via the CLI and TUI, or let an AI agent manage your projects through the built-in MCP server — or both.

---

## Two Interfaces, One System

### Human Interface — CLI & TUI

![Local Brain TUI](docs/images/tui-screenshot.png)

- **Zero-friction capture** — `brain add "..."` takes under a second, no context switching
- **Batch curation** — process and organize during dedicated time blocks with `brain refile` and `brain plan`
- **Full CLI** — every operation is a command; scriptable and composable
- **Optional TUI** — visual project and task overview with `brain tui`

### AI Interface — MCP Server + Agent Skills

<!-- TODO: Add screenshot of chatting with Claude about projects/todos. Save to docs/images/claude-chat-screenshot.png -->
![Claude chat with Local Brain](docs/images/claude-chat-screenshot.png)

- **18 purpose-built MCP tools** — designed to minimize LLM round-trips while giving complete access to your projects
- **7 bundled [Agent Skills](https://agentskills.io)** — teach your AI agent *how* to be a project manager: daily briefings, inbox triage, weekly reviews, project planning, and more
- **Personal assistant** — run `brain storm` to open Claude in your brains directory for an interactive session
- **Works with any skills-compatible agent** — Claude Code, Codex, Gemini CLI, OpenCode

---

## Quick Start

### Install

```bash
# macOS/Linux (Homebrew)
brew tap SanderMoon/tap
brew install brain

# Or build from source
git clone https://github.com/SanderMoon/local-brain.git && cd local-brain
make install && make install-mcp
```

### Set up

```bash
brain init                   # create your first brain
```

### Connect your AI agent

```bash
brain mcp install            # register the MCP server with detected agents
brain skill install          # install skills (daily briefing, triage, planning, ...)
```

This auto-detects your installed agents (Claude Code, Codex, Gemini CLI, OpenCode) and configures both the MCP server and skills for each. You can target a specific agent with `--agent claude`.

### Use it

```bash
# CLI: capture → organize → act
brain add "Fix authentication bug"
brain refile                        # move inbox items to projects
brain todo ls --priority 1          # see what matters today

# AI: just talk to your agent
# "Good morning"                    → daily briefing
# "Process my inbox"                → guided triage
# "Let's plan the website redesign" → project breakdown
```

**[📖 Full Documentation →](https://sandermoon.github.io/local-brain/)**

---

## Core Workflow: Capture → Curate

**Phase 1: Capture** (< 1 second)

Dump everything to your inbox — no metadata, no decisions, no interruptions.

```bash
brain add "Fix auth bug in login"
brain add "Review Sarah's PR"
brain add "Update deployment docs"
```

**Phase 2: Curate** (dedicated time blocks)

Organize, prioritize, and act on your items in batch.

```bash
brain refile      # move inbox items to projects
brain plan        # add priorities, due dates, tags
brain todo ls     # see your task list
```

Or let your AI agent handle it — with skills installed, just say "process my inbox" and it walks you through each item.

---

## Agent Skills

Local Brain ships 7 bundled skills that follow the [Agent Skills open standard](https://agentskills.io). Skills teach your AI agent structured workflows — turning it from a generic chatbot into a capable project manager.

| Skill | What it does |
|-------|-------------|
| `brain-daily` | Morning briefing — priorities, overdue tasks, blocked items |
| `brain-setup` | First-time setup wizard — conversational onboarding |
| `brain-capture` | Quick capture with smart metadata inference |
| `brain-triage` | Systematic inbox processing — refile, categorize, discard |
| `brain-plan` | Break goals into concrete tasks with priorities and deadlines |
| `brain-focus` | Deep work session with distraction capture |
| `brain-review` | Weekly review of progress, stale work, and priorities |

```bash
brain skill install          # install all skills to detected agents
brain skill upgrade          # update skills after a brain version upgrade
brain skill status           # check what's installed where
```

Together, these skills form a complete **Capture → Organize → Plan → Execute → Reflect** loop — inspired by GTD, adapted for AI-assisted workflows.

> **Customizing skills:** Installed skills live as plain SKILL.md files in your agent's skills directory (e.g., `~/.claude/skills/`). You're free to edit them — but note that `brain skill upgrade` will overwrite your changes. To preserve customizations, either skip upgrading that skill or keep your edits in a separate file alongside SKILL.md.

---

## Key Concepts

**Brains**: Top-level workspaces (e.g., "Work", "Personal"). Only one is active at a time, symlinked to `~/brain`.

**Projects**: Focus areas within a brain (e.g., "website-redesign"). Each has `notes.md`, `todo.md`, and optional git repo links.

**Dump**: Your inbox (`00_dump.md`) for rapid capture. Process it regularly with `brain refile`.

---

## Documentation

- **[🚀 Quickstart Guide](https://sandermoon.github.io/local-brain/)** — Get started in 3 minutes
- **[📦 Installation](https://sandermoon.github.io/local-brain/installation/)** — All installation methods
- **[📖 Command Reference](https://sandermoon.github.io/local-brain/commands/)** — Complete command documentation
- **[🤖 MCP Server Setup](docs/mcp-server.md)** — AI agent integration (Claude Desktop, Claude Code)
- **[💻 Development Guide](https://sandermoon.github.io/local-brain/development/)** — Architecture and contributing

---

## Contributing

Contributions are welcome! See the [Development Guide](https://sandermoon.github.io/local-brain/development/) for setup, architecture, and contributing guidelines.

```bash
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

MIT License — See [LICENSE](LICENSE) file for details.
