/**
 * @file models.ts
 */

export type ISODate      = string
export type ISOTimestamp = string
export type Kind         = "monthly" | "global"
export type ScrapeStage  = "scrape" | "persist"

export type User = {
    uid:         number
    username:    string
    memberSince: ISODate | null
}

export type Scrape = {
    scrape_id:        number
    kind:             Kind,
    period:           ISODate | null
    scraped_at:       ISOTimestamp
    duration_ms:      number
    scrape_attempts:  number
    persist_attempts: number
}

export type ScrapeEntry = {
    scrape_id:       number
    uid:             number
    rank:            number
    total_points:    number
    solved:          number
    attempted:       number
    acceptance_rate: number
    avg_time_s:      number
    avg_points:      number
    last_active:     ISOTimestamp
}

export type LeaderboardEntry = ScrapeEntry & { user: User }

export type Leaderboard = {
    scrape: Scrape,
    entries: LeaderboardEntry[]
}
