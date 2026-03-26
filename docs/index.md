# Local Brain

> Plain markdown files. Full project management. Managed by AI, or by you.

Local Brain turns a folder of markdown files into a project management system: a CLI, a TUI, and an AI agent interface. Your projects, tasks, and notes stay in plain text on your machine. You can manage them yourself, hand them off to an AI agent, or do both.

## Quick Start

```bash
brew tap SanderMoon/tap
brew install brain
brain init
```

### Connect your AI agent

```bash
brain mcp install            # register the MCP server with detected agents
brain skill install           # install agent skills (briefing, triage, planning, ...)
```

Auto-detects installed agents (Claude Code, Codex, Gemini CLI, OpenCode). Target a specific one with `--agent claude`.

### Start using it

```bash
brain add "Ship the new API endpoint"    # capture to inbox
brain                                     # launch the TUI
# Or open your agent and say "what's on my plate?"
```

## Three Ways to Use It

### 1. AI agent

Connect via the built-in MCP server and agent skills, then talk naturally:

- *"Good morning"* -- get a daily briefing with priorities, overdue tasks, and blockers
- *"Process my inbox"* -- guided triage of captured items
- *"Let's plan the website redesign"* -- break a project into tasks with priorities and deadlines

The agent reads and writes the same markdown files. No sync, no database, no lock-in.

### 2. CLI & TUI

```bash
brain add "Fix auth bug"            # capture to inbox (< 1 second)
brain refile                        # move inbox items to projects
brain todo ls --priority 1          # see high-priority tasks
brain todo done abc123              # mark a task done
brain plan                          # promote backlog tasks, set metadata
brain daily                         # create/open today's daily note
brain                               # launch the TUI
```

### 3. Your editor

Open `~/brains/work/` as an Obsidian vault, VS Code workspace, or just use vim. Edit the markdown directly; Local Brain picks up changes instantly.

## What It Looks Like on Disk

A **brain** is a workspace folder. A **project** is a subfolder with `todo.md` and `notes.md`. The **dump** (`00_dump.md`) is your inbox for rapid capture. Only one brain is active at a time, symlinked to `~/brain`.

```
~/brains/work/                          <- a brain (workspace)
├── 00_dump.md                          <- inbox for quick captures
├── 00_daily/                           <- daily notes
│   └── 2026-03-25.md
├── website-redesign/                   <- a project
│   ├── todo.md
│   ├── notes.md
│   └── .repos                          <- linked git repositories
└── api-migration/
    ├── todo.md
    └── notes.md
```

## Agent Skills

Local Brain ships 7 skills that follow the [Agent Skills open standard](https://agentskills.io). These teach your AI agent structured project management workflows.

| Skill | Trigger | What it does |
|-------|---------|-------------|
| `brain-daily` | *"Good morning"* | Briefing with priorities, overdue tasks, blocked items |
| `brain-capture` | *"Remind me to..."* | Quick capture with smart metadata inference |
| `brain-triage` | *"Process my inbox"* | Walk through each inbox item: refile, tag, or discard |
| `brain-plan` | *"Let's plan..."* | Break goals into tasks with priorities and deadlines |
| `brain-focus` | *"Let's work on..."* | Deep work session on one project |
| `brain-review` | *"Weekly review"* | Progress check across all projects |
| `brain-setup` | *"Set up my brain"* | Guided first-time onboarding |

```bash
brain skill install          # install all skills
brain skill status           # check what's installed
brain skill upgrade          # update after a brain version upgrade
```

## Next Steps

- **[Installation Guide](installation.md)** -- Complete installation for all platforms
- **[Command Reference](commands.md)** -- Full CLI documentation
- **[MCP Server](mcp-server.md)** -- AI agent integration details
- **[Development Guide](development.md)** -- Architecture, contributing, and API docs
