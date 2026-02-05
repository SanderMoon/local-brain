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

// UnifiedSearch searches both todos and notes
func UnifiedSearch(activeDir string, query string, project string, includeTodos bool, includeNotes bool, searchNoteContent bool, createdAfter string, createdBefore string, completedAfter string, completedBefore string) ([]UnifiedSearchResult, error) {
	var results []UnifiedSearchResult

	// Default to both if neither specified
	if !includeTodos && !includeNotes {
		includeTodos = true
		includeNotes = true
	}

	// Search todos
	if includeTodos {
		allTodos, err := ParseAllTodos(activeDir, true)
		if err == nil {
			// Apply filters
			filteredTodos := SearchTodos(allTodos, query, project, "", nil)

			// Apply temporal filters
			if createdAfter != "" || createdBefore != "" || completedAfter != "" || completedBefore != "" {
				filteredTodos = filterTodosByTemporal(filteredTodos, createdAfter, createdBefore, completedAfter, completedBefore)
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
	if includeNotes {
		noteResults, err := SearchNotes(activeDir, query, project, searchNoteContent)
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
func filterTodosByTemporal(todos []TodoItem, createdAfter string, createdBefore string, completedAfter string, completedBefore string) []TodoItem {
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

		if include {
			filtered = append(filtered, todo)
		}
	}

	return filtered
}

// FilterTodosByTemporal is a public helper for filtering todos by dates
// Used by MCP tools and CLI
func FilterTodosByTemporal(todos []TodoItem, createdAfter string, createdBefore string, completedAfter string, completedBefore string) []TodoItem {
	return filterTodosByTemporal(todos, createdAfter, createdBefore, completedAfter, completedBefore)
}
