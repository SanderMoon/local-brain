# Agent Skills

Local Brain ships 7 agent skills: plain SKILL.md files that give AI coding agents step-by-step project management workflows. They follow the [Agent Skills open standard](https://agentskills.io), get installed into your agent's config directory, and load automatically at session start.

**MCP tools** are single operations (fetch tasks, create projects). **Skills** are workflows that chain multiple tool calls together with decision logic and formatting. You need both: the MCP server provides the tools, skills provide the playbook.

---

## The 7 Bundled Skills

Together, these skills form a **Capture, Organize, Plan, Execute, Reflect** loop.

### brain-daily

**Purpose:** Morning briefing and daily planning.

**When to use it:** At the start of your workday, or whenever you want a status overview.

**Example triggers:**
- "Good morning"
- "What should I focus on today?"
- "What's on my plate?"
- "Catch me up"

**What it does:**
1. Fetches the daily briefing (overdue, due-today, in-progress, blocked items)
2. Shows a project overview table with open vs. backlog counts
3. Lists actionable items grouped by urgency
4. Recommends a single focus task
5. Offers follow-up actions (triage inbox, promote backlog, set context)

Backlog items are excluded from the actionable list and summarized separately.

---

### brain-capture

**Purpose:** Quick capture of tasks, ideas, and notes with zero friction.

**When to use it:** Any time you want to save a thought without breaking your flow.

**Example triggers:**
- "Remind me to update the API docs"
- "Add a task: fix the auth bug"
- "Note to self: check Redis config"
- "I need to email Sarah about the contract"

**What it does:**
1. Parses natural language for action items, due dates, urgency, and project associations
2. Routes to the right place: known project gets a direct task, ambiguous goes to inbox, non-actionable becomes a note
3. Handles multiple items in one message
4. Confirms with a single line. No follow-up questions.

Metadata inference: "by Friday" becomes a due date, "urgent" becomes priority 1, "eventually" becomes priority 3.

---

### brain-triage

**Purpose:** Process inbox items systematically.

**When to use it:** When your inbox has accumulated items that need to be sorted into projects.

**Example triggers:**
- "Process my inbox"
- "Triage my dump"
- "Clean up my inbox"
- After a daily briefing shows pending inbox items

**What it does:**
1. Fetches all inbox items and existing projects
2. Groups related items and suggests actions: refile, create project, convert to note, or discard
3. Waits for confirmation before executing
4. Supports batch operations ("refile all of these to backend-v2")
5. Presents a summary when done

For large inboxes (15+ items), offers a quick pass for obvious refiles vs. a thorough pass for every item.

---

### brain-plan

**Purpose:** Break goals into concrete, actionable tasks with priorities and deadlines.

**When to use it:** When starting a new project or when a vague goal needs structure.

**Example triggers:**
- "Let's plan the website redesign"
- "Break this down into tasks"
- "How should I approach this?"
- "New project: mobile app v2"

**What it does:**
1. Determines scope: new project, existing project needing tasks, or vague goal
2. Brainstorms tasks, breaks them into day-sized pieces, identifies dependencies
3. Adds metadata: priorities, deadlines working back from milestones, logical ordering
4. Creates the project (if new), all tasks, and planning notes capturing goals and decisions
5. New tasks default to backlog. Only tasks committed to immediately get set to open.

Planning notes preserve context that would be lost when the session ends.

---

### brain-focus

**Purpose:** Deep work session on a single project.

**When to use it:** When you are ready to work on one thing without distractions.

**Example triggers:**
- "Let's work on the API migration"
- "Focus on backend-v2"
- "Deep work time"

**What it does:**
1. Sets project context and loads the full dashboard (tasks, notes, repos)
2. Suggests the next action: continue in-progress work or pick the highest-priority open task
3. Marks the chosen task as in-progress
4. During the session: captures stray thoughts to inbox, tracks completions, suggests next tasks
5. On wrap-up: summarizes accomplishments, offers to save a session note

One project, one context. Anything off-topic goes to inbox.

---

### brain-review

**Purpose:** Weekly review of all projects and progress.

**When to use it:** At the end of the week, or whenever you want a full picture of your workspace.

**Example triggers:**
- "Weekly review"
- "How did this week go?"
- "What's the state of everything?"
- "Friday review"

**What it does:**
1. Gathers all projects, recent completions, open tasks, blocked items, inbox state
2. Celebrates progress first (completions grouped by project)
3. Reviews each active project: flags stale tasks (2+ weeks), surfaces blocked work, checks overdue items
4. Offers options for stale work: keep, reprioritize, drop, or break down
5. Processes inbox if items are pending
6. Asks about the coming week, generates a review note

Takes 5-15 minutes. Respects "skip" if a project is on track.

---

### brain-setup

**Purpose:** Guided first-time setup of a Local Brain workspace.

**When to use it:** Right after installing Local Brain, or when the brain is empty.

**Example triggers:**
- "Set up my brain"
- "I'm new here"
- "Help me organize my projects"
- First interaction with an empty brain

**What it does:**
1. Checks current state (empty brain, existing projects, or inbox-only)
2. Asks what you are working on conversationally
3. Creates 3-5 projects with initial tasks from your description
4. Runs a mini daily briefing to show the system working
5. Suggests saying "good morning" tomorrow

Keeps it to 3-5 exchanges. Infers structure from natural language.

---

## Supported Agents

Skills are installed as SKILL.md files in each agent's configuration directory. Local Brain auto-detects which agents you have installed.

| Agent | ID | Config Directory | Skills Directory |
|-------|----|-----------------|-----------------|
| Claude Code | `claude` | `~/.claude` | `~/.claude/skills/` |
| Codex | `codex` | `~/.codex` | `~/.codex/skills/` |
| Gemini CLI | `gemini` | `~/.gemini` | `~/.gemini/skills/` |
| OpenCode | `opencode` | `~/.config/opencode` | `~/.config/opencode/skills/` |

Detection checks whether the config directory exists. Undetected agents are skipped unless you target them with `--agent`.

Each skill is installed as `<skills-dir>/<skill-name>/SKILL.md`. For example, the daily briefing skill for Claude Code lives at `~/.claude/skills/brain-daily/SKILL.md`.

---

## Management Commands

### `brain skill list`

List all bundled skills with their descriptions.

```bash
brain skill list
```

### `brain skill install [name]`

Install skills to all detected AI agents.

```bash
brain skill install              # install all 7 skills
brain skill install brain-daily  # install only the daily briefing skill
```

**Options:**
- `--agent <id>` - Target a specific agent instead of all detected ones. Valid values: `claude`, `codex`, `gemini`, `opencode`, `all` (default: `all`)
- `--force` - Overwrite existing skill files

If no agents are detected, the command prints a list of supported agents with their expected config directories.

Already-installed skills are skipped unless `--force` is used.

### `brain skill status`

Show the installation status of all skills across all known agents.

```bash
brain skill status
```

Prints a table with each skill as a row and each agent as a column. Values are `installed`, `not installed`, or `-` (agent not detected).

### `brain skill upgrade`

Upgrade all installed skills to the latest version bundled with your `brain` binary.

```bash
brain skill upgrade
brain skill upgrade --agent claude
```

**Options:**
- `--agent <id>` - Target a specific agent (default: `all`)

This overwrites installed SKILL.md files with the versions from the current binary. Skills that are not installed are skipped (use `brain skill install` to add new skills first).

Run this after updating Local Brain to pick up any skill improvements.

### `brain skill remove <name>`

Remove a skill from all detected agents.

```bash
brain skill remove brain-setup
brain skill remove brain-daily --agent claude
```

**Options:**
- `--agent <id>` - Target a specific agent (default: `all`)

---

## Customizing Skills

Installed SKILL.md files are plain markdown. Edit them directly:

```bash
vim ~/.claude/skills/brain-daily/SKILL.md
```

You can change trigger phrases, adjust workflow steps, modify tone, or add project-specific instructions.

**Note:** `brain skill upgrade` overwrites installed files. Back up customized skills before upgrading, or skip upgrading the ones you have modified.

---

## Creating Your Own Skills

Add a SKILL.md file to your agent's skills directory.

### 1. Create the skill directory

```bash
mkdir -p ~/.claude/skills/my-custom-skill
```

### 2. Write the SKILL.md

Create `~/.claude/skills/my-custom-skill/SKILL.md` with this structure:

```markdown
---
name: my-custom-skill
description: One-line description of what this skill does and when to use it.
compatibility: Requires local-brain MCP server (brain-mcp) to be configured.
---

# my-custom-skill: Human-Readable Title

Brief explanation of what this skill does.

## Trigger Phrases

- "phrase one", "phrase two"
- Describe when this skill should activate

## Workflow

### Step 1 - Do the first thing

Explain what the agent should do, which MCP tools to call, and what to present to the user.

### Step 2 - Do the next thing

Continue with the workflow steps.

## Tool Sequence

Summarize the MCP tool calls in order:

    get_brain_overview()
      -> some_tool()
      -> another_tool()

## Notes

- Guidelines for the agent's behavior
- Edge cases to handle
- Tone guidance
```

### 3. Tips

- Name the exact MCP tools the agent should call and in what order.
- One skill, one workflow. Multiple workflows = multiple skills.
- Test by starting a fresh agent session and trying the trigger phrases.

### 4. Multi-agent installation

To use a custom skill with multiple agents, copy the SKILL.md file to each agent's skills directory:

```bash
cp -r ~/.claude/skills/my-custom-skill ~/.gemini/skills/
cp -r ~/.claude/skills/my-custom-skill ~/.codex/skills/
```

Custom skills are not managed by `brain skill install` or `brain skill upgrade`. You manage them yourself.

---

## The Agent Skills Standard

Local Brain's skills follow the [Agent Skills open standard](https://agentskills.io): YAML frontmatter (`name`, `description`, `compatibility`) plus a markdown body with workflow instructions. Any agent that supports the standard can load them.
