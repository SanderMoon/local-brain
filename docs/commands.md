# Command Reference

## Brain Management

### `brain new [name]`

Create a new brain workspace.

```bash
brain new work
brain new personal
```

Creates `~/brains/<name>` with: `00_dump.md`, `01_active/`, `02_areas/`, `03_resources/`, `99_archive/`, `_templates/`.

**Notes:**
- Defaults to `default` if no name given
- Auto-switches to the new brain if it's the first one or named "default"
- Creates a symlink at `~/brain` pointing to the new brain

---

### `brain switch [name]`

Switch to a different brain workspace.

```bash
brain switch work
brain switch          # interactive fzf selection
```

Updates the `~/brain` symlink. Previous brain remains intact.

---

### `brain current`

Show the currently active brain.

```bash
brain current
brain current --name-only
brain current --path-only
```

**Options:**
- `--name-only` - Output only the brain name (useful for scripts/prompts)
- `--path-only` - Output only the path

---

### `brain list`

List all registered brains.

```bash
brain list
brain list --paths
brain list --json
```

**Options:**
- `--paths` - Show full filesystem paths
- `--json` - Output JSON format

Active brain is marked with `*`.

---

### `brain path [brain-name]`

Show the filesystem path of a brain. Defaults to the active brain.

```bash
brain path
brain path work
cd $(brain path)      # navigate to brain directory
```

---

### `brain rename <old> <new>`

Rename a brain. Updates the directory, config, and symlink if active.

```bash
brain rename work work-2024
```

---

### `brain delete <name>`

Permanently delete a brain. **Cannot be undone.**

```bash
brain delete old-brain
```

Requires `y/N` confirmation. Removes the `~/brain` symlink if it pointed to the deleted brain.

---

### `brain import [root-path]`

Scan a directory for brain structures and register them.

```bash
brain import                        # Scan ~/brains (default)
brain import ~/Dropbox/Brains       # Scan a custom directory
```

**Notes:**
- Looks for subdirectories containing `01_active/` and `00_dump.md`
- Registers paths only (does not copy files)
- Skips already-registered brains

---

### `brain config <key> [value]`

Get or set configuration values. Stored in `~/.config/brain/config.json`.

```bash
brain config editor           # Get current editor
brain config editor nvim      # Set editor
```

**Available Keys:**
- `editor` - Preferred text editor

---

## Capture Workflow

### `brain add [text]`

Quick capture to dump.

```bash
brain add "Fix authentication bug"     # Task capture (instant)
brain add                              # Note capture (opens editor)
```

**With text:** Appends as `- [ ] your text #captured:YYYY-MM-DD`.

**Without text:** Prompts for note title, opens `$EDITOR`, saves as indented note block.

---

### `brain dump ls [--json]`

List items in the dump with stable IDs.

```bash
brain dump ls
brain dump ls --json
```

**Options:**
- `--json` - Output JSON with IDs, types, and line numbers

IDs are stable (MD5 hash of content + line number + file mtime).

---

## Curation Workflow

### `brain refile [<id> <project>]`

Move items from dump to projects.

```bash
brain refile                    # Interactive: walk through each item
brain refile a1b2c3 backend-api # Direct: refile by ID
```

**Behavior:**
- **Tasks** append to project's `todo.md` with backlog status (`[~]`). Use `brain plan` to promote.
- **Notes** are saved as `YYYY-MM-DD-title-slug.md` in the project's `notes/` directory.
- Items retain their `#captured:` timestamp and are removed from dump after refiling.

Interactive mode shows fzf project selection with `[SKIP]` and `[TRASH]` options.

---

### `brain plan`

Promote backlog tasks and set metadata. Ideal for weekly planning sessions.

```bash
brain plan
```

**Workflow:**
1. Select tasks via fzf (Tab for multi-select)
2. For each, prompts for status, priority (1-3), due date, and tags
3. All fields optional; press Enter to skip or accept defaults
4. Backlog items auto-promote to "open" on Enter

**Natural language dates:** `today`, `tomorrow`, `+3d`, `next-friday`, `YYYY-MM-DD`. Enter `clear` to remove metadata.

---

### `brain review`

AI-powered project analysis. **Not yet implemented** (stub).

---

## Project Management

### `brain project new <name>`

Create a new project with boilerplate (`notes.md`, `todo.md`, `.repos`).

```bash
brain project new website-redesign
```

**Options:**
- `--section <section>` - PARA section (default: `01_active`; valid: `01_active`, `02_areas`, `03_resources`)

Names: letters, numbers, hyphens, and underscores only. Auto-focuses the new project.

---

### `brain project list [--json]`

**Alias:** `brain project ls`

List all projects. Selected project marked with `*`.

```bash
brain project list
brain project list --json
```

