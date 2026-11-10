/**
 * @file actions.ts
 * 
 * Thin server actions for client components
 */

"use server"

import { ISODate, Kind, Leaderboard, TrajectoryMetric, TrajectoryPoint } from "./models";
import { getLatestLeaderboard, getTrajectories } from "./queries";

export async function loadLeaderboard(
    kind: Kind, period?: ISODate
): Promise<Leaderboard | null> {
    return getLatestLeaderboard(kind, period)
}

const METRICS = new Set<TrajectoryMetric>([
    "total_points", "solved", "attempted",
    "acceptance_rate", "avg_time_s", "avg_points"
])
const MAX_UIDS = 25

export async function loadTrajectories(
    kind: Kind, period: ISODate | undefined,
    uids: number[], metric: TrajectoryMetric
): Promise<TrajectoryPoint[]> {
    if (!METRICS.has(metric)) throw new Error("Unknown metric")
        return getTrajectories(kind, period, uids.slice(0, MAX_UIDS), metric)
}
