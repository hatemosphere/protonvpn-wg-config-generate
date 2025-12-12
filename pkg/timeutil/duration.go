// Package timeutil provides time-related utility functions.
package timeutil

import (
	"fmt"
	"strings"
	"time"
)

// pluralize returns singular or plural form based on count
func pluralize(count int, singular, plural string) string {
	if count == 1 {
		return fmt.Sprintf("1 %s", singular)
	}
	return fmt.Sprintf("%d %s", count, plural)
}

// formatTimeComponents formats non-zero time components into a string
func formatTimeComponents(parts []string) string {
	return strings.Join(parts, " ")
}

// formatWeeksAndDays formats weeks and remaining days
func formatWeeksAndDays(days int) []string {
	var parts []string
	weeks := days / 7
	remainingDays := days % 7

	if weeks > 0 {
		parts = append(parts, pluralize(weeks, "week", "weeks"))
	}
	if remainingDays > 0 {
		parts = append(parts, pluralize(remainingDays, "day", "days"))
	}
	return parts
}

// formatHoursMinutes formats hours and minutes
func formatHoursMinutes(hours, minutes int) string {
	switch {
	case hours > 0 && minutes > 0:
		return fmt.Sprintf("%d hours %d minutes", hours, minutes)
	case hours > 0:
		return fmt.Sprintf("%d hours", hours)
	default:
		return fmt.Sprintf("%d minutes", minutes)
	}
}

// formatDaysHoursMinutes formats days, hours, and minutes into a string
func formatDaysHoursMinutes(days, hours, minutes int) string {
	// Simple days format (less than a week)
	if days < 7 {
		return formatSimpleDays(days, hours, minutes)
	}

	// Weeks format
	parts := formatWeeksAndDays(days)
	if hours > 0 || minutes > 0 {
		parts = append(parts, formatHoursMinutes(hours, minutes))
	}
	return formatTimeComponents(parts)
}

// formatSimpleDays formats days less than a week with hours and minutes
func formatSimpleDays(days, hours, minutes int) string {
	if hours == 0 && minutes == 0 {
		return pluralize(days, "day", "days")
	}
	if minutes == 0 {
		return fmt.Sprintf("%d days %d hours", days, hours)
	}
	return fmt.Sprintf("%d days %d hours %d minutes", days, hours, minutes)
}

// HumanizeDuration converts a duration to a human-readable format with high precision.
// It shows progressively less detail for longer durations:
// - Less than a minute: "less than a minute"
// - Less than an hour: "X minutes"
// - Less than a day: "X hours Y minutes"
// - Less than a week: "X days Y hours Z minutes"
// - Less than a month: "X weeks Y days Z hours W minutes"
// - Longer periods: months and years with appropriate detail
func HumanizeDuration(d time.Duration) string {
	if d < 0 {
		return "expired"
	}

	if d < time.Minute {
		return "less than a minute"
	}

	if d < time.Hour {
		return pluralize(int(d.Minutes()), "minute", "minutes")
	}

	if d < 24*time.Hour {
		return formatHoursDuration(d)
	}

	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days < 30 {
		return formatDaysHoursMinutes(days, hours, minutes)
	}

	if days < 365 {
		return formatMonthsDuration(days)
	}

	return formatYearsDuration(days)
}

// formatHoursDuration formats durations less than a day
func formatHoursDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	if minutes == 0 {
		return pluralize(hours, "hour", "hours")
	}
	return fmt.Sprintf("%d hours %d minutes", hours, minutes)
}

// formatMonthsDuration formats durations between 30 and 365 days
func formatMonthsDuration(days int) string {
	months := days / 30
	remainingDays := days % 30
	if remainingDays == 0 {
		return pluralize(months, "month", "months")
	}
	return fmt.Sprintf("%d months %d days", months, remainingDays)
}

// formatYearsDuration formats durations of 365 days or more
func formatYearsDuration(days int) string {
	years := days / 365
	remainingDays := days % 365
	if remainingDays == 0 {
		return pluralize(years, "year", "years")
	}
	if remainingDays < 30 {
		return fmt.Sprintf("%d years %d days", years, remainingDays)
	}
	months := remainingDays / 30
	return fmt.Sprintf("%d years %d months", years, months)
}
