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

// ---------------------------------------------------------------------
// Unit tests: pure helpers, no database needed.
// ---------------------------------------------------------------------

func TestScrapePeriod(t *testing.T) {
	tests := []struct {
		name string
		at   time.Time
	}{
		{"first of month", time.Date(2026, time.April, 1, 0, 0, 0, 0, models.SiteLocation)},
		{"mid month", time.Date(2026, time.April, 15, 12, 0, 0, 0, models.SiteLocation)},
		{"last instant of month", time.Date(2026, time.April, 30, 23, 59, 59, 999_999_999, models.SiteLocation)},
		// Already May in UTC, still April 30 on the site's (Eastern) clock.
		{"utc rolled over, site has not", time.Date(2026, time.May, 1, 3, 0, 0, 0, time.UTC)},
	}

	// Postgres DATE has no zone, so the site's calendar date is encoded as UTC midnight.
	want := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)

	for _, tt := range tests {
		t.Run("monthly/"+tt.name, func(t *testing.T) {
			got := scrapePeriod(models.MonthlyLeaderboard, tt.at)
			if !got.Valid || !got.Time.Equal(want) {
				t.Errorf("period = %+v, want %v", got, want)
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

	joined := time.Date(2026, time.April, 8, 22, 22, 0, 0, models.SiteLocation)
	want := time.Date(2026, time.April, 8, 0, 0, 0, 0, time.UTC)
	if got := memberSince(joined); !got.Valid || !got.Time.Equal(want) {
		t.Errorf("memberSince(%v) = %+v, want %v", joined, got, want)
	}
}

// ---------------------------------------------------------------------
// Integration tests: real Postgres, one throwaway schema per test.
//
//	docker run --rm -e POSTGRES_PASSWORD=pw -p 5433:5432 postgres:16
//	export TEST_DATABASE_URL=postgres://postgres:pw@localhost:5433/postgres
// ---------------------------------------------------------------------

const schemaFile = "../../supabase/migrations/20261019030650_init_schema.sql"

// newTestStore applies the schema into a fresh, uniquely named Postgres schema
// and returns a Store whose connections only see that schema. Because nothing
// is shared between tests, they can run in parallel.
func newTestStore(t *testing.T) *Store {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping database tests")
	}
	ctx := context.Background()

	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	schema := fmt.Sprintf("test_%016x", rand.Uint64()) // random: parallel tests can share a clock tick
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	// Cleanups run LIFO: the test pool below closes before this drops the schema.
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
		t.Fatalf("read schema: %v", err)
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
		UID:            uid,
		Rank:           rank,
		Username:       username,
		LastActive:     t0,
		Solved:         5499,
		Attempted:      5512,
		AcceptanceRate: 99.76,
		AverageTime:    157,
		AveragePoints:  543.41,
		TotalPoints:    2_988_210,
	}
}

func statsAt(at time.Time) *scraper.Stats {
	return &scraper.Stats{At: at, Duration: 800 * time.Millisecond}
}

// persist is PersistScrape with the common arguments filled in.
func persist(t *testing.T, st *Store, lb models.Leaderboard, at time.Time, users ...*models.User) {
	t.Helper()
	if err := st.PersistScrape(context.Background(), lb, users, statsAt(at), oneAttempt); err != nil {
		t.Fatalf("PersistScrape: %v", err)
	}
}

func lastTwo(t *testing.T, st *Store, lb models.Leaderboard) (prev, curr []*models.User) {
	t.Helper()
	prev, curr, err := st.LastTwoScrapes(context.Background(), lb)
	if err != nil {
		t.Fatalf("LastTwoScrapes: %v", err)
	}
	return prev, curr
}

func byUID(users []*models.User) map[int]*models.User {
	m := make(map[int]*models.User, len(users))
	for _, u := range users {
		m[u.UID] = u
	}
	return m
}

func assertSameUser(t *testing.T, got, want *models.User) {
	t.Helper()
	if got == nil {
		t.Fatalf("user %d (%s) missing", want.UID, want.Username)
	}
	near := func(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

	if got.UID != want.UID || got.Rank != want.Rank || got.Username != want.Username ||
		got.Solved != want.Solved || got.Attempted != want.Attempted || got.TotalPoints != want.TotalPoints ||
		!near(got.AcceptanceRate, want.AcceptanceRate) || !near(got.AverageTime, want.AverageTime) ||
		!near(got.AveragePoints, want.AveragePoints) || !got.LastActive.Equal(want.LastActive) {
		t.Errorf("user round-trip mismatch:\n got  %+v\n want %+v", got, want)
	}
}

func TestStore_EmptyDatabase(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)

	if prev, curr := lastTwo(t, st, models.MonthlyLeaderboard); prev != nil || curr != nil {
		t.Errorf("prev=%v curr=%v, want both nil", prev, curr)
	}
}

func TestStore_RoundTrip(t *testing.T) {
	t.Parallel()
	for _, lb := range []models.Leaderboard{models.MonthlyLeaderboard, models.GlobalLeaderboard} {
		t.Run(lb.String(), func(t *testing.T) {
			t.Parallel()
			st := newTestStore(t)
			want := []*models.User{testUser(1, "beeron", 1), testUser(2, "zeroshift", 2)}

			persist(t, st, lb, t0, want...)

			prev, curr := lastTwo(t, st, lb)
			if prev != nil {
				t.Errorf("prev = %v, want nil after one scrape", prev)
			}
			if len(curr) != len(want) {
				t.Fatalf("curr has %d users, want %d", len(curr), len(want))
			}
			got := byUID(curr)
			for _, w := range want {
				assertSameUser(t, got[w.UID], w)
			}
		})
	}
}

// The scrapes row keeps the metadata the daemon passes in: retry counts and
// the period derived from the leaderboard kind.
func TestStore_ScrapeMetadata(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)
	ctx := context.Background()

	att := Attempts{Scrape: 3, Persist: 2}
	if err := st.PersistScrape(ctx, models.MonthlyLeaderboard, nil, statsAt(t0), att); err != nil {
		t.Fatal(err)
	}

	var scrapeAtt, persistAtt, durationMs int32
	var period time.Time
	err := st.pool.QueryRow(ctx,
		`SELECT scrape_attempts, persist_attempts, duration_ms, period FROM scrapes`,
	).Scan(&scrapeAtt, &persistAtt, &durationMs, &period)
	if err != nil {
		t.Fatal(err)
	}
	if scrapeAtt != 3 || persistAtt != 2 {
		t.Errorf("attempts = {%d %d}, want {3 2}", scrapeAtt, persistAtt)
	}
	if durationMs != 800 {
		t.Errorf("duration_ms = %d, want 800", durationMs)
	}
	if period.Day() != 1 || period.Month() != time.April {
		t.Errorf("period = %v, want April 1", period)
	}
}

