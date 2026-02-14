---
name: brain-triage
description: Process inbox items systematically — refile to projects, add metadata, or discard. Use when the inbox has items, after a daily briefing reveals pending captures, or when the user says "process my inbox", "triage", or "clean up my dump".
compatibility: Requires local-brain MCP server (brain-mcp) to be configured.
---

# brain-triage: Inbox Processing

Systematically process items in the inbox (00_dump.md), helping the user decide where each item belongs. This is the "clarify and organize" phase of the GTD loop.

## Trigger Phrases

- "Process my inbox", "Triage my dump"
- "Clean up my inbox", "Organize my captures"
- "What's in my inbox?"
- After a daily briefing shows inbox items

## Workflow

### Step 1 — Fetch Inbox and Context

Call `get_dump_items` to retrieve all inbox items.

If the inbox is empty:
> "Inbox zero — nothing to process. Nice work."

If items exist, also call `get_brain_overview` to get the list of existing projects for refile suggestions.

### Step 2 — Group and Present

Before processing item-by-item, scan all items and group related ones:
- Items that mention the same project or topic
- Items captured close together that form a sequence

Present a summary:
> "You have [N] items in your inbox. I see [X] that look related to [project], [Y] that might be new topics. Let's process them."

### Step 3 — Process Each Item (or Group)

For each item or group of related items:

1. **Show the item** with its ID and content
2. **Suggest an action** based on content analysis:
   - **Refile to [project]** — if it clearly matches an existing project
   - **Create new project** — if it represents a distinct new effort
   - **Convert to note** — if it's reference material, not an action
   - **Add metadata** — if it's a task but needs priority/due date
   - **Discard** — if it's stale, duplicate, or no longer relevant
3. **Wait for confirmation** — never refile without the user agreeing

For batches of related items going to the same project, suggest a batch refile:
> "These 4 items all look like they belong in [project]. Refile all of them?"

### Step 4 — Execute Actions

Based on user decisions:
- `refile_item` — for moves to existing projects (use batch array for multiple items)
- `create_project` then `refile_item` — for new project creation
- `create_project_note` — for converting items to notes
- `update_todo` — for adding priority or due dates after refile
- `delete_todo` — for discarding items (confirm first)

### Step 5 — Summary

After processing all items:
> "Done! Processed [N] items: [X] refiled, [Y] new projects created, [Z] discarded. Inbox is at [remaining] items."

If inbox is now empty, celebrate briefly.

## Tool Sequence

```
get_dump_items()
get_brain_overview()
  → group and present
  → for each item/group:
    → suggest action
    → await confirmation
    → refile_item() / create_project() / create_project_note() / delete_todo()
  → present summary
```

## Notes

- Speed over perfection — if the user says "yeah, refile all of those", do it in one batch call
- When suggesting a project for refile, explain *why* you think it matches ("this mentions the API, which relates to your backend-v2 project")
- If an item is ambiguous, offer 2-3 options rather than guessing
- Keep momentum — do not over-discuss each item. Quick decisions are the goal
- If the inbox has > 15 items, offer to do a "quick pass" (obvious refiles first) vs. "thorough pass" (every item)
- Tone: systematic and efficient — like processing email with a good filter
