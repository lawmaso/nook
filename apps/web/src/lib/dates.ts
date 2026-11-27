import { TZDate } from "@date-fns/tz"
import { addDays, differenceInCalendarDays, format, isValid, parse, startOfMonth, subDays, subMonths, endOfMonth } from "date-fns"

import type { Kind } from "@/lib/leaderboards"

/** The leaderboard site's clock; monthly periods and day buckets use it. */
export const SITE_TZ = "America/New_York"

export type Bucket = "hour" | "day" | "week"

/** An inclusive range of calendar days in site time, as yyyy-MM-dd strings. */
export type DayRange = { from: string; to: string }

const DAY = "yyyy-MM-dd"
// Reference date for parse(); fixed so day parsing never reads the clock.
const REF = new Date(0)

export function siteToday(now: Date = new Date()): TZDate {
  return new TZDate(now, SITE_TZ)
}

export function toDay(d: Date): string {
  return format(d, DAY)
}

function parseDay(value: unknown): Date | null {
  if (typeof value !== "string") return null
  const d = parse(value, DAY, REF)
  return isValid(d) ? d : null
}

export function defaultRange(kind: Kind, now: Date = new Date()): DayRange {
  const today = siteToday(now)
  const from = kind === "monthly" ? startOfMonth(today) : subDays(today, 89)
  return { from: toDay(from), to: toDay(today) }
}

/** Read ?from=&to= (yyyy-MM-dd, site time), falling back to the kind's default. */
export function parseRange(
  params: { from?: unknown; to?: unknown },
  kind: Kind,
  now: Date = new Date(),
): DayRange {
  const from = parseDay(params.from)
  const to = parseDay(params.to)
  if (!from || !to) return defaultRange(kind, now)
  return from <= to ? { from: toDay(from), to: toDay(to) } : { from: toDay(to), to: toDay(from) }
}

/** Convert a day range to [start, end) instants for querying. */
export function rangeBounds({ from, to }: DayRange): { start: string; end: string } {
  const [fy, fm, fd] = from.split("-").map(Number)
  const [ty, tm, td] = to.split("-").map(Number)
  const start = new TZDate(fy, fm - 1, fd, SITE_TZ)
  const end = addDays(new TZDate(ty, tm - 1, td, SITE_TZ), 1)
  return { start: start.toISOString(), end: end.toISOString() }
}

export function rangeDays({ from, to }: DayRange): number {
  return differenceInCalendarDays(parse(to, DAY, REF), parse(from, DAY, REF)) + 1
}

export function pickBucket(range: DayRange): Bucket {
  const days = rangeDays(range)
  if (days <= 3) return "hour"
  if (days <= 120) return "day"
  return "week"
}

export type RangePreset = { label: string; range: DayRange }

export function rangePresets(now: Date = new Date()): RangePreset[] {
  const today = siteToday(now)
  const lastMonth = subMonths(today, 1)
  return [
    { label: "Last 24h", range: { from: toDay(subDays(today, 1)), to: toDay(today) } },
    { label: "Last 7 days", range: { from: toDay(subDays(today, 6)), to: toDay(today) } },
    { label: "Last 30 days", range: { from: toDay(subDays(today, 29)), to: toDay(today) } },
    { label: "This month", range: { from: toDay(startOfMonth(today)), to: toDay(today) } },
    { label: "Last month", range: { from: toDay(startOfMonth(lastMonth)), to: toDay(endOfMonth(lastMonth)) } },
    { label: "Last 90 days", range: { from: toDay(subDays(today, 89)), to: toDay(today) } },
  ]
}

/** Formats a bucket instant for chart axes/tooltips in site time. */
export function formatBucket(iso: string, bucket: Bucket, style: "axis" | "tooltip" = "axis"): string {
  const d = new TZDate(iso, SITE_TZ)
  if (bucket === "hour") return format(d, style === "axis" ? "MMM d, ha" : "MMM d, h:mm a")
  if (bucket === "week") return style === "axis" ? format(d, "MMM d") : `Week of ${format(d, "MMM d, yyyy")}`
  return format(d, style === "axis" ? "MMM d" : "EEE, MMM d, yyyy")
}

export function formatDayLabel(day: string): string {
  return format(parse(day, DAY, REF), "MMM d, yyyy")
}
