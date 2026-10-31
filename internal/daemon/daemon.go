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
	ScrapeUsers(
		ctx context.Context,
		lb models.Leaderboard,
	) ([]*models.User, *scraper.Stats, error)
}

type Store interface {
	PersistScrape(
		ctx context.Context,
		lb models.Leaderboard,
		users []*models.User,
		stats *scraper.Stats,
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
	sc Scraper,
	st Store,
	cfg Config,
	log *slog.Logger,
) *Daemon {
	return &Daemon{
		scraper: sc,
		store:   st,
		cfg:     cfg,
		log:     log,
	}
}

func (d *Daemon) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	for _, job := range d.cfg.Jobs {
		wg.Go(func() { d.runJob(ctx, job) })
	}
	wg.Wait()
	return ctx.Err()
}

func (d *Daemon) runJob(ctx context.Context, job Job) {
	t := time.NewTimer(0) // run first cycle immediately
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			d.cycle(ctx, job)
			t.Reset(job.Interval)
		}
	}
}

func (d *Daemon) cycle(ctx context.Context, job Job) {
	kind := job.Leaderboard.DBKind()
	log := d.log.With("leaderboard", kind)

	var (
		users []*models.User
		stats *scraper.Stats
	)

	// Scrape leaderboard users (with retry)
	err := retry(ctx, d.cfg.Backoff, log, "scrape", func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, job.Timeout)
		defer cancel()

		var err error
		users, stats, err = d.scraper.ScrapeUsers(
			ctx, job.Leaderboard,
		)
		return err
	})
	if err != nil {
		log.Error("cycle failed", "err", err)
		return
	}

	// Persist scrape using the store (with retry)
	err = retry(ctx, d.cfg.Backoff, log, "persist", func(ctx context.Context) error {
		return d.store.PersistScrape(ctx, job.Leaderboard, users, stats)
	})
	if err != nil {
		log.Error("cycle failed", "err", err)
		return
	}

	log.Info(
		"cycle complete", "users", len(users),
		"scrape_time", stats.Duration,
	)
}
