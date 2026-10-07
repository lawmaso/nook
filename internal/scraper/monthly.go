package scraper

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"

	"nook/internal/models"
	"nook/internal/parser"
)

// Parses a [monthlyLeaderboardRow] into [*models.User]
func parseMonthRow(r monthlyLeaderboardRow) (*models.User, error) {
	uid, err := parser.UID(r.profileHref)
	if err != nil {
		return nil, fmt.Errorf(
			"parsing href %q: %w", r.profileHref, err,
		)
	}

	rank, err := parser.Rank(r.rank)
	if err != nil {
		return nil, fmt.Errorf(
			"parsing rank %q: %w", r.rank, err,
		)
	}

	username := r.username

	lastActive, err := parser.Date(r.lastActive)
	if err != nil {
		return nil, fmt.Errorf(
			"parsing last active %q: %w", r.lastActive, err,
		)
	}

	solved, attempted, err := parser.SolveAttempted(r.solvedFraction)
	if err != nil {
		return nil, fmt.Errorf(
			"parsing solved fraction %q: %w", r.solvedFraction, err,
		)
	}

	acceptanceRate := parser.AcceptanceRate(solved, attempted)

	avgTime, err := parser.AvgTimeMonthly(r.avgTime)
	if err != nil {
		return nil, fmt.Errorf(
			"parsing average time %q: %w", r.avgTime, err,
		)
	}

	avgPoints, err := parser.AvgPoints(r.avgPoints)
	if err != nil {
		return nil, fmt.Errorf(
			"parsing average points %q: %w", r.avgPoints, err,
		)
	}

	totalPoints, err := parser.TotalPoints(r.totalPoints)
	if err != nil {
		return nil, fmt.Errorf(
			"parsing total points %q: %w", r.totalPoints, err,
		)
	}

	return &models.User{
		UID:            uid,
		Rank:           rank,
		Username:       username,
		LastActive:     lastActive,
		Solved:         solved,
		Attempted:      attempted,
		AcceptanceRate: acceptanceRate,
		AverageTime:    avgTime,
		AveragePoints:  avgPoints,
		TotalPoints:    totalPoints,
	}, nil
}

// Returns all currents users on the monthly leaderboard
func getMonthlyLeaderboardUsers(ctx context.Context) ([]*models.User, error) {
	// Fetch monthly leaderboard text
	data, err := getLeaderboardText(models.MonthlyLeaderboard)
	if err != nil {
		return nil, fmt.Errorf("getting monthly leaderboard text: %w", err)
	}

	// Preallocated monthly users slice
	users := make([]*models.User, len(data))

	// Wait group implementation
	// var wg sync.WaitGroup
	// for index, row := range data {
	// 	wg.Go(func() {
	// 		user, _ := parseMonthRow(row)
	// 		users[index] = user
	// 	})
	// }
	// wg.Wait()

	g, _ := errgroup.WithContext(ctx)

	for index, row := range data {
		g.Go(func() error {
			monthRow, err := row.Monthly()
			if err != nil {
				return fmt.Errorf(
					"adapting raw row into monthly: %w", err,
				)
			}

			user, err := parseMonthRow(monthRow)
			if err != nil {
				return fmt.Errorf(
					"processing row index %d: %w", index, err,
				)
			}

			users[index] = user
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf(
			"one or more monthly row parses failed: %w", err,
		)
	}
	return users, nil
}
