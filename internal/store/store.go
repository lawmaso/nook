package store

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"nook/internal/models"
	"nook/internal/scraper"
	"nook/internal/sqlcgen"
)

type Store struct {
	pool *pgxpool.Pool    // connection pool
	q    *sqlcgen.Queries // type-safe queries
}

type Attempts struct{ Scrape, Persist int }

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool, sqlcgen.New(pool)}
}

// Persists a single leaderboard scrape
func (st *Store) PersistScrape(
	ctx context.Context,
	lb models.Leaderboard,
	users []*models.User,
	stats *scraper.Stats,
	attempts Attempts,
) error {
	// NOTE: validate before opening a transaction since
	// a duplicate UID violates scrape_entries primary-key constraint
	// halfway through the COPYFROM
	sorted := make([]*models.User, len(users))
	copy(sorted, users)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].UID < sorted[j].UID })

	for i := 1; i < len(users); i++ {
		if sorted[i].UID == sorted[i-1].UID {
			return fmt.Errorf("duplicate uid: %d", sorted[i].UID)
		}
	}

	tx, err := st.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := st.q.WithTx(tx)

	scrapeID, err := qtx.InsertScrape(ctx, sqlcgen.InsertScrapeParams{
		Kind:            sqlcgen.Leaderboard((lb.DBKind())),
		Period:          scrapePeriod(lb, stats.At),
		ScrapedAt:       pgtype.Timestamptz{Time: stats.At, Valid: true},
		DurationMs:      int32(stats.Duration.Milliseconds()),
		ScrapeAttempts:  int32(attempts.Scrape),
		PersistAttempts: int32(attempts.Persist),
	})
	if err != nil {
		return fmt.Errorf("insert scrape: %w", err)
	}

	if len(sorted) > 0 {
		if err := upsertUsers(ctx, qtx, sorted); err != nil {
			return err
		}
		if err := insertEntries(ctx, qtx, scrapeID, sorted); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func upsertUsers(
	ctx context.Context,
	qtx *sqlcgen.Queries,
	users []*models.User,
) error {
	params := make([]sqlcgen.UpsertUsersParams, len(users))
	for i, u := range users {
		params[i] = sqlcgen.UpsertUsersParams{
			Uid:         int32(u.UID),
			Username:    u.Username,
			MemberSince: memberSince(u.MemberSince),
		}
	}

	var errs []error
	qtx.UpsertUsers(ctx, params).Exec(func(i int, err error) {
		if err != nil {
			errs = append(errs, fmt.Errorf("upsert user %q: %w", users[i].Username, err))
		}
	})

	return errors.Join(errs...)
}

func insertEntries(
	ctx context.Context,
	qtx *sqlcgen.Queries,
	scrapeID int64,
	users []*models.User,
) error {
	params := make([]sqlcgen.InsertScrapeEntriesParams, len(users))
	for i, u := range users {
		params[i] = sqlcgen.InsertScrapeEntriesParams{
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
		}
	}

	n, err := qtx.InsertScrapeEntries(ctx, params)
	if err != nil {
		return fmt.Errorf("copy entries: %w", err)
	}

	if int(n) != len(params) {
		return fmt.Errorf("copy entries: wrote %d of %d rows", n, len(params))
	}

	return nil
}

func (st *Store) PersistScrapeFailure(
	ctx context.Context,
	lb models.Leaderboard,
	stage string,
	attempts Attempts,
	cause error,
) error {
	return st.q.InsertScrapeFailure(ctx, sqlcgen.InsertScrapeFailureParams{
		Kind:            sqlcgen.Leaderboard(lb.DBKind()),
		Stage:           sqlcgen.ScrapeStage(stage),
		FailedAt:        pgtype.Timestamptz{Time: time.Now(), Valid: true},
		ScrapeAttempts:  int32(attempts.Scrape),
		PersistAttempts: int32(attempts.Persist),
		Error:           cause.Error(),
	})
}

func scrapePeriod(lb models.Leaderboard, scrapedAt time.Time) pgtype.Date {
	if lb != models.MonthlyLeaderboard {
		return pgtype.Date{Valid: false} // NULL since false
	}

	return pgtype.Date{
		Time:  models.MonthStart(scrapedAt),
		Valid: true,
	}
}

func memberSince(t time.Time) pgtype.Date {
	if t.IsZero() {
		return pgtype.Date{Valid: false} // NULL since false
	}

	return pgtype.Date{
		Time:  models.DayStart(t),
		Valid: true,
	}
}
