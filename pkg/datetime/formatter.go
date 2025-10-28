package datetime

import (
	"log"
	"time"
)

type Formatter struct{}

func NewFormatter() *Formatter {
	return &Formatter{}
}

var rssDateFormats = []string{
	time.RFC1123Z,    // "Mon, 02 Jan 2006 15:04:05 -0700"
	time.RFC1123,     // "Mon, 02 Jan 2006 15:04:05 MST"
	time.RFC822Z,     // "02 Jan 06 15:04 -0700"
	time.RFC822,      // "02 Jan 06 15:04 MST"
	time.RFC3339,     // "2006-01-02T15:04:05Z07:00"
	time.RFC3339Nano, // "2006-01-02T15:04:05.999999999Z07:00"
	"2006-01-02T15:04:05-0700",
	"2006-01-02T15:04:05-07:00",
	"2006-01-02T15:04:05.000-07:00",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05 -0700",
	"2006-01-02 15:04:05 -07:00",
	"2006-01-02 15:04:05",
	"Mon, 2 Jan 2006 15:04:05 -0700",
	"Mon, 2 Jan 2006 15:04:05 -07:00",
	"Mon, 2 Jan 2006 15:04:05 MST",
	"Mon, 2 Jan 2006 15:04:05 GMT",
	"2 Jan 2006 15:04:05 -0700",
	"2 Jan 2006 15:04:05 -07:00",
	"2 Jan 2006 15:04:05",
	"2006-01-02T15:04:05.000Z",
	"2006-01-02T15:04:05.00Z",
	"2006-01-02T15:04:05.0Z",
	"2006-01-02T15:04:05Z",
	"Mon, 02 Jan 2006 15:04:05 GMT",
	"Mon, 02 Jan 2006 15:04:05 UTC",
	"02 Jan 2006 15:04:05 GMT",
	"02 Jan 2006 15:04:05 UTC",
	"January 2, 2006 15:04:05",
	"January 2, 2006, 15:04:05",
	"Jan 2, 2006 15:04:05",
	"Jan 2, 2006, 15:04:05",
	"2006/01/02 15:04:05",
	"02/01/2006 15:04:05",
	"2006-01-02",
	"02/01/2006",
	"01/02/2006",
}

func (f *Formatter) ParseRSSDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		log.Printf("DEBUG: Empty date string, using current time")
		return time.Now().UTC(), nil
	}

	for _, format := range rssDateFormats {
		if parsedTime, err := time.Parse(format, dateStr); err == nil {
			return parsedTime.UTC(), nil
		}
	}

	log.Printf("WARNING: Could not parse date '%s' with any known format, using fallback date", dateStr)
	fallbackDate := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	return fallbackDate, nil
}

func (f *Formatter) FormatForDisplay(t time.Time) string {
	local := t.Local()
	now := time.Now()

	if isSameDay(local, now) {
		return "Today"
	}

	yesterday := now.AddDate(0, 0, -1)
	if isSameDay(local, yesterday) {
		return "Yesterday"
	}

	weekAgo := now.AddDate(0, 0, -7)
	if local.After(weekAgo) {
		return local.Format("Monday")
	}

	if local.Year() == now.Year() {
		return local.Format("January 2")
	}

	return local.Format("January 2, 2006")
}

func (f *Formatter) FormatForGrouping(t time.Time) string {
	local := t.Local()
	return local.Format("2006-01-02")
}

func (f *Formatter) NormalizeToUTC(t time.Time) time.Time {
	return t.UTC()
}

func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}