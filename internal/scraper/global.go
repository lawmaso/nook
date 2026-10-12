package scraper

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"

	"nook/internal/models"
	"nook/internal/parser"
)

// Parses a [globalLeaderboardRow] into [*models.User]
func parseGlobalRow(r globalLeaderboardRow) (*models.User, error) {
	uid, err := parser.UID(r.profileHref)
	if err != nil {
		return nil, fmt.Errorf("parsing href %q: %w", r.profileHref, err)
	}

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

	rawLifetime, err := getLifetimeData(uid)
	if err != nil {
		return nil, fmt.Errorf("getting lifetime stats text: %w", err)
	}

	lifetime, err := rawLifetimeData.Lifetime(rawLifetime)
	if err != nil {
		return nil, fmt.Errorf("adapting raw lifetime row: %w", err)
	}

	solved, err := parser.Solved(lifetime.solved)
	if err != nil {
		return nil, fmt.Errorf("parsing solved: %w", err)
	}

	attempted, err := parser.Solved(lifetime.attempted)
	if err != nil {
		return nil, fmt.Errorf("parsing attempted: %w", err)
	}

	acceptanceRate := parser.AcceptanceRate(solved, attempted)
	avgPoints, err := parser.AvgPointsGlobal(totalPoints, solved)

	if err != nil {
		return nil, fmt.Errorf("parsing global avg points: %w", err)
	}

	avgTime, err := parser.AvgTimeGlobal(lifetime.avgTime)
	if err != nil {
		return nil, fmt.Errorf("parsing avg time: %w", err)
	}

	return &models.User{
		UID:            uid,
		Rank:           rank,
		Username:       username,
		LastActive:     lastGame,
		Solved:         solved,
		Attempted:      attempted,
		AcceptanceRate: acceptanceRate,
		AverageTime:    avgTime,
		AveragePoints:  avgPoints,
		TotalPoints:    totalPoints,
	}, nil
}

func getGlobalLeaderboardUsers(ctx context.Context) ([]*models.User, error) {
	data, err := getLeaderboardText(models.GlobalLeaderboard)
	if err != nil {
		return nil, fmt.Errorf("getting global leaderboard text: %w", err)
	}

	users := make([]*models.User, len(data))

	g, _ := errgroup.WithContext(ctx)
	for index, row := range data {
		g.Go(func() error {
			globalRow, err := row.Global()
			if err != nil {
				return fmt.Errorf(
					"adapting raw row into global: %w", err,
				)
			}

			user, err := parseGlobalRow(globalRow)
			if err != nil {
				return fmt.Errorf(
					"parsing global row: %w", err,
				)
			}

			users[index] = user
			return nil
		})
	}

	// Block until all goroutines finish
	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf(
			"one or more global row parses failed: %w", err,
		)
	}

	return users, nil
}
