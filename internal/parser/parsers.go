package parser

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"nook/internal/models"
)

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

	rank, err := strconv.Atoi(rankStr)
	if err != nil {
		return -1, fmt.Errorf("invalid rank string %q: %w", rankStr, err)
	}

	return rank, nil
}

func Date(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	layouts := []string{models.MonthDateLayout, models.GlobalDateLayout}

	var errs []error
	for _, layout := range layouts {
		// NOTE: parse in location since the passed string
		// is already in the site's location (ET)
		t, err := time.ParseInLocation(layout, s, models.SiteLocation)
		if err == nil {
			return t, nil
		}

		errs = append(errs, fmt.Errorf("layout %q: %w", layout, err))
	}

	return time.Time{}, fmt.Errorf(
		"no matching date layout for %q: %w",
		s, errors.Join(errs...),
	)
}

func Solved(s string) (int, error) {
	return strconv.Atoi(
		strings.ReplaceAll(strings.Fields(s)[0], ",", ""),
	)
}

func Attempted(s string) (int, error) {
	return Solved(s)
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
	trimmed := strings.TrimSpace(s)

	parts := strings.Split(trimmed, ":")
	if len(parts) > 3 {
		return 0, fmt.Errorf("invalid avg time %q: expected [[h:]m:]s", s)
	}

	var total float64
	for _, p := range parts {
		n, err := strconv.ParseUint(p, 10, 32)
		if err != nil {
			return 0, fmt.Errorf("invalid time %q: %w", s, err)
		}
		total = total*60 + float64(n)
	}

	return total, nil
}

func AvgTimeMonthly(s string) (float64, error) {
	return strconv.ParseFloat(strings.Fields(s)[0], 64)
}

func AvgPoints(s string) (float64, error) {
	return strconv.ParseFloat(strings.Fields(s)[0], 64)
}

func AvgPointsGlobal(totalPoints int, solved int) (float64, error) {
	return float64(totalPoints / solved), nil
}

func TotalPoints(s string) (int, error) {
	return strconv.Atoi(strings.ReplaceAll(s, ",", ""))
}

// Integers with thousands separators: "5,888" to 5888
func Count(s string) (int, error) {
	return strconv.Atoi(
		strings.ReplaceAll(strings.TrimSpace(s), ",", ""))
}

// Clock format: "2:35" or "1:02:37"
func AvgTimeClock(s string) (float64, error) {
	parts := strings.Split(strings.TrimSpace(s), ":")
	if lenParts := len(parts); lenParts < 2 || lenParts > 3 {
		return -1, fmt.Errorf("unexpected clock format %q", s)
	}

	secs := 0
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return -1, fmt.Errorf("bad clock value %q", s)
		}
		secs = 60*secs + n
	}

	return float64(secs), nil
}

func AvgPointsDivide(total, solved int) float64 {
	if solved == 0 {
		return 0
	}

	return math.Round(float64(total)/float64(solved)*100) / 100
}
