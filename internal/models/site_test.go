package models

import (
	"testing"
	"time"
)

func TestSiteLocation(t *testing.T) {
	if SiteLocation == nil {
		t.Fatal("SiteLocation is nil")
	}

	if got, want := SiteLocation.String(), "America/New_York"; got != want {
		t.Errorf("SiteLocation = %q, want %q", got, want)
	}
}

func TestMonthStart(t *testing.T) {
	tests := []struct {
		name  string
		input time.Time
	}{
		{"start of month (AM)", time.Date(2026, 4, 4, 5, 44, 1, 0, SiteLocation)},
		{"start of month (PM)", time.Date(2026, 4, 4, 16, 4, 0, 0, SiteLocation)},
		{"middle of month (AM)", time.Date(2026, 4, 16, 5, 44, 1, 0, SiteLocation)},
		{"middle of month (PM)", time.Date(2026, 4, 16, 16, 4, 0, 0, SiteLocation)},
		{"end of month (AM)", time.Date(2026, 4, 30, 5, 44, 1, 0, SiteLocation)},
		{"end of month (PM)", time.Date(2026, 4, 30, 16, 4, 0, 0, SiteLocation)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MonthStart(tt.input)
			want := time.Date(
				tt.input.Year(), tt.input.Month(),
				1, 0, 0, 0, 0,
				time.UTC, // normalize to UTC as standardized timezone
			)

			if got != want {
				t.Errorf("MonthStart() = %v, want %v", got, want)
			}
		})
	}
}

func TestDayStart(t *testing.T) {
	tests := []struct {
		name  string
		input time.Time
	}{
		{"morning", time.Date(2026, 4, 1, 0, 0, 0, 0, SiteLocation)},
		{"noon", time.Date(2026, 4, 1, 12, 0, 0, 0, SiteLocation)},
		{"night", time.Date(2026, 4, 1, 22, 10, 0, 0, SiteLocation)},
		{"before midnight", time.Date(2026, 4, 1, 23, 59, 59, 0, SiteLocation)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DayStart(tt.input)
			want := time.Date(
				tt.input.Year(), tt.input.Month(), tt.input.Day(),
				0, 0, 0, 0,
				time.UTC, // normalize to UTC as standardized timezone
			)

			if got != want {
				t.Errorf("DayStart() = %v, want %v", got, want)
			}
		})
	}
}