func TestStore_LastTwoScrapesOrdersByScrapeTime(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)
	lb := models.MonthlyLeaderboard

	// Inserted out of order: ordering must come from scraped_at, not insert order.
	// Distinct uids, since usernames live on the shared users row.
	persist(t, st, lb, t0.Add(time.Hour), testUser(2, "middle", 1))
	persist(t, st, lb, t0, testUser(1, "oldest", 1))
	persist(t, st, lb, t0.Add(2*time.Hour), testUser(3, "newest", 1))

	prev, curr := lastTwo(t, st, lb)
	if len(curr) != 1 || curr[0].Username != "newest" {
		t.Errorf("curr = %+v, want newest", curr)
	}
	if len(prev) != 1 || prev[0].Username != "middle" {
		t.Errorf("prev = %+v, want middle", prev)
	}
}

func TestStore_KindsAreSeparate(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)

	persist(t, st, models.GlobalLeaderboard, t0, testUser(1, "g", 1))
	persist(t, st, models.MonthlyLeaderboard, t0.Add(time.Hour), testUser(2, "m", 1))

	if _, curr := lastTwo(t, st, models.GlobalLeaderboard); len(curr) != 1 || curr[0].UID != 1 {
		t.Errorf("global = %+v, want only uid 1", curr)
	}
	if _, curr := lastTwo(t, st, models.MonthlyLeaderboard); len(curr) != 1 || curr[0].UID != 2 {
		t.Errorf("monthly = %+v, want only uid 2", curr)
	}
}

