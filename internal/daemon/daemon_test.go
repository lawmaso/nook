package daemon

import (
	"context"
	"errors"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"nook/internal/models"
	"nook/internal/scraper"
	"nook/internal/store"
)

type scrapeFunc func(
	ctx context.Context,
	lb models.Leaderboard,
	n int,
) ([]*models.User, *scraper.Stats, error)

type fakeScraper struct {
	mu    sync.Mutex
	calls map[models.Leaderboard]int // calls[lb] = number of scrape attempts for lb
	fn    scrapeFunc
}

func (f *fakeScraper) ScrapeUsers(
	ctx context.Context,
	lb models.Leaderboard,
) ([]*models.User, *scraper.Stats, error) {
	f.mu.Lock()
	f.calls[lb]++

	n := f.calls[lb]
	f.mu.Unlock()

	return f.fn(ctx, lb, n)
}

// Returns the current number of scrape calls made by the fake scraper
func (f *fakeScraper) Calls(lb models.Leaderboard) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls[lb]
}

// A scrape that is ok
func scrapeOk(_ context.Context, lb models.Leaderboard, n int) ([]*models.User, *scraper.Stats, error) {
	users := []*models.User{{
		Rank:     1,
		UID:      n,
		Username: lb.DBKind(),
	}}
	stats := &scraper.Stats{At: time.Now()}
	return users, stats, nil
}

// A scrape that hangs
func scrapeHang(ctx context.Context, _ models.Leaderboard, _ int) ([]*models.User, *scraper.Stats, error) {
	<-ctx.Done()
	return nil, nil, ctx.Err()
}

func newFakeScraper(fn scrapeFunc) *fakeScraper {
	return &fakeScraper{
		calls: map[models.Leaderboard]int{},
		fn:    fn,
	}
}

// A successful persist in the store
type save struct {
	lb       models.Leaderboard
	users    []*models.User
	attempts store.Attempts
}

// A failed persist in the store
type fail struct {
	lb       models.Leaderboard
	stage    string
	attempts store.Attempts
	cause    error
	at       time.Time
}

type fakeStore struct {
	mu        sync.Mutex
	persistFn func(n int) error // nil err means the persist succeeded
	calls     int
	saves     []save
	fails     []fail
}

func (f *fakeStore) PersistScrape(
	_ context.Context,
	lb models.Leaderboard,
	users []*models.User,
	_ *scraper.Stats,
	attempts store.Attempts,
) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.calls++
	if f.persistFn != nil {
		if err := f.persistFn(f.calls); err != nil {
			return err
		}
	}

	f.saves = append(f.saves, save{lb, users, attempts})
	return nil
}

func (f *fakeStore) PersistScrapeFailure(
	ctx context.Context,
	lb models.Leaderboard,
	stage string,
	attempts store.Attempts,
	cause error,
) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.fails = append(f.fails, fail{lb, stage, attempts, cause, time.Now()})
	return nil
}

func (f *fakeStore) Saved() []save {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]save(nil), f.saves...)
}

func (f *fakeStore) Fails() []fail {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]fail(nil), f.fails...)
}

func newFakeStore(persistFn func(int) error) *fakeStore {
	if persistFn == nil {
		persistFn = nil
	}

	return &fakeStore{
		persistFn: persistFn,
		saves:     make([]save, 0),
		fails:     make([]fail, 0),
	}
}

var (
	monthlyJob = Job{
		Leaderboard: models.MonthlyLeaderboard,
		Interval:    1 * time.Hour,
		Timeout:     30 * time.Second,
	}
	globalJob = Job{
		Leaderboard: models.GlobalLeaderboard,
		Interval:    2 * time.Hour,
		Timeout:     1 * time.Minute,
	}
)

var testBackoff = Backoff{
	Initial:     1 * time.Second,
	Max:         10 * time.Second,
	Factor:      2,
	MaxAttempts: 3,
}

var pastBackoff = 5 * time.Second

func testConfig(testJobs ...Job) Config {
	return Config{
		Jobs:    testJobs,
		Backoff: testBackoff,
	}
}

func startDaemon(scraper Scraper, store Store, cfg Config) (stop func() error) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)

	// run the daemon in a goroutine
	go func() { done <- New(scraper, store, cfg, testLog).Run(ctx) }()

	// return daemon stop func
	return func() error {
		cancel()
		return <-done
	}
}

func mustStop(t *testing.T, stop func() error) {
	t.Helper()
	if err := stop(); !errors.Is(err, context.Canceled) {
		t.Errorf("run returned %v, want context.Canceled", err)
	}
}

