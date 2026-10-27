package scraper

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func mustDoc(t *testing.T, html string) *goquery.Document {
	t.Helper()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("building doc: %v", err)
	}
	return doc
}

func TestExtractLeaderboardRows(t *testing.T) {
	t.Run("skips header, trims text, captures profile link", func(t *testing.T) {
		doc := mustDoc(t, `
			<table class="winners"><tbody>
				<tr><th>Rank</th><th>Username</th></tr>
				<tr><td> 1. </td><td><a href="profile.php?u=1"> u1 </a></td></tr>
				<tr><td>2.</td><td><a href="profile.php?u=2">u2</a></td></tr>
			</tbody></table>
		`)

		rows, err := extractLeaderboardRows(doc)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if n := len(rows); n != 2 {
			t.Errorf("got %d rows, want 2", n)
		}

		if want := []string{"1.", "u1"}; !slices.Equal(rows[0].cols, want) {
			t.Errorf("rows[0].cols = %q, want %q", rows[0].cols, want)
		}

		for i, r := range rows {
			if want := fmt.Sprintf("profile.php?u=%d", i+1); r.profileHref != want {
				t.Errorf("rows[%d].profileHref = %q, want %q", i, r.profileHref, want)
			}
		}
	})

	t.Run("row without a link has empty href", func(t *testing.T) {
		doc := mustDoc(t, leaderboardHTML(rawLeaderboardRow{
			cols: []string{"1.", "ghost"},
		}))

		rows, err := extractLeaderboardRows(doc)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}

		if n := len(rows); n != 1 || rows[0].profileHref != "" {
			t.Errorf("rows = %+v, want one row with empty href", rows)
		}
	})

	errCases := []struct {
		name string
		html string
	}{
		{"no table", "<p>paragraph</p>"},
		{"table without 'winners' class", "<table><tbody><tr><td>1.</td></tr></tbody></table>"},
		{"header only", leaderboardHTML()},
	}

	for _, tt := range errCases {
		t.Run("error/"+tt.name, func(t *testing.T) {
			if _, err := extractLeaderboardRows(
				mustDoc(t, tt.html),
			); err == nil {
				t.Fatal("expected err, got nil")
			}
		})
	}
}

func TestRawLeaderboardRow_Monthly(t *testing.T) {
	cols := []string{"1.", "u1", "date", "1/10", "10%", "60 secs.", "555.5 pts.", "555"}

	t.Run("maps columns in order", func(t *testing.T) {
		got, err := rawLeaderboardRow{profileHref: "href", cols: cols}.Monthly()
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}

		want := monthlyLeaderboardRow{
			profileHref: "href", rank: "1.", username: "u1",
			lastActive: "date", solvedFraction: "1/10", rate: "10%",
			avgTime: "60 secs.", avgPoints: "555.5 pts.",
			totalPoints: "555",
		}

		if got != want {
			t.Errorf("Monthly() = %+v, want %+v", got, want)
		}
	})

	for _, n := range []int{0, 4, 7, 9} {
		t.Run("wrong column count", func(t *testing.T) {
			row := rawLeaderboardRow{cols: make([]string, n)}
			if _, err := row.Monthly(); err == nil {
				t.Errorf("Monthly() with %d cols: expected err", n)
			}
		})
	}
}

func TestRawLeaderboardRow_Global(t *testing.T) {
	cols := []string{"1.", "u1", "date", "100,100"}

	t.Run("maps columns in order", func(t *testing.T) {
		got, err := rawLeaderboardRow{profileHref: "href", cols: cols}.Global()
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}

		want := globalLeaderboardRow{
			profileHref: "href", rank: "1.", username: "u1",
			lastGame: "date", totalPoints: "100,100",
		}

		if got != want {
			t.Errorf("Global() = %+v, want %+v", got, want)
		}
	})

	for _, n := range []int{0, 3, 5, 8} {
		t.Run("wrong column count", func(t *testing.T) {
			row := rawLeaderboardRow{cols: make([]string, n)}
			if _, err := row.Global(); err == nil {
				t.Errorf("Global() with %d cols: expected err", n)
			}
		})
	}
}

