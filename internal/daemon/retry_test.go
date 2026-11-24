package daemon

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"testing/synctest"
	"time"
)

var (
	testErr = errors.New("testErr")
	testLog = slog.New(slog.DiscardHandler)
)

// Pre-jitter delays
var retryBackoff = Backoff{
	Initial:     time.Second,
	Max:         4 * time.Second,
	Factor:      2,
	MaxAttempts: 6,
}

// Returns a fn that fails n times before succeeding
func failTimes(n int, calls *int) func(context.Context) error {
	return func(ctx context.Context) error {
		*calls++
		if *calls <= n {
			return testErr
		}
		return nil
	}
}

func TestRetry_SucceedsFirstTry(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		start := time.Now()
		calls := 0

		err := retry(t.Context(), retryBackoff, testLog, "op", failTimes(0, &calls))
		if err != nil {
			t.Fatalf("err = %v, want nil", err)
		}

		if calls != 1 {
			t.Errorf("calls = %d, want 1", calls)
		}

		// NOTE: this essentially is asserting that virtual time (in the synctest bubble) hasn't elapsed
		if wait := time.Since(start); wait != 0 {
			t.Errorf("waited %v, want no backoff after a success", wait)
		}
	})
}

func TestRetry_SucceedsUntilSuccess(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		calls := 0
		failures := 2

		err := retry(t.Context(), retryBackoff, testLog, "op", failTimes(failures, &calls))
		if err != nil {
			t.Fatalf("err = %v, want nil", err)
		}

		if wantedCalls := failures + 1; calls != wantedCalls {
			t.Errorf("calls = %d, want %d (%d fails + 1 success)", calls, wantedCalls, failures)
		}
	})
}

func TestRetry_GivesUpAfterMaxAttempts(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		calls := 0

		err := retry(t.Context(), retryBackoff, testLog, "op", failTimes(8, &calls))

		if !errors.Is(err, testErr) {
			t.Fatalf("err = %v, want it to be %v", err, testErr)
		}

		if maxAttempts := retryBackoff.MaxAttempts; calls != maxAttempts {
			t.Errorf("calls = %d, want %d", calls, maxAttempts)
		}
	})
}

// Each time must land in the range [d/2, d]
// NOTE: jitter is pseudo-random, so multiple rounds are run
func TestRetry_BackoffGrowsWithinBounds(t *testing.T) {
	wantedDelays := []time.Duration{
		1 * time.Second,
		2 * time.Second,
		4 * time.Second,
		4 * time.Second,
		4 * time.Second,
	}

	for range 100 {
		synctest.Test(t, func(t *testing.T) {
			var at []time.Time // fn call times
			fn := func(context.Context) error {
				at = append(at, time.Now())
				return testErr
			}

			_ = retry(t.Context(), retryBackoff, testLog, "op", fn)

			if len(at) != len(wantedDelays)+1 {
				t.Fatalf("got %d calls, want %d", len(at), len(wantedDelays)+1)
			}

			for i, d := range wantedDelays {
				gap := at[i+1].Sub(at[i])

				if gap < d/2 || gap > d {
					t.Errorf(
						"wait before attempt %d = %v, must be in range [%v, %v]",
						i+2, gap, d/2, d,
					)
				}
			}
		})
	}
}

func TestRetry_CancelDuringBackoffReturnsImmediately(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())

		start := time.Now()
		calls := 0

		done := make(chan error, 1)
		go func() { done <- retry(ctx, retryBackoff, testLog, "op", failTimes(100, &calls)) }()

		// NOTE: this wait ensures that the retry call in
		// the goroutine is on its backoff timer
		synctest.Wait()
		cancel()
		err := <-done

		if !errors.Is(err, context.Canceled) {
			t.Fatalf("err = %v, want context.Canceled", err)
		}

		if calls != 1 {
			t.Errorf("calls = %d, want 1 (no attempt after cancel)", calls)
		}

		if waited := time.Since(start); waited != 0 {
			t.Errorf("waited %v, want cancel to cut the backoff short", waited)
		}
	})
}

// If the fn passed to retry fails since the ctx is finished, retrying is useless: stop and report the err
func TestRetry_DoesNotRetryOnceContextIsDone(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		calls := 0
		err := retry(ctx, retryBackoff, testLog, "op", failTimes(100, &calls))

		if !errors.Is(err, context.Canceled) {
			t.Fatalf("err = %v, want context.Canceled", err)
		}

		if calls != 1 {
			t.Errorf("calls = %d, want 1", calls)
		}
	})
}
