package main

import (
	"fmt"
	"os"

	"nook/internal/models"
	"nook/internal/scraper"
)

func main() {
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
			fmt.Println(user.String())
		}
	}

	if failed {
		os.Exit(1)
	}
}