func TestDaemon_RunsImmediately(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sc, st := newFakeScraper(scrapeOk), &fakeStore{}
		stop := startDaemon(sc, st, testConfig(monthlyJob))

		synctest.Wait()

		saved := st.Saved()
		if n := len(saved); n != 1 {
			t.Fatalf("saved %d scrapes at t=0, want 1", n)
		}

		if att := saved[0].attempts; att != (store.Attempts{Scrape: 1, Persist: 1}) {
			t.Errorf("attempts = %+v, want {1 1}", att)
		}

		if got := saved[0].users; len(got) != 1 || got[0].Username != models.MonthlyLeaderboard.DBKind() {
			t.Errorf("users = %+v, want exactly what the scraper returned", got)
		}

		mustStop(t, stop)
	})
}

func TestDaemon_RunsOnGivenInterval(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sc, st := newFakeScraper(scrapeOk), &fakeStore{}
		stop := startDaemon(sc, st, testConfig(monthlyJob))

		synctest.Wait()

		// wait until just before the next daemon cycle
		time.Sleep(monthlyJob.Interval - time.Nanosecond)
		synctest.Wait()

		if monthlyN := sc.Calls(models.MonthlyLeaderboard); monthlyN != 1 {
			t.Fatalf("monthly scrapes just before the interval = %d, want 1", monthlyN)
		}

		if globalN := sc.Calls(models.GlobalLeaderboard); globalN != 0 {
			t.Errorf("global scrapes just before the interval = %d, want 0", globalN)
		}

		// let daemon run the next cycle
		time.Sleep(time.Nanosecond)
		synctest.Wait()

		if monthlyN := sc.Calls(models.MonthlyLeaderboard); monthlyN != 2 {
			t.Errorf("monthly scrapes at the interval = %d, want 2", monthlyN)
		}

		if globalN := sc.Calls(models.GlobalLeaderboard); globalN != 0 {
			t.Errorf("global scrapes at the interval = %d, want 0", globalN)
		}

		mustStop(t, stop)
	})
}

func TestDaemon_NonOverlappingIntervals(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		const scrapeDuration = 10 * time.Second

		var starts []time.Time
		var mu sync.Mutex

		sc := newFakeScraper(func(ctx context.Context, lb models.Leaderboard, n int) ([]*models.User, *scraper.Stats, error) {
			mu.Lock()
			starts = append(starts, time.Now())
			mu.Unlock()

			time.Sleep(scrapeDuration)
			return scrapeOk(ctx, lb, n)
		})
		stop := startDaemon(sc, &fakeStore{}, testConfig(monthlyJob))

		time.Sleep(2 * (monthlyJob.Interval + scrapeDuration))
		synctest.Wait()
		mustStop(t, stop)

		if n := len(starts); n < 2 {
			t.Fatalf("got %d scrapes, want at least 2", n)
		}

		wantedGap := monthlyJob.Interval + scrapeDuration
		if gap := starts[1].Sub(starts[0]); gap != wantedGap {
			t.Errorf("gap between scrape starts = %v, want %v", gap, wantedGap)
		}
	})
}

func TestDaemon_StuckJobDoesNotBlockOthers(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sc := newFakeScraper(func(ctx context.Context, lb models.Leaderboard, n int) ([]*models.User, *scraper.Stats, error) {
			if lb == models.GlobalLeaderboard {
				return scrapeHang(ctx, lb, n)
			}

			return scrapeOk(ctx, lb, n)
		})

		st := &fakeStore{}
		stop := startDaemon(sc, st, testConfig(monthlyJob, globalJob))

		synctest.Wait() // global job is hung; monthly should be done

		saved := st.Saved()
		if len(saved) != 1 || saved[0].lb != models.MonthlyLeaderboard {
			t.Errorf("saved = %+v, want one monthly scrape", saved)
		}

		mustStop(t, stop)
	})
}

func TestDaemon_ScrapeRetriesThenSucceeds(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sc := newFakeScraper(func(ctx context.Context, lb models.Leaderboard, n int) ([]*models.User, *scraper.Stats, error) {
			if n < 3 {
				return nil, nil, testErr
			}
			return scrapeOk(ctx, lb, n)
		})
		st := &fakeStore{}
		stop := startDaemon(sc, st, testConfig(monthlyJob))

		time.Sleep(pastBackoff)
		synctest.Wait()

		saved := st.Saved()
		if n := len(saved); n != 1 {
			t.Fatalf("saved %d scrapes, want 1", n)
		}

		wantedAttempts := store.Attempts{Scrape: 3, Persist: 1}
		if att := saved[0].attempts; att != wantedAttempts {
			t.Errorf("attempts = %+v, want %+v", att, wantedAttempts)
		}

		if f := st.Fails(); len(f) != 0 {
			t.Errorf("fails = %+v, want none", f)
		}

		mustStop(t, stop)
	})
}

