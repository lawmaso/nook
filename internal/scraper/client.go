package scraper

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"nook/internal/models"

	"github.com/PuerkitoBio/goquery"
)

type ScrapeStats struct {
	ScrapedAt  time.Time
	ScrapeTime time.Duration
}

type Client struct {
	http    *http.Client
	timeout time.Duration
}

func NewClient() *Client {
	return &Client{
		http: &http.Client{Timeout: 8 * time.Second}, // per request
	}
}

// Returns all users on a leaderboard
func (client *Client) GetLeaderboardUsers(
	ctx context.Context,
	lb models.Leaderboard,
) ([]*models.User, *ScrapeStats, error) {
	start := time.Now()

	var (
		users []*models.User
		err   error
	)

	if lb == models.MonthlyLeaderboard {
		users, err = client.getMonthlyLeaderboardUsers(ctx)
	} else {
		users, err = client.getGlobalLeaderboardUsers(ctx)
	}

	stats := &ScrapeStats{
		ScrapedAt:  start,
		ScrapeTime: time.Since(start),
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
