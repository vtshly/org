package parser

import (
	"fmt"
	"time"
)

// parseOrgDate parses org-mode date format
func parseOrgDate(dateStr string) (time.Time, error) {
	// Org-mode format: 2024-01-15 Mon 10:00
	formats := []string{
		"2006-01-02 Mon 15:04",
		"2006-01-02 Mon",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// parseClockTimestamp parses org-mode clock timestamp format
func parseClockTimestamp(timestampStr string) (time.Time, error) {
	// Org-mode clock format: [2024-01-15 Mon 10:00]
	formats := []string{
		"2006-01-02 Mon 15:04",
		"2006-01-02 Mon 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timestampStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse clock timestamp: %s", timestampStr)
}

// formatClockTimestamp formats a time as org-mode clock timestamp
func formatClockTimestamp(t time.Time) string {
	return t.Format("2006-01-02 Mon 15:04")
}

// FormatOrgDate formats a time as org-mode date
func FormatOrgDate(t time.Time) string {
	return t.Format("2006-01-02 Mon")
}
