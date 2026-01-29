# Command Reference

Complete documentation of all Local Brain commands.

---

## Brain Management

### `brain new [name]`

**Description:** Create a new brain workspace

**Usage:**
```bash
brain new work
brain new personal
```

Creates a new brain directory in `~/brains/<name>` with the following structure:
- `00_dump.md` - Inbox for quick captures
- `01_active/` - Active projects directory
- `99_archive/` - Archived projects directory

**Notes:**
- If no name is provided, prompts interactively
- Automatically switches to the new brain
- Creates a symlink at `~/brain` pointing to the new brain

---

### `brain switch [name]`

**Description:** Switch to a different brain workspace

**Usage:**
```bash
brain switch work
brain switch personal
```

**Options:**
- If no name provided, shows interactive selection with fzf

**Notes:**
- Updates the `~/brain` symlink to point to the selected brain
- Previous brain remains intact and can be switched back to anytime

---

### `brain current`

**Description:** Show the currently active brain name

**Usage:**
```bash
brain current
```

**Output:**
```
work
```

**Notes:**
- Returns exit code 1 if no brain is active
- Useful for shell prompts and scripts

---

### `brain list`

**Description:** List all registered brains

**Usage:**
```bash
brain list
```

**Output:**
```
Available Brains:
-----------------
 * work           (active)
   personal
   research
```

The active brain is marked with `*`.

---

### `brain path`

**Description:** Show the filesystem path of the active brain

**Usage:**
```bash
brain path
```

**Output:**
```
/Users/username/brains/work
```

**Notes:**
- Useful for scripting and integrations
- Can be used with `cd $(brain path)` to navigate to brain directory

---

### `brain rename <old> <new>`

**Description:** Rename a brain

**Usage:**
```bash
brain rename work work-2024
```

**Notes:**
- Renames both the directory and updates the configuration
- If renaming the active brain, the symlink is updated automatically

---

### `brain delete [name]`

**Description:** Permanently delete a brain

**Usage:**
```bash
brain delete old-brain
```

**Options:**
- If no name provided, shows interactive selection

**Notes:**
- **DESTRUCTIVE OPERATION** - Cannot be undone
- Requires typing the brain name to confirm deletion
- Cannot delete the currently active brain (switch first)
- All projects, notes, and tasks in the brain are permanently deleted

---

### `brain import [path]`

**Description:** Import an existing brain directory

**Usage:**
```bash
brain import ~/Dropbox/my-old-brain
brain import /path/to/synced-brain
```

**Notes:**
- Registers an existing brain directory with Local Brain
- Useful for importing synced brains from cloud storage
- Does not copy files, only registers the path

---

## Capture Workflow

### `brain add [text]`

**Description:** Quick capture to dump

**Usage:**
```bash
# Quick task capture
brain add "Fix authentication bug"
brain add "Email Sarah about proposal"

# Note capture (opens editor)
brain add
```

**Modes:**

**Quick Capture** (with text):
- Appends as a checkbox task: `- [ ] your text #captured:YYYY-MM-DD`
- Takes < 1 second
- No metadata required

**Editor Mode** (no text):
- Prompts for note title
- Opens your editor (vim/nvim/nano)
- Saves as indented note block with title

**Examples:**
```bash
# Task
brain add "Review PR #123"
# Result in dump: - [ ] Review PR #123 #captured:2026-01-29

# Note (opens editor)
brain add
# Prompts: Note title: Meeting Notes
# Opens editor for content
```

**Notes:**
- Uses `$EDITOR` environment variable
- Captured tasks can be enriched later with `brain plan`
- Notes are multi-line and support full markdown

---

### `brain dump ls [--json]`

**Description:** List items in the dump with stable IDs

**Usage:**
```bash
# Human-readable
brain dump ls

# JSON output for scripts
brain dump ls --json
```

**Options:**
- `--json` - Output JSON format with IDs for programmatic access

**Output (human):**
```
a1b2c3  [Task] Fix authentication bug #captured:2026-01-29
d4e5f6  [Note] Meeting notes from standup #captured:2026-01-29
```

