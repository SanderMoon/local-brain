# File Format Reference

## Directory Layout

A brain is a directory. Projects are subdirectories. Everything is plain text.

```
~/brains/work/                      <- a "brain" (workspace)
├── 00_dump.md                      <- inbox for quick captures
├── 00_daily/                       <- daily notes
│   ├── 2026-03-24.md
│   └── 2026-03-25.md
├── 01_active/                      <- active projects live here
│   ├── website-redesign/
│   │   ├── todo.md                 <- tasks
│   │   ├── notes.md                <- note index (legacy inline notes)
│   │   ├── notes/                  <- individual note files
│   │   │   └── 2026-03-20-architecture-decisions.md
│   │   ├── description.md          <- project description (optional)
│   │   └── .repos                  <- linked git repositories
│   └── api-migration/
│       ├── todo.md
│       ├── notes.md
│       └── notes/
├── 02_areas/                       <- long-running areas of responsibility
├── 03_resources/                   <- reference material
├── 99_archive/                     <- archived projects (timestamped)
│   └── old-project-20260301/
└── _templates/                     <- templates for new files
    ├── new-note.md
    ├── new-project.md
    └── daily-note.md
```

### Top-level directories

| Directory | Purpose |
|-----------|---------|
| `00_dump.md` | Quick capture inbox. Items land here and get refiled to projects. |
| `00_daily/` | Daily notes, one file per day (`YYYY-MM-DD.md`). |
| `01_active/` | Active projects. Each subfolder is a project. |
| `02_areas/` | Long-running areas of responsibility (not time-bound). |
| `03_resources/` | Reference material and resources. |
| `99_archive/` | Archived projects. Format: `{name}-YYYYMMDD/`. |
| `_templates/` | Templates for new notes, projects, and daily notes. |

### Project directory contents

Each project folder inside `01_active/` can contain:

| File | Required | Purpose |
|------|----------|---------|
| `todo.md` | Yes | Task list with sections for active and completed work. |
| `notes.md` | Yes | Note index file. Can contain inline notes (legacy) or links to note files. |
| `notes/` | No | Directory of individual note files. |
| `description.md` | No | Free-form project description. |
| `.repos` | No | List of linked git repository URLs. |

---

## todo.md

### Structure

```markdown
# Tasks

## Active

- [ ] Open task waiting to be worked on
- [>] Task currently in progress
- [-] Blocked task
- [~] Backlog item (not yet promoted to active)

## Completed

- [x] Finished task
```

### Checkbox syntax

| Checkbox | Status | Meaning |
|----------|--------|---------|
| `- [ ]` | open | Ready to work on |
| `- [>]` | in-progress | Currently being worked on |
| `- [-]` | blocked | Waiting on something |
| `- [~]` | backlog | Captured but not yet planned |
| `- [x]` | done | Completed |

New tasks created via `brain add` or `brain refile` default to backlog (`- [~]`). Use `brain plan` or `brain todo status` to promote them.

### Metadata tags

Inline metadata tags are appended after the task text.

**Priority** (1 = high, 2 = medium, 3 = low):

```markdown
- [ ] Fix critical bug p:1
- [ ] Update docs p:2
- [ ] Refactor utils p:3
```

Legacy format with hash prefix is also supported: `#p:1`, `#p:2`, `#p:3`.

**Due dates** (YYYY-MM-DD):

```markdown
- [ ] Ship release due:2026-04-01
```

Legacy format: `#due:2026-04-01`.

**Freeform tags** (always use `#` prefix):

```markdown
- [ ] Fix login flow #bug #urgent
- [ ] Add dark mode #feature #design
```

Tags are hashtags without colons. Hashtags with colons (like `#p:1`, `#due:...`, `#captured:...`) are metadata, not freeform tags.

**Combined example:**

```markdown
- [ ] Finalize color palette p:1 due:2026-04-01 #design
```

### System metadata (HTML comments)

System metadata is stored in HTML comments at the end of task lines. Invisible when rendering markdown, parsed by Local Brain.

```markdown
- [ ] Fix auth bug <!-- id:a1b2c3 captured:2026-03-20 -->
- [x] Write tests <!-- id:d4e5f6 captured:2026-03-15 done:2026-03-22 -->
```

Format: `<!-- key:value key:value -->`

| Field | Format | Purpose |
|-------|--------|---------|
| `id` | 6-character hex string | Stable identifier for the task. |
| `captured` | `YYYY-MM-DD` | Date the task was created. |
| `done` | `YYYY-MM-DD` | Date the task was completed (only on done tasks). |

The `id` is a random 6-character hex string (3 bytes). Used to reference tasks in CLI commands (e.g., `brain todo done a1b2c3`).

**Legacy inline format:** Older files may use `#id:a1b2c3`, `#captured:2026-03-20`, `#done:2026-03-22` instead of HTML comments. Both formats are supported.

---

## 00_dump.md (Inbox)

### Format

