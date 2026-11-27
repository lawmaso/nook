"use client"

import { Area, AreaChart, CartesianGrid, Line, LineChart, XAxis, YAxis } from "recharts"

import {
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from "@/components/ui/chart"
import { formatBucket, type Bucket } from "@/lib/dates"
import { formatCompact, formatLarge, percent, secondsLabel, timeGap } from "@/lib/format"

/** Serializable format names (functions can't cross the server/client boundary). */
export type ValueFormat = "number" | "compact" | "percent" | "seconds" | "duration" | "rank"

const FORMATTERS: Record<ValueFormat, { axis: (n: number) => string; full: (n: number) => string }> = {
  number: { axis: formatCompact, full: formatLarge },
  compact: { axis: formatCompact, full: formatCompact },
  percent: { axis: (n) => `${Math.round(n)}%`, full: (n) => percent(Math.round(n * 100) / 100) },
  seconds: { axis: (n) => `${Math.round(n)}s`, full: (n) => secondsLabel(Math.round(n * 10) / 10) },
  duration: { axis: (n) => timeGap(n).split(" ")[0], full: timeGap },
  rank: { axis: (n) => `#${n}`, full: (n) => `#${formatLarge(n)}` },
}

export type Series = {
  key: string
  label: string
  /** CSS color; defaults to --chart-N by position. */
  color?: string
  /** Put this series on a secondary right-hand axis. */
  right?: boolean
}

export type ChartRow = { bucket: string } & Record<string, number | string | null>

export function MetricChart({
  data,
  series,
  bucket,
  kind = "line",
  format = "number",
  rightFormat = format,
  reversed = false,
  legend = series.length > 1,
  className = "aspect-auto h-64 w-full",
}: {
  data: ChartRow[]
  series: Series[]
  bucket: Bucket
  kind?: "line" | "area"
  format?: ValueFormat
  rightFormat?: ValueFormat
  /** Invert the Y axis (for ranks, where lower is better). */
  reversed?: boolean
  legend?: boolean
  className?: string
}) {
  if (data.length === 0) {
    return (
      <div className="flex h-64 items-center justify-center rounded-lg border border-dashed text-sm text-muted-foreground">
        No scrapes in this range.
      </div>
    )
  }

  const config = Object.fromEntries(
    series.map((s, i) => [s.key, { label: s.label, color: s.color ?? `var(--chart-${(i % 5) + 1})` }]),
  ) satisfies ChartConfig
  const formatFor = (key: string) => FORMATTERS[series.find((s) => s.key === key)?.right ? rightFormat : format]
  const hasRight = series.some((s) => s.right)
  const Chart = kind === "area" ? AreaChart : LineChart

  return (
    <ChartContainer config={config} className={className}>
      <Chart data={data} margin={{ left: 4, right: hasRight ? 4 : 12, top: 8 }} accessibilityLayer>
        <CartesianGrid vertical={false} />
        <XAxis
          dataKey="bucket"
          tickLine={false}
          axisLine={false}
          tickMargin={8}
          minTickGap={32}
          tickFormatter={(v: string) => formatBucket(v, bucket)}
        />
        <YAxis
          yAxisId="left"
          tickLine={false}
          axisLine={false}
          width={56}
          reversed={reversed}
          domain={reversed ? [1, "auto"] : ["auto", "auto"]}
          allowDecimals={format !== "rank"}
          tickFormatter={(v: number) => FORMATTERS[format].axis(v)}
        />
        {hasRight && (
          <YAxis
            yAxisId="right"
            orientation="right"
            tickLine={false}
            axisLine={false}
            width={56}
            domain={["auto", "auto"]}
            tickFormatter={(v: number) => FORMATTERS[rightFormat].axis(v)}
          />
        )}
        <ChartTooltip
          content={
            <ChartTooltipContent
              labelFormatter={(_, payload) => {
                const iso = payload?.[0]?.payload?.bucket
                return typeof iso === "string" ? formatBucket(iso, bucket, "tooltip") : null
              }}
              formatter={(value, name, item) => (
                <div className="flex w-full items-center gap-2">
                  <div className="size-2.5 shrink-0 rounded-[2px]" style={{ background: item.color }} />
                  <span className="text-muted-foreground">{config[String(name)]?.label ?? name}</span>
                  <span className="ml-auto font-mono font-medium text-foreground tabular-nums">
                    {typeof value === "number" ? formatFor(String(name)).full(value) : String(value)}
                  </span>
                </div>
              )}
            />
          }
        />
        {legend && (
          <ChartLegend
            content={<ChartLegendContent />}
            itemSorter={(item) => series.findIndex((s) => s.key === item.dataKey)}
          />
        )}
        {series.map((s) =>
          kind === "area" ? (
            <Area
              key={s.key}
              yAxisId={s.right ? "right" : "left"}
              dataKey={s.key}
              type="monotone"
              stroke={`var(--color-${s.key})`}
              fill={`var(--color-${s.key})`}
              fillOpacity={0.15}
              strokeWidth={2}
              connectNulls
            />
          ) : (
            <Line
              key={s.key}
              yAxisId={s.right ? "right" : "left"}
              dataKey={s.key}
              type="monotone"
              stroke={`var(--color-${s.key})`}
              strokeWidth={2}
              dot={data.length <= 2}
              connectNulls
            />
          ),
        )}
      </Chart>
    </ChartContainer>
  )
}
