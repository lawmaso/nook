"use client"

import { columns } from "./columns"
import { DataTable } from "./data-table"
import { getLatestScrapeEntries } from "@/lib/queries"
import { Selector } from "./selector"
import { useEffect, useState } from "react"
import { LeaderboardEntry } from "@/lib/models"

type Selection = LeaderboardEntry | null

export function Calculator() {
    const [entries, setEntries] = useState<LeaderboardEntry[] | null>(null)
    const [first, setFirst] = useState<Selection>(null)
    const [second, setSecond] = useState<Selection>(null)

    useEffect(() => {
        let cancelled = false

        getLatestScrapeEntries("monthly")
            .then((r) => !cancelled && setEntries(r))
        
        return () => {
            cancelled = true
        }
    })

    if (!entries) return
    return (
        <div className="h-full min-h-0 flex">
            <DataTable columns={columns} data={entries} />
            <div className="flex flex-col">
                <Selector
                    entries={entries}
                    value={first}
                    onValueChange={setFirst}
                    placeholder="Select first user"
                />
                <Selector
                    entries={entries}
                    value={second}
                    onValueChange={setSecond}
                    placeholder="Select second user"
                />
            </div>
        </div>
    )
}
