---
name: brain-daily
description: Run the morning brain briefing and plan the day. Use when starting work, reviewing priorities, or when the user asks "what should I work on today?" or "what's on my plate?".
compatibility: Requires local-brain MCP server (brain-mcp) to be configured.
---

# brain-daily: Daily Planning Workflow

Start the day with a structured overview of the user's Local Brain workspace. This is the "engage" phase — the daily habit that keeps work intentional.

## Trigger Phrases

- "Good morning", "start my day", "morning briefing"
- "What should I focus on today?"
- "What's on my plate?" / "Catch me up" / "What's due?"

## Pre-flight Check

Call `get_brain_overview` first. If the brain is empty (no projects, no dump items), do not run a briefing on nothing. Instead suggest:

> "Your brain is empty — let's set it up first. Tell me about what you're working on."

This naturally transitions to the `brain-setup` workflow.

## Workflow

### Step 1 — Fetch the Briefing

Call `get_daily_briefing` to retrieve:
- Overdue tasks (most urgent)
- Tasks due today
- High-priority open tasks
- In-progress work
- Blocked items
- Recent completions
- Inbox items awaiting refile

### Step 2 — Summarize Clearly

Present a concise, scannable summary. Do not dump raw data. Group by urgency:

1. **Overdue** — flag prominently with task IDs
2. **Due today** — list with IDs
3. **High priority** — P1 tasks without due dates
4. **In progress** — what was already underway
5. **Backlog** — count of items in backlog across projects (brief; this is for awareness, not action)
6. **Inbox** — count of quick-capture items waiting

Keep it to 10 items max; offer to drill into an area if there's more.

If everything is clear (nothing overdue, nothing due, low inbox):
> "Clean slate today. Here's what's in progress and what you could pick up next."

### Step 3 — Recommend a Focus

Based on the briefing:
- Suggest the single most important task to start with and explain why
- If inbox has > 5 items, suggest a quick triage pass first
- Flag any blocked tasks that might need a quick unblock

### Step 4 — Offer Follow-up Actions

After the summary, offer natural next steps based on what the briefing revealed:

- **Inbox items pending?** → "Want to do a quick triage?" (transitions to `brain-triage`)
- **Backlog needs planning?** → "You have N items in backlog — want to promote some for this week?" (use `update_todo` to set status to "open")
- **Clear top priority?** → "Want to focus on [project]?" (`set_context`)
- **Want to log the plan?** → "I can create today's daily note." (`create_daily_note`)
- **Tasks need updating?** → "Want to update any statuses or priorities?" (`update_todo`)

Do not list all options every time. Offer the 1-2 most relevant based on the actual briefing state.

## Tool Sequence

```
get_brain_overview()
  → if empty: suggest brain-setup, stop
get_daily_briefing()
  → present summary
  → recommend focus
  → (optionally) create_daily_note()
  → (optionally) set_context(project_name: ...)
  → (optionally) refile_item / update_todo / create_todo_in_project
```

## Notes

- Use `query_todos` with `due_before: today` and `status: open` for precise overdue filtering if needed.
- Do not make changes without user confirmation for destructive actions.
- After the briefing, let the user steer. Offer options; do not auto-execute a full plan.
- Tone: crisp and energizing — like a good morning standup with a trusted colleague, not a status report.
