// Port of apps/desktop/core/analysis.py and the stat-direction rules in
// apps/desktop/ui/components/comparison.py.

import type { LeaderboardEntry } from "@/lib/leaderboards"

/** Max points a single puzzle can award (used for the "perfect play" estimate). */
export const MAX_PUZZLE_POINTS = 563

type Scored = Pick<LeaderboardEntry, "total_points" | "avg_points" | "avg_time_s">

function lower<T extends Scored>(u1: T, u2: T): T {
  return u1.total_points < u2.total_points ? u1 : u2
}

export function pointGap(u1: Scored, u2: Scored): number {
  return Math.abs(u2.total_points - u1.total_points)
}

/** [puzzles at max score, puzzles at the lower user's average score] */
export function puzzleGap(u1: Scored, u2: Scored): [number, number] {
  const lo = lower(u1, u2)
  const gap = pointGap(u1, u2)
  return [gap / MAX_PUZZLE_POINTS, gap / lo.avg_points]
}

/** Seconds the lower user needs at their average pace to close the gap. */
export function timeGap(u1: Scored, u2: Scored): number {
  const lo = lower(u1, u2)
  return puzzleGap(u1, u2)[1] * lo.avg_time_s
}

export type Direction = "higher" | "lower"
export type Verdict = "better" | "worse" | "equal"

/** Which side wins a stat, given whether higher or lower is better. */
export function compareStat(a: number, b: number, better: Direction): [Verdict, Verdict] {
  if (a === b) return ["equal", "equal"]
  const aWins = better === "higher" ? a > b : a < b
  return aWins ? ["better", "worse"] : ["worse", "better"]
}
