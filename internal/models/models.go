package models

import (
	"fmt"
	"time"
)

type Leaderboard int

const (
	MonthlyLeaderboard Leaderboard = iota
	GlobalLeaderboard
)

func (lb Leaderboard) String() string {
	switch lb {
	case MonthlyLeaderboard:
		return "MonthlyLeaderboard"
	case GlobalLeaderboard:
		return "GlobalLeaderboard"
	default:
		return fmt.Sprintf("Leaderboard(%d)", int(lb))
	}
}

func (lb Leaderboard) DBKind() string {
	switch lb {
	case MonthlyLeaderboard:
		return "monthly"
	case GlobalLeaderboard:
		return "global"
	default:
		return fmt.Sprintf("unknown(%d)", int(lb))
	}
}

func ParseLeaderboard(s string) (Leaderboard, error) {
	switch s {
	case "monthly":
		return MonthlyLeaderboard, nil
	case "global":
		return GlobalLeaderboard, nil
	default:
		return 0, fmt.Errorf("unknown leaderboard %q", s)
	}
}

type User struct {
	UID            int
	Rank           int
	Username       string
	LastActive     time.Time
	Solved         int
	Attempted      int
	AcceptanceRate float64
	AverageTime    float64
	AveragePoints  float64
	TotalPoints    int
}

func (u *User) String() string {
	rank := fmt.Sprintf("%3d", u.Rank)                                        // right-align in 3 chars; '%d' right-aligns by default
	username := fmt.Sprintf("%-33s", "@"+u.Username)                          // left-align in 33 chars; '-' = left
	totalPoints := fmt.Sprintf("%-15s", fmt.Sprintf("%d pts", u.TotalPoints)) // left-align in 15 chars

	return fmt.Sprintf("%s %s %s", rank, username, totalPoints)
}

func (u *User) DetailedString() string {
	lastActive := "never"
	if !u.LastActive.IsZero() {
		// lastActive = u.LastActive.Format("2006-01-02 15:04:05")
		lastActive = u.LastActive.Format("2006-01-02")
	}

	uid := fmt.Sprintf("UID: %-6d", u.UID)
	rank := fmt.Sprintf("Rank: %-4d", u.Rank)
	username := fmt.Sprintf("@%-25s", u.Username)
	last := fmt.Sprintf("LastActive: %-11s", lastActive)

	fraction := fmt.Sprintf("%d/%d", u.Solved, u.Attempted)
	solved := fmt.Sprintf("Solved: %-12s", fraction)

	rateString := fmt.Sprintf("(%.2f%%)", u.AcceptanceRate)
	rate := fmt.Sprintf("%-10s", rateString)

	avgTimeStr := fmt.Sprintf("%.2fs", u.AverageTime)
	avgTime := fmt.Sprintf("AvgTime: %-9s", avgTimeStr)
	avgPts := fmt.Sprintf("AvgPts: %-8.2f", u.AveragePoints)
	total := fmt.Sprintf("TotalPoints: %-6d", u.TotalPoints)

	return fmt.Sprintf(
		"User{%s %s %s %s %s %s %s %s %s}",
		uid, rank, username, last, solved, rate, avgTime, avgPts, total,
	)
}
