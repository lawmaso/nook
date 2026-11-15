/**
 * @file calculator.tsx
 */

"use client"

import { useEffect, useMemo, useState } from "react"
import { ArrowDownIcon, BracketsIcon, ClockArrowRightIcon, UsersIcon } from "lucide-react"
import { getLatestScrapeEntries } from "@/lib/queries"
import { Kind, LeaderboardEntry } from "@/lib/models"

import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "../ui/table"
import { Popover, PopoverTrigger, PopoverContent } from "../ui/popover"
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from "../ui/empty"

import { columns } from "./columns"
import { DataTable } from "./data-table"
import { Selector } from "./selector"
import { formatDecimal, formatDuration, formatLarge, formatPercent } from "./format"
import { Tabs, TabsList, TabsTrigger } from "../ui/tabs"
import { Spinner } from "../ui/spinner"

type Selection = LeaderboardEntry | null

function formatSigned(
    value: number,
    formatter: (value: number) => string
) {
    const sign =
        value > 0 ? "+" :
        value < 0 ? "-" :
        ""

    return `${sign}${formatter(Math.abs(value))}`
}

export function Calculator() {
    const [kind, setKind] = useState<Kind>("global")
    const [entries, setEntries] = useState<LeaderboardEntry[] | null>(null)
    const [userOne, setUserOne] = useState<Selection>(null)
    const [userTwo, setUserTwo] = useState<Selection>(null)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        let cancelled = false
        setLoading(true)
        setError(null)
        setUserOne(null)
        setUserTwo(null)

        getLatestScrapeEntries(kind)
            .then((result) => {
                if (cancelled) return

                let recentEntries = result ?? []
                setEntries(recentEntries)
                // setUserOne(recentEntries[0] ?? null)
                // setUserTwo(recentEntries[1] ?? null)
            })
            .catch((error: unknown) => {
                console.error()

                if (!cancelled) {
                    setEntries([])
                    setError(error instanceof Error
                        ? error.message
                        : "Couldn't load entries"
                    )
                }
            })
            .finally(() => {
                setLoading(false)
            })

        return () => {
            cancelled = true
        }
    }, [kind])

    const comparison = useMemo((): ComparisonResult | null => {
        if (!userOne || !userTwo) return null

        // NOTE: lower is the user with less points
        const [lower, higher] = userOne.rank > userTwo.rank
            ? [userOne, userTwo]
            : [userTwo, userOne]

        const MAX_POINTS_PER_PUZZLE = 563

        const gap             = higher.total_points - lower.total_points
        const minPuzzles      = gap / MAX_POINTS_PER_PUZZLE
        const expectedPuzzles = gap / lower.avg_points
        const minTime         = minPuzzles * lower.avg_time_s
        const expectedTime    = expectedPuzzles * lower.avg_time_s
        
        return {
            lower,
            higher,
            gap,
            minPuzzles,
            expectedPuzzles,
            minTime,
            expectedTime
        }
    }, [userOne, userTwo])

    function handleKind(next: string) {
        const k = next as Kind
        setKind(k)
    }

    if (!entries || loading) {
        return (
            <Empty className="border h-full">
                <EmptyHeader>
                    <EmptyMedia variant="icon">
                        <Spinner />
                    </EmptyMedia>
                    <EmptyTitle>Loading leaderboard</EmptyTitle>
                    <EmptyDescription className="text-sm">
                        Please wait while the leaderboard loads
                    </EmptyDescription>
                </EmptyHeader>
            </Empty>
        )
    }

    return (
        <div className="h-full min-h-0 flex gap-4">
            <DataTable
                columns={columns}
                data={entries}
            />

            <div className="flex min-w-0 flex-1 flex-col gap-4">
                {error && (
                    <p className="text-sm text-destructive">
                        {error}
                    </p>
                )}

                <div>
                    <Tabs value={kind} onValueChange={handleKind}>
                    <TabsList className="w-full">
                        <TabsTrigger value="monthly">Monthly</TabsTrigger>
                        <TabsTrigger value="global">Global</TabsTrigger>
                    </TabsList>
                </Tabs>
                </div>
                <div className="flex gap-2">
                    <Selector
                        entries={entries}
                        value={userOne}
                        onValueChange={setUserOne}
                    />
                    <Selector
                        entries={entries}
                        value={userTwo}
                        onValueChange={setUserTwo}
                    />
                </div>

                {comparison ? (
                    <ComparisonResults comparison={comparison} />
                ) : (
                    <Empty className="border border-solid">
                        <EmptyHeader>
                            <EmptyMedia variant="icon">
                                <UsersIcon />
                            </EmptyMedia>
                            <EmptyTitle>No users selected</EmptyTitle>
                            <EmptyDescription className="text-sm">
                                Select two users to compare individual metrics and estimate
                                the gap between them
                            </EmptyDescription>
                        </EmptyHeader>
                    </Empty>
                )}
            </div>
        </div>
    )
}

type ComparisonResult = {
    lower:           LeaderboardEntry
    higher:          LeaderboardEntry
    gap:             number
    minPuzzles:      number
    expectedPuzzles: number
    minTime:         number
    expectedTime:    number
}

type TableRowData = {
    category:         string
    higherUserMetric: number
    lowerUserMetric:  number
    formatter:        (v: number) => string
    invert?:          boolean
}

