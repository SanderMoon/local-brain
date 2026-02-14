---
name: brain-setup
description: Set up a new Local Brain workspace through guided conversation. Use when the brain is empty, when the user just installed brain, or when they say "set up my brain", "get started", or "I'm new here".
compatibility: Requires local-brain MCP server (brain-mcp) to be configured.
---

# brain-setup: First-Time Setup Wizard

Guide the user through setting up their Local Brain workspace conversationally. This should feel like talking to a new executive assistant on their first day — curious, helpful, and efficient.

## Trigger Phrases

- "Set up my brain", "Get started", "I'm new here"
- "Help me organize my projects"
- "I just installed brain"
- First interaction when the brain is empty

## Pre-flight Check

Call `get_brain_overview` to assess the current state:
- If **no brain exists or brain is completely empty** (no projects, no dump items): run the full setup
- If **projects already exist**: offer to add more projects or refine existing ones — do not re-run full setup
- If **only dump items exist**: offer to organize them into projects as part of setup

## Workflow

### Step 1 — Welcome and Context

Briefly explain what Local Brain does in one or two sentences. Do not lecture. Then ask:

> "What are you working on right now? Tell me about your projects — they can be work tasks, side projects, personal goals, anything."

Let the user talk naturally. Extract project names, descriptions, and any mentioned tasks or deadlines from their response.

### Step 2 — Create Projects

For each project identified:
1. Suggest a short, kebab-case name (e.g., "website-redesign", "q1-planning")
2. Confirm the name and description with the user
3. Call `create_project` for each

Do not create more than 3-5 projects in the first pass. If the user mentions more, note them and say: "Let's start with these — you can always add more later."

### Step 3 — Populate Initial Tasks

For each project, ask:

> "What needs to happen next for [project]? Any deadlines?"

Convert responses into concrete tasks with:
- Clear, actionable descriptions (imperative form: "Ship v2", not "v2 shipping")
- Due dates when mentioned ("by Friday" → the actual date)
- Priority 1 for anything the user emphasizes as urgent or important

Call `create_todo_in_project` in batches per project.

### Step 4 — Handle "I Don't Know"

If the user is vague or says "I just want to try it":
- Create a single project called "getting-started"
- Add a few example tasks: "Add your first real project", "Try the daily briefing tomorrow morning", "Capture a quick thought with brain dump"
- Explain these are just examples they can delete

### Step 5 — Demonstrate Value

After creating at least one project with tasks, run a mini version of the daily briefing:

> "Here's what your morning briefing will look like tomorrow:"

Call `get_daily_briefing` and present a condensed summary. This shows the system working immediately.

### Step 6 — Close the Loop

End with:
1. A summary of what was created (X projects, Y tasks)
2. Suggest the natural next step: "Say 'good morning' tomorrow and I'll brief you on your day."
3. Mention they can capture quick thoughts any time: "Just tell me to remember something and I'll save it to your inbox."

## Tool Sequence

```
get_brain_overview()
  → assess state
  → conversational Q&A
  → create_project() (per project)
  → create_todo_in_project() (per project)
  → get_daily_briefing()
  → present summary and next steps
```

## Notes

- Never create empty projects — every project should have at least one task
- Infer structure from natural language; do not make the user fill out forms
- Keep the conversation to 3-5 exchanges max — do not over-interview
- If the user provides a wall of text, parse it intelligently and confirm your interpretation
- Tone: warm, efficient, slightly enthusiastic — like a competent assistant who's genuinely excited to help organize
