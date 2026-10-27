package scraper

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

func validGlobalRow() globalLeaderboardRow {
	return globalLeaderboardRow{
		profileHref: "profile.php?u=888",
		rank:        "1.",
		username:    "Fioner",
		lastGame:    globalDate(),
		totalPoints: "100,100",
	}
}

func validLifetimeData() lifetimeData {
	return lifetimeData{
		memberSince: globalDate(),
		totalPoints: "100,100",
		solved:      "182",
		successRate: "100%",
		avgTime:     "1:00",
		attempted:   182,
	}
}

func TestParseGlobalRow(t *testing.T) {
	t.Run("valid row", func(t *testing.T) {
		got, err := parseGlobalRow(888, validGlobalRow(), validLifetimeData())
		if err != nil {
			t.Fatalf("expected err: %v", err)
		}

		if got.UID != 888 {
			t.Fatalf("UID = %d, want 888", got.UID)
		}

		if got.Rank != 1 {
			t.Fatalf("Rank = %d, want 1", got.Rank)
		}

		if got.Username != "Fioner" {
			t.Fatalf("Username = %s, want Fioner", got.Username)
		}

		if got.LastActive.IsZero() || got.MemberSince.IsZero() {
			t.Fatalf(
				"dates should be set: LastActive=%v MemberSince=%v",
				got.LastActive, got.MemberSince,
			)
		}

		if got.Solved != 182 || got.Attempted != 182 {
			t.Fatalf("Solved/Attempted = %d/%d, want 182/182", got.Solved, got.Attempted)
		}

		if got.AcceptanceRate != 100 {
			t.Fatalf("AcceptanceRate = %v", got.AcceptanceRate)
		}

		if got.AverageTime != 60 {
			t.Fatalf("AverageTime = %v, expected 60", got.AverageTime)
		}

		if got.AveragePoints != 550 {
			t.Fatalf("AveragePoints = %v, want 550", got.AveragePoints)
		}

		if got.TotalPoints != 100_100 {
			t.Fatalf("TotalPoints = %d, want 100,000", got.TotalPoints)
		}
	})

	t.Run("user with no solves does not divide by zero", func(t *testing.T) {
		lt := validLifetimeData()
		lt.solved = "0"
		lt.attempted = 0

		got, err := parseGlobalRow(888, validGlobalRow(), lt)
		if err != nil {
			t.Fatalf("expected err: %v", err)
		}

		if got.AcceptanceRate != 0 || got.AveragePoints != 0 {
			t.Errorf(
				"rate/avg = %v/%v, want 0/0",
				got.AcceptanceRate, got.AveragePoints,
			)
		}
	})

	errCases := []struct {
		name   string
		mutate func(*globalLeaderboardRow, *lifetimeData)
	}{
		{"bad rank", func(r *globalLeaderboardRow, _ *lifetimeData) { r.rank = "0th" }},
		{"bad last game", func(r *globalLeaderboardRow, _ *lifetimeData) { r.lastGame = "yesterday" }},
		{"bad total points on global row", func(r *globalLeaderboardRow, _ *lifetimeData) { r.totalPoints = "none" }},
		{"bad member since", func(_ *globalLeaderboardRow, lt *lifetimeData) { lt.memberSince = "epoch" }},
		{"bad solved", func(_ *globalLeaderboardRow, lt *lifetimeData) { lt.solved = "not a lot" }},
		{"bad avg time", func(_ *globalLeaderboardRow, lt *lifetimeData) { lt.avgTime = "" }},
	}

	for _, tt := range errCases {
		t.Run("error/"+tt.name, func(t *testing.T) {
			r, lt := validGlobalRow(), validLifetimeData()
			tt.mutate(&r, &lt)

			got, err := parseGlobalRow(888, r, lt)
			if err == nil {
				t.Fatalf("expected err, got user %+v", got)
			}
			if got != nil {
				t.Fatalf("user = %+v on err, want nil", got)
			}
		})
	}
}

// Serves a 2-user global leaderboard plus both profile pages
func globalRoutes() map[string]stub {
	date := globalDate()

	return map[string]stub{
		globalLeaderboardURL: {http.StatusOK, leaderboardHTML(
			rawLeaderboardRow{
				profileHref: "profile.php?u=1",
				cols:        []string{"1.", "uFirst", date, "5,000"},
			},
			rawLeaderboardRow{
				profileHref: "profile.php?u=2",
				cols:        []string{"2.", "uSecond", date, "1,000"},
			},
		)},
		fmt.Sprintf(playerProfileURL, 1): {http.StatusOK, profileHTML(
			date,
			[]string{"5,000", "10", "100%", "1:00"},
			[]string{"5", "5"},
		)},
		fmt.Sprintf(playerProfileURL, 2): {http.StatusOK, profileHTML(
			date,
			[]string{"1,000", "2", "100%", "1:20"},
			[]string{"1", "1"},
		)},
	}
}

