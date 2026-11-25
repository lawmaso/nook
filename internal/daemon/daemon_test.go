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

type scrapeFunc func(ctx context.Context, lb models.Leaderboard, n int) ([]*models.User, *scraper.Stats, error)

func scrapeOk(_ context.Context, lb models.Leaderboard, n int) ([]*models.User, *scraper.Stats, error) {
	users := []*models.User{{
		Rank:     1,
		UID:      n,
		Username: lb.DBKind(),
	}}
	stats := &scraper.Stats{At: time.Now()}

	return users, stats, nil
}

func scrapeHang(ctx context.Context, lb models.Leaderboard, n int) ([]*models.User, *scraper.Stats, error) {
	<-ctx.Done()
	return nil, nil, ctx.Err()
}

type fakeScraper struct {
	mu    sync.Mutex
	calls map[models.Leaderboard]int
	fn    scrapeFunc
}

func newFakeScraper(fn scrapeFunc) *fakeScraper {
	if fn == nil {
		fn = scrapeOk
	}

	return &fakeScraper{
		calls: map[models.Leaderboard]int{},
		fn:    fn,
	}
}

func (f *fakeScraper) ScrapeUsers(ctx context.Context, lb models.Leaderboard) ([]*models.User, *scraper.Stats, error) {
	f.mu.Lock()
	f.calls[lb]++

	n := f.calls[lb]
	f.mu.Unlock()

	return f.fn(ctx, lb, n)
}

// Returns the current number of scrape calls the fake scraper has made
func (f *fakeScraper) Calls(lb models.Leaderboard) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls[lb]
}

type persisted struct {
	lb       models.Leaderboard
	users    []*models.User
	attempts store.Attempts
}

type failed struct {
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
	saved     []persisted
	failures  []failed
}

func (f *fakeStore) PersistScrape(_ context.Context, lb models.Leaderboard, users []*models.User, _ *scraper.Stats, att store.Attempts) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.calls++
	if f.persistFn != nil {
		if err := f.persistFn(f.calls); err != nil {
			return err
		}
	}
	f.saved = append(f.saved, persisted{lb, users, att})
	return nil
}

func (f *fakeStore) PersistScrapeFailure(_ context.Context, lb models.Leaderboard, stage string, att store.Attempts, cause error) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.failures = append(f.failures, failed{lb, stage, att, cause, time.Now()})
	return nil
}

func (f *fakeStore) Saved() []persisted {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]persisted(nil), f.saved...)
}

func (f *fakeStore) Failures() []failed {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]failed(nil), f.failures...)
}

// ---------------------------------------------------------------------
// Harness
// ---------------------------------------------------------------------

var (
	monthlyJob = Job{Leaderboard: models.MonthlyLeaderboard, Interval: 15 * time.Minute, Timeout: 30 * time.Second}
	globalJob  = Job{Leaderboard: models.GlobalLeaderboard, Interval: time.Hour, Timeout: time.Minute}
)

// Waits before jitter: 1s, 2s. Worst case for a whole retry loop is 3s.
var testBackoff = Backoff{Initial: time.Second, Max: 10 * time.Second, Factor: 2, MaxAttempts: 3}

const pastBackoff = 5 * time.Second

func testConfig(jobs ...Job) Config {
	return Config{Jobs: jobs, Backoff: testBackoff}
}

// startDaemon runs the daemon in the background. Call the returned stop func
// before the synctest bubble ends: it cancels Run and returns its error.
func startDaemon(sc Scraper, st Store, cfg Config) (stop func() error) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- New(sc, st, cfg, testLog).Run(ctx) }()
	return func() error {
		cancel()
		return <-done
	}
}

func mustStop(t *testing.T, stop func() error) {
	t.Helper()
	if err := stop(); !errors.Is(err, context.Canceled) {
		t.Errorf("Run returned %v, want context.Canceled", err)
	}
}

// ---------------------------------------------------------------------
// Scheduling
// ---------------------------------------------------------------------

func TestDaemon_FirstCycleRunsImmediately(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sc, st := newFakeScraper(nil), &fakeStore{}
		stop := startDaemon(sc, st, testConfig(monthlyJob))

		synctest.Wait() // no fake time has passed

		saved := st.Saved()
		if len(saved) != 1 {
			t.Fatalf("saved %d scrapes at t=0, want 1", len(saved))
		}
		if saved[0].attempts != (store.Attempts{Scrape: 1, Persist: 1}) {
			t.Errorf("attempts = %+v, want {1 1}", saved[0].attempts)
		}
		if got := saved[0].users; len(got) != 1 || got[0].Username != "monthly" {
			t.Errorf("users = %+v, want exactly what the scraper returned", got)
		}
		mustStop(t, stop)
	})
}