**Output (JSON):**
```json
[
  {
    "id": "a1b2c3",
    "type": "task",
    "content": "Fix authentication bug #captured:2026-01-29",
    "line": 5
  }
]
```

**Notes:**
- IDs are stable and based on MD5 hash of content + line number + file mtime
- JSON output includes line numbers for precise file editing

---

### `brain dump rm <id>`

**Description:** Remove an item from the dump by ID

**Usage:**
```bash
brain dump rm a1b2c3
```

**Notes:**
- Permanently deletes the item from dump
- Use `brain refile` instead to move items to projects
- Useful for removing spam or accidental captures

---

## Curation Workflow

### `brain refile [<id> <project>]`

**Description:** Move items from dump to projects

**Usage:**

**Interactive Mode** (recommended):
```bash
brain refile
```

Shows each dump item one by one with:
- Interactive project selection (fzf)
- Special options: `[SKIP]` and `[TRASH]`
- Progress counter

**Direct Mode:**
```bash
brain refile a1b2c3 backend-api
```

Moves specific item by ID to a specific project.

**Behavior:**
- **Tasks** → Appended to project's `todo.md`
- **Notes** → Created as separate markdown files in project's `notes/` directory

**Examples:**
```bash
# Interactive refiling
brain refile

# Direct refiling
brain refile a1b2c3 work
brain refile d4e5f6 personal
```

**Notes:**
- Items retain their `#captured:YYYY-MM-DD` timestamp
- Notes are saved as `YYYY-MM-DD-title-slug.md`
- Removed from dump after successful refiling

---

### `brain plan`

**Description:** Interactive batch task planning

**Usage:**
```bash
brain plan
```

**Workflow:**
1. Shows interactive task selection (fzf)
2. For each selected task, prompts for:
   - **Priority** (1=high, 2=medium, 3=low)
   - **Due date** (YYYY-MM-DD, tomorrow, +3d, next-friday)
   - **Tags** (comma/space separated)
   - **State** (open, in-progress, blocked)
3. All fields optional - press Enter to skip
4. Loops until cancelled (Esc)

**Examples:**
```
Select task to plan (Esc to exit)
Task: Fix authentication bug (backend-api)
--------------------------------------------------------------
Current priority: none
Current due date: none
Current tags: none
Current state: open
--------------------------------------------------------------
Priority (1=high, 2=medium, 3=low, clear, or Enter to skip): 1
✓ Set priority to 1 (high)

Due date (YYYY-MM-DD, tomorrow, +3d, next-friday, clear, or Enter to skip): next-friday
✓ Set due date to 2026-02-07

Tags (comma or space separated, or Enter to skip): bug security
✓ Added tags: #bug #security

State (open, in-progress, blocked, or Enter to skip): in-progress
✓ Set state to in-progress
```

**Natural Language Dates:**
- `today` / `tomorrow` / `yesterday`
- `+3d` - 3 days from now
- `next-monday`, `next-friday`, etc.
- `YYYY-MM-DD` - Explicit date

**Notes:**
- Ideal for weekly planning sessions
- Complements `brain add` for capture-curate workflow
- Existing tags shown as suggestions
- Can clear metadata by entering `clear`

---

### `brain review`

**Description:** Review and process items (if implemented)

**Usage:**
```bash
brain review
```

**Notes:**
- Check implementation status with `brain review --help`
- May provide weekly review workflows

---

## Project Management

### `brain project new <name>`

**Description:** Create a new project

**Usage:**
```bash
brain project new website-redesign
brain project new blog-posts
```

**Creates:**
- Project directory in `01_active/<name>/`
- `notes.md` with initial template
- `todo.md` with sample tasks
- `.repos` file for git repository links

**Name Requirements:**
- Letters, numbers, hyphens, and underscores only
- No spaces (use hyphens instead)

**Notes:**
- Automatically selects (focuses) the new project
- Creates boilerplate structure to get started quickly

---

### `brain project list [--json]`

**Aliases:** `brain project ls`

**Description:** List all active projects

