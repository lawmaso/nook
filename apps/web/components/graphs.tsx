/**
 * @file graphs.tsx
 */

"use client"

import { ISODate, Kind, LeaderboardEntry, ScrapeEntry, TrajectoryPoint } from "@/lib/models"
import React, { useEffect, useMemo, useState } from "react"
import { ChartConfig, ChartContainer, ChartTooltip, ChartTooltipContent } from "./ui/chart"
import { Area, AreaChart, Bar, BarChart, CartesianGrid, XAxis, YAxis } from "recharts"
import { loadTrajectories } from "@/lib/actions"
import { userColor } from "@/lib/colors"
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from "./ui/empty"
import { CircleAlertIcon, FolderOpenIcon } from "lucide-react"
import { Spinner } from "./ui/spinner"

export type GraphProps = {
    entries: LeaderboardEntry[]
    n:       number
    bins:    number
    kind:    Kind
    period?: ISODate
}

export type LeaderboardGraph = {
    id:          string
    title:       string
    description: string
    usesN:       boolean
    usesBins:    boolean
    render:      (props: GraphProps) => React.ReactNode
}

type Metric = keyof Pick<
    ScrapeEntry,
    "total_points" | "solved" | "attempted" |
    "acceptance_rate" | "avg_time_s" | "avg_points"
>

const compact = new Intl.NumberFormat("en-US", {
    notation: "compact",
    maximumFractionDigits: 1
})

function ResponsiveContainer({ config, children }: {
    config: ChartConfig
    children: React.ReactNode
}) {
    return (
        <ChartContainer
            config={config}
            className="h-full min-h-0 w-full aspect-auto"
        >
            {children}
        </ChartContainer>
    )
}

function NBar({
    entries, n, metric, label, reverse
}: GraphProps & { metric: Metric; label: string, reverse: boolean }) {
    const data = useMemo(
        () => {
            let sorted = [...entries].sort((a, b) =>
                reverse ? b[metric] - a[metric] : a[metric] - b[metric]
            )

            return sorted
                .slice(entries.length - n, entries.length)
                .map((e) => ({
                    name:  e.user.username,
                    value: e[metric]
                }))
        },
        [entries, n, metric]
    )

    const config = {
        value: {
            label: label,
            color: "var(--chart-1)"
        }
    } satisfies ChartConfig

    return (
        <ResponsiveContainer config={config}>
            <BarChart data={data}>
                <CartesianGrid />
                <XAxis
                    dataKey="name"
                    type="category"
                    height={80}
                    angle={-30}
                    textAnchor="end"
                    interval={n <= 25
                        ? 0
                        : 1
                    }
                />
                <YAxis
                    type="number"
                    tickFormatter={(v) => compact.format(v)}
                />
                <ChartTooltip content={<ChartTooltipContent />} />
                <Bar dataKey="value" fill="var(--chart-3)" />
            </BarChart>
        </ResponsiveContainer>
    )
}

function buildHistogram(values: number[], bins: number) {
    if (values.length == 0) return []

    let min = Infinity
    let max = -Infinity

    for (const v of values) {
        if (v < min) min = v
        if (v > max) max = v
    }

    const binWidth = ((max - min) / bins) || 1

    const hist = Array.from({ length: bins }, (_v, i) => {
        const left = min + (i * binWidth)
        const right = left + binWidth

        return {
            start: compact.format(left),
            range: `${compact.format(left)} - ${compact.format(right)}`,
            count: 0
        }
    })

    for (const v of values) {
        const idx = Math.min(hist.length - 1, Math.floor((v - min) / binWidth))
        hist[idx].count++
    }

    return hist
}

function Histogram({
    entries, metric, bins
}: GraphProps & { metric: Metric }) {
    const data = useMemo(
        () => buildHistogram(entries.map((e) => e[metric]), bins),
        [entries, metric, bins]
    )

    const config = {
        count: {
            label: "Users",
            color: "var(--chart-3)"
        }
    } satisfies ChartConfig

    return (
        <ResponsiveContainer config={config}>
            <BarChart data={data}>
                <CartesianGrid />
                <XAxis
                    dataKey="start"
                    minTickGap={20}
                />
                <YAxis />
                <ChartTooltip
                    content={
                        <ChartTooltipContent
                            labelFormatter={(_, payload) => payload?.[0]?.payload?.range}
                        />
                    }
                />
                <Bar dataKey="count" fill="var(--chart-3)"/>
            </BarChart>
        </ResponsiveContainer>
    )
}