func TestDaemon_ScrapeGivesUpThenRunsNextCycle(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sc := newFakeScraper(func(ctx context.Context, lb models.Leaderboard, n int) ([]*models.User, *scraper.Stats, error) {
			return nil, nil, testErr
		})
		st := &fakeStore{}
		stop := startDaemon(sc, st, testConfig(monthlyJob))

		time.Sleep(pastBackoff)
		synctest.Wait()

		f := st.Fails()
		if nFails := len(f); nFails != 1 {
			t.Fatalf("got %d fails, want 1", nFails)
		}

		if f[0].stage != "scrape" || !errors.Is(f[0].cause, testErr) {
			t.Errorf("fails = %+v, want {%d 0}", f[0].attempts, testBackoff.MaxAttempts)
		}

		if len(st.Saved()) != 0 {
			t.Errorf("persist at least one scrape that failed")
		}

		time.Sleep(monthlyJob.Interval + pastBackoff)
		synctest.Wait()

		wantedScrapeCalls := 2 * testBackoff.MaxAttempts
		if n := sc.Calls(models.MonthlyLeaderboard); n != wantedScrapeCalls {
			t.Errorf("scrape calls = %d, want %d after next cycle", n, wantedScrapeCalls)
		}

		mustStop(t, stop)
	})
}

func TestDaemon_RetriesPersistWithoutRescraping(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sc := newFakeScraper(scrapeOk)
		st := &fakeStore{
			persistFn: func(n int) error {
				if n <= 1 {
					return testErr
				}
				return nil
			},
		}

		stop := startDaemon(sc, st, testConfig(monthlyJob))
		time.Sleep(pastBackoff)
		synctest.Wait()

		if n := sc.Calls(models.MonthlyLeaderboard); n != 1 {
			t.Errorf("scrape calls = %d, want 1", n)
		}

		saved := st.Saved()

		wantedAttempts := store.Attempts{Scrape: 1, Persist: 2}
		if len(saved) != 1 || saved[0].attempts != wantedAttempts {
			t.Errorf("saved = %+v, want one save with attempts %+v", saved, wantedAttempts)
		}

		mustStop(t, stop)
	})
}

func TestDaemon_TestPersistGivesUp(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sc := newFakeScraper(scrapeOk)
		st := &fakeStore{
			persistFn: func(n int) error {
				return testErr
			},
		}

		stop := startDaemon(sc, st, testConfig(monthlyJob))
		time.Sleep(pastBackoff)
		synctest.Wait()

		// check no persists succeeded
		saved := st.Saved()
		if n := len(saved); n != 0 {
			t.Fatalf("saved %d scrapes, want 0", n)
		}

		// check only persist failures exist
		f := st.Fails()
		if nFails := len(f); nFails != 1 {
			t.Errorf("got %d fails, want 1", nFails)
		}

		// check persist exhausted all possible attempts
		wantedAttempts := store.Attempts{Scrape: 1, Persist: testBackoff.MaxAttempts}
		if gotAttempts := f[0].attempts; gotAttempts != wantedAttempts {
			t.Errorf("attempts = %+v, want {1 %d}", gotAttempts, testBackoff.MaxAttempts)
		}

		mustStop(t, stop)
	})
}

func TestDaemon_ShutsDownWhileIdle(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		st := &fakeStore{}
		stop := startDaemon(newFakeScraper(scrapeOk), st, testConfig(monthlyJob, globalJob))
		synctest.Wait() // both jobs are done; they're now waiting on their timers

		mustStop(t, stop) // kill daemon goroutine
		if f := st.Fails(); len(f) != 0 {
			t.Errorf("recorded fails = %+v, want none for an idle shutdown", f)
		}
	})
}

func TestDaemon_ShutdownMidScrapeRecordsAsFailure(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		st := &fakeStore{}
		stop := startDaemon(newFakeScraper(scrapeHang), st, testConfig(monthlyJob))
		synctest.Wait() // wait until in backoff period for a scrape attempt

		mustStop(t, stop)
		f := st.Fails()
		if len(f) != 1 || f[0].stage != "scrape" || !errors.Is(f[0].cause, context.Canceled) {
			t.Errorf("fails = %+v, want one scrape failure caused by context.Canceled", f)
		}
	})
}
