/**
 * @file calculator.tsx
 */

"use client"

import { ReactNode, useMemo, useState } from "react"
import { ArrowDownIcon, BracketsIcon, ClockArrowRightIcon, UsersIcon } from "lucide-react"
import { ISODate, Kind, LeaderboardEntry } from "@/lib/models"

import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "../ui/table"
import { Popover, PopoverTrigger, PopoverContent } from "../ui/popover"
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from "../ui/empty"
import { Tabs, TabsList, TabsTrigger } from "../ui/tabs"
import { Spinner } from "../ui/spinner"

import { columns } from "./columns"
import { DataTable } from "./data-table"
import { Selector } from "./selector"
import { formatDecimal, formatDuration, formatLarge, formatPercent } from "./format"
import { useLatestLeaderboard } from "@/hooks/useLatestLeaderboard"
import { LucideIcon } from "lucide-react/types"

type Selection = LeaderboardEntry | null
type SelectedIds = {
    userOne: number | null
    userTwo: number | null
}

const INITIAL_KIND: Kind = "global"
const NO_ENTRIES: LeaderboardEntry[] = []
const NO_SELECTION: SelectedIds = { userOne: null, userTwo: null }

function formatAge(iso: ISODate): string {
    const minutes = Math.max(0, Math.round((Date.now() - new Date(iso).getTime()) / 60_000))

    if (minutes < 1) return "just now"
    if (minutes < 60) return `${minutes} minutes ago`
    return `${Math.round(minutes / 60)}h ago`
}

