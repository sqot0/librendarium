package librusapi

import (
	"fmt"
	"time"
)

// BuildEventTimes parses the date and time strings from a HomeWork and returns the start and end times
// for a calendar event in the Europe/Warsaw timezone.
func BuildEventTimes(hw HomeWork) (time.Time, time.Time, error) {
	loc, err := time.LoadLocation("Europe/Warsaw")
	if err != nil {
		loc = time.UTC
	}

	date, err := time.Parse("2006-01-02", hw.Date)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid date %q: %w", hw.Date, err)
	}

	parseTime := func(s string, defaultHour int) (int, int, int) {
		t, err := time.Parse("15:04:05", s)
		if err != nil {
			return defaultHour, 0, 0
		}
		return t.Hour(), t.Minute(), t.Second()
	}

	startHour, startMin, startSec := parseTime(hw.TimeFrom, 8)
	endHour, endMin, endSec := parseTime(hw.TimeTo, 16)

	start := time.Date(date.Year(), date.Month(), date.Day(), startHour, startMin, startSec, 0, loc)
	end := time.Date(date.Year(), date.Month(), date.Day(), endHour, endMin, endSec, 0, loc)

	return start, end, nil
}
