package scraper

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"nook/internal/models"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

// A canned HTTP response
type stub struct {
	status int
	body   string
}

// Returns a scraper [*Client] whose requests are answered from routes,
// keyed by a full URL. Unknown URLs get 404 status
func newTestClient(routes map[string]stub) *Client {
	c := NewClient()
	c.http.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		s, ok := routes[req.URL.String()]
		if !ok {
			s = stub{status: http.StatusNotFound}
		}

		return &http.Response{
			StatusCode: s.status,
			Status:     fmt.Sprintf("%d %s", s.status, http.StatusText(s.status)),
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(s.body)),
			Request:    req,
		}, nil
	})
	return c
}

var refDate = time.Date(2026, time.April, 19, 0, 0, 0, 0, models.SiteLocation)

func monthDate() string  { return refDate.Format(models.MonthDateLayout) }
func globalDate() string { return refDate.Format(models.GlobalDateLayout) }

// Builds a (rough) leaderboard page: one header row + a row per argument.
// The profile link is placed on the second column if it exists.
func leaderboardHTML(rows ...rawLeaderboardRow) string {
	var b strings.Builder

	b.WriteString(`<table class="winners"><tbody>`)
	b.WriteString(`<tr><th>Rank</th><th>Username</th></tr>`)

	for _, r := range rows {
		b.WriteString("<tr>")

		for j, c := range r.cols {
			if j == 1 && r.profileHref != "" {
				fmt.Fprintf(&b, `<td><a href=%q>%s</a></td>`, r.profileHref, c)
			} else {
				fmt.Fprintf(&b, `<td>%s</td>`, c)
			}
		}
		b.WriteString("</tr>")
	}

	b.WriteString("</tbody></table>")
	return b.String()
}

// Builds a (rough) player profile page.
func profileHTML(memberSince string, totals, played []string) string {
	var b strings.Builder

	fmt.Fprintf(&b,
		`<div id="tabs-1"><table><tbody><tr><td>Name</td><td>x</td><td>%s</td></tr></tbody></table></div>`,
		memberSince,
	)

	for _, t := range totals {
		fmt.Fprintf(&b, `<div class="sc-total"><b>%s</b></div>`, t)
	}

	for _, p := range played {
		fmt.Fprintf(&b, `<div class="sc-stats"><b>%s</b></div>`, p)
	}

	return b.String()
}

func TestClient_get(t *testing.T) {
	const url = "https://example.test/page"

	t.Run("200 returns a parsed document", func(t *testing.T) {
		c := newTestClient(map[string]stub{
			url: {http.StatusOK, `<p id="x">bello</p>`},
		})

		doc, err := c.get(context.Background(), url)
		if err != nil {
			t.Fatalf("get() expected err: %v", err)
		}

		if got := doc.Find("p#x").Text(); got != "bello" {
			t.Errorf("document text = %q, want %q", got, "bello")
		}
	})

	t.Run("non-200 is an error mentioning status", func(t *testing.T) {
		c := newTestClient(map[string]stub{
			url: {http.StatusInternalServerError, ""},
		})

		_, err := c.get(context.Background(), url)
		if err == nil {
			t.Fatal("get() expected err for 500, got nil")
		}

		if !strings.Contains(err.Error(), "500") {
			t.Errorf("error %q should mention status 500", err)
		}
	})

	t.Run("404 for unknown url", func(t *testing.T) {
		c := newTestClient(nil)
		if _, err := c.get(context.Background(), url); err == nil {
			t.Fatal("get() expected err for 404, got nil")
		}
	})

	t.Run("transport failure is wrapped", func(t *testing.T) {
		bad := errors.New("bad")

		c := NewClient()
		c.http.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, bad
		})

		_, err := c.get(context.Background(), url)
		if !errors.Is(err, bad) {
			t.Errorf("get() err = %v, want it to wrap %v", err, bad)
		}
	})

	t.Run("invalid url", func(t *testing.T) {
		c := newTestClient(nil)
		if _, err := c.get(context.Background(), "://test"); err == nil {
			t.Fatal("get() expected err for malformed url, got nil")
		}
	})
}

func TestGetLeaderboardUsers_ReturnsStatsOnFail(t *testing.T) {
	c := newTestClient(map[string]stub{
		monthlyLeaderboardURL: {http.StatusInternalServerError, ""},
	})

	for _, lb := range []models.Leaderboard{
		models.MonthlyLeaderboard,
		models.GlobalLeaderboard,
	} {
		t.Run(lb.DBKind()+"_stats", func(t *testing.T) {
			users, stats, err := c.GetLeaderboardUsers(t.Context(), lb)

			if err == nil {
				t.Fatal("expected err, got nil")
			}

			if users != nil {
				t.Errorf("users = %v, want nil on err", users)
			}

			if stats == nil {
				t.Fatal("stats should be returned on err")
			}

			if stats.ScrapedAt.IsZero() {
				t.Error("stats.ScrapedAt should be set")
			}
		})
	}
}

func TestGetLeaderboardUsers_DispatchesByLeaderboard(t *testing.T) {
	c := newTestClient(map[string]stub{
		monthlyLeaderboardURL: {http.StatusOK, leaderboardHTML(rawLeaderboardRow{
			profileHref: "profile.php?u=10305",
			cols: []string{
				"1.", "", monthDate(), "1/4", "25%", "60", "563.0", "563",
			},
		})},
	})

	users, stats, err := c.GetLeaderboardUsers(context.Background(), models.MonthlyLeaderboard)
	if err != nil {
		t.Fatalf("monthly: unexpected err: %v", err)
	}

	if n := len(users); n != 1 || stats == nil {
		t.Fatalf("monthly: got %d users, stats=%v", n, stats)
	}

	if _, _, err := c.GetLeaderboardUsers(
		context.Background(), models.GlobalLeaderboard,
	); err == nil {
		t.Fatalf("global: expected err because only monthly url is served")
	}
}
