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

const binOptions = [
    { label: "10 bins", value: 10 },
    { label: "15 bins", value: 15 },
    { label: "25 bins", value: 25 },
    { label: "50 bins", value: 50 },
    { label: "75 bins", value: 75 },
    { label: "100 bins", value: 100 },
]

type ExplorerProps = {
    periods: ISODate[]
    initial: Leaderboard | null
}

export function LeaderboardExplorer({ periods, initial }: ExplorerProps) {
    const [kind, setKind]               = useState<Kind>("monthly")
    const [period, setPeriod]           = useState<ISODate | undefined>(initial?.scrape.period ?? periods[0])
    const [leaderboard, setLeaderboard] = useState<Leaderboard | null>(initial)
    const [graphId, setGraphId]         = useState(GRAPHS[GRAPHS.length-1].graphs[0].id)
    const [n, setN]                     = useState(100)
    const [bins, setBins]               = useState(15)
    const [loading, setLoading]         = useState(false)
    const [error, setError]             = useState<string | null>(null)
    const [dateRange, setDateRange]     = useState<DateRange | undefined>({
        from: parseISODate(period!),
        to: new Date(Date.now())
    })

    const graph = GRAPHS
        .flatMap((group) => group.graphs)
        .find((g) => g.id === graphId) ?? GRAPHS[GRAPHS.length-1].graphs[0]

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
        <div className="grid grid-cols-1 gap-4 lg:h-full lg:min-h-0 lg:grid-cols-[auto_minmax(0,1fr)] lg:gap-6">
            <div className="flex min-w-0 flex-col gap-4 lg:min-h-0 lg:overflow-y-auto">
                <Tabs value={kind} onValueChange={handleKind}>
                    <TabsList className="w-full">
                        <TabsTrigger value="monthly">Monthly</TabsTrigger>
                        <TabsTrigger value="global">Global</TabsTrigger>
                    </TabsList>
                </Tabs>

                <div className="max-w-full overflow-x-auto">
                    <Calendar
                        mode="range"
                        selected={dateRange}
                        defaultMonth={period ? parseISODate(period) : undefined}
                        disabled={(d: Date) => !periodByMonth.has(dateMonthKey(d))}
                        captionLayout="dropdown"
                        className="mx-auto border lg:mx-0"
                    />
                </div>
            </div>

            <Card className="flex min-w-0 flex-col overflow-hidden lg:min-h-0">
                <CardHeader className="gap-4">
                    <div className="flex flex-col gap-4 sm:flex-row sm:flex-wrap sm:items-start sm:justify-between">
                        <div className="min-w-0">
                            <CardTitle className="text-lg">{graph.title}</CardTitle>
                            <CardDescription className="text-base">{graph.description}</CardDescription>
                        </div>

                        <div className="flex w-full flex-wrap gap-2 sm:w-auto">
                            <Select value={graph.title} onValueChange={(v) => v && setGraphId(v)}>
                                <SelectTrigger className="w-full sm:w-[220px]">
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
                                            {idx < GRAPHS.length - 1 && <SelectSeparator />}
                                        </React.Fragment>
                                    ))}
                                </SelectContent>
                            </Select>

                            {graph.usesN && (
                                <Select items={nOptions} value={n} onValueChange={(v) => v && setN(v)}>
                                    <SelectTrigger className="min-w-0 flex-1 sm:flex-none">
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

                            {graph.usesBins && (
                                <Select items={binOptions} value={bins} onValueChange={(v) => v && setBins(v)}>
                                    <SelectTrigger className="min-w-0 flex-1 sm:flex-none">
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent alignItemWithTrigger>
                                        <SelectGroup>
                                            {binOptions.map((item) => (
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

                <CardContent
                    className={cn(
                        "h-100 min-w-0 w-full transition-opacity sm:h-[30rem] lg:h-auto lg:min-h-0 lg:flex-1",
                        loading && "opacity-50"
                    )}
                >
                    {error ? (
                        <p className="text-sm text-destructive">{error}</p>
                    ) : !leaderboard || leaderboard.entries.length === 0 ? (
                        <Empty className="h-full">
                            <EmptyHeader>
                                <EmptyMedia variant="icon">
                                    <FolderOpenIcon />
                                </EmptyMedia>
                                <EmptyTitle>No scrapes found</EmptyTitle>
                                <EmptyDescription>
                                    The selected date range hasn't happened yet, or we're too far into the future
                                </EmptyDescription>
                            </EmptyHeader>
                        </Empty>
                    ) : (
                        graph.render({
                            entries: leaderboard.entries,
                            n,
                            bins,
                            kind,
                            period,
                        })
                    )}
                </CardContent>
            </Card>
        </div>
    )
}