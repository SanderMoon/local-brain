package validation

import "fmt"

// MCPError represents a structured error response for MCP tools
type MCPError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e *MCPError) Error() string {
	return e.Message
}

// NewValidationError creates a validation error
func NewValidationError(field, reason string) error {
	return &MCPError{
		Code:    "VALIDATION_ERROR",
		Message: fmt.Sprintf("validation failed for %s: %s", field, reason),
		Details: map[string]interface{}{
			"field":  field,
			"reason": reason,
		},
	}
}

// NewNotFoundError creates a not found error with available options
func NewNotFoundError(resource, name string, available []string) error {
	return &MCPError{
		Code:    fmt.Sprintf("%s_NOT_FOUND", resource),
		Message: fmt.Sprintf("%s '%s' not found", resource, name),
		Details: map[string]interface{}{
			"requested":  name,
			"available":  available,
			"resource":   resource,
		},
	}
}

// NewItemNotFoundError creates an error for missing todo/dump items
func NewItemNotFoundError(itemType, id string) error {
	return &MCPError{
		Code:    "ITEM_NOT_FOUND",
		Message: fmt.Sprintf("%s with ID '%s' not found", itemType, id),
		Details: map[string]interface{}{
			"item_type": itemType,
			"id":        id,
		},
	}
}

// NewProjectNotFoundError creates an error for missing projects with suggestions
func NewProjectNotFoundError(name string, availableProjects []string) error {
	return NewNotFoundError("PROJECT", name, availableProjects)
}

// NewBrainNotFoundError creates an error for missing brains
func NewBrainNotFoundError(name string, availableBrains []string) error {
	return NewNotFoundError("BRAIN", name, availableBrains)
}