**Options:**
- `--json` - Output JSON format
- `--section <section>` - PARA section (default: `01_active`; valid: `01_active`, `02_areas`, `03_resources`)

---

### `brain project select [name]`

Focus on a specific project. Focus persists across sessions.

```bash
brain project select backend-api
brain project select              # Interactive fzf selection
```

---

### `brain project current`

Show the currently focused project. Exit code 1 if none focused.

```bash
brain project current
```

---

### `brain project describe`

Edit or display a project description (stored in `description.md`).

```bash
brain project describe              # Edit in editor
brain project describe --show       # Display
brain project describe --show --json
```

**Options:**
- `--show` - Display instead of editing
- `--json` - JSON output (with `--show`)

---

### `brain project clone <url> [name]`

Import a git repository as a new project. Creates project, links repo, clones to `~/dev/`, and focuses.

```bash
brain project clone https://github.com/user/repo.git
brain project clone https://github.com/user/repo.git my-project
```

Project name is extracted from the URL if not provided.

---

### `brain project link <url>`

Link a git repository to the focused project's `.repos` file.

```bash
brain project link https://github.com/user/repo.git
```

Multiple repos can be linked to one project. Verifies remote accessibility.

---

### `brain project pull`

Clone or update all linked repositories for the focused project.

```bash
brain project pull
```

Not-yet-cloned repos are cloned to `~/dev/<repo-name>/`. Already-cloned repos are pulled. Lines starting with `#` in `.repos` are skipped.

---

### `brain project archive <name>`

Move a project from `01_active/` to `99_archive/` (appends `_YYYYMMDD`).

```bash
brain project archive old-project
brain project archive              # Archives focused project
```

Does not delete anything. Code repositories in `~/dev/` are not affected.

---

### `brain project move <project> [target-brain]`

Move a project to another brain.

```bash
brain project move backend-api work
brain project move backend-api     # Interactive brain selection
```

Fails if a project with the same name exists in the target brain.

---

### `brain project delete <name>`

Permanently delete a project. **Cannot be undone.** Consider `brain project archive` instead.

```bash
brain project delete old-project
brain project delete               # Deletes focused project
```

Requires typing the project name to confirm. Does not delete repos in `~/dev/`.

---

## Todo Management

### `brain todo`

Interactive fuzzy search through tasks. Select a task to open it in your editor at the exact line.

```bash
brain todo
```

Requires fzf. Preview uses `bat` if available.

---

### `brain todo ls`

List tasks with filters.

```bash
brain todo ls
brain todo ls --status in-progress --priority 1
brain todo ls --tag bug --tag security --tag-mode and
brain todo ls --overdue --sort deadline
brain todo ls --all --json
```

**Options:**
- `--json` - Output JSON format
- `--all` - Include completed tasks
- `--priority <1-3>` - Filter by priority level
- `--no-priority` - Show only unprioritized tasks
- `--status <state>` - Filter: backlog, open, in-progress, blocked, done
- `--tag <tag>` - Filter by tag (repeatable)
- `--tag-mode <and|or>` - Tag filter mode (default: or)
- `--due-today` - Tasks due today
- `--due-this-week` - Tasks due within 7 days
- `--overdue` - Tasks past due date
- `--sort <field>` - Sort by: priority, deadline, project, status
- `--completed-after <YYYY-MM-DD>` - Completed on/after date (implies `--all`)
- `--completed-before <YYYY-MM-DD>` - Completed on/before date (implies `--all`)

**Output format:** `ID [P1] [>] Task content #tags (project) [Due: DATE]`

Default sort: overdue/upcoming first, then by priority. Filters can be combined.

---

### `brain todo done [ID...]`

Mark tasks as complete (`[x]`). Tasks stay in todo.md and can be reopened.

```bash
brain todo done abc123
brain todo done abc123 def456     # Multiple
brain todo done                   # Interactive multi-select
```

---

### `brain todo delete [id]`

Permanently delete a task. **Cannot be undone.** Requires confirmation.

```bash
brain todo delete abc123
brain todo delete                 # Interactive selection
```

---

### `brain todo reopen [id]`

Reopen a completed task (changes `[x]` back to `[ ]`).

```bash
brain todo reopen abc123
brain todo reopen                 # Interactive selection from completed tasks
```

---

### `brain todo start <id>`

Mark task as in-progress (`[>]`).

```bash
brain todo start abc123
```

---

### `brain todo block <id>`

Mark task as blocked (`[-]`).

```bash
brain todo block abc123
```

---

### `brain todo unblock <id>`

Unblock a task (changes `[-]` back to `[ ]`).

```bash
brain todo unblock abc123
```

---

### `brain todo status <id> <state>`

Set task status directly. Valid states: `backlog` (`[~]`), `open` (`[ ]`), `in-progress` (`[>]`), `blocked` (`[-]`), `done` (`[x]`).

