/**
 * @file actions.ts
 * 
 * Thin server actions for client components
 */

"use server"

import { ISODate, Kind, Leaderboard, TrajectoryPoint } from "./models";
import { getLatestLeaderboard, getTrajectories } from "./queries";

export async function loadLeaderboard(
    kind: Kind, period?: ISODate
): Promise<Leaderboard | null> {
    return getLatestLeaderboard(kind, period)
}

export async function loadTrajectories(
    kind: Kind, period: ISODate | undefined,
    uids: number[]
): Promise<TrajectoryPoint[]> {
    return getTrajectories(kind, period, uids)
}