func TestGetGlobalLeaderboardUsers(t *testing.T) {
	t.Run("combines leaderboard rows with profile data in page order", func(t *testing.T) {
		c := newTestClient(globalRoutes())

		users, err := c.getGlobalLeaderboardUsers(context.Background())
		if err != nil {
			t.Fatalf("expected err: %v", err)
		}

		if n := len(users); n != 2 {
			t.Errorf("got %d users, want 2", n)
		}

		u1, u2 := users[0], users[1]
		if u1.UID != 1 || u1.Rank != 1 || u1.Solved != 10 || u1.Attempted != 10 || u1.AverageTime != 60 {
			t.Errorf("u1 = %+v", u1)
		}

		if u2.UID != 2 || u2.Rank != 2 || u2.Solved != 2 || u2.Attempted != 2 || u2.AverageTime != 80 {
			t.Errorf("u2 = %+v", u2)
		}

	})

	t.Run("many rows keep order under concurrency", func(t *testing.T) {
		// Run with -race to catch potentially unsynchronized writes
		const n = 40
		date := globalDate()

		rows := make([]rawLeaderboardRow, 0, n)
		routes := map[string]stub{}

		for i := 0; i < n; i++ {
			uid := 1000 + i
			rows = append(rows, rawLeaderboardRow{
				profileHref: fmt.Sprintf("profile.php?u=%d", uid),
				cols:        []string{fmt.Sprintf("%d.", i+1), fmt.Sprintf("user%d", i), date, "100"},
			})
			routes[fmt.Sprintf(playerProfileURL, uid)] = stub{http.StatusOK, profileHTML(
				date, []string{"100", "10", "100%", "2:00"}, []string{"10"},
			)}
		}

		routes[globalLeaderboardURL] = stub{http.StatusOK, leaderboardHTML(rows...)}

		users, err := newTestClient(routes).getGlobalLeaderboardUsers(context.Background())
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if len(users) != n {
			t.Fatalf("got %d users, want %d", len(users), n)
		}
		for i, u := range users {
			if u == nil {
				t.Fatalf("users[%d] is nil", i)
			}
			if want := 1000 + i; u.UID != want {
				t.Errorf("users[%d].UID = %d, want %d (order broken)", i, u.UID, want)
			}
		}

		errCases := []struct {
			name   string
			mutate func(map[string]stub)
		}{
			{"leaderboard page fails", func(r map[string]stub) {
				r[globalLeaderboardURL] = stub{http.StatusInternalServerError, ""}
			}},
			{"one profile page fails", func(r map[string]stub) {
				r[fmt.Sprintf(playerProfileURL, 2)] = stub{http.StatusInternalServerError, ""}
			}},
			{"one profile page missing", func(r map[string]stub) {
				delete(r, fmt.Sprintf(playerProfileURL, 2))
			}},
			{"profile page has wrong number of totals", func(r map[string]stub) {
				r[fmt.Sprintf(playerProfileURL, 1)] = stub{http.StatusOK, profileHTML(
					globalDate(), []string{"1,000", "10"}, nil,
				)}
			}},
			{"row has wrong column count", func(r map[string]stub) {
				r[globalLeaderboardURL] = stub{http.StatusOK, leaderboardHTML(rawLeaderboardRow{
					profileHref: "profile.php?u=1", cols: []string{"1.", "u1"},
				})}
			}},
			{"row has no uid in href", func(r map[string]stub) {
				r[globalLeaderboardURL] = stub{http.StatusOK, leaderboardHTML(rawLeaderboardRow{
					profileHref: "profile.php", cols: []string{"1.", "u1", globalDate(), "1,000"},
				})}
			}},
		}

		for _, tt := range errCases {
			t.Run("error/"+tt.name, func(t *testing.T) {
				routes := globalRoutes()
				tt.mutate(routes)

				users, err := newTestClient(routes).getGlobalLeaderboardUsers(
					context.Background(),
				)
				if err == nil {
					t.Fatalf("expected err, got %d users", len(users))
				}
				if users != nil {
					t.Errorf("users = %v on err, want nil", users)
				}

			})
		}
	})

}
