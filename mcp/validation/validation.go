package validation

import (
	"fmt"
	"regexp"
	"time"

	"github.com/sandermoonemans/local-brain/pkg/api"
)

// ValidateTodoID checks if the ID is a 6-character hex string
func ValidateTodoID(id string) error {
	if !regexp.MustCompile(`^[a-f0-9]{6}$`).MatchString(id) {
		return fmt.Errorf("invalid todo ID: must be 6-character hex string (got: %s)", id)
	}
	return nil
}

// ValidateProjectName validates project name using the existing API validator
func ValidateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}
	return api.ValidateProjectName(name)
}

// ValidateDueDate checks if the date is in YYYY-MM-DD format
func ValidateDueDate(date string) error {
	if date == "" {
		return nil // Empty date is valid (clears due date)
	}
	_, err := time.Parse("2006-01-02", date)
	if err != nil {
		return fmt.Errorf("invalid due date: must be YYYY-MM-DD format (got: %s)", date)
	}
	return nil
}

// ValidatePriority checks if priority is in valid range (1-3) or nil
func ValidatePriority(p *int) error {
	if p == nil {
		return nil // nil is valid (clears priority)
	}
	if *p < 1 || *p > 3 {
		return fmt.Errorf("priority must be 1 (high), 2 (medium), or 3 (low), got: %d", *p)
	}
	return nil
}

// ValidateTodoStatus checks if status is one of the allowed values
func ValidateTodoStatus(status string) error {
	validStatuses := map[string]bool{
		"open":        true,
		"in-progress": true,
		"blocked":     true,
		"done":        true,
	}
	if !validStatuses[status] {
		return fmt.Errorf("invalid status: must be one of [open, in-progress, blocked, done], got: %s", status)
	}
	return nil
}

// ValidateBrainName checks if brain name is non-empty
func ValidateBrainName(name string) error {
	if name == "" {
		return fmt.Errorf("brain name cannot be empty")
	}
	return nil
}

// ValidateNonEmpty checks if a required string field is non-empty
func ValidateNonEmpty(fieldName, value string) error {
	if value == "" {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}
	return nil
}
