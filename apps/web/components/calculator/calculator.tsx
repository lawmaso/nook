
"use client"

import { useEffect, useMemo, useState } from "react"

import { columns } from "./columns"
import { DataTable } from "./data-table"
import { Selector } from "./selector"

import { getLatestScrapeEntries } from "@/lib/queries"
import { LeaderboardEntry } from "@/lib/models"

import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "../ui/card"

import {
    formatDecimal,
    formatDuration,
    formatLarge,
    formatSeconds,
} from "./format"

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
    const [entries, setEntries] = useState<LeaderboardEntry[] | null>(null)
    const [first, setFirst] = useState<Selection>(null)
    const [second, setSecond] = useState<Selection>(null)
    const [loadError, setLoadError] = useState<string | null>(null)

    useEffect(() => {
        let cancelled = false

        getLatestScrapeEntries("monthly")
            .then((result) => {
                if (cancelled) return

                const loadedEntries = result ?? []

                setEntries(loadedEntries)
                setFirst(loadedEntries[0] ?? null)
                setSecond(loadedEntries[1] ?? null)
            })
            .catch((error: unknown) => {
                console.error("Failed to load leaderboard entries:", error)

                if (!cancelled) {
                    setEntries([])
                    setLoadError("Failed to load leaderboard entries.")
                }
            })

        return () => {
            cancelled = true
        }
    }, [])

    // Always compare the lower-scoring user against the higher-scoring user,
    // regardless of which selector they were chosen in.
    const comparison = useMemo(() => {
        if (!first || !second) return null

        const [lo, hi] =
            first.total_points <= second.total_points
                ? [first, second]
                : [second, first]

        const pointsGap = Math.max(
            0,
            hi.total_points - lo.total_points
        )

        // Minimum possible puzzle count, assuming 563 points per puzzle.
        const minimumPuzzles = pointsGap / 563.0

        // Expected puzzle count, based on the lower-scoring user's average.
        const expectedPuzzles =
            Number.isFinite(lo.avg_points) && lo.avg_points > 0
                ? pointsGap / lo.avg_points
                : null

        // Treat avg_time_s as seconds per puzzle.
        const avgTimeS =
            Number.isFinite(lo.avg_time_s) && lo.avg_time_s >= 0
                ? lo.avg_time_s
                : null

        // Estimate the time required to complete each puzzle estimate.
        const minimumTimeS =
            avgTimeS !== null
                ? minimumPuzzles * avgTimeS
                : null

        const expectedTimeS =
            avgTimeS !== null && expectedPuzzles !== null
                ? expectedPuzzles * avgTimeS
                : null

        return {
            lo,
            hi,
            pointsGap,
            minimumPuzzles,
            expectedPuzzles,
            avgTimeS,
            minimumTimeS,
            expectedTimeS,
        }
    }, [first, second])

    if (entries === null) {
        return (
            <div className="p-4 text-sm text-muted-foreground">
                Loading leaderboard...
            </div>
        )
    }

    return (
        <div className="h-full min-h-0 flex gap-4">
            <DataTable
                columns={columns}
                data={entries}
            />

            <div className="flex min-w-0 flex-1 flex-col gap-4">
                {loadError && (
                    <p className="text-sm text-destructive">
                        {loadError}
                    </p>
                )}

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

                {comparison ? (
                    <>
                        <ComparisonResults comparison={comparison} />
                        <CatchUpEstimate comparison={comparison} />
                    </>
                ) : (
                    <Card>
                        <CardContent className="pt-6 text-sm text-muted-foreground">
                            Select two users to compare their leaderboard
                            statistics.
                        </CardContent>
                    </Card>
                )}
            </div>
        </div>
    )
}