function Distribution({
    entries, metric, bins
}: GraphProps & { metric: Metric; bins: number }) {
    const data = useMemo(
        () => buildHistogram(entries.map((e) => e[metric]), bins),
        [entries, metric, bins]
    )

    const config = {
        count: {
            label: "Users",
            color: "var(--chart-3)"
        }
    } satisfies ChartConfig

    return (
        <ChartContainer config={config}>
            <AreaChart data={data}>
                <CartesianGrid />
                <XAxis dataKey="start"/>
                <YAxis />
                <ChartTooltip content={<ChartTooltipContent />} />
                <Area
                    dataKey="count"
                    type="natural"
                    fill="none"
                />
            </AreaChart>
        </ChartContainer>
    )
}

const nGraph = (
    id: string, title: string,
    description: string, metric: Metric,
    reverse: boolean = false
): LeaderboardGraph => ({
    id,
    title,
    description,
    usesN: true,
    usesBins: false,
    render: function (props: GraphProps): React.ReactNode {
        return <NBar {...props} metric={metric} label={title} reverse={reverse}/>
    }
})

function Trajectory({
    entries, n, kind, period, metric
}: GraphProps & { metric: Metric }) {
    const shown = useMemo(
        () => [...entries]
            .sort((a, b) => b[metric] - a[metric])
            .slice(),
        [entries, n, metric]
    )

    const uidKey = shown.map((e) => e.uid).join(",")

    const [points, setPoints]   = useState<TrajectoryPoint[]>([])
    const [loading, setLoading] = useState(true)
    const [error, setError]     = useState<string | null>(null)

    useEffect(() => {
        let cancelled = false

        setLoading(true)
        setError(null)
        
        loadTrajectories(kind, period, shown.map((e) => e.uid), metric)
            .then((r) => !cancelled && setPoints(r))
            .catch((e) => !cancelled && setError(e instanceof Error ? e.message : "Trajectory load failed"))
            .finally(() => !cancelled && setLoading(false))
    }, [kind, period, uidKey, metric])

    const data = useMemo(() => {
        const byTime = new Map<number, Record<string, number>>()

        for (const p of points) {
            const row = byTime.get(p.t) ?? { t: p.t }

            row[`u${p.uid}`] = p.value
            byTime.set(p.t, row)
        }

        return [...byTime.values()].sort((a, b) => a.t - b.t)
    }, [points])


    const config = useMemo(
        () => Object.fromEntries(
            shown.map((e, _i) => [
                `u${e.uid}`,
                {
                    label: e.user.username,
                    color: "var(--chart-1)"
                }
            ])
        ) as ChartConfig,
        [shown]
    )

    const Placeholder = ({ icon, title, description, content }: {
        icon: React.ReactNode,
        title: string
        description: string
        content?: string
    }) => (
        <Empty className="border min-h-full">
            <EmptyHeader>
                <EmptyMedia variant="icon">{icon}</EmptyMedia>
                <EmptyTitle className="text-xl">{title}</EmptyTitle>
                <EmptyDescription>{description}</EmptyDescription>
            </EmptyHeader>
            {content && (
                <EmptyContent>
                    {content}
                </EmptyContent>
            )}
        </Empty>
        
    )

    if (error) {
        return Placeholder({
            icon: <CircleAlertIcon />,
            title: "Error",
            description: error
        })
    } else if (loading) {
        return Placeholder({
            icon: <Spinner />,
            title: "Loading trajectory",
            description: "Using scrape entries to create individual trajectories"
        })
    } else if (data.length <= 1) {
        return Placeholder({
            icon: <FolderOpenIcon />,
            title: "Not enough scrapes",
            description: "We don't have enough scrape entries to form valid trajectories"
        })
    }

    const day = (t: number) =>
        new Date(t).toLocaleDateString("en-US", {
            month: "short",
            day: "numeric"
        })

    return (
        <ResponsiveContainer config={config}>
            <AreaChart data={data}>
                <CartesianGrid />
                <XAxis
                    dataKey="t"
                    type="number"
                    scale="time"
                    domain={["dataMin", "dataMax"]}
                    tickFormatter={day}
                />
                <YAxis
                    tickFormatter={(v) => compact.format(v)}
                    domain={["auto", "auto"]}
                />
                <ChartTooltip
                    content={
                        <ChartTooltipContent
                            labelFormatter={(_, payload) =>
                                new Date(payload?.[0]?.payload?.t).toLocaleString("en-US", {
                                    timeZoneName: "shortGeneric"
                                }).replace(",", " @")
                            }
                        />
                    }
                />
                {shown.map((e, idx) => (
                    <Area
                        key={e.uid}
                        dataKey={`u${e.uid}`}
                        type="stepAfter"
                        stroke={userColor(idx)}
                        fillOpacity={0}
                        strokeWidth={1.44}
                        isAnimationActive={false}
                        connectNulls
                    />
                ))}
            </AreaChart>
        </ResponsiveContainer>
    )
}

