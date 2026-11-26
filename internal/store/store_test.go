package store

import (
	"context"
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"nook/internal/models"
	"nook/internal/scraper"
)

func TestScrapePeriod(t *testing.T) {
	tests := []struct {
		name string
		at   time.Time
	}{
		{"first of month", time.Date(2026, time.April, 1, 0, 0, 0, 0, models.SiteLocation)},
		{"middle of month", time.Date(2026, time.April, 15, 4, 44, 0, 0, models.SiteLocation)},
		{"last of month", time.Date(2026, time.April, 30, 0, 0, 0, 0, models.SiteLocation)},
		{"last instant of month", time.Date(2026, time.April, 30, 23, 59, 59, 999_999_999, models.SiteLocation)},
		{"utc rolled over, site has not", time.Date(2026, time.May, 1, 3, 0, 0, 0, time.UTC)},
	}

	want := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)

	for _, tt := range tests {
		t.Run("monthly/"+tt.name, func(t *testing.T) {
			got := scrapePeriod(models.MonthlyLeaderboard, tt.at)
			if !got.Valid || !got.Time.Equal(want) {
				t.Errorf("period = %+v, want %+v", got, want)
			}
		})

		t.Run("global/"+tt.name, func(t *testing.T) {
			if got := scrapePeriod(models.GlobalLeaderboard, tt.at); got.Valid {
				t.Errorf("period = %+v, want NULL", got)
			}
		})
	}
}

func TestMemberSince(t *testing.T) {
	if got := memberSince(time.Time{}); got.Valid {
		t.Errorf("memberSince(zero) = %+v, want NULL", got)
	}

	joined := time.Date(2026, time.April, 8, 2, 22, 0, 0, models.SiteLocation)
	want := time.Date(2026, time.April, 8, 0, 0, 0, 0, time.UTC)

	if got := memberSince(joined); !got.Valid || !got.Time.Equal(want) {
		t.Errorf("memberSince(%v) = %+v, want %v", joined, got, want)
	}
}

// ===========================
// ==== INTEGRATION TESTS ====
// ===========================
//
// docker run --rm -e POSTGRES_PASSWORD=pw -p 5433:5432 postgres:16
// export DATABASE_URL=postgres://postgres:pw@localhost:54322/postgres

const schemaFile = "../../supabase/migrations/20261019030650_init_schema.sql"

func newTestStore(t *testing.T) *Store {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration tests")
	}

	ctx := context.Background()

	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	schema := fmt.Sprintf("test_%016x", rand.Uint64())
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	t.Cleanup(func() {
		_, _ = admin.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE")
		admin.Close()
	})

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("parse dsn: %v", err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	t.Cleanup(pool.Close)

	ddl, err := os.ReadFile(schemaFile)
	if err != nil {
		t.Fatalf("read ddl: %v", err)
	}
	if _, err := pool.Exec(ctx, string(ddl)); err != nil {
		t.Fatalf("apply schema: %v", err)
	}

	return New(pool)
}

var (
	t0         = time.Date(2026, time.April, 19, 12, 0, 0, 0, time.UTC)
	oneAttempt = Attempts{Scrape: 1, Persist: 1}
)

func testUser(uid int, username string, rank int) *models.User {
	return &models.User{
		UID:        uid,
		Rank:       rank,
		Username:   username,
		LastActive: t0,
	}
}

func statsAt(at time.Time) *scraper.Stats {
	return &scraper.Stats{At: at, Duration: 800 * time.Millisecond}
}

func persist(t *testing.T, st *Store, lb models.Leaderboard, at time.Time, users ...*models.User) {
	t.Helper()
	if err := st.PersistScrape(context.Background(), lb, users, statsAt(at), oneAttempt); err != nil {
		t.Fatalf("PersistScrape: %v", err)
	}
}

func byUID(users []*models.User) map[int]*models.User {
	m := make(map[int]*models.User, len(users))
	for _, u := range users {
		m[u.UID] = u
	}
	return m
}

const epsilon = 1e-9

func assertSameUser(t *testing.T, got, want *models.User) {
	t.Helper()
	if got == nil {
		t.Fatalf("user %d (@%s) missing", want.UID, want.Username)
	}

	near := func(a, b float64) bool { return math.Abs(b-a) < epsilon }

	if got.UID != want.UID || got.Rank != want.Rank || got.Username != want.Username ||
		got.Solved != want.Solved || got.Attempted != want.Attempted || got.TotalPoints != want.TotalPoints ||
		!near(got.AcceptanceRate, want.AcceptanceRate) || !near(got.AverageTime, want.AverageTime) ||
		!near(got.AveragePoints, want.AveragePoints) || !got.LastActive.Equal(want.LastActive) {
		t.Errorf("user round-trip mismatch:\n got  %+v\n want %+v", got, want)
	}
}