**Usage:**
```bash
# Human-readable
brain project list

# JSON output
brain project list --json
```

**Options:**
- `--json` - Output JSON format

**Output (human):**
```
Active Projects:
----------------
 * backend-api        (selected) [Repos: 1, Tasks: 12]
   frontend                      [Repos: 2, Tasks: 5]
   documentation                 [Repos: 0, Tasks: 3]
```

**Output (JSON):**
```json
[
  {
    "name": "backend-api",
    "focused": true,
    "repo_count": 1,
    "task_count": 12
  }
]
```

**Notes:**
- Shows task counts and linked repo counts
- Selected project marked with `*`

---

### `brain project select [name]`

**Description:** Focus on a specific project

**Usage:**
```bash
brain project select backend-api
brain project select  # Interactive selection
```

**Options:**
- If no name provided, shows interactive fzf selection

**Notes:**
- Focused project used as default for commands like `brain project link`
- Focus persists across sessions

---

### `brain project current`

**Description:** Show currently focused project

**Usage:**
```bash
brain project current
```

**Output:**
```
backend-api
```

**Notes:**
- Returns exit code 1 if no project is focused
- Useful for shell prompts and scripts

---

### `brain project clone <url> [name]`

**Description:** Import a git repository as a new project

**Usage:**
```bash
brain project clone https://github.com/user/repo.git
brain project clone https://github.com/user/repo.git my-project
```

**Workflow:**
1. Creates new project (name extracted from URL if not provided)
2. Links the repository
3. Clones to `~/dev/`
4. Focuses the project

**Examples:**
```bash
# Auto-name from URL
brain project clone https://github.com/user/awesome-tool.git
# Creates project: awesome-tool

# Custom name
brain project clone https://github.com/user/repo.git backend-api
# Creates project: backend-api
```

**Notes:**
- One-command setup for code-based projects
- Repository cloned to `~/dev/<repo-name>/`
- Use `brain go` afterward to enter dev mode

---

### `brain project link <url>`

**Description:** Link a git repository to current/focused project

**Usage:**
```bash
brain project link https://github.com/user/repo.git
```

**Notes:**
- Links repository URL to focused project's `.repos` file
- Multiple repos can be linked to one project
- Use `brain project pull` to clone/update linked repos
- Verifies remote accessibility (warns if unreachable)

---

### `brain project pull`

**Description:** Clone/update all linked repositories

**Usage:**
```bash
brain project pull
```

**Behavior:**
- Reads `.repos` file from focused project
- For each repository:
  - **If not cloned**: Clones to `~/dev/<repo-name>/`
  - **If already cloned**: Runs `git pull`

**Output:**
```
Project: backend-api
  backend-service
    Cloning...

  frontend-app
    Updating...
```

**Notes:**
- Clones to `~/dev/` directory by default
- Skips repos that are commented out in `.repos` (lines starting with `#`)

---

### `brain project archive <name>`

**Description:** Archive a project

**Usage:**
```bash
brain project archive old-project
brain project archive  # Archives focused project
```

**Behavior:**
- Moves project from `01_active/` to `99_archive/`
- Appends timestamp: `project-name_YYYYMMDD`
- Clears focus if archiving focused project

**Notes:**
- Does not delete the project, just moves it
- Archived projects can be manually moved back to `01_active/` if needed
- Code repositories in `~/dev/` are not affected

---

### `brain project move <project> [target-brain]`

**Description:** Move a project to another brain

**Usage:**
```bash
brain project move backend-api work
brain project move backend-api  # Interactive brain selection
```

**Options:**
- If target brain not provided, shows interactive selection

**Notes:**
- Moves entire project directory between brains
- Clears focus if moving focused project
- Target brain must exist
- Cannot move if project with same name exists in target brain

---

### `brain project delete <name>`

**Description:** Permanently delete a project

**Usage:**
```bash
brain project delete old-project
brain project delete  # Deletes focused project
```

**Notes:**
- **DESTRUCTIVE OPERATION** - Cannot be undone
- Requires typing project name to confirm
- Consider using `brain project archive` instead
- Does not delete code repositories in `~/dev/`

