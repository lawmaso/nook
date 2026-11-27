import { format } from "date-fns"
import { SwordsIcon } from "lucide-react"
import type { Metadata } from "next"
import { connection } from "next/server"
import Link from "next/link"
import { notFound } from "next/navigation"

import { ChartCard, summarize } from "@/components/charts/chart-card"
import { MetricChart, type ChartRow } from "@/components/charts/metric-chart"
import { DateRangePicker } from "@/components/date-range-picker"
import { KindSwitch } from "@/components/kind-switch"
import { MigrationNotice } from "@/components/migration-notice"
import { RelativeTime } from "@/components/relative-time"
import { buttonVariants } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { parseRange, pickBucket, rangeBounds } from "@/lib/dates"
import { formatLarge, percent, secondsLabel } from "@/lib/format"
import { isKind, KIND_LABELS, KINDS, type Kind } from "@/lib/leaderboards"
import { getUserHistory, getUserProfile, type UserProfile } from "@/lib/queries"
import { first, kindHref } from "@/lib/search-params"
import { cn } from "@/lib/utils"

function parseUid(raw: string): number | null {
  if (!/^\d{1,10}$/.test(raw)) return null
  const uid = Number(raw)
  return Number.isSafeInteger(uid) && uid <= 2_147_483_647 ? uid : null
}

export async function generateMetadata({ params }: PageProps<"/u/[uid]">): Promise<Metadata> {
  const uid = parseUid((await params).uid)
  const profile = uid === null ? null : await getUserProfile(uid)
  return { title: profile?.user.username ?? "User" }
}

function StandingCard({ kind, standing }: { kind: Kind; standing: UserProfile["current"][Kind] }) {
  return (
    <Card size="sm">
      <CardHeader>
        <CardDescription>{KIND_LABELS[kind]}</CardDescription>
        <CardTitle className="font-heading text-2xl tabular-nums">
          {standing ? `#${formatLarge(standing.rank)}` : "Unranked"}
        </CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-3 text-sm">
        {standing ? (
          <>
            <dl className="grid grid-cols-2 gap-x-4 gap-y-1 tabular-nums">
              <dt className="text-muted-foreground">Points</dt>
              <dd>{formatLarge(standing.total_points)}</dd>
              <dt className="text-muted-foreground">Solved</dt>
              <dd>
                {formatLarge(standing.solved)} / {formatLarge(standing.attempted)}
              </dd>
              <dt className="text-muted-foreground">Acceptance</dt>
              <dd>{percent(standing.acceptance_rate)}</dd>
              <dt className="text-muted-foreground">Avg time</dt>
              <dd>{secondsLabel(standing.avg_time_s)}</dd>
              <dt className="text-muted-foreground">Last active</dt>
              <dd>
                <RelativeTime iso={standing.last_active} />
              </dd>
            </dl>
            <Link
              href={`/leaderboard/${kind}?a=${standing.uid}`}
              className={cn(buttonVariants({ variant: "outline", size: "sm" }), "self-start")}
            >
              <SwordsIcon /> Compare on {KIND_LABELS[kind].toLowerCase()} board
            </Link>
          </>
        ) : (
          <p className="text-muted-foreground">Not on the latest {kind} snapshot.</p>
        )}
      </CardContent>
    </Card>
  )
}

export default async function UserPage({ params, searchParams }: PageProps<"/u/[uid]">) {
  const uid = parseUid((await params).uid)
  if (uid === null) notFound()

  const profile = await getUserProfile(uid)
  if (!profile) notFound()

  const sp = await searchParams
  await connection() // default date range depends on "now"
  const requested = first(sp.kind)
  const kind: Kind = isKind(requested) ? requested : profile.current.monthly ? "monthly" : "global"
  const range = parseRange({ from: first(sp.from), to: first(sp.to) }, kind)
  const bucket = pickBucket(range)
  const { start, end } = rangeBounds(range)

  const { rows, missingRpc } = await getUserHistory(uid, kind, start, end, bucket)
  const data: ChartRow[] = rows.map((r) => ({
    bucket: r.bucket,
    rank: r.rank,
    points: r.total_points,
    solved: r.solved,
    attempted: r.attempted,
    acceptance: r.acceptance_rate,
    avgTime: r.avg_time_s,
    avgPoints: r.avg_points,
  }))
  const col = (key: keyof ChartRow) => data.map((d) => d[key] as number | null)

  const { user } = profile

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-1">
        <h1 className="font-heading text-2xl font-semibold tracking-tight">{user.username}</h1>
        <p className="text-sm text-muted-foreground">
          uid {user.uid}
          {user.member_since && <> · member since {format(`${user.member_since}T00:00:00`, "MMMM d, yyyy")}</>}
        </p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        {KINDS.map((k) => (
          <StandingCard key={k} kind={k} standing={profile.current[k]} />
        ))}
      </div>

      <div className="flex flex-wrap items-center justify-between gap-2">
        <h2 className="font-heading text-lg font-semibold">History</h2>
        <div className="flex flex-wrap items-center gap-2">
          <KindSwitch active={kind} hrefFor={(k) => kindHref(`/u/${uid}`, k, sp)} />
          <DateRangePicker range={range} />
        </div>
      </div>

      {missingRpc ? (
        <MigrationNotice />
      ) : (
        <div className="grid gap-6 lg:grid-cols-2">
          <ChartCard
            title="Rank"
            description={`Position on the ${kind} board`}
            summary={summarize(col("rank"), (n) => `#${formatLarge(n)}`, false)}
          >
            <MetricChart data={data} bucket={bucket} format="rank" reversed series={[{ key: "rank", label: "Rank" }]} />
          </ChartCard>

          <ChartCard title="Total points" summary={summarize(col("points"), formatLarge)}>
            <MetricChart
              data={data}
              bucket={bucket}
              kind="area"
              series={[{ key: "points", label: "Points", color: "var(--chart-2)" }]}
            />
          </ChartCard>

          <ChartCard title="Solved vs attempted" summary={summarize(col("solved"), formatLarge)}>
            <MetricChart
              data={data}
              bucket={bucket}
              series={[
                { key: "solved", label: "Solved" },
                { key: "attempted", label: "Attempted" },
              ]}
            />
          </ChartCard>

          <ChartCard
            title="Efficiency"
            description="Acceptance rate and average solve time"
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
