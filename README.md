# Local Brain

```
 ______     ______     ______     __     __   __
/\  == \   /\  == \   /\  __ \   /\ \   /\ "-.\ \
\ \  __<   \ \  __<   \ \  __ \  \ \ \  \ \ \-.  \
 \ \_____\  \ \_\ \_\  \ \_\ \_\  \ \_\  \ \_\\"\_\
  \/_____/   \/_/ /_/   \/_/\/_/   \/_/   \/_/ \/_/
```

> Plain markdown files. Full project management. Managed by AI, or by you.

[![Documentation](https://img.shields.io/badge/docs-sandermoon.github.io-blue)](https://sandermoon.github.io/local-brain/)
[![Go Version](https://img.shields.io/github/go-mod/go-version/SanderMoon/local-brain)](https://github.com/SanderMoon/local-brain)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Local Brain turns a folder of markdown files into a project management system: a CLI, a TUI, and an AI agent interface. Your projects, tasks, and notes stay in plain text on your machine. You can manage them yourself, hand them off to an AI agent, or do both.

Open the same files in Obsidian, grep them from the terminal, or let Claude organize your week. It's just markdown.

---

## What it looks like on disk

```
~/brains/work/                          ← a "brain" (workspace)
├── 00_dump.md                          ← inbox for quick captures
├── website-redesign/                   ← a project
│   ├── todo.md
│   ├── notes.md
│   └── .repos                          ← linked git repositories
├── api-migration/
│   ├── todo.md
│   └── notes.md
└── ...
```

A **brain** is a folder. A **project** is a subfolder. Tasks and notes are markdown files. That's it.

Here's what a `todo.md` looks like:

```markdown
# Website Redesign

## Active

- [ ] Finalize color palette #p:1 due:2026-04-01 #design
- [ ] Set up CI/CD pipeline #p:2 #infra
- [ ] Write copy for landing page

## Completed

- [x] Create wireframes #p:1 #design #captured:2026-03-01 #done:2026-03-18
```

And the inbox (`00_dump.md`), where quick captures land before you organize them:

```markdown
# Dump

- [ ] Fix authentication bug #captured:2026-03-25
- [ ] Review PR #452 #captured:2026-03-25

[Note] Architecture meeting #captured:2026-03-25
    Decided to move to microservices
    Redis for session caching
    Follow up with platform team
```

These files work in any text editor, any markdown tool, and any git workflow. Local Brain provides the API layer on top.

---

## Three ways to use it

### 1. AI agent (the main event)

![Claude chat with Local Brain](docs/images/claude-chat-screenshot.png)

Connect your AI agent via the built-in MCP server and agent skills. Then just talk to it:

- *"Good morning"* : get a daily briefing with priorities, overdue tasks, and blockers
- *"Process my inbox"* : guided triage of captured items
- *"Let's plan the website redesign"* : break a project into tasks with priorities and deadlines
- *"What should I focus on today?"* : context-aware recommendations

The agent reads and writes the same markdown files. No sync, no database, no lock-in.

### 2. CLI & TUI

![Local Brain TUI](docs/images/tui-screenshot.png)

```bash
brain add "Fix auth bug"            # capture to inbox (< 1 second)
brain refile                        # move inbox items to projects
brain todo ls --priority 1          # see high-priority tasks
brain todo done abc123              # mark a task done
brain tui                           # visual project overview
```

### 3. Your editor

Open `~/brains/work/` as an Obsidian vault, VS Code workspace, or just use vim. Edit the markdown directly; Local Brain picks up changes instantly.

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
brain skill install          # install agent skills (briefing, triage, planning, ...)
```

Auto-detects installed agents (Claude Code, Codex, Gemini CLI, OpenCode). Target a specific one with `--agent claude`.

### Start using it

```bash
# Capture something
brain add "Ship the new API endpoint"

# Or just open your agent and say "what's on my plate?"
```

---

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

Together these form a **Capture, Organize, Plan, Execute, Reflect** loop, inspired by GTD and built for AI-assisted workflows.

```bash
brain skill install          # install all skills
brain skill status           # check what's installed
brain skill upgrade          # update after a brain version upgrade
```

---

## Why Local Brain?

**Your files are the database.** No proprietary format, no cloud sync, no vendor lock-in. Every piece of data is a markdown file you can read, edit, grep, and version control.

**AI-native, not AI-only.** The MCP server gives agents full access to your projects, but you're never locked out. Edit a `todo.md` by hand, and the agent sees the change. Let the agent reorganize your tasks, and you see the result in your editor.

**Works with your tools.** Obsidian, VS Code, vim, Syncthing, git... anything that reads files works with Local Brain. The CLI and TUI are there for when you want structured operations.

---

## Documentation

- **[Quickstart Guide](https://sandermoon.github.io/local-brain/)** · Get started in 3 minutes
- **[Installation](https://sandermoon.github.io/local-brain/installation/)** · All installation methods and shell integration
- **[Command Reference](https://sandermoon.github.io/local-brain/commands/)** · Full CLI documentation
- **[MCP Server](docs/mcp-server.md)** · AI agent integration details
- **[Development Guide](https://sandermoon.github.io/local-brain/development/)** · Architecture and contributing

---

## Contributing

```bash
git clone https://github.com/SanderMoon/local-brain.git
cd local-brain
make build && make test
```

See the [Development Guide](https://sandermoon.github.io/local-brain/development/) for architecture details.

---

## License

MIT. See [LICENSE](LICENSE) for details.
