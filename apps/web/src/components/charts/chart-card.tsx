import { TrendingDownIcon, TrendingUpIcon } from "lucide-react"

import { Card, CardAction, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { cn } from "@/lib/utils"

export type Summary = {
  value: string
  /** Signed change over the range, already formatted (e.g. "+12,340"). */
  delta?: string
  direction?: "up" | "down" | "flat"
  /** Whether "up" is good (green) for this metric. */
  upIsGood?: boolean
}

export function ChartCard({
  title,
  description,
  summary,
  className,
  children,
}: {
  title: string
  description?: string
  summary?: Summary
  className?: string
  children: React.ReactNode
}) {
  const good =
    summary?.direction === "flat" || !summary?.direction
      ? null
      : (summary.direction === "up") === (summary.upIsGood ?? true)

  return (
    <Card className={cn("min-w-0", className)}>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
        {description && <CardDescription>{description}</CardDescription>}
        {summary && (
          <CardAction className="text-right">
            <div className="font-heading text-lg font-semibold tabular-nums">{summary.value}</div>
            {summary.delta && (
              <div
                className={cn(
                  "flex items-center justify-end gap-1 text-xs tabular-nums",
                  good === null ? "text-muted-foreground" : good ? "text-emerald-600 dark:text-emerald-400" : "text-destructive",
                )}
              >
                {summary.direction === "up" && <TrendingUpIcon className="size-3.5" />}
                {summary.direction === "down" && <TrendingDownIcon className="size-3.5" />}
                {summary.delta}
              </div>
            )}
          </CardAction>
        )}
      </CardHeader>
      <CardContent>{children}</CardContent>
    </Card>
  )
}

/** Summarize a series as its last value and change since the first value. */
export function summarize(
  values: (number | null | undefined)[],
  format: (n: number) => string,
  upIsGood = true,
): Summary | undefined {
  const present = values.filter((v): v is number => typeof v === "number" && Number.isFinite(v))
  if (present.length === 0) return undefined
  const first = present[0]
  const last = present[present.length - 1]
  const diff = last - first
  return {
    value: format(last),
    delta: present.length < 2 ? undefined : diff === 0 ? "no change" : `${diff > 0 ? "+" : "−"}${format(Math.abs(diff))}`,
    direction: diff > 0 ? "up" : diff < 0 ? "down" : "flat",
    upIsGood,
  }
}
