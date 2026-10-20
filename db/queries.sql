-----------------------------
---- Nook Daemon queries ----
-----------------------------

-- name: UpsertUser :exec
INSERT INTO users (uid, username, member_since)
VALUES (@uid, @username, @member_since)
ON CONFLICT (uid) DO UPDATE SET
    username = EXCLUDED.username,
    member_since = COALESCE(EXCLUDED.member_since, users.member_since);

-- name: InsertScrape :one
INSERT INTO scrapes (kind, scraped_at, duration_ms, period)
VALUES (@kind, @scraped_at, @duration_ms, @period)
RETURNING scrape_id;

-- name: InsertScrapeEntry :exec
INSERT INTO scrape_entries (
    scrape_id, uid, rank, total_points,
    solved, attempted, acceptance_rate,
    avg_time_s, avg_points, last_active
) VALUES (
    @scrape_id, @uid, @rank, @total_points,
    @solved, @attempted, @acceptance_rate,
    @avg_time_s, @avg_points, @last_active
);

--------------------------
---- Nook API queries ----
--------------------------

-- name: GetLatestScrapes :many
SELECT scrape_id, kind, scraped_at, duration_ms, period
FROM scrapes
WHERE kind = @kind
ORDER BY scraped_at DESC
LIMIT sqlc.arg('limit');

-- name: GetEntriesForScrape :many
SELECT e.uid, u.username, e.rank, e.total_points,
       e.solved, e.attempted, e.acceptance_rate,
       e.avg_time_s, e.avg_points, e.last_active
FROM scrape_entries e
JOIN users u USING (uid)
WHERE e.scrape_id = @scrape_id
ORDER BY e.rank ASC;
