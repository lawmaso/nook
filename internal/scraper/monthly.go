package scraper

import (
	"context"
	"fmt"

	"nook/internal/models"
	"nook/internal/parser"
)

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
func (client *Client) getMonthlyLeaderboardUsers(ctx context.Context) ([]*models.User, error) {
	data, err := client.getLeaderboardText(ctx, models.MonthlyLeaderboard)
	if err != nil {
		return nil, fmt.Errorf("getting monthly leaderboard text: %w", err)
	}

	// No network calls per row, so no goroutines or errgroup needed
	users := make([]*models.User, len(data))
	for i, row := range data {
		mr, err := row.Monthly()
		if err != nil {
			return nil, fmt.Errorf("row %d: adapting to monthly", i, err)
		}

		user, err := parseMonthRow(mr)
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", i, err)
		}

		users[i] = user
	}

	return users, nil
}
