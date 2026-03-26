# Workflows

Practical patterns for using Local Brain day-to-day.

---

## Multi-Brain Workflows

### When to use brains vs projects

A **brain** represents a life context: work, personal, freelance. A **project** is a focus area within that context: a feature you're shipping, a trip you're planning, a client engagement.

If two things never share tasks or notes, they belong in separate brains. If they might, use separate projects within the same brain.

Most people start with a single brain and add more only when the separation is genuinely useful.

### Switching context

```bash
brain switch work            # switch directly
brain switch                 # interactive selection
brain current                # check which brain is active
```

When you switch, the `~/brain` symlink re-points to the new brain's directory. All `brain` commands now operate on it, and your shell prompt updates (if configured).

### Archiving projects

When a project is done or paused, archive it instead of deleting:

```bash
brain project archive api-migration    # moves to 99_archive/
brain project delete api-migration     # permanent removal (asks for confirmation)
```

Prefer archiving unless you are certain you won't need the data.

---

## Editor Integration

### Obsidian

Open a brain directory (`~/brains/work/`) as an Obsidian vault. Your projects, notes, and tasks appear as navigable markdown files. Edits sync both ways since they share the same files on disk.

Tips:
- Open `~/brains/` as the vault to see all brains at once
- The `_templates/` directory works with Obsidian's Templates plugin
- Obsidian's daily notes pair well with `brain daily`

### VS Code

Open a brain directory directly, or create a `.code-workspace` to include both your brain and related code repos:

```json
{
  "folders": [
    { "path": "~/brains/work", "name": "Brain: Work" },
    { "path": "~/dev/api-server", "name": "API Server" }
  ]
}
```

### Editor configuration

Local Brain checks these in order when launching an editor (e.g., `brain edit`):

1. Brain config editor (`brain config editor nvim`)
2. `nvim` if installed
3. `vim` if installed
4. `$EDITOR` environment variable

### The storm command

`brain storm` launches an AI agent in your brains root directory with access to all your brain data:

```bash
brain storm                      # Claude Code (default)
brain storm --agent aider        # aider
brain storm -a llm               # llm
```

Useful for brainstorming, weekly reviews, or bulk reorganization. Create a `CLAUDE.md` in `~/brains/` to give Claude Code standing instructions about how you want your brain managed.

---

## Sync

### File sync (Syncthing, Dropbox, iCloud)

Brains are plain directories of markdown files, so any file sync tool works.

**Syncthing** (recommended for privacy): add `~/brains/` as a shared folder. Conflicts produce `.sync-conflict` files you resolve manually.

**Cloud sync**: either set `BRAIN_ROOT` to a path inside your sync folder, or symlink:

```bash
export BRAIN_ROOT=~/Dropbox/brains
# or
ln -s ~/brains ~/Dropbox/brains
```

Watch out for sync conflicts on `todo.md` if you edit on two devices before syncing.

### Git for version control

You can version-control individual brains or the entire brains directory:

```bash
cd ~/brains/work && git init && git add -A && git commit -m "Initial brain state"
```

Suggested `.gitignore`:

```
.obsidian/
.vscode/
*.swp
*~
.sync-conflict-*
```

Auto-commit script (run via cron):

```bash
#!/bin/bash
cd ~/brains/work || exit
git add -A
git diff --cached --quiet || git commit -m "auto: $(date +%Y-%m-%d)"
```

Git and file sync can coexist: file sync for real-time access across devices, git for history and rollback.