---

## Todo Management

### `brain todo`

**Description:** Interactive fuzzy search through tasks

**Usage:**
```bash
brain todo
```

**Behavior:**
- Shows all open tasks in fzf
- Preview shows file location with context (uses `bat` if available)
- Select task → Opens in editor at exact line

**Notes:**
- Requires fzf for interactive mode
- Opens file at the specific line number of the task
- Great for quickly finding and editing tasks

---

### `brain todo ls`

**Description:** List tasks with filters

**Usage:**
```bash
# List all open tasks
brain todo ls

# Filter by status
brain todo ls --status in-progress
brain todo ls --status blocked

# Filter by priority
brain todo ls --priority 1
brain todo ls --no-priority

# Filter by tags
brain todo ls --tag bug
brain todo ls --tag bug --tag security --tag-mode and

# Filter by due date
brain todo ls --due-today
brain todo ls --due-this-week
brain todo ls --overdue

# Sorting
brain todo ls --sort priority
brain todo ls --sort deadline
brain todo ls --sort project
brain todo ls --sort status

# Include completed tasks
brain todo ls --all

# JSON output
brain todo ls --json
```

**Options:**
- `--json` - Output JSON format
- `--all` - Include completed tasks
- `--priority <1-3>` - Filter by priority level
- `--no-priority` - Show only unprioritized tasks
- `--status <state>` - Filter by status (open, in-progress, blocked, done)
- `--tag <tag>` - Filter by tag (can specify multiple)
- `--tag-mode <and|or>` - Tag filter mode (default: or)
- `--due-today` - Tasks due today
- `--due-this-week` - Tasks due within 7 days
- `--overdue` - Tasks past due date
- `--sort <field>` - Sort by priority, deadline, project, or status

**Output:**
```
abc123 [P1] [>] Fix authentication bug #bug #security (backend-api) [Due: 2026-02-07]
def456 [P2] [ ] Update documentation (docs) [Due: 2026-02-10]
ghi789      [ ] Refactor API handlers (backend-api)
```

**Format:**
- `ID` - Unique task identifier
- `[P1]` - Priority badge (1=high, 2=medium, 3=low)
- `[>]` - Status checkbox (`[ ]`=open, `[>]`=in-progress, `[-]`=blocked, `[x]`=done)
- Task content
- Tags (with `#`)
- `(project)` - Project name
- `[Due: DATE]` - Due date (shows `[OVERDUE]` if past due)

**Notes:**
- Default sort: overdue/upcoming tasks first, then by priority
- Filters can be combined
- JSON output includes file paths and line numbers for editing

---

### `brain todo done [id]`

**Description:** Mark task as complete

**Usage:**
```bash
# By ID
brain todo done abc123

# Interactive selection
brain todo done
```

