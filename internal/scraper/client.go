package scraper

import (
	"context"
	"time"

	"nook/internal/models"
)

type ScrapeStats struct {
	ScrapedAt  time.Time
	ScrapeTime time.Duration
}

// Returns all current users on a leaderboard
func GetLeaderboardUsers(lb models.Leaderboard) ([]*models.User, *ScrapeStats, error) {
	lbFetch := getMonthlyLeaderboardUsers
	if lb == models.GlobalLeaderboard {
		lbFetch = getGlobalLeaderboardUsers
	}

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	start := time.Now()
	users, err := lbFetch(ctx)
	stats := &ScrapeStats{
		ScrapedAt:  start,
		ScrapeTime: time.Since(start),
	}

	if err != nil {
		return nil, stats, err
	}

	return users, stats, nil
}