func TestDaemon_RunsOnInterval(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sc, st := newFakeScraper(nil), &fakeStore{}
		stop := startDaemon(sc, st, testConfig(monthlyJob))

		time.Sleep(monthlyJob.Interval - time.Nanosecond)
		synctest.Wait()
		if n := sc.Calls(models.MonthlyLeaderboard); n != 1 {
			t.Fatalf("scrapes just before the interval = %d, want 1", n)
		}

		time.Sleep(time.Nanosecond)
		synctest.Wait()
		if n := sc.Calls(models.MonthlyLeaderboard); n != 2 {
			t.Fatalf("scrapes at the interval = %d, want 2", n)
		}
		mustStop(t, stop)
	})
}

// The next cycle is scheduled from when the last one ended, so slow cycles
// never overlap or bunch up.
func TestDaemon_IntervalStartsAfterCycleEnds(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		const scrapeTime = 10 * time.Second
		var starts []time.Time
		var mu sync.Mutex
		sc := newFakeScraper(func(ctx context.Context, lb models.Leaderboard, n int) ([]*models.User, *scraper.Stats, error) {
			mu.Lock()
			starts = append(starts, time.Now())
			mu.Unlock()
			time.Sleep(scrapeTime)
			return scrapeOk(ctx, lb, n)
		})
		stop := startDaemon(sc, &fakeStore{}, testConfig(monthlyJob))

		time.Sleep(2 * (monthlyJob.Interval + scrapeTime))
		synctest.Wait()
		mustStop(t, stop)

		if len(starts) < 2 {
			t.Fatalf("got %d scrapes, want at least 2", len(starts))
		}
		if gap := starts[1].Sub(starts[0]); gap != scrapeTime+monthlyJob.Interval {
			t.Errorf("gap between scrape starts = %v, want %v", gap, scrapeTime+monthlyJob.Interval)
		}
	})
}

func TestDaemon_StuckJobDoesNotBlockOtherJobs(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sc := newFakeScraper(func(ctx context.Context, lb models.Leaderboard, n int) ([]*models.User, *scraper.Stats, error) {
			if lb == models.MonthlyLeaderboard {
				return scrapeHang(ctx, lb, n)
			}
			return scrapeOk(ctx, lb, n)
		})
		st := &fakeStore{}
		stop := startDaemon(sc, st, testConfig(monthlyJob, globalJob))

		synctest.Wait() // monthly is hung; global should already be done

		saved := st.Saved()
		if len(saved) != 1 || saved[0].lb != models.GlobalLeaderboard {
			t.Errorf("saved = %+v, want one global scrape", saved)
		}
		mustStop(t, stop)
	})
}

// ---------------------------------------------------------------------
// Retries and failures
// ---------------------------------------------------------------------

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
		if len(saved) != 1 {
			t.Fatalf("saved %d scrapes, want 1", len(saved))
		}
		if saved[0].attempts != (store.Attempts{Scrape: 3, Persist: 1}) {
			t.Errorf("attempts = %+v, want {3 1}", saved[0].attempts)
		}
		if f := st.Failures(); len(f) != 0 {
			t.Errorf("recorded failures %+v, want none", f)
		}
		mustStop(t, stop)
	})
}

func TestDaemon_ScrapeGivesUpThenNextCycleRuns(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sc := newFakeScraper(func(context.Context, models.Leaderboard, int) ([]*models.User, *scraper.Stats, error) {
			return nil, nil, testErr
		})
		st := &fakeStore{}
		stop := startDaemon(sc, st, testConfig(monthlyJob))

		time.Sleep(pastBackoff)
		synctest.Wait()

		f := st.Failures()
		if len(f) != 1 {
			t.Fatalf("recorded %d failures, want 1", len(f))
		}
		if f[0].stage != "scrape" || !errors.Is(f[0].cause, testErr) {
			t.Errorf("failure = %+v, want stage scrape wrapping errBoom", f[0])
		}
		if f[0].attempts != (store.Attempts{Scrape: testBackoff.MaxAttempts, Persist: 0}) {
			t.Errorf("attempts = %+v, want {%d 0}", f[0].attempts, testBackoff.MaxAttempts)
		}
		if len(st.Saved()) != 0 {
			t.Error("persisted a scrape that failed")
		}

		// A failed cycle must not kill the job. The next one starts an interval
		// after this one gave up, then needs its own backoff time to finish.
		time.Sleep(monthlyJob.Interval + pastBackoff)
		synctest.Wait()
		if n := sc.Calls(models.MonthlyLeaderboard); n != 2*testBackoff.MaxAttempts {
			t.Errorf("scrape calls = %d, want %d after a second cycle", n, 2*testBackoff.MaxAttempts)
		}
		mustStop(t, stop)
	})
}