```bash
brain todo status abc123 in-progress
```

---

### `brain todo prio <id> <priority>`

Set task priority. Adds/updates `#p:N` tag.

```bash
brain todo prio abc123 1          # High (P1)
brain todo prio abc123 2          # Medium (P2)
brain todo prio abc123 3          # Low (P3)
brain todo prio abc123 0          # Clear priority
```

---

### `brain todo due <id> <date>`

Set task due date. Adds/updates `#due:YYYY-MM-DD` tag. Overdue tasks show `[OVERDUE]` in listings.

```bash
brain todo due abc123 2026-02-15
brain todo due abc123 tomorrow
brain todo due abc123 +7d
brain todo due abc123 next-friday
brain todo due abc123 clear       # Remove due date
```

---

### `brain todo schedule`

Batch schedule unscheduled tasks interactively. Loops through open tasks without due dates, prompting for each.

```bash
brain todo schedule
```

Press Esc to exit. Supports all date formats from `brain todo due`.

---

### `brain todo tag <id> <tags...>`

Add or remove tags on a task. Tags are stored as `#tagname` inline.

```bash
brain todo tag abc123 bug security
brain todo tag abc123 --rm security
```

**Options:**
- `--rm` - Remove specified tags instead of adding

---

### `brain todo tags`

List all tags in use with counts.

```bash
brain todo tags
```

---

## Note Management

### `brain note`

Select and open a note in the focused project via fzf.

```bash
brain note
```

**Options:**
- `--section <section>` - PARA section (default: `01_active`; valid: `01_active`, `02_areas`, `03_resources`)

Requires a focused project. To create notes: `brain add` then `brain refile`.

---

### `brain note ls`

List note files in the focused project.

```bash
brain note ls
brain note ls --json
```

**Options:**
- `--json` - Output JSON format
- `--section <section>` - PARA section (default: `01_active`; valid: `01_active`, `02_areas`, `03_resources`)

---

### `brain note delete`

Select and delete a note from the focused project. **Cannot be undone.**

```bash
brain note delete
```

**Options:**
- `--section <section>` - PARA section (default: `01_active`; valid: `01_active`, `02_areas`, `03_resources`)

Shows fzf selection with preview. Requires `y/N` confirmation.

---

## Context & Dev Mode

### `brain go`

Enter project context.

```bash
brain go
```

**No linked repos:** Opens a new shell in the project directory.

**With linked repos:** Launches a tmux session with windows for the project directory and each linked repository. Activates Python venv if present. Attaches to existing session if already running.

---

### `brain storm [--agent <name>]`

Open an AI agent in your brains root directory (`~/brains` or `$BRAIN_ROOT`).

```bash
brain storm                    # Launch Claude Code
brain storm --agent aider      # Launch aider
```

**Options:**
- `--agent`, `-a` (string, default: `claude`) - CLI agent binary to launch

The agent process replaces the current shell (exec). Requires the agent binary on `$PATH`.

---

## Sync & Utilities

### `brain sync [status|scan|rescan]`

Syncthing control wrapper. Requires Syncthing installed and running.

```bash
brain sync              # Force rescan (default)
brain sync status       # Show Syncthing status
```

---

### `brain daily`

Create or open today's daily note at `{brain}/00_daily/YYYY-MM-DD.md`.

```bash
brain daily
brain daily --open      # Open in editor after creating
```

Includes a briefing section with overdue todos from all projects.

---

## Migration

### `brain migrate`

Migrate brain files to portable markdown standards. Dry-run by default.

```bash
brain migrate                          # Dry-run
brain migrate --apply                  # Write changes
brain migrate --project backend-api    # Single project
brain migrate --section 02_areas       # Different PARA section
```

**Options:**
- `--apply` - Write changes (default is dry-run)
- `--project <name>` - Migrate only the named project
- `--section <section>` - PARA section (default: `01_active`; valid: `01_active`, `02_areas`, `03_resources`)

**Operations:**
1. **Notes frontmatter** - Adds YAML frontmatter to legacy note files
2. **Todos HTML comments** - Moves inline `#id:`, `#captured:`, `#done:` tags into HTML comments
3. **Notes index links** - Appends missing relative links to `notes.md`

Idempotent and safe to run multiple times. Uses atomic writes.

---

### `brain migrate-ids`

Add stable IDs (`#id:` tags) to all todos that don't have them.

```bash
brain migrate-ids
```

Safe to run multiple times. IDs use the same MD5 algorithm as the original bash version.

---

## MCP Server Management

Register the brain-mcp server with AI coding agents. Auto-detects installed agents.

### `brain mcp install [--agent <id>]`

Register brain-mcp with detected AI agents.

```bash
brain mcp install
brain mcp install --agent claude
```

