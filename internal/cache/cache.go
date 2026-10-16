package cache

import (
	"fmt"
	"sync"

	"nook/internal/models"
	"nook/internal/scraper"
)

// A cached leaderboard
type CachedLeaderboard struct {
	Prev  []*models.User
	Curr  []*models.User
	Stats *scraper.ScrapeStats
}

// The main interface for the leaderboard cache
type LeaderboardCache struct {
	mu                 sync.RWMutex
	monthlyLeaderboard CachedLeaderboard
	globalLeaderboard  CachedLeaderboard
	OnRefreshed        func(lb models.Leaderboard, curr []*models.User)
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
	// NOTE: Allow the scrape to occur outside the lock since
	// we don't want the network call holding a write lock (and blocking readers)
	users, stats, err := scraper.GetLeaderboardUsers(lb)
	if err != nil {
		return fmt.Errorf(
			"cache refresh failed",
		)
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

	// Store the passed data into the cache
	cachedLeaderboard.Prev = prev
	cachedLeaderboard.Curr = curr
}
