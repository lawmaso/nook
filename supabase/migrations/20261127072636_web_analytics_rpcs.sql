-- Aggregation RPCs for the web app (apps/web).
--
-- PostgREST caps responses at 1000 rows, and a month of hourly monthly scrapes
-- is ~72k scrape_entries, so time series are aggregated here. Both functions
-- pick the LAST scrape in each bucket (hour/day/week, truncated in site time,
-- America/New_York) so every point is a real leaderboard snapshot.
--
-- SECURITY INVOKER keeps the existing "public read" RLS policies in force.

CREATE OR REPLACE FUNCTION public.leaderboard_metrics(
    p_kind   leaderboard,
    p_from   timestamptz,
    p_to     timestamptz,
    p_bucket text DEFAULT 'day'
)
RETURNS TABLE (
    bucket         timestamptz,
    scrape_id      bigint,
    scraped_at     timestamptz,
    period         date,
    entrants       int,
    top1_points    int,
    top10_points   int,
    top50_points   int,
    top100_points  int,
    median_points  double precision,
    total_solved   bigint,
    avg_acceptance double precision,
    avg_time_s     double precision,
    active_24h     int
)
LANGUAGE sql STABLE SECURITY INVOKER
SET search_path = public
AS $$
    WITH picked AS (
        SELECT DISTINCT ON (date_trunc(p_bucket, s.scraped_at AT TIME ZONE 'America/New_York'))
            date_trunc(p_bucket, s.scraped_at AT TIME ZONE 'America/New_York')
                AT TIME ZONE 'America/New_York' AS bucket,
            s.scrape_id,
            s.scraped_at,
            s.period
        FROM scrapes s
        WHERE s.kind = p_kind
          AND s.scraped_at >= p_from
          AND s.scraped_at <  p_to
          AND p_bucket IN ('hour', 'day', 'week')
        ORDER BY date_trunc(p_bucket, s.scraped_at AT TIME ZONE 'America/New_York'),
                 s.scraped_at DESC
    )
    SELECT
        p.bucket,
        p.scrape_id,
        p.scraped_at,
        p.period,
        COUNT(*)::int,
        (array_agg(e.total_points ORDER BY e.rank, e.total_points DESC))[1],
        (array_agg(e.total_points ORDER BY e.rank, e.total_points DESC))[10],
        (array_agg(e.total_points ORDER BY e.rank, e.total_points DESC))[50],
        (array_agg(e.total_points ORDER BY e.rank, e.total_points DESC))[100],
        percentile_cont(0.5) WITHIN GROUP (ORDER BY e.total_points),
        SUM(e.solved),
        AVG(e.acceptance_rate),
        AVG(e.avg_time_s),
        (COUNT(*) FILTER (WHERE e.last_active >= p.scraped_at - interval '24 hours'))::int
    FROM picked p
    JOIN scrape_entries e ON e.scrape_id = p.scrape_id
    GROUP BY p.bucket, p.scrape_id, p.scraped_at, p.period
    ORDER BY p.bucket;
$$;

CREATE OR REPLACE FUNCTION public.user_history(
    p_uid    int,
    p_kind   leaderboard,
    p_from   timestamptz,
    p_to     timestamptz,
    p_bucket text DEFAULT 'day'
)
RETURNS TABLE (
    bucket          timestamptz,
    scraped_at      timestamptz,
    period          date,
    rank            int,
    total_points    int,
    solved          int,
    attempted       int,
    acceptance_rate double precision,
    avg_time_s      double precision,
    avg_points      double precision
)
LANGUAGE sql STABLE SECURITY INVOKER
SET search_path = public
AS $$
    SELECT DISTINCT ON (date_trunc(p_bucket, s.scraped_at AT TIME ZONE 'America/New_York'))
        date_trunc(p_bucket, s.scraped_at AT TIME ZONE 'America/New_York')
            AT TIME ZONE 'America/New_York',
        s.scraped_at,
        s.period,
        e.rank,
        e.total_points,
        e.solved,
        e.attempted,
        e.acceptance_rate,
        e.avg_time_s,
        e.avg_points
    FROM scrape_entries e
    JOIN scrapes s ON s.scrape_id = e.scrape_id
    WHERE e.uid = p_uid
      AND s.kind = p_kind
      AND s.scraped_at >= p_from
      AND s.scraped_at <  p_to
      AND p_bucket IN ('hour', 'day', 'week')
    ORDER BY date_trunc(p_bucket, s.scraped_at AT TIME ZONE 'America/New_York'),
             s.scraped_at DESC;
$$;

REVOKE EXECUTE ON FUNCTION public.leaderboard_metrics(leaderboard, timestamptz, timestamptz, text) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION public.user_history(int, leaderboard, timestamptz, timestamptz, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION public.leaderboard_metrics(leaderboard, timestamptz, timestamptz, text) TO anon, authenticated;
GRANT EXECUTE ON FUNCTION public.user_history(int, leaderboard, timestamptz, timestamptz, text) TO anon, authenticated;
