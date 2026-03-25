package api

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sandermoonemans/local-brain/pkg/testutil"
)

// makeTodoMD wraps lines inside a standard todo.md with ## Active / ## Completed sections.
func makeTodoMD(activeLines ...string) string {
	content := "# Tasks\n\n## Active\n\n"
	for _, l := range activeLines {
		content += l + "\n"
	}
	content += "\n## Completed\n\n"
	return content
}

// makeFullTodoMD creates a todo.md with both active and completed lines.
func makeFullTodoMD(activeLines []string, completedLines []string) string {
	content := "# Tasks\n\n## Active\n\n"
	for _, l := range activeLines {
		content += l + "\n"
	}
	content += "\n## Completed\n\n"
	for _, l := range completedLines {
		content += l + "\n"
	}
	return content
}

func TestGetDailyBriefing_EmptyBrain(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	briefing, err := GetDailyBriefing(tb.ActiveDirPath, tb.DumpPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if briefing.Summary.TotalOpen != 0 {
		t.Errorf("expected 0 total open, got %d", briefing.Summary.TotalOpen)
	}
	if briefing.Summary.Overdue != 0 {
		t.Errorf("expected 0 overdue, got %d", briefing.Summary.Overdue)
	}
	if len(briefing.Urgent.Overdue) != 0 {
		t.Errorf("expected empty overdue list, got %d items", len(briefing.Urgent.Overdue))
	}
	if len(briefing.HighPriority) != 0 {
		t.Errorf("expected empty high priority list, got %d items", len(briefing.HighPriority))
	}
	if len(briefing.Upcoming) != 0 {
		t.Errorf("expected empty upcoming list, got %d items", len(briefing.Upcoming))
	}
	if len(briefing.Summary.ProjectSummaries) != 0 {
		t.Errorf("expected empty project summaries, got %d", len(briefing.Summary.ProjectSummaries))
	}
	if len(briefing.Context.RecentNotes) != 0 {
		t.Errorf("expected empty recent notes, got %d", len(briefing.Context.RecentNotes))
	}
}

func TestGetDailyBriefing_Categorization(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	now := time.Now()

	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
	today := now.Format("2006-01-02")
	inThreeDays := now.AddDate(0, 0, 3).Format("2006-01-02")
	inTenDays := now.AddDate(0, 0, 10).Format("2006-01-02")
	threeDaysAgo := now.AddDate(0, 0, -3).Format("2006-01-02")

	projectDir := tb.AddProject("alpha")

	activeLines := []string{
		fmt.Sprintf("- [ ] Overdue task due:%s", yesterday),
		fmt.Sprintf("- [ ] Due today task due:%s", today),
		fmt.Sprintf("- [ ] Due this week due:%s", inThreeDays),
		fmt.Sprintf("- [ ] Upcoming task due:%s", inTenDays),
		"- [ ] High priority task p:1",
		"- [>] In progress task",
		"- [-] Blocked task",
		"- [~] Backlog task",
	}
	completedLines := []string{
		fmt.Sprintf("- [x] Completed task <!-- done:%s -->", threeDaysAgo),
	}

	tb.WriteFile(filepath.Join(projectDir, "todo.md"), makeFullTodoMD(activeLines, completedLines))

	briefing, err := GetDailyBriefing(tb.ActiveDirPath, tb.DumpPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := len(briefing.Urgent.Overdue); got != 1 {
		t.Errorf("expected 1 overdue, got %d", got)
	}
	if got := len(briefing.Urgent.DueToday); got != 1 {
		t.Errorf("expected 1 due_today, got %d", got)
	}
	if got := len(briefing.Urgent.DueThisWeek); got != 1 {
		t.Errorf("expected 1 due_this_week, got %d", got)
	}
	if got := len(briefing.Upcoming); got != 1 {
		t.Errorf("expected 1 upcoming, got %d", got)
	}
	if got := len(briefing.HighPriority); got != 1 {
		t.Errorf("expected 1 high priority, got %d", got)
	}
	if got := len(briefing.InProgress); got != 1 {
		t.Errorf("expected 1 in-progress, got %d", got)
	}
	if got := len(briefing.Blocked); got != 1 {
		t.Errorf("expected 1 blocked, got %d", got)
	}
	if got := briefing.Summary.CompletionsLast3Days; got != 1 {
		t.Errorf("expected 1 recent completion, got %d", got)
	}
	// 7 open (non-backlog) tasks
	if got := briefing.Summary.TotalOpen; got != 7 {
		t.Errorf("expected 7 total open, got %d", got)
	}
	// 1 backlog task
	if got := briefing.Summary.TotalBacklog; got != 1 {
		t.Errorf("expected 1 total backlog, got %d", got)
	}
	if got := len(briefing.Backlog); got != 1 {
		t.Errorf("expected 1 backlog item, got %d", got)
	}
}

func TestGetDailyBriefing_Deduplication(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	now := time.Now()

	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
	inThreeDays := now.AddDate(0, 0, 3).Format("2006-01-02")

	projectDir := tb.AddProject("dedup")

	tb.WriteFile(filepath.Join(projectDir, "todo.md"), makeTodoMD(
		fmt.Sprintf("- [ ] Overdue high priority p:1 due:%s", yesterday),
		"- [ ] Plain high priority p:1",
		fmt.Sprintf("- [ ] Week high priority p:1 due:%s", inThreeDays),
	))

	briefing, err := GetDailyBriefing(tb.ActiveDirPath, tb.DumpPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The overdue item should be in Overdue only
	if got := len(briefing.Urgent.Overdue); got != 1 {
		t.Fatalf("expected 1 overdue, got %d", got)
	}

	overdueID := briefing.Urgent.Overdue[0].ID

	// HighPriority should have 2 items: the plain p:1 and the this-week p:1
	// The overdue p:1 must NOT appear in HighPriority
	if got := len(briefing.HighPriority); got != 2 {
		t.Errorf("expected 2 high priority (plain + week), got %d", got)
	}
	for _, hp := range briefing.HighPriority {
		if hp.ID == overdueID {
			t.Error("overdue p:1 item must NOT appear in HighPriority")
		}
	}

	// The DueThisWeek p:1 item should appear in BOTH DueThisWeek and HighPriority
	if got := len(briefing.Urgent.DueThisWeek); got != 1 {
		t.Fatalf("expected 1 due_this_week, got %d", got)
	}
	weekID := briefing.Urgent.DueThisWeek[0].ID
	foundInHP := false
	for _, hp := range briefing.HighPriority {
		if hp.ID == weekID {
			foundInHP = true
			break
		}
	}
	if !foundInHP {
		t.Error("due_this_week p:1 item should also appear in HighPriority")
	}
}

func TestGetDailyBriefing_BacklogExcludedFromHighPriority(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	projectDir := tb.AddProject("backlog-test")
	tb.WriteFile(filepath.Join(projectDir, "todo.md"), makeTodoMD(
		"- [ ] Open high priority p:1",
		"- [~] Backlog high priority p:1",
		"- [~] Backlog normal task",
		"- [ ] Open normal task",
	))

	briefing, err := GetDailyBriefing(tb.ActiveDirPath, tb.DumpPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only the open p:1 item should be in HighPriority
	if got := len(briefing.HighPriority); got != 1 {
		t.Errorf("expected 1 high priority (open only), got %d", got)
	}
	if len(briefing.HighPriority) > 0 && briefing.HighPriority[0].Status != "open" {
		t.Errorf("expected high priority item to be open, got %q", briefing.HighPriority[0].Status)
	}

	// Backlog should contain both backlog items
	if got := len(briefing.Backlog); got != 2 {
		t.Errorf("expected 2 backlog items, got %d", got)
	}

	// TotalOpen should not include backlog
	if got := briefing.Summary.TotalOpen; got != 2 {
		t.Errorf("expected 2 total open (excluding backlog), got %d", got)
	}
	if got := briefing.Summary.TotalBacklog; got != 2 {
		t.Errorf("expected 2 total backlog, got %d", got)
	}

	// Project summary should show open/backlog split
	if got := len(briefing.Summary.ProjectSummaries); got != 1 {
		t.Fatalf("expected 1 project summary, got %d", got)
	}
	ps := briefing.Summary.ProjectSummaries[0]
	if ps.OpenCount != 2 {
		t.Errorf("expected open_count=2, got %d", ps.OpenCount)
	}
	if ps.BacklogCount != 2 {
		t.Errorf("expected backlog_count=2, got %d", ps.BacklogCount)
	}
	if ps.TaskCount != 4 {
		t.Errorf("expected task_count=4, got %d", ps.TaskCount)
	}
}

func TestGetDailyBriefing_Sorting(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	now := time.Now()

	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
	threeDaysAgo := now.AddDate(0, 0, -3).Format("2006-01-02")
	fiveDaysAgo := now.AddDate(0, 0, -5).Format("2006-01-02")
	today := now.Format("2006-01-02")

	projectDir := tb.AddProject("sorting")

	p1, p3 := 1, 3

	tb.WriteFile(filepath.Join(projectDir, "todo.md"), makeTodoMD(
		// Overdue items — added in unsorted order
		fmt.Sprintf("- [ ] Overdue recent due:%s", yesterday),
		fmt.Sprintf("- [ ] Overdue oldest due:%s", fiveDaysAgo),
		fmt.Sprintf("- [ ] Overdue middle due:%s", threeDaysAgo),
		// Due today items with different priorities
		fmt.Sprintf("- [ ] Today low p:%d due:%s", p3, today),
		fmt.Sprintf("- [ ] Today high p:%d due:%s", p1, today),
	))

	briefing, err := GetDailyBriefing(tb.ActiveDirPath, tb.DumpPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Overdue should be sorted ascending (oldest first)
	if got := len(briefing.Urgent.Overdue); got != 3 {
		t.Fatalf("expected 3 overdue items, got %d", got)
	}
	if briefing.Urgent.Overdue[0].DueDate != fiveDaysAgo {
		t.Errorf("first overdue should be oldest (%s), got %s", fiveDaysAgo, briefing.Urgent.Overdue[0].DueDate)
	}
	if briefing.Urgent.Overdue[1].DueDate != threeDaysAgo {
		t.Errorf("second overdue should be middle (%s), got %s", threeDaysAgo, briefing.Urgent.Overdue[1].DueDate)
	}
	if briefing.Urgent.Overdue[2].DueDate != yesterday {
		t.Errorf("third overdue should be most recent (%s), got %s", yesterday, briefing.Urgent.Overdue[2].DueDate)
	}

	// DueToday should be sorted by priority ascending (p:1 first)
	if got := len(briefing.Urgent.DueToday); got != 2 {
		t.Fatalf("expected 2 due_today items, got %d", got)
	}
	if briefing.Urgent.DueToday[0].Priority == nil || *briefing.Urgent.DueToday[0].Priority != 1 {
		t.Error("first due_today item should have priority 1")
	}
	if briefing.Urgent.DueToday[1].Priority == nil || *briefing.Urgent.DueToday[1].Priority != 3 {
		t.Error("second due_today item should have priority 3")
	}
}

func TestGetDailyBriefing_RecentNotes(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	now := time.Now()
	today := now.Format("2006-01-02")

	projectDir := tb.AddProject("notes-proj")
	notesDir := filepath.Join(projectDir, "notes")
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		t.Fatalf("failed to create notes dir: %v", err)
	}

	// Write a recent note (mtime = now by default)
	recentNotePath := filepath.Join(notesDir, "2026-01-01-recent.md")
	tb.WriteFile(recentNotePath, fmt.Sprintf("---\ntitle: Recent Note\ndate: %s\nproject: notes-proj\ntags: []\n---\n\n# Recent Note\n", today))

	// Write an old note and backdate its mtime to 10 days ago
	oldNotePath := filepath.Join(notesDir, "2026-01-01-old.md")
	tenDaysAgo := now.AddDate(0, 0, -10)
	tb.WriteFile(oldNotePath, fmt.Sprintf("---\ntitle: Old Note\ndate: %s\nproject: notes-proj\ntags: []\n---\n\n# Old Note\n", tenDaysAgo.Format("2006-01-02")))
	if err := os.Chtimes(oldNotePath, tenDaysAgo, tenDaysAgo); err != nil {
		t.Fatalf("failed to backdate old note: %v", err)
	}

	briefing, err := GetDailyBriefing(tb.ActiveDirPath, tb.DumpPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := len(briefing.Context.RecentNotes); got != 1 {
		t.Fatalf("expected 1 recent note, got %d", got)
	}
	if briefing.Context.RecentNotes[0].Title != "Recent Note" {
		t.Errorf("expected title 'Recent Note', got %q", briefing.Context.RecentNotes[0].Title)
	}
}

func TestGetDailyBriefing_RecentNotes_Cap(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	now := time.Now()
	today := now.Format("2006-01-02")

	projectDir := tb.AddProject("many-notes")
	notesDir := filepath.Join(projectDir, "notes")
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		t.Fatalf("failed to create notes dir: %v", err)
	}

	// Create 20 recent notes
	for i := 0; i < 20; i++ {
		notePath := filepath.Join(notesDir, fmt.Sprintf("2026-01-01-note-%02d.md", i))
		tb.WriteFile(notePath, fmt.Sprintf("---\ntitle: Note %d\ndate: %s\nproject: many-notes\ntags: []\n---\n\n# Note %d\n", i, today, i))
	}

	briefing, err := GetDailyBriefing(tb.ActiveDirPath, tb.DumpPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := len(briefing.Context.RecentNotes); got != 15 {
		t.Errorf("expected recent notes capped at 15, got %d", got)
	}
}

func TestGetDailyBriefing_ProjectSummaries(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	now := time.Now()
	inTwoDays := now.AddDate(0, 0, 2).Format("2006-01-02")

	// Project A: 3 tasks + description with ## Summary section
	projA := tb.AddProject("proj-a")
	tb.WriteFile(filepath.Join(projA, "todo.md"), makeTodoMD(
		fmt.Sprintf("- [ ] Task 1 due:%s", inTwoDays),
		"- [ ] Task 2",
		"- [ ] Task 3",
	))
	tb.WriteFile(filepath.Join(projA, "description.md"),
		"# Project A\n\n## Summary\nThis is the project summary.\nAnother summary line.\n\n## Details\nNot relevant.\n")

	// Project B: 1 task, no description
	projB := tb.AddProject("proj-b")
	tb.WriteFile(filepath.Join(projB, "todo.md"), makeTodoMD(
		"- [ ] Task B1",
	))

	briefing, err := GetDailyBriefing(tb.ActiveDirPath, tb.DumpPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := len(briefing.Summary.ProjectSummaries); got != 2 {
		t.Fatalf("expected 2 project summaries, got %d", got)
	}

	// Most active project (proj-a with 3 tasks) should be first
	if briefing.Summary.ProjectSummaries[0].Name != "proj-a" {
		t.Errorf("expected proj-a first (more tasks), got %s", briefing.Summary.ProjectSummaries[0].Name)
	}
	if briefing.Summary.ProjectSummaries[0].TaskCount != 3 {
		t.Errorf("expected proj-a to have 3 tasks, got %d", briefing.Summary.ProjectSummaries[0].TaskCount)
	}
	if briefing.Summary.ProjectSummaries[0].Summary != "This is the project summary.\nAnother summary line." {
		t.Errorf("unexpected summary for proj-a: %q", briefing.Summary.ProjectSummaries[0].Summary)
	}

	if briefing.Summary.ProjectSummaries[1].Name != "proj-b" {
		t.Errorf("expected proj-b second, got %s", briefing.Summary.ProjectSummaries[1].Name)
	}
	if briefing.Summary.ProjectSummaries[1].Summary != "" {
		t.Errorf("expected empty summary for proj-b (no description.md), got %q", briefing.Summary.ProjectSummaries[1].Summary)
	}
}

func TestGetDailyBriefing_Upcoming(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	now := time.Now()

	inThreeDays := now.AddDate(0, 0, 3).Format("2006-01-02")
	inTenDays := now.AddDate(0, 0, 10).Format("2006-01-02")
	inTwentyDays := now.AddDate(0, 0, 20).Format("2006-01-02")
	inThirtyOneDays := now.AddDate(0, 0, 31).Format("2006-01-02")

	projectDir := tb.AddProject("upcoming")

	tb.WriteFile(filepath.Join(projectDir, "todo.md"), makeTodoMD(
		fmt.Sprintf("- [ ] This week task due:%s", inThreeDays),
		fmt.Sprintf("- [ ] Upcoming ten days due:%s", inTenDays),
		fmt.Sprintf("- [ ] Upcoming twenty days due:%s", inTwentyDays),
		fmt.Sprintf("- [ ] Beyond month due:%s", inThirtyOneDays),
	))

	briefing, err := GetDailyBriefing(tb.ActiveDirPath, tb.DumpPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// DueThisWeek should only contain the 3-day task
	if got := len(briefing.Urgent.DueThisWeek); got != 1 {
		t.Errorf("expected 1 due_this_week, got %d", got)
	}

	// Upcoming should contain 10-day and 20-day tasks only
	if got := len(briefing.Upcoming); got != 2 {
		t.Fatalf("expected 2 upcoming, got %d", got)
	}
	// Sorted by due date ascending
	if briefing.Upcoming[0].DueDate != inTenDays {
		t.Errorf("first upcoming should be in 10 days, got %s", briefing.Upcoming[0].DueDate)
	}
	if briefing.Upcoming[1].DueDate != inTwentyDays {
		t.Errorf("second upcoming should be in 20 days, got %s", briefing.Upcoming[1].DueDate)
	}

	// Beyond-month task should not appear anywhere
	if got := briefing.Summary.DueNextMonth; got != 2 {
		t.Errorf("expected DueNextMonth=2, got %d", got)
	}
}

func TestExtractSummarySection(t *testing.T) {
	cases := []struct {
		name        string
		input       string
		wantSummary string
	}{
		{
			name:        "no summary section",
			input:       "# Title\n\nSome content here.\n",
			wantSummary: "",
		},
		{
			name:        "summary section only",
			input:       "# Title\n\n## Summary\nThis is the summary.\n",
			wantSummary: "This is the summary.",
		},
		{
			name:        "summary section with following section",
			input:       "# Title\n\n## Summary\nFirst line.\nSecond line.\n\n## Details\nNot this.\n",
			wantSummary: "First line.\nSecond line.",
		},
		{
			name:        "empty summary section",
			input:       "# Title\n\n## Summary\n\n## Details\nNot this.\n",
			wantSummary: "",
		},
		{
			name:        "empty input",
			input:       "",
			wantSummary: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractSummarySection(tc.input)
			if got != tc.wantSummary {
				t.Errorf("extractSummarySection(%q) = %q, want %q", tc.input, got, tc.wantSummary)
			}
		})
	}
}

func TestPriorityLess(t *testing.T) {
	p := func(n int) *int { return &n }

	cases := []struct {
		a, b *int
		want bool
	}{
		{p(1), p(2), true},  // 1 < 2
		{p(2), p(1), false}, // 2 > 1
		{p(1), p(1), false}, // equal
		{nil, p(1), false},  // nil sorts after non-nil
		{p(1), nil, true},   // non-nil sorts before nil
		{nil, nil, false},   // both nil = equal
	}

	for _, tc := range cases {
		got := priorityLess(tc.a, tc.b)
		if got != tc.want {
			aStr := "<nil>"
			bStr := "<nil>"
			if tc.a != nil {
				aStr = fmt.Sprintf("%d", *tc.a)
			}
			if tc.b != nil {
				bStr = fmt.Sprintf("%d", *tc.b)
			}
			t.Errorf("priorityLess(%s, %s) = %v, want %v", aStr, bStr, got, tc.want)
		}
	}
}
