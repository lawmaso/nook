import { Suspense } from "react";
import { getLatestLeaderboard, getScrapePeriods } from "@/lib/queries";
import { LeaderboardExplorer } from "@/components/leaderboard-explorer";

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
        <main className="h-dvh overflow-hidden p-6 flex flex-col">
            <h1 className="text-2xl font-semibold tracking-tight mb-2">Nook</h1>
            <Suspense>
                <Explorer />
            </Suspense>
        </main>
    )
}
