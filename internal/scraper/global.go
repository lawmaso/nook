package scraper

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"

	"nook/internal/models"
	"nook/internal/parser"
)

func parseGlobalRow(
	uid int,
	r globalLeaderboardRow,
	lt lifetimeData,
) (*models.User, error) {
	rank, err := parser.Rank(r.rank)
	if err != nil {
		return nil, fmt.Errorf("parsing rank %q: %w", r.rank, err)
	}

	username := r.username

	lastGame, err := parser.Date(r.lastGame)
	if err != nil {
		return nil, fmt.Errorf("parsing lastGame %q: %w", r.lastGame, err)
	}

	totalPoints, err := parser.TotalPoints(r.totalPoints)
	if err != nil {
		return nil, fmt.Errorf("parsing totalPoints %q: %w", r.totalPoints, err)
	}

	solved, err := parser.Count(lt.solved)
	if err != nil {
		return nil, fmt.Errorf("parsing solved %q: %w", lt.solved, err)
	}

	avgTime, err := parser.AvgTimeClock(lt.avgTime)
	if err != nil {
		return nil, fmt.Errorf("parsing avgTime: %q: %w", lt.avgTime, err)
	}

	return &models.User{
		UID:            uid,
		Rank:           rank,
		Username:       username,
		LastActive:     lastGame,
		Solved:         solved,
		Attempted:      lt.attempted,
		AcceptanceRate: parser.AcceptanceRate(solved, lt.attempted),
		AverageTime:    avgTime,
		AveragePoints:  parser.AvgPointsDivide(totalPoints, solved),
		TotalPoints:    totalPoints,
	}, nil
}

func (client *Client) getGlobalLeaderboardUsers(
	ctx context.Context,
) ([]*models.User, error) {
	data, err := client.getLeaderboardText(ctx, models.GlobalLeaderboard)
	if err != nil {
		return nil, fmt.Errorf("getting global leaderboard text: %w", err)
	}

	users := make([]*models.User, len(data))

	g, ctx := errgroup.WithContext(ctx) // keep derived ctx so one failure cancels the rest
	g.SetLimit(10)                      // bound concurrent profile fetches

	for i, row := range data {
		g.Go(func() error {
			gr, err := row.Global()
			if err != nil {
				return fmt.Errorf(
					"row %d: adapting to global: %w", i, err,
				)
			}

			uid, err := parser.UID(gr.profileHref)
			if err != nil {
				return fmt.Errorf(
					"row %d: parsing href %q: %w", i, gr.profileHref, err,
				)
			}

			raw, err := client.getLifetimeData(ctx, uid)
			if err != nil {
				return fmt.Errorf(
					"row %d: lifetime data: %w", i, err,
				)
			}

			lt, err := raw.Lifetime()
			if err != nil {
				return fmt.Errorf(
					"row %d: adapting lifetime data: %w", i, err,
				)
			}

			user, err := parseGlobalRow(uid, gr, lt)
			if err != nil {
				return fmt.Errorf(
					"row %d: %w", i, err,
				)
			}

			users[i] = user
			return nil
		})
	}

	// Block until all goroutines finish
	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("global scrape failed: %w", err)
	}

	return users, nil
}
