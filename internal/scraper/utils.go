package scraper

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"nook/internal/models"
	"nook/internal/parser"
)

const (
	globalLeaderboardURL  = "https://starbattle.puzzlebaron.com/halloffame.php"
	monthlyLeaderboardURL = "https://starbattle.puzzlebaron.com/contest1.php"
	playerProfileURL      = "https://starbattle.puzzlebaron.com/profile.php?u=%d"
)

// Raw leaderboard row text data
type rawLeaderboardRow struct {
	profileHref string   // link to user profile
	cols        []string // leaderboard columns
}

// Monthly row text data
type monthlyLeaderboardRow struct {
	profileHref    string
	rank           string
	username       string
	lastActive     string
	solvedFraction string
	rate           string
	avgTime        string
	avgPoints      string
	totalPoints    string
}

// Global row text data
type globalLeaderboardRow struct {
	profileHref string
	rank        string
	username    string
	lastGame    string
	totalPoints string
}

// Extracts and returns the text from a leaderboard
func (client *Client) getLeaderboardText(
	ctx context.Context,
	lb models.Leaderboard,
) ([]rawLeaderboardRow, error) {
	url := monthlyLeaderboardURL
	if lb == models.GlobalLeaderboard {
		url = globalLeaderboardURL
	}

	doc, err := client.get(ctx, url)
	if err != nil {
		return nil, err
	}

	return extractLeaderboardRows(doc)
}

// Extracts the raw leaderboard rows
func extractLeaderboardRows(doc *goquery.Document) ([]rawLeaderboardRow, error) {
	rows := doc.Find("table.winners tbody tr")

	// Expect the header row + at least one row
	if rows.Length() < 2 {
		return nil, errors.New("no lb rows found")
	}

	data := make([]rawLeaderboardRow, 0, rows.Length()-1)
	rows.Each(func(i int, row *goquery.Selection) {
		// Skip the header row
		if i == 0 {
			return
		}

		var r rawLeaderboardRow
		row.Find("td").Each(func(_ int, td *goquery.Selection) {
			r.cols = append(r.cols, strings.TrimSpace(td.Text()))
			if href, ok := td.Find("a").Attr("href"); ok {
				r.profileHref = href
			}
		})
		data = append(data, r)
	})

	return data, nil
}

// Adapts [rawLeaderboardRow] into [monthlyLeaderboardRow]
func (row rawLeaderboardRow) Monthly() (monthlyLeaderboardRow, error) {
	expectedCols := 8
	if len(row.cols) != expectedCols {
		return monthlyLeaderboardRow{}, fmt.Errorf(
			"expected %d columns, got %d", expectedCols, len(row.cols),
		)
	}

	return monthlyLeaderboardRow{
		profileHref:    row.profileHref,
		rank:           row.cols[0],
		username:       row.cols[1],
		lastActive:     row.cols[2],
		solvedFraction: row.cols[3],
		rate:           row.cols[4],
		avgTime:        row.cols[5],
		avgPoints:      row.cols[6],
		totalPoints:    row.cols[7],
	}, nil
}

// Adapts [rawLeaderboardRow] into [globalLeaderboardRow]
func (row rawLeaderboardRow) Global() (globalLeaderboardRow, error) {
	expectedCols := 4
	if len(row.cols) != expectedCols {
		return globalLeaderboardRow{}, fmt.Errorf(
			"expected %d columns, got %d", expectedCols, len(row.cols),
		)
	}

	return globalLeaderboardRow{
		profileHref: row.profileHref,
		rank:        row.cols[0],
		username:    row.cols[1],
		lastGame:    row.cols[2],
		totalPoints: row.cols[3],
	}, nil
}

// Is this even needed?
// The new site restructing forces us out of
// the generalization of parsing rows of the lifetime table
// (which no longer exists)
type rawLifetimeData struct {
	memberSince string
	totals      []string
	played      []string
}

type lifetimeData struct {
	memberSince string // ex: April 15, 2026
	totalPoints string // ex: 2,971,338
	solved      string // ex: 5,468
	successRate string // ex: 99.9% or 100%
	avgTime     string // ex: 2:37
	attempted   int    // derived: sum of per-puzzle 'played' counts
}

func (client *Client) getLifetimeData(
	ctx context.Context,
	uid int,
) (rawLifetimeData, error) {
	doc, err := client.get(ctx, fmt.Sprintf(playerProfileURL, uid))
	if err != nil {
		return rawLifetimeData{}, err
	}

	data, err := extractLifetimeData(doc)
	if err != nil {
		return rawLifetimeData{}, fmt.Errorf("UID %d: %w", uid, err)
	}

	return data, nil
}

func extractLifetimeData(doc *goquery.Document) (rawLifetimeData, error) {
	var data rawLifetimeData

	table := doc.Find("div#tabs-1 table")
	if table.Length() != 1 {
		return rawLifetimeData{}, errors.New("finding table in div#tabs-1")
	}

	td := table.Find("td").Eq(2)
	if td.Length() != 1 {
		return rawLifetimeData{}, errors.New("finding member since td")
	}
	data.memberSince = strings.TrimSpace(td.Text())

	totals := doc.Find("div.sc-total")
	if totals.Length() == 0 {
		return rawLifetimeData{}, errors.New("no div.sc-total found")
	}

	var err error
	totals.EachWithBreak(func(i int, total *goquery.Selection) bool {
		b := total.Find("b")
		if bs := b.Length(); bs != 1 {
			err = fmt.Errorf("sc-total %d: expected 1 <b>, got %d", i, bs)
			return false
		}

		data.totals = append(data.totals, strings.TrimSpace(b.Text()))
		return true
	})
	if err != nil {
		return rawLifetimeData{}, err
	}

	// Zero div.sc-stats is still valid; player just hasn't played at all (#larper)
	doc.Find("div.sc-stats").EachWithBreak(func(i int, stat *goquery.Selection) bool {
		b := stat.Find("b").Eq(0)
		if b.Length() != 1 {
			err = fmt.Errorf("sc-stats %d: played count not found", i)
			return false
		}

		data.played = append(data.played, strings.TrimSpace(b.Text()))
		return true
	})
	if err != nil {
		return rawLifetimeData{}, err
	}

	return data, nil
}

// Adapts [rawLifetimeData] into [lifetimeData]
func (data rawLifetimeData) Lifetime() (lifetimeData, error) {
	// Expect: totalPoints, solved, successRate, avgTime
	if totals := len(data.totals); totals != 4 {
		return lifetimeData{}, fmt.Errorf("expected 4 totals, got %d", totals)
	}

	attempted := 0
	for _, p := range data.played {
		n, err := parser.Count(p)
		if err != nil {
			return lifetimeData{}, fmt.Errorf("parsing played count %q: %w", p, err)
		}
		attempted += n
	}

	return lifetimeData{
		memberSince: data.memberSince,
		totalPoints: data.totals[0],
		solved:      data.totals[1],
		successRate: data.totals[2],
		avgTime:     data.totals[3],
		attempted:   attempted,
	}, nil
}