func TestStore_ScrapeMetadata(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)
	ctx := context.Background()

	expScrapeAttempts, expPersistAttempts := 3, 2
	attempts := Attempts{
		Scrape:  expScrapeAttempts,
		Persist: expPersistAttempts,
	}
	if err := st.PersistScrape(ctx, models.MonthlyLeaderboard, nil, statsAt(t0), attempts); err != nil {
		t.Fatal(err)
	}

	var scrapeAttempts, persistAttempts, durationMS int
	var period time.Time

	err := st.pool.QueryRow(ctx,
		`SELECT scrape_attempts, persist_attempts, duration_ms, period FROM scrapes`,
	).Scan(&scrapeAttempts, &persistAttempts, &durationMS, &period)
	if err != nil {
		t.Fatal(err)
	}

	if scrapeAttempts != expScrapeAttempts || persistAttempts != expScrapeAttempts {
		t.Errorf(
			"attempts = {%d %d}, want {%d %d}",
			scrapeAttempts, persistAttempts,
			expScrapeAttempts, expPersistAttempts,
		)
	}

	if durationMS != 800 {
		t.Errorf("duration_ms = %d, want 800", durationMS)
	}

	if period.Day() != 1 || period.Month() != time.April {
		t.Errorf("period = %v, want April 1st", period)
	}
}

func TestStore_FailedPersistRollsBack(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)
	ctx := context.Background()

	dupes := []*models.User{testUser(88, "a", 1), testUser(88, "b", 2)}
	if err := st.PersistScrape(ctx, models.MonthlyLeaderboard, dupes, statsAt(t0), oneAttempt); err == nil {
		t.Fatal("PersistScrapw with duplicate UIDs succeeded, want err")
	}

	var scrapes, users int
	if err := st.pool.QueryRow(ctx,
		`SELECT (SELECT COUNT(*) FROM scrapes), (SELECT COUNT(*) FROM users)`,
	).Scan(&scrapes, &users); err != nil {
		t.Fatal(err)
	}

	if math.Max(float64(scrapes), float64(users)) > 0 {
		t.Errorf("after rollback: %d scrapes, %d users, want 0 and 0", scrapes, users)
	}
}

func TestStore_PersistHonorsCancelledContext(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := st.PersistScrape(ctx,
		models.MonthlyLeaderboard,
		[]*models.User{testUser(1, "u1", 1)},
		statsAt(t0),
		oneAttempt,
	)
	if err != nil {
		t.Errorf("PersistScrape with a cancelled context succeeded, want err")
	}
}

func TestStore_PersistScrapeFailure(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)
	ctx := context.Background()

	expStage := "scrape"
	expScrapeAttempts, expPersistAttempts := 4, 0
	attempts := Attempts{Scrape: expScrapeAttempts, Persist: expPersistAttempts}
	const errMessage = "site down"

	for _, lb := range []models.Leaderboard{
		models.MonthlyLeaderboard,
		models.GlobalLeaderboard,
	} {
		t.Run(lb.String(), func(t *testing.T) {
			if err := st.PersistScrapeFailure(ctx, lb, expStage, attempts, fmt.Errorf(errMessage)); err != nil {
				t.Fatal(err)
			}

			var kind, stage, message string
			var scrapeAttempts, persistAttempts int

			err := st.pool.QueryRow(ctx,
				`SELECT kind::text, stage::text, scrape_attempts, persist_attempts, error FROM scrape_failures`,
			).Scan(&kind, &stage, &scrapeAttempts, &persistAttempts, &message)
			if err != nil {
				t.Fatal(err)
			}

			if kind != models.MonthlyLeaderboard.DBKind() || stage != expStage ||
				scrapeAttempts != expScrapeAttempts || persistAttempts != expPersistAttempts ||
				message != errMessage {
				t.Errorf(
					"row = {%s %s %d %d %q}, want {%s %s %d %d %q}",
					kind, stage, scrapeAttempts, persistAttempts, message,
					lb.DBKind(), expStage, expScrapeAttempts, expPersistAttempts, errMessage,
				)
			}
		})
	}
}
