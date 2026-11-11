/**
 * @queries.ts
 */

import { ISODate, Kind, Leaderboard, LeaderboardEntry, Scrape, ScrapeEntry, TrajectoryMetric, TrajectoryPoint, TrajectoryRow, User } from "./models";
import { supabase } from "./supabase";


const PAGE_SIZE = 1000  // max rows per request

async function fetchAll<T>(
    buildPage: (from: number, to: number) => PromiseLike<{
        data: T[] | null
        error: Error | null
    }>
): Promise<T[]> {
    const rows: T[] = []

    for (let from = 0; ; from += 1000) {
        const { data, error } = await buildPage(from, from + PAGE_SIZE - 1)

        if (error) throw error
        rows.push(...(data ?? []))

        if (!data || data.length < PAGE_SIZE) break
    }

    return rows
}

export async function getScrapePeriods(kind: Kind = "monthly"): Promise<ISODate[]> {
    const { data, error } = await supabase
        .from("scrapes")
        .select("period")
        .eq("kind", kind)
        .not("period", "is", null)
        .order("period", { ascending: false })
    
    if (error) throw error
    return [...new Set((data ?? []).map((r) => r.period as ISODate))]
}

export async function getScrapes(kind: Kind, period?: ISODate | null): Promise<Scrape[]> {
    let query = supabase
        .from("scrapes")
        .select("*")
        .eq("kind", kind)
        .order("scraped_at", { ascending: false })
    
    if (kind === "monthly" && period) {
        query = query.eq("period", period)
    }

    const { data, error } = await query
    if (error) throw error
    return (data ?? []) as Scrape[]
}

export async function getLatestScrape(kind: Kind, period?: ISODate): Promise<Scrape | null> {
    let query = supabase
        .from("scrapes")
        .select("*")
        .eq("kind", kind)

    if (kind === "global") query = query.is("period", null)
    else if (period)       query = query.eq("period", period)
    else                   query = query.order("period", { ascending: false })

    const { data, error } = await query
        .order("scraped_at", { ascending: false })
        .limit(1)
        .maybeSingle()

    if (error) throw error
    return (data as Scrape | null) ?? null
}

export async function getUsers(): Promise<User[]> {
    return fetchAll<User>((from, to) =>
        supabase
            .from("users")
            .select("*")
            .order("uid")
            .range(from, to)
    )
}

export async function getUser(uid: number): Promise<User | null> {
    const { data, error } = await supabase
        .from("users")
        .select("*")
        .eq("uid", uid)
        .maybeSingle()
    
    if (error) throw error
    return (data as User | null) ?? null
}

export async function searchUsers(term: string, limit: 20): Promise<User[]> {
    const { data, error } = await supabase
        .from("users")
        .select("*")
        .ilike("username", `%${term}%`)
        .order("username")
        .limit(limit)

    if (error) throw error
    return (data ?? []) as User[]
}

export async function getScrapeEntries(scrapeId: number): Promise<LeaderboardEntry[]> {
    const { data, error } = await supabase
        .from("scrape_entries")
        .select("*, users(*)")
        .eq("scrape_id", scrapeId)
        .order("rank")
        .limit(100)  // max 100 entries per scrape
    
    if (error) throw error
    return data.map(({ users, ...entry }) => ({
        ...entry,
        user: users
    }))
}


export async function getLatestLeaderboard(
    kind: Kind, period?: ISODate
): Promise<Leaderboard | null> {
    const scrape = await getLatestScrape(kind, period)
    if (!scrape) return null
    
    return {
        scrape,
        entries: await getScrapeEntries(scrape.scrape_id)
    }
}

export async function getLatestScrapeEntries(kind: Kind): Promise<ScrapeEntry[] | null> {
    const leaderboard = await getLatestLeaderboard(kind)
    if (!leaderboard) return null

    return leaderboard.entries
}

const MAX_TRAJECTORY_POINTS = 80

function downsample<T>(items: T[], max: number): T[] {
    if (items.length <= max) return items

    const step = (items.length - 1) / (max - 1)
    return Array.from({ length: max }, (_v, i) => items[Math.round(i * step)])
}

type ScrapeStub = Pick<Scrape, "scrape_id" | "scraped_at">
type MetricRow = { scrape_id: number; uid: number } & Partial<Record<TrajectoryMetric, number>>

export async function getTrajectories(
    kind: Kind,
    period: ISODate | undefined,
    uids: number[],
    metric: TrajectoryMetric
): Promise<TrajectoryPoint[]> {
    if (uids.length === 0) return []

    const scrapes = await fetchAll<ScrapeStub>((from, to) => {
        let query = supabase
            .from("scrapes")
            .select("scrape_id, scraped_at")
        
        if (kind === "global") query = query.is("period", null)
        else if (period)       query = query.eq("period", period)

        return query.order("scraped_at").range(from, to)
    })

    const picked = downsample(scrapes, MAX_TRAJECTORY_POINTS)
    if (picked.length === 0) return []
    const timeById = new Map(picked.map((s) => [s.scrape_id, Date.parse(s.scraped_at)]))

    const rows = await fetchAll<MetricRow>((from, to) =>
        supabase
            .from("scrape_entries")
            .select(`scrape_id, uid, ${metric}`)
            .in("scrape_id", [...timeById.keys()])
            .in("uid", uids)
            .order("scrape_id")
            .range(from, to) as unknown as PromiseLike<{ data: MetricRow[] | null; error: Error | null}>
    )

    return rows.map((r) => ({
        uid: r.uid,
        t: timeById.get(r.scrape_id)!,
        value: r[metric] ?? 0
    }))
}
