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

// Is this even needed?
// The new site restructing forces us out of
// the generalization of parsing rows of the lifetime table
// (which no longer exists)
type rawLifetimeData struct {
	data []string
}

type lifetimeData struct {
	memberSince string // ex: April 15, 2026
	totalPoints string // ex: 2,971,338
	attempted   string // ex: 5,468
	solved      string // ex: 5,468
	successRate string // ex: 99.9% or 100%
	avgTime     string // ex: 2:37
}

func getLifetimeData(uid int) (rawLifetimeData, error) {
	profileURL := fmt.Sprintf(playerProfileURL, strconv.Itoa(uid))
	response, err := http.Get(profileURL)

	if err != nil {
		return rawLifetimeData{}, fmt.Errorf("GET %q: %w", profileURL, err)
	}
	defer response.Body.Close()

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return rawLifetimeData{}, fmt.Errorf("creating document: %w", err)
	}

	// Result raw lifetime data
	data := rawLifetimeData{}

	// Find member since text
	table := doc.Find("div#tabs-1 table")
	if table.Length() != 1 {
		return rawLifetimeData{}, fmt.Errorf("finding table in div#tabs-1")
	}

	td := table.Find("td").Eq(2)
	if td.Length() != 1 {
		return rawLifetimeData{}, fmt.Errorf("finding member since td")
	}

	// Append the trimmed member since text
	data.data = append(data.data, strings.TrimSpace(td.Text()))

	// Find the aggregates binded to the profile
	totals := doc.Find("div.sc-total")
	if totals.Length() == 0 {
		return rawLifetimeData{}, fmt.Errorf(".sc-total divs not found")
	}

	// Append each string total to data columns
	totals.Each(func(_ int, total *goquery.Selection) {
		strData := total.Find("b")

		// Ignore malformed aggregations; expected length
		// of calling function should handle the error
		if strData.Length() != 1 {
			return
		}
		data.data = append(data.data, strings.TrimSpace(strData.Text()))
	})

	// Due to site UI restructuring, we have to aggregate
	// the attempts of each puzzle variation...
	stats := doc.Find("div.sc-stats")
	if stats.Length() == 0 {
		return rawLifetimeData{}, fmt.Errorf(".sc-stats divs not found")
	}

	// Append the played aggregation for each puzzle type.
	// Does not concern itself with casting to int, then casting
	// back to string since that wouldn't be clean; just get the text data.
	stats.Each(func(_ int, stat *goquery.Selection) {
		playedData := stat.Find("b").Eq(0)

		if playedData.Length() != 1 {
			return
		}

		data.data = append(data.data, strings.TrimSpace(playedData.Text()))
	})

	return data, nil
}

// Adapts [rawLifetimeData] into [lifetimeData]
func (data rawLifetimeData) Lifetime() (lifetimeData, error) {
	// Expect at least 5 since the aggregations
	// invariantly exist on each user profile whether
	// or not they've played any puzzles
	expectedLength := 5
	length := len(data.data)

	if length < expectedLength {
		return lifetimeData{}, fmt.Errorf(
			"got %d entries, expected %d", length, expectedLength,
		)
	}

	var attempted int
	if length == 5 {
		// Edge case: player hasn't played any puzzles
		attempted = 0
	} else {
		// Otherwise: length > 5, so at least 1 puzzle
		// variation has been played by the user

		// Aggregate attempts so wrapper is cleaner and
		// we don't have to account for each individual puzzle type
		for _, attempts := range data.data[5:] {
			att, err := strconv.Atoi(strings.ReplaceAll(attempts, ",", ""))
			if err != nil {
				return lifetimeData{}, fmt.Errorf(
					"converting attempt %q to integer", attempts,
				)
			}
			attempted += att
		}
	}

	return lifetimeData{
		memberSince: data.data[0],
		totalPoints: data.data[1],
		attempted:   strconv.Itoa(attempted),
		solved:      data.data[2],
		successRate: data.data[3],
		avgTime:     data.data[4],
	}, nil
}
