import { format } from "date-fns"
import { ClockIcon, TimerIcon, UsersIcon } from "lucide-react"
import type { Metadata } from "next"
import { notFound } from "next/navigation"
import { Suspense } from "react"

import { KindSwitch } from "@/components/kind-switch"
import { ComparePanel } from "@/components/leaderboard/compare-panel"
import { LeaderboardTable } from "@/components/leaderboard/leaderboard-table"
import { RelativeTime } from "@/components/relative-time"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { formatLarge } from "@/lib/format"
import { isKind, KIND_LABELS, KINDS } from "@/lib/leaderboards"
import { getLatestLeaderboard } from "@/lib/queries"

export function generateStaticParams() {
  return KINDS.map((kind) => ({ kind }))
}

export async function generateMetadata({ params }: PageProps<"/leaderboard/[kind]">): Promise<Metadata> {
  const { kind } = await params
  return { title: isKind(kind) ? `${KIND_LABELS[kind]} leaderboard` : "Leaderboard" }
}

export default async function LeaderboardPage({ params }: PageProps<"/leaderboard/[kind]">) {
  const { kind } = await params
  if (!isKind(kind)) notFound()

  const { scrape, entries } = await getLatestLeaderboard(kind)
  const periodLabel = scrape?.period ? format(`${scrape.period}T00:00:00`, "MMMM yyyy") : "All time"

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div className="flex flex-col gap-1">
          <h1 className="font-heading text-2xl font-semibold tracking-tight">{KIND_LABELS[kind]} leaderboard</h1>
          <p className="text-sm text-muted-foreground">{periodLabel}</p>
        </div>
        <KindSwitch active={kind} hrefFor={(k) => `/leaderboard/${k}`} />
      </div>

      {scrape ? (
        <div className="flex flex-wrap gap-2">
          <Badge variant="secondary">
            <ClockIcon /> Updated <RelativeTime iso={scrape.scraped_at} />
          </Badge>
          <Badge variant="secondary">
            <UsersIcon /> {formatLarge(entries.length)} users
          </Badge>
          <Badge variant="outline">
            <TimerIcon /> Scraped in {formatLarge(scrape.duration_ms)} ms
          </Badge>
        </div>
      ) : null}

      {entries.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center text-sm text-muted-foreground">
            No {kind} scrapes yet. Start the daemon (<code>go run ./cmd/daemon</code>) to populate this board.
          </CardContent>
        </Card>
      ) : (
        <div className="grid items-start gap-6 lg:grid-cols-[minmax(0,1fr)_380px]">
          <Card className="min-w-0">
            <CardHeader>
              <CardTitle>Standings</CardTitle>
            </CardHeader>
            <CardContent>
              <LeaderboardTable entries={entries} />
            </CardContent>
          </Card>
          <div className="lg:sticky lg:top-20">
            <Suspense fallback={<Skeleton className="h-[560px] w-full rounded-xl" />}>
              <ComparePanel entries={entries} />
            </Suspense>
          </div>
        </div>
      )}
    </div>
  )
}
