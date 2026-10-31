package scraper

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"nook/internal/models"

	"github.com/PuerkitoBio/goquery"
)

type Stats struct {
	At       time.Time     // when the scrape was initiated
	Duration time.Duration // how long the scrape took
}

type Client struct {
	http    *http.Client  // reusable http client; go supports pooling
	timeout time.Duration // max time allowed per scrape
}

func NewClient() *Client {
	return &Client{
		http: &http.Client{Timeout: 8 * time.Second},
	}
}

// Returns all users on a leaderboard
func (client *Client) ScrapeUsers(
	ctx context.Context,
	lb models.Leaderboard,
) ([]*models.User, *Stats, error) {
	start := time.Now() // non-wall time; no need to convert to site location

	var (
		users []*models.User
		err   error
	)

	if lb == models.MonthlyLeaderboard {
		users, err = client.getMonthlyLeaderboardUsers(ctx)
	} else {
		users, err = client.getGlobalLeaderboardUsers(ctx)
	}

	stats := &Stats{
		At:       start,
		Duration: time.Since(start),
	}
	if err != nil {
		return nil, stats, err
	}

	return users, stats, nil
}

func (client *Client) get(
	ctx context.Context,
	url string,
) (*goquery.Document, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request for %q: %w", url, err)
	}

	resp, err := client.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %q: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %q: status %s", url, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parsing html from %q: %w", url, err)
	}

	return doc, nil
}