// A persist failure retries the write only; the data already scraped is reused.
func TestDaemon_PersistRetriesWithoutRescraping(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sc := newFakeScraper(nil)
		st := &fakeStore{persistFn: func(n int) error {
			if n == 1 {
				return testErr
			}
			return nil
		}}
		stop := startDaemon(sc, st, testConfig(monthlyJob))

		time.Sleep(pastBackoff)
		synctest.Wait()

		if n := sc.Calls(models.MonthlyLeaderboard); n != 1 {
			t.Errorf("scrape calls = %d, want 1", n)
		}
		saved := st.Saved()
		if len(saved) != 1 || saved[0].attempts != (store.Attempts{Scrape: 1, Persist: 2}) {
			t.Errorf("saved = %+v, want one scrape with attempts {1 2}", saved)
		}
		mustStop(t, stop)
	})
}

func TestDaemon_PersistGivesUp(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		st := &fakeStore{persistFn: func(int) error { return testErr }}
		stop := startDaemon(newFakeScraper(nil), st, testConfig(monthlyJob))

		time.Sleep(pastBackoff)
		synctest.Wait()

		f := st.Failures()
		if len(f) != 1 || f[0].stage != "persist" {
			t.Fatalf("failures = %+v, want one at stage persist", f)
		}
		if f[0].attempts != (store.Attempts{Scrape: 1, Persist: testBackoff.MaxAttempts}) {
			t.Errorf("attempts = %+v, want {1 %d}", f[0].attempts, testBackoff.MaxAttempts)
		}
		mustStop(t, stop)
	})
}

// Job.Timeout bounds each attempt, not the whole retry loop.
func TestDaemon_TimeoutIsPerAttempt(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		start := time.Now()
		st := &fakeStore{}
		stop := startDaemon(newFakeScraper(scrapeHang), st, testConfig(monthlyJob))

		time.Sleep(time.Duration(testBackoff.MaxAttempts)*monthlyJob.Timeout + pastBackoff)
		synctest.Wait()

		f := st.Failures()
		if len(f) != 1 {
			t.Fatalf("recorded %d failures, want 1", len(f))
		}
		if !errors.Is(f[0].cause, context.DeadlineExceeded) {
			t.Errorf("cause = %v, want context.DeadlineExceeded", f[0].cause)
		}
		minElapsed := time.Duration(testBackoff.MaxAttempts) * monthlyJob.Timeout
		if elapsed := f[0].at.Sub(start); elapsed < minElapsed {
			t.Errorf("gave up after %v, want at least %v (one full timeout per attempt)", elapsed, minElapsed)
		}
		mustStop(t, stop)
	})
}

// ---------------------------------------------------------------------
// Shutdown
// ---------------------------------------------------------------------

func TestDaemon_ShutdownWhileIdle(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		st := &fakeStore{}
		stop := startDaemon(newFakeScraper(nil), st, testConfig(monthlyJob, globalJob))
		synctest.Wait() // both jobs done with cycle 1, waiting on their timers

		mustStop(t, stop)
		if f := st.Failures(); len(f) != 0 {
			t.Errorf("recorded failures %+v, want none for an idle shutdown", f)
		}
	})
}

// Shutting down mid-scrape is still recorded as a failure, with a context that
// outlives the cancelled one so the write can land.
func TestDaemon_ShutdownMidScrapeRecordsFailure(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		st := &fakeStore{}
		stop := startDaemon(newFakeScraper(scrapeHang), st, testConfig(monthlyJob))
		synctest.Wait()

		mustStop(t, stop)

		f := st.Failures()
		if len(f) != 1 || f[0].stage != "scrape" || !errors.Is(f[0].cause, context.Canceled) {
			t.Errorf("failures = %+v, want one scrape failure caused by context.Canceled", f)
		}
	})
}
