// Package timeutil provides time-related utility functions.
package timeutil

import (
	"fmt"
	"time"
)

// formatDaysHoursMinutes formats days, hours, and minutes into a string
//
//nolint:gocyclo // Formatting function with many output cases
func formatDaysHoursMinutes(days, hours, minutes int) string {
	if days < 7 {
		if days == 1 && hours == 0 && minutes == 0 {
			return "1 day"
		}
		if hours == 0 && minutes == 0 {
			return fmt.Sprintf("%d days", days)
		}
		if minutes == 0 {
			return fmt.Sprintf("%d days %d hours", days, hours)
		}
		return fmt.Sprintf("%d days %d hours %d minutes", days, hours, minutes)
	}

	// Weeks
	weeks := days / 7
	remainingDays := days % 7

	result := ""
	if weeks > 0 {
		if weeks == 1 {
			result = "1 week"
		} else {
			result = fmt.Sprintf("%d weeks", weeks)
		}
	}
	if remainingDays > 0 {
		if result != "" {
			result += " "
		}
		if remainingDays == 1 {
			result += "1 day"
		} else {
			result += fmt.Sprintf("%d days", remainingDays)
		}
	}
	if hours > 0 || minutes > 0 {
		if result != "" {
			result += " "
		}
		switch {
		case hours > 0 && minutes > 0:
			result += fmt.Sprintf("%d hours %d minutes", hours, minutes)
		case hours > 0:
			result += fmt.Sprintf("%d hours", hours)
		default:
			result += fmt.Sprintf("%d minutes", minutes)
		}
	}
	return result
}

// HumanizeDuration converts a duration to a human-readable format with high precision.
// It shows progressively less detail for longer durations:
// - Less than a minute: "less than a minute"
// - Less than an hour: "X minutes"
// - Less than a day: "X hours Y minutes"
// - Less than a week: "X days Y hours Z minutes"
// - Less than a month: "X weeks Y days Z hours W minutes"
// - Longer periods: months and years with appropriate detail
//
//nolint:gocyclo // Duration formatting requires many time-range cases
func HumanizeDuration(d time.Duration) string {
	if d < 0 {
		return "expired"
	}

	// Less than a minute
	if d < time.Minute {
		return "less than a minute"
	}

	// Less than an hour
	if d < time.Hour {
		minutes := int(d.Minutes())
		if minutes == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", minutes)
	}

	// Less than a day
	if d < 24*time.Hour {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		if hours == 1 && minutes == 0 {
			return "1 hour"
		}
		if minutes == 0 {
			return fmt.Sprintf("%d hours", hours)
		}
		return fmt.Sprintf("%d hours %d minutes", hours, minutes)
	}

	// Days
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	// Less than a month
	if days < 30 {
		return formatDaysHoursMinutes(days, hours, minutes)
	}

	// Months to years
	if days < 365 {
		months := days / 30
		remainingDays := days % 30
		if months == 1 && remainingDays == 0 {
			return "1 month"
		}
		if remainingDays == 0 {
			return fmt.Sprintf("%d months", months)
		}
		return fmt.Sprintf("%d months %d days", months, remainingDays)
	}

	// Years
	years := days / 365
	remainingDays := days % 365
	if years == 1 && remainingDays == 0 {
		return "1 year"
	}
	if remainingDays == 0 {
		return fmt.Sprintf("%d years", years)
	}
	if remainingDays < 30 {
		return fmt.Sprintf("%d years %d days", years, remainingDays)
	}
	months := remainingDays / 30
	return fmt.Sprintf("%d years %d months", years, months)
}
