package utils

import (
	"log"
	"time"
)

// Common RSS date formats
var rssDateFormats = []string{
	time.RFC1123Z,    // "Mon, 02 Jan 2006 15:04:05 -0700"
	time.RFC1123,     // "Mon, 02 Jan 2006 15:04:05 MST"
	time.RFC822Z,     // "02 Jan 06 15:04 -0700"
	time.RFC822,      // "02 Jan 06 15:04 MST"
	time.RFC3339,     // "2006-01-02T15:04:05Z07:00"
	time.RFC3339Nano, // "2006-01-02T15:04:05.999999999Z07:00"
	"2006-01-02T15:04:05-0700",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05 -0700",
	"2006-01-02 15:04:05",
	"Mon, 2 Jan 2006 15:04:05 -0700",
	"Mon, 2 Jan 2006 15:04:05 MST",
	"2 Jan 2006 15:04:05 -0700",
	"2 Jan 2006 15:04:05",
	"2006-01-02T15:04:05.000Z",
	"2006-01-02T15:04:05Z",
	"Mon, 02 Jan 2006 15:04:05 GMT",
	"02 Jan 2006 15:04:05 GMT",
}

// ParseRSSDate attempts to parse a date string using various RSS date formats
func ParseRSSDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Now(), nil
	}

	// Try each format until one works
	for _, format := range rssDateFormats {
		if parsedTime, err := time.Parse(format, dateStr); err == nil {
			// Convert to UTC for consistency
			return parsedTime.UTC(), nil
		}
	}

	// If no format works, log the issue and return current time
	log.Printf("Warning: Could not parse date '%s', using current time", dateStr)
	return time.Now().UTC(), nil
}

// FormatDateForDisplay formats a time for display purposes
func FormatDateForDisplay(t time.Time) string {
	// Convert to local timezone for display
	local := t.Local()

	// Check if it's today
	now := time.Now()
	if isSameDay(local, now) {
		return "Today"
	}

	// Check if it's yesterday
	yesterday := now.AddDate(0, 0, -1)
	if isSameDay(local, yesterday) {
		return "Yesterday"
	}

	// Check if it's within the last week
	weekAgo := now.AddDate(0, 0, -7)
	if local.After(weekAgo) {
		return local.Format("Monday")
	}

	// Check if it's this year
	if local.Year() == now.Year() {
		return local.Format("January 2")
	}

	// Otherwise show full date
	return local.Format("January 2, 2006")
}

// FormatDateForGrouping formats a time for grouping purposes (consistent UTC)
func FormatDateForGrouping(t time.Time) string {
	// Use local time for grouping so items appear under the right day
	local := t.Local()
	return local.Format("2006-01-02")
}

// isSameDay checks if two times are on the same day
func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// NormalizeToUTC converts any time to UTC for consistent storage
func NormalizeToUTC(t time.Time) time.Time {
	return t.UTC()
}
