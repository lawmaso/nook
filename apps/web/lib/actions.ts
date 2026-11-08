/**
 * @file actions.ts
 * 
 * Thin server actions for client components
 */

"use server"

import { ISODate, Kind, Leaderboard } from "./models";
import { getLatestLeaderboard } from "./queries";

export async function loadLeaderboard(
    kind: Kind, period?: ISODate
): Promise<Leaderboard | null> {
    return getLatestLeaderboard(kind, period)
}
