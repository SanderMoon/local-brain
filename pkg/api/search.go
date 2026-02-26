package api

// SearchResultType distinguishes between todo and note results
type SearchResultType string

const (
	ResultTypeTodo SearchResultType = "todo"
	ResultTypeNote SearchResultType = "note"
)

// UnifiedSearchResult wraps either a todo or note
type UnifiedSearchResult struct {
	Type SearchResultType  `json:"type"`
	Todo *TodoItem         `json:"todo,omitempty"`
	Note *NoteSearchResult `json:"note,omitempty"`
}

// SearchOptions contains all parameters for UnifiedSearch
type SearchOptions struct {
	Query             string
	Project           string
	IncludeTodos      bool
	IncludeNotes      bool
	SearchNoteContent bool
	// Todo-specific filters
	Status           string
	Priority         *int
	Tags             []string
	IncludeCompleted bool
	// Temporal filters
	CreatedAfter    string
	CreatedBefore   string
	CompletedAfter  string
	CompletedBefore string
	DueAfter        string
	DueBefore       string
}

// UnifiedSearch searches both todos and notes with optional filtering
func UnifiedSearch(activeDir string, opts SearchOptions) ([]UnifiedSearchResult, error) {
	var results []UnifiedSearchResult

	// Default to both if neither specified
	if !opts.IncludeTodos && !opts.IncludeNotes {
		opts.IncludeTodos = true
		opts.IncludeNotes = true
	}

	// Search todos
	if opts.IncludeTodos {
		// Include completed if explicitly requested or if filtering by status/query/tags
		includeCompleted := opts.IncludeCompleted || opts.Status != "" || opts.Query != "" || len(opts.Tags) > 0
		allTodos, err := ParseAllTodos(activeDir, includeCompleted)
		if err == nil {
			// Apply content/status/tag filters
			filteredTodos := SearchTodos(allTodos, opts.Query, opts.Project, opts.Status, opts.Tags)

			// Apply temporal filters
			if opts.CreatedAfter != "" || opts.CreatedBefore != "" || opts.CompletedAfter != "" || opts.CompletedBefore != "" || opts.DueAfter != "" || opts.DueBefore != "" {
				filteredTodos = filterTodosByTemporal(filteredTodos, opts.CreatedAfter, opts.CreatedBefore, opts.CompletedAfter, opts.CompletedBefore, opts.DueAfter, opts.DueBefore)
			}

			// Apply priority filter
			if opts.Priority != nil {
				var priorityFiltered []TodoItem
				for _, todo := range filteredTodos {
					if todo.Priority != nil && *todo.Priority == *opts.Priority {
						priorityFiltered = append(priorityFiltered, todo)
					}
				}
				filteredTodos = priorityFiltered
			}

			for i := range filteredTodos {
				results = append(results, UnifiedSearchResult{
					Type: ResultTypeTodo,
					Todo: &filteredTodos[i],
				})
			}
		}
	}

	// Search notes
	if opts.IncludeNotes {
		noteResults, err := SearchNotes(activeDir, opts.Query, opts.Project, opts.SearchNoteContent)
		if err == nil {
			for i := range noteResults {
				results = append(results, UnifiedSearchResult{
					Type: ResultTypeNote,
					Note: &noteResults[i],
				})
			}
		}
	}

	return results, nil
}

// filterTodosByTemporal applies date range filters
func filterTodosByTemporal(todos []TodoItem, createdAfter string, createdBefore string, completedAfter string, completedBefore string, dueAfter string, dueBefore string) []TodoItem {
	var filtered []TodoItem

	for _, todo := range todos {
		include := true

		// Filter by captured date - exclude if no date when filtering by dates
		if createdAfter != "" {
			if todo.CapturedDate == "" || todo.CapturedDate < createdAfter {
				include = false
			}
		}
		if createdBefore != "" && include {
			if todo.CapturedDate == "" || todo.CapturedDate > createdBefore {
				include = false
			}
		}

		// Filter by completed date - exclude if no date when filtering by dates
		if completedAfter != "" && include {
			if todo.CompletedDate == "" || todo.CompletedDate < completedAfter {
				include = false
			}
		}
		if completedBefore != "" && include {
			if todo.CompletedDate == "" || todo.CompletedDate > completedBefore {
				include = false
			}
		}

		// Filter by due date - exclude if no date when filtering by dates
		if dueAfter != "" && include {
			if todo.DueDate == "" || todo.DueDate < dueAfter {
				include = false
			}
		}
		if dueBefore != "" && include {
			if todo.DueDate == "" || todo.DueDate > dueBefore {
				include = false
			}
		}

		if include {
			filtered = append(filtered, todo)
		}
	}

	return filtered
}

// FilterTodosByTemporal is a public helper for filtering todos by dates
// Used by MCP tools and CLI
func FilterTodosByTemporal(todos []TodoItem, createdAfter string, createdBefore string, completedAfter string, completedBefore string, dueAfter string, dueBefore string) []TodoItem {
	return filterTodosByTemporal(todos, createdAfter, createdBefore, completedAfter, completedBefore, dueAfter, dueBefore)
}