const hGraph = (
    id: string, title: string,
    description: string, metric: Metric
): LeaderboardGraph => ({
    id,
    title,
    description,
    usesN: false,
    usesBins: true,
    render: function (props: GraphProps): React.ReactNode {
        return <Histogram {...props} metric={metric} />
    }
})

const tGraph = (
    id: string, title: string,
    description: string, metric: Metric
): LeaderboardGraph => ({
    id,
    title,
    description,
    usesN: true,
    usesBins: false,
    render: function (props: GraphProps): React.ReactNode {
        return <Trajectory {...props} metric={metric} />
    }
})

type GraphGroup = {
    title: string
    graphs: LeaderboardGraph[]
}

export const GRAPHS: GraphGroup[] = [
    {
        title: "Distributions",
        graphs: [
            nGraph(
                "n-user-points",
                "User points",
                "A distribution of individual user points",
                "total_points"
            ),
            nGraph(
                "n-avg-time",
                "Average time",
                "A distribution of individual user time averages (s)",
                "avg_time_s",
                true
            ),
            nGraph(
                "n-acceptance-rate",
                "Acceptance rate",
                "A distribution of individual acceptance rates",
                "acceptance_rate"
            ),
            nGraph(
                "n-solved",
                "Total solved",
                "A distribution of individual solves",
                "solved"
            ),
            nGraph(
                "n-avg-points",
                "Average points",
                "A distribution of individual average points per puzzle",
                "avg_points"
            ),
        ]
    },
    {
        title: "Histograms",
        graphs: [
            hGraph(
                "h-user-points",
                "User points",
                "A histogram of collective user points",
                "total_points"
            ),
            hGraph(
                "h-avg-time",
                "Average time",
                "A histogram of collective user time averages (s)",
                "avg_time_s",
            ),
            hGraph(
                "h-acceptance-rate",
                "Acceptance rate",
                "A histogram of collective acceptance rates",
                "acceptance_rate"
            ),
            hGraph(
                "h-solved",
                "Total solved",
                "A histogram of collective solves",
                "solved"
            ),
            hGraph(
                "h-avg-points",
                "Average points",
                "A histogram of collective average points per puzzle",
                "avg_points"
            ),
        ]
    },
    {
        title: "Trajectories",
        graphs: [
            tGraph(
                "t-total-points",
                "User points",
                "A trajectory of individual user points",
                "total_points"
            ),
            tGraph(
                "t-avg-time",
                "Average time",
                "A trajectory of individual user time averages (s)",
                "avg_time_s"
            ),
            tGraph(
                "t-acceptance-rate",
                "Acceptance rate",
                "A trajectory of individual acceptance rates",
                "acceptance_rate"
            ),
            tGraph(
                "t-solved",
                "Total solved",
                "A trajectory of individual solves",
                "solved"
            ),
            tGraph(
                "t-avg-points",
                "Average points",
                "A trajectory of individual average points per puzzle",
                "avg_points"
            ),
        ]
    }
]
