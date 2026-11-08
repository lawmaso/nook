import { LeaderboardExplorer } from "@/components/leaderboard-explorer";
import { getLatestLeaderboard, getScrapePeriods } from "@/lib/queries";
import { Suspense } from "react";

async function Explorer() {
    const periods = await getScrapePeriods()
    const initial = await getLatestLeaderboard("monthly", periods[0])

    return (
        <LeaderboardExplorer
            periods={periods}
            initial={initial}    
        />
    )
}

export default function Home() {
    return (
        <main className="max-w p-6">
            <h1 className="text-2xl font-semibold tracking-tight">Nook</h1>
            <Suspense>
                <Explorer />
            </Suspense>
        </main>
    )
}