import type { Metadata } from "next"
import { connection } from "next/server"

import { ChartCard, summarize } from "@/components/charts/chart-card"
import { MetricChart, type ChartRow } from "@/components/charts/metric-chart"
import { DateRangePicker } from "@/components/date-range-picker"
import { KindSwitch } from "@/components/kind-switch"
import { MigrationNotice } from "@/components/migration-notice"
import { parseRange, pickBucket, rangeBounds } from "@/lib/dates"
import { formatCompact, formatLarge, percent } from "@/lib/format"
import { KIND_LABELS, parseKind } from "@/lib/leaderboards"
import { getLeaderboardMetrics } from "@/lib/queries"
import { first, kindHref } from "@/lib/search-params"

export const metadata: Metadata = { title: "Analytics" }

const BUCKET_LABEL = { hour: "hourly", day: "daily", week: "weekly" } as const

export default async function AnalyticsPage({ searchParams }: PageProps<"/analytics">) {
  const params = await searchParams
  await connection() // default date range depends on "now"
  const kind = parseKind(first(params.kind))
  const range = parseRange({ from: first(params.from), to: first(params.to) }, kind)
  const bucket = pickBucket(range)
  const { start, end } = rangeBounds(range)

  const { rows, missingRpc } = await getLeaderboardMetrics(kind, start, end, bucket)

  const data: ChartRow[] = rows.map((r) => ({
    bucket: r.bucket,
    top1: r.top1_points,
    top10: r.top10_points,
    top50: r.top50_points,
    top100: r.top100_points,
    median: r.median_points,
    entrants: r.entrants,
    active: r.active_24h,
    solved: r.total_solved,
    acceptance: r.avg_acceptance,
    avgTime: r.avg_time_s,
  }))
  const col = (key: keyof ChartRow) => data.map((d) => d[key] as number | null)

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div className="flex flex-col gap-1">
          <h1 className="font-heading text-2xl font-semibold tracking-tight">Analytics</h1>
          <p className="text-sm text-muted-foreground">
            {KIND_LABELS[kind]} leaderboard, last snapshot per {bucket} ({BUCKET_LABEL[bucket]} points, US Eastern).
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <KindSwitch active={kind} hrefFor={(k) => kindHref("/analytics", k, params)} />
          <DateRangePicker range={range} />
        </div>
      </div>

      {kind === "monthly" && (
        <p className="text-xs text-muted-foreground">
          Monthly standings reset on the 1st of each month (US Eastern), so series drop at period boundaries.
        </p>
      )}

      {missingRpc ? (
        <MigrationNotice />
      ) : (
        <div className="grid gap-6 lg:grid-cols-2">
          <ChartCard
            title="Point cutoffs"
            description="Points held by #1, #10, #50 and #100"
            summary={summarize(col("top1"), formatCompact)}
            className="lg:col-span-2"
          >
            <MetricChart
              data={data}
              bucket={bucket}
              format="number"
              className="aspect-auto h-80 w-full"
              series={[
                { key: "top1", label: "#1" },
                { key: "top10", label: "#10" },
                { key: "top50", label: "#50" },
                { key: "top100", label: "#100" },
              ]}
            />
          </ChartCard>

          <ChartCard
            title="Median points"
            description="Midpoint of everyone on the board"
            summary={summarize(col("median"), formatLarge)}
          >
            <MetricChart data={data} bucket={bucket} kind="area" series={[{ key: "median", label: "Median" }]} />
          </ChartCard>

          <ChartCard
            title="Activity"
            description="Users active in the 24h before each snapshot"
            summary={summarize(col("active"), formatLarge)}
          >
            <MetricChart
              data={data}
              bucket={bucket}
              series={[
                { key: "active", label: "Active (24h)" },
                { key: "entrants", label: "On board" },
              ]}
            />
          </ChartCard>

          <ChartCard
            title="Puzzles solved"
            description="Total solves across the board"
            summary={summarize(col("solved"), formatLarge)}
          >
            <MetricChart
              data={data}
              bucket={bucket}
              kind="area"
              series={[{ key: "solved", label: "Solved", color: "var(--chart-2)" }]}
            />
          </ChartCard>

          <ChartCard
            title="Efficiency"
            description="Average acceptance rate and solve time"
            summary={summarize(col("acceptance"), (n) => percent(Math.round(n * 100) / 100))}
          >
            <MetricChart
              data={data}
              bucket={bucket}
              format="percent"
              rightFormat="seconds"
              series={[
                { key: "acceptance", label: "Acceptance" },
                { key: "avgTime", label: "Avg time", right: true },
              ]}
            />
          </ChartCard>
        </div>
      )}
    </div>
  )
}