**Options:**
- `--agent` - Target: `claude`, `codex`, `gemini`, `opencode`, or `all` (default: `all`)

**Supported agents:**

| ID | Name | Config written to |
|----|------|------------------|
| `claude` | Claude Code | `~/.claude.json` |
| `codex` | Codex | `~/.codex/config.toml` |
| `gemini` | Gemini CLI | `~/.gemini/settings.json` |
| `opencode` | OpenCode | `~/.config/opencode/opencode.json` |

Re-running is safe; already-registered agents are skipped. Binary path is auto-detected.

---

### `brain mcp remove [--agent <id>]`

Remove brain-mcp registration from AI agents.

```bash
brain mcp remove
brain mcp remove --agent claude
```

**Options:**
- `--agent` - Target: `claude`, `codex`, `gemini`, `opencode`, or `all` (default: `all`)

---

### `brain mcp status`

Show MCP registration status across all known agents.

```bash
brain mcp status
```

Status values: `registered`, `not registered`, `-` (agent not detected).

---

## Skill Management

Install SKILL.md files into AI coding agents to teach them Local Brain workflows.

### `brain skill list`

List all bundled skills.

```bash
brain skill list
```

---

### `brain skill install [name] [--agent <id>] [--force]`

Install skill(s) to detected AI agents.

```bash
brain skill install                              # All skills, all agents
brain skill install --agent claude               # All skills, one agent
brain skill install brain-daily --agent claude   # One skill, one agent
brain skill install --force                      # Overwrite existing
```

**Options:**
- `--agent` - Target: `claude`, `codex`, `gemini`, `opencode`, or `all` (default: `all`)
- `--force` - Overwrite existing skill files

**Skills directories:**

| ID | Skills directory |
|----|-----------------|
| `claude` | `~/.claude/skills/` |
| `codex` | `~/.codex/skills/` |
| `gemini` | `~/.gemini/skills/` |
| `opencode` | `~/.config/opencode/skills/` |

Re-running without `--force` is safe; existing skills are skipped. OpenCode also reads `~/.claude/skills/` natively.

---

### `brain skill upgrade [--agent <id>]`

Upgrade installed skills to the latest bundled version. Only upgrades already-installed skills. **Overwrites local modifications.**

```bash
brain skill upgrade
brain skill upgrade --agent claude
```

**Options:**
- `--agent` - Target: `claude`, `codex`, `gemini`, `opencode`, or `all` (default: `all`)

---

### `brain skill remove <name> [--agent <id>]`

Remove a skill from AI agents.

```bash
brain skill remove brain-daily
brain skill remove brain-daily --agent claude
```

**Options:**
- `--agent` - Target: `claude`, `codex`, `gemini`, `opencode`, or `all` (default: `all`)

---

### `brain skill status`

Show installation status of all bundled skills across all known agents.

```bash
brain skill status
```

Status values: `installed`, `not installed`, `-` (agent not detected).

---

## Global Options

- `--help` - Show command help
- `--version` - Show version information

---

## Task States Reference

Local Brain uses markdown checkboxes with different symbols:

| Checkbox | State | Description |
|----------|-------|-------------|
| `[~]` | backlog | Not yet planned for active work |
| `[ ]` | open | Planned for current work cycle |
| `[>]` | in-progress | Currently working |
| `[-]` | blocked | Waiting on external dependency |
| `[x]` | done | Completed |

---

## Metadata Tags Reference

Tasks support inline metadata tags:

| Tag | Format | Example | Description |
|-----|--------|---------|-------------|
| Priority | `#p:N` | `#p:1` | Priority 1-3 (1=high) |
| Due Date | `#due:DATE` | `#due:2026-02-15` | Task deadline |
| Captured | `#captured:DATE` | `#captured:2026-01-29` | When item was added |
| Done | `#done:DATE` | `#done:2026-01-30` | When completed |
| Custom Tags | `#tagname` | `#bug #security` | Free-form labels |

**Example Task:**
```markdown
- [>] Fix authentication bug #p:1 #due:2026-02-15 #bug #security #captured:2026-01-29
```

---

## JSON API Usage

Many commands support `--json` for scripting:

```bash
brain dump ls --json | jq '.[0].id'
brain project list --json | jq '.[] | select(.focused==true) | .name'
brain todo ls --priority 1 --json | jq '.[].content'
```

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `BRAIN_ROOT` | `~/brains` | Root directory for all brains |
| `BRAIN_SYMLINK` | `~/brain` | Active brain symlink location |
| `BRAIN_CONFIG_DIR` | `~/.config/brain` | Config directory |
| `EDITOR` | `vi` | Editor for notes |

---

## Next Steps

- [Installation Guide](installation.md)
- [Quickstart](index.md)
- [Development Guide](development.md)
