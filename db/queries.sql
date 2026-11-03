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
INSERT INTO scrapes (
    kind, period,
    scraped_at, duration_ms,
    scrape_attempts, persist_attempts
) VALUES (
    @kind, @period,
    @scraped_at, @duration_ms,
    @scrape_attempts, @persist_attempts
)
RETURNING scrape_id;

-- name: InsertScrapeFailure :exec
INSERT INTO scrape_failures (
    kind, stage, failed_at,
    scrape_attempts, persist_attempts,
    error
) VALUES (
    @kind, @stage, @failed_at,
    @scrape_attempts, @persist_attempts,
    @error
);

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

-- name: GetScrapeEntries :many
SELECT e.uid, u.username, e.rank, e.total_points,
       e.solved, e.attempted, e.acceptance_rate,
       e.avg_time_s, e.avg_points, e.last_active
FROM scrape_entries e
JOIN users u USING (uid)
WHERE e.scrape_id = @scrape_id
ORDER BY e.rank ASC;

-- name: ReliabilitySince :one
WITH s AS (
    SELECT
        COUNT(*) AS successes,
        COUNT(*) FILTER (WHERE scrape_attempts > 1 OR persist_attempts > 1) AS recovered
    FROM scrapes WHERE scraped_at >= @since
), f AS (
    SELECT COUNT(*) AS failures
    FROM scrape_failures WHERE failed_at >= @since
)
SELECT s.successes, s.recovered, f.failures,
       ROUND(100.0 * s.successes / NULLIF(s.successes + s.failures, 0), 2)::float8 AS success_pct,
       ROUND(100.0 * s.recovered / NULLIF(s.recovered + s.failures, 0), 2)::float8 AS recovery_pct
FROM s, f;
