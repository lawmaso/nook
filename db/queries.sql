-- name: UpsertUser :exec
INSERT INTO users (uid, username)
VALUES ($1, $2)
ON CONFLICT (uid) DO UPDATE SET
    username     = EXCLUDED.username,
    member_since = COALESCE(EXCLUDED.member_since, users.member_since);

-- name: InsertScrape :one
INSERT INTO scrapes (kind, scraped_at, duration_ms, period)
VALUES ($1, $2, $3, $4)
RETURNING scrape_id;

-- name: InsertScrapeEntry :exec
INSERT INTO scrape_entries (scrape_id, uid, rank, total_points, data)
VALUES ($1, $2, $3, $4, $5);
