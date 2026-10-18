package cache

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"

	"nook/internal/models"
)

// Reads REFRESH_INTERVAL (seconds) from the environment
func ParseRefreshInterval() time.Duration {
	const defaultInterval = 20 * time.Second

	raw := os.Getenv("REFRESH_INTERVAL")
	// fmt.Println(raw)
	if raw == "" {
		return defaultInterval
	}

	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds <= 0 {
		slog.Warn("invalid REFRESH_INTERVAL, using default",
			"value", raw, "default", defaultInterval,
		)
		return defaultInterval
	}

	return time.Duration(seconds) * time.Second
}

// Launches a background goroutine loop that refreshes each leaderboard
// on a fixed interval until the context is cancelled
func (cache *LeaderboardCache) LaunchRefresher(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Prime the cache immediately on launch
		cache.refreshAll()

		for {
			select {
			case <-ctx.Done():
				slog.Info("refresher halted")
				return
			case <-ticker.C:
				cache.refreshAll()
			}
		}
	}()
}

// Refreshes a cached leaderboard with retry logic
func (cache *LeaderboardCache) refreshWithRetry(refresh func() error, attempts int, backoff time.Duration) error {
	var err error
	for i := range attempts {
		if err = refresh(); err == nil {
			return nil
		}

		// Simple linear backoff between refresh attempts
		if i < attempts-1 {
			time.Sleep(backoff * time.Duration(i+1))
		}
	}

	return err
}

// Refreshes both leaderboards
func (cache *LeaderboardCache) refreshAll() error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error

	leaderboards := []models.Leaderboard{models.MonthlyLeaderboard, models.GlobalLeaderboard}

	// Refresh both leaderboards concurrently
	for _, lb := range leaderboards {
		wg.Go(func() {
			// Attempt to refresh 3 times with a 10 second backoff on each attempt
			err := cache.refreshWithRetry(func() error { return cache.Refresh(lb) }, 3, 10*time.Second)

			if err != nil {
				slog.Error("refresh failed after retries", "leaderboard", lb.String(), "error", err)

				mu.Lock()
				errs = append(errs, fmt.Errorf("%s: %w", lb, err))
				mu.Unlock()
				return
			}

			if cache.OnRefreshed != nil {
				snapshot, snapErr := cache.GetCachedLeaderboard(lb)
				if snapErr != nil {
					slog.Error("failed to read back cache after refresh",
						"leaderboard", lb.String(), "error", snapErr)
				}
				cache.OnRefreshed(lb, snapshot.Curr)
			}
		})
	}

	wg.Wait()
	return errors.Join(errs...)
}