export function Calculator() {
    const [kind, setKind] = useState<Kind>(INITIAL_KIND)
    const [selected, setSelected] = useState<SelectedIds>(NO_SELECTION)

    const { data, isFetching, isError, error } = useLatestLeaderboard(kind)
    const entries = data?.entries ?? NO_ENTRIES

    const userOne: Selection = useMemo(
        () => entries.find((e) => e.user.uid === selected.userOne) ?? entries[0] ?? null,
        [entries, selected.userOne]
    )

    const userTwo: Selection = useMemo(
        () => entries.find((e) => e.user.uid === selected.userTwo) ?? entries[1] ?? null,
        [entries, selected.userTwo]
    )

    const comparison = useMemo((): ComparisonResult | null => {
        if (!userOne || !userTwo) return null

        // NOTE: lower is the user with a "lower" rank
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
        setKind(next as Kind)
        setSelected(NO_SELECTION)
    }

    return (
        <div className="grid grid-cols-1 gap-4 lg:h-full lg:min-h-0 lg:grid-cols-2 lg:grid-rows-[auto_minmax(0,1fr)]">
            <div className="flex flex-col gap-2 lg:col-start-2 lg:row-start-1">
                {isError && (
                    <p className="text-sm text-destructive">
                        {error instanceof Error ? error.message : "Couldn't load leaderboard"}
                        {entries.length > 0 && ". Showing latest leaderboard"}
                    </p>
                )}

                <Tabs value={kind} onValueChange={handleKind}>
                    <TabsList className="w-full">
                        <TabsTrigger value="monthly">Monthly</TabsTrigger>
                        <TabsTrigger value="global">Global</TabsTrigger>
                    </TabsList>
                </Tabs>

                {data && (
                    <p className="text-xs text-muted-foreground">
                        Scraped {formatAge(data.scrape.scraped_at)}
                    </p>
                )}
            </div>

            <div className="flex max-h-[28rem] min-h-0 flex-col overflow-auto lg:col-start-1 lg:row-span-2 lg:row-start-1 lg:max-h-none lg:h-full">
                {isFetching ? (
                    <Empty className="h-full min-h-64 border border-solid">
                        <EmptyHeader>
                            <EmptyMedia variant="icon"><Spinner /></EmptyMedia>
                            <EmptyTitle>Loading leaderboard</EmptyTitle>
                            <EmptyDescription className="text-sm">
                                Please wait while the leaderboard loads
                            </EmptyDescription>
                        </EmptyHeader>
                    </Empty>
                ) : (
                    <DataTable columns={columns} data={entries} />
                )}
            </div>

            <div className="flex min-w-0 flex-col gap-4 lg:col-start-2 lg:row-start-2 lg:min-h-0 lg:overflow-y-auto">
                <div className="flex flex-col gap-2 sm:flex-row">
                    <div className="min-w-0 sm:flex-1">
                        <Selector
                            entries={entries}
                            value={userOne}
                            onValueChange={(entry) =>
                                setSelected((s) => ({ ...s, userOne: entry?.user.uid ?? null }))
                            }
                        />
                    </div>
                    <div className="min-w-0 sm:flex-1">
                        <Selector
                            entries={entries}
                            value={userTwo}
                            onValueChange={(entry) =>
                                setSelected((s) => ({ ...s, userTwo: entry?.user.uid ?? null }))
                            }
                        />
                    </div>
                </div>

                {comparison ? (
                    <ComparisonResults comparison={comparison} />
                ) : (
                    <Empty className="border border-solid">
                        <EmptyHeader>
                            <EmptyMedia variant="icon"><UsersIcon /></EmptyMedia>
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

    // return (
    //     <div className="h-full min-h-0 flex gap-4 overflow-y-auto sm:flex-col lg:overflow-visible">
    //         <div className="flex h-96 min-h-0 shrink-0 flex-col overflow-auto lg:order-1 lg:h-auto lg:shrink">
    //             {isFetching ? (
    //                 <Empty className="h-full min-h-64 border border-solid">
    //                     <EmptyHeader>
    //                         <EmptyMedia variant="icon">
    //                             <Spinner />
    //                         </EmptyMedia>
    //                         <EmptyTitle>Loading leaderboard</EmptyTitle>
    //                         <EmptyDescription className="text-sm">
    //                             Please wait while the leaderboard loads
    //                         </EmptyDescription>
    //                     </EmptyHeader>
    //                 </Empty>
    //             ) : (
    //                 <DataTable
    //                     columns={columns}
    //                     data={entries}
    //                 />
    //             )}
    //         </div>

    //         <div className="flex min-w-0 shrink-0 flex-col gap-4 lg:order-2 lg:flex-1 lg:shrink">
    //             {isError && (
    //                 <p className="text-sm text-destructive">
    //                     {error instanceof Error ? error.message : "Couldn't load leaderboard"}
    //                     {entries.length > 0 && ". Showing latest leaderboard"}
    //                 </p>
    //             )}

    //             <Tabs value={kind} onValueChange={handleKind}>
    //                 <TabsList className="w-full">
    //                     <TabsTrigger value="monthly">Monthly</TabsTrigger>
    //                     <TabsTrigger value="global">Global</TabsTrigger>
    //                 </TabsList>
    //             </Tabs>

    //             {data && (
    //                 <p className="-mt-2 text-xs text-muted-foreground">
    //                     Scraped {formatAge(data.scrape.scraped_at)}
    //                 </p>
    //             )}

    //             <div className="flex flex-col gap-2 sm:flex-row">
    //                 <div className="min-w-0 sm:flex-1">
    //                     <Selector
    //                         entries={entries}
    //                         value={userOne}
    //                         onValueChange={(entry) => 
    //                             setSelected((s) => ({ ...s, userOne: entry?.user.uid ?? null }))
    //                         }
    //                         />
    //                 </div>
    //                 <div className="min-w-0 sm:flex-1">
    //                     <Selector
    //                         entries={entries}
    //                         value={userTwo}
    //                         onValueChange={(entry) => 
    //                             setSelected((s) => ({ ...s, userTwo: entry?.user.uid ?? null }))
    //                         }
    //                     />
    //                 </div>
    //             </div>

    //             {comparison ? (
    //                 <ComparisonResults comparison={comparison} />
    //             ) : (
    //                 <Empty className="border border-solid">
    //                     <EmptyHeader>
    //                         <EmptyMedia variant="icon">
    //                             <UsersIcon />
    //                         </EmptyMedia>
    //                         <EmptyTitle>No users selected</EmptyTitle>
    //                         <EmptyDescription className="text-sm">
    //                             Select two users to compare individual metrics and estimate
    //                             the gap between them
    //                         </EmptyDescription>
    //                     </EmptyHeader>
    //                 </Empty>
    //             )}
    //         </div>
    //     </div>
    // )
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

function Handle({ name }: { name: string }) {
    return <strong className="font-bold whitespace-nowrap">@{name}</strong>
}

type StatProps = {
    icon:      LucideIcon
    label:     string
    values:    number[]
    formatter: (v: number) => string
    description:  ReactNode
}

function Stat({ icon: Icon, label, values, formatter, description }: StatProps) {
    const value = values.map(formatter).join("-")

    return (
        <Popover>
            <PopoverTrigger nativeButton={false} delay={10} closeDelay={100} render={
                <div className="flex min-w-0 flex-col items-center text-center">
                    <div className="flex flex-col items-center gap-1 sm:flex-row sm:gap-2">
                        <Icon className="h-4 w-4 shrink-0 sm:h-5 sm:w-5" color="var(--chart-1)" />
                        <p className="cursor-pointer text-lg hover:underline sm:text-2xl">{value}</p>
                    </div>
                    <p className="text-xs sm:text-sm">{label}</p>
                </div>
            }/>
            <PopoverContent className="flex w-64 max-w-[calc(100vw-2rem)] flex-col gap-0.5">
                <div>{description}</div>
            </PopoverContent>
        </Popover>
    )
}

function ComparisonResults({ comparison }: { comparison: ComparisonResult }) {
    const { higher, lower } = comparison 
    const higherLabel = `@${higher.user.username}`
    const lowerLabel = `@${lower.user.username}`
    const higherHandle = <Handle name={higher.user.username} />
    const lowerHandle = <Handle name={lower.user.username} />
    
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

    const statData: StatProps[] = [
        {
            icon: ArrowDownIcon,
            label: "points behind",
            formatter: formatLarge,
            values: [comparison.gap],
            description: <>{lowerHandle} is this many points behind {higherHandle}</>
        },
        {
            icon: BracketsIcon,
            label: "puzzles to close gap",
            values: [comparison.minPuzzles, comparison.expectedPuzzles],
            formatter: formatLarge,
            description: <>An estimate of the puzzles {lowerHandle} must complete to reach {higherHandle}</>
        },
        {
            icon: ClockArrowRightIcon,
            label: "to close gap",
            values: [comparison.expectedTime],
            formatter: formatDuration,
            description: <>An estimate of the time gap between {lowerHandle} and {higherHandle}</>
        },
    ]


    return (
        <div className="flex flex-col gap-4 border p-3 sm:p-4 2xl:flex-row">
            <Table className="min-w-0 overflow-x-auto 2xl:flex-1">
                <TableHeader>
                    <TableRow>
                        <TableHead className="text-base font-bold sm:text-lg">Metric</TableHead>
                        <TableHead className="max-w-28 truncate text-base font-bold sm:text-lg">{higherLabel}</TableHead>
                        <TableHead className="max-w-28 truncate text-base font-bold sm:text-lg">{lowerLabel}</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {tableRowData.map((row) => (
                        <TableRow key={row.category}>
                            <TableCell className="text-sm whitespace-normal">{row.category}</TableCell>
                            <TableCell className="text-sm">{row.formatter(row.higherUserMetric)}</TableCell>
                            <TableCell className="text-sm">{row.formatter(row.lowerUserMetric)}</TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>

            <div className="grid grid-cols-3 items-start gap-2 2xl:flex 2xl:w-xs 2xl:flex-col 2xl:items-center 2xl:justify-center 2xl:gap-3">
                {statData.map((stat) => (
                    <Stat key={stat.label} {...stat} />
                ))}
            </div>
        </div>
    )
}
