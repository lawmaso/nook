CREATE TYPE leaderboard AS ENUM ('monthly', 'global');

CREATE TABLE users (
    uid          INT PRIMARY KEY,
    username     TEXT NOT NULL,
    member_since DATE  -- NULL until uid is found in a global scrape
);

CREATE TABLE scrapes (
    scrape_id   BIGINT      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    kind        leaderboard NOT NULL,
    scraped_at  TIMESTAMPTZ NOT NULL,  -- scrape start (from ScrapeStats)
    duration_ms INTEGER     NOT NULL CHECK (duration_ms >= 0),
    period      DATE,  -- first day of month (ET); NULL for global scrapes

    CONSTRAINT scrape_period_matches_kind CHECK (
        (kind = 'monthly' AND period is NOT NULL AND EXTRACT(DAY FROM period) = 1) OR
        (kind = 'global'  AND period is NULL)
    )
);

CREATE INDEX idx_scrapes_kind_period_time ON scrapes (kind, period, scraped_at);

CREATE TABLE scrape_entries (
    scrape_id       BIGINT           NOT NULL REFERENCES scrapes (scrape_id) ON DELETE CASCADE,
    uid             INT              NOT NULL REFERENCES users (uid),
    rank            INT              NOT NULL CHECK (rank > 0),
    total_points    INT              NOT NULL,
    solved          INT              NOT NULL,
    attempted       INT              NOT NULL,
    acceptance_rate DOUBLE PRECISION NOT NULL,
    avg_time_s      DOUBLE PRECISION NOT NULL,
    avg_points      DOUBLE PRECISION NOT NULL,
    last_active     TIMESTAMPTZ      NOT NULL,
    PRIMARY KEY (scrape_id, uid)
);

CREATE INDEX idx_scrape_entries_uid ON scrape_entries (uid, scrape_id);
