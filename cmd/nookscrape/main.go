package main

import (
	"flag"
	"fmt"
	"os"

	"nook/internal/models"
	"nook/internal/scraper"
)

func main() {
	detailed := flag.Bool("detailed", false, "print detailed user info")
	flag.Parse()

	failed := false

	for _, model := range []models.Leaderboard{
		models.MonthlyLeaderboard,
		models.GlobalLeaderboard,
	} {
		users, _, err := scraper.GetLeaderboardUsers(model)
		if err != nil {
			fmt.Printf("%s\n", err.Error())
			failed = true
			continue
		}

		fmt.Printf("---- %v\n", model)
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
}