// A user seen again is upserted (latest username wins), and each scrape keeps
// its own entry for them.
func TestStore_SameUserAcrossScrapes(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)
	lb := models.MonthlyLeaderboard

	persist(t, st, lb, t0, testUser(1, "old_name", 5))
	persist(t, st, lb, t0.Add(time.Hour), testUser(1, "new_name", 2))

	prev, curr := lastTwo(t, st, lb)
	if len(prev) != 1 || prev[0].Rank != 5 || len(curr) != 1 || curr[0].Rank != 2 {
		t.Errorf("prev=%+v curr=%+v, want rank 5 then 2", prev, curr)
	}
	if curr[0].Username != "new_name" || prev[0].Username != "new_name" {
		t.Errorf("usernames = %q/%q, want the latest name on both", prev[0].Username, curr[0].Username)
	}
}

func TestStore_EmptyScrapeIsStillRecorded(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)
	lb := models.MonthlyLeaderboard

	persist(t, st, lb, t0, testUser(1, "u", 1))
	persist(t, st, lb, t0.Add(time.Hour)) // no users

	prev, curr := lastTwo(t, st, lb)
	if len(curr) != 0 || len(prev) != 1 {
		t.Errorf("prev=%+v curr=%+v, want the empty scrape as curr", prev, curr)
	}
}

// Any error mid-transaction must leave no trace: no scrapes row, no users.
func TestStore_FailedPersistRollsBack(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)
	ctx := context.Background()

	dupes := []*models.User{testUser(1, "a", 1), testUser(1, "b", 2)}
	if err := st.PersistScrape(ctx, models.MonthlyLeaderboard, dupes, statsAt(t0), oneAttempt); err == nil {
		t.Fatal("PersistScrape with duplicate uids succeeded, want an error")
	}

	var scrapes, users int
	if err := st.pool.QueryRow(ctx,
		`SELECT (SELECT count(*) FROM scrapes), (SELECT count(*) FROM users)`,
	).Scan(&scrapes, &users); err != nil {
		t.Fatal(err)
	}
	if scrapes != 0 || users != 0 {
		t.Errorf("after rollback: %d scrapes, %d users, want 0 and 0", scrapes, users)
	}
}

func TestStore_PersistHonorsCancelledContext(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := st.PersistScrape(ctx, models.MonthlyLeaderboard, []*models.User{testUser(1, "u", 1)}, statsAt(t0), oneAttempt)
	if err == nil {
		t.Fatal("PersistScrape with a cancelled context succeeded, want an error")
	}
}

func TestStore_PersistScrapeFailure(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)
	ctx := context.Background()

	att := Attempts{Scrape: 4, Persist: 0}
	if err := st.PersistScrapeFailure(ctx, models.GlobalLeaderboard, "scrape", att, fmt.Errorf("site down")); err != nil {
		t.Fatal(err)
	}

	var kind, stage, msg string
	var scrapeAtt, persistAtt int32
	err := st.pool.QueryRow(ctx,
		`SELECT kind::text, stage::text, scrape_attempts, persist_attempts, error FROM scrape_failures`,
	).Scan(&kind, &stage, &scrapeAtt, &persistAtt, &msg)
	if err != nil {
		t.Fatal(err)
	}
	if kind != "global" || stage != "scrape" || scrapeAtt != 4 || persistAtt != 0 || msg != "site down" {
		t.Errorf("row = {%s %s %d %d %q}, want {global scrape 4 0 \"site down\"}", kind, stage, scrapeAtt, persistAtt, msg)
	}
}
