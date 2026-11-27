"use client"

import { ChevronDownIcon, ChevronUpIcon } from "lucide-react"
import Link from "next/link"
import { useSearchParams } from "next/navigation"
import { useCallback, useMemo } from "react"

import { ElapsedTicker } from "@/components/relative-time"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Separator } from "@/components/ui/separator"
import { compareStat, pointGap, puzzleGap, timeGap, type Direction, type Verdict } from "@/lib/analysis"
import * as fmt from "@/lib/format"
import type { LeaderboardEntry } from "@/lib/leaderboards"
import { cn } from "@/lib/utils"

const VERDICT_CLASS: Record<Verdict, string> = {
  better: "text-emerald-600 dark:text-emerald-400",
  worse: "text-destructive",
  equal: "text-muted-foreground",
}

const STATS: {
  label: string
  value: (u: LeaderboardEntry) => number
  format: (n: number) => string
  better: Direction
}[] = [
  { label: "Total pts", value: (u) => u.total_points, format: fmt.formatLarge, better: "higher" },
  { label: "Solved", value: (u) => u.solved, format: fmt.formatLarge, better: "higher" },
  { label: "Avg pts", value: (u) => u.avg_points, format: fmt.formatLarge, better: "higher" },
  { label: "Avg time", value: (u) => u.avg_time_s, format: fmt.secondsLabel, better: "lower" },
  { label: "Acceptance", value: (u) => u.acceptance_rate, format: fmt.percent, better: "higher" },
]

function UserPicker({
  label,
  entries,
  index,
  onChange,
}: {
  label: string
  entries: LeaderboardEntry[]
  index: number
  onChange: (index: number) => void
}) {
  const items = useMemo(
    () => entries.map((e) => ({ value: e.uid, label: `#${e.rank} ${e.username}` })),
    [entries],
  )

  return (
    <div className="flex flex-col gap-1.5">
      <span className="text-xs font-medium text-muted-foreground">{label}</span>
      <div className="flex gap-1">
        <Select
          items={items}
          value={entries[index]?.uid ?? null}
          onValueChange={(uid) => {
            const next = entries.findIndex((e) => e.uid === uid)
            if (next >= 0) onChange(next)
          }}
        >
          <SelectTrigger className="min-w-0 flex-1" aria-label={label}>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {items.map((item) => (
              <SelectItem key={item.value} value={item.value}>
                {item.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Button
          variant="outline"
          size="icon"
          aria-label={`${label}: next higher rank`}
          disabled={index <= 0}
          onClick={() => onChange(index - 1)}
        >
          <ChevronUpIcon />
        </Button>
        <Button
          variant="outline"
          size="icon"
          aria-label={`${label}: next lower rank`}
          disabled={index >= entries.length - 1}
          onClick={() => onChange(index + 1)}
        >
          <ChevronDownIcon />
        </Button>
      </div>
    </div>
  )
}

function Stat({ value, label }: { value: string; label: string }) {
  return (
    <div className="flex items-baseline justify-between gap-3 rounded-lg bg-muted/50 px-3 py-2">
      <span className="font-heading text-lg font-semibold tabular-nums">{value}</span>
      <span className="text-right text-xs text-muted-foreground">{label}</span>
    </div>
  )
}

/** Head-to-head comparison, ported from the desktop app's compare panel. */
export function ComparePanel({ entries }: { entries: LeaderboardEntry[] }) {
  const searchParams = useSearchParams()

  const indexOf = useCallback(
    (param: string | null, fallback: number) => {
      const uid = Number(param)
      const i = param ? entries.findIndex((e) => e.uid === uid) : -1
      return i >= 0 ? i : Math.min(fallback, entries.length - 1)
    },
    [entries],
  )

  const ia = indexOf(searchParams.get("a"), 0)
  const ib = indexOf(searchParams.get("b"), 1)

  const select = (which: "a" | "b", index: number) => {
    // Shallow URL update: Next syncs useSearchParams without a server round trip.
    const params = new URLSearchParams(window.location.search)
    params.set(which, String(entries[index].uid))
    window.history.replaceState(null, "", `?${params}`)
  }

  // Desktop orders the pair so u1 is the higher-ranked user.
  const [u1, u2] = entries[ia].rank <= entries[ib].rank ? [entries[ia], entries[ib]] : [entries[ib], entries[ia]]
  const lower = u1.total_points < u2.total_points ? u1 : u2
  const same = u1.uid === u2.uid

  return (
    <Card>
      <CardHeader>
        <CardTitle>Compare users</CardTitle>
        <CardDescription>How far apart two users are, and what it takes to close the gap.</CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        <UserPicker label="User 1" entries={entries} index={ia} onChange={(i) => select("a", i)} />
        <UserPicker label="User 2" entries={entries} index={ib} onChange={(i) => select("b", i)} />

        <Separator />

        {same ? (
          <p className="text-sm text-muted-foreground">Pick two different users to compare.</p>
        ) : (
          <>
            <div className="flex flex-col gap-2">
              <Stat value={fmt.pointGap(pointGap(u1, u2))} label="points behind" />
              <Stat value={fmt.puzzleGap(puzzleGap(u1, u2))} label="puzzles to close gap" />
              <Stat value={fmt.timeGap(timeGap(u1, u2))} label="to close gap" />
            </div>
            <p className="text-xs text-muted-foreground">
              Puzzles: at max score – at {lower.username}&apos;s average. Time: at {lower.username}&apos;s average pace.
            </p>

            <table className="w-full text-sm">
              <thead>
                <tr className="text-left">
                  <th className="pb-2 font-medium text-muted-foreground">Metric</th>
                  <th className="pb-2 font-medium text-emerald-600 dark:text-emerald-400">#{u1.rank}</th>
                  <th className="pb-2 font-medium text-destructive">#{u2.rank}</th>
                </tr>
              </thead>
              <tbody className="tabular-nums">
                {STATS.map(({ label, value, format, better }) => {
                  const [a, b] = [value(u1), value(u2)]
                  const [va, vb] = compareStat(a, b, better)
                  return (
                    <tr key={label} className="border-t">
                      <td className="py-1.5 text-muted-foreground">{label}</td>
                      <td className={cn("py-1.5", VERDICT_CLASS[va])}>{format(a)}</td>
                      <td className={cn("py-1.5", VERDICT_CLASS[vb])}>{format(b)}</td>
                    </tr>
                  )
                })}
              </tbody>
            </table>

            <div className="flex flex-col gap-2 text-sm">
              {[
                { u: u1, color: "text-emerald-600 dark:text-emerald-400" },
                { u: u2, color: "text-destructive" },
              ].map(({ u, color }) => (
                <div key={u.uid} className="flex flex-col">
                  <Link href={`/u/${u.uid}`} className={cn("font-medium underline-offset-4 hover:underline", color)}>
                    [#{u.rank}] {u.username}
                  </Link>
                  <span className="text-muted-foreground">
                    Last active: <ElapsedTicker iso={u.last_active} className="tabular-nums" /> ago
                  </span>
                </div>
              ))}
            </div>
          </>
        )}
      </CardContent>
    </Card>
  )
}
