package api

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ProjectBriefing summarises a single project for the daily briefing.
type ProjectBriefing struct {
	Name      string `json:"name"`
	TaskCount int    `json:"task_count"`
	Summary   string `json:"summary,omitempty"` // text from ## Summary section of description.md
}

// DailyBriefing provides a comprehensive overview for starting the day.
type DailyBriefing struct {
	Summary struct {
		TotalOpen            int               `json:"total_open"`
		Overdue              int               `json:"overdue"`
		DueToday             int               `json:"due_today"`
		DueThisWeek          int               `json:"due_this_week"`
		DueNextMonth         int               `json:"due_next_month"`
		HighPriority         int               `json:"high_priority"`
		InProgress           int               `json:"in_progress"`
		Blocked              int               `json:"blocked"`
		InboxCount           int               `json:"inbox_count"`
		CompletionsLast3Days int               `json:"completions_last_3_days"`
		ProjectSummaries     []ProjectBriefing `json:"project_summaries"`
	} `json:"summary"`
	Urgent struct {
		Overdue     []TodoItem `json:"overdue"`
		DueToday    []TodoItem `json:"due_today"`
		DueThisWeek []TodoItem `json:"due_this_week"`
	} `json:"urgent"`
	Upcoming     []TodoItem      `json:"upcoming"`
	HighPriority []TodoItem      `json:"high_priority"`
	InProgress   []TodoItem      `json:"in_progress"`
	Blocked      []TodoItem      `json:"blocked"`
	Context      BriefingContext `json:"context"`
}

// BriefingContext holds supporting context items for the daily briefing.
type BriefingContext struct {
	RecentCompletions []TodoItem     `json:"recent_completions"`
	InboxItems        []DumpItemJSON `json:"inbox_items"`
	RecentNotes       []NoteFile     `json:"recent_notes"`
}

// priorityLess returns true if priority a should sort before b.
// Lower number = higher priority (1 is highest). nil sorts last.
func priorityLess(a, b *int) bool {
	if a == nil && b == nil {
		return false
	}
	if a == nil {
		return false
	}
	if b == nil {
		return true
	}
	return *a < *b
}

// extractSummarySection returns the body of the "## Summary" section from a
// description.md content string. Returns empty string if the section is absent.
func extractSummarySection(description string) string {
	lines := strings.Split(description, "\n")
	inSummary := false
	var summaryLines []string

	for _, line := range lines {
		if strings.TrimSpace(line) == "## Summary" {
			inSummary = true
			continue
		}
		if inSummary {
			// Stop at the next heading
			if strings.HasPrefix(strings.TrimSpace(line), "## ") {
				break
			}
			summaryLines = append(summaryLines, line)
		}
	}

	return strings.TrimSpace(strings.Join(summaryLines, "\n"))
}