**Behavior:**
- Changes checkbox from `[ ]` to `[x]`
- Keeps task in todo.md (doesn't delete)
- Can be reopened with `brain todo reopen`

---

### `brain todo delete [id]`

**Description:** Permanently delete a task

**Usage:**
```bash
# By ID
brain todo delete abc123

# Interactive selection
brain todo delete
```

**Notes:**
- **DESTRUCTIVE** - Requires confirmation
- Removes line from todo.md
- Cannot be undone (unless using version control)

---

### `brain todo reopen [id]`

**Description:** Reopen a completed task

**Usage:**
```bash
# By ID
brain todo reopen abc123

# Interactive selection from completed tasks
brain todo reopen
```

**Behavior:**
- Changes checkbox from `[x]` back to `[ ]`
- Task becomes visible in default listings again

---

### `brain todo start <id>`

**Description:** Mark task as in-progress

**Usage:**
```bash
brain todo start abc123
```

**Behavior:**
- Changes checkbox to `[>]`
- Signals active work on the task

---

### `brain todo block <id>`

**Description:** Mark task as blocked

**Usage:**
```bash
brain todo block abc123
```

**Behavior:**
- Changes checkbox to `[-]`
- Indicates task is waiting on external dependency

---

### `brain todo unblock <id>`

**Description:** Unblock a task (set to open)

**Usage:**
```bash
brain todo unblock abc123
```

**Behavior:**
- Changes checkbox from `[-]` back to `[ ]`

---

### `brain todo status <id> <state>`

**Description:** Set task status directly

**Usage:**
```bash
brain todo status abc123 open
brain todo status abc123 in-progress
brain todo status abc123 blocked
brain todo status abc123 done
```

**Valid States:**
- `open` - Not started (`[ ]`)
- `in-progress` - Currently working (`[>]`)
- `blocked` - Waiting on dependency (`[-]`)
- `done` - Completed (`[x]`)

---

### `brain todo prio <id> <priority>`

**Description:** Set task priority

**Usage:**
```bash
brain todo prio abc123 1  # High priority
brain todo prio abc123 2  # Medium priority
brain todo prio abc123 3  # Low priority
brain todo prio abc123 0  # Clear priority
```

**Priority Levels:**
- `1` - High (P1)
- `2` - Medium (P2)
- `3` - Low (P3)
- `0` or `clear` - No priority

**Behavior:**
- Adds/updates `#p:N` tag in task line
- Affects sort order in `brain todo ls`

---

### `brain todo due <id> <date>`

**Description:** Set task due date

**Usage:**
```bash
# Explicit date
brain todo due abc123 2026-02-15

# Natural language
brain todo due abc123 tomorrow
brain todo due abc123 next-friday
brain todo due abc123 +7d

# Clear due date
brain todo due abc123 clear
```

**Date Formats:**
- `YYYY-MM-DD` - Explicit date
- `today`, `tomorrow`, `yesterday`
- `+Nd` - N days from now (e.g., `+7d`)
- `next-monday`, `next-tuesday`, etc.
- `clear` / `none` - Remove due date

**Behavior:**
- Adds/updates `#due:YYYY-MM-DD` tag
- Shows as `[Due: DATE]` in listings
- Overdue tasks highlighted with `[OVERDUE]`

---

### `brain todo tag <id> <tags...>`

**Description:** Add tags to a task

**Usage:**
```bash
# Single tag
brain todo tag abc123 bug

# Multiple tags
brain todo tag abc123 bug security urgent

# Remove tags
brain todo tag abc123 --rm bug security
```

**Options:**
- `--rm` - Remove specified tags instead of adding

**Examples:**
```bash
# Add tags
brain todo tag abc123 bug security
# Result: - [ ] Fix auth bug #bug #security

# Remove tags
brain todo tag abc123 --rm security
# Result: - [ ] Fix auth bug #bug
```

**Notes:**
- Tags are stored as `#tagname` in the task line
- Tags can be filtered with `brain todo ls --tag`
- Use lowercase for consistency (convention)

---

### `brain todo tags`

**Description:** List all tags used across tasks

**Usage:**
```bash
brain todo tags
```

**Output:**
```
Tags in use:
  bug (5 tasks)
  security (3 tasks)
  documentation (2 tasks)
  performance (1 task)
```

**Notes:**
- Shows tag usage statistics
- Useful for discovering existing tags before adding new ones
- Helps maintain tag consistency

---

## Note Management

### `brain note [project]`

**Description:** Open project notes in editor

**Usage:**
```bash
# Open focused project's notes
brain note

# Open specific project's notes
brain note backend-api
```

**Behavior:**
- Opens `notes.md` in your editor
- Uses `$EDITOR` environment variable

**Notes:**
- Notes are free-form markdown
- Can include documentation, meeting notes, ideas, etc.
- Timestamped notes from dump are stored in `notes/` subdirectory

---

### `brain note ls [project]`

**Description:** List note files

**Usage:**
```bash
# List notes in focused project
brain note ls

# List notes in specific project
brain note ls backend-api
```

**Output:**
```
notes.md
notes/2026-01-20-meeting-notes.md
notes/2026-01-25-architecture-ideas.md
```

**Notes:**
- Shows `notes.md` and all files in `notes/` directory
- Timestamped notes created by refiling from dump

---

## Context & Dev Mode

### `brain go`

**Description:** Enter project context

**Usage:**
```bash
brain go
```

**Behavior:**

**No linked repos:**
- Opens new shell in project directory

**With linked repos:**
- Launches tmux session
- Creates windows for:
  - Project directory (notes/todos)
  - Each linked repository
- Activates Python venv if present

**Examples:**
```bash
# Simple project
cd ~/brain/01_active/documentation
brain go
# Opens shell in documentation/

# Project with repos
cd ~/brain/01_active/backend-api
brain go
# Launches tmux with windows for:
#   1. backend-api (project)
#   2. backend-service (code repo)
#   3. frontend-app (code repo)
```

**Notes:**
- Requires tmux for multi-repo projects
- Automatically names tmux session after project
- Attaches to existing session if already running

---

## Sync & Utilities

### `brain sync`

**Description:** Sync with external storage (if implemented)

**Usage:**
```bash
brain sync
```

**Notes:**
- Check `brain sync --help` for implementation status
- May integrate with Syncthing, Dropbox, etc.
- Brain directories are plain files, easily syncable manually

---

## Global Options

Available on all commands:

- `--help` - Show command help
- `--version` - Show version information

**Examples:**
```bash
brain --help
brain todo ls --help
brain --version
```

---

## Task States Reference

Local Brain uses markdown checkboxes with different symbols:

| Checkbox | State | Description |
|----------|-------|-------------|
| `[ ]` | open | Not started |
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

Many commands support `--json` for programmatic access:

```bash
# Get dump items with IDs
brain dump ls --json | jq '.[0].id'

# List projects
brain project list --json | jq '.[] | select(.focused==true) | .name'

# Filter tasks
brain todo ls --priority 1 --json | jq '.[].content'
```

**Non-interactive Commands:**

These commands support scripting without user interaction:

```bash
# Refile by ID
brain refile abc123 backend-api

# Set task metadata
brain todo prio abc123 1
brain todo due abc123 2026-02-15
brain todo tag abc123 bug security
brain todo status abc123 in-progress
```

**Example AI Agent Workflow:**
```bash
#!/bin/bash
# Get first dump item
ID=$(brain dump ls --json | jq -r '.[0].id')

# Refile to focused project
PROJECT=$(brain project current)
brain refile "$ID" "$PROJECT"

# Set high priority
brain todo prio "$ID" 1
```

---

## Environment Variables

Customize Local Brain behavior:

```bash
# Root directory for all brains (default: ~/brains)
export BRAIN_ROOT="$HOME/Dropbox/Brains"

# Active brain symlink location (default: ~/brain)
export BRAIN_SYMLINK="$HOME/ActiveBrain"

# Config directory (default: ~/.config/brain)
export BRAIN_CONFIG_DIR="$HOME/.config/brain"

# Editor for notes (default: vi)
export EDITOR="nvim"
```

Add to `~/.bashrc` or `~/.zshrc` to make permanent.

---

## Tips & Best Practices

**Daily Workflow:**
```bash
# Morning: Quick captures throughout the day
brain add "Task 1"
brain add "Task 2"

# End of day: Process dump
brain refile

# Weekly: Enrich tasks
brain plan
```

**Keyboard-Driven Workflow:**
- Use shell aliases: `alias ba='brain add'`
- Use fzf for all interactive selections
- Set up shell completion for tab completion

**Project Organization:**
- Use hyphens in project names: `backend-api` not `backend_api`
- Archive completed projects regularly
- One brain per major context (Work, Personal, Learning)

**Task Management:**
- Capture without metadata (speed)
- Enrich during planning (batch)
- Use priorities sparingly (only what's urgent)
- Tags for categorization, not micro-management

**Code Integration:**
- Link repos early: `brain project link <url>`
- Pull repos before work: `brain project pull`
- Use `brain go` for instant dev environment

---

## Next Steps

- [Installation Guide](installation.md) - Set up Local Brain
- [Quickstart](index.md) - Get started quickly
- [Development Guide](development.md) - Contribute or extend Local Brain
