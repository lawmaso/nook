/**
 * @file leaderboard-explorer.tsx
 */

"use client"

import React, { useMemo, useState } from "react"
import { ISODate, Kind, Leaderboard } from "@/lib/models"
import { loadLeaderboard } from "@/lib/actions"
import { Tabs, TabsList, TabsTrigger } from "./ui/tabs"
import { Calendar } from "./ui/calendar"
import { DateRange } from "react-day-picker"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "./ui/card"
import { Select, SelectContent, SelectGroup, SelectItem, SelectLabel, SelectSeparator, SelectTrigger, SelectValue } from "./ui/select"
import { GRAPHS } from "./graphs"
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from "./ui/empty"
import { FolderOpenIcon } from "lucide-react"
import { cn } from "@/lib/utils"

const monthKey = (iso: string): string => iso.slice(0, 7)
const dateMonthKey = (d: Date): string => `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}`
const parseISODate = (iso: ISODate): Date => {
    const [y, m, d] = iso.slice(0, 10).split("-").map(Number)
    return new Date(y, m - 1, d)
}

const nOptions = [
    { label: "Top 10", value: 10 },
    { label: "Top 15", value: 15 },
    { label: "Top 25", value: 25 },
    { label: "Top 50", value: 50 },
    { label: "Top 75", value: 75 },
    { label: "All", value: 100 },
]

type ExplorerProps = {
    periods: ISODate[]
    initial: Leaderboard | null
}

export function LeaderboardExplorer({ periods, initial }: ExplorerProps) {
    const [kind, setKind]               = useState<Kind>("monthly")
    const [period, setPeriod]           = useState<ISODate | undefined>(initial?.scrape.period ?? periods[0])
    const [leaderboard, setLeaderboard] = useState<Leaderboard | null>(initial)
    const [graphId, setGraphId]         = useState(GRAPHS[0].graphs[0].id)
    const [n, setN]                     = useState(100)
    const [loading, setLoading]         = useState(false)
    const [error, setError]             = useState<string | null>(null)
    const [dateRange, setDateRange]     = useState<DateRange | undefined>({
        from: parseISODate(period!),
        to: new Date(Date.now())
    })

    const graph = GRAPHS
        .flatMap((group) => group.graphs)
        .find((g) => g.id === graphId) ?? GRAPHS[0].graphs[0]

    const periodByMonth = useMemo(
        () => new Map(periods.map((p) => [monthKey(p), p])),
        [periods]
    )
    async function load(nextKind: Kind, nextPeriod?: ISODate) {
        setLoading(true)
        setError(null)

        try {
            setLeaderboard(await loadLeaderboard(nextKind, nextPeriod))
        } catch (e) {
            setError(e instanceof Error ? e.message : "fail")
        } finally {
            setLoading(false)
        }
    }

    function handleKind(next: string) {
        const k = next as Kind
        setKind(k)
        load(k, k === "monthly" ? period : undefined)
    }

    return (
        <div className="grid h-full min-h-0 gap-6 lg:grid-cols-[auto_1fr]">
            <div className="flex flex-col gap-4">
                <Tabs value={kind} onValueChange={handleKind}>
                    <TabsList className="w-full">
                        <TabsTrigger value="monthly">Monthly</TabsTrigger>
                        <TabsTrigger value="global">Global</TabsTrigger>
                    </TabsList>
                </Tabs>

                <Calendar
                    mode="range"
                    selected={dateRange}
                    defaultMonth={period ? parseISODate(period) : undefined }
                    disabled={(d: Date) => {
                        const periodExists = periodByMonth.has(dateMonthKey(d))
                        return !periodExists
                    }}
                    captionLayout="dropdown"
                    className="border"
                />
            </div>

            <Card className="min-h-0 flex flex-col overflow-hidden">
                <CardHeader className="gap-4">
                    <div className="flex flex-wrap items-start justify-between gap-4">
                        <div>
                            <CardTitle className="text-lg">{graph.title}</CardTitle>
                            <CardDescription className="text-md">{graph.description}</CardDescription>
                        </div>
                        <div className="flex gap-2">
                            <Select value={graph.title} onValueChange={(v) => v && setGraphId(v)}>
                                <SelectTrigger className="w-[220px]">
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    {GRAPHS.map((group, idx) => (
                                        <React.Fragment key={group.title}>
                                            <SelectGroup>
                                                <SelectLabel>{group.title}</SelectLabel>
                                                {group.graphs.map((g) => (
                                                    <SelectItem key={g.id} value={g.id}>
                                                        {g.title}
                                                    </SelectItem>
                                                ))}
                                            </SelectGroup>
                                            {idx < GRAPHS.length && <SelectSeparator />}
                                        </React.Fragment>
                                    ))}
                                </SelectContent>
                            </Select>

                            {graph.usesN && (
                                <Select items={nOptions} defaultValue={100} onValueChange={(v) => v && setN(v)}>
                                    <SelectTrigger>
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent alignItemWithTrigger>
                                        <SelectGroup>
                                            {nOptions.map((item) => (
                                                <SelectItem key={item.value} value={item.value}>
                                                    {item.label}
                                                </SelectItem>
                                            ))}
                                        </SelectGroup>
                                    </SelectContent>
                                </Select>
                            )}
                        </div>
                    </div>
                </CardHeader>
                <CardContent className={cn(
                    "min-h-0 flex-1 transition-opacity",
                    loading ? "opacity-50" : ""
                )}>
                        {error ? (
                            <p className="text-sm text-destructive">{error}</p>
                        ) : !leaderboard || leaderboard.entries.length === 0 ? (
                            <Empty>
                                <EmptyHeader>
                                    <EmptyMedia variant="icon">
                                        <FolderOpenIcon />
                                    </EmptyMedia>
                                    <EmptyTitle>No Scrapes Found</EmptyTitle>
                                    <EmptyDescription>
                                        The selected scrape range hasn't happened yet, or we're too far into the future
                                    </EmptyDescription>
                                </EmptyHeader>
                            </Empty>
                        ) : (
                            graph.render({
                                entries: leaderboard.entries,
                                n,
                                kind,
                                period
                            })
                        )}
                </CardContent>
            </Card>
        </div>
    )
}