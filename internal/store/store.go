package store

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"nook/internal/models"
	"nook/internal/parser"
	"nook/internal/scraper"
	"nook/internal/sqlcgen"
)

type Store struct {
	pool *pgxpool.Pool    // connection pool
	q    *sqlcgen.Queries // type-safe queries
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool, sqlcgen.New(pool)}
}

// Persist a single scrape cycle
func (st *Store) SaveScrape(
	ctx context.Context,
	lb models.Leaderboard,
	users []*models.User,
	stats *scraper.ScrapeStats,
) error {
	tx, err := st.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := st.q.WithTx(tx)

	scrapeID, err := qtx.InsertScrape(ctx, sqlcgen.InsertScrapeParams{
		Kind:       sqlcgen.Leaderboard((lb.DBKind())),
		ScrapedAt:  pgtype.Timestamptz{Time: stats.ScrapedAt, Valid: true},
		DurationMs: int32(stats.ScrapeTime.Milliseconds()),
		Period:     scrapePeriod(lb, stats.ScrapedAt),
	})
	if err != nil {
		return fmt.Errorf("insert scrape: %w", err)
	}

	sorted := make([]*models.User, len(users))
	copy(sorted, users)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].UID < sorted[j].UID })

	for _, u := range sorted {
		if err := qtx.UpsertUser(ctx, sqlcgen.UpsertUserParams{
			Uid:      int32(u.UID),
			Username: u.Username,
		}); err != nil {
			return fmt.Errorf("upsert user %q: %w", u.Username, err)
		}

		if err := qtx.InsertScrapeEntry(ctx, sqlcgen.InsertScrapeEntryParams{
			ScrapeID:       scrapeID,
			Uid:            int32(u.UID),
			Rank:           int32(u.Rank),
			TotalPoints:    int32(u.TotalPoints),
			Solved:         int32(u.Solved),
			Attempted:      int32(u.Attempted),
			AcceptanceRate: u.AcceptanceRate,
			AvgTimeS:       u.AverageTime,
			AvgPoints:      u.AveragePoints,
			LastActive:     pgtype.Timestamptz{Time: u.LastActive, Valid: true},
		}); err != nil {
			return fmt.Errorf("insert entry %q: %w", u.Username, err)
		}
	}

	return tx.Commit(ctx)
}

func scrapePeriod(lb models.Leaderboard, scrapedAt time.Time) pgtype.Date {
	if lb != models.MonthlyLeaderboard {
		return pgtype.Date{}
	}

	t := scrapedAt.In(parser.Eastern)
	return pgtype.Date{
		Time:  time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC),
		Valid: true,
	}
}

func (st *Store) LastTwoScrapes(
	ctx context.Context,
	lb models.Leaderboard,
) (prev, curr []*models.User, err error) {
	var limit int32 = 2
	scrapes, err := st.q.GetLatestScrapes(ctx, sqlcgen.GetLatestScrapesParams{
		Kind:  sqlcgen.Leaderboard(lb.DBKind()),
		Limit: limit,
	})

	if err != nil {
		return nil, nil, fmt.Errorf(": %w", err)
	}

	load := func(scrapeID int64) ([]*models.User, error) {
		entries, err := st.q.GetEntriesForScrape(ctx, scrapeID)
		if err != nil {
			return nil, fmt.Errorf("get entries for scrape %d: %w", scrapeID, err)
		}

		users := make([]*models.User, 0, len(entries))
		for _, e := range entries {
			users = append(users, &models.User{
				UID:            int(e.Uid),
				Rank:           int(e.Rank),
				Username:       e.Username,
				LastActive:     e.LastActive.Time,
				Solved:         int(e.Solved),
				Attempted:      int(e.Attempted),
				AcceptanceRate: e.AcceptanceRate,
				AverageTime:    e.AvgTimeS,
				AveragePoints:  e.AvgPoints,
				TotalPoints:    int(e.TotalPoints),
			})
		}

		return users, nil
	}

	totalScrapes := len(scrapes)

	// Process latest scrape (i.e., most recent)
	if totalScrapes > 0 {
		if curr, err = load(scrapes[0].ScrapeID); err != nil {
			return nil, nil, err
		}
	}

	// Process second latest scrape
	if totalScrapes > 1 {
		if prev, err = load(scrapes[1].ScrapeID); err != nil {
			return nil, nil, err
		}
	}

	return prev, curr, nil
}
