import { cacheLife, cacheTag } from "next/cache"

import type { FunctionReturns, Tables } from "@/lib/database.types"
import type { Bucket } from "@/lib/dates"
import type { Kind, LeaderboardEntry } from "@/lib/leaderboards"
import { supabase } from "@/lib/supabase"

export type Scrape = Pick<Tables<"scrapes">, "scrape_id" | "kind" | "period" | "scraped_at" | "duration_ms">

export type Leaderboard = {
  scrape: Scrape | null
  entries: LeaderboardEntry[]
}

export type MetricsRow = FunctionReturns<"leaderboard_metrics">[number]
export type HistoryRow = FunctionReturns<"user_history">[number]

/** RPC results carry a flag for "migration not applied" instead of throwing. */
export type RpcResult<T> = { rows: T[]; missingRpc: boolean }

const PAGE_SIZE = 1000 // PostgREST's default max-rows

const ENTRY_COLUMNS =
  "uid, rank, total_points, solved, attempted, acceptance_rate, avg_time_s, avg_points, last_active, users(username)"

async function getLatestScrape(kind: Kind): Promise<Scrape | null> {
  "use cache"
  cacheLife("scrape")
  cacheTag(`scrapes:${kind}`)

  // Same as GetLatestScrapes in db/queries.sql
  const { data, error } = await supabase
    .from("scrapes")
    .select("scrape_id, kind, period, scraped_at, duration_ms")
    .eq("kind", kind)
    .order("scraped_at", { ascending: false })
    .limit(1)
    .maybeSingle()

  if (error) throw new Error(`Failed to load latest ${kind} scrape: ${error.message}`)
  return data
}

export async function getLatestLeaderboard(kind: Kind): Promise<Leaderboard> {
  "use cache"
  cacheLife("scrape")
  cacheTag(`scrapes:${kind}`)

  const scrape = await getLatestScrape(kind)
  if (!scrape) return { scrape: null, entries: [] }

  // Same as GetScrapeEntries in db/queries.sql, paged past PostgREST's row cap.
  const entries: LeaderboardEntry[] = []
  for (let offset = 0; ; offset += PAGE_SIZE) {
    const { data, error } = await supabase
      .from("scrape_entries")
      .select(ENTRY_COLUMNS)
      .eq("scrape_id", scrape.scrape_id)
      .order("rank", { ascending: true })
      .range(offset, offset + PAGE_SIZE - 1)

    if (error) throw new Error(`Failed to load ${kind} leaderboard: ${error.message}`)

    for (const { users, ...entry } of data) {
      entries.push({ ...entry, username: users?.username ?? `uid ${entry.uid}` })
    }
    if (data.length < PAGE_SIZE) break
  }

  return { scrape, entries }
}

export type UserProfile = {
  user: Tables<"users">
  current: Record<Kind, (LeaderboardEntry & { scraped_at: string }) | null>
}

export async function getUserProfile(uid: number): Promise<UserProfile | null> {
  "use cache"
  cacheLife("scrape")
  cacheTag(`user:${uid}`)

  const { data: user, error } = await supabase
    .from("users")
    .select("uid, username, member_since")
    .eq("uid", uid)
    .maybeSingle()

  if (error) throw new Error(`Failed to load user ${uid}: ${error.message}`)
  if (!user) return null

  const current = await Promise.all(
    (["monthly", "global"] as const).map(async (kind) => {
      const scrape = await getLatestScrape(kind)
      if (!scrape) return null

      const { data, error } = await supabase
        .from("scrape_entries")
        .select(ENTRY_COLUMNS)
        .eq("scrape_id", scrape.scrape_id)
        .eq("uid", uid)
        .maybeSingle()

      if (error) throw new Error(`Failed to load ${kind} standing: ${error.message}`)
      if (!data) return null

      const { users, ...entry } = data
      return { ...entry, username: users?.username ?? user.username, scraped_at: scrape.scraped_at }
    }),
  )

  return { user, current: { monthly: current[0], global: current[1] } }
}

function isMissingFunction(error: { code?: string } | null): boolean {
  // PGRST202: function not found in the schema cache (migration not applied)
  return error?.code === "PGRST202"
}

export async function getLeaderboardMetrics(
  kind: Kind,
  start: string,
  end: string,
  bucket: Bucket,
): Promise<RpcResult<MetricsRow>> {
  "use cache"
  cacheLife("scrape")
  cacheTag(`scrapes:${kind}`)

  const { data, error } = await supabase.rpc("leaderboard_metrics", {
    p_kind: kind,
    p_from: start,
    p_to: end,
    p_bucket: bucket,
  })

  if (isMissingFunction(error)) return { rows: [], missingRpc: true }
  if (error) throw new Error(`Failed to load ${kind} metrics: ${error.message}`)
  return { rows: data ?? [], missingRpc: false }
}

export async function getUserHistory(
  uid: number,
  kind: Kind,
  start: string,
  end: string,
  bucket: Bucket,
): Promise<RpcResult<HistoryRow>> {
  "use cache"
  cacheLife("scrape")
  cacheTag(`user:${uid}`)

  const { data, error } = await supabase.rpc("user_history", {
    p_uid: uid,
    p_kind: kind,
    p_from: start,
    p_to: end,
    p_bucket: bucket,
  })

  if (isMissingFunction(error)) return { rows: [], missingRpc: true }
  if (error) throw new Error(`Failed to load history for user ${uid}: ${error.message}`)
  return { rows: data ?? [], missingRpc: false }
}

export type UserHit = Pick<Tables<"users">, "uid" | "username">

export async function searchUsers(query: string): Promise<UserHit[]> {
  "use cache"
  cacheLife("scrape")

  const q = query.trim()
  if (!q) return []

  const escaped = q.replace(/[\\%_]/g, (c) => `\\${c}`)
  const { data, error } = await supabase
    .from("users")
    .select("uid, username")
    .ilike("username", `%${escaped}%`)
    .order("username")
    .limit(10)

  if (error) throw new Error(`User search failed: ${error.message}`)
  return data
}
