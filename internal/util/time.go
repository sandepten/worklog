package util

import "time"

// TruncateToLocalDay returns a time representing midnight of the same day in the local timezone of t.
func TruncateToLocalDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