// GetDailyBriefing generates a comprehensive daily briefing.
func GetDailyBriefing(activeDir, dumpPath string) (*DailyBriefing, error) {
	briefing := &DailyBriefing{}

	// Time anchors
	now := time.Now()
	today := now.Format("2006-01-02")
	weekFromNow := now.AddDate(0, 0, 7).Format("2006-01-02")
	monthFromNow := now.AddDate(0, 0, 30).Format("2006-01-02")
	threeDaysAgo := now.AddDate(0, 0, -3).Format("2006-01-02")
	sevenDaysAgo := now.AddDate(0, 0, -7)

	// Parse all todos (including completed for recent completions)
	allTodos, err := ParseAllTodos(activeDir, true)
	if err != nil {
		return nil, err
	}

	// Parse dump items (non-fatal if absent)
	dumpItems, err := ParseDumpToJSON(dumpPath)
	if err != nil {
		dumpItems = []DumpItemJSON{}
	}

	// Single-pass categorisation
	var openTodos []TodoItem
	var completedRecently []TodoItem
	projectTaskCount := make(map[string]int)

	// urgentIDs tracks items already in overdue/due_today to deduplicate HighPriority
	urgentIDs := make(map[string]bool)

	for _, todo := range allTodos {
		if todo.Status == "done" {
			if todo.CompletedDate >= threeDaysAgo {
				completedRecently = append(completedRecently, todo)
			}
			continue
		}

		openTodos = append(openTodos, todo)
		projectTaskCount[todo.Project]++

		// Date-based urgency categorisation
		if todo.DueDate != "" {
			if todo.DueDate < today {
				briefing.Urgent.Overdue = append(briefing.Urgent.Overdue, todo)
				urgentIDs[todo.ID] = true
			} else if todo.DueDate == today {
				briefing.Urgent.DueToday = append(briefing.Urgent.DueToday, todo)
				urgentIDs[todo.ID] = true
			} else if todo.DueDate <= weekFromNow {
				briefing.Urgent.DueThisWeek = append(briefing.Urgent.DueThisWeek, todo)
			} else if todo.DueDate <= monthFromNow {
				briefing.Upcoming = append(briefing.Upcoming, todo)
				briefing.Summary.DueNextMonth++
			}
		}

		// High priority (p:1) — exclude items already in overdue or due_today
		if todo.Priority != nil && *todo.Priority == 1 && !urgentIDs[todo.ID] {
			briefing.HighPriority = append(briefing.HighPriority, todo)
		}

		// Status-based categorisation
		if todo.Status == "in-progress" {
			briefing.InProgress = append(briefing.InProgress, todo)
		} else if todo.Status == "blocked" {
			briefing.Blocked = append(briefing.Blocked, todo)
		}
	}

	// Sort all buckets
	sort.Slice(briefing.Urgent.Overdue, func(i, j int) bool {
		return briefing.Urgent.Overdue[i].DueDate < briefing.Urgent.Overdue[j].DueDate
	})

	sort.Slice(briefing.Urgent.DueToday, func(i, j int) bool {
		return priorityLess(briefing.Urgent.DueToday[i].Priority, briefing.Urgent.DueToday[j].Priority)
	})

	sort.Slice(briefing.Urgent.DueThisWeek, func(i, j int) bool {
		a, b := briefing.Urgent.DueThisWeek[i], briefing.Urgent.DueThisWeek[j]
		if a.DueDate != b.DueDate {
			return a.DueDate < b.DueDate
		}
		return priorityLess(a.Priority, b.Priority)
	})

	sort.Slice(briefing.Upcoming, func(i, j int) bool {
		a, b := briefing.Upcoming[i], briefing.Upcoming[j]
		if a.DueDate != b.DueDate {
			return a.DueDate < b.DueDate
		}
		return priorityLess(a.Priority, b.Priority)
	})

	sort.Slice(briefing.HighPriority, func(i, j int) bool {
		a, b := briefing.HighPriority[i], briefing.HighPriority[j]
		aHasDate := a.DueDate != ""
		bHasDate := b.DueDate != ""
		if aHasDate != bHasDate {
			return aHasDate
		}
		if aHasDate && a.DueDate != b.DueDate {
			return a.DueDate < b.DueDate
		}
		return priorityLess(a.Priority, b.Priority)
	})

	sort.Slice(briefing.InProgress, func(i, j int) bool {
		return priorityLess(briefing.InProgress[i].Priority, briefing.InProgress[j].Priority)
	})

	sort.Slice(briefing.Blocked, func(i, j int) bool {
		return priorityLess(briefing.Blocked[i].Priority, briefing.Blocked[j].Priority)
	})

	sort.Slice(completedRecently, func(i, j int) bool {
		return completedRecently[i].CompletedDate > completedRecently[j].CompletedDate
	})

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
	// DueNextMonth was incremented inline above

	// Build ProjectSummaries — only projects with open tasks
	entries, err := os.ReadDir(activeDir)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			name := entry.Name()
			taskCount := projectTaskCount[name]
			if taskCount == 0 {
				continue
			}
			projectDir := filepath.Join(activeDir, name)
			var summary string
			if desc, descErr := ReadProjectDescription(projectDir); descErr == nil {
				summary = extractSummarySection(desc)
			}
			briefing.Summary.ProjectSummaries = append(briefing.Summary.ProjectSummaries, ProjectBriefing{
				Name:      name,
				TaskCount: taskCount,
				Summary:   summary,
			})
		}
		// Sort by task count descending (most active project first)
		sort.Slice(briefing.Summary.ProjectSummaries, func(i, j int) bool {
			return briefing.Summary.ProjectSummaries[i].TaskCount > briefing.Summary.ProjectSummaries[j].TaskCount
		})
	}

	// Collect recent notes (modified in last 7 days), cap at 15
	var allRecentNotes []NoteFile
	if entries != nil {
		for _, entry := range entries {
			if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			projectDir := filepath.Join(activeDir, entry.Name())
			notes, err := ListNotes(projectDir)
			if err != nil {
				continue
			}
			for _, note := range notes {
				if note.ModTime.After(sevenDaysAgo) {
					allRecentNotes = append(allRecentNotes, note)
				}
			}
		}
		sort.Slice(allRecentNotes, func(i, j int) bool {
			return allRecentNotes[i].ModTime.After(allRecentNotes[j].ModTime)
		})
		if len(allRecentNotes) > 15 {
			allRecentNotes = allRecentNotes[:15]
		}
	}

	// Fill context
	briefing.Context.RecentCompletions = completedRecently
	briefing.Context.InboxItems = dumpItems
	briefing.Context.RecentNotes = allRecentNotes

	return briefing, nil
}