function ComparisonResults({ comparison }: { comparison: ComparisonResult }) {
    const { higher, lower } = comparison 
    const higherLabel = `@${higher.user.username}`
    const lowerLabel  = `@${lower.user.username}`

    const tableRowData: TableRowData[] = [
        {
            category: "Total points",
            higherUserMetric: higher.total_points,
            lowerUserMetric: lower.total_points,
            formatter: formatLarge
        },
        {
            category: "Solved",
            higherUserMetric: higher.solved,
            lowerUserMetric: lower.solved,
            formatter: formatLarge
        },
        {
            category: "Average points",
            higherUserMetric: higher.avg_points,
            lowerUserMetric: lower.avg_points,
            formatter: formatDecimal
        },
        {
            category: "Average time",
            higherUserMetric: higher.avg_time_s,
            lowerUserMetric: lower.avg_time_s,
            formatter: formatDuration,
            invert: true
        },
        {
            category: "Acceptance",
            higherUserMetric: higher.acceptance_rate,
            lowerUserMetric: lower.acceptance_rate,
            formatter: formatPercent
        },

    ]

    return (
        <div className="flex flex-row-reverse border p-4">
            <div className="w-xs flex items-center justify-center">
                <div className="flex flex-col gap-3">
                    <Popover>
                        <PopoverTrigger nativeButton={false} delay={10} closeDelay={100} render={
                            <div className="flex flex-col items-center">
                                <div className="flex items-center gap-2">
                                    <ArrowDownIcon className="h-5 w-5" color="var(--chart-1)" />
                                    <p className="text-2xl cursor-pointer hover:underline whitespace-nowrap">{formatLarge(comparison.gap)}</p>
                                </div>
                                <p className="text-sm">points behind</p>
                            </div>
                        }/>
                        <PopoverContent className="flex w-64 flex-col gap-0.5">
                            <div>
                                <strong className="whitespace-nowrap font-bold">@{lower.user.username}{" "}</strong>
                                is this many points behind <strong className="whitespace-nowrap font-bold">@{higher.user.username}{" "}</strong>
                            </div>
                        </PopoverContent>
                    </Popover>

                    <Popover>
                        <PopoverTrigger nativeButton={false} delay={10} closeDelay={100} render={
                            <div className="flex flex-col items-center">
                                <div className="flex items-center gap-2">
                                    <BracketsIcon className="h-5 w-5" color="var(--chart-1)" />
                                    <p className="text-2xl cursor-pointer hover:underline whitespace-nowrap">
                                        {formatLarge(comparison.minPuzzles)}
                                        {"-"}
                                        {formatLarge(comparison.expectedPuzzles)}
                                    </p>
                                </div>
                                <p className="text-sm">puzzles to close gap</p>
                            </div>
                        }/>
                        <PopoverContent className="flex w-64 flex-col gap-0.5">
                            <div>
                                An estimate of the puzzles <strong className="whitespace-nowrap font-bold">@{lower.user.username}{" "}</strong>
                                must complete to reach <strong className="whitespace-nowrap font-bold">@{higher.user.username}{" "}</strong>
                            </div>
                        </PopoverContent>
                    </Popover>

                    <Popover>
                        <PopoverTrigger nativeButton={false} delay={10} closeDelay={100} render={
                            <div className="flex flex-col items-center">
                                <div className="flex items-center gap-2">
                                    <ClockArrowRightIcon className="h-5 w-5" color="var(--chart-1)" />
                                    <p className="text-2xl cursor-pointer hover:underline whitespace-nowrap">{formatDuration(comparison.expectedTime)}</p>
                                </div>
                                <p className="text-sm">to close gap</p>
                            </div>
                        }/>
                        <PopoverContent className="flex w-64 flex-col gap-0.5">
                            <div>
                                An estimate of the time gap between <strong className="whitespace-nowrap font-bold">@{lower.user.username}{" "}</strong>
                                and <strong className="whitespace-nowrap font-bold">@{higher.user.username}{" "}</strong>
                            </div>
                        </PopoverContent>
                    </Popover>
                </div>
            </div>

            <Table>
                <TableHeader>
                    <TableRow>
                        <TableHead className="font-bold text-lg">Metric</TableHead>
                        <TableHead className="font-bold text-lg">{higherLabel}</TableHead>
                        <TableHead className="font-bold text-lg">{lowerLabel}</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {tableRowData.map((row) => {
                        const higherUserLabel = row.formatter(row.higherUserMetric)
                        const lowerUserLabel  = row.formatter(row.lowerUserMetric)

                        const isEqual = row.higherUserMetric == row.lowerUserMetric
                        const isHigherBetter = !row.invert
                            ? row.higherUserMetric > row.lowerUserMetric
                            : row.higherUserMetric < row.lowerUserMetric

                        const baseStyle   = "text-sm font-bold"
                        const betterStyle = `${baseStyle} text-green-400`
                        const worseStyle  = `${baseStyle} text-red-400`

                        return (
                            <TableRow key={row.category}>
                                <TableCell className="text-sm">{row.category}</TableCell>
                                {/* <TableCell className={isEqual ? "text-sm" : (isHigherBetter ? betterStyle : worseStyle)}>{higherUserLabel}</TableCell> */}
                                {/* <TableCell className={isEqual ? "text-sm" : (isHigherBetter ? worseStyle : betterStyle)}>{lowerUserLabel}</TableCell> */}
                                <TableCell className="text-sm">{higherUserLabel}</TableCell>
                                <TableCell className="text-sm">{lowerUserLabel}</TableCell>
                            </TableRow>
                        )
                    })}
                </TableBody>
            </Table>
        </div>
    )
}
