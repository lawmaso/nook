import type { Enums } from "@/lib/database.types"

export type Kind = Enums<"leaderboard">

export const KINDS = ["monthly", "global"] as const satisfies readonly Kind[]

export const KIND_LABELS: Record<Kind, string> = {
  monthly: "Monthly",
  global: "Global",
}

export function isKind(value: unknown): value is Kind {
  return typeof value === "string" && (KINDS as readonly string[]).includes(value)
}

export function parseKind(value: unknown, fallback: Kind = "monthly"): Kind {
  return isKind(value) ? value : fallback
}

export type LeaderboardEntry = {
  uid: number
  username: string
  rank: number
  total_points: number
  solved: number
  attempted: number
  acceptance_rate: number // 0–100
  avg_time_s: number
  avg_points: number
  last_active: string
}