```markdown
# Dump

Quick capture landing zone. Process with `brain refile`.

- [ ] Fix authentication bug <!-- id:a1b2c3 captured:2026-03-25 -->
- [ ] Review PR #452 <!-- id:d4e5f6 captured:2026-03-25 -->

[Note] Architecture meeting <!-- id:789abc captured:2026-03-25 -->
    Decided to move to microservices
    Redis for session caching
    Follow up with platform team
```

### Tasks in the dump

Same syntax as todo.md tasks. Always `- [ ]` (open) checkbox:

```markdown
- [ ] Task description <!-- id:abc123 captured:2026-03-25 -->
```

### Notes in the dump

Notes use a `[Note]` prefix on the title line, followed by indented content (4 spaces):

```markdown
[Note] Meeting notes <!-- id:abc123 captured:2026-03-25 -->
    First line of content
    Second line of content
    More content here
```

The parser treats consecutive lines indented with 4 spaces as part of the note body. A non-indented line (or end of file) terminates the note.

Markdown headers (`#`, `##`, etc.) and blank lines are skipped by the parser. Everything else is either a task or a note.

---

## Notes

Notes exist as individual files in `notes/`, or as inline entries in `notes.md`.

### Note files (notes/ directory)

Files live in `{project}/notes/` with YAML frontmatter.

**Filename convention:** `YYYY-MM-DD-slug.md`

```
notes/
├── 2026-03-20-architecture-decisions.md
├── 2026-03-22-sprint-planning.md
└── 2026-03-25-api-design.md
```

**File format:**

```markdown
---
title: Architecture Decisions
date: 2026-03-20
project: website-redesign
tags: []
---

# Architecture Decisions

Content goes here.
```

**Frontmatter fields:**

| Field | Format | Purpose |
|-------|--------|---------|
| `title` | Free text | Display title of the note. |
| `date` | `YYYY-MM-DD` | Creation date. |
| `project` | Project name | Which project this note belongs to. |
| `tags` | YAML list | Optional tags (e.g., `[meeting, architecture]`). |

### notes.md (index file)

Index file. When note files are created, a link is added under `## Notes`:

```markdown
# project-name Notes

## Notes

- [2026-03-25 API Design](notes/2026-03-25-api-design.md)
- [2026-03-22 Sprint Planning](notes/2026-03-22-sprint-planning.md)
```

**Legacy inline format:** Older files may contain inline notes using `[Note]` syntax (same as dump). The parser also supports a fallback where the first line is `# Title` followed by `Created: YYYY-MM-DD`.

---

## .repos

One git URL per line. Empty lines and `#`-prefixed lines are ignored.

```
https://github.com/user/repo.git
git@github.com:user/another-repo.git
# This is a comment
https://github.com/user/third-repo.git
```

Repo names are extracted from URLs and mapped to `{dev_dir}/{repo-name}` (default `~/dev/`, configurable via `brain config dev_dir`).

---

## description.md

Free-form markdown. No required structure. Used by AI agents and the TUI for project context.

```markdown
Website redesign project for Q2 2026. Goal is to modernize the frontend
with a new design system and improve page load times by 50%.

Key stakeholders: design team, frontend team, product.
```

---

## Daily Notes (00_daily/)

One file per day, named `YYYY-MM-DD.md`. Created by `brain daily` or MCP.

```markdown
---
title: 2026-03-25
date: 2026-03-25
---

# 2026-03-25

## Daily Briefing

- [ ] [website-redesign] Finalize color palette (due: 2026-03-20)

## Today's Focus

-

## Notes

```

`## Daily Briefing` is auto-populated with overdue tasks at creation time.

---

## config.json

Stored at `~/.config/brain/config.json`.

```json
{
  "current": "work",
  "brains": {
    "work": {
      "path": "/Users/username/brains/work",
      "created": "2026-01-01",
      "focus": "website-redesign"
    },
    "personal": {
      "path": "/Users/username/brains/personal",
      "created": "2026-01-15"
    }
  },
  "editor": "nvim"
}
```

**Top-level fields:**

| Field | Type | Purpose |
|-------|------|---------|
| `current` | string | Name of the active brain. |
| `brains` | object | Map of brain name to brain info. |
| `editor` | string | Preferred editor (optional). Falls back to `$EDITOR`. |

**Brain info fields:**

| Field | Type | Purpose |
|-------|------|---------|
| `path` | string | Absolute path to the brain directory. |
| `created` | string | Creation date (`YYYY-MM-DD`). |
| `focus` | string | Currently focused project within this brain (optional). |

The active brain is symlinked to `~/brain` (or `~/brain-dev` in dev mode).

**Environment variable overrides** (mainly for testing):

| Variable | Default | Purpose |
|----------|---------|---------|
| `BRAIN_ROOT` | `~/brains` | Root directory where brains are stored. |
| `BRAIN_SYMLINK` | `~/brain` | Symlink location for the active brain. |
| `BRAIN_CONFIG_DIR` | `~/.config/brain` | Configuration directory. |
| `BRAIN_CONFIG_PATH` | `~/.config/brain/config.json` | Configuration file path. |
