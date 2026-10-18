package daemon

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"time"
)

type Backoff struct {
	Initial     time.Duration
	Max         time.Duration
	Factor      float64
	MaxAttempts int
}

// Calls a function fn using a [Backoff] configuration
func retry(
	ctx context.Context, b Backoff, log *slog.Logger,
	op string, fn func(context.Context) error,
) error {
	delay := b.Initial

	for attempt := 1; ; attempt++ {
		err := fn(ctx)

		// Function fn returned no error; no more attempts needed
		if err == nil {
			return nil
		}

		// Shutting down; don't retry anymore
		if ctxErr := ctx.Err(); ctxErr != nil {
			return fmt.Errorf("%s: %w", op, ctxErr)
		}

		// Ran out of retry attempts
		if attempt >= b.MaxAttempts {
			return fmt.Errorf("%s: gave up after %d attempts: %w", op, attempt, err)
		}

		wait := delay/2 + rand.N(delay/2+1)
		log.Warn(
			"attempt failed, backing off", "op",
			op, "attempt", attempt, "retry_in", wait,
			"err", err,
		)

		t := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			t.Stop()
			return fmt.Errorf("%s: %w", op, ctx.Err())
		case <-t.C:
		}

		delay = min(time.Duration(float64(delay)*b.Factor), b.Max)
	}
}