// Displays the difference between the two selected users.
function ComparisonResults({
    comparison,
}: {
    comparison: {
        lo: LeaderboardEntry
        hi: LeaderboardEntry
        pointsGap: number
        minimumPuzzles: number
        expectedPuzzles: number | null
        avgTimeS: number | null
        minimumTimeS: number | null
        expectedTimeS: number | null
    }
}) {
    const { lo, hi } = comparison

    const loLabel = `${lo.rank}. ${lo.user.username}`
    const hiLabel = `${hi.rank}. ${hi.user.username}`

    return (
        <Card>
            <CardHeader>
                <CardTitle>User comparison</CardTitle>
                <CardDescription>
                    Differences are calculated as Hi − Lo.
                </CardDescription>
            </CardHeader>

            <CardContent>
                <div className="overflow-x-auto">
                    <table className="w-full min-w-[28rem] text-sm">
                        <thead>
                            <tr className="text-muted-foreground">
                                <th className="px-2 py-2 text-left font-medium">
                                    Metric
                                </th>
                                <th className="px-2 py-2 text-right font-medium">
                                    {loLabel}
                                </th>
                                <th className="px-2 py-2 text-right font-medium">
                                    {hiLabel}
                                </th>
                                <th className="px-2 py-2 text-right font-medium">
                                    Difference
                                </th>
                            </tr>
                        </thead>

                        <tbody>
                            <tr className="border-t">
                                <th className="px-2 py-3 text-left font-medium">
                                    Average points
                                </th>
                                <td className="px-2 py-3 text-right tabular-nums">
                                    {formatDecimal(lo.avg_points)}
                                </td>
                                <td className="px-2 py-3 text-right tabular-nums">
                                    {formatDecimal(hi.avg_points)}
                                </td>
                                <td className="px-2 py-3 text-right tabular-nums font-medium">
                                    {formatSigned(
                                        hi.avg_points - lo.avg_points,
                                        formatDecimal
                                    )}
                                </td>
                            </tr>

                            <tr className="border-t">
                                <th className="px-2 py-3 text-left font-medium">
                                    Average time
                                </th>
                                <td className="px-2 py-3 text-right tabular-nums">
                                    {formatSeconds(lo.avg_time_s)}
                                </td>
                                <td className="px-2 py-3 text-right tabular-nums">
                                    {formatSeconds(hi.avg_time_s)}
                                </td>
                                <td className="px-2 py-3 text-right tabular-nums font-medium">
                                    {formatSigned(
                                        hi.avg_time_s - lo.avg_time_s,
                                        formatSeconds
                                    )}
                                </td>
                            </tr>

                            <tr className="border-t">
                                <th className="px-2 py-3 text-left font-medium">
                                    Total points
                                </th>
                                <td className="px-2 py-3 text-right tabular-nums">
                                    {formatLarge(lo.total_points)}
                                </td>
                                <td className="px-2 py-3 text-right tabular-nums">
                                    {formatLarge(hi.total_points)}
                                </td>
                                <td className="px-2 py-3 text-right tabular-nums font-medium">
                                    {formatSigned(
                                        hi.total_points - lo.total_points,
                                        formatLarge
                                    )}
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </CardContent>
        </Card>
    )
}

// Estimates how many puzzles and how much time the lower-scoring user
// needs to close the total-points gap.
function CatchUpEstimate({
    comparison,
}: {
    comparison: {
        lo: LeaderboardEntry
        hi: LeaderboardEntry
        pointsGap: number
        minimumPuzzles: number
        expectedPuzzles: number | null
        avgTimeS: number | null
        minimumTimeS: number | null
        expectedTimeS: number | null
    }
}) {
    const {
        lo,
        hi,
        pointsGap,
        minimumPuzzles,
        expectedPuzzles,
        avgTimeS,
        minimumTimeS,
        expectedTimeS,
    } = comparison

    return (
        <Card>
            <CardHeader>
                <CardTitle>Catch-up estimate</CardTitle>

                <CardDescription>
                    {hi.user.username} is ahead of {lo.user.username} by{" "}
                    {formatLarge(pointsGap)} points.

                    {avgTimeS !== null
                        ? ` Time estimates use ${formatSeconds(avgTimeS)} per puzzle from ${lo.user.username}.`
                        : " A valid average time is required to estimate catch-up time."}
                </CardDescription>
            </CardHeader>

            <CardContent>
                <div className="grid grid-cols-2 gap-4">
                    <div className="border p-4">
                        <p className="text-sm text-muted-foreground">
                            Minimum estimate
                        </p>

                        <p className="mt-2 text-2xl font-semibold tabular-nums">
                            {formatDecimal(minimumPuzzles)}
                        </p>

                        <p className="text-sm text-muted-foreground">
                            puzzles
                        </p>

                        <div className="my-3 border-t" />

                        <p className="text-sm text-muted-foreground">
                            Estimated time
                        </p>

                        <p className="mt-1 text-lg font-semibold tabular-nums">
                            {minimumTimeS !== null
                                ? formatDuration(minimumTimeS)
                                : "—"}
                        </p>

                        <p className="mt-2 text-xs text-muted-foreground">
                            Points gap ÷ 563.0
                        </p>
                    </div>

                    <div className="border p-4">
                        <p className="text-sm text-muted-foreground">
                            Expected estimate
                        </p>

                        <p className="mt-2 text-2xl font-semibold tabular-nums">
                            {expectedPuzzles !== null
                                ? formatDecimal(expectedPuzzles)
                                : "—"}
                        </p>

                        <p className="text-sm text-muted-foreground">
                            puzzles
                        </p>

                        <div className="my-3 border-t" />

                        <p className="text-sm text-muted-foreground">
                            Estimated time
                        </p>

                        <p className="mt-1 text-lg font-semibold tabular-nums">
                            {expectedTimeS !== null
                                ? formatDuration(expectedTimeS)
                                : "—"}
                        </p>

                        <p className="mt-2 text-xs text-muted-foreground">
                            Points gap ÷ {lo.user.username}'s average points
                        </p>
                    </div>
                </div>

                <p className="mt-4 text-xs text-muted-foreground">
                    These are estimates based on the selected lower-scoring
                    user's current averages, not guaranteed completion times.
                </p>
            </CardContent>
        </Card>
    )
}
