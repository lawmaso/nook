package daemon

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"nook/internal/models"
	"nook/internal/scraper"
)

type Scraper interface {
	GetLeaderboardUsers(
		ctx context.Context,
		lb models.Leaderboard,
	) ([]*models.User, *scraper.ScrapeStats, error)
}

type Store interface {
	SaveScrape(
		ctx context.Context,
		lb models.Leaderboard,
		users []*models.User,
		stats *scraper.ScrapeStats,
	) error
}

type Job struct {
	Leaderboard models.Leaderboard
	Interval    time.Duration // time between scrapes
	Timeout     time.Duration // time per scrape attempt
}

type Config struct {
	Jobs    []Job
	Backoff Backoff
}

func DefaultConfig() Config {
	return Config{
		Jobs: []Job{
			{ // Monthly leaderboard scrape job
				Leaderboard: models.MonthlyLeaderboard,
				Interval:    15 * time.Minute,
				Timeout:     30 * time.Second,
			},
			{ // Global leaderboard scrape job
				Leaderboard: models.GlobalLeaderboard,
				Interval:    1 * time.Hour,
				Timeout:     1 * time.Minute,
			},
		},
		Backoff: Backoff{
			Initial:     1 * time.Second,
			Max:         1 * time.Minute,
			Factor:      2,
			MaxAttempts: 4,
		},
	}
}

type Daemon struct {
	scraper Scraper
	store   Store
	cfg     Config
	log     *slog.Logger
}

func New(
	scraper Scraper,
	store Store,
	cfg Config,
	log *slog.Logger,
) *Daemon {
	return &Daemon{
		scraper: scraper,
		store:   store,
		cfg:     cfg,
		log:     log,
	}
}

func (daemon *Daemon) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	for _, job := range daemon.cfg.Jobs {
		wg.Go(func() { daemon.runJob(ctx, job) })
	}
	wg.Wait()
	return ctx.Err()
}

func (daemon *Daemon) runJob(ctx context.Context, job Job) {
	t := time.NewTimer(0) // run first cycle immediately
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			daemon.cycle(ctx, job)
			t.Reset(job.Interval)
		}
	}
}

func (daemon *Daemon) cycle(ctx context.Context, job Job) {
	kind := job.Leaderboard.DBKind()
	log := daemon.log.With("leaderboard", kind)

	var (
		users []*models.User
		stats *scraper.ScrapeStats
	)

	// Scrape leaderboard users (with retry)
	err := retry(ctx, daemon.cfg.Backoff, log, "scrape", func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, job.Timeout)
		defer cancel()

		var err error
		users, stats, err = daemon.scraper.GetLeaderboardUsers(
			ctx, job.Leaderboard,
		)
		return err
	})
	if err != nil {
		log.Error("cycle failed", "err", err)
		return
	}

	// Persist scrape using the store (with retry)
	err = retry(ctx, daemon.cfg.Backoff, log, "persist", func(ctx context.Context) error {
		return daemon.store.SaveScrape(ctx, job.Leaderboard, users, stats)
	})
	if err != nil {
		log.Error("cycle failed", "err", err)
		return
	}

	log.Info(
		"cycle complete", "users", len(users),
		"scrape_time", stats.ScrapeTime,
	)
}
