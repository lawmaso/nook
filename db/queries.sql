-- name: UpsertUser :exec
INSERT INTO users (uid, username)
VALUES ($1, $2)
ON CONFLICT (uid) DO UPDATE SET username = EXCLUDED.username;

-- name: InsertScrape :one
INSERT INTO leaderboard_scrapes (leaderboard_kind, scraped_at)
VALUES ($1, now())
RETURNING scrape_id;

-- name: InsertEntry :exec
INSERT INTO leaderboard_entries (scrape_id, uid, data)
VALUES ($1, $2, $3);

-- name: GetLatestScrapes :many
SELECT scrape_id, leaderboard_kind, scraped_at
FROM leaderboard_scrapes
WHERE leaderboard_kind = $1
ORDER BY scraped_at DESC
LIMIT $2;

-- name: GetEntriesForScrape :many
SELECT uid, data
FROM leaderboard_entries
WHERE scrape_id = $1;

-- name: GetScrapesInRange :many
SELECT scrape_id, leaderboard_kind, scraped_at
FROM leaderboard_scrapes
WHERE leaderboard_kind = $1
    AND scraped_at BETWEEN $2 and $3
ORDER BY scraped_at ASC;