func TestExtractLifetimeData(t *testing.T) {
	t.Run("valid profile", func(t *testing.T) {
		html := profileHTML(
			"  April 15, 2026 ",
			[]string{"2,971,338", "5,468", "99.9%", "2:37"},
			[]string{"6", "4"},
		)

		got, err := extractLifetimeData(mustDoc(t, html))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.memberSince != "April 15, 2026" {
			t.Errorf("memberSince = %q", got.memberSince)
		}
		if want := []string{"2,971,338", "5,468", "99.9%", "2:37"}; !slices.Equal(got.totals, want) {
			t.Errorf("totals = %q, want %q", got.totals, want)
		}
		if want := []string{"6", "4"}; !slices.Equal(got.played, want) {
			t.Errorf("played = %q, want %q", got.played, want)
		}
	})

	t.Run("player who has never played is valid", func(t *testing.T) {
		html := profileHTML("April 15, 2026", []string{"0", "0", "0%", "0:00"}, nil)

		got, err := extractLifetimeData(mustDoc(t, html))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got.played) != 0 {
			t.Errorf("played = %q, want empty", got.played)
		}
	})

	const table = `<div id="tabs-1"><table><tbody><tr><td>a</td><td>b</td><td>April 15, 2026</td></tr></tbody></table></div>`

	errCases := []struct {
		name string
		html string
	}{
		{"no tabs-1 table", `<div class="sc-total"><b>1</b></div>`},
		{"too few tds for member since", `<div id="tabs-1"><table><tbody><tr><td>a</td><td>b</td></tr></tbody></table></div><div class="sc-total"><b>1</b></div>`},
		{"no sc-total", table},
		{"sc-total without <b>", table + `<div class="sc-total">1</div>`},
		{"sc-total with two <b>", table + `<div class="sc-total"><b>1</b><b>2</b></div>`},
		{"sc-stats without <b>", table + `<div class="sc-total"><b>1</b></div><div class="sc-stats">nothing</div>`},
	}
	for _, tt := range errCases {
		t.Run("error/"+tt.name, func(t *testing.T) {
			if _, err := extractLifetimeData(mustDoc(t, tt.html)); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestRawLifetimeData_Lifetime(t *testing.T) {
	totals := []string{"2,971,338", "5,468", "99.9%", "2:37"}

	t.Run("maps totals and sums played counts", func(t *testing.T) {
		raw := rawLifetimeData{
			memberSince: "April 15, 2026",
			totals:      totals,
			played:      []string{"6", "1,004", "10"},
		}

		got, err := raw.Lifetime()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := lifetimeData{
			memberSince: "April 15, 2026",
			totalPoints: "2,971,338",
			solved:      "5,468",
			successRate: "99.9%",
			avgTime:     "2:37",
			attempted:   1020,
		}
		if got != want {
			t.Errorf("Lifetime() = %+v, want %+v", got, want)
		}
	})

	t.Run("no played counts means zero attempted", func(t *testing.T) {
		got, err := rawLifetimeData{totals: totals}.Lifetime()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.attempted != 0 {
			t.Errorf("attempted = %d, want 0", got.attempted)
		}
	})

	t.Run("wrong number of totals", func(t *testing.T) {
		for _, n := range []int{0, 3, 5} {
			raw := rawLifetimeData{totals: make([]string, n)}
			if _, err := raw.Lifetime(); err == nil {
				t.Errorf("Lifetime() with %d totals: expected error", n)
			}
		}
	})

	t.Run("unparsable played count", func(t *testing.T) {
		raw := rawLifetimeData{totals: totals, played: []string{"6", "lots"}}
		if _, err := raw.Lifetime(); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
