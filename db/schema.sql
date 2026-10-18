CREATE TABLE users (
    uid      INT PRIMARY KEY,
    username TEXT NOT NULL
);

CREATE TABLE scrapes(
    scrape_id        BIGSERIAL PRIMARY KEY,
    leaderboard_kind TEXT NOT NULL,
    scraped_at       TIMESTAMPTZ NOT NULL,

);

CREATE INDEX idx_scrape_kind_time
    ON scrapes (leaderboard_kind, scraped_at DESC);

CREATE TABLE scrape_entries (
    scrape_id BIGINT NOT NULL REFERENCES scrapes (scrape_id) ON DELETE CASCADE,
    uid       INT NOT NULL REFERENCES users (uid),
    data      JSONB NOT NULL,
    PRIMARY KEY (scrape_id, uid)
);

CREATE INDEX idx_entries_uid ON scrape_entries (uid);
