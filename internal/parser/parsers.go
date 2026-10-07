package parser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var eastern *time.Location

func init() {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic(err)
	}
	eastern = loc
}

func UID(s string) (int, error) {
	sep := "?u="
	_, uidStr, found := strings.Cut(s, sep)

	if !found {
		return -1, fmt.Errorf(
			"seperator %q not found in %q", sep, s,
		)
	}

	return strconv.Atoi(uidStr)
}

func Rank(s string) (int, error) {
	sep := "."
	rankStr, _, found := strings.Cut(s, sep)

	if !found {
		return -1, fmt.Errorf(
			"seperator %q not found in %q", sep, s,
		)
	}
	return strconv.Atoi(rankStr)
}

func Date(s string) (time.Time, error) {
	layouts := []string{
		"January 2, 2006, 3:04 pm", // monthly leaderboard format
		"January 2, 2006",          // global leaderboard format
	}

	var errs []error
	for _, layout := range layouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			return t.In(eastern), nil
		}

		errs = append(errs, fmt.Errorf("layout %q: %w", layout, err))
	}

	return time.Time{}, fmt.Errorf(
		"no matching date layout for %q: %w",
		s, errors.Join(errs...),
	)
}

func Solved(s string) (int, error) {
	return strconv.Atoi(strings.Fields(s)[0])
}

func Attempted(s string) (int, error) {
	return strconv.Atoi(strings.Fields(s)[0])
}

func SolveAttempted(s string) (int, int, error) {
	sep := "/"
	solvedStr, attemptedStr, found := strings.Cut(s, sep)

	if !found {
		return -1, -1, fmt.Errorf(
			"seperator %q not found in %q", sep, s,
		)
	}

	solved, err := strconv.Atoi(solvedStr)
	if err != nil {
		return -1, -1, err
	}

	attempted, err := strconv.Atoi(attemptedStr)
	if err != nil {
		return solved, -1, err
	}

	return solved, attempted, err
}

func AcceptanceRate(solved, attempted int) float64 {
	if attempted == 0 {
		return 0
	}

	rate := (float64(solved) / float64(attempted)) * 100
	return float64(int(rate*100)) / 100
}

func AvgTimeGlobal(s string) (float64, error) {
	sep := "\u00a0" // \xa0 == non-breaking space
	timeStr, _, found := strings.Cut(s, sep)

	if !found {
		return -1.0, fmt.Errorf(
			"seperator %q not found in %q", sep, timeStr,
		)
	}

	return strconv.ParseFloat(timeStr, 64)
}

func AvgTimeMonthly(s string) (float64, error) {
	return strconv.ParseFloat(strings.Fields(s)[0], 64)
}

func AvgPoints(s string) (float64, error) {
	return strconv.ParseFloat(strings.Fields(s)[0], 64)
}

func TotalPoints(s string) (int, error) {
	return strconv.Atoi(strings.ReplaceAll(s, ",", ""))
}
