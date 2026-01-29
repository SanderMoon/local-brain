package dateutil

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParseNaturalDate converts natural language date input to ISO format (YYYY-MM-DD)
// Supports:
//   - ISO dates: 2026-02-15
//   - Today/tomorrow/yesterday
//   - Relative dates: +3d, -2w, +1m, +1y
//   - Day names: monday, tuesday, next-friday, this-saturday
func ParseNaturalDate(input string) (string, error) {
	input = strings.ToLower(strings.TrimSpace(input))
	if input == "" {
		return "", fmt.Errorf("empty date input")
	}

	now := time.Now()

	// Try ISO date first
	if matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, input); matched {
		// Validate it's a real date
		_, err := time.Parse("2006-01-02", input)
		if err != nil {
			return "", fmt.Errorf("invalid date format: %s", input)
		}
		return input, nil
	}

	// Handle keywords
	switch input {
	case "today":
		return now.Format("2006-01-02"), nil
	case "tomorrow":
		return now.AddDate(0, 0, 1).Format("2006-01-02"), nil
	case "yesterday":
		return now.AddDate(0, 0, -1).Format("2006-01-02"), nil
	}

	// Handle relative dates: +3d, -2w, +1m, +1y
	if strings.HasPrefix(input, "+") || strings.HasPrefix(input, "-") {
		return parseRelativeDate(input, now)
	}

	// Handle day names: monday, tuesday, next-friday, this-saturday
	if date, err := parseDayName(input, now); err == nil {
		return date, nil
	}

	return "", fmt.Errorf("unrecognized date format: %s", input)
}

// parseRelativeDate parses relative date strings like +3d, -2w, +1m, +1y
func parseRelativeDate(input string, now time.Time) (string, error) {
	pattern := regexp.MustCompile(`^([+-])(\d+)([dwmy])$`)
	matches := pattern.FindStringSubmatch(input)

	if matches == nil {
		return "", fmt.Errorf("invalid relative date format: %s", input)
	}

	sign := matches[1]
	amountStr := matches[2]
	unit := matches[3]

	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		return "", fmt.Errorf("invalid amount: %s", amountStr)
	}

	if sign == "-" {
		amount = -amount
	}

	var result time.Time
	switch unit {
	case "d":
		result = now.AddDate(0, 0, amount)
	case "w":
		result = now.AddDate(0, 0, amount*7)
	case "m":
		result = now.AddDate(0, amount, 0)
	case "y":
		result = now.AddDate(amount, 0, 0)
	}

	return result.Format("2006-01-02"), nil
}

// parseDayName parses day names like "monday", "next-friday", "this-saturday"
func parseDayName(input string, now time.Time) (string, error) {
	dayNames := map[string]time.Weekday{
		"sunday":    time.Sunday,
		"monday":    time.Monday,
		"tuesday":   time.Tuesday,
		"wednesday": time.Wednesday,
		"thursday":  time.Thursday,
		"friday":    time.Friday,
		"saturday":  time.Saturday,
	}

	// Parse "next-friday" or "this-saturday"
	var prefix string
	var dayName string

	if strings.HasPrefix(input, "next-") {
		prefix = "next"
		dayName = strings.TrimPrefix(input, "next-")
	} else if strings.HasPrefix(input, "this-") {
		prefix = "this"
		dayName = strings.TrimPrefix(input, "this-")
	} else {
		dayName = input
	}

	targetWeekday, ok := dayNames[dayName]
	if !ok {
		return "", fmt.Errorf("unrecognized day name: %s", input)
	}

	currentWeekday := now.Weekday()
	daysUntil := int(targetWeekday - currentWeekday)

	// Default behavior (no prefix): next occurrence of the day
	if prefix == "" || prefix == "next" {
		if daysUntil <= 0 {
			daysUntil += 7
		}
		if prefix == "next" && daysUntil < 7 {
			// "next-friday" means the Friday of next week, not this week
			daysUntil += 7
		}
	} else if prefix == "this" {
		// "this-friday" means the Friday of this week
		if daysUntil < 0 {
			daysUntil += 7
		}
	}

	result := now.AddDate(0, 0, daysUntil)
	return result.Format("2006-01-02"), nil
}
