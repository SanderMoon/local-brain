package dateutil

import (
	"strings"
	"testing"
	"time"
)

func TestParseNaturalDate_ISO(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"2026-02-15", "2026-02-15", false},
		{"2024-01-01", "2024-01-01", false},
		{"2025-12-31", "2025-12-31", false},
		{"2026-13-01", "", true},  // Invalid month
		{"2026-02-30", "", true},  // Invalid day
		{"not-a-date", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseNaturalDate(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error for input %s, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %s: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, result)
				}
			}
		})
	}
}

func TestParseNaturalDate_Keywords(t *testing.T) {
	now := time.Now()
	today := now.Format("2006-01-02")
	tomorrow := now.AddDate(0, 0, 1).Format("2006-01-02")
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")

	tests := []struct {
		input    string
		expected string
	}{
		{"today", today},
		{"tomorrow", tomorrow},
		{"yesterday", yesterday},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseNaturalDate(tt.input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestParseNaturalDate_Relative(t *testing.T) {
	now := time.Now()

	tests := []struct {
		input    string
		expected time.Time
	}{
		{"+1d", now.AddDate(0, 0, 1)},
		{"+3d", now.AddDate(0, 0, 3)},
		{"-2d", now.AddDate(0, 0, -2)},
		{"+1w", now.AddDate(0, 0, 7)},
		{"+2w", now.AddDate(0, 0, 14)},
		{"-1w", now.AddDate(0, 0, -7)},
		{"+1m", now.AddDate(0, 1, 0)},
		{"+3m", now.AddDate(0, 3, 0)},
		{"-1m", now.AddDate(0, -1, 0)},
		{"+1y", now.AddDate(1, 0, 0)},
		{"-1y", now.AddDate(-1, 0, 0)},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseNaturalDate(tt.input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			expected := tt.expected.Format("2006-01-02")
			if result != expected {
				t.Errorf("Expected %s, got %s", expected, result)
			}
		})
	}
}

func TestParseNaturalDate_DayNames(t *testing.T) {
	now := time.Now()

	// Test each day name
	dayNames := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}

	for _, dayName := range dayNames {
		t.Run(dayName, func(t *testing.T) {
			result, err := ParseNaturalDate(dayName)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			// Result should be a valid date
			parsed, err := time.Parse("2006-01-02", result)
			if err != nil {
				t.Errorf("Result is not a valid date: %s", result)
			}
			// Should be in the future (or today if it's that day)
			if parsed.Before(now.Truncate(24 * time.Hour)) {
				t.Errorf("Day name %s resolved to past date: %s", dayName, result)
			}
			// Should be within next 7 days
			maxDate := now.AddDate(0, 0, 7)
			if parsed.After(maxDate) {
				t.Errorf("Day name %s resolved to date too far in future: %s", dayName, result)
			}
		})
	}
}

func TestParseNaturalDate_NextDay(t *testing.T) {
	now := time.Now()

	// Test "next-" prefix
	tests := []string{"next-monday", "next-friday", "next-saturday"}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			result, err := ParseNaturalDate(input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			// Result should be a valid date
			parsed, err := time.Parse("2006-01-02", result)
			if err != nil {
				t.Errorf("Result is not a valid date: %s", result)
			}
			// Should be at least 7 days in the future
			minDate := now.AddDate(0, 0, 7)
			if parsed.Before(minDate.Truncate(24 * time.Hour)) {
				t.Errorf("next-%s resolved to date too close: %s (should be at least 7 days away)", input, result)
			}
		})
	}
}

func TestParseNaturalDate_ThisDay(t *testing.T) {
	now := time.Now()

	// Test "this-" prefix
	tests := []string{"this-monday", "this-friday", "this-saturday"}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			result, err := ParseNaturalDate(input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			// Result should be a valid date
			parsed, err := time.Parse("2006-01-02", result)
			if err != nil {
				t.Errorf("Result is not a valid date: %s", result)
			}
			// Should be within this week (next 7 days)
			maxDate := now.AddDate(0, 0, 7)
			if parsed.After(maxDate) {
				t.Errorf("this-%s resolved to date beyond this week: %s", input, result)
			}
			// Should not be in the past
			if parsed.Before(now.Truncate(24 * time.Hour)) {
				t.Errorf("this-%s resolved to past date: %s", input, result)
			}
		})
	}
}

func TestParseNaturalDate_Invalid(t *testing.T) {
	tests := []string{
		"",
		"invalid",
		"123",
		"not a date",
		"+abc",
		"next-notaday",
		"this-notaday",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := ParseNaturalDate(input)
			if err == nil {
				t.Errorf("Expected error for invalid input: %s", input)
			}
		})
	}
}

func TestParseRelativeDate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		input       string
		shouldError bool
	}{
		{"+1d", false},
		{"-1d", false},
		{"+1w", false},
		{"+1m", false},
		{"+1y", false},
		{"+abc", true},
		{"1d", true}, // Missing sign
		{"+d", true}, // Missing amount
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseRelativeDate(tt.input, now)
			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error for input %s, got result: %s", tt.input, result)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %s: %v", tt.input, err)
				}
				// Verify result is a valid date
				_, parseErr := time.Parse("2006-01-02", result)
				if parseErr != nil {
					t.Errorf("Result is not a valid date: %s", result)
				}
			}
		})
	}
}

func TestParseDayName(t *testing.T) {
	now := time.Now()

	tests := []struct {
		input       string
		shouldError bool
	}{
		{"monday", false},
		{"tuesday", false},
		{"wednesday", false},
		{"thursday", false},
		{"friday", false},
		{"saturday", false},
		{"sunday", false},
		{"next-monday", false},
		{"this-friday", false},
		{"notaday", true},
		{"next-notaday", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseDayName(tt.input, now)
			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error for input %s, got result: %s", tt.input, result)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %s: %v", tt.input, err)
				}
				// Verify result is a valid date
				_, parseErr := time.Parse("2006-01-02", result)
				if parseErr != nil {
					t.Errorf("Result is not a valid date: %s", result)
				}
			}
		})
	}
}

func TestParseNaturalDate_CaseInsensitive(t *testing.T) {
	tests := []string{
		"TODAY",
		"Tomorrow",
		"YESTERDAY",
		"Monday",
		"FRIDAY",
		"Next-Friday",
		"THIS-SATURDAY",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := ParseNaturalDate(input)
			if err != nil {
				t.Errorf("Should handle case-insensitive input %s: %v", input, err)
			}
		})
	}
}

func TestParseNaturalDate_Whitespace(t *testing.T) {
	tests := map[string]string{
		"  today  ":   "today",
		"\ttomorrow\t": "tomorrow",
		" +3d ":       "+3d",
	}

	for input, normalized := range tests {
		t.Run(input, func(t *testing.T) {
			result1, err1 := ParseNaturalDate(input)
			result2, err2 := ParseNaturalDate(normalized)

			if err1 != nil || err2 != nil {
				if err1 != nil && err2 != nil {
					// Both failed, that's okay if they match
					if !strings.Contains(err1.Error(), err2.Error()) && !strings.Contains(err2.Error(), err1.Error()) {
						t.Errorf("Different errors for whitespace: %v vs %v", err1, err2)
					}
				} else {
					t.Errorf("Whitespace handling inconsistent: %v vs %v", err1, err2)
				}
			} else if result1 != result2 {
				t.Errorf("Whitespace affected result: %s vs %s", result1, result2)
			}
		})
	}
}
