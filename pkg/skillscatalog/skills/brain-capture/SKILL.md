---
name: brain-capture
description: Quickly capture tasks, ideas, and notes with smart metadata inference. Use when the user says "remind me", "save this", "add a task", "I need to", "don't let me forget", or any variant of quick capture.
compatibility: Requires local-brain MCP server (brain-mcp) to be configured.
---

# brain-capture: Smart Quick Capture

Capture thoughts, tasks, and notes instantly with minimal friction. Speed is everything — the capture must be faster than the thought fading. Infer metadata from natural language; do not interrogate.

## Trigger Phrases

- "Remind me to...", "Don't let me forget..."
- "Add a task: ...", "Save this: ..."
- "I need to...", "Note to self: ..."
- "Quick thought: ...", "Idea: ..."
- Any statement that implies something should be remembered

## Workflow

### Step 1 — Parse the Input

Extract from the user's natural language:

| Signal | Example | Inference |
|--------|---------|-----------|
| Action verb | "Ship the feature" | Task (not note) |
| Time reference | "by Friday", "next week", "before March" | Due date |
| Urgency words | "urgent", "ASAP", "critical", "important" | Priority 1 |
| Project mention | "for the website", "re: backend" | Project association |
| "idea", "thought", "note" | "Idea for the landing page" | Note (not task) |
| Multiple items | "I need to X, Y, and Z" | Multiple captures |

### Step 2 — Route the Capture

**If a project is clearly identified** (user mentions it by name or context is unambiguous):
- Call `create_todo_in_project` directly with inferred metadata
- Confirm: "Added to [project]: [task] (due Friday, P1)"

**If the project is ambiguous or no project context:**
- Call `add_to_dump` to send to inbox
- Confirm: "Captured to inbox: [task]"
- Do NOT ask "which project?" — that breaks the speed contract. Inbox exists for this.

**If it's a note, not a task:**
- Call `add_to_dump` with type "note" and a title
- Confirm: "Saved note: [title]"

**If multiple items are mentioned:**
- Parse each one separately
- Capture all in a single batch call
- Confirm with a list: "Captured 3 items: ..."

### Step 3 — Confirm and Return

Confirmation must be **one line**. Examples:

> "Captured: 'Review PR for auth module' -> inbox"

> "Added to backend-v2: 'Fix rate limiting bug' (P1, due 2026-02-18)"

> "Saved 3 tasks to inbox: 'Update docs', 'Email Sarah', 'Book flight'"

Then **stop**. Do not offer follow-up actions. Do not ask if they want to add more. The user will say more if they want more. Return control immediately.

## Tool Sequence

```
# Direct to project (clear association):
create_todo_in_project(project, [{content, priority?, due_date?, tags?}])

# To inbox (ambiguous or no project):
add_to_dump(type: "task", content)

# Notes:
add_to_dump(type: "note", title, content)

# Multiple items:
create_todo_in_project(project, [{...}, {...}, {...}])
# or
add_to_dump() × N
```

## Metadata Inference Rules

### Due Dates
- "today" → today's date
- "tomorrow" → tomorrow's date
- "Friday" / "next Friday" → the next occurrence of that day
- "next week" → next Monday
- "end of month" → last day of current month
- "by March 1st" → 2026-03-01

### Priority
- "urgent", "ASAP", "critical", "blocker" → Priority 1
- "important", "should", "need to" → Priority 2
- "eventually", "sometime", "low priority", "when I get to it" → Priority 3
- No signal → do not set priority (leave null)

### Tags
- Do not add tags unless the user explicitly mentions them
- If the user says "tag it as X" or "#X", add the tag

## Notes

- Speed is the primary metric. One tool call, one confirmation line, done.
- Never ask clarifying questions for a simple capture. Bias toward action.
- If you're unsure whether something is a task or note, default to task — tasks are easier to reclassify later
- The inbox is the safety net. When in doubt, capture to inbox. Triage happens later.
- This skill activates frequently — it should be nearly invisible in its efficiency
