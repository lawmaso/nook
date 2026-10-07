package scraper

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"

	"nook/internal/models"
)

const (
	globalLeaderboardURL  = "https://starbattle.puzzlebaron.com/halloffame.php"
	monthlyLeaderboardURL = "https://starbattle.puzzlebaron.com/contest1.php"
	playerProfileURL      = "https://starbattle.puzzlebaron.com/profile.php?u=%s"
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

// Extracts and returns the text content from a leaderboard table.
//
// Processes rows concurrently with goroutines, writing each to a preallocated
// slice to preserve rank order without explicit synchronization overhead.
func getLeaderboardText(lb models.Leaderboard) ([]rawLeaderboardRow, error) {
	url := monthlyLeaderboardURL
	if lb == models.GlobalLeaderboard {
		url = globalLeaderboardURL
	}

	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching %q: %w", url, err)
	}
	defer response.Body.Close()

	// Parse response body
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, fmt.Errorf("parsing response body")
	}

	// Find all leaderboard table rows
	rows := doc.Find("table.winners tbody tr")

	// Preallocated so each goroutine can write directly to data[rowIndex-1],
	// preservering rank order without relying on completion order
	// No mutex/channel needed: concurrent writes to distinct slice indices are
	// race-free since each goroutine only touches its own memory location (&data[rowIndex-1])
	data := make([]rawLeaderboardRow, rows.Length()-1)
	var wg sync.WaitGroup

	// Concurrently process each row
	rows.Each(func(rowIndex int, row *goquery.Selection) {
		// Skip header row
		if rowIndex == 0 {
			return
		}

		// Process this row in it's own goroutine
		wg.Go(func() {
			// Extract the column text of this row
			row.Find("td").Each(func(colIndex int, td *goquery.Selection) {
				data[rowIndex-1].cols = append(
					data[rowIndex-1].cols,
					strings.TrimSpace(td.Text()),
				)

				// Extract the profile link (if it exists)
				if link := td.Find("a"); link.Length() != 0 {
					if href, exists := link.Attr("href"); exists {
						data[rowIndex-1].profileHref = href
					}
				}
			})
		})
	})

	// Block until all goroutines finish
	wg.Wait()

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

type rawLifetimeRow struct {
	cols []string
}

type lifetimeRow struct {
	memberSince string
	totalPoints string
	attempted   string
	solved      string
	successRate string
	avgPoints   string
	avgTime     string
}

func getLifetimeStatsText(uid int) (rawLifetimeRow, error) {
	profileURL := fmt.Sprintf(playerProfileURL, strconv.Itoa(uid))
	response, err := http.Get(profileURL)

	if err != nil {
		return rawLifetimeRow{}, fmt.Errorf("GET %q: %w", profileURL, err)
	}
	defer response.Body.Close()

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return rawLifetimeRow{}, fmt.Errorf("")
	}

	// Find the lifetime statistic table
	table := doc.Find("div#tabs-1 table")
	if table.Length() != 1 {
		return rawLifetimeRow{}, fmt.Errorf(
			"finding lifetime table for UID %d", uid,
		)
	}

	// Extract rows
	rows := table.Find("tbody tr")
	if rows.Length() == 0 {
		return rawLifetimeRow{}, fmt.Errorf(
			"finding lifetime table rows for UID %d", uid,
		)
	}

	// Extract table data
	var data rawLifetimeRow
	rows.Each(func(index int, row *goquery.Selection) {
		// Skip the header
		if index == 0 {
			return
		}

		// Process columns
		cols := row.Find("td")
		if cols.Length() == 2 {
			columnVal := cols.Eq(1).Text()

			// Only append the actual column values
			data.cols = append(data.cols, strings.TrimSpace(columnVal))
		}
	})

	// Get average time (stored in the last scorecard table)
	table = doc.Find("table.scorecard_table").Eq(-1)
	if table.Length() == 0 {
		return rawLifetimeRow{}, fmt.Errorf(
			"finding last scorecard table for UID %d", uid,
		)
	}

	timeRow := table.Find("tbody tr td").Last()
	data.cols = append(data.cols, strings.TrimSpace(timeRow.Text()))

	return data, nil
}

// Adapts [rawLifetimeRow] into [lifetimeRow]
func (r rawLifetimeRow) Lifetime() (lifetimeRow, error) {
	expectedCols := 7
	if len(r.cols) != expectedCols {
		return lifetimeRow{}, fmt.Errorf(
			"expected %d columns, got %d", expectedCols, len(r.cols),
		)
	}

	return lifetimeRow{
		memberSince: r.cols[0],
		totalPoints: r.cols[1],
		attempted:   r.cols[2],
		solved:      r.cols[3],
		successRate: r.cols[4],
		avgPoints:   r.cols[5],
		avgTime:     r.cols[6],
	}, nil
}
