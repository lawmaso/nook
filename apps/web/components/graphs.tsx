/**
 * @file graphs.tsx
 */

"use client"
import { ISODate, Kind, LeaderboardEntry, ScrapeEntry } from "@/lib/models"
import React, { useMemo } from "react"
import { ChartConfig, ChartContainer, ChartTooltip, ChartTooltipContent } from "./ui/chart"
import { Bar, BarChart, CartesianGrid, XAxis, YAxis } from "recharts"

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
    maximumFractionDigits: 2
})

function NBar({
    entries, n, metric, label
}: GraphProps & { metric: Metric; label: string}) {
    const data = useMemo(
        () => [...entries]
            .sort((a, b) => a[metric] - b[metric])
            .map((e) => ({
                name:  e.user.username,
                value: e[metric]
            })),
        [entries, n, metric]
    )

    const config = {
        value: {
            label: label,
            color: "var(--chart-1)"
        }
    } satisfies ChartConfig

    return (
        <ChartContainer config={config}>
            <BarChart data={data}>
                <CartesianGrid />
                <XAxis
                    dataKey="name"
                    type="category"
                    angle={-40}
                    textAnchor="end"
                    height={100}
                    interval={2}
                />
                <YAxis
                    type="number"
                    tickFormatter={(v) => compact.format(v)}
                />
                <ChartTooltip content={<ChartTooltipContent />} />
                <Bar dataKey="value" fill="var(--chart-1)" />
            </BarChart>
        </ChartContainer>
    )
}

const nGraph = (
    id: string, title: string,
    description: string, metric: Metric
): LeaderboardGraph => ({
    id,
    title,
    description,
    usesN: true,
    render: function (props: GraphProps): React.ReactNode {
        return <NBar {...props} metric={metric} label={title}/>
    }
})

export const GRAPHS = [
    nGraph("n-user-points", "User points", "User points", "total_points",)
]
