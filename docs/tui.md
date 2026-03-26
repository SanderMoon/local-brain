# TUI Reference

Launch with `brain` (no arguments). Press `q` or `Ctrl+C` to exit.

## Layout

- **Sidebar** (left): brains and projects
- **Content** (right): notes, todo list, or kanban board
- **Status bar** (bottom): current brain, project, filters, and keybinding hints

`Tab` toggles focus between sidebar and content.

## Keybindings

### Global

| Key | Action |
|-----|--------|
| `q` / `Ctrl+C` | Quit |
| `Tab` | Toggle sidebar / content focus |
| `1` / `2` / `3` | Switch view (Notes, List, Kanban) |
| `?` | Toggle help overlay |
| `r` | Refresh data |
| `c` | Toggle completed todos |
| `a` | Toggle all projects / current project |

### Sidebar

| Key | Action |
|-----|--------|
| `j` / `k` | Move selection |
| `h` / `l` | Switch Brains / Projects section |
| `Enter` | Activate brain or project |

### Todo List

| Key | Action |
|-----|--------|
| `j` / `k` | Move selection |
| `e` | Open in editor (jumps to line) |
| `p` | Set priority (1/2/3) |
| `s` | Cycle status: backlog > open > in-progress > blocked > done |
| `d` | Set due date (`YYYY-MM-DD`, `tomorrow`, `+3d`, etc.) |
| `t` | Add tags (space-separated) |

### Kanban

| Key | Action |
|-----|--------|
| `h` / `l` | Move column focus left / right |
| `j` / `k` | Move selection within column |
| `e` | Open in editor |
| `m` / `b` | Move task forward / backward one status |
| `B` | Toggle backlog column |

### Notes

| Key | Action |
|-----|--------|
| `j` / `k` | Move selection |
| `e` | Open in editor |
| `p` | Toggle preview pane |

### Input Prompts

`Enter` submits, `Esc` cancels.

## Inline Editing

In the Todo List, you can modify task metadata without leaving the TUI. Press `e` for full editing in your configured editor (suspends the TUI, opens at the correct line, refreshes on return).

## Kanban Board

Five columns: **Backlog**, **Open**, **In Progress**, **Blocked**, **Done**.

- Tasks sorted by priority within each column (P1 first).
- Columns scroll independently; `^`/`v` indicators appear when content overflows.
- Backlog is collapsed by default (shows item count only). Press `B` to expand.
- Each card shows: priority badge + content, project name, and due date or tags.

## Notes View

Split layout by default: note list on the left, preview on the right. Press `p` to toggle the preview pane. Each entry shows title and creation date.

## Filters

| Toggle | Key | Default | Note |
|--------|-----|---------|------|
| Show completed | `c` | Off | Kanban Done column always shows completed tasks regardless. |
| All projects | `a` | Off | Show tasks from every project. |

Active filters are shown in the status bar.
