---
name: brain-review
description: Run a weekly review of all projects, tasks, and progress. Use at the end of the week, when the user says "weekly review", "let's review", "how did this week go?", or "what's the state of everything?".
compatibility: Requires local-brain MCP server (brain-mcp) to be configured.
---

# brain-review: Weekly Review

Walk the user through a comprehensive review of their entire workspace. This is the "reflect" phase — the practice that keeps the system trustworthy. Inspired by GTD's weekly review: get clear, get current, get creative.

## Trigger Phrases

- "Weekly review", "Let's review", "Review my projects"
- "How did this week go?", "What's the state of everything?"
- "End of week check-in", "Friday review"
- "What did I accomplish this week?"

## Workflow

### Step 1 — Gather Everything

Make these calls to build the full picture:

1. `get_brain_overview` — all projects and their stats
2. `query_todos` with `status: done` and `completed_after: [7 days ago]` — recent completions
3. `query_todos` with `status: open` — all open tasks
4. `query_todos` with `status: blocked` — blocked items
5. `get_dump_items` — inbox state

### Step 2 — Celebrate Progress

Start positive. Present what was accomplished this week:

> "This week you completed [N] tasks across [X] projects."

List the completed tasks grouped by project. Acknowledge effort — this is what keeps people motivated.

### Step 3 — Review Each Active Project

For each project with open tasks:

1. Show project name, open task count, and any in-progress/blocked items
2. Ask: "Is this still active? Anything changed?"
3. Look for signals:
   - **Stale tasks**: open tasks with no recent activity — ask if they're still relevant
   - **Blocked tasks**: surface them prominently — "This has been blocked. Can we unblock it?"
   - **Missing deadlines**: overdue tasks — "This was due [date]. What's the status?"

Keep this conversational, not interrogative. If a project is on track, a quick "Looks good, moving on" is fine.

### Step 4 — Identify Stale Work

Query for tasks that have been open a long time without progress:
- Tasks created > 2 weeks ago with status still `open` (never moved to `in-progress`)
- Flag these specifically:

> "These tasks have been sitting for a while. Want to reprioritize, defer, or drop any of them?"

Options for each:
- **Keep as-is** — still relevant, just not yet
- **Reprioritize** — change priority or add a due date to create urgency
- **Drop** — mark as done or delete if no longer needed
- **Break down** — maybe the task is too big; split into smaller tasks

### Step 5 — Process Inbox

If the inbox has items:
> "You have [N] items in your inbox. Want to do a quick triage?"

If yes, transition to the `brain-triage` workflow. If not, note it as a todo for next session.

### Step 6 — Look Ahead

Ask about the coming week:

> "Anything coming up next week? New deadlines, priorities, or projects?"

Create any new tasks or adjust priorities based on the response.

### Step 7 — Generate Review Note

Offer to create a weekly review note:

Call `create_daily_note` (or `create_project_note` in a "reviews" project) with a summary:
- Tasks completed this week
- Key decisions made during review
- Focus areas for next week
- Any items deferred or dropped

## Tool Sequence

```
get_brain_overview()
query_todos(status: done, completed_after: [7 days ago])
query_todos(status: open)
query_todos(status: blocked)
get_dump_items()
  → celebrate completions
  → review each project
  → surface stale/blocked work
  → (optionally) triage inbox
  → look ahead
  → update_todo() / delete_todo() / create_todo_in_project() as needed
  → create_daily_note() or create_project_note() with summary
```

## Notes

- The weekly review should take 5-15 minutes, not 45. Keep it moving.
- Do not read back every single task — summarize and highlight what needs attention
- If a project has > 10 open tasks, suggest the user focus on the top 3-5 rather than reviewing all
- Respect "skip" — if the user says a project is fine, move on without drilling in
- Tone: reflective and supportive — like a good coach checking in, not an auditor
- This is the most important skill for long-term system trust. If reviews happen, the system stays current.
