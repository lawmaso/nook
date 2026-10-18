CREATE TABLE users (
    uid          INT         PRIMARY KEY,
    username     TEXT        NOT NULL,
    member_since TIMESTAMPTZ  -- NULL until uid is found in a global scrape
);

CREATE TABLE scrapes(
    scrape_id   BIGSERIAL   PRIMARY KEY,
    kind        TEXT        NOT NULL CHECK (kind IN ('monthly', 'global')),
    scraped_at  TIMESTAMPTZ NOT NULL,
    duration_ms INTEGER     NOT NULL CHECK (duration_ms >= 0),
    period      DATE  -- first day of month; NULL for global scrapes

    CONSTRAINT scrape_period_matches_kind CHECK (
        (kind = 'monthly' AND period IS NOT NULL) OR
        (kind = 'global'  AND period IS NULL)
    )
);

CREATE INDEX idx_scrapes_kind_period_time
    ON scrapes (kind, period, scraped_at);

CREATE TABLE scrape_entries (
    scrape_id    BIGINT NOT NULL REFERENCES scrapes (scrape_id) ON DELETE CASCADE,
    uid          INT    NOT NULL REFERENCES users (uid),
    rank         INT    NOT NULL,
    total_points INT    NOT NULL,
    data         JSONB  NOT NULL,  -- remaining fields; keep for future extendibility
    PRIMARY KEY (scrape_id, uid)
);

CREATE INDEX idx_scrape_entries_uid ON scrape_entries (uid, scrape_id);

ALTER TABLE users          ENABLE ROW LEVEL SECURITY;
ALTER TABLE scrapes        ENABLE ROW LEVEL SECURITY;
ALTER TABLE scrape_entries ENABLE ROW LEVEL SECURITY;

CREATE POLICY "public read" ON users          FOR SELECT TO anon USING (true);
CREATE POLICY "public read" ON scrapes        FOR SELECT TO anon USING (true);
CREATE POLICY "public read" ON scrape_entries FOR SELECT TO anon USING (true);