---
name: brain-focus
description: Start a focused work session on a specific project. Use when the user says "let's work on", "I'm focusing on", "deep work", "time to work on [project]", or wants to concentrate on a single project without distractions.
compatibility: Requires local-brain MCP server (brain-mcp) to be configured.
---

# brain-focus: Deep Work Session

Set up and maintain a focused work context for a single project. The agent becomes a project-scoped collaborator — loading full context, suggesting next actions, and deflecting distractions to the inbox.

## Trigger Phrases

- "Let's work on [project]", "Focus on [project]"
- "I'm going to work on [project] now"
- "Deep work time", "Focus mode"
- "What should I do next on [project]?"

## Workflow

### Step 1 — Set the Focus

Determine which project to focus on:
- If the user names a project: use it directly
- If ambiguous: call `get_brain_overview` and ask which project

Call `set_context` to set the focused project.

### Step 2 — Load Project Context

Call `get_project_context` with `include_note_content: "preview"` to get:
- All open and in-progress tasks
- Recent notes
- Linked repositories

Present a concise project dashboard:

> **[Project Name]**
> [Description]
>
> In progress: [list]
> Up next: [highest priority open tasks]
> Blocked: [any blocked items]
> Recent notes: [titles]

### Step 3 — Suggest Next Action

Based on the project state, recommend what to work on:

1. **In-progress tasks first** — continue what was already started
2. **Highest priority unblocked tasks** — if nothing is in-progress
3. **Blocked items** — if something could be unblocked quickly, suggest that

> "You were working on [task]. Want to continue, or pick something else?"

Mark the chosen task as `in-progress` with `update_todo`.

### Step 4 — During the Session

While the user is working in focus mode:

**Capture stray thoughts to inbox:**
When the user mentions something unrelated to the current project:
> "Got it — I'll save that to your inbox so you can deal with it later."

Call `add_to_dump` for the stray thought. Return focus to the project immediately. Do not switch context.

**Track progress:**
When the user completes something:
- Update the task status to `done` via `update_todo`
- Suggest the next task from the project

**Add new tasks:**
If the user discovers new work during the session:
- Add it to the current project via `create_todo_in_project`
- Do not break focus to plan or organize — just capture and continue

### Step 5 — Session Wrap-up

When the user signals they're done ("that's enough", "I'm stopping", "wrapping up"):

1. Summarize what was accomplished:
   > "Good session. You completed [N] tasks on [project]: [list]. [M] tasks remaining."

2. Update any in-progress tasks that weren't finished — leave them as `in-progress` for next time

3. Offer to create a session note:
   > "Want me to save a note about what you worked on?"

   If yes, call `create_project_note` with a session summary.

4. Mention if any stray thoughts were captured:
   > "I saved [N] items to your inbox during the session. You can triage them later."

## Tool Sequence

```
set_context(project_name: project)
get_project_context(project, include_note_content: "preview")
  → present dashboard
  → suggest next action
  → update_todo(status: "in-progress") for chosen task
  → during session:
    → add_to_dump() for stray thoughts
    → update_todo(status: "done") for completions
    → create_todo_in_project() for discovered work
  → on wrap-up:
    → summarize
    → create_project_note() (optional)
```

## Focus Principles

- **One project, one context.** Do not switch between projects during a focus session. Anything off-topic goes to inbox.
- **Momentum over perfection.** Finishing a task and moving to the next is better than perfecting one task endlessly.
- **Minimal interruption.** Capture stray thoughts silently — one line confirmation, then back to work.
- **Clear boundaries.** The session has a start and an end. The wrap-up summary creates closure.

## Notes

- If the user hasn't set up any projects yet, suggest running `brain-setup` or `brain-plan` first
- Do not proactively suggest focus sessions — this skill activates when the user explicitly wants to focus
- Keep the dashboard brief. The user is here to work, not to read a report.
- If all tasks in the project are done, congratulate and suggest planning next steps or archiving the project
- Tone: calm, focused, and supportive — like a good pair programmer who keeps you on track
