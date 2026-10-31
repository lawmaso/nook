package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"nook/internal/models"
	"nook/internal/scraper"
)

func main() {
	detailed := flag.Bool("detailed", false, "print detailed user info")
	timeout := flag.Duration("timeout", 10*time.Second, "deadline per leaderboard scrape")
	flag.Parse()

	client := scraper.NewClient()
	failed := false

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	for _, lb := range []models.Leaderboard{
		models.MonthlyLeaderboard,
		models.GlobalLeaderboard,
	} {
		// Check if interrupted between scrapes
		if ctx.Err() != nil {
			os.Exit(128 + int(syscall.SIGINT))
		}

		users, stats, err := scrapeOne(ctx, client, lb, *timeout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v: %v\n", lb, err)
			failed = true
			continue
		}

		fmt.Printf(
			"---- %v (%d users, %s)\n",
			lb, len(users),
			stats.Duration.Round(time.Millisecond),
		)

		for _, user := range users {
			if *detailed {
				fmt.Println(user.DetailedString())
			} else {
				fmt.Println(user.String())
			}
		}
		fmt.Println()
	}

	if failed {
		os.Exit(1)
	}
	os.Exit(0)
}

// Each scrape gets its own deadline; derived from original signal ctx
func scrapeOne(
	ctx context.Context,
	client *scraper.Client,
	lb models.Leaderboard,
	timeout time.Duration,
) ([]*models.User, *scraper.Stats, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return client.ScrapeUsers(ctx, lb)
}
