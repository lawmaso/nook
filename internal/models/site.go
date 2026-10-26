package models

import "time"

var SiteLocation *time.Location

const (
	MonthDateLayout  = "January 2, 2006, 3:04 pm" // monthly leaderboard format
	GlobalDateLayout = "January 2, 2006"          // global leaderboard format
)

func init() {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic(err)
	}
	SiteLocation = loc
}

func MonthStart(t time.Time) time.Time {
	t = t.In(SiteLocation)
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func DayStart(t time.Time) time.Time {
	t = t.In(SiteLocation)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
