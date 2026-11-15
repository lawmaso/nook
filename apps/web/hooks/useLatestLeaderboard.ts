/**
 * @file useLatestLeaderboard.ts
 */

import { useQuery, useQueryClient } from "@tanstack/react-query"
import { Kind, Leaderboard } from "@/lib/models"
import { getLatestScrape, getScrapeEntries } from "@/lib/queries"

const SCRAPE_INTERVAL_MS: Record<Kind, number> = {
    monthly: 60 * 60_000,
    global:  60 * 60_000 * 2
}

const VERSION_CHECK_MS = 5 * 60_000

export function useLatestLeaderboard(kind: Kind) {
    const queryClient = useQueryClient()
    const queryKey = ["latest-leaderboard", kind] as const

    return useQuery({
        queryKey,
        queryFn: async (): Promise<Leaderboard | null> => {
            const scrape = await getLatestScrape(kind)
            if (!scrape) return null

            const cached = queryClient.getQueryData<Leaderboard | null>(queryKey)
            if (cached?.scrape?.scrape_id == scrape.scrape_id) return cached
            
            return {
                scrape,
                entries: await getScrapeEntries(scrape.scrape_id)
            }
        },
        staleTime: Math.min(SCRAPE_INTERVAL_MS[kind] / 20, VERSION_CHECK_MS),
        refetchInterval: VERSION_CHECK_MS,
        refetchOnWindowFocus: true
    })
}
