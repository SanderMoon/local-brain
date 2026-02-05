package validation

import (
	"testing"
)

func TestValidateTodoID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"valid 6-char hex", "abc123", false},
		{"valid all lowercase", "fedcba", false},
		{"valid all numbers", "123456", false},
		{"invalid uppercase", "ABC123", true},
		{"invalid too short", "abc12", true},
		{"invalid too long", "abc1234", true},
		{"invalid non-hex", "xyz123", true},
		{"invalid special chars", "abc-12", true},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTodoID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTodoID(%q) error = %v, wantErr %v", tt.id, err, tt.wantErr)
			}
		})
	}
}

func TestValidateDueDate(t *testing.T) {
	tests := []struct {
		name    string
		date    string
		wantErr bool
	}{
		{"valid date", "2026-02-04", false},
		{"empty date (clear)", "", false},
		{"invalid format", "02-04-2026", true},
		{"invalid format 2", "2026/02/04", true},
		{"invalid date", "2026-13-01", true},
		{"invalid day", "2026-02-30", true},
		{"not a date", "not-a-date", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDueDate(tt.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDueDate(%q) error = %v, wantErr %v", tt.date, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePriority(t *testing.T) {
	high := 1
	medium := 2
	low := 3
	invalid := 4
	zero := 0

	tests := []struct {
		name     string
		priority *int
		wantErr  bool
	}{
		{"nil (clear priority)", nil, false},
		{"valid high", &high, false},
		{"valid medium", &medium, false},
		{"valid low", &low, false},
		{"invalid too high", &invalid, true},
		{"invalid zero", &zero, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePriority(tt.priority)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePriority(%v) error = %v, wantErr %v", tt.priority, err, tt.wantErr)
			}
		})
	}
}

func TestValidateTodoStatus(t *testing.T) {
	tests := []struct {
		name    string
		status  string
		wantErr bool
	}{
		{"valid open", "open", false},
		{"valid in-progress", "in-progress", false},
		{"valid blocked", "blocked", false},
		{"valid done", "done", false},
		{"invalid status", "completed", true},
		{"invalid status 2", "todo", true},
		{"empty status", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTodoStatus(tt.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTodoStatus(%q) error = %v, wantErr %v", tt.status, err, tt.wantErr)
			}
		})
	}
}

func TestValidateBrainName(t *testing.T) {
	tests := []struct {
		name    string
		brain   string
		wantErr bool
	}{
		{"valid name", "work", false},
		{"valid with hyphen", "work-brain", false},
		{"empty name", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBrainName(tt.brain)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBrainName(%q) error = %v, wantErr %v", tt.brain, err, tt.wantErr)
			}
		})
	}
}

func TestValidateNonEmpty(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		value     string
		wantErr   bool
	}{
		{"valid value", "content", "something", false},
		{"empty value", "content", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNonEmpty(tt.fieldName, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNonEmpty(%q, %q) error = %v, wantErr %v", tt.fieldName, tt.value, err, tt.wantErr)
			}
		})
	}
}
