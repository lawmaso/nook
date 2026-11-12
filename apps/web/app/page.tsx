import { Suspense } from "react";
import { getLatestLeaderboard, getScrapePeriods } from "@/lib/queries";
import { LeaderboardExplorer } from "@/components/leaderboard-explorer";
import { Calculator } from "@/components/calculator/calculator";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

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
            <Tabs className="flex-1 min-h-0 flex flex-col" defaultValue="calculator">
                <TabsList className="ml-auto shrink-0">
                    <TabsTrigger value="graphs">Graphs</TabsTrigger>
                    <TabsTrigger value="calculator">Calculator</TabsTrigger>
                </TabsList>
                <TabsContent value="graphs" className="flex-1 min-h-0">
                    <Suspense>
                        <Explorer />
                    </Suspense>
                </TabsContent>
                <TabsContent value="calculator" className="flex-1 min-h-0">
                    <Suspense>
                        <Calculator />
                    </Suspense>
                </TabsContent>
            </Tabs>
        </main>
    )
}
