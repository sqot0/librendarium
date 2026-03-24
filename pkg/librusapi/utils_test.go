package librusapi

import (
	"testing"
	"time"
)

func TestBuildEventTimes(t *testing.T) {
	hw := HomeWork{
		ID:       1,
		Date:     "2026-03-25",
		TimeFrom: "08:15:00",
		TimeTo:   "09:00:00",
	}

	start, end, err := BuildEventTimes(hw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loc, _ := time.LoadLocation("Europe/Warsaw")
	expectedStart := time.Date(2026, 3, 25, 8, 15, 0, 0, loc)
	expectedEnd := time.Date(2026, 3, 25, 9, 0, 0, 0, loc)

	if !start.Equal(expectedStart) {
		t.Errorf("expected start %v, got %v", expectedStart, start)
	}
	if !end.Equal(expectedEnd) {
		t.Errorf("expected end %v, got %v", expectedEnd, end)
	}
}

func TestBuildEventTimesDefault(t *testing.T) {
	hw := HomeWork{
		ID:       2,
		Date:     "2026-03-25",
		TimeFrom: "",
		TimeTo:   "",
	}

	start, end, err := BuildEventTimes(hw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loc, _ := time.LoadLocation("Europe/Warsaw")
	expectedStart := time.Date(2026, 3, 25, 8, 0, 0, 0, loc)
	expectedEnd := time.Date(2026, 3, 25, 16, 0, 0, 0, loc)

	if !start.Equal(expectedStart) {
		t.Errorf("expected start %v, got %v", expectedStart, start)
	}
	if !end.Equal(expectedEnd) {
		t.Errorf("expected end %v, got %v", expectedEnd, end)
	}
}

func TestBuildEventTimesInvalidDate(t *testing.T) {
	hw := HomeWork{
		ID:   3,
		Date: "not-a-date",
	}

	_, _, err := BuildEventTimes(hw)
	if err == nil {
		t.Error("expected error for invalid date, got nil")
	}
}
