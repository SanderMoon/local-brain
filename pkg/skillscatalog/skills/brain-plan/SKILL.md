---
name: brain-plan
description: Plan a project by breaking goals into concrete tasks with priorities and deadlines. Use when the user says "let's plan", "break this down", "I need to figure out how to", "help me plan", or when starting a new project.
compatibility: Requires local-brain MCP server (brain-mcp) to be configured.
---

# brain-plan: Project Planning

Help the user decompose a goal into concrete, actionable tasks with priorities, deadlines, and logical ordering. This is collaborative planning — the agent facilitates, the user decides.

## Trigger Phrases

- "Let's plan [project]", "Help me plan..."
- "Break this down", "How should I approach this?"
- "I need to figure out how to..."
- "New project: ...", "I want to start..."
- "What are the steps for...?"

## Workflow

### Step 1 — Identify the Scope

Determine if this is:
- **A new project** that needs to be created
- **An existing project** that needs task breakdown
- **A vague goal** that needs clarification first

If new: ask for a one-sentence description of what "done" looks like.
If existing: call `get_project_context` to see current state.
If vague: help clarify with one question — "What would it look like if this was finished?"

### Step 2 — Brainstorm Tasks

Ask:
> "What are the main things that need to happen to get this done?"

Let the user brain-dump. Then help structure their response:

1. **Break large items into smaller tasks** — if something takes more than a day of effort, it probably should be 2-3 tasks
2. **Make tasks actionable** — each should start with a verb: "Write", "Build", "Research", "Contact", "Deploy"
3. **Identify dependencies** — "What needs to happen first?"
4. **Spot missing steps** — "I notice you mentioned X but not Y — do we need that?"

Present the proposed task list and ask for confirmation before creating anything.

### Step 3 — Add Metadata

For the confirmed tasks:

1. **Sequence** — order tasks logically (dependencies first)
2. **Priority** — ask about the first 2-3 most important tasks: "Which of these is the most critical to get right?"
3. **Deadlines** — if the project has an overall deadline, work backwards: "If this needs to be done by [date], when should each piece be finished?"
4. **First task** — identify the single next action: "If you were starting right now, what would you do first?"

### Step 4 — Create Everything

Execute the plan:
1. `create_project` if new (with description)
2. `create_todo_in_project` with the full task list, including priorities and due dates
   - New tasks default to **backlog** status. Only set `status: "open"` for tasks the user wants to work on immediately (this week).
   - Ask: "Which of these do you want to start on now?" — promote those to open, leave the rest in backlog.
3. `set_context` to focus the new/existing project

### Step 5 — Document the Plan

Create project notes that capture context beyond what fits in task titles. Tasks say *what*; notes say *why* and *how*.

**Always create a planning note** via `create_project_note`:
- Project goal (the "definition of done")
- Key milestones or phases
- Decisions made during planning and their rationale
- Open questions, risks, or unknowns

**For larger features**, also create implementation notes for complex areas:
- If the user describes technical approach, architecture, or design decisions — capture these in a separate note (e.g., "Architecture: auth flow", "Design: API schema")
- If the user shares requirements, constraints, or feature requests from others — save the original context as a note, not just the distilled tasks
- If there are multiple approaches discussed and one was chosen — document what was considered and why

The goal is statefulness: when the user (or agent) returns to this project days or weeks later, the notes should provide enough context to pick up without re-explaining everything.

**Rule of thumb:** If the planning conversation surfaces information that would be lost when the chat session ends, it belongs in a note.

### Step 6 — Close with Next Action

End with clarity:
> "Your plan for [project] is set: [N] tasks, first milestone by [date]. The next thing to do is: [first task]. Want to start on that now?"

## Tool Sequence

```
get_project_context(project) OR get_brain_overview()
  → collaborative planning conversation
  → create_project() (if new)
  → create_todo_in_project(project, [tasks with metadata])
  → create_project_note(project, "Plan: [project]", planning summary)
  → create_project_note(project, "Design: [topic]", implementation details) (if applicable)
  → create_project_note(project, "Requirements: [topic]", feature context) (if applicable)
  → set_context(project_name: project)
```

## Planning Principles

- **Concrete over abstract**: "Improve performance" is not a task. "Profile the API endpoint and identify top 3 bottlenecks" is.
- **Small over large**: A task should be completable in one sitting. If it feels big, break it down.
- **Ordered over scattered**: Tasks should have a natural sequence. What unlocks what?
- **Complete enough, not perfect**: A plan with 5 good tasks beats 20 over-specified ones. You can always add more later.
- **Backlog by default**: New tasks go to backlog. Only promote to open what the user commits to working on soon. This keeps the active list focused and manageable.

## Notes

- Do not over-plan. 5-10 tasks is ideal for most personal projects. More than 15 and the plan itself becomes overhead.
- Push for a "definition of done" — planning without a clear finish line is just brainstorming
- If the user says "I don't know where to start", suggest a research/discovery task as the first step: "Research options for X" or "Prototype a basic version of Y"
- Respect the user's domain expertise — suggest structure, not solutions. They know their work better than you do.
- Tone: collaborative and structured — like a good project kickoff meeting
