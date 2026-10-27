package scraper

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"nook/internal/models"
)

func validMonthlyRow() monthlyLeaderboardRow {
	return monthlyLeaderboardRow{
		profileHref:    "profile.php?u=99",
		rank:           "3.",
		username:       "LeeSedol",
		lastActive:     monthDate(),
		solvedFraction: "1/10",
		rate:           "10%",
		avgTime:        "66.3",
		avgPoints:      "563.0",
		totalPoints:    "563",
	}
}

func TestParseMonthRow(t *testing.T) {
	t.Run("valid row", func(t *testing.T) {
		got, err := parseMonthRow(validMonthlyRow())

		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}

		if got.UID != 99 {
			t.Errorf("UID = %d, want 99", got.UID)
		}

		if got.Rank != 3 {
			t.Errorf("Rank = %d, want 3", got.Rank)
		}

		if got.Username != "LeeSedol" {
			t.Errorf("Username = %q, want LeeSedol", got.Username)
		}

		if got.Solved != 1 || got.Attempted != 10 {
			t.Errorf("Solved/Attempted = %d/%d, want 1/10", got.Solved, got.Attempted)
		}

		if got.AcceptanceRate != 10 {
			t.Errorf("AcceptanceRate = %v, want 10", got.AcceptanceRate)
		}

		if got.AverageTime != 66.3 {
			t.Errorf("AverageTime = %v, want 66.3", got.AverageTime)
		}
		if got.AveragePoints != 563.0 {
			t.Errorf("AveragePoints = %v, want 563.0", got.AveragePoints)
		}
		if got.TotalPoints != 563.0 {
			t.Errorf("TotalPoints = %d, want 563.0", got.TotalPoints)
		}

		if got.LastActive.Format(models.MonthDateLayout) != monthDate() &&
			got.LastActive.Format(models.GlobalDateLayout) != monthDate() {
			t.Errorf("LastActive = %v, want date %q", got.LastActive, monthDate())
		}
	})

	errCases := []struct {
		name   string
		mutate func(*monthlyLeaderboardRow)
	}{
		{"href without uid", func(r *monthlyLeaderboardRow) { r.profileHref = "profile.php" }},
		{"empty href", func(r *monthlyLeaderboardRow) { r.profileHref = "" }},
		{"bad rank", func(r *monthlyLeaderboardRow) { r.rank = "first" }},
		{"bad date", func(r *monthlyLeaderboardRow) { r.lastActive = "yesterday" }},
		{"bad solved fraction", func(r *monthlyLeaderboardRow) { r.solvedFraction = "lots" }},
		{"bad avg time", func(r *monthlyLeaderboardRow) { r.avgTime = "fast" }},
		{"bad avg points", func(r *monthlyLeaderboardRow) { r.avgPoints = "many" }},
		{"bad total points", func(r *monthlyLeaderboardRow) { r.totalPoints = "some" }},
	}

	for _, tt := range errCases {
		t.Run("error/"+tt.name, func(t *testing.T) {
			r := validMonthlyRow()
			tt.mutate(&r)

			got, err := parseMonthRow(r)
			if err == nil {
				t.Fatalf("expected err, got user %+v", got)
			}
			if got != nil {
				t.Errorf("user = %+v on err, want nil", got)
			}
		})
	}
}

func TestGetMonthlyLeaderboardUsers(t *testing.T) {
	createMonthRow := func(href, rank, username string) rawLeaderboardRow {
		return rawLeaderboardRow{
			profileHref: href,
			cols: []string{
				rank, username, monthDate(), "1/4", "25%", "60", "563.1", "563",
			},
		}
	}

	t.Run("returns users in page order", func(t *testing.T) {
		c := newTestClient(map[string]stub{
			monthlyLeaderboardURL: {http.StatusOK, leaderboardHTML(
				createMonthRow("profile.php?u=100", "1.", "u1"),
				createMonthRow("profile.php?u=101", "2.", "u2"),
			)},
		})

		users, err := c.getMonthlyLeaderboardUsers(context.Background())
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}

		if n := len(users); n != 2 {
			t.Fatalf("got %d users, want 2", n)
		}

		if users[0].Username != "u1" || users[0].UID != 100 || users[0].Rank != 1 {
			t.Errorf("users[0] = %+v", users[0])
		}

		if users[1].Username != "u2" || users[1].UID != 101 || users[1].Rank != 2 {
			t.Errorf("users[1] = %+v", users[1])
		}
	})

	errCases := []struct {
		name  string
		stub  stub
		wants string // substring expected in the err
	}{
		{"server err", stub{http.StatusInternalServerError, ""}, "500"},
		{"no leaderboard rows", stub{http.StatusOK, leaderboardHTML()}, ""},
		{
			"wrong column count",
			stub{http.StatusOK, leaderboardHTML(rawLeaderboardRow{
				profileHref: "profile.php?u=88", cols: []string{"1.", "Drake"},
			})},
			"row 0",
		},
		{
			"second row unparsable",
			stub{http.StatusOK, leaderboardHTML(
				createMonthRow("profile.php?u=100", "1.", "u1"),
				createMonthRow("profile.php?u=101", "x.", "u2"),
			)},
			"row 1",
		},
	}

	for _, tt := range errCases {
		t.Run("error/"+tt.name, func(t *testing.T) {
			c := newTestClient(map[string]stub{monthlyLeaderboardURL: tt.stub})

			users, err := c.getMonthlyLeaderboardUsers(context.Background())
			if err == nil {
				t.Fatalf("expected err, got %d users", len(users))
			}

			if tt.wants != "" && !strings.Contains(err.Error(), tt.wants) {
				t.Errorf("err %q should contain %q", err, tt.wants)
			}
		})
	}
}
