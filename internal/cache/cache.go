package cache

import (
	"context"
	"fmt"
	"sync"

	"nook/internal/models"
	"nook/internal/scraper"
)

// A cached leaderboard
type CachedLeaderboard struct {
	Prev  []*models.User
	Curr  []*models.User
	Stats *scraper.Stats
}

// The main interface for the leaderboard cache
type LeaderboardCache struct {
	scraper            *scraper.Client
	mu                 sync.RWMutex
	monthlyLeaderboard CachedLeaderboard
	globalLeaderboard  CachedLeaderboard
	OnRefreshed        func(lb models.Leaderboard, curr []*models.User)
}

func New(sc *scraper.Client) *LeaderboardCache {
	return &LeaderboardCache{scraper: sc}
}

// Helper that returns a reference to a cached leaderboard
func getCachedLeaderboard(cache *LeaderboardCache, lb models.Leaderboard) (*CachedLeaderboard, error) {
	switch lb {
	case models.MonthlyLeaderboard:
		return &cache.monthlyLeaderboard, nil
	case models.GlobalLeaderboard:
		return &cache.globalLeaderboard, nil
	default:
		return nil, fmt.Errorf(
			"unknown leaderboard type %q", lb,
		)
	}
}

// Returns a cached leaderboard
func (cache *LeaderboardCache) GetCachedLeaderboard(lb models.Leaderboard) (CachedLeaderboard, error) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	cachedLeaderboard, err := getCachedLeaderboard(cache, lb)
	if err != nil {
		return CachedLeaderboard{}, err
	}

	return *cachedLeaderboard, nil
}

// Refreshes a cached leaderboard
func (cache *LeaderboardCache) Refresh(lb models.Leaderboard) error {
	ctx := context.Background()

	// NOTE: Allow the scrape to occur outside the lock since
	// we don't want the network call holding a write lock (and blocking readers)
	users, stats, err := cache.scraper.ScrapeUsers(ctx, lb)
	if err != nil {
		return fmt.Errorf("cache refresh failed")
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	cachedLeaderboard, err := getCachedLeaderboard(cache, lb)
	if err != nil {
		return err
	}

	cachedLeaderboard.Prev = cachedLeaderboard.Curr
	cachedLeaderboard.Curr = users
	cachedLeaderboard.Stats = stats

	return nil
}

// Seeds the initial in-memory state for a leaderboard
func (cache *LeaderboardCache) Seed(lb models.Leaderboard, curr, prev []*models.User) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	cachedLeaderboard, err := getCachedLeaderboard(cache, lb)
	if err != nil {
		return
	}

	// Store the passed data in the cache
	cachedLeaderboard.Prev = prev
	cachedLeaderboard.Curr = curr
}
