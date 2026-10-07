package store

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"

	"nook/internal/models"
	"nook/internal/sqlcgen"
)

type Store struct {
	pool *pgxpool.Pool
	q    *sqlcgen.Queries

	// Serializes all writes
	writeMu sync.Mutex
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool, q: sqlcgen.New(pool)}
}

// Persists one refresh cycle: a scrape row + every user
// entry for it (in a single transaction).
func (s *Store) SaveScrape(ctx context.Context, lb models.Leaderboard, users []*models.User) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.q.WithTx(tx)

	scrapeId, err := qtx.InsertScrape(ctx, lb.DBKind())
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

		data, err := json.Marshal(u)
		if err != nil {
			return fmt.Errorf("marshal user %q: %w", u.Username, err)
		}

		if err := qtx.InsertEntry(ctx, sqlcgen.InsertEntryParams{
			ScrapeID: scrapeId,
			Uid:      int32(u.UID),
			Data:     data,
		}); err != nil {
			return fmt.Errorf("insert entry %q: %w", u.Username, err)
		}
	}

	return tx.Commit(ctx)
}

// Returns [curr, prev] (most recent first) for a given leaderboard.
// Used once at startup to warm the in-memory cache so a restart doesn't
// serve empty datauntil the next refresh tick.
func (s *Store) LatestTwo(ctx context.Context, lb models.Leaderboard) (curr, prev []*models.User, err error) {
	scrapes, err := s.q.GetLatestScrapes(ctx, sqlcgen.GetLatestScrapesParams{
		LeaderboardKind: lb.DBKind(),
		Limit:           2,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("get latest scrapes: %w", err)
	}

	load := func(scrapeID int64) ([]*models.User, error) {
		entries, err := s.q.GetEntriesForScrape(ctx, scrapeID)
		if err != nil {
			return nil, fmt.Errorf("get entries for scrape %d: %w", scrapeID, err)
		}
		users := make([]*models.User, 0, len(entries))
		for _, e := range entries {
			var u models.User
			if err := json.Unmarshal(e.Data, &u); err != nil {
				return nil, fmt.Errorf("unmarshal entry %d: %w", e.Uid, err)
			}
			users = append(users, &u)
		}
		return users, nil
	}

	if len(scrapes) > 0 {
		if curr, err = load(scrapes[0].ScrapeID); err != nil {
			return nil, nil, err
		}
	}
	if len(scrapes) > 1 {
		if prev, err = load(scrapes[1].ScrapeID); err != nil {
			return nil, nil, err
		}
	}
	return curr, prev, nil
}
