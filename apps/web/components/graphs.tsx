/**
 * @file graphs.tsx
 */

"use client"
import { ISODate, Kind, LeaderboardEntry, ScrapeEntry } from "@/lib/models"
import React, { useMemo } from "react"
import { ChartConfig, ChartContainer, ChartTooltip, ChartTooltipContent } from "./ui/chart"
import { Area, AreaChart, Bar, BarChart, CartesianGrid, XAxis, YAxis } from "recharts"

export type GraphProps = {
    entries: LeaderboardEntry[]
    n:       number
    kind:    Kind
    period?: ISODate
}

export type LeaderboardGraph = {
    id:          string
    title:       string
    description: string
    usesN:       boolean
    render:      (props: GraphProps) => React.ReactNode
}

type Metric = keyof Pick<
    ScrapeEntry,
    "total_points" | "solved" | "attempted" |
    "acceptance_rate" | "avg_time_s" | "avg_points"
>

const compact = new Intl.NumberFormat("en", {
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
    render: function (props: GraphProps): React.ReactNode {
        return <NBar {...props} metric={metric} label={title} reverse={reverse}/>
    }
})

const hGraph = (
    id: string, title: string,
    description: string, metric: Metric,
    bins: number = 90
): LeaderboardGraph => ({
    id,
    title,
    description,
    usesN: false,
    render: function (props: GraphProps): React.ReactNode {
        return <Histogram {...props} metric={metric} bins={bins} />
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
                "User points distribution",
                "A distribution of individual user points",
                "total_points"
            ),
            nGraph(
                "n-avg-time",
                "Average time distribution",
                "A distribution of individual user time averages (s)",
                "avg_time_s",
                true
            ),
            nGraph(
                "n-acceptance-rate",
                "Accentance rate distribution",
                "A distribution of individual acceptance rates",
                "acceptance_rate"
            ),
            nGraph(
                "n-solved",
                "Total solved distribution",
                "A distribution of individual solves",
                "solved"
            ),
            nGraph(
                "n-avg-points",
                "Average points distribution",
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
                "User points histogram",
                "A histogram of collective user points",
                "total_points"
            )
        ]
    },
]

