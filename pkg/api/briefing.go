package api

import (
	"time"
)

// DailyBriefing provides a comprehensive overview for starting the day
type DailyBriefing struct {
	Summary struct {
		TotalOpen            int            `json:"total_open"`
		Overdue              int            `json:"overdue"`
		DueToday             int            `json:"due_today"`
		DueThisWeek          int            `json:"due_this_week"`
		HighPriority         int            `json:"high_priority"`
		InProgress           int            `json:"in_progress"`
		Blocked              int            `json:"blocked"`
		InboxCount           int            `json:"inbox_count"`
		CompletionsLast3Days int            `json:"completions_last_3_days"`
		ProjectsWithTasks    map[string]int `json:"projects_with_tasks"`
	} `json:"summary"`
	Urgent struct {
		Overdue     []TodoItem `json:"overdue"`
		DueToday    []TodoItem `json:"due_today"`
		DueThisWeek []TodoItem `json:"due_this_week"`
	} `json:"urgent"`
	HighPriority      []TodoItem      `json:"high_priority"`
	InProgress        []TodoItem      `json:"in_progress"`
	Blocked           []TodoItem      `json:"blocked"`
	Context           BriefingContext `json:"context"`
}

type BriefingContext struct {
	RecentCompletions []TodoItem      `json:"recent_completions"`
	InboxItems        []DumpItemJSON  `json:"inbox_items"`
}

// GetDailyBriefing generates a comprehensive daily briefing
func GetDailyBriefing(activeDir, dumpPath string) (*DailyBriefing, error) {
	briefing := &DailyBriefing{}
	briefing.Summary.ProjectsWithTasks = make(map[string]int)

	// Get today's date
	today := time.Now().Format("2006-01-02")
	weekFromNow := time.Now().AddDate(0, 0, 7).Format("2006-01-02")
	threeDaysAgo := time.Now().AddDate(0, 0, -3).Format("2006-01-02")

	// Parse all todos (including completed for recent completions)
	allTodos, err := ParseAllTodos(activeDir, true)
	if err != nil {
		return nil, err
	}

	// Parse dump items
	dumpItems, err := ParseDumpToJSON(dumpPath)
	if err != nil {
		// Don't fail if dump doesn't exist
		dumpItems = []DumpItemJSON{}
	}

	// Categorize todos
	var openTodos []TodoItem
	var completedRecently []TodoItem

	for _, todo := range allTodos {
		if todo.Status == "done" {
			// Check if completed in last 3 days
			if todo.CompletedDate >= threeDaysAgo {
				completedRecently = append(completedRecently, todo)
			}
			continue
		}

		// Count open todos
		openTodos = append(openTodos, todo)
		briefing.Summary.ProjectsWithTasks[todo.Project]++

		// Categorize by urgency (date-based)
		if todo.DueDate != "" {
			if todo.DueDate < today {
				briefing.Urgent.Overdue = append(briefing.Urgent.Overdue, todo)
			} else if todo.DueDate == today {
				briefing.Urgent.DueToday = append(briefing.Urgent.DueToday, todo)
			} else if todo.DueDate <= weekFromNow {
				briefing.Urgent.DueThisWeek = append(briefing.Urgent.DueThisWeek, todo)
			}
		}

		// High priority (p:1) - ALL of them, regardless of due date
		if todo.Priority != nil && *todo.Priority == 1 {
			briefing.HighPriority = append(briefing.HighPriority, todo)
		}

		// Status-based categorization
		if todo.Status == "in-progress" {
			briefing.InProgress = append(briefing.InProgress, todo)
		} else if todo.Status == "blocked" {
			briefing.Blocked = append(briefing.Blocked, todo)
		}
	}

	// Fill summary counts
	briefing.Summary.TotalOpen = len(openTodos)
	briefing.Summary.Overdue = len(briefing.Urgent.Overdue)
	briefing.Summary.DueToday = len(briefing.Urgent.DueToday)
	briefing.Summary.DueThisWeek = len(briefing.Urgent.DueThisWeek)
	briefing.Summary.HighPriority = len(briefing.HighPriority)
	briefing.Summary.InProgress = len(briefing.InProgress)
	briefing.Summary.Blocked = len(briefing.Blocked)
	briefing.Summary.InboxCount = len(dumpItems)
	briefing.Summary.CompletionsLast3Days = len(completedRecently)

	// Fill context
	briefing.Context.RecentCompletions = completedRecently
	briefing.Context.InboxItems = dumpItems

	return briefing, nil
}
